package cerberus

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/core/hermes"
)

// testMCPAuth mirrors the adapter in cmd/kuronuma-api: Hermes verifies tokens that Cerberus issued.
type testMCPAuth struct{ o *OAuth }

func (a testMCPAuth) Authenticate(ctx context.Context, bearer string) (hermes.Caller, error) {
	info, err := a.o.VerifyAccessToken(ctx, bearer)
	if errors.Is(err, ErrNotFound) {
		return hermes.Caller{}, hermes.ErrUnauthorized
	}
	if err != nil {
		return hermes.Caller{}, err
	}
	return hermes.Caller{UserID: info.UserID, ClientID: info.ClientID, Scopes: info.Scopes}, nil
}

// TestAssistantConnectsEndToEnd walks the same path a real MCP client takes:
// 401 -> discovery -> register -> sign-in + consent -> token -> tool call -> refresh.
func TestAssistantConnectsEndToEnd(t *testing.T) {
	h := newOH(t)
	mcp := hermes.New(hermes.Options{Name: "kuronuma", ResourceMetadataURL: h.o.MetadataURL()}, testMCPAuth{h.o}, quietLog())
	mcp.AddTool(hermes.Tool{
		Name: "whoami", Scope: "plutus:read", ReadOnly: true, InputSchema: map[string]any{"type": "object"},
		Handler: func(_ context.Context, c hermes.Caller, _ json.RawMessage) (string, error) {
			return "user=" + c.UserID, nil
		},
	})
	mcp.AddTool(hermes.Tool{
		Name: "write_something", Scope: "plutus:write", InputSchema: map[string]any{"type": "object"},
		Handler: func(context.Context, hermes.Caller, json.RawMessage) (string, error) { return "written", nil },
	})
	h.mux.Handle("/mcp", mcp)

	rpc := func(token, body string) *http.Response {
		hdr := map[string]string{"Authorization": "Bearer " + token}
		if token == "" {
			hdr = nil
		}
		return h.send("POST", "/mcp", "application/json", body, nil, hdr).Result()
	}
	const initBody = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}`

	// 1. No token: 401 pointing at the protected-resource metadata.
	resp := rpc("", initBody)
	challenge := resp.Header.Get("WWW-Authenticate")
	if resp.StatusCode != http.StatusUnauthorized || !strings.HasPrefix(challenge, `Bearer resource_metadata="`) {
		t.Fatalf("expected 401 with resource_metadata, got %d %q", resp.StatusCode, challenge)
	}
	metaURL := strings.TrimSuffix(strings.TrimPrefix(challenge, `Bearer resource_metadata="`), `"`)
	if metaURL != h.o.MetadataURL() {
		t.Fatalf("challenge points at %q, want %q", metaURL, h.o.MetadataURL())
	}

	// 2. Follow the advertised URL (as a client would) to find the authorization server.
	fetch := func(abs string) map[string]any {
		u, err := url.Parse(abs)
		if err != nil || u.Scheme+"://"+u.Host != testOrigin {
			t.Fatalf("discovery URL %q is not on our origin", abs)
		}
		rec := h.send("GET", u.RequestURI(), "", "", nil, nil)
		if rec.Code != 200 {
			t.Fatalf("GET %s = %d", abs, rec.Code)
		}
		var m map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &m)
		return m
	}
	meta := fetch(metaURL)
	if meta["resource"] != testOrigin+"/mcp" {
		t.Fatalf("resource = %v", meta["resource"])
	}
	issuer := meta["authorization_servers"].([]any)[0].(string)
	as := fetch(issuer + "/.well-known/oauth-authorization-server")
	for _, k := range []string{"authorization_endpoint", "token_endpoint", "registration_endpoint"} {
		if !strings.HasPrefix(as[k].(string), testOrigin+"/api/oauth/") {
			t.Fatalf("%s = %v", k, as[k])
		}
	}

	// 3. Register, authorize (read-only), exchange.
	cid := h.registerClient("Claude", "https://claude.ai/api/mcp/auth_callback")
	code := h.authorizeCode(cid, "https://claude.ai/api/mcp/auth_callback", "plutus:read")
	tr := decodeToken(t, h.token(codeForm(cid, code, "https://claude.ai/api/mcp/auth_callback", testVerifier)))
	if tr.AccessToken == "" {
		t.Fatal("no access token")
	}

	// 4. Use the token against the MCP endpoint.
	resp = rpc(tr.AccessToken, initBody)
	if resp.StatusCode != 200 {
		t.Fatalf("initialize with token: %d", resp.StatusCode)
	}
	list := h.send("POST", "/mcp", "application/json", `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`, nil,
		map[string]string{"Authorization": "Bearer " + tr.AccessToken})
	if !strings.Contains(list.Body.String(), `"whoami"`) || strings.Contains(list.Body.String(), "write_something") {
		t.Fatalf("read-only grant must see only read tools: %s", list.Body.String())
	}
	callBody := func(tool string) string {
		return `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"` + tool + `","arguments":{}}}`
	}
	call := h.send("POST", "/mcp", "application/json", callBody("whoami"), nil, map[string]string{"Authorization": "Bearer " + tr.AccessToken})
	if !strings.Contains(call.Body.String(), "user=user-g-123") {
		t.Fatalf("whoami: %s", call.Body.String())
	}
	denied := h.send("POST", "/mcp", "application/json", callBody("write_something"), nil, map[string]string{"Authorization": "Bearer " + tr.AccessToken})
	if strings.Contains(denied.Body.String(), "written") || !strings.Contains(denied.Body.String(), `"isError":true`) {
		t.Fatalf("a read-only grant must not run write tools: %s", denied.Body.String())
	}

	// 5. Refresh keeps working through Hermes; the garbage token does not.
	tr2 := decodeToken(t, h.token(url.Values{"grant_type": {"refresh_token"}, "client_id": {cid}, "refresh_token": {tr.RefreshToken}}))
	if r := rpc(tr2.AccessToken, initBody); r.StatusCode != 200 {
		t.Fatalf("refreshed token rejected: %d", r.StatusCode)
	}
	if r := rpc("garbage", initBody); r.StatusCode != http.StatusUnauthorized {
		t.Fatalf("garbage token: %d", r.StatusCode)
	}
}
