package content

import (
	"strings"
	"testing"
)

const exampleBrief = "Son 30 günde sepetine ürün ekleyip alışveriş yapmayan müşterilere %30 indirim"

func TestSystemPrompt(t *testing.T) {
	p := SystemPrompt()
	for _, s := range []string{"160", "80", "{{first_name}}", "plain text", "<brief>", ToolName} {
		if !strings.Contains(p, s) {
			t.Errorf("system prompt lacks %q", s)
		}
	}
	if strings.Contains(p, exampleBrief) {
		t.Error("system prompt holds the brief")
	}
	if p != SystemPrompt() {
		t.Error("system prompt is not constant")
	}
}

func TestUserMessage(t *testing.T) {
	m := UserMessage("tr", "users in Istanbul", 20, exampleBrief)
	for _, s := range []string{"tr", "Turkish", "users in Istanbul", "20", "<brief>\n" + exampleBrief + "\n</brief>"} {
		if !strings.Contains(m, s) {
			t.Errorf("user message lacks %q", s)
		}
	}
	if lang, ok := LanguageOf(UserMessage("en", "d", 5, "b")); lang != "en" || !ok {
		t.Errorf("LanguageOf = %q, %v", lang, ok)
	}
	if lang, ok := LanguageOf(m); lang != "tr" || !ok {
		t.Errorf("LanguageOf = %q, %v", lang, ok)
	}
	if _, ok := LanguageOf("<brief>x</brief>"); ok {
		t.Error("LanguageOf accepted other text")
	}
}

func TestRepairMessage(t *testing.T) {
	m := RepairMessage([]FieldError{{Path: "sms", Message: "too long"}})
	if !strings.Contains(m, "sms: too long") || !strings.Contains(m, ToolName) {
		t.Errorf("repair message = %q", m)
	}
}

func TestToolSchema(t *testing.T) {
	s := ToolSchema()
	req, _ := s["required"].([]any)
	if len(req) != 3 || req[0] != "sms" || req[1] != "email_subject" || req[2] != "email_body" {
		t.Errorf("required = %v", req)
	}
	if s["additionalProperties"] != false {
		t.Error("additionalProperties is not false")
	}
	props := s["properties"].(map[string]any)
	if len(props) != 3 {
		t.Errorf("props = %v", props)
	}
	for k, v := range props {
		if _, has := v.(map[string]any)["maxLength"]; has {
			t.Errorf("%s has maxLength", k)
		}
	}
}
