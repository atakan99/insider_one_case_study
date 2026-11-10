-- +goose Up
CREATE INDEX send_jobs_sent_at ON send_jobs (sent_at) WHERE sent_at IS NOT NULL;

-- +goose Down
DROP INDEX send_jobs_sent_at;
