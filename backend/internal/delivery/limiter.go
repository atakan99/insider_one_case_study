package delivery

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter limits the sends of one channel to perSec each second, with no
// burst, and can be paused after a 429.
//
// It has two steps. Wait paces the claims. Admit runs just before the
// request and keeps the request starts at least gap apart, so a slow
// preparation cannot bunch the sends of tokens that were granted in time.
type Limiter struct {
	rl    *rate.Limiter
	clock Clock
	gap   time.Duration

	mu          sync.Mutex
	pausedUntil time.Time
	nextSend    time.Time
}

// gapMargin makes the gap between two request starts 2% longer than
// 1s/perSec. Network jitter then cannot put perSec+1 requests into one
// second at the provider.
const gapMargin = 50

// NewLimiter returns a Limiter with a burst of 1.
func NewLimiter(perSec int, clock Clock) *Limiter {
	gap := time.Second / time.Duration(perSec)
	return &Limiter{
		rl:    rate.NewLimiter(rate.Limit(perSec), 1),
		clock: clock,
		gap:   gap + gap/gapMargin,
	}
}

// Admit blocks until gap has passed since the start of the last admitted
// request. Call it just before the request. A cancelled Admit keeps its slot.
func (l *Limiter) Admit(ctx context.Context) error {
	l.mu.Lock()
	now := l.clock.Now()
	at := l.nextSend
	if at.Before(now) {
		at = now
	}
	l.nextSend = at.Add(l.gap)
	l.mu.Unlock()
	if d := at.Sub(now); d > 0 {
		return l.clock.Sleep(ctx, d)
	}
	return ctx.Err()
}

// Pause makes Wait block for at least d from now. A shorter pause does not
// shorten a longer one.
func (l *Limiter) Pause(d time.Duration) {
	until := l.clock.Now().Add(d)
	l.mu.Lock()
	if until.After(l.pausedUntil) {
		l.pausedUntil = until
	}
	l.mu.Unlock()
}

// Wait blocks until a send is allowed. It returns ctx.Err() when ctx is cancelled.
// A pause that starts while Wait sleeps for a token is honoured too.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		if err := l.waitPause(ctx); err != nil {
			return err
		}
		now := l.clock.Now()
		r := l.rl.ReserveN(now, 1)
		if err := l.clock.Sleep(ctx, r.DelayFrom(now)); err != nil {
			r.CancelAt(now)
			return err
		}
		if !l.paused() {
			return nil
		}
	}
}

func (l *Limiter) paused() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.clock.Now().Before(l.pausedUntil)
}

func (l *Limiter) waitPause(ctx context.Context) error {
	for {
		l.mu.Lock()
		until := l.pausedUntil
		l.mu.Unlock()
		d := until.Sub(l.clock.Now())
		if d <= 0 {
			return nil
		}
		if err := l.clock.Sleep(ctx, d); err != nil {
			return err
		}
	}
}
