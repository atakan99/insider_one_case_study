package mockprovider

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"unicode/utf8"
)

const (
	maxSMSBody      = 160
	maxEmailSubject = 80
	maxReference    = 128
)

var e164 = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)

type smsRequest struct {
	To        string `json:"to"`
	Body      string `json:"body"`
	Reference string `json:"reference"`
}

type emailRequest struct {
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Reference string `json:"reference"`
}

func validateSMS(r smsRequest) error {
	if r.To == "" {
		return errors.New("to is required")
	}
	if !e164.MatchString(r.To) {
		return errors.New("to must be an E.164 phone number, for example +905321234567")
	}
	if err := checkLength("body", r.Body, 1, maxSMSBody); err != nil {
		return err
	}
	return checkLength("reference", r.Reference, 0, maxReference)
}

func validateEmail(r emailRequest) error {
	if r.To == "" {
		return errors.New("to is required")
	}
	addr, err := mail.ParseAddress(r.To)
	if err != nil || addr.Name != "" || addr.Address != r.To {
		return errors.New("to must be one email address with no display name")
	}
	if err := checkLength("subject", r.Subject, 1, maxEmailSubject); err != nil {
		return err
	}
	if r.Body == "" {
		return errors.New("body is required")
	}
	return checkLength("reference", r.Reference, 0, maxReference)
}

// checkLength counts runes, not bytes.
func checkLength(field, v string, minLen, maxLen int) error {
	n := utf8.RuneCountInString(v)
	if n < minLen {
		return fmt.Errorf("%s is required", field)
	}
	if n > maxLen {
		return fmt.Errorf("%s is longer than %d characters", field, maxLen)
	}
	return nil
}
