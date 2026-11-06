package config

import (
	"strings"
	"testing"
)

func env(m map[string]string) func(string) string {
	return func(key string) string { return m[key] }
}

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":8080")
	}
	if cfg.LLMMode != "mock" {
		t.Errorf("LLMMode = %q, want %q", cfg.LLMMode, "mock")
	}
	if cfg.DatabaseURL != "postgres://x" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://x")
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	_, err := Load(env(map[string]string{}))
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("err = %v, want error that mentions DATABASE_URL", err)
	}
}

func TestLoad_AnthropicWithoutKey(t *testing.T) {
	_, err := Load(env(map[string]string{
		"DATABASE_URL": "postgres://x",
		"LLM_MODE":     "anthropic",
	}))
	if err == nil || !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Fatalf("err = %v, want error that mentions ANTHROPIC_API_KEY", err)
	}
}

func TestLoad_InvalidLLMMode(t *testing.T) {
	_, err := Load(env(map[string]string{
		"DATABASE_URL": "postgres://x",
		"LLM_MODE":     "gpt",
	}))
	if err == nil || !strings.Contains(err.Error(), "LLM_MODE") {
		t.Fatalf("err = %v, want error that mentions LLM_MODE", err)
	}
}
