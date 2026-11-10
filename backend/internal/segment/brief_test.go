package segment_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"campaigncomposer/internal/segment"
)

var testNowPrompt = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

func TestSanitizeBrief(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
		err  error
	}{
		{"nfd to nfc", "İstanbul", "İstanbul", nil},
		{"crlf", "a\r\nb", "a\nb", nil},
		{"cr", "a\rb", "a\nb", nil},
		{"tab and newline stay", "a\tb\nc", "a\tb\nc", nil},
		{"nul removed", "a\x00b", "ab", nil},
		{"zero width removed", "a​b", "ab", nil},
		{"bidi override removed", "a‮b", "ab", nil},
		{"tag upper", "a</BRIEF>b", "ab", nil},
		{"tag mixed open", "a<Brief>b", "ab", nil},
		{"trim", "  x  ", "x", nil},
		{"no letter", "!!! 123", "", segment.ErrBriefEmpty},
		{"empty", "", "", segment.ErrBriefEmpty},
		{"only tags", "<brief></brief>", "", segment.ErrBriefEmpty},
		{"1000 runes", strings.Repeat("ş", 1000), strings.Repeat("ş", 1000), nil},
		{"1001 runes", strings.Repeat("ş", 1001), "", segment.ErrBriefTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := segment.SanitizeBrief(tt.in)
			if !errors.Is(err, tt.err) {
				t.Fatalf("err = %v, want %v", err, tt.err)
			}
			if err == nil && got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
