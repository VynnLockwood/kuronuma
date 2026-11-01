package hermes

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/VynnLockwood/kuronuma-mcp/apps/api/internal/core/webguard"
)

// SessionUser resolves the signed-in user of a request from the session cookie. main.go adapts the sign-in code to
// it, so this package does not depend on it.
type SessionUser interface {
	UserID(r *http.Request) (string, bool)
}

const (
	maxWebBody        = 64 << 10
	genericWebFailure = "Something went wrong on the server. Please try again."
)

// WebHandler lets the website's own pages run the registered tools as the signed-in user: the dashboard uses the
// same handlers as the assistants, so there is one place where the rules live.
//
// The request is {"tool": "...", "arguments": {...}}. The user comes only from the verified session cookie, never
// from the body. The checks run in this order: method, the browser checks (webguard), the session, the body size,
// the JSON, the tool name. A request without a session never gets its body read. Nothing here sends CORS headers.
func WebHandler(s *Server, auth SessionUser, publicURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeWeb(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
			return
		}
		if err := webguard.Check(r, publicURL); err != nil {
			s.log.Warn("web call refused", "reason", err.Error())
			writeWeb(w, http.StatusForbidden, map[string]any{"error": "forbidden"})
			return
		}
		userID, ok := auth.UserID(r)
		if !ok || userID == "" {
			writeWeb(w, http.StatusUnauthorized, map[string]any{"error": "not signed in"})
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebBody))
		if err != nil {
			var tooBig *http.MaxBytesError
			if errors.As(err, &tooBig) {
				writeWeb(w, http.StatusRequestEntityTooLarge, map[string]any{"error": "request too large"})
				return
			}
			writeWeb(w, http.StatusBadRequest, map[string]any{"error": "unreadable request"})
			return
		}
		var req struct {
			Tool      string          `json:"tool"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(body, &req); err != nil || req.Tool == "" {
			writeWeb(w, http.StatusBadRequest, map[string]any{"error": "send {\"tool\": ..., \"arguments\": ...}"})
			return
		}

		caller := Caller{UserID: userID, ClientID: "web", Scopes: s.AllScopes()}
		text, err := s.CallAs(r.Context(), caller, req.Tool, req.Arguments)
		if err != nil {
			var te ToolError
			switch {
			case errors.Is(err, ErrUnknownTool):
				writeWeb(w, http.StatusNotFound, map[string]any{"error": "unknown tool"})
			case errors.As(err, &te):
				writeWeb(w, http.StatusUnprocessableEntity, map[string]any{"error": te.Msg})
			default:
				// CallAs has logged the cause of a failing tool; nothing about it goes to the browser.
				writeWeb(w, http.StatusInternalServerError, map[string]any{"error": genericWebFailure})
			}
			return
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(text), &result); err != nil || result == nil {
			s.log.Error("web call: tool result is not a JSON object", "tool", req.Tool)
			writeWeb(w, http.StatusInternalServerError, map[string]any{"error": genericWebFailure})
			return
		}
		writeWeb(w, http.StatusOK, map[string]any{"result": result})
	})
}

func writeWeb(w http.ResponseWriter, status int, body map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
