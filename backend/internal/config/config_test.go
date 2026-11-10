package config

import (
	"strings"
	"testing"
	"time"
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

func TestLoadMockProvider_Defaults(t *testing.T) {
	cfg, err := LoadMockProvider(env(map[string]string{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPAddr != ":8081" {
		t.Errorf("HTTPAddr = %q, want :8081", cfg.HTTPAddr)
	}
	if cfg.RatePerSec != 100 || cfg.Burst != 10 {
		t.Errorf("rate = %v/%d, want 100/10", cfg.RatePerSec, cfg.Burst)
	}
	if cfg.ErrorRate != 0.03 || cfg.SlowRate != 0.03 || cfg.LostRate != 0.01 {
		t.Errorf("rates = %v %v %v, want 0.03 0.03 0.01", cfg.ErrorRate, cfg.SlowRate, cfg.LostRate)
	}
	if cfg.SlowDelay != 5*time.Second || cfg.LatencyMax != 50*time.Millisecond {
		t.Errorf("durations = %v %v, want 5s 50ms", cfg.SlowDelay, cfg.LatencyMax)
	}
}

func TestLoadMockProvider_Values(t *testing.T) {
	cfg, err := LoadMockProvider(env(map[string]string{
		"HTTP_ADDR":         ":9000",
		"MOCK_RATE_PER_SEC": "20.5",
		"MOCK_RATE_BURST":   "3",
		"MOCK_ERROR_RATE":   "0.1",
		"MOCK_SLOW_RATE":    "0.2",
		"MOCK_LOST_RATE":    "0.3",
		"MOCK_SLOW_DELAY":   "2s",
		"MOCK_LATENCY_MAX":  "0",
		"MOCK_RANDOM":       "7",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := MockProviderConfig{
		HTTPAddr: ":9000", RatePerSec: 20.5, Burst: 3,
		ErrorRate: 0.1, SlowRate: 0.2, LostRate: 0.3,
		SlowDelay: 2 * time.Second, LatencyMax: 0, Random: 7,
	}
	if cfg != want {
		t.Errorf("cfg = %+v, want %+v", cfg, want)
	}
}

func TestLoadMockProvider_Errors(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"error rate over 1", map[string]string{"MOCK_ERROR_RATE": "1.5"}, "MOCK_ERROR_RATE"},
		{"slow rate negative", map[string]string{"MOCK_SLOW_RATE": "-0.1"}, "MOCK_SLOW_RATE"},
		{"lost rate not a number", map[string]string{"MOCK_LOST_RATE": "abc"}, "MOCK_LOST_RATE"},
		{"sum over 1", map[string]string{"MOCK_ERROR_RATE": "0.5", "MOCK_SLOW_RATE": "0.4", "MOCK_LOST_RATE": "0.2"}, "sum"},
		{"slow delay without unit", map[string]string{"MOCK_SLOW_DELAY": "5"}, "MOCK_SLOW_DELAY"},
		{"latency negative", map[string]string{"MOCK_LATENCY_MAX": "-1s"}, "MOCK_LATENCY_MAX"},
		{"rate zero", map[string]string{"MOCK_RATE_PER_SEC": "0"}, "MOCK_RATE_PER_SEC"},
		{"burst zero", map[string]string{"MOCK_RATE_BURST": "0"}, "MOCK_RATE_BURST"},
		{"random negative", map[string]string{"MOCK_RANDOM": "-3"}, "MOCK_RANDOM"},
		{"error rate NaN", map[string]string{"MOCK_ERROR_RATE": "NaN"}, "MOCK_ERROR_RATE"},
		{"slow rate Inf", map[string]string{"MOCK_SLOW_RATE": "Inf"}, "MOCK_SLOW_RATE"},
		{"rate per sec Inf", map[string]string{"MOCK_RATE_PER_SEC": "+Inf"}, "MOCK_RATE_PER_SEC"},
		{"rate per sec NaN", map[string]string{"MOCK_RATE_PER_SEC": "NaN"}, "MOCK_RATE_PER_SEC"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadMockProvider(env(tc.env))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want error that mentions %q", err, tc.want)
			}
		})
	}
}

func TestLoad_LLMSettings(t *testing.T) {
	cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AnthropicModel != "claude-opus-5-5" {
		t.Errorf("AnthropicModel = %q", cfg.AnthropicModel)
	}
	want := []string{"http://localhost:3000", "http://localhost:5173"}
	if strings.Join(cfg.CORSOrigins, "|") != strings.Join(want, "|") {
		t.Errorf("CORSOrigins = %v, want %v", cfg.CORSOrigins, want)
	}

	cfg, err = Load(env(map[string]string{
		"DATABASE_URL":      "postgres://x",
		"LLM_MODE":          "anthropic",
		"ANTHROPIC_API_KEY": "test-key",
		"ANTHROPIC_MODEL":   "claude-sonnet-5-5",
		"CORS_ORIGINS":      " https://a.example , https://b.example,, ",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AnthropicModel != "claude-sonnet-5-5" || cfg.AnthropicAPIKey != "test-key" {
		t.Errorf("cfg = %+v", cfg)
	}
	want = []string{"https://a.example", "https://b.example"}
	if strings.Join(cfg.CORSOrigins, "|") != strings.Join(want, "|") {
		t.Errorf("CORSOrigins = %v, want %v", cfg.CORSOrigins, want)
	}
}
