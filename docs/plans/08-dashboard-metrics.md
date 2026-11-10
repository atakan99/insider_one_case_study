# 08 — Dashboard and Metrics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task by task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A `/dashboard` page shows the delivery statistics of all campaigns, the throughput, the failure reasons, a campaign table and the system health. The api and the worker expose `/metrics` in the Prometheus text format.

**Architecture:** This slice adds one package, one migration, two endpoints, two `/metrics` handlers and one page:

```
store/migrations   00004_dashboard.sql: partial index on send_jobs (sent_at)
internal/config    Config.WorkerURL, Config.ProviderURL
internal/delivery  queue.go    Queue.Depth (XLEN, XPENDING)
                   metrics.go  Metrics: attempt, result, wait and retry counters, stream gauges
                   consumer.go calls Metrics (nil-safe)
internal/dashboard types.go    Stats, Totals, Throughput, Failure, CampaignRow, Health
                   store.go    Postgres: Stats, InProgress
                   health.go   Checker: db, redis, worker, provider, streams
                   collector.go prometheus.Collector for campaign_send_jobs, campaign_campaigns
internal/httpapi   dashboard.go GET /api/dashboard/stats, GET /api/dashboard/health
                   metrics.go   request counter and histogram middleware
                   health.go    /readyz also pings Redis
cmd/api            Checker, collector, /metrics, middleware
cmd/worker         Metrics, /metrics on :8082
frontend           lib/dashboard.ts, pages/dashboard.vue, components/dashboard/*.vue,
                   sidebar item, / redirect
```

`internal/dashboard` imports only `delivery` and `store` from `internal/`. It selects no user column. It reads `send_jobs` and `campaigns` only.

**Tech Stack:** Go 1.27, `pgx/v5`, goose, `go-redis/v9`, `github.com/prometheus/client_golang` (new), Vue 3, shadcn-vue (`Card`, `Badge`, `Table`), `@unovis/vue` through `components/ui/chart`, Vitest.

**Spec:** `docs/specs/08-dashboard-metrics/goal.md` and `requirements.md`.

## Global Constraints

- Delivery statuses: `queued`, `sending`, `retrying`, `sent`, `failed`, `unknown`, `skipped`. In progress: `queued`, `sending`, `retrying`.
- Channels: `sms`, `email`. Streams `send:sms`, `send:email`. Group `workers`.
- Throughput: 120 buckets of 5 s. `start = now.Truncate(5s) - 600s`. A bucket value is count / 5. The window is `[start, start + 600s)`.
- Failures: top 8 pairs of `status` and `last_error`, for `failed`, `unknown` and `skipped`. Order by the total, then by `status`, then by `reason`.
- Campaign table: 20 newest campaigns that are not `draft`, by `id` descending. `duration_ms` only for `completed`, else `null`.
- Health: each check has a 1 s timeout. All four run at the same time. The endpoint always returns `200`.
- Poll: 5 s. No request while `document.visibilityState === 'hidden'`. No second request while one is in flight.
- Error codes: `db_unavailable` with `503` for a database error. Use `writeError` of `internal/httpapi/errors.go`.
- Metric names, types and labels exactly as in `goal.md` section "Metrics". The `route` label is `r.Pattern`, or `unmatched` when it is empty. No metric has a campaign id label or a user id label.
- No response, no metric and no log line has a name, a phone, an email or message text.
- UI: no hard-coded colour. Use the tokens of slice 07: SMS `--chart-1`, email `--chart-2`, limit line `--muted-foreground`. Status colours as `status-badge.vue`. No horizontal scroll at 360 px.
- Tests use `llm.Mock`. No test reads `ANTHROPIC_API_KEY`. Postgres tests use `storetest.NewDB`. Redis tests use `deliverytest.NewRedis` and skip without `TEST_REDIS_URL`.
- Do not commit after each task. The user commits the slice after review.

## Review Focus

1. **Redis is up, but the worker never ran, so the streams and the group do not exist.** `XPENDING` gives `NOGROUP` and `XLEN` gives 0. The health part must show `length: 0`, `pending: 0` and `redis: ok`, not an error. Pinned by Task 2 `TestQueue_DepthNoStream`.
2. **A response takes longer than 5 s, for example while Postgres is slow.** The poll must not start a second request. Requests must not stack up. Pinned by Task 5 `does not start a second poll while one is in flight`.
3. **A scanner requests many unknown paths.** The `route` label must stay `unmatched`. It must not create one series for each path. Pinned by Task 3 `TestMetrics_UnmatchedRoute`.
4. **A retry moves a completed campaign back to `sending`.** The table must show `duration_ms: null` and the `sending` badge. It must not show the old duration. Pinned by Task 1 `TestStats_DurationAfterRetry`.
5. **The marketer leaves `/dashboard` while a request runs.** The late response must not write state, and no new timer must start. Pinned by Task 5 `stops polling after unmount, also with a request in flight`.

---

### Task 1: Dashboard store

**Files:**
- Create: `backend/internal/store/migrations/00004_dashboard.sql`
- Create: `backend/internal/dashboard/types.go`, `backend/internal/dashboard/store.go`
- Test: `backend/internal/dashboard/store_test.go`

**Interfaces:**
- Consumes: `delivery.Channel`, `delivery.Channels`, `delivery.Status` and its constants. `storetest.NewDB`, `store.Migrate`.
- Produces:

```go
type ChannelTotals struct {
    Sent, Failed, Unknown, Skipped, InProgress int64 // json: sent, failed, unknown, skipped, in_progress
}
type Throughput struct {
    BucketSeconds int       `json:"bucket_seconds"` // always 5
    Start         time.Time `json:"start"`
    SMS           []float64 `json:"sms"`   // len 120
    Email         []float64 `json:"email"` // len 120
}
type Failure struct {
    Status string `json:"status"`
    Reason string `json:"reason"`
    SMS    int64  `json:"sms"`
    Email  int64  `json:"email"`
}
type CampaignRow struct {
    ID          int64  `json:"id"`
    Description string `json:"description"`
    Status      string `json:"status"`
    Total, Sent, Failed, Unknown int64 // json: total, sent, failed, unknown
    DurationMS  *int64 `json:"duration_ms"`
}
type Stats struct {
    GeneratedAt  time.Time                         `json:"generated_at"`
    Totals       map[delivery.Channel]ChannelTotals `json:"totals"`
    Throughput   Throughput                        `json:"throughput"`
    Failures     []Failure                         `json:"failures"`      // never nil
    MeanAttempts map[delivery.Channel]*float64     `json:"mean_attempts"` // nil value when no sent job
    Campaigns    []CampaignRow                     `json:"campaigns"`     // never nil
}

func NewStore(pool *pgxpool.Pool) *Store
func (s *Store) Stats(ctx context.Context, now time.Time) (Stats, error)
func (s *Store) InProgress(ctx context.Context) (map[delivery.Channel]int64, error)
func (s *Store) Counts(ctx context.Context) (jobs map[delivery.Channel]map[delivery.Status]int64, campaigns map[string]int64, err error)
```

- [ ] **Step 1: Write the failing tests.** Each test makes a database with `storetest.NewDB` and migrates it. It inserts users, campaigns and send jobs with SQL. Use one helper `insertJob(t, db, campaignID, userID, channel, status, attempts, lastError, sentAt, createdAt, updatedAt)`.
  - `TestStats_Totals`: 3 sent, 1 failed, 2 unknown, 1 skipped, 1 queued, 1 sending, 1 retrying on SMS, and 2 sent on email. Expect SMS `{3,1,2,1,3}` and email `{2,0,0,0,0}`. Compare with a `SELECT channel, status, count(*) ... GROUP BY 1, 2` in the test (FR-1, FR-2).
  - `TestStats_ThroughputBuckets`: `now = 12:00:07`. Expect `Start = 11:50:05` and 120 values per channel. 50 SMS at `11:55:00.000` give 10 at index 59. One SMS at `11:50:04.999` is in no bucket. One SMS at `11:50:05.000` gives 0.2 at index 0. One SMS at `12:00:05` is in no bucket. All other values are 0 (FR-3, FR-4).
  - `TestStats_Failures`: 10 reason pairs. Expect 8 rows, highest total first. `sent` and `queued` jobs never occur. A pair with SMS 2 and email 3 gives one row with `sms: 2, email: 3` (FR-5).
  - `TestStats_MeanAttempts`: SMS sent with attempts 1, 1, 2 gives `1.333...`. Email with no sent job gives `nil` (FR-6).
  - `TestStats_Campaigns`: 22 non-draft campaigns and 1 draft. Expect 20 rows, ids descending, no draft. A `completed` campaign with jobs from `created_at 10:00:00` to `updated_at 10:00:03.120` gives `duration_ms: 3120`. A `queued` campaign gives `nil`. A campaign with no send job gives total 0 (FR-7, FR-8).
  - `TestStats_DurationAfterRetry`: a campaign that went `completed`, then `sending`, gives `status: "sending"` and `duration_ms: nil`.
  - `TestStats_Empty`: no send job gives zero totals, 240 zero values, `failures: []`, `campaigns: []` and both `mean_attempts` nil.
  - `TestInProgress`: counts `queued + sending + retrying` for each channel.
  - `TestCounts`: matches the `GROUP BY` of `TestStats_Totals`, and `campaigns` by status includes `draft`.

- [ ] **Step 2: Run the tests.** Run: `make test`. Expected: FAIL, because package `dashboard` does not exist.

- [ ] **Step 3: Write `00004_dashboard.sql`.** Up: `CREATE INDEX send_jobs_sent_at ON send_jobs (sent_at) WHERE sent_at IS NOT NULL;`. Down: `DROP INDEX send_jobs_sent_at;`.

- [ ] **Step 4: Implement `Store` in `store.go`.** `Stats` runs all its queries in one `REPEATABLE READ, READ ONLY` transaction, so that the parts agree. The throughput query uses the bucket index in SQL:

```sql
SELECT channel, floor(extract(epoch FROM sent_at - $1) / 5)::int AS b, count(*)
FROM send_jobs WHERE sent_at >= $1 AND sent_at < $1 + interval '600 seconds'
GROUP BY 1, 2
```

  Duration is `extract(epoch FROM max(updated_at) - min(created_at)) * 1000`, rounded to an integer, and only when `campaigns.status = 'completed'`. The campaign rows come from one query: the 20 campaigns, `LEFT JOIN` the per-campaign `count(*) FILTER (WHERE ...)` aggregates.

- [ ] **Step 5: Run the tests.** Run: `make test`. Expected: PASS.

- [ ] **Step 6: Check the index.** In the test database with 50,000 send jobs, run `EXPLAIN` on the throughput query. Expected: `Index Scan` or `Bitmap Index Scan` on `send_jobs_sent_at`. Write the result in the task report (NFR-4).

---

### Task 2: Queue depth and the health checker

**Files:**
- Modify: `backend/internal/delivery/queue.go`
- Modify: `backend/internal/config/config.go`
- Create: `backend/internal/dashboard/health.go`
- Test: `backend/internal/delivery/queue_test.go`, `backend/internal/config/config_test.go`, `backend/internal/dashboard/health_test.go`

**Interfaces:**
- Consumes: Task 1 `Store.InProgress`. `deliverytest.NewRedis`.
- Produces:

```go
// delivery
func (q *Queue) Depth(ctx context.Context, ch Channel) (length, pending int64, err error)

// config: two new fields of Config
WorkerURL   string // WORKER_URL, default "http://localhost:8082"
ProviderURL string // PROVIDER_URL, default "http://localhost:8081"

// dashboard
type StreamDepth struct {
    Channel    delivery.Channel `json:"channel"`
    Length     *int64           `json:"length"`  // nil when Redis is down
    Pending    *int64           `json:"pending"` // nil when Redis is down
    InProgress int64            `json:"in_progress"`
}
type Health struct {
    Services map[string]string `json:"services"` // keys db, redis, worker, provider; values "ok" or "down"
    Streams  []StreamDepth     `json:"streams"`  // sms first, then email
}
type Pinger interface{ Ping(ctx context.Context) error }
type Checker struct {
    DB          Pinger
    Redis       Pinger
    Queue       *delivery.Queue
    Store       *Store
    HTTP        *http.Client
    WorkerURL   string
    ProviderURL string
    Timeout     time.Duration // 0 means 1 s
}
func (c *Checker) Check(ctx context.Context) Health
```

- [ ] **Step 1: Write the failing tests.**
  - `TestQueue_Depth` (Redis): push 3 ids on `send:sms`, read 2 with a consumer. Expect `length 3, pending 2`.
  - `TestQueue_DepthNoStream` (Redis): an empty Redis gives `0, 0, nil` for both channels.
  - `TestLoad_ServiceURLs`: defaults, and `WORKER_URL=http://worker:8082` overrides. An URL with no scheme returns an error.
  - `TestChecker_AllOK`: fake worker and provider with `httptest.Server` that answer `200`. Expect four `ok`.
  - `TestChecker_WorkerDown`: the worker answers `500`. The provider sleeps 3 s. Expect both `down`, and `Check` returns in under 2 s (FR-10).
  - `TestChecker_RedisDown`: a `Redis` pinger that fails and a queue on a closed client. Expect `redis: down`, `length` and `pending` nil, and `in_progress` from Postgres (FR-12).
  - `TestChecker_Streams` (Redis and Postgres): expect `length` and `pending` as `TestQueue_Depth` (FR-11).

- [ ] **Step 2: Run the tests.** Run: `make test`. Expected: FAIL.

- [ ] **Step 3: Implement `Queue.Depth`.** Use `XLEN` and `XPENDING <key> workers` (summary form). Treat `NOGROUP` and a missing key as 0, as `Lag` does.

- [ ] **Step 4: Implement the two config fields** in `Load`, with `url.Parse` and a check for the `http` or `https` scheme.

- [ ] **Step 5: Implement `Checker.Check`.** Run the four checks and the depth reads in goroutines, each with its own `context.WithTimeout(ctx, timeout)`. An HTTP check is `ok` only for status `200` from `GET <url>/healthz`. Log the cause of a `down` at `Warn`, without the URL query.

- [ ] **Step 6: Run the tests.** Run: `make test`. Expected: PASS.

---

### Task 3: Dashboard endpoints, api metrics and `/readyz`

**Files:**
- Create: `backend/internal/httpapi/dashboard.go`, `backend/internal/httpapi/metrics.go`, `backend/internal/dashboard/collector.go`
- Modify: `backend/internal/httpapi/health.go`, `backend/cmd/api/main.go`, `docker-compose.yml`, `.env.example`, `backend/go.mod`, `backend/go.sum`
- Test: `backend/internal/httpapi/dashboard_test.go`, `backend/internal/httpapi/metrics_test.go`, `backend/internal/httpapi/health_test.go`, `backend/internal/dashboard/collector_test.go`

**Interfaces:**
- Consumes: Task 1 `Store.Stats`, `Store.Counts`. Task 2 `Checker.Check`, `Config.WorkerURL`, `Config.ProviderURL`.
- Produces:

```go
// httpapi
type DashboardService interface {
    Stats(ctx context.Context, now time.Time) (dashboard.Stats, error)
}
type HealthChecker interface {
    Check(ctx context.Context) dashboard.Health
}
func RegisterDashboard(mux *http.ServeMux, stats DashboardService, health HealthChecker, now func() time.Time)
func NewHTTPMetrics(reg prometheus.Registerer) *HTTPMetrics
func (m *HTTPMetrics) Wrap(next http.Handler) http.Handler
// RegisterHealth gets a second Pinger for Redis:
func RegisterHealth(mux *http.ServeMux, db, redis Pinger)

// dashboard
func NewCollector(s *Store, log *slog.Logger) prometheus.Collector
```

- [ ] **Step 1: Write the failing tests.**
  - `TestDashboardStats_OK`: a fake service. The body has the keys `generated_at, totals, throughput, failures, mean_attempts, campaigns`, and `mean_attempts.email` is `null`.
  - `TestDashboardStats_DBError`: expect `503` and `"code":"db_unavailable"`.
  - `TestDashboardHealth_AlwaysOK`: a fake checker with `worker: down`. Expect `200` and the body.
  - `TestDashboard_NoPersonalData`: decode both bodies to `map[string]any` and walk all keys. No key is `user_id`, `first_name`, `phone`, `email_address`, `subject` or `body`. Each value under a key `sms` or `email` is a number, an array of numbers or `null` (NFR-2).
  - `TestReadyz_RedisDown`: DB up, Redis down. Expect `503` with `{"status":"unavailable","db":"ok","redis":"down"}` (FR-13). Update the old `/readyz` tests to pass a Redis pinger that is up.
  - `TestMetrics_Route`: wrap a mux with `GET /api/campaigns/{id}`. Request `/api/campaigns/1`. Scrape the registry with `testutil.CollectAndCount` and `testutil.ToFloat64`. Expect `campaign_http_requests_total{route="GET /api/campaigns/{id}",method="GET",code="200"} 1` (FR-16).
  - `TestMetrics_UnmatchedRoute`: request 50 unknown paths. Expect one series with `route="unmatched"` and value 50.
  - `TestMetrics_StreamNotInHistogram`: a route that ends in `/stream` counts in the counter and has no histogram sample.
  - `TestCollector` (Postgres): the jobs of Task 1 `TestStats_Totals`. `testutil.GatherAndCompare` gives `campaign_send_jobs{channel="sms",status="sent"} 3` and `campaign_campaigns{status="completed"} N` (FR-14, FR-15).
  - `TestCollector_DBDown`: a closed pool gives no samples and no panic.

- [ ] **Step 2: Run the tests.** Run: `make test`. Expected: FAIL.

- [ ] **Step 3: Add the dependency.** Run `go get github.com/prometheus/client_golang@latest` and `go mod tidy` in the `golang:1.27-alpine` container, as the Makefile does.

- [ ] **Step 4: Implement `RegisterDashboard`, `HTTPMetrics` and `NewCollector`.**
  - `Wrap` reads `r.Pattern` after `next.ServeHTTP` returns. The mux sets it on the same request. It records the status code with a small `ResponseWriter` wrapper that keeps `http.Flusher` working, because the SSE handlers need it.
  - A route is a stream route when the pattern ends in `/stream` or is `POST /api/content/generate`. Stream routes get no histogram sample.
  - Histogram buckets: `prometheus.DefBuckets`.
  - The collector runs `Store.Counts` with a 2 s timeout inside `Collect`. On an error it logs and sends no metric.

- [ ] **Step 5: Wire `cmd/api`.**
  - Make one `prometheus.NewRegistry()`. Register `collectors.NewGoCollector()`, `collectors.NewProcessCollector(...)`, `NewHTTPMetrics(reg)` and `dashboard.NewCollector(...)`.
  - Add `GET /metrics` with `promhttp.HandlerFor(reg, promhttp.HandlerOpts{})`.
  - The server handler becomes `httpapi.CORS(cfg.CORSOrigins, metrics.Wrap(mux))`.
  - Move the Redis client before `RegisterHealth`, and pass it as the Redis pinger.
  - `RegisterDashboard(mux, dashboard.NewStore(pool), &dashboard.Checker{...}, time.Now)`.

- [ ] **Step 6: Update compose and `.env.example`.** The `api` service gets `WORKER_URL: http://worker:8082` and `PROVIDER_URL: http://mock-provider:8081`. `.env.example` lists both with the localhost defaults.

- [ ] **Step 7: Run the tests.** Run: `make verify`. Expected: exit 0.

---

### Task 4: Worker metrics

**Files:**
- Create: `backend/internal/delivery/metrics.go`
- Modify: `backend/internal/delivery/consumer.go`, `backend/internal/delivery/worker.go`, `backend/cmd/worker/main.go`
- Test: `backend/internal/delivery/metrics_test.go`, `backend/internal/delivery/consumer_test.go`, `backend/cmd/worker/main_test.go` (create it if it does not exist)

**Interfaces:**
- Consumes: Task 2 `Queue.Depth`.
- Produces:

```go
type Metrics struct { /* unexported prometheus vectors */ }
func NewMetrics(reg prometheus.Registerer, q *Queue) *Metrics
func (m *Metrics) Attempt(ch Channel, kind OutcomeKind, d time.Duration) // attempts_total + duration histogram
func (m *Metrics) Result(ch Channel, s Status)                          // send_results_total
func (m *Metrics) Waited(ch Channel, d time.Duration)                   // rate_limit_wait_seconds_total
func (m *Metrics) RetryScheduled(ch Channel)                            // retries_scheduled_total
func (k OutcomeKind) String() string // "delivered", "retryable", "permanent", "ambiguous"
// Consumer and Worker get a field:
Metrics *Metrics // nil means no metrics
```

All methods of `*Metrics` do nothing on a nil receiver.

- [ ] **Step 1: Write the failing tests.**
  - `TestOutcomeKind_String`: the four values.
  - `TestMetrics_NilSafe`: all four methods on a nil `*Metrics` do not panic.
  - `TestConsumer_Metrics` (Postgres and Redis): run the forced-failure table of slice 07 FR-11 with a `Metrics` on a new registry. After `503` then `202`: `campaign_send_attempts_total{channel="sms",outcome="retryable"} 1`, `{outcome="delivered"} 1`, `campaign_retries_scheduled_total{channel="sms"} 1`, `campaign_send_results_total{channel="sms",status="retrying"} 1` and `{status="sent"} 1` (FR-17).
  - `TestConsumer_MetricsSkip`: a `no_consent` skip gives `send_results_total{status="skipped"} 1` and no attempt sample.
  - `TestMetrics_StreamGauges` (Redis): 3 entries on `send:sms` give `campaign_stream_length{channel="sms"} 3`.
  - `TestWorkerMetrics_EndToEnd`: extend the slice 07 end-to-end test. After the run, the sum of `send_results_total{status="sent"}` equals the number of `sent` send jobs (FR-18). Scrape the registry text. No label value matches `^[0-9]+$`, and none has `@` (FR-19).

- [ ] **Step 2: Run the tests.** Run: `make test`. Expected: FAIL.

- [ ] **Step 3: Implement `Metrics`.** The stream gauges are a `prometheus.Collector` inside `Metrics` that calls `Queue.Depth` with a 1 s timeout in `Collect`. On an error it sends no sample. The duration histogram uses the buckets `0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5`.

- [ ] **Step 4: Call `Metrics` in `Consumer.Handle` and `finish`.**
  - Measure the time of `Limiter.Wait` with `c.Clock.Now()` and call `Waited`.
  - Measure the time of `Provider.Send` and call `Attempt` with `o.Kind`.
  - In `finish`, after a successful `Store.Finish`, call `Result(r.Status)`. When `r.Status == StatusRetrying`, also call `RetryScheduled`.
  - `Worker.Run` passes `w.Metrics` to each `Consumer`.

- [ ] **Step 5: Wire `cmd/worker`.** Make one registry with the Go and process collectors and `delivery.NewMetrics(reg, queue)`. Add `GET /metrics` to the mux of `healthHandler`. Pass the metrics to the `Worker`.

- [ ] **Step 6: Run the tests.** Run: `make verify`. Expected: exit 0.

---

### Task 5: Dashboard page

**Files:**
- Create: `frontend/src/lib/dashboard.ts`, `frontend/src/pages/dashboard.vue`
- Create: `frontend/src/components/dashboard/totals-cards.vue`, `throughput-chart.vue`, `failures-card.vue`, `campaigns-table.vue`, `health-panel.vue`
- Modify: `frontend/src/pages/index.vue`, `frontend/src/constants/sidebar-data.ts`, `frontend/src/router/__tests__/routes.test.ts`, `frontend/src/components/app-sidebar/__tests__/app-sidebar.test.ts`
- Test: `frontend/src/lib/__tests__/dashboard.test.ts`, `frontend/src/components/dashboard/__tests__/*.test.ts`, `frontend/src/pages/__tests__/dashboard.test.ts`

**Interfaces:**
- Consumes: the JSON of Task 3. `apiFetch` of `lib/api-client.ts`. `call` error mapping as in `lib/delivery.ts`. `StatusBadge` of `components/campaign/status-badge.vue`.
- Produces:

```ts
export interface ChannelTotals { sent: number, failed: number, unknown: number, skipped: number, in_progress: number }
export interface DashboardStats { generated_at: string, totals: Record<DeliveryChannel, ChannelTotals>,
  throughput: { bucket_seconds: number, start: string, sms: number[], email: number[] },
  failures: { status: DeliveryStatus, reason: string, sms: number, email: number }[],
  mean_attempts: Record<DeliveryChannel, number | null>,
  campaigns: { id: number, description: string, status: CampaignStatus, total: number, sent: number, failed: number, unknown: number, duration_ms: number | null }[] }
export interface DashboardHealth { services: Record<'db' | 'redis' | 'worker' | 'provider', 'ok' | 'down'>,
  streams: { channel: DeliveryChannel, length: number | null, pending: number | null, in_progress: number }[] }
export function getDashboardStats(): Promise<DashboardStats>
export function getDashboardHealth(): Promise<DashboardHealth>
export function successRate(t: ChannelTotals): string   // "95.2%" or "–"
export function percent(part: number, total: number): string // "95.2%" or "–" when total is 0
export function formatDuration(ms: number | null): string    // "3.1 s", "2 min 5 s", or "–"
export const DASHBOARD_POLL_MS = 5000
```

- [ ] **Step 1: Write the failing tests.**
  - `lib/__tests__/dashboard.test.ts`: `successRate({sent: 95, failed: 0, unknown: 5, ...})` is `"95.0%"`. All zero is `"–"`. `percent(1, 3)` is `"33.3%"`. `formatDuration(3120)` is `"3.1 s"`, `125000` is `"2 min 5 s"`, `null` is `"–"`.
  - `totals-cards.test.ts`: two rows of six cards. The Unknown card has "may be delivered" (FR-21).
  - `throughput-chart.test.ts`: the chart gets 120 points with `x` as the bucket time and `sms`, `email` and `limit: 100` (FR-22). Test the computed data, not the SVG.
  - `failures-card.test.ts`: one row per failure with both counts. The mean attempts show `"1.04"`, and `null` shows `"–"` (FR-23).
  - `campaigns-table.test.ts`: the seven columns. Each row has a link to `/campaigns/{id}`. An empty list shows "No campaigns sent yet" with a link to `/campaigns/new` (FR-24, FR-29).
  - `health-panel.test.ts`: four badges and two stream rows. `length: null` shows "–" (FR-25).
  - `pages/__tests__/dashboard.test.ts`, with `vi.useFakeTimers()` and mocked `getDashboardStats` and `getDashboardHealth`:
    - `polls both endpoints at once and every 5 seconds` (FR-26).
    - `makes no request while the tab is hidden, and polls at once when it shows` (FR-27). Set `document.visibilityState` with `Object.defineProperty` and send `visibilitychange`.
    - `keeps the last data and shows Could not refresh` (FR-28).
    - `does not start a second poll while one is in flight`.
    - `stops polling after unmount, also with a request in flight` (NFR-5).
  - `routes.test.ts`: remove `/dashboard` from the removed routes. `/` redirects to `/dashboard`. `/dashboard` resolves to the dashboard page.
  - `app-sidebar.test.ts`: the first link is "Dashboard" with `href="/dashboard"` (FR-20).

- [ ] **Step 2: Run the tests.** Run: `cd frontend && pnpm test:run`. Expected: FAIL.

- [ ] **Step 3: Implement `lib/dashboard.ts`** with the interfaces above.

- [ ] **Step 4: Implement the five components.**
  - `throughput-chart.vue` uses `ChartContainer` with `VisXYContainer`. It has two `VisLine` for SMS and email, and one `VisLine` with `lineDashArray: [4, 4]` for the limit.
  - The chart has `VisAxis` for x (time `HH:mm:ss`) and y, and `ChartCrosshair` with `ChartTooltipContent`. The colours come from `ChartConfig` with `var(--chart-1)`, `var(--chart-2)` and `var(--muted-foreground)`.
  - `campaigns-table.vue` puts the `Table` in a `div` with `overflow-x-auto`, so that only the card scrolls at 360 px (NFR-3).
  - The cards use `Card`. The badges use `Badge`. A `down` badge uses the `destructive` variant. An `ok` badge uses the `success` token of slice 07.

- [ ] **Step 5: Implement `pages/dashboard.vue`.**
  - One `refresh()` that calls both endpoints with `Promise.allSettled` and sets `inFlight`. It returns at once when `inFlight` is true, or when the tab is hidden.
  - A `setTimeout` chain of `DASHBOARD_POLL_MS` starts the next refresh after the last one ends. A `visibilitychange` listener calls `refresh()` when the tab shows.
  - An `unmounted` flag drops late responses. `onUnmounted` clears the timer and the listener.
  - A failed part keeps its last data. The page shows "Could not refresh. Last update HH:mm:ss" in `text-muted-foreground`.
  - Layout: one column below `md`, two columns from `lg` for the chart and the failures card.

- [ ] **Step 6: Update the redirect and the sidebar.** `pages/index.vue` redirects to `/dashboard`. `navData` gets a first group "Overview" with `{ title: 'Dashboard', url: '/dashboard', icon: LayoutDashboardIcon }`.

- [ ] **Step 7: Run the checks.** Run: `cd frontend && pnpm test:run && pnpm build && CI=true pnpm lint`. Expected: all pass.

---

### Task 6: Docs and the end-to-end check

**Files:**
- Modify: `docs/architecture.md`, `docs/decisions-log.md`, `tasks/backlog.json`, `tasks/progress.md`

- [ ] **Step 1: Update `docs/architecture.md`.** Flow step 5 describes the dashboard, the two endpoints and the two `/metrics`. The service list names `/metrics` on the api and the worker.

- [ ] **Step 2: Add decisions D-068 and later** to `docs/decisions-log.md`. Write one for each point of "Calls by Claude" in `goal.md`. Write one for each new call of a task. Use the `ste-writing` skill.

- [ ] **Step 3: Set the backlog item.** In `tasks/backlog.json`, `08-dashboard-metrics` gets `"spec": "docs/specs/08-dashboard-metrics/"` and the FR, NFR and acceptance lines of `requirements.md`. `passes` stays `false`.

- [ ] **Step 4: Run the end-to-end check.** Use a compose project `-p cc-verify` on free ports. The user's stack holds 3000, 8080, 8081 and 8082.
  - `docker compose -p cc-verify up -d --build --wait` with `LLM_MODE=mock`.
  - Run the example brief to "Send now". Wait for `completed`.
  - `GET /api/dashboard/stats`: the totals match `GET /api/campaigns/{id}/progress`. The highest throughput value is 100 or lower.
  - `GET /api/dashboard/health`: four `ok`, pending 0.
  - `docker compose -p cc-verify stop mock-provider`. Within 10 s, `provider` is `down`. `start` makes it `ok` again.
  - `curl` the api and the worker `/metrics`. Check the metric names of `goal.md`.
  - Headless Chrome at 1280 px and 360 px, in the light theme and the dark theme. Check the five parts and that `document.documentElement.scrollWidth <= innerWidth`.
  - `docker compose -p cc-verify down -v`.

- [ ] **Step 5: Run `make verify`.** Expected: exit 0.

- [ ] **Step 6: Append the session entry** to `tasks/progress.md`. Use the `ste-writing` skill and run `ste_check.py` on all changed `.md` files.
