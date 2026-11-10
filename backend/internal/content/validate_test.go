package content

import (
	"encoding/json"
	"strings"
	"testing"
)

func mk(sms, subj, body string) []byte {
	b, _ := json.Marshal(map[string]string{"sms": sms, "email_subject": subj, "email_body": body})
	return b
}

func hasErr(errs []FieldError, path, msg string) bool {
	for _, e := range errs {
		if e.Path == path && (msg == "" || strings.Contains(e.Message, msg)) {
			return true
		}
	}
	return false
}

func TestValidate_Limits(t *testing.T) {
	v := "{{first_name}}"
	t.Run("sms 160 rendered passes", func(t *testing.T) {
		_, l, errs := Validate(mk(strings.Repeat("a", 140)+v, "s", "b"), 20)
		if len(errs) != 0 {
			t.Fatalf("errs = %v", errs)
		}
		if l.SMSRunes != 160 {
			t.Errorf("SMSRunes = %d", l.SMSRunes)
		}
	})
	t.Run("sms 161 rendered fails", func(t *testing.T) {
		_, _, errs := Validate(mk(strings.Repeat("a", 141)+v, "s", "b"), 20)
		if !hasErr(errs, "sms", "161 runes after rendering, limit 160") {
			t.Errorf("errs = %v", errs)
		}
	})
	t.Run("subject 80 passes", func(t *testing.T) {
		_, _, errs := Validate(mk("s", strings.Repeat("a", 60)+v, "b"), 20)
		if len(errs) != 0 {
			t.Errorf("errs = %v", errs)
		}
	})
	t.Run("subject 81 fails", func(t *testing.T) {
		_, _, errs := Validate(mk("s", strings.Repeat("a", 61)+v, "b"), 20)
		if !hasErr(errs, "email_subject", "81 runes after rendering, limit 80") {
			t.Errorf("errs = %v", errs)
		}
	})
	t.Run("body 2000 passes", func(t *testing.T) {
		_, _, errs := Validate(mk("s", "s", strings.Repeat("a", 2000)), 20)
		if len(errs) != 0 {
			t.Errorf("errs = %v", errs)
		}
	})
	t.Run("body 2001 fails", func(t *testing.T) {
		_, _, errs := Validate(mk("s", "s", strings.Repeat("a", 2001)), 20)
		if !hasErr(errs, "email_body", "2001 runes after rendering, limit 2000") {
			t.Errorf("errs = %v", errs)
		}
	})
	t.Run("runes not bytes", func(t *testing.T) {
		_, l, errs := Validate(mk(strings.Repeat("ş", 160), strings.Repeat("ş", 80), strings.Repeat("ş", 2000)), 20)
		if len(errs) != 0 {
			t.Fatalf("errs = %v", errs)
		}
		if l.SMSRunes != 160 || l.EmailSubjectRunes != 80 || l.EmailBodyRunes != 2000 {
			t.Errorf("limits = %+v", l)
		}
	})
}

func TestValidate_Cases(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		path string // "" means must pass
		msg  string
	}{
		{"ok", `{"sms":"a","email_subject":"b","email_body":"c"}`, "", ""},
		{"unknown key", `{"sms":"a","email_subject":"b","email_body":"c","x":"d"}`, "x", "unknown key"},
		{"missing field", `{"sms":"a","email_subject":"b"}`, "email_body", "required"},
		{"number for field", `{"sms":1,"email_subject":"b","email_body":"c"}`, "sms", "string"},
		{"null for field", `{"sms":null,"email_subject":"b","email_body":"c"}`, "sms", "string"},
		{"empty after trim", `{"sms":"  ","email_subject":"b","email_body":"c"}`, "sms", "empty"},
		{"newline in subject", `{"sms":"a","email_subject":"b\nc","email_body":"c"}`, "email_subject", "line break"},
		{"CR in subject", `{"sms":"a","email_subject":"b\rc","email_body":"c"}`, "email_subject", "line break"},
		{"bell in sms", `{"sms":"a\u0007","email_subject":"b","email_body":"c"}`, "sms", "control"},
		{"newline in sms", `{"sms":"a\nb","email_subject":"b","email_body":"c"}`, "sms", "control"},
		{"tab in body ok", `{"sms":"a","email_subject":"b","email_body":"c\td\ne"}`, "", ""},
		{"tab in subject", `{"sms":"a","email_subject":"b\tc","email_body":"c"}`, "email_subject", "control"},
		{"tag in sms", `{"sms":"a <b>","email_subject":"b","email_body":"c"}`, "sms", "HTML"},
		{"tag in subject", `{"sms":"a","email_subject":"<b>","email_body":"c"}`, "email_subject", "HTML"},
		{"tag in body", `{"sms":"a","email_subject":"b","email_body":"<b>c"}`, "email_body", "HTML"},
		{"closing tag", `{"sms":"a","email_subject":"b","email_body":"x</a>"}`, "email_body", "HTML"},
		{"comment", `{"sms":"a","email_subject":"b","email_body":"<!-- x"}`, "email_body", "HTML"},
		{"less than ok", `{"sms":"a < b","email_subject":"b","email_body":"c"}`, "", ""},
		{"tag after render in sms", `{"sms":"<{{first_name}}","email_subject":"b","email_body":"c"}`, "sms", "HTML"},
		{"tag after render in subject", `{"sms":"a","email_subject":"<{{first_name}}","email_body":"c"}`, "email_subject", "HTML"},
		{"tag after render in body", `{"sms":"a","email_subject":"b","email_body":"x <{{first_name}}"}`, "email_body", "HTML"},
		{"bad variable", `{"sms":"{{last_name}}","email_subject":"b","email_body":"c"}`, "sms", "variable"},
		{"data after object", `{"sms":"a","email_subject":"b","email_body":"c"} {}`, "", "data after"},
		{"array", `[]`, "", "object"},
		{"not json", `nope`, "", "valid JSON"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, _, errs := Validate([]byte(tc.raw), 20)
			if tc.path == "" && tc.msg == "" {
				if len(errs) != 0 {
					t.Fatalf("errs = %v", errs)
				}
				if c.SMS == "" {
					t.Error("empty content")
				}
				return
			}
			if !hasErr(errs, tc.path, tc.msg) {
				t.Errorf("want error %q at %q, got %v", tc.msg, tc.path, errs)
			}
		})
	}
}

func TestValidate_AllErrors(t *testing.T) {
	_, _, errs := Validate(mk("", "", ""), 20)
	if len(errs) != 3 {
		t.Fatalf("errs = %v", errs)
	}
	for i, p := range []string{"sms", "email_subject", "email_body"} {
		if errs[i].Path != p {
			t.Errorf("errs[%d].Path = %q, want %q", i, errs[i].Path, p)
		}
	}
}

func TestValidate_Trim(t *testing.T) {
	c, l, errs := Validate(mk("  Hi {{first_name}} \n", " Subj\t", "\n Body "), 5)
	if len(errs) != 0 {
		t.Fatalf("errs = %v", errs)
	}
	if c.SMS != "Hi {{first_name}}" || c.EmailSubject != "Subj" || c.EmailBody != "Body" {
		t.Errorf("content = %+v", c)
	}
	want := Limits{FirstNameRunes: 5, SMSRunes: 8, EmailSubjectRunes: 4, EmailBodyRunes: 4}
	if l != want {
		t.Errorf("limits = %+v, want %+v", l, want)
	}
}
