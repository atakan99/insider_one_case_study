// Package content holds the content types, the variable renderer and the
// validator for generated campaign copy. It imports no other internal package.
package content

const (
	MaxSMSRunes     = 160
	MaxSubjectRunes = 80
	MaxBodyRunes    = 2000
	VarFirstName    = "{{first_name}}"
)

// Content is the copy for one language.
type Content struct {
	SMS          string `json:"sms"`
	EmailSubject string `json:"email_subject"`
	EmailBody    string `json:"email_body"`
}

// Limits holds the rendered rune counts that the final event reports.
type Limits struct {
	FirstNameRunes    int `json:"first_name_runes"`
	SMSRunes          int `json:"sms_runes"`
	EmailSubjectRunes int `json:"email_subject_runes"`
	EmailBodyRunes    int `json:"email_body_runes"`
}

// FieldError is one validation error with the path of the field.
type FieldError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}
