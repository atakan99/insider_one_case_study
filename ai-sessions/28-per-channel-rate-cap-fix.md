# Per-channel rate cap fix

- Session: `d8f8302d-048b-4168-9dc1-18f4deb6608e`
- Model: `claude-opus-5-5`
- Time (UTC): 2026-11-11 16:29:16 → 2026-11-11 16:34:12
- Subagent runs: 0

---

## 👤 User  <sub>2026-11-11 16:29:16</sub>

_[IDE: selection]_ 



**Almost, but I would fix one remaining requirement violation before submitting.**

An isolated check reproduced **108 provider requests within one second**, exceeding the required 100-per-channel cap. The [100 ms preparation allowance](~/Desktop/projects/case-studies/insider_one/backend/internal/delivery/consumer.go:106) still lets permitted sends bunch together.

Everything else checked passes:

- Fresh Go integration tests and **234 frontend tests**
- Lint, frontend production build, and Compose configuration validation
- Clean working tree; session exports and submission documents present
- Regression tests for the previous consent and recovery issues

Also correct the `sqlc` claims in [DECISIONS.md](~/Desktop/projects/case-studies/insider_one/DECISIONS.md:7); the implementation uses hand-written SQL.

After the rate-cap fix, I’d do one final fresh-install and browser-flow check. I haven’t verified those or real Claude calls in this pass.


we need a rate cap fix

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one/backend/internal/delivery && ls && wc -l *.go && sed -n 1,200p consumer.go
```

<details><summary>Result</summary>

```
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
     152 consumer.go
     526 consumer_test.go
      43 fakeclock_test.go
      75 limiter.go
      59 limiter_test.go
     128 metrics.go
     139 metrics_test.go
      38 policy.go
      66 policy_test.go
     180 provider.go
     209 provider_test.go
     189 queue.go
     217 queue_test.go
      51 service.go
     133 service_test.go
     543 store.go
     666 store_test.go
     104 types.go
     316 worker.go
     513 worker_test.go
    4347 total
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

	// An entry for a job that cannot 
… [3440 more chars truncated]
```

</details>

**🔧 Bash** ```bash
cat limiter.go limiter_test.go && grep -rn "Limiter\|NewLimiter\|Workers\|concurr" worker.go ../../cmd/*/main.go | head -40
```

<details><summary>Result</summary>

```
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
package delivery

import (
	"context"
	"testing"
	"time"
)

func TestLimiter_Rate(t *testing.T) {
	clk := newFakeClock(testNow)
	l := NewLimiter(100, clk)
	var times []time.Time
	for i := 0; i < 1000; i++ {
		if err := l.Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
		times = append(times, clk.Now())
	}
	// For every call, count the calls in the 1-second window that starts there.
	j := 0
	fo
… [1135 more chars truncated]
```

</details>

**🔧 Bash** ```bash
sed -n 60,140p worker.go && grep -n "MaxPrep\|afterClaim\|108\|per second\|window" consumer_test.go worker_test.go | head -30; grep -rn "replicas\|worker" ../../../docker-compose*.yml ../../../compose*.y*ml 2>/dev/null | head
```

<details><summary>Result</summary>

```
func orDefault(d, def time.Duration) time.Duration {
	if d > 0 {
		return d
	}
	return def
}

// Run blocks until ctx ends. It then stops reading and waits, and gives the
// requests that started up to 10 seconds to end.
func (w *Worker) Run(ctx context.Context) error {
	if w.Clock == nil {
		w.Clock = RealClock()
	}
	if err := w.Queue.Ensure(ctx); err != nil {
		return err
	}
	var wg sync.WaitGroup
	loop := func(every time.Duration, fn func(context.Context)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.tick(ctx, every, fn)
		}()
	}
	loop(orDefault(w.PlanEvery, defaultPlanEvery), w.plan)
	loop(orDefault(w.RetryEvery, defaultRetryEvery), w.pollRetries)
	loop(orDefault(w.RecoverEvery, defaultRecoverEvery), w.recover)

	for _, ch := range Channels {
		c := &Consumer{
			Store: w.Store, Queue: w.Queue, Provider: w.Provider,
			Limiter: NewLimiter(w.Cfg.RatePerSec, w.Clock),
			Channel: ch, Clock: w.Clock, MaxAttempts: w.Cfg.MaxAttempts,
			Jitter: rand.Float64, Log: w.log(), Metrics: w.Metrics,
		}
		for range w.Cfg.Consumers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				w.consume(ctx, c)
			}()
		}
		loop(orDefault(w.RecoverEvery, defaultRecoverEvery), func(ctx context.Context) { w.claimDead(ctx, c) })
	}

	<-ctx.Done()
	w.log().Info("worker stopping")
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(shutdownGrace):
		w.log().Warn("worker stopped with requests still running; the recovery sets them to unknown")
	}
	return nil
}

// tick runs fn now and then after each interval, until ctx ends. A run
// never overlaps the next one.
func (w *Worker) tick(ctx context.Context, every time.Duration, fn func(context.Context)) {
	for ctx.Err() == nil {
		fn(ctx)
		if w.Clock.Sleep(ctx, every) != nil {
			return
		}
	}
}

// push puts the ids on the streams and marks them as enqueued. A failure is
// logged only: the jobs stay queued, and the recovery pushes them again.
func (w *Worker) push(ctx contex
… [789 more chars truncated]
```

</details>

**🔧 Bash** ```bash
sed -n 380,526p consumer_test.go && cat fakeclock_test.go && grep -n "Clock interface" -A8 *.go && grep -n "MaxPrep\|prep\|D-0" ../../../docs/decisions-log.md | head
```

<details><summary>Result</summary>

```
defer cancel()
	p.HTTP.Transport = cancelTransport{base: p.HTTP.Transport, cancel: cancel}
	c := e.consumer(p)
	es := e.entries(t, e.sms[0], 1)
	if err := c.Handle(ctx, es[0]); err != nil {
		t.Fatalf("Handle = %v", err)
	}
	j := getJob(t, e.ctx, e.pool, e.sms[0])
	if j.Status != StatusUnknown && j.Status != StatusSent {
		t.Fatalf("job = %s %q, want unknown or sent", j.Status, j.LastError)
	}
	if strings.Contains(j.LastError, "cancel") {
		t.Fatalf("last_error = %q", j.LastError)
	}
	if n := e.pending(t); n != 0 {
		t.Fatalf("pending = %d", n)
	}
}

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

	// Every consumer stalls after its first claim until all n jobs are
	// claimed and each rate token is older than MaxPrep.
	c.MaxPrep = 50 * time.Millisecond
	var claimed sync.WaitGroup
	claimed.Add(n)
	stall := make(chan struct{})
	go func() {
		claimed.Wait()
		time.Sleep(2 * c.MaxPrep)
		close(stall)
	}()
	var claims atomic.Int32
	c.afterClaim = func() {
		if claims.Add(1) <= n {
			claimed.Done()
		}
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

… [4199 more chars truncated]
```

</details>

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one/backend && ls cmd internal; grep -rln "window\|perSecond\|429" cmd internal --include=*.go | grep -v _test | head; grep -n "RatePerSec\|Consumers" -r internal cmd --include=*.go | grep -v _test | head
```

<details><summary>Result</summary>

```
cmd:
api
mockprovider
seed
worker

internal:
campaign
config
content
dashboard
delivery
httpapi
llm
mockprovider
seed
segment
store
(eval):1: no matches found: --include=*.go
(eval):1: no matches found: --include=*.go
```

</details>

**🔧 Bash** ```bash
grep -rn "RatePerSec\|Consumers\b" --include='*.go' . | grep -v _test | head; grep -rln "429\|Window\|window" --include='*.go' internal/mockprovider cmd/mockprovider
```

<details><summary>Result</summary>

```
cmd/mockprovider/main.go:38:		"rate_per_sec", cfg.RatePerSec, "burst", cfg.Burst,
cmd/mockprovider/main.go:51:			RatePerSec: cfg.RatePerSec,
internal/mockprovider/server.go:37:			channelSMS:   rate.NewLimiter(rate.Limit(o.RatePerSec), o.Burst),
internal/mockprovider/server.go:38:			channelEmail: rate.NewLimiter(rate.Limit(o.RatePerSec), o.Burst),
internal/mockprovider/server.go:105:		rateLimited(retryAfter(lim, now, s.o.RatePerSec))
internal/config/config.go:157:	RatePerSec float64
internal/config/config.go:173:		RatePerSec: 100,
internal/config/config.go:192:			cfg.RatePerSec = f
internal/config/config.go:263:	// RatePerSec is SEND_RATE_PER_SEC, the send rate of all consumers together.
internal/config/config.go:264:	RatePerSec int
internal/mockprovider/server.go
internal/mockprovider/options.go
internal/mockprovider/failure_test.go
internal/mockprovider/limit_test.go
internal/mockprovider/failure.go
```

</details>

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one/backend/internal/delivery && grep -n "func newConsumerEnv" -A60 consumer_test.go | head -90; sed -n 1,25p consumer_test.go
```

<details><summary>Result</summary>

```
51:func newConsumerEnv(t *testing.T) *consumerEnv {
52-	t.Helper()
53-	ctx, pool, s := seeded(t)
54-	_, ids := planned(t, ctx, pool, s)
55-	if len(ids[ChannelSMS]) < 7 {
56-		t.Fatalf("fixture has %d sms jobs, want at least 7", len(ids[ChannelSMS]))
57-	}
58-	q := NewQueue(deliverytest.NewRedis(t, 0), "c1")
59-	if err := q.Ensure(ctx); err != nil {
60-		t.Fatal(err)
61-	}
62-	return &consumerEnv{ctx: ctx, pool: pool, store: s, queue: q, clk: newFakeClock(testNow),
63-		srv: newMock(t).URL, logs: &syncBuf{}, sms: ids[ChannelSMS]}
64-}
65-
66-func (e *consumerEnv) consumer(p *Provider) *Consumer {
67-	return &Consumer{
68-		Store: e.store, Queue: e.queue, Provider: p, Limiter: NewLimiter(1000, e.clk),
69-		Channel: ChannelSMS, Clock: e.clk, MaxAttempts: 5, Jitter: func() float64 { return 0 },
70-		Log: slog.New(slog.NewTextHandler(e.logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
71-	}
72-}
73-
74-// entries pushes id n times and reads the new entries.
75-func (e *consumerEnv) entries(t *testing.T, id int64, n int) []Entry {
76-	t.Helper()
77-	ids := make([]int64, n)
78-	for i := range ids {
79-		ids[i] = id
80-	}
81-	if err := e.queue.Push(e.ctx, ChannelSMS, ids); err != nil {
82-		t.Fatal(err)
83-	}
84-	es, err := e.queue.Read(e.ctx, ChannelSMS, n, time.Second)
85-	if err != nil || len(es) != n {
86-		t.Fatalf("Read = %v, %v", es, err)
87-	}
88-	return es
89-}
90-
91-func (e *consumerEnv) pending(t *testing.T) int64 {
92-	t.Helper()
93-	p, err := e.queue.rdb.XPending(e.ctx, streamKey(ChannelSMS), groupName).Result()
94-	if err != nil {
95-		t.Fatal(err)
96-	}
97-	return p.Count
98-}
99-
100-func (e *consumerEnv) nextAttempt(t *testing.T, id int64) time.Time {
101-	t.Helper()
102-	var at time.Time
103-	if err := e.pool.QueryRow(e.ctx, `SELECT next_attempt_at FROM send_jobs WHERE id = $1`, id).Scan(&at); err != nil {
104-		t.Fatal(err)
105-	}
106-	return at
107-}
108-
109-func TestConsumer_Sequences(t *testing.T) {
110-	e := newConsumerEnv(t)
111-	type step stru
… [324 more chars truncated]
```

</details>

**🔧 Bash** ```bash
python3 - <<'EOF'
import re
p='limiter.go'; s=open(p).read()
s=s.replace('''// Limiter limits the sends of one channel to perSec each second, with no
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
''','''// Limiter limits the sends of one channel to perSec each second, with no
// burst, and can be paused after a 429.
//
// It has two steps. Wait paces the claims. Admit runs just before the
// request and keeps the request starts at least gap apart, so a slow
// preparation cannot bunch the sends of tokens that were granted in time.
type Limiter struct {
	rl    *rate.Limiter
	clock Clock
	gap   time.Duration

	mu          sync.Mutex
	pausedUntil time.Time
	nextSend    time.Time
}

// gapMargin makes the gap between two request starts 2% longer than
// 1s/perSec. Network jitter then cannot put perSec+1 requests into one
// second at the provider.
const gapMargin = 50

// NewLimiter returns a Limiter with a burst of 1.
func NewLimiter(perSec int, clock Clock) *Limiter {
	gap := time.Second / time.Duration(perSec)
	return &Limiter{
		rl:    rate.NewLimiter(rate.Limit(perSec), 1),
		clock: clock,
		gap:   gap + gap/gapMargin,
	}
}

// Admit blocks until gap has passed since the start of the last admitted
// request. Call it just before the request. A cancelled Admit keeps its slot.
func (l *Limiter) Admit(ctx context.Context) error {
	l.mu.Lock()
	now := l.clock.Now()
	at := l.nextSend
	if at.Before(now) {
		at = now
	}
	l.nextSend = at.Add(l.gap)
	l.mu.Unlock()
	if d := at.Sub(now); d > 0 {
		return l.clock.Sleep(ctx, d)
	}
	return ctx.Err()
}
''')
open(p,'w').write(s)

p='consumer.go'; s=open(p).read()
old='''		c.log().Warn("send job prepared too slowly, new rate token", "job_id", job.ID, "channel", c.Channel, "took", prep)
		release()
	}
'''
new='''		c.log().Warn("send job prepared too slowly, new rate token", "job_id", job.ID, "channel", c.Channel, "took", prep)
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
'''
assert old in s; s=s.replace(old,new); open(p,'w').write(s)
EOF
grep -n "func TestConsumer_ConsentRevokedDuringWait" consumer_test.go
```

<details><summary>Result</summary>

```
477:func TestConsumer_ConsentRevokedDuringWait(t *testing.T) {
```

</details>

**🔧 Bash** ```bash
cat >> limiter_test.go <<'EOF'

// Tokens are granted on time, but the first sends are prepared slowly and the
// later ones fast. Admit must still keep every second at perSec or fewer.
func TestLimiter_AdmitSpacesSlowPreparation(t *testing.T) {
	clk := newFakeClock(testNow)
	const perSec = 100
	l := NewLimiter(perSec, clk)
	type send struct{ token, at time.Time }
	var sends []send
	for i := 0; i < 300; i++ {
		if err := l.Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
		sends = append(sends, send{token=[REDACTED]
	}
	// Prep of 90 ms for the first half, 0 for the rest: without Admit, the
	// sends of both halves overlap and pass perSec in one second.
	var at []time.Time
	for i, s := range sends {
		t0 := s.token
		if i < 150 {
			t0 = t0.Add(90 * time.Millisecond)
		}
		if clk.Now().Before(t0) {
			clk.Set(t0)
		}
		if err := l.Admit(context.Background()); err != nil {
			t.Fatal(err)
		}
		at = append(at, clk.Now())
	}
	slices.SortFunc(at, func(a, b time.Time) int { return a.Compare(b) })
	j := 0
	for i := range at {
		for j < len(at) && at[j].Before(at[i].Add(time.Second)) {
			j++
		}
		if n := j - i; n > perSec {
			t.Fatalf("%d sends in the second from send %d", n, i)
		}
	}
}

func TestLimiter_AdmitCancelled(t *testing.T) {
	clk := newFakeClock(testNow)
	l := NewLimiter(1, clk)
	if err := l.Admit(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := l.Admit(ctx); err != context.Canceled {
		t.Fatalf("Admit = %v, want context.Canceled", err)
	}
}
EOF
sed -i '' 's/^\t"context"$/\t"context"\n\t"slices"/' limiter_test.go && head -8 limiter_test.go
```

<details><summary>Result</summary>

```
package delivery

import (
	"context"
	"slices"
	"testing"
	"time"
)
```

</details>

**🔧 Bash** ```bash
python3 - <<'EOF'
p='consumer_test.go'; s=open(p).read()
anchor='// Consent that the user takes back during the token wait must stop the send.'
new='''// The first claims are prepared slowly but within MaxPrep, and the later
// ones fast. Their requests must not bunch: the review found 108 requests in
// one second at 100 per second this way.
func TestConsumer_PacesRequestsAfterSlowPreparation(t *testing.T) {
	e := newConsumerEnv(t)
	const perSec = 50
	n := len(e.sms)
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
	c.MaxPrep = 200 * time.Millisecond
	var claims atomic.Int32
	c.afterClaim = func() {
		if int(claims.Add(1)) <= n/2 {
			time.Sleep(150 * time.Millisecond)
		}
	}

	if err := e.queue.Push(e.ctx, ChannelSMS, e.sms); err != nil {
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
	// k intervals must take at least k*interval. The margin allows for
	// scheduler jitter.
	for i := range at {
		for j := i + 1; j < len(at); j++ {
			if span, want := at[j].Sub(at[i]), time.Duration(j-i)*interval-interval/2; span < want {
				t.Fatalf("requests %d to %d arrived within %v, want at least %v at %d per second", i, j, span, want, perSec)
			}
		}
	}
}

'''
assert anchor in s; s=s.replace(anchor,new+anchor,1); open(p,'w').write(s)
EOF
cd ../.. && gofmt -l . ; grep -n "^test:\|^lint:" -A6 ../Makefile
```

<details><summary>Result</summary>

```
49:test:
50-	@docker network inspect cc-test >/dev/null 2>&1 || docker network create cc-test >/dev/null
51-	@docker rm -f cc-test-pg cc-test-redis >/dev/null 2>&1 || true
52-	docker run -d --name cc-test-pg --network cc-test \
53-		-e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=test \
54-		-p 127.0.0.1:55432:5432 postgres:17-alpine >/dev/null
55-	docker run -d --name cc-test-redis --network cc-test \
--
67:lint:
68-	$(GO) vet ./...
69-	@out="$$($(GOFMT) -l .)" || exit 1; \
70-		if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
71-	CI=true $(PNPM) lint
72-
73-verify: lint test
```

</details>

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one && which go; sed -n 49,66p Makefile
```

<details><summary>Result</summary>

```
/opt/homebrew/bin/go
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
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:30:56</sub>

Let me first confirm the new tests fail without the fix, using test containers.

**🔧 Bash** ```bash
grep -n "^TEST_DATABASE_URL\|^TEST_REDIS_URL" Makefile; docker rm -f cc-tpg cc-tredis >/dev/null 2>&1; docker run -d --name cc-tpg -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=test -p 127.0.0.1:55433:5432 postgres:17-alpine >/dev/null && docker run -d --name cc-tredis -p 127.0.0.1:56380:6379 redis:8-alpine >/dev/null; until docker exec cc-tpg pg_isready -U test -d test -h 127.0.0.1 >/dev/null 2>&1; do sleep 1; done; echo ready
```

<details><summary>Result</summary>

```
9:TEST_DATABASE_URL = postgres://test:test@cc-test-pg:5432/test?sslmode=disable
10:TEST_REDIS_URL = redis://cc-test-redis:6379/0
15:TEST_DATABASE_URL = postgres://test:test@127.0.0.1:55432/test?sslmode=disable
16:TEST_REDIS_URL = redis://127.0.0.1:56379/0
ready
```

</details>

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one/backend && export TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55433/test?sslmode=disable' TEST_REDIS_URL='redis://127.0.0.1:56380/0'
cp internal/delivery/consumer.go /tmp/claude-501/c.go.bak 2>/dev/null || cp internal/delivery/consumer.go "$TMPDIR/c.go.bak"
python3 - <<'EOF'
p='internal/delivery/consumer.go'; s=open(p).read()
s=s.replace('if err := c.Limiter.Admit(ctx); err != nil {','if err := ctx.Err(); err != nil {'); open(p,'w').write(s)
EOF
go test ./internal/delivery -run 'PacesRequests' -count=1 2>&1 | tail -8
git checkout -p internal/delivery/consumer.go </dev/null >/dev/null 2>&1; python3 - <<'EOF'
p='internal/delivery/consumer.go'; s=open(p).read()
s=s.replace('if err := ctx.Err(); err != nil {\n\t\tif rerr','if err := c.Limiter.Admit(ctx); err != nil {\n\t\tif rerr'); open(p,'w').write(s)
EOF
grep -n "Admit" internal/delivery/consumer.go
go test ./internal/delivery -count=1 2>&1 | tail -8
```

<details><summary>Result</summary>

```
--- FAIL: TestConsumer_PacesRequestsAfterSlowPreparation (0.33s)
    consumer_test.go:536: requests 0 to 4 arrived within 69.922ms, want at least 70ms at 50 per second
FAIL
FAIL	campaigncomposer/internal/delivery	1.330s
FAIL
118:	// can still move sends together. Admit spaces the request starts.
119:	if err := c.Limiter.Admit(ctx); err != nil {
ok  	campaigncomposer/internal/delivery	42.727s
```

</details>

**🔧 Bash** ```bash
export TEST_DATABASE_URL='postgres://test:test@127.0.0.1:55433/test?sslmode=disable' TEST_REDIS_URL='redis://127.0.0.1:56380/0'
python3 - <<'EOF'
p='internal/delivery/consumer_test.go'; s=open(p).read()
s=s.replace('''	c.MaxPrep = 200 * time.Millisecond
	var claims atomic.Int32
	c.afterClaim = func() {
		if int(claims.Add(1)) <= n/2 {
			time.Sleep(150 * time.Millisecond)''','''	// A 3-interval stall puts each slow send on the time of a fast one.
	c.MaxPrep = 5 * interval
	var claims atomic.Int32
	c.afterClaim = func() {
		if int(claims.Add(1)) <= n/2 {
			time.Sleep(3 * interval)''')
open(p,'w').write(s)
p='internal/delivery/consumer.go'; s=open(p).read()
s=s.replace('if err := c.Limiter.Admit(ctx); err != nil {\n\t\tif rerr','if err := ctx.Err(); err != nil {\n\t\tif rerr'); open(p,'w').write(s)
EOF
for i in 1 2 3; do go test ./internal/delivery -run 'PacesRequestsAfterSlow' -count=1 2>&1 | grep -E "consumer_test|^ok|FAIL\s"; done
python3 - <<'EOF'
p='internal/delivery/consumer.go'; s=open(p).read()
s=s.replace('if err := ctx.Err(); err != nil {\n\t\tif rerr','if err := c.Limiter.Admit(ctx); err != nil {\n\t\tif rerr'); open(p,'w').write(s)
EOF
for i in 1 2 3 4 5; do go test ./internal/delivery -run 'PacesRequests|Limiter' -count=1 2>&1 | tail -2; done
```

<details><summary>Result</summary>

```
consumer_test.go:537: requests 0 to 2 arrived within 19.885167ms, want at least 30ms at 50 per second
FAIL	campaigncomposer/internal/delivery	0.658s
    consumer_test.go:537: requests 0 to 2 arrived within 18.423417ms, want at least 30ms at 50 per second
FAIL	campaigncomposer/internal/delivery	0.549s
    consumer_test.go:537: requests 0 to 2 arrived within 18.44225ms, want at least 30ms at 50 per second
FAIL	campaigncomposer/internal/delivery	0.599s
ok  	campaigncomposer/internal/delivery	1.195s
ok  	campaigncomposer/internal/delivery	1.076s
ok  	campaigncomposer/internal/delivery	1.072s
ok  	campaigncomposer/internal/delivery	1.068s
ok  	campaigncomposer/internal/delivery	1.138s
```

</details>

**🔧 Bash** ```bash
cd ~/Desktop/projects/case-studies/insider_one && docker rm -f cc-tpg cc-tredis >/dev/null; grep -rl sqlc backend | head; sed -n 40,47p docs/decisions-log.md; grep -n "^## D-" docs/decisions-log.md | tail -3; grep -n "MaxPrep\|rate token\|100 ms" docs/decisions-log.md | head
```

<details><summary>Result</summary>

```
## D-006 Data access: `pgx` + `sqlc`, hand-built segment SQL

- **Context:** Most queries are fixed. Segment queries are dynamic, because the WHERE clause depends on the rule.
- **Options:** An ORM, `sqlc` for all queries, `sqlc` plus a hand-built SQL compiler.
- **Choice:** `pgx` + `sqlc` for fixed queries. `internal/segment` compiles the validated rule into SQL with `$n` placeholders.
- **Why:** `sqlc` cannot generate dynamic queries. Placeholders keep user values out of the SQL text.

## D-007 Redis Streams as the queue
620:## D-087 The read-only pages reuse the message preview
627:## D-088 The campaign page shows one language at a time
634:## D-089 Counts, dates and names use one format
462:## D-066 Duplicate queue entries use no rate token
466:- **Context:** The recovery pushes stale queued jobs again. On a long backlog, an entry waits longer than the stale age. The recovery then adds a second entry for each job. Each duplicate entry used a rate token before `Claim` rejected it. This slowed the real sends and made the backlog bigger on each recovery pass.
594:- **Context:** The consumer took a rate token, then claimed and loaded the send job from Postgres (D-053). A slow database held sends that already had a token. When the database answered, these sends left together with the newly permitted sends. The channel could then go above 100 requests per second. A test with 7 held sends showed 7 requests in less than 1 ms.
606:## D-086 A rate token is good for 100 ms
613:- **Choice:** The order is wait, claim, load, send. The consumer measures the time from the token to the end of the load. If this time is more than `Consumer.MaxPrep` (100 ms), the consumer releases the send job to `queued`. It then takes a new token and claims again. `Store.Claimable` is back (D-066).
616:  - Sends whose tokens are less than 100 ms apart can leave together. At 100 per second, this is 10 requests at most. It is equal to the burst of the mock provider.
617:  - When the claim and the l
… [99 more chars truncated]
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
sed -n 606,619p docs/decisions-log.md && tail -8 docs/decisions-log.md
```

<details><summary>Result</summary>

```
## D-086 A rate token is good for 100 ms

- **Context:** D-084 put the token wait after the claim and the load. A review found two defects in that order. The load reads the consent, so a change of consent during the wait did not stop the send. The send job was `sending` during the wait, so the recovery could set a live send job to `unknown`. The first order (D-053) had neither defect, but a slow database could release held sends together.
- **Options:**
  - Keep the D-084 order and read the consent again after the wait. This puts database work after the token again.
  - Keep the D-084 order and add a column that separates a waiting send job from a started request.
  - Restore the D-053 order and limit the age of a token.
- **Choice:** The order is wait, claim, load, send. The consumer measures the time from the token to the end of the load. If this time is more than `Consumer.MaxPrep` (100 ms), the consumer releases the send job to `queued`. It then takes a new token and claims again. `Store.Claimable` is back (D-066).
- **Why:** The consent read is the last step before the request. A send job is `queued` during a token wait, so the recovery cannot expire it. One rule covers the three defects and needs no migration.
- **Limits:**
  - Sends whose tokens are less than 100 ms apart can leave together. At 100 per second, this is 10 requests at most. It is equal to the burst of the mock provider.
  - When the claim and the load always take more than 100 ms, no message is sent. The worker logs a warning for each new token.
  - A compose run with 1,803 requests gave `rate_limited: 0`, no such warning, and 98 sends at most in one second for each channel.

- **Why:** The page length does not grow with the number of languages. The marketer sees the result of the send without a scroll. The summary uses the progress that the page already loads, so the API does not change.

## D-089 Counts, dates and names use one format

- **Context:** Pages showed `tr`, `TR` and "Turkish" for 
… [714 more chars truncated]
```

</details>

**🔧 Bash** ```bash
cat >> docs/decisions-log.md <<'EOF'

## D-090 A send gate keeps the request starts apart

- **Context:** D-086 lets a token wait up to 100 ms before its request. A review sent slow and fast preparations together. The provider got 108 requests in one second at a limit of 100.
- **Options:**
  - Make `MaxPrep` shorter. This makes the bunch smaller but does not remove it.
  - Take the token after the load. This brings back the defects that D-086 fixed.
  - Add a second gate just before the request.
- **Choice:** `Limiter.Admit` runs after the load and before `Provider.Send`. It keeps the request starts of one channel at least `1s/perSec` plus 2% apart. If the context stops during `Admit`, the consumer releases the send job to `queued`. The token wait and `MaxPrep` stay as in D-086.
- **Why:** With a minimum gap between request starts, one second cannot hold more than `perSec` requests. The 2% margin covers network jitter between the worker and the provider. The token still paces the claims, so a send job stays `queued` during a long wait.
- **Limits:**
  - The maximum rate is approximately 98 per second for each channel.
  - The consent read can be older than the request by the `Admit` wait. This wait is normally less than `MaxPrep`.
  - The gate is in memory. Two worker processes would each send at the full rate.
EOF
python3 .claude/skills/ste-writing/scripts/ste_check.py docs/decisions-log.md | grep -n "D-090\|" | tail -5; python3 .claude/skills/ste-writing/scripts/ste_check.py docs/decisions-log.md | awk -F: '$2>=640'
```

<details><summary>Result</summary>

```
(Bash completed with no output)
```

</details>

**🔧 Bash** ```bash
make verify 2>&1 | tail -25
```

<details><summary>Result</summary>

```
dist/assets/pages-CzooNFwN.js                        0.21 kB │ gzip:  0.19 kB
dist/assets/blank-B5sdXqCT.js                        0.24 kB │ gzip:  0.21 kB
dist/assets/errors-DQFC1xB4.js                       0.31 kB │ gzip:  0.24 kB
dist/assets/_...path_-qP1W3T6N.js                    0.51 kB │ gzip:  0.37 kB
dist/assets/CardTitle-Di-bGsLe.js                    0.75 kB │ gzip:  0.42 kB
dist/assets/custom-error-obCtdLlo.js                 0.83 kB │ gzip:  0.50 kB
dist/assets/_code_-BCefbyvq.js                       0.96 kB │ gzip:  0.60 kB
dist/assets/content-view-Dd3leb-p.js                 1.77 kB │ gzip:  0.79 kB
dist/assets/Primitive-BCtzcCtt.js                    1.77 kB │ gzip:  0.90 kB
dist/assets/button-DJ4hEXov.js                       2.10 kB │ gzip:  0.94 kB
dist/assets/status-badge-BCO-gDt5.js                 2.57 kB │ gzip:  0.94 kB
dist/assets/useId-DooQjZBx.js                        3.16 kB │ gzip:  1.57 kB
dist/assets/format-CN7fFAzd.js                       3.88 kB │ gzip:  1.75 kB
dist/assets/campaigns-DfSA9uQl.js                    4.12 kB │ gzip:  1.63 kB
dist/assets/new-2a2Zax4e.js                         14.64 kB │ gzip:  5.53 kB
dist/assets/_id_-DOB_jnpm.js                        23.22 kB │ gzip:  7.47 kB
dist/assets/review-dialog-DeNgBumc.js               62.98 kB │ gzip: 16.77 kB
dist/assets/runtime-core.esm-bundler-VScysjlC.js    72.52 kB │ gzip: 28.06 kB
dist/assets/DialogTitle-cfHe5FVB.js                 76.56 kB │ gzip: 26.35 kB
dist/assets/default-CrgGG0UF.js                    106.29 kB │ gzip: 26.50 kB
dist/assets/dashboard-nQThyzi7.js                  205.99 kB │ gzip: 67.17 kB
dist/assets/index-BRTuk86L.js                      250.10 kB │ gzip: 84.11 kB

✓ built in 651ms
docker compose config -q
```

</details>

**🔧 Bash** ```bash
make verify > "$TMPDIR/verify.log" 2>&1; echo rc=$?; grep -E "^(ok|FAIL|---)|Tests |Test Files" "$TMPDIR/verify.log" | head -30
```

<details><summary>Result</summary>

```
rc=0
ok  	campaigncomposer/cmd/api	(cached)
ok  	campaigncomposer/cmd/worker	(cached)
ok  	campaigncomposer/internal/campaign	(cached)
ok  	campaigncomposer/internal/config	(cached)
ok  	campaigncomposer/internal/content	(cached)
ok  	campaigncomposer/internal/dashboard	(cached)
ok  	campaigncomposer/internal/delivery	(cached)
ok  	campaigncomposer/internal/httpapi	(cached)
ok  	campaigncomposer/internal/llm	(cached)
ok  	campaigncomposer/internal/mockprovider	(cached)
ok  	campaigncomposer/internal/seed	(cached)
ok  	campaigncomposer/internal/segment	(cached)
ok  	campaigncomposer/internal/store	(cached)
 Test Files  31 passed (31)
      Tests  234 passed (234)
```

</details>

## 🤖 Claude  <sub>2026-11-11 16:34:12</sub>

I fixed the rate cap. Each channel now stays at or below 100 sends in any one-second window. `make verify` passes with all Go tests and 234 frontend tests, and the new tests fail when I take the fix out. I did not commit.

**Cause:** The consumer took a rate token, claimed the job, loaded it, then sent. The 100 ms allowed for the claim and load meant that slow jobs could land on the same moments as fast jobs that came after them. Tokens were granted at the correct rate, but the requests reached the provider in bunches.

**Fix:**
- [limiter.go](backend/internal/delivery/limiter.go) has a new `Limiter.Admit` step. It keeps the start of each request at least 1s/perSec + 2% after the one before (10.2 ms at 100/s). With that minimum gap, one second cannot hold more than 100 requests. The 2% covers network jitter on the way to the provider.
- [consumer.go:116-124](backend/internal/delivery/consumer.go#L116-L124) calls `Admit` after the job is loaded and right before `Provider.Send`. If shutdown happens during that wait, the job goes back to `queued`. The token wait and the 100 ms check stay as they were, so the earlier fixes for consent and recovery still apply.

**Tests:**
- `TestLimiter_AdmitSpacesSlowPreparation` sends 300 tokens with mixed preparation times through a fake clock. It checks that no one-second window has more than 100 sends.
- `TestLimiter_AdmitCancelled` checks that a cancelled `Admit` returns the cancel error.
- `TestConsumer_PacesRequestsAfterSlowPreparation` uses a real provider and real time, with slow claims that overlap fast ones. Without the fix it failed 3 times out of 3: 3 requests arrived in about 19 ms, where 30 ms or more is required. With the fix it passed 5 times out of 5.

**Trade-offs** (recorded as D-090 in [decisions-log.md](docs/decisions-log.md)):
- The top speed is now about 98 sends per second per channel instead of 100.
- The consent check can be a few milliseconds older than the send, because the job now waits briefly after loading.
- The limit is kept in memory, so two worker processes would each send at the full rate. This was already true before the fix.

**Not done:**
- **sqlc wording:** I did not edit `DECISIONS.md`, because CLAUDE.md says you own it. The finding is correct: nothing in `backend/` uses sqlc. D-006 in my decisions log makes the same false claim. Should I correct D-006?
- **Final checks:** I did not do the fresh install, the browser flow, or real Claude calls that the review suggests as a final check.

