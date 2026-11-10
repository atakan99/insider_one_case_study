package dashboard

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"campaigncomposer/internal/delivery"
)

const defaultHealthTimeout = time.Second

// StreamDepth is the state of one send stream. Length and Pending are nil when
// Redis cannot answer.
type StreamDepth struct {
	Channel    delivery.Channel `json:"channel"`
	Length     *int64           `json:"length"`
	Pending    *int64           `json:"pending"`
	InProgress int64            `json:"in_progress"`
}

// Health is the answer of the health check. Services maps db, redis, worker
// and provider to "ok" or "down". Streams lists sms first, then email.
type Health struct {
	Services map[string]string `json:"services"`
	Streams  []StreamDepth     `json:"streams"`
}

// Pinger checks that a dependency answers.
type Pinger interface {
	Ping(ctx context.Context) error
}

// RedisPinger adapts a Redis client to Pinger.
type RedisPinger struct{ Client *redis.Client }

// Ping sends PING to Redis.
func (p RedisPinger) Ping(ctx context.Context) error { return p.Client.Ping(ctx).Err() }

// Checker checks the services and reads the depth of the send streams. Each
// check has its own timeout, and all checks run at the same time.
type Checker struct {
	DB          Pinger
	Redis       Pinger
	Queue       *delivery.Queue
	Store       *Store
	HTTP        *http.Client
	WorkerURL   string
	ProviderURL string
	Timeout     time.Duration // 0 means 1 s
}

// Check runs all checks and returns the result. It does not return an error:
// a failed check shows as "down" or as a nil depth.
func (c *Checker) Check(ctx context.Context) Health {
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = defaultHealthTimeout
	}
	h := Health{
		Services: map[string]string{},
		Streams:  make([]StreamDepth, len(delivery.Channels)),
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	run := func(f func(ctx context.Context)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			f(cctx)
		}()
	}
	service := func(name string, check func(ctx context.Context) error) {
		run(func(ctx context.Context) {
			state := "ok"
			if err := check(ctx); err != nil {
				state = "down"
				slog.WarnContext(ctx, "health check failed", "service", name, "err", err)
			}
			mu.Lock()
			h.Services[name] = state
			mu.Unlock()
		})
	}
	service("db", c.DB.Ping)
	service("redis", c.Redis.Ping)
	service("worker", func(ctx context.Context) error { return c.get(ctx, c.WorkerURL) })
	service("provider", func(ctx context.Context) error { return c.get(ctx, c.ProviderURL) })

	for i, ch := range delivery.Channels {
		h.Streams[i].Channel = ch
		if c.Queue == nil {
			continue
		}
		run(func(ctx context.Context) {
			length, pending, err := c.Queue.Depth(ctx, ch)
			if err != nil {
				slog.WarnContext(ctx, "health check failed", "service", "queue depth", "channel", string(ch), "err", err)
				return
			}
			h.Streams[i].Length, h.Streams[i].Pending = &length, &pending
		})
	}
	if c.Store != nil {
		run(c.inProgress(&h))
	}
	wg.Wait()
	return h
}

func (c *Checker) inProgress(h *Health) func(ctx context.Context) {
	return func(ctx context.Context) {
		in, err := c.Store.InProgress(ctx)
		if err != nil {
			slog.WarnContext(ctx, "health check failed", "service", "in progress", "err", err)
			return
		}
		for i, ch := range delivery.Channels {
			h.Streams[i].InProgress = in[ch]
		}
	}
}

// get is ok only for status 200 from GET base/healthz.
func (c *Checker) get(ctx context.Context, base string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/healthz", nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	cl := c.HTTP
	if cl == nil {
		cl = http.DefaultClient
	}
	resp, err := cl.Do(req)
	if err != nil {
		// url.Error holds the URL. Log only the cause.
		var ue *url.Error
		if errors.As(err, &ue) {
			err = ue.Err
		}
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
