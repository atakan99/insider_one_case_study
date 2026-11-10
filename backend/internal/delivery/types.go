// Package delivery sends campaign messages through a queue and a provider.
package delivery

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
