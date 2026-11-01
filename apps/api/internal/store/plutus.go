package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/modules/plutus"
)

var _ plutus.Store = (*Postgres)(nil)

const txColumns = `id::text, amount_minor, currency, merchant, category, occurred_on, note, COALESCE(reference, '')`

type scanner interface{ Scan(dest ...any) error }

func scanTx(row scanner) (plutus.Transaction, error) {
	var t plutus.Transaction
	err := row.Scan(&t.ID, &t.AmountMinor, &t.Currency, &t.Merchant, &t.Category, &t.OccurredOn, &t.Note, &t.Reference)
	return t, err
}

// AddTransaction inserts a record. When the user already has one with the same reference,
// nothing is inserted and the existing record is returned with created=false.
func (p *Postgres) AddTransaction(ctx context.Context, userID string, t plutus.NewTransaction) (plutus.Transaction, bool, error) {
	var ref any // NULL when there is no reference, so the unique index ignores the row
	if t.Reference != "" {
		ref = t.Reference
	}
	tx, err := scanTx(p.pool.QueryRow(ctx, `
		INSERT INTO transactions (user_id, amount_minor, currency, merchant, category, occurred_on, note, reference)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, reference) WHERE reference IS NOT NULL DO NOTHING
		RETURNING `+txColumns,
		userID, t.AmountMinor, t.Currency, t.Merchant, t.Category, t.OccurredOn, t.Note, ref))
	if err == nil {
		return tx, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return plutus.Transaction{}, false, err
	}
	existing, err := scanTx(p.pool.QueryRow(ctx,
		`SELECT `+txColumns+` FROM transactions WHERE user_id = $1::uuid AND reference = $2`,
		userID, t.Reference))
	return existing, false, err
}

// CountTransactions returns how many transactions the user has.
func (p *Postgres) CountTransactions(ctx context.Context, userID string) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx, `SELECT count(*) FROM transactions WHERE user_id = $1::uuid`, userID).Scan(&n)
	return n, err
}

// FindPossibleDuplicate looks for the same amount, currency, date and merchant (case-insensitive).
func (p *Postgres) FindPossibleDuplicate(ctx context.Context, userID string, t plutus.NewTransaction) (string, bool, error) {
	var id string
	err := p.pool.QueryRow(ctx, `
		SELECT id::text FROM transactions
		 WHERE user_id = $1::uuid AND amount_minor = $2 AND currency = $3
		   AND occurred_on = $4 AND lower(merchant) = lower($5)
		 LIMIT 1`,
		userID, t.AmountMinor, t.Currency, t.OccurredOn, t.Merchant).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	return id, err == nil, err
}

// ListTransactions returns the user's records, newest first.
func (p *Postgres) ListTransactions(ctx context.Context, userID string, f plutus.ListFilter) ([]plutus.Transaction, error) {
	var q strings.Builder
	args := []any{userID}
	q.WriteString(`SELECT ` + txColumns + ` FROM transactions WHERE user_id = $1::uuid`)
	if f.From != nil {
		args = append(args, *f.From)
		fmt.Fprintf(&q, ` AND occurred_on >= $%d`, len(args))
	}
	if f.To != nil {
		args = append(args, *f.To)
		fmt.Fprintf(&q, ` AND occurred_on <= $%d`, len(args))
	}
	if f.Category != "" {
		args = append(args, f.Category)
		fmt.Fprintf(&q, ` AND category = $%d`, len(args))
	}
	args = append(args, f.Limit)
	fmt.Fprintf(&q, ` ORDER BY occurred_on DESC, created_at DESC LIMIT $%d`, len(args))

	rows, err := p.pool.Query(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []plutus.Transaction
	for rows.Next() {
		t, err := scanTx(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Summarize totals the user's records by category, month or merchant, per currency.
func (p *Postgres) Summarize(ctx context.Context, userID string, f plutus.SummaryFilter) ([]plutus.SummaryRow, error) {
	// The grouping expression comes from this fixed list, never from user input.
	var groupExpr string
	switch f.GroupBy {
	case "category":
		groupExpr = `category`
	case "month":
		groupExpr = `to_char(occurred_on, 'YYYY-MM')`
	case "merchant":
		groupExpr = `merchant`
	default:
		return nil, fmt.Errorf("unsupported group_by %q", f.GroupBy)
	}
	order := `SUM(amount_minor) DESC, 1`
	if f.GroupBy == "month" {
		order = `1, currency`
	}
	rows, err := p.pool.Query(ctx, `
		SELECT `+groupExpr+` AS k, currency, SUM(amount_minor)::bigint, COUNT(*)::bigint
		  FROM transactions
		 WHERE user_id = $1::uuid AND occurred_on >= $2 AND occurred_on <= $3
		 GROUP BY k, currency
		 ORDER BY `+order,
		userID, f.From, f.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []plutus.SummaryRow
	for rows.Next() {
		var r plutus.SummaryRow
		if err := rows.Scan(&r.Key, &r.Currency, &r.TotalMinor, &r.Count); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// DeleteTransaction removes one of the user's records and reports whether it existed.
func (p *Postgres) DeleteTransaction(ctx context.Context, userID, id string) (bool, error) {
	tag, err := p.pool.Exec(ctx,
		`DELETE FROM transactions WHERE id = $1::uuid AND user_id = $2::uuid`, id, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// CategoryNames returns the names in the user's category list, sorted. uncategorized is built in and not included.
func (p *Postgres) CategoryNames(ctx context.Context, userID string) ([]string, error) {
	rows, err := p.pool.Query(ctx, `SELECT name FROM categories WHERE user_id = $1::uuid ORDER BY name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}
