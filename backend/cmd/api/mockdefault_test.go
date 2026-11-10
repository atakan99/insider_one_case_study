package main

import (
	"testing"
	"time"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/llm"
	"campaigncomposer/internal/segment"
)

func TestMockDefault(t *testing.T) {
	seg := mockDefault(llm.Request{Tool: llm.Tool{Name: segment.ToolName}, Turns: []llm.Turn{{User: "x"}}})
	if seg.Err != nil || seg.Response.ToolUse == nil || string(seg.Response.ToolUse.Input) != llm.ExampleRuleJSON {
		t.Fatalf("segment step = %+v", seg)
	}
	if seg.Response.ToolUse.Name != segment.ToolName || seg.Response.Assistant != "mock-assistant" {
		t.Fatalf("segment step fields = %+v", seg.Response)
	}
	for _, lang := range []string{"tr", "en", "xx"} {
		msg := content.UserMessage(lang, "desc", 30, "brief")
		s := mockDefault(llm.Request{Tool: llm.Tool{Name: content.ToolName}, Turns: []llm.Turn{{User: msg}}})
		if s.Err != nil || s.Response.ToolUse == nil || s.Response.ToolUse.Name != content.ToolName {
			t.Fatalf("%s: step = %+v", lang, s)
		}
		if _, _, errs := content.Validate(s.Response.ToolUse.Input, 30); len(errs) > 0 {
			t.Fatalf("%s: invalid content: %v", lang, errs)
		}
	}
	// No language line and no turns fall back to en.
	s := mockDefault(llm.Request{Tool: llm.Tool{Name: content.ToolName}})
	if _, _, errs := content.Validate(s.Response.ToolUse.Input, 30); len(errs) > 0 {
		t.Fatalf("fallback invalid: %v", errs)
	}
}

func TestNewContentGenerator(t *testing.T) {
	g := newContentGenerator(llm.NewMock(), segment.Default(), nil)
	if g.LLMTimeout != 45*time.Second {
		t.Fatalf("LLMTimeout = %v", g.LLMTimeout)
	}
	if g.DraftEvery != 100*time.Millisecond {
		t.Fatalf("DraftEvery = %v", g.DraftEvery)
	}
	if g.Now == nil || g.Clock == nil {
		t.Fatal("Now and Clock must be set")
	}
}
