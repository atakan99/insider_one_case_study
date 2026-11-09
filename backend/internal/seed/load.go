package seed

import (
	"context"
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Load inserts the generated data. If users exist and reset is false, it
// does nothing and returns inserted == false.
func Load(ctx context.Context, pool *pgxpool.Pool, o Options, reset bool) (inserted bool, err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users)`).Scan(&exists); err != nil {
		return false, fmt.Errorf("check users: %w", err)
	}
	if exists && !reset {
		return false, nil
	}
	if reset {
		if _, err := tx.Exec(ctx, `TRUNCATE events, users RESTART IDENTITY`); err != nil {
			return false, fmt.Errorf("truncate: %w", err)
		}
	}

	users, events := Generate(o)
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"users"}, userColumns, pgx.CopyFromSlice(len(users), func(i int) ([]any, error) {
		u := users[i]
		return []any{u.ID, u.FirstName, u.LastName, u.Email, u.Phone, u.Language, u.Country, u.City,
			u.BirthDate, u.EmailConsent, u.SMSConsent, u.CreatedAt, u.LastCartAddAt, u.LastPurchaseAt,
			int32(u.OrderCount), money(u.TotalSpentKurus)}, nil
	})); err != nil {
		return false, fmt.Errorf("copy users: %w", err)
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"events"}, eventColumns, pgx.CopyFromSlice(len(events), func(i int) ([]any, error) {
		e := events[i]
		var amount any
		if e.AmountKurus != nil {
			amount = money(*e.AmountKurus)
		}
		return []any{e.ID, e.UserID, e.Type, amount, e.OccurredAt}, nil
	})); err != nil {
		return false, fmt.Errorf("copy events: %w", err)
	}

	for _, table := range []string{"users", "events"} {
		q := fmt.Sprintf(`SELECT setval(pg_get_serial_sequence('%[1]s','id'), coalesce((SELECT max(id) FROM %[1]s), 1), (SELECT count(*) > 0 FROM %[1]s))`, table)
		if _, err := tx.Exec(ctx, q); err != nil {
			return false, fmt.Errorf("set %s sequence: %w", table, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit: %w", err)
	}
	return true, nil
}

var (
	userColumns = []string{"id", "first_name", "last_name", "email", "phone", "language", "country", "city",
		"birth_date", "email_consent", "sms_consent", "created_at", "last_cart_add_at", "last_purchase_at",
		"order_count", "total_spent"}
	eventColumns = []string{"id", "user_id", "type", "amount", "occurred_at"}
)

// money converts kurus to a numeric with two decimals.
func money(kurus int64) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(kurus), Exp: -2, Valid: true}
}
