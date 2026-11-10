package campaign

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/seed"
	"campaigncomposer/internal/segment"
)

const ankaraRule = `{"op":"and","conditions":[{"field":"city","op":"eq","value":"Ankara"}]}`

func seededService(t *testing.T) (context.Context, *pgxpool.Pool, *Service) {
	t.Helper()
	ctx, pool := emptyDB(t)
	o := seed.Options{Random: 42, Now: time.Now(), Users: 500}
	if _, err := seed.Load(ctx, pool, o, false); err != nil {
		t.Fatalf("seed: %v", err)
	}
	svc := &Service{
		Store: NewStore(pool),
		Allow: segment.Default(),
		Count: func(ctx context.Context, w segment.Where) (segment.Audience, error) {
			return segment.CountAudience(ctx, pool, w)
		},
		Sample: func(ctx context.Context, w segment.Where, langs []string, n int) (map[string][]string, error) {
			return segment.SampleFirstNames(ctx, pool, w, langs, n)
		},
		Now: time.Now,
	}
	return ctx, pool, svc
}

// trAudience returns the tr counts of rule.
func trAudience(t *testing.T, ctx context.Context, svc *Service, rule string) segment.LanguageCounts {
	t.Helper()
	r, errs := svc.Allow.ParseRule([]byte(rule))
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	aud, err := svc.Count(ctx, svc.Allow.Compile(r, time.Now()))
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range aud.Languages {
		if l.Code == "tr" {
			return l
		}
	}
	t.Fatal("no tr audience")
	return segment.LanguageCounts{}
}

func makeDraft(t *testing.T, ctx context.Context, svc *Service, rule, sms string) int64 {
	t.Helper()
	l := trAudience(t, ctx, svc, rule)
	id, err := svc.Store.CreateDraft(ctx, Draft{
		Brief: "brief", Description: "desc", Rule: json.RawMessage(rule),
		Contents: []LanguageContent{{Language: "tr", SMSCount: l.SMS, EmailCount: l.Email, FirstNameRunes: int(l.MaxFirstName),
			Content: contentOf(sms)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func body(sms string) json.RawMessage {
	b, _ := json.Marshal(map[string]any{"contents": []map[string]string{{
		"language": "tr", "sms": sms, "email_subject": "Konu {{first_name}}", "email_body": "Merhaba {{first_name}}",
	}}})
	return b
}

func TestService_SaveAndApprove(t *testing.T) {
	ctx, _, svc := seededService(t)
	id := makeDraft(t, ctx, svc, ankaraRule, "old")
	c, err := svc.Save(ctx, id, body("  new {{first_name}} "))
	if err != nil {
		t.Fatal(err)
	}
	if c.Status != StatusDraft || c.Contents[0].SMS != "new {{first_name}}" {
		t.Fatalf("after save: %+v", c)
	}
	got, _ := svc.Get(ctx, id)
	if got.Contents[0].SMS != "new {{first_name}}" {
		t.Fatalf("stored: %+v", got.Contents[0])
	}
	c, repeat, err := svc.Approve(ctx, id, body("final {{first_name}}"))
	if err != nil || repeat {
		t.Fatalf("approve: %v, repeat %v", err, repeat)
	}
	if c.Status != StatusApproved || c.ApprovedAt == nil || c.Contents[0].SMS != "final {{first_name}}" {
		t.Fatalf("after approve: %+v", c)
	}
	l := trAudience(t, ctx, svc, ankaraRule)
	if c.Contents[0].SMSCount != l.SMS || c.Contents[0].EmailCount != l.Email {
		t.Fatalf("counts %+v, want sms %d email %d", c.Contents[0], l.SMS, l.Email)
	}
}

func TestService_SaveInvalid(t *testing.T) {
	ctx, _, svc := seededService(t)
	id := makeDraft(t, ctx, svc, ankaraRule, "old")
	_, err := svc.Save(ctx, id, body("{{last_name}}"))
	var inv *InvalidError
	if !errors.As(err, &inv) || !hasPath(inv.Errors, "contents[0].sms") {
		t.Fatalf("err = %v, want InvalidError on contents[0].sms", err)
	}
}

func TestService_ApproveUsesNewLongestName(t *testing.T) {
	ctx, pool, svc := seededService(t)
	l := trAudience(t, ctx, svc, ankaraRule)
	m := int(l.MaxFirstName)
	sms := "{{first_name}}" + strings.Repeat("a", 160-m-1) // 159 runes rendered
	id := makeDraft(t, ctx, svc, ankaraRule, sms)
	if _, err := svc.Save(ctx, id, body(sms)); err != nil {
		t.Fatalf("save before the new user: %v", err)
	}
	name := strings.Repeat("Ş", m+2)
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, first_name,last_name,email,phone,language,country,city,birth_date,email_consent,sms_consent,created_at,order_count,total_spent)
VALUES (1000001, $1,'K','long@example.com','+905000000001','tr','TR','Ankara','1990-01-01',true,true,now(),0,0)`, name); err != nil {
		t.Fatal(err)
	}
	_, _, err := svc.Approve(ctx, id, body(sms))
	var inv *InvalidError
	if !errors.As(err, &inv) || !hasPath(inv.Errors, "contents[0].sms") {
		t.Fatalf("err = %v, want InvalidError on contents[0].sms", err)
	}
	c, _ := svc.Get(ctx, id)
	if c.Status != StatusDraft {
		t.Fatalf("status = %s, want draft", c.Status)
	}
}

func TestService_ApproveConcurrent(t *testing.T) {
	ctx, _, svc := seededService(t)
	id := makeDraft(t, ctx, svc, ankaraRule, "hi {{first_name}}")
	var wg sync.WaitGroup
	var mu sync.Mutex
	firsts := 0
	approvedAt := map[string]bool{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, repeat, err := svc.Approve(ctx, id, body(fmt.Sprintf("hi %d {{first_name}}", i)))
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				t.Errorf("approve: %v", err)
				return
			}
			if c.ID != id {
				t.Errorf("id = %d", c.ID)
			}
			if !repeat {
				firsts++
			}
			approvedAt[c.ApprovedAt.String()] = true
		}()
	}
	wg.Wait()
	if firsts != 1 || len(approvedAt) != 1 {
		t.Fatalf("first approvals = %d, approved_at values = %d; want 1 and 1", firsts, len(approvedAt))
	}
}

func TestService_SaveApproved(t *testing.T) {
	ctx, _, svc := seededService(t)
	id := makeDraft(t, ctx, svc, ankaraRule, "hi")
	if _, _, err := svc.Approve(ctx, id, body("hi")); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Save(ctx, id, body("hi")); !errors.Is(err, ErrNotDraft) {
		t.Fatalf("err = %v, want ErrNotDraft", err)
	}
}

func TestService_DBTimeout(t *testing.T) {
	ctx, _, svc := seededService(t)
	id := makeDraft(t, ctx, svc, ankaraRule, "hi")
	svc.Count = func(context.Context, segment.Where) (segment.Audience, error) {
		return segment.Audience{}, context.DeadlineExceeded
	}
	if _, _, err := svc.Approve(ctx, id, body("changed")); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}
	c, _ := svc.Get(ctx, id)
	if c.Status != StatusDraft || c.Contents[0].SMS != "hi" {
		t.Fatalf("campaign changed: %+v", c)
	}
}

func TestService_SampleNames(t *testing.T) {
	ctx, _, svc := seededService(t)
	id := makeDraft(t, ctx, svc, ankaraRule, "hi")
	got, err := svc.SampleNames(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Code != "tr" || len(got[0].FirstNames) == 0 || len(got[0].FirstNames) > 5 {
		t.Fatalf("sample = %+v", got)
	}
	if _, err := svc.SampleNames(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown id: %v", err)
	}
}

// With one pool connection, the row lock must not wait for the audience query.
func TestService_ApproveWithOneConnection(t *testing.T) {
	ctx, pool, svc := seededService(t)
	id := makeDraft(t, ctx, svc, ankaraRule, "hi {{first_name}}")
	cfg := pool.Config().Copy()
	cfg.MaxConns = 1
	one, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer one.Close()
	svc.Store = NewStore(one)
	svc.Count = func(ctx context.Context, w segment.Where) (segment.Audience, error) {
		return segment.CountAudience(ctx, one, w)
	}
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, _, err := svc.Approve(cctx, id, body("hi {{first_name}}")); err != nil {
		t.Fatalf("approve with one connection: %v", err)
	}
}

// A failed check stores the new audience snapshot, so the window can show the new limit.
func TestService_FailedApproveStoresNewLongestName(t *testing.T) {
	ctx, pool, svc := seededService(t)
	l := trAudience(t, ctx, svc, ankaraRule)
	m := int(l.MaxFirstName)
	sms := "{{first_name}}" + strings.Repeat("a", 160-m-1)
	id := makeDraft(t, ctx, svc, ankaraRule, sms)
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, first_name,last_name,email,phone,language,country,city,birth_date,email_consent,sms_consent,created_at,order_count,total_spent)
VALUES (1000001, $1,'K','long@example.com','+905000000001','tr','TR','Ankara','1990-01-01',true,true,now(),0,0)`, strings.Repeat("Ş", m+2)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Approve(ctx, id, body(sms)); err == nil {
		t.Fatal("want an error")
	}
	c, _ := svc.Get(ctx, id)
	if c.Status != StatusDraft || c.Contents[0].FirstNameRunes != m+2 || c.Contents[0].SMS != sms {
		t.Fatalf("after failed approve: status %s runes %d (want %d), sms changed %v", c.Status, c.Contents[0].FirstNameRunes, m+2, c.Contents[0].SMS != sms)
	}
}

// A repeat approval needs no audience query, so a database error there cannot turn it into a failure.
func TestService_RepeatApproveSkipsCount(t *testing.T) {
	ctx, _, svc := seededService(t)
	id := makeDraft(t, ctx, svc, ankaraRule, "hi")
	if _, _, err := svc.Approve(ctx, id, body("hi")); err != nil {
		t.Fatal(err)
	}
	svc.Count = func(context.Context, segment.Where) (segment.Audience, error) {
		return segment.Audience{}, context.DeadlineExceeded
	}
	if _, repeat, err := svc.Approve(ctx, id, body("hi")); err != nil || !repeat {
		t.Fatalf("repeat: %v %v", repeat, err)
	}
}
