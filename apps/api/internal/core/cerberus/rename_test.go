package cerberus

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

// The session cookie was renamed from nem_session to kuro_session [keep-nemosyne]. There is deliberately no fallback to the old
// name: people sign in again once, and nothing keeps honouring the old cookie.
func TestOldSessionCookieNameIsNotAccepted(t *testing.T) {
	h := newOH(t)
	me := func(name string) int {
		rec := h.send("GET", "/api/auth/me", "", "", []*http.Cookie{{Name: name, Value: h.session.Value}}, nil)
		return rec.Code
	}
	if got := me("kuro_session"); got != http.StatusOK {
		t.Fatalf("kuro_session: status %d, want 200", got)
	}
	if got := me("nem_session"); got != http.StatusUnauthorized { // [keep-nemosyne]
		t.Fatalf("old cookie name: status %d, want 401", got)
	}
	if h.session.Name != "kuro_session" {
		t.Fatalf("the server sets the cookie %q, want kuro_session", h.session.Name)
	}
}

func TestNewClientsGetTheKuroPrefix(t *testing.T) {
	h := newOH(t)
	if id := h.registerClient("App", testRedirect); !strings.HasPrefix(id, "kuro_c_") {
		t.Fatalf("client id %q, want prefix kuro_c_", id)
	}
}

// Clients registered before the rename keep their nem_c_ ids in the database and must keep working. [keep-nemosyne]
func TestLegacyClientIDsStillWork(t *testing.T) {
	h := newOH(t)
	const legacy = "nem_c_legacy" // [keep-nemosyne]
	if err := h.store.CreateClient(context.Background(), OAuthClient{ID: legacy, Name: "Old app", RedirectURIs: []string{testRedirect}, CreatedAt: h.now}); err != nil {
		t.Fatal(err)
	}
	code := h.authorizeCode(legacy, testRedirect, "")
	rec := h.token(codeForm(legacy, code, testRedirect, testVerifier))
	if tr := decodeToken(t, rec); rec.Code != http.StatusOK || tr.AccessToken == "" {
		t.Fatalf("legacy client exchange: status=%d body=%s", rec.Code, rec.Body.String())
	}
}
