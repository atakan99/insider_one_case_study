# Rate limiting, review dialog, and mock docs

- Session: `22a0e40d-eeb7-4ae5-9612-495cb6c67262`
- Model: `claude-fable-5-2`
- Time (UTC): 2026-11-11 09:39:41 → 2026-11-11 10:45:12
- Subagent runs: 0

---

## 👤 User  <sub>2026-11-11 09:39:41</sub>

_[IDE: opened file]_ 



The project implements the core flow in case.md, but I found two defects and one delivery documentation gap.
1. [P2] Rate limiting happens too early. In [consumer.go (line 48)](~/Desktop/projects/case-studies/insider_one/backend/internal/delivery/consumer.go:48), consumers acquire a token before claiming and loading jobs from Postgres. Database delays can hold already permitted sends, then release them together alongside newly permitted sends. This can exceed the required 100 requests per second per channel. Enforce pacing immediately before the provider request and test actual request timestamps under database delays.
2. [P2] Saving can silently discard newer edits. [review-dialog.vue (line 220)](~/Desktop/projects/case-studies/insider_one/frontend/src/components/campaign/review-dialog.vue:220) snapshots the form, awaits the save, then resets all fields from the response. Inputs remain editable during that request. If the marketer continues typing, those newer edits disappear when the response arrives. Disable editing while saving or preserve edits made after submission.
3. [P3] Default mock behavior needs clearer disclosure. Default startup uses a mock that returns the same Istanbul segment and 30% discount content regardless of the brief ([mockdefault.go (line 13)](~/Desktop/projects/case-studies/insider_one/backend/cmd/api/mockdefault.go:13)). Mock development is explicitly allowed, but the [README (line 5)](~/Desktop/projects/case-studies/insider_one/README.md:5) does not explain this limitation or how to enable Claude. Document both so reviewers can assess the main AI requirement.

**🔧 Skill** ```json
{
  "skill": "superpowers:receiving-code-review"
}
```

<details><summary>Result</summary>

```
Launching skill: superpowers:receiving-code-review
```

</details>

> _[skill instructions loaded: receiving-code-review]_

**🔧 Read** `~/Desktop/projects/case-studies/insider_one/backend/internal/delivery/consumer.go`

<details><summary>Result</summary>

```
1	package delivery
2	
3	import (
4		"context"
5		"fmt"
6		"log/slog"
7	)
8	
9	// Consumer handles the queue entries of one channel.
10	type Consumer struct {
11		Store       *Store
12		Queue       *Queue
13		Provider    *Provider
14		Limiter     *Limiter
15		Channel     Channel
16		Clock       Clock
17		MaxAttempts int
18		Jitter      func() float64
19		Log         *slog.Logger
20		Metrics     *Metrics // nil means no metrics
21	
22		// Hooks for tests. Both are nil in production.
23		afterClaim func()
24		finishErr  func() error
25	}
26	
27	func (c *Consumer) log() *slog.Logger {
28		if c.Log != nil {
29			return c.Log
30		}
31		return slog.Default()
32	}
33	
34	// Handle runs one entry: wait, claim, load, send, write the result, acknowledge.
35	// It returns an error only when the entry must stay unacknowledged.
36	func (c *Consumer) Handle(ctx context.Context, e Entry) error {
37		// The writes after a claim use a context that cannot be cancelled: the
38		// result of an attempt must reach Postgres, also during a shutdown.
39		done := context.WithoutCancel(ctx)
40	
41		// An entry for a job that cannot be claimed must not use a rate token.
42		// On an error here, the normal path decides.
43		if ok, err := c.Store.Claimable(ctx, e.JobID, c.Clock.Now()); err == nil && !ok {
44			return c.Queue.Ack(done, c.Channel, e.StreamID)
45		}
46	
47		waitFrom := c.Clock.Now()
48		err := c.Limiter.Wait(ctx)
49		c.Metrics.Waited(c.Channel, c.Clock.Now().Sub(waitFrom))
50		if err != nil {
51			return err
52		}
53		job, ok, err := c.Store.Claim(ctx, e.JobID, c.Clock.Now())
54		if err != nil {
55			return err
56		}
57		if !ok {
58			return c.Queue.Ack(done, c.Channel, e.StreamID)
59		}
60		if c.afterClaim != nil {
61			c.afterClaim()
62		}
63		if err := ctx.Err(); err != nil {
64			if rerr := c.Store.Release(done, job.ID); rerr != nil {
65				c.log().Error("release job", "job_id", job.ID, "channel", c.Channel, "error", rerr)
66			}
67			return err
68		}
69	
70		msg, skip, err := 
… [2065 more chars truncated]
```

</details>

**🔧 Read** `~/Desktop/projects/case-studies/insider_one/frontend/src/components/campaign/review-dialog.vue`

<details><summary>Result</summary>

```
1	<script setup lang="ts">
2	import { computed, reactive, ref, watch } from 'vue'
3	import { toast } from 'vue-sonner'
4	
5	import type { Campaign, ContentInput } from '@/lib/campaigns'
6	import type { Content } from '@/lib/content'
7	
8	import { Alert, AlertDescription } from '@/components/ui/alert'
9	import {
10	  AlertDialog,
11	  AlertDialogContent,
12	  AlertDialogDescription,
13	  AlertDialogFooter,
14	  AlertDialogHeader,
15	  AlertDialogTitle,
16	} from '@/components/ui/alert-dialog'
17	import { Button } from '@/components/ui/button'
18	import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
19	import { Input } from '@/components/ui/input'
20	import { Label } from '@/components/ui/label'
21	import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
22	import { Spinner } from '@/components/ui/spinner'
23	import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
24	import { Textarea } from '@/components/ui/textarea'
25	import { approveCampaign, CampaignRequestError, getCampaign, getSampleNames, languageName, saveContents } from '@/lib/campaigns'
26	import { checkContent, LIMITS, renderedLength } from '@/lib/limits'
27	
28	import MessagePreview from './message-preview.vue'
29	
30	type Field = keyof Content
31	type Channel = 'sms' | 'email'
32	
33	const props = defineProps<{ campaignId: number }>()
34	const emit = defineEmits<{ approved: [id: number] }>()
35	const open = defineModel<boolean>('open', { default: false })
36	
37	const FALLBACK_NAME = 'Alex'
38	const VAR_LABEL = '{{first_name}}'
39	const FIELDS: Field[] = ['sms', 'email_subject', 'email_body']
40	
41	const campaign = ref<Campaign | null>(null)
42	const loadError = ref('')
43	const edits = reactive<Record<string, Content>>({})
44	const saved = ref<Record<string, Content>>({})
45	const samples = ref<Record<string, string[]>>({})
46	const sampleFailed = ref(false)
47	const chosenName = reactive<Rec
… [18534 more chars truncated]
```

</details>

**🔧 Read** `~/Desktop/projects/case-studies/insider_one/backend/cmd/api/mockdefault.go`

<details><summary>Result</summary>

```
1	package main
2	
3	import (
4		"time"
5	
6		"campaigncomposer/internal/content"
7		"campaigncomposer/internal/httpapi"
8		"campaigncomposer/internal/llm"
9		"campaigncomposer/internal/segment"
10	)
11	
12	// mockDefault answers an unscripted mock call by tool name.
13	func mockDefault(req llm.Request) llm.MockStep {
14		input := []byte(llm.ExampleRuleJSON)
15		if req.Tool.Name == content.ToolName {
16			lang := "en"
17			if len(req.Turns) > 0 {
18				if l, ok := content.LanguageOf(req.Turns[0].User); ok {
19					lang = l
20				}
21			}
22			input = content.MockContent(lang)
23		}
24		return llm.MockStep{Response: llm.Response{
25			ToolUse:   &llm.ToolUse{ID: "toolu_mock", Name: req.Tool.Name, Input: input},
26			Assistant: "mock-assistant",
27		}}
28	}
29	
30	// newContentGenerator builds the handler of POST /api/content/generate.
31	func newContentGenerator(client llm.Client, allow segment.Allowlist, count httpapi.AudienceCounter) *httpapi.ContentGenerator {
32		return &httpapi.ContentGenerator{
33			LLM:        client,
34			Allow:      allow,
35			Count:      count,
36			Now:        time.Now,
37			Clock:      time.Now,
38			LLMTimeout: 45 * time.Second,
39			DraftEvery: 100 * time.Millisecond,
40		}
41	}
42
```

</details>

**🔧 Read** `~/Desktop/projects/case-studies/insider_one/README.md`

<details><summary>Result</summary>

````
1	# Smart Campaign Composer
2	
3	A marketer writes a brief for a campaign. Claude makes a segment rule and SMS and email content from the brief. The backend validates the rule and counts the audience. The marketer edits and approves the content. Then a queue sends the messages to a mock provider, and a dashboard shows the delivery status. No real person gets a message.
4	
5	## Run
6	
7	```bash
8	docker compose up
9	```
10	
11	- Frontend: http://localhost:3000
12	- API: http://localhost:8080 (`/healthz`, `/readyz`)
13	- Mock provider: http://localhost:8081
14	
15	For frontend work with hot reload, run `pnpm dev` in `frontend/` (http://localhost:5173).
16	
17	## Test
18	
19	```bash
20	make test
21	```
22	
23	## Status
24	
25	See [`tasks/backlog.json`](tasks/backlog.json).
26
````

</details>

**🔧 Read** `~/Desktop/projects/case-studies/insider_one/tasks/progress.md`

<details><summary>Result</summary>

```
1	# Progress
2	
3	Append one entry per session. Use the template below. Put the newest entry at the bottom.
4	
5	<!-- Template:
6	## YYYY-MM-DD — <slice id>
7	- Done:
8	- Verified with:
9	- Next:
10	- Open questions:
11	-->
12	
13	## 2026-11-06 — 00-skeleton
14	- Done: Task 1 of `docs/plans/00-skeleton.md`. Moved the case brief to `docs/case.md` and the frontend-trim spec to `docs/specs/01-frontend-trim/`. Wrote `.gitignore`, `CLAUDE.md`, `DECISIONS.md` (D-001 to D-006), `README.md`, `docs/architecture.md`, `tasks/backlog.json`, `tasks/progress.md` and `ai-sessions/.gitkeep`.
15	- Verified with: `git check-ignore` (prints `.env`, `.env.local`, then `exit=1`). `python3 -m json.tool tasks/backlog.json` (prints `ok`).
16	- Next: Task 2 (ste-writing skill, executor and reviewer subagents).
17	- Open questions: None.
18	
19	## 2026-11-06 — 00-skeleton
20	- Done: Task 2 of `docs/plans/00-skeleton.md`. Added the `ste-writing` skill (`SKILL.md`, `references/rules.md`, `references/terms.md`, `scripts/ste_check.py` and its tests). Added the `executor` and `reviewer` subagents. Changed the Writing style line in `CLAUDE.md`. Deleted the template git hooks in `.git/hooks/` (user decision). The reviewer found 6 defects in `ste_check.py`. All 6 are fixed, and each one has a regression test.
21	- Verified with: `python3 -B -m unittest discover -s .claude/skills/ste-writing/scripts` (11 tests, OK). The session lists `executor`, `reviewer` and `ste-writing`.
22	- Next: Task 3 (Go module, config, health endpoints, two binaries).
23	- Open questions: `.gitignore` does not include `__pycache__/`. The tests write it when they run without `-B`.
24	
25	## 2026-11-06 — 00-skeleton
26	- Done: Task 3 of `docs/plans/00-skeleton.md`. Added the Go module `campaigncomposer` (go 1.27, `pgx/v5` v5.12.0), `internal/config`, `internal/httpapi` (health endpoints), `cmd/api` and `cmd/mockprovider`. The reviewer found 3 defects in the two `main.go` files. All 3 are fixed: a 2 s timeout on the startup DB
… [15985 more chars truncated]
```

</details>

**🔧 Bash** ```bash
ls backend/internal/delivery backend/cmd/api backend/cmd/worker && cat backend/internal/delivery/limiter.go && grep -n -i "limiter\|rate token\|token before\|Claimable" docs/decisions-log.md | head -40 && grep -rn "LLM_MODE\|ANTHROPIC" --include=*.go --include=*.yml --include=.env.example --include=Makefile . 2>/dev/null | grep -v _test | grep -v node_modules | head -40
```

<details><summary>Result</summary>

```
backend/cmd/api:
main.go
mockdefault.go
mockdefault_test.go

backend/cmd/worker:
main.go
main_test.go

backend/internal/delivery:
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
	return &Limiter{rl: rate.NewLimiter(rate.Limit(perSec), 1), clock: clock}
}

// Pause makes Wait block for at least d from now. A shorter pause does not
// shorten a longer one.
func (l *Limiter) Pause(d time.Duration) {
	until := l.clock.Now().Add(d)
	l.mu.Lock()
	if until.After(l.pausedUntil) {
		l.pausedUntil = until
	}
	l.mu.Unlock()
}

// Wait blocks until a send is allowed. It returns ctx.Err() when ctx is cancelled.
// A pause that starts while Wait sleeps for a token is honoured too.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		if err := l.waitPause(ctx); err != nil {
			return err
		}
		now := l.clock.Now()
		r := l.rl.ReserveN(now, 1)
		if err := l.clock.Sleep(ctx, r.DelayFrom(now)); err != nil {
			r.CancelAt(now)
			return err
		}
		if !l.paused() {
			return nil
		}
	}
}

func (l *Limiter) paused() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.clock.Now().Before(l.pausedUntil)
}

func (l *Limiter) waitPause(ctx context.Context) error {
	for {
		l.mu.Lock()
		until := l.pausedUntil
		l.mu.Unlock()
		d := until.Sub(l.clock.Now())
		if d <= 0 {
			return nil
		}
		if err := l.clock.Sleep(ctx, d); err != nil {
			return err
		}
	}
}
10:-
… [2207 more chars truncated]
```

</details>

**🔧 Bash** ```bash
sed -n 136,146p docs/decisions-log.md; sed -n 340,376p docs/decisions-log.md; tail -30 docs/decisions-log.md; grep -rn "LLM_MODE\|ANTHROPIC" backend/internal/config backend/cmd docker-compose.yml .env.example Makefile | grep -v _test | head -30; grep -n "func Test" backend/internal/delivery/consumer_test.go backend/internal/delivery/limiter_test.go; cat backend/internal/delivery/fakeclock_test.go; grep -n "ExampleRuleJSON" -A12 backend/internal/llm/*.go | head -40; ls frontend/src/components/campaign/ frontend/src/components/campaign/__tests__ 2>/dev/null
```

<details><summary>Result</summary>

```
- **Why:** The case forbids duplicates, so a possible missed message is better than a possible duplicate. SendGrid has no reliable lookup that the worker can use before a retry. Clear rules in the mock make each path testable.

## D-020 Token bucket from `golang.org/x/time/rate`

- **Context:** The mock provider needs a rate limit for each channel. The worker in slice 07 needs one too.
- **Options:** Our own token bucket, `golang.org/x/time/rate`, a Redis-based limiter.
- **Choice:** `golang.org/x/time/rate`, one limiter for each channel. The default is 100 per second with a burst of 10.
- **Why:** The package is maintained by the Go team and takes an explicit time, so tests can use a fake clock. The mock is one process, so it needs no shared state in Redis.

## D-021 `X-Mock-Fail` header for exact tests


## D-049 32 consumers for each channel

- **Context:** The target is 100 requests per second for each channel. Each `slow` response holds one consumer for 3 s.
- **Options:** Few consumers. 32 consumers. One goroutine for each send job.
- **Choice:** `SEND_CONSUMERS` defaults to 32 for each channel.
- **Why:** The rate limiter is the cap, not the consumers. 32 consumers keep 100 requests per second even when some wait for a `slow` response.

## D-050 The worker plans approved campaigns

- **Context:** The approval is one Postgres transaction (D-040). A push to Redis in the same request can fail on its own.
- **Options:** The approval pushes to Redis. The worker plans the approved campaigns.
- **Choice:** The worker finds the `approved` campaigns and writes the `send_jobs` rows. The unique key on campaign, user and channel blocks a second plan. Then the campaign becomes `queued`.
- **Why:** The approval stays one transaction and does not need Redis. A crash after the plan and before the push is safe, because the recovery pushes the send jobs again (FR-27).

## D-051 One stream for each channel

- **Context:** A slow SMS provider must not stop the email channel.
- *
… [10217 more chars truncated]
```

</details>

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

**🔧 Bash** ```bash
sed -n 1,105p backend/internal/delivery/consumer_test.go; sed -n 225,325p backend/internal/delivery/consumer_test.go; grep -n "Claimable" -r backend --include="*.go"; grep -n "ExampleRuleJSON =" -A25 backend/internal/llm/*.go | head -45; grep -n "func MockContent" -A30 backend/internal/content/*.go | head -50; sed -n 30,60p docker-compose.yml; cat .env.example; grep -n "Clock\b" -A8 backend/internal/delivery/types.go | head -40
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
	"regexp"
	"strings"
	"sync"
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

func (e *consumerEnv) pending(t *testing.T) int64 {
	t.Helper()
	p, err := e.queue.rd
… [8761 more chars truncated]
```

</details>

**🔧 Bash** ```bash
sed -n 150,260p backend/internal/delivery/store.go; grep -n -i "stale\|unknown" backend/internal/delivery/store.go backend/internal/delivery/worker.go | head -40; grep -n "RetryAfter" -B3 -A8 backend/internal/delivery/provider.go | head -60; grep -n "func newMock\|func providerStats" -A22 backend/internal/delivery/*_test.go | head -70
```

<details><summary>Result</summary>

```
return fmt.Errorf("mark enqueued: %w", err)
	}
	return nil
}

const jobColumns = `id, campaign_id, user_id, channel, language, status, attempts, last_error, retryable, updated_at`

func scanJob(row pgx.Row) (Job, error) {
	var j Job
	err := row.Scan(&j.ID, &j.CampaignID, &j.UserID, &j.Channel, &j.Language, &j.Status,
		&j.Attempts, &j.LastError, &j.Retryable, &j.UpdatedAt)
	return j, err
}

// Claim moves a queued or due retrying job to sending and adds 1 to attempts.
// It also moves the campaign from queued to sending. ok is false when no row changed.
func (s *Store) Claim(ctx context.Context, id int64, now time.Time) (Job, bool, error) {
	var j Job
	var ok bool
	err := s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		var e error
		j, e = scanJob(tx.QueryRow(ctx, `UPDATE send_jobs SET status = 'sending', attempts = attempts + 1, updated_at = $2
WHERE id = $1 AND (status = 'queued' OR (status = 'retrying' AND next_attempt_at <= $2))
RETURNING `+jobColumns, id, now))
		if errors.Is(e, pgx.ErrNoRows) {
			return nil
		}
		if e != nil {
			return e
		}
		ok = true
		_, e = tx.Exec(ctx, `UPDATE campaigns SET status = 'sending', updated_at = $2 WHERE id = $1 AND status = 'queued'`, j.CampaignID, now)
		return e
	})
	if err != nil {
		return Job{}, false, fmt.Errorf("claim job: %w", err)
	}
	return j, ok, nil
}

// Claimable reports whether Claim could take the job now. It only saves rate
// tokens: the UPDATE in Claim stays the gate.
func (s *Store) Claimable(ctx context.Context, id int64, now time.Time) (bool, error) {
	var ok bool
	err := s.inTx(ctx, pgx.ReadOnly, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM send_jobs WHERE id = $1
AND (status = 'queued' OR (status = 'retrying' AND next_attempt_at <= $2)))`, id, now).Scan(&ok)
	})
	if err != nil {
		return false, fmt.Errorf("check job: %w", err)
	}
	return ok, nil
}

// Release moves a sending job back to queued and subtracts 1 from attempts.
func (s *Store) Release(ctx context.C
… [7256 more chars truncated]
```

</details>

**🔧 Bash** ```bash
grep -rn "Clock" backend/internal/delivery/*.go | grep -v _test | grep -n "type\|func" ; sed -n 338,352p backend/internal/delivery/store.go; grep -n "ExpireSending\|defaultSendingAfter\|SendingAfter" backend/internal/delivery/worker.go; sed -n 140,150p backend/internal/delivery/worker.go; grep -n "D-053\|D-066\|waits for a token\|token" docs/architecture.md docs/specs/07-send-queue/*.md | head -20; which go; sed -n 1,60p frontend/src/components/campaign/__tests__/review-dialog.test.ts; grep -n "it(\|describe(" frontend/src/components/campaign/__tests__/review-dialog.test.ts
```

<details><summary>Result</summary>

```
2:backend/internal/delivery/limiter.go:22:func NewLimiter(perSec int, clock Clock) *Limiter {
25:backend/internal/delivery/provider.go:21:type Clock interface {
26:backend/internal/delivery/provider.go:27:type realClock struct{}
28:backend/internal/delivery/provider.go:30:func RealClock() Clock { return realClock{} }
29:backend/internal/delivery/provider.go:32:func (realClock) Now() time.Time { return time.Now() }
30:backend/internal/delivery/provider.go:34:func (realClock) Sleep(ctx context.Context, d time.Duration) error {
	return out, nil
}

// ExpireSending sets unknown on the jobs that have been sending for more than
// age. The outcome of such a job is not known, so it is never sent again.
func (s *Store) ExpireSending(ctx context.Context, now time.Time, age time.Duration) (int64, error) {
	var n int64
	err := s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE send_jobs SET status = 'unknown', retryable = false,
last_error = 'worker stopped before the result was known', updated_at = $1
WHERE status = 'sending' AND updated_at < $2`, now, now.Add(-age))
		n = tag.RowsAffected()
		return err
	})
	if err != nil {
250:	if n, err := w.Store.ExpireSending(ctx, now, w.Cfg.Timeout+30*time.Second); err != nil {
		if err := w.Store.MarkEnqueued(ctx, ids[ch], w.Clock.Now()); err != nil {
			w.log().Error("mark enqueued", "channel", ch, "count", len(ids[ch]), "error", err)
		}
	}
}

// withoutBacklog drops the channels whose group still has unread entries.
// The jobs of such a channel are waiting on the stream, and a second entry for
// each would only use rate tokens. When the lag is unknown or cannot be read,
// the jobs are pushed again, which is the safe direction.
func (w *Worker) withoutBacklog(ctx context.Context, ids map[Channel][]int64) map[Channel][]int64 {
docs/architecture.md:77:   - 32 consumers for each channel read the stream (D-049). Each consumer waits for a rate-limit token, then claims the send job with one conditional
… [4845 more chars truncated]
```

</details>

**🔧 Bash** ```bash
cat >> backend/internal/delivery/consumer_test.go <<'EOF'

// A stall between the claim and the request (a slow database) must not let
// the held sends leave together. The provider sees the real request times.
func TestConsumer_PacesRequestsAfterDatabaseStall(t *testing.T) {
	e := newConsumerEnv(t)
	const perSec = 50
	const n = 7
	interval := time.Second / perSec

	var mu sync.Mutex
	var at []time.Time
	mock := newMock(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		at = append(at, time.Now())
		mu.Unlock()
		mock.Config.Handler.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)

	clk := RealClock()
	c := e.consumer(NewProvider(srv.URL, 2*time.Second))
	c.Clock = clk
	c.Limiter = NewLimiter(perSec, clk)

	// Every consumer stalls after its claim until all n jobs are claimed.
	var claimed sync.WaitGroup
	claimed.Add(n)
	stall := make(chan struct{})
	go func() {
		claimed.Wait()
		close(stall)
	}()
	c.afterClaim = func() {
		claimed.Done()
		<-stall
	}

	if err := e.queue.Push(e.ctx, ChannelSMS, e.sms[:n]); err != nil {
		t.Fatal(err)
	}
	es, err := e.queue.Read(e.ctx, ChannelSMS, n, time.Second)
	if err != nil || len(es) != n {
		t.Fatalf("Read = %v, %v", es, err)
	}
	var wg sync.WaitGroup
	for _, en := range es {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.Handle(e.ctx, en); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()

	if len(at) != n {
		t.Fatalf("provider got %d requests, want %d", len(at), n)
	}
	slices.SortFunc(at, func(a, b time.Time) int { return a.Compare(b) })
	// n requests need n-1 intervals. The margin allows for scheduler jitter.
	if span, want := at[n-1].Sub(at[0]), (n-1)*interval*8/10; span < want {
		t.Fatalf("%d requests arrived within %v, want at least %v at %d per second", n, span, want, perSec)
	}
	for i := 1; i < n; i++ {
		if gap := at[i].Sub(at[i-1]); gap < interval/4 {
			t.Fatalf("requests %d and %d are %v apart, want about %v", i-1, i, gap, interval)
		}
	}
}
EOF
sed -i '' 's|^\t"net/http"$|\t"net/http"\n\t"net/http/httptest"|; s|^\t"regexp"$|\t"regexp"\n\t"slices"|' backend/internal/delivery/consumer_test.go
grep -n "test:\|^test\|TEST_DATABASE\|go test" Makefile | head; sed -n 60,90p frontend/src/components/campaign/__tests__/review-dialog.test.ts; sed -n 178,192p frontend/src/components/campaign/__tests__/review-dialog.test.ts
```

<details><summary>Result</summary>

```
8:GO_TEST = docker run --rm --network cc-test -e TEST_DATABASE_URL -e TEST_REDIS_URL -v "$(CURDIR)/backend:/src" -w /src golang:1.27-alpine go
9:TEST_DATABASE_URL = postgres://test:test@cc-test-pg:5432/test?sslmode=disable
15:TEST_DATABASE_URL = postgres://test:test@127.0.0.1:55432/test?sslmode=disable
49:test:
63:	@export TEST_DATABASE_URL='$(TEST_DATABASE_URL)' TEST_REDIS_URL='$(TEST_REDIS_URL)'; $(GO_TEST) test ./...; rc=$$?; \
65:	$(PNPM) test:run
  const el = $(sel)!
  el.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, button: 0 }))
  el.dispatchEvent(new MouseEvent('click', { bubbles: true, button: 0 }))
  await flushPromises()
}

async function type(sel: string, value: string) {
  const el = $(sel) as HTMLInputElement | HTMLTextAreaElement
  el.value = value
  el.dispatchEvent(new Event('input', { bubbles: true }))
  await flushPromises()
}

async function click(sel: string) {
  $(sel)!.click()
  await flushPromises()
}

beforeEach(() => {
  getC.mockReset()
  getS.mockReset()
  save.mockReset()
  approve.mockReset()
  vi.mocked(toast.success).mockReset()
  getS.mockResolvedValue({ tr: ['Abdurrahmangazi', 'Ayşe', 'Mehmet'], en: ['Christopher', 'Ann'] })
})

afterEach(() => {
  document.body.innerHTML = ''
})

  it('save draft sends one PUT and toasts', async () => {
    save.mockResolvedValue(campaign())
    await make()
    await type('[data-test="sms-input"]', 'Hey {{first_name}}')
    await click('[data-test="save"]')
    expect(save).toHaveBeenCalledTimes(1)
    const [id, contents] = save.mock.calls[0]!
    expect(id).toBe(12)
    expect(contents).toEqual([
      { language: 'en', sms: 'Hey {{first_name}}', email_subject: 'Sale {{first_name}}', email_body: 'Hello {{first_name}}' },
      { language: 'tr', sms: 'Merhaba {{first_name}}', email_subject: 'İndirim', email_body: 'Merhaba {{first_name}},\nSepet' },
    ])
    expect(toast.success).toHaveBeenCalledWith('Draft saved.')
  })
```

</details>

**🔧 Bash** ```bash
python3 - <<'EOF'
p='frontend/src/components/campaign/__tests__/review-dialog.test.ts'
s=open(p).read()
anchor="  it('cancel asks only when dirty', async () => {"
new='''  it('keeps the text typed while a save is in progress', async () => {
    let finish!: (c: Campaign) => void
    save.mockReturnValue(new Promise<Campaign>((resolve) => {
      finish = resolve
    }))
    await make()
    await type('[data-test="sms-input"]', 'Hey {{first_name}}')
    await click('[data-test="save"]')
    await type('[data-test="sms-input"]', 'Hey {{first_name}}, one more thing')
    const c = campaign()
    c.contents[0]!.sms = 'Hey {{first_name}}'
    finish(c)
    await flushPromises()
    expect(($('[data-test="sms-input"]') as HTMLTextAreaElement).value).toBe('Hey {{first_name}}, one more thing')
    // The newer text is not saved yet, so a close must ask.
    await click('[data-test="cancel"]')
    expect($('[data-test="discard"]')).not.toBeNull()
    await click('[data-test="discard-confirm"]')
    // Discard goes back to the saved text, not to the text before the save.
    await w_reopen()
  })

'''
EOF
sed -n 40,66p Makefile; sed -n 193,214p frontend/src/components/campaign/__tests__/review-dialog.test.ts
```

<details><summary>Result</summary>

```
curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 localhost:8082/healthz
	@echo
	curl -fsS --retry 15 --retry-delay 2 --retry-all-errors --retry-max-time 30 -o /dev/null -w "frontend %{http_code}\n" localhost:3000/

down:
	docker compose down

# Starts a temporary Postgres and Redis, runs the Go tests and removes the containers.
# The exit code is the exit code of the Go tests.
test:
	@docker network inspect cc-test >/dev/null 2>&1 || docker network create cc-test >/dev/null
	@docker rm -f cc-test-pg cc-test-redis >/dev/null 2>&1 || true
	docker run -d --name cc-test-pg --network cc-test \
		-e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=test \
		-p 127.0.0.1:55432:5432 postgres:17-alpine >/dev/null
	docker run -d --name cc-test-redis --network cc-test \
		-p 127.0.0.1:56379:6379 redis:8-alpine >/dev/null
	@i=0; until docker exec cc-test-pg pg_isready -U test -d test -h 127.0.0.1 >/dev/null 2>&1; do \
		i=$$((i+1)); if [ $$i -gt 60 ]; then echo "test Postgres not ready"; docker rm -f cc-test-pg cc-test-redis >/dev/null; exit 1; fi; \
		sleep 1; done
	@i=0; until docker exec cc-test-redis redis-cli ping >/dev/null 2>&1; do \
		i=$$((i+1)); if [ $$i -gt 60 ]; then echo "test Redis not ready"; docker rm -f cc-test-pg cc-test-redis >/dev/null; exit 1; fi; \
		sleep 1; done
	@export TEST_DATABASE_URL='$(TEST_DATABASE_URL)' TEST_REDIS_URL='$(TEST_REDIS_URL)'; $(GO_TEST) test ./...; rc=$$?; \
		docker rm -f cc-test-pg cc-test-redis >/dev/null 2>&1; exit $$rc
	$(PNPM) test:run

  it('cancel asks only when dirty', async () => {
    const w = await make()
    await click('[data-test="cancel"]')
    expect(w.emitted('update:open')?.at(-1)).toEqual([false])
    const w2 = await make()
    await type('[data-test="sms-input"]', 'changed')
    await click('[data-test="cancel"]')
    expect(w2.emitted('update:open')).toBeUndefined()
    expect(text('[data-test="discard"]')).toContain('Discard your changes?')
    await click('[data-test="dis
… [486 more chars truncated]
```

</details>

**🔧 Bash** ```bash
python3 - <<'EOF'
p='frontend/src/components/campaign/__tests__/review-dialog.test.ts'
s=open(p).read()
anchor="  it('cancel asks only when dirty', async () => {"
new='''  it('keeps the text typed while a save is in progress', async () => {
    let finish!: (c: Campaign) => void
    save.mockReturnValue(new Promise<Campaign>((resolve) => {
      finish = resolve
    }))
    await make()
    await type('[data-test="sms-input"]', 'Hey {{first_name}}')
    await type('[data-test="subject-input"]', 'unused')
    await click('[data-test="save"]')
    await type('[data-test="sms-input"]', 'Hey {{first_name}}, one more thing')
    const c = campaign()
    c.contents[0]!.sms = 'Hey {{first_name}} (saved)'
    c.contents[0]!.email_body = 'Body from the server'
    finish(c)
    await flushPromises()
    const sms = () => ($('[data-test="sms-input"]') as HTMLTextAreaElement).value
    expect(sms()).toBe('Hey {{first_name}}, one more thing')
    expect(toast.success).toHaveBeenCalledWith('Draft saved.')
    // A field with no newer edit takes the saved value.
    await pressTab('[data-test="channel-tab"][data-value="email"]')
    expect(($('[data-test="body-input"]') as HTMLTextAreaElement).value).toBe('Body from the server')
    await pressTab('[data-test="channel-tab"][data-value="sms"]')
    // The newer text is not saved, so a close asks. Discard goes back to the saved text.
    await click('[data-test="cancel"]')
    expect($('[data-test="discard"]')).not.toBeNull()
    await click('[data-test="discard-confirm"]')
    expect(sms()).toBe('Hey {{first_name}} (saved)')
  })

'''
assert anchor in s
s=s.replace(anchor,new+anchor,1)
open(p,'w').write(s)
EOF
grep -n "subject-input" frontend/src/components/campaign/__tests__/review-dialog.test.ts | head -3
docker ps --format '{{.Names}}' | head -20
```

<details><summary>Result</summary>

```
100:    expect($('[data-test="subject-input"]')).not.toBeNull()
139:    await type('[data-test="subject-input"]', 'a\nb')
140:    expect(($('[data-test="subject-input"]') as HTMLInputElement).value).not.toContain('\n')
insider_one-frontend-1
insider_one-worker-1
insider_one-api-1
insider_one-mock-provider-1
insider_one-redis-1
insider_one-postgres-1
```

</details>

**🔧 Bash** ```bash
python3 - <<'EOF'
p='frontend/src/components/campaign/__tests__/review-dialog.test.ts'
s=open(p).read()
s=s.replace("    await type('[data-test=\"subject-input\"]', 'unused')\n","",1)
old=s[s.index("    // The newer text is not saved, so a close asks."):s.index("  it('cancel asks only when dirty'")]
s=s.replace(old,'''    // The newer text is not saved, so a close asks.
    await click('[data-test="cancel"]')
    expect($('[data-test="discard"]')).not.toBeNull()
  })

''',1)
open(p,'w').write(s)
EOF
(docker network inspect cc-test >/dev/null 2>&1 || docker network create cc-test >/dev/null
docker rm -f cc-test-pg cc-test-redis >/dev/null 2>&1
docker run -d --name cc-test-pg --network cc-test -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=test -p 127.0.0.1:55432:5432 postgres:17-alpine >/dev/null
docker run -d --name cc-test-redis --network cc-test -p 127.0.0.1:56379:6379 redis:8-alpine >/dev/null
i=0; until docker exec cc-test-pg pg_isready -U test -d test -h 127.0.0.1 >/dev/null 2>&1; do i=$((i+1)); [ $i -gt 60 ] && break; sleep 1; done)
grep -n "TEST_REDIS_URL =" Makefile
export TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55432/test?sslmode=disable' TEST_REDIS_URL='redis://127.0.0.1:56379/0'
cd backend && go test -count=1 -run 'TestConsumer_PacesRequests' ./internal/delivery/ 2>&1 | tail -15
cd ../frontend && pnpm vitest run src/components/campaign/__tests__/review-dialog.test.ts 2>&1 | tail -25
```

<details><summary>Result</summary>

```
10:TEST_REDIS_URL = redis://cc-test-redis:6379/0
16:TEST_REDIS_URL = redis://127.0.0.1:56379/0
--- FAIL: TestConsumer_PacesRequestsAfterDatabaseStall (0.29s)
    consumer_test.go:460: 7 requests arrived within 81µs, want at least 96ms at 50 per second
FAIL
FAIL	campaigncomposer/internal/delivery	0.745s
FAIL

⎯⎯⎯⎯⎯⎯⎯ Failed Tests 1 ⎯⎯⎯⎯⎯⎯⎯

 FAIL  src/components/campaign/__tests__/review-dialog.test.ts > review dialog > keeps the text typed while a save is in progress
AssertionError: expected 'Hey {{first_name}} (saved)' to be 'Hey {{first_name}}, one more thing' // Object.is equality

Expected: "Hey {{first_name}}, one more thing"
Received: "Hey {{first_name}} (saved)"

 ❯ src/components/campaign/__tests__/review-dialog.test.ts:208:19
    206|     await flushPromises()
    207|     const sms = () => ($('[data-test="sms-input"]') as HTMLTextAreaEle…
    208|     expect(sms()).toBe('Hey {{first_name}}, one more thing')
       |                   ^
    209|     expect(toast.success).toHaveBeenCalledWith('Draft saved.')
    210|     // A field with no newer edit takes the saved value.

⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯[1/1]⎯


 Test Files  1 failed (1)
      Tests  1 failed | 18 passed (19)
   Start at  12:41:26
   Duration  1.80s (tests 68%, import 19%, environment 9%, transform 4%)
```

</details>

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one && python3 - <<'EOF'
p='backend/internal/delivery/consumer.go'
s=open(p).read()
a=s.index("// Handle runs one entry")
b=s.index("\tmsg, skip, err := c.Store.Load(ctx, job)")
s=s[:a]+'''// Handle runs one entry: claim, load, wait, send, write the result, acknowledge.
// It returns an error only when the entry must stay unacknowledged.
func (c *Consumer) Handle(ctx context.Context, e Entry) error {
	// The writes after a claim use a context that cannot be cancelled: the
	// result of an attempt must reach Postgres, also during a shutdown.
	done := context.WithoutCancel(ctx)

	// The claim comes before the rate token, so an entry for a job that
	// cannot be claimed uses no token.
	job, ok, err := c.Store.Claim(ctx, e.JobID, c.Clock.Now())
	if err != nil {
		return err
	}
	if !ok {
		return c.Queue.Ack(done, c.Channel, e.StreamID)
	}
	if c.afterClaim != nil {
		c.afterClaim()
	}
	// Nothing was sent yet, so the job can go back to the queue.
	release := func(cause error) error {
		if rerr := c.Store.Release(done, job.ID); rerr != nil {
			c.log().Error("release job", "job_id", job.ID, "channel", c.Channel, "error", rerr)
		}
		return cause
	}
	if err := ctx.Err(); err != nil {
		return release(err)
	}

'''+s[b:]
old=s[s.index("\tif err != nil {\n\t\t// Nothing was sent, so the job can go back to the queue."):s.index("\tif skip != \"\" {")]
s=s.replace(old,"\tif err != nil {\n\t\treturn release(err)\n\t}\n",1)
old2='''	reference := fmt.Sprintf("%d:%d:%s", job.CampaignID, job.UserID, job.Channel)
'''
new2='''	// The token wait is the last step before the request. No database work
	// sits between the two, so a slow query cannot hold permitted sends and
	// then let them leave together.
	waitFrom := c.Clock.Now()
	err = c.Limiter.Wait(ctx)
	c.Metrics.Waited(c.Channel, c.Clock.Now().Sub(waitFrom))
	if err != nil {
		return release(err)
	}

	reference := fmt.Sprintf("%d:%d:%s", job.CampaignID, job.UserID, job.Channel)
'''
assert old2 in s
s=s.replace(old2,new2,1)
open(p,'w').write(s)

p='backend/internal/delivery/store.go'
s=open(p).read()
a=s.index("// Claimable reports whether"); b=s.index("// Release moves a sending job")
s=s[:a]+s[b:]
open(p,'w').write(s)

p='backend/internal/delivery/worker.go'
s=open(p).read()
old="// each would only use rate tokens. When"
assert old in s
s=s.replace(old,"// each would only add claims that fail. When",1)
open(p,'w').write(s)

p='frontend/src/components/campaign/review-dialog.vue'
s=open(p).read()
old='''  try {
    reset(await saveContents(props.campaignId, sent))
    toast.success('Draft saved.')
'''
new='''  try {
    const c = await saveContents(props.campaignId, sent)
    // The fields stay editable during the request. Text typed after the
    // submit is newer than the response, so it stays and counts as unsaved.
    const newer: Record<string, Partial<Content>> = {}
    for (const s of sent) {
      for (const f of FIELDS) {
        const now = edit
… [466 more chars truncated]
```

<details><summary>Result</summary>

```
// Handle runs one entry: claim, load, wait, send, write the result, acknowledge.
// It returns an error only when the entry must stay unacknowledged.
func (c *Consumer) Handle(ctx context.Context, e Entry) error {
	// The writes after a claim use a context that cannot be cancelled: the
	// result of an attempt must reach Postgres, also during a shutdown.
	done := context.WithoutCancel(ctx)

	// The claim comes before the rate token, so an entry for a job that
	// cannot be claimed uses no token.
	job, ok, err := c.Store.Claim(ctx, e.JobID, c.Clock.Now())
	if err != nil {
		return err
	}
	if !ok {
		return c.Queue.Ack(done, c.Channel, e.StreamID)
	}
	if c.afterClaim != nil {
		c.afterClaim()
	}
	// Nothing was sent yet, so the job can go back to the queue.
	release := func(cause error) error {
		if rerr := c.Store.Release(done, job.ID); rerr != nil {
			c.log().Error("release job", "job_id", job.ID, "channel", c.Channel, "error", rerr)
		}
		return cause
	}
	if err := ctx.Err(); err != nil {
		return release(err)
	}

	msg, skip, err := c.Store.Load(ctx, job)
	if err != nil {
		return release(err)
	}
	if skip != "" {
		c.log().Info("send job skipped", "job_id", job.ID, "channel", c.Channel, "attempt", job.Attempts, "error", skip)
		return c.finish(done, e, job, Result{Status: StatusSkipped, LastError: skip})
	}

	// The token wait is the last step before the request. No database work
	// sits between the two, so a slow query cannot hold permitted sends and
	// then let them leave together.
		if err := w.Store.MarkEnqueued(ctx, ids[ch], w.Clock.Now()); err != nil {
			w.log().Error("mark enqueued", "channel", ch, "count", len(ids[ch]), "error", err)
		}
	}
}

// withoutBacklog drops the channels whose group still has unread entries.
// The jobs of such a channel are waiting on the stream, and a second entry for
// each would only add claims that fail. When the lag is unknown or cannot be read,
// the jobs are pushed again, which is the safe direction.
func (w *Worker)
… [83 more chars truncated]
```

</details>

**🔧 Bash** ```bash
export TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55432/test?sslmode=disable' TEST_REDIS_URL='redis://127.0.0.1:56379/0'
cd backend && gofmt -l . ; go vet ./... && go test -race -count=1 ./internal/delivery/... ./cmd/... 2>&1 | tail -30; go test -count=10 -run 'TestConsumer_PacesRequests' ./internal/delivery/ 2>&1 | tail -5
cd ../frontend && pnpm vitest run src/components/campaign/__tests__/review-dialog.test.ts 2>&1 | tail -8
```

<details><summary>Result</summary>

```
ok  	campaigncomposer/internal/delivery	45.988s
?   	campaigncomposer/internal/delivery/deliverytest	[no test files]
ok  	campaigncomposer/cmd/api	1.298s
?   	campaigncomposer/cmd/mockprovider	[no test files]
?   	campaigncomposer/cmd/seed	[no test files]
ok  	campaigncomposer/cmd/worker	1.713s
ok  	campaigncomposer/internal/delivery	2.747s
 RUN  v5.0.5 ~/Desktop/projects/case-studies/insider_one/frontend


 Test Files  1 passed (1)
      Tests  19 passed (19)
   Start at  12:42:42
   Duration  1.94s (tests 69%, import 18%, environment 8%, transform 6%)
```

</details>

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one && sed -n 1,25p backend/internal/content/mock.go; sed -n 150,160p docs/specs/07-send-queue/goal.md; sed -n 366,369p docs/specs/07-send-queue/goal.md; sed -n 74,80p docs/architecture.md; sed -n 456,468p docs/decisions-log.md; grep -n "description\|Description" backend/internal/llm/mock.go | head -5; tail -5 docs/decisions-log.md | cut -c1-80
```

<details><summary>Result</summary>

```
package content

import "encoding/json"

var mockContents = map[string]Content{
	"tr": {
		SMS:          "Merhaba {{first_name}}, sepetin seni bekliyor! Bugün tamamla, %30 indirimi kaçırma.",
		EmailSubject: "{{first_name}}, sepetinde %30 indirim seni bekliyor",
		EmailBody: "Merhaba {{first_name}},\n\n" +
			"Sepetine eklediğin ürünler hâlâ seni bekliyor.\n\n" +
			"Siparişini bu hafta tamamlarsan %30 indirimden yararlanabilirsin.\n\n" +
			"Görüşmek üzere!",
	},
	"en": {
		SMS:          "Hi {{first_name}}, your cart is waiting! Finish today and get 30% off.",
		EmailSubject: "{{first_name}}, 30% off is waiting in your cart",
		EmailBody: "Hi {{first_name}},\n\n" +
			"The items you added to your cart are still waiting for you.\n\n" +
			"Complete your order this week and enjoy 30% off.\n\n" +
			"See you soon!",
	},
}

// MockContent returns the default tool input JSON for "tr" or "en". Other
// codes get "en".
The sender runs in the worker. Each channel has its own consumers and its own rate limiter. A slow channel does not stop the other channel.

For each entry:

1. **Wait** for a token of the channel rate limiter.
2. **Claim.** `UPDATE send_jobs SET status = 'sending', attempts = attempts + 1 WHERE id = $1 AND status IN ('queued', 'retrying') AND (next_attempt_at IS NULL OR next_attempt_at <= now())`. If no row changes, acknowledge the entry and stop. Another consumer has it, or it is done.
3. **Load.** Read the user and the content of the send job language. Read the user with the consent condition of the channel.
4. **Check.**
   - If the user has no consent now, set `skipped` with `no_consent`.
   - If the user has no phone or no email for the channel, set `skipped` with `no_contact`.
   - Render the template with the user's first name. Check the SMS against 160 runes, or the subject against 80 runes. If the text is over the limit, set `skipped` with `over_limit`. This must not happen, because the approval checked the longest first name.
The user did not dec
… [4068 more chars truncated]
```

</details>

**🔧 Bash** ```bash
sed -n 160,166p docs/specs/07-send-queue/goal.md | cut -c1-160
```

<details><summary>Result</summary>

```
- Render the template with the user's first name. Check the SMS against 160 runes, or the subject against 80 runes. If the text is over the limit, set `skipp
5. **Send** one HTTP request to the provider, with a timeout. Set `reference` to `<campaign id>:<user id>:<channel>`. Set no `Idempotency-Key` header.
6. **Write the result** to Postgres. See "Retry strategy".
7. **Acknowledge** the entry.

The email `body` is the rendered email body. The provider gets plain text. The email has no sender field, because the mock provider has none.
```

</details>

**🔧 Bash** ```bash
python3 - <<'EOF'
def sub(p, pairs):
    s=open(p).read()
    for a,b in pairs:
        assert a in s, a[:50]
        s=s.replace(a,b,1)
    open(p,'w').write(s)

sub('docs/specs/07-send-queue/goal.md',[
("1. **Wait** for a token of the channel rate limiter.\n2. **Claim.**","1. **Claim.**"),
("\n3. **Load.**","\n2. **Load.**"),
("\n4. **Check.**","\n3. **Check.**"),
("\n5. **Send** one HTTP request","\n4. **Wait** for a token of the channel rate limiter. No database work comes between the token and the request (D-084).\n5. **Send** one HTTP request"),
("- The worker waits for a rate-limit token before it claims a send job. On `SIGTERM`, a claimed send job with no request goes back to `queued`.",
 "- The worker claims a send job, then waits for a rate-limit token immediately before the request (D-084 replaced the first order). On `SIGTERM`, a claimed send job with no request goes back to `queued`."),
])
sub('docs/architecture.md',[
("Each consumer waits for a rate-limit token, then claims the send job with one conditional `UPDATE`. The claim is the dedup gate (D-053, D-046).",
 "Each consumer claims the send job with one conditional `UPDATE`. The claim is the dedup gate (D-046). The consumer waits for a rate-limit token immediately before the provider request (D-084)."),
])
sub('docs/decisions-log.md',[
("## D-053 The worker waits for a token before it claims\n","## D-053 The worker waits for a token before it claims\n\n**Replaced by D-084.**\n"),
("## D-066 Duplicate queue entries use no rate token\n","## D-066 Duplicate queue entries use no rate token\n\n**D-084 removed the `Store.Claimable` check. The lag gate of the recovery stays.**\n"),
])
with open('docs/decisions-log.md','a') as f:
    f.write('''
## D-084 The worker waits for a token immediately before the request

- **Context:** The consumer took a rate token, then claimed and loaded the send job from Postgres (D-053). A slow database held sends that already had a token. When the database answered, these sends left together with the newly permitted sends. The channel could then go above 100 requests per second. A test with 7 held sends showed 7 requests in less than 1 ms.
- **Options:** Keep the order and add a second limiter before the request. Move the token wait to the last step before the request.
- **Choice:** The order is claim, load, wait, send. No database work comes between the token and the request. A cancel or an error before the request releases the send job to `queued`. `Store.Claimable` is removed, because the claim now rejects a duplicate entry before the token wait (D-066).
- **Why:** One limiter at the request is the only place that bounds the real request rate. A second limiter adds state and gives no more safety. The cost is that a send job shows `sending` during its token wait. With 32 consumers at 100 per second, this wait is 320 ms or less. After a crash, the recovery sets these send jobs to `unknown`. This is at most 32 for each channel. A `Retry-After` longer than
… [2409 more chars truncated]
```

<details><summary>Result</summary>

```
README.md:23: sentence-length: 29 words: The segment rule of the example brief: users in Istanbul who added to the cart in the last 30 days and did not buy in the last 30 days.
371:**Replaced by D-084.**
464:**D-084 removed the `Store.Claimable` check. The lag gate of the recovery stays.**
590:## D-084 The worker waits for a token immediately before the request
backend/internal/content/prompt.go:51:// name length and the brief. The first line is read back by LanguageOf.
backend/internal/content/prompt.go-52-func UserMessage(lang, description string, firstNameRunes int, brief string) string {
backend/internal/content/prompt.go-53-	name := languageNames[lang]
backend/internal/content/prompt.go-54-	if name == "" {
backend/internal/content/prompt.go-55-		name = lang
backend/internal/content/prompt.go-56-	}
backend/internal/content/prompt.go-57-	var b strings.Builder
backend/internal/content/prompt.go-58-	fmt.Fprintf(&b, "%s%s (%s)\n", languagePrefix, lang, name)
backend/internal/content/prompt.go-59-	fmt.Fprintf(&b, "Audience: %s\n", description)
backend/internal/content/prompt.go-60-	fmt.Fprintf(&b, "Longest first name: %d characters\n\n", firstNameRunes)
backend/internal/content/prompt.go-61-	b.WriteString("<brief>\n" + brief + "\n</brief>")
backend/internal/content/prompt.go-62-	return b.String()
backend/internal/content/prompt.go-63-}
--
backend/internal/content/prompt.go:65:// LanguageOf reads the language code from the first line of a UserMessage.
backend/internal/content/prompt.go:66:func LanguageOf(userMessage string) (string, bool) {
backend/internal/content/prompt.go-67-	line, _, _ := strings.Cut(userMessage, "\n")
backend/internal/content/prompt.go-68-	rest, ok := strings.CutPrefix(line, languagePrefix)
backend/internal/content/prompt.go-69-	if !ok {
backend/internal/content/prompt.go-70-		return "", false
backend/internal/content/prompt.go-71-	}
backend/internal/content/prompt.go-72-	code, _, ok := strings.Cut(rest, " (")
backend/internal/content/prompt.go-7
… [295 more chars truncated]
```

</details>

**🔧 Bash** ```bash
python3 - <<'EOF'
p='README.md'
s=open(p).read()
a="- The segment rule of the example brief: users in Istanbul who added to the cart in the last 30 days and did not buy in the last 30 days.\n"
assert a in s
s=s.replace(a,"- The segment rule of the example brief. It selects users in Istanbul who added to the cart in the last 30 days. It excludes users who bought in the last 30 days.\n")
open(p,'w').write(s)
EOF
make verify > "/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/22a0e40d-eeb7-4ae5-9612-495cb6c67262/scratchpad/verify.log" 2>&1; echo "exit=$?"; grep -E "^(ok|FAIL|---|\?)|Tests |Test Files|error" "/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/22a0e40d-eeb7-4ae5-9612-495cb6c67262/scratchpad/verify.log" | grep -v "no test files" | tail -30
```

<details><summary>Result</summary>

```
exit=0
ok  	campaigncomposer/cmd/api	(cached)
ok  	campaigncomposer/cmd/worker	(cached)
ok  	campaigncomposer/internal/campaign	(cached)
ok  	campaigncomposer/internal/config	(cached)
ok  	campaigncomposer/internal/content	(cached)
ok  	campaigncomposer/internal/dashboard	(cached)
ok  	campaigncomposer/internal/delivery	41.608s
ok  	campaigncomposer/internal/httpapi	(cached)
ok  	campaigncomposer/internal/llm	(cached)
ok  	campaigncomposer/internal/mockprovider	(cached)
ok  	campaigncomposer/internal/seed	(cached)
ok  	campaigncomposer/internal/segment	(cached)
ok  	campaigncomposer/internal/store	(cached)
 Test Files  29 passed (29)
      Tests  222 passed (222)
dist/assets/errors-Dc3ugJO5.js                       0.19 kB │ gzip:  0.17 kB
dist/assets/errors-DQFC1xB4.js                       0.31 kB │ gzip:  0.24 kB
dist/assets/custom-error-lpwJBPQa.js                 0.83 kB │ gzip:  0.50 kB
```

</details>

**🔧 Bash** ```bash
cat >> tasks/progress.md <<'EOF'

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
EOF
python3 .claude/skills/ste-writing/scripts/ste_check.py tasks/progress.md | tail -3; git status --short | grep -v ai-sessions
```

<details><summary>Result</summary>

```
tasks/progress.md:82: sentence-length: 36 words: User decisions: one review window with SMS and Email tabs, a second confirm before the send, a token-based design that works on a phone, failed languages are skipped, and Claude saves a draft when it finishes.
tasks/progress.md:82: sentence-length: 46 words: All 5 are fixed, each with a test that failed first: a stale draft text after "Regenerate content", the row lock that waited for the audience query, a late response for another campaign, old limits after a 422, and the review button after a failed regenerate.
tasks/progress.md:83: sentence-length: 30 words: A compose run with CODE on ports 13000 and 18080 (the running stack holds 3000 and 8080) gave the draft, approve CODE, repeat CODE, PUT CODE and CORS with CODE.
M  .claude/agent-memory/reviewer/MEMORY.md
A  .claude/agent-memory/reviewer/dashboard-metrics-defects.md
M  .claude/agent-memory/reviewer/review-approve-defects.md
 M README.md
 M backend/internal/delivery/consumer.go
 M backend/internal/delivery/consumer_test.go
 M backend/internal/delivery/store.go
 M backend/internal/delivery/worker.go
 M docs/architecture.md
 M docs/decisions-log.md
 M docs/specs/07-send-queue/goal.md
 M frontend/src/components/campaign/__tests__/review-dialog.test.ts
 M frontend/src/components/campaign/review-dialog.vue
 M tasks/backlog.json
 M tasks/progress.md
```

</details>

## 🤖 Claude  <sub>2026-11-11 09:44:32</sub>

All three review findings were correct, and all three are fixed. `make verify` passes; nothing is committed.

## 1. Rate limiting (P2)

- **Cause:** the consumer took a rate token, then did the claim and the load in Postgres. A database stall held permitted sends and released them together.
- **Reproduction:** a new test holds 7 claimed sends, then releases them at the same time. Before the fix, the provider got 7 requests in 81 µs at a limit of 50 per second.
- **Fix:** the order in [consumer.go](backend/internal/delivery/consumer.go) is now claim, load, wait, send. No database work comes between the token and the request.
- **Removed:** `Store.Claimable`. The claim now rejects a duplicate entry before the token wait, so the pre-check had no function.
- **Test:** `TestConsumer_PacesRequestsAfterDatabaseStall` in [consumer_test.go](backend/internal/delivery/consumer_test.go) checks the real request times at the provider. It passed 10 of 10 runs after the fix.

This change replaces D-053, which was my earlier decision. A send job now shows `sending` during its token wait, which is 320 ms or less with 32 consumers. After a crash, the recovery sets those send jobs to `unknown`, at most 32 for each channel. D-084 records this.

## 2. Save discards newer edits (P2)

- **Cause:** `saveDraft` in [review-dialog.vue](frontend/src/components/campaign/review-dialog.vue) reset all fields from the response while the fields stayed editable.
- **Fix:** a field that changed after the submit keeps the typed text. All other fields take the saved value. The kept text counts as unsaved, so a close asks before it discards.
- **Why not disable the fields:** a disabled field drops the focus on each save. D-085 records this.
- **Test:** "keeps the text typed while a save is in progress" in [review-dialog.test.ts](frontend/src/components/campaign/__tests__/review-dialog.test.ts). It failed before the fix.

## 3. Mock disclosure (P3)

[README.md](README.md) has a new section, "Mock mode and Claude". It says that the mock does not read the brief and returns the Istanbul rule and the 30% discount text for each brief. It also gives the four steps to enable Claude with `LLM_MODE=anthropic` and `ANTHROPIC_API_KEY`.

## Other changes

- [docs/decisions-log.md](docs/decisions-log.md): D-084 and D-085, and notes on D-053 and D-066.
- [docs/architecture.md](docs/architecture.md) and [docs/specs/07-send-queue/goal.md](docs/specs/07-send-queue/goal.md): the new step order.
- [tasks/progress.md](tasks/progress.md): a session entry.

## Open points

- A `Retry-After` longer than `SEND_TIMEOUT` plus 30 s can set a waiting send job to `unknown`. The mock provider sends 1 s, so this does not occur now.
- I did not run the compose stack to measure the request rate against the mock provider after the change. The running stack still has the old worker until you rebuild it.

## 👤 User  <sub>2026-11-11 10:39:07</sub>

ai review found these as well 


The three previous findings are fixed in `280a142`. I found **two regressions in the delivery change**, both reproduced with isolated tests in a temporary copy.

1. **[P1] Consent revoked during the rate-limit wait is ignored.**  
   [consumer.go:64](~/Desktop/projects/case-studies/insider_one/backend/internal/delivery/consumer.go:64) loads the recipient and checks consent before waiting for permission to send. If consent is revoked during that wait, the cached message still goes to the provider. My reproduction revoked SMS consent during the wait; the provider received **one request**. Revalidate consent near dispatch while preserving request pacing.

2. **[P2] Recovery can expire a live job that is waiting to send.**  
   Jobs enter `sending` before the limiter wait, but [worker.go:250](~/Desktop/projects/case-studies/insider_one/backend/internal/delivery/worker.go:250) expires them after `SEND_TIMEOUT + 30s`, including time spent waiting. A long 429 pause or a supported low-rate/high-consumer configuration can exceed that threshold. The consumer then sends anyway, while `Finish` cannot update the expired row. My reproduction produced **one provider request with the job permanently marked `unknown`**. Distinguish waiting jobs from active requests, or maintain and verify ownership before dispatch.

The previous fixes check out:

- Pacing now happens after database loading, with a regression test.
- Save responses preserve edits made during the request.
- README explains canned mock responses and enabling Claude.

Fresh project tests pass: all Go packages and **222 frontend tests**. Lint passes. The two additional reproduction tests fail as described above. The working tree remains unchanged.

**🔧 Bash** ```bash
git log --oneline | head -3 && git status --short | grep -v ai-sessions; grep -n "consent" backend/internal/store/migrations/*.sql | head; grep -n "func seeded\|func planned\|func getJob" -A6 backend/internal/delivery/*_test.go | head -40; grep -n "Handle(" -B6 -A12 backend/internal/delivery/worker.go | head -60
```

**🔧 Skill** ```json
{
  "skill": "superpowers:receiving-code-review"
}
```

<details><summary>Result</summary>

```
280a142 export ai and fixes
c5d728f fix(08): final review fixes for the dashboard
f7bb4f4 docs(08): dashboard architecture, decisions D-068 to D-082, backlog and progress
backend/internal/store/migrations/00001_users_events.sql:12:    email_consent    boolean NOT NULL,
backend/internal/store/migrations/00001_users_events.sql:13:    sms_consent      boolean NOT NULL,
backend/internal/delivery/store_test.go:33:func seeded(t *testing.T) (context.Context, *pgxpool.Pool, *Store) {
backend/internal/delivery/store_test.go-34-	t.Helper()
backend/internal/delivery/store_test.go-35-	return seededN(t, 500)
backend/internal/delivery/store_test.go-36-}
backend/internal/delivery/store_test.go-37-
backend/internal/delivery/store_test.go:38:func seededN(t *testing.T, users int) (context.Context, *pgxpool.Pool, *Store) {
backend/internal/delivery/store_test.go-39-	t.Helper()
backend/internal/delivery/store_test.go-40-	url := storetest.NewDB(t)
backend/internal/delivery/store_test.go-41-	ctx := context.Background()
backend/internal/delivery/store_test.go-42-	if err := store.Migrate(ctx, url); err != nil {
backend/internal/delivery/store_test.go-43-		t.Fatalf("migrate: %v", err)
backend/internal/delivery/store_test.go-44-	}
--
backend/internal/delivery/store_test.go:75:func planned(t *testing.T, ctx context.Context, pool *pgxpool.Pool, s *Store) (int64, map[Channel][]int64) {
backend/internal/delivery/store_test.go-76-	t.Helper()
backend/internal/delivery/store_test.go-77-	id := approved(t, ctx, pool, exampleRule, "tr")
backend/internal/delivery/store_test.go-78-	cid, ids, ok, err := s.PlanNext(ctx, segment.Default(), testNow)
backend/internal/delivery/store_test.go-79-	if err != nil || !ok || cid != id {
backend/internal/delivery/store_test.go-80-		t.Fatalf("PlanNext = %d, %v, %v, %v", cid, ok, err, id)
backend/internal/delivery/store_test.go-81-	}
--
backend/internal/delivery/store_test.go:94:func getJob(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id int64) Job {
backend/
… [1221 more chars truncated]
```

</details>

<details><summary>Result</summary>

```
Launching skill: superpowers:receiving-code-review
```

</details>

**🔧 Bash** ```bash
cat >> backend/internal/delivery/consumer_test.go <<'EOF'

// Consent that the user takes back during the token wait must stop the send.
func TestConsumer_ConsentRevokedDuringWait(t *testing.T) {
	e := newConsumerEnv(t)
	srv := newMock(t)
	c := e.consumer(NewProvider(srv.URL, 2*time.Second))
	var once sync.Once
	e.clk.onSleep = func() {
		once.Do(func() {
			if _, err := e.pool.Exec(e.ctx, `UPDATE users SET sms_consent = false
WHERE id = (SELECT user_id FROM send_jobs WHERE id = $1)`, e.sms[0]); err != nil {
				t.Error(err)
			}
		})
	}
	es := e.entries(t, e.sms[0], 1)
	if err := c.Handle(e.ctx, es[0]); err != nil {
		t.Fatal(err)
	}
	if got := providerStats(t, srv.URL)["sms"].Requests; got != 0 {
		t.Fatalf("requests = %d, want 0 after the consent was revoked", got)
	}
	if j := getJob(t, e.ctx, e.pool, e.sms[0]); j.Status != StatusSkipped || j.LastError != SkipNoConsent {
		t.Fatalf("job = %s %q, want skipped %q", j.Status, j.LastError, SkipNoConsent)
	}
}

// A job that waits for a token is not in sending, so the recovery cannot set
// it to unknown while its consumer is alive.
func TestConsumer_WaitingJobIsNotExpired(t *testing.T) {
	e := newConsumerEnv(t)
	srv := newMock(t)
	c := e.consumer(NewProvider(srv.URL, 2*time.Second))
	var once sync.Once
	e.clk.onSleep = func() {
		once.Do(func() {
			if _, err := e.store.ExpireSending(e.ctx, testNow.Add(time.Hour), time.Second); err != nil {
				t.Error(err)
			}
		})
	}
	es := e.entries(t, e.sms[0], 1)
	if err := c.Handle(e.ctx, es[0]); err != nil {
		t.Fatal(err)
	}
	if got := providerStats(t, srv.URL)["sms"].Requests; got != 1 {
		t.Fatalf("requests = %d, want 1", got)
	}
	if j := getJob(t, e.ctx, e.pool, e.sms[0]); j.Status != StatusSent {
		t.Fatalf("job = %s, want sent", j.Status)
	}
}
EOF
(docker rm -f cc-test-pg cc-test-redis >/dev/null 2>&1
docker run -d --name cc-test-pg --network cc-test -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=test -p 127.0.0.1:55432:5432 postgres:17-alpine >/dev/null
docker run -d --name cc-test-redis --network cc-test -p 127.0.0.1:56379:6379 redis:8-alpine >/dev/null
i=0; until docker exec cc-test-pg pg_isready -U test -d test -h 127.0.0.1 >/dev/null 2>&1; do i=$((i+1)); [ $i -gt 60 ] && break; sleep 1; done)
export TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55432/test?sslmode=disable' TEST_REDIS_URL='redis://127.0.0.1:56379/0'
cd backend && go test -count=1 -run 'TestConsumer_(ConsentRevoked|WaitingJob)' ./internal/delivery/ 2>&1 | tail -12
```

<details><summary>Result</summary>

```
--- FAIL: TestConsumer_ConsentRevokedDuringWait (0.16s)
    consumer_test.go:488: requests = 1, want 0 after the consent was revoked
--- FAIL: TestConsumer_WaitingJobIsNotExpired (0.10s)
    consumer_test.go:517: job = unknown, want sent
FAIL
FAIL	campaigncomposer/internal/delivery	3.307s
FAIL
```

</details>

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one && python3 - <<'EOF'
p='backend/internal/delivery/consumer.go'
s=open(p).read()
a=s.index("// Handle runs one entry")
b=s.index("\treference := fmt.Sprintf(")
new='''// defaultMaxPrep is the default of Consumer.MaxPrep.
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
			return err
		}
		if !ok {
			return c.Queue.Ack(done, c.Channel, e.StreamID)
		}
		if c.afterClaim != nil {
			c.afterClaim()
		}
		// Nothing was sent yet, so the job can go back to the queue.
		release := func() {
			if rerr := c.Store.Release(done, job.ID); rerr != nil {
				c.log().Error("release job", "job_id", job.ID, "channel", c.Channel, "error", rerr)
			}
		}
		if err := ctx.Err(); err != nil {
			release()
			return err
		}

		// The load reads the consent. It is the last step before the
		// request, so a change of consent during the wait stops the send.
		var skip string
		msg, skip, err = c.Store.Load(ctx, job)
		if err != nil {
			release()
			return err
		}
		if skip != "" {
			c.log().Info("send job skipped", "job_id", job.ID, "channel", c.Channel, "attempt", job.Attempts, "error", skip)
			return c.finish(done, e, job, Result{Status: StatusSkipped, LastError: skip})
		}

		// A token is good for MaxPrep. When the database was slower, the
		// sends that it held would leave together and break the rate. Such
		// a job goes back to queued and takes a new token.
		prep := c.Clock.Now().Sub(granted)
		if prep <= maxPrep {
			break
		}
		c.log().Warn("send job prepared too slowly, new rate token", "job_id", job.ID, "channel", c.Channel, "took", prep)
		release()
	}

'''
s=s[:a]+new+s[b:]
s=s.replace('''	Log         *slog.Logger
''','''	Log         *slog.Logger
	// MaxPrep is the longest time between a rate token and the request.
	// The claim and the load run in this time. 0 means 
… [2300 more chars truncated]
```

<details><summary>Result</summary>

```
internal/delivery/consumer.go
ok  	campaigncomposer/internal/delivery	46.599s
ok  	campaigncomposer/cmd/api	1.258s
ok  	campaigncomposer/cmd/worker	1.663s
ok  	campaigncomposer/internal/delivery	13.631s
```

</details>

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one && gofmt -w backend/internal/delivery/consumer.go && grep -n "ports:" -A2 docker-compose.yml | grep -v "^--" ; ls; docker ps --format '{{.Names}} {{.Ports}}' | cut -c1-90
```

<details><summary>Result</summary>

```
16:    # No ports: Postgres is reachable only inside the compose network.
17-    # This avoids a conflict with a Postgres on the host's port 5432.
18-
49:    ports:
50-      - "8080:8080"
51-    depends_on:
102:    ports:
103-      - "8081:8081"
104-
124:    ports:
125-      - "8082:8082"
126-    depends_on:
138:    ports:
139-      - "3000:8080"
140-    depends_on:
CLAUDE.md
DECISIONS.md
Dockerfile
Makefile
README.md
ai-sessions
backend
docker-compose.override.yml
docker-compose.yml
docs
frontend
tasks
cc-test-redis 127.0.0.1:56379->6379/tcp
cc-test-pg 127.0.0.1:55432->5432/tcp
insider_one-frontend-1 0.0.0.0:3000->8080/tcp, [::]:3000->8080/tcp
insider_one-worker-1 0.0.0.0:8082->8082/tcp, [::]:8082->8082/tcp
insider_one-api-1 0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp
insider_one-mock-provider-1 0.0.0.0:8081->8081/tcp, [::]:8081->8081/tcp
insider_one-redis-1 6379/tcp
insider_one-postgres-1 127.0.0.1:5433->5432/tcp
```

</details>

**🔧 Bash** ```bash
grep -rn 'HandleFunc\|Handle("' backend/cmd/api/main.go backend/internal/httpapi/*.go | grep -v _test | cut -c1-170 | head -30; grep -n "rate_limited\|json:\"" backend/internal/mockprovider/*.go | grep -v _test | head -20
```

<details><summary>Result</summary>

```
backend/cmd/api/main.go:142:	mux.Handle("GET /metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
backend/internal/httpapi/campaigns.go:31:	mux.HandleFunc("GET /api/campaigns", h.list)
backend/internal/httpapi/campaigns.go:32:	mux.HandleFunc("GET /api/campaigns/{id}", h.get)
backend/internal/httpapi/campaigns.go:33:	mux.HandleFunc("PUT /api/campaigns/{id}/contents", h.save)
backend/internal/httpapi/campaigns.go:34:	mux.HandleFunc("POST /api/campaigns/{id}/approve", h.approve)
backend/internal/httpapi/campaigns.go:35:	mux.HandleFunc("GET /api/campaigns/{id}/sample", h.sample)
backend/internal/httpapi/content.go:43:	mux.HandleFunc("POST /api/content/generate", g.generate)
backend/internal/httpapi/dashboard.go:32:	mux.HandleFunc("GET /api/dashboard/stats", h.getStats)
backend/internal/httpapi/dashboard.go:33:	mux.HandleFunc("GET /api/dashboard/health", h.getHealth)
backend/internal/httpapi/delivery.go:37:	mux.HandleFunc("GET /api/campaigns/{id}/progress", h.progress)
backend/internal/httpapi/delivery.go:38:	mux.HandleFunc("GET /api/campaigns/{id}/progress/stream", h.stream)
backend/internal/httpapi/delivery.go:39:	mux.HandleFunc("GET /api/campaigns/{id}/send-jobs", h.sendJobs)
backend/internal/httpapi/delivery.go:40:	mux.HandleFunc("POST /api/campaigns/{id}/retry", h.retry)
backend/internal/httpapi/health.go:30:	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
backend/internal/httpapi/health.go:34:	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
backend/internal/httpapi/segments.go:34:	mux.HandleFunc("POST /api/segments/preview", p.preview)
backend/internal/mockprovider/server.go:87:		writeError(w, http.StatusTooManyRequests, "rate_limited", "too many requests for channel "+channel)
backend/internal/mockprovider/store.go:17:	ID         string    `json:"id"`
backend/internal/mockprovider/store.go:18:	Channel    string    `json:"channel"`
backend/internal/mockprovider/store.go:19:	To         string    `jso
… [1346 more chars truncated]
```

</details>

**🔧 Bash** ```bash
S="/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/22a0e40d-eeb7-4ae5-9612-495cb6c67262/scratchpad" && cat > "$S/ports.yml" <<'EOF'
services:
  api:
    ports: !override
      - "18080:8080"
  mock-provider:
    ports: !override
      - "18081:8081"
  worker:
    ports: !override
      - "18082:8082"
  frontend:
    ports: !override
      - "13000:8080"
EOF
LLM_MODE=mock docker compose -p cc-verify -f docker-compose.yml -f "$S/ports.yml" up -d --build --wait > "$S/up.log" 2>&1; echo "up=$?"; tail -3 "$S/up.log"
curl -sS -N -X POST localhost:18080/api/content/generate -H 'Content-Type: application/json' -d '{"brief":"Istanbul users who added to cart in the last 30 days but did not buy. 30% off."}' | tail -c 600
```

<details><summary>Result</summary>

```
up=0
 Container cc-verify-postgres-1 Healthy 
 Container cc-verify-redis-1 Healthy 
 Container cc-verify-api-1 Healthy 
{"error":{"code":"invalid_request","message":"The request body must be JSON with a brief and a rule, of at most 16 KiB."}}
```

</details>

**🔧 Bash** ```bash
B='Istanbul users who added to cart in the last 30 days but did not buy. 30% off.'
P=$(curl -sS -X POST localhost:18080/api/segments/preview -H 'Content-Type: application/json' -d "{\"brief\":\"$B\"}"); echo "$P" | cut -c1-500
```

<details><summary>Result</summary>

```
{"rule":{"op":"and","conditions":[{"field":"city","op":"eq","value":"Istanbul"},{"field":"last_cart_add_at","op":"within_days","value":30},{"op":"or","conditions":[{"field":"last_purchase_at","op":"is_null"},{"field":"last_purchase_at","op":"older_than_days","value":30}]}]},"description":"city is Istanbul AND last cart add in the last 30 days AND (no purchase OR last purchase more than 30 days ago)","unmapped":[],"audience":{"email":148,"sms":101,"reachable":183,"matched":212,"no_consent":29,"la
```

</details>

**🔧 Bash** ```bash
S="/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/22a0e40d-eeb7-4ae5-9612-495cb6c67262/scratchpad"
curl -sS -N -X POST localhost:18080/api/content/generate -H 'Content-Type: application/json' -d '{"brief":"Istanbul users, 30% off.","rule":{"op":"and","conditions":[{"field":"city","op":"eq","value":"Istanbul"}]}}' > "$S/gen.txt"; grep -o '"campaign_id":[0-9]*' "$S/gen.txt" | head -2; tail -c 300 "$S/gen.txt"; echo; curl -sS localhost:18080/api/campaigns | cut -c1-300
```

<details><summary>Result</summary>

```
ğin ürünler hâlâ seni bekliyor.\n\nSiparişini bu hafta tamamlarsan %30 indirimden yararlanabilirsin.\n\nGörüşmek üzere!"},"limits":{"first_name_runes":7,"sms_runes":76,"email_subject_runes":44,"email_body_runes":148}}

event: campaign
data: {"id":1,"status":"draft"}

event: done
data: {}


{"campaigns":[{"id":1,"status":"draft","description":"city is Istanbul","languages":["tr"],"created_at":"2026-11-11T10:43:07.650226Z","approved_at":null}]}
```

</details>

**🔧 Bash** ```bash
S="/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/22a0e40d-eeb7-4ae5-9612-495cb6c67262/scratchpad"
C=$(curl -sS localhost:18080/api/campaigns/1 | python3 -c 'import json,sys; c=json.load(sys.stdin); print(json.dumps({"contents":[{k:x[k] for k in ("language","sms","email_subject","email_body")} for x in c["contents"]]}))')
curl -sS -X POST localhost:18080/api/campaigns/1/approve -H 'Content-Type: application/json' -d "$C" | cut -c1-200; echo
for i in $(seq 1 40); do P=$(curl -sS localhost:18080/api/campaigns/1/progress); echo "$P" | grep -q '"status":"completed"' && break; /bin/sleep 1 2>/dev/null || python3 -c 'import time; time.sleep(1)'; done; echo "$P" | cut -c1-600; echo
curl -sS localhost:18081/v1/stats; echo
curl -sS 'localhost:18081/v1/messages?limit=100000' | cut -c1-300
```

<details><summary>Result</summary>

```
{"id":1,"status":"approved","brief":"Istanbul users, 30% off.","description":"city is Istanbul","skipped_languages":[],"created_at":"2026-11-11T10:43:07.650226Z","updated_at":"2026-11-11T10:43:15.0411

{"campaign_id":1,"status":"completed","channels":{"email":{"total":1014,"queued":0,"sending":0,"retrying":0,"sent":966,"failed":0,"unknown":48,"skipped":0,"retryable_failed":0},"sms":{"total":737,"queued":0,"sending":0,"retrying":0,"sent":716,"failed":0,"unknown":21,"skipped":0,"retryable_failed":0}}}

{"email":{"requests":1051,"delivered":1008,"rejected":0,"rate_limited":0,"errors":37,"slow":37,"lost":11,"duplicate_deliveries":0},"sms":{"requests":752,"delivered":737,"rejected":0,"rate_limited":0,"errors":15,"slow":16,"lost":5,"duplicate_deliveries":0}}

{"error":{"code":"invalid_request","message":"limit must be an integer from 1 to 1000"}}
```

</details>

**🔧 Bash** ```bash
docker logs cc-verify-worker-1 2>&1 | grep -c "prepared too slowly"; docker exec cc-verify-postgres-1 psql -U campaign -d campaign -Atc "select channel, max(n), count(*) from (select channel, date_trunc('second', sent_at) s, count(*) n from send_jobs where sent_at is not null group by 1,2) t group by 1"; S="/private/tmp/claude-501/-Users-USER-Desktop-projects-case-studies-insider-one/22a0e40d-eeb7-4ae5-9612-495cb6c67262/scratchpad"; docker compose -p cc-verify -f docker-compose.yml -f "$S/ports.yml" down -v > /dev/null 2>&1; echo down=$?
```

<details><summary>Result</summary>

```
0
email|98|14
sms|98|9
down=0
```

</details>

**🔧 Bash** ```bash
python3 - <<'EOF'
def sub(p, pairs):
    s=open(p).read()
    for a,b in pairs:
        assert a in s, a[:50]
        s=s.replace(a,b,1)
    open(p,'w').write(s)

sub('docs/specs/07-send-queue/goal.md',[
("1. **Claim.**","1. **Wait** for a token of the channel rate limiter. The send job stays `queued` during the wait.\n2. **Claim.**"),
("\n2. **Load.**","\n3. **Load.**"),
("\n3. **Check.**","\n4. **Check.**"),
("\n4. **Wait** for a token of the channel rate limiter. No database work comes between the token and the request (D-084).\n5. **Send**",
 "\n5. **Check the token age.** If steps 2 to 4 took more than 100 ms, release the send job to `queued` and go to step 1 (D-086).\n6. **Send**"),
("\n6. **Write the result**","\n7. **Write the result**"),
("\n7. **Acknowledge**","\n8. **Acknowledge**"),
("- The worker claims a send job, then waits for a rate-limit token immediately before the request (D-084 replaced the first order). On `SIGTERM`,",
 "- The worker waits for a rate-limit token before it claims a send job. A token is good for 100 ms (D-086). On `SIGTERM`,"),
])
sub('docs/architecture.md',[
("Each consumer claims the send job with one conditional `UPDATE`. The claim is the dedup gate (D-046). The consumer waits for a rate-limit token immediately before the provider request (D-084).",
 "Each consumer waits for a rate-limit token, then claims the send job with one conditional `UPDATE`. The claim is the dedup gate (D-053, D-046). A token is good for 100 ms. If the claim and the load take longer, the send job goes back to `queued` and takes a new token (D-086)."),
])
sub('docs/decisions-log.md',[
("\n\n**Replaced by D-084.**\n","\n\n**D-084 replaced this order. D-086 restored it and added a limit on the token age.**\n"),
("\n\n**D-084 removed the `Store.Claimable` check. The lag gate of the recovery stays.**\n","\n\n**D-084 removed the `Store.Claimable` check. D-086 restored it.**\n"),
("## D-084 The worker waits for a token immediately before the request\n","## D-084 The worker waits for a token immediately before the request\n\n**Replaced by D-086. This order caused two defects.**\n"),
])
with open('docs/decisions-log.md','a') as f:
    f.write('''
## D-086 A rate token is good for 100 ms

- **Context:** D-084 put the token wait after the claim and the load. A review found two defects in that order. The load reads the consent, so a change of consent during the wait did not stop the send. The send job was `sending` during the wait, so the recovery could set a live send job to `unknown`. The first order (D-053) had neither defect, but a slow database could release held sends together.
- **Options:**
  - Keep the D-084 order and read the consent again after the wait. This puts database work after the token again.
  - Keep the D-084 order and add a column that separates a waiting send job from a started request.
  - Restore the D-053 order and limit the age of a token.
- **Choice:** The order is wait, claim, load, send. The consumer measures the t
… [2531 more chars truncated]
```

<details><summary>Result</summary>

```
docs/specs/07-send-queue/goal.md:161: sentence-length: 26 words: **Check the token age.** If steps 2 to 4 took more than 100 ms, release the send job to CODE and go to step 1 (D-086).
exit=0
ok  	campaigncomposer/internal/delivery	42.646s
      Tests  222 passed (222)
```

</details>

## 🤖 Claude  <sub>2026-11-11 10:45:12</sub>

Both findings were correct, and both came from my previous fix (D-084). They are fixed now, `make verify` passes, and nothing is committed.

## Fix

The worker order in [consumer.go](backend/internal/delivery/consumer.go) is again wait, claim, load, send, with one new rule: a rate token is good for 100 ms. If the claim and the load take longer, the send job goes back to `queued` and takes a new token.

| Finding | How the new order covers it |
|---|---|
| Consent revoked during the wait (P1) | The load reads the consent after the wait. It is the last step before the request. |
| Recovery expires a waiting job (P2) | A send job is `queued` during the wait. It is `sending` only around the request. |
| Pacing after a database stall (first review) | A send that a slow database held has an old token, so it takes a new one. |

I also restored `Store.Claimable`, so a duplicate queue entry again uses no rate token.

## Evidence

- **Reproduction:** `TestConsumer_ConsentRevokedDuringWait` showed one request sent, and `TestConsumer_WaitingJobIsNotExpired` showed a job left in `unknown`, before the fix.
- **After the fix:** both tests and the pacing test of the first review passed 15 of 15 runs with `-race`.
- **Compose run:** a separate stack sent 1,803 requests to the mock provider. It gave `rate_limited: 0`, `duplicate_deliveries: 0`, no token-age warning, and 98 sends at most in one second for each channel.

## Limits of this design

- Sends whose tokens are less than 100 ms apart can leave together. At 100 per second, this is 10 requests at most, which equals the burst of the mock provider. A stall can thus put a channel up to 10 requests above 100 in one second.
- If the claim and the load always take more than 100 ms, the worker sends nothing. It logs a warning for each new token.

## Docs

- [docs/decisions-log.md](docs/decisions-log.md): D-086 replaces D-084. D-053 and D-066 are in effect again.
- [docs/architecture.md](docs/architecture.md) and [docs/specs/07-send-queue/goal.md](docs/specs/07-send-queue/goal.md): the step order.
- [tasks/progress.md](tasks/progress.md): a session entry.

The running stack on ports 3000 and 8080 still has the old worker until you rebuild it.

