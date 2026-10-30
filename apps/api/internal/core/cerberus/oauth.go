package cerberus

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
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

const consentDomain = "nemosyne-oauth-consent-v1\x00"

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

// CleanupExpired removes expired codes and tokens.
func (o *OAuth) CleanupExpired(ctx context.Context) (int64, error) {
	return o.store.DeleteExpiredOAuth(ctx, o.opt.Now())
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
		"resource_name":            "Nemosyne",
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

func (o *OAuth) register(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	noStore(w)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var req struct {
		RedirectURIs            []string `json:"redirect_uris"`
		ClientName              string   `json:"client_name"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
		GrantTypes              []string `json:"grant_types"`
		ResponseTypes           []string `json:"response_types"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		oauthError(w, http.StatusBadRequest, "invalid_client_metadata", "request body must be a JSON object")
		return
	}
	if n := len(req.RedirectURIs); n == 0 || n > 5 {
		oauthError(w, http.StatusBadRequest, "invalid_redirect_uri", "provide between 1 and 5 redirect_uris")
		return
	}
	for _, u := range req.RedirectURIs {
		if !validRedirectURI(u) {
			oauthError(w, http.StatusBadRequest, "invalid_redirect_uri", "redirect_uris must be https, loopback http, or a custom app scheme without fragments")
			return
		}
	}
	if m := req.TokenEndpointAuthMethod; m != "" && m != "none" {
		oauthError(w, http.StatusBadRequest, "invalid_client_metadata", "only public clients (token_endpoint_auth_method \"none\") are supported")
		return
	}
	for _, g := range req.GrantTypes {
		if g != "authorization_code" && g != "refresh_token" {
			oauthError(w, http.StatusBadRequest, "invalid_client_metadata", "unsupported grant_type "+cleanText(g, 40))
			return
		}
	}
	for _, rt := range req.ResponseTypes {
		if rt != "code" {
			oauthError(w, http.StatusBadRequest, "invalid_client_metadata", "only response_type \"code\" is supported")
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
	c := OAuthClient{ID: "nem_c_" + id, Name: name, RedirectURIs: req.RedirectURIs, CreatedAt: o.opt.Now()}
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

type authzErr struct {
	code, desc string
	redirect   bool // true when it is safe to send the error back to the client's redirect_uri
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
		return req, &authzErr{"invalid_request", "client_id and redirect_uri are required", false}
	}
	client, err := o.store.GetClient(ctx, req.ClientID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return req, &authzErr{"invalid_client", "unknown client_id", false}
		}
		o.log.Error("get oauth client failed", "err", err)
		return req, &authzErr{"server_error", "temporary failure", false}
	}
	req.Client = client
	if !slices.Contains(client.RedirectURIs, req.RedirectURI) {
		return req, &authzErr{"invalid_request", "redirect_uri is not registered for this client", false}
	}

	// From here on the redirect_uri is trusted, so errors go back to the client.
	if v.Get("response_type") != "code" {
		return req, &authzErr{"unsupported_response_type", "response_type must be code", true}
	}
	if v.Get("code_challenge_method") != "S256" || !reChallenge.MatchString(req.Challenge) {
		return req, &authzErr{"invalid_request", "PKCE with code_challenge_method=S256 is required", true}
	}
	scopes, ok := o.parseScopes(req.ScopeRaw)
	if !ok {
		return req, &authzErr{"invalid_scope", "unknown scope requested", true}
	}
	req.Scopes = scopes
	if req.Resource != "" && req.Resource != o.Resource() {
		return req, &authzErr{"invalid_target", "resource must be " + o.Resource(), true}
	}
	return req, nil
}

// parseScopes validates a space-separated scope list. An empty list means every supported scope.
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
	o.renderConsent(w, user, req)
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
			Scopes:        req.Scopes,
			Resource:      resource,
			ExpiresAt:     o.opt.Now().Add(5 * time.Minute),
		})
		if err != nil {
			o.log.Error("save auth code failed", "err", err)
			o.errorPage(w, http.StatusInternalServerError, "Temporary failure. Please try again.")
			return
		}
		o.log.Info("oauth consent approved", "client_id", req.ClientID, "scopes", strings.Join(req.Scopes, " "))
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
	if !e.redirect {
		o.errorPage(w, http.StatusBadRequest, e.desc)
		return
	}
	o.redirect(w, r, req, map[string]string{"error": e.code, "error_description": e.desc})
}

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

var consentTmpl = template.Must(template.New("consent").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Connect to Nemosyne</title>
<style>
body{margin:0;background:#f7f4ee;color:#1d1b16;font:16px/1.5 system-ui,-apple-system,"Segoe UI",Roboto,sans-serif}
main{max-width:30rem;margin:3rem auto;padding:0 1.25rem}
h1{font:600 1.6rem Georgia,serif;margin:0 0 .25rem}
.card{background:#fff;border:1px solid #e4ded0;border-radius:.75rem;padding:1.25rem;margin:1rem 0}
.muted{color:#6b665b;font-size:.9rem}
.warn{background:#fdf1e3;color:#7a4310;border-radius:.5rem;padding:.6rem .9rem;font-size:.9rem}
ul{padding-left:1.2rem;margin:.5rem 0}
.row{display:flex;gap:.75rem;margin-top:1rem}
button{font:inherit;border-radius:.5rem;padding:.6rem 1rem;border:1px solid #2f4a8a;cursor:pointer}
.approve{background:#2f4a8a;color:#fff}.deny{background:transparent;color:#2f4a8a}
code{background:#f1ede3;padding:.1rem .3rem;border-radius:.25rem;word-break:break-all}
</style></head><body><main>
<h1>Nemosyne</h1>
<p class="muted">Signed in as {{.Email}}</p>
<div class="card">
<p><strong>{{.ClientName}}</strong> wants to connect to your Nemosyne account.</p>
<p class="muted">It will be able to:</p>
<ul>{{range .Scopes}}<li>{{.Description}}</li>{{end}}</ul>
<p class="muted">After you decide, you will be sent to <code>{{.Redirect}}</code>.</p>
<p class="warn">Only approve if you started this connection yourself. The app name is chosen by the app, not verified by us.</p>
<form method="post" action="/api/oauth/authorize">
<input type="hidden" name="response_type" value="code">
<input type="hidden" name="client_id" value="{{.ClientID}}">
<input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
<input type="hidden" name="code_challenge" value="{{.Challenge}}">
<input type="hidden" name="code_challenge_method" value="S256">
<input type="hidden" name="state" value="{{.State}}">
<input type="hidden" name="scope" value="{{.Scope}}">
<input type="hidden" name="resource" value="{{.Resource}}">
<input type="hidden" name="csrf" value="{{.CSRF}}">
<div class="row">
<button class="approve" type="submit" name="decision" value="approve">Approve</button>
<button class="deny" type="submit" name="decision" value="deny">Deny</button>
</div></form></div></main></body></html>`))

var errorTmpl = template.Must(template.New("error").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Nemosyne</title>
<style>body{margin:0;background:#f7f4ee;color:#1d1b16;font:16px/1.5 system-ui,sans-serif}
main{max-width:30rem;margin:3rem auto;padding:0 1.25rem}h1{font:600 1.6rem Georgia,serif}</style></head>
<body><main><h1>Nemosyne</h1><p>{{.}}</p></main></body></html>`))

func securePageHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; frame-ancestors 'none'")
}

func (o *OAuth) renderConsent(w http.ResponseWriter, user User, req authzReq) {
	var scopes []Scope
	for _, name := range req.Scopes {
		for _, s := range o.opt.Scopes {
			if s.Name == name {
				scopes = append(scopes, s)
			}
		}
	}
	securePageHeaders(w)
	err := consentTmpl.Execute(w, map[string]any{
		"Email":       user.Email,
		"ClientName":  req.Client.Name,
		"Scopes":      scopes,
		"Redirect":    redirectLabel(req.RedirectURI),
		"ClientID":    req.ClientID,
		"RedirectURI": req.RedirectURI,
		"Challenge":   req.Challenge,
		"State":       req.State,
		"Scope":       strings.Join(req.Scopes, " "),
		"Resource":    req.Resource,
		"CSRF":        o.csrf(user.ID, req),
	})
	if err != nil {
		o.log.Error("render consent failed", "err", err)
	}
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
	client, err := o.store.GetClient(r.Context(), f.Get("client_id"))
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
