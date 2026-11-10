package segment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

const exampleRule = `{
  "op": "and",
  "conditions": [
    { "field": "city", "op": "eq", "value": "Istanbul" },
    { "field": "last_cart_add_at", "op": "within_days", "value": 30 },
    { "op": "or", "conditions": [
      { "field": "last_purchase_at", "op": "is_null" },
      { "field": "last_purchase_at", "op": "older_than_days", "value": 30 }
    ]}
  ]
}`

func cond(field, op, value string) string {
	if value == "" {
		return fmt.Sprintf(`{"field":%q,"op":%q}`, field, op)
	}
	return fmt.Sprintf(`{"field":%q,"op":%q,"value":%s}`, field, op, value)
}

func root(conds ...string) string {
	return `{"op":"and","conditions":[` + strings.Join(conds, ",") + `]}`
}

func nest(depth int) string {
	s := cond("order_count", "gt", "1")
	for i := 0; i < depth; i++ {
		s = `{"op":"or","conditions":[` + s + `]}`
	}
	return s
}

func TestParseRule_Valid(t *testing.T) {
	r, errs := Default().ParseRule([]byte(exampleRule))
	if len(errs) != 0 {
		t.Fatalf("errors: %v", errs)
	}
	got, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var want bytes.Buffer
	if err := json.Compact(&want, []byte(exampleRule)); err != nil {
		t.Fatal(err)
	}
	if string(got) != want.String() {
		t.Fatalf("got  %s\nwant %s", got, want.String())
	}
}

func TestParseRule_Rejects(t *testing.T) {
	many := make([]string, 21)
	for i := range many {
		many[i] = `"Istanbul"`
	}
	conds21 := make([]string, 21)
	for i := range conds21 {
		conds21[i] = cond("order_count", "gt", "1")
	}
	tests := []struct {
		name, raw, path string
	}{
		{"not json", `{`, ""},
		{"trailing data", root() + ` {}`, ""},
		{"too large", `{"op":"and","conditions":[],"x":"` + strings.Repeat("a", MaxRuleBytes) + `"}`, ""},
		{"root not object", `[]`, ""},
		{"root is condition", cond("city", "eq", `"Ankara"`), ""},
		{"unknown key in root", `{"op":"and","conditions":[],"extra":1}`, "extra"},
		{"unknown key in condition", root(`{"field":"city","op":"eq","value":"Ankara","x":1}`), "conditions[0].x"},
		{"unknown field", root(cond("email", "eq", `"a@b.c"`)), "conditions[0].field"},
		{"injected field", root(cond("city; DROP TABLE users", "eq", `"Ankara"`)), "conditions[0].field"},
		{"unknown operator", root(cond("city", "like", `"A%"`)), "conditions[0].op"},
		{"wrong operator for field", root(cond("city", "gt", `"Ankara"`)), "conditions[0].op"},
		{"group op bad", `{"op":"not","conditions":[]}`, "op"},
		{"string for within_days", root(cond("last_cart_add_at", "within_days", `"30"`)), "conditions[0].value"},
		{"city not on list", root(cond("city", "eq", `"Paris"`)), "conditions[0].value"},
		{"sql in city", root(cond("city", "eq", `"Istanbul' OR 1=1 --"`)), "conditions[0].value"},
		{"days too big", root(cond("created_at", "within_days", "366")), "conditions[0].value"},
		{"days zero", root(cond("created_at", "within_days", "0")), "conditions[0].value"},
		{"age too low", root(cond("age", "gte", "17")), "conditions[0].value"},
		{"order count negative", root(cond("order_count", "eq", "-1")), "conditions[0].value"},
		{"between low over high", root(cond("age", "between", "[40,30]")), "conditions[0].value"},
		{"between wrong length", root(cond("age", "between", "[30]")), "conditions[0].value"},
		{"between not list", root(cond("age", "between", "30")), "conditions[0].value"},
		{"in 21 values", root(cond("city", "in", "["+strings.Join(many, ",")+"]")), "conditions[0].value"},
		{"in empty", root(cond("city", "in", "[]")), "conditions[0].value"},
		{"in item bad", root(cond("city", "in", `["Ankara","Nope"]`)), "conditions[0].value[1]"},
		{"in not list", root(cond("city", "in", `"Ankara"`)), "conditions[0].value"},
		{"eq with list", root(cond("city", "eq", `["Ankara"]`)), "conditions[0].value"},
		{"missing value", root(cond("city", "eq", "")), "conditions[0].value"},
		{"is_null with value", root(cond("last_purchase_at", "is_null", "5")), "conditions[0].value"},
		{"depth 4", root(nest(3)), "conditions[0].conditions[0].conditions[0]"},
		{"21 conditions", root(conds21...), ""},
		{"second condition bad", root(cond("city", "eq", `"Ankara"`), cond("city", "eq", `"Atlantis"`)), "conditions[1].value"},
		{"field not string", root(`{"field":1,"op":"eq","value":1}`), "conditions[0].field"},
		{"condition not object", root(`"x"`), "conditions[0]"},
		{"conditions not list", `{"op":"and","conditions":{}}`, "conditions"},
	}
	a := Default()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := a.ParseRule([]byte(tt.raw))
			if len(errs) == 0 {
				t.Fatal("expected errors")
			}
			if errs[0].Path != tt.path {
				t.Fatalf("path = %q, want %q (%v)", errs[0].Path, tt.path, errs)
			}
		})
	}
}

func TestParseRule_DepthThreeIsValid(t *testing.T) {
	_, errs := Default().ParseRule([]byte(root(nest(2))))
	if len(errs) != 0 {
		t.Fatal(errs)
	}
}

func TestParseRule_TwentyConditionsValid(t *testing.T) {
	c := make([]string, 20)
	for i := range c {
		c[i] = cond("order_count", "gt", "1")
	}
	if _, errs := Default().ParseRule([]byte(root(c...))); len(errs) != 0 {
		t.Fatal(errs)
	}
}

func TestParseRule_AllErrors(t *testing.T) {
	raw := root(
		cond("city", "eq", `"Paris"`),
		cond("age", "gte", "5"),
		cond("last_cart_add_at", "within_days", `"30"`),
	)
	_, errs := Default().ParseRule([]byte(raw))
	if len(errs) != 3 {
		t.Fatalf("got %d errors: %v", len(errs), errs)
	}
	seen := map[string]bool{}
	for _, e := range errs {
		seen[e.Path] = true
	}
	if len(seen) != 3 {
		t.Fatalf("paths not distinct: %v", errs)
	}
}

func TestParseRule_EmptyGroups(t *testing.T) {
	a := Default()
	r, errs := a.ParseRule([]byte(`{"op":"and","conditions":[]}`))
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	if got, _ := json.Marshal(r); string(got) != `{"op":"and","conditions":[]}` {
		t.Fatalf("marshal: %s", got)
	}
	_, errs = a.ParseRule([]byte(root(`{"op":"or","conditions":[]}`)))
	if len(errs) == 0 || errs[0].Path != "conditions[0]" {
		t.Fatalf("errs = %v", errs)
	}
}

func TestParseRule_IntegerLiterals(t *testing.T) {
	a := Default()
	ok := func(kind, op, v string, want bool) {
		t.Helper()
		_, errs := a.ParseRule([]byte(root(cond(kind, op, v))))
		if (len(errs) == 0) != want {
			t.Errorf("%s %s %s: valid=%v errs=%v", kind, op, v, len(errs) == 0, errs)
		}
	}
	ok("last_cart_add_at", "within_days", "30", true)
	ok("last_cart_add_at", "within_days", "30.0", false)
	ok("last_cart_add_at", "within_days", "3e1", false)
	ok("last_cart_add_at", "within_days", `"30"`, false)
	ok("total_spent", "gt", "99.95", true)
	ok("total_spent", "gt", "99.955", false)
	ok("total_spent", "gt", "100", true)
	ok("total_spent", "gt", "99.950", true)
	ok("total_spent", "gt", "1e2", false)
	ok("total_spent", "gt", "-1", false)
	ok("total_spent", "between", "[1.5, 2.25]", true)
	ok("total_spent", "between", "[2.25, 1.5]", false)
}

func TestParseRule_NullValue(t *testing.T) {
	a := Default()
	for _, raw := range []string{
		root(cond("last_purchase_at", "is_null", "")),
		root(cond("last_purchase_at", "is_null", "null")),
		root(cond("last_purchase_at", "is_not_null", "null")),
	} {
		if _, errs := a.ParseRule([]byte(raw)); len(errs) != 0 {
			t.Errorf("%s: %v", raw, errs)
		}
	}
	if _, errs := a.ParseRule([]byte(root(cond("city", "eq", "null")))); len(errs) == 0 {
		t.Error("eq with null must fail")
	}
}

func TestParseRule_CityErrorListsValues(t *testing.T) {
	_, errs := Default().ParseRule([]byte(root(cond("city", "eq", `"İstanbul"`))))
	if len(errs) != 1 || !strings.Contains(errs[0].Message, "Istanbul") || !strings.Contains(errs[0].Message, "Ankara") {
		t.Fatalf("errs = %v", errs)
	}
}

func TestParseRule_RoundTripLists(t *testing.T) {
	raw := root(cond("city", "in", `["Ankara","Izmir"]`), cond("age", "between", "[18,30]"), cond("total_spent", "lt", "99.5"))
	r, errs := Default().ParseRule([]byte(raw))
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	got, _ := json.Marshal(r)
	if string(got) != raw {
		t.Fatalf("got %s", got)
	}
	c := r.Root().Nodes[0].(*Condition)
	if c.Field.Name != "city" || len(c.Values) != 2 || c.Values[1].Str != "Izmir" {
		t.Fatalf("bad tree: %+v", c)
	}
}

func TestClone(t *testing.T) {
	a := Default()
	b := a.Clone()
	b.Fields[0].Enum[0] = "X"
	b.Fields[0].Ops[0] = "x"
	if a.Fields[0].Enum[0] == "X" || a.Fields[0].Ops[0] == "x" {
		t.Fatal("clone shares slices")
	}
}

func TestDecodeAnswer(t *testing.T) {
	a := Default()
	rule := root(cond("city", "eq", `"Ankara"`))
	long := func(n int) string { return `"` + strings.Repeat("é", n) + `"` }
	six := `["a","b","c","d","e","f"]`
	tests := []struct {
		name, raw string
		wantOK    bool
		path      string
	}{
		{"ok", `{"status":"ok","rule":` + rule + `,"unmapped":["x"]}`, true, ""},
		{"ok no unmapped", `{"status":"ok","rule":` + rule + `}`, true, ""},
		{"ok five unmapped", `{"status":"ok","rule":` + rule + `,"unmapped":["a","b","c","d","e"]}`, true, ""},
		{"unsupported", `{"status":"unsupported","reason":"no data"}`, true, ""},
		{"six unmapped", `{"status":"ok","rule":` + rule + `,"unmapped":` + six + `}`, false, "unmapped"},
		{"201 rune item", `{"status":"ok","rule":` + rule + `,"unmapped":[` + long(201) + `]}`, false, "unmapped[0]"},
		{"200 rune item", `{"status":"ok","rule":` + rule + `,"unmapped":[` + long(200) + `]}`, true, ""},
		{"501 rune reason", `{"status":"unsupported","reason":` + long(501) + `}`, false, "reason"},
		{"500 rune reason", `{"status":"unsupported","reason":` + long(500) + `}`, true, ""},
		{"unknown status", `{"status":"maybe"}`, false, "status"},
		{"missing status", `{}`, false, "status"},
		{"ok no rule", `{"status":"ok"}`, false, "rule"},
		{"ok bad rule", `{"status":"ok","rule":` + root(cond("city", "eq", `"Paris"`)) + `}`, false, "rule.conditions[0].value"},
		{"unsupported no reason", `{"status":"unsupported"}`, false, "reason"},
		{"unsupported with rule", `{"status":"unsupported","reason":"r","rule":` + rule + `}`, false, "rule"},
		{"unknown key", `{"status":"unsupported","reason":"r","x":1}`, false, "x"},
		{"unmapped not strings", `{"status":"ok","rule":` + rule + `,"unmapped":[1]}`, false, "unmapped[0]"},
		{"trailing data", `{"status":"unsupported","reason":"r"} x`, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans, errs := a.DecodeAnswer([]byte(tt.raw))
			if tt.wantOK {
				if len(errs) != 0 {
					t.Fatalf("errs = %v", errs)
				}
				if ans.Status == "" {
					t.Fatal("empty status")
				}
				return
			}
			if len(errs) == 0 || errs[0].Path != tt.path {
				t.Fatalf("errs = %v, want first path %q", errs, tt.path)
			}
		})
	}
	ans, _ := a.DecodeAnswer([]byte(`{"status":"ok","rule":` + rule + `,"unmapped":["x"]}`))
	if len(ans.Unmapped) != 1 || ans.Rule.Root().Nodes[0].(*Condition).Values[0].Str != "Ankara" {
		t.Fatalf("answer = %+v", ans)
	}
}
