# Smart Campaign Composer

## Layout

- `backend/` — Go module. One binary per process in `cmd/`, shared packages in `internal/`.
- `frontend/` — Vue app. Slice `01-frontend-trim` creates it from `shadcn-vue-admin-main/`.
- `docs/` — case brief (`case.md`), specs (`specs/NN-name/`), plans (`plans/NN-name.md`), `architecture.md`.
- `tasks/` — `backlog.json` (slices and acceptance checks) and `progress.md` (session log).
- `ai-sessions/` — Claude session exports. Remove API keys and personal data before you commit them.

## Commands

- `make up` — build and start all services, then wait for the health checks.
- `make down` — stop all services.
- `make test` — run all tests. Works without Go (uses a `golang` container).
- `make lint` — run `go vet` and `gofmt`.
- `make verify` — run lint, tests and the compose config check.

## Workflow

1. At session start, read `tasks/progress.md`.
2. Pick the first item in `tasks/backlog.json` with `"passes": false`.
3. Set `passes` to `true` only after `make verify` passes and the user approves.
4. At session end, append an entry to `tasks/progress.md`.

## Hard rules

- Never execute LLM output as SQL. Segment rules are JSON. The backend validates them against an allowlist of fields and operators. Then it compiles them to parameterized queries.
- Never send on a channel without that user's consent for that channel.
- SMS must be 160 characters or fewer. Email subject must be 80 characters or fewer. Check the limits after you render `{{first_name}}`-style variables.
- Tests use the mock LLM. No test reads `ANTHROPIC_API_KEY`.
- Write a `DECISIONS.md` entry for each non-obvious choice.

## Writing style

TBD. Task 2 Step 9 of `docs/plans/00-skeleton.md` replaces this line.
