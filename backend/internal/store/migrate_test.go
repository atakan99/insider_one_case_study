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
