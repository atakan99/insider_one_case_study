# Requirements: Send queue

Read `goal.md` first. This file lists the checks that the result must pass.

## Functional requirements

### Campaign status

- **FR-1.** The migration allows the statuses `draft`, `approved`, `queued`, `sending` and `completed`. A Postgres test shows that `approved_at` cannot be null for the last four.
- **FR-2.** A Postgres test shows that a campaign cannot move back to an earlier status. The only exception is `completed` to `sending`.
- **FR-3.** The trigger of slice 06 still blocks a change of the content, the brief, the rule and `approved_at` for each status after `draft`.

### Planner

- **FR-4.** For an approved campaign with the example rule, the planner creates one send job for each user with consent on each channel. The counts match `CountAudience` for the same rule and the languages of the campaign.
- **FR-5.** Users without consent for a channel get no send job on that channel. A Postgres test checks this with users that have SMS consent only, email consent only, both, and none.
- **FR-6.** Users with a skipped language get no send job.
- **FR-7.** Two planner runs on one campaign at the same time give one set of send jobs. The unique key blocks a second row.
- **FR-8.** After the plan, the campaign is `queued` and each send job is on the stream of its channel.
- **FR-9.** A campaign with no send job goes to `completed`.

### Sender

- **FR-10.** A `202` sets `sent`, `provider_id`, `sent_at` and `attempts = 1`.
- **FR-11.** A table test with forced failures gives these results:
  - `429` then `202`: `sent`, `attempts = 2`. The second attempt comes after `Retry-After`.
  - `500` then `202`: `sent`, `attempts = 2`.
  - `503` five times: `failed`, `attempts = 5`, `last_error` with `503`.
  - `400`: `failed`, `attempts = 1`. No second attempt.
  - `slow`: `unknown`, `attempts = 1`. No second attempt.
  - `lost`: `unknown`, `attempts = 1`. No second attempt.
  - Connection refused, then `202`: `sent`, `attempts = 2`.
- **FR-12.** The backoff before attempts 2, 3, 4 and 5 is 1, 2, 4 and 8 s, each with up to 20% jitter. A test with a fake clock checks `next_attempt_at`. A due retry goes on the stream within 1 s.
- **FR-13.** Two consumers that get the same entry at the same time make one attempt. A test pushes one send job two times and checks that the provider got one request.
- **FR-14.** The request has `reference` `<campaign id>:<user id>:<channel>` and no `Idempotency-Key` header.
- **FR-15.** The SMS body and the email subject and body have `{{first_name}}` replaced with the user's first name.
- **FR-16.** A user whose consent changes to false after the plan gets `skipped` with `no_consent`. The provider gets no request.
- **FR-17.** A rendered SMS over 160 runes, or a subject over 80 runes, gives `skipped` with `over_limit`. The provider gets no request. The test sets a long first name directly in the database.
- **FR-18.** The HTTP client makes one request for each attempt. A test with `lost` checks that the provider got one request.

### Rate limit

- **FR-19.** With a fake clock and 1,000 send jobs on one channel, no 1-second window has more than 100 attempts on that channel.
- **FR-20.** SMS and email each send 100 per second at the same time. A slow SMS provider does not slow the email channel.
- **FR-21.** After a `429` with `Retry-After: 2`, the channel makes no attempt for 2 seconds.

### Recovery

- **FR-22.** A `queued` send job with `enqueued_at` null or older than 120 seconds goes on the stream again.
- **FR-23.** A `retrying` send job goes on the stream when `next_attempt_at` is in the past.
- **FR-24.** A `sending` send job that is older than `SEND_TIMEOUT` plus 30 seconds becomes `unknown`.
- **FR-24a.** On `SIGTERM`, a send job that the worker claimed but did not send goes back to `queued`. A test stops the worker during a token wait. The send job is `queued`, not `unknown`, and is sent later.
- **FR-25.** A stream entry of a dead consumer goes to a live consumer after 60 seconds idle.
- **FR-26.** A campaign is `completed` when no send job is `queued`, `sending` or `retrying`.
- **FR-27.** A test stops the worker after the plan and before the push. A new worker sends all send jobs. Each provider `reference` occurs one time.

### End to end

- **FR-28.** A test approves a campaign with the example rule. It runs the worker against the real mock provider handler with the default random failures. All send jobs end in a terminal status. The campaign is `completed`. `GET /v1/stats` has `duplicate_deliveries: 0` for both channels.
- **FR-29.** In the same test, each `sent` send job has one message in the provider with its `reference`. No `failed` send job has a message in the provider.

### API

- **FR-30.** `GET /api/campaigns/{id}/progress` returns the counts in the shape of `goal.md`. The counts match a `GROUP BY` in the test.
- **FR-31.** For a draft, the progress and the stream return `409 campaign_not_sendable`. An unknown id returns `404`. An id of `abc`, `0` or `-1` returns `400`.
- **FR-32.** The progress stream sends a `progress` event at once. It sends no event in a second with no change. It sends `done` and closes when the campaign is `completed`.
- **FR-33.** `GET /api/campaigns/{id}/send-jobs` filters by `status` and `channel` and pages with `after`. A bad value returns `400`. The response has no name, no phone, no email and no text. A test checks the JSON keys.
- **FR-34.** `POST /api/campaigns/{id}/retry` on a `completed` campaign with retryable failed send jobs sets them to `queued` with `attempts = 0`. It sets the campaign to `sending` and returns the number. A `400` failure stays `failed`. `unknown` send jobs do not change.
- **FR-35.** A retry on a campaign that is not `completed` returns `409 campaign_not_completed`. A retry with nothing to retry returns `409 nothing_to_retry`.
- **FR-36.** Two retries at the same time give one retry. The second returns `409`.

### Frontend

- **FR-37.** After "Send now" succeeds, the page goes to `/campaigns/{id}` and opens the sending popup.
- **FR-38.** The stepper marks Approved, Queued, Sending and Done from the campaign status. A test checks each status.
- **FR-39.** The popup has one row for each channel with a total above 0. The bar segments have widths in proportion to sent, retrying, failed and unknown.
- **FR-40.** The popup updates from `progress` events. When the stream fails, it calls `GET .../progress` every 2 seconds.
- **FR-41.** At `completed`, the popup shows the totals text. "Open delivery list" closes the popup and scrolls to the list.
- **FR-42.** "View progress" shows while the campaign is `approved`, `queued` or `sending`.
- **FR-43.** "Retry N failed" shows only when the campaign is `completed` and `retryable_failed` is above 0. It opens a confirm popup first. "Send again" calls the retry endpoint and opens the sending popup.
- **FR-44.** The send job list shows "User #id", channel, language, status, attempts and last error. The `unknown` status shows "may be delivered". The chips filter by status. "Load more" gets the next page.
- **FR-45.** `/campaigns` shows the badges Draft, Approved, Queued, Sending and Completed.
- **FR-46.** All new colours come from theme tokens. The popup and the list work in the light theme and the dark theme.

## Non-functional requirements

- **NFR-1.** No test reads `ANTHROPIC_API_KEY`. The `worker` service gets no `env_file`.
- **NFR-2.** The worker logs no name, phone, email or message text. A test captures the logs of an end-to-end run and checks this.
- **NFR-3.** The send job list endpoint and the progress endpoints return no personal data except the user id.
- **NFR-4.** `internal/delivery` gets user rows only through the `segment` package, with the consent condition of the channel.
- **NFR-5.** At 360 px width, `/campaigns/{id}` and the sending popup have no horizontal scroll.
- **NFR-6.** `make test` starts a temporary Redis and removes it. It works with no local Go.

## Acceptance

- `make verify` exits 0.
- A clean `docker compose up --wait` starts the `worker` service. Its `/healthz` reports Postgres and Redis as `ok`.
- With the example brief and `LLM_MODE=mock`, the flow from "Find audience" to "Send now" ends with the campaign `completed`. The popup shows the four steps done.
- After the run, the mock provider `GET /v1/stats` has `duplicate_deliveries: 0` for both channels.
- The mock provider's `rate_limited` counter stays low. The plan sets a number after a first run.
- A headless Chrome run at 1280 px and 360 px shows the popup, the list and the retry button.
