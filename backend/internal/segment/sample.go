package segment

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// sampleQuery builds the sample query for w. The language list and the limit
// take the two placeholders after the arguments of w. It reads only id,
// first_name, language, the consent columns and the columns of the rule.
func sampleQuery(w Where, _ int) string {
	langArg := len(w.Args()) + 1
	limitArg := langArg + 1
	return fmt.Sprintf(`WITH r AS (
    SELECT id, first_name, language FROM users
    WHERE (%s) AND (email_consent OR sms_consent) AND language = ANY($%d::text[])
), names AS (
    SELECT language, first_name, min(id) AS first_id FROM r GROUP BY language, first_name
), ranked AS (
    SELECT language, first_name,
        row_number() OVER (PARTITION BY language ORDER BY char_length(first_name) DESC, first_name) AS by_length,
        first_id
    FROM names
), ordered AS (
    SELECT language, first_name,
        row_number() OVER (PARTITION BY language ORDER BY (by_length = 1) DESC, first_id) AS pos
    FROM ranked
)
SELECT language, first_name FROM ordered WHERE pos <= $%d ORDER BY language, pos`, w.SQL(), langArg, limitArg)
}

// SampleFirstNames returns up to n distinct first names of the reachable users
// of each language in langs. The first name of each list is the longest (ties:
// by name). The others come from the users with the lowest id. It returns first
// names only, never a user row.
func SampleFirstNames(ctx context.Context, pool *pgxpool.Pool, w Where, langs []string, n int) (map[string][]string, error) {
	if w.sql == "" {
		return nil, ErrEmptyWhere
	}
	args := append(append([]any{}, w.Args()...), langs, n)
	out := map[string][]string{}
	err := inTx(ctx, pool, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, sampleQuery(w, len(langs)), args...)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var lang, name string
			if err := rs.Scan(&lang, &name); err != nil {
				return err
			}
			out[lang] = append(out[lang], name)
		}
		return rs.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("sample first names: %w", err)
	}
	return out, nil
}
