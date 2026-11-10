// Package delivery sends campaign messages through a queue and a provider.
package delivery

import (
	"errors"
	"time"
)

// Channel is a delivery channel.
type Channel string

// Channels.
const (
	ChannelSMS   Channel = "sms"
	ChannelEmail Channel = "email"
)

// Channels lists every channel.
var Channels = []Channel{ChannelSMS, ChannelEmail}

// Status is the state of a send job.
type Status string

// Send job statuses. Terminal: sent, failed, unknown, skipped.
const (
	StatusQueued   Status = "queued"
	StatusSending  Status = "sending"
	StatusRetrying Status = "retrying"
	StatusSent     Status = "sent"
	StatusFailed   Status = "failed"
	StatusUnknown  Status = "unknown"
	StatusSkipped  Status = "skipped"
)

// Skip reasons that Load returns.
const (
	SkipNoConsent = "no_consent"
	SkipNoContact = "no_contact"
	SkipOverLimit = "over_limit"
)

// Errors of the store.
var (
	ErrNotFound       = errors.New("delivery: campaign not found")
	ErrNotSendable    = errors.New("delivery: campaign has not been approved")
	ErrNotCompleted   = errors.New("delivery: campaign is not completed")
	ErrNothingToRetry = errors.New("delivery: no failed job can be retried")
)

// Job is one row of send_jobs.
type Job struct {
	ID         int64     `json:"id"`
	CampaignID int64     `json:"campaign_id"`
	UserID     int64     `json:"user_id"`
	Channel    Channel   `json:"channel"`
	Language   string    `json:"language"`
	Status     Status    `json:"status"`
	Attempts   int       `json:"attempts"`
	LastError  string    `json:"last_error"`
	Retryable  bool      `json:"retryable"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Result is the new state after an attempt. NextAttemptAt is set for
// StatusRetrying only.
type Result struct {
	Status        Status
	LastError     string
	Retryable     bool
	ProviderID    string
	NextAttemptAt *time.Time
}

// Message is a rendered message for one recipient.
type Message struct{ To, Subject, Body string }

// ChannelProgress counts the jobs of one channel by status.
type ChannelProgress struct {
	Total           int64 `json:"total"`
	Queued          int64 `json:"queued"`
	Sending         int64 `json:"sending"`
	Retrying        int64 `json:"retrying"`
	Sent            int64 `json:"sent"`
	Failed          int64 `json:"failed"`
	Unknown         int64 `json:"unknown"`
	Skipped         int64 `json:"skipped"`
	RetryableFailed int64 `json:"retryable_failed"`
}

// Progress is the delivery state of one campaign. Channels has both keys.
type Progress struct {
	CampaignID int64                       `json:"campaign_id"`
	Status     string                      `json:"status"`
	Channels   map[Channel]ChannelProgress `json:"channels"`
}

// JobFilter selects send jobs. Empty Status and Channel match all.
// After is the last id of the previous page.
type JobFilter struct {
	Status  Status
	Channel Channel
	Limit   int
	After   int64
}
