package segment

import (
	"reflect"
	"regexp"
	"sort"
	"testing"
	"unicode/utf8"

	"campaigncomposer/internal/seed"
)

// goSample computes the expected sample from the seed users.
func goSample(users []seed.User, match func(seed.User) bool, n int) map[string][]string {
	type entry struct {
		name string
		id   int64
	}
	byLang := map[string]map[string]int64{}
	for _, u := range users {
		if !match(u) || !(u.EmailConsent || u.SMSConsent) {
			continue
		}
		m := byLang[u.Language]
		if m == nil {
			m = map[string]int64{}
			byLang[u.Language] = m
		}
		if id, ok := m[u.FirstName]; !ok || u.ID < id {
			m[u.FirstName] = u.ID
		}
	}
	out := map[string][]string{}
	for lang, m := range byLang {
		var es []entry
		for name, id := range m {
			es = append(es, entry{name, id})
		}
		longest := es[0].name
		for _, e := range es {
			lr, cr := utf8.RuneCountInString(e.name), utf8.RuneCountInString(longest)
			if lr > cr || (lr == cr && e.name < longest) {
				longest = e.name
			}
		}
		sort.Slice(es, func(i, j int) bool { return es[i].id < es[j].id })
		names := []string{longest}
		for _, e := range es {
			if len(names) == n {
				break
			}
			if e.name != longest {
				names = append(names, e.name)
			}
		}
		out[lang] = names
	}
	return out
}

func exampleMatch(u seed.User) bool {
	return u.City == "Istanbul" && within(u.LastCartAddAt, 30) &&
		(u.LastPurchaseAt == nil || olderThan(u.LastPurchaseAt, 30))
}

func TestSampleFirstNames(t *testing.T) {
	ctx, pool, users := seedDB(t)
	w := compile(t, exampleRule)
	got, err := SampleFirstNames(ctx, pool, w, []string{"tr", "en"}, 5)
	if err != nil {
		t.Fatal(err)
	}
	want := goSample(users, exampleMatch, 5)
	if len(want) == 0 {
		t.Fatal("test data has no matching user")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SampleFirstNames = %v, want %v", got, want)
	}
}

func TestSampleFirstNames_LongestFirst(t *testing.T) {
	ctx, pool, _ := seedDB(t)
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, first_name,last_name,email,phone,language,country,city,birth_date,email_consent,sms_consent,created_at,order_count,total_spent)
VALUES (1000001, 'Abdurrahmangazi Şükrü','K','long@example.com','+905000000001','tr','TR','Ankara','1990-01-01',false,true,now(),0,0)`); err != nil {
		t.Fatal(err)
	}
	w := compile(t, `{"op":"and","conditions":[{"field":"city","op":"eq","value":"Ankara"}]}`)
	got, err := SampleFirstNames(ctx, pool, w, []string{"tr"}, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got["tr"]) == 0 || got["tr"][0] != "Abdurrahmangazi Şükrü" {
		t.Fatalf("first tr name = %v, want the long name first", got["tr"])
	}
	if len(got["tr"]) > 5 {
		t.Fatalf("got %d names, want at most 5", len(got["tr"]))
	}
	seen := map[string]bool{}
	for _, n := range got["tr"] {
		if seen[n] {
			t.Fatalf("duplicate name %q in %v", n, got["tr"])
		}
		seen[n] = true
	}
}

func TestSampleFirstNames_NoConsentExcluded(t *testing.T) {
	ctx, pool, _ := seedDB(t)
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, first_name,last_name,email,phone,language,country,city,birth_date,email_consent,sms_consent,created_at,order_count,total_spent)
VALUES (1000002, 'Nobodywantsmessagesatall','K','none@example.com','+905000000002','tr','TR','Ankara','1990-01-01',false,false,now(),0,0)`); err != nil {
		t.Fatal(err)
	}
	w := compile(t, `{"op":"and","conditions":[{"field":"city","op":"eq","value":"Ankara"}]}`)
	got, err := SampleFirstNames(ctx, pool, w, []string{"tr", "en"}, 5)
	if err != nil {
		t.Fatal(err)
	}
	for lang, names := range got {
		for _, n := range names {
			if n == "Nobodywantsmessagesatall" {
				t.Fatalf("%s: user with no consent is in the sample", lang)
			}
		}
	}
}

func TestSampleFirstNames_Stable(t *testing.T) {
	ctx, pool, _ := seedDB(t)
	w := compile(t, exampleRule)
	a, err := SampleFirstNames(ctx, pool, w, []string{"tr", "en"}, 5)
	if err != nil {
		t.Fatal(err)
	}
	b, err := SampleFirstNames(ctx, pool, w, []string{"tr", "en"}, 5)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("two calls differ: %v and %v", a, b)
	}
}

func TestSampleQuery_Columns(t *testing.T) {
	q := sampleQuery(compile(t, `{"op":"and","conditions":[{"field":"city","op":"eq","value":"Istanbul"}]}`), 2)
	for _, col := range []string{"last_name", "email", "phone", "birth_date", "total_spent"} {
		if regexp.MustCompile(`\b` + col + `\b`).MatchString(q) {
			t.Errorf("sample query reads %s:\n%s", col, q)
		}
	}
}

func TestSampleFirstNames_EmptyWhere(t *testing.T) {
	if _, err := SampleFirstNames(t.Context(), nil, Where{}, []string{"tr"}, 5); err != ErrEmptyWhere {
		t.Fatalf("err = %v, want ErrEmptyWhere", err)
	}
}
