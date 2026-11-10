package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"campaigncomposer/internal/delivery"
)

type fakeDelivery struct {
	mu       sync.Mutex
	progress []delivery.Progress // returned in order; the last one repeats
	calls    int
	err      error
	jobs     []delivery.Job
	next     *int64
	gotID    int64
	gotF     delivery.JobFilter
	retried  int
}

func (f *fakeDelivery) Progress(_ context.Context, id int64) (delivery.Progress, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gotID = id
	if f.err != nil {
		return delivery.Progress{}, f.err
	}
	i := min(f.calls, len(f.progress)-1)
	f.calls++
	return f.progress[i], nil
}

func (f *fakeDelivery) ListJobs(_ context.Context, id int64, fl delivery.JobFilter) ([]delivery.Job, *int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gotID, f.gotF = id, fl
	return f.jobs, f.next, f.err
}

func (f *fakeDelivery) Retry(_ context.Context, id int64) (int, delivery.Progress, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gotID = id
	if f.err != nil {
		return 0, delivery.Progress{}, f.err
	}
	return f.retried, f.progress[0], nil
}

func prog(status string, sent int64) delivery.Progress {
	return delivery.Progress{CampaignID: 7, Status: status, Channels: map[delivery.Channel]delivery.ChannelProgress{
		delivery.ChannelSMS:   {Total: 10, Sent: sent, Queued: 10 - sent},
		delivery.ChannelEmail: {Total: 4, Failed: 1, RetryableFailed: 1},
	}}
}

func deliveryMux(f *fakeDelivery, poll time.Duration) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterDelivery(mux, f, poll)
	return mux
}

func do(mux http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestProgress_Shape(t *testing.T) {
	f := &fakeDelivery{progress: []delivery.Progress{prog("sending", 3)}}
	rec := do(deliveryMux(f, time.Second), "GET", "/api/campaigns/7/progress")
	if rec.Code != 200 {
		t.Fatalf("code %d: %s", rec.Code, rec.Body)
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["campaign_id"] != float64(7) || got["status"] != "sending" {
		t.Errorf("body = %v", got)
	}
	ch := got["channels"].(map[string]any)
	sms := ch["sms"].(map[string]any)
	for _, k := range []string{"total", "queued", "sending", "retrying", "sent", "failed", "unknown", "skipped", "retryable_failed"} {
		if _, ok := sms[k]; !ok {
			t.Errorf("sms lacks %s", k)
		}
	}
	if _, ok := ch["email"]; !ok || sms["sent"] != float64(3) {
		t.Errorf("channels = %v", ch)
	}
}

func TestDelivery_BadID(t *testing.T) {
	f := &fakeDelivery{progress: []delivery.Progress{prog("sending", 0)}}
	mux := deliveryMux(f, time.Second)
	for _, id := range []string{"abc", "0", "-1"} {
		for _, rt := range [][2]string{{"GET", "progress"}, {"GET", "progress/stream"}, {"GET", "send-jobs"}, {"POST", "retry"}} {
			if rec := do(mux, rt[0], "/api/campaigns/"+id+"/"+rt[1]); rec.Code != 400 {
				t.Errorf("%s %s id %s = %d", rt[0], rt[1], id, rec.Code)
			}
		}
	}
}

func TestDelivery_ErrorMapping(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{delivery.ErrNotFound, 404, "not_found"},
		{delivery.ErrNotSendable, 409, "campaign_not_sendable"},
		{delivery.ErrNotCompleted, 409, "campaign_not_completed"},
		{delivery.ErrNothingToRetry, 409, "nothing_to_retry"},
		{errors.New("boom"), 503, "db_unavailable"},
	}
	for _, c := range cases {
		f := &fakeDelivery{err: c.err}
		mux := deliveryMux(f, time.Second)
		for _, rt := range [][2]string{{"GET", "progress"}, {"GET", "progress/stream"}, {"GET", "send-jobs"}, {"POST", "retry"}} {
			rec := do(mux, rt[0], "/api/campaigns/7/"+rt[1])
			var b errorBody
			_ = json.Unmarshal(rec.Body.Bytes(), &b)
			if rec.Code != c.status || b.Error.Code != c.code {
				t.Errorf("%v on %s = %d %s", c.err, rt[1], rec.Code, b.Error.Code)
			}
		}
	}
}

func TestSendJobs_Query(t *testing.T) {
	f := &fakeDelivery{}
	mux := deliveryMux(f, time.Second)
	for _, q := range []string{"status=bogus", "channel=push", "limit=0", "limit=201", "limit=x", "after=-1", "after=0", "after=x"} {
		if rec := do(mux, "GET", "/api/campaigns/7/send-jobs?"+q); rec.Code != 400 {
			t.Errorf("%s = %d", q, rec.Code)
		}
	}
	for _, s := range []string{"queued", "sending", "retrying", "sent", "failed", "unknown", "skipped"} {
		if rec := do(mux, "GET", "/api/campaigns/7/send-jobs?status="+s); rec.Code != 200 {
			t.Errorf("status %s = %d", s, rec.Code)
		}
	}
	if rec := do(mux, "GET", "/api/campaigns/7/send-jobs?status=failed&channel=email&limit=200&after=12"); rec.Code != 200 {
		t.Fatalf("good query = %d", rec.Code)
	}
	want := delivery.JobFilter{Status: "failed", Channel: "email", Limit: 200, After: 12}
	if f.gotF != want || f.gotID != 7 {
		t.Errorf("filter = %+v id %d", f.gotF, f.gotID)
	}
	do(mux, "GET", "/api/campaigns/7/send-jobs")
	if f.gotF.Limit != 50 {
		t.Errorf("default limit = %d", f.gotF.Limit)
	}
}

func TestSendJobs_Keys(t *testing.T) {
	next := int64(9)
	f := &fakeDelivery{next: &next, jobs: []delivery.Job{{ID: 9, CampaignID: 7, UserID: 3, Channel: "sms", Language: "tr",
		Status: "failed", Attempts: 5, LastError: "x", Retryable: true, UpdatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}}}
	rec := do(deliveryMux(f, time.Second), "GET", "/api/campaigns/7/send-jobs")
	var got struct {
		SendJobs  []map[string]any `json:"send_jobs"`
		NextAfter *int64           `json:"next_after"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || len(got.SendJobs) != 1 || got.NextAfter == nil || *got.NextAfter != 9 {
		t.Fatalf("body %s, %v", rec.Body, err)
	}
	keys := map[string]bool{}
	for k := range got.SendJobs[0] {
		keys[k] = true
	}
	want := []string{"id", "user_id", "channel", "language", "status", "attempts", "last_error", "updated_at"}
	if len(keys) != len(want) {
		t.Errorf("keys = %v", keys)
	}
	for _, k := range want {
		if !keys[k] {
			t.Errorf("missing %s", k)
		}
	}
}

func TestSendJobs_Empty(t *testing.T) {
	rec := do(deliveryMux(&fakeDelivery{}, time.Second), "GET", "/api/campaigns/7/send-jobs")
	if !strings.Contains(rec.Body.String(), `"send_jobs":[]`) || !strings.Contains(rec.Body.String(), `"next_after":null`) {
		t.Errorf("body = %s", rec.Body)
	}
}

func TestRetry_Response(t *testing.T) {
	f := &fakeDelivery{progress: []delivery.Progress{prog("sending", 3)}, retried: 4}
	rec := do(deliveryMux(f, time.Second), "POST", "/api/campaigns/7/retry")
	var got struct {
		Retried  int `json:"retried"`
		Progress struct {
			Status string `json:"status"`
		} `json:"progress"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || rec.Code != 200 || got.Retried != 4 || got.Progress.Status != "sending" {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
}

func TestProgressStream(t *testing.T) {
	f := &fakeDelivery{progress: []delivery.Progress{prog("sending", 3), prog("sending", 3), prog("sending", 3), prog("sending", 5), prog("completed", 10)}}
	rec := do(deliveryMux(f, 10*time.Millisecond), "GET", "/api/campaigns/7/progress/stream")
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type %q", ct)
	}
	evs := parseSSE(t, rec.Body.String())
	var names []string
	for _, e := range evs {
		names = append(names, e.Event)
	}
	if strings.Join(names, ",") != "progress,progress,progress,done" {
		t.Fatalf("events = %v", names)
	}
	var last delivery.Progress
	_ = json.Unmarshal(evs[2].Data, &last)
	if last.Status != "completed" {
		t.Errorf("last progress = %+v", last)
	}
}

func TestProgressStream_Draft(t *testing.T) {
	rec := do(deliveryMux(&fakeDelivery{err: delivery.ErrNotSendable}, time.Second), "GET", "/api/campaigns/7/progress/stream")
	if rec.Code != 409 || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		t.Errorf("%d %s", rec.Code, rec.Header().Get("Content-Type"))
	}
}

func TestProgressStream_Disconnect(t *testing.T) {
	f := &fakeDelivery{progress: []delivery.Progress{prog("sending", 3)}}
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest("GET", "/api/campaigns/7/progress/stream", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	done := make(chan struct{})
	go func() { deliveryMux(f, 5*time.Millisecond).ServeHTTP(rec, req); close(done) }()
	time.Sleep(30 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not stop after disconnect")
	}
}
