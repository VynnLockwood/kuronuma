# Kuronuma

**Kuronuma (คุโรนุมะ)** is an open-source memory for AI assistants.

> Memory for your AI assistant.

Kuronuma keeps your spending, tasks, to-dos and notes in one place that you own, and lets your own AI assistant (Claude and others that speak [MCP](https://modelcontextprotocol.io)) read and write them through a small set of scoped tools. You sign in with Google, you approve each assistant, and the assistant can only do what you allowed.

Source: [github.com/VynnLockwood/kuronuma](https://github.com/VynnLockwood/kuronuma) · MIT licence · self-hostable with Docker Compose.

![Architecture diagram. Your browser and your AI assistant reach a reverse proxy over HTTPS. The proxy sends page requests to the web dashboard, sign-in requests to the sign-in and OAuth service (which talks to Google), and tool requests to the MCP endpoint. The MCP endpoint calls the memory modules, which read and write a private PostgreSQL database. The sign-in service also stores users, sessions and tokens in that database.](docs/assets/architecture.png)

## Status

Early, with little real-world use yet. In CI the API has automated tests; the web app is type-checked and built, and the deployment files are config-checked (nothing there starts the full stack).

### Working now

- **Sign-in with Google**, with an optional list of allowed email addresses.
- **Web dashboard in English and Thai:** an overview with totals, a spending page (add, filter and delete records) and a tasks and notes page (add, edit, complete and delete).
- **MCP endpoint with 15 tools:** spending (add, list, totals, delete), tasks, to-dos and notes (add, list, update, delete), categories shared by everything (list, create, rename, merge) and payee rules that file new spending under a category automatically (set, list, delete).
- **Assistants sign in with OAuth.** A consent page shows what each assistant asks for; tokens are scoped, and every query is limited to the signed-in user.
- **A Kuronuma voice:** connected assistants are asked to sound quiet, sincere and brief when they talk about what Kuronuma keeps.
- **Self-hosting:** Docker Compose runs the web app, the API, PostgreSQL and an nginx front door with rate limits.

### In progress

- Dashboard pages for categories and payee rules.
- A dashboard page that shows which assistants are connected, with a way to disconnect them.

### Planned

- Encryption of sensitive fields.
- Routines and work data as new kinds of memory.

## Getting started

You need Docker.

```bash
# 1. Configure
cp .env.example .env
#    set POSTGRES_PASSWORD, SESSION_SECRET (openssl rand -base64 48),
#    GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET, and ALLOWED_EMAILS=you@gmail.com

# 2. Run
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

### Web pages

| Route | What it is |
|---|---|
| `/` and `/th` | Public homepage in English and Thai |
| `/dashboard` and `/th/dashboard` | Sign in and see your memory (overview, spending, tasks and notes) |
| `/connect` | The MCP URL and setup steps for assistants |

Optional build-time settings for the web app: `NEXT_PUBLIC_REPO_URL` sets the "Source code" link in the footer, and `NEXT_PUBLIC_SITE_URL` (for example `https://kuronuma.example.com`) enables the social-preview image. The Docker build does not pass either yet.

### Connect an assistant

Once signed in, open **/connect** on your deployment for the MCP URL (`https://your-domain/mcp`) and setup steps.

- **Claude:** Settings, Connectors, add a custom connector, paste the URL, then sign in with Google and approve.
- **Claude Code:** `claude mcp add --transport http kuronuma https://your-domain/mcp`, then `/mcp` to sign in.

Remote assistants need a public HTTPS address; if you run Kuronuma on your own computer, only apps on that computer can reach it. A connector added before a tool existed keeps its old permissions, so re-add or re-authorize it to see new tools.

Kuronuma also tells the connected assistant how to sound (see "Voice" in [docs/design.md](docs/design.md)). Self-hosters who want a plain voice change the `VoiceGuide` text in the API's instructions file, which the design document points to.

### Develop without Docker

Enable the commit template once: `git config commit.template .gitmessage` (Conventional Commits; CI checks PR titles).

```bash
# API (needs a Postgres and the env vars from .env.example)
cd apps/api && go test ./... && go run ./cmd/kuronuma-api
# Web (proxies /api to http://localhost:8081)
cd apps/web && npm install && npm run dev
```

### CI

GitHub Actions (`.github/workflows/ci.yml`) runs on every PR: Go vet and race tests, a check that `go.mod`/`go.sum` are committed and tidy, the web typecheck and build, Docker image builds, and `docker compose config` and `nginx -t`. Dependabot proposes weekly dependency updates.

## How it is built

| Part | Choice |
|---|---|
| Web app | Next.js (TypeScript), Tailwind CSS |
| API | Go |
| Database | PostgreSQL |
| Front door | nginx |
| Packaging | Docker Compose |
| Protocol | MCP over Streamable HTTP, OAuth for remote assistants |

[docs/design.md](docs/design.md) has the full design, the decisions behind it, and the code name of each part.

## Agents and diagrams

AI agents working in this repository must read [AGENTS.md](AGENTS.md) first; each agent keeps a log in `agent-logs/<name>/`. Agent skills live in [skills/](skills/README.md); the architecture picture above was made with the `diagram-design` skill and its editable source is [docs/assets/architecture.html](docs/assets/architecture.html).

## Security

- Found a vulnerability? Please report it privately, not in a public issue. See [SECURITY.md](SECURITY.md).
- Running your own Kuronuma: set `ALLOWED_EMAILS` unless you want any verified Google account to be able to sign in, choose a strong `POSTGRES_PASSWORD` and `SESSION_SECRET`, terminate TLS in front of nginx, and keep `.env` out of git.
- Behind a CDN or load balancer every visitor shares that proxy's address, so configure `real_ip_header` and `set_real_ip_from` in `deploy/nginx/nginx.conf` first, or the per-IP rate limits will apply to everyone together.

## License

MIT. See [LICENSE](LICENSE).
