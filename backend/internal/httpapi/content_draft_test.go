package httpapi

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

	"campaigncomposer/internal/campaign"
	"campaigncomposer/internal/content"
	"campaigncomposer/internal/llm"
	"campaigncomposer/internal/seed"
	"campaigncomposer/internal/segment"
	"campaigncomposer/internal/store"
	"campaigncomposer/internal/store/storetest"
)

// fakeDrafts records the draft saves.
type fakeDrafts struct {
	mu       sync.Mutex
	created  []campaign.Draft
	replaced map[int64]campaign.Draft
	err      error
}

func (f *fakeDrafts) CreateDraft(_ context.Context, d campaign.Draft) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return 0, f.err
	}
	f.created = append(f.created, d)
	return 41, nil
}

func (f *fakeDrafts) ReplaceDraft(_ context.Context, id int64, d campaign.Draft) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	if f.replaced == nil {
		f.replaced = map[int64]campaign.Draft{}
	}
	f.replaced[id] = d
	return nil
}

func (f *fakeDrafts) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.created) + len(f.replaced)
}

func draftGenerator(c llm.Client, f *langCounter, d *fakeDrafts) *ContentGenerator {
	g := newGenerator(c, f)
	g.Drafts = d
	return g
}

type streamErr struct {
	Language *string `json:"language"`
	Code     string  `json:"code"`
	Message  string  `json:"message"`
}

func lastEvents(evs []sseEvent, n int) []string {
	var out []string
	for _, e := range evs[len(evs)-n:] {
		out = append(out, e.Event)
	}
	return out
}

func TestGenerate_SavesDraft(t *testing.T) {
	m := contentMock(defaultContent)
	f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
	d := &fakeDrafts{}
	evs := streamEvents(t, serveContent(draftGenerator(m, f, d), genBody("  "+exampleBrief+"  ", exampleRule())))
	if got := strings.Join(lastEvents(evs, 3), ","); got != "final,campaign,done" {
		t.Fatalf("last events %s", got)
	}
	var ce struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(evs[len(evs)-2].Data, &ce); err != nil || ce.ID != 41 || ce.Status != "draft" {
		t.Fatalf("campaign event %s", evs[len(evs)-2].Data)
	}
	if len(d.created) != 1 {
		t.Fatalf("created %d drafts", len(d.created))
	}
	got := d.created[0]
	brief, _ := segment.SanitizeBrief(exampleBrief)
	rule, _ := segment.Default().ParseRule([]byte(exampleRule()))
	wantRule, _ := rule.MarshalJSON()
	if got.Brief != brief || string(got.Rule) != string(wantRule) || got.Description != segment.Default().Describe(rule) {
		t.Errorf("draft head %+v", got)
	}
	if len(got.Skipped) != 0 {
		t.Errorf("skipped %v", got.Skipped)
	}
	want, _, _ := content.Validate(content.MockContent("tr"), 14)
	if len(got.Contents) != 1 {
		t.Fatalf("contents %+v", got.Contents)
	}
	c := got.Contents[0]
	if c.Language != "tr" || c.Content != want || c.SMSCount != 205 || c.EmailCount != 290 || c.FirstNameRunes != 14 {
		t.Errorf("content %+v", c)
	}
}

func TestGenerate_DraftSkipsFailedLanguage(t *testing.T) {
	m := contentMock(func(lang string, attempt int) llm.MockStep {
		if lang == "en" {
			return llm.MockStep{Err: errors.New("boom")}
		}
		return defaultContent(lang, attempt)
	})
	f := &langCounter{langs: []segment.LanguageCounts{trLang(), enLang()}}
	d := &fakeDrafts{}
	streamEvents(t, serveContent(draftGenerator(m, f, d), genBody(exampleBrief, exampleRule())))
	if len(d.created) != 1 {
		t.Fatalf("created %d", len(d.created))
	}
	got := d.created[0]
	if len(got.Contents) != 1 || got.Contents[0].Language != "tr" || len(got.Skipped) != 1 || got.Skipped[0] != "en" {
		t.Fatalf("draft %+v", got)
	}
}

func TestGenerate_NoDraftWhenAllFail(t *testing.T) {
	m := contentMock(func(string, int) llm.MockStep { return llm.MockStep{Err: errors.New("boom")} })
	f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
	d := &fakeDrafts{}
	evs := streamEvents(t, serveContent(draftGenerator(m, f, d), genBody(exampleBrief, exampleRule())))
	if d.calls() != 0 {
		t.Fatalf("save calls %d", d.calls())
	}
	for _, e := range evs {
		if e.Event == "campaign" {
			t.Fatal("campaign event without a draft")
		}
	}
	if evs[len(evs)-1].Event != "done" {
		t.Fatalf("last %s", evs[len(evs)-1].Event)
	}
}

func TestGenerate_DraftSaveError(t *testing.T) {
	m := contentMock(defaultContent)
	f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
	d := &fakeDrafts{err: errors.New("db down")}
	evs := streamEvents(t, serveContent(draftGenerator(m, f, d), genBody(exampleBrief, exampleRule())))
	if got := strings.Join(lastEvents(evs, 2), ","); got != "error,done" {
		t.Fatalf("last events %s", got)
	}
	var se streamErr
	_ = json.Unmarshal(evs[len(evs)-2].Data, &se)
	if se.Language != nil || se.Code != "draft_not_saved" || se.Message == "" || strings.Contains(string(evs[len(evs)-2].Data), "db down") {
		t.Fatalf("error event %s", evs[len(evs)-2].Data)
	}
}

func TestGenerate_NoDraftAfterDisconnect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := &fakeStream{f: func(cctx context.Context, req llm.Request, on func(llm.StreamEvent)) (llm.Response, error) {
		raw := content.MockContent("tr")
		on(llm.StreamEvent{Delta: string(raw)})
		cancel()
		<-cctx.Done()
		// The call ends with a valid answer after the client left.
		return llm.Response{ToolUse: &llm.ToolUse{ID: "c1", Name: content.ToolName, Input: raw}}, nil
	}}
	f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
	d := &fakeDrafts{}
	done := make(chan struct{})
	go func() {
		serveContentCtx(ctx, draftGenerator(c, f, d), genBody("hello", exampleRule()))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not return")
	}
	if d.calls() != 0 {
		t.Fatalf("saved %d drafts after disconnect", d.calls())
	}
}

func TestGenerate_ReplaceDraft(t *testing.T) {
	m := contentMock(defaultContent)
	f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
	d := &fakeDrafts{}
	b, _ := json.Marshal(exampleBrief)
	evs := streamEvents(t, serveContent(draftGenerator(m, f, d), fmt.Sprintf(`{"brief":%s,"rule":%s,"campaign_id":7}`, b, exampleRule())))
	if len(d.created) != 0 || len(d.replaced) != 1 {
		t.Fatalf("created %d, replaced %v", len(d.created), d.replaced)
	}
	if _, ok := d.replaced[7]; !ok {
		t.Fatalf("replaced %v, want id 7", d.replaced)
	}
	if !strings.Contains(string(evs[len(evs)-2].Data), `"id":7`) {
		t.Fatalf("campaign event %s", evs[len(evs)-2].Data)
	}
}

func TestGenerate_ReplaceConflict(t *testing.T) {
	for _, e := range []error{campaign.ErrConflict, campaign.ErrNotFound} {
		m := contentMock(defaultContent)
		f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
		d := &fakeDrafts{err: e}
		b, _ := json.Marshal(exampleBrief)
		evs := streamEvents(t, serveContent(draftGenerator(m, f, d), fmt.Sprintf(`{"brief":%s,"rule":%s,"campaign_id":7}`, b, exampleRule())))
		var se streamErr
		_ = json.Unmarshal(evs[len(evs)-2].Data, &se)
		if evs[len(evs)-2].Event != "error" || se.Code != "campaign_conflict" || se.Language != nil {
			t.Fatalf("%v: event %s %s", e, evs[len(evs)-2].Event, evs[len(evs)-2].Data)
		}
	}
}

func TestGenerate_BadCampaignID(t *testing.T) {
	for _, id := range []string{"0", "-1", `"7"`, "1.5"} {
		m := contentMock(defaultContent)
		f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
		b, _ := json.Marshal(exampleBrief)
		rec := serveContent(draftGenerator(m, f, &fakeDrafts{}), fmt.Sprintf(`{"brief":%s,"rule":%s,"campaign_id":%s}`, b, exampleRule(), id))
		if rec.Code != 400 || !strings.Contains(rec.Body.String(), "invalid_request") {
			t.Errorf("campaign_id %s: %d %s", id, rec.Code, rec.Body)
		}
		if len(m.Requests()) != 0 {
			t.Errorf("campaign_id %s: LLM called", id)
		}
	}
}

func TestGenerate_DraftInDatabase(t *testing.T) {
	url := storetest.NewDB(t)
	ctx := context.Background()
	if err := store.Migrate(ctx, url); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := seed.Load(ctx, pool, seed.Options{Random: 42, Now: time.Now(), Users: 500}, false); err != nil {
		t.Fatal(err)
	}
	count := func(ctx context.Context, w segment.Where) (segment.Audience, error) {
		return segment.CountAudience(ctx, pool, w)
	}
	st := campaign.NewStore(pool)
	g := &ContentGenerator{
		LLM: contentMock(defaultContent), Allow: segment.Default(), Count: count,
		Now: time.Now, Clock: frozenClock, LLMTimeout: time.Second, DraftEvery: 100 * time.Millisecond,
		Drafts: st,
	}
	rule := `{"op":"and","conditions":[{"field":"city","op":"eq","value":"Istanbul"}]}`
	evs := streamEvents(t, serveContent(g, genBody(exampleBrief, rule)))
	var ce struct {
		ID int64 `json:"id"`
	}
	for _, e := range evs {
		if e.Event == "campaign" {
			_ = json.Unmarshal(e.Data, &ce)
		}
	}
	if ce.ID == 0 {
		t.Fatalf("no campaign event: %v", names(evs))
	}
	c, err := st.Get(ctx, ce.ID)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := segment.Default().ParseRule([]byte(rule))
	aud, err := count(ctx, segment.Default().Compile(r, time.Now()))
	if err != nil {
		t.Fatal(err)
	}
	if c.Status != campaign.StatusDraft || len(c.Contents) != len(aud.Languages) {
		t.Fatalf("campaign %+v, audience languages %d", c, len(aud.Languages))
	}
	for _, l := range aud.Languages {
		found := false
		for _, lc := range c.Contents {
			if lc.Language == l.Code {
				found = true
				if lc.SMSCount != l.SMS || lc.EmailCount != l.Email || lc.FirstNameRunes != int(l.MaxFirstName) {
					t.Errorf("%s: stored %+v, audience %+v", l.Code, lc, l)
				}
			}
		}
		if !found {
			t.Errorf("no content row for %s", l.Code)
		}
	}
	// A regenerate with the same brief and rule replaces the same draft.
	b, _ := json.Marshal(exampleBrief)
	evs = streamEvents(t, serveContent(g, fmt.Sprintf(`{"brief":%s,"rule":%s,"campaign_id":%d}`, b, rule, ce.ID)))
	if !strings.Contains(string(evs[len(evs)-2].Data), fmt.Sprintf(`"id":%d`, ce.ID)) {
		t.Fatalf("regenerate: %s %s", evs[len(evs)-2].Event, evs[len(evs)-2].Data)
	}
	list, _ := st.List(ctx, 10)
	if len(list) != 1 {
		t.Fatalf("campaigns after regenerate: %d", len(list))
	}
}
