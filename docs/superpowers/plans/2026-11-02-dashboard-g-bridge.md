# Dashboard PR G: the web bridge Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline) or superpowers:subagent-driven-development. Steps use checkbox (`- [ ]`) syntax.

**Goal:** `POST /api/web/call` runs one registered tool as the signed-in user, so the dashboard can use the same handlers as the assistants. No new data code, no migration, no new dependency.

**Spec:** `docs/superpowers/specs/2026-10-31-web-dashboard-design.md` (section 1). This is PR G of five (G bridge, H shell + Overview + Spending, I Tasks and notes, J Categories and Payee rules, K Connected assistants). Later PRs depend on this one.

**Architecture:** `hermes` gains `Server.CallAs` (shares one run path with MCP `tools/call`: 20 s timeout, panic recovery, `ToolError` handling, one log line), `Server.AllScopes`, and `WebHandler`. The browser-request checks live in a new leaf package `internal/core/webguard` so PR K can reuse them without editing `hermes`. `main.go` adapts `cerberus.Handler.UserFromRequest` to the one-method `SessionUser` interface; the `cerberus` package is not edited.

## Global Constraints

- The user id comes only from the verified session cookie. Nothing in the body can set it; `arguments` goes to the tool unchanged and tools already reject unknown arguments.
- Caller: `Caller{UserID, ClientID: "web", Scopes: Server.AllScopes()}`.
- Request rules (webguard): method POST; header `X-Kuronuma-Web: 1` (ruling: the spec says `X-Nemosyne-Web`; the product is now Kuronuma and this header is new, so it carries the new name); `Content-Type` media type `application/json`; `Origin`, when present, equal to the origin of `PUBLIC_URL`; `Sec-Fetch-Site`, when present, must be `same-origin` or `none` (ruling: the spec rejects `cross-site`; `same-site` is also refused because a sibling subdomain is not our page). No CORS headers are ever sent. All responses `Cache-Control: no-store`.
- Check order: method (405), webguard (403), session (401), body size (413), JSON and tool name (400), unknown tool (404), run. An unauthenticated caller never gets a body read.
- Body limit 64 KB (`http.MaxBytesReader`).
- Responses exactly as the spec's table. 500 body is the generic sentence; causes are logged.
- Logging: tool name, user id, client `web`, ok, milliseconds. Never arguments or results.
- nginx: `location /api/web/` before `location /api/`, its own zone (300 requests a minute per IP, burst 60), `client_max_body_size 64k`.
- Owner approval needed (AGENTS.md): this adds a new way to act as a signed-in user, and touches `deploy/nginx/nginx.conf`. Approving this plan gives both.
- Commits: `VynnLockwood <120536136+VynnLockwood@users.noreply.github.com>`, no AI trailers. Conventional Commits. Branch `feat/web-bridge`.

## Review Focus

- Isolation both ways: alice's bridge call can never read or change bob's data and the reverse; the body and the header cannot choose the user; every path that reaches a tool has a verified session.
- CSRF: each webguard rule is enforced and a request missing any one is refused; no CORS headers; a cross-site form POST (no custom header) is refused.
- No path logs arguments, results or the cookie; the 500 path returns only the generic sentence.
- `CallAs` and MCP `tools/call` share the run path (no copy that can drift); the MCP behaviour and messages are unchanged.
- The nginx location is matched before `/api/` and the body limit and rate limit apply.

---

### Task 1: `CallAs`, `AllScopes` and the shared run path

**Files:** Modify `apps/api/internal/core/hermes/server.go`, `server_test.go`.

**Interfaces:** Produces `var ErrUnknownTool`, `var ErrMissingScope`, `func (s *Server) AllScopes() []string` (sorted, unique), `func (s *Server) CallAs(ctx context.Context, c Caller, name string, args json.RawMessage) (string, error)`; the tool's own `ToolError` is returned as is.

- [ ] **Step 1: Write the failing tests:** `CallAs` runs a toy tool and returns its text; an unknown name returns `ErrUnknownTool`; a caller without the tool's scope returns `ErrMissingScope`; a `ToolError` comes back as a `ToolError`; a panic comes back as an error (not a crash); empty or `null` arguments reach the handler as `{}`; the 20 s timeout is applied (a handler sees a context deadline); `AllScopes` is the sorted set of the registered tools' scopes; one log line per call has tool, user id, client id, ok and ms and contains no arguments or results; the existing MCP `tools/call` tests still pass unchanged.
- [ ] **Step 2: Run them.** Expected: FAIL (methods missing).
- [ ] **Step 3: Implement** by moving the timeout, `safeCall` and the log line out of `callTool` into one `run` used by both; `callTool` keeps its scope message and result formatting.
- [ ] **Step 4: Verify.** `gotest.sh`, `gotest.sh vet`, `gofmt -l apps/api`.
- [ ] **Step 5: Commit** `feat(hermes): add CallAs so non-MCP callers share the tool run path`.

### Task 2: `webguard`

**Files:** Create `apps/api/internal/core/webguard/webguard.go`, `webguard_test.go`.

**Interfaces:** Produces `func Check(r *http.Request, publicURL string) error` (nil when the request passes) and exported sentinel errors per rule; `const Header = "X-Kuronuma-Web"`.

- [ ] **Step 1: Write the failing tests** (table): passes with all rules met; fails when the header is missing, has another value, when the content type is `text/plain`, `application/x-www-form-urlencoded`, `multipart/form-data`, missing, or `application/json; charset=utf-8` (that one passes), when `Origin` is another origin, a scheme or port variant, or `null`, when `Sec-Fetch-Site` is `cross-site` or `same-site`; passes with no `Origin` and no `Sec-Fetch-Site` (non-browser clients) and with `Sec-Fetch-Site: same-origin`; an invalid `publicURL` makes every request fail closed.
- [ ] **Step 2: Run.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Verify** suite, vet, gofmt (note the harness copies packages named in `gotest.sh`; add `webguard` to the copied list in the scratch script).
- [ ] **Step 5: Commit** `feat(webguard): add the browser-request checks for the web bridge`.

### Task 3: `WebHandler`

**Files:** Create `apps/api/internal/core/hermes/web.go`, `web_test.go`.

**Interfaces:** Consumes Tasks 1 and 2. Produces `type SessionUser interface { UserID(r *http.Request) (string, bool) }` and `func WebHandler(s *Server, auth SessionUser, publicURL string) http.Handler`.

- [ ] **Step 1: Write the failing tests** (toy tools, fake session): 405 for GET, PUT, DELETE; 403 for each webguard failure (header, content type, Origin, Sec-Fetch-Site); 401 with `{"error":"not signed in"}` (and the body is not read: a huge body still gets 401); 413 over 64 KB; 400 for invalid JSON and a missing tool; 404 `{"error":"unknown tool"}`; 200 `{"result": {...}}` with the tool's object; no `arguments` is the same as `{}`; 422 `{"error": "<ToolError text>"}`; 500 with only the generic sentence and the cause in the log; every response has `Cache-Control: no-store` and none has an `Access-Control-*` header; the body cannot choose the user (`{"tool":"whoami","user_id":"bob","arguments":{"user_id":"bob"}}` returns alice's id from a toy `whoami` tool, and a tool that rejects unknown arguments returns 422); two toy users with per-user data: alice reads only alice's, bob only bob's, both ways; log output has the tool, user id, client `web` and `ok` but never an argument value or a result value.
- [ ] **Step 2: Run.** Expected: FAIL.
- [ ] **Step 3: Implement** per the spec's table and the check order above; `result` is the tool text parsed as a JSON object (a tool whose text is not a JSON object is a 500 and an error log).
- [ ] **Step 4: Verify** suite, vet, gofmt.
- [ ] **Step 5: Commit** `feat(hermes): add the web bridge handler`.

### Task 4: Real modules through the bridge, wiring and nginx

**Files:** Modify `apps/api/cmd/kuronuma-api/main.go`, `deploy/nginx/nginx.conf`, `apps/api/internal/modules/plutus/plutus_test.go` (bridge isolation test), `docs/design.md`.

- [ ] **Step 1: Write the failing tests:** in the `plutus` tests (they already import `hermes`), build a server with the plutus module and its fake store, wrap it in `WebHandler` with a fake session, and check: alice adds a transaction through the bridge and lists it; bob's list through the bridge is empty; bob cannot delete alice's id (not found); a refused input returns 422 with the tool's text; the saved record is visible to the MCP path for alice and not bob. nginx: a syntax check with a dummy environment (`nginx -t` is run by CI; locally compare the file against the plan: location order, zone, body size).
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement.** `main.go`: a small adapter type that calls `auth.UserFromRequest` and returns the user id; `mux.Handle("POST /api/web/call", ...)` (the handler itself answers 405 for other methods, so register the path without a method); pass `PUBLIC_URL`. nginx: the new zone and location. `docs/design.md`: a "Web bridge" section (endpoint, checks, identity, logging, limits).
- [ ] **Step 4: Verify.** `gotest.sh`, vet, gofmt; the deploy check from AGENTS.md (`docker compose -f deploy/docker-compose.yml config -q` with dummy env, nginx syntax by reading; CI runs `nginx -t`).
- [ ] **Step 5: Commit** `feat(api): serve the web bridge and rate-limit it in nginx`.

### Task 5: Review and PR

- [ ] Final checks, whole-branch review by a fresh Opus reviewer with this Review Focus (one fix pass, tests first), agent log in `agent-logs/claude/`, PR `feat(api): add the web bridge for the dashboard` with the template (Auth/OAuth is not edited; check "Deploy, CI, or nginx"), the real CI result, "Rulings I made" and "Deferred minors" in the final message.
