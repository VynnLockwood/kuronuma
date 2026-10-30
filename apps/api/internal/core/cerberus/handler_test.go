package cerberus

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

const testOrigin = "http://localhost:8080"

var testSecret = []byte("0123456789abcdef0123456789abcdef-test")

// ---- fakes ----

type fakeProvider struct {
	identity Identity
	err      error

	gotCode, gotVerifier, gotNonce string
}

func (f *fakeProvider) AuthCodeURL(state, nonce, verifier string) string {
	v := url.Values{"state": {state}, "nonce": {nonce}, "verifier": {verifier}}
	return "https://accounts.example/auth?" + v.Encode()
}

func (f *fakeProvider) Exchange(_ context.Context, code, verifier, nonce string) (Identity, error) {
	f.gotCode, f.gotVerifier, f.gotNonce = code, verifier, nonce
	if f.err != nil {
		return Identity{}, f.err
	}
	return f.identity, nil
}

type memStore struct {
	mu       sync.Mutex
	users    map[string]User // by subject
	sessions map[string]memSession
}

type memSession struct {
	user    User
	expires time.Time
}

func newMemStore() *memStore {
	return &memStore{users: map[string]User{}, sessions: map[string]memSession{}}
}

func (m *memStore) UpsertUser(_ context.Context, id Identity) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id.Subject]
	if !ok {
		u = User{ID: "user-" + id.Subject}
	}
	u.Email, u.Name, u.Picture = id.Email, id.Name, id.Picture
	m.users[id.Subject] = u
	return u, nil
}

func (m *memStore) CreateSession(_ context.Context, userID string, h []byte, exp time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.ID == userID {
			m.sessions[hex.EncodeToString(h)] = memSession{user: u, expires: exp}
			return nil
		}
	}
	return errors.New("no such user")
}

func (m *memStore) GetSessionUser(_ context.Context, h []byte, now time.Time) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[hex.EncodeToString(h)]
	if !ok || !s.expires.After(now) {
		return User{}, ErrNotFound
	}
	return s.user, nil
}

func (m *memStore) DeleteSession(_ context.Context, h []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, hex.EncodeToString(h))
	return nil
}

func (m *memStore) DeleteExpiredSessions(_ context.Context, now time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for k, s := range m.sessions {
		if !s.expires.After(now) {
			delete(m.sessions, k)
			n++
		}
	}
	return n, nil
}

// ---- harness ----

type harness struct {
	t     *testing.T
	h     *Handler
	mux   *http.ServeMux
	prov  *fakeProvider
	store *memStore
	now   time.Time
}

func newHarness(t *testing.T, allowed ...string) *harness {
	t.Helper()
	hs := &harness{
		t:     t,
		prov:  &fakeProvider{identity: Identity{Subject: "g-123", Email: "Nolan@Example.com", EmailVerified: true, Name: "Nolan"}},
		store: newMemStore(),
		now:   time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC),
	}
	hs.h = New(hs.prov, hs.store, Options{
		PublicURL:     testOrigin,
		Secret:        testSecret,
		SessionTTL:    24 * time.Hour,
		AllowedEmails: allowed,
		Now:           func() time.Time { return hs.now },
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	hs.mux = http.NewServeMux()
	hs.h.Register(hs.mux)
	return hs
}

func (hs *harness) do(method, target string, cookies []*http.Cookie, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	hs.mux.ServeHTTP(rec, req)
	return rec
}

func cookieNamed(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// startLogin performs the login step and returns the flow cookie and the
// state/nonce/verifier that were sent to the provider.
func (hs *harness) startLogin(next string) (flow *http.Cookie, state, nonce, verifier string) {
	hs.t.Helper()
	rec := hs.do("GET", "/api/auth/google/login?next="+url.QueryEscape(next), nil, nil)
	if rec.Code != http.StatusFound {
		hs.t.Fatalf("login status = %d, want 302", rec.Code)
	}
	loc, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		hs.t.Fatal(err)
	}
	q := loc.Query()
	flow = cookieNamed(rec, flowCookie)
	if flow == nil {
		hs.t.Fatal("login did not set the flow cookie")
	}
	return flow, q.Get("state"), q.Get("nonce"), q.Get("verifier")
}

// ---- tests ----

func TestLoginSetsHardenedFlowCookieAndRedirects(t *testing.T) {
	hs := newHarness(t)
	flow, state, nonce, verifier := hs.startLogin("/dashboard")

	if state == "" || nonce == "" || len(verifier) < 43 {
		t.Fatalf("weak/missing values: state=%q nonce=%q verifier len=%d", state, nonce, len(verifier))
	}
	if state == nonce || nonce == verifier {
		t.Fatal("state, nonce and verifier must be independent random values")
	}
	if !flow.HttpOnly || flow.SameSite != http.SameSiteLaxMode || flow.Path != "/api/auth/google" {
		t.Fatalf("flow cookie not hardened: %+v", flow)
	}
}

func TestLoginWithoutProviderIs503(t *testing.T) {
	hs := newHarness(t)
	h := New(nil, hs.store, Options{PublicURL: testOrigin, Secret: testSecret, SessionTTL: time.Hour},
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	h.Register(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/auth/google/login", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestHappyPathSignInMeAndLogout(t *testing.T) {
	hs := newHarness(t)
	flow, state, nonce, verifier := hs.startLogin("/dashboard")

	rec := hs.do("GET", "/api/auth/google/callback?code=good&state="+url.QueryEscape(state), []*http.Cookie{flow}, nil)
	if rec.Code != http.StatusFound {
		t.Fatalf("callback status = %d, want 302", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != testOrigin+"/dashboard" {
		t.Fatalf("redirect = %q, want %q", got, testOrigin+"/dashboard")
	}
	// The verifier and nonce from login must reach the provider unchanged.
	if hs.prov.gotVerifier != verifier || hs.prov.gotNonce != nonce || hs.prov.gotCode != "good" {
		t.Fatalf("provider got code=%q verifier=%q nonce=%q", hs.prov.gotCode, hs.prov.gotVerifier, hs.prov.gotNonce)
	}

	sess := cookieNamed(rec, sessionCookie)
	if sess == nil || sess.Value == "" {
		t.Fatal("no session cookie set")
	}
	if !sess.HttpOnly || sess.SameSite != http.SameSiteLaxMode || sess.Path != "/" {
		t.Fatalf("session cookie not hardened: %+v", sess)
	}
	if c := cookieNamed(rec, flowCookie); c == nil || c.MaxAge >= 0 {
		t.Fatal("flow cookie should be cleared after callback")
	}
	// The raw token must not be what the database holds.
	if _, ok := hs.store.sessions[sess.Value]; ok {
		t.Fatal("store holds the raw session token")
	}
	if _, ok := hs.store.sessions[hex.EncodeToString(hashToken(sess.Value))]; !ok {
		t.Fatal("store is missing the hashed session token")
	}

	me := hs.do("GET", "/api/auth/me", []*http.Cookie{sess}, nil)
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d, want 200", me.Code)
	}
	if body := me.Body.String(); !strings.Contains(body, `"email":"Nolan@Example.com"`) {
		t.Fatalf("me body = %s", body)
	}

	out := hs.do("POST", "/api/auth/logout", []*http.Cookie{sess}, map[string]string{"Origin": testOrigin})
	if out.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want 204", out.Code)
	}
	if after := hs.do("GET", "/api/auth/me", []*http.Cookie{sess}, nil); after.Code != http.StatusUnauthorized {
		t.Fatalf("me after logout = %d, want 401", after.Code)
	}
}

func TestMeWithoutOrWithBogusCookieIs401(t *testing.T) {
	hs := newHarness(t)
	if rec := hs.do("GET", "/api/auth/me", nil, nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no cookie: %d", rec.Code)
	}
	bogus := &http.Cookie{Name: sessionCookie, Value: "not-a-real-token"}
	if rec := hs.do("GET", "/api/auth/me", []*http.Cookie{bogus}, nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("bogus cookie: %d", rec.Code)
	}
}

func TestExpiredSessionIsRejected(t *testing.T) {
	hs := newHarness(t)
	flow, state, _, _ := hs.startLogin("/")
	rec := hs.do("GET", "/api/auth/google/callback?code=good&state="+url.QueryEscape(state), []*http.Cookie{flow}, nil)
	sess := cookieNamed(rec, sessionCookie)

	hs.now = hs.now.Add(25 * time.Hour) // past the 24h TTL
	if got := hs.do("GET", "/api/auth/me", []*http.Cookie{sess}, nil); got.Code != http.StatusUnauthorized {
		t.Fatalf("expired session status = %d, want 401", got.Code)
	}
}

func TestCallbackRejections(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(hs *harness, flow *http.Cookie, state string) (*http.Cookie, string, string) // returns cookie, state, extra query
		wantErr string
	}{
		{"state mismatch", func(hs *harness, f *http.Cookie, s string) (*http.Cookie, string, string) { return f, "wrong", "" }, "invalid_flow"},
		{"missing flow cookie", func(hs *harness, f *http.Cookie, s string) (*http.Cookie, string, string) { return nil, s, "" }, "invalid_flow"},
		{"tampered flow cookie", func(hs *harness, f *http.Cookie, s string) (*http.Cookie, string, string) {
			bad := *f
			bad.Value = "x" + f.Value
			return &bad, s, ""
		}, "invalid_flow"},
		{"expired flow", func(hs *harness, f *http.Cookie, s string) (*http.Cookie, string, string) {
			hs.now = hs.now.Add(11 * time.Minute)
			return f, s, ""
		}, "invalid_flow"},
		{"user denied at google", func(hs *harness, f *http.Cookie, s string) (*http.Cookie, string, string) {
			return f, s, "&error=access_denied"
		}, "access_denied"},
		{"provider error", func(hs *harness, f *http.Cookie, s string) (*http.Cookie, string, string) {
			hs.prov.err = errors.New("boom")
			return f, s, ""
		}, "exchange_failed"},
		{"unverified email", func(hs *harness, f *http.Cookie, s string) (*http.Cookie, string, string) {
			hs.prov.identity.EmailVerified = false
			return f, s, ""
		}, "email_unverified"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hs := newHarness(t)
			flow, state, _, _ := hs.startLogin("/")
			c, st, extra := tc.mutate(hs, flow, state)
			var cookies []*http.Cookie
			if c != nil {
				cookies = append(cookies, c)
			}
			rec := hs.do("GET", "/api/auth/google/callback?code=good&state="+url.QueryEscape(st)+extra, cookies, nil)
			if rec.Code != http.StatusFound {
				t.Fatalf("status = %d, want 302", rec.Code)
			}
			if want := testOrigin + "/?auth_error=" + tc.wantErr; rec.Header().Get("Location") != want {
				t.Fatalf("redirect = %q, want %q", rec.Header().Get("Location"), want)
			}
			if cookieNamed(rec, sessionCookie) != nil {
				t.Fatal("a session must not be created on failure")
			}
			if len(hs.store.sessions) != 0 {
				t.Fatal("store must have no sessions after failure")
			}
		})
	}
}

func TestAllowList(t *testing.T) {
	hs := newHarness(t, "someone-else@example.com")
	flow, state, _, _ := hs.startLogin("/")
	rec := hs.do("GET", "/api/auth/google/callback?code=good&state="+url.QueryEscape(state), []*http.Cookie{flow}, nil)
	if want := testOrigin + "/?auth_error=not_allowed"; rec.Header().Get("Location") != want {
		t.Fatalf("redirect = %q, want %q", rec.Header().Get("Location"), want)
	}
	if len(hs.store.users) != 0 {
		t.Fatal("a rejected user must not be stored")
	}

	// Case-insensitive match for an allowed address.
	hs2 := newHarness(t, "nolan@example.com")
	flow, state, _, _ = hs2.startLogin("/")
	rec = hs2.do("GET", "/api/auth/google/callback?code=good&state="+url.QueryEscape(state), []*http.Cookie{flow}, nil)
	if cookieNamed(rec, sessionCookie) == nil {
		t.Fatal("allow-listed user (different case) should sign in")
	}
}

func TestFlowCookieIsSingleUseAtTheBrowserButSignedAgainstForgery(t *testing.T) {
	// A cookie sealed with a different secret must be refused.
	forged, err := sealFlow([]byte("a-completely-different-secret-value!!"), flowState{
		State: "s", Nonce: "n", Verifier: "v", Next: "/", Expires: time.Now().Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	hs := newHarness(t)
	rec := hs.do("GET", "/api/auth/google/callback?code=good&state=s",
		[]*http.Cookie{{Name: flowCookie, Value: forged}}, nil)
	if want := testOrigin + "/?auth_error=invalid_flow"; rec.Header().Get("Location") != want {
		t.Fatalf("redirect = %q, want %q", rec.Header().Get("Location"), want)
	}
}

func TestLogoutRefusesCrossOrigin(t *testing.T) {
	hs := newHarness(t)
	rec := hs.do("POST", "/api/auth/logout", nil, map[string]string{"Origin": "https://evil.example"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestLogoutRejectsGET(t *testing.T) {
	hs := newHarness(t)
	rec := hs.do("GET", "/api/auth/logout", nil, nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestNextIsSanitisedDuringLogin(t *testing.T) {
	hs := newHarness(t)
	flow, state, _, _ := hs.startLogin("//evil.example/steal")
	rec := hs.do("GET", "/api/auth/google/callback?code=good&state="+url.QueryEscape(state), []*http.Cookie{flow}, nil)
	if got := rec.Header().Get("Location"); got != testOrigin+"/" {
		t.Fatalf("redirect = %q, want site root", got)
	}
}
