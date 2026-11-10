package segment

import (
	"strconv"
	"strings"
)

// opPhrases holds the fixed words for the comparison operators. The text
// goes after "<field phrase> is ".
var opPhrases = map[Operator]string{
	OpEq:      "",
	OpNeq:     "not ",
	OpIn:      "one of ",
	OpNotIn:   "none of ",
	OpGt:      "greater than ",
	OpGte:     "at least ",
	OpLt:      "less than ",
	OpLte:     "at most ",
	OpBetween: "between ",
}

// nullPhrases gives the words for is_null and is_not_null per field. A field
// that is not here gets a generic phrase.
var nullPhrases = map[string][2]string{
	"last_cart_add_at": {"no cart add", "has a cart add"},
	"last_purchase_at": {"no purchase", "has a purchase"},
}

// Describe makes an English description of a validated rule. The words are
// fixed, so the same rule always gives the same text.
func (a Allowlist) Describe(r Rule) string {
	g := r.Root()
	if len(g.Nodes) == 0 {
		return "all users"
	}
	return describeGroup(g)
}

func describeGroup(g Group) string {
	sep := " AND "
	if g.Op == GroupOr {
		sep = " OR "
	}
	parts := make([]string, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		switch n := n.(type) {
		case *Group:
			if len(n.Nodes) == 0 {
				continue
			}
			parts = append(parts, "("+describeGroup(*n)+")")
		case *Condition:
			parts = append(parts, describeCond(n))
		}
	}
	return strings.Join(parts, sep)
}

func describeCond(c *Condition) string {
	f := c.Field
	switch c.Op {
	case OpWithinDays:
		return f.Phrase + " in the last " + days(c.Values[0].Int)
	case OpOlderThanDays:
		return f.Phrase + " more than " + days(c.Values[0].Int) + " ago"
	case OpIsNull, OpIsNotNull:
		i := 0
		if c.Op == OpIsNotNull {
			i = 1
		}
		if p, ok := nullPhrases[f.Name]; ok {
			return p[i]
		}
		return f.Phrase + [2]string{" is not set", " is set"}[i]
	case OpBetween:
		return f.Phrase + " is between " + text(f, c.Values[0]) + " and " + text(f, c.Values[1])
	case OpIn, OpNotIn:
		items := make([]string, len(c.Values))
		for i, v := range c.Values {
			items[i] = text(f, v)
		}
		return f.Phrase + " is " + opPhrases[c.Op] + strings.Join(items, ", ")
	}
	return f.Phrase + " is " + opPhrases[c.Op] + text(f, c.Values[0])
}

func days(n int64) string {
	if n == 1 {
		return "1 day"
	}
	return strconv.FormatInt(n, 10) + " days"
}

func text(f *FieldSpec, v Value) string {
	switch f.Kind {
	case KindEnum:
		return v.Str
	case KindDecimal:
		return v.Dec
	default:
		return strconv.FormatInt(v.Int, 10)
	}
}
