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

## 2026-11-06 — 01-frontend-trim
- Done: Task 1 of `docs/plans/01-frontend-trim.md`. Copied the template to `frontend/` and removed the rows of the "What we remove" table. Fixed `package.json`, `pnpm-workspace.yaml` and `.env.example` before the first install. Added the three campaign placeholder pages and the `/` redirect. New tests cover the routes, the sidebar and the page search. The reviewer found 4 defects. All are fixed: stale globals in the generated `auto-import.d.ts`, a 401 that showed no toast, and two weak sidebar assertions.
- Verified with: `pnpm test:run` (55 tests), `pnpm build`, `CI=true pnpm lint`. The sidebar test fails on three code changes made on purpose. `.git/hooks` has no hook files after `pnpm install`.
- Next: Task 2 (unused UI components, packages and files). The template `frontend/README.md` must be replaced, because it names removed code (NFR-5).
- Open questions: `__pycache__/` in `.gitignore` (from slice 00).

## 2026-11-06 — 01-frontend-trim
- Done: Task 2 of `docs/plans/01-frontend-trim.md`. Deleted 29 UI components (kept the 34 on the list and `kbd`). Deleted the decision 3 files that no kept file uses. Removed 10 packages. `universal-cookie` stays, because `useCookies` in `default.vue` needs it as a peer. Replaced the template `frontend/README.md` and the `index.html` title (NFR-5). The reviewer found 2 defects. Both are fixed: a coverage glob for the deleted `src/services/`, and theme-editor exports in `constants/themes.ts`.
- Verified with: `pnpm test:run` (55 tests), `pnpm build`, `CI=true pnpm lint`. The NFR-4 import grep has no matches. The NFR-5 text grep matches only the negative assertions in `__tests__`. `pnpm peers check` shows only the two issues that the template also has.
- Next: Task 3 (frontend in Docker and compose, pnpm in the Makefile, delete the template).
- Open questions: the Task 2 Step 4 text grep conflicts with the Task 1 tests. It needs `--exclude-dir=__tests__`. `__pycache__/` in `.gitignore` (from slice 00).

## 2026-11-06 — 01-frontend-trim
- Done: Task 3 of `docs/plans/01-frontend-trim.md`. The root `Dockerfile` has targets `frontend-build` and `frontend` (nginx). Compose serves the frontend on port 3000. The Makefile runs pnpm in a container when pnpm is missing. Added D-009 and D-010. Deleted `shadcn-vue-admin-main/`. User decision: the env validator uses the `.env.example` values as defaults, so `pnpm dev` works without `.env`. The reviewer found 7 defects. Fixed 6: the pnpm store in the host folder, root-owned files on Linux, the nginx redirect port, a double Cache-Control header, missing files that returned the app, and the NFR-3 acceptance text.
- Verified with: headless Chrome checks of FR-1 to FR-8 on `pnpm dev` (no `.env`) and on the compose build. `curl` checks of the nginx headers and status codes. A build-arg override reaches the built JS. `PATH=/usr/bin:/bin:/usr/local/bin make lint test verify` exit 0, with no `.pnpm-store` left. `make verify` and `make down` exit 0.
- Next: user review of slice 01. The user sets `01-frontend-trim.passes`. Then slice 02-seed-users (no spec yet).
- Open questions: `CLAUDE.md` is out of date (reviewer defect 6). I did not change it, because the request came from a subagent. `__pycache__/` in `.gitignore`. The Task 2 Step 4 grep needs `--exclude-dir=__tests__`.

## 2026-11-09 — 02-seed-users
- Done: spec `docs/specs/02-seed-users/`, plan `docs/plans/02-seed-users.md` and Tasks 1 to 4. Added `internal/store` (goose migrations, `Migrate`), `internal/store/storetest`, `internal/seed` (`Generate`, `Load`), `cmd/seed` and the one-shot `seed` service in compose. `make test` starts a temporary test Postgres (D-017). User decisions: `gofakeit` instead of Mockaroo (D-012), `language` and `country` columns with one language for each campaign (D-013), amounts in TRY, goose (D-015). The reviews found 2 Important defects, and both are fixed: gofakeit seed 0 was not repeatable, and the example-brief SQL test used a wrong rule. The final review asked for 6 fixes and 2 STE fixes. All are done.
- Verified with: `make verify` exit 0. The DB tests ran against Postgres and did not skip. `make test` also passes with no local Go. A clean start with `docker compose -p cc-verify up --wait` gave exit 0, 5006 users, 1406 in Istanbul, `seed` exited 0 and `api` running. A mutation `Exp: 0` makes `TestLoad_AmountScale` fail.
- Next: user review of slice 02. The user sets `02-seed-users.passes`. Then slice 03-mock-provider.
- Open questions: four Minor items stay open. `storetest` replaces the URL path only. The constraint tests check only `err != nil`. A Feb 29 seed time can give age 17. `Load` takes no lock. On an old volume, "last 30 days" segments become empty. `SEED_RESET=true` fixes this, and the README slice must document it.
