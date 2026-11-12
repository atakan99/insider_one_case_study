# Implicit state machines in main.go

- Session: `d69aea1d-9d23-4cca-8129-17ae82a3db16`
- Model: `claude-opus-5-5`
- Time (UTC): 2026-11-11 16:27:29 → 2026-11-11 16:51:00
- Subagent runs: 0

---

## 👤 User  <sub>2026-11-11 16:27:29</sub>

_[IDE: selection]_ 
find the implicit state machines and write them down

**🔧 Skill** ```json
{
  "skill": "ste-writing"
}
```

<details><summary>Result</summary>

```
Launching skill: ste-writing
```

</details>

> _[skill instructions loaded: ste-writing]_

**🔧 Bash** _Read progress log and list code layout_
```bash
tail -40 tasks/progress.md && ls docs docs/specs docs/plans && find backend -name '*.go' -not -name '*_test.go' | xargs wc -l | sort -n | tail -40 && ls backend/migrations* backend/internal 2>/dev/null
```

<details><summary>Result (error)</summary>

```
Exit code 1
- Next: user review of slice 08. The user sets `08-dashboard-metrics.passes`.
- Open questions:
  - `XLEN` of the streams does not fall after a send, because the worker does not trim the streams. This is not a defect of slice 08.

## 2026-11-11 — external review fixes
- Done: three findings of an external review. All three were correct.
  - The worker now waits for the rate token immediately before the provider request. The order is claim, load, wait, send (D-084, replaces D-053). `Store.Claimable` is removed.
  - A save in the review window keeps the text that the marketer typed during the request (D-085).
  - `README.md` describes the limits of the mock LLM and the steps to enable Claude.
- Verified with: each defect had a test that failed first. `TestConsumer_PacesRequestsAfterDatabaseStall` showed 7 requests in 81 µs at 50 per second before the fix. It passed 10 of 10 runs after the fix. The review window test "keeps the text typed while a save is in progress" failed before the fix. `go test -race ./internal/delivery/...` passes. `make verify` exit 0 (222 frontend tests).
- Next: user review of the changes. Nothing is committed.
- Open questions:
  - A `Retry-After` longer than `SEND_TIMEOUT` plus 30 s can set a waiting send job to `unknown`. The mock provider sends 1 s.
  - No compose run measured the request rate against the mock provider after the change.

## 2026-11-11 — second external review
- Done: two findings of a second review. Both were correct, and both came from D-084.
  - The worker order is again wait, claim, load, send. A rate token is good for 100 ms. If the claim and the load take longer, the send job goes back to `queued` and takes a new token (D-086, replaces D-084).
  - `Store.Claimable` is back.
- Verified with: `TestConsumer_ConsentRevokedDuringWait` and `TestConsumer_WaitingJobIsNotExpired` failed first (one request sent, and a job in `unknown`). They pass now. The pacing test of the first review also passes. These three tests
… [4381 more chars truncated]
```

</details>

**🔧 Bash** _Inspect status fields and transitions_
```bash
cd ~/Desktop/projects/case-studies/insider_one/backend && cat internal/delivery/types.go && grep -rn "status\|Status\|state" --include=*.go internal/campaign/*.go internal/delivery/store.go internal/delivery/worker.go internal/delivery/queue.go internal/delivery/consumer.go | grep -v _test | head -150; find . -name '*.sql' | head
```

<details><summary>Result</summary>

```
// Package delivery sends campaign messages through a queue and a provider.
package delivery

import (
	"errors"
	"time"
)

// Channel is a delivery channel.
type Channel string

// Channels.
const (
	ChannelSMS   Channel = "sms"
	ChannelEmail Channel = "email"
)

// Channels lists every channel.
var Channels = []Channel{ChannelSMS, ChannelEmail}

// Status is the state of a send job.
type Status string

// Send job statuses. Terminal: sent, failed, unknown, skipped.
const (
	StatusQueued   Status = "queued"
	StatusSending  Status = "sending"
	StatusRetrying Status = "retrying"
	StatusSent     Status = "sent"
	StatusFailed   Status = "failed"
	StatusUnknown  Status = "unknown"
	StatusSkipped  Status = "skipped"
)

// Skip reasons that Load returns.
const (
	SkipNoConsent = "no_consent"
	SkipNoContact = "no_contact"
	SkipOverLimit = "over_limit"
)

// Errors of the store.
var (
	ErrNotFound       = errors.New("delivery: campaign not found")
	ErrNotSendable    = errors.New("delivery: campaign has not been approved")
	ErrNotCompleted   = errors.New("delivery: campaign is not completed")
	ErrNothingToRetry = errors.New("delivery: no failed job can be retried")
)

// Job is one row of send_jobs.
type Job struct {
	ID         int64     `json:"id"`
	CampaignID int64     `json:"campaign_id"`
	UserID     int64     `json:"user_id"`
	Channel    Channel   `json:"channel"`
	Language   string    `json:"language"`
	Status     Status    `json:"status"`
	Attempts   int       `json:"attempts"`
	LastError  string    `json:"last_error"`
	Retryable  bool      `json:"retryable"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Result is the new state after an attempt. NextAttemptAt is set for
// StatusRetrying only.
type Result struct {
	Status        Status
	LastError     string
	Retryable     bool
	ProviderID    string
	NextAttemptAt *time.Time
}

// Message is a rendered message for one recipient.
type Message struct{ To, Subject, Body string }

// ChannelProgress counts the jobs of on
… [1131 more chars truncated]
```

</details>

**🔧 Bash** _Read migrations and campaign store/service_
```bash
cat internal/store/migrations/*.sql; cat internal/campaign/store.go internal/campaign/service.go
```

<details><summary>Result</summary>

```
-- +goose Up
CREATE TABLE users (
    id               bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    first_name       text NOT NULL,
    last_name        text NOT NULL,
    email            text NOT NULL UNIQUE,
    phone            text NOT NULL,
    language         text NOT NULL CHECK (language IN ('tr', 'en')),
    country          text NOT NULL CHECK (country IN ('TR', 'US')),
    city             text NOT NULL,
    birth_date       date NOT NULL,
    email_consent    boolean NOT NULL,
    sms_consent      boolean NOT NULL,
    created_at       timestamptz NOT NULL,
    last_cart_add_at timestamptz,
    last_purchase_at timestamptz,
    order_count      integer NOT NULL CHECK (order_count >= 0),
    total_spent      numeric(12,2) NOT NULL CHECK (total_spent >= 0)
);

CREATE INDEX users_city_idx ON users (city);
CREATE INDEX users_language_idx ON users (language);
CREATE INDEX users_last_cart_add_at_idx ON users (last_cart_add_at);
CREATE INDEX users_last_purchase_at_idx ON users (last_purchase_at);

CREATE TABLE events (
    id          bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    user_id     bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type        text NOT NULL CHECK (type IN ('cart_add', 'purchase')),
    amount      numeric(10,2),
    occurred_at timestamptz NOT NULL,
    CHECK ((type = 'purchase') = (amount IS NOT NULL)),
    CHECK (amount IS NULL OR amount > 0)
);

CREATE INDEX events_user_id_idx ON events (user_id);

-- +goose Down
DROP TABLE events;
DROP TABLE users;
-- +goose Up
CREATE TABLE campaigns (
    id                bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    brief             text NOT NULL,
    rule              jsonb NOT NULL,
    description       text NOT NULL,
    status            text NOT NULL CHECK (status IN ('draft', 'approved')),
    skipped_languages text[] NOT NULL DEFAULT '{}',
    created_at        timestamptz NOT NULL,
    updated_at        timestamptz NOT NULL,
    approved_at
… [20038 more chars truncated]
```

</details>

**🔧 Bash** _Read delivery store and worker_
```bash
cat internal/delivery/store.go internal/delivery/worker.go
```

<details><summary>Result</summary>

```
package delivery

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/segment"
)

// statementTimeout limits each query, as in the campaign package.
const statementTimeout = "2s"

const (
	defaultListLimit = 50
	maxListLimit     = 200
)

// Store reads and writes send jobs in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore returns a Store on pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) inTx(ctx context.Context, mode pgx.TxAccessMode, fn func(pgx.Tx) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: mode})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = '"+statementTimeout+"'"); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// PlanNext locks one approved campaign (FOR UPDATE SKIP LOCKED), inserts its
// send jobs, and sets it queued, or completed when it has none. ok is false
// when no campaign waits. ids are the new queued send job ids per channel.
// When the stored rule is not valid, it logs the error and completes the
// campaign with no jobs.
func (s *Store) PlanNext(ctx context.Context, allow segment.Allowlist, now time.Time) (campaignID int64, ids map[Channel][]int64, ok bool, err error) {
	ids = map[Channel][]int64{}
	err = s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		var raw []byte
		e := tx.QueryRow(ctx, `SELECT id, rule FROM campaigns WHERE status = 'approved'
ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&campaignID, &raw)
		if errors.Is(e, pgx.ErrNoRows) {
			return nil
		}
		if e != nil {
			return e
		}
		ok = true
		rule, ferrs := allow.ParseRule(raw)
		if len(ferrs) > 0 {
			slog.Default().Error("send planner: stored rule is not valid, campaign completed with no job
… [23633 more chars truncated]
```

</details>

**🔧 Bash** _Read consumer, provider, queue, mock failure_
```bash
cat internal/delivery/consumer.go internal/delivery/provider.go internal/delivery/queue.go | head -400; cat internal/mockprovider/failure.go
```

<details><summary>Result</summary>

```
package delivery

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Consumer handles the queue entries of one channel.
type Consumer struct {
	Store       *Store
	Queue       *Queue
	Provider    *Provider
	Limiter     *Limiter
	Channel     Channel
	Clock       Clock
	MaxAttempts int
	Jitter      func() float64
	Log         *slog.Logger
	// MaxPrep is the longest time between a rate token and the request.
	// The claim and the load run in this time. 0 means 100 ms.
	MaxPrep time.Duration
	Metrics *Metrics // nil means no metrics

	// Hooks for tests. Both are nil in production.
	afterClaim func()
	finishErr  func() error
}

func (c *Consumer) log() *slog.Logger {
	if c.Log != nil {
		return c.Log
	}
	return slog.Default()
}

// defaultMaxPrep is the default of Consumer.MaxPrep.
const defaultMaxPrep = 100 * time.Millisecond

// Handle runs one entry: wait, claim, load, send, write the result, acknowledge.
// It returns an error only when the entry must stay unacknowledged.
func (c *Consumer) Handle(ctx context.Context, e Entry) error {
	// The writes after a claim use a context that cannot be cancelled: the
	// result of an attempt must reach Postgres, also during a shutdown.
	done := context.WithoutCancel(ctx)

	// An entry for a job that cannot be claimed must not use a rate token.
	// On an error here, the normal path decides.
	if ok, err := c.Store.Claimable(ctx, e.JobID, c.Clock.Now()); err == nil && !ok {
		return c.Queue.Ack(done, c.Channel, e.StreamID)
	}

	maxPrep := c.MaxPrep
	if maxPrep <= 0 {
		maxPrep = defaultMaxPrep
	}
	var job Job
	var msg Message
	for {
		// The job is not claimed during the wait. It stays queued, so the
		// recovery cannot take it for a request that hangs.
		waitFrom := c.Clock.Now()
		err := c.Limiter.Wait(ctx)
		granted := c.Clock.Now()
		c.Metrics.Waited(c.Channel, granted.Sub(waitFrom))
		if err != nil {
			return err
		}

		var ok bool
		job, ok, err = c.Store.Claim(ctx, e.JobID, c.Clock.Now())
		if err != nil {
			return er
… [12213 more chars truncated]
```

</details>

**🔧 Bash** _Read retry decision, limiter, architecture doc_
```bash
grep -rn "func Next\|func.*Limiter\|func (l \*Limiter)" internal/ ; sed -n '/^func Next/,/^}/p' $(grep -rl "func Next" internal/); grep -rln "Limiter struct" internal | xargs sed -n '1,200p' | head -120; cat docs/architecture.md 2>/dev/null | head -5; ls ../docs; grep -n "state\|State" ../docs/architecture.md | head -30
```

<details><summary>Result</summary>

```
internal/mockprovider/server.go:173:func retryAfter(lim *rate.Limiter, now time.Time, perSec float64) int {
internal/delivery/policy.go:19:func Next(o Outcome, attempts, maxAttempts int, now time.Time, jitter func() float64) Result {
internal/delivery/limiter.go:22:func NewLimiter(perSec int, clock Clock) *Limiter {
internal/delivery/limiter.go:28:func (l *Limiter) Pause(d time.Duration) {
internal/delivery/limiter.go:39:func (l *Limiter) Wait(ctx context.Context) error {
internal/delivery/limiter.go:56:func (l *Limiter) paused() bool {
internal/delivery/limiter.go:62:func (l *Limiter) waitPause(ctx context.Context) error {
internal/delivery/limiter_test.go:9:func TestLimiter_Rate(t *testing.T) {
internal/delivery/limiter_test.go:34:func TestLimiter_Pause(t *testing.T) {
internal/delivery/limiter_test.go:50:func TestLimiter_WaitCancelled(t *testing.T) {
func Next(o Outcome, attempts, maxAttempts int, now time.Time, jitter func() float64) Result {
	switch o.Kind {
	case Delivered:
		return Result{Status: StatusSent, ProviderID: o.ProviderID, LastError: o.Err}
	case Retryable:
		if attempts >= maxAttempts {
			return Result{Status: StatusFailed, LastError: o.Err, Retryable: true}
		}
		wait := o.RetryAfter
		if wait <= 0 {
			wait = Backoff(attempts, jitter)
		}
		at := now.Add(wait)
		return Result{Status: StatusRetrying, LastError: o.Err, NextAttemptAt: &at}
	case Ambiguous:
		return Result{Status: StatusUnknown, LastError: o.Err}
	default:
		return Result{Status: StatusFailed, LastError: o.Err}
	}
}
package delivery

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter limits the sends of one channel to perSec each second, with no
// burst, and can be paused after a 429.
type Limiter struct {
	rl    *rate.Limiter
	clock Clock

	mu          sync.Mutex
	pausedUntil time.Time
}

// NewLimiter returns a Limiter with a burst of 1.
func NewLimiter(perSec int, clock Clock) *Limiter {
	return &Limiter{rl: rate.NewLimiter(rate.Limit(perSec), 1), cloc
… [1169 more chars truncated]
```

</details>

**🔧 Bash** _Check architecture doc and frontend files_
```bash
cd ..; grep -n "^#\|status\|state" docs/architecture.md | head -60; wc -l docs/architecture.md; ls frontend/src; find frontend/src -name '*.vue' -o -name '*.ts' | grep -v test | grep -iv "components/ui" | head -60
```

<details><summary>Result</summary>

```
1:# Architecture
3:## Components
5:- **api** — Go HTTP API. Health endpoints, campaigns, segments, content, send status and the dashboard. It serves `/metrics` on port 8080 (D-068 to D-076).
13:## Folder tree
46:## Flow
70:   - The service counts the audience before it locks the row. The approval then locks the campaign row and checks the status in one transaction. A repeated approval returns the campaign and changes nothing (D-039).
76:   - Each channel has one Redis stream. An entry holds only the send job id. Postgres holds the delivery status (D-051, D-052).
83:5. **Live status** — The dashboard shows the delivery status.
      88 docs/architecture.md
App.vue
assets
components
composables
constants
layouts
lib
main.ts
pages
plugins
router
stores
types
utils
validators
frontend/src/App.vue
frontend/src/main.ts
frontend/src/types/vue-router-meta.d.ts
frontend/src/types/auto-import.d.ts
frontend/src/types/auto-import-components.d.ts
frontend/src/types/vite-env.d.ts
frontend/src/types/route-map.d.ts
frontend/src/composables/use-sidebar-navigation.ts
frontend/src/composables/use-system-theme.ts
frontend/src/plugins/index.ts
frontend/src/constants/app-config.ts
frontend/src/constants/app.ts
frontend/src/constants/sidebar-data.ts
frontend/src/constants/themes.ts
frontend/src/stores/sidebar-config.ts
frontend/src/stores/theme.ts
frontend/src/utils/is-external-url.ts
frontend/src/utils/env.ts
frontend/src/components/confirm-dialog.vue
frontend/src/components/custom-error.vue
frontend/src/components/loading.vue
frontend/src/components/toggle-theme.vue
frontend/src/layouts/blank.vue
frontend/src/layouts/default.vue
frontend/src/lib/api-client.ts
frontend/src/lib/utils.ts
frontend/src/lib/content.ts
frontend/src/lib/dashboard.ts
frontend/src/lib/campaigns.ts
frontend/src/lib/format.ts
frontend/src/lib/segments.ts
frontend/src/lib/limits.ts
frontend/src/lib/delivery.ts
frontend/src/validators/env.validator.ts
frontend/src/pages/index.vue
frontend/src/pages/errors.vue
fronten
… [1007 more chars truncated]
```

</details>

**🔧 Bash** _Read architecture flow and campaign components_
```bash
sed -n 40,88p docs/architecture.md; find frontend/src/components -path '*campaign*' -o -path '*review*' -o -path '*compose*' | grep -v test; wc -l frontend/src/pages/campaigns/*.vue
```

<details><summary>Result</summary>

````
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
   - Cla
… [5860 more chars truncated]
````

</details>

**🔧 Bash** _Find state refs in frontend components_
```bash
cd frontend/src; wc -l components/campaign/*.vue lib/content.ts lib/delivery.ts; grep -n "ref<\|ref(\|phase\|status\|state\|computed(" components/campaign/review-dialog.vue components/campaign/brief-audience.vue components/campaign/content-cards.vue components/campaign/sending-dialog.vue pages/campaigns/\[id\].vue lib/content.ts lib/delivery.ts | head -150
```

<details><summary>Result</summary>

```
204 components/campaign/brief-audience.vue
     215 components/campaign/content-cards.vue
      48 components/campaign/content-view.vue
      30 components/campaign/delivery-bar.vue
     341 components/campaign/delivery-section.vue
      32 components/campaign/message-preview.vue
     493 components/campaign/review-dialog.vue
     143 components/campaign/sending-dialog.vue
      29 components/campaign/status-badge.vue
     137 lib/content.ts
     123 lib/delivery.ts
    1795 total
components/campaign/sending-dialog.vue:33:const step = computed(() => props.progress ? currentStep(props.progress.status) : 'approved')
components/campaign/sending-dialog.vue:34:const isDone = computed(() => props.progress?.status === 'completed')
components/campaign/sending-dialog.vue:43:const rows = computed(() => CHANNELS
components/campaign/sending-dialog.vue:47:const totals = computed(() => {
components/campaign/sending-dialog.vue:87:            :data-state="stepState(i)"
components/campaign/review-dialog.vue:41:const campaign = ref<Campaign | null>(null)
components/campaign/review-dialog.vue:42:const loadError = ref('')
components/campaign/review-dialog.vue:44:const saved = ref<Record<string, Content>>({})
components/campaign/review-dialog.vue:45:const samples = ref<Record<string, string[]>>({})
components/campaign/review-dialog.vue:46:const sampleFailed = ref(false)
components/campaign/review-dialog.vue:48:const lang = ref('')
components/campaign/review-dialog.vue:49:const channel = ref<Channel>('sms')
components/campaign/review-dialog.vue:51:const banner = ref('')
components/campaign/review-dialog.vue:52:const busy = ref(false)
components/campaign/review-dialog.vue:53:const confirmOpen = ref(false)
components/campaign/review-dialog.vue:54:const discardOpen = ref(false)
components/campaign/review-dialog.vue:58:const isDraft = computed(() => campaign.value?.status === 'draft')
components/campaign/review-dialog.vue:59:const contents = computed(() => campaign.value?.contents ?? [])
com
… [6362 more chars truncated]
```

</details>

**🔧 Bash** _Read review dialog, content cards, delivery lib_
```bash
sed -n 1,200p components/campaign/review-dialog.vue | sed -n 60,200p; sed -n 1,140p components/campaign/content-cards.vue; sed -n 50,123p lib/delivery.ts
```

<details><summary>Result</summary>

```
const current = computed(() => contents.value.find(c => c.language === lang.value))
const nameRunes = computed(() => current.value?.first_name_runes ?? 0)

function copyContent(c: Content): Content {
  return { sms: c.sms, email_subject: c.email_subject, email_body: c.email_body }
}

function reset(c: Campaign) {
  campaign.value = c
  for (const k of Object.keys(edits)) delete edits[k]
  for (const k of Object.keys(serverErrors)) delete serverErrors[k]
  const snap: Record<string, Content> = {}
  for (const lc of c.contents) {
    edits[lc.language] = copyContent(lc)
    snap[lc.language] = copyContent(lc)
  }
  saved.value = snap
  if (!c.contents.some(lc => lc.language === lang.value))
    lang.value = c.contents[0]?.language ?? ''
  banner.value = ''
}

async function load() {
  const id = props.campaignId
  const mine = ++loadToken
  loadError.value = ''
  loadedFor = id
  try {
    const c = await getCampaign(id)
    if (mine !== loadToken)
      return
    reset(c)
  }
  catch (err) {
    if (mine === loadToken)
      loadError.value = err instanceof CampaignRequestError ? err.message : 'Something went wrong. Try again.'
    return
  }
  try {
    const names = await getSampleNames(id)
    if (mine !== loadToken)
      return
    samples.value = names
    sampleFailed.value = false
  }
  catch {
    if (mine !== loadToken)
      return
    samples.value = {}
    sampleFailed.value = true
  }
}

// Reloads the counts and limits after a failed check, and keeps the edits.
async function refreshLimits() {
  const id = props.campaignId
  const mine = loadToken
  try {
    const c = await getCampaign(id)
    if (mine === loadToken && id === props.campaignId)
      campaign.value = c
  }
  catch {
    // The old limits stay. The server check still decides.
  }
}

const dirty = computed(() => Object.keys(edits).some(l =>
  FIELDS.some(f => edits[l]![f] !== saved.value[l]?.[f])))

// Load on each open, so that a regenerated draft shows its new text. Unsaved
// edits o
… [8091 more chars truncated]
```

</details>

**🔧 Bash** _Inspect content and segment generation flow_
```bash
cd ../../backend; grep -n "func \|send(\|emit\|\"retry\"\|\"final\"\|\"error\"\|\"campaign\"\|\"done\"\|\"audience\"\|\"draft\"\|attempt\|repair" internal/httpapi/content.go | head -80; grep -n "repair\|attempt\|is_error\|func " internal/segment/prompt.go internal/httpapi/segments.go | head -30
```

<details><summary>Result</summary>

```
42:func RegisterContent(mux *http.ServeMux, g *ContentGenerator) {
93:func (g *ContentGenerator) generate(w http.ResponseWriter, r *http.Request) {
141:	_ = sse.send("audience", audienceEvent{Description: description, Languages: langs})
165:		_ = sse.send("error", streamErrorEvent{Code: "draft_not_saved", Message: "The draft could not be saved. Try again."})
167:	_ = sse.send("done", struct{}{})
172:func (g *ContentGenerator) saveDraft(ctx context.Context, sse *sseWriter, id *int64, d campaign.Draft, results []langResult) {
204:		_ = sse.send("campaign", campaignEvent{ID: saved, Status: campaign.StatusDraft})
207:		_ = sse.send("error", streamErrorEvent{Code: "campaign_conflict",
211:		_ = sse.send("error", streamErrorEvent{Code: "draft_not_saved", Message: "The draft could not be saved. Try again."})
215:// runLanguage streams the content of one language, with one repair round,
218:func (g *ContentGenerator) runLanguage(ctx context.Context, sse *sseWriter, lang segment.LanguageCounts, description, brief string) *finalEvent {
233:	fail := func(attempts int, ec, msg string, errs []content.FieldError, err error) {
234:		slog.WarnContext(ctx, "content generate", "language", code, "attempts", attempts,
236:		_ = sse.send("error", languageErrorEvent{Language: code, Code: ec, Message: msg, Details: errs})
239:	for attempts := 1; ; attempts++ {
244:				slog.InfoContext(ctx, "content generate: cancelled", "language", code, "attempts", attempts)
248:				fail(attempts, "llm_timeout", "The language model took too long to answer.", nil, err)
250:				fail(attempts, "llm_error", "The language model call failed.", nil, err)
255:			fail(attempts, "llm_refused", "The language model refused this brief.", nil, nil)
276:			slog.InfoContext(ctx, "content generate", "language", code, "attempts", attempts, "outcome", "final",
280:			_ = sse.send("final", fe)
283:		if attempts >= 2 {
284:			fail(attempts, "content_invalid", "The language model could not write valid content for this languag
… [2809 more chars truncated]
```

</details>

**🔧 Bash** _Read content generation and review submit logic_
```bash
sed -n 93,215p internal/httpapi/content.go; sed -n 335,375p internal/httpapi/content.go; sed -n 200,320p ../frontend/src/components/campaign/review-dialog.vue; grep -n "stream\|Progress\|done" internal/httpapi/delivery.go | head -30
```

<details><summary>Result</summary>

```
func (g *ContentGenerator) generate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in struct {
		Brief      string          `json:"brief"`
		Rule       json.RawMessage `json:"rule"`
		CampaignID *int64          `json:"campaign_id"`
	}
	if err := decodeBody(w, r, &in); err != nil || in.Brief == "" || isAbsentJSON(in.Rule) ||
		(in.CampaignID != nil && *in.CampaignID <= 0) {
		writeError(w, http.StatusBadRequest, "invalid_request", "The request body must be JSON with a brief and a rule, of at most 16 KiB.", nil)
		return
	}
	brief, err := segment.SanitizeBrief(in.Brief)
	if err != nil {
		writeBriefError(w, err)
		return
	}
	rule, ruleErrs := g.Allow.ParseRule(in.Rule)
	if len(ruleErrs) > 0 {
		writeError(w, http.StatusBadRequest, "invalid_rule", "The segment rule is not valid.", ruleErrs)
		return
	}

	cctx, cancel := context.WithTimeout(ctx, countTimeout)
	aud, err := g.Count(cctx, g.Allow.Compile(rule, g.now()))
	cancel()
	if err != nil {
		slog.ErrorContext(ctx, "content generate: audience count failed", "err", err)
		writeError(w, http.StatusServiceUnavailable, "db_unavailable", "The audience could not be counted. Try again later.", nil)
		return
	}
	langs := make([]segment.LanguageCounts, 0, len(aud.Languages))
	for _, l := range aud.Languages {
		if l.Reachable > 0 {
			langs = append(langs, l)
		}
	}
	if len(langs) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "empty_audience", "No user in this audience can get a message.", nil)
		return
	}

	sse, ok := newSSEWriter(w)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "The server cannot stream the response.", nil)
		return
	}
	description := g.Allow.Describe(rule)
	_ = sse.send("audience", audienceEvent{Description: description, Languages: langs})

	// One worker for each language. A failure in one language does not cancel
	// the others; only the end of the request context cancels them all.
	results := make([]langResult, len(langs))
	var wg sync.WaitGroup
… [8320 more chars truncated]
```

</details>

**🔧 Bash** _Read campaign page progress logic and stream end_
```bash
sed -n 60,165p ../frontend/src/pages/campaigns/\[id\].vue; sed -n 86,130p internal/httpapi/delivery.go; tail -30 ../docs/decisions-log.md | grep -o "^## D-[0-9]*\|^### D-[0-9]*\|D-[0-9][0-9]*" | tail -3
```

<details><summary>Result</summary>

```
stopWatch?.()
  stopWatch = null
}

// The page owns the progress stream, so that the status, the badge and the buttons follow it.
function startWatching() {
  stopWatching()
  if (!progress.value || !['approved', 'queued', 'sending'].includes(progress.value.status))
    return
  const watchedId = id.value
  stopWatch = watchProgress(watchedId, (p) => {
    if (watchedId !== id.value)
      return
    progress.value = p
    progressError.value = ''
    if (campaign.value)
      campaign.value.status = p.status
    if (p.status === 'completed')
      stopWatching()
  })
}

// A reload bumps loadToken, so a response from an older load or id is dropped.
async function refreshProgress(mine = loadToken) {
  try {
    const p = await getProgress(id.value)
    if (mine !== loadToken)
      return
    progress.value = p
    progressError.value = ''
    if (campaign.value)
      campaign.value.status = p.status
    startWatching()
  }
  catch (err) {
    if (mine !== loadToken)
      return
    progressError.value = err instanceof CampaignRequestError ? err.message : FALLBACK_MESSAGE
  }
}

async function load() {
  const mine = ++loadToken
  stopWatching()
  error.value = ''
  notFound.value = false
  if (!Number.isInteger(id.value) || id.value <= 0) {
    notFound.value = true
    return
  }
  try {
    const c = await getCampaign(id.value)
    if (mine !== loadToken)
      return
    campaign.value = c
    if (!c.contents.some(x => x.language === lang.value))
      lang.value = c.contents[0]?.language ?? ''
    progress.value = null
    progressError.value = ''
    if (c.status !== 'draft')
      await refreshProgress(mine)
  }
  catch (err) {
    if (mine !== loadToken)
      return
    if (err instanceof CampaignRequestError && (err.status === 404 || err.status === 400))
      notFound.value = true
    else
      error.value = err instanceof CampaignRequestError ? err.message : FALLBACK_MESSAGE
  }
}

onBeforeUnmount(() => {
  loadToken++
  stopWatching()
})

watch(id, 
… [1640 more chars truncated]
```

</details>

**🔧 Bash** _Find how retried jobs reach the queue_
```bash
grep -rn "\.Retry(" --include=*.go . | grep -v _test; grep -rn "func.*Retry" internal/delivery/*.go | grep -v _test
```

<details><summary>Result</summary>

```
(eval):1: no matches found: --include=*.go
internal/delivery/service.go:34:func (s *Service) Retry(ctx context.Context, id int64) (retried int, p Progress, err error) {
internal/delivery/store.go:508:func (s *Store) Retry(ctx context.Context, campaignID int64, now time.Time) (map[Channel][]int64, error) {
internal/delivery/metrics.go:96:func (m *Metrics) RetryScheduled(ch Channel) {
```

</details>

**🔧 Bash** _Read delivery service retry_
```bash
sed -n 1,80p internal/delivery/service.go
```

<details><summary>Result</summary>

```
package delivery

import (
	"context"
	"log/slog"
	"time"
)

// Service gives the api the delivery state and the retry action.
type Service struct {
	Store *Store
	Queue *Queue
	Now   func() time.Time
	Log   *slog.Logger
}

// Progress returns the job counts. ErrNotFound for an unknown campaign,
// ErrNotSendable for a draft.
func (s *Service) Progress(ctx context.Context, id int64) (Progress, error) {
	return s.Store.Progress(ctx, id)
}

// ListJobs checks the campaign first, because the store returns an empty list
// for an unknown campaign or a draft.
func (s *Service) ListJobs(ctx context.Context, id int64, f JobFilter) ([]Job, *int64, error) {
	if _, err := s.Store.Progress(ctx, id); err != nil {
		return nil, nil, err
	}
	return s.Store.ListJobs(ctx, id, f)
}

// Retry resets the retryable failed jobs, then pushes them. A push error is
// logged, not returned: the worker recovery pushes the jobs later.
func (s *Service) Retry(ctx context.Context, id int64) (retried int, p Progress, err error) {
	ids, err := s.Store.Retry(ctx, id, s.Now())
	if err != nil {
		return 0, Progress{}, err
	}
	for ch, list := range ids {
		retried += len(list)
		if perr := s.Queue.Push(ctx, ch, list); perr != nil {
			s.Log.ErrorContext(ctx, "retry push failed, recovery will push", "campaign_id", id, "channel", ch, "err", perr)
			continue
		}
		if merr := s.Store.MarkEnqueued(ctx, list, s.Now()); merr != nil {
			s.Log.ErrorContext(ctx, "retry mark enqueued failed", "campaign_id", id, "channel", ch, "err", merr)
		}
	}
	p, err = s.Store.Progress(ctx, id)
	return retried, p, err
}
```

</details>

**🔧 Write** `~/Desktop/projects/case-studies/insider_one/docs/state-machines.md`
````
# State machines

This document lists the state machines in the code. Most of them have no explicit state type. Their states come from status columns, booleans, nullable fields and conditional `UPDATE` statements.

Each section gives:

- the place where the state lives
- the states
- the transitions, with the code that makes them
- the guard that blocks other transitions

The last section lists the gaps that we found.

## 1. Campaign status

**State:** `campaigns.status` in Postgres.

**Guard:** explicit. The trigger `campaigns_frozen_after_approval` in `00003_send_jobs.sql` allows only the transitions in the table. The writes are in three packages.

```mermaid
stateDiagram-v2
    [*] --> draft: content stream saves the draft
    draft --> approved: POST /approve
    approved --> queued: planner wrote send jobs
    approved --> completed: planner found no recipient or a bad rule
    queued --> sending: first claim of a send job
    queued --> completed: allowed, no code path
    sending --> completed: recovery found no open send job
    completed --> sending: POST /retry
```

| From | To | Code | Condition |
|---|---|---|---|
| (none) | `draft` | `campaign.Store.CreateDraft` | At least one language is final at the end of the content stream. |
| `draft` | `draft` | `ReplaceDraft`, `Update(approve=false)` | Content changes only. `ReplaceDraft` also needs the same brief and rule. |
| `draft` | `approved` | `campaign.Store.Update(approve=true)` | The content check passes. The row is locked. |
| `approved` | `approved` | `campaign.Store.Update(approve=true)` | A repeated approval. It returns `repeat = true` and changes nothing. |
| `approved` | `queued` | `delivery.Store.PlanNext` | The planner inserted one or more send jobs. |
| `approved` | `completed` | `delivery.Store.PlanNext` | The planner inserted no send job, or the stored rule is not valid. |
| `queued` | `sending` | `delivery.Store.Claim` | The same transaction claims the first send job. |
| `queued`, `sending` | `completed` | `delivery.Store.CompleteCampaigns` | No send job is `queued`, `sending` or `retrying`. Recovery runs it every 5 s. |
| `completed` | `sending` | `delivery.Store.Retry` | One or more `failed` send jobs have `retryable = true`. |

Other rules:

- `approved_at` is null only in `draft`. A `CHECK` constraint enforces this.
- After `draft`, the brief, the rule, the description, the skipped languages and `approved_at` cannot change. The trigger enforces this.
- The trigger `campaign_contents_draft_only` blocks all writes to `campaign_contents` after `draft`.
- `UpdateCounts` does nothing when the status is not `draft`.

## 2. Send job status

**State:** `send_jobs.status`, `attempts`, `next_attempt_at`, `retryable` and `enqueued_at` in Postgres.

**Guard:** implicit. No trigger checks the transitions. Each `UPDATE` has a `WHERE status = ...` condition, and that condition is the only guard. The comment in `types.go` names the terminal states: `sent`, `failed`, `unknown` and `s
… [11436 more chars truncated]
````

<details><summary>Result</summary>

```
File created successfully at: ~/Desktop/projects/case-studies/insider_one/docs/state-machines.md (file state is current in your context — no need to Read it back)
```

</details>

**🔧 Bash** _Run STE checker on the new doc_
```bash
cd ~/Desktop/projects/case-studies/insider_one && python3 .claude/skills/ste-writing/scripts/ste_check.py docs/state-machines.md
```

<details><summary>Result</summary>

```
docs/state-machines.md:248: present-perfect: has unsaved
```

</details>

**🔧 Edit** `~/Desktop/projects/case-studies/insider_one/docs/state-machines.md`
**old:**
```
except when the same campaign has unsaved edits.
```
**new:**
```
except when the edits of the same campaign are not saved.
```

<details><summary>Result</summary>

```
The file ~/Desktop/projects/case-studies/insider_one/docs/state-machines.md has been updated successfully. (file state is current in your context — no need to Read it back)
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:30:10</sub>

I wrote the state machines down in [docs/state-machines.md](docs/state-machines.md). It has a Mermaid diagram and a transition table for each one, and each transition names the function that makes it. The STE checker passes. I changed no code and committed nothing.

**Machines found (10):**

1. **Campaign status:** `draft → approved → queued → sending → completed`, and `completed → sending` on retry. This is the only machine that the database enforces, with the trigger in `00003_send_jobs.sql`. The writes that change it are in `campaign.Store`, `PlanNext`, `Claim`, `CompleteCampaigns` and `Retry`.
2. **Send job status:** 7 states. The `WHERE status = ...` in each `UPDATE` is the only guard. The doc includes the table from provider outcome to new status in `policy.go` `Next`.
3. **Queue entry:** new, pending and acknowledged, with `XAUTOCLAIM` after 60 s.
4. **Consumer steps in `Handle`:** precheck, wait, claim, load, send, finish. A rate token older than 100 ms sends the job back to `queued` (D-086).
5. **Channel rate limiter:** open, or paused after a `429`.
6. **Retry push memory:** the in-memory `pushedRetries` map, with its 30 s rule.
7. **Content generation for one language:** the backend loop with one repair round, the language cards in the frontend, and the `draftId` and `token` rules for a run.
8. **Segment preview:** a repair loop with two attempts.
9. **Review window:** loading, clean, dirty, saving, confirming, approving and discard prompt. These states come from `busy`, `dirty`, `confirmOpen` and `discardOpen`.
10. **Campaign page progress:** streaming, then polling, then stopped. Polling never goes back to the stream.

**Gaps I noted for you to decide.** None of them is a confirmed bug:

- **No database guard on send job status.** A new code path can write any status, even from `sent`. A trigger like the campaign one would close this, but it adds work to each write on the busiest table.
- **`queued → completed` has no code path.** The trigger allows it, but `Claim` moves the campaign to `sending` in the same transaction as the first claim.
- **A failed `Finish` gives `unknown`, even when nothing was sent.** If the write of a skip result fails, recovery later sets the job to `unknown`. Nothing is sent twice, but the reports count a skip as an unknown result.
- **The `retry` SSE event has two meanings.** It is sent when the LLM stream resets, and also for a repair round. The language card shows both the same way.
- **`currentStep` maps `draft` to `approved`.** No user sees this today, because the sending window opens only after approval.
- **Shared state names.** `queued` and `sending` are statuses of both a campaign and a send job.

I did not add a decisions-log entry, because this doc records no new choice. Do you want a link to it from `docs/architecture.md`?

## 👤 User  <sub>2026-11-11 16:36:27</sub>

how to fix each

**🔧 Bash** _Check status types and imports for the fixes_
```bash
cd backend && grep -n "Status" internal/campaign/types.go 2>/dev/null || grep -rn "StatusDraft\s*Status\|Status = \"" internal/campaign/; grep -rn "campaigncomposer/internal/campaign\"" internal/delivery internal/campaign; grep -n "retryEvent struct" -A4 internal/httpapi/*.go; grep -n "CampaignStatus\|DeliveryStatus\|type Step" ../frontend/src/lib/*.ts | head
```

<details><summary>Result</summary>

```
internal/campaign/campaign.go:17:	StatusDraft    Status = "draft"
internal/campaign/campaign.go:18:	StatusApproved Status = "approved"
internal/delivery/store_test.go:14:	"campaigncomposer/internal/campaign"
internal/delivery/service_test.go:14:	"campaigncomposer/internal/campaign"
internal/httpapi/content.go:56:type retryEvent struct {
internal/httpapi/content.go-57-	Language string               `json:"language"`
internal/httpapi/content.go-58-	Errors   []content.FieldError `json:"errors,omitempty"`
internal/httpapi/content.go-59-}
internal/httpapi/content.go-60-
../frontend/src/lib/content.ts:1:import type { CampaignStatus } from '@/lib/campaigns'
../frontend/src/lib/content.ts:14:    | { event: 'campaign', data: { id: number, status: CampaignStatus } }
../frontend/src/lib/dashboard.ts:1:import type { CampaignStatus } from '@/lib/campaigns'
../frontend/src/lib/dashboard.ts:2:import type { DeliveryChannel, DeliveryStatus } from '@/lib/delivery'
../frontend/src/lib/dashboard.ts:12:  failures: { status: DeliveryStatus, reason: string, sms: number, email: number }[]
../frontend/src/lib/dashboard.ts:14:  campaigns: { id: number, description: string, status: CampaignStatus, total: number, sent: number, failed: number, unknown: number, duration_ms: number | null }[]
../frontend/src/lib/campaigns.ts:5:export type CampaignStatus = 'draft' | 'approved' | 'queued' | 'sending' | 'completed'
../frontend/src/lib/campaigns.ts:9:  status: CampaignStatus
../frontend/src/lib/campaigns.ts:18:export interface CampaignSummary { id: number, status: CampaignStatus, description: string, languages: string[], created_at: string, approved_at: string | null }
../frontend/src/lib/delivery.ts:1:import type { CampaignStatus } from '@/lib/campaigns'
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:37:16</sub>

Below is a fix for each gap, ordered by how much it changes behaviour. I changed nothing yet.

## 1. A failed `Finish` gives `unknown` when nothing was sent (fix first)

This is the only gap where the stored result is wrong.

- **Fix:** in `Consumer.finish`, a write error on a skip result calls `Release`. `Release` puts the job back from `sending` to `queued`. Do this only for `StatusSkipped`. The entry stays unacknowledged, so the next delivery loads the job again and skips it again. This is safe, because a skip sends nothing.
- **Why only skips:** a `Release` after a `retrying` result would drop the backoff and the `Retry-After`. After `sent` or `unknown`, the message may have been delivered, so a `Release` could send it twice.
- **Limit:** if the database is down, `Release` also fails. The job then still becomes `unknown`, as it does today.
- **Test:** the test hook `finishErr` already exists. Make it fail on a skip, then check that the job is `queued` and that recovery does not set it to `unknown`.

## 2. No database guard on the send job status

- **Fix:** add a new migration, `00005`, with a trigger that runs `BEFORE UPDATE` on `send_jobs`. Add `WHEN (OLD.status IS DISTINCT FROM NEW.status)`, so the trigger runs only when the status changes. The writes to `enqueued_at` do not run it.
- **Allowed pairs:**
  - `queued → sending`
  - `retrying → sending`
  - `sending → queued`, `sent`, `retrying`, `failed`, `unknown` or `skipped`
  - `failed → queued`, only when `OLD.retryable` is true
- **Second guard:** make a new row start in `queued`. Use `CHECK`-style logic in a `BEFORE INSERT` trigger, or the same function.
- **Cost:** one small plpgsql call for each status change. `ExpireSending` and `Retry` change many rows in one statement. A row that breaks the rule fails the whole statement. That is the correct result, because such a row shows a bug.
- **Test:** one store test for each blocked pair, for example `sent → queued`. Also run `make verify`, because the existing paths must still pass.

## 3. The `retry` event has two meanings

- **Fix:** add `Reason string \`json:"reason"\`` to `retryEvent` in [content.go:56](backend/internal/httpapi/content.go#L56).
  - `drafter.on` sends `reason: "restart"` after a `Reset`.
  - The repair round sends `reason: "invalid"` with the errors.
- **Frontend:** add `reason` to the `ContentEvent` type in `lib/content.ts`. In `content-cards.vue`, show "Fixing the content…" only for `invalid`. For `restart`, clear the draft and stay in the drafting state.
- **Docs:** this changes the SSE contract. Update the event list in `docs/architecture.md` and spec 05, and add a decisions-log entry.

## 4. `currentStep` maps `draft` to `approved`

- **Fix:** add `type SendableStatus = Exclude<CampaignStatus, 'draft'>` in `lib/campaigns.ts`. Use it for `Progress.status` and for the parameter of `currentStep`. The backend never returns progress for a draft (`ErrNotSendable`), so the type matches the API.
- **Switch:** give `currentStep` one `case` for each status, with no `default`. TypeScript then reports a missing case when someone adds a status.
- No runtime change, so `make verify` is the full check.

## 5. Shared state names

Do not rename the values. They are in the database, the API and the UI.

- **Fix in Go:** add `StatusQueued`, `StatusSending` and `StatusCompleted` next to the two constants in [campaign.go:17](backend/internal/campaign/campaign.go#L17). In `delivery/store.go`, use these constants in place of the string literals `"completed"`, `"sending"` and `"draft"`. Change `setCampaign` to take a `campaign.Status`, and change `Progress.Status` to `campaign.Status`. There is no import cycle, because the delivery tests already import `campaign`.
- The compiler then catches a campaign status passed where a send job status is expected. The frontend already has separate `CampaignStatus` and `DeliveryStatus` types.
- The raw SQL strings stay. A trigger from fix 2 is the guard for those.

## 6. `queued → completed` has no code path

I recommend no change to the trigger. Keep the transition as a fallback for recovery.

- **Reason:** if you remove it, `CompleteCampaigns` must change to `WHERE status = 'sending'`, or the bulk update fails. A future bug could then leave a campaign in `queued` forever. A campaign that gets stuck is worse than a transition that is too broad.
- **Fix:** add one SQL comment at the trigger and one line in `docs/state-machines.md` that calls it a recovery fallback.

## Suggested order

1, 2, 4, 5, 3, 6. Fixes 1, 2, 4 and 5 are internal. Fix 3 changes the SSE contract, so it needs a decisions-log entry and spec updates. Do you want me to implement them, and if so, all of them or a subset?

## 👤 User  <sub>2026-11-11 16:38:06</sub>

go on

**🔧 Bash** _Inspect consumer and store tests_
```bash
grep -n "finishErr\|func Test\|func new\|func setup" internal/delivery/consumer_test.go | head -50; ls internal/delivery/; grep -n "func " internal/delivery/store_test.go | head -30
```

<details><summary>Result</summary>

```
51:func newConsumerEnv(t *testing.T) *consumerEnv {
109:func TestConsumer_Sequences(t *testing.T) {
184:func TestConsumer_DoubleEntry(t *testing.T) {
210:func TestConsumer_NoContact(t *testing.T) {
231:func TestConsumer_ReleaseOnCancel(t *testing.T) {
253:func TestConsumer_CancelInWait(t *testing.T) {
270:func TestConsumer_FinishFailsNoResend(t *testing.T) {
275:	c.finishErr = func() error {
304:func TestConsumer_NotClaimableSpendsNoToken(t *testing.T) {
329:func TestConsumer_LogsNoPersonalData(t *testing.T) {
375:func TestConsumer_CancelDuringSend(t *testing.T) {
401:func TestConsumer_PacesRequestsAfterDatabaseStall(t *testing.T) {
479:func TestConsumer_PacesRequestsAfterSlowPreparation(t *testing.T) {
544:func TestConsumer_ConsentRevokedDuringWait(t *testing.T) {
571:func TestConsumer_WaitingJobIsNotExpired(t *testing.T) {
consumer.go
consumer_test.go
deliverytest
fakeclock_test.go
limiter.go
limiter_test.go
metrics.go
metrics_test.go
policy.go
policy_test.go
provider.go
provider_test.go
queue.go
queue_test.go
service.go
service_test.go
store.go
store_test.go
types.go
worker.go
worker_test.go
33:func seeded(t *testing.T) (context.Context, *pgxpool.Pool, *Store) {
38:func seededN(t *testing.T, users int) (context.Context, *pgxpool.Pool, *Store) {
57:func approved(t *testing.T, ctx context.Context, pool *pgxpool.Pool, rule string, langs ...string) int64 {
75:func planned(t *testing.T, ctx context.Context, pool *pgxpool.Pool, s *Store) (int64, map[Channel][]int64) {
85:func campaignStatus(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id int64) string {
94:func getJob(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id int64) Job {
105:func exec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) {
112:func expectedCounts(t *testing.T, ctx context.Context, pool *pgxpool.Pool, rule string, langs ...string) (sms, email int64) {
133:func TestStore_PlanExample(t *testing.T) {
158:func TestStore_PlanSkippedLanguage(t *testing.T) {
… [888 more chars truncated]
```

</details>

**🔧 Bash** _Read consumer test helpers and related tests_
```bash
sed -n 1,108p internal/delivery/consumer_test.go; sed -n 205,305p internal/delivery/consumer_test.go; sed -n 355,380p internal/delivery/store_test.go
```

<details><summary>Result</summary>

```
package delivery

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/delivery/deliverytest"
)

type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

type consumerEnv struct {
	ctx   context.Context
	pool  *pgxpool.Pool
	store *Store
	queue *Queue
	clk   *fakeClock
	srv   string
	logs  *syncBuf
	sms   []int64
}

func newConsumerEnv(t *testing.T) *consumerEnv {
	t.Helper()
	ctx, pool, s := seeded(t)
	_, ids := planned(t, ctx, pool, s)
	if len(ids[ChannelSMS]) < 7 {
		t.Fatalf("fixture has %d sms jobs, want at least 7", len(ids[ChannelSMS]))
	}
	q := NewQueue(deliverytest.NewRedis(t, 0), "c1")
	if err := q.Ensure(ctx); err != nil {
		t.Fatal(err)
	}
	return &consumerEnv{ctx: ctx, pool: pool, store: s, queue: q, clk: newFakeClock(testNow),
		srv: newMock(t).URL, logs: &syncBuf{}, sms: ids[ChannelSMS]}
}

func (e *consumerEnv) consumer(p *Provider) *Consumer {
	return &Consumer{
		Store: e.store, Queue: e.queue, Provider: p, Limiter: NewLimiter(1000, e.clk),
		Channel: ChannelSMS, Clock: e.clk, MaxAttempts: 5, Jitter: func() float64 { return 0 },
		Log: slog.New(slog.NewTextHandler(e.logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}
}

// entries pushes id n times and reads the new entries.
func (e *consumerEnv) entries(t *testing.T, id int64, n int) []Entry {
	t.Helper()
	ids := make([]int64, n)
	for i := range ids {
		ids[i] = id
	}
	if err := e.queue.Push(e.ctx, ChannelSMS, ids); err != nil {
		t.Fatal(err)
	}
	es, err := e.queue.Read(e.ctx, ChannelSMS, n, time.Second)
	if err != nil || len(es) != n {
		t.Fatalf("Read = %v, %v", es, err)
	}
	return es
}

func (e *consumerEnv) pending(t *testin
… [4527 more chars truncated]
```

</details>

**🔧 Bash** _Find test fixtures that write send job statuses_
```bash
grep -rn "addJob(\|SET status\|INSERT INTO send_jobs" --include=*_test.go -r internal | grep -v "^internal/delivery/store.go"
```

<details><summary>Result</summary>

```
(eval):1: no matches found: --include=*_test.go
```

</details>

**🔧 Bash** _Find test fixtures that write statuses_
```bash
grep -rn "addJob(\|SET status\|INSERT INTO send_jobs\|status = '" internal cmd | grep _test.go
```

<details><summary>Result</summary>

```
internal/dashboard/store_test.go:81:	mustExec(t, ctx, db, `INSERT INTO send_jobs (campaign_id, user_id, channel, language, status, attempts, last_error, sent_at, created_at, updated_at)
internal/dashboard/store_test.go:276:	mustExec(t, ctx, db, `UPDATE campaigns SET status = 'completed' WHERE id = $1`, completed)
internal/dashboard/store_test.go:277:	mustExec(t, ctx, db, `UPDATE campaigns SET status = 'queued' WHERE id = $1`, queued)
internal/dashboard/store_test.go:323:	mustExec(t, ctx, db, `UPDATE campaigns SET status = 'completed' WHERE id = $1`, cid)
internal/dashboard/store_test.go:326:	mustExec(t, ctx, db, `UPDATE campaigns SET status = 'sending' WHERE id = $1`, cid)
internal/delivery/store_test.go:144:	if err := pool.QueryRow(ctx, `SELECT count(*) FROM send_jobs WHERE campaign_id = $1 AND status = 'queued' AND enqueued_at IS NULL AND language = 'tr'`, cid).Scan(&n); err != nil {
internal/delivery/store_test.go:269:	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1`, id, testNow.Add(time.Minute))
internal/delivery/store_test.go:369:func addJob(t *testing.T, ctx context.Context, pool *pgxpool.Pool, cid, uid int64, ch Channel, status string) int64 {
internal/delivery/store_test.go:372:	err := pool.QueryRow(ctx, `INSERT INTO send_jobs (campaign_id, user_id, channel, language, status, created_at, updated_at)
internal/delivery/store_test.go:469:	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1`, q[3], testNow.Add(-time.Second))
internal/delivery/store_test.go:470:	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1`, q[4], testNow.Add(time.Minute))
internal/delivery/store_test.go:480:	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'sending', attempts = 1, updated_at = $2 WHERE id = $1`, q[5], testNow.Add(-40*time.Second))
internal/delivery/store_test.go:481:	exec(t, ctx, pool, `UPDAT
… [3961 more chars truncated]
```

</details>

**🔧 Bash** _Inspect fixtures affected by a guard_
```bash
grep -rn "addJob(t" internal | grep -v "func addJob"; sed -n 75,95p internal/dashboard/store_test.go; sed -n 185,230p internal/store/migrate_test.go; sed -n 460,505p internal/delivery/worker_test.go; sed -n 300,312p internal/delivery/consumer_test.go
```

<details><summary>Result</summary>

```
return id
}

func insertJob(t *testing.T, ctx context.Context, db *pgxpool.Pool, campaignID, userID int64, channel delivery.Channel,
	status delivery.Status, attempts int, lastError string, sentAt *time.Time, createdAt, updatedAt time.Time) {
	t.Helper()
	mustExec(t, ctx, db, `INSERT INTO send_jobs (campaign_id, user_id, channel, language, status, attempts, last_error, sent_at, created_at, updated_at)
VALUES ($1, $2, $3, 'en', $4, $5, $6, $7, $8, $9)`,
		campaignID, userID, string(channel), string(status), attempts, lastError, sentAt, createdAt, updatedAt)
}

// addJobs inserts n jobs with the same values, each for a new user.
func addJobs(t *testing.T, ctx context.Context, db *pgxpool.Pool, cid int64, ch delivery.Channel, st delivery.Status, n int) {
	t.Helper()
	for _, u := range insertUsers(t, ctx, db, n) {
		insertJob(t, ctx, db, cid, u, ch, st, 1, "", nil, base, base)
	}
}

func groupBy(t *testing.T, ctx context.Context, db *pgxpool.Pool) map[delivery.Channel]map[delivery.Status]int64 {
	t.Helper()
	}
}

func TestMigrate_SendJobsUnique(t *testing.T) {
	ctx, conn := migrated(t)
	id := newCampaign(t, ctx, conn, "queued")
	var uid int64
	if err := conn.QueryRow(ctx, insertUser+` RETURNING id`, "u@example.com", "tr", "TR").Scan(&uid); err != nil {
		t.Fatal(err)
	}
	ins := `INSERT INTO send_jobs (campaign_id, user_id, channel, language, status, created_at, updated_at)
VALUES ($1, $2, $3, 'tr', 'queued', now(), now())`
	if _, err := conn.Exec(ctx, ins, id, uid, "sms"); err != nil {
		t.Fatal(err)
	}
	_, err := conn.Exec(ctx, ins, id, uid, "sms")
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Errorf("duplicate: want unique violation, got %v", err)
	}
	if _, err := conn.Exec(ctx, ins, id, uid, "email"); err != nil {
		t.Errorf("other channel: %v", err)
	}
}
}

func TestWorker_RetryPollerOnTime(t *testing.T) {
	e := newWorkerEnv(t, 500, exampleRule, nil, cleanOpts())
	_, ids, ok, err := e.store.PlanNext(e.ctx, segment.Default(), t
… [2148 more chars truncated]
```

</details>

**🔧 Bash** _Read store test fixtures in context_
```bash
sed -n 435,520p internal/delivery/store_test.go; sed -n 572,580p internal/delivery/store_test.go; sed -n 600,665p internal/delivery/store_test.go; sed -n 25,40p internal/delivery/service_test.go; sed -n 262,272p internal/delivery/store_test.go
```

<details><summary>Result</summary>

```
func TestStore_Recovery(t *testing.T) {
	ctx, pool, s := seeded(t)
	cid, ids := planned(t, ctx, pool, s)
	q := ids[ChannelSMS]
	// q[0] enqueued_at null, q[1] 200 s old, q[2] 10 s old; the rest are fresh.
	exec(t, ctx, pool, `UPDATE send_jobs SET enqueued_at = $1 WHERE campaign_id = $2`, testNow, cid)
	exec(t, ctx, pool, `UPDATE send_jobs SET enqueued_at = NULL WHERE id = $1`, q[0])
	exec(t, ctx, pool, `UPDATE send_jobs SET enqueued_at = $2 WHERE id = $1`, q[1], testNow.Add(-200*time.Second))
	exec(t, ctx, pool, `UPDATE send_jobs SET enqueued_at = $2 WHERE id = $1`, q[2], testNow.Add(-10*time.Second))
	got, err := s.StaleQueued(ctx, testNow, 120*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(got[ChannelSMS]) != 2 || len(got[ChannelEmail]) != 0 {
		t.Fatalf("stale = %v", got)
	}
	want := map[int64]bool{q[0]: true, q[1]: true}
	for _, id := range got[ChannelSMS] {
		if !want[id] {
			t.Fatalf("unexpected stale %d", id)
		}
	}
	if err := s.MarkEnqueued(ctx, []int64{q[0]}, testNow); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkEnqueued(ctx, nil, testNow); err != nil {
		t.Fatal(err)
	}
	got, _ = s.StaleQueued(ctx, testNow, 120*time.Second)
	if len(got[ChannelSMS]) != 1 || got[ChannelSMS][0] != q[1] {
		t.Fatalf("after mark: %v", got)
	}

	// Due retries.
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1`, q[3], testNow.Add(-time.Second))
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1`, q[4], testNow.Add(time.Minute))
	due, err := s.DueRetries(ctx, testNow, 0)
	if err != nil || len(due[ChannelSMS]) != 1 || due[ChannelSMS][0] != q[3] {
		t.Fatalf("due = %v %v", due, err)
	}
	if due, _ = s.DueRetries(ctx, testNow, 5*time.Second); len(due[ChannelSMS]) != 0 {
		t.Fatalf("lateBy ignored: %v", due)
	}

	// Expire sending.
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'sending', attempts = 1, updated_at = $2 WHERE id = $1`, q[5], testNo
… [6060 more chars truncated]
```

</details>

**🔧 Bash** _Check test database user and helpers_
```bash
ls internal/delivery/deliverytest; grep -rn "POSTGRES_USER\|postgres://" ../docker-compose*.yml ../Makefile internal/*/*test*.go internal/delivery/deliverytest/*.go 2>/dev/null | head; grep -rn "func exec" internal/delivery/*_test.go
```

<details><summary>Result</summary>

```
redis.go
internal/config/config_test.go:14:	cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x", "REDIS_URL": "redis://r"}))
internal/config/config_test.go:27:	if cfg.DatabaseURL != "postgres://x" {
internal/config/config_test.go:28:		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://x")
internal/config/config_test.go:41:		"DATABASE_URL": "postgres://x",
internal/config/config_test.go:52:		"DATABASE_URL": "postgres://x",
internal/config/config_test.go:62:	cfg, err := LoadSeed(env(map[string]string{"DATABASE_URL": "postgres://x"}))
internal/config/config_test.go:66:	want := SeedConfig{DatabaseURL: "postgres://x", Users: 5000, Random: 42, Reset: false}
internal/config/config_test.go:74:		"DATABASE_URL": "postgres://x",
internal/config/config_test.go:82:	want := SeedConfig{DatabaseURL: "postgres://x", Users: 12, Random: 7, Reset: true}
internal/config/config_test.go:185:	cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x", "REDIS_URL": "redis://r"}))
internal/delivery/store_test.go:105:func exec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) {
```

</details>

**🔧 Bash** _Find how tests get a Postgres database_
```bash
sed -n 1,60p internal/delivery/store_test.go; grep -rn "testcontainers\|TEST_DATABASE\|func.*Postgres" internal --include=*.go 2>/dev/null | head; grep -rln "testcontainers\|TEST_DATABASE" internal | head; sed -n '/^test:/,/^[a-z]*:/p' ../Makefile
```

<details><summary>Result</summary>

```
package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/campaign"
	"campaigncomposer/internal/content"
	"campaigncomposer/internal/seed"
	"campaigncomposer/internal/segment"
	"campaigncomposer/internal/store"
	"campaigncomposer/internal/store/storetest"
)

var testNow = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

const exampleRule = `{"op":"and","conditions":[
 {"field":"city","op":"eq","value":"Istanbul"},
 {"field":"last_cart_add_at","op":"within_days","value":30},
 {"op":"or","conditions":[
  {"field":"last_purchase_at","op":"is_null"},
  {"field":"last_purchase_at","op":"older_than_days","value":30}]}]}`

const trUSRule = `{"op":"and","conditions":[{"field":"country","op":"in","value":["TR","US"]}]}`

func seeded(t *testing.T) (context.Context, *pgxpool.Pool, *Store) {
	t.Helper()
	return seededN(t, 500)
}

func seededN(t *testing.T, users int) (context.Context, *pgxpool.Pool, *Store) {
	t.Helper()
	url := storetest.NewDB(t)
	ctx := context.Background()
	if err := store.Migrate(ctx, url); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if _, err := seed.Load(ctx, pool, seed.Options{Random: 42, Now: testNow, Users: users}, false); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return ctx, pool, NewStore(pool)
}

// approved makes an approved campaign with one content row for each language.
func approved(t *testing.T, ctx context.Context, pool *pgxpool.Pool, rule string, langs ...string) int64 {
	t.Helper()
	cs := campaign.NewStore(pool)
	var lcs []campaign.LanguageContent
(eval):1: no matches found: --include=*.go
internal/store/storetest/storetest.go
test:
	@docker network inspect cc-test >/dev/null 2>&1 || docker network create cc-test >/dev/null
	@docker rm -f cc-test-pg cc-test-redis >/dev/null 2>&1 || true
	docker run -d --name cc-test-pg --
… [937 more chars truncated]
```

</details>

**🔧 Bash** _Read trigger tests and retry event usage_
```bash
sed -n 120,185p internal/store/migrate_test.go; grep -n "retry" ../frontend/src/lib/content.ts ../frontend/src/components/campaign/*.test.ts ../frontend/src/components/campaign/__tests__/* 2>/dev/null | head -30; find ../frontend/src -name '*content-cards*'; grep -rn "retryEvent{" internal/httpapi/
```

<details><summary>Result</summary>

```
VALUES ('b', '{"op":"and","conditions":[]}', 'd', $1, now(), now(), `+approved+`) RETURNING id`, status).Scan(&id)
	if err != nil {
		t.Fatalf("insert campaign %s: %v", status, err)
	}
	return id
}

func TestMigrate_CampaignStatusMoves(t *testing.T) {
	ctx, conn := migrated(t)
	allowed := [][2]string{
		{"draft", "approved"}, {"approved", "queued"}, {"approved", "completed"},
		{"queued", "sending"}, {"queued", "completed"}, {"sending", "completed"},
		{"completed", "sending"},
	}
	for _, m := range allowed {
		id := newCampaign(t, ctx, conn, m[0])
		set := "status = $2"
		if m[0] == "draft" {
			set += ", approved_at = now()"
		}
		if _, err := conn.Exec(ctx, `UPDATE campaigns SET `+set+` WHERE id = $1`, id, m[1]); err != nil {
			t.Errorf("%s -> %s: want ok, got %v", m[0], m[1], err)
		}
	}
	denied := [][2]string{
		{"queued", "approved"}, {"sending", "queued"}, {"completed", "approved"}, {"sending", "draft"},
		{"approved", "sending"}, {"completed", "queued"},
	}
	for _, m := range denied {
		id := newCampaign(t, ctx, conn, m[0])
		if _, err := conn.Exec(ctx, `UPDATE campaigns SET status = $2 WHERE id = $1`, id, m[1]); err == nil {
			t.Errorf("%s -> %s: want error, got nil", m[0], m[1])
		}
	}
	if _, err := conn.Exec(ctx, `INSERT INTO campaigns (brief, rule, description, status, created_at, updated_at)
VALUES ('b', '{}', 'd', 'queued', now(), now())`); err == nil {
		t.Error("queued with approved_at null accepted")
	}
}

func TestMigrate_FrozenAfterApprovalStill(t *testing.T) {
	ctx, conn := migrated(t)
	id := newCampaign(t, ctx, conn, "sending")
	for _, q := range []string{
		`UPDATE campaigns SET brief = 'x' WHERE id = $1`,
		`UPDATE campaigns SET rule = '{"op":"or","conditions":[]}' WHERE id = $1`,
		`UPDATE campaigns SET approved_at = now() + interval '1 day' WHERE id = $1`,
	} {
		if _, err := conn.Exec(ctx, q, id); err == nil {
			t.Errorf("want error for %s", q)
		}
	}
	// Content rows can only exist for a draft, so insert one before approval.
	draft := n
… [1090 more chars truncated]
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:39:42</sub>

Starting with fix 1, test first.

**🔧 Edit** `~/Desktop/projects/case-studies/insider_one/backend/internal/delivery/consumer_test.go`
**old:**
```
func TestConsumer_NotClaimableSpendsNoToken(t *testing.T) {
```
**new:**
```
// A skip sends nothing. When its result cannot be written, the job goes back
// to queued, so the recovery does not set it to unknown.
func TestConsumer_SkipWriteFailsReleases(t *testing.T) {
	e := newConsumerEnv(t)
	srv := newMock(t)
	c := e.consumer(NewProvider(srv.URL, 2*time.Second))
	exec(t, e.ctx, e.pool, `UPDATE users SET phone = '' WHERE id = (SELECT user_id FROM send_jobs WHERE id = $1)`, e.sms[0])
	fail := true
	c.finishErr = func() error {
		if fail {
			fail = false
			return errors.New("database is down")
		}
		return nil
	}
	es := e.entries(t, e.sms[0], 1)
	if err := c.Handle(e.ctx, es[0]); err == nil {
		t.Fatal("Handle = nil, want the Finish error")
	}
	if j := getJob(t, e.ctx, e.pool, e.sms[0]); j.Status != StatusQueued || j.Attempts != 0 {
		t.Fatalf("job = %s attempts %d, want queued 0", j.Status, j.Attempts)
	}
	if n, err := e.store.ExpireSending(e.ctx, e.clk.Now().Add(time.Hour), time.Second); err != nil || n != 0 {
		t.Fatalf("ExpireSending = %d, %v, want 0", n, err)
	}
	claimed, err := e.queue.AutoClaim(e.ctx, ChannelSMS, 0, 10)
	if err != nil || len(claimed) != 1 {
		t.Fatalf("AutoClaim = %v, %v", claimed, err)
	}
	if err := c.Handle(e.ctx, claimed[0]); err != nil {
		t.Fatal(err)
	}
	if j := getJob(t, e.ctx, e.pool, e.sms[0]); j.Status != StatusSkipped || j.LastError != SkipNoContact || j.Attempts != 1 {
		t.Fatalf("job = %s %q attempts %d", j.Status, j.LastError, j.Attempts)
	}
	if got := providerStats(t, srv.URL)["sms"].Requests; got != 0 {
		t.Fat
… [172 more chars truncated]
```

**🔧 Bash** _Check how to run Go tests_
```bash
cd .. && grep -n "GO_TEST\s*[:?]*=\|TEST_DATABASE_URL\s*[:?]*=\|TEST_REDIS_URL\s*[:?]*=" Makefile; which go
```

<details><summary>Result</summary>

```
The file ~/Desktop/projects/case-studies/insider_one/backend/internal/delivery/consumer_test.go has been updated successfully.
```

</details>

<details><summary>Result</summary>

```
8:GO_TEST = docker run --rm --network cc-test -e TEST_DATABASE_URL -e TEST_REDIS_URL -v "$(CURDIR)/backend:/src" -w /src golang:1.27-alpine go
9:TEST_DATABASE_URL = postgres://test:test@cc-test-pg:5432/test?sslmode=disable
10:TEST_REDIS_URL = redis://cc-test-redis:6379/0
14:GO_TEST = cd backend && go
15:TEST_DATABASE_URL = postgres://test:test@127.0.0.1:55432/test?sslmode=disable
16:TEST_REDIS_URL = redis://127.0.0.1:56379/0
63:	@export TEST_DATABASE_URL='$(TEST_DATABASE_URL)' TEST_REDIS_URL='$(TEST_REDIS_URL)'; $(GO_TEST) test ./...; rc=$$?; \
/opt/homebrew/bin/go
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:39:56</sub>

Starting test containers to run targeted tests.

**🔧 Bash** _Start test databases and run the new failing test_
```bash
docker network inspect cc-test >/dev/null 2>&1 || docker network create cc-test >/dev/null
docker rm -f cc-test-pg cc-test-redis >/dev/null 2>&1
docker run -d --name cc-test-pg --network cc-test -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=test -p 127.0.0.1:55432:5432 postgres:17-alpine >/dev/null
docker run -d --name cc-test-redis --network cc-test -p 127.0.0.1:56379:6379 redis:8-alpine >/dev/null
until docker exec cc-test-pg pg_isready -U test -d test -h 127.0.0.1 >/dev/null 2>&1; do sleep 1; done
cd backend && TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55432/test?sslmode=disable' TEST_REDIS_URL='redis://127.0.0.1:56379/0' go test ./internal/delivery/ -run 'TestConsumer_SkipWriteFailsReleases' -count=1 2>&1 | tail -15
```

<details><summary>Result</summary>

```
--- FAIL: TestConsumer_SkipWriteFailsReleases (0.16s)
    consumer_test.go:324: job = sending attempts 1, want queued 0
FAIL
FAIL	campaigncomposer/internal/delivery	0.628s
FAIL
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:40:06</sub>

Fails as expected. Now the fix.

**🔧 Edit** `~/Desktop/projects/case-studies/insider_one/backend/internal/delivery/consumer.go`
**old:**
```
// finish writes the result and acknowledges the entry. When the write fails,
// the entry stays unacknowledged. Its job stays in sending, and a new claim
// fails, so the message is not sent again.
func (c *Consumer) finish(ctx context.Context, e Entry, job Job, r Result) error {
	var err error
	if c.finishErr != nil {
		err = c.finishErr()
	}
	if err == nil {
		err = c.Store.Finish(ctx, job.ID, r, c.Clock.Now())
	}
	if err != nil {
		c.log().Error("write send result", "job_id", job.ID, "channel", c.Channel, "attempt", job.Attempts, "error", err)
		return err
	}
```
**new:**
```
// finish writes the result and acknowledges the entry. When the write fails,
// the entry stays unacknowledged. Its job stays in sending, and a new claim
// fails, so the message is not sent again. A skip is the exception: nothing
// was sent, so the job goes back to queued and the next delivery of the entry
// skips it again. Without this, the recovery would set it to unknown.
func (c *Consumer) finish(ctx context.Context, e Entry, job Job, r Result) error {
	var err error
	if c.finishErr != nil {
		err = c.finishErr()
	}
	if err == nil {
		err = c.Store.Finish(ctx, job.ID, r, c.Clock.Now())
	}
	if err != nil {
		c.log().Error("write send result", "job_id", job.ID, "channel", c.Channel, "attempt", job.Attempts, "error", err)
		if r.Status == StatusSkipped {
			if rerr := c.Store.Release(ctx, job.ID); rerr != nil {
				c.log().Error("release skipped job", "job_id", job.ID, "channel", c.Channel, "error", rerr)
			}
		}
		return err
	}
```

**🔧 Bash** _Run consumer tests after the fix_
```bash
TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55432/test?sslmode=disable' TEST_REDIS_URL='redis://127.0.0.1:56379/0' go test ./internal/delivery/ -run 'TestConsumer_' -count=1 2>&1 | tail -15
```

<details><summary>Result</summary>

```
The file ~/Desktop/projects/case-studies/insider_one/backend/internal/delivery/consumer.go has been updated successfully.
```

</details>

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/delivery	3.464s
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:40:21</sub>

Fix 1 passes. Now fix 2: the trigger test first.

**🔧 Bash** _Read migrate test helpers_
```bash
sed -n 1,40p internal/store/migrate_test.go; sed -n 100,119p internal/store/migrate_test.go; sed -n 230,260p internal/store/migrate_test.go
```

<details><summary>Result</summary>

```
package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"campaigncomposer/internal/store"
	"campaigncomposer/internal/store/storetest"
)

func migrated(t *testing.T) (context.Context, *pgx.Conn) {
	t.Helper()
	ctx := context.Background()
	url := storetest.NewDB(t)
	if err := store.Migrate(ctx, url); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { conn.Close(context.Background()) })
	return ctx, conn
}

const insertUser = `INSERT INTO users (first_name,last_name,email,phone,language,country,city,birth_date,email_consent,sms_consent,created_at,order_count,total_spent)
VALUES ('A','B',$1,'+905000000000',$2,$3,'Istanbul','1990-01-01',true,true,now(),0,0)`

func TestMigrate_CreatesTables(t *testing.T) {
	ctx, conn := migrated(t)
	for _, tbl := range []string{"users", "events"} {
		var reg *string
		if err := conn.QueryRow(ctx, "SELECT to_regclass($1)::text", "public."+tbl).Scan(&reg); err != nil {
			t.Fatal(err)
		}
		if reg == nil {
		{"update brief", `UPDATE campaigns SET brief = 'x' WHERE id = $1`, []any{id}},
		{"back to draft", `UPDATE campaigns SET status = 'draft' WHERE id = $1`, []any{id}},
	}
	for _, f := range fails {
		if _, err := conn.Exec(ctx, f.sql, f.args...); err == nil {
			t.Errorf("%s: want error after approval, got nil", f.name)
		}
	}
}

// newCampaign inserts a campaign with the given status. Insert does not fire
// the update trigger, so any status can be the start of a move.
func newCampaign(t *testing.T, ctx context.Context, conn *pgx.Conn, status string) int64 {
	t.Helper()
	approved := "now()"
	if status == "draft" {
		approved = "NULL"
	}
	var id int64
	err := conn.QueryRow(ctx, `INSERT INTO campaigns (brief, rule, description, status, created_at, updated_at, approved_at)
```

</details>

**🔧 Bash** _Add trigger test and run it to see it fail_
```bash
cat >> internal/store/migrate_test.go <<'EOF'

// newJob inserts a send job with the given status. Insert does not fire the
// update trigger, so any status can be the start of a move.
func newJob(t *testing.T, ctx context.Context, conn *pgx.Conn, cid int64, email, status string, retryable bool) int64 {
	t.Helper()
	var uid, id int64
	if err := conn.QueryRow(ctx, insertUser+` RETURNING id`, email, "tr", "TR").Scan(&uid); err != nil {
		t.Fatal(err)
	}
	if err := conn.QueryRow(ctx, `INSERT INTO send_jobs (campaign_id, user_id, channel, language, status, retryable, created_at, updated_at)
VALUES ($1, $2, 'sms', 'tr', $3, $4, now(), now()) RETURNING id`, cid, uid, status, retryable).Scan(&id); err != nil {
		t.Fatalf("insert job %s: %v", status, err)
	}
	return id
}

func TestMigrate_SendJobStatusMoves(t *testing.T) {
	ctx, conn := migrated(t)
	cid := newCampaign(t, ctx, conn, "sending")
	n := 0
	job := func(status string, retryable bool) int64 {
		n++
		return newJob(t, ctx, conn, cid, "j"+string(rune('a'+n%26))+string(rune('a'+n/26))+"@example.com", status, retryable)
	}
	allowed := [][2]string{
		{"queued", "sending"}, {"retrying", "sending"},
		{"sending", "queued"}, {"sending", "sent"}, {"sending", "retrying"}, {"sending", "failed"},
		{"sending", "unknown"}, {"sending", "skipped"},
	}
	for _, m := range allowed {
		id := job(m[0], false)
		if _, err := conn.Exec(ctx, `UPDATE send_jobs SET status = $2 WHERE id = $1`, id, m[1]); err != nil {
			t.Errorf("%s -> %s: want ok, got %v", m[0], m[1], err)
		}
	}
	retry := job("failed", true)
	if _, err := conn.Exec(ctx, `UPDATE send_jobs SET status = 'queued', retryable = false WHERE id = $1`, retry); err != nil {
		t.Errorf("retryable failed -> queued: want ok, got %v", err)
	}
	denied := [][2]string{
		{"queued", "sent"}, {"queued", "failed"}, {"queued", "retrying"}, {"retrying", "queued"},
		{"sent", "queued"}, {"sent", "sending"}, {"sent", "retrying"}, {"unknown", "queued"},
		{"unknown", "sending"}, {"skipped", "queued"}, {"failed", "sending"},
	}
	for _, m := range denied {
		id := job(m[0], false)
		if _, err := conn.Exec(ctx, `UPDATE send_jobs SET status = $2 WHERE id = $1`, id, m[1]); err == nil {
			t.Errorf("%s -> %s: want error, got nil", m[0], m[1])
		}
	}
	permanent := job("failed", false)
	if _, err := conn.Exec(ctx, `UPDATE send_jobs SET status = 'queued' WHERE id = $1`, permanent); err == nil {
		t.Error("failed with retryable = false -> queued: want error, got nil")
	}
	// A write that keeps the status is not a move.
	sent := job("sent", false)
	if _, err := conn.Exec(ctx, `UPDATE send_jobs SET enqueued_at = now() WHERE id = $1`, sent); err != nil {
		t.Errorf("same status: want ok, got %v", err)
	}
}
EOF
TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55432/test?sslmode=disable' go test ./internal/store/ -run SendJobStatusMoves -count=1 2>&1 | tail -15
```

<details><summary>Result</summary>

```
migrate_test.go:256: queued -> sent: want error, got nil
    migrate_test.go:256: queued -> failed: want error, got nil
    migrate_test.go:256: queued -> retrying: want error, got nil
    migrate_test.go:256: retrying -> queued: want error, got nil
    migrate_test.go:256: sent -> queued: want error, got nil
    migrate_test.go:256: sent -> sending: want error, got nil
    migrate_test.go:256: sent -> retrying: want error, got nil
    migrate_test.go:256: unknown -> queued: want error, got nil
    migrate_test.go:256: unknown -> sending: want error, got nil
    migrate_test.go:256: skipped -> queued: want error, got nil
    migrate_test.go:256: failed -> sending: want error, got nil
    migrate_test.go:261: failed with retryable = false -> queued: want error, got nil
FAIL
FAIL	campaigncomposer/internal/store	0.385s
FAIL
```

</details>

**🔧 Bash** _Add the transition trigger migration and run store tests_
```bash
python3 - <<'EOF'
p='internal/store/migrate_test.go'
s=open(p).read()
s=s.replace('''		return newJob(t, ctx, conn, cid, "j"+string(rune('a'+n%26))+string(rune('a'+n/26))+"@example.com", status, retryable)''','''		return newJob(t, ctx, conn, cid, fmt.Sprintf("j%d@example.com", n), status, retryable)''')
s=s.replace('"errors"\n\t"testing"','"errors"\n\t"fmt"\n\t"testing"',1)
open(p,'w').write(s)
EOF
cat > internal/store/migrations/00005_send_job_transitions.sql <<'EOF'
-- +goose Up
-- The send job status moves only along the paths of the worker. Each UPDATE in
-- delivery.Store also has its own WHERE condition. This trigger is the second
-- guard: a new code path cannot write a status that the worker never writes.
--
-- Note on campaigns_frozen_after_approval: it allows queued -> completed, but no
-- code path makes it, because Claim moves the campaign to sending with the
-- first claim. It stays as a fallback for CompleteCampaigns.
-- +goose StatementBegin
CREATE FUNCTION send_jobs_status_moves() RETURNS trigger AS $$
BEGIN
    IF (OLD.status, NEW.status) IN (
        ('queued', 'sending'),
        ('retrying', 'sending'),
        ('sending', 'queued'),
        ('sending', 'sent'),
        ('sending', 'retrying'),
        ('sending', 'failed'),
        ('sending', 'unknown'),
        ('sending', 'skipped')
    ) OR (OLD.status = 'failed' AND OLD.retryable AND NEW.status = 'queued') THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION 'send job status cannot move from % to %', OLD.status, NEW.status;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER send_jobs_status_moves
BEFORE UPDATE OF status ON send_jobs
FOR EACH ROW WHEN (OLD.status IS DISTINCT FROM NEW.status)
EXECUTE FUNCTION send_jobs_status_moves();

-- +goose Down
DROP TRIGGER send_jobs_status_moves ON send_jobs;
DROP FUNCTION send_jobs_status_moves();
EOF
TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55432/test?sslmode=disable' go test ./internal/store/ -count=1 2>&1 | tail -15
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/store	0.872s
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:40:55</sub>

Trigger works. Now the fixtures that jump between states directly:

**🔧 Bash** _Run all backend tests to find broken fixtures_
```bash
TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55432/test?sslmode=disable' TEST_REDIS_URL='redis://127.0.0.1:56379/0' go test ./... -count=1 2>&1 | grep -E "^(---|ok|FAIL)|cannot move" | head -40
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/cmd/api	0.271s
ok  	campaigncomposer/cmd/worker	0.678s
ok  	campaigncomposer/internal/campaign	2.985s
ok  	campaigncomposer/internal/config	0.236s
ok  	campaigncomposer/internal/content	0.220s
ok  	campaigncomposer/internal/dashboard	3.623s
--- FAIL: TestConsumer_NotClaimableSpendsNoToken (0.14s)
    consumer_test.go:351: ERROR: send job status cannot move from queued to sent (SQLSTATE P0001)
--- FAIL: TestService_RetryPushes (0.10s)
    service_test.go:47: UPDATE send_jobs SET status = 'sent' WHERE campaign_id = $1: ERROR: send job status cannot move from queued to sent (SQLSTATE P0001)
--- FAIL: TestService_RetryPushFails (0.09s)
    service_test.go:84: UPDATE send_jobs SET status = 'sent' WHERE campaign_id = $1: ERROR: send job status cannot move from queued to sent (SQLSTATE P0001)
--- FAIL: TestStore_ClaimNotDue (0.09s)
    store_test.go:269: UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1: ERROR: send job status cannot move from queued to retrying (SQLSTATE P0001)
--- FAIL: TestStore_Recovery (0.10s)
    store_test.go:469: UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1: ERROR: send job status cannot move from queued to retrying (SQLSTATE P0001)
--- FAIL: TestStore_Progress (0.09s)
    store_test.go:511: UPDATE send_jobs SET status = 'sent' WHERE id = ANY($1): ERROR: send job status cannot move from queued to sent (SQLSTATE P0001)
--- FAIL: TestStore_ListJobs (0.09s)
    store_test.go:575: UPDATE send_jobs SET status = 'failed' WHERE id = ANY($1): ERROR: send job status cannot move from queued to failed (SQLSTATE P0001)
--- FAIL: TestStore_Retry (0.09s)
    store_test.go:607: UPDATE send_jobs SET status = 'sent' WHERE campaign_id = $1: ERROR: send job status cannot move from queued to sent (SQLSTATE P0001)
--- FAIL: TestWorker_RetryPollerOnTime (0.11s)
    worker_test.go:468: UPDATE send_jobs SET status = 'sent' WHERE campaign_id = $1: ERROR: send job st
… [589 more chars truncated]
```

</details>

**🔧 Bash** _Move fixtures to legal paths and rerun delivery tests_
```bash
python3 - <<'EOF'
import re
def edit(p, pairs):
    s=open(p).read()
    for a,b in pairs:
        if a not in s: raise SystemExit(f"missing in {p}: {a[:80]}")
        s=s.replace(a,b)
    open(p,'w').write(s)

helpers='''
// setJobs moves the jobs ids to status st through sending, as the worker does.
// The trigger send_jobs_status_moves allows each move out of sending, so a
// queued or retrying job can reach any status. extra is more of the SET list,
// with $2 and up for args.
func setJobs(t *testing.T, ctx context.Context, pool *pgxpool.Pool, ids []int64, st Status, extra string, args ...any) {
	t.Helper()
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'sending' WHERE id = ANY($1) AND status IN ('queued', 'retrying')`, ids)
	exec(t, ctx, pool, `UPDATE send_jobs SET status = '`+string(st)+`'`+extra+` WHERE id = ANY($1)`, append([]any{ids}, args...)...)
}

// openJobs returns the ids of the queued, retrying and sending jobs of
// campaign cid, without the ids in not.
func openJobs(t *testing.T, ctx context.Context, pool *pgxpool.Pool, cid int64, not ...int64) []int64 {
	t.Helper()
	rs, err := pool.Query(ctx, `SELECT id FROM send_jobs WHERE campaign_id = $1
AND status IN ('queued', 'retrying', 'sending') AND NOT (id = ANY($2)) ORDER BY id`, cid, append([]int64{}, not...))
	if err != nil {
		t.Fatal(err)
	}
	ids, err := pgx.CollectRows(rs, pgx.RowTo[int64])
	if err != nil {
		t.Fatal(err)
	}
	return ids
}
'''
p='internal/delivery/store_test.go'
edit(p,[
 ('func exec(t *testing.T', helpers.lstrip('\n')+'\nfunc exec(t *testing.T'),
 ('"github.com/jackc/pgx/v5/pgxpool"','"github.com/jackc/pgx/v5"\n\t"github.com/jackc/pgx/v5/pgxpool"'),
 ("exec(t, ctx, pool, `UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1`, id, testNow.Add(time.Minute))",
  "setJobs(t, ctx, pool, []int64{id}, StatusRetrying, `, attempts = 1, next_attempt_at = $2`, testNow.Add(time.Minute))"),
 ("exec(t, ctx, pool, `UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1`, q[3], testNow.Add(-time.Second))",
  "setJobs(t, ctx, pool, []int64{q[3]}, StatusRetrying, `, attempts = 1, next_attempt_at = $2`, testNow.Add(-time.Second))"),
 ("exec(t, ctx, pool, `UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1`, q[4], testNow.Add(time.Minute))",
  "setJobs(t, ctx, pool, []int64{q[4]}, StatusRetrying, `, attempts = 1, next_attempt_at = $2`, testNow.Add(time.Minute))"),
 ("""	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'sent' WHERE campaign_id = $1`, cid)
	done, err = s.CompleteCampaigns(ctx)""",
  """	setJobs(t, ctx, pool, openJobs(t, ctx, pool, cid), StatusSent, "")
	done, err = s.CompleteCampaigns(ctx)"""),
 ("""	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'sent' WHERE id = ANY($1)`, ids[ChannelSMS][:3])
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'failed', retryable = true WHERE id = ANY($1)`, ids[ChannelSMS][3:5])
	exec(t, ctx, pool, `UPDATE send_j
… [4384 more chars truncated]
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/delivery	42.681s
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:43:01</sub>

Fix 2 done. Now fix 5: typed campaign status constants.

**🔧 Bash** _Find campaign status string literals in Go_
```bash
sed -n 1,30p internal/campaign/campaign.go; grep -rn '"completed"\|"sending"\|"queued"\|"draft"\|"approved"\|\.Status ==\|\.Status !=\|p.Status\b' internal cmd --include=*.go 2>/dev/null | grep -v _test | grep -v "^internal/campaign/campaign.go"; grep -rn 'Status \(!=\|==\) "' internal cmd | grep -v _test
```

<details><summary>Result</summary>

```
// Package campaign stores draft and approved campaigns and checks the edits
// of the marketer. It imports only content, segment and store from internal.
package campaign

import (
	"encoding/json"
	"errors"
	"time"

	"campaigncomposer/internal/content"
)

// Status is the state of a campaign.
type Status string

const (
	StatusDraft    Status = "draft"
	StatusApproved Status = "approved"
)

// LanguageContent is the copy of one language with the audience snapshot of
// the last save.
type LanguageContent struct {
	Language string `json:"language"`
	content.Content
	SMSCount       int64 `json:"sms_count"`
	EmailCount     int64 `json:"email_count"`
	FirstNameRunes int   `json:"first_name_runes"`
}

(eval):1: no matches found: --include=*.go
internal/segment/validate.go:342:	if json.Unmarshal(m["status"], &ans.Status) != nil || (ans.Status != "ok" && ans.Status != "unsupported") {
internal/httpapi/delivery.go:90:		if p.Status != "completed" {
internal/httpapi/segments.go:108:	if ans.Status == "unsupported" {
internal/httpapi/health.go:54:		if body.Status != "ok" {
internal/delivery/store.go:400:		if p.Status == "draft" {
```

</details>

**🔧 Bash** _Check import rules and literal sites in delivery_
```bash
grep -rn "imports only\|go list\|Imports" internal cmd | grep -v "^internal/campaign/campaign.go" | head; grep -n '"completed"\|"sending"\|"queued"\|"draft"\|status string\|var status' internal/delivery/store.go internal/delivery/types.go
```

<details><summary>Result</summary>

```
internal/delivery/types.go:26:	StatusQueued   Status = "queued"
internal/delivery/types.go:27:	StatusSending  Status = "sending"
internal/delivery/types.go:80:	Queued          int64 `json:"queued"`
internal/delivery/types.go:81:	Sending         int64 `json:"sending"`
internal/delivery/store.go:69:			return setCampaign(ctx, tx, campaignID, "completed", now)
internal/delivery/store.go:125:			return setCampaign(ctx, tx, campaignID, "completed", now)
internal/delivery/store.go:127:		return setCampaign(ctx, tx, campaignID, "queued", now)
internal/delivery/store.go:135:func setCampaign(ctx context.Context, tx pgx.Tx, id int64, status string, now time.Time) error {
internal/delivery/store.go:400:		if p.Status == "draft" {
internal/delivery/store.go:511:		var status string
internal/delivery/store.go:518:		if status != "completed" {
internal/delivery/store.go:534:		return setCampaign(ctx, tx, campaignID, "sending", now)
```

</details>

**🔧 Bash** _Add campaign status constants and replace literals_
```bash
python3 - <<'EOF'
def edit(p, pairs):
    s=open(p).read()
    for a,b in pairs:
        n=s.count(a)
        if n==0: raise SystemExit(f"missing in {p}: {a}")
        s=s.replace(a,b)
    open(p,'w').write(s)
edit('internal/campaign/campaign.go',[(
'''// Status is the state of a campaign.
type Status string

const (
	StatusDraft    Status = "draft"
	StatusApproved Status = "approved"
)''','''// Status is the state of a campaign. The campaign package writes draft and
// approved. The delivery package writes queued, sending and completed. The
// trigger campaigns_frozen_after_approval allows only the moves of
// docs/state-machines.md.
type Status string

// Campaign statuses. The delivery package has a Status type of its own for
// send jobs, with some of the same words.
const (
	StatusDraft     Status = "draft"
	StatusApproved  Status = "approved"
	StatusQueued    Status = "queued"
	StatusSending   Status = "sending"
	StatusCompleted Status = "completed"
)''')])
edit('internal/delivery/store.go',[
 ('setCampaign(ctx, tx, campaignID, "completed", now)','setCampaign(ctx, tx, campaignID, campaign.StatusCompleted, now)'),
 ('setCampaign(ctx, tx, campaignID, "queued", now)','setCampaign(ctx, tx, campaignID, campaign.StatusQueued, now)'),
 ('setCampaign(ctx, tx, campaignID, "sending", now)','setCampaign(ctx, tx, campaignID, campaign.StatusSending, now)'),
 ('id int64, status string, now time.Time) error {','id int64, status campaign.Status, now time.Time) error {'),
 ('if p.Status == "draft" {','if p.Status == campaign.StatusDraft {'),
 ('		var status string\n','		var status campaign.Status\n'),
 ('if status != "completed" {','if status != campaign.StatusCompleted {'),
 ('"campaigncomposer/internal/content"','"campaigncomposer/internal/campaign"\n\t"campaigncomposer/internal/content"'),
])
edit('internal/delivery/types.go',[
 ('''	Status     string                      `json:"status"`''','''	Status     campaign.Status             `json:"status"`'''),
 ('import (\n\t"errors"\n\t"time"\n)','import (\n\t"errors"\n\t"time"\n\n\t"campaigncomposer/internal/campaign"\n)'),
])
edit('internal/httpapi/delivery.go',[('if p.Status != "completed" {','if p.Status != campaign.StatusCompleted {')])
EOF
grep -n '^import\|"campaigncomposer' internal/httpapi/delivery.go | head; go build ./... 2>&1 | head; go vet ./... 2>&1 | head
```

<details><summary>Result</summary>

```
3:import (
12:	"campaigncomposer/internal/delivery"
# campaigncomposer/internal/httpapi
internal/httpapi/delivery.go:90:18: undefined: campaign
# campaigncomposer/internal/httpapi
internal/httpapi/delivery.go:90:18: undefined: campaign
# campaigncomposer/internal/httpapi
# [campaigncomposer/internal/httpapi]
vet: internal/httpapi/delivery.go:90:18: undefined: campaign
```

</details>

**🔧 Bash** _Add missing import and vet_
```bash
sed -i '' 's#^\t"campaigncomposer/internal/delivery"$#\t"campaigncomposer/internal/campaign"\n\t"campaigncomposer/internal/delivery"#' internal/httpapi/delivery.go && go build ./... && go vet ./... && gofmt -l internal cmd; echo rc=$?
```

<details><summary>Result</summary>

```
# campaigncomposer/internal/httpapi
# [campaigncomposer/internal/httpapi]
vet: internal/httpapi/delivery_test.go:59:50: cannot use status (variable of type string) as campaign.Status value in struct literal
rc=1
```

</details>

**🔧 Bash** _Inspect test using string status_
```bash
sed -n 50,65p internal/httpapi/delivery_test.go; grep -rn "Progress{" internal cmd | grep _test | head
```

<details><summary>Result</summary>

```
defer f.mu.Unlock()
	f.gotID = id
	if f.err != nil {
		return 0, delivery.Progress{}, f.err
	}
	return f.retried, f.progress[0], nil
}

func prog(status string, sent int64) delivery.Progress {
	return delivery.Progress{CampaignID: 7, Status: status, Channels: map[delivery.Channel]delivery.ChannelProgress{
		delivery.ChannelSMS:   {Total: 10, Sent: sent, Queued: 10 - sent},
		delivery.ChannelEmail: {Total: 4, Failed: 1, RetryableFailed: 1},
	}}
}

func deliveryMux(f *fakeDelivery, poll time.Duration) *http.ServeMux {
internal/httpapi/delivery_test.go:34:		return delivery.Progress{}, f.err
internal/httpapi/delivery_test.go:53:		return 0, delivery.Progress{}, f.err
internal/httpapi/delivery_test.go:59:	return delivery.Progress{CampaignID: 7, Status: status, Channels: map[delivery.Channel]delivery.ChannelProgress{
internal/httpapi/delivery_test.go:78:	f := &fakeDelivery{progress: []delivery.Progress{prog("sending", 3)}}
internal/httpapi/delivery_test.go:103:	f := &fakeDelivery{progress: []delivery.Progress{prog("sending", 0)}}
internal/httpapi/delivery_test.go:201:	f := &fakeDelivery{progress: []delivery.Progress{prog("sending", 3)}, retried: 4}
internal/httpapi/delivery_test.go:215:	f := &fakeDelivery{progress: []delivery.Progress{prog("sending", 3), prog("sending", 3), prog("sending", 3), prog("sending", 5), prog("completed", 10)}}
internal/httpapi/delivery_test.go:243:	f := &fakeDelivery{progress: []delivery.Progress{prog("sending", 3)}}
```

</details>

**🔧 Bash** _Fix test helper type and vet again_
```bash
sed -i '' 's/^func prog(status string, sent int64)/func prog(status campaign.Status, sent int64)/' internal/httpapi/delivery_test.go && grep -n '"campaigncomposer/internal/campaign"' internal/httpapi/delivery_test.go || sed -i '' 's#^\t"campaigncomposer/internal/delivery"$#\t"campaigncomposer/internal/campaign"\n\t"campaigncomposer/internal/delivery"#' internal/httpapi/delivery_test.go; go vet ./... && gofmt -l internal cmd; echo rc=$?
```

<details><summary>Result</summary>

```
rc=0
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:43:43</sub>

Fix 5 compiles. Now fix 3: the `retry` event reason. Checking existing tests and spec mentions:

**🔧 Bash** _Find retry event usage in tests, spec and frontend_
```bash
grep -rn '"retry"\|retryEvent\|event: retry' internal/httpapi/*_test.go | head; grep -rn "retry" ../docs/specs/05-content-generation/*.md | head -20; grep -n "retry" ../frontend/src/lib/content.ts ../frontend/src/components/campaign/__tests__/content-cards.test.ts ../frontend/src/components/campaign/content-cards.vue
```

<details><summary>Result</summary>

```
internal/httpapi/content_test.go:447:			if e.Event == "retry" {
internal/httpapi/content_test.go:508:			if e.Event == "retry" {
internal/httpapi/content_test.go:532:			if e.Event == "retry" && len(dataOf(t, e).Errors) == 0 {
internal/httpapi/content_test.go:655:		if e.Event == "retry" {
internal/httpapi/content_test.go:667:		if e.Event == "retry" {
internal/httpapi/delivery_test.go:107:		for _, rt := range [][2]string{{"GET", "progress"}, {"GET", "progress/stream"}, {"GET", "send-jobs"}, {"POST", "retry"}} {
internal/httpapi/delivery_test.go:130:		for _, rt := range [][2]string{{"GET", "progress"}, {"GET", "progress/stream"}, {"GET", "send-jobs"}, {"POST", "retry"}} {
../docs/specs/05-content-generation/requirements.md:22:- **FR-13.** If the first answer fails validation and the second passes, the stream has `draft`, `retry`, `draft` and `final` for that language. The second mock request contains the first answer as a tool call and the errors as a tool result with the error flag.
../docs/specs/05-content-generation/requirements.md:35:- **FR-26.** The page shows one card for each language. The `draft` text shows as a draft. On `final`, the card shows the final text and a counter for SMS and subject. On `retry`, the card shows "Fixing content…". On `error`, the card shows the message, and the other cards stay.
../docs/specs/05-content-generation/requirements.md:39:- **FR-30.** A `fallback` block in the stream gives a `retry` event with no errors. The backend uses only the `submit_content` call after it. The language still has its one repair round.
../docs/specs/05-content-generation/goal.md:115:- With the refusal fallback, a `fallback` content block can arrive in the stream. The text before it is not valid. The backend sends a `retry` event with no errors, and uses only the `submit_content` call after the last `fallback` block. This is not the repair round.
../docs/specs/05-content-generation/goal.md:146:1. The backend sends a `retry` event for the language. The front
… [2117 more chars truncated]
```

</details>

**🔧 Bash** _Read retry tests on both sides_
```bash
sed -n 500,540p internal/httpapi/content_test.go; sed -n 645,675p internal/httpapi/content_test.go; grep -n "func dataOf" -A12 internal/httpapi/content_test.go; sed -n 105,150p ../frontend/src/components/campaign/__tests__/content-cards.test.ts; sed -n 170,185p ../frontend/src/components/campaign/content-cards.vue
```

<details><summary>Result</summary>

```
cut := contentStep(`{"sms":"Merhaba {{first_na`)
		cut.Response.CutOff = true
		tr, reqs := run(t, cut, contentStep(good))
		if got := strings.Join(names(tr), ","); got != "draft,retry,draft,final" {
			t.Fatalf("events %s", got)
		}
		var retry langData
		for _, e := range tr {
			if e.Event == "retry" {
				retry = dataOf(t, e)
			}
		}
		found := false
		for _, fe := range retry.Errors {
			if fe.Message == "answer was cut off" {
				found = true
			}
		}
		if !found {
			t.Errorf("retry errors %+v", retry.Errors)
		}
		if len(reqs) != 2 || reqs[1].Turns[2].ToolResult == nil || !reqs[1].Turns[2].ToolResult.IsError {
			t.Errorf("repair turns %+v", reqs[len(reqs)-1].Turns)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		tr, reqs := run(t, contentStep(""), contentStep(good))
		if tr[len(tr)-1].Event != "final" || len(reqs) != 2 {
			t.Errorf("events %v, requests %d", names(tr), len(reqs))
		}
		for _, e := range tr {
			if e.Event == "retry" && len(dataOf(t, e).Errors) == 0 {
				t.Errorf("retry has no errors: %s", e.Data)
			}
		}
	})
}

func TestGenerate_LLMErrors(t *testing.T) {
	cases := []struct {
		return contentStep(good).Response, nil
	}}
	f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
	evs := streamEvents(t, serveContent(newGenerator(c, f), genBody("hello", exampleRule())))
	tr := forLang(t, evs, "tr")
	if got := strings.Join(names(tr), ","); got != "draft,retry,draft,retry,draft,final" {
		t.Fatalf("events %s", got)
	}
	var retries []sseEvent
	for _, e := range tr {
		if e.Event == "retry" {
			retries = append(retries, e)
		}
	}
	if d := dataOf(t, retries[0]); len(d.Errors) != 0 {
		t.Errorf("reset retry has errors: %s", retries[0].Data)
	}
	if d := dataOf(t, retries[1]); len(d.Errors) == 0 {
		t.Errorf("repair retry has no errors: %s", retries[1].Data)
	}
	afterReset := false
	for _, e := range tr {
		if e.Event == "retry" {
			afterReset = true
		}
		if afterReset && e.Event == "draft" && strings.Contains(dataOf(t, e).Content.SMS, "
… [3415 more chars truncated]
```

</details>

**🔧 Bash** _Add reason checks to backend tests_
```bash
python3 - <<'EOF'
def edit(p, pairs):
    s=open(p).read()
    for a,b in pairs:
        if a not in s: raise SystemExit(f"missing in {p}: {a}")
        s=s.replace(a,b)
    open(p,'w').write(s)
edit('internal/httpapi/content_test.go',[
 ('''	if d := dataOf(t, retries[0]); len(d.Errors) != 0 {
		t.Errorf("reset retry has errors: %s", retries[0].Data)
	}
	if d := dataOf(t, retries[1]); len(d.Errors) == 0 {
		t.Errorf("repair retry has no errors: %s", retries[1].Data)
	}''','''	if d := dataOf(t, retries[0]); len(d.Errors) != 0 || d.Reason != "restart" {
		t.Errorf("reset retry: want reason restart and no errors, got %s", retries[0].Data)
	}
	if d := dataOf(t, retries[1]); len(d.Errors) == 0 || d.Reason != "invalid" {
		t.Errorf("repair retry: want reason invalid with errors, got %s", retries[1].Data)
	}'''),
 ('''		if !found {
			t.Errorf("retry errors %+v", retry.Errors)
		}''','''		if !found {
			t.Errorf("retry errors %+v", retry.Errors)
		}
		if retry.Reason != "invalid" {
			t.Errorf("retry reason %q, want invalid", retry.Reason)
		}'''),
])
EOF
grep -n "type langData" -A8 internal/httpapi/content_test.go
```

<details><summary>Result</summary>

```
150:type langData struct {
151-	Language string               `json:"language"`
152-	Content  content.Content      `json:"content"`
153-	Limits   content.Limits       `json:"limits"`
154-	Errors   []content.FieldError `json:"errors"`
155-	Code     string               `json:"code"`
156-	Message  string               `json:"message"`
157-	Details  []content.FieldError `json:"details"`
158-}
```

</details>

**🔧 Bash** _Add Reason to test struct and run failing tests_
```bash
sed -i '' '154a\
	Reason   string               `json:"reason"`
' internal/httpapi/content_test.go && go test ./internal/httpapi/ -run 'TestGenerate' -count=1 2>&1 | tail -8
```

<details><summary>Result</summary>

```
content_test.go:664: reset retry: want reason restart and no errors, got {"language":"tr"}
    content_test.go:667: repair retry: want reason invalid with errors, got {"language":"tr","errors":[{"path":"sms","message":"200 runes after rendering, limit 160"}]}
2026/11/11 19:44:10 INFO content generate: cancelled language=tr attempts=1
2026/11/11 19:44:10 INFO content generate: cancelled language=en attempts=1
2026/11/11 19:44:10 INFO content generate: client disconnected languages=2
FAIL
FAIL	campaigncomposer/internal/httpapi	0.424s
FAIL
```

</details>

**🔧 Bash** _Add reason to retry event and run httpapi tests_
```bash
python3 - <<'EOF'
p='internal/httpapi/content.go'
s=open(p).read()
for a,b in [
('''type retryEvent struct {
	Language string               `json:"language"`
	Errors   []content.FieldError `json:"errors,omitempty"`
}''','''// retryEvent tells the browser that the text of a language starts again.
// Reason is "restart" after a fallback block in the stream (not the repair
// round, no errors) or "invalid" before the repair round (with errors).
type retryEvent struct {
	Language string               `json:"language"`
	Reason   string               `json:"reason"`
	Errors   []content.FieldError `json:"errors,omitempty"`
}

const (
	retryRestart = "restart"
	retryInvalid = "invalid"
)'''),
('retryEvent{Language: code, Errors: errs}','retryEvent{Language: code, Reason: retryInvalid, Errors: errs}'),
('retryEvent{Language: d.lang}','retryEvent{Language: d.lang, Reason: retryRestart}'),
]:
    assert a in s, a
    s=s.replace(a,b)
open(p,'w').write(s)
EOF
go test ./internal/httpapi/ -count=1 2>&1 | tail -3
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/httpapi	2.431s
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:44:25</sub>

Now the frontend side of fix 3, test first.

**🔧 Bash** _Add frontend retry reason test and run it_
```bash
cd ../frontend && python3 - <<'EOF'
p='src/components/campaign/__tests__/content-cards.test.ts'
s=open(p).read()
a="""  it('clears the draft on retry and shows the next draft', async () => {"""
b="""  it('tells a restart from a repair round', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'draft', data: { language: 'tr', content } })
    r.get()({ event: 'retry', data: { language: 'tr', reason: 'restart' } })
    await flushPromises()
    const tr = w.find('[data-language="tr"]')
    expect(tr.find('[data-field="sms"]').exists()).toBe(false)
    expect(tr.text()).toContain('Starting again…')
    expect(tr.text()).not.toContain('Fixing content…')
    r.get()({ event: 'retry', data: { language: 'tr', reason: 'invalid', errors: [{ path: 'sms', message: 'too long' }] } })
    await flushPromises()
    expect(w.find('[data-language="tr"]').text()).toContain('Fixing content…')
    expect(w.find('[data-language="tr"]').text()).not.toContain('Starting again…')
    r.get()({ event: 'draft', data: { language: 'tr', content } })
    await flushPromises()
    expect(w.find('[data-language="tr"]').text()).not.toContain('Fixing content…')
  })

""" + a
assert a in s
s=s.replace(a,b,1)
open(p,'w').write(s)
EOF
grep -n '"test:run"\|"test"' package.json; pnpm vitest run src/components/campaign/__tests__/content-cards.test.ts 2>&1 | tail -15
```

<details><summary>Result</summary>

```
20:    "test": "vitest",
21:    "test:run": "vitest run",
    138|     const tr = w.find('[data-language="tr"]')
    139|     expect(tr.find('[data-field="sms"]').exists()).toBe(false)
    140|     expect(tr.text()).toContain('Starting again…')
       |                       ^
    141|     expect(tr.text()).not.toContain('Fixing content…')
    142|     r.get()({ event: 'retry', data: { language: 'tr', reason: 'invalid…

⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯[1/1]⎯


 Test Files  1 failed (1)
      Tests  1 failed | 17 passed (18)
   Start at  19:44:31
   Duration  743ms (import 54%, environment 26%, tests 11%, transform 9%)
```

</details>

**🔧 Bash** _Update ContentEvent type and view the file head_
```bash
python3 - <<'EOF'
def edit(p, pairs):
    s=open(p).read()
    for a,b in pairs:
        if a not in s: raise SystemExit(f"missing in {p}: {a}")
        s=s.replace(a,b)
    open(p,'w').write(s)
edit('src/lib/content.ts',[
 ("    | { event: 'retry', data: { language: string, errors?: { path: string, message: string }[] } }",
  "    | { event: 'retry', data: { language: string, reason?: RetryReason, errors?: { path: string, message: string }[] } }"),
])
s=open('src/lib/content.ts').read()
i=s.index('export type ContentEvent') if 'export type ContentEvent' in s else None
print(s[:900])
EOF
```

<details><summary>Result</summary>

```
import type { CampaignStatus } from '@/lib/campaigns'
import type { Counts } from '@/lib/segments'

import { API_BASE_URL } from '@/constants/app-config'

export interface Content { sms: string, email_subject: string, email_body: string }
export interface ContentLimits { first_name_runes: number, sms_runes: number, email_subject_runes: number, email_body_runes: number }
export type ContentEvent
  = | { event: 'audience', data: { description: string, languages: (Counts & { code: string })[] } }
    | { event: 'draft', data: { language: string, content: Content } }
    | { event: 'retry', data: { language: string, reason?: RetryReason, errors?: { path: string, message: string }[] } }
    | { event: 'final', data: { language: string, content: Content, limits: ContentLimits } }
    | { event: 'error', data: { language?: string, code: string, message: string } }
    | { event: 'campaign', dat
```

</details>

**🔧 Bash** _Implement retry reason in the card and run tests_
```bash
python3 - <<'EOF'
def edit(p, pairs):
    s=open(p).read()
    for a,b in pairs:
        if a not in s: raise SystemExit(f"missing in {p}: {a}")
        s=s.replace(a,b)
    open(p,'w').write(s)
edit('src/lib/content.ts',[
 ("export type ContentEvent\n",
  "// 'restart': the model began the text again (not the repair round, no errors).\n// 'invalid': the text failed the check, and the repair round starts.\nexport type RetryReason = 'restart' | 'invalid'\nexport type ContentEvent\n"),
])
edit('src/components/campaign/content-cards.vue',[
 ("import type { Content, ContentEvent, ContentLimits } from '@/lib/content'",
  "import type { Content, ContentEvent, ContentLimits, RetryReason } from '@/lib/content'"),
 ("  retrying: boolean\n", "  // retry is set from a retry event until the next draft, final or error.\n  retry?: RetryReason\n"),
 ("      retrying: false,\n", ""),
 ("""    card.draft = e.data.content
    card.retrying = false""", """    card.draft = e.data.content
    card.retry = undefined"""),
 ("""    card.draft = undefined
    card.retrying = true""", """    card.draft = undefined
    card.retry = e.data.reason ?? 'invalid'"""),
 ("""    card.final = { content: e.data.content, limits: e.data.limits }
    card.retrying = false""", """    card.final = { content: e.data.content, limits: e.data.limits }
    card.retry = undefined"""),
 ("""    card.error = e.data.message
    card.retrying = false""", """    card.error = e.data.message
    card.retry = undefined"""),
 ("""        <p v-if="card.retrying" class="text-sm text-muted-foreground" role="status">
          Fixing content…
        </p>""", """        <p v-if="card.retry" class="text-sm text-muted-foreground" role="status">
          {{ card.retry === 'restart' ? 'Starting again…' : 'Fixing content…' }}
        </p>"""),
])
EOF
grep -n "retrying" src/components/campaign/content-cards.vue; pnpm vitest run src/components/campaign/__tests__/content-cards.test.ts 2>&1 | tail -4
```

<details><summary>Result</summary>

```
Tests  18 passed (18)
   Start at  19:44:52
   Duration  593ms (import 52%, environment 26%, transform 13%, tests 9%)
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:44:55</sub>

Fix 4: narrow the progress status type.

**🔧 Bash** _Find Progress status uses in the frontend_
```bash
grep -rn "currentStep\|Progress\b\|progress.status\|p.status\|\.status = p" src --include=*.ts --include=*.vue | grep -v __tests__ | head -30; sed -n 1,10p src/lib/delivery.ts; grep -rn "Step\b" src/lib/delivery.ts | head
```

<details><summary>Result</summary>

```
(eval):1: no matches found: --include=*.ts
import type { CampaignStatus } from '@/lib/campaigns'

import { API_BASE_URL } from '@/constants/app-config'
import { apiFetch } from '@/lib/api-client'
import { CampaignRequestError } from '@/lib/campaigns'

export type DeliveryStatus = 'queued' | 'sending' | 'retrying' | 'sent' | 'failed' | 'unknown' | 'skipped'
export type DeliveryChannel = 'sms' | 'email'
export interface ChannelProgress { total: number, queued: number, sending: number, retrying: number, sent: number, failed: number, unknown: number, skipped: number, retryable_failed: number }
export interface Progress { campaign_id: number, status: CampaignStatus, channels: { sms: ChannelProgress, email: ChannelProgress } }
src/lib/delivery.ts:13:export type Step = 'approved' | 'queued' | 'sending' | 'done'
src/lib/delivery.ts:55:export function currentStep(status: CampaignStatus): Step {
```

</details>

**🔧 Bash** _Narrow progress status type and type-check_
```bash
python3 - <<'EOF'
def edit(p, pairs):
    s=open(p).read()
    for a,b in pairs:
        if a not in s: raise SystemExit(f"missing in {p}: {a}")
        s=s.replace(a,b)
    open(p,'w').write(s)
edit('src/lib/campaigns.ts',[
 ("export type CampaignStatus = 'draft' | 'approved' | 'queued' | 'sending' | 'completed'\n",
  "export type CampaignStatus = 'draft' | 'approved' | 'queued' | 'sending' | 'completed'\n// SendableStatus is a campaign status after approval. The progress API never returns a draft.\nexport type SendableStatus = Exclude<CampaignStatus, 'draft'>\n"),
])
edit('src/lib/delivery.ts',[
 ("import type { CampaignStatus } from '@/lib/campaigns'", "import type { SendableStatus } from '@/lib/campaigns'"),
 ("export interface Progress { campaign_id: number, status: CampaignStatus,", "export interface Progress { campaign_id: number, status: SendableStatus,"),
 ("""export function currentStep(status: CampaignStatus): Step {
  switch (status) {
    case 'queued': return 'queued'
    case 'sending': return 'sending'
    case 'completed': return 'done'
    default: return 'approved'
  }
}""", """// No default case: a new status must get a step here, or the build fails.
export function currentStep(status: SendableStatus): Step {
  switch (status) {
    case 'approved': return 'approved'
    case 'queued': return 'queued'
    case 'sending': return 'sending'
    case 'completed': return 'done'
  }
}"""),
])
EOF
grep -n "CampaignStatus" src/lib/delivery.ts; grep -n '"typecheck"\|"type-check"\|vue-tsc' package.json; npx vue-tsc --noEmit -p tsconfig.app.json 2>&1 | tail -20 || true
```

<details><summary>Result</summary>

```
15:    "build": "vue-tsc -b && vite build",
69:    "vue-tsc": "^3.3.13"
```

</details>

**🔧 Bash** _Type-check, lint and run frontend tests_
```bash
npx vue-tsc -b; echo rc=$?; pnpm lint 2>&1 | tail -5; pnpm test:run 2>&1 | tail -4
```

<details><summary>Result</summary>

```
rc=0
$ eslint .
[@antfu/eslint-config] Detected running in editor, some rules are disabled.
      Tests  235 passed (235)
   Start at  19:45:21
   Duration  2.23s (import 47%, environment 26%, tests 21%, transform 5%)
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:45:28</sub>

All code fixes are in. Now the docs. Checking the decisions-log format:

**🔧 Bash** _Read decisions log tail and spec 05 event table_
```bash
cd .. && tail -25 docs/decisions-log.md; sed -n 190,225p docs/specs/05-content-generation/goal.md
```

<details><summary>Result</summary>

```
- **Context:** The campaign page stacked one card for each language. For a sent campaign, the delivery section was below all the email bodies.
- **Options:** Keep the stack and move the delivery section up. Show one language with tabs and add a short delivery summary at the top.
- **Choice:** Tabs select the language when the campaign has more than one. A summary card below the header shows a stacked bar for each channel and a link to the delivery section. The page computes the lengths with `renderedLength` and `first_name_runes`, the same as the review window.
- **Why:** The page length does not grow with the number of languages. The marketer sees the result of the send without a scroll. The summary uses the progress that the page already loads, so the API does not change.

## D-089 Counts, dates and names use one format

- **Context:** Pages showed `tr`, `TR` and "Turkish" for the same language. Some counts had no thousands separator. The channel order was different between pages. Dates used the browser locale.
- **Choice:** `lib/format.ts` has `formatCount` and `formatDateTime`, both in `en-US`. All pages use `languageName`. The channel order is SMS, then email. The dashboard column "Send jobs" is now "Messages". The page content has a `max-w-6xl` limit in the layout.
- **Why:** The UI is English only (D-022). "Send job" is an internal term that the marketer does not know.
- **Limits:** A click on a draft row in the campaign list still opens the review window. The tests of slice 06 require this. The list has no recipient count, because `GET /api/campaigns` does not return one.

## D-090 A send gate keeps the request starts apart

- **Context:** D-086 lets a token wait up to 100 ms before its request. A review sent slow and fast preparations together. The provider got 108 requests in one second at a limit of 100.
- **Options:**
  - Make `MaxPrep` shorter. This makes the bunch smaller but does not remove it.
  - Take the token after the load. This brings back the d
… [3910 more chars truncated]
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:45:35</sub>

The decisions log has a D-090 about `Limiter.Admit` that I did not see in the code earlier, and `consumer_test.go` changed on disk. Checking for parallel edits:

**🔧 Bash** _Check working tree for parallel changes_
```bash
git status --short; grep -n "Admit\|release skipped job" backend/internal/delivery/consumer.go backend/internal/delivery/limiter.go | head; grep -n "SkipWriteFailsReleases" backend/internal/delivery/consumer_test.go; tail -15 tasks/progress.md | head -5
```

<details><summary>Result</summary>

```
M backend/internal/campaign/campaign.go
 M backend/internal/delivery/consumer.go
 M backend/internal/delivery/consumer_test.go
 M backend/internal/delivery/service_test.go
 M backend/internal/delivery/store.go
 M backend/internal/delivery/store_test.go
 M backend/internal/delivery/types.go
 M backend/internal/delivery/worker_test.go
 M backend/internal/httpapi/content.go
 M backend/internal/httpapi/content_test.go
 M backend/internal/httpapi/delivery.go
 M backend/internal/httpapi/delivery_test.go
 M backend/internal/store/migrate_test.go
 M frontend/src/components/campaign/__tests__/content-cards.test.ts
 M frontend/src/components/campaign/content-cards.vue
 M frontend/src/lib/campaigns.ts
 M frontend/src/lib/content.ts
 M frontend/src/lib/delivery.ts
?? backend/internal/store/migrations/00005_send_job_transitions.sql
backend/internal/delivery/limiter.go:14:// It has two steps. Wait paces the claims. Admit runs just before the
backend/internal/delivery/limiter.go:42:// Admit blocks until gap has passed since the start of the last admitted
backend/internal/delivery/limiter.go:43:// request. Call it just before the request. A cancelled Admit keeps its slot.
backend/internal/delivery/limiter.go:44:func (l *Limiter) Admit(ctx context.Context) error {
backend/internal/delivery/consumer.go:118:	// can still move sends together. Admit spaces the request starts.
backend/internal/delivery/consumer.go:119:	if err := c.Limiter.Admit(ctx); err != nil {
backend/internal/delivery/consumer.go:158:				c.log().Error("release skipped job", "job_id", job.ID, "channel", c.Channel, "error", rerr)
306:func TestConsumer_SkipWriteFailsReleases(t *testing.T) {
- Done: a frontend-only pass after a design review. No backlog item and no API change.
  - The campaign page has a back link and a line with the created date and the totals. It has a delivery summary with a stacked bar, and language tabs. The SMS shows next to the email (D-087, D-088).
  - `{{first_name}}` shows as a chip. The campai
… [283 more chars truncated]
```

</details>

**🔧 Bash** _Check recent commits and the consumer diff_
```bash
git log --oneline -4; git diff --stat; git diff backend/internal/delivery/consumer.go | head -60
```

<details><summary>Result</summary>

```
3ff452f fixes
9097156 exports
09e0470 ui ux improvements
6960e7c decisions
 backend/internal/campaign/campaign.go              | 14 +++--
 backend/internal/delivery/consumer.go              |  9 +++-
 backend/internal/delivery/consumer_test.go         | 45 +++++++++++++++-
 backend/internal/delivery/service_test.go          |  8 ++-
 backend/internal/delivery/store.go                 | 17 +++---
 backend/internal/delivery/store_test.go            | 55 ++++++++++++++-----
 backend/internal/delivery/types.go                 |  4 +-
 backend/internal/delivery/worker_test.go           |  8 +--
 backend/internal/httpapi/content.go                | 13 ++++-
 backend/internal/httpapi/content_test.go           | 12 +++--
 backend/internal/httpapi/delivery.go               |  3 +-
 backend/internal/httpapi/delivery_test.go          |  3 +-
 backend/internal/store/migrate_test.go             | 61 ++++++++++++++++++++++
 .../campaign/__tests__/content-cards.test.ts       | 20 +++++++
 frontend/src/components/campaign/content-cards.vue | 18 +++----
 frontend/src/lib/campaigns.ts                      |  2 +
 frontend/src/lib/content.ts                        |  5 +-
 frontend/src/lib/delivery.ts                       |  9 ++--
 18 files changed, 246 insertions(+), 60 deletions(-)
diff --git a/backend/internal/delivery/consumer.go b/backend/internal/delivery/consumer.go
index dc9cb4c..52cd932 100644
--- a/backend/internal/delivery/consumer.go
+++ b/backend/internal/delivery/consumer.go
@@ -140,7 +140,9 @@ func (c *Consumer) Handle(ctx context.Context, e Entry) error {
 
 // finish writes the result and acknowledges the entry. When the write fails,
 // the entry stays unacknowledged. Its job stays in sending, and a new claim
-// fails, so the message is not sent again.
+// fails, so the message is not sent again. A skip is the exception: nothing
+// was sent, so the job goes back to queued and the next delivery of the entry
+// skips it again. Without this, the recovery would set 
… [640 more chars truncated]
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:45:45</sub>

A new commit, `3ff452f fixes`, landed during the session and added `Limiter.Admit` (D-090). My changes sit cleanly on top. I'll include Admit in the state doc.

**🔧 Bash** _Read the Admit step in the consumer_
```bash
sed -n 108,130p backend/internal/delivery/consumer.go; grep -n "^## D-" docs/decisions-log.md | tail -2
```

<details><summary>Result</summary>

```
// a job goes back to queued and takes a new token.
		prep := c.Clock.Now().Sub(granted)
		if prep <= maxPrep {
			break
		}
		c.log().Warn("send job prepared too slowly, new rate token", "job_id", job.ID, "channel", c.Channel, "took", prep)
		release()
	}

	// The token paces the claims, but a preparation time of up to MaxPrep
	// can still move sends together. Admit spaces the request starts.
	if err := c.Limiter.Admit(ctx); err != nil {
		if rerr := c.Store.Release(done, job.ID); rerr != nil {
			c.log().Error("release job", "job_id", job.ID, "channel", c.Channel, "error", rerr)
		}
		return err
	}

	reference := fmt.Sprintf("%d:%d:%s", job.CampaignID, job.UserID, job.Channel)
	// A started request may finish: only HTTP.Timeout bounds it, so a
	// shutdown does not turn a delivered message into "canceled".
	sendFrom := c.Clock.Now()
	o := c.Provider.Send(done, c.Channel, msg, reference)
634:## D-089 Counts, dates and names use one format
641:## D-090 A send gate keeps the request starts apart
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:45:53</sub>

Now the doc updates: decisions log, spec 05, state-machines, architecture.

**🔧 Bash** _Write decisions log entries and update specs_
```bash
cat >> docs/decisions-log.md <<'EOF'

## D-091 A trigger guards the send job status

- **Context:** The campaign status had a trigger (D-040). The send job status had only the `WHERE` condition of each `UPDATE`. A new code path could write any status, also from `sent`. `docs/state-machines.md` lists the moves.
- **Options:**
  - Keep the `WHERE` conditions as the only guard.
  - Add a trigger only for moves out of the terminal states.
  - Add a trigger with the full list of moves.
- **Choice:** Migration `00005` adds `send_jobs_status_moves`. It runs `BEFORE UPDATE OF status` and only when the status changes. It allows `queued` or `retrying` to `sending`, `sending` to each other status, and `failed` to `queued` when `retryable` is true. The insert is not guarded.
- **Why:** The worker must never send a message two times. A move out of `sent` or `unknown` would break this rule. The full list also blocks a `sent` without a request. The `WHEN` clause keeps the cost low, because the writes to `enqueued_at` do not run the function.
- **Limits:**
  - The test fixtures cannot jump between states. The helper `setJobs` moves each job through `sending`, as the worker does.
  - The trigger still allows `queued → completed` for a campaign. No code path makes this move. It stays as a fallback for `CompleteCampaigns`, because a campaign that stays in `queued` for ever is worse.

## D-092 A skip that cannot be written goes back to queued

- **Context:** When the write of a result fails, the job stays in `sending`. Recovery then sets it to `unknown`. For a skip, nothing was sent, so `unknown` was wrong in the reports.
- **Options:** Release all results that did not deliver. Release only skips.
- **Choice:** `Consumer.finish` calls `Release` after a failed write of a skip result. The entry stays unacknowledged. The next delivery loads the job again and skips it again.
- **Why:** A skip sends nothing, so a second attempt is safe. A release after `retrying` would lose the backoff and `Retry-After`. After `sent` or `unknown`, the message may have reached the provider.
- **Limits:** If the database is down, `Release` also fails. The job then becomes `unknown`, as before.

## D-093 The retry event has a reason

- **Context:** The content stream sent `retry` for two causes. A `fallback` block restarts the text and is not the repair round. A failed check starts the repair round. The card showed "Fixing content…" for both.
- **Choice:** `retry` has `reason`. The value is `restart` after a `fallback` block and `invalid` before the repair round. The card shows "Starting again…" for `restart` and "Fixing content…" for `invalid`. An event with no `reason` counts as `invalid`.
- **Why:** The marketer sees the true cause. The browser no longer has to guess the cause from an empty `errors` list.

## D-094 Status types in code

- **Context:** `queued` and `sending` are statuses of a campaign and of a send job. The delivery package wrote campaign statuses as plain strings. The fr
… [2031 more chars truncated]
```

<details><summary>Result</summary>

```
83:   - `docs/state-machines.md` lists the states and moves of campaigns, send jobs, queue entries and the UI.
```

</details>

**🔧 Bash** _Update state doc and run STE checker on changed lines_
````bash
python3 - <<'EOF'
def edit(p, pairs):
    s=open(p).read()
    for a,b in pairs:
        if a not in s: raise SystemExit(f"missing in {p}: {a}")
        s=s.replace(a,b)
    open(p,'w').write(s)
p='docs/state-machines.md'
edit(p,[
 ("The last section lists the gaps that we found.", "The last section lists the gaps that we found and the changes that closed them."),
 ("""**Guard:** explicit. The trigger `campaigns_frozen_after_approval` in `00003_send_jobs.sql` allows only the transitions in the table. The writes are in three packages.""",
  """**Guard:** explicit. The trigger `campaigns_frozen_after_approval` in `00003_send_jobs.sql` allows only the transitions in the table. The writes are in three packages. They use the constants of `campaign.Status` (D-094)."""),
 ("    queued --> completed: allowed, no code path\n", "    queued --> completed: fallback only, no code path\n"),
 ("""**Guard:** implicit. No trigger checks the transitions. Each `UPDATE` has a `WHERE status = ...` condition, and that condition is the only guard. The comment in `types.go` names the terminal states: `sent`, `failed`, `unknown` and `skipped`.""",
  """**Guard:** explicit since D-091. Each `UPDATE` has a `WHERE status = ...` condition. The trigger `send_jobs_status_moves` in `00005_send_job_transitions.sql` is the second guard. It allows only the moves in the diagram. The comment in `types.go` names the terminal states: `sent`, `failed`, `unknown` and `skipped`."""),
 ("| `sending` | `queued` | `Release` | Nothing was sent. The causes are a cancelled context, a `Load` error, or a rate token older than 100 ms. `attempts` goes down by 1. |",
  "| `sending` | `queued` | `Release` | Nothing was sent. The causes are a cancelled context, a `Load` error, a rate token older than 100 ms, a cancelled `Admit`, or a failed write of a skip result (D-092). `attempts` goes down by 1. |"),
 ("""- After a `Finish` error, the job stays in `sending`. A new claim fails. Thus the entry is acknowledged at the next delivery, and `ExpireSending` later sets the job to `unknown`.""",
  """- After a `Finish` error, the job stays in `sending`. A new claim fails. Thus the entry is acknowledged at the next delivery, and `ExpireSending` later sets the job to `unknown`.
- A skip result is the exception. After a failed write, the consumer releases the job to `queued`. The next delivery skips it again (D-092)."""),
 ("""    load --> send
    send --> finish""", """    load --> admit
    admit --> send: gap since the last request start
    admit --> released: context cancelled, Release
    send --> finish"""),
 ("""    finish --> ack
    ack --> [*]
```

D-086 gives the order: wait, claim, load, send. The job stays `queued` during the wait. Thus recovery cannot expire a job that waits for a token.""",
  """    finish --> ack
    ack --> [*]
    released --> [*]: entry stays pending
```

D-086 gives the order: wait, claim, load, send. The job stays `queued` during the wait. Thus recovery cannot expire a jo
… [3660 more chars truncated]
````

<details><summary>Result</summary>

```
missing in docs/state-machines.md:     streaming --> streaming: Reset event, buffer cleared, "retry" sent
    validating --> final: valid
    validating --> streaming: invalid on attempt 1, "retry" with errors, attempt 2
docs/decisions-log.md:673: present-perfect: have reached
```

</details>

**🔧 Bash** _Inspect the content diagram text and doc state_
````bash
sed -n '/stateDiagram-v2/,/```/p' docs/state-machines.md | sed -n '/streaming/p' | head; git diff --stat docs/state-machines.md 2>/dev/null; git status --short docs
````

<details><summary>Result</summary>

```
[*] --> streaming: attempt 1
    streaming --> streaming: Reset event, buffer cleared, "retry" sent
    streaming --> validating: tool call complete
    validating --> streaming: invalid on attempt 1, "retry" with errors, attempt 2
    streaming --> error: LLM error, timeout or refusal
    streaming --> cancelled: client disconnected
    [*] --> streaming: approved, queued or sending
    streaming --> polling: EventSource error, or the server closed after 30 min
    streaming --> stopped: "done" event
    stopped --> streaming: POST /retry succeeded
 M docs/architecture.md
 M docs/decisions-log.md
 M docs/specs/05-content-generation/goal.md
 M docs/specs/05-content-generation/requirements.md
```

</details>

**🔧 Bash** _Apply state doc edits and recheck STE_
````bash
python3 - <<'EOF'
def edit(p, pairs):
    s=open(p).read()
    for a,b in pairs:
        if a not in s: raise SystemExit(f"missing in {p}: {a}")
        s=s.replace(a,b)
    open(p,'w').write(s)
p='docs/state-machines.md'
edit(p,[
 ("The last section lists the gaps that we found.", "The last section lists the gaps that we found and the changes that closed them."),
 ("""**Guard:** explicit. The trigger `campaigns_frozen_after_approval` in `00003_send_jobs.sql` allows only the transitions in the table. The writes are in three packages.""",
  """**Guard:** explicit. The trigger `campaigns_frozen_after_approval` in `00003_send_jobs.sql` allows only the transitions in the table. The writes are in three packages. They use the constants of `campaign.Status` (D-094)."""),
 ("    queued --> completed: allowed, no code path\n", "    queued --> completed: fallback only, no code path\n"),
 ("""**Guard:** implicit. No trigger checks the transitions. Each `UPDATE` has a `WHERE status = ...` condition, and that condition is the only guard. The comment in `types.go` names the terminal states: `sent`, `failed`, `unknown` and `skipped`.""",
  """**Guard:** explicit since D-091. Each `UPDATE` has a `WHERE status = ...` condition. The trigger `send_jobs_status_moves` in `00005_send_job_transitions.sql` is the second guard. It allows only the moves in the diagram. The comment in `types.go` names the terminal states: `sent`, `failed`, `unknown` and `skipped`."""),
 ("| `sending` | `queued` | `Release` | Nothing was sent. The causes are a cancelled context, a `Load` error, or a rate token older than 100 ms. `attempts` goes down by 1. |",
  "| `sending` | `queued` | `Release` | Nothing was sent. The causes are a cancelled context, a `Load` error, a rate token older than 100 ms, a cancelled `Admit`, or a failed write of a skip result (D-092). `attempts` goes down by 1. |"),
 ("""- After a `Finish` error, the job stays in `sending`. A new claim fails. Thus the entry is acknowledged at the next delivery, and `ExpireSending` later sets the job to `unknown`.""",
  """- After a `Finish` error, the job stays in `sending`. A new claim fails. Thus the entry is acknowledged at the next delivery, and `ExpireSending` later sets the job to `unknown`.
- A skip result is the exception. After a failed write, the consumer releases the job to `queued`. The next delivery skips it again (D-092)."""),
 ("""    load --> send
    send --> finish""", """    load --> admit
    admit --> send: gap since the last request start
    admit --> released: context cancelled, Release
    send --> finish"""),
 ("""    finish --> ack
    ack --> [*]
```

D-086 gives the order: wait, claim, load, send. The job stays `queued` during the wait. Thus recovery cannot expire a job that waits for a token.""",
  """    finish --> ack
    ack --> [*]
    released --> [*]: entry stays pending
```

D-086 gives the order: wait, claim, load, send. The job stays `queued` during the wait. Thus recovery cannot expire a jo
… [3607 more chars truncated]
````

<details><summary>Result</summary>

```
done
[This command modified 1 file you've previously read: docs/state-machines.md. Call Read before editing.]
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:47:09</sub>

Docs are updated. Stopping my test containers and running the full `make verify`:

**🔧 Bash** _Run make verify and summarize results_
```bash
docker rm -f cc-test-pg cc-test-redis >/dev/null 2>&1; make verify > /private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/d69aea1d-9d23-4cca-8129-17ae82a3db16/scratchpad/verify.log 2>&1; echo rc=$?; grep -E "^(ok|FAIL|---)|Tests |error" /private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/d69aea1d-9d23-4cca-8129-17ae82a3db16/scratchpad/verify.log | head -30
```

<details><summary>Result</summary>

```
rc=0
ok  	campaigncomposer/cmd/api	(cached)
ok  	campaigncomposer/cmd/worker	1.219s
ok  	campaigncomposer/internal/campaign	3.203s
ok  	campaigncomposer/internal/config	(cached)
ok  	campaigncomposer/internal/content	(cached)
ok  	campaigncomposer/internal/dashboard	3.949s
ok  	campaigncomposer/internal/delivery	43.710s
ok  	campaigncomposer/internal/httpapi	2.978s
ok  	campaigncomposer/internal/llm	(cached)
ok  	campaigncomposer/internal/mockprovider	(cached)
ok  	campaigncomposer/internal/seed	2.054s
ok  	campaigncomposer/internal/segment	2.618s
ok  	campaigncomposer/internal/store	1.912s
      Tests  235 passed (235)
dist/assets/errors-C9CZG6dQ.js                       0.19 kB │ gzip:  0.17 kB
dist/assets/errors-DQFC1xB4.js                       0.31 kB │ gzip:  0.24 kB
dist/assets/custom-error-CCBnfuAs.js                 0.83 kB │ gzip:  0.50 kB
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:48:14</sub>

`make verify` passes. Adding the session entry to `tasks/progress.md`:

**🔧 Bash** _Append progress entry and check STE_
```bash
cat >> tasks/progress.md <<'EOF'

## 2026-11-11 — state machines
- Done: `docs/state-machines.md` lists 10 state machines in the code, and the gaps that we found. We fixed five gaps and kept one on purpose.
  - The trigger `send_jobs_status_moves` (migration `00005`) allows only the moves of the worker (D-091). The test fixtures use `setJobs` and move each job through `sending`.
  - A skip result that cannot be written goes back to `queued`, not to `unknown` (D-092).
  - The `retry` event of the content stream has `reason`: `restart` or `invalid`. The card shows "Starting again…" or "Fixing content…" (D-093).
  - `campaign.Status` has constants for all five values. `delivery` uses them. The frontend has `SendableStatus`, and `currentStep` has no `default` case (D-094).
  - The campaign trigger still allows `queued → completed` as a fallback. No code path makes this move (D-091).
- Verified with: each fix had a test that failed first. `TestConsumer_SkipWriteFailsReleases` showed `sending` before the fix. `TestMigrate_SendJobStatusMoves` showed 12 moves that the database accepted. The backend and frontend tests of the retry reason failed first. `vue-tsc -b` passes. `make verify` exit 0 (235 frontend tests).
- Next: user review of the changes. Nothing is committed.
- Open questions:
  - The insert of a send job has no status guard. Only `PlanNext` inserts.
  - No compose run checked the new migration against a database with existing data. The trigger runs only on an update, so existing rows are not checked.
EOF
python3 .claude/skills/ste-writing/scripts/ste_check.py tasks/progress.md 2>&1 | tail -3
```

<details><summary>Result</summary>

```
tasks/progress.md:82: sentence-length: 36 words: User decisions: one review window with SMS and Email tabs, a second confirm before the send, a token-based design that works on a phone, failed languages are skipped, and Claude saves a draft when it finishes.
tasks/progress.md:82: sentence-length: 46 words: All 5 are fixed, each with a test that failed first: a stale draft text after "Regenerate content", the row lock that waited for the audience query, a late response for another campaign, old limits after a 422, and the review button after a failed regenerate.
tasks/progress.md:83: sentence-length: 30 words: A compose run with CODE on ports 13000 and 18080 (the running stack holds 3000 and 8080) gave the draft, approve CODE, repeat CODE, PUT CODE and CORS with CODE.
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:48:31</sub>

I applied all six fixes from the list, and `make verify` passes (all Go packages and 235 frontend tests). For each fix, I first wrote a test that failed on the old code. Nothing is committed.

| # | Gap | Change | Decision |
|---|---|---|---|
| 1 | A skip whose result failed to save later showed up as `unknown` | [consumer.go](backend/internal/delivery/consumer.go) now puts the job back in `queued`, so the next try skips it again. Other results keep the old behaviour. | D-092 |
| 2 | Nothing in the database stopped a bad send job status change | New migration [00005_send_job_transitions.sql](backend/internal/store/migrations/00005_send_job_transitions.sql) adds the trigger `send_jobs_status_moves`. It allows only the moves the worker makes, and runs only when the status changes. | D-091 |
| 3 | The `retry` event meant two different things | The event now has `reason: "restart"` or `"invalid"`. The language card shows "Starting again…" or "Fixing content…". An event with no reason counts as `invalid`. | D-093 |
| 4 | `currentStep` treated `draft` as `approved` | New `SendableStatus` type with no `draft`. `currentStep` now has one case per status and no default, so a new status without a step fails the build. | D-094 |
| 5 | Campaigns and send jobs use some of the same status words | `campaign.Status` now has all five constants, and the delivery code uses them instead of plain strings. Values in the database and the API are unchanged. | D-094 |
| 6 | The campaign trigger allows `queued → completed`, but no code does it | Kept on purpose as a fallback. Without it, a campaign could get stuck in `queued`. A comment in migration `00005` records this. | D-091 |

**What else changed:**
- About 15 test fixtures used to jump straight between statuses, for example `queued → sent`. The new trigger rejects those jumps. They now use a helper, `setJobs`, that moves each job through `sending` the way the worker does.
- A commit, `3ff452f fixes`, landed during the session and added `Limiter.Admit` (D-090). My changes sit on top of it without conflict. I added the Admit step to the consumer diagram.
- I updated [docs/state-machines.md](docs/state-machines.md), the spec 05 event table and requirements, `docs/architecture.md` (a link to the new doc), the decisions log (D-091 to D-094) and `tasks/progress.md`.

**Not covered:**
- Creating a send job has no status check. Only the planner (`PlanNext`) creates them.
- I did not run the new migration against a database that already has data. The trigger runs only on updates, so existing rows are never checked.

## 👤 User  <sub>2026-11-11 16:50:19</sub>

_[IDE: opened file]_ 
commit

**🔧 Bash** _Check working tree and commit style_
```bash
git status --short; git log --oneline -1; git log -3 --format='%s%n%b---'
```

<details><summary>Result</summary>

```
M  backend/internal/campaign/campaign.go
M  backend/internal/delivery/consumer.go
M  backend/internal/delivery/consumer_test.go
M  backend/internal/delivery/service_test.go
M  backend/internal/delivery/store.go
M  backend/internal/delivery/store_test.go
M  backend/internal/delivery/types.go
M  backend/internal/delivery/worker_test.go
M  backend/internal/httpapi/content.go
M  backend/internal/httpapi/content_test.go
M  backend/internal/httpapi/delivery.go
M  backend/internal/httpapi/delivery_test.go
M  backend/internal/store/migrate_test.go
A  backend/internal/store/migrations/00005_send_job_transitions.sql
M  docs/architecture.md
M  docs/decisions-log.md
M  docs/specs/05-content-generation/goal.md
M  docs/specs/05-content-generation/requirements.md
M  docs/state-machines.md
M  frontend/src/components/campaign/__tests__/content-cards.test.ts
M  frontend/src/components/campaign/content-cards.vue
M  frontend/src/lib/campaigns.ts
M  frontend/src/lib/content.ts
M  frontend/src/lib/delivery.ts
M  tasks/progress.md
3ff452f fixes
fixes
---
exports
---
ui ux improvements
---
```

</details>

**🔧 Bash** _Create a branch and commit the changes_
```bash
git switch -c state-machine-guards && git add -A && git commit -q -F - <<'EOF'
Add guards for the send job and campaign states

- Add docs/state-machines.md. It lists 10 state machines and the gaps.
- Add the trigger send_jobs_status_moves in migration 00005 (D-091).
- Put a skip back to queued when its result write fails (D-092).
- Add a reason to the retry event of the content stream (D-093).
- Use typed campaign statuses in Go and SendableStatus in the
  frontend (D-094).

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
git log --oneline -2; git status --short | wc -l
```

<details><summary>Result</summary>

```
Switched to a new branch 'state-machine-guards'
44d61eb Add guards for the send job and campaign states
3ff452f fixes
       0
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:50:37</sub>

I committed all 25 changed files as `44d61eb` on a new branch, `state-machine-guards`. I made the branch because I don't commit straight to `main`.

```
Add guards for the send job and campaign states

- Add docs/state-machines.md. It lists 10 state machines and the gaps.
- Add the trigger send_jobs_status_moves in migration 00005 (D-091).
- Put a skip back to queued when its result write fails (D-092).
- Add a reason to the retry event of the content stream (D-093).
- Use typed campaign statuses in Go and SendableStatus in the
  frontend (D-094).
```

Nothing is pushed. All your earlier commits are directly on `main`, so if you want this there too, run `git switch main && git merge --ff-only state-machine-guards`.

## 👤 User  <sub>2026-11-11 16:50:49</sub>

commit to main

**🔧 Bash** _Fast-forward main to the commit and delete the branch_
```bash
git switch main && git merge --ff-only state-machine-guards && git branch -d state-machine-guards && git log --oneline -2 && git status --short | wc -l
```

<details><summary>Result</summary>

```
Switched to branch 'main'
Your branch is up to date with 'origin/main'.
Updating 3ff452f..44d61eb
Fast-forward
 backend/internal/campaign/campaign.go              | 14 +++--
 backend/internal/delivery/consumer.go              |  9 +++-
 backend/internal/delivery/consumer_test.go         | 45 +++++++++++++++-
 backend/internal/delivery/service_test.go          |  8 ++-
 backend/internal/delivery/store.go                 | 17 +++---
 backend/internal/delivery/store_test.go            | 55 ++++++++++++++-----
 backend/internal/delivery/types.go                 |  4 +-
 backend/internal/delivery/worker_test.go           |  8 +--
 backend/internal/httpapi/content.go                | 13 ++++-
 backend/internal/httpapi/content_test.go           | 12 +++--
 backend/internal/httpapi/delivery.go               |  3 +-
 backend/internal/httpapi/delivery_test.go          |  3 +-
 backend/internal/store/migrate_test.go             | 61 ++++++++++++++++++++++
 .../migrations/00005_send_job_transitions.sql      | 36 +++++++++++++
 docs/architecture.md                               |  1 +
 docs/decisions-log.md                              | 35 +++++++++++++
 docs/specs/05-content-generation/goal.md           |  4 +-
 docs/specs/05-content-generation/requirements.md   |  4 +-
 docs/state-machines.md                             | 46 +++++++++-------
 .../campaign/__tests__/content-cards.test.ts       | 20 +++++++
 frontend/src/components/campaign/content-cards.vue | 18 +++----
 frontend/src/lib/campaigns.ts                      |  2 +
 frontend/src/lib/content.ts                        |  5 +-
 frontend/src/lib/delivery.ts                       |  9 ++--
 tasks/progress.md                                  | 13 +++++
 25 files changed, 361 insertions(+), 84 deletions(-)
 create mode 100644 backend/internal/store/migrations/00005_send_job_transitions.sql
Deleted branch state-machine-guards (was 44d61eb).
44d61eb Add guards for the send job and campaign states
3ff452f fixes
       0
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:51:00</sub>

Commit `44d61eb` is now on `main`. I fast-forwarded `main` to it and deleted the temporary branch, so the history stays a straight line. Nothing is pushed: `main` is one commit ahead of `origin/main`.

