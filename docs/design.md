# Nemosyne design (draft, pending approval)

Draft of decisions agreed so far in brainstorming. Not yet a final spec.

## Purpose

A hosted, open-source service (with a self-host option) that lets a user's AI assistant read and write their personal data across life areas. Spending, captured from bank slip photos, is the first module.

## Decisions so far

- Interface to assistants is an MCP server; the chat assistant is the primary UI.
- The user's own assistant reads slip photos with its own vision. It shows the extracted fields for confirmation, then calls an `add_transaction` tool. Images are discarded after extraction; only structured fields are stored.
- Hosted service plus self-host path (Docker Compose). Open core.
- Modular: a small core plus independent modules. Build the core and Plutus (spending) first; add Atlas (tasks) second and extract shared code only after two modules prove what is common.

## Architecture

```
Browser / AI assistant
        | HTTPS
   Propylaea (nginx)
     /      -> Olympus (Next.js)
     /api   -> Go API
     /mcp   -> Hermes (Go, streaming)
                 |
              Gaia (Postgres, internal network only)
```

- **Go backend:** Go MCP SDK over Streamable HTTP; `pgx` + `sqlc`; `goose` migrations. Each module implements a small interface registering its MCP tools, REST routes and migrations. Core never imports a module.
- **Olympus:** pure client of the Go API; no business logic or database access. Types come from a generated OpenAPI client.
- **Propylaea:** TLS termination, path routing, rate limiting, security headers. `/mcp` needs `proxy_buffering off` and long read timeouts.
- **Hades:** AES-GCM envelope encryption for sensitive fields; keys from env or KMS.

## Repo layout (planned)

```
apps/web                 Olympus (Next.js)
apps/api                 Go service
  internal/core          hermes, cerberus, hades
  internal/modules       plutus, atlas, horae, hephaestus
api/openapi.yaml
deploy/                  docker-compose.yml, nginx/
docs/
```

## Security principles

- Scoped, per-module permissions for assistants.
- Sanitize every free-text field returned to the assistant (prompt-injection defense).
- Per-user isolation, audit logs, full export and delete-everything.
- Mask account numbers; store only what reports need.

## Branding and web (Olympus)

- **Names:** Nemo (เนโม่) is the friendly name used in the UI and copy. Nemosyne remains the technical name: repository, MCP server name (`nemosyne`), cookies and environment variables do not change, so existing connectors keep working.
- **Logo:** the owner's illustration (a geometric dragonfly among clouds, black ink on transparent) is used as the hero artwork. The small mark used in the header and favicon is a vector simplification traced from it (`public/brand/nemo-mark.svg`). The source image is 256 px wide and the design system forbids rendering it larger than 256 px, so the hero shows it at 256 px; a vector original would remove that limit.
- **Theme and design system:** the web app follows the owner's "nemosyne" design system, light theme: white ground, black ink, hairlines instead of shadows, square 2px corners, one action colour (ink inverted) and no brand accent. Tokens live in `apps/web/app/tokens.css`; `apps/web/app/globals.css` uses only those tokens (`npm run check:design` in `apps/web` enforces no hex colours, shadows, gradients or large radii outside the allowed rules). Typefaces, self-hosted via `@fontsource` (no requests to Google Fonts): Barlow for display and controls, Geist for body, Geist Mono for labels. Prompt is kept for Thai because the other three have no Thai glyphs; Thai text is never uppercased or tracked. Text and badge colors meet WCAG AA contrast (checked from the live tokens).
- **Positioning:** Nemo is a memory for your AI assistant. Spending (Plutus) is the first module attached to it, not the product. Copy only describes what is built; modules that are not built are labeled "Coming soon".
- **Landing motion (landing page only):** a hero field reacts to the cursor (a hairline grid lit around the pointer, a crosshair tick that snaps to the nearest grid node, five small dragonfly sprites that wander and veer away from the pointer) and sections reveal on scroll, with a 1 px scroll-progress line under the header. This is a deliberate exception to the design system's flat-motion rule, requested by the owner. Everything is off under `prefers-reduced-motion`, sprites stay static without a fine pointer (touch), the loop pauses when the hero is off-screen or the tab is hidden, and content stays visible without JavaScript. Dashboard and Connect stay flat. No animation library is used.
- **Routes:** `/` English homepage, `/th` Thai homepage, `/dashboard` (moved from `/`), `/connect`. Sign-in now returns to `/dashboard`. The API still reports sign-in errors as `/?auth_error=...`; the homepage forwards that to `/dashboard`, so no backend change was needed.
- **Copy rules:** describe only what is built (modules not built are labeled "Coming soon"), no invented testimonials or user counts, and state the hosted-trust limit plainly (see Trust model in open decisions).
- **Known gaps:** the favicons and the social-preview image (`public/og.png`) still use the earlier art. The social-preview image (`public/og.png`) is only referenced when `NEXT_PUBLIC_SITE_URL` is set at build time, which the Dockerfile and compose file do not pass yet. Dashboard and Connect pages are English only. Hosted-instance copy such as pricing is not written.

## Auth (Cerberus): Google sign-in (decided)

Google is the only identity provider for now. Users never create a Nemosyne password.

- **Web sign-in (built first):** Authorization Code flow with PKCE and OIDC against Google. The callback verifies the ID token (signature, issuer, audience, nonce, `email_verified`), upserts the user keyed on Google's `sub` (not email), and creates a server-side session. The browser only holds an opaque, random session token in an `HttpOnly`, `SameSite=Lax`, `Secure` (outside dev) cookie. The database stores only the SHA-256 of that token, so a database leak does not yield usable sessions.
- **MCP clients (built):** Google cannot serve MCP clients directly, because they expect the server they connect to to be an OAuth authorization server. Cerberus therefore issues its own tokens and uses the Google session only to know who is approving. Implemented as stdlib-only code in `internal/core/cerberus/oauth*.go`:
  - Discovery: `/.well-known/oauth-protected-resource` (RFC 9728) and `/.well-known/oauth-authorization-server` (RFC 8414). `/mcp` answers 401 with `WWW-Authenticate: Bearer resource_metadata=...`.
  - Dynamic client registration (RFC 7591) at `/api/oauth/register`: public clients only (no client secret). Redirect URIs must be https, loopback http, or a custom app scheme, with no fragments or userinfo.
  - `/api/oauth/authorize`: PKCE with S256 is mandatory; the `redirect_uri` must exactly match a registered one (otherwise an error page, never a redirect); `resource` must be our MCP URL. Signed-out users are sent through Google sign-in and return to the same URL. The consent page shows the app's self-declared name (unverified, HTML-escaped), the redirect host and the requested scopes. The form carries an HMAC token bound to the user and exact parameters, and POSTs must be same-origin.
  - `/api/oauth/token`: authorization codes are single-use (5 minutes). Access tokens last 1 hour, refresh tokens 30 days and rotate on every use. Codes and tokens are stored only as SHA-256 hashes. Refresh cannot widen scope.
- **Scopes:** assistant tokens are scoped per module (for example `plutus:read`, `plutus:write`), never account-wide.
- **Consequences:** the user's Google account is the root of trust. Account recovery is Google's job. Deleting a Nemosyne account removes our data but does not touch the Google account.

## MCP endpoint (Hermes)

`POST /mcp`, MCP over Streamable HTTP, stateless JSON responses, tools only. `GET /mcp` returns 405 (no server-initiated stream). Supported protocol versions are negotiated (2025-11-25, 2025-06-18, 2025-03-26, 2024-11-05). Batch requests are rejected.

- The protocol layer (`internal/core/hermes`) is hand-written on the standard library: initialize, ping, tools/list, tools/call. This keeps `go.mod` unchanged and the code testable without downloading modules. Revisit the official Go SDK later; tools are defined behind `hermes.Tool`, so swapping the transport would not touch modules.
- Every request needs a valid bearer token. The caller's user id comes only from the verified token. `tools/list` shows only tools the token's scopes allow; calling others returns an error result.
- Tool failures: `hermes.ToolError` messages reach the assistant; anything else is logged and returned as a generic message. Panics are contained. Logs record user, client, tool and duration, never arguments or results.
- User-supplied text is cleaned with `hermes.Untrusted` (control and invisible characters removed, length capped) and tool output carries a note that those fields are data, not instructions. This reduces prompt-injection risk; it cannot eliminate it.
- CORS is open (`*`) on discovery, registration, token and `/mcp` because they use bearer tokens, never cookies.

## Plutus (spending)

Tools: `add_transaction` (`plutus:write`), `delete_transaction` (`plutus:write`, destructive), `list_transactions` and `get_spending_summary` (`plutus:read`). The tool descriptions instruct the assistant to show what it read from a slip and get the user's confirmation before saving.

- Amounts are stored as hundredths of the currency unit (`bigint`); input with more than two decimals is rejected rather than rounded. Totals are reported per currency and never mixed.
- Duplicates: a unique index on `(user_id, reference)` blocks saving the same slip twice. Without a reference, an identical amount, currency, date and merchant is flagged and only saved if the user confirms (`allow_duplicate`).
- Slip images are read by the user's own assistant and are never sent to or stored by Nemosyne.

## Build status

- Live: hosted with working Google sign-in; dependencies locked (`go.sum`); CI, Dependabot, PR/issue templates and agent rules (`AGENTS.md`) in place.
- Built, not yet used with a real assistant: OAuth server, MCP endpoint, Plutus tools, `/connect` page.
- Verified: unit and integration tests (stdlib), SQL and migrations against a real Postgres (up, down, up), browser smoke test of the web pages. The Go-to-Postgres driver wiring was only compile-checked against stubs and is covered by CI's build.
- Known gaps:
  - No "connected apps" page to review or revoke an assistant's access yet; there is no refresh-token reuse detection.
  - Registered OAuth clients are never expired.
  - Merchant and note are stored in plaintext. Hades (field encryption) is not built.
  - Hosted assistants share an IP at the nginx rate limiter; limits are generous but unmeasured.

## Open decisions

- **First target users:** technical early adopters, a specific community, or the general public.
- **MCP transport:** keep the hand-written stdlib layer or move to the official Go SDK.
- **Trust model wording:** how to describe hosted encryption honestly (see the privacy discussion; a Trust model section is still to be written).
- ~~License~~: MIT (decided).

## Mission

Spending is where Nemo starts. The mission is to make Nemo a good memory for more kinds of data you want your assistant to keep: tasks (Atlas), routines (Horae) and work data (Hephaestus), and to get better at organising what it already stores (for example, grouping spending by who you paid). These are goals, not release dates. A module is described as available only when it works. The landing page has a matching "Mission" section.
