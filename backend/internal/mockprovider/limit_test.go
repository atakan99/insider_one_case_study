package mockprovider

import (
	"net/http"
	"strconv"
	"testing"
	"time"
)

func limitOptions(clock *fakeClock, perSec float64, burst int) Options {
	o := normalOptions(clock)
	o.RatePerSec = perSec
	o.Burst = burst
	return o
}

func TestRateLimit_Burst(t *testing.T) {
	srv := startServer(t, limitOptions(newFakeClock(), 100, 10))
	accepted, limited := 0, 0
	for range 200 {
		resp := post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil)
		switch resp.StatusCode {
		case http.StatusAccepted:
			accepted++
		case http.StatusTooManyRequests:
			limited++
			ra, err := strconv.Atoi(resp.Header.Get("Retry-After"))
			if err != nil || ra < 1 {
				t.Fatalf("Retry-After = %q, want an integer of 1 or more", resp.Header.Get("Retry-After"))
			}
			if e := decode[errorBody](t, resp); e.Error.Code != "rate_limited" {
				t.Fatalf("code = %q, want rate_limited", e.Error.Code)
			}
		default:
			t.Fatalf("status = %d", resp.StatusCode)
		}
	}
	if accepted != 10 || limited != 190 {
		t.Fatalf("accepted %d, limited %d, want 10 and 190", accepted, limited)
	}
	st := stats(t, srv.URL)["sms"]
	if st.RateLimited != 190 || st.Requests != 200 || st.Delivered != 10 {
		t.Fatalf("sms = %+v", st)
	}
}

func TestRateLimit_Steady(t *testing.T) {
	clock := newFakeClock()
	srv := startServer(t, limitOptions(clock, 100, 10))
	for i := range 100 {
		clock.Advance(10 * time.Millisecond)
		if resp := post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil); resp.StatusCode != http.StatusAccepted {
			t.Fatalf("request %d: status %d, want 202", i, resp.StatusCode)
		}
	}
}

func TestRateLimit_RetryAfterRoundsUp(t *testing.T) {
	// At 0.5 per second, the next token comes after 2 s.
	srv := startServer(t, limitOptions(newFakeClock(), 0.5, 1))
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil), http.StatusAccepted)
	resp := post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil)
	wantStatus(t, resp, http.StatusTooManyRequests)
	if ra := resp.Header.Get("Retry-After"); ra != "2" {
		t.Fatalf("Retry-After = %q, want 2", ra)
	}
}

func TestRateLimit_PerChannel(t *testing.T) {
	srv := startServer(t, limitOptions(newFakeClock(), 100, 2))
	for range 2 {
		wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil), http.StatusAccepted)
	}
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil), http.StatusTooManyRequests)
	wantStatus(t, post(t, srv.URL, "/v1/email", email("ali@example.com", "s", "b", ""), nil), http.StatusAccepted)
}

func TestRateLimit_429UsesNoToken(t *testing.T) {
	srv := startServer(t, limitOptions(newFakeClock(), 100, 1))
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), map[string]string{"X-Mock-Fail": "429"}), http.StatusTooManyRequests)
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil), http.StatusAccepted)
	// The bucket is now empty. A rate-limited request must not make it go negative.
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil), http.StatusTooManyRequests)
}
