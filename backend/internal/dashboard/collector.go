package dashboard

import (
	"context"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const collectTimeout = 2 * time.Second

type collector struct {
	store *Store
	log   *slog.Logger
	jobs  *prometheus.Desc
	camps *prometheus.Desc
}

// NewCollector reads the send job and campaign counts from Postgres at each
// scrape. If the query fails, it logs the error and sends no metric.
func NewCollector(s *Store, log *slog.Logger) prometheus.Collector {
	return &collector{
		store: s,
		log:   log,
		jobs:  prometheus.NewDesc("campaign_send_jobs", "Send jobs by channel and status.", []string{"channel", "status"}, nil),
		camps: prometheus.NewDesc("campaign_campaigns", "Campaigns by status.", []string{"status"}, nil),
	}
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.jobs
	ch <- c.camps
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), collectTimeout)
	defer cancel()
	jobs, camps, err := c.store.Counts(ctx)
	if err != nil {
		c.log.Error("metrics: count query failed", "err", err)
		return
	}
	for channel, byStatus := range jobs {
		for status, n := range byStatus {
			ch <- prometheus.MustNewConstMetric(c.jobs, prometheus.GaugeValue, float64(n), string(channel), string(status))
		}
	}
	for status, n := range camps {
		ch <- prometheus.MustNewConstMetric(c.camps, prometheus.GaugeValue, float64(n), status)
	}
}
