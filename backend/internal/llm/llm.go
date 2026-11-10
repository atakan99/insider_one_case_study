// Package llm is the client for the language model. It has the Anthropic
// client and a mock client for tests and for LLM_MODE=mock.
package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"campaigncomposer/internal/config"
)

// Tool is the one tool that the model can call. Schema is the full JSON schema
// of the tool input: type, properties, required and additionalProperties.
type Tool struct {
	Name        string
	Description string
	Schema      map[string]any
	Eager       bool // send eager_input_streaming: true, so the input streams unbuffered
}

// ToolUse is a tool call that the model made.
type ToolUse struct {
	ID    string
	Name  string
	Input json.RawMessage
}

// ToolResult answers an earlier ToolUse.
type ToolResult struct {
	ToolUseID string
	Content   string
	IsError   bool
}

// Turn is one entry of the history. Set exactly one field.
type Turn struct {
	User       string      // a user text turn, or
	ToolResult *ToolResult // a user tool-result turn, or
	Assistant  any         // an assistant turn, exactly as Response.Assistant gave it
}

// Request is one call to the model.
type Request struct {
	System string
	Tool   Tool
	Turns  []Turn
}

// Response is the answer of the model.
type Response struct {
	ToolUse   *ToolUse // nil if the model did not call the tool
	Assistant any      // opaque; append it as Turn{Assistant: ...} for a repair round
	Refused   bool     // stop_reason == refusal
	CutOff    bool     // stop_reason == max_tokens
}

// StreamEvent is one event of a streaming call.
type StreamEvent struct {
	Delta string // a fragment of the tool input JSON, or
	Reset bool   // a fallback block started; drop the fragments so far
}

// Client calls the model.
type Client interface {
	Generate(ctx context.Context, req Request) (Response, error)
	// Stream is Generate with a callback for each fragment of the tool input.
	// on is called from the calling goroutine, in order. The returned Response
	// is built like the one from Generate, except that ToolUse.Input is the raw
	// input as streamed. It can be cut off or not valid JSON, so the caller's
	// validator must check it.
	Stream(ctx context.Context, req Request, on func(StreamEvent)) (Response, error)
}

// New returns the client for cfg.LLMMode. In mock mode, mockDefault gives the
// answer of each call; nil keeps ExampleRuleJSON for every call.
func New(cfg config.Config, mockDefault MockDefault) (Client, error) {
	switch cfg.LLMMode {
	case config.LLMModeMock:
		if mockDefault != nil {
			return NewMockFunc(mockDefault), nil
		}
		return NewMock(), nil
	case config.LLMModeAnthropic:
		return NewAnthropic(cfg.AnthropicAPIKey, cfg.AnthropicModel), nil
	default:
		return nil, fmt.Errorf("llm: unknown mode %q", cfg.LLMMode)
	}
}
