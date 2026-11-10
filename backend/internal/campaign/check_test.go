package campaign

import (
	"strings"
	"testing"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/segment"
)

func audOf(langs ...segment.LanguageCounts) map[string]segment.LanguageCounts {
	m := map[string]segment.LanguageCounts{}
	for _, l := range langs {
		m[l.Code] = l
	}
	return m
}

func lcount(code string, reachable, sms, email, maxName int64) segment.LanguageCounts {
	return segment.LanguageCounts{Code: code, Counts: segment.Counts{SMS: sms, Email: email, Reachable: reachable, Matched: reachable}, MaxFirstName: maxName}
}

const okTR = `{"language":"tr","sms":"  Merhaba {{first_name}}  ","email_subject":"Konu","email_body":"Gövde"}`
const okEN = `{"language":"en","sms":"Hi {{first_name}}","email_subject":"Subject","email_body":"Body"}`

func paths(errs []content.FieldError) []string {
	var p []string
	for _, e := range errs {
		p = append(p, e.Path)
	}
	return p
}

func hasPath(errs []content.FieldError, path string) bool {
	for _, e := range errs {
		if e.Path == path {
			return true
		}
	}
	return false
}

func TestCheckContents_Valid(t *testing.T) {
	aud := audOf(lcount("tr", 100, 80, 90, 12), lcount("en", 10, 8, 9, 6))
	got, errs := CheckContents([]byte(`{"contents":[`+okTR+`,`+okEN+`]}`), []string{"en", "tr"}, aud)
	if len(errs) > 0 {
		t.Fatalf("errors: %v", errs)
	}
	if len(got) != 2 || got[0].Language != "en" || got[1].Language != "tr" {
		t.Fatalf("got %+v", got)
	}
	tr := got[1]
	if tr.SMS != "Merhaba {{first_name}}" {
		t.Errorf("sms not trimmed: %q", tr.SMS)
	}
	if tr.SMSCount != 80 || tr.EmailCount != 90 || tr.FirstNameRunes != 12 {
		t.Errorf("counts = %+v", tr)
	}
}

func TestCheckContents(t *testing.T) {
	aud := audOf(lcount("tr", 100, 80, 90, 20), lcount("en", 10, 8, 9, 6))
	long := func(n int) string { return strings.Repeat("a", n) }
	tests := []struct {
		name, body string
		langs      []string
		want       string // a path that must be in the errors
	}{
		{"not an object", `[]`, []string{"tr"}, ""},
		{"no contents", `{}`, []string{"tr"}, "contents"},
		{"empty array", `{"contents":[]}`, []string{"tr"}, "contents"},
		{"missing draft language", `{"contents":[` + okTR + `]}`, []string{"tr", "en"}, "contents"},
		{"language not in draft", `{"contents":[` + okTR + `,` + okEN + `]}`, []string{"tr"}, "contents[1].language"},
		{"same language twice", `{"contents":[` + okTR + `,` + okTR + `]}`, []string{"tr"}, "contents[1].language"},
		{"unknown key", `{"contents":[{"language":"tr","sms":"a","email_subject":"b","email_body":"c","x":1}]}`, []string{"tr"}, "contents[0].x"},
		{"entry not an object", `{"contents":["x"]}`, []string{"tr"}, "contents[0]"},
		// 140 runes + {{first_name}} rendered with 20 runes = 160 passes; 141 fails.
		{"sms 161 runes", `{"contents":[{"language":"tr","sms":"{{first_name}}` + long(141) + `","email_subject":"b","email_body":"c"}]}`, []string{"tr"}, "contents[0].sms"},
		{"subject 81 runes", `{"contents":[{"language":"tr","sms":"a","email_subject":"{{first_name}}` + long(61) + `","email_body":"c"}]}`, []string{"tr"}, "contents[0].email_subject"},
		{"subject newline", `{"contents":[{"language":"tr","sms":"a","email_subject":"b\nc","email_body":"c"}]}`, []string{"tr"}, "contents[0].email_subject"},
		{"wrong variable", `{"contents":[{"language":"tr","sms":"{{last_name}}","email_subject":"b","email_body":"c"}]}`, []string{"tr"}, "contents[0].sms"},
		{"html tag", `{"contents":[{"language":"tr","sms":"a","email_subject":"b","email_body":"<b>c</b>"}]}`, []string{"tr"}, "contents[0].email_body"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := CheckContents([]byte(tc.body), tc.langs, aud)
			if len(errs) == 0 {
				t.Fatal("want errors, got none")
			}
			if tc.want != "" && !hasPath(errs, tc.want) {
				t.Fatalf("paths %v, want %q", paths(errs), tc.want)
			}
		})
	}
}

func TestCheckContents_LimitBoundary(t *testing.T) {
	aud := audOf(lcount("tr", 100, 80, 90, 20))
	body := `{"contents":[{"language":"tr","sms":"{{first_name}}` + strings.Repeat("a", 140) + `","email_subject":"{{first_name}}` + strings.Repeat("b", 60) + `","email_body":"c"}]}`
	if _, errs := CheckContents([]byte(body), []string{"tr"}, aud); len(errs) > 0 {
		t.Fatalf("160 and 80 rendered runes must pass: %v", errs)
	}
}

func TestCheckContents_TwoBadEntries(t *testing.T) {
	aud := audOf(lcount("tr", 100, 80, 90, 5), lcount("en", 10, 8, 9, 5))
	body := `{"contents":[{"language":"tr","sms":"","email_subject":"b","email_body":"c"},{"language":"en","sms":"a","email_subject":"","email_body":"c"}]}`
	_, errs := CheckContents([]byte(body), []string{"tr", "en"}, aud)
	if !hasPath(errs, "contents[0].sms") || !hasPath(errs, "contents[1].email_subject") {
		t.Fatalf("paths %v", paths(errs))
	}
}

func TestCheckContents_NoReachable(t *testing.T) {
	aud := audOf(lcount("tr", 0, 0, 0, 0))
	_, errs := CheckContents([]byte(`{"contents":[`+okTR+`]}`), []string{"tr"}, aud)
	if !hasPath(errs, "contents[0].language") {
		t.Fatalf("paths %v, want contents[0].language", paths(errs))
	}
	_, errs = CheckContents([]byte(`{"contents":[`+okTR+`]}`), []string{"tr"}, audOf())
	if !hasPath(errs, "contents[0].language") {
		t.Fatalf("language missing from audience: paths %v", paths(errs))
	}
}
