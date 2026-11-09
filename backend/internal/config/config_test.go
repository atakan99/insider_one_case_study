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

func TestLoadSeed_Defaults(t *testing.T) {
	cfg, err := LoadSeed(env(map[string]string{"DATABASE_URL": "postgres://x"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := SeedConfig{DatabaseURL: "postgres://x", Users: 5000, Random: 42, Reset: false}
	if cfg != want {
		t.Errorf("cfg = %+v, want %+v", cfg, want)
	}
}

func TestLoadSeed_Values(t *testing.T) {
	cfg, err := LoadSeed(env(map[string]string{
		"DATABASE_URL": "postgres://x",
		"SEED_USERS":   "12",
		"SEED_RANDOM":  "7",
		"SEED_RESET":   "true",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := SeedConfig{DatabaseURL: "postgres://x", Users: 12, Random: 7, Reset: true}
	if cfg != want {
		t.Errorf("cfg = %+v, want %+v", cfg, want)
	}
}

func TestLoadSeed_Errors(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"missing url", map[string]string{}, "DATABASE_URL"},
		{"users not a number", map[string]string{"DATABASE_URL": "x", "SEED_USERS": "abc"}, "SEED_USERS"},
		{"users zero", map[string]string{"DATABASE_URL": "x", "SEED_USERS": "0"}, "SEED_USERS"},
		{"random negative", map[string]string{"DATABASE_URL": "x", "SEED_RANDOM": "-1"}, "SEED_RANDOM"},
		{"reset bad", map[string]string{"DATABASE_URL": "x", "SEED_RESET": "maybe"}, "SEED_RESET"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadSeed(env(tc.env))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want error that mentions %s", err, tc.want)
			}
		})
	}
}
