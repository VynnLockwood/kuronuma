// Package store is Gaia's doorway: the Postgres-backed implementation of storage.
package store

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/core/cerberus"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Postgres implements cerberus.Store on a pgx connection pool.
type Postgres struct {
	pool *pgxpool.Pool
}

var _ cerberus.Store = (*Postgres)(nil)

// Open connects to Postgres and verifies the connection.
func Open(ctx context.Context, databaseURL string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Postgres{pool: pool}, nil
}

// Close releases the pool.
func (p *Postgres) Close() { p.pool.Close() }

// Ping checks database connectivity.
func (p *Postgres) Ping(ctx context.Context) error { return p.pool.Ping(ctx) }

// Migrate applies all pending migrations.
func (p *Postgres) Migrate(ctx context.Context) error {
	db := stdlib.OpenDBFromPool(p.pool)
	defer db.Close()
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.UpContext(ctx, db, "migrations")
}

// UpsertUser creates the user or refreshes profile fields, keyed on the Google subject.
func (p *Postgres) UpsertUser(ctx context.Context, id cerberus.Identity) (cerberus.User, error) {
	var u cerberus.User
	err := p.pool.QueryRow(ctx, `
		INSERT INTO users (google_sub, email, name, picture_url)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (google_sub) DO UPDATE
		   SET email = EXCLUDED.email,
		       name = EXCLUDED.name,
		       picture_url = EXCLUDED.picture_url,
		       last_login_at = now()
		RETURNING id::text, email, name, picture_url`,
		id.Subject, id.Email, id.Name, id.Picture,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Picture)
	return u, err
}

// CreateSession stores a hashed session token.
func (p *Postgres) CreateSession(ctx context.Context, userID string, tokenHash []byte, expiresAt time.Time) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2::uuid, $3)`,
		tokenHash, userID, expiresAt)
	return err
}

// GetSessionUser returns the user for an unexpired session, or cerberus.ErrNotFound.
func (p *Postgres) GetSessionUser(ctx context.Context, tokenHash []byte, now time.Time) (cerberus.User, error) {
	var u cerberus.User
	err := p.pool.QueryRow(ctx, `
		SELECT u.id::text, u.email, u.name, u.picture_url
		  FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token_hash = $1 AND s.expires_at > $2`,
		tokenHash, now,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Picture)
	if errors.Is(err, pgx.ErrNoRows) {
		return cerberus.User{}, cerberus.ErrNotFound
	}
	return u, err
}

// DeleteSession removes one session.
func (p *Postgres) DeleteSession(ctx context.Context, tokenHash []byte) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return err
}

// DeleteExpiredSessions removes expired sessions and returns how many were removed.
func (p *Postgres) DeleteExpiredSessions(ctx context.Context, now time.Time) (int64, error) {
	tag, err := p.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= $1`, now)
	return tag.RowsAffected(), err
}
