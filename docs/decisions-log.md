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
- **Choice:** The api sends CORS headers for the origins in `CORS_ORIGINS`. The default is `http://localhost:3000,http://localhost:5173`. The api answers `OPTIONS` preflight requests.
- **Why:** The frontend already calls `VITE_SERVER_API_URL` directly. CORS needs no change in the nginx and Vite configs.

## D-029 Strictness rules of the rule validator

- **Context:** The validator reads a tree and an answer from Claude. Four points in the spec are open. They are the tree depth, the node type, the keys that do not apply to a status, and duplicate JSON keys.
- **Options:** Count all nodes or only groups for the depth. Decide the node type from `op` or from `field`. Reject or ignore empty values for keys that do not apply. Detect duplicate keys or not.
- **Choice:** The depth counts group levels. The root is level 1. A condition in a level 3 group is valid. A node with a `field` key is a condition. All other nodes are groups. A value of `null`, `""` or `[]` counts as absent for a key that does not apply to the status. A non-empty value for such a key fails. The validator does not detect duplicate JSON keys. The last one wins, and the validator still checks that value.
- **Why:** The example rule has a condition in a level 2 group. A count of all nodes would reject a rule that the spec allows. The tool schema is flat. Thus the model can send empty values for keys that do not apply. A repair round for these would turn a true `unsupported` answer into `segment_invalid`. A duplicate key cannot add a field that the validator does not check, so it gives no way to attack.
