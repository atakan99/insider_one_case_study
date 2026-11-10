package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"campaigncomposer/internal/campaign"
)

// maxCampaignBodyBytes limits the body of a save or an approval.
const maxCampaignBodyBytes = 64 * 1024

// CampaignService is the part of campaign.Service that the handlers use.
type CampaignService interface {
	Get(ctx context.Context, id int64) (campaign.Campaign, error)
	List(ctx context.Context) ([]campaign.Summary, error)
	Save(ctx context.Context, id int64, body json.RawMessage) (campaign.Campaign, error)
	Approve(ctx context.Context, id int64, body json.RawMessage) (campaign.Campaign, bool, error)
	SampleNames(ctx context.Context, id int64) ([]campaign.SampleNames, error)
}

type campaignHandlers struct{ svc CampaignService }

// RegisterCampaigns adds the campaign routes.
func RegisterCampaigns(mux *http.ServeMux, svc CampaignService) {
	h := campaignHandlers{svc: svc}
	mux.HandleFunc("GET /api/campaigns", h.list)
	mux.HandleFunc("GET /api/campaigns/{id}", h.get)
	mux.HandleFunc("PUT /api/campaigns/{id}/contents", h.save)
	mux.HandleFunc("POST /api/campaigns/{id}/approve", h.approve)
	mux.HandleFunc("GET /api/campaigns/{id}/sample", h.sample)
}

// campaignID reads the id path value. It writes 400 and returns false when the
// id is not a positive integer.
func campaignID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_request", "The campaign id must be a positive integer.", nil)
		return 0, false
	}
	return id, true
}

// contentsBody reads a save or approval body. It writes 400 and returns false
// when the body is not a JSON object with a contents key.
func contentsBody(w http.ResponseWriter, r *http.Request) (json.RawMessage, bool) {
	var raw json.RawMessage
	var m map[string]json.RawMessage
	if err := decodeBodyLimit(w, r, &raw, maxCampaignBodyBytes); err != nil ||
		json.Unmarshal(raw, &m) != nil || m == nil || isAbsentJSON(m["contents"]) {
		writeError(w, http.StatusBadRequest, "invalid_request", "The request body must be JSON with contents, of at most 64 KiB.", nil)
		return nil, false
	}
	return raw, true
}

// writeCampaignError maps a service error to the response.
func writeCampaignError(w http.ResponseWriter, r *http.Request, op string, err error) {
	var inv *campaign.InvalidError
	switch {
	case errors.Is(err, campaign.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "No campaign has this id.", nil)
	case errors.Is(err, campaign.ErrNotDraft):
		writeError(w, http.StatusConflict, "campaign_not_draft", "Only a draft can change. This campaign is already approved.", nil)
	case errors.As(err, &inv):
		writeError(w, http.StatusUnprocessableEntity, "content_invalid", "Some messages are not valid. Fix them and try again.", inv.Errors)
	default:
		slog.ErrorContext(r.Context(), "campaign "+op+" failed", "err", err)
		writeError(w, http.StatusServiceUnavailable, "db_unavailable", "The campaign could not be loaded or saved. Try again later.", nil)
	}
}

func (h campaignHandlers) list(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context())
	if err != nil {
		writeCampaignError(w, r, "list", err)
		return
	}
	if list == nil {
		list = []campaign.Summary{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"campaigns": list})
}

func (h campaignHandlers) get(w http.ResponseWriter, r *http.Request) {
	id, ok := campaignID(w, r)
	if !ok {
		return
	}
	c, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeCampaignError(w, r, "get", err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func logSave(r *http.Request, op string, c campaign.Campaign, repeat bool) {
	langs := make([]string, 0, len(c.Contents))
	var sms, email int64
	for _, lc := range c.Contents {
		langs = append(langs, lc.Language)
		sms += lc.SMSCount
		email += lc.EmailCount
	}
	slog.InfoContext(r.Context(), "campaign "+op, "campaign_id", c.ID, "languages", langs,
		"sms_count", sms, "email_count", email, "repeat", repeat)
}

func (h campaignHandlers) save(w http.ResponseWriter, r *http.Request) {
	id, ok := campaignID(w, r)
	if !ok {
		return
	}
	body, ok := contentsBody(w, r)
	if !ok {
		return
	}
	c, err := h.svc.Save(r.Context(), id, body)
	if err != nil {
		writeCampaignError(w, r, "save", err)
		return
	}
	logSave(r, "save", c, false)
	writeJSON(w, http.StatusOK, c)
}

func (h campaignHandlers) approve(w http.ResponseWriter, r *http.Request) {
	id, ok := campaignID(w, r)
	if !ok {
		return
	}
	body, ok := contentsBody(w, r)
	if !ok {
		return
	}
	c, repeat, err := h.svc.Approve(r.Context(), id, body)
	if err != nil {
		writeCampaignError(w, r, "approve", err)
		return
	}
	logSave(r, "approve", c, repeat)
	writeJSON(w, http.StatusOK, c)
}

func (h campaignHandlers) sample(w http.ResponseWriter, r *http.Request) {
	id, ok := campaignID(w, r)
	if !ok {
		return
	}
	names, err := h.svc.SampleNames(r.Context(), id)
	if err != nil {
		writeCampaignError(w, r, "sample", err)
		return
	}
	if names == nil {
		names = []campaign.SampleNames{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"languages": names})
}
