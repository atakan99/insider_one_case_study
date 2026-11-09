// Package config loads process settings from environment variables.
package config

import (
	"errors"
	"fmt"
	"strconv"
)

// LLM modes.
const (
	LLMModeMock      = "mock"
	LLMModeAnthropic = "anthropic"
)

// Config holds the settings that the api process reads at startup.
type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	LLMMode         string
	AnthropicAPIKey string
}

// Load reads the config through getenv. Pass os.Getenv in production and a
// map-backed function in tests.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		HTTPAddr:        getenv("HTTP_ADDR"),
		DatabaseURL:     getenv("DATABASE_URL"),
		LLMMode:         getenv("LLM_MODE"),
		AnthropicAPIKey: getenv("ANTHROPIC_API_KEY"),
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
