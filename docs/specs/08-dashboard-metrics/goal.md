# Goal: Dashboard and metrics

## Why

Step 5 of the case flow says that a dashboard shows the delivery status. The case also asks for live delivery status, metrics endpoints and health check endpoints.

Slice 07 shows the progress of one campaign on `/campaigns/{id}`. This slice adds a dashboard that shows the delivery statistics of all campaigns together. A reviewer opens the app and sees at once:

- how many messages went out on each channel, and how many failed,
- that the worker keeps each channel at 100 messages per second or fewer,
- why messages failed,
- how each recent campaign did,
- that each service is up and how full the queues are.

The same numbers go to a `/metrics` endpoint in the Prometheus text format.

## Input

- The case brief, flow step 5 and the section with the live status, metrics and health checks.
- Slice 07: the `send_jobs` table, the delivery statuses, the streams `send:sms` and `send:email`, the group `workers`, and the worker `/healthz` on port 8082. See `docs/specs/07-send-queue/goal.md`.
- Slice 03: the mock provider `/healthz` and `/v1/stats`.
- The api `/healthz` and `/readyz` of slice 00.
- The chart component `frontend/src/components/ui/chart` (Unovis). It is in the frontend already.
- User decisions of 2026-11-10:
  - The dashboard shows the delivery statistics.
  - The dashboard has five parts: totals, throughput over time, a small failure card, a campaign table and the live system health.
  - The metrics also go to a `/metrics` endpoint in the Prometheus format.

## Output

1. A page `/dashboard` and a "Dashboard" item at the top of the sidebar. `/` goes to `/dashboard`.
2. Two api endpoints:
   - `GET /api/dashboard/stats` returns the totals, the throughput, the failure breakdown and the campaign table.
   - `GET /api/dashboard/health` returns the state of each service and the queue depth.
3. `GET /metrics` on the api (port 8080) and on the worker (port 8082).
4. A Redis check in the api `/readyz`.

## Dashboard page

The page has five parts, from top to bottom. At 360 px width, each part takes the full width.

### 1. Totals

One row of cards for each channel, SMS and email. Each card shows a number:

| Card | Value |
|---|---|
| Sent | Send jobs with `sent`. |
| Failed | Send jobs with `failed`. |
| Unknown | Send jobs with `unknown`. The card shows "may be delivered". |
| Skipped | Send jobs with `skipped`. |
| In progress | Send jobs with `queued`, `sending` or `retrying`. |
| Success rate | sent / (sent + failed + unknown), as a percent with one decimal. It shows "–" when the divisor is 0. |

The totals include all campaigns and all time.

### 2. Throughput

A line chart of the messages per second over the last 10 minutes.

- One line for SMS and one line for email.
- The x axis has 120 buckets of 5 seconds. The value of a bucket is the number of send jobs with `sent_at` in the bucket, divided by 5.
- A dashed line marks the limit of 100 per second.
- Buckets with no send are 0. The chart has no gaps.

The 5-second buckets show a peak of 100 per second for a short campaign. A bucket of 1 minute hides the peak.

### 3. Failures

A small card with the top 8 reasons. One row is one pair of delivery status and `last_error`, with the count of each channel. The statuses are `failed`, `unknown` and `skipped`.

Examples: `failed 503 unavailable`, `unknown timeout`, `skipped no_consent`.

Under the list, the card shows the mean attempts of the `sent` send jobs for each channel, with two decimals.

### 4. Campaigns

A table of the 20 newest campaigns that are not `draft`. Newest first.

| Column | Value |
|---|---|
| Campaign | `#id` and the description. A link to `/campaigns/{id}`. |
| Status | The badge of slice 07. |
| Send jobs | The count of all send jobs. |
| Sent | Percent of all send jobs. |
| Failed | Percent of all send jobs. |
| Unknown | Percent of all send jobs. |
| Duration | From the earliest `created_at` to the latest `updated_at` of its send jobs. Only for `completed`. Else "–". |

A campaign with no send job shows 0 and "–" for the percents. At 360 px, the table scrolls inside its card. The page itself has no horizontal scroll.

### 5. System health

One badge for each service: api database, Redis, worker and mock provider. Each badge shows `ok` or `down`.

Under the badges, one row for each stream:

| Column | Value |
|---|---|
| Stream | `send:sms` or `send:email`. |
| Length | `XLEN`. |
| Pending | The pending count of the group `workers` (`XPENDING`). |
| In progress | Send jobs of the channel with `queued`, `sending` or `retrying`. |

### Live update

- The page gets `stats` and `health` at once, then every 5 seconds.
- The page stops the requests while the browser tab is hidden. It gets both at once when the tab shows again.
- A failed request keeps the last data and shows a small "Could not refresh" text with the time of the last success.
- A new response does not reset the chart zoom, the scroll or the table.

### Empty state

With no send job, the totals show 0 and the chart shows a flat line. The table shows "No campaigns sent yet" with a link to `/campaigns/new`.

## API

### `GET /api/dashboard/stats`

```json
{
  "generated_at": "2026-11-10T12:00:00Z",
  "totals": {
    "sms":   {"sent": 95, "failed": 0, "unknown": 6, "skipped": 0, "in_progress": 0},
    "email": {"sent": 142, "failed": 0, "unknown": 6, "skipped": 0, "in_progress": 0}
  },
  "throughput": {
    "bucket_seconds": 5,
    "start": "2026-11-10T11:50:00Z",
    "sms":   [0, 0, 19.2],
    "email": [0, 0, 20.0]
  },
  "failures": [
    {"status": "unknown", "reason": "timeout", "sms": 4, "email": 5}
  ],
  "mean_attempts": {"sms": 1.04, "email": 1.02},
  "campaigns": [
    {"id": 1, "description": "...", "status": "completed", "total": 249,
     "sent": 237, "failed": 0, "unknown": 12, "duration_ms": 3120}
  ]
}
```

- `throughput.sms` and `throughput.email` have 120 values each. Value `i` is the bucket that starts at `start + i * 5 s`.
- `start` is the current time rounded down to 5 seconds, minus 600 seconds.
- `duration_ms` is `null` when the campaign is not `completed`.
- The client computes the success rate and the percents.

### `GET /api/dashboard/health`

```json
{
  "services": {"db": "ok", "redis": "ok", "worker": "ok", "provider": "down"},
  "streams": [
    {"channel": "sms",   "length": 0, "pending": 0, "in_progress": 0},
    {"channel": "email", "length": 0, "pending": 0, "in_progress": 0}
  ]
}
```

- The api checks the four services at the same time. Each check has a timeout of 1 second.
- The api calls `WORKER_URL/healthz` and `PROVIDER_URL/healthz`. Compose sets `WORKER_URL=http://worker:8082` and `PROVIDER_URL=http://mock-provider:8081` for the `api` service.
- A service that does not answer `200` in 1 second is `down`.
- The endpoint returns `200` also when a service is `down`. The body tells the state.
- When Redis is `down`, `length` and `pending` are `null`.

### Errors

A database error returns `503` with the error code `db_unavailable`, as in the delivery endpoints of slice 07. The log has the cause.

## Metrics

Both processes use `github.com/prometheus/client_golang`. Each `/metrics` also has the Go and process collectors of the library.

### api `/metrics` (port 8080)

| Metric | Type | Labels |
|---|---|---|
| `campaign_send_jobs` | gauge | `channel`, `status` |
| `campaign_campaigns` | gauge | `status` |
| `campaign_http_requests_total` | counter | `route`, `method`, `code` |
| `campaign_http_request_duration_seconds` | histogram | `route`, `method` |

- The two gauges come from one `GROUP BY` query at each scrape. A query error gives no gauge values in that scrape and a log line.
- `route` is the pattern of the mux, for example `GET /api/campaigns/{id}`. It is never the raw path.
- The SSE routes count in the counter. They do not go in the histogram.

### worker `/metrics` (port 8082)

| Metric | Type | Labels |
|---|---|---|
| `campaign_send_attempts_total` | counter | `channel`, `outcome` |
| `campaign_provider_request_duration_seconds` | histogram | `channel` |
| `campaign_send_results_total` | counter | `channel`, `status` |
| `campaign_rate_limit_wait_seconds_total` | counter | `channel` |
| `campaign_retries_scheduled_total` | counter | `channel` |
| `campaign_stream_length` | gauge | `channel` |
| `campaign_stream_pending` | gauge | `channel` |

- `outcome` is one of `delivered`, `retryable`, `permanent` and `ambiguous`. These are the outcome kinds of the provider client.
- `status` is the delivery status that an attempt or a skip writes.
- The two stream gauges come from Redis at each scrape.

## Health checks

- api `/healthz` stays as it is: it always returns `200`.
- api `/readyz` checks Postgres and Redis. It returns `503` when one is down. The body has `db` and `redis`.
- worker `/healthz` stays as it is.
- `make up` keeps its curl checks.

## Privacy

- No response and no metric has a user id, a name, a phone, an email or message text.
- No metric has a campaign id label.
- The campaign table shows the description. The description is not personal data. The marketer wrote it.

## Tests

- Postgres tests for each query of `stats`, with send jobs that the test inserts. They check each count against a `GROUP BY` in the test.
- A test for the throughput buckets with fixed times: a send at the bucket edge, a send before `start`, and empty buckets.
- A test for `health` with a fake worker and a fake provider: one `ok`, one slow, one `500`.
- A test that scrapes each `/metrics` and checks the metric names and labels. It checks that no label value is a user id or a campaign id.
- Frontend tests for each part of the page and for the empty state.
- Frontend tests with a fake timer for the 5-second poll, the hidden tab and the "Could not refresh" text.
- All tests use the mock LLM. No test reads `ANTHROPIC_API_KEY`.

## Non-goals

- Opens, clicks and conversions. The data has no such events.
- A split by language.
- A Prometheus server or Grafana in compose.
- Server-Sent Events for the dashboard. Polling every 5 seconds is enough for totals.
- A date filter or a time window that the user picks.
- Alerts.
- A change to `/campaigns/{id}` or to the sending popup.

## Calls by Claude

The user did not decide these points. Claude chose them. Each one gets an entry in `docs/decisions-log.md`.

- `/` goes to `/dashboard`, not to `/campaigns/new`. The reviewer sees the statistics first.
- Two endpoints, `stats` and `health`. A slow health check does not delay the statistics.
- Polling every 5 seconds, not Server-Sent Events.
- The throughput chart has 5-second buckets over 10 minutes. A short campaign still shows its peak rate.
- A failure reason is the pair of delivery status and `last_error`. The worker writes a short, fixed text there, so no parse is necessary.
- `/metrics` uses `prometheus/client_golang`, a new Go dependency.
- The api gauges come from Postgres at each scrape. The worker counters live in the worker process and start at 0 after a restart.
- No metric has a campaign id label, so that the number of series stays small.
- The api `/readyz` also checks Redis, because the api reads the streams for `health`.
