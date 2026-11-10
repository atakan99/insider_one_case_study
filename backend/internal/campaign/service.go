package campaign

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"campaigncomposer/internal/content"
	"campaigncomposer/internal/segment"
)

// countTimeout limits the audience query of a save or an approval.
const countTimeout = 2 * time.Second

// listLimit is the size of the campaign list.
const listLimit = 100

// sampleSize is the number of sample first names for each language.
const sampleSize = 5

// AudienceCounter counts the audience of a compiled rule.
type AudienceCounter func(ctx context.Context, w segment.Where) (segment.Audience, error)

// Sampler returns sample first names of a compiled rule.
type Sampler func(ctx context.Context, w segment.Where, langs []string, n int) (map[string][]string, error)

// SampleNames are the preview names of one language.
type SampleNames struct {
	Code       string   `json:"code"`
	FirstNames []string `json:"first_names"`
}

// Service joins the store, the rule validator and the audience query.
type Service struct {
	Store  *Store
	Allow  segment.Allowlist
	Count  AudienceCounter
	Sample Sampler
	Now    func() time.Time
}

func (s *Service) now() time.Time {
	if s.Now == nil {
		return time.Now()
	}
	return s.Now()
}

// Get returns campaign id.
func (s *Service) Get(ctx context.Context, id int64) (Campaign, error) { return s.Store.Get(ctx, id) }

// List returns the newest campaigns.
func (s *Service) List(ctx context.Context) ([]Summary, error) { return s.Store.List(ctx, listLimit) }

// Save checks and stores the edits of draft id.
func (s *Service) Save(ctx context.Context, id int64, body json.RawMessage) (Campaign, error) {
	c, _, err := s.update(ctx, id, false, body)
	return c, err
}

// Approve checks and stores the last edits of draft id and approves it. A
// campaign that is already approved comes back with repeat = true, unchanged.
func (s *Service) Approve(ctx context.Context, id int64, body json.RawMessage) (Campaign, bool, error) {
	return s.update(ctx, id, true, body)
}

// compile parses the stored rule again and compiles it.
func (s *Service) compile(c Campaign) (segment.Where, error) {
	r, errs := s.Allow.ParseRule(c.Rule)
	if len(errs) > 0 {
		fe := make([]content.FieldError, 0, len(errs))
		for _, e := range errs {
			fe = append(fe, content.FieldError{Path: "rule", Message: e.Error()})
		}
		return segment.Where{}, &InvalidError{Errors: fe}
	}
	return s.Allow.Compile(r, s.now()), nil
}

// update counts the audience first, then locks the row for the check and the
// write only. Thus the row lock never waits for the audience query, and a
// second connection is not needed while the lock is held. The stored rule of a
// draft never changes (ReplaceDraft keeps it), so the count stays valid.
func (s *Service) update(ctx context.Context, id int64, approve bool, body json.RawMessage) (Campaign, bool, error) {
	c, err := s.Store.Get(ctx, id)
	if err != nil {
		return Campaign{}, false, err
	}
	if c.Status != StatusDraft {
		// A repeat approval or ErrNotDraft; Update decides without a check.
		return s.Store.Update(ctx, id, approve, func(Campaign) ([]LanguageContent, error) { return nil, ErrNotDraft })
	}
	w, err := s.compile(c)
	if err != nil {
		return Campaign{}, false, err
	}
	cctx, cancel := context.WithTimeout(ctx, countTimeout)
	aud, err := s.Count(cctx, w)
	cancel()
	if err != nil {
		return Campaign{}, false, err
	}
	byLang := map[string]segment.LanguageCounts{}
	for _, l := range aud.Languages {
		byLang[l.Code] = l
	}
	out, repeat, err := s.Store.Update(ctx, id, approve, func(locked Campaign) ([]LanguageContent, error) {
		langs := make([]string, 0, len(locked.Contents))
		for _, lc := range locked.Contents {
			langs = append(langs, lc.Language)
		}
		contents, errs := CheckContents(body, langs, byLang)
		if len(errs) > 0 {
			return nil, &InvalidError{Errors: errs}
		}
		return contents, nil
	})
	var inv *InvalidError
	if errors.As(err, &inv) {
		// Keep the new audience snapshot, so the review window shows the limit
		// that failed. The content stays as it was.
		_ = s.Store.UpdateCounts(ctx, id, byLang)
	}
	return out, repeat, err
}

// SampleNames returns up to 5 preview first names for each language of
// campaign id. It uses the stored rule.
func (s *Service) SampleNames(ctx context.Context, id int64) ([]SampleNames, error) {
	c, err := s.Store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	w, err := s.compile(c)
	if err != nil {
		return nil, err
	}
	langs := make([]string, 0, len(c.Contents))
	for _, lc := range c.Contents {
		langs = append(langs, lc.Language)
	}
	cctx, cancel := context.WithTimeout(ctx, countTimeout)
	defer cancel()
	names, err := s.Sample(cctx, w, langs, sampleSize)
	if err != nil {
		return nil, err
	}
	out := make([]SampleNames, 0, len(langs))
	for _, l := range langs {
		n := names[l]
		if n == nil {
			n = []string{}
		}
		out = append(out, SampleNames{Code: l, FirstNames: n})
	}
	return out, nil
}
