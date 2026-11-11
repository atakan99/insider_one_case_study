-- +goose Up
-- The send job status moves only along the paths of the worker. Each UPDATE in
-- delivery.Store also has its own WHERE condition. This trigger is the second
-- guard: a new code path cannot write a status that the worker never writes.
--
-- Note on campaigns_frozen_after_approval: it allows queued -> completed, but no
-- code path makes it, because Claim moves the campaign to sending with the
-- first claim. It stays as a fallback for CompleteCampaigns.
-- +goose StatementBegin
CREATE FUNCTION send_jobs_status_moves() RETURNS trigger AS $$
BEGIN
    IF (OLD.status, NEW.status) IN (
        ('queued', 'sending'),
        ('retrying', 'sending'),
        ('sending', 'queued'),
        ('sending', 'sent'),
        ('sending', 'retrying'),
        ('sending', 'failed'),
        ('sending', 'unknown'),
        ('sending', 'skipped')
    ) OR (OLD.status = 'failed' AND OLD.retryable AND NEW.status = 'queued') THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION 'send job status cannot move from % to %', OLD.status, NEW.status;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER send_jobs_status_moves
BEFORE UPDATE OF status ON send_jobs
FOR EACH ROW WHEN (OLD.status IS DISTINCT FROM NEW.status)
EXECUTE FUNCTION send_jobs_status_moves();

-- +goose Down
DROP TRIGGER send_jobs_status_moves ON send_jobs;
DROP FUNCTION send_jobs_status_moves();
