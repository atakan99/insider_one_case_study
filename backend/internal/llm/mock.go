package llm

import (
	"context"
	"errors"
	"sync"
)

// ExampleRuleJSON is the default answer of the mock: an "ok" answer with the
// example rule of the segment spec.
const ExampleRuleJSON = `{"status":"ok","rule":{"op":"and","conditions":[{"field":"city","op":"eq","value":"Istanbul"},{"field":"last_cart_add_at","op":"within_days","value":30},{"op":"or","conditions":[{"field":"last_purchase_at","op":"is_null"},{"field":"last_purchase_at","op":"older_than_days","value":30}]}]},"unmapped":[]}`

// MockStep is one scripted answer of the mock.
type MockStep struct {
	Response Response
	Err      error
}

// Mock is a Client for tests. It is safe for concurrent use.
type Mock struct {
	mu       sync.Mutex
	steps    []MockStep
	scripted bool
	next     int
	requests []Request
}

// NewMock returns a mock. With steps, it returns them in order and then fails
// with "mock: no more steps". With no steps, every call returns ExampleRuleJSON.
func NewMock(steps ...MockStep) *Mock {
	return &Mock{steps: steps, scripted: len(steps) > 0}
}

// Generate records the request and returns the next answer.
func (m *Mock) Generate(_ context.Context, req Request) (Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = append(m.requests, req)
	if !m.scripted {
		return Response{
			ToolUse:   &ToolUse{ID: "toolu_mock", Name: req.Tool.Name, Input: []byte(ExampleRuleJSON)},
			Assistant: "mock-assistant",
		}, nil
	}
	if m.next >= len(m.steps) {
		return Response{}, errors.New("mock: no more steps")
	}
	s := m.steps[m.next]
	m.next++
	return s.Response, s.Err
}

// Requests returns a copy of the requests that the mock received.
func (m *Mock) Requests() []Request {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Request(nil), m.requests...)
}
