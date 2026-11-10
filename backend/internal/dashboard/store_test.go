package dashboard

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/delivery"
	"campaigncomposer/internal/store"
	"campaigncomposer/internal/store/storetest"
)

var base = time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)

func newDB(t *testing.T) (context.Context, *pgxpool.Pool, *Store) {
	t.Helper()
	url := storetest.NewDB(t)
	ctx := context.Background()
	if err := store.Migrate(ctx, url); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return ctx, pool, NewStore(pool)
}

func mustExec(t *testing.T, ctx context.Context, db *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := db.Exec(ctx, sql, args...); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

// insertUsers makes n users and returns their ids.
func insertUsers(t *testing.T, ctx context.Context, db *pgxpool.Pool, n int) []int64 {
	t.Helper()
	rows, err := db.Query(ctx, `INSERT INTO users (first_name, last_name, email, phone, language, country, city,
birth_date, email_consent, sms_consent, created_at, order_count, total_spent)
SELECT 'A', 'B', gen_random_uuid() || '@example.com', '+90', 'en', 'TR', 'Istanbul', DATE '1990-01-01', true, true, now(), 0, 0
FROM generate_series(1, $1) g RETURNING id`, n)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return ids
}

func insertCampaign(t *testing.T, ctx context.Context, db *pgxpool.Pool, status, desc string) int64 {
	t.Helper()
	var id int64
	err := db.QueryRow(ctx, `INSERT INTO campaigns (brief, rule, description, status, created_at, updated_at, approved_at)
VALUES ('b', '{}', $1, $2, $3, $3, CASE WHEN $2 = 'draft' THEN NULL ELSE $3::timestamptz END) RETURNING id`,
		desc, status, base).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func insertJob(t *testing.T, ctx context.Context, db *pgxpool.Pool, campaignID, userID int64, channel delivery.Channel,
	status delivery.Status, attempts int, lastError string, sentAt *time.Time, createdAt, updatedAt time.Time) {
	t.Helper()
	mustExec(t, ctx, db, `INSERT INTO send_jobs (campaign_id, user_id, channel, language, status, attempts, last_error, sent_at, created_at, updated_at)
VALUES ($1, $2, $3, 'en', $4, $5, $6, $7, $8, $9)`,
		campaignID, userID, string(channel), string(status), attempts, lastError, sentAt, createdAt, updatedAt)
}

// addJobs inserts n jobs with the same values, each for a new user.
func addJobs(t *testing.T, ctx context.Context, db *pgxpool.Pool, cid int64, ch delivery.Channel, st delivery.Status, n int) {
	t.Helper()
	for _, u := range insertUsers(t, ctx, db, n) {
		insertJob(t, ctx, db, cid, u, ch, st, 1, "", nil, base, base)
	}
}

func groupBy(t *testing.T, ctx context.Context, db *pgxpool.Pool) map[delivery.Channel]map[delivery.Status]int64 {
	t.Helper()
	rows, err := db.Query(ctx, `SELECT channel, status, count(*) FROM send_jobs GROUP BY 1, 2`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[delivery.Channel]map[delivery.Status]int64{}
	for rows.Next() {
		var ch delivery.Channel
		var st delivery.Status
		var n int64
		if err := rows.Scan(&ch, &st, &n); err != nil {
			t.Fatal(err)
		}
		if out[ch] == nil {
			out[ch] = map[delivery.Status]int64{}
		}
		out[ch][st] = n
	}
	return out
}

func seedTotals(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	t.Helper()
	cid := insertCampaign(t, ctx, db, "sending", "totals")
	for st, n := range map[delivery.Status]int{
		delivery.StatusSent: 3, delivery.StatusFailed: 1, delivery.StatusUnknown: 2, delivery.StatusSkipped: 1,
		delivery.StatusQueued: 1, delivery.StatusSending: 1, delivery.StatusRetrying: 1,
	} {
		addJobs(t, ctx, db, cid, delivery.ChannelSMS, st, n)
	}
	addJobs(t, ctx, db, cid, delivery.ChannelEmail, delivery.StatusSent, 2)
}

func TestStats_Totals(t *testing.T) {
	ctx, db, s := newDB(t)
	seedTotals(t, ctx, db)
	st, err := s.Stats(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := st.Totals[delivery.ChannelSMS], (ChannelTotals{3, 1, 2, 1, 3}); got != want {
		t.Errorf("sms totals = %+v, want %+v", got, want)
	}
	if got, want := st.Totals[delivery.ChannelEmail], (ChannelTotals{2, 0, 0, 0, 0}); got != want {
		t.Errorf("email totals = %+v, want %+v", got, want)
	}
	g := groupBy(t, ctx, db)
	for _, ch := range delivery.Channels {
		tt := st.Totals[ch]
		if tt.Sent != g[ch][delivery.StatusSent] || tt.Failed != g[ch][delivery.StatusFailed] ||
			tt.Unknown != g[ch][delivery.StatusUnknown] || tt.Skipped != g[ch][delivery.StatusSkipped] {
			t.Errorf("%s totals %+v differ from GROUP BY %v", ch, tt, g[ch])
		}
	}
}

func TestStats_ThroughputBuckets(t *testing.T) {
	ctx, db, s := newDB(t)
	cid := insertCampaign(t, ctx, db, "sending", "tp")
	users := insertUsers(t, ctx, db, 53)
	day := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	at := func(h, m, sec, ms int) *time.Time {
		v := day.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec)*time.Second + time.Duration(ms)*time.Millisecond)
		return &v
	}
	for _, u := range users[:50] {
		insertJob(t, ctx, db, cid, u, delivery.ChannelSMS, delivery.StatusSent, 1, "", at(11, 55, 0, 0), base, base)
	}
	insertJob(t, ctx, db, cid, users[50], delivery.ChannelSMS, delivery.StatusSent, 1, "", at(11, 50, 4, 999), base, base)
	insertJob(t, ctx, db, cid, users[51], delivery.ChannelSMS, delivery.StatusSent, 1, "", at(11, 50, 5, 0), base, base)
	insertJob(t, ctx, db, cid, users[52], delivery.ChannelSMS, delivery.StatusSent, 1, "", at(12, 0, 5, 0), base, base)

	st, err := s.Stats(ctx, *at(12, 0, 7, 0))
	if err != nil {
		t.Fatal(err)
	}
	tp := st.Throughput
	if want := *at(11, 50, 5, 0); !tp.Start.Equal(want) {
		t.Errorf("start = %v, want %v", tp.Start, want)
	}
	if tp.BucketSeconds != 5 || len(tp.SMS) != 120 || len(tp.Email) != 120 {
		t.Fatalf("shape = %d, %d, %d", tp.BucketSeconds, len(tp.SMS), len(tp.Email))
	}
	for i, v := range tp.SMS {
		want := 0.0
		switch i {
		case 59:
			want = 10
		case 0:
			want = 0.2
		}
		if math.Abs(v-want) > 1e-9 {
			t.Errorf("sms[%d] = %v, want %v", i, v, want)
		}
	}
	for i, v := range tp.Email {
		if v != 0 {
			t.Errorf("email[%d] = %v, want 0", i, v)
		}
	}
}

func TestStats_Failures(t *testing.T) {
	ctx, db, s := newDB(t)
	cid := insertCampaign(t, ctx, db, "completed", "fail")
	type pair struct {
		st       delivery.Status
		reason   string
		sms, eml int
	}
	pairs := []pair{
		{delivery.StatusFailed, "r10", 6, 4}, {delivery.StatusFailed, "r9", 9, 0}, {delivery.StatusFailed, "r8", 0, 8},
		{delivery.StatusSkipped, "r7", 7, 0}, {delivery.StatusSkipped, "r6", 3, 3}, {delivery.StatusUnknown, "timeout", 2, 3},
		{delivery.StatusUnknown, "r4", 4, 0}, {delivery.StatusSkipped, "r3", 0, 3},
		{delivery.StatusFailed, "r2", 1, 1}, {delivery.StatusUnknown, "r1", 0, 1},
	}
	for _, p := range pairs {
		for _, u := range insertUsers(t, ctx, db, p.sms) {
			insertJob(t, ctx, db, cid, u, delivery.ChannelSMS, p.st, 1, p.reason, nil, base, base)
		}
		for _, u := range insertUsers(t, ctx, db, p.eml) {
			insertJob(t, ctx, db, cid, u, delivery.ChannelEmail, p.st, 1, p.reason, nil, base, base)
		}
	}
	addJobs(t, ctx, db, cid, delivery.ChannelSMS, delivery.StatusSent, 20)
	addJobs(t, ctx, db, cid, delivery.ChannelEmail, delivery.StatusQueued, 20)

	st, err := s.Stats(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	wantReasons := []string{"r10", "r9", "r8", "r7", "r6", "timeout", "r4", "r3"}
	if len(st.Failures) != 8 {
		t.Fatalf("failures = %+v, want 8 rows", st.Failures)
	}
	for i, f := range st.Failures {
		if f.Reason != wantReasons[i] {
			t.Errorf("row %d reason = %q, want %q", i, f.Reason, wantReasons[i])
		}
		if f.Status == "sent" || f.Status == "queued" {
			t.Errorf("row %d has status %q", i, f.Status)
		}
	}
	if f := st.Failures[5]; f.Status != "unknown" || f.SMS != 2 || f.Email != 3 {
		t.Errorf("timeout row = %+v, want unknown sms 2 email 3", f)
	}
}

func TestStats_MeanAttempts(t *testing.T) {
	ctx, db, s := newDB(t)
	cid := insertCampaign(t, ctx, db, "completed", "ma")
	for i, u := range insertUsers(t, ctx, db, 4) {
		attempts := []int{1, 1, 2, 5}[i]
		status := delivery.StatusSent
		if i == 3 {
			status = delivery.StatusFailed // does not count
		}
		insertJob(t, ctx, db, cid, u, delivery.ChannelSMS, status, attempts, "", nil, base, base)
	}
	st, err := s.Stats(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	got := st.MeanAttempts[delivery.ChannelSMS]
	if got == nil || math.Abs(*got-4.0/3) > 1e-9 {
		t.Errorf("sms mean attempts = %v, want 1.333", got)
	}
	if v, ok := st.MeanAttempts[delivery.ChannelEmail]; !ok || v != nil {
		t.Errorf("email mean attempts = %v (present %v), want present nil", v, ok)
	}
}

func TestStats_Campaigns(t *testing.T) {
	ctx, db, s := newDB(t)
	insertCampaign(t, ctx, db, "draft", "the draft")
	var ids []int64
	for i := 0; i < 22; i++ {
		ids = append(ids, insertCampaign(t, ctx, db, "approved", fmt.Sprintf("c%d", i)))
	}
	completed, queued := ids[21], ids[20]
	mustExec(t, ctx, db, `UPDATE campaigns SET status = 'completed' WHERE id = $1`, completed)
	mustExec(t, ctx, db, `UPDATE campaigns SET status = 'queued' WHERE id = $1`, queued)
	u := insertUsers(t, ctx, db, 4)
	t0 := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
	insertJob(t, ctx, db, completed, u[0], delivery.ChannelSMS, delivery.StatusSent, 1, "", nil, t0, t0.Add(time.Second))
	insertJob(t, ctx, db, completed, u[1], delivery.ChannelSMS, delivery.StatusFailed, 1, "x", nil, t0.Add(time.Second), t0.Add(3120*time.Millisecond))
	insertJob(t, ctx, db, completed, u[2], delivery.ChannelEmail, delivery.StatusUnknown, 1, "timeout", nil, t0.Add(time.Second), t0.Add(2*time.Second))
	insertJob(t, ctx, db, completed, u[3], delivery.ChannelEmail, delivery.StatusSkipped, 0, "no_consent", nil, t0.Add(time.Second), t0.Add(2*time.Second))
	insertJob(t, ctx, db, queued, u[0], delivery.ChannelSMS, delivery.StatusQueued, 0, "", nil, t0, t0.Add(time.Second))

	st, err := s.Stats(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Campaigns) != 20 {
		t.Fatalf("campaigns = %d, want 20", len(st.Campaigns))
	}
	for i, c := range st.Campaigns {
		if c.Description == "the draft" || c.Status == "draft" {
			t.Errorf("row %d is a draft", i)
		}
		if i > 0 && c.ID >= st.Campaigns[i-1].ID {
			t.Errorf("ids not descending at %d: %d after %d", i, c.ID, st.Campaigns[i-1].ID)
		}
	}
	byID := map[int64]CampaignRow{}
	for _, c := range st.Campaigns {
		byID[c.ID] = c
	}
	c := byID[completed]
	if c.Total != 4 || c.Sent != 1 || c.Failed != 1 || c.Unknown != 1 || c.Status != "completed" || c.Description != "c21" {
		t.Errorf("completed row = %+v", c)
	}
	if c.DurationMS == nil || *c.DurationMS != 3120 {
		t.Errorf("duration = %v, want 3120", c.DurationMS)
	}
	if q := byID[queued]; q.DurationMS != nil || q.Total != 1 {
		t.Errorf("queued row = %+v, want nil duration total 1", q)
	}
	if e := byID[ids[19]]; e.Total != 0 || e.Sent != 0 || e.DurationMS != nil {
		t.Errorf("empty row = %+v", e)
	}
}

func TestStats_DurationAfterRetry(t *testing.T) {
	ctx, db, s := newDB(t)
	cid := insertCampaign(t, ctx, db, "approved", "retry")
	mustExec(t, ctx, db, `UPDATE campaigns SET status = 'completed' WHERE id = $1`, cid)
	u := insertUsers(t, ctx, db, 1)
	insertJob(t, ctx, db, cid, u[0], delivery.ChannelSMS, delivery.StatusRetrying, 1, "", nil, base, base.Add(time.Second))
	mustExec(t, ctx, db, `UPDATE campaigns SET status = 'sending' WHERE id = $1`, cid)
	st, err := s.Stats(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Campaigns) != 1 || st.Campaigns[0].Status != "sending" || st.Campaigns[0].DurationMS != nil {
		t.Errorf("campaigns = %+v, want status sending and nil duration", st.Campaigns)
	}
}

func TestStats_Empty(t *testing.T) {
	ctx, db, s := newDB(t)
	_ = db
	st, err := s.Stats(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	for _, ch := range delivery.Channels {
		if st.Totals[ch] != (ChannelTotals{}) {
			t.Errorf("%s totals = %+v", ch, st.Totals[ch])
		}
		if v, ok := st.MeanAttempts[ch]; !ok || v != nil {
			t.Errorf("%s mean attempts = %v (present %v)", ch, v, ok)
		}
	}
	for _, series := range [][]float64{st.Throughput.SMS, st.Throughput.Email} {
		if len(series) != 120 {
			t.Fatalf("series length = %d", len(series))
		}
		for i, v := range series {
			if v != 0 {
				t.Errorf("value %d = %v", i, v)
			}
		}
	}
	if st.Failures == nil || len(st.Failures) != 0 {
		t.Errorf("failures = %#v, want empty non-nil", st.Failures)
	}
	if st.Campaigns == nil || len(st.Campaigns) != 0 {
		t.Errorf("campaigns = %#v, want empty non-nil", st.Campaigns)
	}
}

func TestInProgress(t *testing.T) {
	ctx, db, s := newDB(t)
	seedTotals(t, ctx, db)
	addJobs(t, ctx, db, insertCampaign(t, ctx, db, "queued", "more"), delivery.ChannelEmail, delivery.StatusQueued, 2)
	got, err := s.InProgress(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got[delivery.ChannelSMS] != 3 || got[delivery.ChannelEmail] != 2 {
		t.Errorf("in progress = %v, want sms 3 email 2", got)
	}
}

func TestCounts(t *testing.T) {
	ctx, db, s := newDB(t)
	seedTotals(t, ctx, db)
	insertCampaign(t, ctx, db, "draft", "d")
	jobs, camps, err := s.Counts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := groupBy(t, ctx, db)
	for ch, m := range want {
		for st, n := range m {
			if jobs[ch][st] != n {
				t.Errorf("jobs[%s][%s] = %d, want %d", ch, st, jobs[ch][st], n)
			}
		}
	}
	if camps["draft"] != 1 || camps["sending"] != 1 {
		t.Errorf("campaigns = %v, want draft 1 sending 1", camps)
	}
}
