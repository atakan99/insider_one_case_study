package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const unmatchedRoute = "unmatched"

// streamRoute is the content stream. Other stream routes end in /stream.
const contentStreamRoute = "POST /api/content/generate"

// knownMethods bound the method label. Any other method counts as "other".
var knownMethods = map[string]bool{
	http.MethodGet: true, http.MethodPost: true, http.MethodPut: true, http.MethodPatch: true,
	http.MethodDelete: true, http.MethodHead: true, http.MethodOptions: true,
}

// HTTPMetrics counts requests and times them, by route pattern.
type HTTPMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// NewHTTPMetrics registers the request counter and the duration histogram.
func NewHTTPMetrics(reg prometheus.Registerer) *HTTPMetrics {
	m := &HTTPMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "campaign_http_requests_total",
			Help: "HTTP requests by route pattern, method and status code.",
		}, []string{"route", "method", "code"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "campaign_http_request_duration_seconds",
			Help:    "Duration of HTTP requests by route pattern. Stream routes are not timed.",
			Buckets: prometheus.DefBuckets,
		}, []string{"route"}),
	}
	reg.MustRegister(m.requests, m.duration)
	return m
}

// Wrap records each request after the handler returns. The label is the mux
// pattern, never the raw path, so that the label set stays small.
func (m *HTTPMetrics) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)

		// The mux sets r.Pattern on this same request.
		route := r.Pattern
		if route == "" {
			route = unmatchedRoute
		}
		method := r.Method
		if !knownMethods[method] {
			method = "other"
		}
		code := sw.status
		if code == 0 {
			code = http.StatusOK
		}
		m.requests.WithLabelValues(route, method, strconv.Itoa(code)).Inc()
		if !isStreamRoute(route) {
			m.duration.WithLabelValues(route).Observe(time.Since(start).Seconds())
		}
	})
}

func isStreamRoute(pattern string) bool {
	return strings.HasSuffix(pattern, "/stream") || pattern == contentStreamRoute
}

// statusWriter records the status code. It keeps http.Flusher for the SSE
// handlers, and Unwrap lets http.ResponseController reach the real writer.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	if s.status == 0 {
		s.status = code
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusWriter) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

func (s *statusWriter) Flush() {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	http.NewResponseController(s.ResponseWriter).Flush()
}

func (s *statusWriter) Unwrap() http.ResponseWriter { return s.ResponseWriter }
