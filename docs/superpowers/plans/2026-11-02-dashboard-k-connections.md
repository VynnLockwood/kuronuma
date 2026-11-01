# Dashboard PR K: Connected assistants Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline) or superpowers:subagent-driven-development. Steps use checkbox (`- [ ]`) syntax.

**Goal:** A Connected assistants tab where the user sees which assistants can reach their data and disconnects one. This is new code in the sign-in area, so it has its own PR and explicit owner approval.

**Spec:** `docs/superpowers/specs/2026-10-31-web-dashboard-design.md` (section 3). PR K of five; needs G (the `webguard` package and the dashboard shell from H).

**Architecture:** An assistant connection is the pair (user, client) in `oauth_tokens`, so there is no migration. Two handlers in a new file `apps/api/internal/core/cerberus/connections.go`, registered by `OAuth.Register`: `GET /api/web/connections` and `DELETE /api/web/connections/{client_id}`. Both resolve the user from the session cookie (`UserFromRequest`); the DELETE also runs `webguard.Check` (the GET is a read of the user's own data and needs no custom header, but still refuses a foreign `Origin` or a cross-site `Sec-Fetch-Site`; decide and record: running the full check on both is simpler and is the default). `OAuthStore` gains `ListConnections(ctx, userID, now)` and `DeleteConnection(ctx, userID, clientID)`; the Postgres implementation is two queries on `oauth_tokens` joined to `oauth_clients`.

## Global Constraints

- Every query filters by the session user's id; the client id from the path only narrows it. Deleting another user's connection, or a client the user has no tokens for, changes nothing and answers the same as success with `revoked: 0` (do not reveal existence).
- List response per connection: `client_id`, `name` (the client's self-declared name, sanitised with `hermes.Untrusted`-style cleaning already used for the consent page), `redirect` (host for web URLs, scheme for app URLs, via the existing `redirectLabel`), `scopes` (union across the live tokens), `connected_at` (earliest token `created_at` among the user's tokens for that client). Only clients with live (unexpired) tokens appear. Sorted by `connected_at`, newest first. `Cache-Control: no-store`.
- Revoke deletes the user's access and refresh tokens for that client and nothing else (no client, no codes, no other user's rows).
- Not covered: a last-used time (not recorded), an absolute connection lifetime.
- Logging: client id and user id and the count revoked; never tokens or hashes.
- Web: the tab calls these endpoints through `lib/web.ts` (extend it with `listConnections()` and `revokeConnection(id)` using the same header, result handling and error mapping as `callTool`). Scope names are shown as the plain-words areas the consent page uses (Spending, Categories, Tasks, to-dos and notes; read or change), translated EN and TH; unknown scopes show their raw name. Disconnect uses the confirm dialog, explains that the assistant will have to be connected again, and refetches.
- Owner approval needed (AGENTS.md): sign-in area (`cerberus`, `OAuthStore`, the store). Approving this plan gives it. Branch `feat/dashboard-connections`. Commits as the owner, no AI trailers.

## Review Focus

- Isolation both ways: a user can list and revoke only their own connections; a guessed client id of someone else's connection revokes nothing and looks the same as a miss.
- The revoke really ends access: after it, the access token and the refresh token for that client fail at the MCP endpoint and the token endpoint; another client of the same user keeps working.
- CSRF: the DELETE refuses a request without the header, with a foreign `Origin`, or cross-site; the list and revoke need a session.
- No migration; the queries use existing columns and indexes; nothing is logged that identifies a token.
- The names and redirect labels are text only in the UI; a hostile client name renders literally.
- The tab handles zero connections (explains how to connect, links to `/connect`), an error, and a revoke that races with expiry.

---

### Task 1: Store and handlers

**Files:** Modify `apps/api/internal/core/cerberus/oauth_types.go` (interface), create `connections.go`, `connections_test.go`; modify `oauth.go` (`Register`), `apps/api/internal/store/oauth.go`, the in-memory store in the cerberus tests.

- [ ] **Step 1: Write the failing tests** (standard-library harness): GET with no session is 401; a user with two clients with live tokens gets two entries with name, redirect label, scopes and `connected_at`; expired-only clients are absent; another user's connection is absent; the list is sorted newest first; DELETE with no session is 401, with no header or a foreign `Origin` or `Sec-Fetch-Site: cross-site` is 403; DELETE of the user's own client deletes both tokens and `{revoked: n}`; after it, `GetToken`/the token endpoint refuse both tokens while a second client of the same user still works; DELETE of another user's client or an unknown id changes nothing and answers `revoked: 0`; a hostile client name comes back sanitised; responses are `no-store`; logs contain ids and counts but no token material.
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement** (interface methods, in-memory and Postgres versions, handlers, registration).
- [ ] **Step 4: Verify.** `gotest.sh`, vet, gofmt; the real store compiles in CI only (say so); extend `gotest.sh` copy list if a new package is touched.
- [ ] **Step 5: Commit** `feat(cerberus): list and revoke a user's connected assistants`.

### Task 2: The tab

**Files:** Create `components/dashboard/Connections.tsx`; extend `lib/web.ts`, copy, the shell registry. Tests: `/home/claude/scratch/dash-connections-check.py`.

- [ ] **Step 1: Write the failing test** (faked API, EN and TH, 1280 and 390): the list shows name, host, areas in plain words, and the connected date; zero connections shows the empty state with a link to `/connect`; disconnect asks first, calls the DELETE with the header, refetches, and the entry is gone; an error shows the message with retry; a hostile name renders literally; keyboard-only walk through disconnect; no horizontal scroll at 390.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Verify.** Test PASS, `tsc`, `check:design`, `build`, earlier scratch checks, screenshots reviewed.
- [ ] **Step 5: Commit** `feat(web): add the Connected assistants tab`.

### Task 3: Docs, review and PR

- [ ] `docs/design.md` (connections endpoints and the tab), Thai strings list, whole-branch review by a fresh Opus reviewer with this Review Focus (one fix pass, tests first), agent log, PR `feat(cerberus): add Connected assistants to the dashboard` with the template (check Auth / OAuth), the real CI result, "Rulings I made" and "Deferred minors".
