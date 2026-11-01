package plutus

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/core/hermes"
)

// cookieSession stands in for the real session: the cookie "s" names the user.
type cookieSession struct{}

func (cookieSession) UserID(r *http.Request) (string, bool) {
	c, err := r.Cookie("s")
	if err != nil || (c.Value != "alice" && c.Value != "bob") {
		return "", false
	}
	return c.Value, true
}

// The real spending tools, reached through the web bridge, keep users apart in both directions, and a record saved
// through the bridge is the same record the assistants see.
func TestSpendingThroughTheWebBridgeIsPerUser(t *testing.T) {
	st := &memStore{}
	st.setNames("alice", "food")
	st.setNames("bob", "food")
	srv := hermes.New(hermes.Options{Name: "t", ResourceMetadataURL: "https://x/meta"}, auth{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	New(st, func() time.Time { return time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC) }).Register(srv)
	web := hermes.WebHandler(srv, cookieSession{}, "https://kuronuma.example")

	bridge := func(user, tool string, args map[string]any) (int, map[string]any) {
		body, _ := json.Marshal(map[string]any{"tool": tool, "arguments": args})
		r := httptest.NewRequest(http.MethodPost, "/api/web/call", strings.NewReader(string(body)))
		r.Header.Set("X-Kuronuma-Web", "1")
		r.Header.Set("Content-Type", "application/json")
		r.AddCookie(&http.Cookie{Name: "s", Value: user})
		rec := httptest.NewRecorder()
		web.ServeHTTP(rec, r)
		var out map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("not JSON: %q", rec.Body.String())
		}
		return rec.Code, out
	}
	listed := func(user string) int {
		code, out := bridge(user, "list_transactions", map[string]any{})
		if code != 200 {
			t.Fatalf("%s list: %d %v", user, code, out)
		}
		return int(out["result"].(map[string]any)["count"].(float64))
	}

	code, out := bridge("alice", "add_transaction", map[string]any{"amount": "120.50", "currency": "THB", "merchant": "Grab", "occurred_on": "2026-10-28", "category": "food"})
	if code != 200 || out["result"].(map[string]any)["status"] != "saved" {
		t.Fatalf("alice add: %d %v", code, out)
	}
	id := out["result"].(map[string]any)["transaction"].(map[string]any)["id"].(string)

	if listed("alice") != 1 || listed("bob") != 0 {
		t.Fatalf("alice sees %d, bob sees %d; want 1 and 0", listed("alice"), listed("bob"))
	}
	// Bob cannot delete alice's record by id: it looks like a miss, and alice still has it.
	if code, out := bridge("bob", "delete_transaction", map[string]any{"id": id}); code != http.StatusUnprocessableEntity || !strings.Contains(out["error"].(string), "No record") {
		t.Fatalf("bob delete: %d %v", code, out)
	}
	if listed("alice") != 1 {
		t.Fatal("alice's record must survive bob's attempt")
	}
	// And the other way round.
	if code, _ := bridge("bob", "add_transaction", map[string]any{"amount": "10", "currency": "THB", "merchant": "Cafe", "occurred_on": "2026-10-28"}); code != 200 {
		t.Fatalf("bob add: %d", code)
	}
	if listed("alice") != 1 || listed("bob") != 1 {
		t.Fatalf("after bob added: alice %d bob %d", listed("alice"), listed("bob"))
	}
	// A refused input is a 422 with the tool's own text.
	code, out = bridge("alice", "add_transaction", map[string]any{"amount": "5", "currency": "baht", "merchant": "x", "occurred_on": "2026-10-28"})
	if code != http.StatusUnprocessableEntity || !strings.Contains(out["error"].(string), "currency") {
		t.Fatalf("refused input: %d %v", code, out)
	}
	// The assistants (MCP, bearer tokens) see the same records: alice's one, and not bob's.
	e := &env{t: t, srv: srv, store: st}
	if list, _, _ := e.call("alice", "list_transactions", map[string]any{}); list["count"].(float64) != 1 {
		t.Fatalf("alice over MCP: %v", list)
	}
	// No session, no data.
	r := httptest.NewRequest(http.MethodPost, "/api/web/call", strings.NewReader(`{"tool":"list_transactions"}`))
	r.Header.Set("X-Kuronuma-Web", "1")
	r.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	web.ServeHTTP(rec, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no session: %d", rec.Code)
	}
}
