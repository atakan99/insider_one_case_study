package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"campaigncomposer/internal/llm"
	"campaigncomposer/internal/segment"
)

type fakeCount struct {
	mu    sync.Mutex
	calls int
	err   error
}

func (f *fakeCount) count(_ context.Context, _ segment.Where) (segment.Audience, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.err != nil {
		return segment.Audience{}, f.err
	}
	return segment.Audience{Counts: segment.Counts{Email: 290, SMS: 205, Reachable: 330, Matched: 412, NoConsent: 82}, Languages: []segment.LanguageCounts{}}, nil
}

func newPreview(m *llm.Mock, f *fakeCount) http.Handler {
	mux := http.NewServeMux()
	RegisterSegments(mux, &SegmentPreviewer{
		LLM: m, Allow: segment.Default(), Count: f.count,
		Now: time.Now, LLMTimeout: time.Second,
	})
	return mux
}

func post(h http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/segments/preview", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func briefBody(b string) string {
	j, _ := json.Marshal(map[string]string{"brief": b})
	return string(j)
}

type errBody struct {
	Error struct {
		Code    string               `json:"code"`
		Message string               `json:"message"`
		Details []segment.FieldError `json:"details"`
	} `json:"error"`
}

func decodeErr(t *testing.T, rec *httptest.ResponseRecorder) errBody {
	t.Helper()
	var e errBody
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatalf("decode: %v: %s", err, rec.Body.String())
	}
	return e
}

func toolStep(raw string) llm.MockStep {
	return llm.MockStep{Response: llm.Response{
		ToolUse:   &llm.ToolUse{ID: "t1", Name: segment.ToolName, Input: json.RawMessage(raw)},
		Assistant: "assistant-" + raw[:5],
	}}
}

const badCity = `{"status":"ok","rule":{"op":"and","conditions":[{"field":"city","op":"eq","value":"Atlantis"}]},"unmapped":[]}`

func TestPreview_Example(t *testing.T) {
	f := &fakeCount{}
	rec := post(newPreview(llm.NewMock(), f), briefBody("Istanbul cart users"))
	if rec.Code != 200 {
		t.Fatalf("code %d: %s", rec.Code, rec.Body)
	}
	var got struct {
		Rule        json.RawMessage  `json:"rule"`
		Description string           `json:"description"`
		Unmapped    []string         `json:"unmapped"`
		Audience    segment.Audience `json:"audience"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	var ex struct {
		Rule json.RawMessage `json:"rule"`
	}
	_ = json.Unmarshal([]byte(llm.ExampleRuleJSON), &ex)
	var a, b any
	_ = json.Unmarshal(got.Rule, &a)
	_ = json.Unmarshal(ex.Rule, &b)
	if !jsonEq(a, b) {
		t.Errorf("rule %s", got.Rule)
	}
	if got.Description == "" || got.Audience.Matched != 412 {
		t.Errorf("body %s", rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"unmapped":[]`) || !strings.Contains(rec.Body.String(), `"languages":[]`) {
		t.Errorf("want empty arrays: %s", rec.Body)
	}
}

func jsonEq(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return bytes.Equal(x, y)
}

func TestPreview_BadBody(t *testing.T) {
	bodies := []string{
		briefBody(strings.Repeat("a", 16*1024)),
		"{",
		"{}",
		// Trailing data after the object.
		`{"brief":"hello"} garbage`,
		`{"brief":"hello"}{"brief":"again"}`,
		// A valid object padded with whitespace past the 16 KiB limit.
		`{"brief":"hello"}` + strings.Repeat(" ", 16*1024),
	}
	for _, body := range bodies {
		m := llm.NewMock()
		rec := post(newPreview(m, &fakeCount{}), body)
		if rec.Code != 400 || decodeErr(t, rec).Error.Code != "invalid_request" {
			t.Errorf("%.20q: %d %s", body, rec.Code, rec.Body)
		}
		if n := len(m.Requests()); n != 0 {
			t.Errorf("%.20q: LLM got %d calls, want 0", body, n)
		}
	}
}

func TestPreview_TrailingWhitespaceWithinLimit(t *testing.T) {
	rec := post(newPreview(llm.NewMock(), &fakeCount{}), `{"brief":"hello"}`+"\n  \n")
	if rec.Code != 200 {
		t.Fatalf("got %d %s, want 200", rec.Code, rec.Body)
	}
}

func TestPreview_BadBrief(t *testing.T) {
	for _, b := range []string{"!!!", strings.Repeat("a", 1001)} {
		m := llm.NewMock()
		rec := post(newPreview(m, &fakeCount{}), briefBody(b))
		if rec.Code != 400 || decodeErr(t, rec).Error.Code != "invalid_brief" {
			t.Errorf("%d %s", rec.Code, rec.Body)
		}
		if len(m.Requests()) != 0 {
			t.Error("llm called")
		}
	}
}

func TestPreview_Repair(t *testing.T) {
	s1 := toolStep(badCity)
	m := llm.NewMock(s1, toolStep(llm.ExampleRuleJSON))
	rec := post(newPreview(m, &fakeCount{}), briefBody("hello"))
	if rec.Code != 200 {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	reqs := m.Requests()
	if len(reqs) != 2 {
		t.Fatalf("requests %d", len(reqs))
	}
	turns := reqs[1].Turns
	n := len(turns)
	if n != 3 || turns[n-2].Assistant != s1.Response.Assistant {
		t.Fatalf("turns %+v", turns)
	}
	tr := turns[n-1].ToolResult
	if tr == nil || !tr.IsError || tr.ToolUseID != "t1" || !strings.Contains(tr.Content, "conditions[0].value") {
		t.Errorf("tool result %+v", tr)
	}
}

func TestPreview_RepairNoToolCall(t *testing.T) {
	m := llm.NewMock(llm.MockStep{Response: llm.Response{Assistant: "text"}}, toolStep(llm.ExampleRuleJSON))
	rec := post(newPreview(m, &fakeCount{}), briefBody("hello"))
	if rec.Code != 200 {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	turns := m.Requests()[1].Turns
	last := turns[len(turns)-1]
	if last.User == "" || !strings.Contains(last.User, segment.ToolName) {
		t.Errorf("last turn %+v", last)
	}
}

func TestPreview_RepairFails(t *testing.T) {
	f := &fakeCount{}
	m := llm.NewMock(toolStep(badCity), toolStep(badCity))
	rec := post(newPreview(m, f), briefBody("hello"))
	e := decodeErr(t, rec)
	if rec.Code != 422 || e.Error.Code != "segment_invalid" || len(e.Error.Details) == 0 {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
	if f.calls != 0 || len(m.Requests()) != 2 {
		t.Errorf("calls %d reqs %d", f.calls, len(m.Requests()))
	}
}

func TestPreview_Unsupported(t *testing.T) {
	f := &fakeCount{}
	m := llm.NewMock(toolStep(`{"status":"unsupported","reason":"No product data."}`))
	rec := post(newPreview(m, f), briefBody("hello"))
	e := decodeErr(t, rec)
	if rec.Code != 422 || e.Error.Code != "segment_unsupported" || e.Error.Message != "No product data." {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
	if f.calls != 0 || len(m.Requests()) != 1 {
		t.Errorf("calls %d reqs %d", f.calls, len(m.Requests()))
	}
}

func TestPreview_Unmapped(t *testing.T) {
	raw := `{"status":"ok","rule":{"op":"and","conditions":[{"field":"city","op":"eq","value":"Istanbul"}]},"unmapped":["likes shoes"]}`
	rec := post(newPreview(llm.NewMock(toolStep(raw)), &fakeCount{}), briefBody("hello"))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"unmapped":["likes shoes"]`) {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
}

func TestPreview_LLMErrors(t *testing.T) {
	cases := []struct {
		name string
		step llm.MockStep
		code int
		ec   string
	}{
		{"error", llm.MockStep{Err: errors.New("secret sdk text")}, 502, "llm_error"},
		{"timeout", llm.MockStep{Err: context.DeadlineExceeded}, 504, "llm_timeout"},
		{"refused", llm.MockStep{Response: llm.Response{Refused: true}}, 502, "llm_refused"},
	}
	for _, c := range cases {
		m := llm.NewMock(c.step)
		rec := post(newPreview(m, &fakeCount{}), briefBody("hello"))
		e := decodeErr(t, rec)
		if rec.Code != c.code || e.Error.Code != c.ec || strings.Contains(rec.Body.String(), "secret") {
			t.Errorf("%s: %d %s", c.name, rec.Code, rec.Body)
		}
		if len(m.Requests()) != 1 {
			t.Errorf("%s: requests %d", c.name, len(m.Requests()))
		}
	}
}

func TestPreview_DBError(t *testing.T) {
	rec := post(newPreview(llm.NewMock(), &fakeCount{err: errors.New("pg: boom")}), briefBody("hello"))
	if rec.Code != 503 || decodeErr(t, rec).Error.Code != "db_unavailable" || strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
}

func TestPreview_PromptShape(t *testing.T) {
	m := llm.NewMock()
	post(newPreview(m, &fakeCount{}), briefBody("  my brief here  "))
	r := m.Requests()[0]
	if strings.Contains(r.System, "my brief here") {
		t.Error("brief in system prompt")
	}
	if r.Turns[0].User != segment.UserMessage("my brief here") {
		t.Errorf("user turn %q", r.Turns[0].User)
	}
}

func TestPreview_LogsNoBrief(t *testing.T) {
	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	defer slog.SetDefault(old)
	rec := post(newPreview(llm.NewMock(), &fakeCount{}), briefBody("secret-brief-text-123"))
	if rec.Code != 200 {
		t.Fatalf("%d", rec.Code)
	}
	if buf.Len() == 0 {
		t.Error("no log output")
	}
	if strings.Contains(buf.String(), "secret-brief-text-123") {
		t.Errorf("brief in logs: %s", buf.String())
	}
}
