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
	return parseResponse(msg), nil
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

func parseResponse(msg *anthropic.BetaMessage) Response {
	resp := Response{
		Assistant: msg.ToParam(),
		Refused:   msg.StopReason == anthropic.BetaStopReasonRefusal,
	}
	for _, block := range msg.Content {
		if v, ok := block.AsAny().(anthropic.BetaToolUseBlock); ok {
			resp.ToolUse = &ToolUse{ID: v.ID, Name: v.Name, Input: json.RawMessage(v.JSON.Input.Raw())}
			break
		}
	}
	return resp
}
