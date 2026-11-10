package segment

import "testing"

func TestDescribe_Example(t *testing.T) {
	got := Default().Describe(mustRule(t, exampleRule))
	want := "city is Istanbul AND last cart add in the last 30 days AND (no purchase OR last purchase more than 30 days ago)"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
	if again := Default().Describe(mustRule(t, exampleRule)); again != got {
		t.Errorf("not deterministic: %q vs %q", again, got)
	}
}

func TestDescribe_EachFieldAndOperator(t *testing.T) {
	tests := []struct{ cond, want string }{
		{cond("city", "eq", `"Istanbul"`), "city is Istanbul"},
		{cond("city", "neq", `"Istanbul"`), "city is not Istanbul"},
		{cond("city", "in", `["Istanbul","Ankara"]`), "city is one of Istanbul, Ankara"},
		{cond("city", "not_in", `["Istanbul","Ankara"]`), "city is none of Istanbul, Ankara"},
		{cond("country", "eq", `"TR"`), "country is TR"},
		{cond("country", "in", `["TR","US"]`), "country is one of TR, US"},
		{cond("language", "neq", `"tr"`), "language is not tr"},
		{cond("language", "not_in", `["tr"]`), "language is none of tr"},
		{cond("age", "eq", "30"), "age is 30"},
		{cond("age", "gt", "30"), "age is greater than 30"},
		{cond("age", "gte", "30"), "age is at least 30"},
		{cond("age", "lt", "30"), "age is less than 30"},
		{cond("age", "lte", "30"), "age is at most 30"},
		{cond("age", "between", "[25,30]"), "age is between 25 and 30"},
		{cond("created_at", "within_days", "7"), "sign-up in the last 7 days"},
		{cond("created_at", "older_than_days", "1"), "sign-up more than 1 day ago"},
		{cond("last_cart_add_at", "within_days", "1"), "last cart add in the last 1 day"},
		{cond("last_cart_add_at", "older_than_days", "30"), "last cart add more than 30 days ago"},
		{cond("last_cart_add_at", "is_null", ""), "no cart add"},
		{cond("last_cart_add_at", "is_not_null", ""), "has a cart add"},
		{cond("last_purchase_at", "within_days", "30"), "last purchase in the last 30 days"},
		{cond("last_purchase_at", "older_than_days", "30"), "last purchase more than 30 days ago"},
		{cond("last_purchase_at", "is_null", ""), "no purchase"},
		{cond("last_purchase_at", "is_not_null", ""), "has a purchase"},
		{cond("order_count", "eq", "3"), "order count is 3"},
		{cond("order_count", "gt", "3"), "order count is greater than 3"},
		{cond("order_count", "between", "[2,5]"), "order count is between 2 and 5"},
		{cond("total_spent", "eq", "100.50"), "total spent is 100.50"},
		{cond("total_spent", "gte", "100"), "total spent is at least 100"},
		{cond("total_spent", "lte", "100"), "total spent is at most 100"},
		{cond("total_spent", "lt", "100"), "total spent is less than 100"},
		{cond("total_spent", "between", "[10.5,99.99]"), "total spent is between 10.5 and 99.99"},
	}
	for _, tc := range tests {
		got := Default().Describe(mustRule(t, root(tc.cond)))
		if got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.cond, got, tc.want)
		}
	}
}

func TestDescribe_Empty(t *testing.T) {
	if got := Default().Describe(Rule{}); got != "all users" {
		t.Errorf("got %q", got)
	}
}

func TestDescribe_CoversAllowlist(t *testing.T) {
	// Every operator the allowlist permits must have a phrase.
	for _, f := range Default().Fields {
		for _, op := range f.Ops {
			if _, ok := opPhrases[op]; !ok && op != OpWithinDays && op != OpOlderThanDays && op != OpIsNull && op != OpIsNotNull {
				t.Errorf("%s %s has no phrase", f.Name, op)
			}
		}
	}
}
