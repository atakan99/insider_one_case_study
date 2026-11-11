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

## 2026-11-09 — 03-mock-provider
- Done: spec `docs/specs/03-mock-provider/`, plan `docs/plans/03-mock-provider.md` and Tasks 1 to 4. Added `config.LoadMockProvider` and the `MOCK_*` env vars in compose. Added the package `internal/mockprovider`. It has the send endpoints, a token bucket for each channel, random and forced failures, an in-memory store, `/v1/messages` and `/v1/stats`. User decisions: the provider does no dedup, as Twilio and SendGrid (D-018). After a timeout or a dropped connection, the worker marks `unknown` and does not retry (D-019). Added D-020 and D-021. The final review found 2 Important defects, and both are fixed: `NaN` passed the rate check, and IDs were used again after a reset. I re-graded 2 Minor findings to Important and fixed them: `env_file` gave the API key to `mock-provider`, and NFR-1 had no test with failures.
- Verified with: `make verify` exit 0. `go test -race -count=5 ./internal/mockprovider/` passes. A compose run with `-p cc-verify` gave `202`, `503`, `429` with `Retry-After: 1`, a lost response (curl exit 52) and matching stats.
- Next: user review of slice 03. The user sets `03-mock-provider.passes`. Then slice 04-segment-rules.
- Open questions: four Minor items stay open. The decoder accepts data after the JSON object. The 10 s shutdown timeout is too short for `MOCK_SLOW_DELAY` over 10 s. Two tests can be stronger (slow disconnect, bucket after a `429`). Slice 07 must not send `Idempotency-Key` or `X-Idempotency-Key`.

## 2026-11-10 — 04-segment-rules
- Done: Task 8 of `docs/plans/04-segment-rules.md`. Compose passes `ANTHROPIC_MODEL` and `CORS_ORIGINS` to the `api`. `.env.example` lists both. The `seed` service lost its `env_file` and reads `SEED_*` through Compose. Thus only the `api` gets `ANTHROPIC_API_KEY`. `docs/architecture.md` Flow step 1 describes the segment flow (D-022 to D-028).
- Verified with: a clean `docker compose -p cc-verify up -d --build --wait` with no `.env` and `LLM_MODE=mock`. The preview request gave `200`, `Access-Control-Allow-Origin: http://localhost:3000`, email 148, SMS 101, and one language `tr`. A headless Chrome run on `/campaigns/new` showed the same counts in the panel. `make verify` exit 0. The NFR-6 import check prints nothing for `internal/segment` and only `campaigncomposer/internal/config` for `internal/llm`.
- Next: user review of slice 04. The user sets `04-segment-rules.passes`. Then slice 05.
- Open questions: no live call with a real key tested the strict `anyOf` tool schema and `fallbacks: "default"` on `claude-opus-5-5`. On an old volume, "last 30 days" segments can be empty. `SEED_RESET=true` fixes this (D-026).

## 2026-11-10 — 05-content-generation
- Done: slice 05 content generation. The api has `POST /api/content/generate`, an SSE stream with one Claude call for each language (D-030 to D-035). `internal/content` validates the limits after rendering with the longest first name. The frontend shows live draft cards on `/campaigns/new`. `docs/architecture.md` Flow step 2 describes the content flow.
- Verified with: a clean `docker compose -p cc-verify up -d --build --wait` with no `.env` and `LLM_MODE=mock`. `POST /api/content/generate` with the example brief gave `audience`, 8 `draft`, `final` for `tr` (`sms_runes` 76) and `done`. A headless Chrome run on `/campaigns/new` showed the `tr` card with counters 76/160 and 44/80. `make verify` exit 0. The NFR-3 import check prints only `campaigncomposer/internal/config`.
- Next: user review of slice 05. The user sets `05-content-generation.passes`. Then slice 06-review-approve.
- Open questions:
  - No live call with a real key tested eager streaming together with `strict` on `submit_content`.
  - The Minor items in `.superpowers/sdd/05-content-generation/progress.md` go to the final review.

## 2026-11-10 — 06-review-approve
- Done: spec `docs/specs/06-review-approve/` (with the mockup), plan `docs/plans/06-review-approve.md` and Tasks 1 to 9, run inline. User decisions: one review window with SMS and Email tabs, a second confirm before the send, a token-based design that works on a phone, failed languages are skipped, and Claude saves a draft when it finishes. Added migration `00002_campaigns`, `segment.SampleFirstNames`, `internal/campaign` (store, check, service), five campaign endpoints, the draft save in the content stream, the review window, `/campaigns` and `/campaigns/{id}`. Added D-036 to D-045. The final review found 1 Critical and 4 Important defects. All 5 are fixed, each with a test that failed first: a stale draft text after "Regenerate content", the row lock that waited for the audience query, a late response for another campaign, old limits after a 422, and the review button after a failed regenerate.
- Verified with: `make verify` exit 0 (Go tests with Postgres, 146 frontend tests, lint, build). A compose run with `-p cc-verify` on ports 13000 and 18080 (the running stack holds 3000 and 8080) gave the draft, approve `200`, repeat `200`, PUT `409` and CORS with `PUT`. Headless Chrome at 1280 px and 360 px completed the flow to "Approved" with no horizontal scroll.
- Next: user review of slice 06. The user sets `06-review-approve.passes`. Then slice 07-send-queue.
- Open questions:
  - Five Minor review items stay open. See the final message of the session.
  - Slice 07 must widen the `campaigns.status` CHECK and keep `approved_at` set for the later statuses.

## 2026-11-10 — 07-send-queue
- Done: slice 07 send queue. Added a migration for the new campaign statuses and `send_jobs`. Added `internal/delivery` (planner, Redis streams, limiter, sender, retry, recovery) and `cmd/worker`. Added four endpoints: progress, progress stream, send-jobs and retry. Added the sending popup, the delivery section and the status badges. Added D-046 to D-065 and the architecture Flow step 4.
- Verified with: `make verify` exit 0. A compose run with `-p cc-verify` on ports 13000, 18080, 18081 and 18082 (the running stack holds 3000, 8080 and 8081). The worker `/healthz` gave `{"db":"ok","redis":"ok","status":"ok"}`. The example brief gave campaign 1 with email 148 and SMS 101. The campaign went to `completed`. The final progress was SMS 95 sent and 6 unknown, email 142 sent and 6 unknown. `GET /v1/stats` gave `duplicate_deliveries: 0` for both channels, and `rate_limited: 0` for 257 requests. The retry gave `409 nothing_to_retry`. Headless Chrome at 1280 px and 360 px completed the flow to Done and showed the list and the retry button. Both widths had no horizontal scroll. The retry button run used three send jobs that I set to `failed` in the temporary database. Thus that run shows duplicate deliveries in its stats, and the retry itself ran from the popup.
- Next: user review of slice 07. The user sets `07-send-queue.passes`. Then slice 08-dashboard-metrics.
- Open questions:
  - There is one worker replica. A shared rate limiter is a non-goal.
  - `unknown` send jobs are never retried.
  - The progress stream has no SSE keepalive comment.

## 2026-11-10 — 08-dashboard-metrics
- Done: slice 08 dashboard and metrics. Added the dashboard store and the health checker. Added the two dashboard endpoints, the api `/metrics`, the worker `/metrics` and the Redis check in `/readyz`. Added the `/dashboard` page. `/` goes to `/dashboard`. Added D-068 to D-082 and the architecture Flow step 5. The backlog item has the spec and the acceptance lines.
- Verified with: a compose run with `-p cc-verify` on ports 13000, 18080, 18081 and 18082, with `LLM_MODE=mock`. The example brief gave campaign 1 with SMS 96 sent and 5 unknown, email 140 sent and 8 unknown. `GET /api/dashboard/stats` matched `GET /api/campaigns/1/progress`. The highest throughput value was 21.2 per second. `GET /api/dashboard/health` gave four `ok` and pending 0. `docker compose stop mock-provider` made `provider` `down` at once, and `start` made it `ok` in 1 s. Both `/metrics` returned the metrics of `goal.md`. Headless Chrome at 1280 px showed the five parts in the light theme and the dark theme with no horizontal scroll.
- Fix wave after the final review: the grid is `grid grid-cols-1 gap-6 lg:grid-cols-2` with `min-w-0` on both cards. At 360 px, `document.documentElement.scrollWidth` is 360 in the light theme and the dark theme. NFR-3 holds. The x-axis labels of the chart no longer overlap. The page sets `stats` and `health` independently, so a slow `health` call does not delay the statistics. The api Redis client and a second worker client for the stream collector honor context deadlines (D-083). A new test covers the labels of the api `/metrics` (FR-19).
- NFR-4: 50,000 send jobs in 5 campaigns, 4,276 with `sent_at` in the last 10 minutes. Five calls of `GET /api/dashboard/stats` took 25, 19, 19, 18 and 19 ms. The limit is 300 ms.
- Next: user review of slice 08. The user sets `08-dashboard-metrics.passes`.
- Open questions:
  - `XLEN` of the streams does not fall after a send, because the worker does not trim the streams. This is not a defect of slice 08.

## 2026-11-11 — external review fixes
- Done: three findings of an external review. All three were correct.
  - The worker now waits for the rate token immediately before the provider request. The order is claim, load, wait, send (D-084, replaces D-053). `Store.Claimable` is removed.
  - A save in the review window keeps the text that the marketer typed during the request (D-085).
  - `README.md` describes the limits of the mock LLM and the steps to enable Claude.
- Verified with: each defect had a test that failed first. `TestConsumer_PacesRequestsAfterDatabaseStall` showed 7 requests in 81 µs at 50 per second before the fix. It passed 10 of 10 runs after the fix. The review window test "keeps the text typed while a save is in progress" failed before the fix. `go test -race ./internal/delivery/...` passes. `make verify` exit 0 (222 frontend tests).
- Next: user review of the changes. Nothing is committed.
- Open questions:
  - A `Retry-After` longer than `SEND_TIMEOUT` plus 30 s can set a waiting send job to `unknown`. The mock provider sends 1 s.
  - No compose run measured the request rate against the mock provider after the change.

## 2026-11-11 — second external review
- Done: two findings of a second review. Both were correct, and both came from D-084.
  - The worker order is again wait, claim, load, send. A rate token is good for 100 ms. If the claim and the load take longer, the send job goes back to `queued` and takes a new token (D-086, replaces D-084).
  - `Store.Claimable` is back.
- Verified with: `TestConsumer_ConsentRevokedDuringWait` and `TestConsumer_WaitingJobIsNotExpired` failed first (one request sent, and a job in `unknown`). They pass now. The pacing test of the first review also passes. These three tests passed 15 of 15 runs with `-race`. A compose run with `-p cc-verify` sent 1,803 requests: `rate_limited: 0`, `duplicate_deliveries: 0`, no token-age warning, and 98 sends at most in one second for each channel. `make verify` exit 0.
- Next: user review of the changes. Nothing is committed.
- Open questions:
  - Sends whose tokens are less than 100 ms apart can leave together. This is 10 requests at most at 100 per second.

## 2026-11-11 — UI and UX pass
- Done: a frontend-only pass after a design review. No backlog item and no API change.
  - The campaign page has a back link and a line with the created date and the totals. It has a delivery summary with a stacked bar, and language tabs. The SMS shows next to the email (D-087, D-088).
  - `{{first_name}}` shows as a chip. The campaign page shows the SMS length and the subject length.
  - The compose page uses the same content view. The suggested briefs are short labels. The cards show a skeleton before the first draft.
  - Counts, dates, language names and the channel order use one format on all pages (D-089).
  - The campaign list has a "New campaign" button. Its rows answer the Space key.
  - The page content has a `max-w-6xl` limit.
  - The review window has one height on a desktop: 720 px, or the screen height minus 32 px if that is less. A change of tab does not resize it.
- Verified with: `make verify` exit 0 (234 frontend tests, 12 new). Chrome through the debugging protocol showed `/campaigns/5`, `/campaigns`, `/campaigns/new` and `/dashboard` at 1440 px, and the first three at 390 px. `document.documentElement.scrollWidth` was 390 at 390 px. The compose page run used stub responses in the browser, because the running api has `LLM_MODE=anthropic`.
- Review window check: the height was the same on the SMS tab and the Email tab at 1440, 1024, 800 and 390 px.
- Next: user review of the changes. Nothing is committed.
- Open questions:
  - A click on a draft row in the list opens the review window, and other rows open the campaign page.
  - The list has no recipient count, because the API does not return one.
  - The dashboard and the list still show "Loading…" text or nothing during the first load.

## 2026-11-11 — state machines
- Done: `docs/state-machines.md` lists 10 state machines in the code, and the gaps that we found. We fixed five gaps and kept one on purpose.
  - The trigger `send_jobs_status_moves` (migration `00005`) allows only the moves of the worker (D-091). The test fixtures use `setJobs` and move each job through `sending`.
  - A skip result that cannot be written goes back to `queued`, not to `unknown` (D-092).
  - The `retry` event of the content stream has `reason`: `restart` or `invalid`. The card shows "Starting again…" or "Fixing content…" (D-093).
  - `campaign.Status` has constants for all five values. `delivery` uses them. The frontend has `SendableStatus`, and `currentStep` has no `default` case (D-094).
  - The campaign trigger still allows `queued → completed` as a fallback. No code path makes this move (D-091).
- Verified with: each fix had a test that failed first. `TestConsumer_SkipWriteFailsReleases` showed `sending` before the fix. `TestMigrate_SendJobStatusMoves` showed 12 moves that the database accepted. The backend and frontend tests of the retry reason failed first. `vue-tsc -b` passes. `make verify` exit 0 (235 frontend tests).
- Next: user review of the changes. Nothing is committed.
- Open questions:
  - The insert of a send job has no status guard. Only `PlanNext` inserts.
  - No compose run checked the new migration against a database with existing data. The trigger runs only on an update, so existing rows are not checked.
