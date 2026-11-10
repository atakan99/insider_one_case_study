package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func metricsMux() *http.ServeMux {
	mux := http.NewServeMux()
	ok := func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) }
	mux.HandleFunc("GET /api/campaigns/{id}", ok)
	mux.HandleFunc("GET /api/campaigns/{id}/progress/stream", func(w http.ResponseWriter, _ *http.Request) {
		w.(http.Flusher).Flush()
		w.Write([]byte("data: x\n\n"))
	})
	return mux
}

func TestMetrics_Route(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewHTTPMetrics(reg)
	h := m.Wrap(metricsMux())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/campaigns/1", nil))

	c := m.requests.WithLabelValues("GET /api/campaigns/{id}", "GET", "200")
	if got := testutil.ToFloat64(c); got != 1 {
		t.Errorf("counter = %v, want 1", got)
	}
	if n := testutil.CollectAndCount(m.requests); n != 1 {
		t.Errorf("counter series = %d, want 1", n)
	}
	if n := testutil.CollectAndCount(m.duration); n != 1 {
		t.Errorf("histogram series = %d, want 1", n)
	}
}

func TestMetrics_UnmatchedRoute(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewHTTPMetrics(reg)
	h := m.Wrap(metricsMux())
	for i := 0; i < 50; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope/"+strconv.Itoa(i), nil))
	}
	if n := testutil.CollectAndCount(m.requests); n != 1 {
		t.Fatalf("counter series = %d, want 1", n)
	}
	if got := testutil.ToFloat64(m.requests.WithLabelValues("unmatched", "GET", "404")); got != 50 {
		t.Errorf("unmatched = %v, want 50", got)
	}
}

func TestMetrics_StreamNotInHistogram(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewHTTPMetrics(reg)
	h := m.Wrap(metricsMux())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/campaigns/1/progress/stream", nil))
	if got := testutil.ToFloat64(m.requests.WithLabelValues("GET /api/campaigns/{id}/progress/stream", "GET", "200")); got != 1 {
		t.Errorf("counter = %v, want 1", got)
	}
	if n := testutil.CollectAndCount(m.duration); n != 0 {
		t.Errorf("histogram series = %d, want 0", n)
	}
}

func TestMetrics_KeepsFlusher(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := NewHTTPMetrics(reg).Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, ok := w.(http.Flusher); !ok {
			t.Error("wrapped writer is not an http.Flusher")
		}
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Errorf("ResponseController.Flush: %v", err)
		}
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))
}
