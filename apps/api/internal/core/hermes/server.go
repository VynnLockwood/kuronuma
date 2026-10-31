// Package hermes is the MCP gateway: it speaks the Model Context Protocol over
// Streamable HTTP (stateless JSON responses) so a user's AI assistant can call
// Kuronuma's tools. It depends only on the standard library.
//
// Scope: tools only. Server-initiated streams (GET /mcp) are not offered.
package hermes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"
)

// supportedVersions are the MCP protocol revisions this server can speak, newest first.
var supportedVersions = []string{"2025-11-25", "2025-06-18", "2025-03-26", "2024-11-05"}

// ErrUnauthorized is returned by an Authenticator for a missing, unknown or expired token.
var ErrUnauthorized = errors.New("hermes: unauthorized")

// Caller is the verified identity behind a request. UserID comes from the token, never from arguments.
type Caller struct {
	UserID   string
	ClientID string
	Scopes   []string
}

// HasScope reports whether the token grants scope.
func (c Caller) HasScope(scope string) bool { return slices.Contains(c.Scopes, scope) }

// Authenticator turns a bearer token into a Caller.
type Authenticator interface {
	Authenticate(ctx context.Context, bearer string) (Caller, error)
}

// Options configures a Server.
type Options struct {
	Name                string
	Title               string
	Version             string
	Instructions        string
	ResourceMetadataURL string // advertised in WWW-Authenticate so clients can find the OAuth server
}

// Server is an http.Handler for the MCP endpoint.
type Server struct {
	opt   Options
	auth  Authenticator
	log   *slog.Logger
	tools map[string]Tool
	order []string
}

// New creates a Server. Register tools with AddTool before serving.
func New(opt Options, auth Authenticator, log *slog.Logger) *Server {
	return &Server{opt: opt, auth: auth, log: log, tools: map[string]Tool{}}
}

// ToolError is an error whose message is safe to show to the assistant (validation problems etc.).
type ToolError struct{ Msg string }

func (e ToolError) Error() string { return e.Msg }

// Errorf builds a ToolError.
func Errorf(format string, a ...any) error { return ToolError{Msg: fmt.Sprintf(format, a...)} }

// Tool describes one callable tool.
type Tool struct {
	Name        string
	Title       string
	Description string
	InputSchema map[string]any
	Scope       string // required OAuth scope
	ReadOnly    bool
	Destructive bool
	// Handler returns the text for the assistant. A ToolError becomes an error result the
	// assistant can read; any other error is logged and reported generically.
	Handler func(ctx context.Context, c Caller, args json.RawMessage) (string, error)
}

// AddTool registers a tool. It panics on a duplicate name (a programming error).
func (s *Server) AddTool(t Tool) {
	if t.Name == "" || t.Handler == nil || t.Scope == "" {
		panic("hermes: tool needs a name, handler and scope")
	}
	if _, dup := s.tools[t.Name]; dup {
		panic("hermes: duplicate tool " + t.Name)
	}
	s.tools[t.Name] = t
	s.order = append(s.order, t.Name)
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

const (
	errParse          = -32700
	errInvalidRequest = -32600
	errMethodNotFound = -32601
	errInvalidParams  = -32602
)

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		// No server-initiated SSE stream is offered.
		w.Header().Set("Allow", "POST, OPTIONS")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	caller, ok := s.authenticate(w, r)
	if !ok {
		return
	}
	if v := r.Header.Get("MCP-Protocol-Version"); v != "" && !slices.Contains(supportedVersions, v) {
		http.Error(w, "unsupported MCP-Protocol-Version", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "request body too large or unreadable", http.StatusRequestEntityTooLarge)
		return
	}
	if t := bytes.TrimSpace(body); len(t) > 0 && t[0] == '[' {
		writeRPC(w, http.StatusBadRequest, rpcResponse{Error: &rpcError{errInvalidRequest, "batch requests are not supported"}})
		return
	}
	var req rpcRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeRPC(w, http.StatusBadRequest, rpcResponse{Error: &rpcError{errParse, "invalid JSON"}})
		return
	}
	// A response or notification from the client needs no reply.
	if req.Method == "" || len(req.ID) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	if req.JSONRPC != "2.0" || string(req.ID) == "null" {
		writeRPC(w, http.StatusBadRequest, rpcResponse{Error: &rpcError{errInvalidRequest, "invalid JSON-RPC request"}})
		return
	}

	result, rerr := s.dispatch(r.Context(), caller, req)
	writeRPC(w, http.StatusOK, rpcResponse{ID: req.ID, Result: result, Error: rerr})
}

func writeRPC(w http.ResponseWriter, status int, resp rpcResponse) {
	resp.JSONRPC = "2.0"
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) authenticate(w http.ResponseWriter, r *http.Request) (Caller, bool) {
	scheme, tok, found := strings.Cut(r.Header.Get("Authorization"), " ")
	tok = strings.TrimSpace(tok)
	if !found || !strings.EqualFold(scheme, "Bearer") || tok == "" {
		s.unauthorized(w, false)
		return Caller{}, false
	}
	c, err := s.auth.Authenticate(r.Context(), tok)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			s.unauthorized(w, true)
		} else {
			// Do not return 401 for our own outages: clients would discard good tokens.
			s.log.Error("mcp authenticate failed", "err", err)
			http.Error(w, "temporarily unavailable", http.StatusServiceUnavailable)
		}
		return Caller{}, false
	}
	return c, true
}

func (s *Server) unauthorized(w http.ResponseWriter, invalid bool) {
	v := `Bearer resource_metadata="` + s.opt.ResourceMetadataURL + `"`
	if invalid {
		v = `Bearer error="invalid_token", resource_metadata="` + s.opt.ResourceMetadataURL + `"`
	}
	w.Header().Set("WWW-Authenticate", v)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}

func (s *Server) dispatch(ctx context.Context, c Caller, req rpcRequest) (any, *rpcError) {
	switch req.Method {
	case "initialize":
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(req.Params, &p)
		version := supportedVersions[0]
		if slices.Contains(supportedVersions, p.ProtocolVersion) {
			version = p.ProtocolVersion
		}
		res := map[string]any{
			"protocolVersion": version,
			"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
			"serverInfo":      map[string]any{"name": s.opt.Name, "title": s.opt.Title, "version": s.opt.Version},
		}
		if s.opt.Instructions != "" {
			res["instructions"] = s.opt.Instructions
		}
		return res, nil
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return map[string]any{"tools": s.listTools(c)}, nil
	case "tools/call":
		return s.callTool(ctx, c, req.Params)
	default:
		return nil, &rpcError{errMethodNotFound, "method not found: " + cleanName(req.Method)}
	}
}

func (s *Server) listTools(c Caller) []map[string]any {
	out := []map[string]any{}
	for _, name := range s.order {
		t := s.tools[name]
		if !c.HasScope(t.Scope) {
			continue // only advertise what this token may call
		}
		out = append(out, map[string]any{
			"name":        t.Name,
			"title":       t.Title,
			"description": t.Description,
			"inputSchema": t.InputSchema,
			"annotations": map[string]any{
				"title":           t.Title,
				"readOnlyHint":    t.ReadOnly,
				"destructiveHint": t.Destructive,
				"openWorldHint":   false,
			},
		})
	}
	return out
}

func (s *Server) callTool(ctx context.Context, c Caller, params json.RawMessage) (any, *rpcError) {
	var p struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil || p.Name == "" {
		return nil, &rpcError{errInvalidParams, "tools/call needs a tool name"}
	}
	t, ok := s.tools[p.Name]
	if !ok {
		return nil, &rpcError{errInvalidParams, "unknown tool: " + cleanName(p.Name)}
	}
	if !c.HasScope(t.Scope) {
		return toolResult("This connection does not have permission for this tool (needs scope "+t.Scope+").", true), nil
	}
	args := p.Arguments
	if len(bytes.TrimSpace(args)) == 0 || string(bytes.TrimSpace(args)) == "null" {
		args = json.RawMessage(`{}`)
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	start := time.Now()
	text, err := s.safeCall(ctx, t, c, args)
	ok = err == nil
	// Log who called what, never the arguments or results.
	s.log.Info("mcp tool call", "tool", t.Name, "user_id", c.UserID, "client_id", c.ClientID, "ok", ok, "ms", time.Since(start).Milliseconds())

	if err != nil {
		var te ToolError
		if errors.As(err, &te) {
			return toolResult(te.Msg, true), nil
		}
		s.log.Error("tool failed", "tool", t.Name, "err", err)
		return toolResult("Something went wrong on the server. Please try again.", true), nil
	}
	return toolResult(text, false), nil
}

// safeCall runs a handler and converts a panic into an error so one bad tool cannot crash the server.
func (s *Server) safeCall(ctx context.Context, t Tool, c Caller, args json.RawMessage) (text string, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("panic in tool %s: %v", t.Name, rec)
		}
	}()
	return t.Handler(ctx, c, args)
}

func toolResult(text string, isError bool) map[string]any {
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
		"isError": isError,
	}
}

// cleanName keeps client-supplied names short and printable when echoed in errors.
func cleanName(s string) string { return Untrusted(s, 60) }

// setCORS allows browser-based MCP clients. Requests are authenticated by bearer token, not cookies.
func setCORS(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, MCP-Protocol-Version, Mcp-Session-Id")
	h.Set("Access-Control-Expose-Headers", "WWW-Authenticate")
}
