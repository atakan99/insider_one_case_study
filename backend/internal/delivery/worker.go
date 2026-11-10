package delivery

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"campaigncomposer/internal/config"
	"campaigncomposer/internal/segment"
)

// Default intervals and limits of the worker.
const (
	defaultPlanEvery    = time.Second
	defaultRetryEvery   = 500 * time.Millisecond
	defaultRecoverEvery = 5 * time.Second
	defaultClaimIdle    = 60 * time.Second
	defaultStaleAfter   = 120 * time.Second
	// lateRetry is how late a retry must be before the recovery pushes it.
	lateRetry = 30 * time.Second
	// shutdownGrace is how long Run waits for the requests that started.
	shutdownGrace = 10 * time.Second
	readBlock     = time.Second
	errorPause    = time.Second
	claimBatch    = 100
)

// Worker plans approved campaigns, sends the queued messages and repairs the
// queue after a crash. Run is the entry point.
type Worker struct {
	Store    *Store
	Queue    *Queue
	Provider *Provider
	Allow    segment.Allowlist
	Clock    Clock
	Cfg      config.WorkerConfig
	Log      *slog.Logger

	// Intervals. Zero means the default. Tests set short ones.
	PlanEvery, RetryEvery, RecoverEvery, ClaimIdle time.Duration
	// StaleAfter is the age after which a queued job is pushed again.
	StaleAfter time.Duration

	// pushedRetries remembers the retrying jobs that are on the stream, so
	// that a job that waits behind other entries is not pushed at each poll.
	retryMu       sync.Mutex
	pushedRetries map[int64]time.Time
}

func (w *Worker) log() *slog.Logger {
	if w.Log != nil {
		return w.Log
	}
	return slog.Default()
}

func orDefault(d, def time.Duration) time.Duration {
	if d > 0 {
		return d
	}
	return def
}

// Run blocks until ctx ends. It then stops reading and waits, and gives the
// requests that started up to 10 seconds to end.
func (w *Worker) Run(ctx context.Context) error {
	if w.Clock == nil {
		w.Clock = RealClock()
	}
	if err := w.Queue.Ensure(ctx); err != nil {
		return err
	}
	var wg sync.WaitGroup
	loop := func(every time.Duration, fn func(context.Context)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.tick(ctx, every, fn)
		}()
	}
	loop(orDefault(w.PlanEvery, defaultPlanEvery), w.plan)
	loop(orDefault(w.RetryEvery, defaultRetryEvery), w.pollRetries)
	loop(orDefault(w.RecoverEvery, defaultRecoverEvery), w.recover)

	for _, ch := range Channels {
		c := &Consumer{
			Store: w.Store, Queue: w.Queue, Provider: w.Provider,
			Limiter: NewLimiter(w.Cfg.RatePerSec, w.Clock),
			Channel: ch, Clock: w.Clock, MaxAttempts: w.Cfg.MaxAttempts,
			Jitter: rand.Float64, Log: w.log(),
		}
		for range w.Cfg.Consumers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				w.consume(ctx, c)
			}()
		}
		loop(orDefault(w.RecoverEvery, defaultRecoverEvery), func(ctx context.Context) { w.claimDead(ctx, c) })
	}

	<-ctx.Done()
	w.log().Info("worker stopping")
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(shutdownGrace):
		w.log().Warn("worker stopped with requests still running; the recovery sets them to unknown")
	}
	return nil
}

// tick runs fn now and then after each interval, until ctx ends. A run
// never overlaps the next one.
func (w *Worker) tick(ctx context.Context, every time.Duration, fn func(context.Context)) {
	for ctx.Err() == nil {
		fn(ctx)
		if w.Clock.Sleep(ctx, every) != nil {
			return
		}
	}
}

// push puts the ids on the streams and marks them as enqueued. A failure is
// logged only: the jobs stay queued, and the recovery pushes them again.
func (w *Worker) push(ctx context.Context, ids map[Channel][]int64) {
	for _, ch := range Channels {
		if len(ids[ch]) == 0 {
			continue
		}
		if err := w.Queue.Push(ctx, ch, ids[ch]); err != nil {
			w.log().Error("push to queue", "channel", ch, "count", len(ids[ch]), "error", err)
			continue
		}
		if err := w.Store.MarkEnqueued(ctx, ids[ch], w.Clock.Now()); err != nil {
			w.log().Error("mark enqueued", "channel", ch, "count", len(ids[ch]), "error", err)
		}
	}
}

// withoutBacklog drops the channels whose group still has unread entries.
// The jobs of such a channel are waiting on the stream, and a second entry for
// each would only use rate tokens. When the lag is unknown or cannot be read,
// the jobs are pushed again, which is the safe direction.
func (w *Worker) withoutBacklog(ctx context.Context, ids map[Channel][]int64) map[Channel][]int64 {
	out := map[Channel][]int64{}
	for ch, list := range ids {
		if len(list) == 0 {
			continue
		}
		lag, err := w.Queue.Lag(ctx, ch)
		if err != nil {
			w.logErr(ctx, "stream lag", err)
		} else if lag > 0 {
			continue
		}
		out[ch] = list
	}
	return out
}

func (w *Worker) plan(ctx context.Context) {
	for ctx.Err() == nil {
		cid, ids, ok, err := w.Store.PlanNext(ctx, w.Allow, w.Clock.Now())
		if err != nil {
			if ctx.Err() == nil {
				w.log().Error("plan campaign", "error", err)
			}
			return
		}
		if !ok {
			return
		}
		w.log().Info("campaign planned", "campaign_id", cid,
			"sms", len(ids[ChannelSMS]), "email", len(ids[ChannelEmail]))
		w.push(ctx, ids)
	}
}

// pollRetries pushes the retrying jobs that are due. A job that stays due
// is not pushed again for lateRetry: it waits on the stream.
func (w *Worker) pollRetries(ctx context.Context) {
	now := w.Clock.Now()
	due, err := w.Store.DueRetries(ctx, now, 0)
	if err != nil {
		w.logErr(ctx, "due retries", err)
		return
	}
	w.push(ctx, w.freshRetries(due, now, true))
}

// freshRetries returns the jobs of due that are not on the stream yet. When
// forget is true, it first forgets the jobs that are not due any more: a
// later retry of such a job is a new one.
func (w *Worker) freshRetries(due map[Channel][]int64, now time.Time, forget bool) map[Channel][]int64 {
	w.retryMu.Lock()
	defer w.retryMu.Unlock()
	if w.pushedRetries == nil {
		w.pushedRetries = map[int64]time.Time{}
	}
	if forget {
		still := map[int64]bool{}
		for _, ids := range due {
			for _, id := range ids {
				still[id] = true
			}
		}
		for id := range w.pushedRetries {
			if !still[id] {
				delete(w.pushedRetries, id)
			}
		}
	}
	out := map[Channel][]int64{}
	for ch, ids := range due {
		for _, id := range ids {
			if at, ok := w.pushedRetries[id]; ok && now.Sub(at) < lateRetry {
				continue
			}
			w.pushedRetries[id] = now
			out[ch] = append(out[ch], id)
		}
	}
	return out
}

func (w *Worker) logErr(ctx context.Context, msg string, err error) {
	if ctx.Err() == nil {
		w.log().Error(msg, "error", err)
	}
}

func (w *Worker) recover(ctx context.Context) {
	now := w.Clock.Now()
	if ids, err := w.Store.StaleQueued(ctx, now, orDefault(w.StaleAfter, defaultStaleAfter)); err != nil {
		w.logErr(ctx, "stale queued", err)
	} else {
		w.push(ctx, w.withoutBacklog(ctx, ids))
	}
	if ids, err := w.Store.DueRetries(ctx, now, lateRetry); err != nil {
		w.logErr(ctx, "late retries", err)
	} else {
		w.push(ctx, w.freshRetries(ids, now, false))
	}
	if n, err := w.Store.ExpireSending(ctx, now, w.Cfg.Timeout+30*time.Second); err != nil {
		w.logErr(ctx, "expire sending", err)
	} else if n > 0 {
		w.log().Warn("send jobs set to unknown", "count", n)
	}
	done, err := w.Store.CompleteCampaigns(ctx)
	if err != nil {
		w.logErr(ctx, "complete campaigns", err)
		return
	}
	for _, id := range done {
		p, err := w.Store.Progress(ctx, id)
		if err != nil {
			w.logErr(ctx, "campaign progress", err)
			continue
		}
		w.log().Info("campaign completed", "campaign_id", id, "sms", p.Channels[ChannelSMS], "email", p.Channels[ChannelEmail])
	}
}

// claimDead handles the entries of consumers that stopped without an ack.
func (w *Worker) claimDead(ctx context.Context, c *Consumer) {
	for ctx.Err() == nil {
		es, err := w.Queue.AutoClaim(ctx, c.Channel, orDefault(w.ClaimIdle, defaultClaimIdle), claimBatch)
		for _, e := range es {
			w.handle(ctx, c, e)
		}
		if err != nil {
			w.logErr(ctx, "autoclaim", err)
			return
		}
		if len(es) == 0 {
			return
		}
	}
}

func (w *Worker) handle(ctx context.Context, c *Consumer, e Entry) {
	if err := c.Handle(ctx, e); err != nil && ctx.Err() == nil {
		w.log().Error("handle entry", "channel", c.Channel, "job_id", e.JobID, "error", err)
	}
}

// consume reads one entry at a time and handles it, until ctx ends.
func (w *Worker) consume(ctx context.Context, c *Consumer) {
	for ctx.Err() == nil {
		es, err := w.Queue.Read(ctx, c.Channel, 1, readBlock)
		for _, e := range es {
			w.handle(ctx, c, e)
		}
		if err == nil || ctx.Err() != nil {
			continue
		}
		if IsNoGroup(err) {
			w.log().Warn("stream or group lost, creating it again", "channel", c.Channel)
			if eerr := w.Queue.Ensure(ctx); eerr != nil {
				w.logErr(ctx, "ensure queue", eerr)
				_ = w.Clock.Sleep(ctx, errorPause)
			}
			continue
		}
		if !errors.Is(err, context.Canceled) {
			w.log().Error("read queue", "channel", c.Channel, "error", err)
		}
		_ = w.Clock.Sleep(ctx, errorPause)
	}
}
