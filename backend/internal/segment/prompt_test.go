package segment_test

import (
	"encoding/json"
	"strings"
	"testing"

	"campaigncomposer/internal/segment"
)

func TestSystemPrompt_ListsAllowlist(t *testing.T) {
	a := segment.Default()
	p := a.SystemPrompt()
	for _, f := range a.Fields {
		if !strings.Contains(p, f.Name) {
			t.Errorf("prompt lacks field %q", f.Name)
		}
		for _, op := range f.Ops {
			if !strings.Contains(p, string(op)) {
				t.Errorf("prompt lacks operator %q", op)
			}
		}
		for _, e := range f.Enum {
			if !strings.Contains(p, e) {
				t.Errorf("prompt lacks enum value %q", e)
			}
		}
	}
	for _, s := range []string{"submit_segment", "unsupported", "unmapped"} {
		if !strings.Contains(p, s) {
			t.Errorf("prompt lacks %q", s)
		}
	}
	if strings.Contains(p, "<brief>\n") {
		t.Error("prompt contains a brief block")
	}
}

func TestSystemPrompt_NoClock(t *testing.T) {
	if a, b := segment.Default().SystemPrompt(), segment.Default().SystemPrompt(); a != b {
		t.Error("prompt is not stable")
	}
}

func walk(t *testing.T, v any, fn func(m map[string]any)) {
	t.Helper()
	switch x := v.(type) {
	case map[string]any:
		fn(x)
		for _, c := range x {
			walk(t, c, fn)
		}
	case []any:
		for _, c := range x {
			walk(t, c, fn)
		}
	}
}

func roundTrip(t *testing.T, a segment.Allowlist) map[string]any {
	t.Helper()
	b, err := json.Marshal(a.ToolSchema())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "$ref") {
		t.Fatal("schema has $ref")
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestToolSchema(t *testing.T) {
	a := segment.Default()
	m := roundTrip(t, a)
	objects, groups := 0, 0
	var fieldEnums [][]any
	walk(t, m, func(n map[string]any) {
		if n["type"] == "object" {
			objects++
			if ap, ok := n["additionalProperties"]; !ok || ap != false {
				t.Errorf("object without additionalProperties=false: %v", n)
			}
		}
		props, _ := n["properties"].(map[string]any)
		if props == nil {
			return
		}
		if _, ok := props["conditions"]; ok {
			groups++
		}
		if f, ok := props["field"].(map[string]any); ok {
			fieldEnums = append(fieldEnums, f["enum"].([]any))
		}
	})
	if objects == 0 {
		t.Fatal("no objects")
	}
	// Root rule group (level 1) plus level 2 and level 3 groups.
	if groups != 3 {
		t.Errorf("group levels = %d, want 3", groups)
	}
	if len(fieldEnums) == 0 {
		t.Fatal("no field enum")
	}
	for _, e := range fieldEnums {
		if len(e) != len(a.Fields) {
			t.Fatalf("field enum has %d names, want %d", len(e), len(a.Fields))
		}
		for i, f := range a.Fields {
			if e[i] != f.Name {
				t.Errorf("enum[%d] = %v, want %s", i, e[i], f.Name)
			}
		}
	}
	req, _ := m["required"].([]any)
	// rule must be required: with strict decoding the model leaves out an
	// optional rule and the answer has status ok with no rule. The model sends
	// null for rule when the status is unsupported.
	if len(req) != 2 || req[0] != "status" || req[1] != "rule" {
		t.Errorf("required = %v", req)
	}
	rule, _ := m["properties"].(map[string]any)["rule"].(map[string]any)
	anyOf, _ := rule["anyOf"].([]any)
	hasNull := false
	for _, v := range anyOf {
		if n, _ := v.(map[string]any); n["type"] == "null" {
			hasNull = true
		}
	}
	if !hasNull {
		t.Errorf("rule schema does not allow null: %v", rule)
	}
}

func TestAllowlist_FakeFieldReachesAll(t *testing.T) {
	a := segment.Default().Clone()
	a.Fields = append(a.Fields, segment.FieldSpec{
		Name: "zz_test", Column: "order_count", Kind: segment.KindInt,
		Ops: []segment.Operator{segment.OpEq}, Min: 0, Max: 9, Phrase: "zz phrase",
	})
	r, errs := a.ParseRule([]byte(`{"op":"and","conditions":[{"field":"zz_test","op":"eq","value":3}]}`))
	if len(errs) > 0 {
		t.Fatalf("ParseRule: %v", errs)
	}
	w := a.Compile(r, testNowPrompt)
	if !strings.Contains(w.SQL(), "order_count = $1") {
		t.Errorf("sql = %s", w.SQL())
	}
	if d := a.Describe(r); !strings.Contains(d, "zz phrase") {
		t.Errorf("describe = %s", d)
	}
	if !strings.Contains(a.SystemPrompt(), "zz_test") {
		t.Error("prompt lacks zz_test")
	}
	if !strings.Contains(mustJSON(t, a.ToolSchema()), `"zz_test"`) {
		t.Error("schema lacks zz_test")
	}
}

func TestUserMessage(t *testing.T) {
	got := segment.UserMessage("hello")
	if got != "<brief>\nhello\n</brief>" {
		t.Errorf("got %q", got)
	}
}

func TestRepairMessage(t *testing.T) {
	got := segment.RepairMessage([]segment.FieldError{
		{Path: "conditions[0].value", Message: "must be an integer"},
		{Path: "status", Message: "bad status"},
	})
	for _, s := range []string{"conditions[0].value", "must be an integer", "status", "bad status", segment.ToolName} {
		if !strings.Contains(got, s) {
			t.Errorf("missing %q in %q", s, got)
		}
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
