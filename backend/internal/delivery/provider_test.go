package delivery

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"campaigncomposer/internal/mockprovider"
)

// failTransport sets X-Mock-Fail on each request. modes are used one for
// each request, in order. After the last one, "none" is used.
type failTransport struct {
	base  http.RoundTripper
	mu    sync.Mutex
	modes []string
}

func (f *failTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	f.mu.Lock()
	mode := "none"
	if len(f.modes) > 0 {
		mode, f.modes = f.modes[0], f.modes[1:]
	}
	f.mu.Unlock()
	r = r.Clone(r.Context())
	r.Header.Set("X-Mock-Fail", mode)
	return f.base.RoundTrip(r)
}

func (f *failTransport) set(modes ...string) {
	f.mu.Lock()
	f.modes = modes
	f.mu.Unlock()
}

func newMock(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(mockprovider.New(mockprovider.Options{
		RatePerSec: 1000, Burst: 1000, SlowDelay: 200 * time.Millisecond, LatencyMax: 0, Random: 1,
	}))
	t.Cleanup(srv.Close)
	return srv
}

func providerStats(t *testing.T, base string) map[string]mockprovider.ChannelStats {
	t.Helper()
	resp, err := http.Get(base + "/v1/stats")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]mockprovider.ChannelStats
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

// providerWith returns a provider on base that forces the given modes.
func providerWith(base string, timeout time.Duration, modes ...string) (*Provider, *failTransport) {
	p := NewProvider(base, timeout)
	ft := &failTransport{base: p.HTTP.Transport, modes: modes}
	p.HTTP.Transport = ft
	return p, ft
}

var smsMsg = Message{To: "+905321234567", Body: "hello"}

func TestProvider_Outcomes(t *testing.T) {
	srv := newMock(t)
	tests := []struct {
		mode       string
		kind       OutcomeKind
		retryAfter time.Duration
		errHas     string
	}{
		{"none", Delivered, 0, ""},
		{"429", Retryable, time.Second, "429"},
		{"500", Retryable, 0, "500"},
		{"503", Retryable, 0, "503"},
		{"slow", Ambiguous, 0, "timeout"},
		{"lost", Ambiguous, 0, ""},
	}
	for _, tc := range tests {
		t.Run(tc.mode, func(t *testing.T) {
			p, _ := providerWith(srv.URL, 50*time.Millisecond, tc.mode)
			o := p.Send(context.Background(), ChannelSMS, smsMsg, "1:2:sms")
			if o.Kind != tc.kind || o.RetryAfter != tc.retryAfter {
				t.Fatalf("Send = %+v, want kind %v retry-after %v", o, tc.kind, tc.retryAfter)
			}
			if tc.kind == Delivered && o.ProviderID == "" {
				t.Fatal("Delivered with no provider id")
			}
			if tc.kind != Delivered && o.Err == "" {
				t.Fatal("no error text")
			}
			if tc.errHas != "" && !strings.Contains(o.Err, tc.errHas) {
				t.Fatalf("Err = %q, want it to contain %q", o.Err, tc.errHas)
			}
		})
	}
	t.Run("400", func(t *testing.T) {
		p := NewProvider(srv.URL, time.Second)
		o := p.Send(context.Background(), ChannelSMS, Message{}, "1:2:sms")
		if o.Kind != Permanent || !strings.Contains(o.Err, "400") {
			t.Fatalf("Send = %+v, want Permanent 400", o)
		}
	})
	t.Run("email", func(t *testing.T) {
		p := NewProvider(srv.URL, time.Second)
		o := p.Send(context.Background(), ChannelEmail, Message{To: "a@example.com", Subject: "s", Body: "b"}, "1:2:email")
		if o.Kind != Delivered {
			t.Fatalf("Send = %+v", o)
		}
	})
}

func deadURL(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return "http://" + addr
}

func TestProvider_ConnectionRefused(t *testing.T) {
	p := NewProvider(deadURL(t), time.Second)
	o := p.Send(context.Background(), ChannelSMS, smsMsg, "1:2:sms")
	if o.Kind != Retryable || o.Err != "connection refused" {
		t.Fatalf("Send = %+v, want Retryable connection refused", o)
	}
}

func TestProvider_OneRequestPerAttempt(t *testing.T) {
	srv := newMock(t)
	p, _ := providerWith(srv.URL, time.Second, "lost")
	if o := p.Send(context.Background(), ChannelSMS, smsMsg, "1:2:sms"); o.Kind != Ambiguous {
		t.Fatalf("Send = %+v, want Ambiguous", o)
	}
	if got := providerStats(t, srv.URL)["sms"].Requests; got != 1 {
		t.Fatalf("requests = %d, want 1", got)
	}
}

func TestProvider_NonJSON202(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()
	o := NewProvider(srv.URL, time.Second).Send(context.Background(), ChannelSMS, smsMsg, "1:2:sms")
	if o.Kind != Delivered || o.ProviderID != "" {
		t.Fatalf("Send = %+v, want Delivered with no id", o)
	}
}

func TestProvider_Other4xxPermanent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	o := NewProvider(srv.URL, time.Second).Send(context.Background(), ChannelSMS, smsMsg, "1:2:sms")
	if o.Kind != Permanent || !strings.Contains(o.Err, "401") {
		t.Fatalf("Send = %+v", o)
	}
}

func TestProvider_RequestShape(t *testing.T) {
	type rec struct {
		path string
		key  string
		body map[string]any
	}
	var got []rec
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b map[string]any
		_ = json.NewDecoder(r.Body).Decode(&b)
		got = append(got, rec{r.URL.Path, r.Header.Get("Idempotency-Key"), b})
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"id":"x"}`))
	}))
	defer srv.Close()
	p := NewProvider(srv.URL, time.Second)
	p.Send(context.Background(), ChannelSMS, smsMsg, "7:42:sms")
	p.Send(context.Background(), ChannelEmail, Message{To: "a@example.com", Subject: "s", Body: "b"}, "7:42:email")
	if len(got) != 2 {
		t.Fatalf("%d requests", len(got))
	}
	if got[0].path != "/v1/sms" || got[0].body["reference"] != "7:42:sms" || got[0].body["to"] != "+905321234567" ||
		got[0].body["body"] != "hello" || len(got[0].body) != 3 {
		t.Fatalf("sms request = %+v", got[0])
	}
	if got[1].path != "/v1/email" || got[1].body["subject"] != "s" || got[1].body["reference"] != "7:42:email" || len(got[1].body) != 4 {
		t.Fatalf("email request = %+v", got[1])
	}
	if got[0].key != "" || got[1].key != "" {
		t.Fatal("Idempotency-Key header is set")
	}
}
