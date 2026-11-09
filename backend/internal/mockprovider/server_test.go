package mockprovider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeClock is a clock that the test moves by hand.
type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{t: time.Date(2026, 11, 9, 10, 0, 0, 0, time.FixedZone("TRT", 3*3600))}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

// normalOptions has no random failures, no latency and a rate limit that
// the tests do not reach.
func normalOptions(clock *fakeClock) Options {
	return Options{
		RatePerSec: 1e6,
		Burst:      1e6,
		Random:     1,
		Now:        clock.Now,
		Sleep:      func(time.Duration) {},
	}
}

func startServer(t *testing.T, o Options) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(New(o))
	t.Cleanup(srv.Close)
	return srv
}

func post(t *testing.T, base, path, body string, header map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, base+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func decode[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	var v T
	b, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("decode %q: %v", b, err)
	}
	return v
}

type acceptBody struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type listBody struct {
	Total    int       `json:"total"`
	Messages []Message `json:"messages"`
}

func get(t *testing.T, base, path string) *http.Response {
	t.Helper()
	resp, err := http.Get(base + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func listMessages(t *testing.T, base, query string) listBody {
	t.Helper()
	resp := get(t, base, "/v1/messages"+query)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /v1/messages%s: status %d", query, resp.StatusCode)
	}
	return decode[listBody](t, resp)
}

func stats(t *testing.T, base string) map[string]ChannelStats {
	t.Helper()
	return decode[map[string]ChannelStats](t, get(t, base, "/v1/stats"))
}

func sms(to, body, ref string) string {
	b, _ := json.Marshal(map[string]string{"to": to, "body": body, "reference": ref})
	return string(b)
}

func email(to, subject, body, ref string) string {
	b, _ := json.Marshal(map[string]string{"to": to, "subject": subject, "body": body, "reference": ref})
	return string(b)
}

func wantStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want %d, body %s", resp.StatusCode, want, b)
	}
}

func TestSend_SMSAccepted(t *testing.T) {
	clock := newFakeClock()
	srv := startServer(t, normalOptions(clock))
	resp := post(t, srv.URL, "/v1/sms", sms("+905321234567", "Merhaba Ayşe", "c1:u1:sms"), nil)
	wantStatus(t, resp, http.StatusAccepted)
	got := decode[acceptBody](t, resp)
	if got.ID == "" || got.Status != "accepted" {
		t.Fatalf("body = %+v, want non-empty id and status accepted", got)
	}
	list := listMessages(t, srv.URL, "")
	if list.Total != 1 || len(list.Messages) != 1 {
		t.Fatalf("list = %+v, want 1 message", list)
	}
	m := list.Messages[0]
	want := Message{ID: got.ID, Channel: "sms", To: "+905321234567", Body: "Merhaba Ayşe", Reference: "c1:u1:sms", ReceivedAt: clock.Now().UTC()}
	if m != want {
		t.Fatalf("message = %+v, want %+v", m, want)
	}
	if m.ReceivedAt.Location() != time.UTC {
		t.Fatalf("received_at location = %v, want UTC", m.ReceivedAt.Location())
	}
}

func TestSend_EmailAccepted(t *testing.T) {
	clock := newFakeClock()
	srv := startServer(t, normalOptions(clock))
	resp := post(t, srv.URL, "/v1/email", email("ali@example.com", "Sepetin seni bekliyor", "<p>Merhaba</p>", ""), nil)
	wantStatus(t, resp, http.StatusAccepted)
	got := decode[acceptBody](t, resp)
	if got.ID == "" || got.Status != "accepted" {
		t.Fatalf("body = %+v", got)
	}
	list := listMessages(t, srv.URL, "")
	if list.Total != 1 {
		t.Fatalf("total = %d, want 1", list.Total)
	}
	m := list.Messages[0]
	if m.Channel != "email" || m.To != "ali@example.com" || m.Subject != "Sepetin seni bekliyor" || m.Body != "<p>Merhaba</p>" || m.ID != got.ID {
		t.Fatalf("message = %+v", m)
	}
}

func TestSend_Invalid(t *testing.T) {
	cases := []struct {
		name, path, body string
	}{
		{"sms body 161", "/v1/sms", sms("+905321234567", strings.Repeat("a", 161), "")},
		{"sms body empty", "/v1/sms", sms("+905321234567", "", "")},
		{"sms to not E.164", "/v1/sms", sms("05321234567", "hi", "")},
		{"sms to missing", "/v1/sms", `{"body":"hi"}`},
		{"email subject 81", "/v1/email", email("ali@example.com", strings.Repeat("ğ", 81), "b", "")},
		{"email subject empty", "/v1/email", email("ali@example.com", "", "b", "")},
		{"email to bad", "/v1/email", email("not-an-email", "s", "b", "")},
		{"email to with display name", "/v1/email", email("Ali <ali@example.com>", "s", "b", "")},
		{"email body empty", "/v1/email", email("ali@example.com", "s", "", "")},
		{"reference 129", "/v1/sms", sms("+905321234567", "hi", strings.Repeat("r", 129))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := startServer(t, normalOptions(newFakeClock()))
			resp := post(t, srv.URL, tc.path, tc.body, nil)
			wantStatus(t, resp, http.StatusBadRequest)
			if e := decode[errorBody](t, resp); e.Error.Code != "invalid_request" || e.Error.Message == "" {
				t.Fatalf("error = %+v, want code invalid_request with a message", e)
			}
			if total := listMessages(t, srv.URL, "").Total; total != 0 {
				t.Fatalf("total = %d, want 0", total)
			}
		})
	}
}

func TestSend_RuneCount(t *testing.T) {
	srv := startServer(t, normalOptions(newFakeClock()))
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", strings.Repeat("ş", 160), ""), nil), http.StatusAccepted)
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", strings.Repeat("ş", 161), ""), nil), http.StatusBadRequest)
	wantStatus(t, post(t, srv.URL, "/v1/email", email("ali@example.com", strings.Repeat("ğ", 80), "b", ""), nil), http.StatusAccepted)
}

func TestSend_BadJSON(t *testing.T) {
	big := `{"to":"+905321234567","body":"` + strings.Repeat("a", 2<<20) + `"}`
	for name, body := range map[string]string{
		"wrong type": `{"to": 123, "body": "hi"}`,
		"not json":   `not json`,
		"too large":  big,
	} {
		t.Run(name, func(t *testing.T) {
			srv := startServer(t, normalOptions(newFakeClock()))
			resp := post(t, srv.URL, "/v1/sms", body, nil)
			wantStatus(t, resp, http.StatusBadRequest)
			if e := decode[errorBody](t, resp); e.Error.Code != "invalid_request" {
				t.Fatalf("code = %q, want invalid_request", e.Error.Code)
			}
		})
	}
}

func TestSend_RepeatDelivers(t *testing.T) {
	srv := startServer(t, normalOptions(newFakeClock()))
	body := sms("+905321234567", "hi", "c1:u1:sms")
	r1 := post(t, srv.URL, "/v1/sms", body, nil)
	r2 := post(t, srv.URL, "/v1/sms", body, nil)
	wantStatus(t, r1, http.StatusAccepted)
	wantStatus(t, r2, http.StatusAccepted)
	a, b := decode[acceptBody](t, r1), decode[acceptBody](t, r2)
	if a.ID == b.ID {
		t.Fatalf("ids are equal: %q", a.ID)
	}
	if total := listMessages(t, srv.URL, "").Total; total != 2 {
		t.Fatalf("total = %d, want 2", total)
	}
	if d := stats(t, srv.URL)["sms"].DuplicateDeliveries; d != 1 {
		t.Fatalf("duplicate_deliveries = %d, want 1", d)
	}
}

func TestSend_EmptyReferenceNotDuplicate(t *testing.T) {
	srv := startServer(t, normalOptions(newFakeClock()))
	for range 3 {
		wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", ""), nil), http.StatusAccepted)
	}
	if d := stats(t, srv.URL)["sms"].DuplicateDeliveries; d != 0 {
		t.Fatalf("duplicate_deliveries = %d, want 0", d)
	}
}

func TestSend_DuplicatePerChannel(t *testing.T) {
	srv := startServer(t, normalOptions(newFakeClock()))
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "hi", "c1:u1"), nil), http.StatusAccepted)
	wantStatus(t, post(t, srv.URL, "/v1/email", email("ali@example.com", "s", "b", "c1:u1"), nil), http.StatusAccepted)
	st := stats(t, srv.URL)
	if st["sms"].DuplicateDeliveries != 0 || st["email"].DuplicateDeliveries != 0 {
		t.Fatalf("stats = %+v, want no duplicates", st)
	}
}

func TestMessages_Filters(t *testing.T) {
	clock := newFakeClock()
	srv := startServer(t, normalOptions(clock))
	sends := []struct{ path, body string }{
		{"/v1/sms", sms("+905320000001", "1", "r1")},
		{"/v1/sms", sms("+905320000002", "2", "r2")},
		{"/v1/email", email("a@example.com", "s", "3", "r1")},
		{"/v1/sms", sms("+905320000001", "4", "r2")},
		{"/v1/email", email("b@example.com", "s", "5", "r2")},
	}
	for _, s := range sends {
		wantStatus(t, post(t, srv.URL, s.path, s.body, nil), http.StatusAccepted)
		clock.Advance(time.Second)
	}
	bodies := func(l listBody) string {
		var out []string
		for _, m := range l.Messages {
			out = append(out, m.Body)
		}
		return strings.Join(out, ",")
	}
	cases := []struct {
		query     string
		wantTotal int
		want      string
	}{
		{"", 5, "5,4,3,2,1"},
		{"?channel=sms", 3, "4,2,1"},
		{"?channel=email", 2, "5,3"},
		{"?to=%2B905320000001", 2, "4,1"},
		{"?reference=r2", 3, "5,4,2"},
		{"?channel=sms&reference=r2", 2, "4,2"},
		{"?limit=2&offset=1", 5, "4,3"},
		{"?offset=10", 5, ""},
	}
	for _, tc := range cases {
		got := listMessages(t, srv.URL, tc.query)
		if got.Total != tc.wantTotal || bodies(got) != tc.want {
			t.Errorf("%q: total %d bodies %q, want %d %q", tc.query, got.Total, bodies(got), tc.wantTotal, tc.want)
		}
	}
}

func TestMessages_EmptyIsArray(t *testing.T) {
	srv := startServer(t, normalOptions(newFakeClock()))
	b, _ := io.ReadAll(get(t, srv.URL, "/v1/messages").Body)
	if !strings.Contains(string(b), `"messages":[]`) {
		t.Fatalf("body = %s, want an empty messages array", b)
	}
}

func TestMessages_BadQuery(t *testing.T) {
	srv := startServer(t, normalOptions(newFakeClock()))
	for _, q := range []string{"limit=0", "limit=1001", "limit=x", "offset=-1", "channel=push"} {
		resp := get(t, srv.URL, "/v1/messages?"+q)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", q, resp.StatusCode)
			continue
		}
		if e := decode[errorBody](t, resp); e.Error.Code != "invalid_request" {
			t.Errorf("%s: code %q, want invalid_request", q, e.Error.Code)
		}
	}
}

func TestStats_Counts(t *testing.T) {
	srv := startServer(t, normalOptions(newFakeClock()))
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "a", ""), nil), http.StatusAccepted)
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "b", ""), nil), http.StatusAccepted)
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("bad", "c", ""), nil), http.StatusBadRequest)
	wantStatus(t, post(t, srv.URL, "/v1/email", email("ali@example.com", "s", "b", ""), nil), http.StatusAccepted)
	st := stats(t, srv.URL)
	if want := (ChannelStats{Requests: 3, Delivered: 2, Rejected: 1}); st["sms"] != want {
		t.Errorf("sms = %+v, want %+v", st["sms"], want)
	}
	if want := (ChannelStats{Requests: 1, Delivered: 1}); st["email"] != want {
		t.Errorf("email = %+v, want %+v", st["email"], want)
	}
}

func TestReset(t *testing.T) {
	srv := startServer(t, normalOptions(newFakeClock()))
	wantStatus(t, post(t, srv.URL, "/v1/sms", sms("+905321234567", "a", "r1"), nil), http.StatusAccepted)
	wantStatus(t, post(t, srv.URL, "/v1/email", email("ali@example.com", "s", "b", ""), nil), http.StatusAccepted)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/v1/messages", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	wantStatus(t, resp, http.StatusNoContent)

	if total := listMessages(t, srv.URL, "").Total; total != 0 {
		t.Fatalf("total = %d, want 0", total)
	}
	st := stats(t, srv.URL)
	if st["sms"] != (ChannelStats{}) || st["email"] != (ChannelStats{}) {
		t.Fatalf("stats = %+v, want all 0", st)
	}
	resp = post(t, srv.URL, "/v1/sms", sms("+905321234567", "a", "r1"), nil)
	wantStatus(t, resp, http.StatusAccepted)
	if d := stats(t, srv.URL)["sms"].DuplicateDeliveries; d != 0 {
		t.Fatalf("duplicate_deliveries = %d, want 0 after reset", d)
	}
	// An id is unique for each delivery, also across a reset.
	if id := decode[acceptBody](t, resp).ID; id == "sms_00000001" {
		t.Fatalf("id %q was used before the reset", id)
	}
}

func TestConcurrent_DuplicateCount(t *testing.T) {
	srv := startServer(t, normalOptions(newFakeClock()))
	var wg sync.WaitGroup
	errs := make(chan error, 500)
	for g := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 10 {
				body := sms("+905321234567", fmt.Sprintf("g%d-%d", g, i), "same-ref")
				resp, err := http.Post(srv.URL+"/v1/sms", "application/json", strings.NewReader(body))
				if err != nil {
					errs <- err
					continue
				}
				resp.Body.Close()
				if resp.StatusCode != http.StatusAccepted {
					errs <- fmt.Errorf("status %d", resp.StatusCode)
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	st := stats(t, srv.URL)["sms"]
	if st.Delivered != 500 || st.DuplicateDeliveries != 499 || st.Requests != 500 {
		t.Fatalf("sms = %+v, want delivered 500, duplicates 499", st)
	}
}

// TestConcurrent_CountersWithFailures checks NFR-1 with all modes active:
// delivered is the count of 202 responses plus lost responses.
func TestConcurrent_CountersWithFailures(t *testing.T) {
	o := normalOptions(newFakeClock())
	o.ErrorRate, o.SlowRate, o.LostRate = 0.2, 0.2, 0.2
	srv := startServer(t, o)
	var (
		wg               sync.WaitGroup
		mu               sync.Mutex
		accepted, failed int64
		transport        int64
	)
	for g := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 10 {
				body := sms("+905321234567", fmt.Sprintf("g%d-%d", g, i), "")
				resp, err := http.Post(srv.URL+"/v1/sms", "application/json", strings.NewReader(body))
				mu.Lock()
				switch {
				case err != nil:
					transport++
				case resp.StatusCode == http.StatusAccepted:
					accepted++
				default:
					failed++
				}
				mu.Unlock()
				if err == nil {
					resp.Body.Close()
				}
			}
		}()
	}
	wg.Wait()
	st := stats(t, srv.URL)["sms"]
	if st.Lost == 0 || st.Errors == 0 || st.Slow == 0 {
		t.Fatalf("sms = %+v, want each failure mode at least once", st)
	}
	if st.Delivered != accepted+st.Lost {
		t.Fatalf("delivered %d, want 202 count %d + lost %d", st.Delivered, accepted, st.Lost)
	}
	if transport != st.Lost || failed != st.Errors || st.Requests != 500 {
		t.Fatalf("transport errors %d, other failures %d, sms = %+v", transport, failed, st)
	}
}

func TestHealthz(t *testing.T) {
	srv := startServer(t, normalOptions(newFakeClock()))
	wantStatus(t, get(t, srv.URL, "/healthz"), http.StatusOK)
}
