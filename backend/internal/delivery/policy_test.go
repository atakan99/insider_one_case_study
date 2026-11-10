package delivery

import (
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	zero := func() float64 { return 0 }
	one := func() float64 { return 1 }
	want0 := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}
	want1 := []time.Duration{1200 * time.Millisecond, 2400 * time.Millisecond, 4800 * time.Millisecond, 9600 * time.Millisecond}
	for i := 1; i <= 4; i++ {
		if got := Backoff(i, zero); got != want0[i-1] {
			t.Errorf("Backoff(%d, 0) = %v, want %v", i, got, want0[i-1])
		}
		if got := Backoff(i, one); got != want1[i-1] {
			t.Errorf("Backoff(%d, 1) = %v, want %v", i, got, want1[i-1])
		}
	}
}

func TestNext(t *testing.T) {
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	zero := func() float64 { return 0 }
	at := func(d time.Duration) *time.Time { x := now.Add(d); return &x }
	tests := []struct {
		name     string
		o        Outcome
		attempts int
		want     Result
	}{
		{"202", Outcome{Kind: Delivered, ProviderID: "m1"}, 1, Result{Status: StatusSent, ProviderID: "m1"}},
		{"429", Outcome{Kind: Retryable, RetryAfter: 3 * time.Second, Err: "429 rate_limited"}, 1,
			Result{Status: StatusRetrying, LastError: "429 rate_limited", NextAttemptAt: at(3 * time.Second)}},
		{"500 first", Outcome{Kind: Retryable, Err: "500 provider_error"}, 1,
			Result{Status: StatusRetrying, LastError: "500 provider_error", NextAttemptAt: at(time.Second)}},
		{"503 fourth", Outcome{Kind: Retryable, Err: "503 unavailable"}, 4,
			Result{Status: StatusRetrying, LastError: "503 unavailable", NextAttemptAt: at(8 * time.Second)}},
		{"503 last", Outcome{Kind: Retryable, Err: "503 unavailable"}, 5,
			Result{Status: StatusFailed, LastError: "503 unavailable", Retryable: true}},
		{"429 last", Outcome{Kind: Retryable, RetryAfter: time.Second, Err: "429 rate_limited"}, 5,
			Result{Status: StatusFailed, LastError: "429 rate_limited", Retryable: true}},
		{"connection refused", Outcome{Kind: Retryable, Err: "connection refused"}, 2,
			Result{Status: StatusRetrying, LastError: "connection refused", NextAttemptAt: at(2 * time.Second)}},
		{"400", Outcome{Kind: Permanent, Err: "400 invalid_request"}, 1,
			Result{Status: StatusFailed, LastError: "400 invalid_request"}},
		{"timeout", Outcome{Kind: Ambiguous, Err: "timeout"}, 1,
			Result{Status: StatusUnknown, LastError: "timeout"}},
		{"dropped", Outcome{Kind: Ambiguous, Err: "connection lost"}, 3,
			Result{Status: StatusUnknown, LastError: "connection lost"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Next(tc.o, tc.attempts, 5, now, zero)
			if got.Status != tc.want.Status || got.LastError != tc.want.LastError ||
				got.Retryable != tc.want.Retryable || got.ProviderID != tc.want.ProviderID {
				t.Fatalf("Next = %+v, want %+v", got, tc.want)
			}
			if (got.NextAttemptAt == nil) != (tc.want.NextAttemptAt == nil) ||
				got.NextAttemptAt != nil && !got.NextAttemptAt.Equal(*tc.want.NextAttemptAt) {
				t.Fatalf("NextAttemptAt = %v, want %v", got.NextAttemptAt, tc.want.NextAttemptAt)
			}
		})
	}
}
