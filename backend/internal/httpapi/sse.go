package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
)

// sseWriter serialises events from many goroutines and flushes after each one.
type sseWriter struct {
	mu      sync.Mutex
	w       http.ResponseWriter
	flusher http.Flusher
}

// newSSEWriter sends the stream headers and status 200. It returns false,
// and writes nothing, if w cannot flush.
func newSSEWriter(w http.ResponseWriter) (*sseWriter, bool) {
	f, ok := w.(http.Flusher)
	if !ok {
		return nil, false
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	f.Flush()
	return &sseWriter{w: w, flusher: f}, true
}

// send writes one event with data as a single JSON line, then flushes.
func (s *sseWriter) send(event string, data any) error {
	// json.Marshal escapes control characters, so the data is one line.
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	b.WriteString("event: ")
	b.WriteString(event)
	b.WriteString("\ndata: ")
	b.Write(raw)
	b.WriteString("\n\n")

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.w.Write(b.Bytes()); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}
