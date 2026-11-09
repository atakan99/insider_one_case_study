package mockprovider

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// sleepRecorder is a fake Sleep that records each duration and returns at once.
type sleepRecorder struct {
	mu    sync.Mutex
	calls []time.Duration
}

func (s *sleepRecorder) Sleep(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, d)
}

func (s *sleepRecorder) Calls() []time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]time.Duration(nil), s.calls...)
}

func TestForced(t *testing.T) {
	const slowDelay = 5 * time.Second
	cases := []struct {
		header        string
		status        int
		code          string
		wantDelivered int64
		check         func(t *testing.T, st ChannelStats, sleeps []time.Duration)
	}{
		{"500", 500, "provider_error", 0, func(t *testing.T, st ChannelStats, _ []time.Duration) {
			if st.Errors != 1 {
				t.Errorf("errors = %d, want 1", st.Errors)
			}
		}},
		{"503", 503, "unavailable", 0, func(t *testing.T, st ChannelStats, _ []time.Duration) {
			if st.Errors != 1 {
				t.Errorf("errors = %d, want 1", st.Errors)
			}
		}},
		{"slow", 202, "", 1, func(t *testing.T, st ChannelStats, sleeps []time.Duration) {
			if st.Slow != 1 {
				t.Errorf("slow = %d, want 1", st.Slow)
			}
			if len(sleeps) != 1 || sleeps[0] != slowDelay {
				t.Errorf("sleeps = %v, want [%v]", sleeps, slowDelay)
			}
		}},
		{"429", 429, "rate_limited", 0, func(t *testing.T, st ChannelStats, _ []time.Duration) {
			if st.RateLimited != 1 {
				t.Errorf("rate_limited = %d, want 1", st.RateLimited)
			}
		}},
		{"none", 202, "", 1, nil},
		{"boom", 400, "invalid_request", 0, func(t *testing.T, st ChannelStats, _ []time.Duration) {
			if st.Rejected != 1 {
				t.Errorf("rejected = %d, want 1", st.Rejected)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.header, func(t *testing.T) {
			rec := &sleepRecorder{}
			o := normalOptions(newFakeClock())
			o.ErrorRate = 1 // "none" must turn this off
			o.SlowDelay = slowDelay
			o.Sleep = rec.Sleep
			srv := startServer(t, o)
			resp := post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), map[string]string{"X-Mock-Fail": tc.header})
			wantStatus(t, resp, tc.status)
			if tc.code != "" {
				if e := decode[errorBody](t, resp); e.Error.Code != tc.code {
					t.Fatalf("code = %q, want %q", e.Error.Code, tc.code)
				}
			}
			if tc.header == "429" && resp.Header.Get("Retry-After") != "1" {
				t.Fatalf("Retry-After = %q, want 1", resp.Header.Get("Retry-After"))
			}
			st := stats(t, srv.URL)["sms"]
			if st.Delivered != tc.wantDelivered {
				t.Fatalf("delivered = %d, want %d", st.Delivered, tc.wantDelivered)
			}
			if tc.check != nil {
				tc.check(t, st, rec.Calls())
			}
		})
	}
}

func TestForced_Lost(t *testing.T) {
	srv := startServer(t, normalOptions(newFakeClock()))
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/sms", strings.NewReader(sms("+905321234567", "hi", "c1:u1:sms")))
	req.Header.Set("X-Mock-Fail", "lost")
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		resp.Body.Close()
		t.Fatalf("got status %d, want a transport error", resp.StatusCode)
	}
	st := stats(t, srv.URL)["sms"]
	if st.Delivered != 1 || st.Lost != 1 {
		t.Fatalf("sms = %+v, want delivered 1 and lost 1", st)
	}
}

func TestForced_ValidationFirst(t *testing.T) {
	srv := startServer(t, normalOptions(newFakeClock()))
	resp := post(t, srv.URL, "/v1/sms", sms("+905321234567", "", ""), map[string]string{"X-Mock-Fail": "500"})
	wantStatus(t, resp, http.StatusBadRequest)
}

func TestSlow_StoresAfterClientLeaves(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	clientDone := make(chan struct{})
	o := normalOptions(newFakeClock())
	o.SlowDelay = 5 * time.Second
	o.Sleep = func(d time.Duration) {
		if d == o.SlowDelay {
			cancel()     // the client times out during the slow wait
			<-clientDone // and has left before the provider stores the message
		}
	}
	srv := startServer(t, o)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL+"/v1/sms", strings.NewReader(sms("+905321234567", "hi", "")))
	req.Header.Set("X-Mock-Fail", "slow")
	resp, err := http.DefaultClient.Do(req)
	close(clientDone)
	if err == nil {
		resp.Body.Close()
		t.Fatalf("got status %d, want a context error", resp.StatusCode)
	}

	deadline := time.Now().Add(time.Second)
	for listMessages(t, srv.URL, "").Total != 1 {
		if time.Now().After(deadline) {
			t.Fatal("message not stored after the client left")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestRandom_ZeroRates(t *testing.T) {
	srv := startServer(t, normalOptions(newFakeClock()))
	for i := range 1000 {
		if resp := post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil); resp.StatusCode != http.StatusAccepted {
			t.Fatalf("request %d: status %d", i, resp.StatusCode)
		}
	}
}

func TestRandom_Distribution(t *testing.T) {
	o := Options{ErrorRate: 0.03, SlowRate: 0.03, LostRate: 0.01, Random: 1}
	p := newPicker(o)
	const n = 10000
	counts := map[mode]int{}
	for range n {
		counts[p.pick()]++
	}
	want := map[mode]float64{
		modeNone:     1 - 0.07,
		modeError500: 0.015,
		modeError503: 0.015,
		modeSlow:     0.03,
		modeLost:     0.01,
	}
	for m, rate := range want {
		got := float64(counts[m]) / n
		if got < rate-0.01 || got > rate+0.01 {
			t.Errorf("mode %v: rate %.4f, want %.4f ± 0.01", m, got, rate)
		}
	}
	if counts[modeForced429] != 0 {
		t.Errorf("random picker returned forced429")
	}
}

func TestRandom_SameSeedSameModes(t *testing.T) {
	o := Options{ErrorRate: 0.2, SlowRate: 0.2, LostRate: 0.2, Random: 9}
	a, b := newPicker(o), newPicker(o)
	for i := range 1000 {
		if x, y := a.pick(), b.pick(); x != y {
			t.Fatalf("pick %d: %v != %v", i, x, y)
		}
	}
}

func TestLatency(t *testing.T) {
	rec := &sleepRecorder{}
	o := normalOptions(newFakeClock())
	o.LatencyMax = 50 * time.Millisecond
	o.Sleep = rec.Sleep
	srv := startServer(t, o)
	for range 20 {
		wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil), http.StatusAccepted)
	}
	calls := rec.Calls()
	if len(calls) != 20 {
		t.Fatalf("sleep calls = %d, want 20", len(calls))
	}
	for _, d := range calls {
		if d < 0 || d > 50*time.Millisecond {
			t.Fatalf("latency %v, want 0 to 50ms", d)
		}
	}

	rec0 := &sleepRecorder{}
	o0 := normalOptions(newFakeClock())
	o0.Sleep = rec0.Sleep
	srv0 := startServer(t, o0)
	wantStatus(t, post(t, srv0.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil), http.StatusAccepted)
	if n := len(rec0.Calls()); n != 0 {
		t.Fatalf("sleep calls with LatencyMax 0 = %d, want 0", n)
	}
}
