// Package dashboard reads the delivery metrics that the dashboard shows.
package dashboard

import (
	"time"

	"campaigncomposer/internal/delivery"
)

// ChannelTotals counts the send jobs of one channel over all campaigns.
// InProgress is queued + sending + retrying.
type ChannelTotals struct {
	Sent       int64 `json:"sent"`
	Failed     int64 `json:"failed"`
	Unknown    int64 `json:"unknown"`
	Skipped    int64 `json:"skipped"`
	InProgress int64 `json:"in_progress"`
}

// Throughput is the messages per second of the last 10 minutes in 120
// buckets of 5 seconds. Value i is the bucket that starts at Start + i*5 s.
type Throughput struct {
	BucketSeconds int       `json:"bucket_seconds"`
	Start         time.Time `json:"start"`
	SMS           []float64 `json:"sms"`
	Email         []float64 `json:"email"`
}

// Failure is one pair of status and last_error, with the count per channel.
type Failure struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
	SMS    int64  `json:"sms"`
	Email  int64  `json:"email"`
}

// CampaignRow is one row of the campaign table. DurationMS is nil unless the
// campaign is completed.
type CampaignRow struct {
	ID          int64  `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Total       int64  `json:"total"`
	Sent        int64  `json:"sent"`
	Failed      int64  `json:"failed"`
	Unknown     int64  `json:"unknown"`
	DurationMS  *int64 `json:"duration_ms"`
}

// Stats is the body of GET /api/dashboard/stats.
type Stats struct {
	GeneratedAt  time.Time                          `json:"generated_at"`
	Totals       map[delivery.Channel]ChannelTotals `json:"totals"`
	Throughput   Throughput                         `json:"throughput"`
	Failures     []Failure                          `json:"failures"`
	MeanAttempts map[delivery.Channel]*float64      `json:"mean_attempts"`
	Campaigns    []CampaignRow                      `json:"campaigns"`
}
