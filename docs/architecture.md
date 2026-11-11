# Architecture

## Components

- **api** — Go HTTP API. Health endpoints, campaigns, segments, content, send status and the dashboard. It serves `/metrics` on port 8080 (D-068 to D-076).
- **mock-provider** — Go service that acts as an SMS and email provider, like Twilio or SendGrid. It sometimes returns `429`, `5xx`, slow responses and lost responses. It does no dedup (D-018).
- **worker** — Go service. It plans approved campaigns, sends the messages to the mock provider and reports `/healthz` and `/metrics` on port 8082 (D-046 to D-065, D-074).
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
    segment/        rule DSL, validation, SQL builder, sample first names
    campaign/       draft and approved campaigns, edit check, approval
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
   - A user is in the count of a channel only if the user has consent for that channel. The preview count applies consent with `FILTER` columns in `CountAudience`. `Recipients` is the only function that returns user rows. It always adds the consent condition for its channel. Slices 05 to 07 get recipients only through `Recipients`. `SampleFirstNames` returns first names only, for the preview in step 3.
   - The api sends CORS headers for the origins in `CORS_ORIGINS` (D-028). The default model is Claude Opus 5.5 (D-027).
2. **Content** — The UI streams the SMS and email content.
   - `POST /api/content/generate` takes the brief and the rule. The browser sends the rule back. The backend validates the rule again and runs the audience query again (D-032).
   - The backend makes one Claude call for each language, all at the same time. Each call has its own validation and its own repair round (D-030).
   - Claude must answer with the tool `submit_content`. The tool has `eager_input_streaming: true` and `strict: true`, so the drafts stream while Claude writes. The validator is the gate, because the API does not validate an eager stream (D-033).
   - The validator checks the limits after it renders `{{first_name}}` with the longest first name of the language audience (D-031).
   - The response is a Server-Sent Events stream with the events `audience`, `draft`, `retry`, `final`, `error` and `done`.
   - `cmd/api` gives the mock client its default answers, so the flow works with `LLM_MODE=mock` and no key (D-034).
   - The api server has no write timeout, because a stream can last 90 s.
3. **Approval** — The marketer previews, edits and approves the content.
   - At the end of the content stream, the backend saves a draft campaign with the final languages. Failed languages are skipped. The stream sends a `campaign` event with the id, then `done` (D-036). No draft is saved after a client disconnect or when all languages fail.
   - "Regenerate content" sends the draft id. The backend replaces the content of that draft. It compares the brief and the rule (as `jsonb`) with the stored ones (D-037).
   - The review window loads the campaign and sample first names (`GET /api/campaigns/{id}/sample`). The browser counts the limits with the stored `first_name_runes`. The backend makes the decision.
   - `PUT /api/campaigns/{id}/contents` saves the edits. `POST /api/campaigns/{id}/approve` saves the last edits and approves. Both validate the stored rule again, run the audience query again and check the content with the new longest first name (D-038).
   - The service counts the audience before it locks the row. The approval then locks the campaign row and checks the status in one transaction. A repeated approval returns the campaign and changes nothing (D-039).
   - Database triggers block a change of the content, the brief, the rule and the approval time after approval (D-040).
   - The campaign stores the counts at the last save, but no recipient list (D-041). The send step gets the recipients at send time.
4. **Send** — The worker plans and sends the messages.
   - The mock provider delivers each request, also a repeat. The worker owns the dedup. It does not retry after a timeout or a dropped connection (D-018, D-019).
   - The approval does not touch Redis. The worker finds approved campaigns and writes one `send_jobs` row for each reachable user and channel. The unique key blocks a second plan (D-050). The campaign becomes `queued`.
   - Each channel has one Redis stream. An entry holds only the send job id. Postgres holds the delivery status (D-051, D-052).
   - 32 consumers for each channel read the stream (D-049). Each consumer waits for a rate-limit token, then claims the send job with one conditional `UPDATE`. The claim is the dedup gate (D-053, D-046). A token is good for 100 ms. If the claim and the load take longer, the send job goes back to `queued` and takes a new token (D-086).
   - The consumer checks consent again and renders the message. It sends one request for each attempt. A connection error before the write gives a retry. A timeout or a lost connection after the write gives `unknown` (D-057, D-058).
   - A failed attempt that can work later gives `retrying` with `next_attempt_at`. The delay is 1, 2, 4 and 8 s with 20% jitter, for 5 attempts at most. A retry poller every 500 ms puts due retries on the stream (D-047, D-054, D-055).
   - Recovery runs in the worker. It puts old `queued` and due `retrying` send jobs on the stream again. It sets a `sending` send job that hangs to `unknown`. It gives the entry of a dead consumer to a live consumer. It sets a campaign to `completed` when no send job is open (D-056).
   - `POST /api/campaigns/{id}/retry` sets the retryable `failed` send jobs of a `completed` campaign to `queued` with `attempts = 0`. The campaign becomes `sending` (D-059, D-060).
   - `GET /api/campaigns/{id}/progress` and its stream give the counts. `GET /api/campaigns/{id}/send-jobs` gives the list with user ids only (D-061).
5. **Live status** — The dashboard shows the delivery status.
   - `/` goes to `/dashboard`. The page polls `GET /api/dashboard/stats` and `GET /api/dashboard/health` at once, then every 5 s. It stops while the tab is hidden (D-068, D-069, D-070).
   - `stats` gives the totals, the throughput in 120 buckets of 5 s, the top 8 failure reasons, the mean attempts and a table. The table lists the 20 newest campaigns that are not `draft` (D-071, D-072).
   - `health` checks Postgres, Redis, the worker `/healthz` and the provider `/healthz` at the same time. Each check has a timeout of 1 s. It also gives `XLEN` and `XPENDING` of each stream. It returns `200` when a service is `down` (D-069).
   - The api `/readyz` checks Postgres and Redis (D-076).
   - The api `/metrics` has gauges from one `GROUP BY` query at each scrape and the HTTP request metrics. The worker `/metrics` has counters in the process and stream gauges from Redis. No metric has a campaign id label (D-073, D-074, D-075).
