package hermes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"
)

func webCaller(s *Server) Caller {
	return Caller{UserID: "u-web", ClientID: "web", Scopes: s.AllScopes()}
}

func TestCallAsRunsATool(t *testing.T) {
	s := newServer()
	text, err := s.CallAs(context.Background(), webCaller(s), "echo", json.RawMessage(`{"a":1}`))
	if err != nil || text != `user=u-web args={"a":1}` {
		t.Fatalf("got %q, %v", text, err)
	}
	// No arguments, or null, reach the handler as an empty object.
	for _, raw := range []json.RawMessage{nil, json.RawMessage(``), json.RawMessage(`null`), json.RawMessage(`  `)} {
		text, err := s.CallAs(context.Background(), webCaller(s), "echo", raw)
		if err != nil || text != `user=u-web args={}` {
			t.Errorf("args %q: got %q, %v", raw, text, err)
		}
	}
}

func TestCallAsErrors(t *testing.T) {
	s := newServer()
	ctx := context.Background()
	if _, err := s.CallAs(ctx, webCaller(s), "nope", nil); !errors.Is(err, ErrUnknownTool) {
		t.Errorf("unknown tool: %v", err)
	}
	if _, err := s.CallAs(ctx, Caller{UserID: "u", Scopes: []string{"t:read"}}, "danger", nil); !errors.Is(err, ErrMissingScope) {
		t.Errorf("missing scope: %v", err)
	}
	_, err := s.CallAs(ctx, webCaller(s), "invalid", nil)
	var te ToolError
	if !errors.As(err, &te) || te.Msg != "amount must be positive" {
		t.Errorf("tool error: %v", err)
	}
	_, err = s.CallAs(ctx, webCaller(s), "broken", nil)
	if err == nil || errors.As(err, &te) {
		t.Errorf("a failure that is not a ToolError must stay a plain error: %v", err)
	}
	_, err = s.CallAs(ctx, webCaller(s), "panics", nil)
	if err == nil || !strings.Contains(err.Error(), "panic") {
		t.Errorf("a panic must come back as an error: %v", err)
	}
}

func TestCallAsAppliesTheToolTimeout(t *testing.T) {
	s := newServer()
	var left time.Duration
	s.AddTool(Tool{
		Name: "deadline", Scope: "t:read", InputSchema: map[string]any{"type": "object"},
		Handler: func(ctx context.Context, _ Caller, _ json.RawMessage) (string, error) {
			d, ok := ctx.Deadline()
			if !ok {
				return "", errors.New("no deadline")
			}
			left = time.Until(d)
			return "ok", nil
		},
	})
	if _, err := s.CallAs(context.Background(), webCaller(s), "deadline", nil); err != nil {
		t.Fatal(err)
	}
	if left <= 15*time.Second || left > 20*time.Second {
		t.Fatalf("deadline in %v, want about 20s", left)
	}
}

func TestAllScopesIsTheSortedSetOfToolScopes(t *testing.T) {
	s := newServer()
	if got, want := s.AllScopes(), []string{"t:read", "t:write"}; !slices.Equal(got, want) {
		t.Fatalf("AllScopes = %v, want %v", got, want)
	}
}

// One log line per call: who and what, never the arguments or the result.
func TestCallAsLogsWhoAndWhatButNotArgumentsOrResults(t *testing.T) {
	var buf bytes.Buffer
	s := New(Options{Name: "t"}, fakeAuth{}, slog.New(slog.NewTextHandler(&buf, nil)))
	s.AddTool(Tool{
		Name: "secretive", Scope: "t:read", InputSchema: map[string]any{"type": "object"},
		Handler: func(context.Context, Caller, json.RawMessage) (string, error) { return "RESULT-SECRET-9", nil },
	})
	if _, err := s.CallAs(context.Background(), webCaller(s), "secretive", json.RawMessage(`{"note":"ARG-SECRET-7"}`)); err != nil {
		t.Fatal(err)
	}
	log := buf.String()
	for _, want := range []string{"tool=secretive", "user_id=u-web", "client_id=web", "ok=true", "ms="} {
		if !strings.Contains(log, want) {
			t.Errorf("log lacks %q:\n%s", want, log)
		}
	}
	if n := strings.Count(log, "\n"); n != 1 {
		t.Errorf("want one log line, got %d:\n%s", n, log)
	}
	for _, secret := range []string{"ARG-SECRET-7", "RESULT-SECRET-9"} {
		if strings.Contains(log, secret) {
			t.Errorf("log must not contain %q:\n%s", secret, log)
		}
	}
}
