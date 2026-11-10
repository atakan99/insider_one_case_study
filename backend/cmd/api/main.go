// Command api runs the campaign HTTP API.
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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"campaigncomposer/internal/campaign"
	"campaigncomposer/internal/config"
	"campaigncomposer/internal/dashboard"
	"campaigncomposer/internal/delivery"
	"campaigncomposer/internal/httpapi"
	"campaigncomposer/internal/llm"
	"campaigncomposer/internal/segment"
)

const (
	shutdownTimeout    = 10 * time.Second
	startupPingTimeout = 2 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("api stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// pgxpool.New connects lazily. It fails only on an invalid DATABASE_URL.
	// If Postgres is down, the process keeps running and /readyz reports it.
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	// Bound the startup ping, so that an unreachable host cannot delay the listener.
	pingCtx, cancelPing := context.WithTimeout(ctx, startupPingTimeout)
	err = pool.Ping(pingCtx)
	cancelPing()
	if err != nil {
		slog.Warn("initial db ping failed", "err", err)
	}

	// The client connects lazily, so a Redis outage does not stop the api.
	redisOpt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return err
	}
	// Without this option a context deadline does not bound a Redis read. A slow
	// Redis would then hold /readyz and the health checks for the read timeout.
	redisOpt.ContextTimeoutEnabled = true
	rdb := redis.NewClient(redisOpt)
	defer rdb.Close()
	redisPing := dashboard.RedisPinger{Client: rdb}

	mux := http.NewServeMux()
	httpapi.RegisterHealth(mux, pool, redisPing)

	llmClient, err := llm.New(cfg, mockDefault)
	if err != nil {
		return err
	}
	if m, ok := llmClient.(*llm.Mock); ok {
		// Slow, small fragments, so that the demo shows the stream.
		m.SetStream(16, 30*time.Millisecond)
	}
	slog.Info("llm configured", "model", cfg.AnthropicModel, "llm_mode", cfg.LLMMode)
	count := func(ctx context.Context, w segment.Where) (segment.Audience, error) {
		return segment.CountAudience(ctx, pool, w)
	}
	httpapi.RegisterSegments(mux, &httpapi.SegmentPreviewer{
		LLM:        llmClient,
		Allow:      segment.Default(),
		Count:      count,
		Now:        time.Now,
		LLMTimeout: 20 * time.Second,
	})
	campaigns := campaign.NewStore(pool)
	gen := newContentGenerator(llmClient, segment.Default(), count)
	gen.Drafts = campaigns
	httpapi.RegisterContent(mux, gen)
	httpapi.RegisterCampaigns(mux, &campaign.Service{
		Store: campaigns,
		Allow: segment.Default(),
		Count: campaign.AudienceCounter(count),
		Sample: func(ctx context.Context, w segment.Where, langs []string, n int) (map[string][]string, error) {
			return segment.SampleFirstNames(ctx, pool, w, langs, n)
		},
		Now: time.Now,
	})

	queue := delivery.NewQueue(rdb, "api")
	httpapi.RegisterDelivery(mux, &delivery.Service{
		Store: delivery.NewStore(pool),
		Queue: queue,
		Now:   time.Now,
		Log:   slog.Default(),
	}, time.Second)

	dash := dashboard.NewStore(pool)
	httpapi.RegisterDashboard(mux, dash, &dashboard.Checker{
		DB:          pool,
		Redis:       redisPing,
		Queue:       queue,
		Store:       dash,
		HTTP:        &http.Client{},
		WorkerURL:   cfg.WorkerURL,
		ProviderURL: cfg.ProviderURL,
	}, time.Now)

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		dashboard.NewCollector(dash, slog.Default()),
	)
	metrics := httpapi.NewHTTPMetrics(reg)
	mux.Handle("GET /metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	// No WriteTimeout: the content SSE stream can last 90 s (NFR-9).
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.CORS(cfg.CORSOrigins, metrics.Wrap(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errc := make(chan error, 1)
	go func() {
		slog.Info("api listening", "addr", cfg.HTTPAddr, "llm_mode", cfg.LLMMode)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	// Restore default signal handling, so a second Ctrl-C stops the process at once.
	stop()

	slog.Info("api shutting down")
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
