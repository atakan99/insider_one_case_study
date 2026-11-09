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
   - TBD per slice.
2. **Content** — The UI streams the SMS and email content.
   - TBD per slice.
3. **Approval** — The marketer previews, edits and approves the content.
   - TBD per slice.
4. **Send** — A queue sends the messages to the mock provider.
   - The mock provider delivers each request, also a repeat. The worker owns the dedup. It does not retry after a timeout or a dropped connection (D-018, D-019).
5. **Live status** — The dashboard shows the delivery status.
   - TBD per slice.
