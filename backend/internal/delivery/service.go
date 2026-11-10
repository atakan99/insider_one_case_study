package delivery

import (
	"context"
	"log/slog"
	"time"
)

// Service gives the api the delivery state and the retry action.
type Service struct {
	Store *Store
	Queue *Queue
	Now   func() time.Time
	Log   *slog.Logger
}

// Progress returns the job counts. ErrNotFound for an unknown campaign,
// ErrNotSendable for a draft.
func (s *Service) Progress(ctx context.Context, id int64) (Progress, error) {
	return s.Store.Progress(ctx, id)
}

// ListJobs checks the campaign first, because the store returns an empty list
// for an unknown campaign or a draft.
func (s *Service) ListJobs(ctx context.Context, id int64, f JobFilter) ([]Job, *int64, error) {
	if _, err := s.Store.Progress(ctx, id); err != nil {
		return nil, nil, err
	}
	return s.Store.ListJobs(ctx, id, f)
}

// Retry resets the retryable failed jobs, then pushes them. A push error is
// logged, not returned: the worker recovery pushes the jobs later.
func (s *Service) Retry(ctx context.Context, id int64) (retried int, p Progress, err error) {
	ids, err := s.Store.Retry(ctx, id, s.Now())
	if err != nil {
		return 0, Progress{}, err
	}
	for ch, list := range ids {
		retried += len(list)
		if perr := s.Queue.Push(ctx, ch, list); perr != nil {
			s.Log.ErrorContext(ctx, "retry push failed, recovery will push", "campaign_id", id, "channel", ch, "err", perr)
			continue
		}
		if merr := s.Store.MarkEnqueued(ctx, list, s.Now()); merr != nil {
			s.Log.ErrorContext(ctx, "retry mark enqueued failed", "campaign_id", id, "channel", ch, "err", merr)
		}
	}
	p, err = s.Store.Progress(ctx, id)
	return retried, p, err
}
