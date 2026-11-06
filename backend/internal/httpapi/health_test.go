package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakePinger struct {
	err   error
	block bool
}

func (f fakePinger) Ping(ctx context.Context) error {
	if f.block {
		<-ctx.Done()
		return ctx.Err()
	}
	return f.err
}

func serve(t *testing.T, db Pinger, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	RegisterHealth(mux, db)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func assertJSON(t *testing.T, rec *httptest.ResponseRecorder, code int, body string) {
	t.Helper()
	if rec.Code != code {
		t.Errorf("status = %d, want %d", rec.Code, code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != body {
		t.Errorf("body = %s, want %s", got, body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestHealthz_OK(t *testing.T) {
	rec := serve(t, fakePinger{}, http.MethodGet, "/healthz")
	assertJSON(t, rec, http.StatusOK, `{"status":"ok"}`)
}

func TestReadyz_DBUp(t *testing.T) {
	rec := serve(t, fakePinger{}, http.MethodGet, "/readyz")
	assertJSON(t, rec, http.StatusOK, `{"status":"ok","db":"ok"}`)
}

func TestReadyz_DBDown(t *testing.T) {
	rec := serve(t, fakePinger{err: errors.New("conn refused")}, http.MethodGet, "/readyz")
	assertJSON(t, rec, http.StatusServiceUnavailable, `{"status":"unavailable","db":"down"}`)
	if strings.Contains(rec.Body.String(), "conn refused") {
		t.Errorf("body leaks the internal error: %s", rec.Body.String())
	}
}

func TestReadyz_PingTimeout(t *testing.T) {
	start := time.Now()
	rec := serve(t, fakePinger{block: true}, http.MethodGet, "/readyz")
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("readyz took %s, want < 3s", elapsed)
	}
	assertJSON(t, rec, http.StatusServiceUnavailable, `{"status":"unavailable","db":"down"}`)
}

func TestHealthz_MethodNotAllowed(t *testing.T) {
	rec := serve(t, fakePinger{}, http.MethodPost, "/healthz")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
