package segment

import (
	"encoding/json"
)

// Limits from the spec.
const (
	MaxRuleBytes  = 16 * 1024
	MaxDepth      = 3 // levels of groups; the root is level 1
	MaxConditions = 20
	MaxListValues = 20
)

// GroupOp is "and" or "or".
type GroupOp string

const (
	GroupAnd GroupOp = "and"
	GroupOr  GroupOp = "or"
)

// Node is a *Group or a *Condition.
type Node interface{ isNode() }

// Group joins child nodes with And or Or.
type Group struct {
	Op    GroupOp
	Nodes []Node
}

// Value is one validated value. Which member is set depends on the field
// kind: Str for KindEnum, Int for KindInt, KindDays and KindAge, and Dec
// (the number text, at most 2 decimals) for KindDecimal.
type Value struct {
	Str string
	Int int64
	Dec string
}

// Condition is a validated test on one field. Values has no entry for
// is_null and is_not_null, two entries for between, one or more for in and
// not_in, and one entry otherwise.
type Condition struct {
	Field  *FieldSpec
	Op     Operator
	Values []Value
}

func (*Group) isNode()     {}
func (*Condition) isNode() {}

// Rule is a validated segment rule. Only ParseRule makes a non-zero Rule.
// The zero Rule is an empty "and" group, which means all users.
type Rule struct{ root Group }

// Root returns the root group.
func (r Rule) Root() Group {
	g := r.root
	if g.Op == "" {
		g.Op = GroupAnd
	}
	return g
}

// FieldError is one validation error.
type FieldError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

func (e FieldError) Error() string {
	if e.Path == "" {
		return e.Message
	}
	return e.Path + ": " + e.Message
}

// MarshalJSON writes the rule in the same shape as the input.
func (r Rule) MarshalJSON() ([]byte, error) { return json.Marshal(marshalGroup(r.Root())) }

type jsonGroup struct {
	Op         GroupOp `json:"op"`
	Conditions []any   `json:"conditions"`
}

type jsonCond struct {
	Field string          `json:"field"`
	Op    Operator        `json:"op"`
	Value json.RawMessage `json:"value,omitempty"`
}

func marshalGroup(g Group) jsonGroup {
	out := jsonGroup{Op: g.Op, Conditions: make([]any, 0, len(g.Nodes))}
	for _, n := range g.Nodes {
		switch n := n.(type) {
		case *Group:
			out.Conditions = append(out.Conditions, marshalGroup(*n))
		case *Condition:
			out.Conditions = append(out.Conditions, marshalCond(n))
		}
	}
	return out
}

func marshalCond(c *Condition) jsonCond {
	out := jsonCond{Field: c.Field.Name, Op: c.Op}
	one := func(v Value) any {
		switch c.Field.Kind {
		case KindEnum:
			return v.Str
		case KindDecimal:
			return json.Number(v.Dec)
		default:
			return v.Int
		}
	}
	var val any
	switch c.Op {
	case OpIsNull, OpIsNotNull:
		return out
	case OpIn, OpNotIn, OpBetween:
		list := make([]any, len(c.Values))
		for i, v := range c.Values {
			list[i] = one(v)
		}
		val = list
	default:
		val = one(c.Values[0])
	}
	out.Value, _ = json.Marshal(val)
	return out
}
