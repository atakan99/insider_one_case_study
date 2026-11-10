package content

import (
	"strings"
	"testing"
)

func TestMockContent(t *testing.T) {
	for _, lang := range []string{"tr", "en", "xx"} {
		c, _, errs := Validate(MockContent(lang), 30)
		if len(errs) > 0 {
			t.Fatalf("%s: %v", lang, errs)
		}
		if c.SMS == "" {
			t.Errorf("%s: empty sms", lang)
		}
	}
	c, _, _ := Validate(MockContent("tr"), 30)
	if !strings.Contains(c.SMS, "%30") {
		t.Errorf("tr sms = %q", c.SMS)
	}
}
