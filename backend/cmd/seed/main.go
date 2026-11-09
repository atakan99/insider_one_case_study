// Command seed applies the migrations and loads fake users and events.
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/config"
	"campaigncomposer/internal/seed"
	"campaigncomposer/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("seed failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadSeed(os.Getenv)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := store.Migrate(ctx, cfg.DatabaseURL); err != nil {
		return err
	}
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	inserted, err := seed.Load(ctx, pool, seed.Options{Random: cfg.Random, Now: time.Now().UTC(), Users: cfg.Users}, cfg.Reset)
	if err != nil {
		return err
	}
	if !inserted {
		slog.Info("users exist, skipping")
		return nil
	}
	var users, events int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM users), (SELECT count(*) FROM events)`).Scan(&users, &events); err != nil {
		return err
	}
	slog.Info("seeded", "users", users, "events", events)
	return nil
}
