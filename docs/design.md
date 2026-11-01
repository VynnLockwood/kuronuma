# Kuronuma design (draft, pending approval)

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

- **Names:** Kuronuma (คุโรนุมะ) is the name used in the UI, copy, repository (`kuronuma-mcp`), MCP server and cookies (`kuro_session`, `kuro_flow`). The database name, database user and Compose project name stay `nemosyne` so the data volume is not orphaned (see `deploy/docker-compose.yml`). <!-- [keep-nemosyne] -->
- **Logo:** the owner's illustration (a geometric dragonfly among clouds, black ink on transparent). The web app shows it at 28 px next to the wordmark in the header and footer, and at 80 px on the sign-in card and the closing call to action; the source image is 256 px, so it is never rendered larger than that. The favicon is `app/icon.png`.
- **Theme and design system:** the Blush look: a warm white page (`#fffafb`), near-black plum ink (`#2a2024`), one raspberry action colour (`#c23b69`), soft pink surfaces (`#fdf0f4`) and borders (`#efd3dd`), pill-shaped buttons, large-radius cards (20 to 32 px) and no gradients. Styling is Tailwind CSS v4 on shadcn/ui-style components (`apps/web/components/ui/`: Button, Card, Badge, Input, Tabs on Radix, Alert, Skeleton, and an Accordion on native `<details>`). Colour tokens, including the four motif colours (`--motif-*`), are CSS variables in the `:root` block of `apps/web/app/globals.css`. `npm run check:design` in `apps/web` enforces hex colours only in that block, no gradients, the pinned background and primary values, and WCAG contrast for every text pair (main ink 7:1, other text and button labels at least 4.5:1, large-only text and focus rings 3:1). The consent and error pages served by the API use the same colours as Go constants and a test checks their contrast. Typefaces, self-hosted via `@fontsource` (no requests to Google Fonts): Inter for all UI text, Geist Mono for URLs and code. Prompt follows Inter for Thai glyphs; Thai headings are never tracked tighter. Six small original SVG motifs (sparkle, sprout, cloud, letter, petal, star) decorate the landing page and are always `aria-hidden`, never focusable and never over text.
- **Positioning:** Kuronuma is a memory for your AI assistant. Spending (Plutus) is the first module attached to it, not the product. Copy only describes what is built; modules that are not built are labeled "Coming soon".
- **Motion (landing page and shared components):** the hero enters in a short stagger; motifs twinkle or drift slowly; sections fade up as they scroll into view; feature, module and step cards lift and show a sparkle on hover; the main buttons get a flat shimmer on hover (on every page that uses them); tab panels fade in; FAQ answers open with a smooth height change. Hover effects apply only on devices that can hover. Everything is off under `prefers-reduced-motion`, and content stays visible without JavaScript. Because gradients are banned, the link underline fades in by colour instead of sliding in. `@view-transition` cross-fades full page loads only; client-side `<Link>` navigation does not animate. No animation library is used.
- **Letter field:** about 14 faint paper letters (8 below 640 px) drift behind every page and ease away from a mouse pointer (`components/LetterField.tsx`, physics in `lib/letter-physics.ts`). Letters are drawn at 18 to 28% opacity: above 28%, muted text or a link crossed by a letter's outline would fall below 4.5:1. The layer is fixed, behind the content, `aria-hidden` and click-through; touch screens get drift only; under reduced motion the letters are drawn once and never move. It is the only continuous script animation: one `requestAnimationFrame` loop that writes only `transform` and stops while the tab is hidden. Low-end phones will feel it first; if it ever shows jank, lower the letter count in `LetterField.tsx`.
- **Routes:** `/` English homepage, `/th` Thai homepage, `/dashboard` (moved from `/`), `/connect`. Sign-in now returns to `/dashboard`. The API still reports sign-in errors as `/?auth_error=...`; the homepage forwards that to `/dashboard`, so no backend change was needed.
- **Copy rules:** describe only what is built (modules not built are labeled "Coming soon"), no invented testimonials or user counts, and state the hosted-trust limit plainly (see Trust model in open decisions).
- **Known gaps:** the favicons still use the earlier art. The social-preview image (`public/og.png`) is only referenced when `NEXT_PUBLIC_SITE_URL` is set at build time, which the Dockerfile and compose file do not pass yet. Dashboard and Connect pages are English only. Hosted-instance copy such as pricing is not written.

## Auth (Cerberus): Google sign-in (decided)

Google is the only identity provider for now. Users never create a Kuronuma password.

- **Web sign-in (built first):** Authorization Code flow with PKCE and OIDC against Google. The callback verifies the ID token (signature, issuer, audience, nonce, `email_verified`), upserts the user keyed on Google's `sub` (not email), and creates a server-side session. The browser only holds an opaque, random session token in an `HttpOnly`, `SameSite=Lax`, `Secure` (outside dev) cookie. The database stores only the SHA-256 of that token, so a database leak does not yield usable sessions.
- **MCP clients (built):** Google cannot serve MCP clients directly, because they expect the server they connect to to be an OAuth authorization server. Cerberus therefore issues its own tokens and uses the Google session only to know who is approving. Implemented as stdlib-only code in `internal/core/cerberus/oauth*.go`:
  - Discovery: `/.well-known/oauth-protected-resource` (RFC 9728) and `/.well-known/oauth-authorization-server` (RFC 8414). `/mcp` answers 401 with `WWW-Authenticate: Bearer resource_metadata=...`.
  - Dynamic client registration (RFC 7591) at `/api/oauth/register`: every client is registered as a public client and no secret is issued. A registration may ask for `client_secret_post` or `client_secret_basic` (server-side clients such as Gemini do); it is accepted and the response says `none`, as RFC 7591 section 3.2.1 allows. Key-based methods are refused. Up to 10 redirect URIs, each https, loopback http, or a custom app scheme, with no fragments or userinfo. The token endpoint reads the client id from the form or from an HTTP Basic header and ignores any password. Refused registrations and authorize requests are logged with the reason (never the state, challenge, redirect paths or client name).
  - `/api/oauth/authorize`: PKCE with S256 is mandatory; the `redirect_uri` must exactly match a registered one (otherwise an error page, never a redirect); `resource` must be our MCP URL. Signed-out users are sent through Google sign-in and return to the same URL. The consent page (English or Thai: `?lang=`, else the browser language) asks "Connect <app> to your Kuronuma?", shows the app's self-declared name (unverified, HTML-escaped), the redirect host, a short "check before you approve" checklist, and the requested scopes grouped by area in plain words. The user can approve "Read and change" (everything requested) or "Read only" (only the `:read` scopes of the request; the server derives the granted scopes from the checked request, never from the form, and refuses read-only for a request with no read scope). Its security policy allows no scripts; the signed-in email is wrapped in `email_off` markers so Cloudflare's email obfuscation cannot replace it with a link that needs a script. The form carries an HMAC token bound to the user and exact parameters, and POSTs must be same-origin.
  - `/api/oauth/token`: authorization codes are single-use (5 minutes). Access tokens last 1 hour, refresh tokens 30 days and rotate on every use. Codes and tokens are stored only as SHA-256 hashes. Refresh cannot widen scope.
- **Scopes:** assistant tokens are scoped per module (for example `plutus:read`, `plutus:write`), never account-wide.
- **Consequences:** the user's Google account is the root of trust. Account recovery is Google's job. Deleting a Kuronuma account removes our data but does not touch the Google account.

## MCP endpoint (Hermes)

`POST /mcp`, MCP over Streamable HTTP, stateless JSON responses, tools only. `GET /mcp` returns 405 (no server-initiated stream). Supported protocol versions are negotiated (2025-11-25, 2025-06-18, 2025-03-26, 2024-11-05). Batch requests are rejected.

- The protocol layer (`internal/core/hermes`) is hand-written on the standard library: initialize, ping, tools/list, tools/call. This keeps `go.mod` unchanged and the code testable without downloading modules. Revisit the official Go SDK later; tools are defined behind `hermes.Tool`, so swapping the transport would not touch modules.
- Every request needs a valid bearer token. The caller's user id comes only from the verified token. `tools/list` shows only tools the token's scopes allow; calling others returns an error result.
- Tool failures: `hermes.ToolError` messages reach the assistant; anything else is logged and returned as a generic message. Panics are contained. Logs record user, client, tool and duration, never arguments or results.
- User-supplied text is cleaned with `hermes.Untrusted` (control and invisible characters removed, length capped) and tool output carries a note that those fields are data, not instructions. This reduces prompt-injection risk; it cannot eliminate it.
- CORS is open (`*`) on discovery, registration, token and `/mcp` because they use bearer tokens, never cookies.

## Plutus (spending)

Tools: `add_transaction` (`plutus:write`), `delete_transaction` (`plutus:write`, destructive), `list_transactions` and `get_spending_summary` (`plutus:read`). The tool descriptions instruct the assistant to show what it read from a slip and get the user's confirmation before saving. A category the user does not have yet is created when the record is saved (see Themis).

- Amounts are stored as hundredths of the currency unit (`bigint`); input with more than two decimals is rejected rather than rounded. Totals are reported per currency and never mixed.
- Duplicates: a unique index on `(user_id, reference)` blocks saving the same slip twice. Without a reference, an identical amount, currency, date and merchant is flagged and only saved if the user confirms (`allow_duplicate`).
- Slip images are read by the user's own assistant and are never sent to or stored by Kuronuma.
- Payee rules: the user can teach Kuronuma what a payee is (`set_payee_rule`, `list_payee_rules`, `delete_payee_rule`, using the plutus scopes). A rule applies when the normalised merchant contains the rule's payee; the longest match wins, then the most recently updated. `add_transaction` resolves its category as the one given (always wins, even an explicit `uncategorized`), else the matching rule, else `uncategorized`, and reports `category_source` and `matched_payee`. Rules only affect new transactions, are limited to 500 per user, and follow their category when it is renamed or merged (Themis).

## Themis (categories)

One category list per user, shared by every module. Records refer to a category by name, and `uncategorized` is always valid. Saving spending, a task, a to-do or a note with a name the user does not have yet creates that category on the spot (`themis.Ensure`) and the result says `category_created`, so an assistant can file something under a new category without a separate step. Assistants are told to reuse an existing category when one fits. A user can have at most 100 categories; past that a new name is refused with a message to pick an existing one or merge some. `set_payee_rule` is the exception: its category must already be in the list, and an unknown one is refused with the user's list in the message.

Tools: `list_categories` (`categories:read`), `create_category` and `rename_category` (`categories:write`), `merge_categories` (`categories:write`, destructive). Rename and merge update every record that uses the name in one transaction. A name that is in use but missing from the list shows as `unlisted` so it can be merged.

- Migration 00004 builds each user's starting list from the categories they already use.
- Existing connectors keep the scopes they were granted; the user re-authorizes to see the new tools. `add_transaction` still works with `plutus:write` alone.
- Next: payee rules (Plutus) and flexible memory items (Atlas) build on this list. See `docs/superpowers/specs/2026-10-31-memory-and-categories-design.md`.

## Atlas (tasks, to-dos and notes)

One flexible memory type for anything the user asks the assistant to remember. Each item has a kind (`task` for work items, `todo` for personal checklist items, `note` for facts, decisions and preferences), a status (`open`, `done`, `archived`), a priority, an optional due date, a category from the shared list (see Themis) and up to 10 tags. Completing an item sets `completed_at`; reopening clears it.

Tools: `list_memories` (`atlas:read`), `add_memory` and `update_memory` (`atlas:write`), `delete_memory` (`atlas:write`, destructive). `list_memories` takes filters (kind, status, category, tag, inclusive `due_before` and `due_after`, limit up to 200) and a text query: up to 5 words, each of which must appear as a substring of the title, details or a tag, ignoring case. Search is substring matching rather than Postgres full-text search because full-text search does not split Thai words; `%`, `_` and `\` in a query match themselves. Order: open items first, then due date (undated last), priority, newest.

- Migration 00006. New scopes `atlas:read` and `atlas:write`; existing connectors must be re-authorized to see the tools.
- No duplicate detection: the tool descriptions tell the assistant to list before adding.
- Not built yet: a web screen for browsing memories, reminders, recurring items (Horae). The landing page, `/th` and the dashboard show Atlas as Available.

## Build status

- Live: hosted with working Google sign-in; dependencies locked (`go.sum`); CI, Dependabot, PR/issue templates and agent rules (`AGENTS.md`) in place.
- Built, not yet used with a real assistant: OAuth server, MCP endpoint, Plutus tools (including payee rules), Themis category tools, Atlas memory tools, `/connect` page.
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

Spending is where Kuronuma starts. The mission is to make Kuronuma a good memory for more kinds of data you want your assistant to keep: tasks (Atlas), routines (Horae) and work data (Hephaestus), and to get better at organising what it already stores (for example, grouping spending by who you paid). These are goals, not release dates. A module is described as available only when it works. The landing page has a matching "Mission" section.
