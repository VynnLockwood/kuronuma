// Package cerberus guards the gates: it signs users in (Google today) and
// manages their sessions. Later it will also act as the OAuth server for MCP clients.
package cerberus

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned by a Store when a row does not exist or has expired.
var ErrNotFound = errors.New("cerberus: not found")

// Identity is what an upstream identity provider tells us about a person.
type Identity struct {
	Subject       string // stable provider user id (Google "sub"); the account key
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// User is a Nemosyne account.
type User struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// Provider is an upstream OpenID Connect identity provider.
type Provider interface {
	// AuthCodeURL builds the provider's authorization URL (PKCE S256 derived from verifier).
	AuthCodeURL(state, nonce, verifier string) string
	// Exchange trades the code for tokens and returns the verified identity.
	Exchange(ctx context.Context, code, verifier, nonce string) (Identity, error)
}

// Store persists users and sessions. Session tokens are only ever stored hashed.
type Store interface {
	UpsertUser(ctx context.Context, id Identity) (User, error)
	CreateSession(ctx context.Context, userID string, tokenHash []byte, expiresAt time.Time) error
	GetSessionUser(ctx context.Context, tokenHash []byte, now time.Time) (User, error)
	DeleteSession(ctx context.Context, tokenHash []byte) error
	DeleteExpiredSessions(ctx context.Context, now time.Time) (int64, error)
}
