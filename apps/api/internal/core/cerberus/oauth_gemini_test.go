package cerberus

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// Gemini's custom connector registers the way a server-side RFC 7591 client does. Public reports from other MCP
// servers show it sends six callbacks on oauth-redirect.googleusercontent.com (Google documents six) and asks for a
// client-secret login method. Refusing either made registration fail, and Gemini then asked the user for a client
// ID and secret instead. The ids below are made up.
func geminiRedirects() []string {
	out := []string{"https://oauth-redirect.googleusercontent.com/r/user_bound_custom-mcp-1234567890-mcp_example_com"}
	for i := 2; i <= 6; i++ {
		out = append(out, fmt.Sprintf("https://oauth-redirect.googleusercontent.com/r/user_bound_custom-mcp-1234567890-mcp_example_com_%d", i))
	}
	return out
}

func (h *oh) registerRaw(body map[string]any) (*http.Response, map[string]any) {
	h.t.Helper()
	b, _ := json.Marshal(body)
	rec := h.send("POST", "/api/oauth/register", "application/json", string(b), nil, nil)
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Result(), out
}

func basicAuth(id, secret string) map[string]string {
	return map[string]string{"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(id+":"+secret))}
}

func TestRegisterAcceptsServerSideClients(t *testing.T) {
	for _, method := range []string{"client_secret_post", "client_secret_basic"} {
		t.Run(method, func(t *testing.T) {
			h := newOH(t)
			res, out := h.registerRaw(map[string]any{
				"redirect_uris":              geminiRedirects(),
				"token_endpoint_auth_method": method,
				"grant_types":                []string{"authorization_code", "refresh_token"},
				"response_types":             []string{"code"},
			})
			if res.StatusCode != http.StatusCreated {
				t.Fatalf("status = %d, want 201: %v", res.StatusCode, out)
			}
			if out["token_endpoint_auth_method"] != "none" {
				t.Errorf("the response must say what the server actually does (none), got %v", out["token_endpoint_auth_method"])
			}
			if _, has := out["client_secret"]; has {
				t.Error("no secret is issued: clients are public and protected by PKCE")
			}
			if uris, _ := out["redirect_uris"].([]any); len(uris) != 6 {
				t.Errorf("redirect_uris = %v, want all six echoed back", out["redirect_uris"])
			}
		})
	}
}

func TestRegisterStillRefusesWhatItCannotServe(t *testing.T) {
	h := newOH(t)
	for _, method := range []string{"private_key_jwt", "tls_client_auth", "client_secret_jwt"} {
		res, out := h.registerRaw(map[string]any{"redirect_uris": []string{"https://a.example/cb"}, "token_endpoint_auth_method": method})
		if res.StatusCode != http.StatusBadRequest || out["error"] != "invalid_client_metadata" {
			t.Errorf("%s: status=%d %v, want 400 invalid_client_metadata", method, res.StatusCode, out)
		}
	}
	eleven := make([]string, 11)
	for i := range eleven {
		eleven[i] = fmt.Sprintf("https://a.example/cb%d", i)
	}
	if res, out := h.registerRaw(map[string]any{"redirect_uris": eleven}); res.StatusCode != http.StatusBadRequest || out["error"] != "invalid_redirect_uri" {
		t.Errorf("eleven redirect_uris: status=%d %v, want 400 invalid_redirect_uri", res.StatusCode, out)
	}
	ten := eleven[:10]
	if res, out := h.registerRaw(map[string]any{"redirect_uris": ten}); res.StatusCode != http.StatusCreated {
		t.Errorf("ten redirect_uris: status=%d %v, want 201", res.StatusCode, out)
	}
}

// A client that logs in with a secret sends its client_id in the Authorization header, not in the form. The
// server issued no secret, so the secret is ignored; the client is still protected by PKCE and the code binding.
func TestTokenEndpointAcceptsBasicClientAuth(t *testing.T) {
	h := newOH(t)
	uris := geminiRedirects()
	_, reg := h.registerRaw(map[string]any{"redirect_uris": uris, "token_endpoint_auth_method": "client_secret_basic"})
	cid, _ := reg["client_id"].(string)
	if cid == "" {
		t.Fatalf("registration failed: %v", reg)
	}
	form := func(code string) url.Values {
		return url.Values{"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {uris[0]}, "code_verifier": {testVerifier}}
	}
	post := func(f url.Values, hdr map[string]string) tokenResp {
		rec := h.send("POST", "/api/oauth/token", "application/x-www-form-urlencoded", f.Encode(), nil, hdr)
		tr := decodeToken(t, rec)
		if tr.AccessToken == "" && tr.Error == "" {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
		return tr
	}

	// client_secret_basic: id and (ignored) secret in the header, nothing in the form.
	tr := post(form(h.authorizeCode(cid, uris[0], "")), basicAuth(cid, "ignored-secret"))
	if tr.AccessToken == "" || tr.RefreshToken == "" {
		t.Fatalf("code exchange with Basic auth failed: %+v", tr)
	}
	// ... and the refresh, the same way.
	rf := post(url.Values{"grant_type": {"refresh_token"}, "refresh_token": {tr.RefreshToken}}, basicAuth(cid, "ignored-secret"))
	if rf.AccessToken == "" {
		t.Fatalf("refresh with Basic auth failed: %+v", rf)
	}
	// client_secret_post: id and a secret in the form; the secret is ignored.
	f := form(h.authorizeCode(cid, uris[0], ""))
	f.Set("client_id", cid)
	f.Set("client_secret", "ignored-secret")
	if tr := post(f, nil); tr.AccessToken == "" {
		t.Fatalf("code exchange with client_secret_post failed: %+v", tr)
	}
}

func TestTokenEndpointRejectsBadClientAuth(t *testing.T) {
	h := newOH(t)
	cid := h.registerClient("App", testRedirect)
	other := h.registerClient("Other", "http://localhost:8888/cb")
	cases := map[string]struct {
		formID string
		hdr    map[string]string
	}{
		"header and form name different clients": {other, basicAuth(cid, "x")},
		"unknown id in the header":               {"", basicAuth("nem_c_nope", "x")},
		"empty id in the header":                 {"", basicAuth("", "x")},
		"header is not base64":                   {"", map[string]string{"Authorization": "Basic !!!not-base64"}},
		"header has no colon":                    {"", map[string]string{"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("nocolon"))}},
		"bearer scheme is not client auth":       {"", map[string]string{"Authorization": "Bearer " + cid}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			code := h.authorizeCode(cid, testRedirect, "")
			f := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {testRedirect}, "code_verifier": {testVerifier}}
			if c.formID != "" {
				f.Set("client_id", c.formID)
			}
			rec := h.send("POST", "/api/oauth/token", "application/x-www-form-urlencoded", f.Encode(), nil, c.hdr)
			if tr := decodeToken(t, rec); rec.Code != http.StatusUnauthorized || tr.Error != "invalid_client" || tr.AccessToken != "" {
				t.Fatalf("status=%d body=%s, want 401 invalid_client", rec.Code, rec.Body.String())
			}
		})
	}
}

// When a client is refused at registration the server must say why in its own log, without writing anything a
// client sent that could be secret (paths and query strings of redirect URIs, the client name).
func TestRegistrationRejectionIsLoggedSafely(t *testing.T) {
	var buf bytes.Buffer
	h := newOHLog(t, slog.New(slog.NewTextHandler(&buf, nil)))
	res, _ := h.registerRaw(map[string]any{
		"client_name": "Super Secret Client Name",
		"redirect_uris": []string{
			"https://oauth-redirect.googleusercontent.com/r/user_bound_custom-mcp-1?secret=abc123",
			"https://a.example/cb?token=SECRET",
		},
		"token_endpoint_auth_method": "private_key_jwt",
	})
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
	log := buf.String()
	for _, want := range []string{"oauth client registration rejected", "code=invalid_client_metadata", "auth_method=private_key_jwt", "redirect_count=2", "https://oauth-redirect.googleusercontent.com", "https://a.example"} {
		if !strings.Contains(log, want) {
			t.Errorf("log lacks %q:\n%s", want, log)
		}
	}
	for _, secret := range []string{"abc123", "SECRET", "user_bound_custom-mcp", "Super Secret Client Name"} {
		if strings.Contains(log, secret) {
			t.Errorf("log must not contain %q:\n%s", secret, log)
		}
	}
}

// A refused authorize request is logged too (what was wrong, which client, where it would redirect) but never
// the state or the PKCE challenge, and an unknown scope is named so a client that asks for one can be diagnosed.
func TestAuthorizeRejectionIsLoggedSafely(t *testing.T) {
	var buf bytes.Buffer
	h := newOHLog(t, slog.New(slog.NewTextHandler(&buf, nil)))
	cid := h.registerClient("App", testRedirect)
	vals := authzValues(cid, testRedirect, "st-very-secret-state", "openid profile")
	rec := h.send("GET", "/api/oauth/authorize?"+vals.Encode(), "", "", []*http.Cookie{h.session}, nil)

	if shown := rec.Header().Get("Location") + rec.Body.String(); !strings.Contains(shown, "openid") {
		t.Errorf("the refusal must name the scope it did not recognise: %s", shown)
	}
	log := buf.String()
	for _, want := range []string{"oauth authorize rejected", "code=invalid_scope", "client_id=" + cid, "redirect_origin=http://localhost:7777"} {
		if !strings.Contains(log, want) {
			t.Errorf("log lacks %q:\n%s", want, log)
		}
	}
	for _, secret := range []string{"st-very-secret-state", challengeFor(testVerifier)} {
		if strings.Contains(log, secret) {
			t.Errorf("log must not contain %q:\n%s", secret, log)
		}
	}
}
