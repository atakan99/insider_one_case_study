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
