package segment

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// MaxBriefRunes is the longest brief that SanitizeBrief accepts.
const MaxBriefRunes = 1000

var (
	// ErrBriefEmpty means the brief has no letter.
	ErrBriefEmpty = errors.New("brief has no letter")
	// ErrBriefTooLong means the brief has more than MaxBriefRunes runes.
	ErrBriefTooLong = errors.New("brief is too long")
)

var briefTag = regexp.MustCompile(`(?i)<\s*/?\s*brief\s*>`)

// SanitizeBrief cleans a brief and checks it. It normalizes to NFC, makes
// line breaks "\n", drops control and format characters (except \n and \t),
// removes the <brief> delimiters and trims spaces.
func SanitizeBrief(s string) (string, error) {
	s = norm.NFC.String(strings.ToValidUTF8(s, ""))
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}
		if unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, s)
	// Repeat, so that "<br<brief>ief>" cannot rebuild a tag.
	for {
		t := briefTag.ReplaceAllString(s, "")
		if t == s {
			break
		}
		s = t
	}
	s = strings.TrimSpace(s)
	if !strings.ContainsFunc(s, unicode.IsLetter) {
		return "", ErrBriefEmpty
	}
	if utf8.RuneCountInString(s) > MaxBriefRunes {
		return "", ErrBriefTooLong
	}
	return s, nil
}
