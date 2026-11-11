package store_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"campaigncomposer/internal/store"
	"campaigncomposer/internal/store/storetest"
)

func migrated(t *testing.T) (context.Context, *pgx.Conn) {
	t.Helper()
	ctx := context.Background()
	url := storetest.NewDB(t)
	if err := store.Migrate(ctx, url); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { conn.Close(context.Background()) })
	return ctx, conn
}

const insertUser = `INSERT INTO users (first_name,last_name,email,phone,language,country,city,birth_date,email_consent,sms_consent,created_at,order_count,total_spent)
VALUES ('A','B',$1,'+905000000000',$2,$3,'Istanbul','1990-01-01',true,true,now(),0,0)`

func TestMigrate_CreatesTables(t *testing.T) {
	ctx, conn := migrated(t)
	for _, tbl := range []string{"users", "events"} {
		var reg *string
		if err := conn.QueryRow(ctx, "SELECT to_regclass($1)::text", "public."+tbl).Scan(&reg); err != nil {
			t.Fatal(err)
		}
		if reg == nil {
			t.Errorf("table %s missing", tbl)
		}
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	ctx := context.Background()
	url := storetest.NewDB(t)
	if err := store.Migrate(ctx, url); err != nil {
		t.Fatal(err)
	}
	if err := store.Migrate(ctx, url); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
}

func TestMigrate_RejectsBadValues(t *testing.T) {
	ctx, conn := migrated(t)
	if _, err := conn.Exec(ctx, insertUser, "ok@example.com", "tr", "TR"); err != nil {
		t.Fatalf("valid insert: %v", err)
	}
	if _, err := conn.Exec(ctx, insertUser, "l@example.com", "de", "TR"); err == nil {
		t.Error("language de accepted")
	}
	if _, err := conn.Exec(ctx, insertUser, "c@example.com", "en", "GB"); err == nil {
		t.Error("country GB accepted")
	}
	if _, err := conn.Exec(ctx, `INSERT INTO events (user_id,type,amount,occurred_at)
SELECT id,'cart_add',10,now() FROM users LIMIT 1`); err == nil {
		t.Error("cart_add with amount accepted")
	}
}

func TestMigrate_CampaignTriggers(t *testing.T) {
	ctx, conn := migrated(t)
	var id int64
	if err := conn.QueryRow(ctx, `INSERT INTO campaigns (brief, rule, description, status, created_at, updated_at)
VALUES ('b', '{"op":"and","conditions":[]}', 'd', 'draft', now(), now()) RETURNING id`).Scan(&id); err != nil {
		t.Fatalf("insert campaign: %v", err)
	}
	insContent := `INSERT INTO campaign_contents (campaign_id, language, sms, email_subject, email_body, sms_count, email_count, first_name_runes)
VALUES ($1, $2, 's', 'subj', 'body', 1, 1, 5)`
	if _, err := conn.Exec(ctx, insContent, id, "tr"); err != nil {
		t.Fatalf("insert content: %v", err)
	}
	if _, err := conn.Exec(ctx, `UPDATE campaign_contents SET sms = 's2' WHERE campaign_id = $1`, id); err != nil {
		t.Fatalf("update draft content: %v", err)
	}
	if _, err := conn.Exec(ctx, `UPDATE campaigns SET status = 'approved', approved_at = now() WHERE id = $1`, id); err != nil {
		t.Fatalf("approve: %v", err)
	}
	fails := []struct {
		name string
		sql  string
		args []any
	}{
		{"update sms", `UPDATE campaign_contents SET sms = 'x' WHERE campaign_id = $1`, []any{id}},
		{"delete content", `DELETE FROM campaign_contents WHERE campaign_id = $1`, []any{id}},
		{"insert content", insContent, []any{id, "en"}},
		{"update brief", `UPDATE campaigns SET brief = 'x' WHERE id = $1`, []any{id}},
		{"back to draft", `UPDATE campaigns SET status = 'draft' WHERE id = $1`, []any{id}},
	}
	for _, f := range fails {
		if _, err := conn.Exec(ctx, f.sql, f.args...); err == nil {
			t.Errorf("%s: want error after approval, got nil", f.name)
		}
	}
}

// newCampaign inserts a campaign with the given status. Insert does not fire
// the update trigger, so any status can be the start of a move.
func newCampaign(t *testing.T, ctx context.Context, conn *pgx.Conn, status string) int64 {
	t.Helper()
	approved := "now()"
	if status == "draft" {
		approved = "NULL"
	}
	var id int64
	err := conn.QueryRow(ctx, `INSERT INTO campaigns (brief, rule, description, status, created_at, updated_at, approved_at)
VALUES ('b', '{"op":"and","conditions":[]}', 'd', $1, now(), now(), `+approved+`) RETURNING id`, status).Scan(&id)
	if err != nil {
		t.Fatalf("insert campaign %s: %v", status, err)
	}
	return id
}

func TestMigrate_CampaignStatusMoves(t *testing.T) {
	ctx, conn := migrated(t)
	allowed := [][2]string{
		{"draft", "approved"}, {"approved", "queued"}, {"approved", "completed"},
		{"queued", "sending"}, {"queued", "completed"}, {"sending", "completed"},
		{"completed", "sending"},
	}
	for _, m := range allowed {
		id := newCampaign(t, ctx, conn, m[0])
		set := "status = $2"
		if m[0] == "draft" {
			set += ", approved_at = now()"
		}
		if _, err := conn.Exec(ctx, `UPDATE campaigns SET `+set+` WHERE id = $1`, id, m[1]); err != nil {
			t.Errorf("%s -> %s: want ok, got %v", m[0], m[1], err)
		}
	}
	denied := [][2]string{
		{"queued", "approved"}, {"sending", "queued"}, {"completed", "approved"}, {"sending", "draft"},
		{"approved", "sending"}, {"completed", "queued"},
	}
	for _, m := range denied {
		id := newCampaign(t, ctx, conn, m[0])
		if _, err := conn.Exec(ctx, `UPDATE campaigns SET status = $2 WHERE id = $1`, id, m[1]); err == nil {
			t.Errorf("%s -> %s: want error, got nil", m[0], m[1])
		}
	}
	if _, err := conn.Exec(ctx, `INSERT INTO campaigns (brief, rule, description, status, created_at, updated_at)
VALUES ('b', '{}', 'd', 'queued', now(), now())`); err == nil {
		t.Error("queued with approved_at null accepted")
	}
}

func TestMigrate_FrozenAfterApprovalStill(t *testing.T) {
	ctx, conn := migrated(t)
	id := newCampaign(t, ctx, conn, "sending")
	for _, q := range []string{
		`UPDATE campaigns SET brief = 'x' WHERE id = $1`,
		`UPDATE campaigns SET rule = '{"op":"or","conditions":[]}' WHERE id = $1`,
		`UPDATE campaigns SET approved_at = now() + interval '1 day' WHERE id = $1`,
	} {
		if _, err := conn.Exec(ctx, q, id); err == nil {
			t.Errorf("want error for %s", q)
		}
	}
	// Content rows can only exist for a draft, so insert one before approval.
	draft := newCampaign(t, ctx, conn, "draft")
	if _, err := conn.Exec(ctx, `INSERT INTO campaign_contents (campaign_id, language, sms, email_subject, email_body, sms_count, email_count, first_name_runes)
VALUES ($1, 'tr', 's', 'subj', 'body', 1, 1, 5)`, draft); err != nil {
		t.Fatal(err)
	}
	for _, to := range []string{"approved", "queued", "sending"} {
		if _, err := conn.Exec(ctx, `UPDATE campaigns SET status = $2, approved_at = coalesce(approved_at, now()) WHERE id = $1`, draft, to); err != nil {
			t.Fatalf("move to %s: %v", to, err)
		}
	}
	if _, err := conn.Exec(ctx, `UPDATE campaign_contents SET sms = 'x' WHERE campaign_id = $1`, draft); err == nil {
		t.Error("content update accepted while sending")
	}
}

func TestMigrate_SendJobsUnique(t *testing.T) {
	ctx, conn := migrated(t)
	id := newCampaign(t, ctx, conn, "queued")
	var uid int64
	if err := conn.QueryRow(ctx, insertUser+` RETURNING id`, "u@example.com", "tr", "TR").Scan(&uid); err != nil {
		t.Fatal(err)
	}
	ins := `INSERT INTO send_jobs (campaign_id, user_id, channel, language, status, created_at, updated_at)
VALUES ($1, $2, $3, 'tr', 'queued', now(), now())`
	if _, err := conn.Exec(ctx, ins, id, uid, "sms"); err != nil {
		t.Fatal(err)
	}
	_, err := conn.Exec(ctx, ins, id, uid, "sms")
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Errorf("duplicate: want unique violation, got %v", err)
	}
	if _, err := conn.Exec(ctx, ins, id, uid, "email"); err != nil {
		t.Errorf("other channel: %v", err)
	}
}

// newJob inserts a send job with the given status. Insert does not fire the
// update trigger, so any status can be the start of a move.
func newJob(t *testing.T, ctx context.Context, conn *pgx.Conn, cid int64, email, status string, retryable bool) int64 {
	t.Helper()
	var uid, id int64
	if err := conn.QueryRow(ctx, insertUser+` RETURNING id`, email, "tr", "TR").Scan(&uid); err != nil {
		t.Fatal(err)
	}
	if err := conn.QueryRow(ctx, `INSERT INTO send_jobs (campaign_id, user_id, channel, language, status, retryable, created_at, updated_at)
VALUES ($1, $2, 'sms', 'tr', $3, $4, now(), now()) RETURNING id`, cid, uid, status, retryable).Scan(&id); err != nil {
		t.Fatalf("insert job %s: %v", status, err)
	}
	return id
}

func TestMigrate_SendJobStatusMoves(t *testing.T) {
	ctx, conn := migrated(t)
	cid := newCampaign(t, ctx, conn, "sending")
	n := 0
	job := func(status string, retryable bool) int64 {
		n++
		return newJob(t, ctx, conn, cid, fmt.Sprintf("j%d@example.com", n), status, retryable)
	}
	allowed := [][2]string{
		{"queued", "sending"}, {"retrying", "sending"},
		{"sending", "queued"}, {"sending", "sent"}, {"sending", "retrying"}, {"sending", "failed"},
		{"sending", "unknown"}, {"sending", "skipped"},
	}
	for _, m := range allowed {
		id := job(m[0], false)
		if _, err := conn.Exec(ctx, `UPDATE send_jobs SET status = $2 WHERE id = $1`, id, m[1]); err != nil {
			t.Errorf("%s -> %s: want ok, got %v", m[0], m[1], err)
		}
	}
	retry := job("failed", true)
	if _, err := conn.Exec(ctx, `UPDATE send_jobs SET status = 'queued', retryable = false WHERE id = $1`, retry); err != nil {
		t.Errorf("retryable failed -> queued: want ok, got %v", err)
	}
	denied := [][2]string{
		{"queued", "sent"}, {"queued", "failed"}, {"queued", "retrying"}, {"retrying", "queued"},
		{"sent", "queued"}, {"sent", "sending"}, {"sent", "retrying"}, {"unknown", "queued"},
		{"unknown", "sending"}, {"skipped", "queued"}, {"failed", "sending"},
	}
	for _, m := range denied {
		id := job(m[0], false)
		if _, err := conn.Exec(ctx, `UPDATE send_jobs SET status = $2 WHERE id = $1`, id, m[1]); err == nil {
			t.Errorf("%s -> %s: want error, got nil", m[0], m[1])
		}
	}
	permanent := job("failed", false)
	if _, err := conn.Exec(ctx, `UPDATE send_jobs SET status = 'queued' WHERE id = $1`, permanent); err == nil {
		t.Error("failed with retryable = false -> queued: want error, got nil")
	}
	// A write that keeps the status is not a move.
	sent := job("sent", false)
	if _, err := conn.Exec(ctx, `UPDATE send_jobs SET enqueued_at = now() WHERE id = $1`, sent); err != nil {
		t.Errorf("same status: want ok, got %v", err)
	}
}
