package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/delivery"
)

const (
	bucketSeconds = 5
	bucketCount   = 120
	windowSeconds = bucketSeconds * bucketCount
	failureRows   = 8
	campaignRows  = 20
)

// Store reads dashboard data from Postgres. It selects from send_jobs and
// campaigns only.
type Store struct{ pool *pgxpool.Pool }

// NewStore returns a Store on pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// Stats returns all parts of the dashboard from one snapshot of the data.
func (s *Store) Stats(ctx context.Context, now time.Time) (Stats, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return Stats{}, fmt.Errorf("dashboard: begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // read-only

	st := Stats{
		GeneratedAt:  now.UTC(),
		Totals:       map[delivery.Channel]ChannelTotals{},
		MeanAttempts: map[delivery.Channel]*float64{},
		Failures:     []Failure{},
		Campaigns:    []CampaignRow{},
	}
	for _, ch := range delivery.Channels {
		st.Totals[ch] = ChannelTotals{}
		st.MeanAttempts[ch] = nil
	}

	rows, err := tx.Query(ctx, `SELECT channel, status, count(*) FROM send_jobs GROUP BY 1, 2`)
	if err != nil {
		return Stats{}, fmt.Errorf("dashboard: totals: %w", err)
	}
	for rows.Next() {
		var ch delivery.Channel
		var status delivery.Status
		var n int64
		if err := rows.Scan(&ch, &status, &n); err != nil {
			rows.Close()
			return Stats{}, fmt.Errorf("dashboard: totals: %w", err)
		}
		t := st.Totals[ch]
		switch status {
		case delivery.StatusSent:
			t.Sent = n
		case delivery.StatusFailed:
			t.Failed = n
		case delivery.StatusUnknown:
			t.Unknown = n
		case delivery.StatusSkipped:
			t.Skipped = n
		case delivery.StatusQueued, delivery.StatusSending, delivery.StatusRetrying:
			t.InProgress += n
		}
		st.Totals[ch] = t
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return Stats{}, fmt.Errorf("dashboard: totals: %w", err)
	}

	start := now.UTC().Truncate(bucketSeconds * time.Second).Add(-windowSeconds * time.Second)
	st.Throughput = Throughput{
		BucketSeconds: bucketSeconds,
		Start:         start,
		SMS:           make([]float64, bucketCount),
		Email:         make([]float64, bucketCount),
	}
	rows, err = tx.Query(ctx, `SELECT channel, floor(extract(epoch FROM sent_at - $1::timestamptz) / 5)::int AS b, count(*)
FROM send_jobs WHERE sent_at >= $1::timestamptz AND sent_at < $1::timestamptz + interval '600 seconds'
GROUP BY 1, 2`, start)
	if err != nil {
		return Stats{}, fmt.Errorf("dashboard: throughput: %w", err)
	}
	for rows.Next() {
		var ch delivery.Channel
		var b int
		var n int64
		if err := rows.Scan(&ch, &b, &n); err != nil {
			rows.Close()
			return Stats{}, fmt.Errorf("dashboard: throughput: %w", err)
		}
		if b < 0 || b >= bucketCount {
			continue
		}
		switch ch {
		case delivery.ChannelSMS:
			st.Throughput.SMS[b] = float64(n) / bucketSeconds
		case delivery.ChannelEmail:
			st.Throughput.Email[b] = float64(n) / bucketSeconds
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return Stats{}, fmt.Errorf("dashboard: throughput: %w", err)
	}

	rows, err = tx.Query(ctx, `SELECT status, last_error,
  count(*) FILTER (WHERE channel = 'sms'), count(*) FILTER (WHERE channel = 'email')
FROM send_jobs WHERE status IN ('failed', 'unknown', 'skipped')
GROUP BY status, last_error
ORDER BY count(*) DESC, status, last_error
LIMIT $1`, failureRows)
	if err != nil {
		return Stats{}, fmt.Errorf("dashboard: failures: %w", err)
	}
	for rows.Next() {
		var f Failure
		if err := rows.Scan(&f.Status, &f.Reason, &f.SMS, &f.Email); err != nil {
			rows.Close()
			return Stats{}, fmt.Errorf("dashboard: failures: %w", err)
		}
		st.Failures = append(st.Failures, f)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return Stats{}, fmt.Errorf("dashboard: failures: %w", err)
	}

	rows, err = tx.Query(ctx, `SELECT channel, avg(attempts)::float8 FROM send_jobs WHERE status = 'sent' GROUP BY channel`)
	if err != nil {
		return Stats{}, fmt.Errorf("dashboard: mean attempts: %w", err)
	}
	for rows.Next() {
		var ch delivery.Channel
		var v float64
		if err := rows.Scan(&ch, &v); err != nil {
			rows.Close()
			return Stats{}, fmt.Errorf("dashboard: mean attempts: %w", err)
		}
		st.MeanAttempts[ch] = &v
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return Stats{}, fmt.Errorf("dashboard: mean attempts: %w", err)
	}

	rows, err = tx.Query(ctx, `SELECT c.id, c.description, c.status,
  coalesce(a.total, 0), coalesce(a.sent, 0), coalesce(a.failed, 0), coalesce(a.unknown, 0),
  CASE WHEN c.status = 'completed' AND a.total > 0
       THEN round(extract(epoch FROM a.last_update - a.first_create) * 1000)::bigint END
FROM (SELECT id, description, status FROM campaigns WHERE status <> 'draft' ORDER BY id DESC LIMIT $1) c
LEFT JOIN LATERAL (
  SELECT count(*) AS total,
         count(*) FILTER (WHERE status = 'sent') AS sent,
         count(*) FILTER (WHERE status = 'failed') AS failed,
         count(*) FILTER (WHERE status = 'unknown') AS unknown,
         min(created_at) AS first_create, max(updated_at) AS last_update
  FROM send_jobs WHERE campaign_id = c.id
) a ON true
ORDER BY c.id DESC`, campaignRows)
	if err != nil {
		return Stats{}, fmt.Errorf("dashboard: campaigns: %w", err)
	}
	for rows.Next() {
		var c CampaignRow
		if err := rows.Scan(&c.ID, &c.Description, &c.Status, &c.Total, &c.Sent, &c.Failed, &c.Unknown, &c.DurationMS); err != nil {
			rows.Close()
			return Stats{}, fmt.Errorf("dashboard: campaigns: %w", err)
		}
		st.Campaigns = append(st.Campaigns, c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return Stats{}, fmt.Errorf("dashboard: campaigns: %w", err)
	}
	return st, nil
}

// InProgress counts the send jobs that are queued, sending or retrying for
// each channel. Both channels are always present.
func (s *Store) InProgress(ctx context.Context) (map[delivery.Channel]int64, error) {
	out := map[delivery.Channel]int64{}
	for _, ch := range delivery.Channels {
		out[ch] = 0
	}
	rows, err := s.pool.Query(ctx, `SELECT channel, count(*) FROM send_jobs
WHERE status IN ('queued', 'sending', 'retrying') GROUP BY channel`)
	if err != nil {
		return nil, fmt.Errorf("dashboard: in progress: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ch delivery.Channel
		var n int64
		if err := rows.Scan(&ch, &n); err != nil {
			return nil, fmt.Errorf("dashboard: in progress: %w", err)
		}
		out[ch] = n
	}
	return out, rows.Err()
}

// Counts returns the send jobs by channel and status and the campaigns by
// status. Rows with a count of 0 are absent.
func (s *Store) Counts(ctx context.Context) (jobs map[delivery.Channel]map[delivery.Status]int64, campaigns map[string]int64, err error) {
	jobs = map[delivery.Channel]map[delivery.Status]int64{}
	campaigns = map[string]int64{}
	rows, err := s.pool.Query(ctx, `SELECT channel, status, count(*) FROM send_jobs GROUP BY 1, 2`)
	if err != nil {
		return nil, nil, fmt.Errorf("dashboard: job counts: %w", err)
	}
	for rows.Next() {
		var ch delivery.Channel
		var st delivery.Status
		var n int64
		if err := rows.Scan(&ch, &st, &n); err != nil {
			rows.Close()
			return nil, nil, fmt.Errorf("dashboard: job counts: %w", err)
		}
		if jobs[ch] == nil {
			jobs[ch] = map[delivery.Status]int64{}
		}
		jobs[ch][st] = n
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("dashboard: job counts: %w", err)
	}
	rows, err = s.pool.Query(ctx, `SELECT status, count(*) FROM campaigns GROUP BY 1`)
	if err != nil {
		return nil, nil, fmt.Errorf("dashboard: campaign counts: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var st string
		var n int64
		if err := rows.Scan(&st, &n); err != nil {
			return nil, nil, fmt.Errorf("dashboard: campaign counts: %w", err)
		}
		campaigns[st] = n
	}
	return jobs, campaigns, rows.Err()
}
