package campaign

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/store"
	"campaigncomposer/internal/store/storetest"
)

func emptyDB(t *testing.T) (context.Context, *pgxpool.Pool) {
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

const emptyRule = `{"op":"and","conditions":[]}`

func lc(lang, sms string) LanguageContent {
	return LanguageContent{
		Language:       lang,
		Content:        content.Content{SMS: sms, EmailSubject: "Subject " + lang, EmailBody: "Body " + lang},
		SMSCount:       10,
		EmailCount:     20,
		FirstNameRunes: 7,
	}
}

func draft(brief string, contents ...LanguageContent) Draft {
	return Draft{Brief: brief, Description: "desc", Rule: json.RawMessage(emptyRule), Skipped: []string{}, Contents: contents}
}

func TestStore_CreateAndGet(t *testing.T) {
	ctx, pool := emptyDB(t)
	s := NewStore(pool)
	d := draft("brief", lc("tr", "merhaba"), lc("en", "hello"))
	d.Skipped = []string{}
	id, err := s.CreateDraft(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	c, err := s.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != id || c.Status != StatusDraft || c.Brief != "brief" || c.Description != "desc" || c.ApprovedAt != nil {
		t.Fatalf("campaign = %+v", c)
	}
	if len(c.Contents) != 2 || c.Contents[0].Language != "en" || c.Contents[1].Language != "tr" {
		t.Fatalf("contents not sorted by language: %+v", c.Contents)
	}
	if c.Contents[1] != lc("tr", "merhaba") {
		t.Fatalf("tr content = %+v", c.Contents[1])
	}
	if c.SkippedLanguages == nil || len(c.SkippedLanguages) != 0 {
		t.Fatalf("skipped = %#v, want empty non-nil", c.SkippedLanguages)
	}
	var rule map[string]any
	if err := json.Unmarshal(c.Rule, &rule); err != nil || rule["op"] != "and" {
		t.Fatalf("rule = %s", c.Rule)
	}
}

func TestStore_GetNotFound(t *testing.T) {
	ctx, pool := emptyDB(t)
	if _, err := NewStore(pool).Get(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestStore_ReplaceDraft(t *testing.T) {
	ctx, pool := emptyDB(t)
	s := NewStore(pool)
	id, err := s.CreateDraft(ctx, draft("brief", lc("tr", "one"), lc("en", "two")))
	if err != nil {
		t.Fatal(err)
	}
	d := draft("brief", lc("tr", "new"))
	d.Skipped = []string{"en"}
	if err := s.ReplaceDraft(ctx, id, d); err != nil {
		t.Fatal(err)
	}
	c, err := s.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Contents) != 1 || c.Contents[0].SMS != "new" {
		t.Fatalf("contents = %+v", c.Contents)
	}
	if len(c.SkippedLanguages) != 1 || c.SkippedLanguages[0] != "en" {
		t.Fatalf("skipped = %v", c.SkippedLanguages)
	}
	var n int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM campaigns").Scan(&n); err != nil || n != 1 {
		t.Fatalf("campaign count = %d, %v", n, err)
	}
}

func TestStore_ReplaceDraftSameRuleOtherKeyOrder(t *testing.T) {
	ctx, pool := emptyDB(t)
	s := NewStore(pool)
	id, err := s.CreateDraft(ctx, draft("brief", lc("tr", "one")))
	if err != nil {
		t.Fatal(err)
	}
	d := draft("brief", lc("tr", "two"))
	d.Rule = json.RawMessage(`{"conditions":[],"op":"and"}`)
	if err := s.ReplaceDraft(ctx, id, d); err != nil {
		t.Fatalf("ReplaceDraft with reordered keys: %v", err)
	}
}

func TestStore_ReplaceDraftConflict(t *testing.T) {
	ctx, pool := emptyDB(t)
	s := NewStore(pool)
	id, err := s.CreateDraft(ctx, draft("brief", lc("tr", "one")))
	if err != nil {
		t.Fatal(err)
	}
	other := draft("other brief", lc("tr", "x"))
	if err := s.ReplaceDraft(ctx, id, other); !errors.Is(err, ErrConflict) {
		t.Errorf("other brief: err = %v, want ErrConflict", err)
	}
	otherRule := draft("brief", lc("tr", "x"))
	otherRule.Rule = json.RawMessage(`{"op":"or","conditions":[]}`)
	if err := s.ReplaceDraft(ctx, id, otherRule); !errors.Is(err, ErrConflict) {
		t.Errorf("other rule: err = %v, want ErrConflict", err)
	}
	if err := s.ReplaceDraft(ctx, 999, draft("brief", lc("tr", "x"))); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown id: err = %v, want ErrNotFound", err)
	}
	pass := func(c Campaign) ([]LanguageContent, error) { return c.Contents, nil }
	if _, _, err := s.Update(ctx, id, true, pass); err != nil {
		t.Fatal(err)
	}
	if err := s.ReplaceDraft(ctx, id, draft("brief", lc("tr", "x"))); !errors.Is(err, ErrConflict) {
		t.Errorf("approved: err = %v, want ErrConflict", err)
	}
}

func TestStore_UpdateSave(t *testing.T) {
	ctx, pool := emptyDB(t)
	s := NewStore(pool)
	id, err := s.CreateDraft(ctx, draft("brief", lc("tr", "one")))
	if err != nil {
		t.Fatal(err)
	}
	before, _ := s.Get(ctx, id)
	time.Sleep(5 * time.Millisecond)
	c, repeat, err := s.Update(ctx, id, false, func(c Campaign) ([]LanguageContent, error) {
		if c.ID != id || c.Status != StatusDraft {
			t.Errorf("check got %+v", c)
		}
		return []LanguageContent{lc("tr", "edited")}, nil
	})
	if err != nil || repeat {
		t.Fatalf("Update = %v, repeat %v", err, repeat)
	}
	if c.Status != StatusDraft || c.Contents[0].SMS != "edited" || !c.UpdatedAt.After(before.UpdatedAt) {
		t.Fatalf("after save: %+v", c)
	}
}

func TestStore_UpdateApprove(t *testing.T) {
	ctx, pool := emptyDB(t)
	s := NewStore(pool)
	id, err := s.CreateDraft(ctx, draft("brief", lc("tr", "one")))
	if err != nil {
		t.Fatal(err)
	}
	c, repeat, err := s.Update(ctx, id, true, func(Campaign) ([]LanguageContent, error) {
		return []LanguageContent{lc("tr", "final")}, nil
	})
	if err != nil || repeat {
		t.Fatalf("Update = %v, repeat %v", err, repeat)
	}
	if c.Status != StatusApproved || c.ApprovedAt == nil || c.Contents[0].SMS != "final" {
		t.Fatalf("after approve: %+v", c)
	}
	got, _ := s.Get(ctx, id)
	if got.Status != StatusApproved || got.Contents[0].SMS != "final" {
		t.Fatalf("stored: %+v", got)
	}
}

func TestStore_UpdateRepeat(t *testing.T) {
	ctx, pool := emptyDB(t)
	s := NewStore(pool)
	id, err := s.CreateDraft(ctx, draft("brief", lc("tr", "one")))
	if err != nil {
		t.Fatal(err)
	}
	pass := func(c Campaign) ([]LanguageContent, error) { return c.Contents, nil }
	first, _, err := s.Update(ctx, id, true, pass)
	if err != nil {
		t.Fatal(err)
	}
	called := false
	second, repeat, err := s.Update(ctx, id, true, func(Campaign) ([]LanguageContent, error) {
		called = true
		return []LanguageContent{lc("tr", "changed")}, nil
	})
	if err != nil || !repeat || called {
		t.Fatalf("repeat approve: err %v, repeat %v, check called %v", err, repeat, called)
	}
	if !second.ApprovedAt.Equal(*first.ApprovedAt) || second.Contents[0].SMS != "one" {
		t.Fatalf("repeat changed the campaign: %+v", second)
	}
	if _, _, err := s.Update(ctx, id, false, pass); !errors.Is(err, ErrNotDraft) {
		t.Fatalf("save after approval: err = %v, want ErrNotDraft", err)
	}
}

func TestStore_UpdateNotFound(t *testing.T) {
	ctx, pool := emptyDB(t)
	pass := func(c Campaign) ([]LanguageContent, error) { return c.Contents, nil }
	if _, _, err := NewStore(pool).Update(ctx, 999, true, pass); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestStore_UpdateCheckError(t *testing.T) {
	ctx, pool := emptyDB(t)
	s := NewStore(pool)
	id, err := s.CreateDraft(ctx, draft("brief", lc("tr", "one")))
	if err != nil {
		t.Fatal(err)
	}
	boom := errors.New("boom")
	if _, _, err := s.Update(ctx, id, true, func(Campaign) ([]LanguageContent, error) { return nil, boom }); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want boom", err)
	}
	c, _ := s.Get(ctx, id)
	if c.Status != StatusDraft || c.Contents[0].SMS != "one" {
		t.Fatalf("campaign changed after check error: %+v", c)
	}
}

func TestStore_List(t *testing.T) {
	ctx, pool := emptyDB(t)
	s := NewStore(pool)
	var ids []int64
	for _, b := range []string{"a", "b", "c"} {
		id, err := s.CreateDraft(ctx, draft(b, lc("tr", "x"), lc("en", "y")))
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	list, err := s.List(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 || list[0].ID != ids[2] || list[2].ID != ids[0] {
		t.Fatalf("list order: %+v", list)
	}
	if len(list[0].Languages) != 2 || list[0].Languages[0] != "en" || list[0].Languages[1] != "tr" {
		t.Fatalf("languages = %v", list[0].Languages)
	}
	short, err := s.List(ctx, 2)
	if err != nil || len(short) != 2 {
		t.Fatalf("limit 2: %d, %v", len(short), err)
	}
}

func contentOf(sms string) content.Content {
	return content.Content{SMS: sms, EmailSubject: "Konu", EmailBody: "Merhaba"}
}
