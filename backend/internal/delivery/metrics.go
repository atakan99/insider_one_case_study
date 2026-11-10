package delivery

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const depthTimeout = time.Second

// String returns the label value of the outcome kind.
func (k OutcomeKind) String() string {
	switch k {
	case Delivered:
		return "delivered"
	case Retryable:
		return "retryable"
	case Permanent:
		return "permanent"
	case Ambiguous:
		return "ambiguous"
	}
	return "unknown"
}

// Metrics holds the Prometheus metrics of the worker. No label has a user,
// a campaign or message text. A nil *Metrics does nothing.
type Metrics struct {
	attempts *prometheus.CounterVec
	duration *prometheus.HistogramVec
	results  *prometheus.CounterVec
	waited   *prometheus.CounterVec
	retries  *prometheus.CounterVec
}

// NewMetrics registers the metrics on reg. When q is not nil, it also
// registers the stream gauges, which read Redis at each scrape.
func NewMetrics(reg prometheus.Registerer, q *Queue) *Metrics {
	m := &Metrics{
		attempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "campaign_send_attempts_total",
			Help: "Calls to the provider, by channel and outcome kind.",
		}, []string{"channel", "outcome"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "campaign_provider_request_duration_seconds",
			Help:    "Duration of one provider call.",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		}, []string{"channel"}),
		results: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "campaign_send_results_total",
			Help: "Delivery statuses written by an attempt or a skip.",
		}, []string{"channel", "status"}),
		waited: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "campaign_rate_limit_wait_seconds_total",
			Help: "Seconds that consumers waited for the rate limiter.",
		}, []string{"channel"}),
		retries: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "campaign_retries_scheduled_total",
			Help: "Retries that were scheduled.",
		}, []string{"channel"}),
	}
	reg.MustRegister(m.attempts, m.duration, m.results, m.waited, m.retries)
	if q != nil {
		reg.MustRegister(&streamCollector{q: q})
	}
	return m
}

// Attempt counts one provider call and records its duration.
func (m *Metrics) Attempt(ch Channel, kind OutcomeKind, d time.Duration) {
	if m == nil {
		return
	}
	m.attempts.WithLabelValues(string(ch), kind.String()).Inc()
	m.duration.WithLabelValues(string(ch)).Observe(d.Seconds())
}

// Result counts one status that was written to a send job.
func (m *Metrics) Result(ch Channel, s Status) {
	if m == nil {
		return
	}
	m.results.WithLabelValues(string(ch), string(s)).Inc()
}

// Waited adds the time that a consumer waited for the rate limiter.
func (m *Metrics) Waited(ch Channel, d time.Duration) {
	if m == nil || d <= 0 {
		return
	}
	m.waited.WithLabelValues(string(ch)).Add(d.Seconds())
}

// RetryScheduled counts one scheduled retry.
func (m *Metrics) RetryScheduled(ch Channel) {
	if m == nil {
		return
	}
	m.retries.WithLabelValues(string(ch)).Inc()
}

// streamCollector reads the depth of each stream at each scrape. On an error
// it sends no sample for that channel.
type streamCollector struct{ q *Queue }

var (
	streamLengthDesc  = prometheus.NewDesc("campaign_stream_length", "Entries in the send stream.", []string{"channel"}, nil)
	streamPendingDesc = prometheus.NewDesc("campaign_stream_pending", "Entries that a consumer read and did not acknowledge.", []string{"channel"}, nil)
)

func (c *streamCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- streamLengthDesc
	ch <- streamPendingDesc
}

func (c *streamCollector) Collect(out chan<- prometheus.Metric) {
	for _, ch := range Channels {
		ctx, cancel := context.WithTimeout(context.Background(), depthTimeout)
		length, pending, err := c.q.Depth(ctx, ch)
		cancel()
		if err != nil {
			continue
		}
		out <- prometheus.MustNewConstMetric(streamLengthDesc, prometheus.GaugeValue, float64(length), string(ch))
		out <- prometheus.MustNewConstMetric(streamPendingDesc, prometheus.GaugeValue, float64(pending), string(ch))
	}
}
