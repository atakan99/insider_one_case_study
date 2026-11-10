package httpapi

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
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
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, mf := range mfs {
		if mf.GetName() != "campaign_http_request_duration_seconds" {
			continue
		}
		got := map[string]string{}
		for _, l := range mf.Metric[0].Label {
			got[l.GetName()] = l.GetValue()
		}
		if len(got) != 2 || got["route"] != "GET /api/campaigns/{id}" || got["method"] != "GET" {
			t.Errorf("histogram labels = %v, want route and method", got)
		}
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

var idSegment = regexp.MustCompile(`(^|/)[0-9]+(/|$)`)

// FR-19: no label value is a campaign id or an email. The names le, quantile,
// code and version hold numbers by design (D-081).
func TestMetrics_NoIDOrEmailInLabels(t *testing.T) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	m := NewHTTPMetrics(reg)
	mux := http.NewServeMux()
	ok := func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) }
	mux.HandleFunc("GET /api/campaigns/{id}", ok)
	mux.HandleFunc("GET /api/campaigns/{id}/progress", ok)
	h := m.Wrap(mux)
	for _, path := range []string{"/api/campaigns/1", "/api/campaigns/1/progress"} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	skip := map[string]bool{"le": true, "quantile": true, "code": true, "version": true}
	checked := 0
	for _, mf := range mfs {
		for _, mt := range mf.Metric {
			for _, l := range mt.Label {
				if skip[l.GetName()] {
					continue
				}
				checked++
				if l.GetValue() == "1" || idSegment.MatchString(l.GetValue()) || strings.Contains(l.GetValue(), "@") {
					t.Errorf("%s: label %s=%q holds an id or an email", mf.GetName(), l.GetName(), l.GetValue())
				}
			}
		}
	}
	if checked == 0 {
		t.Error("no label was checked")
	}
}
