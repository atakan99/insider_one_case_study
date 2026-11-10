package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
	"unicode/utf8"

	"campaigncomposer/internal/llm"
	"campaigncomposer/internal/segment"
)

const (
	maxBodyBytes = 16 * 1024
	countTimeout = 2 * time.Second
)

// AudienceCounter counts the audience of a compiled rule.
type AudienceCounter func(ctx context.Context, w segment.Where) (segment.Audience, error)

// SegmentPreviewer serves POST /api/segments/preview.
type SegmentPreviewer struct {
	LLM        llm.Client
	Allow      segment.Allowlist
	Count      AudienceCounter
	Now        func() time.Time
	LLMTimeout time.Duration
}

// RegisterSegments adds POST /api/segments/preview.
func RegisterSegments(mux *http.ServeMux, p *SegmentPreviewer) {
	mux.HandleFunc("POST /api/segments/preview", p.preview)
}

type previewResponse struct {
	Rule        segment.Rule     `json:"rule"`
	Description string           `json:"description"`
	Unmapped    []string         `json:"unmapped"`
	Audience    segment.Audience `json:"audience"`
}

const noToolCallText = "You did not call " + segment.ToolName + ". Do not reply with text. Call " + segment.ToolName + " with your answer.\n"

func (p *SegmentPreviewer) preview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	var in struct {
		Brief string `json:"brief"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Brief == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "The request body must be JSON with a brief of at most 16 KiB.", nil)
		return
	}
	brief, err := segment.SanitizeBrief(in.Brief)
	if err != nil {
		msg := "The brief must contain at least one letter."
		if errors.Is(err, segment.ErrBriefTooLong) {
			msg = "The brief must have at most 1000 characters."
		}
		writeError(w, http.StatusBadRequest, "invalid_brief", msg, nil)
		return
	}
	briefLen := utf8.RuneCountInString(brief)

	req := llm.Request{
		System: p.Allow.SystemPrompt(),
		Tool:   llm.Tool{Name: segment.ToolName, Description: "Submit the segment rule for the brief.", Schema: p.Allow.ToolSchema()},
		Turns:  []llm.Turn{{User: segment.UserMessage(brief)}},
	}

	var ans segment.Answer
	var errs []segment.FieldError
	attempts := 0
	for attempts < 2 {
		attempts++
		resp, err := p.generate(ctx, req)
		if err != nil {
			slog.WarnContext(ctx, "segment preview: llm call failed", "brief_runes", briefLen, "attempts", attempts, "err", err)
			if errors.Is(err, context.DeadlineExceeded) {
				writeError(w, http.StatusGatewayTimeout, "llm_timeout", "The language model took too long to answer.", nil)
			} else {
				writeError(w, http.StatusBadGateway, "llm_error", "The language model call failed.", nil)
			}
			return
		}
		if resp.Refused {
			slog.WarnContext(ctx, "segment preview: llm refused", "brief_runes", briefLen, "attempts", attempts)
			writeError(w, http.StatusBadGateway, "llm_refused", "The language model refused this brief.", nil)
			return
		}
		if resp.ToolUse != nil && resp.ToolUse.Name == segment.ToolName {
			ans, errs = p.Allow.DecodeAnswer(resp.ToolUse.Input)
			if len(errs) == 0 {
				break
			}
			req.Turns = append(req.Turns,
				llm.Turn{Assistant: resp.Assistant},
				llm.Turn{ToolResult: &llm.ToolResult{ToolUseID: resp.ToolUse.ID, Content: segment.RepairMessage(errs), IsError: true}})
		} else {
			errs = []segment.FieldError{{Path: "", Message: "no submit_segment call"}}
			req.Turns = append(req.Turns,
				llm.Turn{Assistant: resp.Assistant},
				llm.Turn{User: noToolCallText + segment.RepairMessage(errs)})
		}
		slog.InfoContext(ctx, "segment preview: validation failed", "brief_runes", briefLen, "attempts", attempts, "errors", errs)
	}
	if len(errs) > 0 {
		writeError(w, http.StatusUnprocessableEntity, "segment_invalid", "The language model could not make a valid segment rule. Try rewording the brief.", errs)
		return
	}
	if ans.Status == "unsupported" {
		slog.InfoContext(ctx, "segment preview: unsupported", "brief_runes", briefLen, "attempts", attempts)
		writeError(w, http.StatusUnprocessableEntity, "segment_unsupported", ans.Reason, nil)
		return
	}

	cctx, cancel := context.WithTimeout(ctx, countTimeout)
	defer cancel()
	aud, err := p.Count(cctx, p.Allow.Compile(ans.Rule, p.Now()))
	if err != nil {
		slog.ErrorContext(ctx, "segment preview: audience count failed", "brief_runes", briefLen, "attempts", attempts, "err", err)
		writeError(w, http.StatusServiceUnavailable, "db_unavailable", "The audience could not be counted. Try again later.", nil)
		return
	}
	if aud.Languages == nil {
		aud.Languages = []segment.LanguageCounts{}
	}
	unmapped := ans.Unmapped
	if unmapped == nil {
		unmapped = []string{}
	}
	ruleJSON, _ := ans.Rule.MarshalJSON()
	slog.InfoContext(ctx, "segment preview", "brief_runes", briefLen, "attempts", attempts,
		"rule", string(ruleJSON), "matched", aud.Matched, "reachable", aud.Reachable,
		"email", aud.Email, "sms", aud.SMS, "no_consent", aud.NoConsent)
	writeJSON(w, http.StatusOK, previewResponse{
		Rule: ans.Rule, Description: p.Allow.Describe(ans.Rule), Unmapped: unmapped, Audience: aud,
	})
}

// generate runs one LLM call with its own timeout.
func (p *SegmentPreviewer) generate(ctx context.Context, req llm.Request) (llm.Response, error) {
	cctx, cancel := context.WithTimeout(ctx, p.LLMTimeout)
	defer cancel()
	return p.LLM.Generate(cctx, req)
}
