package segment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	maxUnmapped     = 5
	maxUnmappedRune = 200
	maxReasonRune   = 500
)

// Answer is the decoded LLM answer. Status is "ok" or "unsupported".
type Answer struct {
	Status   string
	Rule     Rule
	Unmapped []string
	Reason   string
}

type validator struct {
	a     Allowlist
	errs  []FieldError
	conds int
}

func (v *validator) fail(path, format string, args ...any) {
	v.errs = append(v.errs, FieldError{Path: path, Message: fmt.Sprintf(format, args...)})
}

// ParseRule validates raw JSON against the allowlist and returns the
// validated rule, or all errors that it found.
func (a Allowlist) ParseRule(raw []byte) (Rule, []FieldError) {
	v := &validator{a: a}
	msg, ok := v.checkDocument(raw)
	if !ok {
		return Rule{}, v.errs
	}
	g := v.group("", msg, 1, true)
	if v.conds > MaxConditions {
		v.fail("", "the rule has more than %d conditions in total", MaxConditions)
	}
	if len(v.errs) > 0 {
		return Rule{}, v.errs
	}
	return Rule{root: g}, nil
}

// checkDocument checks size, JSON syntax, trailing data and that the
// document is an object.
func (v *validator) checkDocument(raw []byte) (json.RawMessage, bool) {
	if len(raw) > MaxRuleBytes {
		v.fail("", "the JSON is more than %d bytes", MaxRuleBytes)
		return nil, false
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	var msg json.RawMessage
	if err := dec.Decode(&msg); err != nil {
		v.fail("", "not valid JSON: %v", err)
		return nil, false
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		v.fail("", "data after the top-level JSON value")
		return nil, false
	}
	return msg, true
}

func join(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

// object decodes raw as a JSON object. It reports keys that are not allowed.
func (v *validator) object(path string, raw json.RawMessage, allowed ...string) (map[string]json.RawMessage, bool) {
	var m map[string]json.RawMessage
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		v.fail(path, "must be a JSON object")
		return nil, false
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		v.fail(path, "must be a JSON object")
		return nil, false
	}
	for k := range m {
		ok := false
		for _, al := range allowed {
			if k == al {
				ok = true
			}
		}
		if !ok {
			v.fail(join(path, k), "unknown key")
		}
	}
	return m, true
}

func (v *validator) group(path string, raw json.RawMessage, depth int, root bool) Group {
	var g Group
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		v.fail(path, "must be a JSON object")
		return g
	}
	var probe map[string]json.RawMessage
	if json.Unmarshal(raw, &probe) == nil {
		if _, isCond := probe["field"]; isCond {
			if root {
				v.fail(path, "the root must be a group with op and conditions")
			} else {
				v.fail(path, "a group cannot have a field key")
			}
			return g
		}
	}
	m, ok := v.object(path, raw, "op", "conditions")
	if !ok {
		return g
	}
	if depth > MaxDepth {
		v.fail(path, "the rule is deeper than %d levels", MaxDepth)
		return g
	}
	opRaw, hasOp := m["op"]
	var op string
	if !hasOp || json.Unmarshal(opRaw, &op) != nil || (op != "and" && op != "or") {
		v.fail(join(path, "op"), `must be "and" or "or"`)
	} else {
		g.Op = GroupOp(op)
	}
	condsRaw, has := m["conditions"]
	var items []json.RawMessage
	if !has || json.Unmarshal(condsRaw, &items) != nil || bytes.TrimSpace(condsRaw)[0] != '[' {
		v.fail(join(path, "conditions"), "must be a list")
		return g
	}
	if len(items) == 0 && !root {
		v.fail(path, "a group must have at least one condition")
	}
	g.Nodes = []Node{}
	for i, item := range items {
		p := join(path, fmt.Sprintf("conditions[%d]", i))
		var probe map[string]json.RawMessage
		if json.Unmarshal(item, &probe) == nil {
			if _, isCond := probe["field"]; isCond {
				v.conds++
				if c := v.condition(p, item); c != nil {
					g.Nodes = append(g.Nodes, c)
				}
				continue
			}
		}
		sub := v.group(p, item, depth+1, false)
		g.Nodes = append(g.Nodes, &sub)
	}
	return g
}

func (v *validator) condition(path string, raw json.RawMessage) *Condition {
	m, ok := v.object(path, raw, "field", "op", "value")
	if !ok {
		return nil
	}
	bad := false
	var name string
	if json.Unmarshal(m["field"], &name) != nil {
		v.fail(join(path, "field"), "must be a string")
		return nil
	}
	spec, ok := v.a.Field(name)
	if !ok {
		v.fail(join(path, "field"), "unknown field %q", truncate(name, 40))
		bad = true
	}
	var opName string
	opRaw, hasOp := m["op"]
	if !hasOp || json.Unmarshal(opRaw, &opName) != nil {
		v.fail(join(path, "op"), "must be a string")
		return nil
	}
	op := Operator(opName)
	if ok && !spec.Allows(op) {
		v.fail(join(path, "op"), "operator %q is not valid for field %q; allowed: %s", truncate(opName, 40), spec.Name, joinOps(spec.Ops))
		bad = true
	}
	if bad {
		return nil
	}
	c := &Condition{Field: spec, Op: op}
	valRaw, hasVal := m["value"]
	vp := join(path, "value")
	switch op {
	case OpIsNull, OpIsNotNull:
		if hasVal && string(bytes.TrimSpace(valRaw)) != "null" {
			v.fail(vp, "must not be set for %s", op)
			return nil
		}
		return c
	}
	if !hasVal {
		v.fail(vp, "is required for %s", op)
		return nil
	}
	n := len(v.errs)
	switch op {
	case OpIn, OpNotIn, OpBetween:
		var items []json.RawMessage
		if t := bytes.TrimSpace(valRaw); len(t) == 0 || t[0] != '[' || json.Unmarshal(valRaw, &items) != nil {
			v.fail(vp, "must be a list")
			return nil
		}
		if op == OpBetween && len(items) != 2 {
			v.fail(vp, "must be a list of exactly 2 values, low first")
			return nil
		}
		if op != OpBetween && (len(items) < 1 || len(items) > MaxListValues) {
			v.fail(vp, "must be a list of 1 to %d values", MaxListValues)
			return nil
		}
		for i, it := range items {
			if val, ok := v.scalar(fmt.Sprintf("%s[%d]", vp, i), spec, it); ok {
				c.Values = append(c.Values, val)
			}
		}
		if op == OpBetween && len(v.errs) == n && numLess(spec, c.Values[1], c.Values[0]) {
			v.fail(vp, "low must be less than or equal to high")
		}
	default:
		if val, ok := v.scalar(vp, spec, valRaw); ok {
			c.Values = append(c.Values, val)
		}
	}
	if len(v.errs) > n {
		return nil
	}
	return c
}

func numLess(spec *FieldSpec, a, b Value) bool {
	if spec.Kind == KindDecimal {
		fa, _ := strconv.ParseFloat(a.Dec, 64)
		fb, _ := strconv.ParseFloat(b.Dec, 64)
		return fa < fb
	}
	return a.Int < b.Int
}

func (v *validator) scalar(path string, spec *FieldSpec, raw json.RawMessage) (Value, bool) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var x any
	if err := dec.Decode(&x); err != nil {
		v.fail(path, "not valid JSON")
		return Value{}, false
	}
	switch spec.Kind {
	case KindEnum:
		s, ok := x.(string)
		if ok {
			for _, e := range spec.Enum {
				if e == s {
					return Value{Str: s}, true
				}
			}
		}
		v.fail(path, "must be one of: %s", strings.Join(spec.Enum, ", "))
		return Value{}, false
	case KindDecimal:
		num, ok := x.(json.Number)
		if !ok || !validDecimal(num.String(), spec) {
			v.fail(path, "must be a number from %s to %s with at most 2 decimals", fmtNum(spec.Min), fmtNum(spec.Max))
			return Value{}, false
		}
		return Value{Dec: num.String()}, true
	default:
		num, ok := x.(json.Number)
		if ok && !strings.ContainsAny(num.String(), ".eE") {
			if i, err := strconv.ParseInt(num.String(), 10, 64); err == nil && float64(i) >= spec.Min && float64(i) <= spec.Max {
				return Value{Int: i}, true
			}
		}
		v.fail(path, "must be an integer from %s to %s", fmtNum(spec.Min), fmtNum(spec.Max))
		return Value{}, false
	}
}

func validDecimal(s string, spec *FieldSpec) bool {
	if strings.ContainsAny(s, "eE") {
		return false
	}
	if i := strings.IndexByte(s, '.'); i >= 0 {
		if len(strings.TrimRight(s[i+1:], "0")) > 2 {
			return false
		}
	}
	f, err := strconv.ParseFloat(s, 64)
	return err == nil && !math.IsInf(f, 0) && f >= spec.Min && f <= spec.Max
}

func fmtNum(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }

func joinOps(ops []Operator) string {
	s := make([]string, len(ops))
	for i, o := range ops {
		s[i] = string(o)
	}
	return strings.Join(s, ", ")
}

func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "..."
}

// DecodeAnswer validates the LLM answer object. Rule errors have the prefix
// "rule." in their path.
func (a Allowlist) DecodeAnswer(raw []byte) (Answer, []FieldError) {
	v := &validator{a: a}
	msg, ok := v.checkDocument(raw)
	if !ok {
		return Answer{}, v.errs
	}
	m, ok := v.object("", msg, "status", "rule", "unmapped", "reason")
	if !ok {
		return Answer{}, v.errs
	}
	var ans Answer
	if json.Unmarshal(m["status"], &ans.Status) != nil || (ans.Status != "ok" && ans.Status != "unsupported") {
		v.fail("status", `must be "ok" or "unsupported"`)
		return Answer{}, v.errs
	}
	str := func(key string, max int) string {
		var s string
		if json.Unmarshal(m[key], &s) != nil {
			v.fail(key, "must be a string")
		} else if utf8.RuneCountInString(s) > max {
			v.fail(key, "must be at most %d characters", max)
		}
		return s
	}
	switch ans.Status {
	case "ok":
		if _, has := m["reason"]; has {
			v.fail("reason", "is not allowed when status is ok")
		}
		if rr, has := m["rule"]; !has {
			v.fail("rule", "is required when status is ok")
		} else {
			rv := &validator{a: a}
			g := rv.group("rule", rr, 1, true)
			if rv.conds > MaxConditions {
				rv.fail("rule", "the rule has more than %d conditions in total", MaxConditions)
			}
			v.errs = append(v.errs, rv.errs...)
			ans.Rule = Rule{root: g}
		}
		if ur, has := m["unmapped"]; has {
			var items []json.RawMessage
			if t := bytes.TrimSpace(ur); t[0] != '[' || json.Unmarshal(ur, &items) != nil {
				v.fail("unmapped", "must be a list of strings")
			} else {
				if len(items) > maxUnmapped {
					v.fail("unmapped", "must have at most %d items", maxUnmapped)
				}
				for i, it := range items {
					var s string
					p := fmt.Sprintf("unmapped[%d]", i)
					if json.Unmarshal(it, &s) != nil {
						v.fail(p, "must be a string")
					} else if utf8.RuneCountInString(s) > maxUnmappedRune {
						v.fail(p, "must be at most %d characters", maxUnmappedRune)
					} else {
						ans.Unmapped = append(ans.Unmapped, s)
					}
				}
			}
		}
	case "unsupported":
		for _, k := range []string{"rule", "unmapped"} {
			if _, has := m[k]; has {
				v.fail(k, "is not allowed when status is unsupported")
			}
		}
		if _, has := m["reason"]; !has {
			v.fail("reason", "is required when status is unsupported")
		} else {
			ans.Reason = str("reason", maxReasonRune)
			if strings.TrimSpace(ans.Reason) == "" && len(v.errs) == 0 {
				v.fail("reason", "must not be empty")
			}
		}
	}
	if len(v.errs) > 0 {
		return Answer{}, v.errs
	}
	return ans, nil
}
