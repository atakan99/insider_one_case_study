package llm

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"

	"campaigncomposer/internal/config"
)

func testTool() Tool {
	return Tool{
		Name:        "submit_segment",
		Description: "Submit the segment.",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{"type": "string", "enum": []any{"ok", "unsupported"}},
			},
			"required":             []any{"status"},
			"additionalProperties": false,
		},
	}
}

func marshal(t *testing.T, v any) map[string]any {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestAnthropic_BuildParams(t *testing.T) {
	c := &anthropicClient{model: "claude-opus-5-5"}
	p := c.buildParams(Request{
		System: "SYSTEM TEXT",
		Tool:   testTool(),
		Turns: []Turn{
			{User: "<brief>x</brief>"},
			{Assistant: anthropic.BetaMessageParam{OfAssistant: &anthropic.BetaAssistantMessageParam{
				Content: []anthropic.BetaContentBlockParamUnion{anthropic.NewBetaToolUseBlock("tu_1", map[string]any{"status": "bad"}, "submit_segment")},
			}}},
			{ToolResult: &ToolResult{ToolUseID: "tu_1", Content: "errors here", IsError: true}},
		},
	})
	m := marshal(t, p)

	if m["model"] != "claude-opus-5-5" {
		t.Errorf("model = %v", m["model"])
	}
	if m["max_tokens"] != float64(16000) {
		t.Errorf("max_tokens = %v", m["max_tokens"])
	}
	if _, ok := m["thinking"]; ok {
		t.Errorf("thinking must not be set")
	}
	if oc, _ := m["output_config"].(map[string]any); oc["effort"] != "low" {
		t.Errorf("output_config = %v", m["output_config"])
	}
	if tc, _ := m["tool_choice"].(map[string]any); tc["type"] != "auto" {
		t.Errorf("tool_choice = %v", m["tool_choice"])
	}
	if m["fallbacks"] != "default" {
		t.Errorf("fallbacks = %v, want default", m["fallbacks"])
	}
	found := false
	for _, b := range p.Betas {
		if b == anthropic.AnthropicBetaServerSideFallback2026_07_01 {
			found = true
		}
	}
	if !found {
		t.Errorf("betas = %v, want server-side-fallback-2026-07-01", p.Betas)
	}

	sys, _ := m["system"].([]any)
	if len(sys) != 1 || sys[0].(map[string]any)["text"] != "SYSTEM TEXT" {
		t.Errorf("system = %v", m["system"])
	}

	tools, _ := m["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("tools = %v", m["tools"])
	}
	tool := tools[0].(map[string]any)
	if tool["name"] != "submit_segment" || tool["strict"] != true {
		t.Errorf("tool = %v", tool)
	}
	schema := tool["input_schema"].(map[string]any)
	if schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Errorf("input_schema = %v", schema)
	}
	if _, ok := schema["properties"].(map[string]any)["status"]; !ok {
		t.Errorf("input_schema properties = %v", schema["properties"])
	}
	if req, _ := schema["required"].([]any); len(req) != 1 || req[0] != "status" {
		t.Errorf("required = %v", schema["required"])
	}

	msgs, _ := m["messages"].([]any)
	if len(msgs) != 3 {
		t.Fatalf("messages = %v", msgs)
	}
	if msgs[0].(map[string]any)["role"] != "user" || msgs[1].(map[string]any)["role"] != "assistant" {
		t.Errorf("roles = %v", msgs)
	}
	last := msgs[2].(map[string]any)
	if last["role"] != "user" {
		t.Errorf("tool result role = %v", last["role"])
	}
	content := last["content"].([]any)
	if len(content) != 1 {
		t.Fatalf("tool result content = %v", content)
	}
	blk := content[0].(map[string]any)
	if blk["type"] != "tool_result" || blk["tool_use_id"] != "tu_1" || blk["is_error"] != true {
		t.Errorf("tool_result block = %v", blk)
	}
}

func TestAnthropic_BuildParams_ToolResultNotError(t *testing.T) {
	c := &anthropicClient{model: "m"}
	p := c.buildParams(Request{Tool: testTool(), Turns: []Turn{{ToolResult: &ToolResult{ToolUseID: "x", Content: "ok"}}}})
	msgs := marshal(t, p)["messages"].([]any)
	blk := msgs[0].(map[string]any)["content"].([]any)[0].(map[string]any)
	if v, ok := blk["is_error"]; ok && v != false {
		t.Errorf("is_error = %v", v)
	}
}

const toolUseBody = `{"id":"msg_1","type":"message","role":"assistant","model":"claude-opus-5-5",
"content":[{"type":"text","text":"ok"},{"type":"tool_use","id":"toolu_1","name":"submit_segment","input":{"status":"unsupported","reason":"no data"}}],
"stop_reason":"tool_use","stop_sequence":null,"usage":{"input_tokens":1,"output_tokens":1}}`

const textBody = `{"id":"msg_2","type":"message","role":"assistant","model":"claude-opus-5-5",
"content":[{"type":"text","text":"hello"}],"stop_reason":"end_turn","stop_sequence":null,"usage":{"input_tokens":1,"output_tokens":1}}`

const refusalBody = `{"id":"msg_3","type":"message","role":"assistant","model":"claude-opus-5-5",
"content":[],"stop_reason":"refusal","stop_sequence":null,"usage":{"input_tokens":1,"output_tokens":1}}`

func decode(t *testing.T, body string) *anthropic.BetaMessage {
	t.Helper()
	var msg anthropic.BetaMessage
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		t.Fatal(err)
	}
	return &msg
}

func TestAnthropic_ParseResponse(t *testing.T) {
	r := parseResponse(decode(t, toolUseBody))
	if r.Refused || r.ToolUse == nil {
		t.Fatalf("tool use response = %+v", r)
	}
	if r.ToolUse.ID != "toolu_1" || r.ToolUse.Name != "submit_segment" {
		t.Errorf("tool use = %+v", r.ToolUse)
	}
	var in map[string]any
	if err := json.Unmarshal(r.ToolUse.Input, &in); err != nil || in["status"] != "unsupported" {
		t.Errorf("input = %s (%v)", r.ToolUse.Input, err)
	}
	if _, ok := r.Assistant.(anthropic.BetaMessageParam); !ok {
		t.Errorf("assistant = %T, want BetaMessageParam", r.Assistant)
	}

	r = parseResponse(decode(t, textBody))
	if r.ToolUse != nil || r.Refused || r.Assistant == nil {
		t.Errorf("text response = %+v", r)
	}

	r = parseResponse(decode(t, refusalBody))
	if !r.Refused || r.ToolUse != nil {
		t.Errorf("refusal response = %+v", r)
	}
}

func TestAnthropic_ErrorsKeepDeadline(t *testing.T) {
	err := mapError(context.DeadlineExceeded)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(err.Error(), "llm") {
		t.Errorf("err = %v", err)
	}
}

func TestNew(t *testing.T) {
	c, err := New(config.Config{LLMMode: config.LLMModeMock})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.(*Mock); !ok {
		t.Errorf("mock mode client = %T", c)
	}
	c, err = New(config.Config{LLMMode: config.LLMModeAnthropic, AnthropicAPIKey: "k", AnthropicModel: "m"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.(*anthropicClient); !ok {
		t.Errorf("anthropic mode client = %T", c)
	}
	if _, err := New(config.Config{LLMMode: "x"}); err == nil {
		t.Error("want error for unknown mode")
	}
}

var _ = option.WithMaxRetries
