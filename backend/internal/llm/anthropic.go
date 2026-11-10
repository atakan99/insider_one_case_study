package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"
)

const maxTokens = 16000

type anthropicClient struct {
	client anthropic.Client
	model  string
}

// NewAnthropic returns a client for the Anthropic API.
func NewAnthropic(apiKey, model string) Client {
	return &anthropicClient{
		client: anthropic.NewClient(option.WithAPIKey(apiKey), option.WithMaxRetries(1)),
		model:  model,
	}
}

// Generate sends the request. It never logs the request.
func (c *anthropicClient) Generate(ctx context.Context, req Request) (Response, error) {
	msg, err := c.client.Beta.Messages.New(ctx, c.buildParams(req))
	if err != nil {
		return Response{}, mapError(err)
	}
	return parseResponse(msg, nil), nil
}

// Stream sends the request as a streaming call. It calls on for each
// input_json_delta and with Reset for each fallback block. It never logs the
// request.
func (c *anthropicClient) Stream(ctx context.Context, req Request, on func(StreamEvent)) (Response, error) {
	stream := c.client.Beta.Messages.NewStreaming(ctx, c.buildParams(req))
	defer stream.Close()

	var msg anthropic.BetaMessage
	// The raw tool input of each block, as streamed. Accumulate replaces an
	// input that is not valid JSON with {}, but the caller must see the real
	// text to report it as cut off or invalid.
	inputs := map[int]string{}
	for stream.Next() {
		ev := stream.Current()
		if err := msg.Accumulate(ev); err != nil {
			return Response{}, fmt.Errorf("llm: %w", err)
		}
		switch e := ev.AsAny().(type) {
		case anthropic.BetaRawContentBlockStartEvent:
			if _, ok := e.ContentBlock.AsAny().(anthropic.BetaFallbackBlock); ok {
				on(StreamEvent{Reset: true})
			}
		case anthropic.BetaRawContentBlockDeltaEvent:
			if d, ok := e.Delta.AsAny().(anthropic.BetaInputJSONDelta); ok && d.PartialJSON != "" {
				inputs[int(e.Index)] += d.PartialJSON
				on(StreamEvent{Delta: d.PartialJSON})
			}
		}
	}
	if err := stream.Err(); err != nil {
		return Response{}, mapError(err)
	}
	return parseResponse(&msg, inputs), nil
}

func mapError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("llm: request timed out: %w", err)
	}
	return fmt.Errorf("llm: %w", err)
}

func (c *anthropicClient) buildParams(req Request) anthropic.BetaMessageNewParams {
	schema := anthropic.BetaToolInputSchemaParam{
		Properties:  req.Tool.Schema["properties"],
		ExtraFields: map[string]any{"additionalProperties": false},
	}
	switch r := req.Tool.Schema["required"].(type) {
	case []string:
		schema.Required = r
	case []any:
		for _, v := range r {
			if s, ok := v.(string); ok {
				schema.Required = append(schema.Required, s)
			}
		}
	}

	tool := anthropic.BetaToolParam{
		Name:        req.Tool.Name,
		Description: anthropic.String(req.Tool.Description),
		InputSchema: schema,
		Strict:      anthropic.Bool(true),
	}
	if req.Tool.Eager {
		tool.EagerInputStreaming = anthropic.Bool(true)
	}

	msgs := make([]anthropic.BetaMessageParam, 0, len(req.Turns))
	for _, t := range req.Turns {
		switch {
		case t.ToolResult != nil:
			msgs = append(msgs, anthropic.NewBetaUserMessage(
				anthropic.NewBetaToolResultBlock(t.ToolResult.ToolUseID, t.ToolResult.Content, t.ToolResult.IsError)))
		case t.Assistant != nil:
			if m, ok := t.Assistant.(anthropic.BetaMessageParam); ok {
				msgs = append(msgs, m)
			}
		default:
			msgs = append(msgs, anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(t.User)))
		}
	}

	return anthropic.BetaMessageNewParams{
		Model:        anthropic.Model(c.model),
		MaxTokens:    maxTokens,
		System:       []anthropic.BetaTextBlockParam{{Text: req.System}},
		Messages:     msgs,
		Tools:        []anthropic.BetaToolUnionParam{{OfTool: &tool}},
		ToolChoice:   anthropic.BetaToolChoiceUnionParam{OfAuto: &anthropic.BetaToolChoiceAutoParam{}},
		OutputConfig: anthropic.BetaOutputConfigParam{Effort: anthropic.BetaOutputConfigEffortLow},
		Fallbacks:    anthropic.BetaFallbacksParamOfDefault(),
		Betas:        []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01},
	}
}

// parseResponse turns the final message into a Response. rawInputs, if not
// nil, gives the streamed raw tool input of a content block by index.
//
// With a server-side fallback, the content has a fallback block at each model
// switch. Only text blocks from before the last fallback block are kept; a
// tool_use before it is not valid. ToolUse is the first tool call after the
// last fallback block. A tool_use whose input is not a JSON object is replayed
// with {} as the input, so that Assistant can be sent back to the API.
func parseResponse(msg *anthropic.BetaMessage, rawInputs map[int]string) Response {
	resp := Response{
		Refused: msg.StopReason == anthropic.BetaStopReasonRefusal,
		CutOff:  msg.StopReason == anthropic.BetaStopReasonMaxTokens,
	}

	lastFallback := -1
	for i, block := range msg.Content {
		if _, ok := block.AsAny().(anthropic.BetaFallbackBlock); ok {
			lastFallback = i
		}
	}

	assistant := msg.ToParam()
	content := make([]anthropic.BetaContentBlockParamUnion, 0, len(msg.Content))
	for i, block := range msg.Content {
		switch v := block.AsAny().(type) {
		case anthropic.BetaTextBlock:
			content = append(content, block.ToParam())
		case anthropic.BetaToolUseBlock:
			if i < lastFallback {
				continue
			}
			raw, ok := rawInputs[i]
			if !ok {
				raw = v.JSON.Input.Raw()
			}
			p := v.ToParam()
			if !isJSONObject(raw) {
				p.Input = map[string]any{}
			}
			content = append(content, anthropic.BetaContentBlockParamUnion{OfToolUse: &p})
			if resp.ToolUse == nil {
				resp.ToolUse = &ToolUse{ID: v.ID, Name: v.Name, Input: json.RawMessage(raw)}
			}
		default:
			if i < lastFallback {
				continue
			}
			content = append(content, block.ToParam())
		}
	}
	assistant.OfAssistant.Content = content
	resp.Assistant = assistant
	return resp
}

func isJSONObject(raw string) bool {
	var obj map[string]json.RawMessage
	return json.Unmarshal([]byte(raw), &obj) == nil && obj != nil
}
