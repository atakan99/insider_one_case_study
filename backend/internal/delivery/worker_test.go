package delivery

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/config"
	"campaigncomposer/internal/delivery/deliverytest"
	"campaigncomposer/internal/mockprovider"
	"campaigncomposer/internal/segment"
)

type workerEnv struct {
	ctx   context.Context
	pool  *pgxpool.Pool
	store *Store
	queue *Queue
	srv   *httptest.Server
	cid   int64
}

// newWorkerEnv seeds users, approves a campaign and starts a provider.
func newWorkerEnv(t *testing.T, users int, rule string, wrap func(http.Handler) http.Handler, opts mockprovider.Options) *workerEnv {
	t.Helper()
	ctx, pool, s := seededN(t, users)
	cid := approved(t, ctx, pool, rule, "tr", "en")
	var h http.Handler = mockprovider.New(opts)
	if wrap != nil {
		h = wrap(h)
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &workerEnv{ctx: ctx, pool: pool, store: s, queue: NewQueue(deliverytest.NewRedis(t, 0), "w1"), srv: srv, cid: cid}
}

func cleanOpts() mockprovider.Options {
	return mockprovider.Options{RatePerSec: 10000, Burst: 10000, SlowDelay: 200 * time.Millisecond, Random: 1}
}

func (e *workerEnv) worker(rate, consumers int, timeout time.Duration) *Worker {
	return &Worker{
		Store: e.store, Queue: e.queue, Provider: NewProvider(e.srv.URL, timeout),
		Allow: segment.Default(), Clock: RealClock(),
		Cfg:       config.WorkerConfig{RatePerSec: rate, MaxAttempts: 5, Consumers: consumers, Timeout: timeout},
		Log:       slog.New(slog.NewTextHandler(&syncBuf{}, nil)),
		PlanEvery: 50 * time.Millisecond, RetryEvery: 50 * time.Millisecond,
		RecoverEvery: 100 * time.Millisecond, ClaimIdle: time.Minute, StaleAfter: time.Minute,
	}
}

// run starts w.Run and returns a stop function that cancels it and returns its error.
func (e *workerEnv) run(t *testing.T, w *Worker) (stop func() error) {
	t.Helper()
	ctx, cancel := context.WithCancel(e.ctx)
	errc := make(chan error, 1)
	go func() { errc <- w.Run(ctx) }()
	var once sync.Once
	var err error
	stop = func() error {
		once.Do(func() {
			cancel()
			select {
			case err = <-errc:
			case <-time.After(20 * time.Second):
				t.Error("Run did not return within 20s of cancel")
			}
		})
		return err
	}
	t.Cleanup(func() { _ = stop() })
	return stop
}

func (e *workerEnv) campaignDone(t *testing.T, limit time.Duration) {
	t.Helper()
	deadline := time.Now().Add(limit)
	for time.Now().Before(deadline) {
		if campaignStatus(t, e.ctx, e.pool, e.cid) == "completed" {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	p, _ := e.store.Progress(e.ctx, e.cid)
	t.Fatalf("campaign not completed within %v: %+v", limit, p)
}

func (e *workerEnv) countStatus(t *testing.T, status ...string) int {
	t.Helper()
	var n int
	if err := e.pool.QueryRow(e.ctx, `SELECT count(*) FROM send_jobs WHERE campaign_id = $1 AND status = ANY($2)`, e.cid, status).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

type jobRef struct {
	id        int64
	status    Status
	reference string
}

func (e *workerEnv) jobs(t *testing.T) []jobRef {
	t.Helper()
	rs, err := e.pool.Query(e.ctx, `SELECT id, status, campaign_id || ':' || user_id || ':' || channel FROM send_jobs WHERE campaign_id = $1`, e.cid)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()
	var out []jobRef
	for rs.Next() {
		var j jobRef
		if err := rs.Scan(&j.id, &j.status, &j.reference); err != nil {
			t.Fatal(err)
		}
		out = append(out, j)
	}
	return out
}

func (e *workerEnv) messageCount(t *testing.T, reference string) int {
	t.Helper()
	resp, err := http.Get(e.srv.URL + "/v1/messages?reference=" + reference)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out struct {
		Total int `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out.Total
}

// checkDelivery asserts that all jobs are terminal, no message is a duplicate,
// each sent job has one message and each failed job has none.
func (e *workerEnv) checkDelivery(t *testing.T, wantSent bool) {
	t.Helper()
	js := e.jobs(t)
	if len(js) == 0 {
		t.Fatal("no send jobs")
	}
	sent := 0
	for _, j := range js {
		switch j.status {
		case StatusSent:
			sent++
			if n := e.messageCount(t, j.reference); n != 1 {
				t.Fatalf("sent job %d (%s) has %d provider messages, want 1", j.id, j.reference, n)
			}
		case StatusFailed:
			if n := e.messageCount(t, j.reference); n != 0 {
				t.Fatalf("failed job %d (%s) has %d provider messages, want 0", j.id, j.reference, n)
			}
		case StatusUnknown, StatusSkipped:
		default:
			t.Fatalf("job %d is %s, want a terminal status", j.id, j.status)
		}
	}
	if wantSent && sent != len(js) {
		t.Fatalf("%d of %d jobs sent", sent, len(js))
	}
	for ch, st := range providerStats(t, e.srv.URL) {
		if st.DuplicateDeliveries != 0 {
			t.Fatalf("%s duplicate deliveries = %d", ch, st.DuplicateDeliveries)
		}
	}
}

// checkLogsNoPersonalData asserts that the logs hold no phone, no email, no
// first name (as a whole word) and no message text of the seeded users.
func (e *workerEnv) checkLogsNoPersonalData(t *testing.T, out string) {
	t.Helper()
	if !strings.Contains(out, "job_id") && !strings.Contains(out, "campaign") {
		t.Fatalf("no log lines:\n%.2000s", out)
	}
	words := map[string]bool{}
	for _, w := range regexp.MustCompile(`[\p{L}\p{N}_]+`).FindAllString(out, -1) {
		words[w] = true
	}
	rs, err := e.pool.Query(e.ctx, `SELECT first_name, phone, email FROM users`)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()
	n := 0
	for rs.Next() {
		var name, phone, email string
		if err := rs.Scan(&name, &phone, &email); err != nil {
			t.Fatal(err)
		}
		n++
		for _, secret := range []string{phone, email, "Hi " + name, "Subject " + name, "Body " + name} {
			if secret != "" && strings.Contains(out, secret) {
				t.Fatalf("log contains %q", secret)
			}
		}
		if name != "" && words[name] {
			t.Fatalf("log contains the first name %q", name)
		}
	}
	if err := rs.Err(); err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("no users to check")
	}
}

func TestWorker_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("long")
	}
	opts := mockprovider.Options{RatePerSec: 100, Burst: 10, ErrorRate: 0.03, SlowRate: 0.03, LostRate: 0.01,
		SlowDelay: 200 * time.Millisecond, LatencyMax: 5 * time.Millisecond, Random: 7}
	e := newWorkerEnv(t, 3000, trUSRule, nil, opts)
	w := e.worker(100, 32, 50*time.Millisecond)
	w.RecoverEvery = 200 * time.Millisecond
	logs := &syncBuf{}
	w.Log = slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	start := time.Now()
	stop := e.run(t, w)
	e.campaignDone(t, 90*time.Second)
	if err := stop(); err != nil {
		t.Fatalf("Run = %v", err)
	}
	p, err := e.store.Progress(e.ctx, e.cid)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("end-to-end took %v, progress %+v, stats %+v", time.Since(start), p.Channels, providerStats(t, e.srv.URL))
	e.checkDelivery(t, false)
	e.checkLogsNoPersonalData(t, logs.String())
	if p.Channels[ChannelSMS].Sent == 0 || p.Channels[ChannelEmail].Sent == 0 {
		t.Fatalf("no sent jobs: %+v", p.Channels)
	}
}

func TestWorker_RecoverRepushGatedByLag(t *testing.T) {
	e := newWorkerEnv(t, 200, exampleRule, nil, cleanOpts())
	if err := e.queue.Ensure(e.ctx); err != nil {
		t.Fatal(err)
	}
	if _, _, ok, err := e.store.PlanNext(e.ctx, segment.Default(), time.Now()); err != nil || !ok {
		t.Fatalf("PlanNext = %v, %v", ok, err)
	}
	// Unread entries on both streams: the queued jobs are not pushed again.
	if err := e.queue.Push(e.ctx, ChannelSMS, []int64{999999}); err != nil {
		t.Fatal(err)
	}
	if err := e.queue.Push(e.ctx, ChannelEmail, []int64{999999}); err != nil {
		t.Fatal(err)
	}
	w := e.worker(100, 1, time.Second)
	w.StaleAfter = time.Nanosecond
	w.recover(e.ctx)
	for _, ch := range Channels {
		if n, _ := e.queue.rdb.XLen(e.ctx, streamKey(ch)).Result(); n != 1 {
			t.Fatalf("%s length = %d with unread entries, want 1", ch, n)
		}
	}
	// Read and acknowledge everything: lag is 0, so the jobs are pushed again.
	for _, ch := range Channels {
		es, err := e.queue.Read(e.ctx, ch, 10, 100*time.Millisecond)
		if err != nil || len(es) != 1 {
			t.Fatalf("Read = %v, %v", es, err)
		}
		if err := e.queue.Ack(e.ctx, ch, es[0].StreamID); err != nil {
			t.Fatal(err)
		}
	}
	w.recover(e.ctx)
	for _, ch := range Channels {
		if n, _ := e.queue.rdb.XLen(e.ctx, streamKey(ch)).Result(); n <= 1 {
			t.Fatalf("%s length = %d with an empty group, want the queued jobs pushed", ch, n)
		}
	}
}

func TestWorker_RestartBeforePush(t *testing.T) {
	e := newWorkerEnv(t, 500, exampleRule, nil, cleanOpts())
	_, ids, ok, err := e.store.PlanNext(e.ctx, segment.Default(), time.Now())
	if err != nil || !ok || len(ids[ChannelSMS])+len(ids[ChannelEmail]) == 0 {
		t.Fatalf("PlanNext = %v, %v, %v", ids, ok, err)
	}
	// The planner "crashed" here: the jobs are queued and nothing is on the stream.
	stop := e.run(t, e.worker(100, 8, time.Second))
	e.campaignDone(t, 60*time.Second)
	if err := stop(); err != nil {
		t.Fatal(err)
	}
	e.checkDelivery(t, true)
}

func TestWorker_StreamLost(t *testing.T) {
	e := newWorkerEnv(t, 500, trUSRule, nil, cleanOpts())
	w := e.worker(100, 8, time.Second)
	w.StaleAfter = 300 * time.Millisecond
	stop := e.run(t, w)
	deadline := time.Now().Add(30 * time.Second)
	for e.countStatus(t, "sent") < 10 {
		if time.Now().After(deadline) {
			t.Fatal("no progress")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err := e.queue.rdb.Del(e.ctx, "send:sms", "send:email").Err(); err != nil {
		t.Fatal(err)
	}
	if e.countStatus(t, "queued") == 0 {
		t.Fatal("run ended before the streams were deleted; test is not valid")
	}
	e.campaignDone(t, 60*time.Second)
	if err := stop(); err != nil {
		t.Fatal(err)
	}
	e.checkDelivery(t, true)
}

func TestWorker_AutoClaimDeadConsumer(t *testing.T) {
	e := newWorkerEnv(t, 500, exampleRule, nil, cleanOpts())
	_, ids, ok, err := e.store.PlanNext(e.ctx, segment.Default(), time.Now())
	if err != nil || !ok {
		t.Fatalf("PlanNext = %v, %v", ok, err)
	}
	dead := NewQueue(e.queue.rdb, "dead")
	if err := dead.Ensure(e.ctx); err != nil {
		t.Fatal(err)
	}
	victim := ids[ChannelSMS][0]
	if err := dead.Push(e.ctx, ChannelSMS, []int64{victim}); err != nil {
		t.Fatal(err)
	}
	// The recovery must not push it again: only the autoclaim can send it.
	if err := e.store.MarkEnqueued(e.ctx, []int64{victim}, time.Now()); err != nil {
		t.Fatal(err)
	}
	es, err := dead.Read(e.ctx, ChannelSMS, 1, time.Second)
	if err != nil || len(es) != 1 {
		t.Fatalf("dead Read = %v, %v", es, err)
	}
	w := e.worker(100, 4, time.Second)
	w.ClaimIdle = 100 * time.Millisecond
	stop := e.run(t, w)
	e.campaignDone(t, 30*time.Second)
	if err := stop(); err != nil {
		t.Fatal(err)
	}
	if j := getJob(t, e.ctx, e.pool, victim); j.Status != StatusSent {
		t.Fatalf("victim job = %s", j.Status)
	}
	e.checkDelivery(t, true)
}

func TestWorker_ChannelsIndependent(t *testing.T) {
	var mu sync.Mutex
	var emails []time.Time
	wrap := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/v1/sms":
				time.Sleep(300 * time.Millisecond)
			case "/v1/email":
				mu.Lock()
				emails = append(emails, time.Now())
				mu.Unlock()
			}
			next.ServeHTTP(w, r)
		})
	}
	e := newWorkerEnv(t, 500, trUSRule, wrap, cleanOpts())
	w := e.worker(20, 4, 3*time.Second)
	e.run(t, w)
	var first time.Time
	deadline := time.Now().Add(10 * time.Second)
	for first.IsZero() {
		mu.Lock()
		if len(emails) > 0 {
			first = emails[0]
		}
		mu.Unlock()
		if time.Now().After(deadline) {
			t.Fatal("no email attempt")
		}
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(time.Until(first.Add(time.Second)))
	mu.Lock()
	n := 0
	for _, at := range emails {
		if !at.After(first.Add(time.Second)) {
			n++
		}
	}
	mu.Unlock()
	if n < 15 {
		t.Fatalf("%d email attempts in 1 s, want at least 15", n)
	}
}

func TestWorker_RetryPollerOnTime(t *testing.T) {
	e := newWorkerEnv(t, 500, exampleRule, nil, cleanOpts())
	_, ids, ok, err := e.store.PlanNext(e.ctx, segment.Default(), time.Now())
	if err != nil || !ok {
		t.Fatalf("PlanNext = %v, %v", ok, err)
	}
	exec(t, e.ctx, e.pool, `UPDATE send_jobs SET status = 'sent' WHERE campaign_id = $1`, e.cid)
	id := ids[ChannelSMS][0]
	due := time.Now().Add(300 * time.Millisecond)
	exec(t, e.ctx, e.pool, `UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1`, id, due)
	w := e.worker(100, 2, time.Second)
	w.RetryEvery = 100 * time.Millisecond
	w.RecoverEvery = time.Hour // only the retry poller can push it
	e.run(t, w)
	for time.Now().Before(due.Add(time.Second)) {
		if getJob(t, e.ctx, e.pool, id).Status == StatusSent {
			var sentAt time.Time
			if err := e.pool.QueryRow(e.ctx, `SELECT sent_at FROM send_jobs WHERE id = $1`, id).Scan(&sentAt); err != nil {
				t.Fatal(err)
			}
			if sentAt.Before(due) {
				t.Fatalf("sent at %v, before it was due at %v", sentAt, due)
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("job still %s at due+1s", getJob(t, e.ctx, e.pool, id).Status)
}

// A retry waits for its turn on a busy stream: the poller must not push the same job again at each poll.
func TestWorker_RetryPollerNoRepeat(t *testing.T) {
	e := newWorkerEnv(t, 500, exampleRule, nil, cleanOpts())
	_, ids, _, err := e.store.PlanNext(e.ctx, segment.Default(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	exec(t, e.ctx, e.pool, `UPDATE send_jobs SET status = 'sent' WHERE campaign_id = $1`, e.cid)
	id := ids[ChannelSMS][0]
	exec(t, e.ctx, e.pool, `UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = now() WHERE id = $1`, id)
	// A queue with no consumer: the entry stays on the stream.
	w := e.worker(100, 1, time.Second)
	w.RetryEvery = 20 * time.Millisecond
	w.RecoverEvery = time.Hour
	w.Cfg.Consumers = 0
	e.run(t, w)
	time.Sleep(500 * time.Millisecond)
	n, err := e.queue.rdb.XLen(e.ctx, "send:sms").Result()
	if err != nil || n != 1 {
		t.Fatalf("stream length = %d, %v; want 1", n, err)
	}
}
