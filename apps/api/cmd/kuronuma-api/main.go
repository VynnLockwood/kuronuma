// Command nemosyne-api is the Go backend for Nemosyne.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/config"
	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/core/cerberus"
	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/core/cerberus/google"
	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/core/hermes"
	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/core/themis"
	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/modules/atlas"
	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/modules/plutus"
	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/store"
)

const version = "0.2.0"

// mcpAuth lets Hermes verify bearer tokens issued by Cerberus.
type mcpAuth struct{ oauth *cerberus.OAuth }

func (a mcpAuth) Authenticate(ctx context.Context, bearer string) (hermes.Caller, error) {
	info, err := a.oauth.VerifyAccessToken(ctx, bearer)
	if errors.Is(err, cerberus.ErrNotFound) {
		return hermes.Caller{}, hermes.ErrUnauthorized
	}
	if err != nil {
		return hermes.Caller{}, err
	}
	return hermes.Caller{UserID: info.UserID, ClientID: info.ClientID, Scopes: info.Scopes}, nil
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.Migrate(ctx); err != nil {
		return err
	}

	// Leave provider nil (not a typed-nil) when Google is not configured.
	var provider cerberus.Provider
	if cfg.GoogleConfigured() {
		dctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		g, err := google.New(dctx, cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL())
		cancel()
		if err != nil {
			return err
		}
		provider = g
		log.Info("google sign-in enabled", "redirect_url", cfg.GoogleRedirectURL())
	} else {
		log.Warn("GOOGLE_CLIENT_ID/GOOGLE_CLIENT_SECRET not set: sign-in is disabled")
	}
	if len(cfg.AllowedEmails) > 0 {
		log.Info("sign-in restricted to allow-list", "count", len(cfg.AllowedEmails))
	}

	auth := cerberus.New(provider, db, cerberus.Options{
		PublicURL:     cfg.PublicURL,
		Secret:        cfg.SessionSecret,
		SessionTTL:    cfg.SessionTTL,
		AllowedEmails: cfg.AllowedEmails,
		SecureCookies: cfg.SecureCookies(),
	}, log)

	// Cerberus as OAuth server for assistants, Hermes as the MCP endpoint, Plutus as the first module.
	oauth := cerberus.NewOAuth(auth, db, cerberus.OAuthOptions{
		PublicURL: cfg.PublicURL,
		Secret:    cfg.SessionSecret,
		Scopes: []cerberus.Scope{
			{Name: plutus.ScopeRead, Description: "Read your spending records and summaries"},
			{Name: plutus.ScopeWrite, Description: "Add and delete your spending records"},
			{Name: themis.ScopeRead, Description: "Read your category list"},
			{Name: themis.ScopeWrite, Description: "Create, rename and merge your categories"},
			{Name: atlas.ScopeRead, Description: "Read your tasks, to-dos and notes"},
			{Name: atlas.ScopeWrite, Description: "Add, change and delete your tasks, to-dos and notes"},
		},
	}, log)
	mcp := hermes.New(hermes.Options{
		Name:                "nemosyne",
		Title:               "Nemosyne",
		Version:             version,
		Instructions:        hermes.DefaultInstructions,
		ResourceMetadataURL: oauth.MetadataURL(),
	}, mcpAuth{oauth: oauth}, log)
	plutus.New(db, time.Now).Register(mcp)
	themis.New(db).Register(mcp)
	atlas.New(db).Register(mcp)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		pctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(pctx); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	auth.Register(mux)
	oauth.Register(mux)
	mux.Handle("/mcp", mcp)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           secureHeaders(logRequests(log, mux)),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	go cleanupExpired(ctx, log, db, oauth)

	errc := make(chan error, 1)
	go func() {
		log.Info("api listening", "addr", cfg.Addr)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		log.Info("shutting down")
		sctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return srv.Shutdown(sctx)
	}
	return nil
}

func cleanupExpired(ctx context.Context, log *slog.Logger, db *store.Postgres, oauth *cerberus.OAuth) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n, err := db.DeleteExpiredSessions(ctx, time.Now())
			if err != nil {
				log.Error("session cleanup failed", "err", err)
			} else if n > 0 {
				log.Info("expired sessions removed", "count", n)
			}
			n, err = oauth.CleanupExpired(ctx)
			if err != nil {
				log.Error("oauth cleanup failed", "err", err)
			} else if n > 0 {
				log.Info("expired oauth codes and tokens removed", "count", n)
			}
		}
	}
}

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func logRequests(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		// Path only: never log the query string (it carries OAuth codes and state).
		log.Info("request", "method", r.Method, "path", r.URL.Path, "status", rec.status,
			"ms", time.Since(start).Milliseconds())
	})
}
