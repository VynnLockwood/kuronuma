package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/VynnLockwood/nemosyne/apps/api/internal/core/themis"
)

var _ themis.Store = (*Postgres)(nil)

// categoryRefs lists every table that has user_id and category columns, with the label the tools report
// it under. Rename and merge update all of them in one transaction. Table names come only from this
// list, never from input. Add a table here when a new module stores a category.
var categoryRefs = []struct{ table, label string }{
	{"transactions", "spending"},
}

// ListCategories returns the user's listed names plus every in-use name that is missing from the list.
func (p *Postgres) ListCategories(ctx context.Context, userID string) ([]themis.Category, error) {
	byName := map[string]*themis.Category{}
	get := func(name string) *themis.Category {
		c, ok := byName[name]
		if !ok {
			c = &themis.Category{Name: name, Uses: map[string]int{}}
			byName[name] = c
		}
		return c
	}

	rows, err := p.pool.Query(ctx, `SELECT name FROM categories WHERE user_id = $1::uuid`, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return nil, err
		}
		get(name).Listed = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, ref := range categoryRefs {
		rows, err := p.pool.Query(ctx,
			`SELECT category, COUNT(*) FROM `+ref.table+` WHERE user_id = $1::uuid GROUP BY category`, userID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var name string
			var n int64
			if err := rows.Scan(&name, &n); err != nil {
				rows.Close()
				return nil, err
			}
			get(name).Uses[ref.label] = int(n)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	out := make([]themis.Category, 0, len(byName))
	for _, c := range byName {
		out = append(out, *c)
	}
	return out, nil
}

// CreateCategory adds name to the user's list; created is false when it was already there.
func (p *Postgres) CreateCategory(ctx context.Context, userID, name string) (bool, error) {
	tag, err := p.pool.Exec(ctx,
		`INSERT INTO categories (user_id, name) VALUES ($1::uuid, $2) ON CONFLICT (user_id, name) DO NOTHING`, userID, name)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// RenameCategory renames a listed category and every record that uses it, in one transaction.
func (p *Postgres) RenameCategory(ctx context.Context, userID, from, to string) (map[string]int, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op after Commit

	tag, err := tx.Exec(ctx, `UPDATE categories SET name = $3 WHERE user_id = $1::uuid AND name = $2`, userID, from, to)
	if isUniqueViolation(err) {
		return nil, themis.ErrExists
	}
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, themis.ErrNotFound
	}

	updated := map[string]int{}
	for _, ref := range categoryRefs {
		tag, err := tx.Exec(ctx,
			`UPDATE `+ref.table+` SET category = $3 WHERE user_id = $1::uuid AND category = $2`, userID, from, to)
		if err != nil {
			return nil, err
		}
		updated[ref.label] = int(tag.RowsAffected())
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return updated, nil
}

// MergeCategories moves every record from the sources to target and removes the sources from the list,
// in one transaction: if any source is unknown nothing changes.
func (p *Postgres) MergeCategories(ctx context.Context, userID string, sources []string, target string) (map[string]int, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op after Commit

	if target != themis.Uncategorized {
		var ok bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM categories WHERE user_id = $1::uuid AND name = $2)`, userID, target).Scan(&ok); err != nil {
			return nil, err
		}
		if !ok {
			return nil, themis.ErrNotFound
		}
	}

	moved := map[string]int{}
	for _, src := range sources {
		if src == target || src == themis.Uncategorized {
			return nil, fmt.Errorf("invalid merge source %q", src)
		}
		removed, err := tx.Exec(ctx, `DELETE FROM categories WHERE user_id = $1::uuid AND name = $2`, userID, src)
		if err != nil {
			return nil, err
		}
		n := removed.RowsAffected()
		for _, ref := range categoryRefs {
			tag, err := tx.Exec(ctx,
				`UPDATE `+ref.table+` SET category = $3 WHERE user_id = $1::uuid AND category = $2`, userID, src, target)
			if err != nil {
				return nil, err
			}
			moved[ref.label] += int(tag.RowsAffected())
			n += tag.RowsAffected()
		}
		if n == 0 {
			return nil, themis.ErrNotFound // neither listed nor in use; the deferred Rollback undoes earlier sources
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return moved, nil
}
