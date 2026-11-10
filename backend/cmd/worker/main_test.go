package main

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"campaigncomposer/internal/delivery"
	"campaigncomposer/internal/delivery/deliverytest"
)

type fakeDB struct{ err error }

func (f fakeDB) Ping(context.Context) error { return f.err }

func get(t *testing.T, h http.Handler) (int, map[string]string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %q: %v", rec.Body.String(), err)
	}
	return rec.Code, body
}

func deadRedis(t *testing.T) *redis.Client {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	rdb := redis.NewClient(&redis.Options{Addr: addr, MaxRetries: -1})
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

func TestHealthz(t *testing.T) {
	good := deliverytest.NewRedis(t, 2)
	tests := []struct {
		name  string
		db    error
		rdb   *redis.Client
		code  int
		db_   string
		redis string
	}{
		{"all up", nil, good, 200, "ok", "ok"},
		{"redis down", nil, deadRedis(t), 503, "ok", "down"},
		{"db down", errors.New("no"), good, 503, "down", "ok"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, body := get(t, healthHandler(fakeDB{tc.db}, tc.rdb, http.NotFoundHandler()))
			if code != tc.code || body["db"] != tc.db_ || body["redis"] != tc.redis {
				t.Fatalf("got %d %v", code, body)
			}
			if want := map[int]string{200: "ok", 503: "unavailable"}[tc.code]; body["status"] != want {
				t.Fatalf("status = %q, want %q", body["status"], want)
			}
		})
	}
}

func TestMetricsRoute(t *testing.T) {
	reg := prometheus.NewRegistry()
	delivery.NewMetrics(reg, nil).Result(delivery.ChannelSMS, delivery.StatusSent)
	h := healthHandler(fakeDB{}, deliverytest.NewRedis(t, 2), promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `campaign_send_results_total{channel="sms",status="sent"} 1`) {
		t.Fatalf("GET /metrics = %d\n%s", rec.Code, rec.Body.String())
	}
}
