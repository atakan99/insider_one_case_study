// Package segment holds the segment rule types, the field allowlist and the
// validator. It imports nothing from other internal packages.
package segment

// Kind is the value type of a field.
type Kind int

const (
	KindEnum    Kind = iota // string from a fixed list
	KindInt                 // integer in [Min, Max]
	KindDecimal             // number in [Min, Max] with at most 2 decimals
	KindDays                // integer number of days in [Min, Max]
	KindAge                 // integer age in years in [Min, Max]
)

// Operator is a comparison operator in a condition.
type Operator string

const (
	OpEq            Operator = "eq"
	OpNeq           Operator = "neq"
	OpIn            Operator = "in"
	OpNotIn         Operator = "not_in"
	OpGt            Operator = "gt"
	OpGte           Operator = "gte"
	OpLt            Operator = "lt"
	OpLte           Operator = "lte"
	OpBetween       Operator = "between"
	OpWithinDays    Operator = "within_days"
	OpOlderThanDays Operator = "older_than_days"
	OpIsNull        Operator = "is_null"
	OpIsNotNull     Operator = "is_not_null"
)

// FieldSpec describes one field that a rule may use.
type FieldSpec struct {
	Name     string // JSON field name
	Column   string // SQL column; "birth_date" for age
	Kind     Kind
	Ops      []Operator
	Enum     []string // KindEnum only
	Min, Max float64  // numeric kinds
	Phrase   string   // used by Describe
}

// Allowlist is the one table of fields and operators that a rule may use.
type Allowlist struct{ Fields []FieldSpec }

// Field returns the spec with the given name.
func (a Allowlist) Field(name string) (*FieldSpec, bool) {
	for i := range a.Fields {
		if a.Fields[i].Name == name {
			return &a.Fields[i], true
		}
	}
	return nil, false
}

// Allows reports whether the field accepts the operator.
func (f *FieldSpec) Allows(op Operator) bool {
	for _, o := range f.Ops {
		if o == op {
			return true
		}
	}
	return false
}

// Clone returns a deep copy.
func (a Allowlist) Clone() Allowlist {
	out := Allowlist{Fields: make([]FieldSpec, len(a.Fields))}
	for i, f := range a.Fields {
		f.Ops = append([]Operator(nil), f.Ops...)
		f.Enum = append([]string(nil), f.Enum...)
		out.Fields[i] = f
	}
	return out
}

var (
	opsEnum  = []Operator{OpEq, OpNeq, OpIn, OpNotIn}
	opsNum   = []Operator{OpEq, OpGt, OpGte, OpLt, OpLte, OpBetween}
	opsTime  = []Operator{OpWithinDays, OpOlderThanDays}
	opsTimeN = []Operator{OpWithinDays, OpOlderThanDays, OpIsNull, OpIsNotNull}
)

// Default returns the allowlist from the slice 04 spec. The city list is
// kept in step with the seed package by a test.
func Default() Allowlist {
	cities := []string{
		"Istanbul", "Ankara", "Izmir", "Bursa", "Antalya", "Adana", "Konya",
		"New York", "Los Angeles", "Chicago", "Houston", "Miami", "Seattle", "Boston",
	}
	return Allowlist{Fields: []FieldSpec{
		{Name: "city", Column: "city", Kind: KindEnum, Ops: opsEnum, Enum: cities, Phrase: "city"},
		{Name: "country", Column: "country", Kind: KindEnum, Ops: opsEnum, Enum: []string{"TR", "US"}, Phrase: "country"},
		{Name: "language", Column: "language", Kind: KindEnum, Ops: opsEnum, Enum: []string{"tr", "en"}, Phrase: "language"},
		{Name: "age", Column: "birth_date", Kind: KindAge, Ops: opsNum, Min: 18, Max: 120, Phrase: "age"},
		{Name: "created_at", Column: "created_at", Kind: KindDays, Ops: opsTime, Min: 1, Max: 365, Phrase: "sign-up"},
		{Name: "last_cart_add_at", Column: "last_cart_add_at", Kind: KindDays, Ops: opsTimeN, Min: 1, Max: 365, Phrase: "last cart add"},
		{Name: "last_purchase_at", Column: "last_purchase_at", Kind: KindDays, Ops: opsTimeN, Min: 1, Max: 365, Phrase: "last purchase"},
		{Name: "order_count", Column: "order_count", Kind: KindInt, Ops: opsNum, Min: 0, Max: 100000, Phrase: "order count"},
		{Name: "total_spent", Column: "total_spent", Kind: KindDecimal, Ops: opsNum, Min: 0, Max: 10000000, Phrase: "total spent"},
	}}
}
