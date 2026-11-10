# Decisions log (Claude)

Claude writes this log. It records the implementation choices that Claude made or recorded during the work. The user owns the root `DECISIONS.md`, which is the case deliverable. Code and docs refer to the IDs in this log, for example `D-008`.

## D-001 Go for the backend

- **Context:** The case allows Go or PHP for the backend.
- **Options:** Go, PHP.
- **Choice:** Go.
- **Why:** Workers and rate limiters need cheap concurrency. Go gives one static binary per service.

## D-002 Postgres as the datastore

- **Context:** We must store users, campaigns and send state. Segment rules filter users.
- **Options:** Postgres, a document store.
- **Choice:** Postgres is the system of record. The queue is in Redis (D-007).
- **Why:** Segment queries need SQL. Postgres also holds the delivery status and the idempotency keys.

## D-003 Backend from scratch, one module, one binary per process

- **Context:** We need `api`, `worker`, `mockprovider` and `seed` processes.
- **Options:** A Go starter template, our own layout.
- **Choice:** One Go module. Each process is a binary in `cmd/`.
- **Why:** A template brings code that we did not choose and must defend. The mock provider imports nothing from `internal/` except `config`. Thus it stays a stand-in for an external service.

## D-004 `make test` without Go

- **Context:** Graders can have Docker but no Go.
- **Options:** Require Go, run Go in a container.
- **Choice:** `make test` uses a `golang` container when `go` is not on the `PATH`.
- **Why:** The case requires that `make test` works. Docker is the only tool that we can expect.

## D-005 Router: stdlib `net/http`

- **Context:** The API needs method routing and path parameters.
- **Options:** stdlib `net/http` (Go 1.22+ patterns), chi.
- **Choice:** stdlib `net/http`.
- **Why:** It covers our routes with zero dependencies. Look at this again if we need middleware other than request-id, recover and timeout.

## D-006 Data access: `pgx` + `sqlc`, hand-built segment SQL

- **Context:** Most queries are fixed. Segment queries are dynamic, because the WHERE clause depends on the rule.
- **Options:** An ORM, `sqlc` for all queries, `sqlc` plus a hand-built SQL compiler.
- **Choice:** `pgx` + `sqlc` for fixed queries. `internal/segment` compiles the validated rule into SQL with `$n` placeholders.
- **Why:** `sqlc` cannot generate dynamic queries. Placeholders keep user values out of the SQL text.

## D-007 Redis Streams as the queue

- **Context:** Sends are asynchronous, at 100 messages per second per channel at most, with retries.
- **Options:** A Postgres table with `FOR UPDATE SKIP LOCKED`, Redis Streams, NATS JetStream.
- **Choice:** Redis Streams (user decision). Postgres keeps the idempotency keys and the delivery status.
- **Why:** Consumer groups give at-least-once delivery and pending-entry retries. The send-queue spec designs streams, retry and the Postgres idempotency check.

## D-008 One root Dockerfile for all app images

- **Context:** The stack has several Go binaries and, from slice 01, a frontend.
- **Options:** One Dockerfile per app, one multi-stage Dockerfile at the repo root.
- **Choice:** One root `Dockerfile` (user decision). Target `go-service` with `ARG CMD` builds each Go binary. Slice 01 adds a frontend target.
- **Why:** One file to read and one build context. `docker-compose.yml` selects the target and runs Postgres and Redis from official images.

## D-009 Frontend from the shadcn-vue-admin template, trimmed

- **Context:** The case asks for a Vue frontend. We need a sidebar layout, a light and dark theme, tables and dialogs.
- **Options:** Start from nothing, use the `shadcn-vue-admin` template (MIT) as it is, use the template and remove what the flow does not use.
- **Choice:** Trim the template into `frontend/`. Remove demo pages, mock login, UI languages, marketing site, theme editor and the template's `AGENTS.md` and `.agents/`. Keep file-based routing and auto-imports.
- **Why:** The shell and components work and have tests. Each kept file has a planned use. Removing file-based routing would change every page file for no gain.

## D-010 Compose serves the built frontend with nginx

- **Context:** `docker compose up` must start the full app. Developers also need hot reload.
- **Options:** nginx with the built `dist/`, `pnpm dev` in a container, `pnpm dev` only on the host.
- **Choice:** Compose builds the `frontend` target and serves it with nginx on port 3000. Local work uses `pnpm dev` on port 5173.
- **Why:** A static build is small and starts fast. nginx sends each unknown path to `index.html`, so deep links work after a reload.

## D-011 Export sessions with a script from the JSONL transcripts

- **Context:** The case asks for an export of each Claude session in `ai-sessions/`. API keys and personal data must not be in the exports.
- **Options:** Run `/export` by hand in each session, copy the raw JSONL files, convert the JSONL files to Markdown with a script.
- **Choice:** `ai-sessions/export_sessions.py` converts each session to one Markdown file and adds the subagent runs. It writes an index to `ai-sessions/README.md`.
- **Why:** The script covers all sessions and the subagent runs in one step. It removes hidden system context and redacts the email, the home path, account IDs and secrets. You can run it again at the end of the case.

## D-012 Seed data from a Go generator with `gofakeit`

- **Context:** The case needs a few thousand fake users. `docker compose up` must load them with no manual step.
- **Options:** Mockaroo files, a committed CSV file, Postgres `generate_series`, a Go generator with `gofakeit`.
- **Choice:** `cmd/seed` generates the users with `gofakeit` and a fixed random seed (user decision). Turkish names and all cities come from fixed lists in the seed package.
- **Why:** The Mockaroo API needs an account key and gives 1,000 rows per file on the free plan. A reviewer cannot generate the data again. A generator runs in compose with no account, and it adds edge-case users that the hard rules need. `gofakeit` has no Turkish name data.

## D-013 Users have a language and a country, campaigns have one language

- **Context:** Users are in Turkey and the US. Later, some users can speak a language that is not the main language of their country.
- **Options:** For storage: a BCP 47 locale, or `language` and `country` columns. For campaigns: one language, or one content variant for each language in the audience.
- **Choice:** Two columns, `language` (`tr`, `en`) and `country` (`TR`, `US`). In slice 02 the language follows the country. A campaign has one language, and the compiler adds `language = <campaign language>` to the rule (user decision).
- **Why:** Two columns are easier to filter and to test than locale strings. The separate column lets a later slice add mixed languages and multi-language campaigns with a new seed, with no schema change.

## D-014 `events` table plus summary columns on `users`

- **Context:** Segments such as "added to cart in the last 30 days and did not buy" need event times.
- **Options:** Summary columns only, an `events` table only, both.
- **Choice:** Both. The seed calculates `last_cart_add_at`, `last_purchase_at`, `order_count` and `total_spent` from `events`. The segment allowlist uses only `users` columns.
- **Why:** The compiler stays a simple `WHERE` clause with no joins or subqueries. The `events` table gives a real shopping history. All seed times are offsets from the seed time, so time-based segments are never empty.

## D-015 Migrations with `goose`, run by the seed process

- **Context:** Slice 02 adds the first tables. Something must create them before the api starts.
- **Options:** `goose`, `golang-migrate`, plain SQL files in the Postgres init folder. The runner: the api at startup, or the one-shot seed process.
- **Choice:** `goose` v3 with the SQL files embedded in `internal/store` (user decision). The `seed` process applies them, and the api starts after `seed` completes.
- **Why:** Embedded files need no extra image or tool. The init folder runs only on an empty volume, so later migrations would not run. One process that applies the migrations prevents a race between api replicas.

## D-016 Seed with `pgx` directly, `sqlc` comes later

- **Context:** D-006 chose `sqlc` for fixed queries. The seed needs only an existence check, a truncate, two bulk copies and two sequence resets.
- **Options:** Set up `sqlc` now, use `pgx` directly in the seed.
- **Choice:** The seed uses `pgx` `CopyFrom` and `Exec` directly. The first slice with fixed app queries sets up `sqlc`.
- **Why:** `sqlc` adds a config file, generated code and a generate step. These simple statements do not justify that now.

## D-017 Temporary Postgres for `make test`

- **Context:** The database tests need Postgres. `make test` must work with Docker and no Go (D-004).
- **Options:** A shared Postgres from compose, a temporary container for each `make test`.
- **Choice:** `make test` starts a temporary `postgres:17-alpine` container named `cc-test-pg` on the network `cc-test`. It publishes the port on `127.0.0.1:55432`. Go in a container joins `cc-test` through `GO_TEST`. Local Go uses the port. `DOCKER_GO` stays without the network, so lint works. Each test creates its own database with `storetest.NewDB`.
- **Why:** Tests never touch the dev data. Parallel packages do not collide. Port 55432 avoids a Postgres on the host port 5432.

## D-018 The mock provider does no dedup

- **Context:** The case says that the same message must not go to the same person twice. Twilio Messaging and SendGrid Mail Send have no `Idempotency-Key` on their send endpoints. A repeat request sends the message again.
- **Options:** The provider honors an `Idempotency-Key` and absorbs repeats. The provider delivers each repeat and the worker prevents duplicates.
- **Choice:** The provider delivers each repeat (user decision). It stores an optional `reference`, as SendGrid `custom_args` does. `GET /v1/stats` counts `duplicate_deliveries` for the audit.
- **Why:** The mock acts like the real providers. Thus the dedup logic is in our worker, where the case wants to see it. Provider dedup would hide a worker bug.

## D-019 Only a timeout or a dropped connection is ambiguous, and the worker does not retry it

- **Context:** With no provider dedup, a retry after an ambiguous failure can send a second copy.
- **Options:** Retry all failures. Check the provider before a retry. Do not retry ambiguous failures.
- **Choice:** In the mock, `500` and `503` mean "not delivered". `slow` and `lost` deliver the message, but the client gets a timeout or a dropped connection. The worker in slice 07 retries `429`, `5xx` and connection refused. It marks a timeout or a dropped connection `unknown` and does not retry (user decision).
- **Why:** The case forbids duplicates, so a possible missed message is better than a possible duplicate. SendGrid has no reliable lookup that the worker can use before a retry. Clear rules in the mock make each path testable.

## D-020 Token bucket from `golang.org/x/time/rate`

- **Context:** The mock provider needs a rate limit for each channel. The worker in slice 07 needs one too.
- **Options:** Our own token bucket, `golang.org/x/time/rate`, a Redis-based limiter.
- **Choice:** `golang.org/x/time/rate`, one limiter for each channel. The default is 100 per second with a burst of 10.
- **Why:** The package is maintained by the Go team and takes an explicit time, so tests can use a fake clock. The mock is one process, so it needs no shared state in Redis.

## D-021 `X-Mock-Fail` header for exact tests

- **Context:** The random failures make the worker tests flaky. The tests must reach each path exactly.
- **Options:** A fixed random seed only, a control endpoint that sets the next failure, a header on each request.
- **Choice:** The `X-Mock-Fail` header sets the mode for one request. The random rates and `MOCK_RANDOM` stay for the demo.
- **Why:** A header affects one request only. Thus parallel tests do not interfere. A control endpoint keeps state that a parallel test can change.

## D-022 The audience gives the languages, and the backend adds no language filter

- **Context:** The operator writes only the brief. D-013 said that a campaign has one language, and that the compiler adds `language = <campaign language>` to each rule.
- **Options:** The operator selects the language. The backend selects the biggest language group and excludes the other users. The backend reports each language of the audience.
- **Choice:** The segment response has a `languages` array. The backend fills it with a `GROUP BY language` on the audience. The compiler adds no language condition. If the brief names a language, Claude puts it in the rule (user decision, 2026-11-10). Slice 05 decides how it uses the array.
- **Why:** No user is excluded because of language. When a later seed mixes languages in a country, the array gets more entries with no code change. This replaces the campaign-language part of D-013. The `language` and `country` columns of D-013 do not change.

## D-023 The backend writes the rule description

- **Context:** The operator must see what the audience is before content generation.
- **Options:** Claude writes a summary. The backend makes the text from the validated rule.
- **Choice:** The backend makes the text from the validated rule, with a fixed phrase for each field and operator.
- **Why:** A summary from Claude can differ from the rule. The backend text always describes the query that runs.

## D-024 One repair round with the validation errors

- **Context:** Claude can return a rule that fails validation.
- **Options:** Fail at once. Retry with the same prompt. Send the errors back and ask for a correction, a fixed number of times.
- **Choice:** Send the validator errors back one time, as a `tool_result` with `is_error: true` for the earlier `tool_use`. If the second answer also fails, return `422`.
- **Why:** Most errors are small, for example a wrong value type, and the error text lets Claude fix them. One round limits the cost and the wait to two calls.

## D-025 Strict tool schema made from the allowlist

- **Context:** The Anthropic client asks Claude for the tool `submit_segment`. Opus 5.5 and Sonnet 5.5 reject a forced `tool_choice`, so the client sends `tool_choice: auto`. The tool input schema can be loose or strict.
- **Options:** A loose `{"type": "object"}` schema. A strict schema with enums for the fields, operators and cities.
- **Choice:** The tool has `strict: true`. `internal/segment` builds its JSON schema from the allowlist and unrolls the 3 levels of the tree, with no `$ref`. The caller gives the tool definition to `internal/llm`. Thus `internal/llm` does not import `internal/segment`. The backend validator still checks every answer.
- **Why:** A strict schema makes fewer bad answers, so the repair round runs less often. The schema is only a help. The validator is the only gate.

## D-026 Relative times use the wall clock

- **Context:** The seed times are offsets from the seed time. The seed skips when users exist, so the data on an old volume gets older each day.
- **Options:** The compiler uses the seed time from a stored value. The compiler uses the wall clock.
- **Choice:** The wall clock, through an injected clock. After about 30 days, the example brief can match no users on an old volume. `SEED_RESET=true` fixes this. The README slice documents it.
- **Why:** A real system uses the real time. A stored seed time adds a table and couples the api to the seed.

## D-027 Claude Opus 5.5 at low effort, with refusal fallbacks

- **Context:** The segment step is a small extraction. It must be reliable and fast.
- **Options:** `claude-opus-5-5`, `claude-sonnet-5-5`, `claude-haiku-5-5`.
- **Choice:** The default is `claude-opus-5-5`, with effort `low`. `ANTHROPIC_MODEL` can change it. The client asks for server-side refusal fallbacks (`fallbacks: "default"`, beta `server-side-fallback-2026-07-01`). If the Go SDK has no binding for this form, it uses the array form `[{model: claude-opus-4-8}]` with beta `server-side-fallback-2026-06-01`.
- **Why:** The Claude API guide sets Opus 5.5 as the default model. Low effort keeps the call fast and cheap for a short extraction. The fallback lets a false refusal of a normal marketing brief still give an answer.

## D-028 CORS on the api

- **Context:** The frontend runs on port 3000 (nginx) or 5173 (`pnpm dev`). It calls the api on port 8080. The browser treats this as a different origin.
- **Options:** A proxy in nginx and Vite. CORS headers on the api.
- **Choice:** The api sends CORS headers for the origins in `CORS_ORIGINS`. The default has `localhost` and `127.0.0.1` on ports 3000 and 5173. The api answers `OPTIONS` preflight requests.
- **Why:** The frontend already calls `VITE_SERVER_API_URL` directly. CORS needs no change in the nginx and Vite configs.

## D-029 Strictness rules of the rule validator

- **Context:** The validator reads a tree and an answer from Claude. Four points in the spec are open. They are the tree depth, the node type, the keys that do not apply to a status, and duplicate JSON keys.
- **Options:** Count all nodes or only groups for the depth. Decide the node type from `op` or from `field`. Reject or ignore empty values for keys that do not apply. Detect duplicate keys or not.
- **Choice:** The depth counts group levels. The root is level 1. A condition in a level 3 group is valid. A node with a `field` key is a condition. All other nodes are groups. A value of `null`, `""` or `[]` counts as absent for a key that does not apply to the status. A non-empty value for such a key fails. The validator does not detect duplicate JSON keys. The last one wins, and the validator still checks that value.
- **Why:** The example rule has a condition in a level 2 group. A count of all nodes would reject a rule that the spec allows. The tool schema is flat. Thus the model can send empty values for keys that do not apply. A repair round for these would turn a true `unsupported` answer into `segment_invalid`. A duplicate key cannot add a field that the validator does not check, so it gives no way to attack.

## D-030 One content call for each language, at the same time

- **Context:** The audience can have more than one language (D-022). Claude writes the SMS and the email for each language.
- **Options:** One call that returns all languages. One call for each language, in sequence. One call for each language, at the same time.
- **Choice:** One call for each language, at the same time. Each call has its own validation and its own repair round.
- **Why:** A short answer streams faster. A failure in one language does not stop the other languages. The prompt names one language, so Claude cannot mix them.

## D-031 The limit check uses the longest first name of the audience

- **Context:** The SMS and subject limits apply after `{{first_name}}` is filled in. At generation time, the backend does not render a message for each user.
- **Options:** Use a fixed name length. Render for each recipient. Use the longest first name of the language audience.
- **Choice:** The audience query gives the maximum `char_length(first_name)` for each language, for users with consent on one or more channels. The validator renders each field with a name of that length. The send slice checks each rendered message again.
- **Why:** If the longest name fits, each name fits. The check needs one aggregate value and no user row.

## D-032 The frontend sends the rule back, and the backend validates it again

- **Context:** Slice 04 stores no rule. The content endpoint needs the rule to calculate the languages and the longest first name.
- **Options:** Store the preview result on the server and send an id. Send the rule back from the browser.
- **Choice:** The browser sends the brief and the rule. The backend runs the slice 04 validator and the audience query again.
- **Why:** No new storage before slice 06. The validator is the gate, so a changed rule from the browser cannot get past the allowlist.

## D-033 Eager tool streaming for content, and the validator stays the gate

- **Context:** The case asks that the content streams onto the page. By default the API buffers each tool parameter, so the text arrives in one piece.
- **Options:** Stream plain text and parse it. Stream the tool input with `eager_input_streaming: true`.
- **Choice:** The tool `submit_content` has `eager_input_streaming: true` and `strict: true`. The backend parses the partial JSON for the drafts and validates the final input. If the API rejects the pair, the live stream stays and `strict` goes. A tool input that is not valid JSON is replayed as `{}` in the repair round. After a `fallback` block, the backend uses only the tool call after that block.
- **Why:** The tool keeps the answer structured, and the drafts stream. With eager streaming the API does not validate the input, but the backend validator always checks it.

## D-034 The mock gets its default answers from cmd/api

- **Context:** One mock client serves the segment and the content endpoints. `internal/llm` must not import other `internal/` packages.
- **Options:** Copy the content into `internal/llm`. Add a language field to `llm.Request`. Inject a default function from `cmd/api`.
- **Choice:** `cmd/api` gives the mock a function that selects the answer from the tool name. `internal/content` owns the default content and reads the language from its own user message.
- **Why:** Each package keeps its own data. `llm.Request` gets no field that only the mock uses.

## D-035 Stream details of the content endpoint

- **Context:** The content endpoint sends `draft` events while Claude writes. The spec did not fix four details. They are the draft throttle after a retry, the last draft, a client that disconnects, and a response that cannot flush.
- **Options:** Reset the throttle after a retry or keep it. Always send the last draft or skip it when the text is the same. Send `error` and `done` after a disconnect or send nothing. Ignore a missing flusher or return an error.
- **Choice:** The draft throttle does not reset after a retry. The handler skips the last draft when the text did not change. After a client disconnect, it sends no `error` and no `done` event. If the response cannot flush, it returns `500` with `internal_error`. `net/http` can always flush, so this path is a guard only.
- **Why:** A throttle that resets would let a retry send a burst of drafts. A repeated draft adds no information. No client reads the events after a disconnect. A stream that cannot flush is not a stream, so the handler fails early and clearly.

## D-036 The content endpoint saves the draft

- **Context:** The user wants drafts. Claude creates a draft when it finishes the content.
- **Options:** The browser posts the final content to a create endpoint. The content endpoint saves the draft at the end of the stream.
- **Choice:** The content endpoint saves the draft after all languages are final or failed. It sends a `campaign` event before `done`. It saves no draft when all languages fail or after a client disconnect.
- **Why:** The stored content is always content that passed the validator. The browser cannot store content that Claude did not write. A disconnected client did not see the result, so a draft would surprise the marketer.

## D-037 "Regenerate content" replaces the same draft

- **Context:** Each content run can make a draft. A marketer who clicks "Regenerate content" several times would get many drafts.
- **Options:** Make a new draft for each run. Replace the draft of the run.
- **Choice:** "Regenerate content" sends `campaign_id`. The backend replaces the content of that draft if the brief and the rule are the same and the campaign is a draft. Else it sends `campaign_conflict`. A new "Find audience" makes a new draft. The rule compare uses `jsonb` equality.
- **Why:** One brief and one audience give one campaign. Postgres `jsonb` changes the key order, so a byte compare would fail on each regenerate.

## D-038 Save and approve check the content again with the current audience

- **Context:** The browser sends edited text. The audience can change between the draft and the approval.
- **Options:** Trust the browser counters. Check with the `first_name_runes` of the draft. Run the audience query again.
- **Choice:** Each save and each approval parses the stored rule, runs the audience query again and checks the content with the new longest name. The counts in the campaign update. A failed check also stores the new counts, so the window shows the new limit. The browser sends no rule after the draft exists.
- **Why:** A new user with a longer name could make an SMS longer than 160 characters at send time. The backend is the gate, as in D-025 and D-032.

## D-039 A row lock stops a double approval

- **Context:** A double click, two tabs or a retried request can approve one draft two times.
- **Options:** An idempotency key from the browser. A row lock and a status check.
- **Choice:** The service counts the audience first, with no lock. Then the approval runs `SELECT ... FOR UPDATE`, checks the status, writes the content and sets `approved` in one transaction. The lock never waits for the audience query. A repeated approval returns `200` with the campaign and changes nothing. A save on an approved campaign returns `409`.
- **Why:** The draft id already names the campaign. A lock gives one approval with no extra key or table. A test with ten parallel approvals proves it.

## D-040 Triggers freeze an approved campaign

- **Context:** After approval, the content must not change. The send step must send what the marketer approved.
- **Options:** Check the status in the store code only. Add database triggers.
- **Choice:** A trigger on `campaign_contents` rejects an insert, update or delete when the campaign is not a draft. A trigger on `campaigns` rejects a change of the brief, rule, description, skipped languages or approval time after approval, and a change back to `draft`. The status can still move forward for slice 07.
- **Why:** The rule holds also for code that a later slice adds and for manual SQL.

## D-041 The campaign stores counts, not recipients

- **Context:** The review window and the campaign page show how many people get each channel.
- **Options:** Store the recipient list at approval. Store only the counts.
- **Choice:** Each content row stores the SMS count, the email count and `first_name_runes` of the last save. The send step gets the recipients with `segment.Recipients` at send time.
- **Why:** No user row goes into the campaign tables. Consent can change before the send, and the send step must use the current consent.

## D-042 Edits are saved with "Save draft" or with the approval

- **Context:** The marketer edits in the review window and can close it.
- **Options:** Autosave each change. Save on a button.
- **Choice:** "Save draft" and "Approve & send" save. Each way to close the window with unsaved edits asks "Discard your changes?". The edits stay while the page stays open.
- **Why:** A save runs the audience query. Autosave would run it on each key press. A clear button gives a clear moment of save.

## D-043 Sample first names for the preview

- **Context:** The preview shows the message with a real name from the audience.
- **Options:** A fixed name. Real users with all fields. First names only.
- **Choice:** `GET /api/campaigns/{id}/sample` returns up to 5 distinct first names of reachable users for each language. The first is the longest. The others come from the lowest user ids. The query reads no surname, email or phone. If the request fails, the preview uses "Alex". The sender shows as "Your brand".
- **Why:** The longest name shows the worst case for the limits. First names are enough for `{{first_name}}`. The sender belongs to the send step.

## D-044 The marketer cannot add or remove a language

- **Context:** The user decided that a failed language is skipped. The spec did not give a way to leave out a final language.
- **Options:** A toggle for each language. Only failed languages are skipped.
- **Choice:** The save and approval must have one entry for each language of the draft. If a language has no reachable users at approval, the approval fails with `content_invalid` on `contents[i].language`. The backend does not drop the language silently.
- **Why:** The user asked to send what Claude made. A silent drop would send less than the confirm popup showed.

## D-045 Close paths and the single-line subject

- **Context:** The review window can close by the close button, "Cancel", Escape or a click outside. The spec asks for an error on a line break in the subject.
- **Options:** Handle each close event. Handle the dialog `update:open` event.
- **Choice:** All close paths go through the dialog `update:open` event, which asks first when edits are not saved. The subject uses a single-line input, so the browser drops line breaks. The backend still rejects `\n` and `\r` in the subject.
- **Why:** One path cannot miss a way to close. A browser cannot send a line break from a single-line input, so the browser check is for paste only, and the backend is the gate.
