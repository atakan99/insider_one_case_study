package delivery

import (
	"context"
	"fmt"
	"log/slog"
	"time"
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
	// MaxPrep is the longest time between a rate token and the request.
	// The claim and the load run in this time. 0 means 100 ms.
	MaxPrep time.Duration
	Metrics *Metrics // nil means no metrics

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

// defaultMaxPrep is the default of Consumer.MaxPrep.
const defaultMaxPrep = 100 * time.Millisecond

// Handle runs one entry: wait, claim, load, send, write the result, acknowledge.
// It returns an error only when the entry must stay unacknowledged.
func (c *Consumer) Handle(ctx context.Context, e Entry) error {
	// The writes after a claim use a context that cannot be cancelled: the
	// result of an attempt must reach Postgres, also during a shutdown.
	done := context.WithoutCancel(ctx)

	// An entry for a job that cannot be claimed must not use a rate token.
	// On an error here, the normal path decides.
	if ok, err := c.Store.Claimable(ctx, e.JobID, c.Clock.Now()); err == nil && !ok {
		return c.Queue.Ack(done, c.Channel, e.StreamID)
	}

	maxPrep := c.MaxPrep
	if maxPrep <= 0 {
		maxPrep = defaultMaxPrep
	}
	var job Job
	var msg Message
	for {
		// The job is not claimed during the wait. It stays queued, so the
		// recovery cannot take it for a request that hangs.
		waitFrom := c.Clock.Now()
		err := c.Limiter.Wait(ctx)
		granted := c.Clock.Now()
		c.Metrics.Waited(c.Channel, granted.Sub(waitFrom))
		if err != nil {
			return err
		}

		var ok bool
		job, ok, err = c.Store.Claim(ctx, e.JobID, c.Clock.Now())
		if err != nil {
			return err
		}
		if !ok {
			return c.Queue.Ack(done, c.Channel, e.StreamID)
		}
		if c.afterClaim != nil {
			c.afterClaim()
		}
		// Nothing was sent yet, so the job can go back to the queue.
		release := func() {
			if rerr := c.Store.Release(done, job.ID); rerr != nil {
				c.log().Error("release job", "job_id", job.ID, "channel", c.Channel, "error", rerr)
			}
		}
		if err := ctx.Err(); err != nil {
			release()
			return err
		}

		// The load reads the consent. It is the last step before the
		// request, so a change of consent during the wait stops the send.
		var skip string
		msg, skip, err = c.Store.Load(ctx, job)
		if err != nil {
			release()
			return err
		}
		if skip != "" {
			c.log().Info("send job skipped", "job_id", job.ID, "channel", c.Channel, "attempt", job.Attempts, "error", skip)
			return c.finish(done, e, job, Result{Status: StatusSkipped, LastError: skip})
		}

		// A token is good for MaxPrep. When the database was slower, the
		// sends that it held would leave together and break the rate. Such
		// a job goes back to queued and takes a new token.
		prep := c.Clock.Now().Sub(granted)
		if prep <= maxPrep {
			break
		}
		c.log().Warn("send job prepared too slowly, new rate token", "job_id", job.ID, "channel", c.Channel, "took", prep)
		release()
	}

	// The token paces the claims, but a preparation time of up to MaxPrep
	// can still move sends together. Admit spaces the request starts.
	if err := c.Limiter.Admit(ctx); err != nil {
		if rerr := c.Store.Release(done, job.ID); rerr != nil {
			c.log().Error("release job", "job_id", job.ID, "channel", c.Channel, "error", rerr)
		}
		return err
	}

	reference := fmt.Sprintf("%d:%d:%s", job.CampaignID, job.UserID, job.Channel)
	// A started request may finish: only HTTP.Timeout bounds it, so a
	// shutdown does not turn a delivered message into "canceled".
	sendFrom := c.Clock.Now()
	o := c.Provider.Send(done, c.Channel, msg, reference)
	c.Metrics.Attempt(c.Channel, o.Kind, c.Clock.Now().Sub(sendFrom))
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
	c.Metrics.Result(c.Channel, r.Status)
	if r.Status == StatusRetrying {
		c.Metrics.RetryScheduled(c.Channel)
	}
	return c.Queue.Ack(ctx, c.Channel, e.StreamID)
}
