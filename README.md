# Smart Campaign Composer

A marketer writes a brief for a campaign. Claude makes a segment rule and SMS and email content from the brief. The backend validates the rule and counts the audience. The marketer edits and approves the content. Then a queue sends the messages to a mock provider, and a dashboard shows the delivery status. No real person gets a message.

## Run

You need Docker with Compose, `make` and `curl`. You do not need Go, Node or an API key.

```bash
make up
```

`make up` builds the images and starts all services in the background. Then it waits until each service answers its health check. The first start loads 5,000 seed users.

Open http://localhost:3000 when the command ends.

| Service | URL | Notes |
|---|---|---|
| Frontend | http://localhost:3000 | The app. `/` goes to `/dashboard`. |
| API | http://localhost:8080 | `/healthz`, `/readyz`, `/metrics` |
| Mock provider | http://localhost:8081 | `/healthz` |
| Worker | http://localhost:8082 | `/healthz`, `/metrics` |

Postgres and Redis have no host ports.

To stop the services:

```bash
make down
```

`make down` keeps the data. To delete the users, the campaigns and the queue, run `docker compose down -v`.

`docker compose up` also starts the app. It stays in the foreground and does not run the health checks.

## Use the app

1. Open http://localhost:3000/campaigns/new.
2. Write a brief, or select a suggested brief.
3. Read the segment rule and the audience count for each channel and each language.
4. Generate the content. The SMS and email drafts show while the LLM writes them.
5. Open the review window. Edit the content, then approve the campaign.
6. Open the campaign page or `/dashboard` to see the delivery status.

The pages are:

- `/campaigns/new` — write a brief and make a campaign.
- `/campaigns` — the list of campaigns.
- `/campaigns/{id}` — the content and the delivery status of one campaign.
- `/dashboard` — totals, throughput and the health of the services.

## Mock mode and Claude

The default start uses a mock LLM. It needs no API key, and all tests use it.

The mock does not read the brief. For each brief, it returns the same answers:

- The segment rule of the example brief. It selects users in Istanbul who added to the cart in the last 30 days. It excludes users who bought in the last 30 days.
- A fixed cart reminder with a 30% discount, in Turkish or English. Other languages get the English text.

All other steps are real in mock mode: the rule validation, the audience count, the limit checks, the review, the send queue and the dashboard.

To see Claude make the segment rule and the content from your brief:

1. Copy `.env.example` to `.env`.
2. In `.env`, set `LLM_MODE=anthropic`.
3. In `.env`, set `ANTHROPIC_API_KEY` to your key.
4. Run `make up` again.

The default model is `claude-opus-5-5`. To change it, set `ANTHROPIC_MODEL` in `.env`. Only the `api` service gets the key. The provider stays a mock in both modes, so no real person gets a message.

## Configuration

All settings are optional. [`.env.example`](.env.example) lists each setting with its default value.

- `SEED_*` — the number of seed users and the random seed.
- `MOCK_*` — the rate limit and the failure rates of the mock provider.
- `SEND_*` — the send rate, the attempts, the timeout and the consumers of the worker.

The mock provider fails on purpose. By default, 3% of the requests get an error, 3% are slow and 1% get no response. Thus some send jobs end as `failed` or `unknown`. Set the three rates to 0 to turn the failures off.

## Safety rules

- The backend never runs LLM output as SQL. A segment rule is JSON. The backend validates it against an allowlist of fields and operators, then compiles it to a parameterized query.
- A user gets a message on a channel only with consent for that channel. The worker checks consent again immediately before each send.
- An SMS has 160 characters or fewer. An email subject has 80 characters or fewer. The backend checks the limits after it renders `{{first_name}}`.
- Each send job goes to the provider one time at most. After a timeout, the worker does not send again. It sets the delivery status to `unknown`.

## Develop

```bash
make test     # all tests: Go with a temporary Postgres and Redis, then the frontend
make lint     # go vet, gofmt and the frontend lint
make verify   # lint, tests, the frontend build and the compose config check
```

These commands work without Go and pnpm on the host. Then they run in containers.

For frontend work with hot reload:

1. Run `make up`.
2. Run `pnpm install` and then `pnpm dev` in `frontend/`.
3. Open http://localhost:5173.

## Layout

- `backend/` — the Go module. `cmd/` has one binary for each process: `api`, `worker`, `mockprovider` and `seed`. `internal/` has the shared packages.
- `frontend/` — the Vue app.
- `docs/` — the case brief, the specs, the plans and the design documents.
- `tasks/` — the backlog and the session log.
- `ai-sessions/` — the exports of the Claude Code sessions.

## Documents

- [`docs/case.md`](docs/case.md) — the case brief.
- [`docs/architecture.md`](docs/architecture.md) — the components and the flow.
- [`docs/state-machines.md`](docs/state-machines.md) — the state machines in the code.
- [`DECISIONS.md`](DECISIONS.md) — the main decisions.
- [`docs/decisions-log.md`](docs/decisions-log.md) — the full log of decisions.
- [`ai-sessions/README.md`](ai-sessions/README.md) — the index of the session exports.

## Status

All nine slices in [`tasks/backlog.json`](tasks/backlog.json) pass. [`tasks/progress.md`](tasks/progress.md) has the log of each session and the open questions.
