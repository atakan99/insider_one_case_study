// Package seed builds fake users and shopping events. It has no database code.
package seed

import (
	"math/rand/v2"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

const (
	defaultUsers   = 5000
	eventWindow    = 180 * 24 * time.Hour
	maxPurchaseLag = 72 * time.Hour
)

// Options sets the generator input. Now must be set by the caller: Generate
// never reads the clock. A zero Now gives meaningless dates. If Users is 0 or
// less, Generate uses 5000.
type Options struct {
	Random uint64    // SEED_RANDOM, default 42
	Now    time.Time // seed time; all times are offsets from it
	Users  int       // number of random users, default 5000
}

// User is one row of the users table.
type User struct {
	ID                                int64
	FirstName, LastName, Email, Phone string
	Language, Country, City           string
	BirthDate, CreatedAt              time.Time
	EmailConsent, SMSConsent          bool
	LastCartAddAt, LastPurchaseAt     *time.Time
	OrderCount                        int
	TotalSpentKurus                   int64
}

// Event is one row of the events table.
type Event struct {
	ID, UserID  int64
	Type        string // "cart_add" or "purchase"
	AmountKurus *int64 // nil for cart_add
	OccurredAt  time.Time
}

type gen struct {
	f      *gofakeit.Faker
	now    time.Time
	users  []User
	events []Event
}

// Generate returns users and events. The output depends only on o.
func Generate(o Options) ([]User, []Event) {
	n := o.Users
	if n <= 0 {
		n = defaultUsers
	}
	g := &gen{f: gofakeit.NewFaker(rand.NewPCG(o.Random, o.Random), false), now: o.Now.UTC().Truncate(time.Second)}
	for i := 0; i < n; i++ {
		g.randomUser()
	}
	g.edgeCases()
	return g.users, g.events
}

func (g *gen) chance(p float64) bool { return g.f.Float64() < p }

// int64n returns a value in [0, n).
func (g *gen) int64n(n int64) int64 { return int64(g.f.IntN(int(n))) }

func (g *gen) pick(list []string) string { return list[g.f.IntN(len(list))] }

func (g *gen) randomUser() {
	u := User{ID: int64(len(g.users) + 1)}
	if g.chance(0.70) {
		u.Country, u.Language = "TR", "tr"
		u.FirstName, u.LastName = g.pick(trFirstNames), g.pick(trLastNames)
		u.Phone = "+905" + g.digits(9)
		if g.chance(0.40) {
			u.City = "Istanbul"
		} else {
			u.City = g.pick(trCities[1:])
		}
	} else {
		u.Country, u.Language = "US", "en"
		u.FirstName, u.LastName = g.f.FirstName(), g.f.LastName()
		u.Phone = "+1" + itoa(201+g.f.IntN(789)) + "55501" + g.digits(2)
		u.City = g.pick(usCities)
	}
	u.Email = asciiSlug(u.FirstName) + "." + asciiSlug(u.LastName) + "." + itoa(int(u.ID)) + "@example.com"
	u.EmailConsent = g.chance(0.70)
	u.SMSConsent = g.chance(0.50)
	u.BirthDate = g.birthDate()
	u.CreatedAt = g.now.Add(-time.Duration(g.int64n(int64(g.now.Sub(g.now.AddDate(-3, 0, 0))/time.Second)+1)) * time.Second)
	g.users = append(g.users, u)
	if g.chance(0.60) {
		g.randomEvents(&g.users[len(g.users)-1])
	}
}

func (g *gen) birthDate() time.Time {
	today := time.Date(g.now.Year(), g.now.Month(), g.now.Day(), 0, 0, 0, 0, time.UTC)
	years := 18 + g.f.IntN(58)
	return today.AddDate(-years, 0, -g.f.IntN(365))
}

func (g *gen) randomEvents(u *User) {
	start := u.CreatedAt
	if floor := g.now.Add(-eventWindow); floor.After(start) {
		start = floor
	}
	span := int64(g.now.Sub(start) / time.Second)
	for i, carts := 0, 1+g.f.IntN(8); i < carts; i++ {
		cart := start.Add(time.Duration(g.int64n(span+1)) * time.Second)
		g.addEvent(u, "cart_add", nil, cart)
		if g.chance(0.40) {
			maxLag := int64(maxPurchaseLag / time.Second)
			if left := int64(g.now.Sub(cart) / time.Second); left < maxLag {
				maxLag = left
			}
			amount := 5000 + g.int64n(495001)
			g.addEvent(u, "purchase", &amount, cart.Add(time.Duration(g.int64n(maxLag+1))*time.Second))
		}
	}
}

// addEvent appends an event and updates the summary fields of u.
func (g *gen) addEvent(u *User, typ string, amount *int64, at time.Time) {
	g.events = append(g.events, Event{ID: int64(len(g.events) + 1), UserID: u.ID, Type: typ, AmountKurus: amount, OccurredAt: at})
	t := at
	if typ == "cart_add" {
		if u.LastCartAddAt == nil || t.After(*u.LastCartAddAt) {
			u.LastCartAddAt = &t
		}
		return
	}
	u.OrderCount++
	u.TotalSpentKurus += *amount
	if u.LastPurchaseAt == nil || t.After(*u.LastPurchaseAt) {
		u.LastPurchaseAt = &t
	}
}

func (g *gen) edgeCases() {
	today := time.Date(g.now.Year(), g.now.Month(), g.now.Day(), 0, 0, 0, 0, time.UTC)
	created := g.now.AddDate(-1, 0, 0)
	add := func(slug, first, last, country, city string, phone string, email, sms bool) *User {
		lang := "en"
		if country == "TR" {
			lang = "tr"
		}
		g.users = append(g.users, User{
			ID: int64(len(g.users) + 1), FirstName: first, LastName: last,
			Email: "edge." + slug + "@example.com", Phone: phone,
			Language: lang, Country: country, City: city,
			BirthDate: today.AddDate(-35, 0, 0), CreatedAt: created,
			EmailConsent: email, SMSConsent: sms,
		})
		return &g.users[len(g.users)-1]
	}
	add("longname", "Bartholomew-Maximilian-Alexander-Jonathn", "Smith", "US", "New York", "+12125550100", true, true)
	add("turkish", "Çağrı", "Şükrüoğlu", "TR", "Istanbul", "+905550000001", true, true)
	add("noconsent", "Nora", "Nogood", "US", "Chicago", "+13125550101", false, false)
	add("smsonly", "Selim", "Sonmez", "TR", "Ankara", "+905550000002", false, true)
	add("emailonly", "Emily", "Eastwood", "US", "Seattle", "+12065550102", true, false)
	brief := add("examplebrief", "Deniz", "Yıldırım", "TR", "Istanbul", "+905550000003", true, true)
	g.addEvent(brief, "cart_add", nil, g.now.Add(-5*24*time.Hour))
}

func (g *gen) digits(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('0' + g.f.IntN(10))
	}
	return string(b)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for ; n > 0; n /= 10 {
		b = append([]byte{byte('0' + n%10)}, b...)
	}
	return string(b)
}

var asciiMap = strings.NewReplacer("ç", "c", "ğ", "g", "ı", "i", "İ", "i", "ö", "o", "ş", "s", "ü", "u",
	"Ç", "c", "Ğ", "g", "Ö", "o", "Ş", "s", "Ü", "u")

// asciiSlug lowercases s, maps Turkish letters to ASCII and drops other characters.
func asciiSlug(s string) string {
	s = strings.ToLower(asciiMap.Replace(s))
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
