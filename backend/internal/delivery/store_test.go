package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/campaign"
	"campaigncomposer/internal/content"
	"campaigncomposer/internal/seed"
	"campaigncomposer/internal/segment"
	"campaigncomposer/internal/store"
	"campaigncomposer/internal/store/storetest"
)

var testNow = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

const exampleRule = `{"op":"and","conditions":[
 {"field":"city","op":"eq","value":"Istanbul"},
 {"field":"last_cart_add_at","op":"within_days","value":30},
 {"op":"or","conditions":[
  {"field":"last_purchase_at","op":"is_null"},
  {"field":"last_purchase_at","op":"older_than_days","value":30}]}]}`

const trUSRule = `{"op":"and","conditions":[{"field":"country","op":"in","value":["TR","US"]}]}`

func seeded(t *testing.T) (context.Context, *pgxpool.Pool, *Store) {
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
	if _, err := seed.Load(ctx, pool, seed.Options{Random: 42, Now: testNow, Users: 500}, false); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return ctx, pool, NewStore(pool)
}

// approved makes an approved campaign with one content row for each language.
func approved(t *testing.T, ctx context.Context, pool *pgxpool.Pool, rule string, langs ...string) int64 {
	t.Helper()
	cs := campaign.NewStore(pool)
	var lcs []campaign.LanguageContent
	for _, l := range langs {
		lcs = append(lcs, campaign.LanguageContent{Language: l, Content: content.Content{
			SMS: "Hi {{first_name}} " + l, EmailSubject: "Subject {{first_name}}", EmailBody: "Body {{first_name}} " + l}})
	}
	id, err := cs.CreateDraft(ctx, campaign.Draft{Brief: "b", Description: "d", Rule: json.RawMessage(rule), Contents: lcs})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := cs.Update(ctx, id, true, func(c campaign.Campaign) ([]campaign.LanguageContent, error) { return c.Contents, nil }); err != nil {
		t.Fatal(err)
	}
	return id
}

func planned(t *testing.T, ctx context.Context, pool *pgxpool.Pool, s *Store) (int64, map[Channel][]int64) {
	t.Helper()
	id := approved(t, ctx, pool, exampleRule, "tr")
	cid, ids, ok, err := s.PlanNext(ctx, segment.Default(), testNow)
	if err != nil || !ok || cid != id {
		t.Fatalf("PlanNext = %d, %v, %v, %v", cid, ok, err, id)
	}
	return cid, ids
}

func campaignStatus(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id int64) string {
	t.Helper()
	var st string
	if err := pool.QueryRow(ctx, `SELECT status FROM campaigns WHERE id = $1`, id).Scan(&st); err != nil {
		t.Fatal(err)
	}
	return st
}

func getJob(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id int64) Job {
	t.Helper()
	var j Job
	err := pool.QueryRow(ctx, `SELECT id, campaign_id, user_id, channel, language, status, attempts, last_error, retryable, updated_at
FROM send_jobs WHERE id = $1`, id).Scan(&j.ID, &j.CampaignID, &j.UserID, &j.Channel, &j.Language, &j.Status, &j.Attempts, &j.LastError, &j.Retryable, &j.UpdatedAt)
	if err != nil {
		t.Fatal(err)
	}
	return j
}

func exec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, sql, args...); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

func expectedCounts(t *testing.T, ctx context.Context, pool *pgxpool.Pool, rule string, langs ...string) (sms, email int64) {
	t.Helper()
	r, errs := segment.Default().ParseRule([]byte(rule))
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	aud, err := segment.CountAudience(ctx, pool, segment.Default().Compile(r, testNow))
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range aud.Languages {
		for _, w := range langs {
			if l.Code == w {
				sms += l.SMS
				email += l.Email
			}
		}
	}
	return
}

func TestStore_PlanExample(t *testing.T) {
	ctx, pool, s := seeded(t)
	cid, ids := planned(t, ctx, pool, s)
	sms, email := expectedCounts(t, ctx, pool, exampleRule, "tr")
	if sms == 0 || email == 0 {
		t.Fatalf("fixture has no audience: sms %d email %d", sms, email)
	}
	if int64(len(ids[ChannelSMS])) != sms || int64(len(ids[ChannelEmail])) != email {
		t.Fatalf("ids sms %d email %d, want %d and %d", len(ids[ChannelSMS]), len(ids[ChannelEmail]), sms, email)
	}
	var n int64
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM send_jobs WHERE campaign_id = $1 AND status = 'queued' AND enqueued_at IS NULL AND language = 'tr'`, cid).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != sms+email {
		t.Fatalf("queued rows %d, want %d", n, sms+email)
	}
	if st := campaignStatus(t, ctx, pool, cid); st != "queued" {
		t.Fatalf("campaign status %s", st)
	}
	if _, _, ok, err := s.PlanNext(ctx, segment.Default(), testNow); err != nil || ok {
		t.Fatalf("second PlanNext ok=%v err=%v", ok, err)
	}
}

func TestStore_PlanSkippedLanguage(t *testing.T) {
	ctx, pool, s := seeded(t)
	cid := approved(t, ctx, pool, trUSRule, "tr")
	_, ids, ok, err := s.PlanNext(ctx, segment.Default(), testNow)
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
	sms, email := expectedCounts(t, ctx, pool, trUSRule, "tr")
	allSMS, _ := expectedCounts(t, ctx, pool, trUSRule, "tr", "en")
	if allSMS == sms {
		t.Fatal("fixture has no en audience")
	}
	if int64(len(ids[ChannelSMS])) != sms || int64(len(ids[ChannelEmail])) != email {
		t.Fatalf("ids %d/%d want %d/%d", len(ids[ChannelSMS]), len(ids[ChannelEmail]), sms, email)
	}
	var n int64
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM send_jobs WHERE campaign_id = $1 AND language <> 'tr'`, cid).Scan(&n); err != nil || n != 0 {
		t.Fatalf("rows in other languages: %d, %v", n, err)
	}
}

func TestStore_PlanConcurrent(t *testing.T) {
	ctx, pool, s := seeded(t)
	cid := approved(t, ctx, pool, exampleRule, "tr")
	var wg sync.WaitGroup
	oks := make([]bool, 2)
	errs := make([]error, 2)
	for i := range oks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, oks[i], errs[i] = s.PlanNext(ctx, segment.Default(), testNow)
		}()
	}
	wg.Wait()
	if errs[0] != nil || errs[1] != nil {
		t.Fatal(errs)
	}
	if oks[0] == oks[1] {
		t.Fatalf("ok = %v, want exactly one true", oks)
	}
	sms, email := expectedCounts(t, ctx, pool, exampleRule, "tr")
	var n int64
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM send_jobs WHERE campaign_id = $1`, cid).Scan(&n); err != nil || n != sms+email {
		t.Fatalf("rows %d, want %d (%v)", n, sms+email, err)
	}
}

func TestStore_PlanNoJobs(t *testing.T) {
	ctx, pool, s := seeded(t)
	exec(t, ctx, pool, `UPDATE users SET email_consent = false, sms_consent = false`)
	cid := approved(t, ctx, pool, exampleRule, "tr")
	got, ids, ok, err := s.PlanNext(ctx, segment.Default(), testNow)
	if err != nil || !ok || got != cid {
		t.Fatalf("PlanNext = %d %v %v", got, ok, err)
	}
	if len(ids[ChannelSMS])+len(ids[ChannelEmail]) != 0 {
		t.Fatalf("ids = %v", ids)
	}
	if st := campaignStatus(t, ctx, pool, cid); st != "completed" {
		t.Fatalf("status %s", st)
	}
}

func TestStore_PlanBadRule(t *testing.T) {
	ctx, pool, s := seeded(t)
	cid := approved(t, ctx, pool, `{"op":"drop table","conditions":[]}`, "tr")
	got, ids, ok, err := s.PlanNext(ctx, segment.Default(), testNow)
	if err != nil || !ok || got != cid || len(ids[ChannelSMS])+len(ids[ChannelEmail]) != 0 {
		t.Fatalf("PlanNext = %d %v %v %v", got, ids, ok, err)
	}
	if st := campaignStatus(t, ctx, pool, cid); st != "completed" {
		t.Fatalf("status %s", st)
	}
}

func TestStore_ClaimOnce(t *testing.T) {
	ctx, pool, s := seeded(t)
	cid, ids := planned(t, ctx, pool, s)
	id := ids[ChannelSMS][0]
	var wg sync.WaitGroup
	oks := make([]bool, 2)
	errs := make([]error, 2)
	for i := range oks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, oks[i], errs[i] = s.Claim(ctx, id, testNow)
		}()
	}
	wg.Wait()
	if errs[0] != nil || errs[1] != nil || oks[0] == oks[1] {
		t.Fatalf("ok = %v err = %v", oks, errs)
	}
	j := getJob(t, ctx, pool, id)
	if j.Attempts != 1 || j.Status != StatusSending {
		t.Fatalf("job %+v", j)
	}
	if st := campaignStatus(t, ctx, pool, cid); st != "sending" {
		t.Fatalf("campaign %s", st)
	}
	// A claim on a sending job changes nothing.
	if _, ok, err := s.Claim(ctx, id, testNow); err != nil || ok {
		t.Fatalf("claim again: %v %v", ok, err)
	}
}

func TestStore_ClaimNotDue(t *testing.T) {
	ctx, pool, s := seeded(t)
	_, ids := planned(t, ctx, pool, s)
	id := ids[ChannelSMS][0]
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1`, id, testNow.Add(time.Minute))
	if _, ok, err := s.Claim(ctx, id, testNow); err != nil || ok {
		t.Fatalf("not due: ok=%v err=%v", ok, err)
	}
	j, ok, err := s.Claim(ctx, id, testNow.Add(2*time.Minute))
	if err != nil || !ok || j.Attempts != 2 || j.Channel != ChannelSMS || j.Status != StatusSending {
		t.Fatalf("due: %+v %v %v", j, ok, err)
	}
}

func TestStore_Release(t *testing.T) {
	ctx, pool, s := seeded(t)
	_, ids := planned(t, ctx, pool, s)
	id := ids[ChannelSMS][0]
	if _, ok, err := s.Claim(ctx, id, testNow); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if err := s.Release(ctx, id); err != nil {
		t.Fatal(err)
	}
	if j := getJob(t, ctx, pool, id); j.Status != StatusQueued || j.Attempts != 0 {
		t.Fatalf("job %+v", j)
	}
	if err := s.Release(ctx, id); err != nil {
		t.Fatal(err)
	}
	if j := getJob(t, ctx, pool, id); j.Attempts != 0 {
		t.Fatalf("second release changed attempts: %+v", j)
	}
}

func TestStore_FinishSent(t *testing.T) {
	ctx, pool, s := seeded(t)
	_, ids := planned(t, ctx, pool, s)
	id := ids[ChannelSMS][0]
	if _, ok, err := s.Claim(ctx, id, testNow); err != nil || !ok {
		t.Fatal(ok, err)
	}
	done := testNow.Add(time.Second)
	if err := s.Finish(ctx, id, Result{Status: StatusSent, ProviderID: "p-1"}, done); err != nil {
		t.Fatal(err)
	}
	var prov string
	var sentAt *time.Time
	var status string
	if err := pool.QueryRow(ctx, `SELECT provider_id, sent_at, status FROM send_jobs WHERE id = $1`, id).Scan(&prov, &sentAt, &status); err != nil {
		t.Fatal(err)
	}
	if prov != "p-1" || sentAt == nil || !sentAt.Equal(done) || status != "sent" {
		t.Fatalf("%q %v %s", prov, sentAt, status)
	}
	// A second Finish on a job that is not sending changes nothing and is no error.
	if err := s.Finish(ctx, id, Result{Status: StatusFailed, LastError: "x"}, done); err != nil {
		t.Fatal(err)
	}
	if j := getJob(t, ctx, pool, id); j.Status != StatusSent {
		t.Fatalf("job %+v", j)
	}
}

func TestStore_FinishRetryingAndFailed(t *testing.T) {
	ctx, pool, s := seeded(t)
	_, ids := planned(t, ctx, pool, s)
	a, b := ids[ChannelSMS][0], ids[ChannelSMS][1]
	for _, id := range []int64{a, b} {
		if _, ok, err := s.Claim(ctx, id, testNow); err != nil || !ok {
			t.Fatal(ok, err)
		}
	}
	next := testNow.Add(5 * time.Second)
	if err := s.Finish(ctx, a, Result{Status: StatusRetrying, LastError: "503", Retryable: true, NextAttemptAt: &next}, testNow); err != nil {
		t.Fatal(err)
	}
	if err := s.Finish(ctx, b, Result{Status: StatusFailed, LastError: "400"}, testNow); err != nil {
		t.Fatal(err)
	}
	var got *time.Time
	if err := pool.QueryRow(ctx, `SELECT next_attempt_at FROM send_jobs WHERE id = $1`, a).Scan(&got); err != nil || got == nil || !got.Equal(next) {
		t.Fatalf("next_attempt_at %v %v", got, err)
	}
	if j := getJob(t, ctx, pool, a); j.Status != StatusRetrying || j.LastError != "503" || !j.Retryable {
		t.Fatalf("%+v", j)
	}
	if j := getJob(t, ctx, pool, b); j.Status != StatusFailed || j.Retryable {
		t.Fatalf("%+v", j)
	}
}

func addUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, first, phone string, email, sms bool) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(ctx, `INSERT INTO users (first_name,last_name,email,phone,language,country,city,birth_date,email_consent,sms_consent,created_at,order_count,total_spent)
VALUES ($1,'B','x' || nextval('users_id_seq') || '@example.com',$2,'tr','TR','Istanbul','1990-01-01',$3,$4,now(),0,0) RETURNING id`,
		first, phone, email, sms).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func addJob(t *testing.T, ctx context.Context, pool *pgxpool.Pool, cid, uid int64, ch Channel, status string) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(ctx, `INSERT INTO send_jobs (campaign_id, user_id, channel, language, status, created_at, updated_at)
VALUES ($1,$2,$3,'tr',$4,$5,$5) RETURNING id`, cid, uid, string(ch), status, testNow).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestStore_LoadSkips(t *testing.T) {
	ctx, pool, s := seeded(t)
	cid := approved(t, ctx, pool, exampleRule, "tr")
	long := strings.Repeat("ş", 200)
	cases := []struct {
		name  string
		first string
		phone string
		email bool
		sms   bool
		ch    Channel
		want  string
	}{
		{"ok", "Ayşe", "+905001112233", true, true, ChannelSMS, ""},
		{"sms no consent", "Ayşe", "+905001112233", true, false, ChannelSMS, SkipNoConsent},
		{"email no consent", "Ayşe", "+905001112233", false, true, ChannelEmail, SkipNoConsent},
		{"no phone", "Ayşe", "", true, true, ChannelSMS, SkipNoContact},
		{"long sms name", long, "+905001112233", true, true, ChannelSMS, SkipOverLimit},
		{"long subject name", strings.Repeat("ş", 75), "+905001112233", true, true, ChannelEmail, SkipOverLimit},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			uid := addUser(t, ctx, pool, c.first, c.phone, c.email, c.sms)
			j := Job{ID: 1, CampaignID: cid, UserID: uid, Channel: c.ch, Language: "tr"}
			_, skip, err := s.Load(ctx, j)
			if err != nil || skip != c.want {
				t.Fatalf("skip = %q, err %v; want %q", skip, err, c.want)
			}
		})
	}
	// The long first name is set by SQL after insert, as in the brief.
	uid := addUser(t, ctx, pool, "x", "+905001112233", true, true)
	exec(t, ctx, pool, `UPDATE users SET first_name = $2 WHERE id = $1`, uid, long)
	if _, skip, err := s.Load(ctx, Job{CampaignID: cid, UserID: uid, Channel: ChannelSMS, Language: "tr"}); err != nil || skip != SkipOverLimit {
		t.Fatalf("skip %q err %v", skip, err)
	}
}

func TestStore_LoadRenders(t *testing.T) {
	ctx, pool, s := seeded(t)
	cid := approved(t, ctx, pool, exampleRule, "tr")
	uid := addUser(t, ctx, pool, "Ayşe", "+905001112233", true, true)
	m, skip, err := s.Load(ctx, Job{CampaignID: cid, UserID: uid, Channel: ChannelSMS, Language: "tr"})
	if err != nil || skip != "" || m.Body != "Hi Ayşe tr" || m.To != "+905001112233" || m.Subject != "" {
		t.Fatalf("sms %+v %q %v", m, skip, err)
	}
	m, skip, err = s.Load(ctx, Job{CampaignID: cid, UserID: uid, Channel: ChannelEmail, Language: "tr"})
	if err != nil || skip != "" || m.Subject != "Subject Ayşe" || m.Body != "Body Ayşe tr" || !strings.HasPrefix(m.To, "x") {
		t.Fatalf("email %+v %q %v", m, skip, err)
	}
	if _, _, err := s.Load(ctx, Job{CampaignID: cid, UserID: uid, Channel: ChannelEmail, Language: "xx"}); err == nil {
		t.Fatal("missing content gave no error")
	}
}

func TestStore_Recovery(t *testing.T) {
	ctx, pool, s := seeded(t)
	cid, ids := planned(t, ctx, pool, s)
	q := ids[ChannelSMS]
	// q[0] enqueued_at null, q[1] 200 s old, q[2] 10 s old; the rest are fresh.
	exec(t, ctx, pool, `UPDATE send_jobs SET enqueued_at = $1 WHERE campaign_id = $2`, testNow, cid)
	exec(t, ctx, pool, `UPDATE send_jobs SET enqueued_at = NULL WHERE id = $1`, q[0])
	exec(t, ctx, pool, `UPDATE send_jobs SET enqueued_at = $2 WHERE id = $1`, q[1], testNow.Add(-200*time.Second))
	exec(t, ctx, pool, `UPDATE send_jobs SET enqueued_at = $2 WHERE id = $1`, q[2], testNow.Add(-10*time.Second))
	got, err := s.StaleQueued(ctx, testNow, 120*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(got[ChannelSMS]) != 2 || len(got[ChannelEmail]) != 0 {
		t.Fatalf("stale = %v", got)
	}
	want := map[int64]bool{q[0]: true, q[1]: true}
	for _, id := range got[ChannelSMS] {
		if !want[id] {
			t.Fatalf("unexpected stale %d", id)
		}
	}
	if err := s.MarkEnqueued(ctx, []int64{q[0]}, testNow); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkEnqueued(ctx, nil, testNow); err != nil {
		t.Fatal(err)
	}
	got, _ = s.StaleQueued(ctx, testNow, 120*time.Second)
	if len(got[ChannelSMS]) != 1 || got[ChannelSMS][0] != q[1] {
		t.Fatalf("after mark: %v", got)
	}

	// Due retries.
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1`, q[3], testNow.Add(-time.Second))
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'retrying', attempts = 1, next_attempt_at = $2 WHERE id = $1`, q[4], testNow.Add(time.Minute))
	due, err := s.DueRetries(ctx, testNow, 0)
	if err != nil || len(due[ChannelSMS]) != 1 || due[ChannelSMS][0] != q[3] {
		t.Fatalf("due = %v %v", due, err)
	}
	if due, _ = s.DueRetries(ctx, testNow, 5*time.Second); len(due[ChannelSMS]) != 0 {
		t.Fatalf("lateBy ignored: %v", due)
	}

	// Expire sending.
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'sending', attempts = 1, updated_at = $2 WHERE id = $1`, q[5], testNow.Add(-40*time.Second))
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'sending', attempts = 1, updated_at = $2 WHERE id = $1`, q[6], testNow.Add(-10*time.Second))
	n, err := s.ExpireSending(ctx, testNow, 33*time.Second)
	if err != nil || n != 1 {
		t.Fatalf("expired %d %v", n, err)
	}
	if j := getJob(t, ctx, pool, q[5]); j.Status != StatusUnknown || j.Retryable || j.LastError == "" {
		t.Fatalf("%+v", j)
	}
	if j := getJob(t, ctx, pool, q[6]); j.Status != StatusSending {
		t.Fatalf("%+v", j)
	}

	// Complete campaigns.
	done, err := s.CompleteCampaigns(ctx)
	if err != nil || len(done) != 0 {
		t.Fatalf("early complete %v %v", done, err)
	}
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'sent' WHERE campaign_id = $1`, cid)
	done, err = s.CompleteCampaigns(ctx)
	if err != nil || len(done) != 1 || done[0] != cid {
		t.Fatalf("complete %v %v", done, err)
	}
	if st := campaignStatus(t, ctx, pool, cid); st != "completed" {
		t.Fatal(st)
	}
}

func TestStore_Progress(t *testing.T) {
	ctx, pool, s := seeded(t)
	cid, ids := planned(t, ctx, pool, s)
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'sent' WHERE id = ANY($1)`, ids[ChannelSMS][:3])
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'failed', retryable = true WHERE id = ANY($1)`, ids[ChannelSMS][3:5])
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'failed' WHERE id = $1`, ids[ChannelEmail][0])
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'skipped' WHERE id = $1`, ids[ChannelEmail][1])
	p, err := s.Progress(ctx, cid)
	if err != nil {
		t.Fatal(err)
	}
	if p.CampaignID != cid || p.Status != "queued" || len(p.Channels) != 2 {
		t.Fatalf("%+v", p)
	}
	for _, ch := range Channels {
		var want ChannelProgress
		rows, err := pool.Query(ctx, `SELECT status, count(*) FROM send_jobs WHERE campaign_id = $1 AND channel = $2 GROUP BY status`, cid, string(ch))
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var st string
			var n int64
			if err := rows.Scan(&st, &n); err != nil {
				t.Fatal(err)
			}
			want.Total += n
			switch Status(st) {
			case StatusQueued:
				want.Queued = n
			case StatusSent:
				want.Sent = n
			case StatusFailed:
				want.Failed = n
			case StatusSkipped:
				want.Skipped = n
			}
		}
		rows.Close()
		if ch == ChannelSMS {
			want.RetryableFailed = 2
		}
		if p.Channels[ch] != want {
			t.Fatalf("%s = %+v, want %+v", ch, p.Channels[ch], want)
		}
	}
	if _, err := s.Progress(ctx, 99999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown: %v", err)
	}
	draft, err := campaign.NewStore(pool).CreateDraft(ctx, campaign.Draft{Brief: "b", Description: "d", Rule: json.RawMessage(exampleRule)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Progress(ctx, draft); !errors.Is(err, ErrNotSendable) {
		t.Fatalf("draft: %v", err)
	}
	// An approved campaign with no jobs has both keys with zeros.
	ap := approved(t, ctx, pool, exampleRule, "tr")
	p, err = s.Progress(ctx, ap)
	if err != nil || len(p.Channels) != 2 || p.Channels[ChannelSMS].Total != 0 {
		t.Fatalf("approved: %+v %v", p, err)
	}
}

func TestStore_ListJobs(t *testing.T) {
	ctx, pool, s := seeded(t)
	cid, ids := planned(t, ctx, pool, s)
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'failed' WHERE id = ANY($1)`, ids[ChannelSMS][:4])
	jobs, next, err := s.ListJobs(ctx, cid, JobFilter{Status: StatusFailed, Channel: ChannelSMS, Limit: 3})
	if err != nil || len(jobs) != 3 || next == nil || *next != jobs[2].ID {
		t.Fatalf("page 1: %d %v %v", len(jobs), next, err)
	}
	for _, j := range jobs {
		if j.Status != StatusFailed || j.Channel != ChannelSMS || j.CampaignID != cid {
			t.Fatalf("%+v", j)
		}
	}
	jobs, next, err = s.ListJobs(ctx, cid, JobFilter{Status: StatusFailed, Channel: ChannelSMS, Limit: 3, After: *next})
	if err != nil || len(jobs) != 1 || next != nil {
		t.Fatalf("page 2: %d %v %v", len(jobs), next, err)
	}
	all, next, err := s.ListJobs(ctx, cid, JobFilter{Limit: 1000})
	if err != nil || next != nil || len(all) != len(ids[ChannelSMS])+len(ids[ChannelEmail]) {
		t.Fatalf("all: %d %v %v", len(all), next, err)
	}
	for i := 1; i < len(all); i++ {
		if all[i].ID <= all[i-1].ID {
			t.Fatal("not ordered by id")
		}
	}
}

func TestStore_Retry(t *testing.T) {
	ctx, pool, s := seeded(t)
	cid, ids := planned(t, ctx, pool, s)
	if _, err := s.Retry(ctx, cid, testNow); !errors.Is(err, ErrNotCompleted) {
		t.Fatalf("queued campaign: %v", err)
	}
	a, b, c := ids[ChannelSMS][0], ids[ChannelSMS][1], ids[ChannelEmail][0]
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'sent' WHERE campaign_id = $1`, cid)
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'failed', attempts = 3, retryable = true, last_error = '503', enqueued_at = $2 WHERE id = ANY($1)`, []int64{a, c}, testNow)
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'failed', attempts = 1, retryable = false, last_error = '400' WHERE id = $1`, b)
	d := ids[ChannelEmail][1]
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'unknown', attempts = 1 WHERE id = $1`, d)
	exec(t, ctx, pool, `UPDATE campaigns SET status = 'completed' WHERE id = $1`, cid)

	var wg sync.WaitGroup
	res := make([]map[Channel][]int64, 2)
	errs := make([]error, 2)
	for i := range res {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res[i], errs[i] = s.Retry(ctx, cid, testNow)
		}()
	}
	wg.Wait()
	okN := 0
	for i := range res {
		switch {
		case errs[i] == nil:
			okN++
			if len(res[i][ChannelSMS]) != 1 || res[i][ChannelSMS][0] != a || len(res[i][ChannelEmail]) != 1 || res[i][ChannelEmail][0] != c {
				t.Fatalf("retry ids %v", res[i])
			}
		case !errors.Is(errs[i], ErrNotCompleted):
			t.Fatalf("second retry: %v", errs[i])
		}
	}
	if okN != 1 {
		t.Fatalf("successes %d", okN)
	}
	if j := getJob(t, ctx, pool, a); j.Status != StatusQueued || j.Attempts != 0 || j.Retryable {
		t.Fatalf("%+v", j)
	}
	var enq, nxt *time.Time
	if err := pool.QueryRow(ctx, `SELECT enqueued_at, next_attempt_at FROM send_jobs WHERE id = $1`, a).Scan(&enq, &nxt); err != nil || enq != nil || nxt != nil {
		t.Fatalf("%v %v %v", enq, nxt, err)
	}
	if j := getJob(t, ctx, pool, b); j.Status != StatusFailed || j.Attempts != 1 {
		t.Fatalf("400 job changed: %+v", j)
	}
	if j := getJob(t, ctx, pool, d); j.Status != StatusUnknown {
		t.Fatalf("unknown job changed: %+v", j)
	}
	if st := campaignStatus(t, ctx, pool, cid); st != "sending" {
		t.Fatal(st)
	}

	// Nothing to retry.
	exec(t, ctx, pool, `UPDATE send_jobs SET status = 'sent' WHERE campaign_id = $1`, cid)
	exec(t, ctx, pool, `UPDATE campaigns SET status = 'completed' WHERE id = $1`, cid)
	if _, err := s.Retry(ctx, cid, testNow); !errors.Is(err, ErrNothingToRetry) {
		t.Fatalf("nothing: %v", err)
	}
	if _, err := s.Retry(ctx, 99999, testNow); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown: %v", err)
	}
}
