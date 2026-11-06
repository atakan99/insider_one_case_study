# 00 — Repo Skeleton Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A repo where `docker compose up` starts Postgres, the Go API and the mock provider, `make verify` passes, and every later slice has a fixed place for its spec, plan, decisions and progress.

**Architecture:** Backend written from scratch, no template. One Go module in `backend/`, built by one Dockerfile with a `CMD` build arg. Target package map (this slice creates only the **bold** ones, later slices add the rest):

```
cmd/   **api**  worker  **mockprovider**  seed
internal/
  **config**   env loading
  **httpapi**  handlers, health, SSE (later)
  segment    rule DSL, validation, SQL builder
  llm        interface, Claude client, fake
  content    channel limits, {{var}} rendering
  delivery   queue, rate limiter, retry, idempotency
  store      sqlc-generated queries + migrations
```
 Postgres is the only infra service in this slice. Queue choice is deferred to the send slice (see Open Decisions). The frontend is out of scope here: `docs/specs/01-frontend-trim/` owns the move from `shadcn-vue-admin-main/` to `frontend/`.

**Tech Stack:** Go 1.27 (stdlib `net/http` routing, `github.com/jackc/pgx/v5`), Postgres 17, Docker Compose v2, GNU Make.

**Spec:** `docs/case.md` (the case brief, moved from root in Task 1) and the repo layout agreed in session (see `docs/architecture.md`).

## Global Constraints

- Backend is Go (decided: `docs/specs/01-frontend-trim/goal.md` says "our Go API"). Log as D-001.
- App reads the key from `ANTHROPIC_API_KEY` in `.env`. Repo ships `.env.example`, never `.env`.
- `docker compose up` must work on a clean clone **with no `.env` file** (LLM runs in mock mode).
- `make test` must work on a machine with Docker but no Go.
- Tests never need a real API key.
- Required root deliverables: `README.md`, `DECISIONS.md`, `ai-sessions/`, `Makefile`, `docker-compose.yml`, `.env.example`.

## Open Decisions (not needed for this slice)

- **Queue:** Postgres table + `SELECT … FOR UPDATE SKIP LOCKED` (recommended: no extra service, send + idempotency key commit in one transaction) vs Redis Streams vs NATS JetStream. Decide in the send-queue spec.
- **Frontend dev in compose:** serve built `frontend/` via nginx in compose, or keep `pnpm dev` local only. Decide in 01-frontend-trim.

## Review Focus

1. Clean clone, no `.env` → `docker compose up` must still start all services. Pinned by Task 4 Step 4.
2. Postgres down while API is up → `/readyz` must return 503, not 200. Pinned by Task 3 test `TestReadyz_DBDown`.
3. API starts before Postgres accepts connections → compose must wait on the Postgres healthcheck, and the API must not crash-loop. Pinned by Task 4 `depends_on: condition: service_healthy`.
4. Host already uses port 5432 (common on dev machines) → compose must not publish Postgres to the host. Pinned by Task 4 Step 4 (`docker compose port postgres 5432` prints nothing).
5. `.env` with a real key gets committed by accident → `.gitignore` must catch `.env` and `.env.local`, and let `.env.example` through. Pinned by Task 1 Step 2.

---

### Task 1: Repo layout, docs stubs, CLAUDE.md, task tracking

**Files:**
- Move: `insider_one_case.md` → `docs/case.md`
- Move: `docs/prompts/01-frontend-trim/` → `docs/specs/01-frontend-trim/` (then delete empty `docs/prompts/`)
- Create: `.gitignore`, `README.md`, `DECISIONS.md`, `CLAUDE.md`, `ai-sessions/.gitkeep`, `docs/architecture.md`, `tasks/backlog.json`, `tasks/progress.md`
- Leave untouched: `shadcn-vue-admin-main/`, `docs/format.md`

**Interfaces:**
- Produces: the file locations every later slice uses. Specs at `docs/specs/NN-name/`, plans at `docs/plans/NN-name.md`, backlog ids `NN-name`.

- [x] **Step 1: Do the moves.** Use `git mv` once the files are tracked, plain `mv` before the first commit.

- [x] **Step 2: Write `.gitignore`, then verify it.**

Must contain: `.env`, `.env.*`, `!.env.example`, `node_modules/`, `dist/`, `coverage/`, `*.log`, `.DS_Store`, `backend/bin/`, `.claude/settings.local.json`, `.claude/agent-memory-local/`, `.vscode/`.

Run:
```bash
touch .env .env.local && git check-ignore .env .env.local; git check-ignore .env.example; echo "exit=$?"; rm .env .env.local
```
Expected: prints `.env` and `.env.local`, then `exit=1` (meaning `.env.example` is not ignored).

- [x] **Step 3: Write `CLAUDE.md` (≤ 60 lines).** Sections and fixed content:
  - **Layout:** one line per top-level dir (`backend/`, `frontend/`, `docs/`, `tasks/`, `ai-sessions/`).
  - **Commands:** `make up`, `make down`, `make test`, `make lint`, `make verify`.
  - **Workflow:** at session start, read `tasks/progress.md`, then pick the first `backlog.json` item with `"passes": false`. Set `passes` to `true` only after `make verify` passes and the user approves. At session end, append to `progress.md`.
  - **Hard rules:**
    - Never execute LLM output as SQL. Segment rules are JSON, validated against an allowlist of fields and operators, then compiled to parameterized queries.
    - Never send on a channel without that user's consent for that channel.
    - SMS ≤ 160 chars and email subject ≤ 80 chars, checked after `{{first_name}}`-style variables are rendered.
    - Tests use the mock LLM. No test reads `ANTHROPIC_API_KEY`.
    - Every non-obvious choice gets a `DECISIONS.md` entry.
  - **Writing style:** placeholder line. Task 2 Step 9 replaces it.

- [x] **Step 4: Write `DECISIONS.md`.** Entry format: `## D-NNN Title` → **Context** / **Options** / **Choice** / **Why**, ≤ 8 lines each. Seed entries:
  - D-001 Go for the backend (concurrency for workers and rate limiters, one static binary per service).
  - D-002 Postgres as the only datastore for now (segment queries need SQL, queue choice deferred).
  - D-003 Backend from scratch, one Go module with one binary per process (`api`, `worker`, `mockprovider`, `seed`). A template would bring code we did not choose and must defend. The mock provider imports nothing from `internal/` except `config`, so it stays a stand-in for an external service.
  - D-004 `make test` falls back to a `golang` container when Go is missing (graders may only have Docker).
  - D-005 Router: stdlib `net/http` (Go 1.22+ method and path-param patterns) over chi. It covers our routes with zero dependencies. Revisit if we need a middleware chain beyond request-id, recover and timeout.
  - D-006 Data access: `pgx` + `sqlc` for fixed queries. Segment queries are dynamic (the WHERE clause depends on the rule), which sqlc cannot generate, so `internal/segment` compiles the validated rule into SQL with `$n` placeholders by hand.

- [x] **Step 5: Write `tasks/backlog.json`.** An array of `{ "id", "title", "spec", "acceptance": [string], "passes": false }`. Entries, in this order:
  `00-skeleton`, `01-frontend-trim`, `02-seed-users`, `03-mock-provider`, `04-segment-rules`, `05-content-generation`, `06-review-approve`, `07-send-queue`, `08-dashboard-metrics`. Only `00` and `01` have a real `spec` path. The rest have `"spec": null` until their spec exists. `00-skeleton.acceptance` = the Goal sentence above split into checks.

- [x] **Step 6: Write `tasks/progress.md`.** Header + one template block:
  ```markdown
  ## YYYY-MM-DD — <slice id>
  - Done:
  - Verified with:
  - Next:
  - Open questions:
  ```

- [x] **Step 7: Write stubs.** `README.md`: title, one-paragraph summary, "Run" (`docker compose up`), "Test" (`make test`), "Status" (link to `tasks/backlog.json`). `docs/architecture.md`: the component list (api, mock-provider, postgres, frontend), the folder tree, and "TBD per slice" under each flow step.

- [x] **Step 8: Verify and commit.**

Run: `python3 -m json.tool tasks/backlog.json > /dev/null && echo ok`
Expected: `ok`

```bash
git add .gitignore README.md DECISIONS.md CLAUDE.md ai-sessions docs tasks
git commit -m "chore: repo layout, docs stubs, task tracking"
```

---

### Task 2: STE writing skill + Claude subagents (executor, reviewer)

**Files:**
- Create: `.claude/skills/ste-writing/SKILL.md`
- Move: `docs/format.md` → `.claude/skills/ste-writing/references/rules.md`
- Create: `.claude/skills/ste-writing/references/terms.md`
- Create: `.claude/skills/ste-writing/scripts/ste_check.py`, `.claude/skills/ste-writing/scripts/test_ste_check.py`
- Create: `.claude/agents/executor.md`, `.claude/agents/reviewer.md`
- Modify: `CLAUDE.md` (Writing style line)

**Interfaces:**
- Consumes: `CLAUDE.md`, `docs/plans/*.md`, `tasks/backlog.json` from Task 1.
- Produces: skill `ste-writing`. Subagents `executor` and `reviewer`, used by the main (planner) session from slice 01 on. `ste_check.py <file.md>...` prints `file:line: rule: text` per finding and always exits 0.

- [ ] **Step 1: Write `SKILL.md` (≤ 80 lines).**
  - Frontmatter `name: ste-writing`. `description:` "Use when writing specs, plans, progress notes, DECISIONS.md, README, CLAUDE.md, agent prompts, subagent reports or commit messages. Applies ASD-STE100 rules. Never use it for app code, comments, UI text, API errors or the app's LLM prompts."
  - **Scope, in (communication between the user and agents):** specs, plans, `tasks/progress.md`, `CLAUDE.md`, agent files, subagent reports, replies to the user, commit messages, `DECISIONS.md`, `README.md`.
  - **Scope, out (state this explicitly):** the app itself. That includes code, code comments, identifiers, UI text, API error messages, the app's LLM prompts, and the SMS and email copy that the app generates.
  - **Checklist (the rules we enforce):** one instruction per sentence. ≤ 20 words per instruction, ≤ 25 per description. Active voice in instructions. No semicolons. No present perfect ("has been"). Max 3-noun clusters. ≤ 6 sentences per paragraph. Lists for sequences and conditions. Do not drop articles or verbs to save words. Use each term from `terms.md` in its one meaning.
  - **Not enforced:** the ~900-word STE dictionary (we cannot check it). Use plain common words instead.
  - Three before/after pairs taken from this repo's own docs.
  - Pointer: "Run `scripts/ste_check.py` on changed `.md` files before you report."

- [ ] **Step 2: Write `terms.md`.** A table `term | meaning | do not use for`. Seed terms: brief, segment rule, audience, channel, consent, template (message text with `{{vars}}`), rendered message, campaign, approval, send job, attempt, delivery status, idempotency key. One meaning per term. Later slices add rows.

- [ ] **Step 3: Write failing tests** in `test_ste_check.py` (stdlib `unittest`):
  - `test_semicolon`: `"Do X; do Y."` → one finding, rule `semicolon`.
  - `test_long_sentence`: a 26-word sentence → rule `sentence-length`. A 20-word sentence → no finding.
  - `test_present_perfect`: `"The job has been sent."` → rule `present-perfect`.
  - `test_skips_code`: a semicolon inside a fenced code block or inline backticks → no finding.
  - `test_skips_tables_and_headings`: lines that start with `|` or `#` → no sentence-length finding.

- [ ] **Step 4: Run, expect FAIL.** `python3 -m unittest discover -s .claude/skills/ste-writing/scripts` → import failure.

- [ ] **Step 5: Implement `ste_check.py`.** Stdlib only. Strip fenced blocks and inline code first. Split sentences on `. ! ?` followed by a space or end of line. Use the 25-word limit. The checker cannot tell an instruction from a description, so the 20-word limit stays a review item. Detect present perfect with `\b(has|have|had) been\b` and `\b(has|have) [a-z]+ed\b`.

- [ ] **Step 6: Run, expect PASS.** Same command as Step 4.

- [ ] **Step 7: Write `executor.md`.**
  ```yaml
  name: executor
  description: Implements ONE task from an approved plan in docs/plans/. Give it the plan path and task number.
  model: sonnet
  tools: Read, Edit, Write, Bash, Grep, Glob
  skills:
    - superpowers:test-driven-development
    - superpowers:verification-before-completion
    - ste-writing
  ```
  Body rules (written in STE): read `CLAUDE.md`, the plan, and the plan's spec. Do only the given task. Write tests first. Run `make verify` before you report. Do not commit. The user commits. Report: files changed, the tail of `make verify` output, each deviation from the plan and its reason.

- [ ] **Step 8: Write `reviewer.md`.**
  ```yaml
  name: reviewer
  description: Read-only review of a finished task or slice against its spec, plan and CLAUDE.md hard rules.
  model: opus
  tools: Read, Grep, Glob, Bash
  memory: project
  skills:
    - ste-writing
  ```
  Body rules: check the diff (`git diff` / `git diff --staged`) against the task's tests, the plan's Review Focus, and the `CLAUDE.md` hard rules. Report each defect as `file:line — problem — failing input`. Report STE findings only for changed docs, in a short separate section after the defects. Do not edit files. Record recurring defect patterns in memory.

- [ ] **Step 9: Update `CLAUDE.md`.** Replace the Writing style line with: "Use the `ste-writing` skill for specs, plans, docs, reports and commit messages. Do not apply it to anything in `backend/` or `frontend/`: code, comments, UI text, API errors, LLM prompts and generated copy."

- [ ] **Step 10: Verify.** In Claude Code run `/agents`, then `/skills`.
  Expected: `executor` (sonnet) and `reviewer` (opus) appear under project agents. `ste-writing` appears under project skills.

- [ ] **Step 11: Commit.**
  ```bash
  git add .claude docs CLAUDE.md
  git commit -m "chore: add ste-writing skill, executor and reviewer subagents"
  ```

---

### Task 3: Go module, config, health endpoints, two binaries

**Files:**
- Create: `backend/go.mod` (module `campaigncomposer`, `go 1.27`)
- Create: `backend/internal/config/config.go`, `backend/internal/config/config_test.go`
- Create: `backend/internal/httpapi/health.go`, `backend/internal/httpapi/health_test.go`
- Create: `backend/cmd/api/main.go`, `backend/cmd/mockprovider/main.go`

**Interfaces:**
- Produces:
  - `config.Config{ HTTPAddr, DatabaseURL, LLMMode, AnthropicAPIKey string }`
  - `config.Load(getenv func(string) string) (config.Config, error)`
  - `httpapi.Pinger interface { Ping(ctx context.Context) error }` (`*pgxpool.Pool` satisfies it)
  - `httpapi.RegisterHealth(mux *http.ServeMux, db httpapi.Pinger)`, which adds `GET /healthz` and `GET /readyz`
  - Env vars: `HTTP_ADDR` (default `:8080`), `DATABASE_URL` (required), `LLM_MODE` (`mock` | `anthropic`, default `mock`), `ANTHROPIC_API_KEY`. Mock provider: `HTTP_ADDR` default `:8081`.

- [ ] **Step 1: Write failing config tests** in `config_test.go`. Use a map-backed `getenv`.
  - `TestLoad_Defaults`: only `DATABASE_URL` set → `HTTPAddr == ":8080"`, `LLMMode == "mock"`, err nil.
  - `TestLoad_MissingDatabaseURL`: empty env → err non-nil, message contains `DATABASE_URL`.
  - `TestLoad_AnthropicWithoutKey`: `LLM_MODE=anthropic`, no key → err contains `ANTHROPIC_API_KEY`.
  - `TestLoad_InvalidLLMMode`: `LLM_MODE=gpt` → err contains `LLM_MODE`.

- [ ] **Step 2: Run, expect FAIL.** `cd backend && go test ./internal/config/` → build failure, `Load` undefined.

- [ ] **Step 3: Implement `config.Load`.**

- [ ] **Step 4: Run, expect PASS.** `cd backend && go test ./internal/config/`

- [ ] **Step 5: Write failing health tests** in `health_test.go`. Use `httptest` and a fake `Pinger` with a configurable error.
  - `TestHealthz_OK`: GET `/healthz` → 200, body `{"status":"ok"}`.
  - `TestReadyz_DBUp`: fake returns nil → 200, body `{"status":"ok","db":"ok"}`.
  - `TestReadyz_DBDown`: fake returns `errors.New("conn refused")` → 503, body `{"status":"unavailable","db":"down"}`. The body must **not** contain `conn refused` (no internal error leak).
  - `TestReadyz_PingTimeout`: fake blocks until ctx is done → 503 within 3s. The handler applies a 2s ping timeout.
  - `TestHealthz_MethodNotAllowed`: POST `/healthz` → 405.

- [ ] **Step 6: Run, expect FAIL.** `cd backend && go test ./internal/httpapi/`

- [ ] **Step 7: Implement `httpapi.RegisterHealth`.** Use Go 1.22+ method patterns (`mux.HandleFunc("GET /healthz", …)`) and `Content-Type: application/json`.

- [ ] **Step 8: Run, expect PASS.** `cd backend && go test ./...`

- [ ] **Step 9: Implement `cmd/api/main.go`.** Load config (exit 1 with message on error), `pgxpool.New`, `httpapi.RegisterHealth`, `http.Server` with `ReadHeaderTimeout: 5s`. Shut down gracefully on SIGINT/SIGTERM (10s). Log with `log/slog` JSON. A failed initial DB connection must not exit: `/readyz` reports it.

- [ ] **Step 10: Implement `cmd/mockprovider/main.go`.** `GET /healthz` → `{"status":"ok"}` on `HTTP_ADDR` (default `:8081`), with the same graceful shutdown. Send endpoints, failures and 429s belong to slice `03-mock-provider`.

- [ ] **Step 11: Verify.** `cd backend && go vet ./... && go build ./... && test -z "$(gofmt -l .)" && echo ok` → `ok`

- [ ] **Step 12: Commit.**
  ```bash
  git add backend
  git commit -m "feat(backend): go module, config, health endpoints, api and mockprovider binaries"
  ```

---

### Task 4: Dockerfile, docker-compose, Makefile, .env.example

**Files:**
- Create: `backend/Dockerfile`, `docker-compose.yml`, `Makefile`, `.env.example`, `backend/.dockerignore`

**Interfaces:**
- Consumes: binaries and env vars from Task 3.
- Produces: services `postgres`, `api` (host port 8080), `mock-provider` (host port 8081). Make targets `up`, `down`, `test`, `lint`, `verify`. Later slices add the `seed` service and `frontend`.

- [ ] **Step 1: Write `backend/Dockerfile`.** Multi-stage: `golang:1.27-alpine` builds `./cmd/${CMD}` with `CGO_ENABLED=0`, final stage `gcr.io/distroless/static-debian12:nonroot`. `ARG CMD` is required.

- [ ] **Step 2: Write `docker-compose.yml`.**
  - `postgres`: `postgres:17-alpine`, user/db/password `campaign`, named volume, healthcheck `pg_isready -U campaign` (interval 2s, retries 15), **no `ports:`**.
  - `api`: build `backend` with `CMD: api`, `env_file: [{ path: .env, required: false }]`, `environment` defaults `DATABASE_URL: postgres://campaign:campaign@postgres:5432/campaign?sslmode=disable`, `LLM_MODE: ${LLM_MODE:-mock}`, `ports: ["8080:8080"]`, `depends_on: postgres: condition: service_healthy`.
  - `mock-provider`: build `backend` with `CMD: mockprovider`, `ports: ["8081:8081"]`.
  - Distroless has no shell, so there's no container healthcheck for api/mock-provider. `make up` checks them with curl instead (Step 3).

- [ ] **Step 3: Write `Makefile`.**
  ```make
  ifeq ($(shell command -v go),)
  GO = docker run --rm -v $(CURDIR)/backend:/src -w /src golang:1.27-alpine go
  else
  GO = cd backend && go
  endif
  ```
  Targets (all `.PHONY`):
  - `up`: `docker compose up --build -d --wait`, then curl `localhost:8080/readyz` and `localhost:8081/healthz` with a retry up to 30s.
  - `down`: `docker compose down`.
  - `test`: `$(GO) test ./...`, then `if [ -d frontend ]; then pnpm -C frontend test:run; fi` (`test` is watch mode and would hang).
  - `lint`: `$(GO) vet ./...` and fail if `gofmt -l` prints anything.
  - `verify`: `lint test`, then `docker compose config -q`.
  - `lint-docs`: run `.claude/skills/ste-writing/scripts/ste_check.py` on all tracked `.md` files outside `frontend/` and `shadcn-vue-admin-main/`. Warnings only. Not part of `verify`.

- [ ] **Step 4: Write `.env.example`.** Keys: `ANTHROPIC_API_KEY=` (empty), `LLM_MODE=mock`, with a comment per key.

- [ ] **Step 5: Verify, clean-clone conditions (Review Focus 1, 3, 4).**
  ```bash
  test ! -f .env && make up && curl -s localhost:8080/readyz && docker compose port postgres 5432; echo "port_exit=$?"
  ```
  Expected: `{"status":"ok","db":"ok"}`, nothing printed for the postgres port, and a non-zero `port_exit`.

- [ ] **Step 6: Verify the DB-down path (Review Focus 2).**
  ```bash
  docker compose stop postgres && curl -s -o /dev/null -w "%{http_code}\n" localhost:8080/readyz; docker compose start postgres
  ```
  Expected: `503`.

- [ ] **Step 7: Verify `make test` without Go (Review Focus, D-004).**
  `PATH=/usr/bin:/bin:/usr/local/bin make test` (Homebrew Go at `/opt/homebrew/bin` is now hidden). Expected: tests run in the container and pass.

- [ ] **Step 8: Run `make verify` and `make down`.** Expected: exit 0.

- [ ] **Step 9: Update tracking and commit.** Set `00-skeleton.passes = true` in `tasks/backlog.json` only after user approval. Append a `progress.md` entry.
  ```bash
  git add backend/Dockerfile backend/.dockerignore docker-compose.yml Makefile .env.example tasks
  git commit -m "chore: docker compose, Makefile, env example"
  ```

---

## Out of scope (owned by later slices)

Frontend move and trim (01), seed users and DB schema/migrations (02), mock provider send/429/latency behavior (03), CI pipeline (bonus, after 08), metrics endpoint (08).
