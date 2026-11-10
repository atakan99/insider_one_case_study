package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

// Clock gives the time and waits. Tests use a fake one.
type Clock interface {
	Now() time.Time
	// Sleep waits for d. It returns ctx.Err() when ctx is cancelled.
	Sleep(ctx context.Context, d time.Duration) error
}

type realClock struct{}

// RealClock returns the Clock of the system.
func RealClock() Clock { return realClock{} }

func (realClock) Now() time.Time { return time.Now() }

func (realClock) Sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// OutcomeKind is the class of the result of one call to the provider.
type OutcomeKind int

// Outcome kinds. Only Retryable allows a new attempt: the provider did not
// deliver. Ambiguous means that the message may have been delivered.
const (
	Delivered OutcomeKind = iota
	Retryable
	Permanent
	Ambiguous
)

// Outcome is the classified result of one call to the provider.
type Outcome struct {
	Kind       OutcomeKind
	ProviderID string
	RetryAfter time.Duration // from a 429
	Err        string        // for example "503 unavailable", "timeout", "connection refused"
}

const maxResponseBody = 1 << 20

// Provider is the HTTP client of the mock provider. One Send is one request.
type Provider struct {
	BaseURL string
	HTTP    *http.Client
}

// NewProvider returns a Provider with the given attempt timeout. The client
// uses no proxy, follows no redirect and does not retry.
func NewProvider(baseURL string, timeout time.Duration) *Provider {
	return &Provider{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP: &http.Client{
			Timeout:       timeout,
			Transport:     &http.Transport{Proxy: nil, MaxIdleConnsPerHost: 64, IdleConnTimeout: 30 * time.Second},
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

// Send makes one request and classifies the result. It does not retry.
func (p *Provider) Send(ctx context.Context, ch Channel, m Message, reference string) Outcome {
	var payload any
	path := "/v1/sms"
	switch ch {
	case ChannelEmail:
		path = "/v1/email"
		payload = map[string]string{"to": m.To, "subject": m.Subject, "body": m.Body, "reference": reference}
	default:
		payload = map[string]string{"to": m.To, "body": m.Body, "reference": reference}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Outcome{Kind: Permanent, Err: "request could not be encoded"}
	}

	// Any call of WroteRequest, also with an error, counts as written: the
	// provider may have the request, so the outcome is not known.
	var written atomic.Bool
	trace := &httptrace.ClientTrace{WroteRequest: func(httptrace.WroteRequestInfo) { written.Store(true) }}
	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), http.MethodPost, p.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return Outcome{Kind: Permanent, Err: "request could not be built"}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTP.Do(req)
	if err != nil {
		return transportOutcome(err, written.Load())
	}
	defer resp.Body.Close()

	switch code := resp.StatusCode; {
	case code == http.StatusAccepted:
		var r struct {
			ID string `json:"id"`
		}
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
		_ = json.Unmarshal(raw, &r) // a body that is not JSON still counts as delivered
		return Outcome{Kind: Delivered, ProviderID: r.ID}
	case code == http.StatusTooManyRequests:
		secs, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		return Outcome{Kind: Retryable, RetryAfter: time.Duration(max(1, secs)) * time.Second, Err: statusText(resp)}
	case code >= 500:
		return Outcome{Kind: Retryable, Err: statusText(resp)}
	default:
		return Outcome{Kind: Permanent, Err: statusText(resp)}
	}
}

// statusText gives "<code> <error code of the body>", for example "503 unavailable".
func statusText(resp *http.Response) string {
	var e struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	_ = json.Unmarshal(raw, &e)
	name := e.Error.Code
	if name == "" {
		name = strings.ToLower(http.StatusText(resp.StatusCode))
	}
	return strings.TrimSpace(fmt.Sprintf("%d %s", resp.StatusCode, name))
}

// transportOutcome classifies an error of the HTTP client. The text is fixed
// and does not hold the URL or the error text of the library.
func transportOutcome(err error, written bool) Outcome {
	if written {
		var ne net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &ne) && ne.Timeout()) {
			return Outcome{Kind: Ambiguous, Err: "timeout"}
		}
		if errors.Is(err, context.Canceled) {
			return Outcome{Kind: Ambiguous, Err: "canceled"}
		}
		return Outcome{Kind: Ambiguous, Err: "connection lost"}
	}
	var dns *net.DNSError
	switch {
	case errors.Is(err, syscall.ECONNREFUSED):
		return Outcome{Kind: Retryable, Err: "connection refused"}
	case errors.As(err, &dns):
		return Outcome{Kind: Retryable, Err: "dns error"}
	case errors.Is(err, context.Canceled):
		return Outcome{Kind: Retryable, Err: "canceled"}
	case errors.Is(err, context.DeadlineExceeded):
		return Outcome{Kind: Retryable, Err: "timeout before request"}
	}
	return Outcome{Kind: Retryable, Err: "connection error"}
}
