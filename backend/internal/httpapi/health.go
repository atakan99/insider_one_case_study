// Package httpapi holds the HTTP handlers of the api process.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// pingTimeout bounds the database check in /readyz.
const pingTimeout = 2 * time.Second

// Pinger checks that a dependency is reachable. *pgxpool.Pool satisfies it,
// and so does dashboard.RedisPinger.
type Pinger interface {
	Ping(ctx context.Context) error
}

type healthBody struct {
	Status string `json:"status"`
	DB     string `json:"db,omitempty"`
	Redis  string `json:"redis,omitempty"`
}

// RegisterHealth adds GET /healthz (liveness) and GET /readyz (readiness).
// /readyz is ready only when both the database and Redis answer.
func RegisterHealth(mux *http.ServeMux, db, redis Pinger) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthBody{Status: "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), pingTimeout)
		defer cancel()
		var dbErr, redisErr error
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); dbErr = db.Ping(ctx) }()
		go func() { defer wg.Done(); redisErr = redis.Ping(ctx) }()
		wg.Wait()
		body := healthBody{Status: "ok", DB: "ok", Redis: "ok"}
		// Log the causes, but do not send them to the client.
		if dbErr != nil {
			slog.WarnContext(r.Context(), "readyz: db ping failed", "err", dbErr)
			body.Status, body.DB = "unavailable", "down"
		}
		if redisErr != nil {
			slog.WarnContext(r.Context(), "readyz: redis ping failed", "err", redisErr)
			body.Status, body.Redis = "unavailable", "down"
		}
		code := http.StatusOK
		if body.Status != "ok" {
			code = http.StatusServiceUnavailable
		}
		writeJSON(w, code, body)
	})
}
