// Command mockprovider stands in for an external SMS and email provider.
// See internal/mockprovider and docs/specs/03-mock-provider/goal.md.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"campaigncomposer/internal/config"
	"campaigncomposer/internal/mockprovider"
)

const shutdownTimeout = 10 * time.Second

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("mockprovider stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadMockProvider(os.Getenv)
	if err != nil {
		return err
	}
	addr := cfg.HTTPAddr
	slog.Info("mockprovider settings",
		"rate_per_sec", cfg.RatePerSec, "burst", cfg.Burst,
		"error_rate", cfg.ErrorRate, "slow_rate", cfg.SlowRate, "lost_rate", cfg.LostRate,
		"slow_delay", cfg.SlowDelay.String(), "latency_max", cfg.LatencyMax.String(),
		"random", cfg.Random)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// No WriteTimeout: slow responses wait for cfg.SlowDelay, and the server
	// must not end them itself.
	srv := &http.Server{
		Addr: addr,
		Handler: mockprovider.New(mockprovider.Options{
			RatePerSec: cfg.RatePerSec,
			Burst:      cfg.Burst,
			ErrorRate:  cfg.ErrorRate,
			SlowRate:   cfg.SlowRate,
			LostRate:   cfg.LostRate,
			SlowDelay:  cfg.SlowDelay,
			LatencyMax: cfg.LatencyMax,
			Random:     cfg.Random,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errc := make(chan error, 1)
	go func() {
		slog.Info("mockprovider listening", "addr", addr)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	// Restore default signal handling, so a second Ctrl-C stops the process at once.
	stop()

	slog.Info("mockprovider shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	// A listen error can race the signal. Do not exit 0 if the server never served.
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
