package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	r := parseResponse(decode(t, toolUseBody), nil, "submit_segment")
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

	r = parseResponse(decode(t, textBody), nil, "submit_segment")
	if r.ToolUse != nil || r.Refused || r.Assistant == nil {
		t.Errorf("text response = %+v", r)
	}

	r = parseResponse(decode(t, refusalBody), nil, "submit_segment")
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
	c, err := New(config.Config{LLMMode: config.LLMModeMock}, nil)
	if err != nil {
		t.Fatal(err)
	}
	m, ok := c.(*Mock)
	if !ok {
		t.Fatalf("mock mode client = %T", c)
	}
	if r, err := m.Generate(context.Background(), Request{}); err != nil || string(r.ToolUse.Input) != ExampleRuleJSON {
		t.Errorf("nil default: %+v, %v", r, err)
	}

	c, err = New(config.Config{LLMMode: config.LLMModeMock}, func(Request) MockStep {
		return MockStep{Response: Response{Refused: true}}
	})
	if err != nil {
		t.Fatal(err)
	}
	if r, err := c.Generate(context.Background(), Request{}); err != nil || !r.Refused {
		t.Errorf("custom default: %+v, %v", r, err)
	}

	c, err = New(config.Config{LLMMode: config.LLMModeAnthropic, AnthropicAPIKey: "k", AnthropicModel: "m"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.(*anthropicClient); !ok {
		t.Errorf("anthropic mode client = %T", c)
	}
	if _, err := New(config.Config{LLMMode: "x"}, nil); err == nil {
		t.Error("want error for unknown mode")
	}
}

func TestAnthropic_BuildParams_Eager(t *testing.T) {
	c := &anthropicClient{model: "m"}
	tool := testTool()
	tool.Eager = true
	p := c.buildParams(Request{Tool: tool})
	tp := p.Tools[0].OfTool
	if !tp.EagerInputStreaming.Valid() || !tp.EagerInputStreaming.Value {
		t.Errorf("EagerInputStreaming = %+v", tp.EagerInputStreaming)
	}
	if !tp.Strict.Valid() || !tp.Strict.Value {
		t.Errorf("Strict = %+v", tp.Strict)
	}
	wire := marshal(t, p)["tools"].([]any)[0].(map[string]any)
	if wire["eager_input_streaming"] != true || wire["strict"] != true {
		t.Errorf("tool = %v", wire)
	}

	// Without Eager the field is not sent.
	p = c.buildParams(Request{Tool: testTool()})
	wire = marshal(t, p)["tools"].([]any)[0].(map[string]any)
	if _, ok := wire["eager_input_streaming"]; ok {
		t.Errorf("eager_input_streaming must not be set: %v", wire)
	}
}

const cutOffBody = `{"id":"msg_4","type":"message","role":"assistant","model":"claude-opus-5-5",
"content":[{"type":"tool_use","id":"toolu_4","name":"submit_content","input":{"sms":"a"}}],
"stop_reason":"max_tokens","stop_sequence":null,"usage":{"input_tokens":1,"output_tokens":1}}`

func TestAnthropic_ParseResponse_CutOff(t *testing.T) {
	r := parseResponse(decode(t, cutOffBody), nil, "submit_content")
	if !r.CutOff || r.Refused {
		t.Errorf("response = %+v", r)
	}
	if r := parseResponse(decode(t, toolUseBody), nil, "submit_segment"); r.CutOff {
		t.Errorf("tool_use stop must not be CutOff")
	}
}

const fallbackBody = `{"id":"msg_5","type":"message","role":"assistant","model":"claude-opus-5-5",
"content":[
 {"type":"text","text":"before"},
 {"type":"tool_use","id":"toolu_old","name":"submit_content","input":{"sms":"old"}},
 {"type":"fallback","from":{"model":"claude-opus-5-5"},"to":{"model":"claude-sonnet-5"},"trigger":{"type":"refusal","category":null}},
 {"type":"text","text":"after"},
 {"type":"tool_use","id":"toolu_new","name":"submit_content","input":{"sms":"new"}}
],
"stop_reason":"tool_use","stop_sequence":null,"usage":{"input_tokens":1,"output_tokens":1}}`

func assistantBlocks(t *testing.T, a any) []map[string]any {
	t.Helper()
	p, ok := a.(anthropic.BetaMessageParam)
	if !ok {
		t.Fatalf("assistant = %T", a)
	}
	m := marshal(t, p)
	if m["role"] != "assistant" {
		t.Errorf("role = %v", m["role"])
	}
	var out []map[string]any
	for _, b := range m["content"].([]any) {
		out = append(out, b.(map[string]any))
	}
	return out
}

func TestAnthropic_ParseResponse_Fallback(t *testing.T) {
	r := parseResponse(decode(t, fallbackBody), nil, "submit_content")
	if r.ToolUse == nil || r.ToolUse.ID != "toolu_new" {
		t.Fatalf("tool use = %+v", r.ToolUse)
	}
	if string(r.ToolUse.Input) != `{"sms":"new"}` {
		t.Errorf("input = %s", r.ToolUse.Input)
	}
	blocks := assistantBlocks(t, r.Assistant)
	var types []string
	for _, b := range blocks {
		types = append(types, fmt.Sprint(b["type"]))
		if b["id"] == "toolu_old" {
			t.Errorf("tool_use before fallback was kept: %v", b)
		}
	}
	if got := strings.Join(types, ","); got != "text,fallback,text,tool_use" {
		t.Errorf("assistant block types = %s", got)
	}
}

const nonObjectInputBody = `{"id":"msg_6","type":"message","role":"assistant","model":"claude-opus-5-5",
"content":[{"type":"tool_use","id":"toolu_6","name":"submit_content","input":"not an object"}],
"stop_reason":"tool_use","stop_sequence":null,"usage":{"input_tokens":1,"output_tokens":1}}`

func TestAnthropic_AssistantInvalidInput(t *testing.T) {
	r := parseResponse(decode(t, nonObjectInputBody), nil, "submit_content")
	if r.ToolUse == nil || string(r.ToolUse.Input) != `"not an object"` {
		t.Fatalf("tool use = %+v", r.ToolUse)
	}
	blocks := assistantBlocks(t, r.Assistant)
	if len(blocks) != 1 {
		t.Fatalf("blocks = %v", blocks)
	}
	if in, ok := blocks[0]["input"].(map[string]any); !ok || len(in) != 0 {
		t.Errorf("replayed input = %#v, want {}", blocks[0]["input"])
	}

	// A raw input override that is cut off (the streaming case) is kept in
	// ToolUse, and the replayed param still has {}.
	r = parseResponse(decode(t, cutOffBody), map[int]string{0: `{"sms":"a`}, "submit_content")
	if string(r.ToolUse.Input) != `{"sms":"a` {
		t.Errorf("tool use input = %s", r.ToolUse.Input)
	}
	blocks = assistantBlocks(t, r.Assistant)
	if in, ok := blocks[0]["input"].(map[string]any); !ok || len(in) != 0 {
		t.Errorf("replayed input = %#v, want {}", blocks[0]["input"])
	}
}

func sse(events ...string) string {
	var sb strings.Builder
	for _, e := range events {
		var v struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal([]byte(e), &v)
		fmt.Fprintf(&sb, "event: %s\ndata: %s\n\n", v.Type, e)
	}
	return sb.String()
}

func streamClient(t *testing.T, body string) *anthropicClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req["stream"] != true {
			t.Errorf("request stream = %v (%v)", req["stream"], err)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return &anthropicClient{
		client: anthropic.NewClient(option.WithBaseURL(srv.URL), option.WithAPIKey("test"), option.WithMaxRetries(0)),
		model:  "claude-opus-5-5",
	}
}

const msgStart = `{"type":"message_start","message":{"id":"msg_s","type":"message","role":"assistant","model":"claude-opus-5-5","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":1,"output_tokens":1}}}`

func TestAnthropic_Stream(t *testing.T) {
	body := sse(
		msgStart,
		`{"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_old","name":"submit_content","input":{}}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"sms\":\"ol"}}`,
		`{"type":"content_block_stop","index":0}`,
		`{"type":"content_block_start","index":1,"content_block":{"type":"fallback","from":{"model":"claude-opus-5-5"},"to":{"model":"claude-sonnet-5"},"trigger":{"type":"refusal","category":null}}}`,
		`{"type":"content_block_stop","index":1}`,
		`{"type":"content_block_start","index":2,"content_block":{"type":"tool_use","id":"toolu_new","name":"submit_content","input":{}}}`,
		`{"type":"content_block_delta","index":2,"delta":{"type":"input_json_delta","partial_json":"{\"sms\":"}}`,
		`{"type":"content_block_delta","index":2,"delta":{"type":"input_json_delta","partial_json":"\"new\"}"}}`,
		`{"type":"content_block_stop","index":2}`,
		`{"type":"message_delta","delta":{"stop_reason":"tool_use","stop_sequence":null},"usage":{"output_tokens":5}}`,
		`{"type":"message_stop"}`,
	)
	c := streamClient(t, body)
	var evs []StreamEvent
	r, err := c.Stream(context.Background(), Request{Tool: Tool{Name: "submit_content"}}, func(ev StreamEvent) { evs = append(evs, ev) })
	if err != nil {
		t.Fatal(err)
	}
	want := []StreamEvent{{Delta: `{"sms":"ol`}, {Reset: true}, {Delta: `{"sms":`}, {Delta: `"new"}`}}
	if fmt.Sprint(evs) != fmt.Sprint(want) {
		t.Errorf("events = %+v, want %+v", evs, want)
	}
	if r.ToolUse == nil || r.ToolUse.ID != "toolu_new" || string(r.ToolUse.Input) != `{"sms":"new"}` {
		t.Errorf("tool use = %+v", r.ToolUse)
	}
	if r.CutOff || r.Refused {
		t.Errorf("response = %+v", r)
	}
	for _, b := range assistantBlocks(t, r.Assistant) {
		if b["id"] == "toolu_old" {
			t.Errorf("tool_use before fallback was kept")
		}
	}
}

func TestAnthropic_StreamCutOff(t *testing.T) {
	body := sse(
		msgStart,
		`{"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_c","name":"submit_content","input":{}}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"sms\":\"cut"}}`,
		`{"type":"content_block_stop","index":0}`,
		`{"type":"message_delta","delta":{"stop_reason":"max_tokens","stop_sequence":null},"usage":{"output_tokens":5}}`,
		`{"type":"message_stop"}`,
	)
	r, err := streamClient(t, body).Stream(context.Background(), Request{Tool: Tool{Name: "submit_content"}}, func(StreamEvent) {})
	if err != nil {
		t.Fatal(err)
	}
	if !r.CutOff {
		t.Errorf("CutOff = false")
	}
	if r.ToolUse == nil || string(r.ToolUse.Input) != `{"sms":"cut` {
		t.Fatalf("tool use = %+v", r.ToolUse)
	}
	blocks := assistantBlocks(t, r.Assistant)
	if in, ok := blocks[0]["input"].(map[string]any); !ok || len(in) != 0 {
		t.Errorf("replayed input = %#v, want {}", blocks[0]["input"])
	}
}

const wrongToolBody = `{"id":"msg_7","type":"message","role":"assistant","model":"claude-opus-5-5",
"content":[
 {"type":"text","text":"hi"},
 {"type":"tool_use","id":"toolu_x","name":"other_tool","input":{"a":1}}
],
"stop_reason":"tool_use","stop_sequence":null,"usage":{"input_tokens":1,"output_tokens":1}}`

const twoCallsBody = `{"id":"msg_8","type":"message","role":"assistant","model":"claude-opus-5-5",
"content":[
 {"type":"tool_use","id":"toolu_wrong","name":"other_tool","input":{"a":1}},
 {"type":"text","text":"hi"},
 {"type":"tool_use","id":"toolu_a","name":"submit_content","input":{"sms":"a"}},
 {"type":"tool_use","id":"toolu_b","name":"submit_content","input":{"sms":"b"}}
],
"stop_reason":"tool_use","stop_sequence":null,"usage":{"input_tokens":1,"output_tokens":1}}`

func TestAnthropic_ParseResponse_WrongToolName(t *testing.T) {
	r := parseResponse(decode(t, wrongToolBody), nil, "submit_content")
	if r.ToolUse != nil {
		t.Fatalf("tool use = %+v, want nil", r.ToolUse)
	}
	blocks := assistantBlocks(t, r.Assistant)
	if len(blocks) != 1 || blocks[0]["type"] != "text" {
		t.Errorf("blocks = %v, want only the text block", blocks)
	}
}

func TestAnthropic_ParseResponse_OneToolCall(t *testing.T) {
	r := parseResponse(decode(t, twoCallsBody), nil, "submit_content")
	if r.ToolUse == nil || r.ToolUse.ID != "toolu_a" || string(r.ToolUse.Input) != `{"sms":"a"}` {
		t.Fatalf("tool use = %+v", r.ToolUse)
	}
	blocks := assistantBlocks(t, r.Assistant)
	var got []string
	for _, b := range blocks {
		got = append(got, fmt.Sprint(b["type"], ":", b["id"]))
	}
	if s := strings.Join(got, ","); s != "text:<nil>,tool_use:toolu_a" {
		t.Errorf("assistant blocks = %s", s)
	}
}
