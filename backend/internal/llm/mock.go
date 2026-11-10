package llm

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ExampleRuleJSON is the default answer of the mock: an "ok" answer with the
// example rule of the segment spec.
const ExampleRuleJSON = `{"status":"ok","rule":{"op":"and","conditions":[{"field":"city","op":"eq","value":"Istanbul"},{"field":"last_cart_add_at","op":"within_days","value":30},{"op":"or","conditions":[{"field":"last_purchase_at","op":"is_null"},{"field":"last_purchase_at","op":"older_than_days","value":30}]}]},"unmapped":[]}`

const defaultStreamChunk = 16

// MockStep is one scripted answer of the mock.
type MockStep struct {
	Response Response
	Err      error
}

// MockDefault gives the answer of an unscripted mock call.
type MockDefault func(req Request) MockStep

// Mock is a Client for tests. It is safe for concurrent use.
type Mock struct {
	mu       sync.Mutex
	steps    []MockStep
	scripted bool
	next     int
	requests []Request
	def      MockDefault
	chunk    int
	delay    time.Duration
}

// NewMock returns a mock. With steps, it returns them in order and then fails
// with "mock: no more steps". With no steps, every call returns ExampleRuleJSON.
func NewMock(steps ...MockStep) *Mock {
	return &Mock{steps: steps, scripted: len(steps) > 0, chunk: defaultStreamChunk}
}

// NewMockFunc returns a mock that calls f for each request. f may be called
// from many goroutines at the same time.
func NewMockFunc(f MockDefault) *Mock {
	return &Mock{def: f, chunk: defaultStreamChunk}
}

// SetStream sets how Stream sends the tool input: fragments of chunk bytes
// (16 if chunk <= 0), with delay between them.
func (m *Mock) SetStream(chunk int, delay time.Duration) {
	if chunk <= 0 {
		chunk = defaultStreamChunk
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chunk, m.delay = chunk, delay
}

// take records the request and returns its answer.
func (m *Mock) take(req Request) MockStep {
	m.mu.Lock()
	m.requests = append(m.requests, req)
	if m.scripted {
		defer m.mu.Unlock()
		if m.next >= len(m.steps) {
			return MockStep{Err: errors.New("mock: no more steps")}
		}
		s := m.steps[m.next]
		m.next++
		return s
	}
	def := m.def
	m.mu.Unlock()
	if def != nil {
		// Called without the lock, so slow defaults do not serialize callers.
		return def(req)
	}
	return MockStep{Response: Response{
		ToolUse:   &ToolUse{ID: "toolu_mock", Name: req.Tool.Name, Input: []byte(ExampleRuleJSON)},
		Assistant: "mock-assistant",
	}}
}

// Generate records the request and returns the next answer.
func (m *Mock) Generate(_ context.Context, req Request) (Response, error) {
	s := m.take(req)
	return s.Response, s.Err
}

// Stream records the request, sends the tool input of the next answer in
// fragments, and then returns the answer. A fragment can end inside a
// multi-byte rune.
func (m *Mock) Stream(ctx context.Context, req Request, on func(StreamEvent)) (Response, error) {
	s := m.take(req)
	if s.Err != nil {
		return Response{}, s.Err
	}
	m.mu.Lock()
	chunk, delay := m.chunk, m.delay
	m.mu.Unlock()

	var input []byte
	if s.Response.ToolUse != nil {
		input = s.Response.ToolUse.Input
	}
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	for i := 0; i < len(input); i += chunk {
		if i > 0 && delay > 0 {
			t := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				t.Stop()
				return Response{}, ctx.Err()
			case <-t.C:
			}
		}
		if err := ctx.Err(); err != nil {
			return Response{}, err
		}
		on(StreamEvent{Delta: string(input[i:min(i+chunk, len(input))])})
	}
	return s.Response, nil
}

// Requests returns a copy of the requests that the mock received.
func (m *Mock) Requests() []Request {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Request(nil), m.requests...)
}
