package campaign

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"campaigncomposer/internal/segment"
)

// statementTimeout limits each query, as in the segment package.
const statementTimeout = "2s"

// Store reads and writes campaigns in Postgres.
type Store struct{ pool *pgxpool.Pool }

// NewStore returns a Store on pool.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) inTx(ctx context.Context, mode pgx.TxAccessMode, fn func(pgx.Tx) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: mode})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = '"+statementTimeout+"'"); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func skipped(d Draft) []string {
	if d.Skipped == nil {
		return []string{}
	}
	return d.Skipped
}

func insertContents(ctx context.Context, tx pgx.Tx, id int64, cs []LanguageContent) error {
	for _, c := range cs {
		if _, err := tx.Exec(ctx, `INSERT INTO campaign_contents
(campaign_id, language, sms, email_subject, email_body, sms_count, email_count, first_name_runes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			id, c.Language, c.SMS, c.EmailSubject, c.EmailBody, c.SMSCount, c.EmailCount, c.FirstNameRunes); err != nil {
			return err
		}
	}
	return nil
}

// CreateDraft inserts a draft and its contents in one transaction.
func (s *Store) CreateDraft(ctx context.Context, d Draft) (int64, error) {
	var id int64
	err := s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `INSERT INTO campaigns
(brief, rule, description, status, skipped_languages, created_at, updated_at)
VALUES ($1, $2, $3, 'draft', $4, now(), now()) RETURNING id`,
			d.Brief, string(d.Rule), d.Description, skipped(d)).Scan(&id); err != nil {
			return err
		}
		return insertContents(ctx, tx, id, d.Contents)
	})
	if err != nil {
		return 0, fmt.Errorf("create draft: %w", err)
	}
	return id, nil
}

// ReplaceDraft replaces the contents, counts and skipped languages of draft id.
// It returns ErrConflict when the campaign is not a draft, or when its brief or
// rule differs from d (rule compared as jsonb). ErrNotFound when id is unknown.
func (s *Store) ReplaceDraft(ctx context.Context, id int64, d Draft) error {
	return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		var st string
		var same bool
		err := tx.QueryRow(ctx, `SELECT status, brief = $2 AND rule = $3::jsonb FROM campaigns WHERE id = $1 FOR UPDATE`,
			id, d.Brief, string(d.Rule)).Scan(&st, &same)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("replace draft: %w", err)
		}
		if st != string(StatusDraft) || !same {
			return ErrConflict
		}
		if _, err := tx.Exec(ctx, `UPDATE campaigns SET description = $2, skipped_languages = $3, updated_at = now() WHERE id = $1`,
			id, d.Description, skipped(d)); err != nil {
			return fmt.Errorf("replace draft: %w", err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM campaign_contents WHERE campaign_id = $1`, id); err != nil {
			return fmt.Errorf("replace draft: %w", err)
		}
		if err := insertContents(ctx, tx, id, d.Contents); err != nil {
			return fmt.Errorf("replace draft: %w", err)
		}
		return nil
	})
}

const selectCampaign = `SELECT id, status, brief, rule, description, skipped_languages, created_at, updated_at, approved_at
FROM campaigns WHERE id = $1`

func getCampaign(ctx context.Context, tx pgx.Tx, id int64, lock bool) (Campaign, error) {
	q := selectCampaign
	if lock {
		q += " FOR UPDATE"
	}
	var c Campaign
	var rule []byte
	err := tx.QueryRow(ctx, q, id).Scan(&c.ID, &c.Status, &c.Brief, &rule, &c.Description,
		&c.SkippedLanguages, &c.CreatedAt, &c.UpdatedAt, &c.ApprovedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, ErrNotFound
	}
	if err != nil {
		return Campaign{}, err
	}
	c.Rule = rule
	if c.SkippedLanguages == nil {
		c.SkippedLanguages = []string{}
	}
	rs, err := tx.Query(ctx, `SELECT language, sms, email_subject, email_body, sms_count, email_count, first_name_runes
FROM campaign_contents WHERE campaign_id = $1 ORDER BY language`, id)
	if err != nil {
		return Campaign{}, err
	}
	defer rs.Close()
	c.Contents = []LanguageContent{}
	for rs.Next() {
		var l LanguageContent
		if err := rs.Scan(&l.Language, &l.SMS, &l.EmailSubject, &l.EmailBody, &l.SMSCount, &l.EmailCount, &l.FirstNameRunes); err != nil {
			return Campaign{}, err
		}
		c.Contents = append(c.Contents, l)
	}
	return c, rs.Err()
}

// Get returns campaign id. ErrNotFound when id is unknown.
func (s *Store) Get(ctx context.Context, id int64) (Campaign, error) {
	var c Campaign
	err := s.inTx(ctx, pgx.ReadOnly, func(tx pgx.Tx) error {
		var err error
		c, err = getCampaign(ctx, tx, id, false)
		return err
	})
	if err != nil && !errors.Is(err, ErrNotFound) {
		err = fmt.Errorf("get campaign: %w", err)
	}
	return c, err
}

// List returns the newest campaigns first, at most limit.
func (s *Store) List(ctx context.Context, limit int) ([]Summary, error) {
	out := []Summary{}
	err := s.inTx(ctx, pgx.ReadOnly, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `SELECT c.id, c.status, c.description, c.created_at, c.approved_at,
       coalesce(array_agg(cc.language ORDER BY cc.language) FILTER (WHERE cc.language IS NOT NULL), '{}')
FROM campaigns c LEFT JOIN campaign_contents cc ON cc.campaign_id = c.id
GROUP BY c.id
ORDER BY c.created_at DESC, c.id DESC
LIMIT $1`, limit)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var sm Summary
			if err := rs.Scan(&sm.ID, &sm.Status, &sm.Description, &sm.CreatedAt, &sm.ApprovedAt, &sm.Languages); err != nil {
				return err
			}
			out = append(out, sm)
		}
		return rs.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list campaigns: %w", err)
	}
	return out, nil
}

// Update locks the campaign row (SELECT ... FOR UPDATE) and calls check with it.
// When the status is not draft: if approve is true, it returns the campaign and
// repeat = true without calling check; if approve is false, it returns ErrNotDraft.
// Otherwise it writes the contents that check returns, sets updated_at, and,
// when approve is true, sets status approved and approved_at, all in one
// transaction. An error from check rolls back and is returned as it is.
func (s *Store) Update(ctx context.Context, id int64, approve bool,
	check func(Campaign) ([]LanguageContent, error)) (Campaign, bool, error) {
	var out Campaign
	var repeat bool
	var checkErr error
	err := s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		c, err := getCampaign(ctx, tx, id, true)
		if err != nil {
			return err
		}
		if c.Status != StatusDraft {
			if !approve {
				return ErrNotDraft
			}
			out, repeat = c, true
			return nil
		}
		contents, err := check(c)
		if err != nil {
			checkErr = err
			return err
		}
		sort.Slice(contents, func(i, j int) bool { return contents[i].Language < contents[j].Language })
		if _, err := tx.Exec(ctx, `DELETE FROM campaign_contents WHERE campaign_id = $1`, id); err != nil {
			return err
		}
		if err := insertContents(ctx, tx, id, contents); err != nil {
			return err
		}
		if approve {
			_, err = tx.Exec(ctx, `UPDATE campaigns SET status = 'approved', approved_at = now(), updated_at = now() WHERE id = $1`, id)
		} else {
			_, err = tx.Exec(ctx, `UPDATE campaigns SET updated_at = now() WHERE id = $1`, id)
		}
		if err != nil {
			return err
		}
		out, err = getCampaign(ctx, tx, id, false)
		return err
	})
	switch {
	case err == nil:
		return out, repeat, nil
	case checkErr != nil, errors.Is(err, ErrNotFound), errors.Is(err, ErrNotDraft):
		return Campaign{}, false, err
	default:
		return Campaign{}, false, fmt.Errorf("update campaign: %w", err)
	}
}

// UpdateCounts stores the audience snapshot of a draft: the counts and the
// longest first name of each language. A language missing from aud gets 0.
// It changes nothing when the campaign is not a draft.
func (s *Store) UpdateCounts(ctx context.Context, id int64, aud map[string]segment.LanguageCounts) error {
	return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
		var st string
		err := tx.QueryRow(ctx, `SELECT status FROM campaigns WHERE id = $1 FOR UPDATE`, id).Scan(&st)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil || st != string(StatusDraft) {
			return err
		}
		rs, err := tx.Query(ctx, `SELECT language FROM campaign_contents WHERE campaign_id = $1`, id)
		if err != nil {
			return err
		}
		langs, err := pgx.CollectRows(rs, pgx.RowTo[string])
		if err != nil {
			return err
		}
		for _, l := range langs {
			c := aud[l]
			if _, err := tx.Exec(ctx, `UPDATE campaign_contents SET sms_count = $3, email_count = $4, first_name_runes = $5
WHERE campaign_id = $1 AND language = $2`, id, l, c.SMS, c.Email, c.MaxFirstName); err != nil {
				return err
			}
		}
		return nil
	})
}
