# Progress

Append one entry per session. Use the template below. Put the newest entry at the bottom.

<!-- Template:
## YYYY-MM-DD — <slice id>
- Done:
- Verified with:
- Next:
- Open questions:
-->

## 2026-11-06 — 00-skeleton
- Done: Task 1 of `docs/plans/00-skeleton.md`. Moved the case brief to `docs/case.md` and the frontend-trim spec to `docs/specs/01-frontend-trim/`. Wrote `.gitignore`, `CLAUDE.md`, `DECISIONS.md` (D-001 to D-006), `README.md`, `docs/architecture.md`, `tasks/backlog.json`, `tasks/progress.md` and `ai-sessions/.gitkeep`.
- Verified with: `git check-ignore` (prints `.env`, `.env.local`, then `exit=1`). `python3 -m json.tool tasks/backlog.json` (prints `ok`).
- Next: Task 2 (ste-writing skill, executor and reviewer subagents).
- Open questions: None.

## 2026-11-06 — 00-skeleton
- Done: Task 2 of `docs/plans/00-skeleton.md`. Added the `ste-writing` skill (`SKILL.md`, `references/rules.md`, `references/terms.md`, `scripts/ste_check.py` and its tests). Added the `executor` and `reviewer` subagents. Changed the Writing style line in `CLAUDE.md`. Deleted the template git hooks in `.git/hooks/` (user decision). The reviewer found 6 defects in `ste_check.py`. All 6 are fixed, and each one has a regression test.
- Verified with: `python3 -B -m unittest discover -s .claude/skills/ste-writing/scripts` (11 tests, OK). The session lists `executor`, `reviewer` and `ste-writing`.
- Next: Task 3 (Go module, config, health endpoints, two binaries).
- Open questions: `.gitignore` does not include `__pycache__/`. The tests write it when they run without `-B`.

## 2026-11-06 — 00-skeleton
- Done: Task 3 of `docs/plans/00-skeleton.md`. Added the Go module `campaigncomposer` (go 1.27, `pgx/v5` v5.12.0), `internal/config`, `internal/httpapi` (health endpoints), `cmd/api` and `cmd/mockprovider`. The reviewer found 3 defects in the two `main.go` files. All 3 are fixed: a 2 s timeout on the startup DB ping, no exit 0 when a listen error races a signal, and a second Ctrl-C that stops the process at once.
- Verified with: `go vet ./... && go build ./... && test -z "$(gofmt -l .)"` (prints `ok`). `go test -count=1 ./...` (config and httpapi pass). A manual run with an unreachable DB host (`10.255.255.1`): `/healthz` gave 200 and `/readyz` gave 503 within 3 s, and SIGTERM shut down cleanly.
- Next: Task 4 (Dockerfile, docker-compose, Makefile, `.env.example`).
- Open questions: `.gitignore` does not include `__pycache__/` (from Task 2).

## 2026-11-06 — 00-skeleton
- Done: Task 4 of `docs/plans/00-skeleton.md`. Added the root `Dockerfile` (target `go-service`), the root `.dockerignore`, `docker-compose.yml` (postgres, redis, api, mock-provider), `Makefile` and `.env.example`. User decisions: one root Dockerfile (D-008), Redis Streams as the queue (D-007, D-002 changed), and the frontend service waits for slice 01. The plan was updated for these decisions and for a new Step 5 check, because Compose v5 prints `:0` for `docker compose port` on a port that is not published. The reviewer found 7 defects. All are fixed: `command -v` under Make 3.81, a gofmt parse error that `lint` ignored, `pg_isready` on the socket during first init, an unquoted volume path, an ARG that split the module cache, the Step 5 check, and the plan snippet.
- Verified with: Step 5 from empty volumes (`{"status":"ok","db":"ok"}`, redis `healthy`, `postgres {}`, `redis {}`). Step 6 (`503`). Step 7 without Go (tests pass in the container). `make verify` exit 0, `make down` exit 0. `make lint` exits 2 on a gofmt parse error. A build without `CMD` fails with a clear message.
- Next: user review of slice 00. The user sets `00-skeleton.passes`. Then slice 01-frontend-trim.
- Open questions: `.gitignore` does not include `__pycache__/` (from Task 2). The api does not read a `REDIS_URL` yet. The send-queue slice adds it.
