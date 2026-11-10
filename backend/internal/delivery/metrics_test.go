package delivery

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// sample returns the value of the series of name with exactly these labels.
// The bool is false when the series does not exist.
func sample(t *testing.T, reg *prometheus.Registry, name string, labels map[string]string) (float64, bool) {
	t.Helper()
	fams, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fams {
		if f.GetName() != name {
			continue
		}
	next:
		for _, m := range f.GetMetric() {
			if len(m.GetLabel()) != len(labels) {
				continue
			}
			for _, l := range m.GetLabel() {
				if labels[l.GetName()] != l.GetValue() {
					continue next
				}
			}
			switch {
			case m.GetCounter() != nil:
				return m.GetCounter().GetValue(), true
			case m.GetGauge() != nil:
				return m.GetGauge().GetValue(), true
			case m.GetHistogram() != nil:
				return float64(m.GetHistogram().GetSampleCount()), true
			}
		}
	}
	return 0, false
}

func wantSample(t *testing.T, reg *prometheus.Registry, name string, labels map[string]string, want float64) {
	t.Helper()
	got, ok := sample(t, reg, name, labels)
	if !ok || got != want {
		t.Fatalf("%s%v = %v (found %v), want %v", name, labels, got, ok, want)
	}
}

func TestOutcomeKind_String(t *testing.T) {
	for k, want := range map[OutcomeKind]string{
		Delivered: "delivered", Retryable: "retryable", Permanent: "permanent", Ambiguous: "ambiguous",
	} {
		if got := k.String(); got != want {
			t.Errorf("OutcomeKind(%d) = %q, want %q", int(k), got, want)
		}
	}
}

func TestMetrics_NilSafe(t *testing.T) {
	var m *Metrics
	m.Attempt(ChannelSMS, Delivered, time.Second)
	m.Result(ChannelSMS, StatusSent)
	m.Waited(ChannelSMS, time.Second)
	m.RetryScheduled(ChannelSMS)
}

func TestMetrics_Counters(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg, nil)
	m.Attempt(ChannelEmail, Permanent, 30*time.Millisecond)
	m.Waited(ChannelEmail, 1500*time.Millisecond)
	wantSample(t, reg, "campaign_send_attempts_total", map[string]string{"channel": "email", "outcome": "permanent"}, 1)
	wantSample(t, reg, "campaign_provider_request_duration_seconds", map[string]string{"channel": "email"}, 1)
	wantSample(t, reg, "campaign_rate_limit_wait_seconds_total", map[string]string{"channel": "email"}, 1.5)
}

func TestMetrics_StreamGauges(t *testing.T) {
	e := newConsumerEnv(t)
	reg := prometheus.NewRegistry()
	NewMetrics(reg, e.queue)
	e.entries(t, e.sms[0], 3)
	wantSample(t, reg, "campaign_stream_length", map[string]string{"channel": "sms"}, 3)
	wantSample(t, reg, "campaign_stream_pending", map[string]string{"channel": "sms"}, 3)
	wantSample(t, reg, "campaign_stream_length", map[string]string{"channel": "email"}, 0)
}

func TestConsumer_Metrics(t *testing.T) {
	e := newConsumerEnv(t)
	srv := newMock(t)
	p, ft := providerWith(srv.URL, 50*time.Millisecond)
	reg := prometheus.NewRegistry()
	c := e.consumer(p)
	c.Metrics = NewMetrics(reg, e.queue)
	id := e.sms[0]

	ft.set("503")
	if err := c.Handle(e.ctx, e.entries(t, id, 1)[0]); err != nil {
		t.Fatal(err)
	}
	e.clk.Set(e.nextAttempt(t, id))
	ft.set("none")
	if err := c.Handle(e.ctx, e.entries(t, id, 1)[0]); err != nil {
		t.Fatal(err)
	}
	sms := func(k, v string) map[string]string { return map[string]string{"channel": "sms", k: v} }
	wantSample(t, reg, "campaign_send_attempts_total", sms("outcome", "retryable"), 1)
	wantSample(t, reg, "campaign_send_attempts_total", sms("outcome", "delivered"), 1)
	wantSample(t, reg, "campaign_retries_scheduled_total", map[string]string{"channel": "sms"}, 1)
	wantSample(t, reg, "campaign_send_results_total", sms("status", "retrying"), 1)
	wantSample(t, reg, "campaign_send_results_total", sms("status", "sent"), 1)
	wantSample(t, reg, "campaign_provider_request_duration_seconds", map[string]string{"channel": "sms"}, 2)
}

func TestConsumer_MetricsSkip(t *testing.T) {
	e := newConsumerEnv(t)
	srv := newMock(t)
	reg := prometheus.NewRegistry()
	c := e.consumer(NewProvider(srv.URL, 2*time.Second))
	c.Metrics = NewMetrics(reg, e.queue)
	exec(t, e.ctx, e.pool, `UPDATE users SET sms_consent = false WHERE id = (SELECT user_id FROM send_jobs WHERE id = $1)`, e.sms[0])
	if err := c.Handle(e.ctx, e.entries(t, e.sms[0], 1)[0]); err != nil {
		t.Fatal(err)
	}
	if j := getJob(t, e.ctx, e.pool, e.sms[0]); j.Status != StatusSkipped || j.LastError != SkipNoConsent {
		t.Fatalf("job = %s %q", j.Status, j.LastError)
	}
	wantSample(t, reg, "campaign_send_results_total", map[string]string{"channel": "sms", "status": "skipped"}, 1)
	fams, _ := reg.Gather()
	for _, f := range fams {
		if strings.HasPrefix(f.GetName(), "campaign_send_attempts_total") || f.GetName() == "campaign_provider_request_duration_seconds" {
			t.Fatalf("attempt sample after a skip: %s", f.GetName())
		}
	}
}
