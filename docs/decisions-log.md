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
