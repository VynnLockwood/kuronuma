package cerberus

import (
	"context"
	"time"
)

// Scope is a permission an assistant can be granted, shown on the consent screen.
type Scope struct {
	Name        string
	Description string
}

// OAuthClient is an MCP client registered through dynamic client registration.
type OAuthClient struct {
	ID           string
	Name         string
	RedirectURIs []string
	CreatedAt    time.Time
}

// AuthCode is a one-time authorization code awaiting exchange.
type AuthCode struct {
	Hash          []byte // SHA-256 of the code
	ClientID      string
	UserID        string
	RedirectURI   string
	CodeChallenge string // PKCE S256 challenge
	Scopes        []string
	Resource      string
	ExpiresAt     time.Time
}

// TokenKind distinguishes access tokens from refresh tokens.
type TokenKind string

const (
	TokenAccess  TokenKind = "access"
	TokenRefresh TokenKind = "refresh"
)

// Token is an issued access or refresh token (stored hashed).
type Token struct {
	Hash      []byte // SHA-256 of the token
	Kind      TokenKind
	ClientID  string
	UserID    string
	Scopes    []string
	ExpiresAt time.Time
}

// TokenInfo is what a verified access token tells the resource server.
type TokenInfo struct {
	UserID   string
	ClientID string
	Scopes   []string
}

// OAuthStore persists OAuth clients, codes and tokens.
type OAuthStore interface {
	CreateClient(ctx context.Context, c OAuthClient) error
	GetClient(ctx context.Context, id string) (OAuthClient, error)

	SaveCode(ctx context.Context, c AuthCode) error
	// ConsumeCode atomically deletes and returns an unexpired code (single use).
	ConsumeCode(ctx context.Context, hash []byte, now time.Time) (AuthCode, error)

	SaveToken(ctx context.Context, t Token) error
	GetToken(ctx context.Context, hash []byte, kind TokenKind, now time.Time) (Token, error)
	// ConsumeRefreshToken atomically deletes and returns an unexpired refresh token (rotation).
	ConsumeRefreshToken(ctx context.Context, hash []byte, now time.Time) (Token, error)

	DeleteExpiredOAuth(ctx context.Context, now time.Time) (int64, error)
}
