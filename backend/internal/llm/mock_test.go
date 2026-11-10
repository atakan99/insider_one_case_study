package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
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

func TestMock_Stream(t *testing.T) {
	input := `{"sms":"abcdefghij","email_subject":"x"}` // 40 bytes
	if len(input) != 40 {
		t.Fatalf("input len = %d", len(input))
	}
	want := Response{ToolUse: &ToolUse{ID: "a", Name: "t", Input: json.RawMessage(input)}, Assistant: "asst"}
	m := NewMock(MockStep{Response: want})
	m.SetStream(16, 0)
	var deltas []string
	got, err := m.Stream(context.Background(), Request{System: "s"}, func(ev StreamEvent) {
		if ev.Reset {
			t.Errorf("unexpected reset")
		}
		deltas = append(deltas, ev.Delta)
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(deltas) != 3 {
		t.Fatalf("deltas = %q, want 3", deltas)
	}
	if strings.Join(deltas, "") != input {
		t.Errorf("joined = %q", strings.Join(deltas, ""))
	}
	if got.ToolUse != want.ToolUse || got.Assistant != "asst" {
		t.Errorf("response = %+v", got)
	}
	if reqs := m.Requests(); len(reqs) != 1 || reqs[0].System != "s" {
		t.Errorf("requests = %+v", reqs)
	}
}

func TestMock_StreamDefaultChunk(t *testing.T) {
	m := NewMock() // ExampleRuleJSON, default chunk 16
	var n int
	var sb strings.Builder
	if _, err := m.Stream(context.Background(), Request{}, func(ev StreamEvent) { n++; sb.WriteString(ev.Delta) }); err != nil {
		t.Fatal(err)
	}
	if want := (len(ExampleRuleJSON) + 15) / 16; n != want {
		t.Errorf("events = %d, want %d", n, want)
	}
	if sb.String() != ExampleRuleJSON {
		t.Errorf("joined = %q", sb.String())
	}
}

func TestMock_StreamStepErr(t *testing.T) {
	m := NewMock(MockStep{Err: errors.New("boom")})
	_, err := m.Stream(context.Background(), Request{}, func(StreamEvent) { t.Error("no events expected") })
	if err == nil || err.Error() != "boom" {
		t.Fatalf("err = %v", err)
	}
}

func TestMock_StreamCancel(t *testing.T) {
	m := NewMock()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := m.Stream(ctx, Request{System: "c"}, func(StreamEvent) {})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if reqs := m.Requests(); len(reqs) != 1 || reqs[0].System != "c" {
		t.Errorf("requests = %+v", reqs)
	}
}

func TestMock_StreamCancelDuringDelay(t *testing.T) {
	m := NewMock()
	m.SetStream(4, time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	var n int
	_, err := m.Stream(ctx, Request{}, func(StreamEvent) { n++; cancel() })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if n != 1 {
		t.Errorf("events = %d, want 1", n)
	}
}

func TestMock_Func(t *testing.T) {
	var mu sync.Mutex
	var seen []string
	m := NewMockFunc(func(req Request) MockStep {
		mu.Lock()
		seen = append(seen, req.System)
		mu.Unlock()
		return MockStep{Response: Response{ToolUse: &ToolUse{ID: "f", Name: req.Tool.Name, Input: json.RawMessage(`{"s":"` + req.System + `"}`)}}}
	})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sys := fmt.Sprintf("req-%d", i)
			var r Response
			var err error
			if i%2 == 0 {
				r, err = m.Generate(context.Background(), Request{System: sys, Tool: Tool{Name: "tool"}})
			} else {
				r, err = m.Stream(context.Background(), Request{System: sys, Tool: Tool{Name: "tool"}}, func(StreamEvent) {})
			}
			if err != nil {
				t.Error(err)
				return
			}
			if r.ToolUse == nil || string(r.ToolUse.Input) != `{"s":"`+sys+`"}` {
				t.Errorf("%s: response = %+v", sys, r.ToolUse)
			}
		}(i)
	}
	wg.Wait()
	if len(seen) != 20 || len(m.Requests()) != 20 {
		t.Fatalf("seen = %d, requests = %d", len(seen), len(m.Requests()))
	}
}
