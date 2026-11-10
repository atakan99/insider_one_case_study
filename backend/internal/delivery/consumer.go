package delivery

import (
	"context"
	"fmt"
	"log/slog"
)

// Consumer handles the queue entries of one channel.
type Consumer struct {
	Store       *Store
	Queue       *Queue
	Provider    *Provider
	Limiter     *Limiter
	Channel     Channel
	Clock       Clock
	MaxAttempts int
	Jitter      func() float64
	Log         *slog.Logger

	// Hooks for tests. Both are nil in production.
	afterClaim func()
	finishErr  func() error
}

func (c *Consumer) log() *slog.Logger {
	if c.Log != nil {
		return c.Log
	}
	return slog.Default()
}

// Handle runs one entry: wait, claim, load, send, write the result, acknowledge.
// It returns an error only when the entry must stay unacknowledged.
func (c *Consumer) Handle(ctx context.Context, e Entry) error {
	// The writes after a claim use a context that cannot be cancelled: the
	// result of an attempt must reach Postgres, also during a shutdown.
	done := context.WithoutCancel(ctx)

	if err := c.Limiter.Wait(ctx); err != nil {
		return err
	}
	job, ok, err := c.Store.Claim(ctx, e.JobID, c.Clock.Now())
	if err != nil {
		return err
	}
	if !ok {
		return c.Queue.Ack(done, c.Channel, e.StreamID)
	}
	if c.afterClaim != nil {
		c.afterClaim()
	}
	if err := ctx.Err(); err != nil {
		if rerr := c.Store.Release(done, job.ID); rerr != nil {
			c.log().Error("release job", "job_id", job.ID, "channel", c.Channel, "error", rerr)
		}
		return err
	}

	msg, skip, err := c.Store.Load(ctx, job)
	if err != nil {
		// Nothing was sent, so the job can go back to the queue.
		if rerr := c.Store.Release(done, job.ID); rerr != nil {
			c.log().Error("release job", "job_id", job.ID, "channel", c.Channel, "error", rerr)
		}
		return err
	}
	if skip != "" {
		c.log().Info("send job skipped", "job_id", job.ID, "channel", c.Channel, "attempt", job.Attempts, "error", skip)
		return c.finish(done, e, job, Result{Status: StatusSkipped, LastError: skip})
	}

	reference := fmt.Sprintf("%d:%d:%s", job.CampaignID, job.UserID, job.Channel)
	// A started request may finish: only HTTP.Timeout bounds it, so a
	// shutdown does not turn a delivered message into "canceled".
	o := c.Provider.Send(done, c.Channel, msg, reference)
	if o.Kind == Retryable && o.RetryAfter > 0 {
		c.Limiter.Pause(o.RetryAfter)
	}
	if o.Kind != Delivered {
		c.log().Warn("send not delivered", "job_id", job.ID, "channel", c.Channel, "attempt", job.Attempts, "error", o.Err)
	}
	return c.finish(done, e, job, Next(o, job.Attempts, c.MaxAttempts, c.Clock.Now(), c.Jitter))
}

// finish writes the result and acknowledges the entry. When the write fails,
// the entry stays unacknowledged. Its job stays in sending, and a new claim
// fails, so the message is not sent again.
func (c *Consumer) finish(ctx context.Context, e Entry, job Job, r Result) error {
	var err error
	if c.finishErr != nil {
		err = c.finishErr()
	}
	if err == nil {
		err = c.Store.Finish(ctx, job.ID, r, c.Clock.Now())
	}
	if err != nil {
		c.log().Error("write send result", "job_id", job.ID, "channel", c.Channel, "attempt", job.Attempts, "error", err)
		return err
	}
	return c.Queue.Ack(ctx, c.Channel, e.StreamID)
}
