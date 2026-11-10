package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"campaigncomposer/internal/delivery"
)

const (
	streamMaxDuration = 30 * time.Minute
	defaultJobsLimit  = 50
	maxJobsLimit      = 200
)

// DeliveryService is the delivery state and the retry action.
type DeliveryService interface {
	Progress(ctx context.Context, id int64) (delivery.Progress, error)
	ListJobs(ctx context.Context, id int64, f delivery.JobFilter) ([]delivery.Job, *int64, error)
	Retry(ctx context.Context, id int64) (int, delivery.Progress, error)
}

type deliveryHandlers struct {
	svc  DeliveryService
	poll time.Duration
}

// RegisterDelivery adds the progress, stream, send-jobs and retry routes.
// poll is the stream check interval.
func RegisterDelivery(mux *http.ServeMux, svc DeliveryService, poll time.Duration) {
	h := deliveryHandlers{svc: svc, poll: poll}
	mux.HandleFunc("GET /api/campaigns/{id}/progress", h.progress)
	mux.HandleFunc("GET /api/campaigns/{id}/progress/stream", h.stream)
	mux.HandleFunc("GET /api/campaigns/{id}/send-jobs", h.sendJobs)
	mux.HandleFunc("POST /api/campaigns/{id}/retry", h.retry)
}

func writeDeliveryError(w http.ResponseWriter, r *http.Request, op string, err error) {
	switch {
	case errors.Is(err, delivery.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "No campaign has this id.", nil)
	case errors.Is(err, delivery.ErrNotSendable):
		writeError(w, http.StatusConflict, "campaign_not_sendable", "This campaign has not been approved, so it has no send state.", nil)
	case errors.Is(err, delivery.ErrNotCompleted):
		writeError(w, http.StatusConflict, "campaign_not_completed", "Only a completed campaign can retry its failed messages.", nil)
	case errors.Is(err, delivery.ErrNothingToRetry):
		writeError(w, http.StatusConflict, "nothing_to_retry", "No failed message can be retried.", nil)
	default:
		slog.ErrorContext(r.Context(), "delivery "+op+" failed", "err", err)
		writeError(w, http.StatusServiceUnavailable, "db_unavailable", "The send state could not be loaded. Try again later.", nil)
	}
}

func (h deliveryHandlers) progress(w http.ResponseWriter, r *http.Request) {
	id, ok := campaignID(w, r)
	if !ok {
		return
	}
	p, err := h.svc.Progress(r.Context(), id)
	if err != nil {
		writeDeliveryError(w, r, "progress", err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h deliveryHandlers) stream(w http.ResponseWriter, r *http.Request) {
	id, ok := campaignID(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), streamMaxDuration)
	defer cancel()
	last, err := h.svc.Progress(ctx, id)
	if err != nil {
		writeDeliveryError(w, r, "progress stream", err)
		return
	}
	sse, ok := newSSEWriter(w)
	if !ok {
		writeError(w, http.StatusInternalServerError, "stream_unsupported", "This server cannot stream.", nil)
		return
	}
	finish := func(p delivery.Progress) bool {
		if p.Status != "completed" {
			return false
		}
		_ = sse.send("done", map[string]any{})
		return true
	}
	if sse.send("progress", last) != nil {
		return
	}
	if finish(last) {
		return
	}
	tick := time.NewTicker(h.poll)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
		p, err := h.svc.Progress(ctx, id)
		if err != nil {
			// Keep the stream open: a short database failure must not end it.
			if ctx.Err() == nil {
				slog.WarnContext(ctx, "delivery progress stream: poll failed", "campaign_id", id, "err", err)
			}
			continue
		}
		if !reflect.DeepEqual(p, last) {
			if sse.send("progress", p) != nil {
				return
			}
			last = p
		}
		if finish(p) {
			return
		}
	}
}

type jobBody struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Channel   string    `json:"channel"`
	Language  string    `json:"language"`
	Status    string    `json:"status"`
	Attempts  int       `json:"attempts"`
	LastError string    `json:"last_error"`
	UpdatedAt time.Time `json:"updated_at"`
}

func badJobsQuery(w http.ResponseWriter) {
	writeError(w, http.StatusBadRequest, "invalid_request",
		"The query must use a valid status, channel (sms or email), limit (1 to 200) and after (a positive integer).", nil)
}

func (h deliveryHandlers) sendJobs(w http.ResponseWriter, r *http.Request) {
	id, ok := campaignID(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	f := delivery.JobFilter{Limit: defaultJobsLimit}
	if s := q.Get("status"); s != "" {
		switch st := delivery.Status(s); st {
		case delivery.StatusQueued, delivery.StatusSending, delivery.StatusRetrying, delivery.StatusSent,
			delivery.StatusFailed, delivery.StatusUnknown, delivery.StatusSkipped:
			f.Status = st
		default:
			badJobsQuery(w)
			return
		}
	}
	if c := q.Get("channel"); c != "" {
		switch ch := delivery.Channel(c); ch {
		case delivery.ChannelSMS, delivery.ChannelEmail:
			f.Channel = ch
		default:
			badJobsQuery(w)
			return
		}
	}
	if s := q.Get("limit"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > maxJobsLimit {
			badJobsQuery(w)
			return
		}
		f.Limit = n
	}
	if s := q.Get("after"); s != "" {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || n <= 0 {
			badJobsQuery(w)
			return
		}
		f.After = n
	}
	jobs, next, err := h.svc.ListJobs(r.Context(), id, f)
	if err != nil {
		writeDeliveryError(w, r, "send jobs", err)
		return
	}
	out := make([]jobBody, 0, len(jobs))
	for _, j := range jobs {
		out = append(out, jobBody{ID: j.ID, UserID: j.UserID, Channel: string(j.Channel), Language: j.Language,
			Status: string(j.Status), Attempts: j.Attempts, LastError: j.LastError, UpdatedAt: j.UpdatedAt})
	}
	writeJSON(w, http.StatusOK, map[string]any{"send_jobs": out, "next_after": next})
}

func (h deliveryHandlers) retry(w http.ResponseWriter, r *http.Request) {
	id, ok := campaignID(w, r)
	if !ok {
		return
	}
	n, p, err := h.svc.Retry(r.Context(), id)
	if err != nil {
		writeDeliveryError(w, r, "retry", err)
		return
	}
	slog.InfoContext(r.Context(), "delivery retry", "campaign_id", id, "retried", n)
	writeJSON(w, http.StatusOK, map[string]any{"retried": n, "progress": p})
}
