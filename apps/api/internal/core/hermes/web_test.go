package hermes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

const webSite = "https://kuronuma.example"

// fakeSession maps the cookie "s" to a user, standing in for the real session.
type fakeSession struct{}

func (fakeSession) UserID(r *http.Request) (string, bool) {
	c, err := r.Cookie("s")
	if err != nil {
		return "", false
	}
	switch c.Value {
	case "alice":
		return "u-alice", true
	case "bob":
		return "u-bob", true
	}
	return "", false
}

type webEnv struct {
	h   http.Handler
	log *bytes.Buffer
}

func newWebEnv() *webEnv {
	buf := &bytes.Buffer{}
	s := New(Options{Name: "t"}, fakeAuth{}, slog.New(slog.NewTextHandler(buf, nil)))
	var mu sync.Mutex
	data := map[string]string{} // per-user, keyed by the caller's user id
	add := func(t Tool) { t.InputSchema = map[string]any{"type": "object"}; s.AddTool(t) }
	add(Tool{Name: "whoami", Scope: "t:read", Handler: func(_ context.Context, c Caller, _ json.RawMessage) (string, error) {
		return Respond(map[string]any{"user": c.UserID, "client": c.ClientID, "scopes": c.Scopes})
	}})
	add(Tool{Name: "put", Scope: "t:write", Handler: func(_ context.Context, c Caller, raw json.RawMessage) (string, error) {
		var a struct {
			Value string `json:"value"`
		}
		if err := DecodeArgs(raw, &a); err != nil {
			return "", err
		}
		mu.Lock()
		data[c.UserID] = a.Value
		mu.Unlock()
		return Respond(map[string]any{"status": "saved"})
	}})
	add(Tool{Name: "get", Scope: "t:read", Handler: func(_ context.Context, c Caller, raw json.RawMessage) (string, error) {
		var a struct{}
		if err := DecodeArgs(raw, &a); err != nil {
			return "", err
		}
		mu.Lock()
		defer mu.Unlock()
		return Respond(map[string]any{"value": data[c.UserID]})
	}})
	add(Tool{Name: "refuse", Scope: "t:read", Handler: func(context.Context, Caller, json.RawMessage) (string, error) {
		return "", Errorf("amount must be positive")
	}})
	add(Tool{Name: "broken", Scope: "t:read", Handler: func(context.Context, Caller, json.RawMessage) (string, error) {
		return "", errors.New("pq: password authentication failed for user secret_db_user")
	}})
	add(Tool{Name: "plain", Scope: "t:read", Handler: func(context.Context, Caller, json.RawMessage) (string, error) {
		return "not a json object", nil
	}})
	add(Tool{Name: "secretive", Scope: "t:read", Handler: func(context.Context, Caller, json.RawMessage) (string, error) {
		return Respond(map[string]any{"leak": "RESULT-SECRET-9"})
	}})
	return &webEnv{h: WebHandler(s, fakeSession{}, webSite), log: buf}
}

// do sends a POST that passes the browser checks unless mutate changes it.
func (e *webEnv) do(cookie, body string, mutate func(r *http.Request)) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/api/web/call", strings.NewReader(body))
	r.Header.Set("X-Kuronuma-Web", "1")
	r.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		r.AddCookie(&http.Cookie{Name: "s", Value: cookie})
	}
	if mutate != nil {
		mutate(r)
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, r)
	return rec
}

func webCall(tool string, args any) string {
	b, _ := json.Marshal(map[string]any{"tool": tool, "arguments": args})
	return string(b)
}

func webDecode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("not JSON: %q", rec.Body.String())
	}
	return out
}

func TestWebMethods(t *testing.T) {
	e := newWebEnv()
	for _, m := range []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodPatch} {
		r := httptest.NewRequest(m, "/api/web/call", nil)
		rec := httptest.NewRecorder()
		e.h.ServeHTTP(rec, r)
		if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != "POST" {
			t.Errorf("%s: %d allow=%q", m, rec.Code, rec.Header().Get("Allow"))
		}
	}
}

func TestWebRefusesBrowserCheckFailures(t *testing.T) {
	e := newWebEnv()
	cases := map[string]func(r *http.Request){
		"no custom header": func(r *http.Request) { r.Header.Del("X-Kuronuma-Web") },
		"wrong header":     func(r *http.Request) { r.Header.Set("X-Kuronuma-Web", "yes") },
		"text/plain":       func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") },
		"form":             func(r *http.Request) { r.Header.Set("Content-Type", "application/x-www-form-urlencoded") },
		"foreign Origin":   func(r *http.Request) { r.Header.Set("Origin", "https://evil.example") },
		"cross-site fetch": func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "cross-site") },
		"same-site fetch":  func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "same-site") },
	}
	for name, mutate := range cases {
		rec := e.do("alice", webCall("whoami", nil), mutate)
		if rec.Code != http.StatusForbidden || webDecode(t, rec)["error"] != "forbidden" {
			t.Errorf("%s: %d %s", name, rec.Code, rec.Body.String())
		}
	}
	// The checks come before the session: a cross-site request without a cookie is "forbidden", not "not signed in".
	if rec := e.do("", webCall("whoami", nil), func(r *http.Request) { r.Header.Del("X-Kuronuma-Web") }); rec.Code != http.StatusForbidden {
		t.Errorf("no cookie and no header: %d", rec.Code)
	}
}

func TestWebNeedsASession(t *testing.T) {
	e := newWebEnv()
	for _, cookie := range []string{"", "mallory"} {
		rec := e.do(cookie, webCall("whoami", nil), nil)
		if rec.Code != http.StatusUnauthorized || webDecode(t, rec)["error"] != "not signed in" {
			t.Errorf("cookie %q: %d %s", cookie, rec.Code, rec.Body.String())
		}
	}
	// A signed-out caller never gets a body read: a huge body is still just 401.
	huge := webCall("whoami", map[string]string{"x": strings.Repeat("a", 200<<10)})
	if rec := e.do("", huge, nil); rec.Code != http.StatusUnauthorized {
		t.Errorf("huge body without a session: %d", rec.Code)
	}
}

func TestWebBodyAndToolErrors(t *testing.T) {
	e := newWebEnv()
	if rec := e.do("alice", webCall("whoami", map[string]string{"x": strings.Repeat("a", 70<<10)}), nil); rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("oversized: %d", rec.Code)
	}
	for name, body := range map[string]string{"not JSON": `{nope`, "no tool": `{}`, "empty tool": `{"tool":""}`, "array": `[1]`} {
		if rec := e.do("alice", body, nil); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %s", name, rec.Code, rec.Body.String())
		}
	}
	rec := e.do("alice", webCall("nope", nil), nil)
	if rec.Code != http.StatusNotFound || webDecode(t, rec)["error"] != "unknown tool" {
		t.Errorf("unknown tool: %d %s", rec.Code, rec.Body.String())
	}
}

func TestWebResults(t *testing.T) {
	e := newWebEnv()
	rec := e.do("alice", webCall("whoami", nil), nil)
	if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("whoami: %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	res := webDecode(t, rec)["result"].(map[string]any)
	if res["user"] != "u-alice" || res["client"] != "web" {
		t.Fatalf("result = %v", res)
	}
	scopes := res["scopes"].([]any)
	if len(scopes) != 2 || scopes[0] != "t:read" || scopes[1] != "t:write" {
		t.Fatalf("a web session carries every scope of the registered tools: %v", scopes)
	}
	// No arguments at all is the same as an empty object.
	if rec := e.do("alice", `{"tool":"whoami"}`, nil); rec.Code != http.StatusOK {
		t.Errorf("no arguments: %d", rec.Code)
	}

	// A ToolError is a 422 with its text, for people to read.
	rec = e.do("alice", webCall("refuse", nil), nil)
	if rec.Code != http.StatusUnprocessableEntity || webDecode(t, rec)["error"] != "amount must be positive" {
		t.Errorf("refuse: %d %s", rec.Code, rec.Body.String())
	}
	// Any other failure is a 500 with a generic sentence; the cause is only in the log.
	rec = e.do("alice", webCall("broken", nil), nil)
	if rec.Code != http.StatusInternalServerError || webDecode(t, rec)["error"] != "Something went wrong on the server. Please try again." {
		t.Errorf("broken: %d %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret_db_user") || !strings.Contains(e.log.String(), "secret_db_user") {
		t.Errorf("the cause must be logged and not returned:\nbody=%s\nlog=%s", rec.Body.String(), e.log.String())
	}
	// A tool whose text is not a JSON object cannot be a result: generic 500.
	if rec := e.do("alice", webCall("plain", nil), nil); rec.Code != http.StatusInternalServerError {
		t.Errorf("plain: %d", rec.Code)
	}
	// Every response is no-store and none carries CORS headers.
	for _, rec := range []*httptest.ResponseRecorder{
		e.do("alice", webCall("whoami", nil), nil), e.do("", webCall("whoami", nil), nil),
		e.do("alice", webCall("refuse", nil), nil), e.do("alice", `{`, nil),
	} {
		if rec.Header().Get("Cache-Control") != "no-store" {
			t.Errorf("Cache-Control = %q", rec.Header().Get("Cache-Control"))
		}
		for k := range rec.Header() {
			if strings.HasPrefix(k, "Access-Control-") {
				t.Errorf("unexpected CORS header %s", k)
			}
		}
	}
}

// The user comes only from the session. Nothing in the body or its arguments can choose it.
func TestWebBodyCannotChooseTheUser(t *testing.T) {
	e := newWebEnv()
	rec := e.do("alice", `{"tool":"whoami","user_id":"u-bob","UserID":"u-bob","arguments":{}}`, nil)
	if got := webDecode(t, rec)["result"].(map[string]any)["user"]; got != "u-alice" {
		t.Fatalf("user = %v, want u-alice", got)
	}
	// A tool that decodes strictly refuses a user id smuggled into its arguments.
	rec = e.do("alice", webCall("get", map[string]string{"user_id": "u-bob"}), nil)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("smuggled argument: %d %s", rec.Code, rec.Body.String())
	}
}

func TestWebUsersAreIsolatedBothWays(t *testing.T) {
	e := newWebEnv()
	if rec := e.do("alice", webCall("put", map[string]string{"value": "alice's secret"}), nil); rec.Code != 200 {
		t.Fatalf("alice put: %d", rec.Code)
	}
	get := func(cookie string) any {
		rec := e.do(cookie, webCall("get", nil), nil)
		if rec.Code != 200 {
			t.Fatalf("%s get: %d %s", cookie, rec.Code, rec.Body.String())
		}
		return webDecode(t, rec)["result"].(map[string]any)["value"]
	}
	if get("alice") != "alice's secret" || get("bob") != "" {
		t.Fatalf("alice=%v bob=%v", get("alice"), get("bob"))
	}
	if rec := e.do("bob", webCall("put", map[string]string{"value": "bob's secret"}), nil); rec.Code != 200 {
		t.Fatalf("bob put: %d", rec.Code)
	}
	if get("alice") != "alice's secret" || get("bob") != "bob's secret" {
		t.Fatalf("after bob wrote: alice=%v bob=%v", get("alice"), get("bob"))
	}
}

func TestWebLogsWhoAndWhatButNotArgumentsOrResults(t *testing.T) {
	e := newWebEnv()
	if rec := e.do("alice", webCall("secretive", map[string]string{"note": "ARG-SECRET-7"}), nil); rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	log := e.log.String()
	for _, want := range []string{"tool=secretive", "user_id=u-alice", "client_id=web", "ok=true"} {
		if !strings.Contains(log, want) {
			t.Errorf("log lacks %q:\n%s", want, log)
		}
	}
	// Arguments, results and the session cookie never reach the log.
	for _, secret := range []string{"ARG-SECRET-7", "RESULT-SECRET-9", "s=alice"} {
		if strings.Contains(log, secret) {
			t.Errorf("log must not contain %q:\n%s", secret, log)
		}
	}
}
