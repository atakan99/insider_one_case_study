# 06 — Review and Approve Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task by task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** When Claude finishes the content, the backend saves a draft campaign. The marketer opens the draft in a review window, previews it with a sample user, edits it and approves it after a second confirm.

**Architecture:** This slice adds one package, one migration, five endpoints and three frontend views:

```
store/migrations  00002_campaigns.sql: campaigns, campaign_contents, immutability triggers
internal/segment  SampleFirstNames (first names only)
internal/campaign types, Store (pgx), CheckContents, Service (Save, Approve, Sample, Get, List)
internal/httpapi  campaigns.go (GET list, GET one, PUT contents, POST approve, GET sample)
                  content.go: optional campaign_id, draft save, `campaign` event
cmd/api           wires campaign.Service and the draft saver
frontend          lib/campaigns.ts, lib/limits.ts, review-dialog.vue, message-preview.vue,
                  Review button in content-cards.vue, pages campaigns/index.vue and [id].vue
```

`internal/campaign` imports only `content`, `segment` and `store` from `internal/`. `internal/httpapi` joins `campaign` with the existing packages.

**Tech Stack:** Go 1.27, `pgx/v5`, goose, Vue 3, shadcn-vue (`Dialog`, `Tabs`, `AlertDialog`, `Select`, `Textarea`, `Input`, `Badge`, `Table`), `vue-sonner`, Vitest.

**Spec:** `docs/specs/06-review-approve/goal.md` and `requirements.md`. The mockup is `docs/specs/06-review-approve/review-mockup.html`.

## Global Constraints

- Limits after rendering with `first_name_runes`: SMS 160 runes, subject 80 runes, body 2,000 runes. The backend uses `content.Validate`. The browser mirrors it in `lib/limits.ts`.
- Status values in this slice: `draft` and `approved`. Only a draft can change. After approval, the content and the brief, rule, description and skipped languages cannot change.
- Body limit 64 KiB for `PUT /api/campaigns/{id}/contents` and `POST /api/campaigns/{id}/approve`. The content endpoint keeps 16 KiB.
- Each query has the `2s` statement timeout of slice 04.
- Error codes: `invalid_request`, `not_found`, `campaign_not_draft`, `content_invalid`, `db_unavailable`. Stream error codes with no language: `draft_not_saved`, `campaign_conflict`.
- No user row in the campaign tables. Only `GET /api/campaigns/{id}/sample` sends first names. No endpoint sends a surname, an email address or a phone number.
- Logs: campaign id, languages, counts, repeat flag. Never the brief, the content text or the API key.
- UI: no hard-coded colour. New token `--warning` in `:root` and `.dark`, mapped as `--color-warning` in `@theme inline`. No horizontal scroll at 360 px. The review window fills the screen below 720 px.
- Preview sender text: `Your brand`. Fallback preview name: `Alex`.
- Tests use `llm.Mock`. No test reads `ANTHROPIC_API_KEY`. Postgres tests use `storetest.NewDB` and skip without `TEST_DATABASE_URL`.
- Do not commit after each task. The user commits the slice after review.

## Review Focus

1. **The marketer clicks "Regenerate content" after a slow first run, and the old stream saves a draft late.** The old stream must not save after the page aborted it. The backend saves no draft after a client disconnect (FR-6). Pinned by Task 4 `TestGenerate_NoDraftAfterDisconnect`.
2. **The jsonb rule comes back with a different key order.** Postgres `jsonb` reorders keys. A byte compare of the rule would turn every regenerate into `campaign_conflict`. `ReplaceDraft` compares with `rule = $n::jsonb` in SQL. Pinned by Task 2 `TestStore_ReplaceDraftSameRuleOtherKeyOrder`.
3. **A long Turkish name, for example `Abdurrahmangazi Şükrü`, joins the audience between the draft and the approval.** The approval must use the new `first_name_runes`, so an SMS that fit before can fail. Pinned by Task 3 `TestService_ApproveUsesNewLongestName` (FR-17).
4. **The marketer opens the same draft in two tabs and approves in both.** One approval wins. The second gets `200` with the approved campaign and changes nothing. Pinned by Task 3 `TestService_ApproveConcurrent` (FR-12).
5. **The marketer closes the review window with unsaved edits by pressing Escape or clicking outside.** The window must ask "Discard your changes?" for each way to close, not only for "Cancel". Pinned by Task 7 `asks before closing with Escape`.

---

### Task 1: Migration and sample names

**Files:**
- Create: `backend/internal/store/migrations/00002_campaigns.sql`, `backend/internal/segment/sample.go`, `backend/internal/segment/sample_test.go`
- Modify: `backend/internal/store/migrate_test.go`

**Interfaces:**
- Produces:

```go
package segment
// SampleFirstNames returns up to n distinct first names of the reachable users
// of each language in langs. The first name of each list is the longest
// (ties: by name). The others come from the users with the lowest id.
func SampleFirstNames(ctx context.Context, pool *pgxpool.Pool, w Where, langs []string, n int) (map[string][]string, error)
```

Tables, as in the spec "Data" section. Extra points:
- `campaigns.status` has `CHECK (status IN ('draft', 'approved'))`. Slice 07 widens it.
- `campaigns.skipped_languages text[] NOT NULL DEFAULT '{}'`.
- `campaign_contents.campaign_id` references `campaigns (id) ON DELETE CASCADE`. `language` has `CHECK (language IN ('tr', 'en'))`.
- Trigger `campaign_contents_draft_only`: `BEFORE INSERT OR UPDATE OR DELETE` on `campaign_contents`. It raises an exception when the parent campaign has a status other than `draft`.
- Trigger `campaigns_frozen_after_approval`: `BEFORE UPDATE` on `campaigns`. When `OLD.status <> 'draft'`, it raises an exception if `brief`, `rule`, `description`, `skipped_languages` or `approved_at` changes, or if `NEW.status = 'draft'`. Slice 07 can still move the status forward.
- The Down section drops both tables and the trigger functions.

- [ ] **Step 1: Write the failing tests**
  - `migrate_test.go` `TestMigrate_CampaignTriggers`: insert a draft and one content row. Update the content row (succeeds). Set the status to `approved` (succeeds). Then each of these fails: an update of `sms`, a delete of the content row, an insert of an `en` row, an update of `brief`, and `status = 'draft'`.
  - `sample_test.go` `TestSampleFirstNames` (Postgres, seed users with `seed.Load`): for the example rule and `n = 5`, compare with a value that the test calculates in Go from `seed.Generate`. Check: at most 5 names, distinct, the first is the longest by rune count (ties by name), the rest follow the lowest user id. Insert a user named `Abdurrahmangazi` with consent and check that it comes first.
  - `TestSampleFirstNames_Stable`: two calls give equal maps.
  - `TestSampleFirstNames_NoConsentExcluded`: a user with the longest name and no consent on any channel is not in the result.
  - `TestSampleQuery_Columns`: build the SQL with an unexported `sampleQuery(w Where, langs int) string` for a rule on `city`. Assert that it does not match `\blast_name\b`, `\bemail\b`, `\bphone\b`, `\bbirth_date\b` or `\btotal_spent\b` (FR-24).
- [ ] **Step 2: Run** `make test`. Expected: the new tests FAIL (no table, no function).
- [ ] **Step 3: Write the migration and `SampleFirstNames`.** Run the query in `inTx`. The extra parameters (`langs`, `n`) take the placeholder numbers after `len(w.Args())`. `w.sql == ""` returns `ErrEmptyWhere`.
- [ ] **Step 4: Run** `make test`. Expected: PASS, with the Postgres tests not skipped.

---

### Task 2: Campaign store

**Files:**
- Create: `backend/internal/campaign/campaign.go`, `store.go`, `store_test.go`

**Interfaces:**
- Consumes: `content.Content`.
- Produces:

```go
package campaign

type Status string
const (StatusDraft Status = "draft"; StatusApproved Status = "approved")

type LanguageContent struct {
	Language string `json:"language"`
	content.Content                       // sms, email_subject, email_body (flattened)
	SMSCount       int64 `json:"sms_count"`
	EmailCount     int64 `json:"email_count"`
	FirstNameRunes int   `json:"first_name_runes"`
}

type Campaign struct {
	ID               int64             `json:"id"`
	Status           Status            `json:"status"`
	Brief            string            `json:"brief"`
	Rule             json.RawMessage   `json:"-"`
	Description      string            `json:"description"`
	SkippedLanguages []string          `json:"skipped_languages"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	ApprovedAt       *time.Time        `json:"approved_at"`
	Contents         []LanguageContent `json:"contents"` // sorted by language
}

type Summary struct {
	ID          int64      `json:"id"`
	Status      Status     `json:"status"`
	Description string     `json:"description"`
	Languages   []string   `json:"languages"`
	CreatedAt   time.Time  `json:"created_at"`
	ApprovedAt  *time.Time `json:"approved_at"`
}

// Draft is the input of CreateDraft and ReplaceDraft.
type Draft struct {
	Brief, Description string
	Rule               json.RawMessage // canonical JSON from segment.Rule.MarshalJSON
	Skipped            []string
	Contents           []LanguageContent
}

var (
	ErrNotFound = errors.New("campaign: not found")
	ErrNotDraft = errors.New("campaign: not a draft")
	ErrConflict = errors.New("campaign: brief or rule differs, or not a draft")
)

type Store struct{ /* pool */ }
func NewStore(pool *pgxpool.Pool) *Store
func (s *Store) CreateDraft(ctx context.Context, d Draft) (int64, error)
// ReplaceDraft replaces the contents, counts and skipped languages of draft id.
// It returns ErrConflict when the campaign is not a draft, or when its brief or
// rule differs from d (rule compared as jsonb). ErrNotFound when id is unknown.
func (s *Store) ReplaceDraft(ctx context.Context, id int64, d Draft) error
func (s *Store) Get(ctx context.Context, id int64) (Campaign, error)
func (s *Store) List(ctx context.Context, limit int) ([]Summary, error) // newest first
// Update locks the campaign row (SELECT ... FOR UPDATE) and calls check with it.
// When the status is not draft: if approve is true, it returns the campaign and
// repeat = true without calling check; if approve is false, it returns ErrNotDraft.
// Otherwise it writes the contents that check returns, sets updated_at, and,
// when approve is true, sets status approved and approved_at, all in one
// transaction. An error from check rolls back and is returned as it is.
func (s *Store) Update(ctx context.Context, id int64, approve bool,
	check func(Campaign) ([]LanguageContent, error)) (c Campaign, repeat bool, err error)
```

- [ ] **Step 1: Write the failing tests** in `store_test.go` (Postgres, migrated empty DB):
  - `TestStore_CreateAndGet`: the fields round-trip, `ApprovedAt` is nil, `Contents` are sorted by language.
  - `TestStore_ReplaceDraft`: same brief and rule replace the contents and the skipped languages. One campaign row remains.
  - `TestStore_ReplaceDraftSameRuleOtherKeyOrder`: the stored rule is `{"op":"and","conditions":[]}`, and `d.Rule` is `{"conditions":[],"op":"and"}`. The result is no error.
  - `TestStore_ReplaceDraftConflict`: a different brief, a different rule, and an approved campaign each give `ErrConflict`. An unknown id gives `ErrNotFound`.
  - `TestStore_UpdateSave` and `TestStore_UpdateApprove`: the content changes. For approve, the status is `approved` and `approved_at` is set.
  - `TestStore_UpdateRepeat`: approve twice. The second call has `repeat = true`, the check function is not called, and `approved_at` does not change. A save after approval gives `ErrNotDraft`.
  - `TestStore_UpdateCheckError`: the check returns an error. Nothing changes, and the error comes back with `errors.Is`.
  - `TestStore_List`: three campaigns, newest first. `limit` cuts the list. `Languages` comes from the content rows.
- [ ] **Step 2: Run** `make test`. Expected: FAIL (package missing).
- [ ] **Step 3: Implement** `campaign.go` and `store.go`. Each method runs with `SET LOCAL statement_timeout = '2s'`, as `segment.inTx` does. `Get` for an unknown id returns `ErrNotFound`.
- [ ] **Step 4: Run** `make test`. Expected: PASS.

---

### Task 3: Content check and campaign service

**Files:**
- Create: `backend/internal/campaign/check.go`, `check_test.go`, `service.go`, `service_test.go`

**Interfaces:**
- Consumes: Task 2 `Store`. `segment.Allowlist.ParseRule`, `Compile`, `segment.LanguageCounts`, `segment.SampleFirstNames`. `content.Validate`.
- Produces:

```go
package campaign

// InvalidError carries the validation errors of a save or an approval.
type InvalidError struct{ Errors []content.FieldError }
func (e *InvalidError) Error() string

// CheckContents decodes body ({"contents": [...]}) and checks it against the
// languages of the draft and the current audience. It returns the trimmed
// contents with the new counts and first_name_runes, sorted by language.
func CheckContents(body json.RawMessage, draftLangs []string, aud map[string]segment.LanguageCounts) ([]LanguageContent, []content.FieldError)

type AudienceCounter func(ctx context.Context, w segment.Where) (segment.Audience, error)
type Sampler func(ctx context.Context, w segment.Where, langs []string, n int) (map[string][]string, error)

type SampleNames struct {
	Code       string   `json:"code"`
	FirstNames []string `json:"first_names"`
}

type Service struct {
	Store  *Store
	Allow  segment.Allowlist
	Count  AudienceCounter
	Sample Sampler
	Now    func() time.Time
}
func (s *Service) Get(ctx context.Context, id int64) (Campaign, error)
func (s *Service) List(ctx context.Context) ([]Summary, error)  // limit 100
func (s *Service) Save(ctx context.Context, id int64, body json.RawMessage) (Campaign, error)
func (s *Service) Approve(ctx context.Context, id int64, body json.RawMessage) (Campaign, bool, error)
func (s *Service) SampleNames(ctx context.Context, id int64) ([]SampleNames, error) // 5 names, languages of the campaign
```

`CheckContents` rules, from the spec "Validation" table. The paths:
- `contents`: not an array, or empty: `contents: must be an array with one entry for each language`.
- `contents[i]`: not an object, or an unknown key: path `contents[i]` or `contents[i].<key>`.
- `contents[i].language`: not in the draft, twice, or no reachable users in `aud`.
- `contents`: a draft language with no entry: path `contents`, message names the language.
- Text fields: call `content.Validate` with the three text fields as a JSON object and the `MaxFirstName` of that language. Prefix each path with `contents[i].`.

`Save` and `Approve` both call `Store.Update` with a check that does these steps:
1. Parse the stored rule with `Allow.ParseRule`. A failure is an `InvalidError` with path `rule`.
2. Run `Count` on the compiled rule with the `2s` timeout. An error returns as it is (the handler maps it to `503`).
3. Call `CheckContents`. Errors become an `InvalidError`.

- [ ] **Step 1: Write the failing tests.**
  - `check_test.go` `TestCheckContents`: a table test with one case for each bullet of requirement FR-15, and the path of each error. A case with two bad entries has errors for both (FR-16). A valid body returns trimmed fields and the counts from `aud`.
  - `TestCheckContents_NoReachable`: `aud` has `tr` with `Reachable: 0`. The error path is `contents[0].language` (FR-18).
  - `service_test.go` (Postgres, seed users, real `Store`, `segment.CountAudience` and `segment.SampleFirstNames`):
    - `TestService_SaveAndApprove`: FR-9 and FR-10.
    - `TestService_ApproveUsesNewLongestName`: make a draft whose SMS fits for the current longest name with 1 rune to spare. Insert a reachable user whose first name is 2 runes longer. `Approve` returns an `InvalidError` with path `contents[0].sms`. The campaign is still a draft (FR-17).
    - `TestService_ApproveConcurrent`: 10 goroutines approve one draft. All return the same id and no error. Exactly one has `repeat = false`. `approved_at` has one value (FR-12).
    - `TestService_SaveApproved`: `ErrNotDraft` (FR-13).
    - `TestService_DBTimeout`: a `Count` that returns `context.DeadlineExceeded`. The error comes back, and the campaign does not change (FR-19).
- [ ] **Step 2: Run** `make test`. Expected: FAIL.
- [ ] **Step 3: Implement** `check.go` and `service.go`. `Service.SampleNames` uses the languages of `Campaign.Contents`.
- [ ] **Step 4: Run** `make test`. Expected: PASS.

---

### Task 4: Draft save in the content endpoint

**Files:**
- Modify: `backend/internal/httpapi/content.go`, `backend/internal/httpapi/content_test.go`

**Interfaces:**
- Consumes: `campaign.Draft`, `campaign.LanguageContent`, `campaign.ErrConflict`, `campaign.ErrNotFound`.
- Produces:

```go
package httpapi

// DraftSaver stores the draft at the end of a content stream.
type DraftSaver interface {
	CreateDraft(ctx context.Context, d campaign.Draft) (int64, error)
	ReplaceDraft(ctx context.Context, id int64, d campaign.Draft) error
}

// ContentGenerator gets one new field:
//	Drafts DraftSaver

type campaignEvent struct {
	ID     int64           `json:"id"`
	Status campaign.Status `json:"status"`
}
// streamErrorEvent is an error event with no language.
type streamErrorEvent struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
```

Changes:
- The request body gets `"campaign_id"` (optional, `*int64`). A value of 0 or less returns `400 invalid_request` before the stream.
- `runLanguage` returns the final content and limits, or nothing when the language failed.
- After `wg.Wait()`, and only when `ctx.Err() == nil`:
  1. If no language is final, send `done`.
  2. Build `campaign.Draft`: the sanitized brief, `rule.MarshalJSON()`, the description, the failed codes as `Skipped`, and one `LanguageContent` for each final language. It takes `SMS`, `Email` and `MaxFirstName` from the `segment.LanguageCounts` of that language.
  3. With no `campaign_id`, call `CreateDraft`. With one, call `ReplaceDraft` and use that id.
  4. On success, send `campaign` with status `draft`. On `ErrConflict` or `ErrNotFound`, send `error` with `code: campaign_conflict` and the message `This draft cannot be updated. Find the audience again to start a new campaign.` On another error, send `error` with `code: draft_not_saved` and `The draft could not be saved. Try again.`
  5. Send `done`.
- The save uses a context with a `2s` timeout.

- [ ] **Step 1: Write the failing tests** with a fake `DraftSaver` that records calls:
  - `TestGenerate_SavesDraft`: FR-1 order of events (`final`, `campaign`, `done`) and the recorded `Draft` fields (FR-2).
  - `TestGenerate_DraftSkipsFailedLanguage`: two languages, one fails. `Skipped` is `["en"]` and `Contents` has `tr` only (FR-3).
  - `TestGenerate_NoDraftWhenAllFail`: no save call and no `campaign` event (FR-4).
  - `TestGenerate_DraftSaveError`: the fake returns an error. The stream has `error` with no `language` and `code: draft_not_saved`, then `done` (FR-5).
  - `TestGenerate_NoDraftAfterDisconnect`: cancel the request context while the mock streams. No save call (FR-6).
  - `TestGenerate_ReplaceDraft`: `campaign_id: 7` calls `ReplaceDraft(7, ...)`, and the `campaign` event has `id: 7` (FR-7).
  - `TestGenerate_ReplaceConflict`: the fake returns `campaign.ErrConflict`. The stream has `error` with `code: campaign_conflict` (FR-8).
  - `TestGenerate_BadCampaignID`: `campaign_id: 0` returns `400 invalid_request` and no LLM call.
  - Update the existing tests that assert the exact event sequence, so that they give a fake saver and expect `campaign` before `done`.
- [ ] **Step 2: Run** `make test`. Expected: the new tests FAIL.
- [ ] **Step 3: Implement** the changes above.
- [ ] **Step 4: Run** `make test`. Expected: PASS.
- [ ] **Step 5: Postgres check** `TestGenerate_DraftInDatabase` (Postgres, seed users, real `campaign.Store` as the saver, default mock): after the stream, `Store.Get` returns a draft with a `tr` row. The counts match `CountAudience` (FR-1, FR-2).

---

### Task 5: Campaign endpoints and wiring

**Files:**
- Create: `backend/internal/httpapi/campaigns.go`, `campaigns_test.go`
- Modify: `backend/internal/httpapi/body.go`, `backend/internal/httpapi/cors.go`, `backend/internal/httpapi/cors_test.go`, `backend/cmd/api/main.go`

**Interfaces:**
- Consumes: Task 3 `campaign.Service` methods and errors. Task 4 `DraftSaver`.
- Produces:

```go
package httpapi

// CampaignService is the part of campaign.Service that the handlers use.
type CampaignService interface {
	Get(ctx context.Context, id int64) (campaign.Campaign, error)
	List(ctx context.Context) ([]campaign.Summary, error)
	Save(ctx context.Context, id int64, body json.RawMessage) (campaign.Campaign, error)
	Approve(ctx context.Context, id int64, body json.RawMessage) (campaign.Campaign, bool, error)
	SampleNames(ctx context.Context, id int64) ([]campaign.SampleNames, error)
}

// RegisterCampaigns adds the five routes of the spec "API" section.
func RegisterCampaigns(mux *http.ServeMux, svc CampaignService)

// decodeBodyLimit is decodeBody with a given byte limit. decodeBody calls it
// with maxBodyBytes.
func decodeBodyLimit(w http.ResponseWriter, r *http.Request, v any, limit int64) error
```

Response shapes: `{"campaigns": [...]}` for the list, the campaign object for one, save and approve, `{"languages": [...]}` for the sample. Error mapping:
- id not a positive integer, or bad body: `400 invalid_request`.
- `ErrNotFound`: `404 not_found`. `ErrNotDraft`: `409 campaign_not_draft`.
- `*InvalidError`: `422 content_invalid` with `details`.
- Any other error: `503 db_unavailable`.
- The body for save and approve must be a JSON object with a `contents` key, at most `64 * 1024` bytes.
- Log one line for each save and approval with the id, the languages, the counts and `repeat`.

`cmd/api`: create `campaign.NewStore(pool)` and a `campaign.Service` with `segment.Default()`, the existing `count`, a sampler that calls `segment.SampleFirstNames` with `pool`, and `time.Now`. Pass the store as `Drafts` to `newContentGenerator`. Call `RegisterCampaigns`. CORS allows `GET, POST, PUT, OPTIONS`.

- [ ] **Step 1: Write the failing tests** in `campaigns_test.go` with a fake `CampaignService`:
  - one test for each row of the spec error table and for each FR-14 body case (`abc`, `0`, `-1`, over 64 KiB, not JSON, no `contents`)
  - `TestApprove_Repeat`: the fake returns `repeat = true`, and the status is `200`
  - `TestGetCampaign_Shape`: the JSON has the keys of the spec example and has no `rule` key
  - `TestList_NoContentText`: the list JSON has no `sms` key (FR-21)
  - `cors_test.go`: the preflight allows `PUT`.
- [ ] **Step 2: Run** `make test`. Expected: FAIL.
- [ ] **Step 3: Implement** `campaigns.go`, `decodeBodyLimit`, the CORS change and the wiring.
- [ ] **Step 4: Run** `make verify`. Expected: exit 0.

---

### Task 6: Frontend API and limit check

**Files:**
- Create: `frontend/src/lib/campaigns.ts`, `frontend/src/lib/limits.ts`, `frontend/src/lib/__tests__/campaigns.test.ts`, `frontend/src/lib/__tests__/limits.test.ts`
- Modify: `frontend/src/lib/content.ts`, `frontend/src/lib/__tests__/content.test.ts`

**Interfaces:**
- Produces:

```ts
// lib/campaigns.ts
export type CampaignStatus = 'draft' | 'approved'
export interface LanguageContent extends Content { language: string, sms_count: number, email_count: number, first_name_runes: number }
export interface Campaign { id: number, status: CampaignStatus, brief: string, description: string, skipped_languages: string[], created_at: string, updated_at: string, approved_at: string | null, contents: LanguageContent[] }
export interface CampaignSummary { id: number, status: CampaignStatus, description: string, languages: string[], created_at: string, approved_at: string | null }
export interface FieldError { path: string, message: string }
export function listCampaigns(): Promise<CampaignSummary[]>
export function getCampaign(id: number): Promise<Campaign>
export function getSampleNames(id: number): Promise<Record<string, string[]>>
export function saveContents(id: number, contents: (Content & { language: string })[]): Promise<Campaign>
export function approveCampaign(id: number, contents: (Content & { language: string })[]): Promise<Campaign>
// The rejected error of save and approve has `code`, `message` and `details: FieldError[]`.
export class CampaignRequestError extends Error { code: string, details: FieldError[], status: number }
export const LANGUAGE_NAMES: Record<string, string> // tr: 'Turkish', en: 'English'

// lib/limits.ts
export const LIMITS: { sms: 160, email_subject: 80, email_body: 2000 }
export function runeLength(s: string): number                       // code points
export function renderedLength(text: string, firstNameRunes: number): number
export function checkContent(c: Content, firstNameRunes: number): Partial<Record<keyof Content, string>>
```

`checkContent` messages (UI text, exact):
- empty after trim: `Write some text.`
- over the limit: `Too long after the name is filled in.`
- a `{` or `}` outside `{{first_name}}`: `Only {{first_name}} is allowed.`
- `\n` or `\r` in the subject: `The subject must be one line.`
- an HTML tag (`<` followed by a letter, `/` or `!`): `HTML is not allowed.`
- a control character (except `\n` and `\t` in the body): `Remove the special characters.`

`content.ts` changes: `generateContent` gets a fifth optional parameter `campaignId?: number` and sends it as `campaign_id`. `ContentEvent` adds `{ event: 'campaign', data: { id: number, status: CampaignStatus } }`. The `error` event has `language?: string`. `KNOWN_EVENTS` adds `campaign`.

- [ ] **Step 1: Write the failing tests.**
  - `limits.test.ts`: the FR-9 cases of slice 05 (`first_name_runes` 20, SMS 154 runes + one variable = 160 passes, 155 fails, subject 80 passes and 81 fails, `ş` counts 1). The FR-30 case (`first_name_runes` 14, 146 runes + one variable gives 160). One case for each message above.
  - `campaigns.test.ts`: each function calls the right method and path, with `fetch` mocked. A `422` rejects with `CampaignRequestError` with `details`.
  - `content.test.ts`: the `campaign` event parses. `campaign_id` is in the body only when given.
- [ ] **Step 2: Run** `make test`. Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run** `make test`. Expected: PASS.

---

### Task 7: Review window

**Files:**
- Create: `frontend/src/components/campaign/review-dialog.vue`, `message-preview.vue`, `__tests__/review-dialog.test.ts`
- Modify: `frontend/src/assets/index.css`

**Interfaces:**
- Consumes: Task 6.
- Produces:

```ts
// review-dialog.vue
props: { campaignId: number }
model: open (boolean, v-model:open)
emits: { approved: [id: number] }
// message-preview.vue
props: { channel: 'sms' | 'email', content: Content, name: string }
```

Behaviour follows the spec sections "Review window", "Save and close" and "Confirm popup". Decisions for the build:
- Load `getCampaign` and `getSampleNames` when `open` becomes true. Keep a local copy of the contents as the edit state. `dirty` is true when the copy differs from the loaded campaign.
- A draft shows the editors and the footer buttons. An approved campaign shows the preview only, with no "Save draft" and no "Approve & send" (FR-42).
- Every close path goes through one `requestClose()`: the close button, "Cancel", Escape and a click outside. Use `DialogContent` events `escape-key-down` and `pointer-down-outside` with `preventDefault()` when `dirty`. With `dirty`, open an `AlertDialog` "Discard your changes?" with "Keep editing" and "Discard".
- Map server error paths to the UI: `contents[i].<field>` goes under that field on the tab of `contents[i].language` from the sent array. All other paths go to the banner.
- The confirm popup sums `sms_count` and `email_count` of all contents.
- `--warning`: light `oklch(0.666 0.179 58.318)`, dark `oklch(0.769 0.188 70.08)`. Map `--color-warning: var(--warning)` in the first `@theme inline` block.
- Layout classes follow the mockup. Below the `md` breakpoint, the dialog uses `h-dvh w-screen max-w-none rounded-none`. The body uses `grid md:grid-cols-2`. Tab lists and footer buttons are full width on a phone.

- [ ] **Step 1: Write the failing tests** in `review-dialog.test.ts`, with `@/lib/campaigns` mocked:
  - `shows tabs for each channel and language` (FR-28, FR-29)
  - `counter at 160 and 161` with red class and both buttons off (FR-30)
  - `counter warns above 90%` with the `text-warning` class (FR-31)
  - `shows field errors and a tab mark` for `{{last_name}}`, an empty field and a subject line break (FR-32)
  - `preview uses the longest sample first` and `renders html as text` (FR-33)
  - `falls back to Alex when the sample fails` (FR-34)
  - `names skipped languages` in the footer and the confirm popup, and `totals in confirm` (FR-35)
  - `save draft sends one PUT and toasts` (FR-36)
  - `cancel asks only when dirty` and `asks before closing with Escape` (FR-37, Review Focus 5)
  - `approve sends once on double click` and `emits approved` (FR-38, FR-39, FR-40)
  - `maps 422 details to fields and banner` (FR-41)
  - `approved campaign is read-only` (FR-42)
- [ ] **Step 2: Run** `make test`. Expected: FAIL.
- [ ] **Step 3: Implement** the components and the token.
- [ ] **Step 4: Run** `make test` and `make lint`. Expected: PASS.

---

### Task 8: Pages and the Review button

**Files:**
- Modify: `frontend/src/components/campaign/content-cards.vue`, `frontend/src/components/campaign/__tests__/content-cards.test.ts`, `frontend/src/pages/campaigns/index.vue`, `frontend/src/pages/campaigns/[id].vue`
- Create: `frontend/src/pages/campaigns/__tests__/index.test.ts`, `frontend/src/pages/campaigns/__tests__/id.test.ts`

**Interfaces:**
- Consumes: Task 6 and Task 7.

`content-cards.vue`:
- Keep `campaignId` from the `campaign` event. "Review & approve" shows after the cards and is on only when `campaignId` is set and no request runs (FR-25).
- "Regenerate content" passes `campaignId` to `generateContent` (FR-26). A new run from a new "Find audience" (a new `runId`) clears `campaignId` first.
- An `error` event with no `language` shows in the content area `Alert`, and the cards stay (FR-27).
- On `approved`, go to `/campaigns/{id}` with the router.

`campaigns/index.vue`: a `Table` on `md` and up, stacked `Card` rows below `md`. Columns: `#id`, description, languages (names), status `Badge` (`Draft` outline, `Approved` default), created time with `toLocaleString()`. A click on a draft opens `review-dialog`. A click on an approved row goes to `/campaigns/{id}`. Reload the list after an approval or a close with a save. Empty state: `No campaigns yet` and a link `Create a campaign` to `/campaigns/new` (FR-43).

`campaigns/[id].vue`: header `Campaign #{id}`, status badge, description. One read-only `Card` per language with the three texts and `{sms_count} SMS · {email_count} email`. The line `Sending starts in a later step.` A draft shows "Review & approve", which opens `review-dialog`. A `404` shows `Campaign not found` and a link to `/campaigns` (FR-44).

- [ ] **Step 1: Write the failing tests** for each FR above, with the API mocked.
- [ ] **Step 2: Run** `make test`. Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run** `make test` and `make lint`. Expected: PASS.

---

### Task 9: Docs and the end-to-end check

**Files:**
- Modify: `docs/architecture.md`, `docs/decisions-log.md`, `tasks/backlog.json`, `tasks/progress.md`

- [ ] **Step 1: `docs/architecture.md`.** Replace "TBD per slice" in Flow step 3 with the draft save, the review window, the save and approve endpoints, the row lock, the triggers and the sample endpoint. In Flow step 1, change "Recipients is the only function that returns user rows" so that it also names `SampleFirstNames`, which returns first names only. Add `campaign/` to the folder tree.
- [ ] **Step 2: `docs/decisions-log.md`.** Add D-036 to D-045, one for each item of the spec section "Calls by Claude", and one for the triggers (Task 1). Use the Context, Options, Choice, Why format.
- [ ] **Step 3: `tasks/backlog.json`.** Set `06-review-approve.spec` to `docs/specs/06-review-approve/` and list FR-1 to FR-44 and NFR-1 to NFR-12 as short acceptance lines. Keep `passes: false`.
- [ ] **Step 4: End-to-end check.** Run `docker compose -p cc-verify up -d --build --wait` with no `.env` and `LLM_MODE=mock`. Then:
  - `POST /api/content/generate` with the example brief ends with `campaign` and `done`.
  - `GET /api/campaigns` lists the draft. `POST /api/campaigns/{id}/approve` with the draft content returns `200` and `approved`. A second call returns `200`. `PUT .../contents` then returns `409`.
  - Headless Chrome at 1280 px and at 360 px: `/campaigns/new` with the example brief, "Review & approve", the review window, "Approve & send", "Send now", then `/campaigns/{id}` with "Approved". At 360 px, `document.documentElement.scrollWidth` equals 360 on `/campaigns` and with the window open (NFR-7).
  - Run `docker compose -p cc-verify down -v`.
- [ ] **Step 5: Run** `make verify`. Expected: exit 0.
- [ ] **Step 6: STE check.** Run `python3 .claude/skills/ste-writing/scripts/ste_check.py` on the changed `.md` files and fix the findings.
- [ ] **Step 7: `tasks/progress.md`.** Append the session entry.

## Out of scope (owned by later slices)

- Sending, the status values after `approved`, the real sender and the recipient list: slice 07.
- The live dashboard on `/campaigns/{id}`: slice 08.
- A retry for one failed language, a draft delete, autosave, and an edit of the brief or the rule.
