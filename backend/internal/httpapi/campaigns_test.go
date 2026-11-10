package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"campaigncomposer/internal/campaign"
	"campaigncomposer/internal/content"
)

// fakeCampaigns is a CampaignService with fixed answers.
type fakeCampaigns struct {
	c       campaign.Campaign
	list    []campaign.Summary
	sample  []campaign.SampleNames
	err     error
	repeat  bool
	gotID   int64
	gotBody json.RawMessage
	calls   int
}

func (f *fakeCampaigns) Get(_ context.Context, id int64) (campaign.Campaign, error) {
	f.gotID, f.calls = id, f.calls+1
	return f.c, f.err
}
func (f *fakeCampaigns) List(context.Context) ([]campaign.Summary, error) {
	f.calls++
	return f.list, f.err
}
func (f *fakeCampaigns) Save(_ context.Context, id int64, b json.RawMessage) (campaign.Campaign, error) {
	f.gotID, f.gotBody, f.calls = id, b, f.calls+1
	return f.c, f.err
}
func (f *fakeCampaigns) Approve(_ context.Context, id int64, b json.RawMessage) (campaign.Campaign, bool, error) {
	f.gotID, f.gotBody, f.calls = id, b, f.calls+1
	return f.c, f.repeat, f.err
}
func (f *fakeCampaigns) SampleNames(_ context.Context, id int64) ([]campaign.SampleNames, error) {
	f.gotID, f.calls = id, f.calls+1
	return f.sample, f.err
}

func sampleCampaign() campaign.Campaign {
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	return campaign.Campaign{
		ID: 12, Status: campaign.StatusDraft, Brief: "b", Rule: json.RawMessage(`{"op":"and","conditions":[]}`),
		Description: "d", SkippedLanguages: []string{"en"}, CreatedAt: now, UpdatedAt: now,
		Contents: []campaign.LanguageContent{{Language: "tr", Content: content.Content{SMS: "s", EmailSubject: "e", EmailBody: "b"},
			SMSCount: 980, EmailCount: 1212, FirstNameRunes: 14}},
	}
}

func serveCampaign(f *fakeCampaigns, method, path, body string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	RegisterCampaigns(mux, f)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func errCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var b errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("decode %s: %v", rec.Body, err)
	}
	return b.Error.Code
}

const okBody = `{"contents":[{"language":"tr","sms":"s","email_subject":"e","email_body":"b"}]}`

func TestGetCampaign_Shape(t *testing.T) {
	f := &fakeCampaigns{c: sampleCampaign()}
	rec := serveCampaign(f, "GET", "/api/campaigns/12", "")
	if rec.Code != 200 || f.gotID != 12 {
		t.Fatalf("%d id %d", rec.Code, f.gotID)
	}
	var m map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &m)
	for _, k := range []string{"id", "status", "description", "brief", "skipped_languages", "created_at", "updated_at", "approved_at", "contents"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing key %s in %s", k, rec.Body)
		}
	}
	if _, ok := m["rule"]; ok {
		t.Errorf("rule leaked: %s", rec.Body)
	}
	c := m["contents"].([]any)[0].(map[string]any)
	for _, k := range []string{"language", "sms", "email_subject", "email_body", "sms_count", "email_count", "first_name_runes"} {
		if _, ok := c[k]; !ok {
			t.Errorf("content missing %s", k)
		}
	}
}

func TestCampaigns_BadID(t *testing.T) {
	for _, id := range []string{"abc", "0", "-1", "1.5"} {
		for _, r := range []struct{ method, path, body string }{
			{"GET", "/api/campaigns/" + id, ""},
			{"PUT", "/api/campaigns/" + id + "/contents", okBody},
			{"POST", "/api/campaigns/" + id + "/approve", okBody},
			{"GET", "/api/campaigns/" + id + "/sample", ""},
		} {
			f := &fakeCampaigns{c: sampleCampaign()}
			rec := serveCampaign(f, r.method, r.path, r.body)
			if rec.Code != 400 || errCode(t, rec) != "invalid_request" || f.calls != 0 {
				t.Errorf("%s %s: %d %s calls %d", r.method, r.path, rec.Code, rec.Body, f.calls)
			}
		}
	}
}

func TestCampaigns_BadBody(t *testing.T) {
	bodies := map[string]string{
		"over 64 KiB": `{"contents":[],"pad":"` + strings.Repeat("a", 64*1024) + `"}`,
		"not json":    `{"contents":`,
		"no contents": `{"x":1}`,
		"array":       `[]`,
		"trailing":    okBody + ` x`,
	}
	for name, b := range bodies {
		for _, p := range []struct{ method, path string }{{"PUT", "/api/campaigns/1/contents"}, {"POST", "/api/campaigns/1/approve"}} {
			f := &fakeCampaigns{c: sampleCampaign()}
			rec := serveCampaign(f, p.method, p.path, b)
			if rec.Code != 400 || errCode(t, rec) != "invalid_request" || f.calls != 0 {
				t.Errorf("%s %s: %d calls %d", name, p.method, rec.Code, f.calls)
			}
		}
	}
}

func TestCampaigns_LargeBodyUnderLimit(t *testing.T) {
	f := &fakeCampaigns{c: sampleCampaign()}
	b := `{"contents":[{"language":"tr","sms":"s","email_subject":"e","email_body":"` + strings.Repeat("a", 20*1024) + `"}]}`
	if rec := serveCampaign(f, "PUT", "/api/campaigns/1/contents", b); rec.Code != 200 {
		t.Fatalf("20 KiB body: %d %s", rec.Code, rec.Body)
	}
}

func TestCampaigns_ErrorMapping(t *testing.T) {
	inv := &campaign.InvalidError{Errors: []content.FieldError{{Path: "contents[0].sms", Message: "too long"}}}
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{campaign.ErrNotFound, 404, "not_found"},
		{campaign.ErrNotDraft, 409, "campaign_not_draft"},
		{inv, 422, "content_invalid"},
		{errors.New("pg down"), 503, "db_unavailable"},
		{context.DeadlineExceeded, 503, "db_unavailable"},
	}
	for _, tc := range cases {
		for _, r := range []struct{ method, path, body string }{
			{"GET", "/api/campaigns/1", ""},
			{"PUT", "/api/campaigns/1/contents", okBody},
			{"POST", "/api/campaigns/1/approve", okBody},
			{"GET", "/api/campaigns/1/sample", ""},
		} {
			if tc.err == campaign.ErrNotDraft && r.method != "PUT" {
				continue
			}
			if _, ok := tc.err.(*campaign.InvalidError); ok && r.method == "GET" {
				continue
			}
			rec := serveCampaign(&fakeCampaigns{err: tc.err}, r.method, r.path, r.body)
			if rec.Code != tc.status || errCode(t, rec) != tc.code {
				t.Errorf("%v %s %s: %d %s", tc.err, r.method, r.path, rec.Code, rec.Body)
			}
			if strings.Contains(rec.Body.String(), "pg down") {
				t.Errorf("internal error text leaked: %s", rec.Body)
			}
		}
	}
	rec := serveCampaign(&fakeCampaigns{err: inv}, "POST", "/api/campaigns/1/approve", okBody)
	if !strings.Contains(rec.Body.String(), `"path":"contents[0].sms"`) {
		t.Errorf("details missing: %s", rec.Body)
	}
}

func TestSave_PassesBody(t *testing.T) {
	f := &fakeCampaigns{c: sampleCampaign()}
	rec := serveCampaign(f, "PUT", "/api/campaigns/12/contents", okBody)
	if rec.Code != 200 || f.gotID != 12 || !strings.Contains(string(f.gotBody), `"contents"`) {
		t.Fatalf("%d %d %s", rec.Code, f.gotID, f.gotBody)
	}
}

func TestApprove_Repeat(t *testing.T) {
	c := sampleCampaign()
	c.Status = campaign.StatusApproved
	for _, repeat := range []bool{false, true} {
		f := &fakeCampaigns{c: c, repeat: repeat}
		rec := serveCampaign(f, "POST", "/api/campaigns/12/approve", okBody)
		if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"status":"approved"`) {
			t.Fatalf("repeat %v: %d %s", repeat, rec.Code, rec.Body)
		}
	}
}

func TestList_NoContentText(t *testing.T) {
	f := &fakeCampaigns{list: []campaign.Summary{{ID: 1, Status: campaign.StatusDraft, Description: "d", Languages: []string{"tr"}}}}
	rec := serveCampaign(f, "GET", "/api/campaigns", "")
	if rec.Code != 200 || !strings.HasPrefix(rec.Body.String(), `{"campaigns":[`) || strings.Contains(rec.Body.String(), "sms") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	rec = serveCampaign(&fakeCampaigns{list: []campaign.Summary{}}, "GET", "/api/campaigns", "")
	if !strings.Contains(rec.Body.String(), `"campaigns":[]`) {
		t.Fatalf("empty list: %s", rec.Body)
	}
}

func TestSample_Shape(t *testing.T) {
	f := &fakeCampaigns{sample: []campaign.SampleNames{{Code: "tr", FirstNames: []string{"Abdurrahmangazi", "Ayşe"}}}}
	rec := serveCampaign(f, "GET", "/api/campaigns/3/sample", "")
	if rec.Code != 200 || f.gotID != 3 || !strings.Contains(rec.Body.String(), `{"languages":[{"code":"tr","first_names":["Abdurrahmangazi","Ayşe"]}]}`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}
