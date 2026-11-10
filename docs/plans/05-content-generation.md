# 05 — Content Generation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task by task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** After the segment step, Claude writes the SMS and the email for each language of the audience. The text streams onto the page. The backend validates it against the limits after it renders `{{first_name}}`.

**Architecture:** This slice adds one package, one LLM method and one endpoint:

```
internal/content   ParseVars, Render, Validate, Limits, ToolSchema, SystemPrompt,
                   UserMessage, RepairMessage, LanguageOf, MockContent, PartialContent
internal/llm       Client.Stream, Response.CutOff, Mock streaming and default function
internal/segment   LanguageCounts.MaxFirstName (json:"-"), filled by CountAudience
internal/httpapi   POST /api/content/generate (ContentGenerator), SSE writer
cmd/api            wires the content generator and the mock default function
frontend           lib/content.ts (stream reader), content-cards.vue, auto-start in brief-audience.vue
```

`internal/content` imports nothing from `internal/`. `internal/httpapi` joins `content`, `segment` and `llm`. `cmd/api` gives the mock its default answers (D-034).

**Tech Stack:** Go 1.27, `github.com/anthropics/anthropic-sdk-go/v2` v2.8.0 (beta streaming), `pgx/v5`, Vue 3, native `fetch` with `ReadableStream`, Vitest.

**Spec:** `docs/specs/05-content-generation/goal.md` and `requirements.md`. Decisions D-030 to D-034.

## Global Constraints

- Limits after rendering with the longest first name: SMS 160 runes, subject 80 runes, body 2,000 runes. Count runes, not bytes.
- The only variable is `{{first_name}}`, exactly, with no spaces. Each `{` or `}` outside it fails validation.
- No control characters except `\n` and `\t` in `email_body`. No `\n` or `\r` in `email_subject`. No HTML tag: `<` followed by a letter, `/` or `!`.
- Tool `submit_content` with `strict: true` and `eager_input_streaming: true`. Model, effort `low`, `tool_choice: auto` and fallbacks as in slice 04 (D-027, D-033).
- One LLM call for each language, at the same time (D-030). Each call has its own `45s` timeout. One repair round for each language.
- Request body at most 16 KiB. Audience query `statement_timeout` `2s`.
- SSE events: `audience`, `draft`, `retry`, `final`, `error`, `done`. At most one `draft` per 100 ms for each language.
- Error codes before the stream: `invalid_request`, `invalid_brief`, `invalid_rule`, `empty_audience`, `db_unavailable`. Error codes in the stream: `content_invalid`, `llm_error`, `llm_refused`, `llm_timeout`.
- Logs: language, attempts, validation errors, rendered rune counts. Never the brief, the content text or the API key.
- No user row leaves the database. The content step uses only `LanguageCounts`.
- Tests use `llm.Mock`. No test reads `ANTHROPIC_API_KEY`. Postgres tests use `storetest.NewDB` and skip without `TEST_DATABASE_URL`.
- Do not commit after each task. The user commits the slice after review.

## Review Focus

1. **The rule that the browser sends back is the JSON that slice 04 wrote.** `Rule.MarshalJSON` output must pass `ParseRule` with no change, also for `is_null` conditions. Pinned by Task 5 `TestGenerate_RuleRoundTrip`.
2. **A long Turkish first name, for example `Şükrü Kağan`.** Postgres `char_length` and the Go rune count must agree on it. Pinned by Task 4 `TestCountAudience_MaxFirstNameRunes`, which inserts a user with that name.
3. **Claude writes `{first_name}` with single braces, or `{{FirstName}}`.** The text would go out with a literal brace. The parser rejects each `{` or `}` that is not part of `{{first_name}}`. Pinned by Task 1 `TestParseVars` cases `{first_name}` and `{{FirstName}}`.
4. **The operator clicks "Find audience" again while content streams.** Events of the old stream must not reach the new cards. Pinned by Task 8 `ignores events from an aborted stream`.
5. **The connection closes before `done`.** The page must not wait forever. Each card with no `final` or `error` shows "The connection closed early. Try again." Pinned by Task 7 `stream ends without done` and Task 8 `marks open cards on early close`.

---

### Task 1: Variables, renderer and validator

**Files:**
- Create: `backend/internal/content/content.go`, `vars.go`, `validate.go`, `vars_test.go`, `validate_test.go`

**Interfaces:**
- Produces:

```go
package content

const (
    MaxSMSRunes     = 160
    MaxSubjectRunes = 80
    MaxBodyRunes    = 2000
    VarFirstName    = "{{first_name}}"
)

type Content struct {
    SMS          string `json:"sms"`
    EmailSubject string `json:"email_subject"`
    EmailBody    string `json:"email_body"`
}

// Limits holds the rendered rune counts that the final event reports.
type Limits struct {
    FirstNameRunes    int `json:"first_name_runes"`
    SMSRunes          int `json:"sms_runes"`
    EmailSubjectRunes int `json:"email_subject_runes"`
    EmailBodyRunes    int `json:"email_body_runes"`
}

type FieldError struct {
    Path    string `json:"path"`
    Message string `json:"message"`
}

// ParseVars checks the variables of one field. It returns the count of {{first_name}}.
func ParseVars(s string) (int, error)
// Render replaces each {{first_name}} with name. No escaping.
func Render(s, name string) string
// Validate decodes the tool input, trims each field and checks it.
// The check renders with a name of firstNameRunes runes.
func Validate(raw []byte, firstNameRunes int) (Content, Limits, []FieldError)
```

- [ ] **Step 1: Write the failing tests**
  - `TestParseVars` (table): `""` → 0. `"Hi {{first_name}}"` → 1. Two variables → 2. Errors for `{{last_name}}`, `{{ first_name }}`, `{{first_name}`, `{{`, `}}`, `{first_name}`, `{{FirstName}}`, `a { b`.
  - `TestRender`: one and two variables, no variable, a name with `<b>` stays as written.
  - `TestValidate_Limits` with `firstNameRunes` 20: SMS of 153 runes with one variable passes (160 rendered). 154 runes fails with the message `161 runes after rendering, limit 160`. Subject at 80 rendered passes, 81 fails. Body at 2,000 passes, 2,001 fails. A field of `ş` runes counts each `ş` as 1.
  - `TestValidate_Cases` (table, one or more for each row of "Validation" in the spec): unknown key, missing field, a number for a field, empty after trim, `\n` in the subject, `\x07` in the SMS, `\t` in the body passes, `<b>` in each field, `</a>`, `<!--`, `a < b` passes, data after the object, not an object.
  - `TestValidate_AllErrors`: an input with three bad fields gives three errors with paths `sms`, `email_subject`, `email_body`.
  - `TestValidate_Trim`: leading and trailing spaces are gone in the returned `Content`. `Limits` has the rendered counts.

- [ ] **Step 2: Run the tests to make sure they fail**

Run: `cd backend && go test ./internal/content/`
Expected: FAIL, the package does not compile.

- [ ] **Step 3: Implement** the three functions. `Validate` decodes with `json.Decoder` and `DisallowUnknownFields`, then checks for data after the object, as slice 04 does. It collects all errors. The rendered name is `strings.Repeat("x", firstNameRunes)`.

- [ ] **Step 4: Run the tests to make sure they pass**

Run: `cd backend && go test ./internal/content/`
Expected: PASS.

---

### Task 2: Prompt, tool schema, mock content and partial parser

**Files:**
- Create: `backend/internal/content/prompt.go`, `mock.go`, `partial.go`, `prompt_test.go`, `mock_test.go`, `partial_test.go`

**Interfaces:**
- Consumes: Task 1 `Content`, `Validate`, `FieldError`, the limit constants.
- Produces:

```go
const ToolName = "submit_content"

func ToolSchema() map[string]any // three required strings, additionalProperties false, no length limits
func SystemPrompt() string       // constant text, no brief, no language
// UserMessage gives the language line, the description, the name length and the brief in <brief> tags.
func UserMessage(lang, description string, firstNameRunes int, brief string) string
func RepairMessage(errs []FieldError) string
// LanguageOf reads the language code back from a UserMessage. ok is false for other text.
func LanguageOf(userMessage string) (lang string, ok bool)
// MockContent returns the default tool input JSON for "tr" or "en". Other codes get "en".
func MockContent(lang string) []byte
// PartialContent parses a prefix of the tool input JSON. It never fails.
func PartialContent(prefix []byte) Content
```

- [ ] **Step 1: Write the failing tests**
  - `TestSystemPrompt`: contains `160`, `80`, `{{first_name}}`, `plain text`, `<brief>`. Does not contain the example brief. Two calls give the same string.
  - `TestUserMessage`: contains `tr`, `Turkish`, the description, `20`, and `<brief>\n…\n</brief>`. `LanguageOf(UserMessage("en", …))` gives `"en", true`. `LanguageOf("<brief>x</brief>")` gives `false`.
  - `TestToolSchema`: `required` is the three keys. `additionalProperties` is false. No `maxLength`.
  - `TestMockContent`: `Validate(MockContent("tr"), 30)` and `Validate(MockContent("en"), 30)` give no errors (FR-24). The `tr` SMS contains `%30`.
  - `TestPartialContent`: take the JSON of a `Content` with `ş`, `\n`, `\"` and `ç` in it. For each byte prefix, `PartialContent` does not panic. Each field is a prefix of the final field. The full input gives the final `Content` (FR-19). A prefix that ends inside `\u00` or inside a multi-byte rune drops the cut part.

- [ ] **Step 2: Run the tests to make sure they fail**

Run: `cd backend && go test ./internal/content/`
Expected: FAIL.

- [ ] **Step 3: Implement.**
  - Write the system prompt and the user message as the spec "Prompt" section says. The language names are `tr` → Turkish, `en` → English. The first user message line is `Language: <code> (<name>)`. `LanguageOf` reads that line.
  - Write the `tr` and `en` mock content for the example brief: a cart reminder with `%30` off, warm and short. Keep the SMS under 120 runes before rendering.
  - `PartialContent` is a small scanner over the three known keys. It closes an open string at the cut point and drops an incomplete escape or rune. Do not use a general JSON library for prefixes.

- [ ] **Step 4: Run the tests to make sure they pass**

Run: `cd backend && go test ./internal/content/`
Expected: PASS.

---

### Task 3: LLM streaming and the mock

**Files:**
- Modify: `backend/internal/llm/llm.go`, `mock.go`, `mock_test.go`, `anthropic.go`, `anthropic_test.go`

**Interfaces:**
- Consumes: the slice 04 `Request`, `Response`, `Tool`, `Turn`.
- Produces:

```go
// StreamEvent is one event of a streaming call.
type StreamEvent struct {
    Delta string // a fragment of the tool input JSON, or
    Reset bool   // a fallback block started; drop the fragments so far
}

type Response struct {
    ToolUse   *ToolUse
    Assistant any
    Refused   bool
    CutOff    bool // new: stop_reason == max_tokens
}

type Client interface {
    Generate(ctx context.Context, req Request) (Response, error)
    Stream(ctx context.Context, req Request, on func(StreamEvent)) (Response, error)
}

type Tool struct {
    Name, Description string
    Schema            map[string]any
    Eager             bool // new: eager_input_streaming
}

// MockDefault gives the answer of an unscripted mock call.
type MockDefault func(req Request) MockStep
func NewMock(steps ...MockStep) *Mock          // unchanged behaviour
func NewMockFunc(f MockDefault) *Mock          // calls f for each request; safe for concurrent use
func (m *Mock) SetStream(chunk int, delay time.Duration) // default chunk 16 bytes, delay 0
func New(cfg config.Config, mockDefault MockDefault) (Client, error) // nil keeps ExampleRuleJSON
```

- [ ] **Step 1: Write the failing tests**
  - `TestMock_Stream`: a step with a 40-byte input and chunk 16 gives 3 `Delta` events. Their join is the input. The returned `Response` is the step response.
  - `TestMock_StreamCancel`: a cancelled context returns `ctx.Err()` and records the request.
  - `TestMock_Func`: `NewMockFunc` calls the function with each request, from 20 goroutines, under `-race`.
  - `TestMock_Default` stays green: `NewMock()` still returns `ExampleRuleJSON`.
  - `TestAnthropic_BuildParams_Eager`: with `Tool.Eager`, the tool has `EagerInputStreaming` true and `Strict` true.
  - `TestAnthropic_ParseResponse_CutOff`: `stop_reason: max_tokens` gives `CutOff`.
  - `TestAnthropic_ParseResponse_Fallback`: a fixed body with a `tool_use`, then a `fallback` block, then a second `tool_use`. `ToolUse` is the second one. `Assistant` has no block from before the `fallback` block, except text.
  - `TestAnthropic_AssistantInvalidInput`: a `tool_use` whose raw input is not a JSON object gives an `Assistant` param with `{}` as the input.

- [ ] **Step 2: Run the tests to make sure they fail**

Run: `cd backend && go test ./internal/llm/`
Expected: FAIL.

- [ ] **Step 3: Implement.**
  - `Stream` on the Anthropic client uses `c.client.Beta.Messages.NewStreaming` with the same params. Accumulate each event into an `anthropic.BetaMessage` with `Accumulate`. For each `input_json_delta`, call `on(StreamEvent{Delta: …})`. For a `content_block_start` of type `fallback`, call `on(StreamEvent{Reset: true})`. Let the compiler give the type names.
  - Move the fallback filter and the `{}` input fix into `parseResponse`, so that `Generate` and `Stream` share it.
  - Update `cmd/api` and each caller of `llm.New` for the new parameter. Pass `nil` until Task 6.

- [ ] **Step 4: Run the tests to make sure they pass**

Run: `cd backend && go test -race ./internal/llm/ ./internal/httpapi/ ./cmd/...`
Expected: PASS. The slice 04 tests still pass.

---

### Task 4: Longest first name in the audience query

**Files:**
- Modify: `backend/internal/segment/audience.go`, `audience_test.go`

**Interfaces:**
- Produces: `LanguageCounts.MaxFirstName int64` with tag `json:"-"`. Thus the preview response of slice 04 does not change.

- [ ] **Step 1: Write the failing tests**
  - `TestCountAudience_MaxFirstName` (Postgres): use the seed users and the example rule. The test calculates the maximum `utf8.RuneCountInString(FirstName)` in Go, over the users with consent on one or more channels. Each language's `MaxFirstName` is equal to it (FR-12).
  - `TestCountAudience_MaxFirstNameRunes` (Postgres): insert a reachable user with first name `Şükrü Kağan`. A rule that selects only this user gives `MaxFirstName` 11.
  - `TestPreview_NoMaxFirstName` in `internal/httpapi`: the preview JSON has no key `MaxFirstName` or `max_first_name`.

- [ ] **Step 2: Run the tests to make sure they fail**

Run: `make test`
Expected: FAIL in the new tests.

- [ ] **Step 3: Implement.** Add `coalesce(max(char_length(first_name)) FILTER (WHERE email_consent OR sms_consent), 0)` to the query and scan it.

- [ ] **Step 4: Run the tests to make sure they pass**

Run: `make test`
Expected: PASS.

---

### Task 5: Language worker and the content endpoint

**Files:**
- Create: `backend/internal/httpapi/content.go`, `sse.go`, `content_test.go`, `sse_test.go`

**Interfaces:**
- Consumes: Task 1 `Validate`, `Content`, `Limits`. Task 2 `ToolName`, `ToolSchema`, `SystemPrompt`, `UserMessage`, `RepairMessage`, `PartialContent`. Task 3 `Client.Stream`, `StreamEvent`, `Response.CutOff`, `Tool.Eager`. Task 4 `MaxFirstName`. Slice 04 `segment.SanitizeBrief`, `Allowlist.ParseRule`, `Allowlist.Compile`, `Allowlist.Describe`, `AudienceCounter`.
- Produces:

```go
type ContentGenerator struct {
    LLM          llm.Client
    Allow        segment.Allowlist
    Count        AudienceCounter
    Now          func() time.Time           // rule clock
    Clock        func() time.Time           // draft throttle clock; tests inject it
    LLMTimeout   time.Duration              // 45s in cmd/api
    DraftEvery   time.Duration              // 100ms in cmd/api
}
func RegisterContent(mux *http.ServeMux, g *ContentGenerator) // POST /api/content/generate

// sseWriter serialises events from many goroutines and flushes after each one.
type sseWriter struct{ /* mu, w, flusher */ }
func newSSEWriter(w http.ResponseWriter) (*sseWriter, bool)
func (s *sseWriter) send(event string, data any) error
```

- [ ] **Step 1: Write the failing tests.** Use `httptest`, a fake counter and a fake clock. Use `llm.NewMockFunc` with a function that picks the answer by `content.LanguageOf(req.Turns[0].User)` and the attempt number. A helper parses the response body into a list of `(event, data)` pairs.
  - `TestGenerate_Example` (FR-1): events are `audience`, one or more `draft`, `final` for `tr`, `done`. Content-Type is `text/event-stream`. `Cache-Control` is `no-cache`.
  - `TestGenerate_BadBody`, `TestGenerate_BadBrief`, `TestGenerate_BadRule`, `TestGenerate_EmptyAudience` (FR-2 to FR-5): JSON errors with the codes. The mock has no requests. For `invalid_rule`, the counter has no calls.
  - `TestGenerate_RuleRoundTrip`: `json.Marshal` of the slice 04 example rule is accepted.
  - `TestGenerate_TwoLanguages` (FR-6, FR-18): the counter returns `tr` and `en`. The mock fails `en` with an error. The stream has `final` for `tr`, `error` with `llm_error` for `en`, then `done`. Each request's user message names its language.
  - `TestGenerate_Repair` (FR-13, FR-14, FR-15, FR-16): four sub-tests. Bad then good gives `draft`, `retry`, `draft`, `final`. The second request has the assistant turn and a tool result with `IsError`. Bad twice gives `error` with `content_invalid`. No tool call gives a repair request that ends with a user text. `CutOff` gives a repair round.
  - `TestGenerate_LLMErrors` (FR-17): error, `context.DeadlineExceeded` and `Refused` give `llm_error`, `llm_timeout` and `llm_refused`. A refusal gives one request only.
  - `TestGenerate_Throttle` (FR-20): the mock sends 10 fragments while the fake clock does not move. The stream has one `draft` before the last, and the last `draft` has the full text.
  - `TestGenerate_Reset` (FR-30): a `Reset` event gives a `retry` event with no errors. The language still gets one repair round after it.
  - `TestGenerate_Disconnect` (FR-21): cancel the request context during the stream. The mock function sees `ctx.Err() != nil`.
  - `TestGenerate_LogsNoText` (NFR-2): capture `slog` output. It does not contain the brief or any content text.
  - `TestSSEWriter`: two goroutines send 100 events each. The output has 200 whole events. Each event is `event: x\ndata: {...}\n\n`.

- [ ] **Step 2: Run the tests to make sure they fail**

Run: `cd backend && go test ./internal/httpapi/`
Expected: FAIL.

- [ ] **Step 3: Implement.**
  - The handler does the backend steps 1 to 9 of the spec "Flow" in sequence. Reuse `decodeBody` and `writeError`.
  - The `audience` event sends the description and the slice 04 language entries that have `Reachable > 0`.
  - Run one goroutine for each language with `errgroup` semantics but no shared cancel: one language's failure must not cancel the others. Cancel all on request context end.
  - Each attempt has its own `context.WithTimeout(ctx, LLMTimeout)`.
  - The worker keeps the fragments in a buffer. On each delta, it sends a `draft` with `PartialContent(buf)` if `Clock()` moved `DraftEvery` since the last draft. When the call ends, it sends one last `draft` before `final` or `retry`.
  - A `retry` from `Reset` clears the buffer and does not use the repair round.
  - Log one line for each language at the end, with the fields in Global Constraints.

- [ ] **Step 4: Run the tests to make sure they pass**

Run: `cd backend && go test -race ./internal/httpapi/`
Expected: PASS.

---

### Task 6: Wiring in cmd/api

**Files:**
- Modify: `backend/cmd/api/main.go`
- Create: `backend/cmd/api/mockdefault.go`, `mockdefault_test.go`

**Interfaces:**
- Consumes: Task 3 `llm.New`, `MockDefault`, `SetStream`. Task 2 `content.LanguageOf`, `content.MockContent`. Task 5 `RegisterContent`.
- Produces: `func mockDefault(req llm.Request) llm.MockStep`. For `segment.ToolName` it returns `llm.ExampleRuleJSON`. For `content.ToolName` it returns `content.MockContent(lang)`.

- [ ] **Step 1: Write the failing test** `TestMockDefault`: a segment request gets the example rule. A content request with a `tr` user message gets JSON that passes `content.Validate(…, 30)`. An `en` one does too.

- [ ] **Step 2: Run the test to make sure it fails**

Run: `cd backend && go test ./cmd/api/`
Expected: FAIL.

- [ ] **Step 3: Implement.** Pass `mockDefault` to `llm.New`. In mock mode, call `SetStream(16, 30*time.Millisecond)`, so that the demo shows the stream. Register the content generator with `LLMTimeout: 45 * time.Second` and `DraftEvery: 100 * time.Millisecond`. Keep `http.Server` with no `WriteTimeout` (NFR-9). Add a comment that says why.

- [ ] **Step 4: Run the tests to make sure they pass**

Run: `make test`
Expected: PASS.

---

### Task 7: Frontend stream reader

**Files:**
- Create: `frontend/src/lib/content.ts`, `frontend/src/lib/__tests__/content.test.ts`

**Interfaces:**
- Produces:

```ts
export interface Content { sms: string, email_subject: string, email_body: string }
export interface ContentLimits { first_name_runes: number, sms_runes: number, email_subject_runes: number, email_body_runes: number }
export type ContentEvent =
  | { event: 'audience', data: { description: string, languages: (Counts & { code: string })[] } }
  | { event: 'draft', data: { language: string, content: Content } }
  | { event: 'retry', data: { language: string, errors?: { path: string, message: string }[] } }
  | { event: 'final', data: { language: string, content: Content, limits: ContentLimits } }
  | { event: 'error', data: { language: string, code: string, message: string } }
  | { event: 'done', data: Record<string, never> }

export class ContentRequestError extends Error { code: string }
// Calls POST {API_BASE_URL}/content/generate. A non-200 JSON error throws ContentRequestError.
// It throws ContentRequestError('stream_closed') if the stream ends before done.
export async function generateContent(brief: string, rule: unknown, onEvent: (e: ContentEvent) => void, signal: AbortSignal): Promise<void>
export function parseSSE(chunk: string, carry: string): { events: ContentEvent[], carry: string }
```

- [ ] **Step 1: Write the failing tests**
  - `parseSSE splits events across chunks`: an event cut in the middle of `data:` comes out whole on the next chunk. Two events in one chunk both come out.
  - `generateContent posts brief and rule`: mock `fetch`, check the URL, the method and the JSON body.
  - `generateContent throws the API message on 400`: the error has the code and the message from the body.
  - `stream ends without done`: the promise rejects with code `stream_closed`.
  - `abort stops reading`: after `abort()`, no more `onEvent` calls.

- [ ] **Step 2: Run the tests to make sure they fail**

Run: `cd frontend && pnpm test:run src/lib/__tests__/content.test.ts`
Expected: FAIL.

- [ ] **Step 3: Implement** with `fetch`, `response.body.getReader()` and `TextDecoder` with `stream: true`. Use `API_BASE_URL`. Set no total timeout. The `AbortSignal` ends the request.

- [ ] **Step 4: Run the tests to make sure they pass**

Run: `cd frontend && pnpm test:run src/lib/__tests__/content.test.ts`
Expected: PASS.

---

### Task 8: Content cards and the automatic start

**Files:**
- Create: `frontend/src/components/campaign/content-cards.vue`, `frontend/src/components/campaign/__tests__/content-cards.test.ts`
- Modify: `frontend/src/components/campaign/brief-audience.vue`, `__tests__/brief-audience.test.ts`, `frontend/src/pages/campaigns/new.vue` (subtitle: "Write a brief. Claude finds the audience and writes the messages.")

**Interfaces:**
- Consumes: Task 7 `generateContent`, `ContentEvent`, `ContentRequestError`.
- Produces: `<ContentCards :brief :rule :run-id />`. A new `runId` aborts the open request, clears the cards and starts again. `brief-audience.vue` sets `runId` after each segment result with one or more languages.

- [ ] **Step 1: Write the failing tests** (mock `@/lib/content`)
  - `starts after a segment result` (FR-25): after a preview with `languages: [tr]`, `generateContent` is called once with the brief and the rule. "Generating content…" shows. No call for a preview with no languages.
  - `streams drafts and shows final` (FR-26): `draft` text shows with the draft style. `final` shows `121 / 160` and the subject counter.
  - `retry and error` (FR-26): `retry` shows "Fixing content…". `error` on `en` shows its message, and the `tr` card stays.
  - `regenerate` (FR-27): the button calls `generateContent` again with the same brief and rule. It is off while a request runs.
  - `ignores events from an aborted stream`: start run 1, then run 2. Events that run 1's `onEvent` sends after that do not change the cards.
  - `marks open cards on early close`: `generateContent` rejects with `stream_closed`. Each card with no `final` shows "The connection closed early. Try again."
  - `shows text, not HTML` (FR-28): a draft with `<b>x</b>` shows the tags as text.
  - `error before the stream` (FR-29): a `ContentRequestError` shows its message. The audience panel stays.
  - `Find audience is off while content runs`.

- [ ] **Step 2: Run the tests to make sure they fail**

Run: `cd frontend && pnpm test:run src/components/campaign`
Expected: FAIL.

- [ ] **Step 3: Implement** with the existing `Card`, `Alert`, `Badge` and `Button` components. Use text interpolation only, never `v-html`. Show the language name (`tr` → Turkish, `en` → English) and the email and SMS counts in the card title. Use `whitespace-pre-wrap` for the body. Keep the brief and the rule of the last segment result in `brief-audience.vue`.

- [ ] **Step 4: Run the checks**

Run: `cd frontend && pnpm test:run && pnpm build && CI=true pnpm lint`
Expected: all pass (NFR-7).

---

### Task 9: Docs and the end-to-end check

**Files:**
- Modify: `docs/architecture.md` (Flow step 2), `tasks/progress.md`

- [ ] **Step 1: Update `docs/architecture.md` Flow step 2** with the content flow: one call for each language, eager streaming, the limit check with the longest first name, the SSE events and D-030 to D-034.

- [ ] **Step 2: Run the end-to-end check** (NFR-8)

Run: `docker compose -p cc-verify up -d --build --wait` with no `.env` and `LLM_MODE=mock`. Then `curl -N -X POST localhost:8080/api/content/generate` with the example brief and rule.
Expected: `audience`, several `draft`, `final` for `tr` with `sms_runes` ≤ 160, `done`. A headless Chrome run on `/campaigns/new` shows the `tr` card fill in and become final.

- [ ] **Step 3: Run `make verify`**

Expected: exit 0 (NFR-10). Then `docker compose -p cc-verify down -v`.

- [ ] **Step 4: Check the import rules** (NFR-3)

Run: `cd backend && go list -f '{{join .Imports "\n"}}' ./internal/content ./internal/llm | grep campaigncomposer`
Expected: only `campaigncomposer/internal/config`, from `internal/llm`.

---

## Out of scope (owned by later slices)

- Storing the content, the campaign record, the preview with a sample user, the edit step and the approval: slice 06.
- Rendering for each recipient at send time, and the second limit check: slice 07.
- Email formatting (heading, button): a later slice.
