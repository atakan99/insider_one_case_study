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

## D-046 The claim is one conditional UPDATE

- **Context:** Two consumers can get the same stream entry. Redis delivers at least once. The mock provider does no dedup (D-018).
- **Options:** A Redis lock for each send job. A Postgres row lock. One conditional `UPDATE`.
- **Choice:** The claim is one `UPDATE` that sets `sending` and counts the attempt only when the status is `queued` or `retrying`. A consumer that updates no row drops the entry.
- **Why:** The database already holds the delivery status. One statement is atomic. Thus Redis is not the dedup gate and the worker needs no second lock.

## D-047 Five attempts with backoff

- **Context:** The mock provider returns `429` and `5xx` at random. The worker must retry without a storm.
- **Options:** A fixed delay. Exponential backoff with no jitter. Exponential backoff with jitter.
- **Choice:** The worker makes 5 attempts at most. The delay before attempts 2, 3, 4 and 5 is 1, 2, 4 and 8 s, plus up to 20% jitter. A `429` counts as an attempt and pauses the channel for `Retry-After`.
- **Why:** Jitter spreads the retries. A pause on the channel stops all consumers from hitting a provider that asked for a wait.

## D-048 The attempt timeout is 3 s

- **Context:** The `slow` response of the mock provider takes 3 s.
- **Options:** A timeout above the slow delay. A timeout below the slow delay.
- **Choice:** `SEND_TIMEOUT` defaults to 3 s. A `slow` response times out and gives `unknown`.
- **Why:** A timeout is ambiguous, because the provider can have sent the message (D-019). The worker does not retry it.

## D-049 32 consumers for each channel

- **Context:** The target is 100 requests per second for each channel. Each `slow` response holds one consumer for 3 s.
- **Options:** Few consumers. 32 consumers. One goroutine for each send job.
- **Choice:** `SEND_CONSUMERS` defaults to 32 for each channel.
- **Why:** The rate limiter is the cap, not the consumers. 32 consumers keep 100 requests per second even when some wait for a `slow` response.

## D-050 The worker plans approved campaigns

- **Context:** The approval is one Postgres transaction (D-040). A push to Redis in the same request can fail on its own.
- **Options:** The approval pushes to Redis. The worker plans the approved campaigns.
- **Choice:** The worker finds the `approved` campaigns and writes the `send_jobs` rows. The unique key on campaign, user and channel blocks a second plan. Then the campaign becomes `queued`.
- **Why:** The approval stays one transaction and does not need Redis. A crash after the plan and before the push is safe, because the recovery pushes the send jobs again (FR-27).

## D-051 One stream for each channel

- **Context:** A slow SMS provider must not stop the email channel.
- **Options:** One stream for all channels. One stream for each channel.
- **Choice:** SMS and email each have a stream, a consumer group, a consumer pool and a rate limiter.
- **Why:** The two channels share no resource that can block them (FR-20).

## D-052 A stream entry has only the send job id

- **Context:** The stream must not hold a second copy of the message or the user.
- **Options:** Put the full message in the entry. Put only the id.
- **Choice:** The entry has the field `job` with the send job id. Postgres is the record of the delivery status. The stream has `MAXLEN ~ 1000000`. The consumer group starts at `0`, so entries that came before the group are delivered. A consumer acknowledges and drops a poison entry (no `job` field, or a value that is not a number).
- **Why:** The entry holds no personal data. A lost or repeated entry cannot change the delivery status. A poison entry would block the group if the worker kept it.

## D-053 The worker waits for a token before it claims

**D-084 replaced this order. D-086 restored it and added a limit on the token age.**

- **Context:** A consumer can wait a long time for a rate-limit token. A claimed send job would show `sending` while it waits.
- **Options:** Claim, then wait. Wait, then claim.
- **Choice:** The consumer waits for a token first, then claims. The limiter checks the `429` pause again after the token wait. On `SIGTERM`, a send job that the worker claimed but did not send goes back to `queued`. Only a crash or a started request gives `unknown`. There is one worker replica, with the limiter in the process. A shared limiter is a non-goal.
- **Why:** The `sending` status then means that a request is in progress or started. A stop during a wait loses no message and creates no false `unknown`.

## D-054 A delayed retry waits in Postgres

- **Context:** A retry has a delay of up to 8 s.
- **Options:** A Redis sorted set. A delayed entry in the stream. A `next_attempt_at` column.
- **Choice:** A `retrying` send job keeps `next_attempt_at` in Postgres. The worker puts it on the stream when the time has come.
- **Why:** Postgres is already the record. After a crash, the recovery finds the same rows. No second store needs a repair.

## D-055 A retry poller runs every 500 ms

- **Context:** The recovery runs every 5 s. A retry with a 1 s delay cannot wait for it.
- **Options:** Use only the recovery. Add a fast poller.
- **Choice:** A poller every 500 ms puts due retries on the stream. It keeps an in-memory set of the retries that it pushed and does not push one again for 30 s. The recovery is the fallback.
- **Why:** A due retry goes on the stream within 1 s (FR-12). The set stops repeat pushes that waste limiter tokens.

## D-056 A hanging `sending` send job becomes `unknown`

- **Context:** After a crash, a send job can stay in `sending`. The worker can have sent the message.
- **Options:** Set it to `queued`. Set it to `unknown`.
- **Choice:** A `sending` send job older than `SEND_TIMEOUT` plus 30 s becomes `unknown`. `Worker.StaleAfter` (default 120 s) is a test hook for the `queued` recovery. `Release` and `CompleteCampaigns` use the database clock. The other store methods use the time that the caller passes.
- **Why:** A retry could send the message two times. The database clock avoids a difference between the worker and the database.

## D-057 The worker checks consent and renders at send time

- **Context:** Consent and the first name can change between the plan and the send. The hard rules need consent for each send.
- **Options:** Store the rendered text at the plan. Render at send time.
- **Choice:** The worker reads the user again, checks the consent of the channel and renders the message. The status `skipped` records why the worker did not send: `no_consent` or `over_limit`. A user who no longer exists gives `skipped` with `no_consent`. A rule that does not parse at plan time gives a `completed` campaign with no send jobs, and the worker logs the error.
- **Why:** The worker never sends without the current consent. It never sends a text over the limits (FR-16, FR-17).

## D-058 `httptrace` tells "not sent" from "maybe sent"

- **Context:** A connection error can happen before or after the request leaves. Only the first case is safe to retry (D-019).
- **Options:** Treat each error as ambiguous. Read the error text. Trace the write.
- **Choice:** The sender uses `httptrace` `WroteRequest`. A connection error before the write gives a retry. A timeout or a lost connection after the write gives `unknown`. `Provider.Send` runs on a context without cancel. `HTTP.Timeout` bounds it, so started requests finish during shutdown.
- **Why:** The trace gives a fact, not a guess from an error string. A request that already started finishes, so shutdown does not create an unknown result.

## D-059 A `retryable` column marks the failed send jobs

- **Context:** The marketer can retry only some failed send jobs. A `400` stays failed.
- **Options:** The api parses `last_error`. A column.
- **Choice:** The worker sets `retryable` when it ends a send job as `failed`. The retry endpoint uses it and the campaign must be `completed`. `last_error` stays after a later `sent`. The spec wins over the plan line "LastError is o.Err in each case".
- **Why:** The api does not depend on the format of an error text. The history of the last error stays for the support view.

## D-060 The retry sets `attempts` to 0

- **Context:** A retried send job is at its limit of 5 attempts.
- **Options:** Keep the count. Reset the count.
- **Choice:** The retry sets `attempts` to 0 and the status to `queued`. The campaign becomes `sending`. A second retry at the same time gets `409`.
- **Why:** The marketer expects a full new set of attempts. A reset is easy to explain.

## D-061 Progress and list limits

- **Context:** The page needs live counts and a list of send jobs without personal data.
- **Options:** Send each change. Send at most one event for each second. Poll only.
- **Choice:** The progress stream sends at most one event for each second. The list returns the user id only. ListJobs has a default limit of 50 and a maximum of 200. The service checks the campaign first, so an unknown id gives `404` and a draft gives `409`.
- **Why:** One event for each second is enough for a bar. A list with ids only holds no personal data (NFR-3).

## D-062 `go-redis/v9` is the Redis client

- **Context:** The worker needs streams, consumer groups, `XAUTOCLAIM` and a blocking read.
- **Options:** `go-redis/v9`. `rueidis`. A hand-written client.
- **Choice:** `github.com/redis/go-redis/v9`. The pool size is `2*Consumers+16`, because each blocking `XREADGROUP` holds one connection.
- **Why:** It is the common client with full stream support. A pool that is too small would starve the blocking consumers.

## D-063 Outcome colours are theme tokens

- **Context:** The send views show five outcomes. FR-46 needs theme tokens for the light and dark themes.
- **Options:** Fixed colours. New tokens.
- **Choice:** Add `--success`, `--success-foreground` and `--info` for both themes. Sent is success, retrying is info, failed is destructive and unknown is warning.
- **Why:** The popup and the list then change with the theme.

## D-064 The seed reset also clears `send_jobs`

- **Context:** `send_jobs.user_id` references `users`. `SEED_RESET=true` truncates `users`.
- **Options:** Leave `send_jobs`. Truncate it too.
- **Choice:** The seed reset also truncates `send_jobs`.
- **Why:** The truncate of `users` would fail with a foreign key error.

## D-065 The page owns the progress stream

- **Context:** The page and the popup both show progress. After a stream error, the page needs a fallback.
- **Options:** The popup opens its own stream. The page opens one stream and passes progress as a prop.
- **Choice:** The page owns the stream. A 409 on retry makes the page refresh its progress. After a stream error, the page polls `GET .../progress` every 2 s and does not open the stream again.
- **Why:** One stream avoids two sources that disagree. Polling is simple and the data is the same (FR-40).

## D-066 Duplicate queue entries use no rate token

**D-084 removed the `Store.Claimable` check. D-086 restored it.**

- **Context:** The recovery pushes stale queued jobs again. On a long backlog, an entry waits longer than the stale age. The recovery then adds a second entry for each job. Each duplicate entry used a rate token before `Claim` rejected it. This slowed the real sends and made the backlog bigger on each recovery pass.
- **Options:** Raise the stale age. Let `Claim` stay the only check. Add a cheap check before the limiter and gate the re-push.
- **Choice:** Before `Limiter.Wait`, the consumer calls `Store.Claimable`. A job that is not `queued`, and not `retrying` with `next_attempt_at` in the past, is acknowledged with no token used. If the check fails, the consumer continues on the normal path. The `Claim` UPDATE stays the real gate. The recovery pushes the stale queued jobs of a channel only when `XINFO GROUPS` shows `lag` 0 for that channel. A missing stream or group counts as lag 0. An unknown lag (negative), or an error, counts as no backlog, so the jobs are pushed.
- **Why:** An unread entry means the job already waits on the stream, so a push adds no value. The safe direction on a doubt is a push, because a duplicate entry is harmless and a lost job is not. The check costs one indexed read for each entry. At a much larger scale, a second entry still costs this read. A larger system must then bound the stream. It can also make the stale age follow the backlog.

## D-067 The page does not reopen the progress stream

- **Context:** `goal.md` says that the stream closes after 30 minutes and that the page opens it again. D-065 gave the page a polling fallback for a stream error.
- **Options:** Reopen the stream after an error or the 30-minute close. Poll until the campaign is `completed`.
- **Choice:** After a stream error or the 30-minute close, the page polls `GET .../progress` every 2 s until the status is `completed`. It does not reopen the stream. This differs from `goal.md`, which says "The page opens it again".
- **Why:** One fallback path is simpler than a reconnect loop. The polled data is the same as the stream data. The cost is one request every 2 s for a campaign that runs longer than 30 minutes.

## D-068 `/` goes to `/dashboard`

- **Context:** Before slice 08, `/` went to `/campaigns/new`. The case brief names a dashboard as the place for the delivery status.
- **Options:** Keep `/campaigns/new`. Go to `/dashboard`.
- **Choice:** `/` goes to `/dashboard`. The sidebar lists "Dashboard" first.
- **Why:** The reviewer sees the statistics first.

## D-069 Two dashboard endpoints

- **Context:** The page needs statistics and the live state of the services. A health check waits up to 1 s for a service that does not answer.
- **Options:** One endpoint for all data. Two endpoints, `stats` and `health`.
- **Choice:** `GET /api/dashboard/stats` and `GET /api/dashboard/health`. `health` checks the four services at the same time, each with a timeout of 1 s. It returns `200` also when a service is `down`.
- **Why:** A slow health check does not delay the statistics. The body tells the state, so the page can show a `down` badge.

## D-070 The dashboard polls every 5 s

- **Context:** The totals change while a campaign runs. Slice 07 uses Server-Sent Events for one campaign.
- **Options:** Server-Sent Events. Polling every 5 s.
- **Choice:** The page polls both endpoints at once every 5 s. It stops while the tab is hidden and polls again when the tab shows.
- **Why:** Polling is simple. A delay of 5 s is enough for totals.

## D-071 Throughput uses 5-second buckets

- **Context:** The chart must show that the worker keeps each channel at 100 messages per second or fewer.
- **Options:** Buckets of 1 minute. Buckets of 5 seconds.
- **Choice:** 120 buckets of 5 s cover the last 10 minutes. A bucket value is the count of `sent_at` in the bucket divided by 5.
- **Why:** A short campaign still shows its peak rate. A bucket of 1 minute hides the peak.

## D-072 A failure reason is the status and `last_error`

- **Context:** The failure card must show why messages did not go out.
- **Options:** Parse `last_error` into codes. Use the pair of delivery status and `last_error`.
- **Choice:** One row is one pair of status and `last_error`. The worker writes a short, fixed text there.
- **Why:** No parse is necessary.

## D-073 `/metrics` uses `prometheus/client_golang`

- **Context:** The case asks for a metrics endpoint. The format is Prometheus text.
- **Options:** A hand-written text writer. The official client library.
- **Choice:** `github.com/prometheus/client_golang`. This is a new Go dependency. Each `/metrics` also has the Go and process collectors.
- **Why:** The library writes the format correctly and has the histogram type.

## D-074 Api gauges come from Postgres, worker counters live in the process

- **Context:** The api and the worker both report metrics.
- **Options:** Count events in each process. Read the state from the database.
- **Choice:** The api gauges `campaign_send_jobs` and `campaign_campaigns` come from one `GROUP BY` query at each scrape. A query error gives no gauge values in that scrape and a log line. The worker counters live in the worker process and start at 0 after a restart. The worker stream gauges come from Redis at each scrape.
- **Why:** The database is the source of truth, and it survives a restart of the api. A counter needs no database read.

## D-075 No metric has a campaign id label

- **Context:** A label with the campaign id adds one series for each campaign.
- **Options:** Label by campaign id. Label by channel and status only.
- **Choice:** No metric has a campaign id label.
- **Why:** The number of series stays small. The labels also hold no personal data.

## D-076 The api `/readyz` checks Redis

- **Context:** The api reads the streams for the `health` endpoint.
- **Options:** Check Postgres only. Check Postgres and Redis.
- **Choice:** `/readyz` checks both. It returns `503` when one is down, and the body has `db` and `redis`. `/healthz` stays as it is.
- **Why:** An api that cannot reach Redis cannot serve `health`.

## D-077 Implementers committed each task

- **Context:** The plan said not to commit per task. The tasks ran in subagents on the branch `08-dashboard-metrics`.
- **Options:** Leave all changes uncommitted until the end. Commit after each task.
- **Choice:** Each implementer committed its task on the branch. Review fixes got their own commits.
- **Why:** A commit for each task gives the reviewers a fixed diff range. This differs from the plan.

## D-078 Failure rows with equal totals sort by status, then reason

- **Context:** `failures` is ordered by the total count, highest first. Rows with the same total had no fixed order.
- **Options:** Leave the order to Postgres. Add a fixed tie-break.
- **Choice:** Rows with equal totals sort by status, then by reason.
- **Why:** The top 8 and the page do not change between two polls with the same data.

## D-079 `dashboard.RedisPinger` adapts go-redis

- **Context:** The health checker needs a ping and the stream counts. Only the ping must be behind an interface, so that a test can fake it. The stream counts use `delivery.Queue`, which uses go-redis.
- **Options:** Use go-redis in the checker. Define a small `Pinger` interface.
- **Choice:** The checker takes a `Pinger` interface. `dashboard.RedisPinger` adapts the go-redis client to it.
- **Why:** A test can give the checker a fake, and the checker has one small dependency.

## D-080 Unknown HTTP methods count as "other"

- **Context:** The `method` label of the api request metrics comes from the client. A client can send any text as a method.
- **Options:** Use the raw method. Clamp unknown methods to "other".
- **Choice:** The metrics middleware keeps the standard methods and records every other method as `other`. The CORS middleware wraps the metrics middleware. Thus a preflight `OPTIONS` that CORS answers is not counted.
- **Why:** A client cannot add series without limit. A preflight is not an api call.

## D-081 The label check excludes four label names

- **Context:** FR-19 says no label value is a user id or a campaign id. Some labels carry numbers by design: `le` and `quantile` hold bucket bounds, `code` holds the HTTP status and `version` holds the Go version.
- **Options:** Check every label value for a number. Skip the four names and check the rest.
- **Choice:** The test skips the label names `le`, `quantile`, `code` and `version`. For all other labels, it checks that no value equals a test user id or campaign id and that no value contains `@`.
- **Why:** A check for any number would fail on the allowed labels. A value that equals a real id, or an email, is the real leak.

## D-082 "Last update" shows the last time both calls succeeded

- **Context:** The page makes two calls. One call can fail while the other succeeds.
- **Options:** Show the last success of any call. Show the last time both calls succeeded.
- **Choice:** After a partial failure, "Last update" shows the last time both calls succeeded.
- **Why:** A partial refresh leaves part of the page old. The time must not claim that the page is current.

## D-083 Redis clients honor context deadlines where a check or a scrape uses them

- **Context:** go-redis v9 sets `ContextTimeoutEnabled` to false by default. A context deadline then does not bound a read. The read timeout of 5 s applies instead. The 1 s timeout of the health checks, `/readyz` and the worker stream collector did not hold against a slow Redis.
- **Options:** Set the option on every client. Set it on the api client and give the worker stream collector its own client.
- **Choice:** The api client sets `ContextTimeoutEnabled`. The worker main client does not change. The stream collector uses `delivery.Queue` on a second client with a pool of 2 and `ContextTimeoutEnabled`.
- **Why:** The blocking reads of the worker consumers must keep the timing of slice 07. A second small client bounds the scrape without a change to the consumers.

## D-084 The worker waits for a token immediately before the request

**Replaced by D-086. This order caused two defects.**

- **Context:** The consumer took a rate token, then claimed and loaded the send job from Postgres (D-053). A slow database held sends that already had a token. When the database answered, these sends left together with the newly permitted sends. The channel could then go above 100 requests per second. A test with 7 held sends showed 7 requests in less than 1 ms.
- **Options:** Keep the order and add a second limiter before the request. Move the token wait to the last step before the request.
- **Choice:** The order is claim, load, wait, send. No database work comes between the token and the request. A cancel or an error before the request releases the send job to `queued`. `Store.Claimable` is removed, because the claim now rejects a duplicate entry before the token wait (D-066).
- **Why:** One limiter at the request is the only place that bounds the real request rate. A second limiter adds state and gives no more safety. The cost is that a send job shows `sending` during its token wait. With 32 consumers at 100 per second, this wait is 320 ms or less. After a crash, the recovery sets these send jobs to `unknown`. This is at most 32 for each channel. A `Retry-After` longer than `SEND_TIMEOUT` plus 30 s can also set a waiting send job to `unknown`. The mock provider sends 1 s.

## D-085 A save keeps the text that the marketer typed during the request

- **Context:** The review window stays editable while a save is in progress. The response replaced all fields. Text that the marketer typed after the submit disappeared, with no message.
- **Options:** Disable the fields during the save. Keep the fields that changed after the submit.
- **Choice:** After the response, a field that differs from the submitted value keeps the typed text. All other fields take the saved value. The kept text counts as unsaved, so a close asks before it discards.
- **Why:** A disabled field drops the focus and stops the marketer for each save. The saved state stays correct, because "unsaved" compares the fields with the response.

## D-086 A rate token is good for 100 ms

- **Context:** D-084 put the token wait after the claim and the load. A review found two defects in that order. The load reads the consent, so a change of consent during the wait did not stop the send. The send job was `sending` during the wait, so the recovery could set a live send job to `unknown`. The first order (D-053) had neither defect, but a slow database could release held sends together.
- **Options:**
  - Keep the D-084 order and read the consent again after the wait. This puts database work after the token again.
  - Keep the D-084 order and add a column that separates a waiting send job from a started request.
  - Restore the D-053 order and limit the age of a token.
- **Choice:** The order is wait, claim, load, send. The consumer measures the time from the token to the end of the load. If this time is more than `Consumer.MaxPrep` (100 ms), the consumer releases the send job to `queued`. It then takes a new token and claims again. `Store.Claimable` is back (D-066).
- **Why:** The consent read is the last step before the request. A send job is `queued` during a token wait, so the recovery cannot expire it. One rule covers the three defects and needs no migration.
- **Limits:**
  - Sends whose tokens are less than 100 ms apart can leave together. At 100 per second, this is 10 requests at most. It is equal to the burst of the mock provider.
  - When the claim and the load always take more than 100 ms, no message is sent. The worker logs a warning for each new token.
  - A compose run with 1,803 requests gave `rate_limited: 0`, no such warning, and 98 sends at most in one second for each channel.

## D-087 The read-only pages reuse the message preview

- **Context:** The campaign page and the compose page showed the SMS, the subject and the body as plain text across the full card width. Lines had about 150 characters. The labels looked like the copy. The review window already had a preview with an SMS bubble and an email frame.
- **Options:** Restyle the plain text. Reuse `message-preview.vue` on both pages.
- **Choice:** A new `content-view.vue` shows the SMS preview next to the email preview. The campaign page and the compose page use it. `message-preview.vue` takes an optional name. With no name, `{{first_name}}` shows as a chip.
- **Why:** One component gives the three places the same look. The chip keeps the token that the marketer types in the review window, so the two views agree. The email body has a `max-w-prose` limit.

## D-088 The campaign page shows one language at a time

- **Context:** The campaign page stacked one card for each language. For a sent campaign, the delivery section was below all the email bodies.
- **Options:** Keep the stack and move the delivery section up. Show one language with tabs and add a short delivery summary at the top.
- **Choice:** Tabs select the language when the campaign has more than one. A summary card below the header shows a stacked bar for each channel and a link to the delivery section. The page computes the lengths with `renderedLength` and `first_name_runes`, the same as the review window.
- **Why:** The page length does not grow with the number of languages. The marketer sees the result of the send without a scroll. The summary uses the progress that the page already loads, so the API does not change.

## D-089 Counts, dates and names use one format

- **Context:** Pages showed `tr`, `TR` and "Turkish" for the same language. Some counts had no thousands separator. The channel order was different between pages. Dates used the browser locale.
- **Choice:** `lib/format.ts` has `formatCount` and `formatDateTime`, both in `en-US`. All pages use `languageName`. The channel order is SMS, then email. The dashboard column "Send jobs" is now "Messages". The page content has a `max-w-6xl` limit in the layout.
- **Why:** The UI is English only (D-022). "Send job" is an internal term that the marketer does not know.
- **Limits:** A click on a draft row in the campaign list still opens the review window. The tests of slice 06 require this. The list has no recipient count, because `GET /api/campaigns` does not return one.

## D-090 A send gate keeps the request starts apart

- **Context:** D-086 lets a token wait up to 100 ms before its request. A review sent slow and fast preparations together. The provider got 108 requests in one second at a limit of 100.
- **Options:**
  - Make `MaxPrep` shorter. This makes the bunch smaller but does not remove it.
  - Take the token after the load. This brings back the defects that D-086 fixed.
  - Add a second gate just before the request.
- **Choice:** `Limiter.Admit` runs after the load and before `Provider.Send`. It keeps the request starts of one channel at least `1s/perSec` plus 2% apart. If the context stops during `Admit`, the consumer releases the send job to `queued`. The token wait and `MaxPrep` stay as in D-086.
- **Why:** With a minimum gap between request starts, one second cannot hold more than `perSec` requests. The 2% margin covers network jitter between the worker and the provider. The token still paces the claims, so a send job stays `queued` during a long wait.
- **Limits:**
  - The maximum rate is approximately 98 per second for each channel.
  - The consent read can be older than the request by the `Admit` wait. This wait is normally less than `MaxPrep`.
  - The gate is in memory. Two worker processes would each send at the full rate.
