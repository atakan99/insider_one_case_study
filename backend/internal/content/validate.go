package content

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

type fieldRule struct {
	key        string
	max        int
	allowBody  bool // newline and tab are allowed
	noLineEnds bool // CR and LF are rejected with a specific message
}

var fieldRules = []fieldRule{
	{key: "sms", max: MaxSMSRunes},
	{key: "email_subject", max: MaxSubjectRunes, noLineEnds: true},
	{key: "email_body", max: MaxBodyRunes, allowBody: true},
}

// Validate decodes the tool input, trims each field and checks it. The check
// renders with a name of firstNameRunes runes. It returns all errors found.
func Validate(raw []byte, firstNameRunes int) (Content, Limits, []FieldError) {
	var errs []FieldError
	fail := func(path, format string, args ...any) {
		errs = append(errs, FieldError{Path: path, Message: fmt.Sprintf(format, args...)})
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	var msg json.RawMessage
	if err := dec.Decode(&msg); err != nil {
		fail("", "not valid JSON")
		return Content{}, Limits{}, errs
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		fail("", "data after the top-level JSON value")
		return Content{}, Limits{}, errs
	}
	var m map[string]json.RawMessage
	if t := bytes.TrimSpace(msg); len(t) == 0 || t[0] != '{' || json.Unmarshal(msg, &m) != nil {
		fail("", "must be a JSON object")
		return Content{}, Limits{}, errs
	}

	// Unknown keys, in a stable order.
	var unknown []string
	for k := range m {
		known := false
		for _, r := range fieldRules {
			if r.key == k {
				known = true
			}
		}
		if !known {
			unknown = append(unknown, k)
		}
	}
	sort.Strings(unknown)
	for _, k := range unknown {
		fail(truncate(k, 40), "unknown key")
	}

	name := strings.Repeat("x", firstNameRunes)
	limits := Limits{FirstNameRunes: firstNameRunes}
	vals := make(map[string]string, len(fieldRules))
	counts := make(map[string]int, len(fieldRules))

	for _, r := range fieldRules {
		rawVal, has := m[r.key]
		if !has {
			fail(r.key, "is required")
			continue
		}
		var s string
		if isNull(rawVal) || json.Unmarshal(rawVal, &s) != nil {
			fail(r.key, "must be a string")
			continue
		}
		s = strings.TrimSpace(s)
		if s == "" {
			fail(r.key, "must not be empty")
			continue
		}
		vals[r.key] = s
		n := len(errs)
		checkText(s, r, func(format string, args ...any) { fail(r.key, format, args...) })
		if _, err := ParseVars(s); err != nil {
			fail(r.key, "%s", err.Error())
		}
		if len(errs) > n {
			continue
		}
		rendered := utf8.RuneCountInString(Render(s, name))
		counts[r.key] = rendered
		if rendered > r.max {
			fail(r.key, "%d runes after rendering, limit %d", rendered, r.max)
		}
	}

	if len(errs) > 0 {
		return Content{}, Limits{}, errs
	}
	limits.SMSRunes = counts["sms"]
	limits.EmailSubjectRunes = counts["email_subject"]
	limits.EmailBodyRunes = counts["email_body"]
	return Content{
		SMS:          vals["sms"],
		EmailSubject: vals["email_subject"],
		EmailBody:    vals["email_body"],
	}, limits, nil
}

// checkText reports control characters and HTML tags. Each kind is reported
// once for the field.
func checkText(s string, r fieldRule, fail func(format string, args ...any)) {
	lineEnd, control := false, false
	for _, c := range s {
		switch {
		case r.noLineEnds && (c == '\n' || c == '\r'):
			lineEnd = true
		case r.allowBody && (c == '\n' || c == '\t'):
		case unicode.IsControl(c):
			control = true
		}
	}
	if lineEnd {
		fail("must not contain a line break")
	}
	if control {
		fail("must not contain control characters")
	}
	rs := []rune(s)
	for i := 0; i+1 < len(rs); i++ {
		if rs[i] == '<' && (rs[i+1] == '/' || rs[i+1] == '!' || unicode.IsLetter(rs[i+1])) {
			fail("must not contain an HTML tag")
			return
		}
	}
}

func isNull(raw json.RawMessage) bool { return string(bytes.TrimSpace(raw)) == "null" }

func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "..."
}
