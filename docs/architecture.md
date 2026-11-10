# Architecture

## Components

- **api** — Go HTTP API. Health endpoints, campaigns, segments, content and send status.
- **mock-provider** — Go service that acts as an SMS and email provider, like Twilio or SendGrid. It sometimes returns `429`, `5xx`, slow responses and lost responses. It does no dedup (D-018).
- **seed** — Go one-shot job. It applies the migrations, loads the seed users, then exits (D-015). The api starts after it completes.
- **postgres** — the datastore (D-002).
- **redis** — the send queue, with Redis Streams (D-007).
- **frontend** — Vue app for the campaign flow. Compose serves the built app with nginx on port 3000 (D-010).

## Folder tree

```
backend/
  cmd/
    api/            HTTP API
    worker/         send worker
    mockprovider/   mock SMS and email provider
    seed/           seed users
  internal/
    config/         env loading
    httpapi/        handlers, health, SSE
    segment/        rule DSL, validation, SQL builder
    llm/            interface, Claude client, fake
    content/        channel limits, {{var}} rendering
    delivery/       queue, rate limiter, retry, idempotency
    mockprovider/   mock provider handlers, rate limit, failure modes, message store
    store/          database access, goose runner
      migrations/   embedded SQL migrations
    seed/           seed user generator and loader
frontend/
docs/
  case.md
  architecture.md
  specs/NN-name/
  plans/NN-name.md
tasks/
  backlog.json
  progress.md
ai-sessions/
```

## Flow

1. **Audience** — Claude makes a segment rule from the brief. The UI shows the audience count.
   - `POST /api/segments/preview` takes the brief. The backend removes control characters and limits the length of the brief before it sends the brief to Claude.
   - Claude must answer with the tool `submit_segment`. The tool has `strict: true` and a JSON schema from the allowlist. The client sends `tool_choice: auto` (D-025).
   - The backend validates the answer against the allowlist of fields and operators (D-029). Claude output is never SQL.
   - If the validation fails, the backend sends the errors back one time as a `tool_result` with `is_error: true`. If the second answer also fails, the api returns `422` (D-024).
   - The backend compiles the valid rule to a parameterized query. Relative times use the wall clock (D-026).
   - The response has the rule, a description that the backend writes (D-023), and the audience counts for each channel and each language. The backend adds no language filter (D-022).
   - A user is in the count of a channel only if the user has consent for that channel. The query goes through `Recipients`, which applies this consent rule.
   - The api sends CORS headers for the origins in `CORS_ORIGINS` (D-028). The default model is Claude Opus 5.5 (D-027).
2. **Content** — The UI streams the SMS and email content.
   - TBD per slice.
3. **Approval** — The marketer previews, edits and approves the content.
   - TBD per slice.
4. **Send** — A queue sends the messages to the mock provider.
   - The mock provider delivers each request, also a repeat. The worker owns the dedup. It does not retry after a timeout or a dropped connection (D-018, D-019).
5. **Live status** — The dashboard shows the delivery status.
   - TBD per slice.
