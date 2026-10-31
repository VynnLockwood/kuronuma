package store

import (
	"context"

	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/modules/plutus"
)

const payeeRuleColumns = `id::text, payee, payee_norm, category, description, updated_at`

// PayeeRules returns all of the user's payee rules, ordered by payee_norm.
func (p *Postgres) PayeeRules(ctx context.Context, userID string) ([]plutus.PayeeRule, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT `+payeeRuleColumns+` FROM payee_rules WHERE user_id = $1::uuid ORDER BY payee_norm`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []plutus.PayeeRule
	for rows.Next() {
		var r plutus.PayeeRule
		if err := rows.Scan(&r.ID, &r.Payee, &r.PayeeNorm, &r.Category, &r.Description, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// SetPayeeRule creates the rule or, when the user already has one for the same payee_norm, updates it.
// created is true only for a new rule. (xmax is 0 on a freshly inserted row and not on an updated one.)
func (p *Postgres) SetPayeeRule(ctx context.Context, userID string, r plutus.PayeeRule) (plutus.PayeeRule, bool, error) {
	var out plutus.PayeeRule
	var created bool
	err := p.pool.QueryRow(ctx, `
		INSERT INTO payee_rules (user_id, payee, payee_norm, category, description)
		VALUES ($1::uuid, $2, $3, $4, $5)
		ON CONFLICT (user_id, payee_norm) DO UPDATE
		   SET payee = EXCLUDED.payee, category = EXCLUDED.category,
		       description = EXCLUDED.description, updated_at = now()
		RETURNING `+payeeRuleColumns+`, (xmax = 0)`,
		userID, r.Payee, r.PayeeNorm, r.Category, r.Description,
	).Scan(&out.ID, &out.Payee, &out.PayeeNorm, &out.Category, &out.Description, &out.UpdatedAt, &created)
	return out, created, err
}

// DeletePayeeRule removes the user's rule for payeeNorm and reports whether it existed.
func (p *Postgres) DeletePayeeRule(ctx context.Context, userID, payeeNorm string) (bool, error) {
	tag, err := p.pool.Exec(ctx,
		`DELETE FROM payee_rules WHERE user_id = $1::uuid AND payee_norm = $2`, userID, payeeNorm)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
