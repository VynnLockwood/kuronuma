// Package google implements cerberus.Provider with Google OpenID Connect.
package google

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/core/cerberus"
)

const googleIssuer = "https://accounts.google.com"

// Provider implements cerberus.Provider using Google's OpenID Connect endpoints.
type Provider struct {
	oauth    *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

// New discovers Google's endpoints (network call) and returns a Provider.
func New(ctx context.Context, clientID, clientSecret, redirectURL string) (*Provider, error) {
	p, err := oidc.NewProvider(ctx, googleIssuer)
	if err != nil {
		return nil, fmt.Errorf("discover google oidc: %w", err)
	}
	return &Provider{
		oauth: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     p.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
		},
		verifier: p.Verifier(&oidc.Config{ClientID: clientID}),
	}, nil
}

// AuthCodeURL implements cerberus.Provider.
func (g *Provider) AuthCodeURL(state, nonce, verifier string) string {
	return g.oauth.AuthCodeURL(state, oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier))
}

// Exchange implements cerberus.Provider. It checks the ID token's signature, issuer,
// audience and expiry (via go-oidc) and the nonce.
func (g *Provider) Exchange(ctx context.Context, code, verifier, nonce string) (cerberus.Identity, error) {
	tok, err := g.oauth.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return cerberus.Identity{}, fmt.Errorf("exchange code: %w", err)
	}
	raw, ok := tok.Extra("id_token").(string)
	if !ok || raw == "" {
		return cerberus.Identity{}, errors.New("google response had no id_token")
	}
	idt, err := g.verifier.Verify(ctx, raw)
	if err != nil {
		return cerberus.Identity{}, fmt.Errorf("verify id_token: %w", err)
	}
	if idt.Nonce != nonce {
		return cerberus.Identity{}, errors.New("id_token nonce mismatch")
	}
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := idt.Claims(&claims); err != nil {
		return cerberus.Identity{}, fmt.Errorf("read claims: %w", err)
	}
	return cerberus.Identity{
		Subject:       idt.Subject,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		Picture:       claims.Picture,
	}, nil
}

var _ cerberus.Provider = (*Provider)(nil)
