package segment

import (
	"strconv"
	"strings"
	"time"
)

// Where is a SQL condition with its parameters. Only Compile makes one.
type Where struct {
	sql  string
	args []any
}

// SQL returns the condition text. It has $n placeholders and no values.
func (w Where) SQL() string { return w.sql }

// Args returns the parameters in $n order.
func (w Where) Args() []any { return w.args }

type compiler struct {
	now  time.Time
	args []any
}

// add stores a parameter and returns its placeholder.
func (c *compiler) add(v any) string {
	c.args = append(c.args, v)
	return "$" + strconv.Itoa(len(c.args))
}

// Compile makes a parameterized WHERE condition from a validated rule. Column
// names come from the allowlist. The clock gives the relative times, so the
// SQL never calls now().
func (a Allowlist) Compile(r Rule, now time.Time) Where {
	c := &compiler{now: now.UTC()}
	sql := c.group(r.Root())
	return Where{sql: sql, args: c.args}
}

func (c *compiler) group(g Group) string {
	if len(g.Nodes) == 0 {
		return "TRUE"
	}
	parts := make([]string, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		switch n := n.(type) {
		case *Group:
			parts = append(parts, c.group(*n))
		case *Condition:
			parts = append(parts, c.condition(n))
		}
	}
	sep := " AND "
	if g.Op == GroupOr {
		sep = " OR "
	}
	return "(" + strings.Join(parts, sep) + ")"
}

var cmpSQL = map[Operator]string{
	OpEq: "=", OpNeq: "<>", OpGt: ">", OpGte: ">=", OpLt: "<", OpLte: "<=",
}

func (c *compiler) condition(cd *Condition) string {
	col := cd.Field.Column
	switch cd.Op {
	case OpIsNull:
		return col + " IS NULL"
	case OpIsNotNull:
		return col + " IS NOT NULL"
	case OpWithinDays:
		return col + " >= " + c.add(c.daysAgo(cd.Values[0].Int))
	case OpOlderThanDays:
		return col + " < " + c.add(c.daysAgo(cd.Values[0].Int))
	}
	if cd.Field.Kind == KindAge {
		return c.age(col, cd)
	}
	switch cd.Op {
	case OpIn:
		return col + " = ANY(" + c.add(c.strings(cd.Values)) + ")"
	case OpNotIn:
		return col + " <> ALL(" + c.add(c.strings(cd.Values)) + ")"
	case OpBetween:
		lo := c.add(c.scalar(cd.Field, cd.Values[0]))
		hi := c.add(c.scalar(cd.Field, cd.Values[1]))
		return col + " BETWEEN " + lo + " AND " + hi
	}
	return col + " " + cmpSQL[cd.Op] + " " + c.add(c.scalar(cd.Field, cd.Values[0]))
}

func (c *compiler) daysAgo(n int64) time.Time { return c.now.AddDate(0, 0, -int(n)) }

// yearsAgo is today (UTC) minus n years, at 00:00 UTC.
func (c *compiler) yearsAgo(n int64) time.Time {
	y, m, d := c.now.Date()
	return time.Date(y-int(n), m, d, 0, 0, 0, 0, time.UTC)
}

// age turns an age test into a test on the birth date. A person who is at
// least n years old was born on or before yearsAgo(n). A person who is at
// most n years old was born after yearsAgo(n+1).
func (c *compiler) age(col string, cd *Condition) string {
	v := cd.Values[0].Int
	switch cd.Op {
	case OpGte:
		return col + " <= " + c.add(c.yearsAgo(v))
	case OpGt:
		return col + " <= " + c.add(c.yearsAgo(v+1))
	case OpLte:
		return col + " > " + c.add(c.yearsAgo(v+1))
	case OpLt:
		return col + " > " + c.add(c.yearsAgo(v))
	case OpEq:
		return c.ageRange(col, v, v)
	default: // OpBetween
		return c.ageRange(col, v, cd.Values[1].Int)
	}
}

func (c *compiler) ageRange(col string, lo, hi int64) string {
	a := c.add(c.yearsAgo(lo))
	b := c.add(c.yearsAgo(hi + 1))
	return "(" + col + " <= " + a + " AND " + col + " > " + b + ")"
}

func (c *compiler) strings(vs []Value) []string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = v.Str
	}
	return out
}

// scalar returns the parameter for one value. Decimals go as text, and
// Postgres reads them as numeric.
func (c *compiler) scalar(f *FieldSpec, v Value) any {
	switch f.Kind {
	case KindEnum:
		return v.Str
	case KindDecimal:
		return v.Dec
	default:
		return v.Int
	}
}
