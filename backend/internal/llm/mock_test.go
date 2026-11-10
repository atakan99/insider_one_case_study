package llm

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
)

func TestMock_Sequence(t *testing.T) {
	m := NewMock(
		MockStep{Response: Response{ToolUse: &ToolUse{ID: "a", Name: "t", Input: json.RawMessage(`{"n":1}`)}}},
		MockStep{Err: errors.New("boom")},
		MockStep{Response: Response{Refused: true}},
	)
	ctx := context.Background()
	r1, err := m.Generate(ctx, Request{System: "s1"})
	if err != nil || r1.ToolUse == nil || r1.ToolUse.ID != "a" {
		t.Fatalf("step 1 = %+v, %v", r1, err)
	}
	if _, err := m.Generate(ctx, Request{System: "s2"}); err == nil || err.Error() != "boom" {
		t.Fatalf("step 2 err = %v, want boom", err)
	}
	r3, err := m.Generate(ctx, Request{System: "s3"})
	if err != nil || !r3.Refused {
		t.Fatalf("step 3 = %+v, %v", r3, err)
	}
	if _, err := m.Generate(ctx, Request{}); err == nil || err.Error() != "mock: no more steps" {
		t.Fatalf("after last step err = %v, want mock: no more steps", err)
	}
	reqs := m.Requests()
	if len(reqs) != 4 || reqs[0].System != "s1" || reqs[1].System != "s2" || reqs[2].System != "s3" {
		t.Fatalf("requests = %+v", reqs)
	}
}

func TestMock_Default(t *testing.T) {
	m := NewMock()
	for i := 0; i < 3; i++ {
		r, err := m.Generate(context.Background(), Request{Tool: Tool{Name: "submit_segment"}})
		if err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
		if r.ToolUse == nil || r.ToolUse.Name != "submit_segment" {
			t.Fatalf("call %d: tool use = %+v", i, r.ToolUse)
		}
		if string(r.ToolUse.Input) != ExampleRuleJSON {
			t.Fatalf("call %d: input = %s", i, r.ToolUse.Input)
		}
	}
}

func TestExampleRuleJSON_Shape(t *testing.T) {
	var v struct {
		Status   string          `json:"status"`
		Rule     json.RawMessage `json:"rule"`
		Unmapped []string        `json:"unmapped"`
	}
	if err := json.Unmarshal([]byte(ExampleRuleJSON), &v); err != nil {
		t.Fatal(err)
	}
	if v.Status != "ok" || len(v.Rule) == 0 || v.Unmapped == nil || len(v.Unmapped) != 0 {
		t.Fatalf("unexpected shape: %+v", v)
	}
}

func TestMock_Concurrent(t *testing.T) {
	m := NewMock()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := m.Generate(context.Background(), Request{}); err != nil {
				t.Error(err)
			}
			_ = m.Requests()
		}()
	}
	wg.Wait()
	if n := len(m.Requests()); n != 50 {
		t.Fatalf("requests = %d, want 50", n)
	}
}
