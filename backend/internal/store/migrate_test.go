package store_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

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
