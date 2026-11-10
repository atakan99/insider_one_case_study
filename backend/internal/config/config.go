// Package config loads process settings from environment variables.
package config

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// LLM modes.
const (
	LLMModeMock      = "mock"
	LLMModeAnthropic = "anthropic"
)

// Config holds the settings that the api process reads at startup.
type Config struct {
	HTTPAddr    string
	DatabaseURL string
	// RedisURL is REDIS_URL. It is required.
	RedisURL        string
	LLMMode         string
	AnthropicAPIKey string
	// AnthropicModel is ANTHROPIC_MODEL. The default is claude-opus-5-5.
	AnthropicModel string
	// CORSOrigins is CORS_ORIGINS, a comma-separated list of allowed origins.
	CORSOrigins []string
}

// Load reads the config through getenv. Pass os.Getenv in production and a
// map-backed function in tests.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		HTTPAddr:        getenv("HTTP_ADDR"),
		DatabaseURL:     getenv("DATABASE_URL"),
		RedisURL:        getenv("REDIS_URL"),
		LLMMode:         getenv("LLM_MODE"),
		AnthropicAPIKey: getenv("ANTHROPIC_API_KEY"),
		AnthropicModel:  getenv("ANTHROPIC_MODEL"),
	}
	if cfg.AnthropicModel == "" {
		cfg.AnthropicModel = "claude-opus-5-5"
	}
	origins := getenv("CORS_ORIGINS")
	if origins == "" {
		origins = "http://localhost:3000,http://localhost:5173,http://127.0.0.1:3000,http://127.0.0.1:5173"
	}
	for _, o := range strings.Split(origins, ",") {
		if o = strings.TrimSpace(o); o != "" {
			cfg.CORSOrigins = append(cfg.CORSOrigins, o)
		}
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}
	if cfg.LLMMode == "" {
		cfg.LLMMode = LLMModeMock
	}

	var errs []error
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	if cfg.RedisURL == "" {
		errs = append(errs, errors.New("REDIS_URL is required"))
	}
	switch cfg.LLMMode {
	case LLMModeMock:
	case LLMModeAnthropic:
		if cfg.AnthropicAPIKey == "" {
			errs = append(errs, errors.New("ANTHROPIC_API_KEY is required when LLM_MODE=anthropic"))
		}
	default:
		errs = append(errs, fmt.Errorf("LLM_MODE must be %q or %q, got %q", LLMModeMock, LLMModeAnthropic, cfg.LLMMode))
	}
	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}
	return cfg, nil
}

// SeedConfig holds the settings that the seed process reads at startup.
type SeedConfig struct {
	DatabaseURL string
	Users       int
	Random      uint64
	Reset       bool
}

// LoadSeed reads the seed config through getenv.
func LoadSeed(getenv func(string) string) (SeedConfig, error) {
	cfg := SeedConfig{DatabaseURL: getenv("DATABASE_URL"), Users: 5000, Random: 42}
	var errs []error
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	if v := getenv("SEED_USERS"); v != "" {
		n, err := strconv.Atoi(v)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("SEED_USERS must be an integer, got %q", v))
		case n < 1:
			errs = append(errs, fmt.Errorf("SEED_USERS must be at least 1, got %d", n))
		default:
			cfg.Users = n
		}
	}
	if v := getenv("SEED_RANDOM"); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			errs = append(errs, fmt.Errorf("SEED_RANDOM must be an unsigned integer, got %q", v))
		} else {
			cfg.Random = n
		}
	}
	if v := getenv("SEED_RESET"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("SEED_RESET must be a boolean, got %q", v))
		} else {
			cfg.Reset = b
		}
	}
	if len(errs) > 0 {
		return SeedConfig{}, errors.Join(errs...)
	}
	return cfg, nil
}

// MockProviderConfig holds the settings that the mock provider reads at startup.
type MockProviderConfig struct {
	HTTPAddr   string
	RatePerSec float64
	Burst      int
	ErrorRate  float64
	SlowRate   float64
	LostRate   float64
	SlowDelay  time.Duration
	LatencyMax time.Duration
	// Random is the seed for the failure picker. It comes from MOCK_RANDOM,
	// or from the clock if MOCK_RANDOM is not set.
	Random uint64
}

// LoadMockProvider reads the mock provider config through getenv.
func LoadMockProvider(getenv func(string) string) (MockProviderConfig, error) {
	cfg := MockProviderConfig{
		HTTPAddr:   getenv("HTTP_ADDR"),
		RatePerSec: 100,
		Burst:      10,
		ErrorRate:  0.03,
		SlowRate:   0.03,
		LostRate:   0.01,
		SlowDelay:  5 * time.Second,
		LatencyMax: 50 * time.Millisecond,
		Random:     uint64(time.Now().UnixNano()),
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8081"
	}

	var errs []error
	if v := getenv("MOCK_RATE_PER_SEC"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) || f <= 0 {
			errs = append(errs, fmt.Errorf("MOCK_RATE_PER_SEC must be a number greater than 0, got %q", v))
		} else {
			cfg.RatePerSec = f
		}
	}
	if v := getenv("MOCK_RATE_BURST"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			errs = append(errs, fmt.Errorf("MOCK_RATE_BURST must be an integer of at least 1, got %q", v))
		} else {
			cfg.Burst = n
		}
	}
	for _, r := range []struct {
		key string
		dst *float64
	}{
		{"MOCK_ERROR_RATE", &cfg.ErrorRate},
		{"MOCK_SLOW_RATE", &cfg.SlowRate},
		{"MOCK_LOST_RATE", &cfg.LostRate},
	} {
		v := getenv(r.key)
		if v == "" {
			continue
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || math.IsNaN(f) || f < 0 || f > 1 {
			errs = append(errs, fmt.Errorf("%s must be a number from 0 to 1, got %q", r.key, v))
			continue
		}
		*r.dst = f
	}
	if sum := cfg.ErrorRate + cfg.SlowRate + cfg.LostRate; sum > 1 {
		errs = append(errs, fmt.Errorf("the sum of MOCK_ERROR_RATE, MOCK_SLOW_RATE and MOCK_LOST_RATE must be 1 or less, got %g", sum))
	}
	for _, d := range []struct {
		key string
		dst *time.Duration
	}{
		{"MOCK_SLOW_DELAY", &cfg.SlowDelay},
		{"MOCK_LATENCY_MAX", &cfg.LatencyMax},
	} {
		v := getenv(d.key)
		if v == "" {
			continue
		}
		dur, err := time.ParseDuration(v)
		if err != nil || dur < 0 {
			errs = append(errs, fmt.Errorf("%s must be a duration of 0 or more, for example 5s, got %q", d.key, v))
			continue
		}
		*d.dst = dur
	}
	if v := getenv("MOCK_RANDOM"); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			errs = append(errs, fmt.Errorf("MOCK_RANDOM must be an unsigned integer, got %q", v))
		} else {
			cfg.Random = n
		}
	}
	if len(errs) > 0 {
		return MockProviderConfig{}, errors.Join(errs...)
	}
	return cfg, nil
}

// WorkerConfig holds the settings that the worker process reads at startup.
type WorkerConfig struct {
	HTTPAddr    string
	DatabaseURL string
	RedisURL    string
	ProviderURL string
	// RatePerSec is SEND_RATE_PER_SEC, the send rate of all consumers together.
	RatePerSec int
	// MaxAttempts is SEND_MAX_ATTEMPTS, the number of provider calls for one job.
	MaxAttempts int
	// Consumers is SEND_CONSUMERS, the number of goroutines that send.
	Consumers int
	// Timeout is SEND_TIMEOUT, the timeout of one provider call.
	Timeout time.Duration
}

// LoadWorker reads the worker config through getenv.
func LoadWorker(getenv func(string) string) (WorkerConfig, error) {
	cfg := WorkerConfig{
		HTTPAddr:    getenv("HTTP_ADDR"),
		DatabaseURL: getenv("DATABASE_URL"),
		RedisURL:    getenv("REDIS_URL"),
		ProviderURL: getenv("PROVIDER_URL"),
		RatePerSec:  100,
		MaxAttempts: 5,
		Consumers:   32,
		Timeout:     3 * time.Second,
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8082"
	}
	var errs []error
	for _, r := range []struct{ key, val string }{
		{"DATABASE_URL", cfg.DatabaseURL},
		{"REDIS_URL", cfg.RedisURL},
		{"PROVIDER_URL", cfg.ProviderURL},
	} {
		if r.val == "" {
			errs = append(errs, fmt.Errorf("%s is required", r.key))
		}
	}
	for _, r := range []struct {
		key      string
		min, max int
		dst      *int
	}{
		{"SEND_RATE_PER_SEC", 1, 100, &cfg.RatePerSec},
		{"SEND_MAX_ATTEMPTS", 1, 10, &cfg.MaxAttempts},
		{"SEND_CONSUMERS", 1, 256, &cfg.Consumers},
	} {
		v := getenv(r.key)
		if v == "" {
			continue
		}
		n, err := strconv.Atoi(v)
		if err != nil || n < r.min || n > r.max {
			errs = append(errs, fmt.Errorf("%s must be an integer from %d to %d, got %q", r.key, r.min, r.max, v))
			continue
		}
		*r.dst = n
	}
	if v := getenv("SEND_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			errs = append(errs, fmt.Errorf("SEND_TIMEOUT must be a duration above 0, for example 3s, got %q", v))
		} else {
			cfg.Timeout = d
		}
	}
	if len(errs) > 0 {
		return WorkerConfig{}, errors.Join(errs...)
	}
	return cfg, nil
}
