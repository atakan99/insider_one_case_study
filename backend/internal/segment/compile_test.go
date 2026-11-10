package segment

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

func mustRule(t *testing.T, js string) Rule {
	t.Helper()
	r, errs := Default().ParseRule([]byte(js))
	if len(errs) > 0 {
		t.Fatalf("rule %s: %v", js, errs)
	}
	return r
}

func ageDate(n int) time.Time {
	return time.Date(2026-n, 11, 10, 0, 0, 0, 0, time.UTC)
}

func TestCompile_EachOperator(t *testing.T) {
	cut := testNow.AddDate(0, 0, -30)
	tests := []struct {
		name string
		rule string
		sql  string
		args []any
	}{
		{"city eq", cond("city", "eq", `"Istanbul"`), "(city = $1)", []any{"Istanbul"}},
		{"city neq", cond("city", "neq", `"Istanbul"`), "(city <> $1)", []any{"Istanbul"}},
		{"city in", cond("city", "in", `["Istanbul","Ankara"]`), "(city = ANY($1))", []any{[]string{"Istanbul", "Ankara"}}},
		{"city not_in", cond("city", "not_in", `["Istanbul","Ankara"]`), "(city <> ALL($1))", []any{[]string{"Istanbul", "Ankara"}}},
		{"country eq", cond("country", "eq", `"TR"`), "(country = $1)", []any{"TR"}},
		{"country neq", cond("country", "neq", `"TR"`), "(country <> $1)", []any{"TR"}},
		{"country in", cond("country", "in", `["TR","US"]`), "(country = ANY($1))", []any{[]string{"TR", "US"}}},
		{"country not_in", cond("country", "not_in", `["TR"]`), "(country <> ALL($1))", []any{[]string{"TR"}}},
		{"language eq", cond("language", "eq", `"tr"`), "(language = $1)", []any{"tr"}},
		{"language neq", cond("language", "neq", `"tr"`), "(language <> $1)", []any{"tr"}},
		{"language in", cond("language", "in", `["tr","en"]`), "(language = ANY($1))", []any{[]string{"tr", "en"}}},
		{"language not_in", cond("language", "not_in", `["en"]`), "(language <> ALL($1))", []any{[]string{"en"}}},
		{"order_count eq", cond("order_count", "eq", "3"), "(order_count = $1)", []any{int64(3)}},
		{"order_count gt", cond("order_count", "gt", "3"), "(order_count > $1)", []any{int64(3)}},
		{"order_count gte", cond("order_count", "gte", "3"), "(order_count >= $1)", []any{int64(3)}},
		{"order_count lt", cond("order_count", "lt", "3"), "(order_count < $1)", []any{int64(3)}},
		{"order_count lte", cond("order_count", "lte", "3"), "(order_count <= $1)", []any{int64(3)}},
		{"order_count between", cond("order_count", "between", "[2,5]"), "(order_count BETWEEN $1 AND $2)", []any{int64(2), int64(5)}},
		{"total_spent eq", cond("total_spent", "eq", "100.50"), "(total_spent = $1)", []any{"100.50"}},
		{"total_spent gt", cond("total_spent", "gt", "100"), "(total_spent > $1)", []any{"100"}},
		{"total_spent gte", cond("total_spent", "gte", "100"), "(total_spent >= $1)", []any{"100"}},
		{"total_spent lt", cond("total_spent", "lt", "100"), "(total_spent < $1)", []any{"100"}},
		{"total_spent lte", cond("total_spent", "lte", "100"), "(total_spent <= $1)", []any{"100"}},
		{"total_spent between", cond("total_spent", "between", "[10.5,99.99]"), "(total_spent BETWEEN $1 AND $2)", []any{"10.5", "99.99"}},
		{"created_at within", cond("created_at", "within_days", "30"), "(created_at >= $1)", []any{cut}},
		{"created_at older", cond("created_at", "older_than_days", "30"), "(created_at < $1)", []any{cut}},
		{"cart within", cond("last_cart_add_at", "within_days", "30"), "(last_cart_add_at >= $1)", []any{cut}},
		{"cart older", cond("last_cart_add_at", "older_than_days", "30"), "(last_cart_add_at < $1)", []any{cut}},
		{"cart is_null", cond("last_cart_add_at", "is_null", ""), "(last_cart_add_at IS NULL)", nil},
		{"cart is_not_null", cond("last_cart_add_at", "is_not_null", ""), "(last_cart_add_at IS NOT NULL)", nil},
		{"purchase within", cond("last_purchase_at", "within_days", "30"), "(last_purchase_at >= $1)", []any{cut}},
		{"purchase older", cond("last_purchase_at", "older_than_days", "30"), "(last_purchase_at < $1)", []any{cut}},
		{"purchase is_null", cond("last_purchase_at", "is_null", ""), "(last_purchase_at IS NULL)", nil},
		{"purchase is_not_null", cond("last_purchase_at", "is_not_null", ""), "(last_purchase_at IS NOT NULL)", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := Default().Compile(mustRule(t, root(tc.rule)), testNow)
			if w.SQL() != tc.sql {
				t.Errorf("SQL = %q, want %q", w.SQL(), tc.sql)
			}
			got := w.Args()
			if len(got) == 0 && len(tc.args) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.args) {
				t.Errorf("Args = %#v, want %#v", got, tc.args)
			}
		})
	}
}

func TestCompile_NoValueInSQL(t *testing.T) {
	rules := []string{
		cond("city", "in", `["Istanbul","Ankara"]`),
		cond("order_count", "between", "[7777,8888]"),
		cond("total_spent", "gte", "12345.67"),
		cond("country", "eq", `"TR"`),
		cond("age", "between", "[25,30]"),
		cond("last_cart_add_at", "within_days", "30"),
	}
	for _, js := range rules {
		w := Default().Compile(mustRule(t, root(js)), testNow)
		var flat []any
		for _, a := range w.Args() {
			switch v := a.(type) {
			case []string:
				for _, s := range v {
					flat = append(flat, s)
				}
			default:
				flat = append(flat, v)
			}
		}
		for _, a := range flat {
			var s string
			switch v := a.(type) {
			case string:
				s = v
			case int64:
				s = strconv.FormatInt(v, 10)
			default:
				continue
			}
			if len(s) >= 3 && strings.Contains(w.SQL(), s) {
				t.Errorf("value %q found in SQL %q", s, w.SQL())
			}
		}
		if strings.Contains(strings.ToLower(w.SQL()), "now()") {
			t.Errorf("SQL has now(): %q", w.SQL())
		}
	}
}

func TestCompile_RelativeTimes(t *testing.T) {
	want := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	w := Default().Compile(mustRule(t, root(cond("last_cart_add_at", "within_days", "30"))), testNow)
	if w.SQL() != "(last_cart_add_at >= $1)" || !w.Args()[0].(time.Time).Equal(want) {
		t.Errorf("within_days: %q %v", w.SQL(), w.Args())
	}
	w = Default().Compile(mustRule(t, root(cond("last_cart_add_at", "older_than_days", "30"))), testNow)
	if w.SQL() != "(last_cart_add_at < $1)" || !w.Args()[0].(time.Time).Equal(want) {
		t.Errorf("older_than_days: %q %v", w.SQL(), w.Args())
	}
}

func TestCompile_Age(t *testing.T) {
	tests := []struct {
		name string
		cond string
		sql  string
		args []any
	}{
		{"gte", cond("age", "gte", "30"), "(birth_date <= $1)", []any{ageDate(30)}},
		{"gt", cond("age", "gt", "30"), "(birth_date <= $1)", []any{ageDate(31)}},
		{"lte", cond("age", "lte", "30"), "(birth_date > $1)", []any{ageDate(31)}},
		{"lt", cond("age", "lt", "30"), "(birth_date > $1)", []any{ageDate(30)}},
		{"eq", cond("age", "eq", "30"), "((birth_date <= $1 AND birth_date > $2))", []any{ageDate(30), ageDate(31)}},
		{"between", cond("age", "between", "[25,30]"), "((birth_date <= $1 AND birth_date > $2))", []any{ageDate(25), ageDate(31)}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// A non-midnight clock must still give midnight UTC dates.
			w := Default().Compile(mustRule(t, root(tc.cond)), testNow)
			if w.SQL() != tc.sql {
				t.Errorf("SQL = %q, want %q", w.SQL(), tc.sql)
			}
			if !reflect.DeepEqual(w.Args(), tc.args) {
				t.Errorf("Args = %#v, want %#v", w.Args(), tc.args)
			}
		})
	}
}

func TestCompile_Groups(t *testing.T) {
	w := Default().Compile(mustRule(t, exampleRule), testNow)
	want := "(city = $1 AND last_cart_add_at >= $2 AND (last_purchase_at IS NULL OR last_purchase_at < $3))"
	if w.SQL() != want {
		t.Errorf("SQL = %q, want %q", w.SQL(), want)
	}
	cut := testNow.AddDate(0, 0, -30)
	if !reflect.DeepEqual(w.Args(), []any{"Istanbul", cut, cut}) {
		t.Errorf("Args = %#v", w.Args())
	}
	w = Default().Compile(mustRule(t, root()), testNow)
	if w.SQL() != "TRUE" || len(w.Args()) != 0 {
		t.Errorf("empty root: %q %v", w.SQL(), w.Args())
	}
	w = Default().Compile(Rule{}, testNow)
	if w.SQL() != "TRUE" {
		t.Errorf("zero rule: %q", w.SQL())
	}
}

func TestCompile_ParamNumberingDepthFirst(t *testing.T) {
	js := `{"op":"or","conditions":[
	  {"op":"and","conditions":[` + cond("city", "eq", `"Ankara"`) + `,` + cond("order_count", "between", "[1,2]") + `]},
	  ` + cond("country", "in", `["TR"]`) + `]}`
	w := Default().Compile(mustRule(t, js), testNow)
	want := "((city = $1 AND order_count BETWEEN $2 AND $3) OR country = ANY($4))"
	if w.SQL() != want {
		t.Errorf("SQL = %q, want %q", w.SQL(), want)
	}
}
