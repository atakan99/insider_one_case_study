# Goal: Send queue

## Why

The case flow has five steps. Step 4 says that the messages go to the mock provider through a queue. The case also gives these rules:

- The queue is asynchronous.
- Each channel sends 100 messages per second or fewer.
- A user does not get the same message two times.
- We design the retry strategy.

This slice builds step 4. After the approval, a worker sends one rendered message to each user with consent, on each channel. A popup shows the steps and the progress. The campaign page lists each send job with its delivery status. The marketer can send the failed send jobs again.

## Input

- The case brief, section "Gönderim", and the hard rules in `CLAUDE.md`.
- Slice 03: the mock provider API, the failure modes, and the section "What the worker must do (slice 07)". See `docs/specs/03-mock-provider/goal.md`.
- Slice 04: `segment.Recipients`, the only function that returns user rows. It always adds the consent condition of its channel.
- Slice 05: the renderer and the validator of the channel limits.
- Slice 06: the approved campaign, the content of each language, and `/campaigns/{id}`. See `docs/specs/06-review-approve/goal.md`.
- Earlier decisions: Redis Streams is the queue (D-007). The provider does no dedup (D-018). The worker does not retry a timeout or a dropped connection (D-019).
- The mockup that the user accepted on 2026-11-10 (option A). A copy is in `docs/specs/07-send-queue/sending-mockup.html`.
- User decisions of 2026-11-10:
  - After "Send now", a popup shows four steps: Approved, Queued, Sending and Done.
  - Under the steps, the popup shows one bar for SMS and one bar for email. Each bar has a colour for each outcome: sent, retrying, failed and unknown.
  - The marketer can close the popup. Sending continues.
  - The campaign page lists each send job with its delivery status.
  - The marketer can send the failed send jobs again. The marketer cannot send the `unknown` send jobs again, because the provider can have them.
  - The popup updates live from the server.
  - The list shows the user number, not the name, the phone or the email.
  - Slice 07 builds the queue, the worker, the popup, the send job list and the retry. Slice 08 builds the live charts, the metrics endpoint and the health checks for the dashboard.

## Output

1. A migration that adds the new campaign statuses and the `send_jobs` table.
2. A Go package `internal/delivery`. It has the planner, the queue, the rate limiter, the sender and the recovery.
3. A binary `cmd/worker` and a `worker` service in `docker-compose.yml`.
4. New endpoints in `internal/httpapi`:
   - `GET /api/campaigns/{id}/progress` returns the counts of each channel and each delivery status.
   - `GET /api/campaigns/{id}/progress/stream` sends the same counts as Server-Sent Events.
   - `GET /api/campaigns/{id}/send-jobs` lists the send jobs.
   - `POST /api/campaigns/{id}/retry` sends the failed send jobs again.
5. A sending popup component.
6. On `/campaigns/{id}`: the progress, the send job list and the retry button.
7. On `/campaigns`: the badges of the new statuses.

## Flow

### Send

1. The marketer clicks "Send now" in the confirm popup of slice 06. The approval request runs as before.
2. On success, the page goes to `/campaigns/{id}` and opens the sending popup. The step "Approved" is done.
3. The worker finds the approved campaign. It creates one send job for each user, channel and language. It then puts the send jobs on the queue. The campaign status becomes `queued`. The step "Queued" is done.
4. The worker sends the send jobs. The first attempt sets the campaign status to `sending`. The bars grow.
5. When no send job is `queued`, `sending` or `retrying`, the campaign status becomes `completed`. The step "Done" is done.
6. The popup shows the totals. "Open delivery list" closes the popup and scrolls to the list.

The marketer can close the popup at any step. A "View progress" button on the campaign page opens it again while the campaign sends.

### Retry

1. When the campaign is `completed` and has retryable failed send jobs, the page shows the button "Retry N failed".
2. A click opens a confirm popup: "Send N failed messages again? Messages with an unknown result are not sent again."
3. "Send again" calls `POST /api/campaigns/{id}/retry`. The campaign status becomes `sending`. The sending popup opens.
4. Steps 4 to 6 of "Send" follow.

## Campaign status

The migration widens the `status` CHECK of `campaigns`.

| Status | Meaning | Set by |
|---|---|---|
| `draft` | Slice 06. | api |
| `approved` | Slice 06. The worker did not plan it yet. | api |
| `queued` | The worker created the send jobs and put them on the queue. | worker |
| `sending` | One or more attempts started. | worker, or api for a retry |
| `completed` | No send job is `queued`, `sending` or `retrying`. | worker |

- The status moves only forward: `approved`, `queued`, `sending`, `completed`. A retry moves `completed` back to `sending`. No other move back is allowed.
- `approved_at` stays set for all statuses after `draft`. The CHECK of slice 06 stays.
- The migration replaces the trigger function of slice 06 with `CREATE OR REPLACE`. The new function keeps the freeze of the content, the brief, the rule and the approval time. It also blocks each backward status move except `completed` to `sending`.
- A campaign with no send job goes from `queued` to `completed` at once.

## Send jobs

### Table `send_jobs`

| Column | Type | Note |
|---|---|---|
| `id` | `bigint` identity | Primary key. |
| `campaign_id` | `bigint` | References `campaigns`. |
| `user_id` | `bigint` | References `users`. |
| `channel` | `text` | `sms` or `email`. |
| `language` | `text` | The language of the content. |
| `status` | `text` | The delivery status. See below. |
| `attempts` | `integer` | The number of attempts. |
| `next_attempt_at` | `timestamptz` | For `retrying`. Null for the other statuses. |
| `last_error` | `text` | A short code and text, for example `503 unavailable`. No personal data. It stays after a later `sent`. |
| `retryable` | `boolean` | True when the last error was `429`, `5xx` or a connection error. The retry endpoint uses it. |
| `provider_id` | `text` | The `id` from the `202` response. |
| `enqueued_at` | `timestamptz` | The last time that a process put the send job on the stream. Null before the first push. |
| `created_at` | `timestamptz` | Not null. |
| `updated_at` | `timestamptz` | Not null. |
| `sent_at` | `timestamptz` | Set for `sent`. |

- The unique key `(campaign_id, user_id, channel)` is the idempotency key. A user gets one send job for each campaign and channel. A second planner run or a double approval cannot add a second one.
- The table stores no phone, no email address and no rendered text. The worker reads them from `users` and `campaign_contents` at send time.
- An index on `(campaign_id, status)` serves the progress counts.

### Delivery status

| Status | Meaning |
|---|---|
| `queued` | The send job waits on the queue. |
| `sending` | The worker claimed it. An attempt runs. |
| `retrying` | The last attempt failed and was not delivered. Another attempt comes at `next_attempt_at`. |
| `sent` | The provider returned `202`. |
| `failed` | The provider did not deliver it and no attempt is left, or the error is permanent. |
| `unknown` | The attempt timed out or the connection dropped. The provider can have the message. |
| `skipped` | The worker did not send it: no consent now, no contact, or the rendered text is over the limit. |

The terminal statuses are `sent`, `failed`, `unknown` and `skipped`.

## Planner

The planner runs in the worker. It checks for `approved` campaigns every second.

1. It locks one `approved` campaign with `FOR UPDATE SKIP LOCKED`.
2. It validates the stored rule again and compiles it.
3. For each channel, it inserts one send job for each user in the audience with consent for that channel. The user's language must be in `campaign_contents`. The insert uses `ON CONFLICT DO NOTHING` on the idempotency key.
4. It sets the campaign status to `queued` and commits.
5. It puts each `queued` send job of the campaign on the queue of its channel.

- The insert gets the users only through the `segment` package. A new function next to `Recipients` returns the ids for the insert. It always adds the consent condition of its channel.
- The sender also needs a new `segment` function. It reads one user by id with the consent condition of the channel. The comment that calls `Recipients` the only function that returns user rows changes. It names the three functions.
- Users with a skipped language get no send job.
- If the rule fails the validation, the campaign goes to `completed` with no send jobs. The worker logs the error. This must not happen, because the api validated the rule at approval.
- A crash after step 4 and before step 5 is safe. The recovery puts the send jobs on the queue.

## Queue

- Redis Streams. One stream for each channel: `send:sms` and `send:email`. One consumer group `workers`.
- An entry has only the send job id. The worker reads the rest from Postgres.
- The queue gives each entry one time or more. The claim in Postgres makes sure that the worker sends each send job one time at most. See "Dedup".
- The worker acknowledges an entry after it writes the result of the attempt to Postgres.
- A retry with a delay does not wait on the stream. The send job gets `retrying` and `next_attempt_at`. A retry poller in the worker checks every 500 ms. It puts each due send job on the stream again. Thus a retry goes out at most about 0.5 s after `next_attempt_at`.

## Sender

The sender runs in the worker. Each channel has its own consumers and its own rate limiter. A slow channel does not stop the other channel.

For each entry:

1. **Wait** for a token of the channel rate limiter.
2. **Claim.** `UPDATE send_jobs SET status = 'sending', attempts = attempts + 1 WHERE id = $1 AND status IN ('queued', 'retrying') AND (next_attempt_at IS NULL OR next_attempt_at <= now())`. If no row changes, acknowledge the entry and stop. Another consumer has it, or it is done.
3. **Load.** Read the user and the content of the send job language. Read the user with the consent condition of the channel.
4. **Check.**
   - If the user has no consent now, set `skipped` with `no_consent`.
   - If the user has no phone or no email for the channel, set `skipped` with `no_contact`.
   - Render the template with the user's first name. Check the SMS against 160 runes, or the subject against 80 runes. If the text is over the limit, set `skipped` with `over_limit`. This must not happen, because the approval checked the longest first name.
5. **Send** one HTTP request to the provider, with a timeout. Set `reference` to `<campaign id>:<user id>:<channel>`. Set no `Idempotency-Key` header.
6. **Write the result** to Postgres. See "Retry strategy".
7. **Acknowledge** the entry.

The email `body` is the rendered email body. The provider gets plain text. The email has no sender field, because the mock provider has none.

## Retry strategy

| Result of the attempt | Delivered? | New delivery status |
|---|---|---|
| `202` | Yes | `sent`, with `provider_id` and `sent_at`. |
| `429` | No | `retrying`. `next_attempt_at` is now plus `Retry-After`. The channel limiter also pauses for `Retry-After`. |
| `500` or `503` | No | `retrying` with backoff. |
| Connection refused, or a DNS error | No | `retrying` with backoff. |
| `400` | No | `failed`. Permanent. No retry. |
| Any other `4xx` | No | `failed`. Permanent. No retry. |
| Timeout after the request was sent | Maybe | `unknown`. No retry. |
| Dropped connection after the request was sent | Maybe | `unknown`. No retry. |

- Backoff: 1 s, 2 s, 4 s and 8 s, each with a random jitter of up to 20%.
- `SEND_MAX_ATTEMPTS` is 5. All attempts count, also `429`. After the last retryable attempt fails, the status is `failed`.
- The attempt timeout `SEND_TIMEOUT` is 3 s. It is shorter than the `MOCK_SLOW_DELAY` of 5 s. Thus a `slow` response gives `unknown`.
- The HTTP client must not retry by itself. Each call to the provider is one attempt.
- A failed send job keeps its `last_error`. The marketer sees it in the list.

### Retryable failed send jobs

The marketer can send a `failed` send job again only if its last error was retryable: `429`, `5xx` or a connection error. A `400` or another `4xx` stays `failed`. The retry button counts only the retryable ones.

## Dedup

The case says that a user does not get the same message two times. These rules make sure of it:

- The idempotency key `(campaign_id, user_id, channel)` allows one send job for each user, campaign and channel.
- The claim in step 2 of the sender is one conditional `UPDATE`. Two consumers with the same entry cannot both claim it.
- The worker sends a new attempt only when the provider did not deliver the last one. These results show that: `429`, `5xx`, or a connection error before the request was sent.
- The worker never sends a send job again after `unknown`. The marketer cannot either.
- A double approval gives one approval (slice 06). The planner runs one time for each campaign.

A test runs a full campaign against the mock provider with random failures. It then checks that `GET /v1/stats` has `duplicate_deliveries: 0`.

## Rate limit

- Each channel has a token bucket with `golang.org/x/time/rate`. The rate is `SEND_RATE_PER_SEC`, default 100. The burst is 1. Thus a channel sends 100 requests per second or fewer.
- The worker waits for a token before each attempt, also for a retry.
- After a `429`, the limiter of that channel stops for `Retry-After` seconds.
- The compose file runs one `worker` replica. The limiter is in that process. A second replica would double the rate. A shared limiter is a non-goal.

## Recovery

The recovery runs in the worker every 5 seconds. It makes the queue safe after a crash or a lost entry.

- It puts on the stream each `queued` send job with `enqueued_at` null or older than 120 seconds. It sets `enqueued_at`. A second entry for a send job that is still on the stream is safe.
- It puts on the stream each send job that is `retrying` with `next_attempt_at` more than 30 seconds in the past. The retry poller missed it.
- It sets `unknown` on each send job that is `sending` for more than `SEND_TIMEOUT` plus 30 seconds. The worker can have sent the request before it stopped.
- It claims the stream entries of a dead consumer with `XAUTOCLAIM` after 60 seconds idle.
- It sets a `queued` or `sending` campaign to `completed` when no send job of the campaign is `queued`, `sending` or `retrying`.

An entry on the stream two times is safe, because the claim lets one attempt through only.

## API

All errors use the slice 04 error shape.

### `GET /api/campaigns/{id}/progress`

Response `200`:

```json
{
  "campaign_id": 12,
  "status": "sending",
  "channels": {
    "sms":   {"total": 980,  "queued": 300, "sending": 10, "retrying": 4, "sent": 660, "failed": 2, "unknown": 3, "skipped": 1, "retryable_failed": 2},
    "email": {"total": 1212, "queued": 400, "sending": 12, "retrying": 2, "sent": 795, "failed": 0, "unknown": 3, "skipped": 0, "retryable_failed": 0}
  }
}
```

- For an `approved` campaign, all counts are 0.
- A draft returns `409 campaign_not_sendable`.

### `GET /api/campaigns/{id}/progress/stream`

- An SSE stream. It sends a `progress` event with the body of `GET .../progress` at once, then every second when a count changed.
- When the campaign status is `completed`, it sends the last `progress` event and a `done` event, then closes.
- The stream closes after 30 minutes. The page opens it again if the campaign is still sending.
- A draft returns `409 campaign_not_sendable` before the stream starts.

### `GET /api/campaigns/{id}/send-jobs`

Query parameters, all optional: `status`, `channel`, `limit` (default 50, maximum 200), `after` (a send job id for the next page).

Response `200`:

```json
{"send_jobs": [{"id": 501, "user_id": 4012, "channel": "sms", "language": "tr", "status": "sent", "attempts": 1, "last_error": "", "updated_at": "2026-11-10T12:00:01Z"}], "next_after": 501}
```

- The list has no name, no phone, no email and no rendered text.
- The order is by `id`. `next_after` is null on the last page.
- A bad `status`, `channel`, `limit` or `after` returns `400 invalid_request`.

### `POST /api/campaigns/{id}/retry`

- No body.
- The api does these steps in one transaction:
  1. Lock the campaign and check that it is `completed`.
  2. Set the retryable `failed` send jobs to `queued`, with `attempts` 0 and no `next_attempt_at`.
  3. Set the campaign to `sending`.
- After the commit, put the send jobs on the queue. The api needs `REDIS_URL` for this. The api service gets a dependency on Redis in compose. If the push fails, the recovery puts them on the queue.
- Response `200`: `{"retried": 14}` and the new progress.

| Status | Code | When |
|---|---|---|
| `400` | `invalid_request` | The id is not valid. |
| `404` | `not_found` | No campaign has this id. |
| `409` | `campaign_not_completed` | The campaign is not `completed`. |
| `409` | `nothing_to_retry` | The campaign has no retryable failed send job. |
| `503` | `db_unavailable` | A query failed. |

### Approval

`POST /api/campaigns/{id}/approve` does not change. The worker finds the approved campaign. The api does not put work on the queue at approval.

## Sending popup

The popup follows option A of the mockup. It uses `Dialog` and the theme tokens.

- Title: "Sending campaign #12". Text: "You can close this. Sending continues in the background."
- A stepper with four steps: Approved, Queued, Sending, Done. A done step has a check mark. The current step has the primary colour.
- One row for each channel with a recipient count above 0: the channel name, "sent / total", and one stacked bar.
- The bar has one segment for each outcome: sent, retrying, failed, unknown. A legend names the colours. Skipped send jobs count in "total" and show in the legend text, not in the bar.
- When the campaign is `completed`, the text changes to "Done. 2,170 sent, 12 failed, 10 unknown."
- Buttons: "Close" and "Open delivery list".
- The popup reads the progress stream. If the stream fails, it calls `GET .../progress` every 2 seconds.
- The colours come from tokens. The slice adds tokens for the outcome colours if the theme has none, in the light theme and the dark theme.
- On a phone, the popup fills the width. The stepper labels go under the dots. The page has no horizontal scroll at 360 px width.

## Campaign page

`/campaigns/{id}` for a campaign that is not a draft:

- The line "Sending starts in a later step" of slice 06 goes away.
- A delivery section shows a summary for each channel: sent, failed, unknown, skipped and total.
- While the campaign is `approved`, `queued` or `sending`, a "View progress" button opens the sending popup.
- When the campaign is `completed` and has retryable failed send jobs, the button "Retry N failed" shows.
- The send job list:
  - Status filter chips: All, Sent, Retrying, Failed, Unknown, Skipped. Each chip shows its count.
  - A channel filter: All, SMS, Email.
  - Columns: Recipient ("User #4012"), Channel, Language, Status, Attempts, Last error.
  - The `unknown` status shows the hint "may be delivered".
  - A "Load more" button gets the next page.
  - On a phone, the rows become stacked cards.

## Campaign list

The status badges on `/campaigns`: Draft, Approved, Queued, Sending, Completed. A click on a campaign that is not a draft goes to `/campaigns/{id}`.

## Worker process

- `cmd/worker` runs the planner, the consumers, the recovery and a small HTTP server with `GET /healthz`. The health check reports Postgres and Redis.
- It stops on `SIGTERM`:
  1. It stops reading new entries and stops the token waits.
  2. A send job that it claimed but did not send goes back to `queued` with `attempts - 1`. The provider did not get it.
  3. It waits up to 10 seconds for the requests that started.
  4. A request that does not end in time stays `sending`. The recovery sets it to `unknown`.
- After a crash, the worker cannot know if a `sending` send job reached the provider. The recovery sets it to `unknown`. This is the safe choice for the dedup.
- Compose publishes port 8082. `make up` checks `/healthz` with curl, as for the api and the mock provider. The distroless image has no shell for a container healthcheck.
- The compose `worker` service starts after the seed and after Redis is healthy. It gets `DATABASE_URL`, `REDIS_URL`, `PROVIDER_URL` and the `SEND_*` settings. It gets no `env_file`. Thus it does not get the API key.

### Logs

For each campaign, the worker logs the plan (campaign id, counts for each channel) and the completion (counts for each delivery status). For each attempt with an error, it logs the send job id, the channel, the attempt number and the error code. It does not log a name, a phone, an email or the message text.

## Settings

| Env var | Process | Default | Note |
|---|---|---|---|
| `REDIS_URL` | worker, api | `redis://redis:6379/0` in compose | Required. |
| `PROVIDER_URL` | worker | `http://mock-provider:8081` in compose | Required. |
| `SEND_RATE_PER_SEC` | worker | `100` | Per channel. 1 to 100. |
| `SEND_MAX_ATTEMPTS` | worker | `5` | 1 to 10. |
| `SEND_TIMEOUT` | worker | `3s` | Must be shorter than `MOCK_SLOW_DELAY`. |
| `SEND_CONSUMERS` | worker | `32` | Consumers per channel. A `slow` response holds a consumer for `SEND_TIMEOUT`. 32 consumers keep 100 per second with the default mock failures. |
| `HTTP_ADDR` | worker | `:8082` | For `/healthz`. |

## Tests

- `make test` also starts a temporary Redis, the same way as the test Postgres. The Go tests read `TEST_REDIS_URL`.
- The sender tests use a fake clock for the backoff and the rate limiter. The timeout cases (`slow`, `lost`) use short real durations, for example a 50 ms timeout and a 200 ms slow delay. The end-to-end test uses the same short durations.
- The end-to-end test uses `httptest` with the real `internal/mockprovider` handler.

## Non-goals

- Live charts, the metrics endpoint and the health page. Slice 08 builds them.
- A retry of `unknown` send jobs.
- A shared rate limiter for more than one worker replica.
- Scheduled sending.
- Cancel or pause of a campaign.
- Delivery receipts or webhooks from the provider.
- A real email sender name or a real provider.

## Calls by Claude

The user did not decide these points. Claude chose them. Each one gets an entry in `docs/decisions-log.md`.

- The worker waits for a rate-limit token before it claims a send job. On `SIGTERM`, a claimed send job with no request goes back to `queued`. Only a crash or a started request gives `unknown`.
- A retry poller every 500 ms puts due retries on the stream. The recovery is the fallback.
- A `retryable` column marks the failed send jobs that the marketer can retry. The api does not parse `last_error`.
- 32 consumers per channel, because each `slow` response holds one consumer for 3 s.
- The worker plans the approved campaigns. The approval does not touch Redis. Thus the approval stays one Postgres transaction.
- One stream for each channel, so that a slow channel does not stop the other.
- A stream entry has only the send job id. Postgres is the record of the delivery status.
- The claim is one conditional `UPDATE`. It is the dedup gate, not Redis.
- A delayed retry waits in Postgres (`next_attempt_at`), not in Redis. The recovery puts it on the stream again.
- 5 attempts with backoff 1, 2, 4 and 8 s and 20% jitter. A `429` counts as an attempt and pauses the channel.
- The attempt timeout is 3 s, shorter than the slow delay of the mock.
- A `sending` send job that hangs after a crash becomes `unknown`, not `queued`. The worker can have sent it.
- The worker checks consent again and renders each message at send time. The `skipped` status records why it did not send.
- The marketer can retry only failed send jobs with a retryable error, and only when the campaign is `completed`.
- The retry sets `attempts` to 0. The marketer gets a full new set of attempts.
- One worker replica with a limiter in the process. A shared limiter is a non-goal.
- The progress stream sends at most one event per second.
- The send job list shows the user id only.
