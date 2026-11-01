package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/core/cerberus"
)

var _ cerberus.OAuthStore = (*Postgres)(nil)

// CreateClient stores a dynamically registered OAuth client.
func (p *Postgres) CreateClient(ctx context.Context, c cerberus.OAuthClient) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO oauth_clients (client_id, client_name, redirect_uris) VALUES ($1, $2, $3)`,
		c.ID, c.Name, c.RedirectURIs)
	return err
}

// GetClient returns a client or cerberus.ErrNotFound.
func (p *Postgres) GetClient(ctx context.Context, id string) (cerberus.OAuthClient, error) {
	var c cerberus.OAuthClient
	err := p.pool.QueryRow(ctx,
		`SELECT client_id, client_name, redirect_uris, created_at FROM oauth_clients WHERE client_id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.RedirectURIs, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return cerberus.OAuthClient{}, cerberus.ErrNotFound
	}
	return c, err
}

// SaveCode stores a hashed authorization code.
func (p *Postgres) SaveCode(ctx context.Context, c cerberus.AuthCode) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO oauth_codes (code_hash, client_id, user_id, redirect_uri, code_challenge, scopes, resource, expires_at)
		VALUES ($1, $2, $3::uuid, $4, $5, $6, $7, $8)`,
		c.Hash, c.ClientID, c.UserID, c.RedirectURI, c.CodeChallenge, c.Scopes, c.Resource, c.ExpiresAt)
	return err
}

// ConsumeCode deletes and returns an unexpired code in one statement, so a code can be used once.
func (p *Postgres) ConsumeCode(ctx context.Context, hash []byte, now time.Time) (cerberus.AuthCode, error) {
	c := cerberus.AuthCode{Hash: hash}
	err := p.pool.QueryRow(ctx, `
		DELETE FROM oauth_codes
		 WHERE code_hash = $1 AND expires_at > $2
		RETURNING client_id, user_id::text, redirect_uri, code_challenge, scopes, resource, expires_at`,
		hash, now,
	).Scan(&c.ClientID, &c.UserID, &c.RedirectURI, &c.CodeChallenge, &c.Scopes, &c.Resource, &c.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return cerberus.AuthCode{}, cerberus.ErrNotFound
	}
	return c, err
}

// SaveToken stores a hashed access or refresh token.
func (p *Postgres) SaveToken(ctx context.Context, t cerberus.Token) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO oauth_tokens (token_hash, kind, client_id, user_id, scopes, expires_at)
		VALUES ($1, $2, $3, $4::uuid, $5, $6)`,
		t.Hash, string(t.Kind), t.ClientID, t.UserID, t.Scopes, t.ExpiresAt)
	return err
}

// GetToken returns an unexpired token of the given kind, or cerberus.ErrNotFound.
func (p *Postgres) GetToken(ctx context.Context, hash []byte, kind cerberus.TokenKind, now time.Time) (cerberus.Token, error) {
	t := cerberus.Token{Hash: hash, Kind: kind}
	err := p.pool.QueryRow(ctx, `
		SELECT client_id, user_id::text, scopes, expires_at
		  FROM oauth_tokens
		 WHERE token_hash = $1 AND kind = $2 AND expires_at > $3`,
		hash, string(kind), now,
	).Scan(&t.ClientID, &t.UserID, &t.Scopes, &t.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return cerberus.Token{}, cerberus.ErrNotFound
	}
	return t, err
}

// ConsumeRefreshToken deletes and returns an unexpired refresh token, so each one is used once (rotation).
func (p *Postgres) ConsumeRefreshToken(ctx context.Context, hash []byte, now time.Time) (cerberus.Token, error) {
	t := cerberus.Token{Hash: hash, Kind: cerberus.TokenRefresh}
	err := p.pool.QueryRow(ctx, `
		DELETE FROM oauth_tokens
		 WHERE token_hash = $1 AND kind = 'refresh' AND expires_at > $2
		RETURNING client_id, user_id::text, scopes, expires_at`,
		hash, now,
	).Scan(&t.ClientID, &t.UserID, &t.Scopes, &t.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return cerberus.Token{}, cerberus.ErrNotFound
	}
	return t, err
}

// DeleteExpiredOAuth removes expired codes and tokens and returns how many rows were removed.
func (p *Postgres) DeleteExpiredOAuth(ctx context.Context, now time.Time) (int64, error) {
	codes, err := p.pool.Exec(ctx, `DELETE FROM oauth_codes WHERE expires_at <= $1`, now)
	if err != nil {
		return 0, err
	}
	tokens, err := p.pool.Exec(ctx, `DELETE FROM oauth_tokens WHERE expires_at <= $1`, now)
	return codes.RowsAffected() + tokens.RowsAffected(), err
}

// DeleteUnusedClients removes clients registered before `before` that hold no code and no token at all, and
// returns how many were removed.
func (p *Postgres) DeleteUnusedClients(ctx context.Context, before time.Time) (int64, error) {
	tag, err := p.pool.Exec(ctx, `
		DELETE FROM oauth_clients c
		 WHERE c.created_at < $1
		   AND NOT EXISTS (SELECT 1 FROM oauth_codes WHERE client_id = c.client_id)
		   AND NOT EXISTS (SELECT 1 FROM oauth_tokens WHERE client_id = c.client_id)`, before)
	return tag.RowsAffected(), err
}
