# Requirements: Dashboard and metrics

Read `goal.md` first. This file lists the checks that the result must pass.

## Functional requirements

### Stats endpoint

- **FR-1.** `GET /api/dashboard/stats` returns the shape of `goal.md`. The totals of each channel match a `GROUP BY channel, status` in the test.
- **FR-2.** `in_progress` is the sum of `queued`, `sending` and `retrying`.
- **FR-3.** `throughput` has 120 values for each channel and `bucket_seconds: 5`. `start` is the current time rounded down to 5 seconds, minus 600 seconds.
- **FR-4.** A test with a fake clock puts 50 SMS `sent_at` values in one bucket. That bucket has the value 10. A send before `start` is not in a bucket. A send at the start of a bucket goes in that bucket.
- **FR-5.** `failures` has at most 8 rows, ordered by the total count, highest first. Each row is one pair of `status` and `reason`. Only `failed`, `unknown` and `skipped` occur.
- **FR-6.** `mean_attempts` is the mean `attempts` of the `sent` send jobs of each channel. With no `sent` send job, it is `null`.
- **FR-7.** `campaigns` has at most 20 rows, newest first. It has no `draft`. The counts of each row match its send jobs.
- **FR-8.** `duration_ms` is set only for `completed`. It is `null` for the other statuses.

### Health endpoint

- **FR-9.** `GET /api/dashboard/health` returns `ok` for each service that answers in 1 second.
- **FR-10.** A worker that answers `500`, or that does not answer in 1 second, gives `worker: "down"`. The endpoint still returns `200` in under 2 seconds.
- **FR-11.** `streams` has `length` from `XLEN` and `pending` from `XPENDING` of the group `workers`. A test with a temporary Redis checks both.
- **FR-12.** With Redis down, `redis` is `down`, and `length` and `pending` are `null`. `in_progress` still comes from Postgres.
- **FR-13.** The api `/readyz` returns `503` when Postgres or Redis is down. The body has `db` and `redis`.

### Metrics

- **FR-14.** The api `/metrics` has `campaign_send_jobs` with the labels `channel` and `status`. The values match the totals of FR-1.
- **FR-15.** The api `/metrics` has `campaign_campaigns` with the label `status`.
- **FR-16.** After a request to `GET /api/campaigns/1`, the api `/metrics` has `campaign_http_requests_total` with `route="GET /api/campaigns/{id}"`.
- **FR-17.** The worker `/metrics` has the metrics of `goal.md`. After a run with forced failures, `campaign_send_attempts_total` has the outcomes `delivered` and `retryable`.
- **FR-18.** The sum of `campaign_send_results_total` for `sent` matches the number of `sent` send jobs after one run on a clean worker.
- **FR-19.** No label value in either `/metrics` is a user id, a campaign id, a name, a phone or an email. A test scrapes both and checks this.

### Dashboard page

- **FR-20.** `/` goes to `/dashboard`. The sidebar has "Dashboard" as the first item.
- **FR-21.** The totals show one row of six cards for each channel. The success rate is sent / (sent + failed + unknown) with one decimal. It is "–" when the divisor is 0.
- **FR-22.** The throughput chart has one line for SMS, one line for email and a dashed line at 100.
- **FR-23.** The failures card shows the rows of `failures` and the mean attempts of each channel.
- **FR-24.** The campaign table shows the columns of `goal.md`. Each row links to `/campaigns/{id}`.
- **FR-25.** The health part shows one badge for each service and one row for each stream.
- **FR-26.** The page calls both endpoints at once and then every 5 seconds. A test with a fake timer checks this.
- **FR-27.** While the tab is hidden, the page makes no request. When the tab shows again, it calls both endpoints at once.
- **FR-28.** A failed request keeps the last data and shows "Could not refresh" with the time of the last success.
- **FR-29.** With no send job, the table shows "No campaigns sent yet" with a link to `/campaigns/new`.
- **FR-30.** All colours come from theme tokens. The page works in the light theme and the dark theme.

## Non-functional requirements

- **NFR-1.** No test reads `ANTHROPIC_API_KEY`. The `worker` service still gets no `env_file`.
- **NFR-2.** No dashboard response and no metric has personal data or message text.
- **NFR-3.** At 360 px width, `/dashboard` has no horizontal scroll. The campaign table scrolls inside its card.
- **NFR-4.** With 50,000 send jobs, `GET /api/dashboard/stats` answers in under 300 ms on the local compose stack. The plan adds an index if `EXPLAIN` shows a full scan for the throughput query.
- **NFR-5.** Leaving `/dashboard` stops the poll. A test checks that no request occurs after the unmount.

## Acceptance

- `make verify` exits 0.
- A clean `docker compose up --wait`, then the example brief to "Send now" with `LLM_MODE=mock`. After the campaign is `completed`, `/dashboard` shows:
  - totals that match `GET /api/campaigns/{id}/progress`,
  - a throughput peak at 100 per second or lower,
  - all four services `ok`, and streams with pending 0.
- `docker compose stop mock-provider` makes the provider badge `down` within 10 seconds. `start` makes it `ok` again.
- `curl localhost:8080/metrics` and `curl localhost:8082/metrics` return the metrics of `goal.md`.
- A headless Chrome run at 1280 px and 360 px shows the five parts in the light theme and the dark theme.
