# Gemini could not connect: registration refused (bug fix)

- Agent: claude
- Date: 2026-10-31
- Branch / PR: fix/gemini-registration (PR link on GitHub)

## Goal
Owner: connecting Nemo works in Claude but not in Gemini. Gemini's Connect to an MCP server dialog showed "The Google redirect URL was rejected by the server. Enter your OAuth client ID and client secret to connect."

## Investigation (systematic debugging)
- Evidence: Gemini falls back to asking for a client ID and secret when dynamic client registration fails. I read every condition under which `POST /api/oauth/register` refuses a request. We log nothing about refusals, and request bodies are deliberately not logged, so the actual Gemini request could not be seen.
- Public reports: other open-source MCP servers hit the identical message and traced it to registration. Gemini registers like a server-side RFC 7591 client: it sends six redirect URIs on `oauth-redirect.googleusercontent.com` (Google's docs also say six) and asks for `client_secret_post` or `client_secret_basic`.
- Reproduction: tests built from that shape failed against the old code with `invalid_redirect_uri: provide between 1 and 5 redirect_uris`, and once that was fixed, with `invalid_client_metadata: only public clients ... are supported`. A third problem showed up one step later: the token endpoint read `client_id` only from the form, but a `client_secret_basic` client sends it in an `Authorization: Basic` header.

## Root causes and fixes (one at a time, each with a failing test first)
1. Redirect URI limit 5, Gemini sends 6: limit raised to 10.
2. Only `token_endpoint_auth_method: none` accepted: `client_secret_post` and `client_secret_basic` are accepted; the client is registered as a public client, no secret is issued, and the response says `none` (RFC 7591 section 3.2.1). Key-based methods (`private_key_jwt`, `tls_client_auth`, `client_secret_jwt`) are still refused.
3. Token endpoint ignored Basic client authentication: it now takes the client id from the form or from a Basic header (the password is ignored because no secret exists; the client is still held to PKCE and the code and refresh-token binding). A header and a form that name different clients, an unreadable header, or an unknown id give 401 `invalid_client`.
4. Diagnostics: refused registrations and refused authorize requests are logged with the reason, client id and origins (scheme and host), never state, PKCE challenge, redirect paths or query strings, or the client name. An unknown scope is named in its error text.

## Verified
- Scratch harness (`gotest.sh`, Go 1.24 directive, `-race`), now including the OAuth package: cerberus, hermes, themis, plutus, atlas all ok; vet clean; `gofmt -l apps/api` empty.
- New tests: Gemini-shaped registration for both secret methods, the refusals that remain, Basic and `client_secret_post` token and refresh exchanges, bad client authentication (6 cases), and the log contents for refused registrations and authorize requests (including that secrets are absent). Two existing registration tests were updated to the new limits (11 redirect URIs; `private_key_jwt`).

## Not verified / known gaps
- I could not test against real Gemini. The cause is established by reproducing the refusal with the publicly documented request shape, not by seeing Gemini's request. If Gemini still fails after deploy, the new log lines show the reason: `docker compose -f deploy/docker-compose.yml logs api | grep "rejected"`.
- Possible next hurdles I did not change because there is no evidence yet: an unauthenticated `HEAD /mcp` answers 405 (not 401); an authorize request with scopes this server does not offer is refused (now named in the message).
- Go 1.26 and `internal/store`/`cmd` compile are covered by CI only.

## Decisions and assumptions
- Changes the sign-in/OAuth area, so it needs the owner's approval under AGENTS.md; the owner asked for the fix in chat.
- Security effect: none weakened. Clients were already public (identified by id, protected by PKCE); the change only lets them say so in a different way and read the id from another place. The cap of 10 redirect URIs still requires each URI to be valid.
