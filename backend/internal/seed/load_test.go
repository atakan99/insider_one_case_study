package seed_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/seed"
	"campaigncomposer/internal/store"
	"campaigncomposer/internal/store/storetest"
)

var testNow = time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)

func setup(t *testing.T) (context.Context, *pgxpool.Pool) {
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
	return ctx, pool
}

func count(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func opts(random uint64, users int) seed.Options {
	return seed.Options{Random: random, Now: testNow, Users: users}
}

func TestLoad_InsertsAll(t *testing.T) {
	ctx, pool := setup(t)
	o := opts(42, 200)
	users, events := seed.Generate(o)
	inserted, err := seed.Load(ctx, pool, o, false)
	if err != nil || !inserted {
		t.Fatalf("Load = %v, %v; want true, nil", inserted, err)
	}
	if got := count(t, ctx, pool, "users"); got != len(users) {
		t.Errorf("users = %d, want %d", got, len(users))
	}
	if got := count(t, ctx, pool, "events"); got != len(events) {
		t.Errorf("events = %d, want %d", got, len(events))
	}
}

func TestLoad_SkipsWhenUsersExist(t *testing.T) {
	ctx, pool := setup(t)
	o := opts(42, 100)
	if _, err := seed.Load(ctx, pool, o, false); err != nil {
		t.Fatal(err)
	}
	u0, e0 := count(t, ctx, pool, "users"), count(t, ctx, pool, "events")
	inserted, err := seed.Load(ctx, pool, opts(43, 50), false)
	if err != nil || inserted {
		t.Fatalf("Load = %v, %v; want false, nil", inserted, err)
	}
	if u, e := count(t, ctx, pool, "users"), count(t, ctx, pool, "events"); u != u0 || e != e0 {
		t.Errorf("counts changed: users %d->%d, events %d->%d", u0, u, e0, e)
	}
}

func TestLoad_Reset(t *testing.T) {
	ctx, pool := setup(t)
	if _, err := seed.Load(ctx, pool, opts(42, 100), false); err != nil {
		t.Fatal(err)
	}
	var email1 string
	if err := pool.QueryRow(ctx, "SELECT email FROM users WHERE id = 1").Scan(&email1); err != nil {
		t.Fatal(err)
	}
	o := opts(43, 150)
	users, events := seed.Generate(o)
	inserted, err := seed.Load(ctx, pool, o, true)
	if err != nil || !inserted {
		t.Fatalf("Load = %v, %v; want true, nil", inserted, err)
	}
	var email2 string
	if err := pool.QueryRow(ctx, "SELECT email FROM users WHERE id = 1").Scan(&email2); err != nil {
		t.Fatal(err)
	}
	if email1 == email2 {
		t.Errorf("user 1 email did not change: %s", email1)
	}
	if got := count(t, ctx, pool, "users"); got != len(users) {
		t.Errorf("users = %d, want %d", got, len(users))
	}
	if got := count(t, ctx, pool, "events"); got != len(events) {
		t.Errorf("events = %d, want %d", got, len(events))
	}
}

func TestLoad_SummariesMatchEvents(t *testing.T) {
	ctx, pool := setup(t)
	if _, err := seed.Load(ctx, pool, opts(42, 500), false); err != nil {
		t.Fatal(err)
	}
	var bad int
	err := pool.QueryRow(ctx, `
SELECT count(*) FROM users u WHERE
  u.last_cart_add_at IS DISTINCT FROM (SELECT max(occurred_at) FROM events e WHERE e.user_id = u.id AND e.type = 'cart_add')
  OR u.last_purchase_at IS DISTINCT FROM (SELECT max(occurred_at) FROM events e WHERE e.user_id = u.id AND e.type = 'purchase')
  OR u.order_count <> (SELECT count(*) FROM events e WHERE e.user_id = u.id AND e.type = 'purchase')
  OR u.total_spent <> (SELECT coalesce(sum(amount), 0) FROM events e WHERE e.user_id = u.id AND e.type = 'purchase')`).Scan(&bad)
	if err != nil {
		t.Fatal(err)
	}
	if bad != 0 {
		t.Errorf("%d users have summaries that differ from events", bad)
	}
}

func TestLoad_SequenceAfterCopy(t *testing.T) {
	ctx, pool := setup(t)
	if _, err := seed.Load(ctx, pool, opts(42, 100), false); err != nil {
		t.Fatal(err)
	}
	var uid int64
	err := pool.QueryRow(ctx, `INSERT INTO users (first_name, last_name, email, phone, language, country, city,
  birth_date, email_consent, sms_consent, created_at, order_count, total_spent)
  VALUES ('A','B','seq.test@example.com','+10000000000','en','US','Austin','1990-01-01',true,true,now(),0,0)
  RETURNING id`).Scan(&uid)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO events (user_id, type, occurred_at) VALUES ($1, 'cart_add', now())`, uid); err != nil {
		t.Fatalf("insert event: %v", err)
	}
}

func TestLoad_ExampleBriefSQL(t *testing.T) {
	ctx, pool := setup(t)
	if _, err := seed.Load(ctx, pool, opts(42, 500), false); err != nil {
		t.Fatal(err)
	}
	var n int
	err := pool.QueryRow(ctx, `
SELECT count(*) FROM users
WHERE city = 'Istanbul' AND language = 'tr'
  AND last_cart_add_at >= $1::timestamptz - interval '30 days'
  AND (last_purchase_at IS NULL OR last_purchase_at < $1::timestamptz - interval '30 days')`, testNow).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	if n < 1 {
		t.Errorf("example brief rule selected %d users, want at least 1", n)
	}
}

func TestLoad_Speed(t *testing.T) {
	ctx, pool := setup(t)
	start := time.Now()
	if _, err := seed.Load(ctx, pool, opts(42, 5000), false); err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d > 10*time.Second {
		t.Errorf("Load took %v, want less than 10s", d)
	}
}

func TestLoad_AmountScale(t *testing.T) {
	ctx, pool := setup(t)
	o := opts(42, 200)
	users, events := seed.Generate(o)
	if _, err := seed.Load(ctx, pool, o, false); err != nil {
		t.Fatal(err)
	}
	var purchase *seed.Event
	for i := range events {
		if events[i].AmountKurus != nil {
			purchase = &events[i]
			break
		}
	}
	if purchase == nil {
		t.Fatal("no purchase event generated")
	}
	var got int64
	if err := pool.QueryRow(ctx, `SELECT (amount * 100)::bigint FROM events WHERE id = $1`, purchase.ID).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != *purchase.AmountKurus {
		t.Errorf("amount * 100 = %d, want %d", got, *purchase.AmountKurus)
	}
	var spender *seed.User
	for i := range users {
		if users[i].TotalSpentKurus > 0 {
			spender = &users[i]
			break
		}
	}
	if spender == nil {
		t.Fatal("no user with purchases")
	}
	if err := pool.QueryRow(ctx, `SELECT (total_spent * 100)::bigint FROM users WHERE id = $1`, spender.ID).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != spender.TotalSpentKurus {
		t.Errorf("total_spent * 100 = %d, want %d", got, spender.TotalSpentKurus)
	}
}
