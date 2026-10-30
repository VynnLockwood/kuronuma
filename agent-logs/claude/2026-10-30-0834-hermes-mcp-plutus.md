# Hermes MCP endpoint, OAuth for assistants, Plutus tools, agent rules

- Agent: claude
- Date: 2026-10-30
- Branch / PR: `feat/hermes-mcp` (no PR opened yet)

## Goal
The owner asked for: (1) an AGENTS.md with per-agent log folders and rules so agents do not break things, (2) a PR/commit template like other open-source projects, (3) a working MCP connection to test with an assistant.

## What I did
- Agent rules: `AGENTS.md`, `CLAUDE.md` (imports AGENTS.md), `agent-logs/README.md`, PR and issue templates, `.gitmessage` commit template, a PR-title (Conventional Commits) check workflow, Dependabot commit prefixes.
- Cerberus OAuth server (`internal/core/cerberus/oauth*.go`): discovery metadata, dynamic client registration, authorize + consent page with CSRF token, token endpoint with PKCE S256, refresh rotation, migration `00002_oauth.sql`.
- Hermes MCP endpoint (`internal/core/hermes`): JSON-RPC over Streamable HTTP, scope-filtered `tools/list`, safe tool calls.
- Plutus (`internal/modules/plutus`): `add_transaction`, `list_transactions`, `get_spending_summary`, `delete_transaction`; migration `00003_transactions.sql`; Postgres store in `internal/store`.
- Wiring in `cmd/nemosyne-api/main.go`, nginx routes (`/mcp`, `/.well-known/`, `/api/oauth/`) with rate limits, `/connect` page in the web app, README and design doc updates.

## Verified
- `gofmt -l .`: no output.
- `go vet` and `go test -race -count=1` for config, hermes, cerberus, plutus (stdlib-only packages): all pass (77 tests including subtests).
- Mutation checks (temporarily breaking PKCE, redirect_uri validation, user scoping, scope filtering): the tests failed each time; code restored afterwards.
- Whole API (including store, google, main) built, vetted and tested in a scratch copy against stub versions of pgx, goose, oauth2 and go-oidc: compiles, tests pass.
- Migrations up, down, up again and every SQL statement used by the Go store, run against a real Postgres via psql with PREPARE (types inferred like pgx): behaved as expected, including the partial unique index, single-use codes, refresh rotation, cross-user delete returning 0 rows, constraints, and cascades.
- `npx tsc --noEmit`, `npm run build`, and a Playwright smoke test of the dashboard and `/connect` page against a stand-in API: pass.
- nginx config parsed with crossplane (no syntax errors); PR-title regex checked with sample titles.

## Not verified / known gaps
- The real `go.mod` requires Go 1.26 and this environment blocks the Go module proxy, so I could not run `go build` or `go test` against the real dependencies. The Go-to-pgx wiring was only checked against stubs; CI will build it for real.
- Not tested with a real assistant (Claude connector, Claude Code) or real Google sign-in.
- Docker and nginx were not run (daemon unavailable). `nginx -t` runs in CI.
- Registered OAuth clients never expire; no refresh-token reuse detection; no page to review or revoke connected apps.
- Merchant and note are stored in plaintext (Hades not built).

## Decisions and assumptions
- Hand-wrote the MCP and OAuth layers on the standard library instead of adding dependencies, so `go.mod` and `go.sum` are unchanged and everything is testable here. Revisit the official Go SDK later.
- This change touches areas that AGENTS.md says need owner approval (Cerberus, migrations, nginx, CI). The owner asked for the MCP work explicitly; it still needs their review before merge.
- Dependabot already opened PRs for Go 1.27, Node 25 and Actions major versions; I did not touch them.
- "agent.md" was created as `AGENTS.md`, the name most agents read automatically.
- Commits authored as the repo owner with no AI trailers, per the owner's instruction.

## Next steps
- Owner: review the PR, merge, redeploy, then connect an assistant at `/connect` and report anything that fails.
- Add a connected-apps page (list and revoke), refresh-token reuse detection, and client expiry.
- Hades field encryption and a written Trust model section.
