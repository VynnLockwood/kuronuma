package olympus

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const testAddr = "127.0.0.1:8082"

// stub is the handler behind the guard; it records whether it was reached.
type stub struct{ calls int }

func (s *stub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.calls++
	_, _ = io.WriteString(w, "ok")
}

func newTestGuard(t *testing.T, addr string, maxBody int64) (*guard, http.Handler, *stub) {
	t.Helper()
	g, err := newGuard(addr, maxBody)
	if err != nil {
		t.Fatal(err)
	}
	s := &stub{}
	return g, g.Wrap(s), s
}

type postOpts struct {
	host, origin, referer string
	token                 string
	omitToken             bool
	file                  string
}

func multipartBody(t *testing.T, o postOpts) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if !o.omitToken {
		if err := mw.WriteField("csrf", o.token); err != nil {
			t.Fatal(err)
		}
	}
	fw, err := mw.CreateFormFile("file", "backup.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(fw, o.file)
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf, mw.FormDataContentType()
}

func post(t *testing.T, h http.Handler, o postOpts) *httptest.ResponseRecorder {
	t.Helper()
	body, ct := multipartBody(t, o)
	req := httptest.NewRequest(http.MethodPost, "/import", body)
	req.Header.Set("Content-Type", ct)
	req.Host = o.host
	if o.host == "" {
		req.Host = testAddr
	}
	if o.origin != "" {
		req.Header.Set("Origin", o.origin)
	}
	if o.referer != "" {
		req.Header.Set("Referer", o.referer)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func get(h http.Handler, path, host string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Host = host
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHostAllowList(t *testing.T) {
	_, h, s := newTestGuard(t, testAddr, 1<<20)
	for _, host := range []string{"127.0.0.1:8082", "localhost:8082", "LOCALHOST:8082", "Localhost:8082"} {
		before := s.calls
		if rec := get(h, "/", host); rec.Code != http.StatusOK || s.calls != before+1 {
			t.Errorf("Host %q: status %d, want 200 and handler reached", host, rec.Code)
		}
	}
	for _, host := range []string{"evil.com", "evil.com:8082", "127.0.0.1:9999", "localhost:9999", "127.0.0.1", "", "localhost.evil.com:8082", "127.0.0.1.evil.com:8082", "0.0.0.0:8082"} {
		before := s.calls
		rec := get(h, "/", host)
		if rec.Code != http.StatusNotFound || s.calls != before {
			t.Errorf("Host %q: status %d, handler reached=%v, want 404 and not reached", host, rec.Code, s.calls != before)
		}
		if rec.Body.Len() != 0 {
			t.Errorf("Host %q: 404 must have an empty body, got %q", host, rec.Body.String())
		}
	}
}

func TestHostAllowListFollowsAddrPort(t *testing.T) {
	_, h, _ := newTestGuard(t, "0.0.0.0:9000", 1<<20)
	if rec := get(h, "/", "127.0.0.1:9000"); rec.Code != http.StatusOK {
		t.Errorf("configured port must be allowed, got %d", rec.Code)
	}
	if rec := get(h, "/", "127.0.0.1:8082"); rec.Code != http.StatusNotFound {
		t.Errorf("default port must not be allowed when another is configured, got %d", rec.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	g, h, _ := newTestGuard(t, testAddr, 1<<20)
	want := map[string]string{
		"Cache-Control":           "no-store",
		"Content-Security-Policy": "default-src 'none'; style-src 'self'; form-action 'self'",
		"X-Frame-Options":         "DENY",
		// same-origin, not no-referrer: with no-referrer Chrome sends "Origin: null" on the import form
		// post, which the guard must refuse, so the owner could never import. Found with a real browser.
		"Referrer-Policy":        "same-origin",
		"X-Content-Type-Options": "nosniff",
	}
	responses := map[string]*httptest.ResponseRecorder{
		"200":          get(h, "/", testAddr),
		"404 bad host": get(h, "/", "evil.com"),
		"405": func() *httptest.ResponseRecorder {
			r := httptest.NewRequest(http.MethodPut, "/", nil)
			r.Host = testAddr
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			return w
		}(),
		"403 no origin": post(t, h, postOpts{token: g.Token()}),
	}
	for name, rec := range responses {
		for k, v := range want {
			if got := rec.Header().Get(k); got != v {
				t.Errorf("%s: header %s = %q, want %q", name, k, got, v)
			}
		}
		if c := rec.Header().Values("Set-Cookie"); len(c) != 0 {
			t.Errorf("%s: must not set cookies, got %v", name, c)
		}
	}
}

func TestImportRequiresToken(t *testing.T) {
	g, h, s := newTestGuard(t, testAddr, 1<<20)
	origin := "http://127.0.0.1:8082"
	for name, o := range map[string]postOpts{
		"missing":   {origin: origin, omitToken: true},
		"empty":     {origin: origin, token: ""},
		"wrong":     {origin: origin, token: "not-the-token"},
		"truncated": {origin: origin, token: g.Token()[:10]},
	} {
		if rec := post(t, h, o); rec.Code != http.StatusForbidden {
			t.Errorf("%s token: status %d, want 403", name, rec.Code)
		}
	}
	if s.calls != 0 {
		t.Fatal("handler must not run without a valid token")
	}
	if rec := post(t, h, postOpts{origin: origin, token: g.Token()}); rec.Code != http.StatusOK || s.calls != 1 {
		t.Fatalf("valid token: status %d calls %d", rec.Code, s.calls)
	}
}

func TestImportRequiresSameOrigin(t *testing.T) {
	g, h, s := newTestGuard(t, testAddr, 1<<20)
	tok := g.Token()
	forbidden := map[string]postOpts{
		"evil origin":          {origin: "http://evil.com", token: tok},
		"null origin":          {origin: "null", token: tok},
		"other port":           {origin: "http://127.0.0.1:9999", token: tok},
		"https scheme":         {origin: "https://127.0.0.1:8082", token: tok},
		"evil referer":         {referer: "http://evil.com/page", token: tok},
		"referer lookalike":    {referer: "http://127.0.0.1:8082.evil.com/", token: tok},
		"neither header":       {token: tok},
		"origin wins over ref": {origin: "http://evil.com", referer: "http://127.0.0.1:8082/", token: tok},
	}
	for name, o := range forbidden {
		if rec := post(t, h, o); rec.Code != http.StatusForbidden {
			t.Errorf("%s: status %d, want 403", name, rec.Code)
		}
	}
	if s.calls != 0 {
		t.Fatal("handler must not run for a cross-origin or origin-less post")
	}
	allowed := map[string]postOpts{
		"127 origin":       {origin: "http://127.0.0.1:8082", token: tok},
		"localhost origin": {origin: "http://localhost:8082", token: tok},
		"referer only":     {referer: "http://127.0.0.1:8082/", token: tok},
	}
	for name, o := range allowed {
		if rec := post(t, h, o); rec.Code != http.StatusOK {
			t.Errorf("%s: status %d, want 200", name, rec.Code)
		}
	}
}

func TestMethodsLimited(t *testing.T) {
	_, h, s := newTestGuard(t, testAddr, 1<<20)
	for _, tc := range []struct{ method, path string }{{http.MethodPut, "/"}, {http.MethodDelete, "/import"}, {http.MethodPatch, "/"}} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req.Host = testAddr
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") == "" {
			t.Errorf("%s %s: status %d Allow=%q, want 405 with Allow", tc.method, tc.path, rec.Code, rec.Header().Get("Allow"))
		}
	}
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	req.Host = "evil.com"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("bad host must win over method check, got %d", rec.Code)
	}
	if s.calls != 0 {
		t.Fatal("handler must not run for disallowed methods")
	}
}

func TestTokenDiffersPerGuard(t *testing.T) {
	a, _, _ := newTestGuard(t, testAddr, 1<<20)
	b, _, _ := newTestGuard(t, testAddr, 1<<20)
	if a.Token() == b.Token() {
		t.Fatal("two guards share a token")
	}
	if len(a.Token()) < 43 {
		t.Fatalf("token too short (%d chars) for 32 random bytes", len(a.Token()))
	}
}

func TestImportTooLarge(t *testing.T) {
	g, h, s := newTestGuard(t, testAddr, 1024)
	rec := post(t, h, postOpts{origin: "http://127.0.0.1:8082", token: g.Token(), file: strings.Repeat("x", 5000)})
	if rec.Code != http.StatusRequestEntityTooLarge || s.calls != 0 {
		t.Fatalf("status %d calls %d, want 413 and handler not reached", rec.Code, s.calls)
	}
}

func TestImportRejectsNonMultipart(t *testing.T) {
	g, h, s := newTestGuard(t, testAddr, 1<<20)
	req := httptest.NewRequest(http.MethodPost, "/import", strings.NewReader(url.Values{"csrf": {g.Token()}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "http://127.0.0.1:8082")
	req.Host = testAddr
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || s.calls != 0 {
		t.Fatalf("status %d calls %d, want 400 and handler not reached", rec.Code, s.calls)
	}
}

func TestHandlerCanReadUploadAfterGuard(t *testing.T) {
	g, err := newGuard(testAddr, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	var got string
	h := g.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, _, err := r.FormFile("file")
		if err != nil {
			t.Errorf("FormFile: %v", err)
			return
		}
		defer f.Close()
		b, _ := io.ReadAll(f)
		got = string(b)
	}))
	post(t, h, postOpts{origin: "http://127.0.0.1:8082", token: g.Token(), file: "hello backup"})
	if got != "hello backup" {
		t.Fatalf("handler read %q, want the uploaded file", got)
	}
}

func TestNewGuardRejectsBadInput(t *testing.T) {
	for _, addr := range []string{"garbage", "", "127.0.0.1", "127.0.0.1:"} {
		if _, err := newGuard(addr, 1024); err == nil {
			t.Errorf("addr %q: want an error", addr)
		}
	}
	if _, err := newGuard(testAddr, 0); err == nil {
		t.Error("maxBody 0: want an error")
	}
}
