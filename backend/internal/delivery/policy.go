package delivery

import (
	"math"
	"time"
)

// Backoff is the wait before the next attempt. attempts is the number of
// attempts so far: 1 to 4 give 1, 2, 4 and 8 s, plus up to 20% jitter. More
// than 4 stays at 8 s. jitter returns a value from 0 to 1.
func Backoff(attempts int, jitter func() float64) time.Duration {
	n := min(max(attempts, 1), 4)
	base := float64(time.Second) * float64(int(1)<<(n-1))
	return time.Duration(math.Round(base * (1 + 0.2*jitter())))
}

// Next maps an outcome to the new state of the send job. attempts is the
// count after this attempt. A retrying result always has NextAttemptAt.
func Next(o Outcome, attempts, maxAttempts int, now time.Time, jitter func() float64) Result {
	switch o.Kind {
	case Delivered:
		return Result{Status: StatusSent, ProviderID: o.ProviderID, LastError: o.Err}
	case Retryable:
		if attempts >= maxAttempts {
			return Result{Status: StatusFailed, LastError: o.Err, Retryable: true}
		}
		wait := o.RetryAfter
		if wait <= 0 {
			wait = Backoff(attempts, jitter)
		}
		at := now.Add(wait)
		return Result{Status: StatusRetrying, LastError: o.Err, NextAttemptAt: &at}
	case Ambiguous:
		return Result{Status: StatusUnknown, LastError: o.Err}
	default:
		return Result{Status: StatusFailed, LastError: o.Err}
	}
}
