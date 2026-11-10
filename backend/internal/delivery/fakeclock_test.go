package delivery

import (
	"context"
	"sync"
	"time"
)

// fakeClock is a Clock whose Sleep moves time forward and does not wait.
type fakeClock struct {
	mu      sync.Mutex
	now     time.Time
	onSleep func() // called at the start of each Sleep, before the cancel check
}

func newFakeClock(t time.Time) *fakeClock { return &fakeClock{now: t} }

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Set(t time.Time) {
	c.mu.Lock()
	c.now = t
	c.mu.Unlock()
}

func (c *fakeClock) Sleep(ctx context.Context, d time.Duration) error {
	if c.onSleep != nil {
		c.onSleep()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if d > 0 {
		c.mu.Lock()
		c.now = c.now.Add(d)
		c.mu.Unlock()
	}
	return nil
}
