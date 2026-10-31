package cerberus

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

const (
	sessionCookie = "kuro_session"
	flowCookie    = "kuro_flow"
	flowTTL       = 10 * time.Minute
)

// Options configures a Handler.
type Options struct {
	PublicURL     string // external origin, no trailing slash
	Secret        []byte // signs the short-lived login-flow cookie
	SessionTTL    time.Duration
	AllowedEmails []string // lowercased; empty = anyone with a verified email
	SecureCookies bool
	Now           func() time.Time // optional, for tests
}

// Handler serves the sign-in HTTP endpoints.
type Handler struct {
	provider Provider // nil when Google is not configured
	store    Store
	opt      Options
	log      *slog.Logger
}

// New builds a Handler. provider may be nil, in which case sign-in returns 503.
func New(provider Provider, store Store, opt Options, log *slog.Logger) *Handler {
	if opt.Now == nil {
		opt.Now = time.Now
	}
	return &Handler{provider: provider, store: store, opt: opt, log: log}
}

// Register mounts the auth routes on mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/google/login", h.login)
	mux.HandleFunc("GET /api/auth/google/callback", h.callback)
	mux.HandleFunc("GET /api/auth/me", h.me)
	mux.HandleFunc("POST /api/auth/logout", h.logout)
}

// UserFromRequest returns the signed-in user for r, if any.
func (h *Handler) UserFromRequest(r *http.Request) (User, bool) {
	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" {
		return User{}, false
	}
	u, err := h.store.GetSessionUser(r.Context(), hashToken(c.Value), h.opt.Now())
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			h.log.Error("session lookup failed", "err", err)
		}
		return User{}, false
	}
	return u, true
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	if h.provider == nil {
		http.Error(w, "Google sign-in is not configured on this server", http.StatusServiceUnavailable)
		return
	}
	state, err1 := randString(32)
	nonce, err2 := randString(32)
	verifier, err3 := randString(32) // 43 chars, valid PKCE verifier
	if err := errors.Join(err1, err2, err3); err != nil {
		h.log.Error("random generation failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	sealed, err := sealFlow(h.opt.Secret, flowState{
		State:    state,
		Nonce:    nonce,
		Verifier: verifier,
		Next:     safeNext(r.URL.Query().Get("next")),
		Expires:  h.opt.Now().Add(flowTTL).Unix(),
	})
	if err != nil {
		h.log.Error("seal flow failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     flowCookie,
		Value:    sealed,
		Path:     "/api/auth/google",
		MaxAge:   int(flowTTL.Seconds()),
		HttpOnly: true,
		Secure:   h.opt.SecureCookies,
		SameSite: http.SameSiteLaxMode, // must be Lax: it has to survive the redirect back from Google
	})
	http.Redirect(w, r, h.provider.AuthCodeURL(state, nonce, verifier), http.StatusFound)
}

func (h *Handler) callback(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	if h.provider == nil {
		http.Error(w, "Google sign-in is not configured on this server", http.StatusServiceUnavailable)
		return
	}

	// The flow cookie is single-use: clear it whatever happens next.
	http.SetCookie(w, &http.Cookie{
		Name: flowCookie, Path: "/api/auth/google", MaxAge: -1,
		HttpOnly: true, Secure: h.opt.SecureCookies, SameSite: http.SameSiteLaxMode,
	})

	c, err := r.Cookie(flowCookie)
	if err != nil {
		h.fail(w, r, "invalid_flow")
		return
	}
	flow, err := openFlow(h.opt.Secret, c.Value, h.opt.Now())
	if err != nil {
		h.fail(w, r, "invalid_flow")
		return
	}

	q := r.URL.Query()
	if q.Get("error") != "" {
		h.fail(w, r, "access_denied")
		return
	}
	if subtle.ConstantTimeCompare([]byte(q.Get("state")), []byte(flow.State)) != 1 {
		h.fail(w, r, "invalid_flow")
		return
	}
	code := q.Get("code")
	if code == "" {
		h.fail(w, r, "invalid_flow")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	id, err := h.provider.Exchange(ctx, code, flow.Verifier, flow.Nonce)
	if err != nil {
		h.log.Warn("google exchange failed", "err", err)
		h.fail(w, r, "exchange_failed")
		return
	}
	if id.Subject == "" || id.Email == "" || !id.EmailVerified {
		h.fail(w, r, "email_unverified")
		return
	}
	if !h.allowed(id.Email) {
		h.log.Info("sign-in rejected: email not on allow-list")
		h.fail(w, r, "not_allowed")
		return
	}

	user, err := h.store.UpsertUser(ctx, id)
	if err != nil {
		h.log.Error("upsert user failed", "err", err)
		h.fail(w, r, "server_error")
		return
	}
	token, hash, err := newSessionToken()
	if err != nil {
		h.log.Error("session token failed", "err", err)
		h.fail(w, r, "server_error")
		return
	}
	expires := h.opt.Now().Add(h.opt.SessionTTL)
	if err := h.store.CreateSession(ctx, user.ID, hash, expires); err != nil {
		h.log.Error("create session failed", "err", err)
		h.fail(w, r, "server_error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(h.opt.SessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   h.opt.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, h.opt.PublicURL+flow.Next, http.StatusFound)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	u, ok := h.UserFromRequest(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	// Defence in depth on top of SameSite=Lax: refuse cross-origin POSTs.
	if o := r.Header.Get("Origin"); o != "" && o != h.opt.PublicURL {
		http.Error(w, "cross-origin request refused", http.StatusForbidden)
		return
	}
	if c, err := r.Cookie(sessionCookie); err == nil && c.Value != "" {
		if err := h.store.DeleteSession(r.Context(), hashToken(c.Value)); err != nil {
			h.log.Error("delete session failed", "err", err)
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: h.opt.SecureCookies, SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) allowed(email string) bool {
	if len(h.opt.AllowedEmails) == 0 {
		return true
	}
	return slices.Contains(h.opt.AllowedEmails, strings.ToLower(email))
}

// fail sends the browser back to the web app with a short, non-sensitive error code.
func (h *Handler) fail(w http.ResponseWriter, r *http.Request, code string) {
	http.Redirect(w, r, h.opt.PublicURL+"/?auth_error="+url.QueryEscape(code), http.StatusFound)
}

func noStore(w http.ResponseWriter) { w.Header().Set("Cache-Control", "no-store") }

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
