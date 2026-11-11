package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/campaign"
	"campaigncomposer/internal/delivery/deliverytest"
)

func newService(t *testing.T, s *Store, closeRedis bool) (*Service, *Queue) {
	t.Helper()
	rdb := deliverytest.NewRedis(t, 0)
	q := NewQueue(rdb, "api")
	if closeRedis {
		_ = rdb.Close()
	}
	return &Service{Store: s, Queue: q, Now: func() time.Time { return testNow }, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}, q
}

// failedCampaign plans a campaign and makes it completed with retryable failed jobs.
func failedCampaign(t *testing.T, ctx context.Context, pool *pgxpool.Pool, s *Store) (int64, map[Channel][]int64) {
	t.Helper()
	cid, ids := planned(t, ctx, pool, s)
	failed := append(append([]int64{}, ids[ChannelSMS][:3]...), ids[ChannelEmail][:2]...)
	setJobs(t, ctx, pool, failed, StatusFailed, ", retryable = true, attempts = 5, last_error = 'x'")
	setJobs(t, ctx, pool, openJobs(t, ctx, pool, cid), StatusSent, "")
	exec(t, ctx, pool, `UPDATE campaigns SET status = 'completed' WHERE id = $1`, cid)
	return cid, map[Channel][]int64{ChannelSMS: ids[ChannelSMS][:3], ChannelEmail: ids[ChannelEmail][:2]}
}

func TestService_RetryPushes(t *testing.T) {
	ctx, pool, s := seeded(t)
	svc, q := newService(t, s, false)
	if err := q.Ensure(ctx); err != nil {
		t.Fatal(err)
	}
	cid, failed := failedCampaign(t, ctx, pool, s)
	n, p, err := svc.Retry(ctx, cid)
	if err != nil || n != 5 {
		t.Fatalf("Retry = %d, %v", n, err)
	}
	if p.Status != "sending" {
		t.Fatalf("status = %s", p.Status)
	}
	for ch, ids := range failed {
		got := map[int64]bool{}
		for {
			es, err := q.Read(ctx, ch, 100, 0)
			if err != nil {
				t.Fatal(err)
			}
			if len(es) == 0 {
				break
			}
			for _, e := range es {
				got[e.JobID] = true
			}
		}
		for _, id := range ids {
			if !got[id] {
				t.Errorf("job %d not on the %s stream", id, ch)
			}
			var set bool
			if err := pool.QueryRow(ctx, `SELECT enqueued_at IS NOT NULL FROM send_jobs WHERE id = $1`, id).Scan(&set); err != nil || !set {
				t.Errorf("job %d enqueued_at set = %v, %v", id, set, err)
			}
		}
	}
}

func TestService_RetryPushFails(t *testing.T) {
	ctx, pool, s := seeded(t)
	svc, _ := newService(t, s, true)
	cid, failed := failedCampaign(t, ctx, pool, s)
	n, _, err := svc.Retry(ctx, cid)
	if err != nil || n != 5 {
		t.Fatalf("Retry = %d, %v", n, err)
	}
	for _, ids := range failed {
		for _, id := range ids {
			var st string
			var null bool
			if err := pool.QueryRow(ctx, `SELECT status, enqueued_at IS NULL FROM send_jobs WHERE id = $1`, id).Scan(&st, &null); err != nil {
				t.Fatal(err)
			}
			if st != "queued" || !null {
				t.Errorf("job %d = %s, enqueued_at null %v", id, st, null)
			}
		}
	}
}

func TestService_RetryErrors(t *testing.T) {
	ctx, pool, s := seeded(t)
	svc, _ := newService(t, s, false)
	if _, _, err := svc.Retry(ctx, 999999); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown: %v", err)
	}
	cid, _ := planned(t, ctx, pool, s)
	if _, _, err := svc.Retry(ctx, cid); !errors.Is(err, ErrNotCompleted) {
		t.Errorf("sending: %v", err)
	}
}

func TestService_ListJobsChecksCampaign(t *testing.T) {
	ctx, pool, s := seeded(t)
	svc, _ := newService(t, s, false)
	if _, _, err := svc.ListJobs(ctx, 999999, JobFilter{}); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown: %v", err)
	}
	draft, err := campaign.NewStore(pool).CreateDraft(ctx, campaign.Draft{Brief: "b", Description: "d", Rule: json.RawMessage(exampleRule)})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ListJobs(ctx, draft, JobFilter{}); !errors.Is(err, ErrNotSendable) {
		t.Errorf("draft: %v", err)
	}
	cid, _ := planned(t, ctx, pool, s)
	jobs, _, err := svc.ListJobs(ctx, cid, JobFilter{Limit: 5})
	if err != nil || len(jobs) != 5 {
		t.Errorf("ListJobs = %d, %v", len(jobs), err)
	}
}
