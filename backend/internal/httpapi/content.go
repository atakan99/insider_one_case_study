package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/llm"
	"campaigncomposer/internal/segment"
)

// ContentGenerator serves POST /api/content/generate. It streams the content
// of each audience language as Server-Sent Events.
type ContentGenerator struct {
	LLM        llm.Client
	Allow      segment.Allowlist
	Count      AudienceCounter
	Now        func() time.Time // rule clock
	Clock      func() time.Time // draft throttle clock; tests inject it
	LLMTimeout time.Duration    // timeout of each LLM call
	DraftEvery time.Duration    // minimum time between two drafts of a language
}

// RegisterContent adds POST /api/content/generate.
func RegisterContent(mux *http.ServeMux, g *ContentGenerator) {
	mux.HandleFunc("POST /api/content/generate", g.generate)
}

type audienceEvent struct {
	Description string                   `json:"description"`
	Languages   []segment.LanguageCounts `json:"languages"`
}

type draftEvent struct {
	Language string          `json:"language"`
	Content  content.Content `json:"content"`
}

type retryEvent struct {
	Language string               `json:"language"`
	Errors   []content.FieldError `json:"errors,omitempty"`
}

type finalEvent struct {
	Language string          `json:"language"`
	Content  content.Content `json:"content"`
	Limits   content.Limits  `json:"limits"`
}

type languageErrorEvent struct {
	Language string               `json:"language"`
	Code     string               `json:"code"`
	Message  string               `json:"message"`
	Details  []content.FieldError `json:"details,omitempty"`
}

const noContentToolText = "You did not call " + content.ToolName + ". Do not reply with text. Call " + content.ToolName + " with your answer.\n"

func (g *ContentGenerator) generate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in struct {
		Brief string          `json:"brief"`
		Rule  json.RawMessage `json:"rule"`
	}
	if err := decodeBody(w, r, &in); err != nil || in.Brief == "" || isAbsentJSON(in.Rule) {
		writeError(w, http.StatusBadRequest, "invalid_request", "The request body must be JSON with a brief and a rule, of at most 16 KiB.", nil)
		return
	}
	brief, err := segment.SanitizeBrief(in.Brief)
	if err != nil {
		writeBriefError(w, err)
		return
	}
	rule, ruleErrs := g.Allow.ParseRule(in.Rule)
	if len(ruleErrs) > 0 {
		writeError(w, http.StatusBadRequest, "invalid_rule", "The segment rule is not valid.", ruleErrs)
		return
	}

	cctx, cancel := context.WithTimeout(ctx, countTimeout)
	aud, err := g.Count(cctx, g.Allow.Compile(rule, g.now()))
	cancel()
	if err != nil {
		slog.ErrorContext(ctx, "content generate: audience count failed", "err", err)
		writeError(w, http.StatusServiceUnavailable, "db_unavailable", "The audience could not be counted. Try again later.", nil)
		return
	}
	langs := make([]segment.LanguageCounts, 0, len(aud.Languages))
	for _, l := range aud.Languages {
		if l.Reachable > 0 {
			langs = append(langs, l)
		}
	}
	if len(langs) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "empty_audience", "No user in this audience can get a message.", nil)
		return
	}

	sse, ok := newSSEWriter(w)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "The server cannot stream the response.", nil)
		return
	}
	description := g.Allow.Describe(rule)
	_ = sse.send("audience", audienceEvent{Description: description, Languages: langs})

	// One worker for each language. A failure in one language does not cancel
	// the others; only the end of the request context cancels them all.
	var wg sync.WaitGroup
	for _, l := range langs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.runLanguage(ctx, sse, l, description, brief)
		}()
	}
	wg.Wait()
	if ctx.Err() != nil {
		slog.InfoContext(ctx, "content generate: client disconnected", "languages", len(langs))
		return
	}
	_ = sse.send("done", struct{}{})
}

// runLanguage streams the content of one language, with one repair round,
// and sends its final or error event.
func (g *ContentGenerator) runLanguage(ctx context.Context, sse *sseWriter, lang segment.LanguageCounts, description, brief string) {
	code := lang.Code
	firstName := int(lang.MaxFirstName)
	req := llm.Request{
		System: content.SystemPrompt(),
		Tool: llm.Tool{
			Name: content.ToolName, Description: "Submit the SMS, the email subject and the email body for the language.",
			Schema: content.ToolSchema(), Eager: true,
		},
		Turns: []llm.Turn{{User: content.UserMessage(code, description, firstName, brief)}},
	}
	d := &drafter{sse: sse, lang: code, clock: g.Clock, every: g.DraftEvery}
	if d.clock == nil {
		d.clock = time.Now
	}
	fail := func(attempts int, ec, msg string, errs []content.FieldError, err error) {
		slog.WarnContext(ctx, "content generate", "language", code, "attempts", attempts,
			"outcome", ec, "errors", errs, "first_name_runes", firstName, "err", err)
		_ = sse.send("error", languageErrorEvent{Language: code, Code: ec, Message: msg, Details: errs})
	}

	for attempts := 1; ; attempts++ {
		d.start()
		resp, err := g.stream(ctx, req, d.on)
		if err != nil {
			if ctx.Err() != nil {
				slog.InfoContext(ctx, "content generate: cancelled", "language", code, "attempts", attempts)
				return
			}
			if errors.Is(err, context.DeadlineExceeded) {
				fail(attempts, "llm_timeout", "The language model took too long to answer.", nil, err)
			} else {
				fail(attempts, "llm_error", "The language model call failed.", nil, err)
			}
			return
		}
		if resp.Refused {
			fail(attempts, "llm_refused", "The language model refused this brief.", nil, nil)
			return
		}
		d.flush()

		toolCall := resp.ToolUse != nil && resp.ToolUse.Name == content.ToolName
		var errs []content.FieldError
		if resp.CutOff {
			errs = append(errs, content.FieldError{Path: "", Message: "answer was cut off"})
		}
		var c content.Content
		var limits content.Limits
		if toolCall {
			var verrs []content.FieldError
			c, limits, verrs = content.Validate(resp.ToolUse.Input, firstName)
			errs = append(errs, verrs...)
		} else {
			errs = append(errs, content.FieldError{Path: "", Message: "no " + content.ToolName + " call"})
		}

		if len(errs) == 0 {
			slog.InfoContext(ctx, "content generate", "language", code, "attempts", attempts, "outcome", "final",
				"first_name_runes", limits.FirstNameRunes, "sms_runes", limits.SMSRunes,
				"email_subject_runes", limits.EmailSubjectRunes, "email_body_runes", limits.EmailBodyRunes)
			_ = sse.send("final", finalEvent{Language: code, Content: c, Limits: limits})
			return
		}
		if attempts >= 2 {
			fail(attempts, "content_invalid", "The language model could not write valid content for this language. Try rewording the brief.", errs, nil)
			return
		}
		slog.InfoContext(ctx, "content generate: validation failed", "language", code, "attempts", attempts, "errors", errs)
		_ = sse.send("retry", retryEvent{Language: code, Errors: errs})

		// A tool input that is not a JSON object is replayed as {} by the
		// client, and the replayed turn has no tool_use other than ToolUse, so
		// the assistant turn can always go back as it is.
		if toolCall {
			req.Turns = append(req.Turns,
				llm.Turn{Assistant: resp.Assistant},
				llm.Turn{ToolResult: &llm.ToolResult{ToolUseID: resp.ToolUse.ID, Content: content.RepairMessage(errs), IsError: true}})
		} else {
			req.Turns = append(req.Turns,
				llm.Turn{Assistant: resp.Assistant},
				llm.Turn{User: noContentToolText + content.RepairMessage(errs)})
		}
	}
}

// stream runs one LLM call with its own timeout.
func (g *ContentGenerator) stream(ctx context.Context, req llm.Request, on func(llm.StreamEvent)) (llm.Response, error) {
	cctx, cancel := context.WithTimeout(ctx, g.LLMTimeout)
	defer cancel()
	return g.LLM.Stream(cctx, req, on)
}

func (g *ContentGenerator) now() time.Time {
	if g.Now == nil {
		return time.Now()
	}
	return g.Now()
}

// isAbsentJSON reports a missing or null JSON value.
func isAbsentJSON(raw json.RawMessage) bool {
	t := bytes.TrimSpace(raw)
	return len(t) == 0 || string(t) == "null"
}

// drafter keeps the tool input fragments of one call and sends throttled
// draft events. It is used from one goroutine only.
type drafter struct {
	sse   *sseWriter
	lang  string
	clock func() time.Time
	every time.Duration

	buf   []byte
	dirty bool      // buf has fragments that no draft has sent
	last  time.Time // time of the last draft; zero before the first
}

// start clears the buffer for a new call.
func (d *drafter) start() {
	d.buf = d.buf[:0]
	d.dirty = false
}

// on handles one stream event.
func (d *drafter) on(ev llm.StreamEvent) {
	if ev.Reset {
		// A fallback block started; the text so far is not valid. This is
		// not the repair round.
		d.start()
		_ = d.sse.send("retry", retryEvent{Language: d.lang})
		return
	}
	if ev.Delta == "" {
		return
	}
	d.buf = append(d.buf, ev.Delta...)
	d.dirty = true
	now := d.clock()
	if d.last.IsZero() || now.Sub(d.last) >= d.every {
		d.send(now)
	}
}

// flush sends a draft with the last text if the last draft did not have it.
func (d *drafter) flush() {
	if d.dirty {
		d.send(d.clock())
	}
}

func (d *drafter) send(now time.Time) {
	_ = d.sse.send("draft", draftEvent{Language: d.lang, Content: content.PartialContent(d.buf)})
	d.last = now
	d.dirty = false
}
