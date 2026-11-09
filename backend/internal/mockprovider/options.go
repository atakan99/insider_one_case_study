// Package mockprovider is a fake SMS and email provider. It acts like
// Twilio Messaging or SendGrid Mail Send: it delivers every valid request,
// also a repeat, and it sometimes fails, is slow, loses the response or
// returns 429. See docs/specs/03-mock-provider/goal.md.
package mockprovider

import "time"

// Options configures the provider. Now and Sleep let tests run with no real
// waits.
type Options struct {
	RatePerSec float64
	Burst      int

	ErrorRate, SlowRate, LostRate float64
	SlowDelay, LatencyMax         time.Duration

	// Random seeds the failure picker.
	Random uint64

	Now   func() time.Time    // nil means time.Now
	Sleep func(time.Duration) // nil means time.Sleep
}
