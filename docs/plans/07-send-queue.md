# 07 — Send Queue Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task by task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** After the approval, a worker sends one rendered message to each user with consent, on each channel, through Redis Streams to the mock provider. The marketer follows the progress in a popup, sees each send job on the campaign page, and can send the retryable failed send jobs again.

**Architecture:** This slice adds one package, one binary, one migration, four endpoints and two frontend components:

```
store/migrations   00003_send_jobs.sql: new statuses, new trigger function, send_jobs
internal/segment   ReachableQuery (planner insert), RecipientByID (sender)
internal/config    WorkerConfig, LoadWorker; Config.RedisURL
internal/delivery  types.go     Status, Channel, Job, Progress, errors
                   store.go     Postgres: plan, claim, release, finish, recovery, progress, list, retry
                   queue.go     Redis Streams: ensure, push, read, ack, autoclaim
                   provider.go  HTTP client to the provider, outcome classification
                   policy.go    retry policy (outcome + attempts -> next state)
                   limiter.go   channel token bucket with pause
                   worker.go    planner loop, consumers, retry poller, recovery, shutdown
                   service.go   api side: progress, list, retry (+ push)
internal/httpapi   delivery.go  progress, progress stream, send-jobs, retry
cmd/worker         main.go, /healthz
cmd/api            Redis client, delivery.Service
frontend           lib/delivery.ts, sending-dialog.vue, delivery-section.vue,
                   [id].vue, index.vue badges, review-dialog route
```

`internal/delivery` imports only `config`, `content`, `segment` and `store` from `internal/`. It never selects user columns itself. User rows come from `segment` with the consent condition (NFR-4).

**Tech Stack:** Go 1.27, `pgx/v5`, goose, `github.com/redis/go-redis/v9` (new), `golang.org/x/time/rate`, `net/http/httptrace`, Vue 3, shadcn-vue (`Dialog`, `AlertDialog`, `Badge`, `Table`, `Button`), Vitest.

**Spec:** `docs/specs/07-send-queue/goal.md` and `requirements.md`. The mockup is `docs/specs/07-send-queue/sending-mockup.html` (option A).

## Global Constraints

- Campaign statuses: `draft`, `approved`, `queued`, `sending`, `completed`. Forward only. The one move back is `completed` to `sending`.
- Delivery statuses: `queued`, `sending`, `retrying`, `sent`, `failed`, `unknown`, `skipped`. Terminal: `sent`, `failed`, `unknown`, `skipped`.
- Idempotency key: `UNIQUE (campaign_id, user_id, channel)` on `send_jobs`.
- Streams `send:sms` and `send:email`, group `workers`. An entry has one field `job` with the send job id.
- `reference` = `<campaign id>:<user id>:<channel>`. No `Idempotency-Key` header.
- Limits after rendering with the user's first name: SMS 160 runes, subject 80 runes. Use `content.Render` and `content.MaxSMSRunes`, `content.MaxSubjectRunes`.
- Defaults: `SEND_RATE_PER_SEC=100` (1..100), `SEND_MAX_ATTEMPTS=5` (1..10), `SEND_TIMEOUT=3s`, `SEND_CONSUMERS=32`, worker `HTTP_ADDR=:8082`. Burst 1.
- Backoff before attempts 2..5: 1 s, 2 s, 4 s, 8 s, each times `1 + U(0, 0.2)`.
- Loops: planner 1 s, retry poller 500 ms, recovery 5 s. Stale `queued` 120 s. Stale `retrying` 30 s past due. Stale `sending` `SEND_TIMEOUT + 30s`. `XAUTOCLAIM` min idle 60 s. Shutdown grace 10 s.
- Skip reasons in `last_error`: `no_consent`, `no_contact`, `over_limit`.
- Error codes: `invalid_request`, `not_found`, `campaign_not_sendable`, `campaign_not_completed`, `nothing_to_retry`, `db_unavailable`.
- No endpoint and no log line has a name, a phone, an email or message text. The send job list has the user id only.
- UI: no hard-coded colour. New tokens `--success` and `--info` in `:root` and `.dark`, mapped in `@theme inline`. Outcome colours: sent `success`, retrying `info`, failed `destructive`, unknown `warning`. No horizontal scroll at 360 px.
- Tests use `llm.Mock`. No test reads `ANTHROPIC_API_KEY`. Postgres tests use `storetest.NewDB`. Redis tests use `TEST_REDIS_URL` and skip without it.
- Do not commit after each task. The user commits the slice after review.

## Review Focus

1. **Redis loses the streams, for example after `docker compose down -v` of only Redis, or a `FLUSHALL`.** Consumers get `NOGROUP`. The worker must create the group again and the recovery must push the `queued` send jobs again, with no duplicate. Pinned by Task 5 `TestWorker_StreamLost`.
2. **Postgres fails when the worker writes the result of a `202`.** The entry is not acknowledged. `XAUTOCLAIM` gives it to another consumer. The claim must fail, because the status is still `sending`, so no second request goes out. The recovery later sets `unknown`. Pinned by Task 4 `TestConsumer_FinishFailsNoResend`.
3. **The popup opens before the planner ran.** The campaign is `approved` and all counts are 0. The popup must show "Approved" as the current step, not "Done" with `0 / 0`. Pinned by Task 7 `shows Approved before the plan`.
4. **The audience has no user with consent on one channel, or on both.** The popup hides a channel with total 0. A campaign with no send job goes to `completed` at once, and the popup shows Done with no bars. Pinned by Task 3 `TestStore_PlanNoJobs` and Task 7 `hides a channel with no recipients`.
5. **A user has SMS consent but an empty phone.** The sender must not call the provider. The send job is `skipped` with `no_contact`. Pinned by Task 4 `TestConsumer_NoContact`.

---

### Task 1: Migration and segment reads

**Files:**
- Create: `backend/internal/store/migrations/00003_send_jobs.sql`
- Modify: `backend/internal/store/migrate_test.go`, `backend/internal/segment/audience.go`, `backend/internal/segment/audience_test.go`

**Interfaces:**
- Produces:

```go
package segment
// ReachableQuery returns a SELECT of (id, language) for the users that match w,
// consent to ch and have a language in langs. Use it as the source of an
// INSERT ... SELECT. The caller appends its own args after args.
func ReachableQuery(w Where, ch Channel, langs []string) (sql string, args []any, err error)

// RecipientByID reads one user with the consent condition of ch.
// ok is false when the user does not exist or has no consent for ch.
func RecipientByID(ctx context.Context, q Querier, id int64, ch Channel) (r Recipient, ok bool, err error)

// Querier is the part of pgxpool.Pool and pgx.Tx that RecipientByID needs.
type Querier interface {
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// ChannelByName returns SMS() for "sms" and Email() for "email".
func ChannelByName(name string) (Channel, error)
```

Migration `00003_send_jobs.sql`:
- Drop the status CHECK of `campaigns` and add `CHECK (status IN ('draft','approved','queued','sending','completed'))`. The name of the old CHECK comes from `00002`. Look it up with `\d campaigns` and drop it by name.
- `CREATE OR REPLACE FUNCTION campaigns_frozen_after_approval()`. Keep all checks of `00002`. Add: when `OLD.status <> NEW.status`, allow only `draft→approved`, `approved→queued`, `approved→completed`, `queued→sending`, `queued→completed`, `sending→completed`, `completed→sending`. Raise an exception for the rest.
- Table `send_jobs` as in the spec, plus `retryable boolean NOT NULL DEFAULT false` and `enqueued_at timestamptz`. `channel CHECK IN ('sms','email')`. `status CHECK IN (...)` with the seven values. `attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0)`. `campaign_id` references `campaigns (id) ON DELETE CASCADE`. `user_id` references `users (id)`.
- `UNIQUE (campaign_id, user_id, channel)`. Index `(campaign_id, status)`. Partial index `(next_attempt_at) WHERE status = 'retrying'`. Partial index `(enqueued_at) WHERE status = 'queued'`.
- Down: drop `send_jobs`, restore the `00002` CHECK and the `00002` function body.

Change the comment above `Recipients`: three functions return user rows (`Recipients`, `ReachableQuery`, `RecipientByID`). Each adds the consent condition of its channel.

- [ ] **Step 1: Write the failing tests**
  - `TestMigrate_CampaignStatusMoves` (FR-1, FR-2): for each allowed move in the list above, the update succeeds. `queued→approved`, `sending→queued`, `completed→approved`, `sending→draft` fail. A status `queued` with `approved_at` null fails.
  - `TestMigrate_FrozenAfterApprovalStill` (FR-3): for status `sending`, an update of `brief`, `rule` or `approved_at` fails. An update of a `campaign_contents` row fails.
  - `TestMigrate_SendJobsUnique`: a second insert with the same `(campaign_id, user_id, channel)` fails with a unique violation. The same user on the other channel succeeds.
  - `TestReachableQuery_Consent` (FR-5): insert four users with SMS only, email only, both, none. With a rule that matches all four, the SMS query returns the first and the third. The email query returns the second and the third.
  - `TestReachableQuery_Languages` (FR-6): `langs = ["tr"]` drops the `en` users.
  - `TestReachableQuery_Columns`: the SQL does not match `\blast_name\b`, `\bemail\b` (as a column), `\bphone\b`.
  - `TestRecipientByID`: a user with consent returns `ok = true` and the phone. Without consent `ok = false`. An unknown id `ok = false`, `err = nil`.
- [ ] **Step 2: Run** `make test` and check that the new tests fail (no migration, undefined functions).
- [ ] **Step 3: Write the migration and the three functions.**
- [ ] **Step 4: Run** `make test`. Expected: PASS, and the earlier `TestMigrate_*` tests still pass. Run the Down and the Up again in `TestMigrate_DownUp` if that test exists.

---

### Task 2: Settings, Redis in tests, and the queue

**Files:**
- Modify: `backend/go.mod`, `backend/go.sum`, `backend/internal/config/config.go`, `backend/internal/config/config_test.go`, `Makefile`
- Create: `backend/internal/delivery/types.go`, `backend/internal/delivery/queue.go`, `backend/internal/delivery/queue_test.go`, `backend/internal/delivery/deliverytest/redis.go`

**Interfaces:**
- Produces:

```go
package config
type WorkerConfig struct {
    HTTPAddr, DatabaseURL, RedisURL, ProviderURL string
    RatePerSec, MaxAttempts, Consumers           int
    Timeout                                      time.Duration
}
func LoadWorker(getenv func(string) string) (WorkerConfig, error)
// Config gets RedisURL (REDIS_URL, required).

package delivery
type Channel string
const (ChannelSMS Channel = "sms"; ChannelEmail Channel = "email")
var Channels = []Channel{ChannelSMS, ChannelEmail}

type Status string
const (StatusQueued Status = "queued"; StatusSending = "sending"; StatusRetrying = "retrying";
       StatusSent = "sent"; StatusFailed = "failed"; StatusUnknown = "unknown"; StatusSkipped = "skipped")

type Entry struct { StreamID string; JobID int64 }

type Queue struct { /* rdb *redis.Client, consumer string */ }
func NewQueue(rdb *redis.Client, consumer string) *Queue
func (q *Queue) Ensure(ctx context.Context) error                          // XGROUP CREATE ... $ MKSTREAM, BUSYGROUP is ok
func (q *Queue) Push(ctx context.Context, ch Channel, ids []int64) error   // pipelined XADD MAXLEN ~ 100000
func (q *Queue) Read(ctx context.Context, ch Channel, n int, block time.Duration) ([]Entry, error)
func (q *Queue) Ack(ctx context.Context, ch Channel, streamID string) error
func (q *Queue) AutoClaim(ctx context.Context, ch Channel, minIdle time.Duration, n int) ([]Entry, error)
func IsNoGroup(err error) bool

package deliverytest
// NewRedis returns a client on TEST_REDIS_URL and flushes its DB. It skips the
// test when TEST_REDIS_URL is empty.
func NewRedis(t testing.TB) *redis.Client
```

`LoadWorker` checks: `DATABASE_URL`, `REDIS_URL`, `PROVIDER_URL` required. `SEND_RATE_PER_SEC` 1..100. `SEND_MAX_ATTEMPTS` 1..10. `SEND_CONSUMERS` 1..256. `SEND_TIMEOUT` a Go duration above 0. It joins all errors, as `Load` does.

Makefile `test`: start `cc-test-redis` (`redis:8-alpine`) next to `cc-test-pg`, on the `cc-test` network and on `127.0.0.1:56379`. Export `TEST_REDIS_URL` (`redis://cc-test-redis:6379/0` in the container branch, `redis://127.0.0.1:56379/0` in the local branch). Pass `-e TEST_REDIS_URL` to `GO_TEST`. Remove both containers after the tests. Wait for `redis-cli ping` (NFR-6).

Tests with Redis must not run in parallel across packages on the same DB. Use DB 0 for `delivery`, DB 1 for `httpapi`, DB 2 for `cmd/worker`. `NewRedis` takes the DB number from an optional argument.

- [ ] **Step 1: Write the failing tests**
  - `config_test.go` `TestLoadWorker_Defaults`: with only the three URLs, the defaults of Global Constraints.
  - `TestLoadWorker_Errors`: table with missing URLs, rate 0 and 101, attempts 11, timeout `abc` and `0s`. All errors come back in one call.
  - `TestLoad_RedisURLRequired`.
  - `queue_test.go` `TestQueue_PushReadAck`: push ids 1, 2, 3 on `sms`. Read gives them in order with the field `job`. After `Ack`, `XPENDING` is 0. Nothing is on `send:email`.
  - `TestQueue_EnsureTwice`: two calls give no error.
  - `TestQueue_AutoClaim`: consumer A reads and does not ack. Consumer B `AutoClaim` with `minIdle = 0` gets the entry.
  - `TestQueue_NoGroup`: after `DEL send:sms`, `Read` returns an error with `IsNoGroup(err) == true`.
- [ ] **Step 2: Run** `make test`. Expected: the new tests fail. Earlier api tests that build `config.Config` without `REDIS_URL` fail too. Fix those fixtures in Step 3.
- [ ] **Step 3: Add `go-redis/v9`, write the config and the queue.**
- [ ] **Step 4: Run** `make test`. Expected: PASS. Run it again with Go removed from `PATH` (`PATH=/usr/bin:/bin:/usr/local/bin make test`). Expected: PASS, and `docker ps -a` shows no `cc-test-*` container.

---

### Task 3: Delivery store

**Files:**
- Create: `backend/internal/delivery/store.go`, `backend/internal/delivery/store_test.go`
- Modify: `backend/internal/delivery/types.go`

**Interfaces:**
- Consumes: `segment.ReachableQuery`, `segment.RecipientByID`, `segment.ChannelByName` (Task 1). `campaign` tables.
- Produces:

```go
package delivery
type Job struct {
    ID, CampaignID, UserID int64
    Channel                Channel
    Language               string
    Status                 Status
    Attempts               int
    LastError              string
    Retryable              bool
    UpdatedAt              time.Time
}

// Result is the new state after an attempt. NextAttemptAt is set for StatusRetrying only.
type Result struct {
    Status        Status
    LastError     string
    Retryable     bool
    ProviderID    string
    NextAttemptAt *time.Time
}

type Message struct { To, Subject, Body string }

type ChannelProgress struct {
    Total, Queued, Sending, Retrying, Sent, Failed, Unknown, Skipped, RetryableFailed int64
} // json tags: total, queued, sending, retrying, sent, failed, unknown, skipped, retryable_failed

type Progress struct {
    CampaignID int64                       `json:"campaign_id"`
    Status     string                      `json:"status"`
    Channels   map[Channel]ChannelProgress `json:"channels"` // always both keys
}

type JobFilter struct { Status Status; Channel Channel; Limit int; After int64 }

var (ErrNotFound, ErrNotSendable, ErrNotCompleted, ErrNothingToRetry error)

type Store struct{ /* pool */ }
func NewStore(pool *pgxpool.Pool) *Store

// PlanNext locks one approved campaign (FOR UPDATE SKIP LOCKED), inserts its
// send jobs, and sets it queued, or completed when it has none. ok is false
// when no campaign waits. ids are the new queued send job ids per channel.
func (s *Store) PlanNext(ctx context.Context, allow segment.Allowlist, now time.Time) (campaignID int64, ids map[Channel][]int64, ok bool, err error)
func (s *Store) MarkEnqueued(ctx context.Context, ids []int64, at time.Time) error

// Claim moves a queued or due retrying job to sending and adds 1 to attempts.
// It also moves the campaign from queued to sending. ok is false when no row changed.
func (s *Store) Claim(ctx context.Context, id int64, now time.Time) (Job, bool, error)
// Release moves a sending job back to queued and subtracts 1 from attempts.
func (s *Store) Release(ctx context.Context, id int64) error
func (s *Store) Finish(ctx context.Context, id int64, r Result, now time.Time) error

// Load reads the user (with consent) and the content of the job language,
// renders it, and returns the message. skip is "" or one of the skip reasons.
func (s *Store) Load(ctx context.Context, j Job) (m Message, skip string, err error)

func (s *Store) DueRetries(ctx context.Context, now time.Time, lateBy time.Duration) (map[Channel][]int64, error)
func (s *Store) StaleQueued(ctx context.Context, now time.Time, age time.Duration) (map[Channel][]int64, error)
func (s *Store) ExpireSending(ctx context.Context, now time.Time, age time.Duration) (int64, error)
func (s *Store) CompleteCampaigns(ctx context.Context) ([]int64, error)

func (s *Store) Progress(ctx context.Context, campaignID int64) (Progress, error)
func (s *Store) ListJobs(ctx context.Context, campaignID int64, f JobFilter) (jobs []Job, nextAfter *int64, err error)
// Retry resets the retryable failed jobs of a completed campaign. It returns their ids per channel.
func (s *Store) Retry(ctx context.Context, campaignID int64, now time.Time) (map[Channel][]int64, error)
```

Points the signatures do not fix:
- `PlanNext` parses `campaigns.rule` with `allow.ParseRule` and compiles it with `now`. Languages = the `language` values of `campaign_contents`. For each channel: `INSERT INTO send_jobs (campaign_id, user_id, channel, language, status, created_at, updated_at) SELECT $n, id, '<ch>', language, 'queued', $n+1, $n+1 FROM (<ReachableQuery>) r ON CONFLICT DO NOTHING RETURNING id`. If the rule fails to parse, log it and set `completed` with no jobs.
- `Claim` uses the conditional `UPDATE` of the spec, `RETURNING` the job columns. Then `UPDATE campaigns SET status = 'sending', updated_at = $now WHERE id = $cid AND status = 'queued'`. One transaction.
- `Finish` writes `status`, `last_error`, `retryable`, `next_attempt_at`, `updated_at`, and for `sent` also `provider_id` and `sent_at`. `WHERE id = $1 AND status = 'sending'`. Zero rows changed is not an error.
- `Load` uses `segment.RecipientByID`. No consent gives `no_consent`. An empty phone (SMS) or email (email) gives `no_contact`. A rendered SMS over 160 runes or subject over 80 runes gives `over_limit`. The body for email is the rendered `email_body`.
- `CompleteCampaigns`: `UPDATE campaigns SET status='completed' WHERE status IN ('queued','sending') AND NOT EXISTS (SELECT 1 FROM send_jobs WHERE campaign_id = campaigns.id AND status IN ('queued','sending','retrying')) RETURNING id`.
- `Progress` uses one `GROUP BY channel, status` and one count of `status='failed' AND retryable`. A draft gives `ErrNotSendable`. An unknown id gives `ErrNotFound`.
- `ListJobs`: `ORDER BY id`, `id > After`, `LIMIT Limit+1` to set `nextAfter`.
- `Retry` locks the campaign `FOR UPDATE`. Not `completed` gives `ErrNotCompleted`. Zero retryable rows gives `ErrNothingToRetry`. It sets those rows to `queued`, `attempts = 0`, `next_attempt_at = NULL`, `enqueued_at = NULL`, `retryable = false`, and the campaign to `sending`.
- Each method uses a context with the `2s` statement timeout of slice 04.

- [ ] **Step 1: Write the failing tests** (Postgres, seed users with `seed.Load`, an approved campaign made with the `campaign` store)
  - `TestStore_PlanExample` (FR-4, FR-8): the counts per channel equal `CountAudience` for the example rule, for the languages of the campaign. The campaign is `queued`.
  - `TestStore_PlanSkippedLanguage` (FR-6).
  - `TestStore_PlanConcurrent` (FR-7): two `PlanNext` calls in goroutines. One gets `ok = true`. The table has one set of rows.
  - `TestStore_PlanNoJobs` (FR-9, Review Focus 4): a rule that matches no user with consent. The campaign is `completed`, `ids` is empty.
  - `TestStore_ClaimOnce` (FR-13): two `Claim` calls on one id in goroutines. Exactly one has `ok = true`. `attempts = 1`. The campaign is `sending`.
  - `TestStore_ClaimNotDue`: a `retrying` job with `next_attempt_at` in the future gives `ok = false`.
  - `TestStore_Release`: after `Claim` and `Release`, the job is `queued` with `attempts = 0`.
  - `TestStore_FinishSent`: `provider_id`, `sent_at` set.
  - `TestStore_LoadSkips` (FR-16, FR-17, Review Focus 5): table with consent set to false, empty phone, and a first name of 200 runes set by SQL. The skip reasons match.
  - `TestStore_LoadRenders` (FR-15): the SMS has the first name in place of `{{first_name}}`.
  - `TestStore_Recovery` (FR-22 to FR-24, FR-26): `StaleQueued` returns a job with `enqueued_at` null and one older than 120 s, not one of 10 s. `DueRetries` with `lateBy = 0` returns a due job. `ExpireSending` sets `unknown` on a 40 s old `sending` job with age 33 s. `CompleteCampaigns` completes a campaign whose jobs are all terminal.
  - `TestStore_Progress` (FR-30): counts equal a `GROUP BY` in the test. A draft gives `ErrNotSendable`.
  - `TestStore_ListJobs` (FR-33): filters and pages. `nextAfter` is nil on the last page.
  - `TestStore_Retry` (FR-34, FR-35, FR-36): a `503` failed job becomes `queued` with `attempts = 0`. A `400` failed job and an `unknown` job do not change. A second concurrent `Retry` gets `ErrNotCompleted`. A campaign with nothing to retry gets `ErrNothingToRetry`.
- [ ] **Step 2: Run** `make test`. Expected: FAIL (undefined).
- [ ] **Step 3: Write `store.go`.**
- [ ] **Step 4: Run** `make test`. Expected: PASS.

---

### Task 4: Provider client, policy, limiter and consumer

**Files:**
- Create: `backend/internal/delivery/provider.go`, `provider_test.go`, `policy.go`, `policy_test.go`, `limiter.go`, `limiter_test.go`, `consumer.go`, `consumer_test.go`

**Interfaces:**
- Consumes: `Store` (Task 3), `Queue` (Task 2).
- Produces:

```go
package delivery
type Clock interface {
    Now() time.Time
    Sleep(ctx context.Context, d time.Duration) error // returns ctx.Err() when cancelled
}
func RealClock() Clock

type OutcomeKind int
const (Delivered OutcomeKind = iota; Retryable; Permanent; Ambiguous)
type Outcome struct {
    Kind       OutcomeKind
    ProviderID string
    RetryAfter time.Duration // from 429
    Err        string        // e.g. "503 unavailable", "timeout", "connection refused"
}

type Provider struct { BaseURL string; HTTP *http.Client }
// NewProvider sets HTTP.Timeout = timeout and a Transport with no proxy.
func NewProvider(baseURL string, timeout time.Duration) *Provider
func (p *Provider) Send(ctx context.Context, ch Channel, m Message, reference string) Outcome

// Next maps an outcome to the new state. attempts is the count after this attempt.
func Next(o Outcome, attempts, maxAttempts int, now time.Time, jitter func() float64) Result
func Backoff(attempts int, jitter func() float64) time.Duration // attempts 1..4 -> 1,2,4,8 s * (1+0.2*jitter())

type Limiter struct { /* rate.Limiter, pausedUntil, clock */ }
func NewLimiter(perSec int, clock Clock) *Limiter
func (l *Limiter) Wait(ctx context.Context) error
func (l *Limiter) Pause(d time.Duration)

type Consumer struct {
    Store    *Store
    Queue    *Queue
    Provider *Provider
    Limiter  *Limiter
    Channel  Channel
    Clock    Clock
    MaxAttempts int
    Jitter   func() float64
    Log      *slog.Logger
}
// Handle runs one entry. It returns an error only when the entry must stay unacknowledged.
func (c *Consumer) Handle(ctx context.Context, e Entry) error
```

Classification in `Send`:
- Use `httptrace.ClientTrace.WroteRequest` to record whether the request was written.
- `202` → `Delivered` with the `id` of the body. `429` → `Retryable` with `Retry-After` (seconds, minimum 1 s). `5xx` → `Retryable`. Other `4xx` → `Permanent`.
- A transport error before the request was written (connection refused, DNS) → `Retryable`. A transport error or a timeout after the request was written → `Ambiguous`.
- A body that is not valid JSON on a `202` is still `Delivered` with an empty `ProviderID`.
- SMS body: `{"to","body","reference"}`. Email body: `{"to","subject","body","reference"}`.

`Next`:
- `Delivered` → `sent`. `Permanent` → `failed`, `retryable = false`. `Ambiguous` → `unknown`.
- `Retryable` with `attempts < maxAttempts` → `retrying`. For a `429`, `next_attempt_at = now + RetryAfter`. For the rest, `next_attempt_at = now + Backoff(attempts)`.
- `Retryable` at `maxAttempts` → `failed`, `retryable = true`.
- `LastError` is `o.Err` in each case.

`Limiter.Wait`: wait until `pausedUntil`, then `ReserveN(clock.Now(), 1)` and `clock.Sleep(r.DelayFrom(now))`. Burst 1.

`Handle` order (spec "Sender"):
1. `Limiter.Wait`.
2. `Store.Claim`. If not ok, `Ack` and return nil.
3. `Store.Load`. On a skip, `Finish` with `skipped`, then `Ack`.
4. `Provider.Send` with `reference`. On a `429`, also call `Limiter.Pause(RetryAfter)`.
5. `Finish` with `Next(...)`, then `Ack`.

Two unexported hooks on `Consumer` serve the tests: `afterClaim func()` and `finishErr func() error`. Both are nil in production.
- If `ctx` is cancelled after the claim and before `Send`, call `Release` with a background context and return `ctx.Err()`.
- If `Finish` fails, log it and return the error. Do not `Ack`.
- Log on each non-`Delivered` outcome: `job_id`, `channel`, `attempt`, `error`. Nothing else about the user.

- [ ] **Step 1: Write the failing tests**
  - `provider_test.go` `TestProvider_Outcomes` (FR-11 base): `httptest` with the real `mockprovider.New` handler and `X-Mock-Fail` set by a test `RoundTripper` for `none`, `429`, `500`, `503`, `slow`, `lost`. Options: `SlowDelay 200ms`, `LatencyMax 0`. Provider timeout `50ms`. Expected kinds: Delivered, Retryable (`RetryAfter 1s`), Retryable, Retryable, Ambiguous, Ambiguous. A `400` from an empty body → Permanent.
  - `TestProvider_ConnectionRefused`: a closed port → Retryable.
  - `TestProvider_OneRequestPerAttempt` (FR-18): `lost` → the provider stats show `requests = 1`.
  - `TestProvider_RequestShape` (FR-14): a recording handler sees `reference` `"7:42:sms"` and no `Idempotency-Key` header.
  - `policy_test.go` `TestNext` table (FR-11, FR-12): each row of the spec "Retry strategy" table. `503` at attempt 5 of 5 → `failed`, `retryable = true`. `Backoff(1..4)` with `jitter = 0` → 1, 2, 4, 8 s. With `jitter = 1` → 1.2, 2.4, 4.8, 9.6 s.
  - `limiter_test.go` `TestLimiter_Rate` (FR-19): fake clock, 1,000 `Wait` calls. Record the fake times. No 1-second window has more than 100.
  - `TestLimiter_Pause` (FR-21): after `Pause(2s)`, the next `Wait` returns at `now + 2s` or later.
  - `consumer_test.go` (Postgres + Redis + `mockprovider.New`):
    - `TestConsumer_Sequences` (FR-10, FR-11): for `429→202`, `500→202`, `503×5`, `400`, `slow`, `lost`, connection refused`→202`, run `Handle` until the job is terminal. For a retry, set the fake clock to `next_attempt_at` and call `Handle` again. Check the status, `attempts` and `last_error` of the spec table.
    - `TestConsumer_DoubleEntry` (FR-13): push one id two times. Two `Handle` calls in goroutines. The provider stats show `requests = 1`.
    - `TestConsumer_NoContact` (Review Focus 5): provider `requests = 0`, status `skipped`, `last_error = no_contact`.
    - `TestConsumer_ReleaseOnCancel` (FR-24a): `afterClaim` cancels `ctx`. The job is `queued`, `attempts = 0`. Provider `requests = 0`.
    - `TestConsumer_CancelInWait`: cancel `ctx` during `Limiter.Wait`. No claim happens. The job stays `queued`.
    - `TestConsumer_FinishFailsNoResend` (Review Focus 2): `finishErr` fails one time. `Handle` returns an error and does not `Ack`. A second `Handle` on the `AutoClaim` entry makes no request. Provider `requests = 1`.
    - `TestConsumer_LogsNoPersonalData` (NFR-2): capture the `slog` output of the sequence test. It has no first name, no phone, no email and no SMS text of the test users.
- [ ] **Step 2: Run** `make test`. Expected: FAIL (undefined).
- [ ] **Step 3: Write the four files.**
- [ ] **Step 4: Run** `make test` and `cd backend && go test -race -count=3 ./internal/delivery/`. Expected: PASS.

---

### Task 5: Worker loops and `cmd/worker`

**Files:**
- Create: `backend/internal/delivery/worker.go`, `worker_test.go`, `backend/cmd/worker/main.go`, `backend/cmd/worker/main_test.go`
- Modify: `docker-compose.yml`, `Makefile` (`up`), `Dockerfile` (comment), `.env.example`

**Interfaces:**
- Consumes: everything of Tasks 2 to 4.
- Produces:

```go
package delivery
type Worker struct {
    Store       *Store
    Queue       *Queue
    Provider    *Provider
    Allow       segment.Allowlist
    Clock       Clock
    Cfg         config.WorkerConfig
    Log         *slog.Logger
    // Intervals; zero means the default of Global Constraints. Tests set short ones.
    PlanEvery, RetryEvery, RecoverEvery, ClaimIdle time.Duration
}
// Run blocks until ctx ends, then shuts down as in the spec "Worker process".
func (w *Worker) Run(ctx context.Context) error
```

`Run`:
- `Queue.Ensure` at start.
- Planner loop: `PlanNext` until `ok = false`, then `Push` and `MarkEnqueued`. Log campaign id and counts per channel.
- For each channel, `Cfg.Consumers` goroutines share one `Limiter`. Each loops `Read(n=1, block=1s)` → `Handle`. On `IsNoGroup`, call `Ensure` and continue.
- Retry poller: `DueRetries(now, 0)` → `Push` → `MarkEnqueued`.
- Recovery: `StaleQueued(120s)` → `Push`, `MarkEnqueued`. `DueRetries(now, 30s)` → `Push`. `ExpireSending(Timeout + 30s)`. `AutoClaim(60s)` → `Handle`. `CompleteCampaigns` → log completion with counts per status.
- Shutdown: cancel the read and wait contexts, give the running `Send` calls their own context with 10 s grace, then return.

`cmd/worker/main.go`: `config.LoadWorker`, pgx pool, `redis.ParseURL`, `NewQueue(rdb, hostname+pid)`, `NewProvider`, `Worker.Run`. An `http.Server` on `Cfg.HTTPAddr` with `GET /healthz`: `{"status":"ok","db":"ok","redis":"ok"}` with `200`, or `503` with the failed part `"down"`. Each check has a 1 s timeout.

Compose service `worker`: `go-service` target with `CMD: worker`. No `env_file`. `DATABASE_URL`, `REDIS_URL: redis://redis:6379/0`, `PROVIDER_URL: http://mock-provider:8081`, `SEND_*` with `${VAR:-default}`. Ports `8082:8082`. `depends_on`: `seed` completed, `redis` healthy, `mock-provider` started. The api gets `REDIS_URL` and `depends_on: redis: service_healthy`. `make up` adds `curl ... localhost:8082/healthz`. `.env.example` lists the `SEND_*` vars with a comment.

- [ ] **Step 1: Write the failing tests**
  - `TestWorker_EndToEnd` (FR-28, FR-29):
    - Seed the users and approve the example campaign.
    - Provider: `mockprovider.New` with the default rates, `SlowDelay 200ms`, `LatencyMax 5ms` and a fixed `Random`.
    - Worker: real clock, `Timeout 50ms`, `RatePerSec 100`, `Consumers 32`, `PlanEvery 50ms`, `RetryEvery 50ms`, `RecoverEvery 200ms`. The real backoff stays.
    - Wait until the campaign is `completed`, with a 90 s limit. Then all jobs are terminal. `GET /v1/stats` `duplicate_deliveries = 0` for both channels. Each `sent` job has exactly one provider message with its `reference`. No `failed` job has one.
  - `TestWorker_RestartBeforePush` (FR-27): call `PlanNext` directly and do not push. Start `Run`. All jobs are sent, each `reference` one time.
  - `TestWorker_StreamLost` (Review Focus 1): during a run, `DEL send:sms send:email`. The run still completes. `duplicate_deliveries = 0`.
  - `TestWorker_AutoClaimDeadConsumer` (FR-25): consumer `dead` reads one entry and stops. `Run` with `ClaimIdle 100ms` sends it.
  - `TestWorker_ChannelsIndependent` (FR-20): an SMS provider path that answers after 300 ms through a wrapping handler. With a fake clock limiter, email still reaches its rate. Check by counting email attempts in 1 s of real time with `RatePerSec 20` to keep the test short.
  - `TestWorker_RetryPollerOnTime` (FR-12): a `retrying` job with `next_attempt_at = now + 300ms` is on the stream before `now + 1.3s`.
  - `TestCompose_WorkerNoKey` (NFR-1) is a Step 4 check: `ANTHROPIC_API_KEY=x docker compose config` shows no `ANTHROPIC_API_KEY` under `worker`.
  - `cmd/worker/main_test.go` `TestHealthz`: with a dead Redis address, `503` and `"redis":"down"`.
- [ ] **Step 2: Run** `make test`. Expected: FAIL.
- [ ] **Step 3: Write `worker.go`, `cmd/worker`, compose, Makefile, `.env.example`.**
- [ ] **Step 4: Run** `make test`, `go test -race ./internal/delivery/ ./cmd/worker/`, `docker compose config -q`. Expected: PASS.

---

### Task 6: Delivery endpoints and api wiring

**Files:**
- Create: `backend/internal/delivery/service.go`, `service_test.go`, `backend/internal/httpapi/delivery.go`, `backend/internal/httpapi/delivery_test.go`
- Modify: `backend/cmd/api/main.go`, `backend/internal/httpapi/cors.go` (only if a header is missing)

**Interfaces:**
- Consumes: `Store.Progress`, `Store.ListJobs`, `Store.Retry`, `Queue.Push`, `Store.MarkEnqueued`.
- Produces:

```go
package delivery
type Service struct { Store *Store; Queue *Queue; Now func() time.Time; Log *slog.Logger }
func (s *Service) Progress(ctx context.Context, id int64) (Progress, error)
func (s *Service) ListJobs(ctx context.Context, id int64, f JobFilter) ([]Job, *int64, error)
// Retry resets the jobs, then pushes them. A push error is logged, not returned:
// the recovery pushes them later.
func (s *Service) Retry(ctx context.Context, id int64) (retried int, p Progress, err error)

package httpapi
type DeliveryService interface {
    Progress(ctx context.Context, id int64) (delivery.Progress, error)
    ListJobs(ctx context.Context, id int64, f delivery.JobFilter) ([]delivery.Job, *int64, error)
    Retry(ctx context.Context, id int64) (int, delivery.Progress, error)
}
// RegisterDelivery adds the four routes. poll is the stream check interval (1 s in main).
func RegisterDelivery(mux *http.ServeMux, svc DeliveryService, poll time.Duration)
```

- Routes: `GET /api/campaigns/{id}/progress`, `GET /api/campaigns/{id}/progress/stream`, `GET /api/campaigns/{id}/send-jobs`, `POST /api/campaigns/{id}/retry`. Reuse `campaignID` and `writeError`.
- Error map: `ErrNotFound` 404 `not_found`, `ErrNotSendable` 409 `campaign_not_sendable`, `ErrNotCompleted` 409 `campaign_not_completed`, `ErrNothingToRetry` 409 `nothing_to_retry`, other 503 `db_unavailable`.
- The stream: call `Progress` before `newSSEWriter`, so that a `404` or `409` is a normal JSON error. Send `progress` at once. Each `poll`, call `Progress` and send only when it differs from the last one. On `completed`, send `progress` and `done`, then return. Stop after 30 minutes or on client disconnect.
- `send-jobs` query: `status` must be one of the seven, `channel` `sms` or `email`, `limit` 1..200 (default 50), `after` a positive integer. Response keys per job exactly: `id`, `user_id`, `channel`, `language`, `status`, `attempts`, `last_error`, `updated_at`. Top level `send_jobs`, `next_after`.
- Retry response: `{"retried": n, "progress": {...}}`.
- `cmd/api`: `config.Load` now has `RedisURL`. Create the Redis client, `delivery.NewQueue(rdb, "api")` and the service. Register with `poll = time.Second`.

- [ ] **Step 1: Write the failing tests**
  - `delivery_test.go` with a fake service:
    - `TestProgress_Shape` (FR-30).
    - `TestDelivery_BadID` (FR-31): `abc`, `0` and `-1` give `400`.
    - `TestDelivery_ErrorMapping`: one case for each error.
    - `TestSendJobs_Query`: a bad `status`, a bad `channel`, `limit=0`, `limit=201` and `after=-1` give `400`. Good values reach the fake.
    - `TestSendJobs_Keys` (FR-33, NFR-3): the exact key set.
    - `TestRetry_Response`.
  - `TestProgressStream` (FR-32): the fake returns `sending` two times with the same counts, then a change, then `completed`. With `poll = 10ms`, the stream has `progress`, `progress`, `progress`, `done`: no event for the repeat. A draft returns `409` with a JSON body.
  - `service_test.go` `TestService_RetryPushes` (Postgres + Redis): after `Retry`, the ids are on the stream of their channel and `enqueued_at` is set.
  - `TestService_RetryPushFails`: a closed Redis client. `Retry` returns `n` and no error. The jobs are `queued` with `enqueued_at` null.
- [ ] **Step 2: Run** `make test`. Expected: FAIL.
- [ ] **Step 3: Write the service, the handlers and the wiring.**
- [ ] **Step 4: Run** `make test`. Expected: PASS.

---

### Task 7: Frontend

**Files:**
- Create: `frontend/src/lib/delivery.ts`, `frontend/src/lib/__tests__/delivery.test.ts`, `frontend/src/components/campaign/sending-dialog.vue`, `frontend/src/components/campaign/delivery-section.vue`, `frontend/src/components/campaign/__tests__/sending-dialog.test.ts`, `frontend/src/components/campaign/__tests__/delivery-section.test.ts`
- Modify: `frontend/src/lib/campaigns.ts` (status type), `frontend/src/pages/campaigns/[id].vue`, `frontend/src/pages/campaigns/index.vue`, the page that handles `approved` from `review-dialog.vue` (`new.vue`, `index.vue`, `[id].vue`), `frontend/src/assets/index.css`, the page tests in `frontend/src/pages/campaigns/__tests__/`

**Interfaces:**
- Consumes: the four endpoints of Task 6.
- Produces:

```ts
// lib/campaigns.ts
export type CampaignStatus = 'draft' | 'approved' | 'queued' | 'sending' | 'completed'

// lib/delivery.ts
export type DeliveryStatus = 'queued' | 'sending' | 'retrying' | 'sent' | 'failed' | 'unknown' | 'skipped'
export interface ChannelProgress { total: number, queued: number, sending: number, retrying: number, sent: number, failed: number, unknown: number, skipped: number, retryable_failed: number }
export interface Progress { campaign_id: number, status: CampaignStatus, channels: { sms: ChannelProgress, email: ChannelProgress } }
export interface SendJob { id: number, user_id: number, channel: 'sms' | 'email', language: string, status: DeliveryStatus, attempts: number, last_error: string, updated_at: string }
export function getProgress(id: number): Promise<Progress>
// Opens the SSE stream. Falls back to getProgress every 2 s on a stream error. Returns a stop function.
export function watchProgress(id: number, on: (p: Progress) => void): () => void
export function listSendJobs(id: number, q: { status?: DeliveryStatus, channel?: 'sms' | 'email', after?: number, limit?: number }): Promise<{ send_jobs: SendJob[], next_after: number | null }>
export function retryCampaign(id: number): Promise<{ retried: number, progress: Progress }>
export type Step = 'approved' | 'queued' | 'sending' | 'done'
export function currentStep(status: CampaignStatus): Step   // approved→approved, queued→queued, sending→sending, completed→done
```

- `sending-dialog.vue` props `{ campaignId: number, open: boolean }`, emits `update:open` and `open-list`. It follows option A of the mockup and the spec "Sending popup" copy exactly: title `Sending campaign #{id}`, text `You can close this. Sending continues in the background.`, done text `Done. {sent} sent, {failed} failed, {unknown} unknown.`, buttons `Close` and `Open delivery list`. Bar segment widths are `count / total * 100%`. It calls `watchProgress` while open and stops it on close.
- `delivery-section.vue` props `{ campaignId: number, progress: Progress }`, emits `view-progress` and `retried`. It has:
  - the summary for each channel
  - "View progress" for `approved`, `queued` and `sending`
  - "Retry N failed" with an `AlertDialog`: text `Send {n} failed messages again? Messages with an unknown result are not sent again.`, buttons `Back` and `Send again`
  - the status chips with counts, the channel filter and the table
  - `may be delivered` under `Unknown`, and `Load more`
  - Rows become cards below 720 px.
- `[id].vue`: remove the line `Sending starts in a later step.`. For a campaign that is not a draft, load the progress and show `delivery-section`. Open `sending-dialog` when the route query has `sending=1`, and remove the query after open. "Open delivery list" scrolls to the section with `scrollIntoView`.
- After `approved` from `review-dialog.vue`, every page routes to `/campaigns/{id}?sending=1` (FR-37).
- `index.vue`: badges Draft, Approved, Queued, Sending, Completed. Only a draft opens the review window.
- Tokens: `--success` and `--info` in `:root` and `.dark`, `--color-success` and `--color-info` in `@theme inline`.

- [ ] **Step 1: Write the failing tests** (Vitest, a mock `EventSource` class on `globalThis`, `fetch` mocked as in the slice 06 tests)
  - `delivery.test.ts`: `currentStep` for each status (FR-38). `watchProgress` calls `on` for each `progress` event. After an `error` event it calls `getProgress` every 2 s (fake timers) (FR-40). The stop function closes the stream and the timer.
  - `sending-dialog.test.ts`: `shows Approved before the plan` (Review Focus 3: status `approved`, totals 0 → step Approved current, no Done text). `marks the steps` for `queued`, `sending`, `completed` (FR-38). `bar widths` for sent 50, retrying 10, failed 20, unknown 20 of 100 → `50%`, `10%`, `20%`, `20%` (FR-39). `hides a channel with no recipients` (Review Focus 4). `done text` at `completed` (FR-41). `Open delivery list` emits `open-list` and closes.
  - `delivery-section.test.ts`: `View progress` shows for `approved`, `queued`, `sending` only (FR-42). `Retry` shows only for `completed` with `retryable_failed > 0`. It opens the confirm first. `Send again` calls `retryCampaign` and emits `retried` (FR-43). Table row text `User #4012` and `may be delivered` for `unknown`. Chips call `listSendJobs` with the status. `Load more` passes `after = next_after` (FR-44). The DOM has no `@` and no `+90`.
  - `pages/campaigns/__tests__/id.test.ts`: `sending=1` opens the dialog. The old line is gone.
  - `pages/campaigns/__tests__/index.test.ts`: the five badges (FR-45).
  - A token test, like the slice 06 `--warning` test: `--success` and `--info` exist in both themes (FR-46).
- [ ] **Step 2: Run** `pnpm -C frontend test:run`. Expected: the new tests fail.
- [ ] **Step 3: Write the components and the page changes.**
- [ ] **Step 4: Run** `pnpm -C frontend test:run`, `pnpm -C frontend build`, `CI=true pnpm -C frontend lint`. Expected: PASS.

---

### Task 8: Docs and the end-to-end check

**Files:**
- Modify: `docs/architecture.md`, `docs/decisions-log.md`, `tasks/backlog.json`, `tasks/progress.md`, `.claude/skills/ste-writing/references/terms.md` (only if a term is missing)

- [ ] **Step 1: Docs.**
  - `docs/architecture.md`: add `worker` to Components and `cmd/worker` to the tree. Replace Flow step 4 with the planner, the streams, the claim, the retry strategy, the recovery and the retry endpoint. Keep the D-018 and D-019 lines.
  - `docs/decisions-log.md`: add D-046 onward in the existing format (Context, Options, Choice, Why):
    - one entry for each line of the spec "Calls by Claude"
    - `go-redis/v9`
    - `httptrace` to tell "not sent" from "maybe sent"
    - the `--success` and `--info` tokens
  - `tasks/backlog.json`: the `07-send-queue` `acceptance` gets the FR and NFR lines of `requirements.md`, as for slice 06.
  - Run `python3 .claude/skills/ste-writing/scripts/ste_check.py docs/architecture.md docs/decisions-log.md`. Fix each finding.
- [ ] **Step 2: Run** `make verify`. Expected: exit 0.
- [ ] **Step 3: Compose check.** If the running stack holds the ports, use `-p cc-verify` with port overrides, as in slice 06.
  - `docker compose up -d --build --wait` with no `.env` and `LLM_MODE=mock`. `curl localhost:8082/healthz` gives `{"status":"ok","db":"ok","redis":"ok"}`.
  - `curl -X DELETE localhost:8081/v1/messages`.
  - Make a campaign with the example brief through the api (preview, generate, approve). Poll `GET /api/campaigns/{id}/progress` until `completed`.
  - `curl localhost:8081/v1/stats`: `duplicate_deliveries` is 0 for both channels. Write down `rate_limited` and set the acceptance number in `backlog.json` (for example, under 2% of `requests`).
  - `POST /api/campaigns/{id}/retry`: `200` when `retryable_failed > 0`, else `409 nothing_to_retry`.
- [ ] **Step 4: Headless Chrome** at 1280 px and 360 px. Go from `/campaigns/new` to "Send now". Check the popup with the four steps, Done, the list with filters and the retry button. Check that 360 px has no horizontal scroll (NFR-5).
- [ ] **Step 5: Progress.** Append the session entry to `tasks/progress.md`. Do not set `passes`. The user does that after review.
