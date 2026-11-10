package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"campaigncomposer/internal/dashboard"
)

// DashboardService reads the dashboard statistics.
type DashboardService interface {
	Stats(ctx context.Context, now time.Time) (dashboard.Stats, error)
}

// HealthChecker checks the services and the queue depth.
type HealthChecker interface {
	Check(ctx context.Context) dashboard.Health
}

type dashboardHandlers struct {
	stats  DashboardService
	health HealthChecker
	now    func() time.Time
}

// RegisterDashboard adds GET /api/dashboard/stats and GET /api/dashboard/health.
// The health route answers 200 even when a service is down.
func RegisterDashboard(mux *http.ServeMux, stats DashboardService, health HealthChecker, now func() time.Time) {
	h := dashboardHandlers{stats: stats, health: health, now: now}
	mux.HandleFunc("GET /api/dashboard/stats", h.getStats)
	mux.HandleFunc("GET /api/dashboard/health", h.getHealth)
}

func (h dashboardHandlers) getStats(w http.ResponseWriter, r *http.Request) {
	st, err := h.stats.Stats(r.Context(), h.now())
	if err != nil {
		slog.ErrorContext(r.Context(), "dashboard stats failed", "err", err)
		writeError(w, http.StatusServiceUnavailable, "db_unavailable", "The statistics could not be loaded. Try again later.", nil)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (h dashboardHandlers) getHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.health.Check(r.Context()))
}
