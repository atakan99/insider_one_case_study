// Package campaign stores draft and approved campaigns and checks the edits
// of the marketer. It imports only content, segment and store from internal.
package campaign

import (
	"encoding/json"
	"errors"
	"time"

	"campaigncomposer/internal/content"
)

// Status is the state of a campaign.
type Status string

const (
	StatusDraft    Status = "draft"
	StatusApproved Status = "approved"
)

// LanguageContent is the copy of one language with the audience snapshot of
// the last save.
type LanguageContent struct {
	Language string `json:"language"`
	content.Content
	SMSCount       int64 `json:"sms_count"`
	EmailCount     int64 `json:"email_count"`
	FirstNameRunes int   `json:"first_name_runes"`
}

// Campaign is one stored campaign. Contents are sorted by language.
type Campaign struct {
	ID               int64             `json:"id"`
	Status           Status            `json:"status"`
	Brief            string            `json:"brief"`
	Rule             json.RawMessage   `json:"-"`
	Description      string            `json:"description"`
	SkippedLanguages []string          `json:"skipped_languages"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	ApprovedAt       *time.Time        `json:"approved_at"`
	Contents         []LanguageContent `json:"contents"`
}

// Summary is one row of the campaign list. It has no content text.
type Summary struct {
	ID          int64      `json:"id"`
	Status      Status     `json:"status"`
	Description string     `json:"description"`
	Languages   []string   `json:"languages"`
	CreatedAt   time.Time  `json:"created_at"`
	ApprovedAt  *time.Time `json:"approved_at"`
}

// Draft is the input of CreateDraft and ReplaceDraft.
type Draft struct {
	Brief, Description string
	Rule               json.RawMessage // canonical JSON from segment.Rule.MarshalJSON
	Skipped            []string
	Contents           []LanguageContent
}

var (
	ErrNotFound = errors.New("campaign: not found")
	ErrNotDraft = errors.New("campaign: not a draft")
	ErrConflict = errors.New("campaign: brief or rule differs, or not a draft")
)
