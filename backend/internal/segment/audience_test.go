package segment

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/seed"
	"campaigncomposer/internal/store"
	"campaigncomposer/internal/store/storetest"
)

func seedDB(t *testing.T) (context.Context, *pgxpool.Pool, []seed.User) {
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
	o := seed.Options{Random: 42, Now: testNow, Users: 500}
	users, _ := seed.Generate(o)
	if _, err := seed.Load(ctx, pool, o, false); err != nil {
		t.Fatalf("load: %v", err)
	}
	return ctx, pool, users
}

func compile(t *testing.T, js string) Where {
	t.Helper()
	return Default().Compile(mustRule(t, js), testNow)
}

// goCounts counts the users that match, by the Go predicate.
func goCounts(users []seed.User, match func(seed.User) bool) (Audience, map[string]Counts) {
	byLang := map[string]Counts{}
	var all Counts
	add := func(c *Counts, u seed.User) {
		c.Matched++
		if u.EmailConsent {
			c.Email++
		}
		if u.SMSConsent {
			c.SMS++
		}
		if u.EmailConsent || u.SMSConsent {
			c.Reachable++
		}
	}
	for _, u := range users {
		if !match(u) {
			continue
		}
		add(&all, u)
		c := byLang[u.Language]
		add(&c, u)
		byLang[u.Language] = c
	}
	all.NoConsent = all.Matched - all.Reachable
	for k, c := range byLang {
		c.NoConsent = c.Matched - c.Reachable
		byLang[k] = c
	}
	return Audience{Counts: all}, byLang
}

func within(t *time.Time, days int) bool {
	return t != nil && !t.Before(testNow.AddDate(0, 0, -days))
}

func olderThan(t *time.Time, days int) bool {
	return t != nil && t.Before(testNow.AddDate(0, 0, -days))
}

func yearsAgo(n int) time.Time {
	return time.Date(testNow.Year()-n, testNow.Month(), testNow.Day(), 0, 0, 0, 0, time.UTC)
}

func TestCountAudience_MatchesGo(t *testing.T) {
	ctx, pool, users := seedDB(t)
	tests := []struct {
		name  string
		rule  string
		match func(seed.User) bool
	}{
		{"example", exampleRule, func(u seed.User) bool {
			return u.City == "Istanbul" && within(u.LastCartAddAt, 30) &&
				(u.LastPurchaseAt == nil || olderThan(u.LastPurchaseAt, 30))
		}},
		{"or group", `{"op":"or","conditions":[
			{"field":"city","op":"eq","value":"Ankara"},
			{"field":"order_count","op":"gte","value":3}]}`, func(u seed.User) bool {
			return u.City == "Ankara" || u.OrderCount >= 3
		}},
		{"age between", `{"op":"and","conditions":[{"field":"age","op":"between","value":[25,40]}]}`,
			func(u seed.User) bool {
				return !u.BirthDate.After(yearsAgo(25)) && u.BirthDate.After(yearsAgo(41))
			}},
		{"empty root", `{"op":"and","conditions":[]}`, func(seed.User) bool { return true }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CountAudience(ctx, pool, compile(t, tc.rule))
			if err != nil {
				t.Fatal(err)
			}
			want, byLang := goCounts(users, tc.match)
			if got.Counts != want.Counts {
				t.Errorf("counts = %+v, want %+v", got.Counts, want.Counts)
			}
			if got.NoConsent != got.Matched-got.Reachable {
				t.Errorf("no_consent %d != matched-reachable", got.NoConsent)
			}
			for _, l := range got.Languages {
				if l.Counts != byLang[l.Code] {
					t.Errorf("lang %s = %+v, want %+v", l.Code, l.Counts, byLang[l.Code])
				}
				if l.NoConsent != l.Matched-l.Reachable {
					t.Errorf("lang %s no_consent mismatch", l.Code)
				}
			}
			wantLangs := 0
			for _, c := range byLang {
				if c.Reachable > 0 {
					wantLangs++
				}
			}
			if len(got.Languages) != wantLangs {
				t.Errorf("languages = %d, want %d", len(got.Languages), wantLangs)
			}
		})
	}
}

func TestCountAudience_Languages(t *testing.T) {
	ctx, pool, _ := seedDB(t)
	got, err := CountAudience(ctx, pool, compile(t,
		`{"op":"and","conditions":[{"field":"country","op":"in","value":["TR","US"]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Languages) != 2 || got.Languages[0].Code != "tr" || got.Languages[1].Code != "en" {
		t.Errorf("languages = %+v, want tr then en", got.Languages)
	}
	got, err = CountAudience(ctx, pool, compile(t, exampleRule))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Languages) != 1 || got.Languages[0].Code != "tr" {
		t.Errorf("example languages = %+v, want only tr", got.Languages)
	}
	got, err = CountAudience(ctx, pool, compile(t, `{"op":"and","conditions":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.NoConsent < 1 {
		t.Errorf("no_consent = %d, want >= 1", got.NoConsent)
	}
}

func TestBuildAudience(t *testing.T) {
	rows := []LanguageCounts{
		{Code: "en", Counts: Counts{Email: 1, SMS: 1, Reachable: 2, Matched: 3, NoConsent: 1}},
		{Code: "xx", Counts: Counts{Matched: 4, NoConsent: 4}},
		{Code: "tr", Counts: Counts{Email: 2, SMS: 2, Reachable: 2, Matched: 2}},
		{Code: "de", Counts: Counts{Email: 5, SMS: 0, Reachable: 5, Matched: 5}},
	}
	got := buildAudience(rows)
	var codes []string
	for _, l := range got.Languages {
		codes = append(codes, l.Code)
	}
	if want := []string{"de", "en", "tr"}; !reflect.DeepEqual(codes, want) {
		t.Errorf("order = %v, want %v", codes, want)
	}
	want := Counts{Email: 8, SMS: 3, Reachable: 9, Matched: 14, NoConsent: 5}
	if got.Counts != want {
		t.Errorf("totals = %+v, want %+v", got.Counts, want)
	}
	empty := buildAudience(nil)
	if empty.Languages == nil || len(empty.Languages) != 0 || empty.Counts != (Counts{}) {
		t.Errorf("empty = %+v", empty)
	}
}

func TestCountAudience_Empty(t *testing.T) {
	ctx, pool, _ := seedDB(t)
	got, err := CountAudience(ctx, pool, compile(t,
		`{"op":"and","conditions":[{"field":"total_spent","op":"gte","value":9999999}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.Counts != (Counts{}) || got.Languages == nil || len(got.Languages) != 0 {
		t.Errorf("got %+v, want zero counts and empty non-nil languages", got)
	}
}

func TestRecipients_Consent(t *testing.T) {
	ctx, pool, users := seedDB(t)
	w := compile(t, `{"op":"and","conditions":[]}`)
	byID := map[int64]seed.User{}
	for _, u := range users {
		byID[u.ID] = u
	}
	for _, tc := range []struct {
		name string
		ch   Channel
		ok   func(seed.User) bool
	}{
		{"sms", SMS(), func(u seed.User) bool { return u.SMSConsent }},
		{"email", Email(), func(u seed.User) bool { return u.EmailConsent }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rs, err := Recipients(ctx, pool, w, tc.ch)
			if err != nil {
				t.Fatal(err)
			}
			want := 0
			for _, u := range users {
				if tc.ok(u) {
					want++
				}
			}
			if len(rs) != want {
				t.Errorf("recipients = %d, want %d", len(rs), want)
			}
			var prev int64
			for _, r := range rs {
				u, ok := byID[r.ID]
				if !ok || !tc.ok(u) {
					t.Fatalf("recipient %d has no consent", r.ID)
				}
				if r.Email == "edge.noconsent@example.com" {
					t.Error("no-consent user returned")
				}
				if r.ID <= prev {
					t.Error("not ordered by id")
				}
				prev = r.ID
			}
		})
	}
	if _, err := Recipients(ctx, pool, w, Channel{}); err != ErrInvalidChannel {
		t.Errorf("zero channel err = %v, want ErrInvalidChannel", err)
	}
	if _, err := Recipients(ctx, pool, w, Channel{column: "id > 0 OR true"}); err != ErrInvalidChannel {
		t.Errorf("unknown column err = %v, want ErrInvalidChannel", err)
	}
}

func TestCountAudience_Timeout(t *testing.T) {
	ctx, pool, _ := seedDB(t)
	old := statementTimeout
	statementTimeout = "1ms"
	t.Cleanup(func() { statementTimeout = old })
	w := Where{sql: "(SELECT true FROM pg_sleep(0.1))"}
	if _, err := CountAudience(ctx, pool, w); err == nil {
		t.Fatal("want timeout error")
	}
}
