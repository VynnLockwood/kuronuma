# Nemosyne

> Memory for your AI assistant. Named for the Greek goddess of memory (classically spelled *Mnemosyne*), mother of the Muses.

Nemosyne is an open-source personal data hub exposed to AI assistants over [MCP](https://modelcontextprotocol.io). Users keep their spending, tasks, routines and work data in one place, and their own assistant reads and writes it through scoped tools.

**Status:** early. Working: Google sign-in (Cerberus), web shell (Olympus), Docker/nginx. Built and awaiting real-world testing: the MCP endpoint (Hermes), OAuth for assistants, and the spending tools (Plutus). See [docs/design.md](docs/design.md).

**AI agents working in this repo:** read [AGENTS.md](AGENTS.md) first. It has the rules, and each agent keeps a log in `agent-logs/<name>/`.

## Getting started

Prerequisites: Docker, and Go 1.26+ once to generate `go.sum`.

```bash
# 1. Generate go.sum with Go 1.26 and commit it (CI fails until it is committed)
cd apps/api && go mod tidy && cd ../..

# 2. Configure
cp .env.example .env
#    set POSTGRES_PASSWORD, SESSION_SECRET (openssl rand -base64 48),
#    GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET, and ALLOWED_EMAILS=you@gmail.com

# 3. Run
docker compose -f deploy/docker-compose.yml --env-file .env up --build
# open http://localhost:8080
```

### Google sign-in setup

1. In the [Google Cloud console](https://console.cloud.google.com/), create a project and configure the **OAuth consent screen** (External; add yourself as a test user while in Testing).
2. **Credentials → Create credentials → OAuth client ID → Web application.**
3. Add this **Authorized redirect URI** (it must match `PUBLIC_URL` exactly):
   `http://localhost:8080/api/auth/google/callback`
   For a hosted deployment also add `https://your-domain/api/auth/google/callback`.
4. Put the client ID and secret in `.env`. The secret stays server-side and is never sent to the browser.

Until a client is configured the app runs with sign-in disabled.

### Connect an assistant

Once signed in, open **/connect** on your deployment for the MCP URL (`https://your-domain/mcp`) and setup steps.

- **Claude:** Settings, Connectors, add a custom connector, paste the URL, then sign in with Google and approve.
- **Claude Code:** `claude mcp add --transport http nemosyne https://your-domain/mcp`, then `/mcp` to sign in.

Remote assistants need a public HTTPS address. Tools: `add_transaction`, `list_transactions`, `get_spending_summary`, `delete_transaction`. Scopes: `plutus:read`, `plutus:write`.

### Updating a deployment

```bash
git pull
docker compose -f deploy/docker-compose.yml --env-file .env up --build -d
docker compose -f deploy/docker-compose.yml --env-file .env restart nginx   # nginx.conf is a bind mount and is not reloaded automatically
```

`PUBLIC_URL` must be exactly the https origin you give the assistant (same host, no trailing slash).

### If connecting an assistant fails

Replace `YOUR-DOMAIN` and run these. Each should return what is shown:

```bash
curl -i https://YOUR-DOMAIN/.well-known/oauth-protected-resource     # 200, JSON with "resource": ".../mcp"
curl -i https://YOUR-DOMAIN/.well-known/oauth-authorization-server   # 200, JSON with "registration_endpoint"
curl -i -X POST https://YOUR-DOMAIN/mcp -H 'Content-Type: application/json' -d '{}'   # 401 with a WWW-Authenticate header
curl -i -X POST https://YOUR-DOMAIN/api/oauth/register -H 'Content-Type: application/json' \
  -d '{"redirect_uris":["https://claude.ai/api/mcp/auth_callback"]}'                  # 201 with a client_id
docker compose -f deploy/docker-compose.yml --env-file .env logs api | grep -E 'oauth|/mcp'
```

An HTML 404 page from the first two means nginx still has the old config (restart it). A 403 or a challenge page means a firewall or CDN in front is blocking the assistant's requests.

### CI

GitHub Actions (`.github/workflows/ci.yml`) runs on every PR: Go vet and race tests, a check that `go.mod`/`go.sum` are committed and tidy, the web typecheck and build, Docker image builds, and `docker compose config` and `nginx -t`. Dependabot proposes weekly dependency updates.

### Develop without Docker

Enable the commit template once: `git config commit.template .gitmessage` (Conventional Commits; CI checks PR titles).


```bash
# API (needs a Postgres and the env vars from .env.example)
cd apps/api && go test ./... && go run ./cmd/nemosyne-api
# Web (proxies /api to http://localhost:8081)
cd apps/web && npm install && npm run dev
```

## First use case

Snap a bank slip, attach it in your assistant's chat, confirm the extracted amount, payee and date, and Nemosyne stores it. Ask for reports any time.

## Stack

| Layer | Choice |
|---|---|
| Frontend | Next.js (TypeScript) |
| Backend | Go |
| Database | PostgreSQL |
| Edge | nginx |
| Packaging | Docker / Docker Compose |
| Protocol | MCP over Streamable HTTP, OAuth for remote clients |

## The pantheon

| Name | Role |
|---|---|
| **Olympus** | Next.js web dashboard |
| **Propylaea** | nginx edge: TLS and routing |
| **Hermes** | MCP gateway between assistants and data |
| **Cerberus** | Auth and OAuth guard |
| **Hades** | Encrypted vault: field-level encryption and keys |
| **Gaia** | PostgreSQL storage |
| **Pythia** | Slip intake: validation, dedupe, review queue |
| **Plutus** | Spending module (first module) |
| **Atlas** | Tasks module |
| **Horae** | Routines module |
| **Hephaestus** | Work data module |
| **Moirai** | Background jobs and scheduling (later) |

## License

MIT. See [LICENSE](LICENSE).
