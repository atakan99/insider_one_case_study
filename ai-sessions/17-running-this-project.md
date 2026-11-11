# Running this project

- Session: `80995dcf-9d9c-4056-8151-9f943a23c7aa`
- Model: `claude-opus-5-5`
- Time (UTC): 2026-11-10 11:18:18 → 2026-11-10 11:26:21
- Subagent runs: 0

---

## 👤 User  <sub>2026-11-10 11:18:18</sub>

_[IDE: opened file]_ 
how can i run this project

**🔧 Bash** _Inspect Makefile, env and compose files_
```bash
ls -a && cat Makefile && ls README* .env* 2>/dev/null; cat .env.example 2>/dev/null; grep -nE "ports:|- \"[0-9]+:|image:|^  [a-z-]+:$" docker-compose.y*ml compose.y*ml 2>/dev/null
```

<details><summary>Result</summary>

```
.
..
.claude
.dockerignore
.env.example
.git
.gitignore
.superpowers
.vscode
CLAUDE.md
DECISIONS.md
Dockerfile
Makefile
README.md
ai-sessions
backend
docker-compose.override.yml
docker-compose.yml
docs
frontend
tasks
# Use local Go when it is installed. Otherwise run Go in a container (D-004).
# The 2>/dev/null makes Make run this through /bin/sh, so the `command` builtin works.
ifeq ($(shell command -v go 2>/dev/null),)
DOCKER_GO = docker run --rm -v "$(CURDIR)/backend:/src" -w /src golang:1.27-alpine
GO = $(DOCKER_GO) go
GOFMT = $(DOCKER_GO) gofmt
# Tests join the cc-test network to reach the test Postgres by name.
GO_TEST = docker run --rm --network cc-test -e TEST_DATABASE_URL -v "$(CURDIR)/backend:/src" -w /src golang:1.27-alpine go
TEST_DATABASE_URL = postgres://test:test@cc-test-pg:5432/test?sslmode=disable
else
GO = cd backend && go
GOFMT = cd backend && gofmt
GO_TEST = cd backend && go
TEST_DATABASE_URL = postgres://test:test@127.0.0.1:55432/test?sslmode=disable
endif

# Same for pnpm. The anonymous /app/node_modules volume keeps Linux packages
# out of the host's frontend/node_modules. The pnpm store stays in the
# container (/tmp). The container runs as root, so at the end it gives the
# files that it wrote (dist/, src/types/) back to the host user.
ifeq ($(shell command -v pnpm 2>/dev/null),)
PNPM = docker run --rm -v "$(CURDIR)/frontend:/app" -v /app/node_modules -e CI=true \
	-e HOST_ID=$(shell id -u):$(shell id -g) -w /app node:24-alpine \
	sh -c 'corepack enable && pnpm install --frozen-lockfile --store-dir /tmp/pnpm-store && pnpm "$$@"; rc=$$?; \
	chown -R "$$HOST_ID" dist src/types 2>/dev/null; exit $$rc' pnpm
else
PNPM = pnpm -C frontend
endif

.PHONY: up down test lint verify lint-docs

up:
	docker compose up --build -d --wait
	curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 localhost:8080/readyz
	@echo
	curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 localhost:8081/healthz
	@echo
	curl -fsS
… [2651 more chars truncated]
```

</details>

**🔧 Bash** _Read README and compose services_
```bash
cat README.md | head -80; grep -nE "^  [a-z-]+:$|ports|[0-9]{4}:" docker-compose.yml docker-compose.override.yml
```

<details><summary>Result</summary>

````
# Smart Campaign Composer

A marketer writes a brief for a campaign. Claude makes a segment rule and SMS and email content from the brief. The backend validates the rule and counts the audience. The marketer edits and approves the content. Then a queue sends the messages to a mock provider, and a dashboard shows the delivery status. No real person gets a message.

## Run

```bash
docker compose up
```

- Frontend: http://localhost:3000
- API: http://localhost:8080 (`/healthz`, `/readyz`)
- Mock provider: http://localhost:8081

For frontend work with hot reload, run `pnpm dev` in `frontend/` (http://localhost:5173).

## Test

```bash
make test
```

## Status

See [`tasks/backlog.json`](tasks/backlog.json).
docker-compose.override.yml:3:  postgres:
docker-compose.override.yml:4:    ports:
docker-compose.override.yml:5:      - "127.0.0.1:5433:5432"
docker-compose.yml:2:  postgres:
docker-compose.yml:16:    # No ports: Postgres is reachable only inside the compose network.
docker-compose.yml:19:  redis:
docker-compose.yml:29:    # No ports, for the same reason as Postgres (host port 6379).
docker-compose.yml:31:  api:
docker-compose.yml:46:    ports:
docker-compose.yml:47:      - "8080:8080"
docker-compose.yml:58:  seed:
docker-compose.yml:78:  mock-provider:
docker-compose.yml:97:    ports:
docker-compose.yml:98:      - "8081:8081"
docker-compose.yml:100:  frontend:
docker-compose.yml:104:    ports:
docker-compose.yml:105:      - "3000:8080"
docker-compose.yml:110:  pgdata:
docker-compose.yml:111:  redisdata:
````

</details>

## 🤖 Claude  <sub>2026-11-10 11:18:33</sub>

Everything runs in Docker. You don't need Go or Node installed on your machine.

## Start it

```bash
cp .env.example .env   # optional; without .env it runs with the mock LLM
make up                # builds, starts every service, waits for the health checks
```

Then open:
- **App:** http://localhost:3000
- **API:** http://localhost:8080 (health checks at `/healthz` and `/readyz`)
- **Mock provider:** http://localhost:8081
- **Postgres from your machine:** `127.0.0.1:5433` (from [docker-compose.override.yml](docker-compose.override.yml))

To stop it, run `make down`.

## Use the real Claude API
By default it uses the mock LLM, which needs no API key. To use Claude instead, set these in `.env`:
```
LLM_MODE=anthropic
ANTHROPIC_API_KEY=[REDACTED]
```
Then run `make up` again.

## What starts
`postgres`, `redis`, `api`, `seed` (loads 5,000 fake users the first time), `mock-provider` (a fake SMS and email sender that fails at random on purpose) and `frontend`. You can change the seed size and the mock failure rates in `.env`. See the comments in [.env.example](.env.example).

## Other commands
- `make test`: runs the Go and frontend tests. It starts a temporary Postgres for them.
- `make verify`: runs lint, the tests, the frontend build and a compose config check.
- **Frontend with hot reload:** start the stack with `make up`, then run `pnpm dev` in `frontend/` and open http://localhost:5173.

If port 3000, 8080, 8081 or 5433 is already in use, `make up` fails. Stop whatever is using that port and try again.

## 👤 User  <sub>2026-11-10 11:26:16</sub>

curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 localhost:8080/readyz
{"status":"ok","db":"ok"}

curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 localhost:8081/healthz
{"status":"ok"}

curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 -o /dev/null -w "frontend %{http_code}\n" localhost:3000/
frontend 200

## 🤖 Claude  <sub>2026-11-10 11:26:21</sub>

All three checks passed, so the stack is running:

- **API** is ready and connected to Postgres (`"db":"ok"`).
- **Mock provider** is healthy.
- **Frontend** returned HTTP 200.

Open **http://localhost:3000** to use the app. The seed service loads the 5,000 test users on the first start, so audience counts can show 0 for a few seconds. To check that it finished, run:

```bash
docker compose logs seed
```

When you're done, run `make down`. Your data stays in the `pgdata` volume, so it's still there the next time you run `make up`.

