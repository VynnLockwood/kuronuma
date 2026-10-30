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

## Open decisions

- **Auth:** run an existing identity server as a container (recommended) vs. build OAuth in Go with a library.
- **First target users:** technical early adopters, a specific community, or the general public.
- **Go MCP SDK:** official SDK vs. `mcp-go`; confirm maturity before committing.
- **License.**
