package delivery

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/segment"
)

// statementTimeout limits each query, as in the campaign package.
const statementTimeout = "2s"

const (
	defaultListLimit = 50
	maxListLimit     = 200
)

// Store reads and writes send jobs in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore returns a Store on pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) inTx(ctx context.Context, mode pgx.TxAccessMode, fn func(pgx.Tx) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: mode})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = '"+statementTimeout+"'"); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// PlanNext locks one approved campaign (FOR UPDATE SKIP LOCKED), inserts its
// send jobs, and sets it queued, or completed when it has none. ok is false
// when no campaign waits. ids are the new queued send job ids per channel.
// When the stored rule is not valid, it logs the error and completes the
// campaign with no jobs.
func (s *Store) PlanNext(ctx context.Context, allow segment.Allowlist, now time.Time) (campaignID int64, ids map[Channel][]int64, ok bool, err error) {
	ids = map[Channel][]int64{}
	err = s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		var raw []byte
		e := tx.QueryRow(ctx, `SELECT id, rule FROM campaigns WHERE status = 'approved'
ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&campaignID, &raw)
		if errors.Is(e, pgx.ErrNoRows) {
			return nil
		}
		if e != nil {
			return e
		}
		ok = true
		rule, ferrs := allow.ParseRule(raw)
		if len(ferrs) > 0 {
			slog.Default().Error("send planner: stored rule is not valid, campaign completed with no jobs",
				"campaign_id", campaignID, "error", ferrs[0].Error())
			return setCampaign(ctx, tx, campaignID, "completed", now)
		}
		where := allow.Compile(rule, now)

		langs := []string{}
		rs, e := tx.Query(ctx, `SELECT language FROM campaign_contents WHERE campaign_id = $1 ORDER BY language`, campaignID)
		if e != nil {
			return e
		}
		for rs.Next() {
			var l string
			if e := rs.Scan(&l); e != nil {
				rs.Close()
				return e
			}
			langs = append(langs, l)
		}
		rs.Close()
		if e := rs.Err(); e != nil {
			return e
		}

		total := 0
		for _, ch := range Channels {
			sch, e := segment.ChannelByName(string(ch))
			if e != nil {
				return e
			}
			q, args, e := segment.ReachableQuery(where, sch, langs)
			if e != nil {
				return e
			}
			n := len(args)
			args = append(args, campaignID, now)
			ins := fmt.Sprintf(`INSERT INTO send_jobs (campaign_id, user_id, channel, language, status, created_at, updated_at)
SELECT $%d, id, '%s', language, 'queued', $%d, $%d FROM (%s) r
ON CONFLICT DO NOTHING RETURNING id`, n+1, string(ch), n+2, n+2, q)
			irs, e := tx.Query(ctx, ins, args...)
			if e != nil {
				return e
			}
			for irs.Next() {
				var id int64
				if e := irs.Scan(&id); e != nil {
					irs.Close()
					return e
				}
				ids[ch] = append(ids[ch], id)
				total++
			}
			irs.Close()
			if e := irs.Err(); e != nil {
				return e
			}
		}
		if total == 0 {
			return setCampaign(ctx, tx, campaignID, "completed", now)
		}
		return setCampaign(ctx, tx, campaignID, "queued", now)
	})
	if err != nil {
		return 0, map[Channel][]int64{}, false, fmt.Errorf("plan next campaign: %w", err)
	}
	return campaignID, ids, ok, nil
}

func setCampaign(ctx context.Context, tx pgx.Tx, id int64, status string, now time.Time) error {
	_, err := tx.Exec(ctx, `UPDATE campaigns SET status = $2, updated_at = $3 WHERE id = $1`, id, status, now)
	return err
}

// MarkEnqueued records that the send jobs ids were pushed to the queue at at.
func (s *Store) MarkEnqueued(ctx context.Context, ids []int64, at time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	err := s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE send_jobs SET enqueued_at = $2 WHERE id = ANY($1)`, ids, at)
		return err
	})
	if err != nil {
		return fmt.Errorf("mark enqueued: %w", err)
	}
	return nil
}

const jobColumns = `id, campaign_id, user_id, channel, language, status, attempts, last_error, retryable, updated_at`

func scanJob(row pgx.Row) (Job, error) {
	var j Job
	err := row.Scan(&j.ID, &j.CampaignID, &j.UserID, &j.Channel, &j.Language, &j.Status,
		&j.Attempts, &j.LastError, &j.Retryable, &j.UpdatedAt)
	return j, err
}

// Claim moves a queued or due retrying job to sending and adds 1 to attempts.
// It also moves the campaign from queued to sending. ok is false when no row changed.
func (s *Store) Claim(ctx context.Context, id int64, now time.Time) (Job, bool, error) {
	var j Job
	var ok bool
	err := s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		var e error
		j, e = scanJob(tx.QueryRow(ctx, `UPDATE send_jobs SET status = 'sending', attempts = attempts + 1, updated_at = $2
WHERE id = $1 AND (status = 'queued' OR (status = 'retrying' AND next_attempt_at <= $2))
RETURNING `+jobColumns, id, now))
		if errors.Is(e, pgx.ErrNoRows) {
			return nil
		}
		if e != nil {
			return e
		}
		ok = true
		_, e = tx.Exec(ctx, `UPDATE campaigns SET status = 'sending', updated_at = $2 WHERE id = $1 AND status = 'queued'`, j.CampaignID, now)
		return e
	})
	if err != nil {
		return Job{}, false, fmt.Errorf("claim job: %w", err)
	}
	return j, ok, nil
}

// Claimable reports whether Claim could take the job now. It only saves rate
// tokens: the UPDATE in Claim stays the gate.
func (s *Store) Claimable(ctx context.Context, id int64, now time.Time) (bool, error) {
	var ok bool
	err := s.inTx(ctx, pgx.ReadOnly, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM send_jobs WHERE id = $1
AND (status = 'queued' OR (status = 'retrying' AND next_attempt_at <= $2)))`, id, now).Scan(&ok)
	})
	if err != nil {
		return false, fmt.Errorf("check job: %w", err)
	}
	return ok, nil
}

// Release moves a sending job back to queued and subtracts 1 from attempts.
func (s *Store) Release(ctx context.Context, id int64) error {
	err := s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE send_jobs SET status = 'queued', attempts = attempts - 1, updated_at = now()
WHERE id = $1 AND status = 'sending'`, id)
		return err
	})
	if err != nil {
		return fmt.Errorf("release job: %w", err)
	}
	return nil
}

// Finish writes the result of an attempt to a sending job. A sent job keeps
// the last_error of an earlier attempt. A job that is not
// sending stays as it is, and this is not an error.
func (s *Store) Finish(ctx context.Context, id int64, r Result, now time.Time) error {
	err := s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		if r.Status == StatusSent {
			_, err := tx.Exec(ctx, `UPDATE send_jobs SET status = $2, last_error = COALESCE(NULLIF($3, ''), last_error), retryable = $4, next_attempt_at = $5,
provider_id = $6, sent_at = $7, updated_at = $7 WHERE id = $1 AND status = 'sending'`,
				id, string(r.Status), r.LastError, r.Retryable, r.NextAttemptAt, r.ProviderID, now)
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE send_jobs SET status = $2, last_error = $3, retryable = $4, next_attempt_at = $5,
updated_at = $6 WHERE id = $1 AND status = 'sending'`,
			id, string(r.Status), r.LastError, r.Retryable, r.NextAttemptAt, now)
		return err
	})
	if err != nil {
		return fmt.Errorf("finish job: %w", err)
	}
	return nil
}

// Load reads the user (with consent) and the content of the job language,
// renders it, and returns the message. skip is "" or one of the skip reasons.
func (s *Store) Load(ctx context.Context, j Job) (m Message, skip string, err error) {
	sch, err := segment.ChannelByName(string(j.Channel))
	if err != nil {
		return Message{}, "", fmt.Errorf("load job %d: %w", j.ID, err)
	}
	err = s.inTx(ctx, pgx.ReadOnly, func(tx pgx.Tx) error {
		r, found, e := segment.RecipientByID(ctx, tx, j.UserID, sch)
		if e != nil {
			return e
		}
		if !found {
			skip = SkipNoConsent
			return nil
		}
		var sms, subject, body string
		if e := tx.QueryRow(ctx, `SELECT sms, email_subject, email_body FROM campaign_contents
WHERE campaign_id = $1 AND language = $2`, j.CampaignID, j.Language).Scan(&sms, &subject, &body); e != nil {
			return e
		}
		switch j.Channel {
		case ChannelSMS:
			if r.Phone == "" {
				skip = SkipNoContact
				return nil
			}
			text := content.Render(sms, r.FirstName)
			if utf8.RuneCountInString(text) > content.MaxSMSRunes {
				skip = SkipOverLimit
				return nil
			}
			m = Message{To: r.Phone, Body: text}
		default:
			if r.Email == "" {
				skip = SkipNoContact
				return nil
			}
			subj := content.Render(subject, r.FirstName)
			if utf8.RuneCountInString(subj) > content.MaxSubjectRunes {
				skip = SkipOverLimit
				return nil
			}
			m = Message{To: r.Email, Subject: subj, Body: content.Render(body, r.FirstName)}
		}
		return nil
	})
	if err != nil {
		return Message{}, "", fmt.Errorf("load job %d: %w", j.ID, err)
	}
	return m, skip, nil
}

func groupIDs(rs pgx.Rows) (map[Channel][]int64, error) {
	defer rs.Close()
	out := map[Channel][]int64{}
	for rs.Next() {
		var id int64
		var ch Channel
		if err := rs.Scan(&id, &ch); err != nil {
			return nil, err
		}
		out[ch] = append(out[ch], id)
	}
	return out, rs.Err()
}

func (s *Store) idsByChannel(ctx context.Context, sql string, args ...any) (map[Channel][]int64, error) {
	var out map[Channel][]int64
	err := s.inTx(ctx, pgx.ReadOnly, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, sql, args...)
		if err != nil {
			return err
		}
		out, err = groupIDs(rs)
		return err
	})
	return out, err
}

// DueRetries returns the retrying jobs whose next_attempt_at is at least
// lateBy before now.
func (s *Store) DueRetries(ctx context.Context, now time.Time, lateBy time.Duration) (map[Channel][]int64, error) {
	out, err := s.idsByChannel(ctx, `SELECT id, channel FROM send_jobs
WHERE status = 'retrying' AND next_attempt_at <= $1 ORDER BY id`, now.Add(-lateBy))
	if err != nil {
		return nil, fmt.Errorf("due retries: %w", err)
	}
	return out, nil
}

// StaleQueued returns the queued jobs that were never pushed (enqueued_at is
// null) or were pushed more than age ago.
func (s *Store) StaleQueued(ctx context.Context, now time.Time, age time.Duration) (map[Channel][]int64, error) {
	out, err := s.idsByChannel(ctx, `SELECT id, channel FROM send_jobs
WHERE status = 'queued' AND (enqueued_at IS NULL OR enqueued_at < $1) ORDER BY id`, now.Add(-age))
	if err != nil {
		return nil, fmt.Errorf("stale queued: %w", err)
	}
	return out, nil
}

// ExpireSending sets unknown on the jobs that have been sending for more than
// age. The outcome of such a job is not known, so it is never sent again.
func (s *Store) ExpireSending(ctx context.Context, now time.Time, age time.Duration) (int64, error) {
	var n int64
	err := s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE send_jobs SET status = 'unknown', retryable = false,
last_error = 'worker stopped before the result was known', updated_at = $1
WHERE status = 'sending' AND updated_at < $2`, now, now.Add(-age))
		n = tag.RowsAffected()
		return err
	})
	if err != nil {
		return 0, fmt.Errorf("expire sending: %w", err)
	}
	return n, nil
}

// CompleteCampaigns sets completed on the queued or sending campaigns that
// have no job in queued, sending or retrying. It returns their ids.
func (s *Store) CompleteCampaigns(ctx context.Context) ([]int64, error) {
	ids := []int64{}
	err := s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `UPDATE campaigns SET status = 'completed', updated_at = now()
WHERE status IN ('queued', 'sending') AND NOT EXISTS (
  SELECT 1 FROM send_jobs WHERE campaign_id = campaigns.id AND status IN ('queued', 'sending', 'retrying'))
RETURNING id`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id int64
			if err := rs.Scan(&id); err != nil {
				return err
			}
			ids = append(ids, id)
		}
		return rs.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("complete campaigns: %w", err)
	}
	return ids, nil
}

// Progress counts the jobs of a campaign by channel and status.
// ErrNotFound for an unknown id, ErrNotSendable for a draft.
func (s *Store) Progress(ctx context.Context, campaignID int64) (Progress, error) {
	p := Progress{CampaignID: campaignID, Channels: map[Channel]ChannelProgress{}}
	for _, ch := range Channels {
		p.Channels[ch] = ChannelProgress{}
	}
	err := s.inTx(ctx, pgx.ReadOnly, func(tx pgx.Tx) error {
		if e := tx.QueryRow(ctx, `SELECT status FROM campaigns WHERE id = $1`, campaignID).Scan(&p.Status); e != nil {
			if errors.Is(e, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return e
		}
		if p.Status == "draft" {
			return ErrNotSendable
		}
		rs, e := tx.Query(ctx, `SELECT channel, status, count(*) FROM send_jobs WHERE campaign_id = $1 GROUP BY channel, status`, campaignID)
		if e != nil {
			return e
		}
		for rs.Next() {
			var ch Channel
			var st Status
			var n int64
			if e := rs.Scan(&ch, &st, &n); e != nil {
				rs.Close()
				return e
			}
			c := p.Channels[ch]
			c.Total += n
			switch st {
			case StatusQueued:
				c.Queued = n
			case StatusSending:
				c.Sending = n
			case StatusRetrying:
				c.Retrying = n
			case StatusSent:
				c.Sent = n
			case StatusFailed:
				c.Failed = n
			case StatusUnknown:
				c.Unknown = n
			case StatusSkipped:
				c.Skipped = n
			}
			p.Channels[ch] = c
		}
		rs.Close()
		if e := rs.Err(); e != nil {
			return e
		}
		rs, e = tx.Query(ctx, `SELECT channel, count(*) FROM send_jobs
WHERE campaign_id = $1 AND status = 'failed' AND retryable GROUP BY channel`, campaignID)
		if e != nil {
			return e
		}
		defer rs.Close()
		for rs.Next() {
			var ch Channel
			var n int64
			if e := rs.Scan(&ch, &n); e != nil {
				return e
			}
			c := p.Channels[ch]
			c.RetryableFailed = n
			p.Channels[ch] = c
		}
		return rs.Err()
	})
	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrNotSendable) {
		return Progress{}, err
	}
	if err != nil {
		return Progress{}, fmt.Errorf("progress: %w", err)
	}
	return p, nil
}

// ListJobs returns the jobs of a campaign ordered by id. A Limit of 0 or less
// means 50, and the maximum is 200. nextAfter is the last id of this page when
// more jobs follow, and nil otherwise.
func (s *Store) ListJobs(ctx context.Context, campaignID int64, f JobFilter) (jobs []Job, nextAfter *int64, err error) {
	limit := f.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	jobs = []Job{}
	err = s.inTx(ctx, pgx.ReadOnly, func(tx pgx.Tx) error {
		rs, e := tx.Query(ctx, `SELECT `+jobColumns+` FROM send_jobs
WHERE campaign_id = $1 AND id > $2
  AND ($3 = '' OR status = $3) AND ($4 = '' OR channel = $4)
ORDER BY id LIMIT $5`, campaignID, f.After, string(f.Status), string(f.Channel), limit+1)
		if e != nil {
			return e
		}
		defer rs.Close()
		for rs.Next() {
			j, e := scanJob(rs)
			if e != nil {
				return e
			}
			jobs = append(jobs, j)
		}
		return rs.Err()
	})
	if err != nil {
		return nil, nil, fmt.Errorf("list jobs: %w", err)
	}
	if len(jobs) > limit {
		jobs = jobs[:limit]
		last := jobs[limit-1].ID
		nextAfter = &last
	}
	return jobs, nextAfter, nil
}

// Retry resets the retryable failed jobs of a completed campaign. It returns their ids per channel.
func (s *Store) Retry(ctx context.Context, campaignID int64, now time.Time) (map[Channel][]int64, error) {
	var out map[Channel][]int64
	err := s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		var status string
		if e := tx.QueryRow(ctx, `SELECT status FROM campaigns WHERE id = $1 FOR UPDATE`, campaignID).Scan(&status); e != nil {
			if errors.Is(e, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return e
		}
		if status != "completed" {
			return ErrNotCompleted
		}
		rs, e := tx.Query(ctx, `UPDATE send_jobs SET status = 'queued', attempts = 0, next_attempt_at = NULL,
enqueued_at = NULL, retryable = false, updated_at = $2
WHERE campaign_id = $1 AND status = 'failed' AND retryable RETURNING id, channel`, campaignID, now)
		if e != nil {
			return e
		}
		out, e = groupIDs(rs)
		if e != nil {
			return e
		}
		if len(out) == 0 {
			return ErrNothingToRetry
		}
		return setCampaign(ctx, tx, campaignID, "sending", now)
	})
	switch {
	case err == nil:
		return out, nil
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrNotCompleted), errors.Is(err, ErrNothingToRetry):
		return nil, err
	}
	return nil, fmt.Errorf("retry failed jobs: %w", err)
}
