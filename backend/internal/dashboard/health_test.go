package dashboard

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"campaigncomposer/internal/delivery"
	"campaigncomposer/internal/delivery/deliverytest"
)

type pinger struct{ err error }

func (p pinger) Ping(context.Context) error { return p.err }

func status(code int, delay time.Duration) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			http.NotFound(w, r)
			return
		}
		select {
		case <-time.After(delay):
		case <-r.Context().Done():
			return
		}
		w.WriteHeader(code)
	}))
	return srv
}

func closedQueue() *delivery.Queue {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	_ = rdb.Close()
	return delivery.NewQueue(rdb, "health")
}

func TestChecker_AllOK(t *testing.T) {
	worker, provider := status(200, 0), status(200, 0)
	defer worker.Close()
	defer provider.Close()
	c := &Checker{
		DB: pinger{}, Redis: pinger{}, Queue: closedQueue(), HTTP: http.DefaultClient,
		WorkerURL: worker.URL, ProviderURL: provider.URL,
	}
	h := c.Check(context.Background())
	for _, k := range []string{"db", "redis", "worker", "provider"} {
		if h.Services[k] != "ok" {
			t.Errorf("services[%s] = %q, want ok", k, h.Services[k])
		}
	}
}

func TestChecker_WorkerDown(t *testing.T) {
	worker, provider := status(500, 0), status(200, 3*time.Second)
	defer worker.Close()
	defer provider.Close()
	c := &Checker{
		DB: pinger{}, Redis: pinger{}, Queue: closedQueue(), HTTP: http.DefaultClient,
		WorkerURL: worker.URL, ProviderURL: provider.URL, Timeout: 500 * time.Millisecond,
	}
	start := time.Now()
	h := c.Check(context.Background())
	if d := time.Since(start); d > 2*time.Second {
		t.Errorf("Check took %v, want under 2s", d)
	}
	if h.Services["worker"] != "down" || h.Services["provider"] != "down" {
		t.Errorf("services = %v, want worker and provider down", h.Services)
	}
	if h.Services["db"] != "ok" {
		t.Errorf("services[db] = %q, want ok", h.Services["db"])
	}
}

func TestChecker_RedisDown(t *testing.T) {
	ctx, pool, s := newDB(t)
	cid := insertCampaign(t, ctx, pool, "queued", "health")
	addJobs(t, ctx, pool, cid, delivery.ChannelSMS, delivery.StatusQueued, 2)
	addJobs(t, ctx, pool, cid, delivery.ChannelEmail, delivery.StatusSent, 1)
	c := &Checker{
		DB: pinger{}, Redis: pinger{err: errors.New("down")}, Queue: closedQueue(), Store: s,
		HTTP: http.DefaultClient, WorkerURL: "http://127.0.0.1:1", ProviderURL: "http://127.0.0.1:1",
		Timeout: 500 * time.Millisecond,
	}
	h := c.Check(ctx)
	if h.Services["redis"] != "down" {
		t.Errorf("services[redis] = %q, want down", h.Services["redis"])
	}
	if len(h.Streams) != 2 || h.Streams[0].Channel != delivery.ChannelSMS || h.Streams[1].Channel != delivery.ChannelEmail {
		t.Fatalf("streams = %+v, want sms then email", h.Streams)
	}
	for _, sd := range h.Streams {
		if sd.Length != nil || sd.Pending != nil {
			t.Errorf("%s depth = %v, %v; want nil, nil", sd.Channel, sd.Length, sd.Pending)
		}
	}
	if h.Streams[0].InProgress != 2 || h.Streams[1].InProgress != 0 {
		t.Errorf("in progress = %d, %d; want 2, 0", h.Streams[0].InProgress, h.Streams[1].InProgress)
	}
}

func TestChecker_Streams(t *testing.T) {
	ctx, _, s := newDB(t)
	rdb := deliverytest.NewRedis(t, 3)
	q := delivery.NewQueue(rdb, "c1")
	if err := q.Ensure(ctx); err != nil {
		t.Fatal(err)
	}
	if err := q.Push(ctx, delivery.ChannelSMS, []int64{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if got, err := q.Read(ctx, delivery.ChannelSMS, 2, 100*time.Millisecond); err != nil || len(got) != 2 {
		t.Fatalf("Read = %v, %v", got, err)
	}
	c := &Checker{
		DB: pinger{}, Redis: RedisPinger{Client: rdb}, Queue: q, Store: s, HTTP: http.DefaultClient,
		WorkerURL: "http://127.0.0.1:1", ProviderURL: "http://127.0.0.1:1", Timeout: 500 * time.Millisecond,
	}
	h := c.Check(ctx)
	if h.Services["redis"] != "ok" {
		t.Errorf("services[redis] = %q, want ok", h.Services["redis"])
	}
	sms, email := h.Streams[0], h.Streams[1]
	if sms.Channel != delivery.ChannelSMS || sms.Length == nil || *sms.Length != 3 || sms.Pending == nil || *sms.Pending != 2 {
		t.Errorf("sms = %+v, want length 3, pending 2", sms)
	}
	if email.Length == nil || *email.Length != 0 || email.Pending == nil || *email.Pending != 0 {
		t.Errorf("email = %+v, want length 0, pending 0", email)
	}
}

// silentRedis listens, accepts connections and never writes.
func silentRedis(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var conns []net.Conn
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			mu.Lock()
			conns = append(conns, c)
			mu.Unlock()
		}
	}()
	t.Cleanup(func() {
		_ = ln.Close()
		mu.Lock()
		defer mu.Unlock()
		for _, c := range conns {
			_ = c.Close()
		}
	})
	return ln.Addr().String()
}

func slowRedisChecker(t *testing.T, timeout time.Duration, ctxTimeout bool) *Checker {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: silentRedis(t), ContextTimeoutEnabled: ctxTimeout})
	t.Cleanup(func() { _ = rdb.Close() })
	return &Checker{
		DB: pinger{}, Redis: RedisPinger{Client: rdb}, Queue: delivery.NewQueue(rdb, "slow"),
		Timeout: timeout,
	}
}

func TestChecker_SlowRedisIsBoundedByTimeout(t *testing.T) {
	c := slowRedisChecker(t, 300*time.Millisecond, true)
	start := time.Now()
	h := c.Check(context.Background())
	if d := time.Since(start); d > 2*time.Second {
		t.Errorf("Check took %v, want under 2s", d)
	}
	if h.Services["redis"] != "down" {
		t.Errorf("redis = %q, want down", h.Services["redis"])
	}
	for _, s := range h.Streams {
		if s.Length != nil || s.Pending != nil {
			t.Errorf("stream %s: length/pending = %v/%v, want nil", s.Channel, s.Length, s.Pending)
		}
	}
}

func TestChecker_ZeroTimeoutMeansOneSecond(t *testing.T) {
	c := slowRedisChecker(t, 0, true)
	start := time.Now()
	h := c.Check(context.Background())
	d := time.Since(start)
	if d < 900*time.Millisecond || d > 2*time.Second {
		t.Errorf("Check took %v, want about 1s", d)
	}
	if h.Services["redis"] != "down" {
		t.Errorf("redis = %q, want down", h.Services["redis"])
	}
}
