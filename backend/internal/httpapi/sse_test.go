package httpapi

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type sseEvent struct {
	Event string
	Data  json.RawMessage
}

// parseSSE splits a response body into events. Each event must be exactly
// "event: name\ndata: {json}\n\n".
func parseSSE(t *testing.T, body string) []sseEvent {
	t.Helper()
	if body == "" {
		return nil
	}
	if !strings.HasSuffix(body, "\n\n") {
		t.Fatalf("body does not end with a blank line: %q", body)
	}
	var out []sseEvent
	for _, block := range strings.Split(strings.TrimSuffix(body, "\n\n"), "\n\n") {
		lines := strings.Split(block, "\n")
		if len(lines) != 2 || !strings.HasPrefix(lines[0], "event: ") || !strings.HasPrefix(lines[1], "data: ") {
			t.Fatalf("bad event block %q", block)
		}
		data := strings.TrimPrefix(lines[1], "data: ")
		if !json.Valid([]byte(data)) {
			t.Fatalf("data is not JSON: %q", data)
		}
		out = append(out, sseEvent{Event: strings.TrimPrefix(lines[0], "event: "), Data: json.RawMessage(data)})
	}
	return out
}

func TestSSEWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	s, ok := newSSEWriter(rec)
	if !ok {
		t.Fatal("recorder must support flushing")
	}
	var wg sync.WaitGroup
	for g := range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 100 {
				if err := s.send("x", map[string]any{"g": g, "i": i, "text": "a\nb"}); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("content type %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("cache control %q", cc)
	}
	if !rec.Flushed {
		t.Error("not flushed")
	}
	evs := parseSSE(t, rec.Body.String())
	if len(evs) != 200 {
		t.Fatalf("got %d events, want 200", len(evs))
	}
	seen := map[[2]int]bool{}
	for _, e := range evs {
		var d struct{ G, I int }
		if e.Event != "x" || json.Unmarshal(e.Data, &d) != nil {
			t.Fatalf("event %+v", e)
		}
		seen[[2]int{d.G, d.I}] = true
	}
	if len(seen) != 200 {
		t.Errorf("distinct events %d", len(seen))
	}
}
