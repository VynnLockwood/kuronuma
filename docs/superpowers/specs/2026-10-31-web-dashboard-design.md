# Web dashboard (design)

Date: 2026-10-31. Status: for owner review. Builds on `docs/design.md` and the memory and categories spec.

## Goal

A signed-in user sees everything Nemo stores about them, as charts and lists, and can manage it in the browser, not only through an assistant. Everything on screen is that user's own data, and every request is tied to the signed-in user.

## Decisions (from the owner)

| Question | Decision |
|---|---|
| How much can the web change? | View everything and manage all of it. Spending can be added by hand and deleted, not edited (matching the assistant tools). |
| Languages | English and Thai, switchable, like the landing page. |
| How does the browser reach the data? | A bridge to the existing tool handlers (approach A), not a second REST API. |
| Charts | Hand-built SVG and CSS using the design tokens. No new dependency. |

Not included in this work: data export, account deletion, editing a spending record, paging past 200 records, push notifications or reminders.

## 1. The bridge endpoint (API)

`POST /api/web/call` runs one registered tool for the signed-in user and returns its result.

- **Request:** `Content-Type: application/json`, header `X-Nemosyne-Web: 1`, the session cookie. Body `{"tool": "list_memories", "arguments": {...}}`, at most 64 KB. `arguments` is optional.
- **Responses** (all `Cache-Control: no-store`):

| Status | Body | When |
|---|---|---|
| 200 | `{"result": {...}}` | The tool ran; `result` is the tool's JSON object. |
| 400 | `{"error": "..."}` | Body is not valid JSON or has no tool name. |
| 401 | `{"error": "not signed in"}` | No valid session. |
| 403 | `{"error": "forbidden"}` | The CSRF checks below failed. |
| 404 | `{"error": "unknown tool"}` | No such tool. |
| 405 | | Any method other than POST. |
| 413 | | Body over 64 KB. |
| 422 | `{"error": "<text>"}` | The tool refused the input (a `ToolError`); the text is written for people and shown in the UI. |
| 500 | `{"error": "Something went wrong on the server. Please try again."}` | Any other failure. Details are logged, not returned. |

- **Identity:** the user id comes only from the verified session cookie (`UserFromRequest`). Nothing in the body can set it; tools already reject unknown arguments. The caller is `Caller{UserID, ClientID: "web", Scopes: every scope of the registered tools}`: the web session is the user acting as themselves, so it needs no consent screen.
- **CSRF:** the cookie stays `SameSite=Lax`. The endpoint also requires the custom header (a cross-site page cannot send it without a CORS preflight, and no CORS headers are served), requires `Content-Type: application/json`, and rejects any request whose `Origin` is present and not `PUBLIC_URL`, or whose `Sec-Fetch-Site` is `cross-site`. POST only.
- **Where it lives:** `internal/core/hermes/web.go` adds `Server.CallAs(ctx, caller, name, args)`, which shares the code path of MCP `tools/call` (argument decoding, 20 s timeout, panic recovery, `ToolError` handling), and `WebHandler(auth SessionUser, publicURL string)`. `SessionUser` is a one-method interface (`UserID(r) (string, bool)`); `main.go` adapts `cerberus` to it. The `cerberus` package is not edited.
- **Logging:** tool name, user id, client `web`, ok, milliseconds. Never arguments or results.
- **nginx:** new `location /api/web/` with its own rate limit (300 requests a minute per IP, burst 60) and `client_max_body_size 64k`, placed before the general `/api/` block.
- **Known limits, inherited from the tools:** results are capped at 200 records; text goes through `hermes.Untrusted`, which collapses line breaks (the open "line breaks in note bodies" decision becomes more visible once there is a web view).

## 2. The web app

- **Routes:** `/dashboard` and `/th/dashboard` render the same components with a `lang` prop. The tab is in the query string (`?tab=spending`). Not signed in: the existing sign-in card. API unreachable: the existing notice.
- **Files:** `lib/web.ts` (typed `callTool(name, args)`: sets the header, parses the three outcomes, never throws raw HTML), `lib/dashboard-copy.ts` (EN and TH), one component per tab under `components/dashboard/`, a small `charts.tsx` (`BarList`, `Columns`), and a native `<dialog>` confirm for every destructive action.
- **Rendering rule:** every piece of user text is a React text node. No `dangerouslySetInnerHTML`, no user-controlled URLs.
- **Tabs:**
  - **Overview:** this month's spending per currency, spending by category (bars), the last 6 months (columns), counts of open, overdue and due-this-week tasks, the five newest notes. Empty states link to `/connect`. Built from `get_spending_summary` (by category, by month), `list_memories`, `list_categories`.
  - **Spending:** date and category filters; a list (newest first, up to 200, with a note when it is full); a summary switch (category, month, merchant) drawn as bars with a table underneath; an add form (amount, currency, merchant, date, category with the user's names suggested, note); delete with confirmation. Handles `possible_duplicate` by asking, then retrying with `allow_duplicate`.
  - **Tasks and notes:** filters (kind, status, category, tag, search), open items first; add; edit (title, details, due date, priority, category, tags); complete, reopen, archive; delete with confirmation.
  - **Categories:** the list with usage per module (spending, payee rules, memory), create, rename, merge (pick sources and a target, show the counts that moved).
  - **Payee rules:** list, set (payee, category, description), delete.
  - **Connected assistants** (section 3).
- **Charts:** SVG and CSS, only design tokens (the `check:design` rule), black and white with the brand accent, readable on a phone, each with a text table for screen readers.
- **States:** loading, empty, error (the tool's own message), and a disabled button while a request runs. After a change the tab refetches; there are no optimistic updates.
- **Accessibility:** labelled controls, keyboard operation, focus returned after a dialog, sufficient contrast, tables for chart data.
- **Language:** all copy in `dashboard-copy.ts`; the Thai text is listed in the PR for a native speaker.

## 3. Connected assistants

The tab lists the assistants the user has connected and lets them disconnect one. An assistant connection is the pair (user, client) in `oauth_tokens`, so no migration is needed.

- `GET /api/web/connections` (session, same CSRF rules for non-GET) returns, per client with live tokens for the user: client name, the host it redirects to, the scopes, and when the connection was first made.
- `DELETE /api/web/connections/{client_id}` deletes that user's tokens for that client (access and refresh) and nothing else; another user's connection is untouched.
- Not covered: a "last used" time (not recorded today) and an absolute lifetime for a connection.
- This is new code in the sign-in area, so it gets its own PR and explicit owner approval.

## 4. Testing

- **Go (the standard-library packages, `-race`):** bridge tests for not signed in (401), missing header (403), foreign `Origin` (403), `Sec-Fetch-Site: cross-site` (403), wrong content type, wrong method (405), oversized body (413), unknown tool (404), a `ToolError` (422 with the text), another failure (500 with a generic message and the cause only in the log), user isolation (alice and bob through the bridge, both ways), the body cannot choose the user, and no arguments or results in log output. Connections: list and revoke are per user.
- **Browser (Playwright with faked API responses):** each tab in EN and TH on desktop and phone width: loading, empty, error, populated, charts, add, edit, delete confirm, merge, a refused input. Screenshots reviewed by eye. Written first and seen failing.
- **Gate checks:** `npx tsc --noEmit`, `npm run check:design`, `npm run build`, `gofmt`, `go vet`, `go test -race`, and CI (the only Go 1.26 compile of `internal/store` and `cmd/`).
- **Not verifiable here:** real Postgres through the bridge, and the real signed-in session. Both are exercised only by CI and after deploy.

## 5. Delivery

Five PRs, merged one at a time. Each has its own agent log. Approvals under `AGENTS.md` are named per PR.

| PR | Content | Needs owner approval |
|---|---|---|
| G | Bridge endpoint, tests, nginx rate limit | Yes: a new way to act as a signed-in user, and nginx |
| H | Dashboard shell, client library, copy EN and TH, Overview and Spending | No |
| I | Tasks and notes | No |
| J | Categories and Payee rules | No |
| K | Connected assistants (API and tab) | Yes: sign-in area |

The security fixes agreed with the owner (repository policy files, deployment defaults, service hardening, the OAuth error page) are separate PRs that go first.

## 6. Risks and open points

- A bug in the bridge would let one user reach another's data, so isolation is tested in both directions and the user id is never read from the request.
- A cross-site scripting bug in the web app would act with the user's full session. User text is only ever rendered as text; a Content-Security-Policy for the web app is decided in the security work, not here.
- The 200-record cap makes long histories incomplete on screen. The Spending tab says so and offers date filters; paging is a follow-up.
- Line breaks in notes are lost by the sanitiser (open owner decision).
