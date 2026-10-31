// Package config loads and validates Kuronuma API settings from the environment.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Config holds all runtime settings for the API.
type Config struct {
	Addr               string
	DatabaseURL        string
	PublicURL          string // external base URL, no trailing slash, e.g. https://kuronuma.example.com
	SessionSecret      []byte
	SessionTTL         time.Duration
	GoogleClientID     string
	GoogleClientSecret string
	// AllowedEmails restricts sign-in to these (lowercased) addresses.
	// Empty means any Google account with a verified email may sign in.
	AllowedEmails []string
}

// SecureCookies reports whether cookies should carry the Secure attribute.
func (c Config) SecureCookies() bool { return strings.HasPrefix(c.PublicURL, "https://") }

// GoogleRedirectURL is the callback URL to register in the Google Cloud console.
func (c Config) GoogleRedirectURL() string { return c.PublicURL + "/api/auth/google/callback" }

// GoogleConfigured reports whether both Google credentials are present.
func (c Config) GoogleConfigured() bool {
	return c.GoogleClientID != "" && c.GoogleClientSecret != ""
}

// Load reads configuration using getenv (pass os.Getenv) and reports every problem at once.
func Load(getenv func(string) string) (Config, error) {
	var errs []error

	cfg := Config{
		Addr:               valueOr(getenv("API_ADDR"), ":8081"),
		DatabaseURL:        getenv("DATABASE_URL"),
		PublicURL:          strings.TrimRight(getenv("PUBLIC_URL"), "/"),
		SessionSecret:      []byte(getenv("SESSION_SECRET")),
		GoogleClientID:     getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: getenv("GOOGLE_CLIENT_SECRET"),
	}

	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}

	if u, err := url.Parse(cfg.PublicURL); err != nil || cfg.PublicURL == "" ||
		(u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Path != "" {
		errs = append(errs, errors.New("PUBLIC_URL must be an origin like https://example.com or http://localhost:8080"))
	}

	if len(cfg.SessionSecret) < 32 {
		errs = append(errs, errors.New("SESSION_SECRET must be at least 32 characters (try: openssl rand -base64 48)"))
	}

	ttl := valueOr(getenv("SESSION_TTL"), "336h")
	d, err := time.ParseDuration(ttl)
	if err != nil || d <= 0 {
		errs = append(errs, fmt.Errorf("SESSION_TTL %q is not a positive duration (e.g. 336h)", ttl))
	}
	cfg.SessionTTL = d

	if (cfg.GoogleClientID == "") != (cfg.GoogleClientSecret == "") {
		errs = append(errs, errors.New("set both GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET, or neither"))
	}

	for _, e := range strings.Split(getenv("ALLOWED_EMAILS"), ",") {
		if e = strings.ToLower(strings.TrimSpace(e)); e != "" {
			cfg.AllowedEmails = append(cfg.AllowedEmails, e)
		}
	}

	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}
	return cfg, nil
}

func valueOr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
