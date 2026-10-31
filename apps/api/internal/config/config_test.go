package config

import (
	"strings"
	"testing"
	"time"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func valid() map[string]string {
	return map[string]string{
		"DATABASE_URL":   "postgres://u:p@localhost:5432/db",
		"PUBLIC_URL":     "http://localhost:8080/",
		"SESSION_SECRET": strings.Repeat("s", 32),
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(env(valid()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":8081" || cfg.PublicURL != "http://localhost:8080" || cfg.SessionTTL != 336*time.Hour {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.SecureCookies() || cfg.GoogleConfigured() {
		t.Fatal("http dev URL must not use Secure cookies; google must be unconfigured")
	}
	if got := cfg.GoogleRedirectURL(); got != "http://localhost:8080/api/auth/google/callback" {
		t.Fatalf("redirect url = %q", got)
	}
}

func TestLoadHTTPSAllowListAndGoogle(t *testing.T) {
	m := valid()
	m["PUBLIC_URL"] = "https://kuronuma.example.com"
	m["GOOGLE_CLIENT_ID"], m["GOOGLE_CLIENT_SECRET"] = "id", "secret"
	m["ALLOWED_EMAILS"] = " A@x.com, ,b@x.com "
	cfg, err := Load(env(m))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.SecureCookies() || !cfg.GoogleConfigured() {
		t.Fatal("https must be Secure; google must be configured")
	}
	if len(cfg.AllowedEmails) != 2 || cfg.AllowedEmails[0] != "a@x.com" || cfg.AllowedEmails[1] != "b@x.com" {
		t.Fatalf("allow-list = %v", cfg.AllowedEmails)
	}
}

func TestLoadReportsEveryProblem(t *testing.T) {
	_, err := Load(env(map[string]string{
		"PUBLIC_URL":       "localhost:8080/path",
		"SESSION_SECRET":   "short",
		"SESSION_TTL":      "forever",
		"GOOGLE_CLIENT_ID": "only-id",
	}))
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{"DATABASE_URL", "PUBLIC_URL", "SESSION_SECRET", "SESSION_TTL", "GOOGLE_CLIENT_ID"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %s:\n%v", want, err)
		}
	}
}
