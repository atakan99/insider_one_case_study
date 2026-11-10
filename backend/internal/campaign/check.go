package campaign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/segment"
)

// InvalidError carries the validation errors of a save or an approval.
type InvalidError struct{ Errors []content.FieldError }

func (e *InvalidError) Error() string {
	parts := make([]string, 0, len(e.Errors))
	for _, fe := range e.Errors {
		parts = append(parts, fe.Path+": "+fe.Message)
	}
	return "campaign: invalid content: " + strings.Join(parts, "; ")
}

var entryKeys = map[string]bool{"language": true, "sms": true, "email_subject": true, "email_body": true}

// CheckContents decodes body ({"contents": [...]}) and checks it against the
// languages of the draft and the current audience. It returns the trimmed
// contents with the new counts and first_name_runes, sorted by language.
func CheckContents(body json.RawMessage, draftLangs []string, aud map[string]segment.LanguageCounts) ([]LanguageContent, []content.FieldError) {
	var errs []content.FieldError
	fail := func(path, format string, args ...any) {
		errs = append(errs, content.FieldError{Path: path, Message: fmt.Sprintf(format, args...)})
	}

	var top map[string]json.RawMessage
	if t := bytes.TrimSpace(body); len(t) == 0 || t[0] != '{' || json.Unmarshal(body, &top) != nil {
		fail("", "must be a JSON object")
		return nil, errs
	}
	var entries []json.RawMessage
	if raw, ok := top["contents"]; !ok || json.Unmarshal(raw, &entries) != nil || len(entries) == 0 {
		fail("contents", "must be an array with one entry for each language")
		return nil, errs
	}

	inDraft := map[string]bool{}
	for _, l := range draftLangs {
		inDraft[l] = true
	}
	seen := map[string]bool{}
	var out []LanguageContent
	for i, raw := range entries {
		base := fmt.Sprintf("contents[%d]", i)
		var m map[string]json.RawMessage
		if t := bytes.TrimSpace(raw); len(t) == 0 || t[0] != '{' || json.Unmarshal(raw, &m) != nil {
			fail(base, "must be a JSON object")
			continue
		}
		var unknown []string
		for k := range m {
			if !entryKeys[k] {
				unknown = append(unknown, k)
			}
		}
		sort.Strings(unknown)
		for _, k := range unknown {
			fail(base+"."+k, "unknown key")
		}

		var lang string
		if json.Unmarshal(m["language"], &lang) != nil || lang == "" {
			fail(base+".language", "must be a language code")
			continue
		}
		switch l, ok := aud[lang]; {
		case !inDraft[lang]:
			fail(base+".language", "%q is not a language of this draft", lang)
			continue
		case seen[lang]:
			fail(base+".language", "%q appears more than once", lang)
			continue
		case !ok || l.Reachable == 0:
			seen[lang] = true
			fail(base+".language", "no user of language %q can get a message now", lang)
			continue
		}
		seen[lang] = true
		l := aud[lang]

		text := map[string]json.RawMessage{}
		for _, k := range []string{"sms", "email_subject", "email_body"} {
			if v, ok := m[k]; ok {
				text[k] = v
			}
		}
		raw, _ := json.Marshal(text)
		c, _, verrs := content.Validate(raw, int(l.MaxFirstName))
		for _, fe := range verrs {
			p := base
			if fe.Path != "" {
				p += "." + fe.Path
			}
			errs = append(errs, content.FieldError{Path: p, Message: fe.Message})
		}
		if len(verrs) == 0 && len(unknown) == 0 {
			out = append(out, LanguageContent{Language: lang, Content: c,
				SMSCount: l.SMS, EmailCount: l.Email, FirstNameRunes: int(l.MaxFirstName)})
		}
	}
	for _, l := range draftLangs {
		if !seen[l] {
			fail("contents", "language %q of the draft has no entry", l)
		}
	}
	if len(errs) > 0 {
		return nil, errs
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Language < out[j].Language })
	return out, nil
}
