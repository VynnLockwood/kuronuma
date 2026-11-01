package cerberus

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
)

var (
	reChallenge = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)       // base64url(SHA-256) is always 43 chars
	reVerifier  = regexp.MustCompile(`^[A-Za-z0-9._~-]{43,128}$`) // RFC 7636
	reScheme    = regexp.MustCompile(`^[a-z][a-z0-9+.-]*$`)
)

const consentDomain = "kuronuma-oauth-consent-v1\x00"

// OAuthOptions configures the OAuth authorization server.
type OAuthOptions struct {
	PublicURL  string // external origin, no trailing slash; also the issuer
	Secret     []byte // derives consent-form CSRF tokens
	Scopes     []Scope
	AccessTTL  time.Duration    // default 1h
	RefreshTTL time.Duration    // default 30d
	Now        func() time.Time // optional, for tests
}

// OAuth is Cerberus acting as an OAuth 2.1 authorization server for MCP clients.
// Users are authenticated through the normal Google sign-in session.
type OAuth struct {
	auth  *Handler
	store OAuthStore
	opt   OAuthOptions
	log   *slog.Logger
}

// NewOAuth builds the authorization server.
func NewOAuth(auth *Handler, store OAuthStore, opt OAuthOptions, log *slog.Logger) *OAuth {
	if opt.Now == nil {
		opt.Now = time.Now
	}
	if opt.AccessTTL == 0 {
		opt.AccessTTL = time.Hour
	}
	if opt.RefreshTTL == 0 {
		opt.RefreshTTL = 30 * 24 * time.Hour
	}
	return &OAuth{auth: auth, store: store, opt: opt, log: log}
}

// Resource is the canonical URL of the MCP endpoint (the OAuth "resource").
func (o *OAuth) Resource() string { return o.opt.PublicURL + "/mcp" }

// MetadataURL is where clients discover how to authorize against the MCP endpoint.
func (o *OAuth) MetadataURL() string {
	return o.opt.PublicURL + "/.well-known/oauth-protected-resource"
}

// Register mounts the OAuth routes on mux.
func (o *OAuth) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /.well-known/oauth-protected-resource", o.protectedResource)
	mux.HandleFunc("GET /.well-known/oauth-protected-resource/mcp", o.protectedResource)
	mux.HandleFunc("GET /.well-known/oauth-authorization-server", o.serverMetadata)
	mux.HandleFunc("POST /api/oauth/register", o.register)
	mux.HandleFunc("GET /api/oauth/authorize", o.authorizeGET)
	mux.HandleFunc("POST /api/oauth/authorize", o.authorizePOST)
	mux.HandleFunc("POST /api/oauth/token", o.token)
	for _, p := range []string{
		"/.well-known/oauth-protected-resource",
		"/.well-known/oauth-protected-resource/mcp",
		"/.well-known/oauth-authorization-server",
		"/api/oauth/register",
		"/api/oauth/token",
	} {
		mux.HandleFunc("OPTIONS "+p, preflight)
	}
}

// VerifyAccessToken checks a bearer token and returns who it belongs to.
// It returns ErrNotFound when the token is unknown or expired.
func (o *OAuth) VerifyAccessToken(ctx context.Context, token string) (TokenInfo, error) {
	t, err := o.store.GetToken(ctx, hashToken(token), TokenAccess, o.opt.Now())
	if err != nil {
		return TokenInfo{}, err
	}
	return TokenInfo{UserID: t.UserID, ClientID: t.ClientID, Scopes: t.Scopes}, nil
}

// unusedClientAge is how old a registered client must be before it can be removed for holding nothing. Registration
// is open, so without this clients that were never used would accumulate forever. It is long on purpose: a client
// also holds nothing once its owner has been away longer than a refresh token lasts (30 days), and that user has to
// remove and re-add the connector, which the error page says (see authzHint).
const unusedClientAge = 180 * 24 * time.Hour

// CleanupExpired removes expired codes and tokens, then clients that are over 180 days old and hold neither.
func (o *OAuth) CleanupExpired(ctx context.Context) (int64, error) {
	now := o.opt.Now()
	n, err := o.store.DeleteExpiredOAuth(ctx, now)
	if err != nil {
		return n, err
	}
	c, err := o.store.DeleteUnusedClients(ctx, now.Add(-unusedClientAge))
	return n + c, err
}

// ---- discovery ----

func (o *OAuth) scopeNames() []string {
	names := make([]string, 0, len(o.opt.Scopes))
	for _, s := range o.opt.Scopes {
		names = append(names, s.Name)
	}
	return names
}

func (o *OAuth) protectedResource(w http.ResponseWriter, _ *http.Request) {
	setCORS(w)
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, map[string]any{
		"resource":                 o.Resource(),
		"authorization_servers":    []string{o.opt.PublicURL},
		"scopes_supported":         o.scopeNames(),
		"bearer_methods_supported": []string{"header"},
		"resource_name":            "Kuronuma",
	})
}

func (o *OAuth) serverMetadata(w http.ResponseWriter, _ *http.Request) {
	setCORS(w)
	w.Header().Set("Cache-Control", "public, max-age=300")
	base := o.opt.PublicURL
	writeJSON(w, http.StatusOK, map[string]any{
		"issuer":                                         base,
		"authorization_endpoint":                         base + "/api/oauth/authorize",
		"token_endpoint":                                 base + "/api/oauth/token",
		"registration_endpoint":                          base + "/api/oauth/register",
		"scopes_supported":                               o.scopeNames(),
		"response_types_supported":                       []string{"code"},
		"grant_types_supported":                          []string{"authorization_code", "refresh_token"},
		"code_challenge_methods_supported":               []string{"S256"},
		"token_endpoint_auth_methods_supported":          []string{"none"},
		"authorization_response_iss_parameter_supported": true,
	})
}

// ---- dynamic client registration (RFC 7591) ----

// maxRedirectURIs is how many redirect URIs one registration may list. Gemini's connector registers six (Google
// documents six app-specific callbacks), so five was too few.
const maxRedirectURIs = 10

// acceptedAuthMethods are the token_endpoint_auth_method values a registration may ask for. Every client is
// registered as "none"; the secret methods are accepted so server-side clients can register at all.
var acceptedAuthMethods = []string{"none", "client_secret_post", "client_secret_basic"}

// registerRequest is the RFC 7591 client metadata this server reads; other fields are ignored.
type registerRequest struct {
	RedirectURIs            []string `json:"redirect_uris"`
	ClientName              string   `json:"client_name"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
}

// rejectRegistration answers a refused registration with 400 and says why in the server log, so an incompatible
// client can be diagnosed. Only coarse facts are logged: the error code and reason, the requested login method,
// grant and response types, how many redirect URIs, and their origins (scheme and host: never paths or query
// strings, which can hold secrets). The client name is not logged. req is nil when the body could not be read.
func (o *OAuth) rejectRegistration(w http.ResponseWriter, req *registerRequest, code, desc string) {
	attrs := []any{"code", code, "reason", desc}
	if req != nil {
		attrs = append(attrs,
			"auth_method", cleanText(req.TokenEndpointAuthMethod, 40),
			"grant_types", cleanList(req.GrantTypes),
			"response_types", cleanList(req.ResponseTypes),
			"redirect_count", len(req.RedirectURIs),
			"redirect_origins", redirectOrigins(req.RedirectURIs))
	}
	o.log.Warn("oauth client registration rejected", attrs...)
	oauthError(w, http.StatusBadRequest, code, desc)
}

// cleanList cleans at most the first ten values of a client-supplied list for logging.
func cleanList(in []string) []string {
	out := make([]string, 0, min(len(in), 10))
	for i, v := range in {
		if i == 10 {
			break
		}
		out = append(out, cleanText(v, 40))
	}
	return out
}

// redirectOrigins reduces redirect URIs to scheme and host (custom app schemes to "scheme://") for logging.
func redirectOrigins(uris []string) []string {
	out := make([]string, 0, min(len(uris), 10))
	for i, s := range uris {
		if i == 10 {
			break
		}
		u, err := url.Parse(s)
		switch {
		case err != nil || u.Scheme == "":
			out = append(out, "(unparseable)")
		case u.Host == "":
			out = append(out, cleanText(u.Scheme, 40)+"://")
		default:
			out = append(out, cleanText(u.Scheme+"://"+u.Host, 100))
		}
	}
	return out
}

func (o *OAuth) register(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	noStore(w)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		o.rejectRegistration(w, nil, "invalid_client_metadata", "request body must be a JSON object")
		return
	}
	if n := len(req.RedirectURIs); n == 0 || n > maxRedirectURIs {
		o.rejectRegistration(w, &req, "invalid_redirect_uri", fmt.Sprintf("provide between 1 and %d redirect_uris", maxRedirectURIs))
		return
	}
	for _, u := range req.RedirectURIs {
		if !validRedirectURI(u) {
			o.rejectRegistration(w, &req, "invalid_redirect_uri", "redirect_uris must be https, loopback http, or a custom app scheme without fragments")
			return
		}
	}
	// Server-side clients such as Gemini ask for a client secret. RFC 7591 section 3.2.1 lets the server answer with
	// the values it actually uses, so they are registered as public clients (no secret is issued, PKCE protects the
	// code) and the response says "none". Key-based methods cannot be served and are refused.
	if m := req.TokenEndpointAuthMethod; m != "" && !slices.Contains(acceptedAuthMethods, m) {
		o.rejectRegistration(w, &req, "invalid_client_metadata", "supported token_endpoint_auth_method values are none, client_secret_post and client_secret_basic; clients are registered as public clients and no secret is issued")
		return
	}
	for _, g := range req.GrantTypes {
		if g != "authorization_code" && g != "refresh_token" {
			o.rejectRegistration(w, &req, "invalid_client_metadata", "unsupported grant_type "+cleanText(g, 40))
			return
		}
	}
	for _, rt := range req.ResponseTypes {
		if rt != "code" {
			o.rejectRegistration(w, &req, "invalid_client_metadata", "only response_type \"code\" is supported")
			return
		}
	}

	id, err := randString(16)
	if err != nil {
		oauthError(w, http.StatusInternalServerError, "server_error", "temporary failure")
		return
	}
	name := cleanText(req.ClientName, 100)
	if name == "" {
		name = "Unnamed app"
	}
	c := OAuthClient{ID: "kuro_c_" + id, Name: name, RedirectURIs: req.RedirectURIs, CreatedAt: o.opt.Now()}
	if err := o.store.CreateClient(r.Context(), c); err != nil {
		o.log.Error("create oauth client failed", "err", err)
		oauthError(w, http.StatusInternalServerError, "server_error", "temporary failure")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"client_id":                  c.ID,
		"client_id_issued_at":        c.CreatedAt.Unix(),
		"client_name":                c.Name,
		"redirect_uris":              c.RedirectURIs,
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "none",
	})
}

// validRedirectURI accepts https URLs, http on loopback (native apps), and custom
// app schemes. It rejects fragments, userinfo and script-capable schemes.
func validRedirectURI(s string) bool {
	if s == "" || len(s) > 512 || strings.Contains(s, "#") {
		return false
	}
	for _, r := range s {
		if r < 0x21 || r == 0x7f {
			return false
		}
	}
	u, err := url.Parse(s)
	if err != nil || u.User != nil {
		return false
	}
	switch u.Scheme {
	case "https":
		return u.Hostname() != ""
	case "http":
		h := u.Hostname()
		return h == "localhost" || h == "127.0.0.1" || h == "::1"
	case "", "javascript", "data", "file", "vbscript", "about", "blob", "ftp":
		return false
	default:
		return reScheme.MatchString(u.Scheme)
	}
}

// ---- authorization endpoint ----

type authzReq struct {
	ClientID    string
	RedirectURI string
	State       string
	Challenge   string
	ScopeRaw    string
	Resource    string
	Scopes      []string
	Client      OAuthClient
}

// authzErr is a problem found before the user has signed in and approved. It is always shown on our own error page,
// never sent back to the client: client registration is open, so a registered redirect_uri only means that someone
// registered it, and redirecting a visitor there would make this server an open redirector. (After the user decides,
// a denial does go back to the client; that is their own choice.)
type authzErr struct {
	code, desc string
}

func (o *OAuth) parseAuthz(ctx context.Context, v url.Values) (authzReq, *authzErr) {
	req := authzReq{
		ClientID:    v.Get("client_id"),
		RedirectURI: v.Get("redirect_uri"),
		State:       v.Get("state"),
		Challenge:   v.Get("code_challenge"),
		ScopeRaw:    v.Get("scope"),
		Resource:    v.Get("resource"),
	}
	if req.ClientID == "" || req.RedirectURI == "" {
		return req, &authzErr{"invalid_request", "client_id and redirect_uri are required"}
	}
	client, err := o.store.GetClient(ctx, req.ClientID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return req, &authzErr{"invalid_client", "unknown client_id"}
		}
		o.log.Error("get oauth client failed", "err", err)
		return req, &authzErr{"server_error", "temporary failure"}
	}
	req.Client = client
	if !slices.Contains(client.RedirectURIs, req.RedirectURI) {
		return req, &authzErr{"invalid_request", "redirect_uri is not registered for this client"}
	}

	// The redirect_uri is registered, but that is not trust (registration is open), so errors stay on our page.
	if v.Get("response_type") != "code" {
		return req, &authzErr{"unsupported_response_type", "response_type must be code"}
	}
	if v.Get("code_challenge_method") != "S256" || !reChallenge.MatchString(req.Challenge) {
		return req, &authzErr{"invalid_request", "PKCE with code_challenge_method=S256 is required"}
	}
	scopes, ok := o.parseScopes(req.ScopeRaw)
	if !ok {
		return req, &authzErr{"invalid_scope", "unknown scope requested: " + cleanText(strings.Join(o.unknownScopes(req.ScopeRaw), " "), 100)}
	}
	req.Scopes = scopes
	if req.Resource != "" && req.Resource != o.Resource() {
		return req, &authzErr{"invalid_target", "resource must be " + o.Resource()}
	}
	return req, nil
}

// parseScopes validates a space-separated scope list. An empty list means every supported scope.
// unknownScopes lists the requested scopes this server does not offer.
func (o *OAuth) unknownScopes(raw string) []string {
	supported := o.scopeNames()
	var out []string
	for _, s := range strings.Fields(raw) {
		if !slices.Contains(supported, s) && !slices.Contains(out, s) {
			out = append(out, s)
		}
	}
	return out
}

func (o *OAuth) parseScopes(raw string) ([]string, bool) {
	supported := o.scopeNames()
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return slices.Clone(supported), true
	}
	var out []string
	for _, s := range fields {
		if !slices.Contains(supported, s) {
			return nil, false
		}
		if !slices.Contains(out, s) {
			out = append(out, s)
		}
	}
	return out, true
}

func (o *OAuth) authorizeGET(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	req, aerr := o.parseAuthz(r.Context(), r.URL.Query())
	if aerr != nil {
		o.fail(w, r, req, aerr)
		return
	}
	user, ok := o.auth.UserFromRequest(r)
	if !ok {
		// Sign in with Google first, then come back to this exact authorize URL.
		http.Redirect(w, r, "/api/auth/google/login?next="+url.QueryEscape(r.URL.RequestURI()), http.StatusFound)
		return
	}
	o.renderConsent(w, r, user, req)
}

func (o *OAuth) authorizePOST(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	if origin := r.Header.Get("Origin"); origin != "" && origin != o.opt.PublicURL {
		o.errorPage(w, http.StatusForbidden, "Cross-origin request refused.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if err := r.ParseForm(); err != nil {
		o.errorPage(w, http.StatusBadRequest, "Malformed request.")
		return
	}
	user, ok := o.auth.UserFromRequest(r)
	if !ok {
		o.errorPage(w, http.StatusUnauthorized, "Your session expired. Start the connection again from your assistant.")
		return
	}
	req, aerr := o.parseAuthz(r.Context(), r.PostForm)
	if aerr != nil {
		o.fail(w, r, req, aerr)
		return
	}
	if !hmac.Equal([]byte(r.PostForm.Get("csrf")), []byte(o.csrf(user.ID, req))) {
		o.errorPage(w, http.StatusForbidden, "This approval form is invalid or expired. Start the connection again.")
		return
	}

	switch r.PostForm.Get("decision") {
	case "deny":
		o.redirect(w, r, req, map[string]string{"error": "access_denied", "error_description": "The user denied the request"})
	case "approve":
		// The access level only narrows what was asked for. The granted scopes come from the request that was
		// checked above (the CSRF token covers it), never from a list in this form.
		granted, access := req.Scopes, r.PostForm.Get("access")
		switch access {
		case "", "full":
			access = "full"
		case "read":
			if granted = readOnlyScopes(req.Scopes); len(granted) == 0 {
				o.errorPage(w, http.StatusBadRequest, "Read-only access is not possible for this request. Go back to your assistant and start the connection again.")
				return
			}
		default:
			o.errorPage(w, http.StatusBadRequest, "Unknown access level.")
			return
		}
		code, err := randString(32)
		if err != nil {
			o.errorPage(w, http.StatusInternalServerError, "Temporary failure. Please try again.")
			return
		}
		resource := req.Resource
		if resource == "" {
			resource = o.Resource()
		}
		err = o.store.SaveCode(r.Context(), AuthCode{
			Hash:          hashToken(code),
			ClientID:      req.ClientID,
			UserID:        user.ID,
			RedirectURI:   req.RedirectURI,
			CodeChallenge: req.Challenge,
			Scopes:        granted,
			Resource:      resource,
			ExpiresAt:     o.opt.Now().Add(5 * time.Minute),
		})
		if err != nil {
			o.log.Error("save auth code failed", "err", err)
			o.errorPage(w, http.StatusInternalServerError, "Temporary failure. Please try again.")
			return
		}
		o.log.Info("oauth consent approved", "client_id", req.ClientID, "access", access, "scopes", strings.Join(granted, " "))
		o.redirect(w, r, req, map[string]string{"code": code})
	default:
		o.errorPage(w, http.StatusBadRequest, "Unknown decision.")
	}
}

// csrf binds the consent form to this user and these exact request parameters.
func (o *OAuth) csrf(userID string, req authzReq) string {
	m := hmac.New(sha256.New, o.opt.Secret)
	m.Write([]byte(consentDomain))
	for _, part := range []string{userID, req.ClientID, req.RedirectURI, req.Challenge, strings.Join(req.Scopes, " "), req.State, req.Resource} {
		m.Write([]byte(part))
		m.Write([]byte{0})
	}
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func (o *OAuth) fail(w http.ResponseWriter, r *http.Request, req authzReq, e *authzErr) {
	// Say why in the log so an incompatible client can be diagnosed. The state, PKCE challenge and full redirect URI
	// are never logged; the redirect is reduced to its origin.
	o.log.Warn("oauth authorize rejected", "code", e.code, "reason", e.desc, "client_id", cleanText(req.ClientID, 60),
		"redirect_origin", redirectOrigins([]string{req.RedirectURI})[0])
	o.errorPage(w, http.StatusBadRequest, e.desc+" "+authzHint)
}

// authzHint tells the person on an error page how to recover. The assistant that started the connection never
// receives these errors, so only the person can act on them. It also covers a client that was removed as unused.
const authzHint = "Go back to your assistant and start the connection again. If it keeps failing, remove Kuronuma from its connectors and add it again."

func (o *OAuth) redirect(w http.ResponseWriter, r *http.Request, req authzReq, params map[string]string) {
	params["state"] = req.State
	params["iss"] = o.opt.PublicURL
	http.Redirect(w, r, buildRedirect(req.RedirectURI, params), http.StatusSeeOther)
}

func buildRedirect(base string, params map[string]string) string {
	u, err := url.Parse(base)
	if err != nil {
		return base
	}
	q := u.Query()
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// ---- consent and error pages ----

func securePageHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; img-src 'self'; frame-ancestors 'none'")
}

func (o *OAuth) errorPage(w http.ResponseWriter, status int, msg string) {
	securePageHeaders(w)
	w.WriteHeader(status)
	_ = errorTmpl.Execute(w, msg)
}

// redirectLabel shows where the user will land: host for web URLs, scheme for app URLs.
func redirectLabel(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		return "the app"
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		return u.Scheme + "://" + u.Host
	}
	return u.Scheme + "://"
}

// ---- token endpoint ----

func (o *OAuth) token(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	if err := r.ParseForm(); err != nil {
		oauthError(w, http.StatusBadRequest, "invalid_request", "malformed form body")
		return
	}
	f := r.PostForm
	clientID, ok := tokenClientID(r, f)
	if !ok {
		oauthError(w, http.StatusUnauthorized, "invalid_client", "malformed client authentication")
		return
	}
	client, err := o.store.GetClient(r.Context(), clientID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			oauthError(w, http.StatusUnauthorized, "invalid_client", "unknown client_id")
			return
		}
		o.log.Error("get oauth client failed", "err", err)
		oauthError(w, http.StatusInternalServerError, "server_error", "temporary failure")
		return
	}
	switch f.Get("grant_type") {
	case "authorization_code":
		o.grantCode(w, r, client, f)
	case "refresh_token":
		o.grantRefresh(w, r, client, f)
	default:
		oauthError(w, http.StatusBadRequest, "unsupported_grant_type", "use authorization_code or refresh_token")
	}
}

// tokenClientID is the client a token request names: the client_id form field, or the user name of an HTTP Basic
// Authorization header (client_secret_basic). The password is ignored on purpose: no secret is ever issued, so a
// client is identified by its id and held to PKCE and the binding of each code and refresh token to that client.
// When both the header and the form name a client they must agree. ok is false for a header that cannot be read.
func tokenClientID(r *http.Request, f url.Values) (id string, ok bool) {
	id = f.Get("client_id")
	scheme, cred, found := strings.Cut(r.Header.Get("Authorization"), " ")
	if !found || !strings.EqualFold(scheme, "Basic") {
		return id, true
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(cred))
	if err != nil {
		return "", false
	}
	user, _, hasColon := strings.Cut(string(raw), ":")
	if !hasColon {
		return "", false
	}
	// RFC 6749 section 2.3.1: the user name is form-urlencoded before it is base64 encoded.
	user, err = url.QueryUnescape(user)
	if err != nil || user == "" {
		return "", false
	}
	if id != "" && id != user {
		return "", false
	}
	return user, true
}

func (o *OAuth) grantCode(w http.ResponseWriter, r *http.Request, client OAuthClient, f url.Values) {
	verifier := f.Get("code_verifier")
	if !reVerifier.MatchString(verifier) {
		oauthError(w, http.StatusBadRequest, "invalid_grant", "code_verifier is missing or malformed")
		return
	}
	// Consuming first makes every code single-use, even if the PKCE check below fails.
	ac, err := o.store.ConsumeCode(r.Context(), hashToken(f.Get("code")), o.opt.Now())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			oauthError(w, http.StatusBadRequest, "invalid_grant", "code is invalid, expired or already used")
			return
		}
		o.log.Error("consume code failed", "err", err)
		oauthError(w, http.StatusInternalServerError, "server_error", "temporary failure")
		return
	}
	if ac.ClientID != client.ID || ac.RedirectURI != f.Get("redirect_uri") {
		oauthError(w, http.StatusBadRequest, "invalid_grant", "code was not issued to this client and redirect_uri")
		return
	}
	if !pkceMatches(verifier, ac.CodeChallenge) {
		oauthError(w, http.StatusBadRequest, "invalid_grant", "PKCE verification failed")
		return
	}
	if res := f.Get("resource"); res != "" && res != ac.Resource {
		oauthError(w, http.StatusBadRequest, "invalid_target", "resource does not match the authorization")
		return
	}
	o.issue(w, r, client.ID, ac.UserID, ac.Scopes)
}

func (o *OAuth) grantRefresh(w http.ResponseWriter, r *http.Request, client OAuthClient, f url.Values) {
	rt, err := o.store.ConsumeRefreshToken(r.Context(), hashToken(f.Get("refresh_token")), o.opt.Now())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			oauthError(w, http.StatusBadRequest, "invalid_grant", "refresh_token is invalid, expired or already used")
			return
		}
		o.log.Error("consume refresh token failed", "err", err)
		oauthError(w, http.StatusInternalServerError, "server_error", "temporary failure")
		return
	}
	if rt.ClientID != client.ID {
		oauthError(w, http.StatusBadRequest, "invalid_grant", "refresh_token was not issued to this client")
		return
	}
	scopes := rt.Scopes
	if raw := f.Get("scope"); raw != "" {
		scopes = strings.Fields(raw)
		for _, s := range scopes {
			if !slices.Contains(rt.Scopes, s) {
				oauthError(w, http.StatusBadRequest, "invalid_scope", "cannot widen scope on refresh")
				return
			}
		}
	}
	o.issue(w, r, client.ID, rt.UserID, scopes)
}

func (o *OAuth) issue(w http.ResponseWriter, r *http.Request, clientID, userID string, scopes []string) {
	now := o.opt.Now()
	access, accessHash, err1 := newSessionToken()
	refresh, refreshHash, err2 := newSessionToken()
	if err := errors.Join(err1, err2); err != nil {
		oauthError(w, http.StatusInternalServerError, "server_error", "temporary failure")
		return
	}
	for _, t := range []Token{
		{Hash: accessHash, Kind: TokenAccess, ClientID: clientID, UserID: userID, Scopes: scopes, ExpiresAt: now.Add(o.opt.AccessTTL)},
		{Hash: refreshHash, Kind: TokenRefresh, ClientID: clientID, UserID: userID, Scopes: scopes, ExpiresAt: now.Add(o.opt.RefreshTTL)},
	} {
		if err := o.store.SaveToken(r.Context(), t); err != nil {
			o.log.Error("save token failed", "err", err)
			oauthError(w, http.StatusInternalServerError, "server_error", "temporary failure")
			return
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  access,
		"token_type":    "Bearer",
		"expires_in":    int(o.opt.AccessTTL.Seconds()),
		"refresh_token": refresh,
		"scope":         strings.Join(scopes, " "),
	})
}

func pkceMatches(verifier, challenge string) bool {
	sum := sha256.Sum256([]byte(verifier))
	got := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(got), []byte(challenge)) == 1
}

// ---- helpers ----

func oauthError(w http.ResponseWriter, status int, code, desc string) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	writeJSON(w, status, map[string]string{"error": code, "error_description": desc})
}

// setCORS lets browser-based MCP clients use the public, cookie-less OAuth endpoints.
func setCORS(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, MCP-Protocol-Version")
}

func preflight(w http.ResponseWriter, _ *http.Request) {
	setCORS(w)
	w.WriteHeader(http.StatusNoContent)
}

// cleanText strips control and invisible formatting characters, collapses whitespace and truncates.
func cleanText(s string, max int) string {
	var b strings.Builder
	n, pending := 0, false
	for _, r := range s {
		switch {
		case unicode.IsControl(r) || unicode.IsSpace(r):
			pending = true
			continue
		case unicode.Is(unicode.Cf, r), r == unicode.ReplacementChar:
			continue
		}
		need := 1
		if pending && n > 0 {
			need = 2
		}
		if n+need > max {
			break
		}
		if pending && n > 0 {
			b.WriteByte(' ')
			n++
		}
		pending = false
		b.WriteRune(r)
		n++
	}
	return b.String()
}
