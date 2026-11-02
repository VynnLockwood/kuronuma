package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/VynnLockwood/kuronuma/apps/api/internal/modules/atlas"
)

var _ atlas.Store = (*Postgres)(nil)

const memoryColumns = `id::text, kind, title, body, status, priority, due_on, category, tags, created_at, updated_at, completed_at`

func scanMemory(row scanner) (atlas.Memory, error) {
	var m atlas.Memory
	err := row.Scan(&m.ID, &m.Kind, &m.Title, &m.Body, &m.Status, &m.Priority, &m.DueOn, &m.Category, &m.Tags, &m.CreatedAt, &m.UpdatedAt, &m.CompletedAt)
	return m, err
}

func nonNil(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}

// AddMemory stores a new item (status open) and returns it.
func (p *Postgres) AddMemory(ctx context.Context, userID string, n atlas.NewMemory) (atlas.Memory, error) {
	return scanMemory(p.pool.QueryRow(ctx, `
		INSERT INTO memories (user_id, kind, title, body, priority, due_on, category, tags)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8)
		RETURNING `+memoryColumns,
		userID, n.Kind, n.Title, n.Body, n.Priority, n.DueOn, n.Category, nonNil(n.Tags)))
}

// CountMemories returns how many items the user has.
func (p *Postgres) CountMemories(ctx context.Context, userID string) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx, `SELECT count(*) FROM memories WHERE user_id = $1::uuid`, userID).Scan(&n)
	return n, err
}

// ListMemories returns the user's items that match f, in the order atlas.Less describes.
func (p *Postgres) ListMemories(ctx context.Context, userID string, f atlas.Filter) ([]atlas.Memory, error) {
	var q strings.Builder
	args := []any{userID}
	q.WriteString(`SELECT ` + memoryColumns + ` FROM memories WHERE user_id = $1::uuid`)
	if f.Kind != "" {
		args = append(args, f.Kind)
		fmt.Fprintf(&q, ` AND kind = $%d`, len(args))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		fmt.Fprintf(&q, ` AND status = $%d`, len(args))
	}
	if f.Category != "" {
		args = append(args, f.Category)
		fmt.Fprintf(&q, ` AND category = $%d`, len(args))
	}
	if f.Tag != "" {
		args = append(args, f.Tag)
		fmt.Fprintf(&q, ` AND $%d = ANY(tags)`, len(args))
	}
	if f.DueBefore != nil {
		args = append(args, *f.DueBefore)
		fmt.Fprintf(&q, ` AND due_on <= $%d`, len(args))
	}
	if f.DueAfter != nil {
		args = append(args, *f.DueAfter)
		fmt.Fprintf(&q, ` AND due_on >= $%d`, len(args))
	}
	for _, term := range f.Terms {
		args = append(args, atlas.LikePattern(term))
		n := len(args)
		fmt.Fprintf(&q, ` AND (title ILIKE $%[1]d ESCAPE '\' OR body ILIKE $%[1]d ESCAPE '\'`+
			` OR EXISTS (SELECT 1 FROM unnest(tags) t WHERE t ILIKE $%[1]d ESCAPE '\'))`, n)
	}
	limit := f.Limit
	if limit <= 0 {
		limit = atlas.DefaultLimit
	}
	if limit > atlas.MaxLimit {
		limit = atlas.MaxLimit
	}
	args = append(args, limit)
	fmt.Fprintf(&q, ` ORDER BY (status <> 'open'), due_on ASC NULLS LAST,`+
		` CASE priority WHEN 'high' THEN 0 WHEN 'normal' THEN 1 ELSE 2 END, created_at DESC LIMIT $%d`, len(args))

	rows, err := p.pool.Query(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []atlas.Memory
	for rows.Next() {
		m, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// UpdateMemory applies the non-nil fields of u. completed_at follows the status: set when the item becomes done
// (kept if it already was), cleared when it leaves done, untouched when the status is not part of the update.
func (p *Postgres) UpdateMemory(ctx context.Context, userID, id string, u atlas.Update) (atlas.Memory, bool, error) {
	sets := []string{"updated_at = now()"}
	args := []any{userID, id}
	set := func(col string, v any) {
		args = append(args, v)
		sets = append(sets, fmt.Sprintf("%s = $%d", col, len(args)))
	}
	if u.Title != nil {
		set("title", *u.Title)
	}
	if u.Body != nil {
		set("body", *u.Body)
	}
	if u.Priority != nil {
		set("priority", *u.Priority)
	}
	if u.Category != nil {
		set("category", *u.Category)
	}
	if u.Tags != nil {
		set("tags", nonNil(*u.Tags))
	}
	if u.SetDue {
		set("due_on", u.DueOn) // a nil *time.Time clears the date
	}
	if u.Status != nil {
		args = append(args, *u.Status)
		n := len(args)
		sets = append(sets, fmt.Sprintf("status = $%d", n),
			fmt.Sprintf("completed_at = CASE WHEN $%d = 'done' THEN COALESCE(completed_at, now()) ELSE NULL END", n))
	}
	m, err := scanMemory(p.pool.QueryRow(ctx,
		`UPDATE memories SET `+strings.Join(sets, ", ")+` WHERE user_id = $1::uuid AND id = $2::uuid RETURNING `+memoryColumns, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return atlas.Memory{}, false, nil
	}
	if err != nil {
		return atlas.Memory{}, false, err
	}
	return m, true, nil
}

// DeleteMemory removes one of the user's items and reports whether it existed.
func (p *Postgres) DeleteMemory(ctx context.Context, userID, id string) (bool, error) {
	tag, err := p.pool.Exec(ctx, `DELETE FROM memories WHERE id = $1::uuid AND user_id = $2::uuid`, id, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
