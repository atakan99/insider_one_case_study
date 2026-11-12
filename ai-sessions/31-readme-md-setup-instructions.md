# README.md setup instructions

- Session: `6617584f-4c57-4791-a6d2-d055e3347d10`
- Model: `claude-fable-5-2`
- Time (UTC): 2026-11-12 06:56:50 → 2026-11-12 06:57:55
- Subagent runs: 0

---

## 👤 User  <sub>2026-11-12 06:56:50</sub>

renew the readme.md i open app with make up ritght?

**🔧 Skill** ```json
{
  "skill": "ste-writing"
}
```

<details><summary>Result</summary>

```
Launching skill: ste-writing
```

</details>

> _[skill instructions loaded: ste-writing]_

**🔧 Bash** _Read the current README, Makefile, compose file, env example and project layout_
```bash
ls -la && echo ---README && cat README.md 2>/dev/null; echo ---MAKE && cat Makefile; echo ---COMPOSE && cat docker-compose.y*ml compose.y*ml 2>/dev/null; echo ---ENV && cat .env.example 2>/dev/null; echo ---PROGRESS && tail -60 tasks/progress.md; echo ---LS && ls backend backend/cmd backend/internal frontend docs docs/specs docs/plans tasks ai-sessions 2>/dev/null
```

<details><summary>Result</summary>

````
total 112
drwxr-xr-x@ 23 USER  staff   736 Nov 11 14:43 .
drwxr-xr-x  12 USER  staff   384 Nov  6 13:23 ..
-rw-r--r--@  1 USER  staff  6148 Nov 11 13:45 .DS_Store
drwxr-xr-x   6 USER  staff   192 Nov  6 14:38 .claude
-rw-r--r--@  1 USER  staff   276 Nov  6 15:26 .dockerignore
-rw-r--r--   1 USER  staff  1828 Nov 11 14:51 .env
-rw-r--r--   1 USER  staff  1717 Nov 10 22:52 .env.example
drwxr-xr-x  15 USER  staff   480 Nov 12 09:56 .git
-rw-r--r--@  1 USER  staff   297 Nov  9 17:46 .gitignore
drwxr-xr-x@  4 USER  staff   128 Nov 10 14:26 .superpowers
drwxr-xr-x   3 USER  staff    96 Nov  6 13:23 .vscode
-rw-r--r--   1 USER  staff  1998 Nov  6 15:41 CLAUDE.md
-rw-r--r--   1 USER  staff  2498 Nov 11 15:12 DECISIONS.md
-rw-r--r--   1 USER  staff  1554 Nov 10 19:52 Dockerfile
-rw-r--r--   1 USER  staff  3831 Nov 10 19:52 Makefile
-rw-r--r--@  1 USER  staff  1744 Nov 11 12:43 README.md
drwxr-xr-x  30 USER  staff   960 Nov 11 19:25 ai-sessions
drwxr-xr-x   6 USER  staff   192 Nov 10 22:52 backend
-rw-r--r--   1 USER  staff   140 Nov  9 17:46 docker-compose.override.yml
-rw-r--r--   1 USER  staff  4660 Nov 10 22:52 docker-compose.yml
drwxr-xr-x   9 USER  staff   288 Nov 11 19:50 docs
drwxr-xr-x  22 USER  staff   704 Nov 12 09:56 frontend
drwxr-xr-x   4 USER  staff   128 Nov 11 19:50 tasks
---README
# Smart Campaign Composer

A marketer writes a brief for a campaign. Claude makes a segment rule and SMS and email content from the brief. The backend validates the rule and counts the audience. The marketer edits and approves the content. Then a queue sends the messages to a mock provider, and a dashboard shows the delivery status. No real person gets a message.

## Run

```bash
docker compose up
```

- Frontend: http://localhost:3000
- API: http://localhost:8080 (`/healthz`, `/readyz`)
- Mock provider: http://localhost:8081

For frontend work with hot reload, run `pnpm dev` in `frontend/` 
… [16092 more chars truncated]
````

</details>

**🔧 Bash** _Read compose services, term list, architecture summary, backlog status and routes_
```bash
cat docker-compose.yml docker-compose.override.yml; echo ---TERMS; cat .claude/skills/ste-writing/references/terms.md; echo ---ARCH; head -80 docs/architecture.md; echo ---BACKLOG; python3 -c "
import json;d=json.load(open('tasks/backlog.json'))
items=d if isinstance(d,list) else d.get('slices',d.get('items',d))
for i in items: print(i.get('id'),i.get('passes'),'|',i.get('title',i.get('description',''))[:90])"; echo ---ROUTES; grep -rn "path:" frontend/src/router/*.ts 2>/dev/null | head -20; cat ai-sessions/README.md | head -30; grep -n "LLM_MODE\|ANTHROPIC_MODEL" .env | sed 's/=.*KEY.*//'
```

<details><summary>Result</summary>

```
services:
  postgres:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: campaign
      POSTGRES_PASSWORD: campaign
      POSTGRES_DB: campaign
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      # Check TCP, not the Unix socket. On first start the init server
      # listens only on the socket, and the api connects over TCP.
      test: ["CMD-SHELL", "pg_isready -h 127.0.0.1 -U campaign -d campaign"]
      interval: 2s
      retries: 15
    # No ports: Postgres is reachable only inside the compose network.
    # This avoids a conflict with a Postgres on the host's port 5432.

  redis:
    image: redis:8-alpine
    # The send queue lives here (D-007). AOF keeps queued jobs across restarts.
    command: ["redis-server", "--appendonly", "yes"]
    volumes:
      - redisdata:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 2s
      retries: 15
    # No ports, for the same reason as Postgres (host port 6379).

  api:
    build:
      context: .
      target: go-service
      args:
        CMD: api
    env_file:
      - path: .env
        required: false
    environment:
      DATABASE_URL: postgres://campaign:campaign@postgres:5432/campaign?sslmode=disable
      REDIS_URL: redis://redis:6379/0
      WORKER_URL: http://worker:8082
      PROVIDER_URL: http://mock-provider:8081
      LLM_MODE: ${LLM_MODE:-mock}
      ANTHROPIC_MODEL: ${ANTHROPIC_MODEL:-claude-opus-5-5}
      # Origins that may call the api from a browser (D-028).
      CORS_ORIGINS: ${CORS_ORIGINS:-http://localhost:3000,http://localhost:5173,http://127.0.0.1:3000,http://127.0.0.1:5173}
    ports:
      - "8080:8080"
    depends_on:
      redis:
        condition: service_healthy
      postgres:
        condition: service_healthy
      # The api starts after the seed applies the migrations (D-015).
      seed:
        condition: service_completed_successfully

  # One-shot job. It applies the migrations, loads the users, then exits.
  #
… [15588 more chars truncated]
```

</details>

**🔧 Bash** _Find frontend routes, model setting usage and dev server settings_
```bash
grep -rnE "path: *['\"]" frontend/src --include=*.ts | grep -v node_modules | head -20; ls frontend/src/pages 2>/dev/null; grep -rn "ANTHROPIC_MODEL\|AnthropicModel" backend --include=*.go | grep -v _test | head; grep -n "pnpm dev\|\"dev\"\|VITE_" frontend/package.json frontend/vite.config.ts frontend/.env* 2>/dev/null | head; grep -n "prerequisite\|Docker" docs/case.md | head
```

<details><summary>Result</summary>

```
(eval):1: no matches found: --include=*.ts
[...path].vue
__tests__
campaigns
dashboard.vue
errors
errors.vue
index.vue
(eval):1: no matches found: --include=*.go
frontend/package.json:14:    "dev": "vite",
frontend/.env.example:1:VITE_SERVER_API_URL=http://localhost:8080
frontend/.env.example:2:VITE_SERVER_API_PREFIX=/api
frontend/.env.example:3:VITE_SERVER_API_TIMEOUT=10000
```

</details>

**🔧 Bash** _List campaign pages and check where the model setting applies_
```bash
ls frontend/src/pages/campaigns; grep -rn "ANTHROPIC_MODEL" backend | grep -v _test | head -5; grep -rn "Model" backend/cmd/api/*.go | head -8
```

<details><summary>Result</summary>

```
[id].vue
__tests__
index.vue
new.vue
backend/internal/config/config.go:28:	// AnthropicModel is ANTHROPIC_MODEL. The default is claude-opus-5-5.
backend/internal/config/config.go:47:		AnthropicModel:  getenv("ANTHROPIC_MODEL"),
backend/cmd/api/main.go:91:	slog.Info("llm configured", "model", cfg.AnthropicModel, "llm_mode", cfg.LLMMode)
```

</details>

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/README.md`
````
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

##
… [2662 more chars truncated]
````

**🔧 Bash** _Run the STE checker on the new README_
```bash
python3 -B .claude/skills/ste-writing/scripts/ste_check.py README.md; grep -n "state\|Suggest\|suggest" frontend/src/pages/campaigns/new.vue | head -5; grep -c "" docs/state-machines.md
```

<details><summary>Result</summary>

```
The file ~/Desktop/projects/case-studies/insider_one/README.md has been updated successfully. (file state is current in your context — no need to Read it back)
```

</details>

<details><summary>Result</summary>

```
283
```

</details>

**🔧 Edit** `~/Desktop/projects/case-studies/insider_one/README.md`
**old:**
```
the states of a campaign and of a send job.
```
**new:**
```
the state machines in the code.
```

<details><summary>Result</summary>

```
The file ~/Desktop/projects/case-studies/insider_one/README.md has been updated successfully. (file state is current in your context — no need to Read it back)
```

</details>

## 🤖 Claude  <sub>2026-11-12 06:57:55</sub>

Written for: a reviewer of the case study who clones the repo and has only Docker.

Yes, `make up` starts the app. It builds the images, starts all services in the background and waits for the health checks. Then you open http://localhost:3000. `make down` stops the services.

I rewrote [README.md](README.md). The STE checker gives no findings. I did not run `make up`, and nothing is committed.

Changes from the old README:

- **Run** uses `make up` as the main command, not `docker compose up`. It lists the prerequisites, the four service URLs (the worker on 8082 was missing) and how to stop and reset.
- **Use the app** is new. It gives the six steps from the brief to the dashboard, and the four pages.
- **Mock mode and Claude** keeps the same text. Step 4 now says `make up`.
- **Configuration** is new. It points to [.env.example](.env.example) and tells the reader that the mock provider fails on purpose, so `failed` and `unknown` send jobs are normal.
- **Safety rules** is new. It states the four rules: no LLM output as SQL, consent for each channel, the length limits, and one send at most for each send job.
- **Develop**, **Layout** and **Documents** are new. They cover `make test`, `make lint`, `make verify`, the hot reload steps and links to the architecture, the state machines and the decisions.
- **Status** says that all nine slices pass.

Your local `.env` has `LLM_MODE=anthropic` and `ANTHROPIC_MODEL=claude-sonnet-5-5`. Thus your `make up` calls the real Claude API with Sonnet, not the mock. A reviewer without `.env` gets the mock.

