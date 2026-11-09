package seed

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

var testNow = time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)

func testOptions() Options {
	return Options{Random: 42, Now: testNow, Users: 5000}
}

var (
	wantTRCities = []string{"Istanbul", "Ankara", "Izmir", "Bursa", "Antalya", "Adana", "Konya"}
	wantUSCities = []string{"New York", "Los Angeles", "Chicago", "Houston", "Miami", "Seattle", "Boston"}
	trPhone      = regexp.MustCompile(`^\+905\d{9}$`)
	usPhone      = regexp.MustCompile(`^\+1\d{3}55501\d{2}$`)
)

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func ageAt(birth, now time.Time) int {
	age := now.Year() - birth.Year()
	if now.Month() < birth.Month() || (now.Month() == birth.Month() && now.Day() < birth.Day()) {
		age--
	}
	return age
}

func isWholeSecondUTC(t time.Time) bool {
	return t.Location() == time.UTC && t.Nanosecond() == 0
}

func TestGenerate_Deterministic(t *testing.T) {
	o := testOptions()
	o.Users = 300
	u1, e1 := Generate(o)
	u2, e2 := Generate(o)
	if !reflect.DeepEqual(u1, u2) || !reflect.DeepEqual(e1, e2) {
		t.Fatal("same options gave different output")
	}
	o.Random = 43
	u3, _ := Generate(o)
	if reflect.DeepEqual(u1, u3) {
		t.Fatal("different Random gave equal users")
	}
}

func TestGenerate_DeterministicZeroSeed(t *testing.T) {
	o := testOptions()
	o.Random = 0
	o.Users = 10
	u1, e1 := Generate(o)
	u2, e2 := Generate(o)
	if !reflect.DeepEqual(u1, u2) || !reflect.DeepEqual(e1, e2) {
		t.Fatal("Random 0 gave different output")
	}
}

func TestGenerate_DefaultUsers(t *testing.T) {
	o := testOptions()
	o.Users = 0
	users, _ := Generate(o)
	const edgeCaseUsers = 6
	if len(users) != 5000+edgeCaseUsers {
		t.Fatalf("got %d users, want %d", len(users), 5000+edgeCaseUsers)
	}
}

func TestGenerate_Fields(t *testing.T) {
	users, events := Generate(testOptions())
	emails := map[string]bool{}
	for i, u := range users {
		if u.ID != int64(i+1) {
			t.Fatalf("user %d has ID %d", i, u.ID)
		}
		if emails[u.Email] {
			t.Fatalf("duplicate email %s", u.Email)
		}
		emails[u.Email] = true
		if !strings.HasSuffix(u.Email, "@example.com") {
			t.Fatalf("bad email %s", u.Email)
		}
		if u.FirstName == "" || u.LastName == "" {
			t.Fatalf("user %d has empty name", u.ID)
		}
		switch u.Country {
		case "TR":
			if u.Language != "tr" || !trPhone.MatchString(u.Phone) || !contains(wantTRCities, u.City) {
				t.Fatalf("bad TR user %+v", u)
			}
		case "US":
			if u.Language != "en" || !usPhone.MatchString(u.Phone) || !contains(wantUSCities, u.City) {
				t.Fatalf("bad US user %+v", u)
			}
		default:
			t.Fatalf("bad country %q", u.Country)
		}
		age := ageAt(u.BirthDate, testNow)
		if age < 18 || age > 75 {
			t.Fatalf("user %d age %d", u.ID, age)
		}
		if u.BirthDate.Hour() != 0 || !isWholeSecondUTC(u.BirthDate) {
			t.Fatalf("birth date not UTC midnight: %v", u.BirthDate)
		}
		if !isWholeSecondUTC(u.CreatedAt) || u.CreatedAt.After(testNow) || u.CreatedAt.Before(testNow.AddDate(-3, 0, 0)) {
			t.Fatalf("bad created_at %v", u.CreatedAt)
		}
	}
	for i, e := range events {
		if e.ID != int64(i+1) {
			t.Fatalf("event %d has ID %d", i, e.ID)
		}
		if !isWholeSecondUTC(e.OccurredAt) {
			t.Fatalf("event time not whole-second UTC: %v", e.OccurredAt)
		}
		switch e.Type {
		case "purchase":
			if e.AmountKurus == nil || *e.AmountKurus < 5000 || *e.AmountKurus > 500000 {
				t.Fatalf("bad purchase amount %v", e.AmountKurus)
			}
		case "cart_add":
			if e.AmountKurus != nil {
				t.Fatal("cart_add has amount")
			}
		default:
			t.Fatalf("bad event type %q", e.Type)
		}
	}
}

func TestGenerate_Ratios(t *testing.T) {
	users, _ := Generate(testOptions())
	n := float64(len(users))
	var tr, email, sms, withEvents, istanbul, trTotal float64
	for _, u := range users {
		if u.Country == "TR" {
			tr++
			trTotal++
			if u.City == "Istanbul" {
				istanbul++
			}
		}
		if u.EmailConsent {
			email++
		}
		if u.SMSConsent {
			sms++
		}
		if u.LastCartAddAt != nil || u.LastPurchaseAt != nil {
			withEvents++
		}
	}
	check := func(name string, got, want float64) {
		if d := got - want; d > 0.03 || d < -0.03 {
			t.Errorf("%s ratio %.3f, want %.2f", name, got, want)
		}
	}
	check("TR", tr/n, 0.70)
	check("email consent", email/n, 0.70)
	check("sms consent", sms/n, 0.50)
	check("with events", withEvents/n, 0.60)
	check("istanbul in TR", istanbul/trTotal, 0.40)
}

func TestGenerate_Summaries(t *testing.T) {
	users, events := Generate(testOptions())
	byUser := map[int64][]Event{}
	for _, e := range events {
		byUser[e.UserID] = append(byUser[e.UserID], e)
	}
	for _, u := range users {
		var lastCart, lastBuy *time.Time
		orders := 0
		var total int64
		for _, e := range byUser[u.ID] {
			at := e.OccurredAt
			if e.Type == "cart_add" && (lastCart == nil || at.After(*lastCart)) {
				lastCart = &at
			}
			if e.Type == "purchase" {
				orders++
				total += *e.AmountKurus
				if lastBuy == nil || at.After(*lastBuy) {
					lastBuy = &at
				}
			}
		}
		if !reflect.DeepEqual(lastCart, u.LastCartAddAt) || !reflect.DeepEqual(lastBuy, u.LastPurchaseAt) ||
			orders != u.OrderCount || total != u.TotalSpentKurus {
			t.Fatalf("summary mismatch for user %d: %+v", u.ID, u)
		}
	}
}

func TestGenerate_EventTimes(t *testing.T) {
	users, events := Generate(testOptions())
	created := map[int64]time.Time{}
	for _, u := range users {
		created[u.ID] = u.CreatedAt
	}
	floor := testNow.Add(-180 * 24 * time.Hour)
	for _, e := range events {
		lo := created[e.UserID]
		if floor.After(lo) {
			lo = floor
		}
		if e.OccurredAt.Before(lo) || e.OccurredAt.After(testNow) {
			t.Fatalf("event %d at %v outside [%v, %v]", e.ID, e.OccurredAt, lo, testNow)
		}
	}
}

func byEmail(users []User, email string) *User {
	for i := range users {
		if users[i].Email == email {
			return &users[i]
		}
	}
	return nil
}

func TestGenerate_EdgeCases(t *testing.T) {
	users, _ := Generate(testOptions())
	long := byEmail(users, "edge.longname@example.com")
	if long == nil || utf8.RuneCountInString(long.FirstName) != 40 || long.Country != "US" || long.Language != "en" {
		t.Fatalf("bad longname user %+v", long)
	}
	tu := byEmail(users, "edge.turkish@example.com")
	if tu == nil || tu.FirstName != "Çağrı" || tu.LastName != "Şükrüoğlu" || tu.Country != "TR" || tu.Language != "tr" {
		t.Fatalf("bad turkish user %+v", tu)
	}
	none := byEmail(users, "edge.noconsent@example.com")
	if none == nil || none.EmailConsent || none.SMSConsent {
		t.Fatalf("bad noconsent user %+v", none)
	}
	sms := byEmail(users, "edge.smsonly@example.com")
	if sms == nil || sms.EmailConsent || !sms.SMSConsent {
		t.Fatalf("bad smsonly user %+v", sms)
	}
	em := byEmail(users, "edge.emailonly@example.com")
	if em == nil || !em.EmailConsent || em.SMSConsent {
		t.Fatalf("bad emailonly user %+v", em)
	}
	if byEmail(users, "edge.examplebrief@example.com") == nil {
		t.Fatal("missing examplebrief user")
	}
	if len(users) != 5006 || users[len(users)-1].ID != 5006 {
		t.Fatalf("got %d users, want 5006", len(users))
	}
}

func TestGenerate_ExampleBrief(t *testing.T) {
	users, _ := Generate(testOptions())
	since := testNow.Add(-30 * 24 * time.Hour)
	for _, u := range users {
		if u.Language == "tr" && u.City == "Istanbul" &&
			u.LastCartAddAt != nil && !u.LastCartAddAt.Before(since) &&
			(u.LastPurchaseAt == nil || u.LastPurchaseAt.Before(since)) {
			return
		}
	}
	t.Fatal("no user matches the example brief rule")
}

func TestGenerate_ExampleBriefUserFixed(t *testing.T) {
	users, events := Generate(testOptions())
	u := byEmail(users, "edge.examplebrief@example.com")
	if u == nil || u.Language != "tr" || u.City != "Istanbul" || !u.EmailConsent || !u.SMSConsent || u.OrderCount != 0 {
		t.Fatalf("bad examplebrief user %+v", u)
	}
	want := testNow.Add(-5 * 24 * time.Hour)
	if u.LastCartAddAt == nil || !u.LastCartAddAt.Equal(want) {
		t.Fatalf("cart add %v, want %v", u.LastCartAddAt, want)
	}
	n := 0
	for _, e := range events {
		if e.UserID == u.ID {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("got %d events, want 1", n)
	}
}

func TestEmailSlug(t *testing.T) {
	if got := asciiSlug("Çağrı İŞÖü O'Neil-Smith"); got != "cagriisouoneilsmith" {
		t.Fatalf("got %q", got)
	}
}
