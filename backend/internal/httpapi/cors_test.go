package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func corsHandler() http.Handler {
	return CORS([]string{"http://localhost:3000"}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func TestCORS_KnownOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/segments/preview", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	corsHandler().ServeHTTP(rec, req)
	h := rec.Header()
	if rec.Code != 204 || h.Get("Access-Control-Allow-Origin") != "http://localhost:3000" ||
		!strings.Contains(h.Get("Access-Control-Allow-Methods"), "POST") ||
		!strings.Contains(h.Get("Access-Control-Allow-Headers"), "Content-Type") ||
		h.Get("Access-Control-Max-Age") != "600" || h.Get("Vary") != "Origin" {
		t.Errorf("%d %v", rec.Code, h)
	}

	req = httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec = httptest.NewRecorder()
	corsHandler().ServeHTTP(rec, req)
	if rec.Code != 200 || rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("simple request: %d %v", rec.Code, rec.Header())
	}
}

func TestCORS_UnknownOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/segments/preview", nil)
	req.Header.Set("Origin", "http://evil.test")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	corsHandler().ServeHTTP(rec, req)
	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "" {
		t.Errorf("allow-origin %q", v)
	}
	if rec.Header().Get("Vary") != "Origin" {
		t.Error("missing Vary")
	}
}
