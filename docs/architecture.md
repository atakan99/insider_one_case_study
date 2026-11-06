# Architecture

## Components

- **api** — Go HTTP API. Health endpoints, campaigns, segments, content and send status.
- **mock-provider** — Go service that acts as an SMS and email provider.
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
    store/          sqlc-generated queries and migrations
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
   - TBD per slice.
5. **Live status** — The dashboard shows the delivery status.
   - TBD per slice.
