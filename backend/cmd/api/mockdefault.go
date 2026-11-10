package main

import (
	"time"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/httpapi"
	"campaigncomposer/internal/llm"
	"campaigncomposer/internal/segment"
)

// mockDefault answers an unscripted mock call by tool name.
func mockDefault(req llm.Request) llm.MockStep {
	input := []byte(llm.ExampleRuleJSON)
	if req.Tool.Name == content.ToolName {
		lang := "en"
		if len(req.Turns) > 0 {
			if l, ok := content.LanguageOf(req.Turns[0].User); ok {
				lang = l
			}
		}
		input = content.MockContent(lang)
	}
	return llm.MockStep{Response: llm.Response{
		ToolUse:   &llm.ToolUse{ID: "toolu_mock", Name: req.Tool.Name, Input: input},
		Assistant: "mock-assistant",
	}}
}

// newContentGenerator builds the handler of POST /api/content/generate.
func newContentGenerator(client llm.Client, allow segment.Allowlist, count httpapi.AudienceCounter) *httpapi.ContentGenerator {
	return &httpapi.ContentGenerator{
		LLM:        client,
		Allow:      allow,
		Count:      count,
		Now:        time.Now,
		Clock:      time.Now,
		LLMTimeout: 45 * time.Second,
		DraftEvery: 100 * time.Millisecond,
	}
}
