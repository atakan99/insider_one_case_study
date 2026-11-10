package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/llm"
	"campaigncomposer/internal/segment"
)

// langCounter is a fake AudienceCounter that returns fixed languages.
type langCounter struct {
	mu    sync.Mutex
	calls int
	langs []segment.LanguageCounts
	err   error
}

func (f *langCounter) count(_ context.Context, _ segment.Where) (segment.Audience, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.err != nil {
		return segment.Audience{}, f.err
	}
	a := segment.Audience{Languages: f.langs}
	for _, l := range f.langs {
		a.Email += l.Email
		a.SMS += l.SMS
		a.Reachable += l.Reachable
		a.Matched += l.Matched
	}
	a.NoConsent = a.Matched - a.Reachable
	return a, nil
}

func (f *langCounter) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func trLang() segment.LanguageCounts {
	return segment.LanguageCounts{Code: "tr", Counts: segment.Counts{Email: 290, SMS: 205, Reachable: 330, Matched: 412, NoConsent: 82}, MaxFirstName: 14}
}

func enLang() segment.LanguageCounts {
	return segment.LanguageCounts{Code: "en", Counts: segment.Counts{Email: 40, SMS: 30, Reachable: 45, Matched: 50, NoConsent: 5}, MaxFirstName: 9}
}

func frozenClock() time.Time { return time.Unix(1_700_000_000, 0) }

func newGenerator(c llm.Client, f *langCounter) *ContentGenerator {
	return &ContentGenerator{
		LLM: c, Allow: segment.Default(), Count: f.count,
		Now: time.Now, Clock: frozenClock,
		LLMTimeout: time.Second, DraftEvery: 100 * time.Millisecond,
	}
}

func serveContent(g *ContentGenerator, body string) *httptest.ResponseRecorder {
	return serveContentCtx(context.Background(), g, body)
}

func serveContentCtx(ctx context.Context, g *ContentGenerator, body string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	RegisterContent(mux, g)
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/content/generate", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// exampleRule is the rule object of llm.ExampleRuleJSON.
func exampleRule() string {
	var ex struct {
		Rule json.RawMessage `json:"rule"`
	}
	_ = json.Unmarshal([]byte(llm.ExampleRuleJSON), &ex)
	return string(ex.Rule)
}

func genBody(brief, rule string) string {
	b, _ := json.Marshal(brief)
	return fmt.Sprintf(`{"brief":%s,"rule":%s}`, b, rule)
}

const exampleBrief = "Son 30 günde sepetine ürün ekleyip almayan İstanbul kullanıcılarına %30 indirim, sıcak ve samimi bir dille."

func contentStep(raw string) llm.MockStep {
	return llm.MockStep{Response: llm.Response{
		ToolUse:   &llm.ToolUse{ID: "c1", Name: content.ToolName, Input: json.RawMessage(raw)},
		Assistant: "assistant-content",
	}}
}

// contentMock answers by the language of the request and the attempt number
// (1 for the first call, 3 for the repair call).
func contentMock(f func(lang string, attempt int) llm.MockStep) *llm.Mock {
	m := llm.NewMockFunc(func(req llm.Request) llm.MockStep {
		lang, _ := content.LanguageOf(req.Turns[0].User)
		return f(lang, len(req.Turns))
	})
	m.SetStream(8, 0)
	return m
}

func defaultContent(lang string, _ int) llm.MockStep {
	return contentStep(string(content.MockContent(lang)))
}

var tooLongSMS = `{"sms":"` + strings.Repeat("a", 200) + `","email_subject":"Hi","email_body":"Body"}`

// fakeStream is an llm.Client whose Stream runs a test function.
type fakeStream struct {
	mu   sync.Mutex
	reqs []llm.Request
	f    func(ctx context.Context, req llm.Request, on func(llm.StreamEvent)) (llm.Response, error)
}

func (c *fakeStream) Generate(context.Context, llm.Request) (llm.Response, error) {
	return llm.Response{}, errors.New("fakeStream: Generate is not used")
}

func (c *fakeStream) Stream(ctx context.Context, req llm.Request, on func(llm.StreamEvent)) (llm.Response, error) {
	c.mu.Lock()
	c.reqs = append(c.reqs, req)
	c.mu.Unlock()
	return c.f(ctx, req, on)
}

func (c *fakeStream) requests() []llm.Request {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]llm.Request(nil), c.reqs...)
}

type langData struct {
	Language string               `json:"language"`
	Content  content.Content      `json:"content"`
	Limits   content.Limits       `json:"limits"`
	Errors   []content.FieldError `json:"errors"`
	Code     string               `json:"code"`
	Message  string               `json:"message"`
	Details  []content.FieldError `json:"details"`
}

func dataOf(t *testing.T, e sseEvent) langData {
	t.Helper()
	var d langData
	if err := json.Unmarshal(e.Data, &d); err != nil {
		t.Fatalf("decode %s: %v", e.Data, err)
	}
	return d
}

// forLang returns the events of one language, in order.
func forLang(t *testing.T, evs []sseEvent, lang string) []sseEvent {
	t.Helper()
	var out []sseEvent
	for _, e := range evs {
		if e.Event == "audience" || e.Event == "done" {
			continue
		}
		if dataOf(t, e).Language == lang {
			out = append(out, e)
		}
	}
	return out
}

// names gives the event names, with runs of the same name merged.
func names(evs []sseEvent) []string {
	var out []string
	for _, e := range evs {
		if len(out) == 0 || out[len(out)-1] != e.Event {
			out = append(out, e.Event)
		}
	}
	return out
}

func streamEvents(t *testing.T, rec *httptest.ResponseRecorder) []sseEvent {
	t.Helper()
	if rec.Code != 200 {
		t.Fatalf("code %d: %s", rec.Code, rec.Body)
	}
	return parseSSE(t, rec.Body.String())
}

func TestGenerate_Example(t *testing.T) {
	m := contentMock(defaultContent)
	f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
	rec := serveContent(newGenerator(m, f), genBody(exampleBrief, exampleRule()))
	evs := streamEvents(t, rec)
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("content type %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("cache control %q", cc)
	}
	if len(evs) < 4 || evs[0].Event != "audience" || evs[len(evs)-1].Event != "done" {
		t.Fatalf("events %v", names(evs))
	}
	if got := names(forLang(t, evs, "tr")); strings.Join(got, ",") != "draft,final" {
		t.Errorf("tr events %v", got)
	}
	if string(evs[len(evs)-1].Data) != "{}" {
		t.Errorf("done data %s", evs[len(evs)-1].Data)
	}

	var aud struct {
		Description string           `json:"description"`
		Raw         []map[string]any `json:"languages"`
	}
	if err := json.Unmarshal(evs[0].Data, &aud); err != nil {
		t.Fatal(err)
	}
	rule, _ := segment.Default().ParseRule([]byte(exampleRule()))
	if aud.Description != segment.Default().Describe(rule) {
		t.Errorf("description %q", aud.Description)
	}
	if len(aud.Raw) != 1 || aud.Raw[0]["code"] != "tr" || aud.Raw[0]["reachable"] != float64(330) ||
		aud.Raw[0]["email"] != float64(290) || aud.Raw[0]["sms"] != float64(205) ||
		aud.Raw[0]["matched"] != float64(412) || aud.Raw[0]["no_consent"] != float64(82) {
		t.Errorf("audience %s", evs[0].Data)
	}
	if strings.Contains(strings.ToLower(string(evs[0].Data)), "first") {
		t.Errorf("audience leaks max first name: %s", evs[0].Data)
	}

	tr := forLang(t, evs, "tr")
	final := dataOf(t, tr[len(tr)-1])
	want, wantLimits, errs := content.Validate(content.MockContent("tr"), 14)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	if final.Content != want || final.Limits != wantLimits || final.Limits.FirstNameRunes != 14 {
		t.Errorf("final %+v", final)
	}
	if last := dataOf(t, tr[len(tr)-2]); last.Content != want {
		t.Errorf("last draft %+v", last.Content)
	}

	reqs := m.Requests()
	if len(reqs) != 1 {
		t.Fatalf("requests %d", len(reqs))
	}
	r := reqs[0]
	if r.Tool.Name != content.ToolName || !r.Tool.Eager || r.System != content.SystemPrompt() || r.Tool.Schema == nil {
		t.Errorf("request %+v", r.Tool)
	}
	brief, _ := segment.SanitizeBrief(exampleBrief)
	if r.Turns[0].User != content.UserMessage("tr", aud.Description, 14, brief) {
		t.Errorf("user message %q", r.Turns[0].User)
	}
}

func TestGenerate_BadBody(t *testing.T) {
	rule := exampleRule()
	bodies := []string{
		genBody(strings.Repeat("a", 16*1024), rule),
		"{",
		"{}",
		`{"brief":"hello"}`,
		`{"rule":` + rule + `}`,
		`{"brief":"hello","rule":null}`,
		`{"brief":"","rule":` + rule + `}`,
		genBody("hello", rule) + " garbage",
		genBody("hello", rule) + strings.Repeat(" ", 16*1024),
	}
	for _, body := range bodies {
		m := contentMock(defaultContent)
		f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
		rec := serveContent(newGenerator(m, f), body)
		if rec.Code != 400 || decodeErr(t, rec).Error.Code != "invalid_request" {
			t.Errorf("%.30q: %d %s", body, rec.Code, rec.Body)
		}
		if len(m.Requests()) != 0 || f.callCount() != 0 {
			t.Errorf("%.30q: llm %d, counter %d", body, len(m.Requests()), f.callCount())
		}
	}
}

func TestGenerate_BadBrief(t *testing.T) {
	for _, b := range []string{"!!!", strings.Repeat("a", 1001)} {
		m := contentMock(defaultContent)
		f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
		rec := serveContent(newGenerator(m, f), genBody(b, exampleRule()))
		if rec.Code != 400 || decodeErr(t, rec).Error.Code != "invalid_brief" {
			t.Errorf("%d %s", rec.Code, rec.Body)
		}
		if len(m.Requests()) != 0 || f.callCount() != 0 {
			t.Error("llm or counter called")
		}
	}
}

func TestGenerate_BadRule(t *testing.T) {
	rules := []string{
		`{"op":"and","conditions":[{"field":"password","op":"eq","value":"x"}]}`,
		`{"op":"and","conditions":[{"field":"city","op":"eq","value":"Istanbul' OR 1=1 --"}]}`,
		`"city = 'Istanbul'"`,
		`[]`,
	}
	for _, rule := range rules {
		m := contentMock(defaultContent)
		f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
		rec := serveContent(newGenerator(m, f), genBody("hello", rule))
		e := decodeErr(t, rec)
		if rec.Code != 400 || e.Error.Code != "invalid_rule" || len(e.Error.Details) == 0 {
			t.Errorf("%s: %d %s", rule, rec.Code, rec.Body)
		}
		if len(m.Requests()) != 0 || f.callCount() != 0 {
			t.Errorf("%s: llm %d, counter %d", rule, len(m.Requests()), f.callCount())
		}
	}
}

func TestGenerate_EmptyAudience(t *testing.T) {
	for _, langs := range [][]segment.LanguageCounts{
		{},
		{{Code: "tr", Counts: segment.Counts{Matched: 4, NoConsent: 4}, MaxFirstName: 0}},
	} {
		m := contentMock(defaultContent)
		rec := serveContent(newGenerator(m, &langCounter{langs: langs}), genBody("hello", exampleRule()))
		if rec.Code != 422 || decodeErr(t, rec).Error.Code != "empty_audience" {
			t.Errorf("%d %s", rec.Code, rec.Body)
		}
		if len(m.Requests()) != 0 {
			t.Error("llm called")
		}
	}
}

func TestGenerate_DBError(t *testing.T) {
	m := contentMock(defaultContent)
	rec := serveContent(newGenerator(m, &langCounter{err: errors.New("pg: boom")}), genBody("hello", exampleRule()))
	if rec.Code != 503 || decodeErr(t, rec).Error.Code != "db_unavailable" || strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
	if len(m.Requests()) != 0 {
		t.Error("llm called")
	}
}

func TestGenerate_RuleRoundTrip(t *testing.T) {
	rule, errs := segment.Default().ParseRule([]byte(exampleRule()))
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	raw, err := json.Marshal(rule)
	if err != nil {
		t.Fatal(err)
	}
	f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
	rec := serveContent(newGenerator(contentMock(defaultContent), f), genBody("hello", string(raw)))
	evs := streamEvents(t, rec)
	if evs[len(evs)-1].Event != "done" || f.callCount() != 1 {
		t.Errorf("events %v", names(evs))
	}
}

func TestGenerate_TwoLanguages(t *testing.T) {
	m := contentMock(func(lang string, attempt int) llm.MockStep {
		if lang == "en" {
			return llm.MockStep{Err: errors.New("secret sdk text")}
		}
		return defaultContent(lang, attempt)
	})
	f := &langCounter{langs: []segment.LanguageCounts{trLang(), enLang()}}
	evs := streamEvents(t, serveContent(newGenerator(m, f), genBody(exampleBrief, exampleRule())))

	var aud struct {
		Languages []struct {
			Code string `json:"code"`
		} `json:"languages"`
	}
	_ = json.Unmarshal(evs[0].Data, &aud)
	if len(aud.Languages) != 2 || aud.Languages[0].Code != "tr" || aud.Languages[1].Code != "en" {
		t.Errorf("audience %s", evs[0].Data)
	}
	tr := forLang(t, evs, "tr")
	if len(tr) == 0 || tr[len(tr)-1].Event != "final" {
		t.Errorf("tr events %v", names(tr))
	}
	en := forLang(t, evs, "en")
	if len(en) != 1 || en[0].Event != "error" {
		t.Fatalf("en events %v", names(en))
	}
	if d := dataOf(t, en[0]); d.Code != "llm_error" || d.Message == "" || strings.Contains(string(en[0].Data), "secret") {
		t.Errorf("en error %s", en[0].Data)
	}
	if evs[len(evs)-1].Event != "done" {
		t.Errorf("last event %s", evs[len(evs)-1].Event)
	}

	seen := map[string]bool{}
	for _, r := range m.Requests() {
		lang, ok := content.LanguageOf(r.Turns[0].User)
		if !ok {
			t.Errorf("no language in %q", r.Turns[0].User)
		}
		seen[lang] = true
	}
	if len(m.Requests()) != 2 || !seen["tr"] || !seen["en"] {
		t.Errorf("requests %d, languages %v", len(m.Requests()), seen)
	}
}

func TestGenerate_Repair(t *testing.T) {
	good := string(content.MockContent("tr"))
	run := func(t *testing.T, first, second llm.MockStep) ([]sseEvent, []llm.Request) {
		t.Helper()
		m := contentMock(func(_ string, attempt int) llm.MockStep {
			if attempt == 1 {
				return first
			}
			return second
		})
		f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
		evs := streamEvents(t, serveContent(newGenerator(m, f), genBody(exampleBrief, exampleRule())))
		if evs[len(evs)-1].Event != "done" {
			t.Errorf("last event %s", evs[len(evs)-1].Event)
		}
		return forLang(t, evs, "tr"), m.Requests()
	}

	t.Run("bad then good", func(t *testing.T) {
		tr, reqs := run(t, contentStep(tooLongSMS), contentStep(good))
		if got := strings.Join(names(tr), ","); got != "draft,retry,draft,final" {
			t.Fatalf("events %s", got)
		}
		for _, e := range tr {
			if e.Event == "retry" {
				d := dataOf(t, e)
				if len(d.Errors) == 0 || d.Errors[0].Path != "sms" {
					t.Errorf("retry %s", e.Data)
				}
			}
		}
		if len(reqs) != 2 {
			t.Fatalf("requests %d", len(reqs))
		}
		turns := reqs[1].Turns
		if len(turns) != 3 || turns[1].Assistant != "assistant-content" {
			t.Fatalf("turns %+v", turns)
		}
		tres := turns[2].ToolResult
		if tres == nil || !tres.IsError || tres.ToolUseID != "c1" || !strings.Contains(tres.Content, "sms") {
			t.Errorf("tool result %+v", tres)
		}
	})

	t.Run("bad twice", func(t *testing.T) {
		tr, reqs := run(t, contentStep(tooLongSMS), contentStep(tooLongSMS))
		if got := strings.Join(names(tr), ","); got != "draft,retry,draft,error" {
			t.Fatalf("events %s", got)
		}
		d := dataOf(t, tr[len(tr)-1])
		if d.Code != "content_invalid" || len(d.Details) == 0 || d.Details[0].Path != "sms" {
			t.Errorf("error %s", tr[len(tr)-1].Data)
		}
		if strings.Contains(string(tr[len(tr)-1].Data), "aaaaaaaaaa") {
			t.Errorf("error has model output: %s", tr[len(tr)-1].Data)
		}
		if len(reqs) != 2 {
			t.Errorf("requests %d", len(reqs))
		}
	})

	t.Run("no tool call", func(t *testing.T) {
		tr, reqs := run(t, llm.MockStep{Response: llm.Response{Assistant: "text answer"}}, contentStep(good))
		if tr[len(tr)-1].Event != "final" {
			t.Errorf("events %v", names(tr))
		}
		if len(reqs) != 2 {
			t.Fatalf("requests %d", len(reqs))
		}
		turns := reqs[1].Turns
		last := turns[len(turns)-1]
		if len(turns) != 3 || turns[1].Assistant != "text answer" || last.ToolResult != nil || !strings.Contains(last.User, content.ToolName) {
			t.Errorf("turns %+v", turns)
		}
	})

	t.Run("cut off", func(t *testing.T) {
		cut := contentStep(`{"sms":"Merhaba {{first_na`)
		cut.Response.CutOff = true
		tr, reqs := run(t, cut, contentStep(good))
		if got := strings.Join(names(tr), ","); got != "draft,retry,draft,final" {
			t.Fatalf("events %s", got)
		}
		var retry langData
		for _, e := range tr {
			if e.Event == "retry" {
				retry = dataOf(t, e)
			}
		}
		found := false
		for _, fe := range retry.Errors {
			if fe.Message == "answer was cut off" {
				found = true
			}
		}
		if !found {
			t.Errorf("retry errors %+v", retry.Errors)
		}
		if len(reqs) != 2 || reqs[1].Turns[2].ToolResult == nil || !reqs[1].Turns[2].ToolResult.IsError {
			t.Errorf("repair turns %+v", reqs[len(reqs)-1].Turns)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		tr, reqs := run(t, contentStep(""), contentStep(good))
		if tr[len(tr)-1].Event != "final" || len(reqs) != 2 {
			t.Errorf("events %v, requests %d", names(tr), len(reqs))
		}
		for _, e := range tr {
			if e.Event == "retry" && len(dataOf(t, e).Errors) == 0 {
				t.Errorf("retry has no errors: %s", e.Data)
			}
		}
	})
}

func TestGenerate_LLMErrors(t *testing.T) {
	cases := []struct {
		name string
		step llm.MockStep
		code string
	}{
		{"error", llm.MockStep{Err: errors.New("secret sdk text")}, "llm_error"},
		{"timeout", llm.MockStep{Err: context.DeadlineExceeded}, "llm_timeout"},
		{"refused", llm.MockStep{Response: llm.Response{Refused: true}}, "llm_refused"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := contentMock(func(string, int) llm.MockStep { return c.step })
			f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
			evs := streamEvents(t, serveContent(newGenerator(m, f), genBody("hello", exampleRule())))
			tr := forLang(t, evs, "tr")
			if len(tr) == 0 {
				t.Fatal("no tr events")
			}
			last := tr[len(tr)-1]
			if d := dataOf(t, last); last.Event != "error" || d.Code != c.code || d.Message == "" || strings.Contains(string(last.Data), "secret") {
				t.Errorf("event %s %s", last.Event, last.Data)
			}
			if len(m.Requests()) != 1 {
				t.Errorf("requests %d", len(m.Requests()))
			}
			if evs[len(evs)-1].Event != "done" {
				t.Errorf("last event %s", evs[len(evs)-1].Event)
			}
		})
	}

	t.Run("attempt timeout", func(t *testing.T) {
		c := &fakeStream{f: func(ctx context.Context, _ llm.Request, _ func(llm.StreamEvent)) (llm.Response, error) {
			<-ctx.Done()
			return llm.Response{}, ctx.Err()
		}}
		g := newGenerator(c, &langCounter{langs: []segment.LanguageCounts{trLang()}})
		g.LLMTimeout = 5 * time.Millisecond
		evs := streamEvents(t, serveContent(g, genBody("hello", exampleRule())))
		tr := forLang(t, evs, "tr")
		if len(tr) != 1 || dataOf(t, tr[0]).Code != "llm_timeout" {
			t.Errorf("events %v", names(tr))
		}
	})
}

func TestGenerate_Throttle(t *testing.T) {
	raw := content.MockContent("tr")
	chunk := (len(raw) + 9) / 10

	t.Run("frozen clock", func(t *testing.T) {
		m := contentMock(defaultContent)
		m.SetStream(chunk, 0)
		f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
		evs := streamEvents(t, serveContent(newGenerator(m, f), genBody("hello", exampleRule())))
		tr := forLang(t, evs, "tr")
		var drafts []langData
		for _, e := range tr {
			if e.Event == "draft" {
				drafts = append(drafts, dataOf(t, e))
			}
		}
		if len(drafts) != 2 {
			t.Fatalf("drafts %d, want 2", len(drafts))
		}
		want, _, _ := content.Validate(raw, 14)
		if drafts[1].Content != want {
			t.Errorf("last draft %+v", drafts[1].Content)
		}
		if drafts[0].Content == want {
			t.Error("first draft already has the full text")
		}
	})

	t.Run("moving clock", func(t *testing.T) {
		m := contentMock(defaultContent)
		m.SetStream(chunk, 0)
		var ticks atomic.Int64
		g := newGenerator(m, &langCounter{langs: []segment.LanguageCounts{trLang()}})
		g.Clock = func() time.Time {
			return frozenClock().Add(time.Duration(ticks.Add(1)) * 100 * time.Millisecond)
		}
		evs := streamEvents(t, serveContent(g, genBody("hello", exampleRule())))
		n := 0
		for _, e := range forLang(t, evs, "tr") {
			if e.Event == "draft" {
				n++
			}
		}
		if n < 10 {
			t.Errorf("drafts %d, want one per fragment", n)
		}
	})
}

func TestGenerate_Reset(t *testing.T) {
	good := string(content.MockContent("tr"))
	c := &fakeStream{f: func(_ context.Context, req llm.Request, on func(llm.StreamEvent)) (llm.Response, error) {
		if len(req.Turns) == 1 {
			on(llm.StreamEvent{Delta: `{"sms":"before-fallback`})
			on(llm.StreamEvent{Reset: true})
			on(llm.StreamEvent{Delta: tooLongSMS})
			return contentStep(tooLongSMS).Response, nil
		}
		on(llm.StreamEvent{Delta: good})
		return contentStep(good).Response, nil
	}}
	f := &langCounter{langs: []segment.LanguageCounts{trLang()}}
	evs := streamEvents(t, serveContent(newGenerator(c, f), genBody("hello", exampleRule())))
	tr := forLang(t, evs, "tr")
	if got := strings.Join(names(tr), ","); got != "draft,retry,draft,retry,draft,final" {
		t.Fatalf("events %s", got)
	}
	var retries []sseEvent
	for _, e := range tr {
		if e.Event == "retry" {
			retries = append(retries, e)
		}
	}
	if d := dataOf(t, retries[0]); len(d.Errors) != 0 {
		t.Errorf("reset retry has errors: %s", retries[0].Data)
	}
	if d := dataOf(t, retries[1]); len(d.Errors) == 0 {
		t.Errorf("repair retry has no errors: %s", retries[1].Data)
	}
	afterReset := false
	for _, e := range tr {
		if e.Event == "retry" {
			afterReset = true
		}
		if afterReset && e.Event == "draft" && strings.Contains(dataOf(t, e).Content.SMS, "before-fallback") {
			t.Errorf("draft after reset has old text: %s", e.Data)
		}
	}
	if len(c.requests()) != 2 {
		t.Errorf("requests %d", len(c.requests()))
	}
}

func TestGenerate_Disconnect(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var mu sync.Mutex
	seen := map[string]error{}
	c := &fakeStream{f: func(cctx context.Context, req llm.Request, on func(llm.StreamEvent)) (llm.Response, error) {
		lang, _ := content.LanguageOf(req.Turns[0].User)
		on(llm.StreamEvent{Delta: `{"sms":"Hi`})
		if lang == "tr" {
			cancel()
		}
		<-cctx.Done()
		mu.Lock()
		seen[lang] = cctx.Err()
		mu.Unlock()
		return llm.Response{}, cctx.Err()
	}}
	f := &langCounter{langs: []segment.LanguageCounts{trLang(), enLang()}}
	done := make(chan struct{})
	go func() {
		serveContentCtx(ctx, newGenerator(c, f), genBody("hello", exampleRule()))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not return after the disconnect")
	}
	mu.Lock()
	defer mu.Unlock()
	if seen["tr"] == nil || seen["en"] == nil {
		t.Errorf("calls did not see a cancelled context: %v", seen)
	}
}

func TestGenerate_LogsNoText(t *testing.T) {
	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	defer slog.SetDefault(old)

	m := contentMock(func(lang string, attempt int) llm.MockStep {
		if lang == "tr" && attempt == 1 {
			return contentStep(tooLongSMS)
		}
		return defaultContent(lang, attempt)
	})
	f := &langCounter{langs: []segment.LanguageCounts{trLang(), enLang()}}
	evs := streamEvents(t, serveContent(newGenerator(m, f), genBody("secret-brief-text-123", exampleRule())))
	if evs[len(evs)-1].Event != "done" {
		t.Fatalf("events %v", names(evs))
	}
	out := buf.String()
	if !strings.Contains(out, `"language":"tr"`) || !strings.Contains(out, `"language":"en"`) || !strings.Contains(out, "sms_runes") {
		t.Errorf("missing language log lines: %s", out)
	}
	var tr, en content.Content
	_ = json.Unmarshal(content.MockContent("tr"), &tr)
	_ = json.Unmarshal(content.MockContent("en"), &en)
	for _, s := range []string{"secret-brief-text-123", tr.SMS, tr.EmailSubject, en.SMS, en.EmailSubject, "Sepetine eklediğin", "aaaaaaaaaa"} {
		if strings.Contains(out, s) {
			t.Errorf("log contains %q: %s", s, out)
		}
	}
}
