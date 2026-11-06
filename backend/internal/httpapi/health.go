// Package httpapi holds the HTTP handlers of the api process.
package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// pingTimeout bounds the database check in /readyz.
const pingTimeout = 2 * time.Second

// Pinger checks that a dependency is reachable. *pgxpool.Pool satisfies it.
type Pinger interface {
	Ping(ctx context.Context) error
}

type healthBody struct {
	Status string `json:"status"`
	DB     string `json:"db,omitempty"`
}

// RegisterHealth adds GET /healthz (liveness) and GET /readyz (readiness).
func RegisterHealth(mux *http.ServeMux, db Pinger) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthBody{Status: "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), pingTimeout)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			// Log the cause, but do not send it to the client.
			slog.WarnContext(r.Context(), "readyz: db ping failed", "err", err)
			writeJSON(w, http.StatusServiceUnavailable, healthBody{Status: "unavailable", DB: "down"})
			return
		}
		writeJSON(w, http.StatusOK, healthBody{Status: "ok", DB: "ok"})
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
