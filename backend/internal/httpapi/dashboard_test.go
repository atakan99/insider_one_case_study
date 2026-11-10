package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"campaigncomposer/internal/dashboard"
	"campaigncomposer/internal/delivery"
)

type fakeStats struct {
	st  dashboard.Stats
	err error
}

func (f fakeStats) Stats(context.Context, time.Time) (dashboard.Stats, error) { return f.st, f.err }

type fakeHealth struct{ h dashboard.Health }

func (f fakeHealth) Check(context.Context) dashboard.Health { return f.h }

func sampleStats() dashboard.Stats {
	mean := 1.5
	return dashboard.Stats{
		GeneratedAt: time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC),
		Totals: map[delivery.Channel]dashboard.ChannelTotals{
			delivery.ChannelSMS:   {Sent: 3, Failed: 1},
			delivery.ChannelEmail: {Sent: 2},
		},
		Throughput: dashboard.Throughput{
			BucketSeconds: 5, Start: time.Date(2026, 11, 10, 9, 50, 0, 0, time.UTC),
			SMS: []float64{0, 1.5}, Email: []float64{0, 0},
		},
		Failures: []dashboard.Failure{{Status: "failed", Reason: "provider_error", SMS: 1}},
		MeanAttempts: map[delivery.Channel]*float64{
			delivery.ChannelSMS: &mean, delivery.ChannelEmail: nil,
		},
		Campaigns: []dashboard.CampaignRow{{ID: 1, Description: "Winter sale", Status: "completed", Total: 4, Sent: 3, Failed: 1}},
	}
}

func sampleHealth() dashboard.Health {
	one := int64(1)
	return dashboard.Health{
		Services: map[string]string{"db": "ok", "redis": "ok", "worker": "down", "provider": "ok"},
		Streams: []dashboard.StreamDepth{
			{Channel: delivery.ChannelSMS, Length: &one, Pending: &one, InProgress: 2},
			{Channel: delivery.ChannelEmail},
		},
	}
}

func getDashboard(t *testing.T, svc DashboardService, hc HealthChecker, path string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	RegisterDashboard(mux, svc, hc, time.Now)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestDashboardStats_OK(t *testing.T) {
	rec := getDashboard(t, fakeStats{st: sampleStats()}, fakeHealth{}, "/api/dashboard/stats")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"generated_at", "totals", "throughput", "failures", "mean_attempts", "campaigns"} {
		if _, ok := body[k]; !ok {
			t.Errorf("body has no key %q", k)
		}
	}
	mean := body["mean_attempts"].(map[string]any)
	if v, ok := mean["email"]; !ok || v != nil {
		t.Errorf("mean_attempts.email = %v (present %v), want null", v, ok)
	}
}

func TestDashboardStats_DBError(t *testing.T) {
	rec := getDashboard(t, fakeStats{err: errors.New("conn refused")}, fakeHealth{}, "/api/dashboard/stats")
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"code":"db_unavailable"`) {
		t.Errorf("body = %s, want code db_unavailable", rec.Body)
	}
	if strings.Contains(rec.Body.String(), "conn refused") {
		t.Errorf("body leaks the internal error: %s", rec.Body)
	}
}

func TestDashboardHealth_AlwaysOK(t *testing.T) {
	rec := getDashboard(t, fakeStats{}, fakeHealth{h: sampleHealth()}, "/api/dashboard/health")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var got dashboard.Health
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Services["worker"] != "down" || got.Services["db"] != "ok" {
		t.Errorf("services = %v", got.Services)
	}
	if len(got.Streams) != 2 || got.Streams[0].Channel != delivery.ChannelSMS || got.Streams[0].InProgress != 2 {
		t.Errorf("streams = %+v", got.Streams)
	}
}

var forbiddenKeys = map[string]bool{
	"user_id": true, "first_name": true, "phone": true, "email_address": true, "subject": true, "body": true,
}

// walk checks every key. A value under the key sms or email must be a number,
// an array of numbers, null, or an object of numbers (NFR-2).
func walk(t *testing.T, path string, v any) {
	t.Helper()
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			if forbiddenKeys[k] {
				t.Errorf("%s has forbidden key %q", path, k)
			}
			if k == "sms" || k == "email" {
				checkNumeric(t, path+"."+k, child, true)
			}
			walk(t, path+"."+k, child)
		}
	case []any:
		for i, child := range x {
			walk(t, path+"[]", child)
			_ = i
		}
	}
}

func checkNumeric(t *testing.T, path string, v any, allowObject bool) {
	t.Helper()
	switch x := v.(type) {
	case nil, float64:
	case []any:
		for _, e := range x {
			if _, ok := e.(float64); !ok {
				t.Errorf("%s has a non-number element %v", path, e)
			}
		}
	case map[string]any:
		if !allowObject {
			t.Errorf("%s is a nested object", path)
			return
		}
		for k, e := range x {
			if _, ok := e.(float64); !ok {
				t.Errorf("%s.%s = %v, want a number", path, k, e)
			}
		}
	default:
		t.Errorf("%s = %v (%T), want number, array of numbers, null or object of numbers", path, v, v)
	}
}

func TestDashboard_NoPersonalData(t *testing.T) {
	for _, path := range []string{"/api/dashboard/stats", "/api/dashboard/health"} {
		rec := getDashboard(t, fakeStats{st: sampleStats()}, fakeHealth{h: sampleHealth()}, path)
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		walk(t, path, body)
	}
}
