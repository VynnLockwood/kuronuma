package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/VynnLockwood/kuronuma/apps/api/internal/modules/plutus"
)

const billColumns = `id::text, name, amount_minor, currency, category, start_month, ended_from`

func scanBill(row scanner) (plutus.Bill, error) {
	var b plutus.Bill
	err := row.Scan(&b.ID, &b.Name, &b.AmountMinor, &b.Currency, &b.Category, &b.StartMonth, &b.EndedFrom)
	return b, err
}

// AddBill inserts a bill for the user.
func (p *Postgres) AddBill(ctx context.Context, userID string, b plutus.NewBill) (plutus.Bill, error) {
	return scanBill(p.pool.QueryRow(ctx, `
		INSERT INTO bills (user_id, name, amount_minor, currency, category, start_month)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		RETURNING `+billColumns,
		userID, b.Name, b.AmountMinor, b.Currency, b.Category, b.StartMonth))
}

// GetBill returns the user's bill, if there is one.
func (p *Postgres) GetBill(ctx context.Context, userID, id string) (plutus.Bill, bool, error) {
	b, err := scanBill(p.pool.QueryRow(ctx,
		`SELECT `+billColumns+` FROM bills WHERE user_id = $1::uuid AND id = $2::uuid`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return plutus.Bill{}, false, nil
	}
	return b, err == nil, err
}

// ListBills returns all of the user's bills, ordered by name, then id.
func (p *Postgres) ListBills(ctx context.Context, userID string) ([]plutus.Bill, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT `+billColumns+` FROM bills WHERE user_id = $1::uuid ORDER BY name, id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []plutus.Bill
	for rows.Next() {
		b, err := scanBill(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// UpdateBill changes only the provided columns. found is false when the user has no such bill.
func (p *Postgres) UpdateBill(ctx context.Context, userID, id string, patch plutus.BillPatch) (plutus.Bill, bool, error) {
	sets := []string{"updated_at = now()"}
	args := []any{userID, id}
	set := func(col string, v any) {
		args = append(args, v)
		sets = append(sets, fmt.Sprintf("%s = $%d", col, len(args)))
	}
	if patch.Name != nil {
		set("name", *patch.Name)
	}
	if patch.Category != nil {
		set("category", *patch.Category)
	}
	if patch.AmountMinor != nil {
		set("amount_minor", *patch.AmountMinor)
	}
	if patch.StartMonth != nil {
		set("start_month", *patch.StartMonth)
	}
	if patch.ClearEnded {
		sets = append(sets, "ended_from = NULL")
	} else if patch.EndedFrom != nil {
		set("ended_from", *patch.EndedFrom)
	}
	b, err := scanBill(p.pool.QueryRow(ctx,
		`UPDATE bills SET `+strings.Join(sets, ", ")+
			` WHERE user_id = $1::uuid AND id = $2::uuid RETURNING `+billColumns, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return plutus.Bill{}, false, nil
	}
	return b, err == nil, err
}

// DeleteBill removes the user's bill and reports whether it existed.
func (p *Postgres) DeleteBill(ctx context.Context, userID, id string) (bool, error) {
	tag, err := p.pool.Exec(ctx, `DELETE FROM bills WHERE user_id = $1::uuid AND id = $2::uuid`, userID, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// CountBills returns how many bills the user has.
func (p *Postgres) CountBills(ctx context.Context, userID string) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx, `SELECT count(*) FROM bills WHERE user_id = $1::uuid`, userID).Scan(&n)
	return n, err
}

// HasPayments reports whether the user has a transaction whose reference starts with bill:<billID>:.
func (p *Postgres) HasPayments(ctx context.Context, userID, billID string) (bool, error) {
	var ok bool
	err := p.pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM transactions
		                WHERE user_id = $1::uuid AND reference LIKE 'bill:' || $2::text || ':%')`,
		userID, billID).Scan(&ok)
	return ok, err
}

// PaymentsByReference returns the user's transactions with exactly the listed references, keyed by reference.
func (p *Postgres) PaymentsByReference(ctx context.Context, userID string, refs []string) (map[string]plutus.Transaction, error) {
	out := map[string]plutus.Transaction{}
	if len(refs) == 0 {
		return out, nil
	}
	rows, err := p.pool.Query(ctx,
		`SELECT `+txColumns+` FROM transactions WHERE user_id = $1::uuid AND reference = ANY($2::text[])`, userID, refs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		t, err := scanTx(rows)
		if err != nil {
			return nil, err
		}
		out[t.Reference] = t
	}
	return out, rows.Err()
}
