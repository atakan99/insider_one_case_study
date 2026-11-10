package content

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPartialContent(t *testing.T) {
	final := Content{
		SMS:          "Merhaba {{first_name}}, \"sepetin\" seni bekliyor ş",
		EmailSubject: "Sepetin seni bekliyor ç",
		EmailBody:    "Selam\n\nİndirim \"%30\" 😀 şimdi\tçok güzel",
	}
	raw, err := json.Marshal(final)
	if err != nil {
		t.Fatal(err)
	}
	// Also an escaped form with \u sequences.
	escaped := []byte(`{"sms":"aşb😀c","email_subject":"s","email_body":"b"}`)
	for _, in := range [][]byte{raw, escaped} {
		var want Content
		if err := json.Unmarshal(in, &want); err != nil {
			t.Fatal(err)
		}
		for i := 0; i <= len(in); i++ {
			got := PartialContent(in[:i])
			if !strings.HasPrefix(want.SMS, got.SMS) || !strings.HasPrefix(want.EmailSubject, got.EmailSubject) || !strings.HasPrefix(want.EmailBody, got.EmailBody) {
				t.Fatalf("prefix %d (%q) gave %+v", i, in[:i], got)
			}
		}
		if got := PartialContent(in); got != want {
			t.Errorf("full input gave %+v, want %+v", got, want)
		}
	}
}

func TestPartialContentCuts(t *testing.T) {
	if got := PartialContent([]byte(`{"sms":"ab\u00`)); got.SMS != "ab" {
		t.Errorf("cut escape: %q", got.SMS)
	}
	if got := PartialContent([]byte(`{"sms":"ab\`)); got.SMS != "ab" {
		t.Errorf("cut backslash: %q", got.SMS)
	}
	if got := PartialContent([]byte("{\"sms\":\"a\xc5")); got.SMS != "a" {
		t.Errorf("cut rune: %q", got.SMS)
	}
	if got := PartialContent([]byte(`{"sms":"a\ud83d`)); got.SMS != "a" {
		t.Errorf("cut pair: %q", got.SMS)
	}
	if got := PartialContent([]byte(`garbage`)); got != (Content{}) {
		t.Errorf("garbage: %+v", got)
	}
	if got := PartialContent(nil); got != (Content{}) {
		t.Errorf("nil: %+v", got)
	}
}
