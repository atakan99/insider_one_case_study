package segment

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// statementTimeout limits each query. Tests change it.
var statementTimeout = "2s"

// Counts holds the audience numbers for one scope.
type Counts struct {
	Email     int64 `json:"email"`
	SMS       int64 `json:"sms"`
	Reachable int64 `json:"reachable"`
	Matched   int64 `json:"matched"`
	NoConsent int64 `json:"no_consent"`
}

// LanguageCounts holds the counts for one language.
type LanguageCounts struct {
	Code string `json:"code"`
	Counts
}

// Audience holds the totals and the counts for each language.
type Audience struct {
	Counts
	Languages []LanguageCounts `json:"languages"`
}

// buildAudience drops the languages with no reachable user, sorts the rest
// and adds the totals. The totals include the dropped languages.
func buildAudience(rows []LanguageCounts) Audience {
	a := Audience{Languages: []LanguageCounts{}}
	for _, r := range rows {
		a.Email += r.Email
		a.SMS += r.SMS
		a.Reachable += r.Reachable
		a.Matched += r.Matched
		if r.Reachable > 0 {
			a.Languages = append(a.Languages, r)
		}
	}
	a.NoConsent = a.Matched - a.Reachable
	sort.Slice(a.Languages, func(i, j int) bool {
		x, y := a.Languages[i], a.Languages[j]
		if x.Reachable != y.Reachable {
			return x.Reachable > y.Reachable
		}
		return x.Code < y.Code
	})
	return a
}

// inTx runs fn in a read-only transaction with the statement timeout.
func inTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
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

// CountAudience counts the users that match w, by language and channel consent.
func CountAudience(ctx context.Context, pool *pgxpool.Pool, w Where) (Audience, error) {
	if w.sql == "" {
		return Audience{}, ErrEmptyWhere
	}
	q := `SELECT language,
       count(*) FILTER (WHERE email_consent),
       count(*) FILTER (WHERE sms_consent),
       count(*) FILTER (WHERE email_consent OR sms_consent),
       count(*)
FROM users
WHERE (` + w.SQL() + `)
GROUP BY language`
	var rows []LanguageCounts
	err := inTx(ctx, pool, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, q, w.Args()...)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var l LanguageCounts
			if err := rs.Scan(&l.Code, &l.Email, &l.SMS, &l.Reachable, &l.Matched); err != nil {
				return err
			}
			l.NoConsent = l.Matched - l.Reachable
			rows = append(rows, l)
		}
		return rs.Err()
	})
	if err != nil {
		return Audience{}, fmt.Errorf("count audience: %w", err)
	}
	return buildAudience(rows), nil
}

// Channel is a message channel. The zero value is not valid.
type Channel struct{ column string }

// Column names for the two valid channels. Constants cannot change at run time.
const (
	smsColumn   = "sms_consent"
	emailColumn = "email_consent"
)

// SMS returns the SMS channel.
func SMS() Channel { return Channel{smsColumn} }

// Email returns the email channel.
func Email() Channel { return Channel{emailColumn} }

// ErrInvalidChannel means the channel is the zero value.
var ErrInvalidChannel = errors.New("segment: invalid channel")

// ErrEmptyWhere means the Where is the zero value.
var ErrEmptyWhere = errors.New("segment: empty where")

// Recipient is a user who can get a message on one channel.
type Recipient struct {
	ID        int64
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Language  string
}

// Recipients returns the users that match w and consent to ch, ordered by id.
// It is the only function that returns user rows.
func Recipients(ctx context.Context, pool *pgxpool.Pool, w Where, ch Channel) ([]Recipient, error) {
	switch ch.column {
	case smsColumn, emailColumn:
	default:
		return nil, ErrInvalidChannel
	}
	if w.sql == "" {
		return nil, ErrEmptyWhere
	}
	q := `SELECT id, first_name, last_name, email, phone, language
FROM users
WHERE (` + w.SQL() + `) AND ` + ch.column + ` = TRUE
ORDER BY id`
	out := []Recipient{}
	err := inTx(ctx, pool, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, q, w.Args()...)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var r Recipient
			if err := rs.Scan(&r.ID, &r.FirstName, &r.LastName, &r.Email, &r.Phone, &r.Language); err != nil {
				return err
			}
			out = append(out, r)
		}
		return rs.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("recipients: %w", err)
	}
	return out, nil
}
