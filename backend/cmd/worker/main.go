// Command worker plans approved campaigns and sends their messages.
// See docs/specs/07-send-queue/goal.md.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"campaigncomposer/internal/config"
	"campaigncomposer/internal/delivery"
	"campaigncomposer/internal/segment"
)

const (
	checkTimeout    = time.Second
	httpStopTimeout = 5 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("worker stopped", "err", err)
		os.Exit(1)
	}
}

type pinger interface {
	Ping(ctx context.Context) error
}

type redisPinger struct{ rdb *redis.Client }

func (r redisPinger) Ping(ctx context.Context) error { return r.rdb.Ping(ctx).Err() }

// healthHandler answers GET /healthz with the state of Postgres and Redis,
// and GET /metrics with metrics.
func healthHandler(db pinger, rdb *redis.Client, metrics http.Handler) http.Handler {
	check := func(ctx context.Context, p pinger) string {
		ctx, cancel := context.WithTimeout(ctx, checkTimeout)
		defer cancel()
		if err := p.Ping(ctx); err != nil {
			return "down"
		}
		return "ok"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		body := map[string]string{"status": "ok", "db": check(r.Context(), db), "redis": check(r.Context(), redisPinger{rdb})}
		code := http.StatusOK
		if body["db"] != "ok" || body["redis"] != "ok" {
			body["status"] = "unavailable"
			code = http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(body)
	})
	mux.Handle("GET /metrics", metrics)
	return mux
}

func run() error {
	cfg, err := config.LoadWorker(os.Getenv)
	if err != nil {
		return err
	}
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("REDIS_URL: %w", err)
	}

	// Each consumer holds one connection while it waits in a blocking read.
	opt.PoolSize = max(opt.PoolSize, 2*cfg.Consumers+16)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// pgxpool.New connects lazily. /healthz reports a database that is down.
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	rdb := redis.NewClient(opt)
	defer rdb.Close()

	host, _ := os.Hostname()
	queue := delivery.NewQueue(rdb, fmt.Sprintf("%s-%d", host, os.Getpid()))
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	metrics := delivery.NewMetrics(reg, queue)
	w := &delivery.Worker{
		Store:    delivery.NewStore(pool),
		Queue:    queue,
		Metrics:  metrics,
		Provider: delivery.NewProvider(cfg.ProviderURL, cfg.Timeout),
		Allow:    segment.Default(),
		Clock:    delivery.RealClock(),
		Cfg:      cfg,
		Log:      slog.Default(),
	}

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           healthHandler(pool, rdb, promhttp.HandlerFor(reg, promhttp.HandlerOpts{})),
		ReadHeaderTimeout: 5 * time.Second,
	}
	errc := make(chan error, 1)
	go func() {
		slog.Info("worker listening", "addr", cfg.HTTPAddr)
		errc <- srv.ListenAndServe()
	}()
	runErr := make(chan error, 1)
	go func() {
		slog.Info("worker settings", "rate_per_sec", cfg.RatePerSec, "max_attempts", cfg.MaxAttempts,
			"consumers", cfg.Consumers, "timeout", cfg.Timeout.String())
		runErr <- w.Run(ctx)
	}()

	select {
	case err := <-errc:
		// The listener failed. Stop the worker, then report the error.
		stop()
		<-runErr
		return err
	case err := <-runErr:
		slog.Info("worker stopped")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), httpStopTimeout)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return err
	}
}
