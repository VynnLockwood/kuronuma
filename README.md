# Nemosyne

> Memory for your AI assistant. Named for the Greek goddess of memory (classically spelled *Mnemosyne*), mother of the Muses.

Nemosyne is an open-source personal data hub exposed to AI assistants over [MCP](https://modelcontextprotocol.io). Users keep their spending, tasks, routines and work data in one place, and their own assistant reads and writes it through scoped tools.

**Status:** early. Google sign-in (Cerberus), the web shell (Olympus), and the Docker/nginx setup exist. MCP (Hermes) and the spending module (Plutus) are next. See [docs/design.md](docs/design.md).

## Getting started

Prerequisites: Docker, and Go 1.24+ once to generate `go.sum`.

```bash
# 1. Generate go.sum (needs network access to the Go module proxy) and commit it
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

### Develop without Docker

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
