# Plugging in a Claude API key

- Session: `f8d6417a-5d33-44f6-9bcd-e3797c483555`
- Model: `claude-sonnet-5-5`
- Time (UTC): 2026-11-11 11:41:37 → 2026-11-11 11:41:49
- Subagent runs: 0

---

## 👤 User  <sub>2026-11-11 11:41:37</sub>

_[IDE: selection]_ 
are we currently plug and play can i plug a claude api key to use this system

**🔧 Bash** ```bash
ls backend/internal/llm; grep -rn "ANTHROPIC\|LLM_" --include=*.go --include=*.yml --include=*.yaml --include=Makefile --include=.env* --include=*.example . 2>/dev/null | grep -v node_modules | grep -v _test | head -40; sed -n 60,130p backend/cmd/api/main.go
```

<details><summary>Result</summary>

```
anthropic.go
anthropic_test.go
llm.go
mock.go
mock_test.go
(eval):1: no matches found: --include=*.go
	// Bound the startup ping, so that an unreachable host cannot delay the listener.
	pingCtx, cancelPing := context.WithTimeout(ctx, startupPingTimeout)
	err = pool.Ping(pingCtx)
	cancelPing()
	if err != nil {
		slog.Warn("initial db ping failed", "err", err)
	}

	// The client connects lazily, so a Redis outage does not stop the api.
	redisOpt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return err
	}
	// Without this option a context deadline does not bound a Redis read. A slow
	// Redis would then hold /readyz and the health checks for the read timeout.
	redisOpt.ContextTimeoutEnabled = true
	rdb := redis.NewClient(redisOpt)
	defer rdb.Close()
	redisPing := dashboard.RedisPinger{Client: rdb}

	mux := http.NewServeMux()
	httpapi.RegisterHealth(mux, pool, redisPing)

	llmClient, err := llm.New(cfg, mockDefault)
	if err != nil {
		return err
	}
	if m, ok := llmClient.(*llm.Mock); ok {
		// Slow, small fragments, so that the demo shows the stream.
		m.SetStream(16, 30*time.Millisecond)
	}
	slog.Info("llm configured", "model", cfg.AnthropicModel, "llm_mode", cfg.LLMMode)
	count := func(ctx context.Context, w segment.Where) (segment.Audience, error) {
		return segment.CountAudience(ctx, pool, w)
	}
	httpapi.RegisterSegments(mux, &httpapi.SegmentPreviewer{
		LLM:        llmClient,
		Allow:      segment.Default(),
		Count:      count,
		Now:        time.Now,
		LLMTimeout: 20 * time.Second,
	})
	campaigns := campaign.NewStore(pool)
	gen := newContentGenerator(llmClient, segment.Default(), count)
	gen.Drafts = campaigns
	httpapi.RegisterContent(mux, gen)
	httpapi.RegisterCampaigns(mux, &campaign.Service{
		Store: campaigns,
		Allow: segment.Default(),
		Count: campaign.AudienceCounter(count),
		Sample: func(ctx context.Context, w segment.Where, langs []string, n int) (map[string][]string, error) {
			return segment.SampleFirstNames(ctx, pool, w, langs, n)
		},
		
… [434 more chars truncated]
```

</details>

**🔧 Bash** ```bash
cat backend/internal/llm/llm.go | head -80; grep -rn "ANTHROPIC\|LLM_MODE" --exclude-dir=node_modules --exclude-dir=.git --exclude-dir=ai-sessions . | grep -v "_test.go" | grep -v "^./docs" | head -20; ls -a | head -30
```

<details><summary>Result</summary>

```
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
	// is built like the one from Generate, e
… [4411 more chars truncated]
```

</details>

## 🤖 Claude  <sub>2026-11-11 11:41:49</sub>

Yes, for the LLM part. The default is mock mode, and a real key needs three steps (from [README.md:31-35](README.md#L31-L35)):

1. Copy `.env.example` to `.env`.
2. Set `LLM_MODE=anthropic`.
3. Set `ANTHROPIC_API_KEY=<your key>`.

Then run `make up`.

- **Mock is the default.** With no `.env`, `LLM_MODE=mock` and the stack runs with no key. The selection happens in `llm.New` ([llm.go](backend/internal/llm/llm.go)), which [main.go](backend/cmd/api/main.go) calls at startup.
- **Real client exists.** [anthropic.go](backend/internal/llm/anthropic.go) is the real client. It is used for segment rules and content generation, including streaming. The default model is `claude-opus-5-5`, and `ANTHROPIC_MODEL` overrides it.
- **Key scope.** Only the `api` service receives the key. The worker gets none, and tests never read it.
- **Delivery stays mocked.** SMS and email still go to the mock provider, so no real person gets a message, even with a real key.

I did not run the system with a real key. The Anthropic client is covered by unit tests (`anthropic_test.go`), but I haven't seen it work against the live API. The progress log only shows verification runs with `LLM_MODE=mock`. Treat your first real-key run as untested. I can try one if you give me a key in `.env`, but I'd rather you set it than paste it here.

