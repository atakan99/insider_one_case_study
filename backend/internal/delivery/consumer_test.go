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

func (e *consumerEnv) pending(t *testing.T) int64 {
	t.Helper()
	p, err := e.queue.rdb.XPending(e.ctx, streamKey(ChannelSMS), groupName).Result()
	if err != nil {
		t.Fatal(err)
	}
	return p.Count
}

func (e *consumerEnv) nextAttempt(t *testing.T, id int64) time.Time {
	t.Helper()
	var at time.Time
	if err := e.pool.QueryRow(e.ctx, `SELECT next_attempt_at FROM send_jobs WHERE id = $1`, id).Scan(&at); err != nil {
		t.Fatal(err)
	}
	return at
}

func TestConsumer_Sequences(t *testing.T) {
	e := newConsumerEnv(t)
	type step struct {
		modes  string // X-Mock-Fail for this call. "refuse" uses a closed port.
		status Status
		att    int
		errHas string
		retry  bool
	}
	cases := []struct {
		name  string
		steps []step
	}{
		{"429 then 202", []step{{"429", StatusRetrying, 1, "429", false}, {"none", StatusSent, 2, "429", false}}},
		{"500 then 202", []step{{"500", StatusRetrying, 1, "500", false}, {"none", StatusSent, 2, "500", false}}},
		{"503 five times", []step{{"503", StatusRetrying, 1, "503", false}, {"503", StatusRetrying, 2, "503", false},
			{"503", StatusRetrying, 3, "503", false}, {"503", StatusRetrying, 4, "503", false}, {"503", StatusFailed, 5, "503", true}}},
		{"400", []step{{"bogus", StatusFailed, 1, "400", false}}},
		{"slow", []step{{"slow", StatusUnknown, 1, "timeout", false}}},
		{"lost", []step{{"lost", StatusUnknown, 1, "connection lost", false}}},
		{"refused then 202", []step{{"refuse", StatusRetrying, 1, "connection refused", false}, {"none", StatusSent, 2, "connection refused", false}}},
	}
	dead := deadURL(t)
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := e.sms[i]
			srv := newMock(t)
			p, ft := providerWith(srv.URL, 50*time.Millisecond)
			c := e.consumer(p)
			for n, st := range tc.steps {
				p.BaseURL = srv.URL
				if st.modes == "refuse" {
					p.BaseURL = dead
				} else {
					ft.set(st.modes)
				}
				if n > 0 {
					e.clk.Set(e.nextAttempt(t, id)) // the retry is due
				}
				before := e.clk.Now()
				es := e.entries(t, id, 1)
				if err := c.Handle(e.ctx, es[0]); err != nil {
					t.Fatalf("step %d: Handle = %v", n, err)
				}
				if st.modes == "429" {
					if at := e.nextAttempt(t, id); !at.Equal(before.Add(time.Second)) {
						t.Fatalf("next_attempt_at = %v, want %v", at, before.Add(time.Second))
					}
					if !c.Limiter.paused() {
						t.Fatal("limiter is not paused after a 429")
					}
				}
				j := getJob(t, e.ctx, e.pool, id)
				if j.Status != st.status || j.Attempts != st.att || j.Retryable != st.retry || !strings.Contains(j.LastError, st.errHas) {
					t.Fatalf("step %d: job = %s attempts %d retryable %v error %q; want %s %d %v %q",
						n, j.Status, j.Attempts, j.Retryable, j.LastError, st.status, st.att, st.retry, st.errHas)
				}
				if st.status == StatusRetrying && e.nextAttempt(t, id).IsZero() {
					t.Fatalf("step %d: no next_attempt_at", n)
				}
			}
			s := providerStats(t, srv.URL)["sms"]
			if s.DuplicateDeliveries != 0 {
				t.Fatalf("duplicate deliveries: %d", s.DuplicateDeliveries)
			}
			if last := tc.steps[len(tc.steps)-1]; last.status == StatusSent && s.Delivered != 1 {
				t.Fatalf("delivered = %d, want 1", s.Delivered)
			}
		})
	}
	if n := e.pending(t); n != 0 {
		t.Fatalf("%d entries left unacknowledged", n)
	}
}

func TestConsumer_DoubleEntry(t *testing.T) {
	e := newConsumerEnv(t)
	srv := newMock(t)
	c := e.consumer(NewProvider(srv.URL, 2*time.Second))
	es := e.entries(t, e.sms[0], 2)
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range es {
		wg.Add(1)
		go func() { defer wg.Done(); errs[i] = c.Handle(e.ctx, es[i]) }()
	}
	wg.Wait()
	if errs[0] != nil || errs[1] != nil {
		t.Fatalf("Handle = %v, %v", errs[0], errs[1])
	}
	if got := providerStats(t, srv.URL)["sms"].Requests; got != 1 {
		t.Fatalf("requests = %d, want 1", got)
	}
	if j := getJob(t, e.ctx, e.pool, e.sms[0]); j.Status != StatusSent || j.Attempts != 1 {
		t.Fatalf("job = %s attempts %d", j.Status, j.Attempts)
	}
	if n := e.pending(t); n != 0 {
		t.Fatalf("%d entries left unacknowledged", n)
	}
}

func TestConsumer_NoContact(t *testing.T) {
	e := newConsumerEnv(t)
	srv := newMock(t)
	c := e.consumer(NewProvider(srv.URL, 2*time.Second))
	exec(t, e.ctx, e.pool, `UPDATE users SET phone = '' WHERE id = (SELECT user_id FROM send_jobs WHERE id = $1)`, e.sms[0])
	es := e.entries(t, e.sms[0], 1)
	if err := c.Handle(e.ctx, es[0]); err != nil {
		t.Fatal(err)
	}
	j := getJob(t, e.ctx, e.pool, e.sms[0])
	if j.Status != StatusSkipped || j.LastError != SkipNoContact {
		t.Fatalf("job = %s %q", j.Status, j.LastError)
	}
	if got := providerStats(t, srv.URL)["sms"].Requests; got != 0 {
		t.Fatalf("requests = %d, want 0", got)
	}
	if n := e.pending(t); n != 0 {
		t.Fatalf("%d entries left unacknowledged", n)
	}
}

func TestConsumer_ReleaseOnCancel(t *testing.T) {
	e := newConsumerEnv(t)
	srv := newMock(t)
	c := e.consumer(NewProvider(srv.URL, 2*time.Second))
	ctx, cancel := context.WithCancel(e.ctx)
	defer cancel()
	c.afterClaim = cancel
	es := e.entries(t, e.sms[0], 1)
	if err := c.Handle(ctx, es[0]); !errors.Is(err, context.Canceled) {
		t.Fatalf("Handle = %v, want context.Canceled", err)
	}
	if j := getJob(t, e.ctx, e.pool, e.sms[0]); j.Status != StatusQueued || j.Attempts != 0 {
		t.Fatalf("job = %s attempts %d, want queued 0", j.Status, j.Attempts)
	}
	if got := providerStats(t, srv.URL)["sms"].Requests; got != 0 {
		t.Fatalf("requests = %d", got)
	}
	if n := e.pending(t); n != 1 {
		t.Fatalf("pending = %d, want the entry to stay", n)
	}
}

func TestConsumer_CancelInWait(t *testing.T) {
	e := newConsumerEnv(t)
	srv := newMock(t)
	c := e.consumer(NewProvider(srv.URL, 2*time.Second))
	c.Limiter.Pause(time.Hour)
	ctx, cancel := context.WithCancel(e.ctx)
	defer cancel()
	e.clk.onSleep = cancel
	es := e.entries(t, e.sms[0], 1)
	if err := c.Handle(ctx, es[0]); !errors.Is(err, context.Canceled) {
		t.Fatalf("Handle = %v, want context.Canceled", err)
	}
	if j := getJob(t, e.ctx, e.pool, e.sms[0]); j.Status != StatusQueued || j.Attempts != 0 {
		t.Fatalf("job = %s attempts %d, want queued 0", j.Status, j.Attempts)
	}
}

func TestConsumer_FinishFailsNoResend(t *testing.T) {
	e := newConsumerEnv(t)
	srv := newMock(t)
	c := e.consumer(NewProvider(srv.URL, 2*time.Second))
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
	if n := e.pending(t); n != 1 {
		t.Fatalf("pending = %d, want 1 (not acknowledged)", n)
	}
	claimed, err := e.queue.AutoClaim(e.ctx, ChannelSMS, 0, 10)
	if err != nil || len(claimed) != 1 {
		t.Fatalf("AutoClaim = %v, %v", claimed, err)
	}
	if err := c.Handle(e.ctx, claimed[0]); err != nil {
		t.Fatal(err)
	}
	if got := providerStats(t, srv.URL)["sms"].Requests; got != 1 {
		t.Fatalf("requests = %d, want 1", got)
	}
	if n := e.pending(t); n != 0 {
		t.Fatalf("pending = %d, want 0", n)
	}
}

func TestConsumer_NotClaimableSpendsNoToken(t *testing.T) {
	e := newConsumerEnv(t)
	srv := newMock(t)
	if _, err := e.pool.Exec(e.ctx, `UPDATE send_jobs SET status = 'sent' WHERE id = $1`, e.sms[0]); err != nil {
		t.Fatal(err)
	}
	c := e.consumer(NewProvider(srv.URL, 2*time.Second))
	c.Limiter = NewLimiter(1, e.clk)
	es := e.entries(t, e.sms[0], 100)
	for _, en := range es {
		if err := c.Handle(e.ctx, en); err != nil {
			t.Fatal(err)
		}
	}
	if d := e.clk.Now().Sub(testNow); d > 5*time.Second {
		t.Fatalf("fake clock advanced %v: tokens were spent on entries that cannot be claimed", d)
	}
	if got := providerStats(t, srv.URL)["sms"].Requests; got != 0 {
		t.Fatalf("requests = %d, want 0", got)
	}
	if n := e.pending(t); n != 0 {
		t.Fatalf("%d entries left unacknowledged", n)
	}
}

func TestConsumer_LogsNoPersonalData(t *testing.T) {
	e := newConsumerEnv(t)
	srv := newMock(t)
	p, ft := providerWith(srv.URL, 50*time.Millisecond)
	c := e.consumer(p)
	modes := []string{"503", "429", "bogus", "slow", "lost"}
	ids := e.sms[:len(modes)]
	for i, id := range ids {
		ft.set(modes[i])
		es := e.entries(t, id, 1)
		if err := c.Handle(e.ctx, es[0]); err != nil {
			t.Fatal(err)
		}
	}
	out := e.logs.String()
	if !strings.Contains(out, "job_id") {
		t.Fatalf("no log lines:\n%s", out)
	}
	for _, id := range ids {
		var name, phone, email string
		if err := e.pool.QueryRow(e.ctx, `SELECT u.first_name, u.phone, u.email FROM users u JOIN send_jobs j ON j.user_id = u.id WHERE j.id = $1`, id).
			Scan(&name, &phone, &email); err != nil {
			t.Fatal(err)
		}
		for _, secret := range []string{phone, email, "Hi " + name} {
			if secret != "" && strings.Contains(out, secret) {
				t.Fatalf("log contains %q:\n%s", secret, out)
			}
		}
		if name != "" && regexp.MustCompile(`\b`+regexp.QuoteMeta(name)+`\b`).MatchString(out) {
			t.Fatalf("log contains the first name %q:\n%s", name, out)
		}
	}
}

// cancelTransport cancels a context when the request starts.
type cancelTransport struct {
	base   http.RoundTripper
	cancel context.CancelFunc
}

func (c cancelTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	c.cancel()
	return c.base.RoundTrip(r)
}

func TestConsumer_CancelDuringSend(t *testing.T) {
	e := newConsumerEnv(t)
	srv := newMock(t)
	p, _ := providerWith(srv.URL, 50*time.Millisecond, "slow")
	ctx, cancel := context.WithCancel(e.ctx)
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

// The first claims are prepared slowly but within MaxPrep, and the later
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
	// A 3-interval stall puts each slow send on the time of a fast one.
	c.MaxPrep = 5 * interval
	var claims atomic.Int32
	c.afterClaim = func() {
		if int(claims.Add(1)) <= n/2 {
			time.Sleep(3 * interval)
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
