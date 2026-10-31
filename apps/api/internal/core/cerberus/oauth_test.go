package cerberus

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	testVerifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	testRedirect = "http://localhost:7777/callback"
)

func quietLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func challengeFor(v string) string {
	sum := sha256.Sum256([]byte(v))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// ---- in-memory OAuth store ----

type memOAuthStore struct {
	mu      sync.Mutex
	clients map[string]OAuthClient
	codes   map[string]AuthCode
	tokens  map[string]Token // key: kind + hex(hash)
}

func newMemOAuthStore() *memOAuthStore {
	return &memOAuthStore{clients: map[string]OAuthClient{}, codes: map[string]AuthCode{}, tokens: map[string]Token{}}
}

func (m *memOAuthStore) CreateClient(_ context.Context, c OAuthClient) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[c.ID] = c
	return nil
}

func (m *memOAuthStore) GetClient(_ context.Context, id string) (OAuthClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.clients[id]
	if !ok {
		return OAuthClient{}, ErrNotFound
	}
	return c, nil
}

func (m *memOAuthStore) SaveCode(_ context.Context, c AuthCode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[hex.EncodeToString(c.Hash)] = c
	return nil
}

func (m *memOAuthStore) ConsumeCode(_ context.Context, h []byte, now time.Time) (AuthCode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := hex.EncodeToString(h)
	c, ok := m.codes[k]
	if !ok || !c.ExpiresAt.After(now) {
		return AuthCode{}, ErrNotFound
	}
	delete(m.codes, k)
	return c, nil
}

func tokenKey(kind TokenKind, h []byte) string { return string(kind) + ":" + hex.EncodeToString(h) }

func (m *memOAuthStore) SaveToken(_ context.Context, t Token) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[tokenKey(t.Kind, t.Hash)] = t
	return nil
}

func (m *memOAuthStore) GetToken(_ context.Context, h []byte, kind TokenKind, now time.Time) (Token, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tokens[tokenKey(kind, h)]
	if !ok || !t.ExpiresAt.After(now) {
		return Token{}, ErrNotFound
	}
	return t, nil
}

func (m *memOAuthStore) ConsumeRefreshToken(_ context.Context, h []byte, now time.Time) (Token, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := tokenKey(TokenRefresh, h)
	t, ok := m.tokens[k]
	if !ok || !t.ExpiresAt.After(now) {
		return Token{}, ErrNotFound
	}
	delete(m.tokens, k)
	return t, nil
}

func (m *memOAuthStore) DeleteExpiredOAuth(_ context.Context, now time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for k, c := range m.codes {
		if !c.ExpiresAt.After(now) {
			delete(m.codes, k)
			n++
		}
	}
	for k, t := range m.tokens {
		if !t.ExpiresAt.After(now) {
			delete(m.tokens, k)
			n++
		}
	}
	return n, nil
}

// ---- harness ----

type oh struct {
	*harness
	o       *OAuth
	store   *memOAuthStore
	session *http.Cookie
}

func newOH(t *testing.T) *oh { return newOHLog(t, quietLog()) }

func newOHLog(t *testing.T, log *slog.Logger) *oh {
	return newOHScopes(t, log, []Scope{{Name: "plutus:read", Description: "Read your spending"}, {Name: "plutus:write", Description: "Add and delete spending"}})
}

func newOHScopes(t *testing.T, log *slog.Logger, scopes []Scope) *oh {
	t.Helper()
	hs := newHarness(t)
	st := newMemOAuthStore()
	o := NewOAuth(hs.h, st, OAuthOptions{
		PublicURL: testOrigin,
		Secret:    testSecret,
		Scopes:    scopes,
		Now:       func() time.Time { return hs.now },
	}, log)
	o.Register(hs.mux)

	flow, state, _, _ := hs.startLogin("/")
	rec := hs.do("GET", "/api/auth/google/callback?code=good&state="+url.QueryEscape(state), []*http.Cookie{flow}, nil)
	sess := cookieNamed(rec, sessionCookie)
	if sess == nil {
		t.Fatal("test sign-in failed")
	}
	return &oh{harness: hs, o: o, store: st, session: sess}
}

func (h *oh) send(method, target, contentType, body string, cookies []*http.Cookie, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)
	return rec
}

func (h *oh) registerClient(name string, redirects ...string) string {
	h.t.Helper()
	body, _ := json.Marshal(map[string]any{"client_name": name, "redirect_uris": redirects, "token_endpoint_auth_method": "none"})
	rec := h.send("POST", "/api/oauth/register", "application/json", string(body), nil, nil)
	if rec.Code != http.StatusCreated {
		h.t.Fatalf("register status = %d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		ClientID string `json:"client_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || out.ClientID == "" {
		h.t.Fatalf("bad register response: %s", rec.Body.String())
	}
	return out.ClientID
}

func authzValues(clientID, redirect, state, scope string) url.Values {
	v := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirect},
		"code_challenge":        {challengeFor(testVerifier)},
		"code_challenge_method": {"S256"},
		"state":                 {state},
	}
	if scope != "" {
		v.Set("scope", scope)
	}
	return v
}

var csrfRe = regexp.MustCompile(`name="csrf" value="([^"]+)"`)

// consent loads the consent page as the signed-in user and returns the CSRF token and page.
func (h *oh) consent(vals url.Values) (csrf, page string) {
	h.t.Helper()
	rec := h.send("GET", "/api/oauth/authorize?"+vals.Encode(), "", "", []*http.Cookie{h.session}, nil)
	if rec.Code != http.StatusOK {
		h.t.Fatalf("consent status = %d body=%s", rec.Code, rec.Body.String())
	}
	m := csrfRe.FindStringSubmatch(rec.Body.String())
	if m == nil {
		h.t.Fatalf("no csrf field on consent page:\n%s", rec.Body.String())
	}
	return m[1], rec.Body.String()
}

func (h *oh) decide(vals url.Values, csrf, decision string) *httptest.ResponseRecorder {
	form := url.Values{}
	for k, v := range vals {
		form[k] = v
	}
	form.Set("csrf", csrf)
	form.Set("decision", decision)
	return h.send("POST", "/api/oauth/authorize", "application/x-www-form-urlencoded", form.Encode(),
		[]*http.Cookie{h.session}, map[string]string{"Origin": testOrigin})
}

// authorizeCode runs consent+approve and returns the authorization code.
func (h *oh) authorizeCode(clientID, redirect, scope string) string {
	h.t.Helper()
	vals := authzValues(clientID, redirect, "st-"+clientID, scope)
	csrf, _ := h.consent(vals)
	rec := h.decide(vals, csrf, "approve")
	if rec.Code != http.StatusSeeOther {
		h.t.Fatalf("approve status = %d body=%s", rec.Code, rec.Body.String())
	}
	loc, _ := url.Parse(rec.Header().Get("Location"))
	code := loc.Query().Get("code")
	if code == "" {
		h.t.Fatalf("no code in redirect: %s", loc)
	}
	return code
}

func (h *oh) token(form url.Values) *httptest.ResponseRecorder {
	return h.send("POST", "/api/oauth/token", "application/x-www-form-urlencoded", form.Encode(), nil, nil)
}

func codeForm(clientID, code, redirect, verifier string) url.Values {
	return url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"redirect_uri":  {redirect},
		"code_verifier": {verifier},
	}
}

type tokenResp struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	Error        string `json:"error"`
}

func decodeToken(t *testing.T, rec *httptest.ResponseRecorder) tokenResp {
	t.Helper()
	var tr tokenResp
	if err := json.Unmarshal(rec.Body.Bytes(), &tr); err != nil {
		t.Fatalf("bad token json: %s", rec.Body.String())
	}
	return tr
}

// ---- tests ----

func TestDiscoveryMetadata(t *testing.T) {
	h := newOH(t)
	for _, path := range []string{"/.well-known/oauth-protected-resource", "/.well-known/oauth-protected-resource/mcp"} {
		rec := h.send("GET", path, "", "", nil, nil)
		if rec.Code != 200 || rec.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Fatalf("%s: status=%d cors=%q", path, rec.Code, rec.Header().Get("Access-Control-Allow-Origin"))
		}
		var m map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &m)
		if m["resource"] != testOrigin+"/mcp" {
			t.Fatalf("%s: resource = %v", path, m["resource"])
		}
		if as, _ := m["authorization_servers"].([]any); len(as) != 1 || as[0] != testOrigin {
			t.Fatalf("%s: authorization_servers = %v", path, m["authorization_servers"])
		}
	}

	rec := h.send("GET", "/.well-known/oauth-authorization-server", "", "", nil, nil)
	var m map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &m)
	want := map[string]string{
		"issuer":                 testOrigin,
		"authorization_endpoint": testOrigin + "/api/oauth/authorize",
		"token_endpoint":         testOrigin + "/api/oauth/token",
		"registration_endpoint":  testOrigin + "/api/oauth/register",
	}
	for k, v := range want {
		if m[k] != v {
			t.Errorf("%s = %v, want %v", k, m[k], v)
		}
	}
	if cm, _ := m["code_challenge_methods_supported"].([]any); len(cm) != 1 || cm[0] != "S256" {
		t.Errorf("code_challenge_methods_supported = %v", m["code_challenge_methods_supported"])
	}

	pre := h.send("OPTIONS", "/api/oauth/token", "", "", nil, nil)
	if pre.Code != http.StatusNoContent || pre.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("preflight: status=%d", pre.Code)
	}
}

func TestRegisterRedirectValidation(t *testing.T) {
	h := newOH(t)
	good := []string{
		"https://claude.ai/api/mcp/auth_callback",
		"http://localhost:6274/oauth/callback",
		"http://127.0.0.1:9000/cb",
		"cursor://anysphere.cursor-retrieval/oauth/callback",
	}
	for _, u := range good {
		body, _ := json.Marshal(map[string]any{"redirect_uris": []string{u}})
		if rec := h.send("POST", "/api/oauth/register", "application/json", string(body), nil, nil); rec.Code != http.StatusCreated {
			t.Errorf("%q should be accepted, got %d %s", u, rec.Code, rec.Body.String())
		}
	}
	bad := []string{
		"http://evil.example/cb", "javascript:alert(1)", "https://user:pw@example.com/cb",
		"https://example.com/cb#frag", "ftp://example.com/cb", "", "example.com/cb",
		"data:text/html,hi", "file:///etc/passwd", "https://exa mple.com/cb",
	}
	for _, u := range bad {
		body, _ := json.Marshal(map[string]any{"redirect_uris": []string{u}})
		rec := h.send("POST", "/api/oauth/register", "application/json", string(body), nil, nil)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "invalid_redirect_uri") {
			t.Errorf("%q should be rejected, got %d %s", u, rec.Code, rec.Body.String())
		}
	}
}

func TestRegisterMetadataValidation(t *testing.T) {
	h := newOH(t)
	cases := map[string]string{
		"not json":     `nope`,
		"no redirects": `{"client_name":"x"}`,
		"too many":     `{"redirect_uris":["https://a.example/1","https://a.example/2","https://a.example/3","https://a.example/4","https://a.example/5","https://a.example/6","https://a.example/7","https://a.example/8","https://a.example/9","https://a.example/10","https://a.example/11"]}`,
		"key-based":    `{"redirect_uris":["https://a.example/cb"],"token_endpoint_auth_method":"private_key_jwt"}`,
		"bad grant":    `{"redirect_uris":["https://a.example/cb"],"grant_types":["password"]}`,
		"bad response": `{"redirect_uris":["https://a.example/cb"],"response_types":["token"]}`,
	}
	for name, body := range cases {
		if rec := h.send("POST", "/api/oauth/register", "application/json", body, nil, nil); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400 (%s)", name, rec.Code, rec.Body.String())
		}
	}

	// Names are cleaned: control and invisible characters removed, length capped.
	id := h.registerClient("  My\nApp\u200b\u202e "+strings.Repeat("x", 300), "https://a.example/cb")
	c, _ := h.store.GetClient(context.Background(), id)
	if strings.ContainsAny(c.Name, "\n\u200b\u202e") || len([]rune(c.Name)) > 100 || !strings.HasPrefix(c.Name, "My App") {
		t.Fatalf("name not cleaned: %q", c.Name)
	}
}

func TestOAuthFullFlow(t *testing.T) {
	h := newOH(t)
	cid := h.registerClient("Test App", testRedirect)
	vals := authzValues(cid, testRedirect, "st-1", "plutus:read")

	// 1. Not signed in: send the user through Google sign-in, returning to this exact URL.
	rec := h.send("GET", "/api/oauth/authorize?"+vals.Encode(), "", "", nil, nil)
	if rec.Code != http.StatusFound || !strings.HasPrefix(rec.Header().Get("Location"), "/api/auth/google/login?next=") {
		t.Fatalf("signed-out authorize: status=%d loc=%q", rec.Code, rec.Header().Get("Location"))
	}
	loc, _ := url.Parse(rec.Header().Get("Location"))
	if next := loc.Query().Get("next"); next != "/api/oauth/authorize?"+vals.Encode() {
		t.Fatalf("next = %q", next)
	}

	// 2. Signed in: consent page lists only the requested scope.
	csrf, page := h.consent(vals)
	for _, want := range []string{"Test App", "See your spending records and totals", "http://localhost:7777", "nolan@example.com"} {
		if !strings.Contains(strings.ToLower(page), strings.ToLower(want)) {
			t.Errorf("consent page missing %q", want)
		}
	}
	if strings.Contains(page, "Add and delete spending records") {
		t.Error("consent page lists a scope that was not requested")
	}

	// 3. Approve: redirected back with code, state and issuer.
	rec = h.decide(vals, csrf, "approve")
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("approve status = %d", rec.Code)
	}
	back, _ := url.Parse(rec.Header().Get("Location"))
	if back.Host != "localhost:7777" || back.Path != "/callback" ||
		back.Query().Get("state") != "st-1" || back.Query().Get("iss") != testOrigin || back.Query().Get("code") == "" {
		t.Fatalf("bad redirect: %s", back)
	}
	code := back.Query().Get("code")

	// 4. Exchange the code with the PKCE verifier.
	rec = h.token(codeForm(cid, code, testRedirect, testVerifier))
	if rec.Code != 200 || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("token status=%d cache=%q body=%s", rec.Code, rec.Header().Get("Cache-Control"), rec.Body.String())
	}
	tr := decodeToken(t, rec)
	if tr.TokenType != "Bearer" || tr.AccessToken == "" || tr.RefreshToken == "" || tr.Scope != "plutus:read" || tr.ExpiresIn != 3600 {
		t.Fatalf("unexpected token response: %+v", tr)
	}
	if _, ok := h.store.tokens[tokenKey(TokenAccess, []byte(tr.AccessToken))]; ok {
		t.Fatal("raw access token must not be a store key")
	}

	info, err := h.o.VerifyAccessToken(context.Background(), tr.AccessToken)
	if err != nil || info.UserID != "user-g-123" || info.ClientID != cid || len(info.Scopes) != 1 || info.Scopes[0] != "plutus:read" {
		t.Fatalf("VerifyAccessToken = %+v, %v", info, err)
	}
	if _, err := h.o.VerifyAccessToken(context.Background(), "garbage"); err != ErrNotFound {
		t.Fatalf("garbage token err = %v, want ErrNotFound", err)
	}

	// 5. Refresh rotates: the old refresh token dies, the new pair works.
	rec = h.token(url.Values{"grant_type": {"refresh_token"}, "client_id": {cid}, "refresh_token": {tr.RefreshToken}})
	if rec.Code != 200 {
		t.Fatalf("refresh status = %d body=%s", rec.Code, rec.Body.String())
	}
	tr2 := decodeToken(t, rec)
	if tr2.AccessToken == tr.AccessToken || tr2.RefreshToken == tr.RefreshToken {
		t.Fatal("refresh must issue new tokens")
	}
	rec = h.token(url.Values{"grant_type": {"refresh_token"}, "client_id": {cid}, "refresh_token": {tr.RefreshToken}})
	if rec.Code != 400 || decodeToken(t, rec).Error != "invalid_grant" {
		t.Fatalf("reused refresh token: status=%d body=%s", rec.Code, rec.Body.String())
	}

	// 6. Access tokens expire.
	h.now = h.now.Add(61 * time.Minute)
	if _, err := h.o.VerifyAccessToken(context.Background(), tr2.AccessToken); err != ErrNotFound {
		t.Fatalf("expired access token err = %v, want ErrNotFound", err)
	}
}

func TestTokenEndpointRejections(t *testing.T) {
	cases := []struct {
		name       string
		mutate     func(h *oh, cid, code string) url.Values
		wantStatus int
		wantErr    string
	}{
		{"wrong verifier", func(h *oh, cid, code string) url.Values {
			return codeForm(cid, code, testRedirect, strings.Repeat("b", 43))
		}, 400, "invalid_grant"},
		{"missing verifier", func(h *oh, cid, code string) url.Values { return codeForm(cid, code, testRedirect, "") }, 400, "invalid_grant"},
		{"wrong redirect_uri", func(h *oh, cid, code string) url.Values {
			return codeForm(cid, code, "http://localhost:7777/other", testVerifier)
		}, 400, "invalid_grant"},
		{"other client", func(h *oh, cid, code string) url.Values {
			other := h.registerClient("Other", "http://localhost:8888/cb")
			return codeForm(other, code, testRedirect, testVerifier)
		}, 400, "invalid_grant"},
		{"unknown client", func(h *oh, cid, code string) url.Values {
			return codeForm("nem_c_nope", code, testRedirect, testVerifier)
		}, 401, "invalid_client"},
		{"unknown code", func(h *oh, cid, code string) url.Values {
			return codeForm(cid, "not-a-code", testRedirect, testVerifier)
		}, 400, "invalid_grant"},
		{"expired code", func(h *oh, cid, code string) url.Values {
			h.now = h.now.Add(6 * time.Minute)
			return codeForm(cid, code, testRedirect, testVerifier)
		}, 400, "invalid_grant"},
		{"resource mismatch", func(h *oh, cid, code string) url.Values {
			f := codeForm(cid, code, testRedirect, testVerifier)
			f.Set("resource", "https://evil.example/mcp")
			return f
		}, 400, "invalid_target"},
		{"bad grant type", func(h *oh, cid, code string) url.Values {
			return url.Values{"grant_type": {"password"}, "client_id": {cid}}
		}, 400, "unsupported_grant_type"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newOH(t)
			cid := h.registerClient("App", testRedirect)
			code := h.authorizeCode(cid, testRedirect, "")
			rec := h.token(tc.mutate(h, cid, code))
			if rec.Code != tc.wantStatus || decodeToken(t, rec).Error != tc.wantErr {
				t.Fatalf("status=%d body=%s, want %d %s", rec.Code, rec.Body.String(), tc.wantStatus, tc.wantErr)
			}
			if decodeToken(t, rec).AccessToken != "" {
				t.Fatal("rejected request must not return a token")
			}
		})
	}
}

func TestAuthorizationCodeIsSingleUse(t *testing.T) {
	h := newOH(t)
	cid := h.registerClient("App", testRedirect)
	code := h.authorizeCode(cid, testRedirect, "")
	if rec := h.token(codeForm(cid, code, testRedirect, testVerifier)); rec.Code != 200 {
		t.Fatalf("first exchange: %d %s", rec.Code, rec.Body.String())
	}
	rec := h.token(codeForm(cid, code, testRedirect, testVerifier))
	if rec.Code != 400 || decodeToken(t, rec).Error != "invalid_grant" {
		t.Fatalf("second exchange: %d %s", rec.Code, rec.Body.String())
	}
}

func TestRefreshTokenBoundToClientAndScope(t *testing.T) {
	h := newOH(t)
	cid := h.registerClient("App", testRedirect)
	other := h.registerClient("Other", "http://localhost:8888/cb")
	code := h.authorizeCode(cid, testRedirect, "plutus:read")
	tr := decodeToken(t, h.token(codeForm(cid, code, testRedirect, testVerifier)))

	// Cannot widen scope.
	rec := h.token(url.Values{"grant_type": {"refresh_token"}, "client_id": {cid}, "refresh_token": {tr.RefreshToken}, "scope": {"plutus:read plutus:write"}})
	if rec.Code != 400 || decodeToken(t, rec).Error != "invalid_scope" {
		t.Fatalf("widen scope: %d %s", rec.Code, rec.Body.String())
	}

	// A different client cannot use it.
	code = h.authorizeCode(cid, testRedirect, "plutus:read")
	tr = decodeToken(t, h.token(codeForm(cid, code, testRedirect, testVerifier)))
	rec = h.token(url.Values{"grant_type": {"refresh_token"}, "client_id": {other}, "refresh_token": {tr.RefreshToken}})
	if rec.Code != 400 || decodeToken(t, rec).Error != "invalid_grant" {
		t.Fatalf("cross-client refresh: %d %s", rec.Code, rec.Body.String())
	}
}

func TestAuthorizeRequestValidation(t *testing.T) {
	cases := []struct {
		name       string
		mutate     func(v url.Values)
		wantPage   bool   // error shown on our page, never redirected
		wantErrKey string // otherwise: error code sent back to the client
	}{
		{"unregistered redirect_uri", func(v url.Values) { v.Set("redirect_uri", "http://localhost:7777/evil") }, true, ""},
		{"unknown client", func(v url.Values) { v.Set("client_id", "nem_c_missing") }, true, ""},
		{"missing client_id", func(v url.Values) { v.Del("client_id") }, true, ""},
		{"missing PKCE", func(v url.Values) { v.Del("code_challenge") }, false, "invalid_request"},
		{"plain PKCE", func(v url.Values) { v.Set("code_challenge_method", "plain") }, false, "invalid_request"},
		{"short challenge", func(v url.Values) { v.Set("code_challenge", "abc") }, false, "invalid_request"},
		{"implicit flow", func(v url.Values) { v.Set("response_type", "token") }, false, "unsupported_response_type"},
		{"unknown scope", func(v url.Values) { v.Set("scope", "plutus:read admin") }, false, "invalid_scope"},
		{"wrong resource", func(v url.Values) { v.Set("resource", "https://evil.example/mcp") }, false, "invalid_target"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newOH(t)
			cid := h.registerClient("App", testRedirect)
			vals := authzValues(cid, testRedirect, "st-9", "")
			tc.mutate(vals)
			rec := h.send("GET", "/api/oauth/authorize?"+vals.Encode(), "", "", []*http.Cookie{h.session}, nil)
			if tc.wantPage {
				if rec.Code != http.StatusBadRequest || rec.Header().Get("Location") != "" {
					t.Fatalf("want error page without redirect, got %d loc=%q", rec.Code, rec.Header().Get("Location"))
				}
				return
			}
			if rec.Code != http.StatusSeeOther {
				t.Fatalf("want redirect with error, got %d", rec.Code)
			}
			loc, _ := url.Parse(rec.Header().Get("Location"))
			if loc.Host != "localhost:7777" || loc.Query().Get("error") != tc.wantErrKey || loc.Query().Get("state") != "st-9" {
				t.Fatalf("redirect = %s", loc)
			}
		})
	}
}

func TestAuthorizeAcceptsMatchingResource(t *testing.T) {
	h := newOH(t)
	cid := h.registerClient("App", testRedirect)
	vals := authzValues(cid, testRedirect, "s", "")
	vals.Set("resource", testOrigin+"/mcp")
	h.consent(vals) // fatals unless 200
}

func TestConsentFormProtections(t *testing.T) {
	h := newOH(t)
	cid := h.registerClient("App", testRedirect)
	vals := authzValues(cid, testRedirect, "st", "plutus:read")
	csrf, _ := h.consent(vals)

	// Wrong or missing CSRF token.
	if rec := h.decide(vals, "wrong", "approve"); rec.Code != http.StatusForbidden {
		t.Errorf("bad csrf: %d", rec.Code)
	}
	if rec := h.decide(vals, "", "approve"); rec.Code != http.StatusForbidden {
		t.Errorf("missing csrf: %d", rec.Code)
	}
	// The token is bound to the exact scopes shown to the user.
	widened := authzValues(cid, testRedirect, "st", "plutus:read plutus:write")
	if rec := h.decide(widened, csrf, "approve"); rec.Code != http.StatusForbidden {
		t.Errorf("widened scope with old csrf: %d", rec.Code)
	}
	// Cross-origin POST.
	form := url.Values{"csrf": {csrf}, "decision": {"approve"}}
	for k, v := range vals {
		form[k] = v
	}
	rec := h.send("POST", "/api/oauth/authorize", "application/x-www-form-urlencoded", form.Encode(),
		[]*http.Cookie{h.session}, map[string]string{"Origin": "https://evil.example"})
	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-origin: %d", rec.Code)
	}
	// No session.
	rec = h.send("POST", "/api/oauth/authorize", "application/x-www-form-urlencoded", form.Encode(), nil,
		map[string]string{"Origin": testOrigin})
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("no session: %d", rec.Code)
	}
	// Unknown decision.
	if rec := h.decide(vals, csrf, "maybe"); rec.Code != http.StatusBadRequest {
		t.Errorf("unknown decision: %d", rec.Code)
	}
	if len(h.store.codes) != 0 {
		t.Fatal("no authorization code may be issued by a rejected request")
	}
}

func TestConsentDenied(t *testing.T) {
	h := newOH(t)
	cid := h.registerClient("App", testRedirect)
	vals := authzValues(cid, testRedirect, "st-deny", "")
	csrf, _ := h.consent(vals)
	rec := h.decide(vals, csrf, "deny")
	loc, _ := url.Parse(rec.Header().Get("Location"))
	if rec.Code != http.StatusSeeOther || loc.Query().Get("error") != "access_denied" || loc.Query().Get("state") != "st-deny" || loc.Query().Get("code") != "" {
		t.Fatalf("deny: status=%d loc=%s", rec.Code, loc)
	}
	if len(h.store.codes) != 0 {
		t.Fatal("deny must not create a code")
	}
}

func TestConsentPageEscapesClientName(t *testing.T) {
	h := newOH(t)
	cid := h.registerClient(`<script>alert(1)</script>`, testRedirect)
	_, page := h.consent(authzValues(cid, testRedirect, "s", ""))
	if strings.Contains(page, "<script>alert(1)") {
		t.Fatal("client name rendered unescaped")
	}
	if !strings.Contains(page, "&lt;script&gt;") {
		t.Fatal("expected escaped client name on the page")
	}
}

func TestCleanupExpired(t *testing.T) {
	h := newOH(t)
	cid := h.registerClient("App", testRedirect)
	code := h.authorizeCode(cid, testRedirect, "")
	h.token(codeForm(cid, code, testRedirect, testVerifier))
	h.now = h.now.Add(60 * 24 * time.Hour)
	n, err := h.o.CleanupExpired(context.Background())
	if err != nil || n != 2 { // access + refresh
		t.Fatalf("cleanup = %d, %v", n, err)
	}
}
