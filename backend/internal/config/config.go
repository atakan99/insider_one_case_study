// Package config loads process settings from environment variables.
package config

import (
	"errors"
	"fmt"
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
