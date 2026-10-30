# Self-host local admin page (Go): design

Status: **draft for owner review**. Date: 2026-10-30. Branch: `docs/selfhost-local-admin-page`. No code yet.

Builds on the self-host scope (`feat/self-host-mode`, "Self-host mode: scope"). Replaces the Next.js frontend **in self-host mode only** with a small page served by the Go binary.

## 1. Goal

Give the single owner a simple page, served by the same Go binary, that:

1. shows the current configuration (read-only), and
2. lets them import a backup of their data with one button.

Success: on a self-hosted install, the owner opens `http://127.0.0.1:8082`, sees mode, URL, key and record status, picks a `backup.json`, and gets a count of what was added and skipped. Nothing on this page is reachable through the tunnel.

## 2. Decisions (owner, 2026-10-30)

| Topic | Decision |
|---|---|
| Frontend | Go, server-rendered, one small page. No Node build, no separate container. |
| Access | **Local-only listener.** Separate port on `127.0.0.1`, not routed through the tunnel, no login. |
| Import format | Versioned Nemosyne JSON. |
| Import semantics | **Merge.** Adds records, skips ones that already exist, deletes nothing. |
| Existing web app | `apps/web` stays in the repo for hosted mode and is not touched here (assumption, see section 10). |

## 3. Out of scope

- Any change to hosted mode, Cerberus, the OAuth server or `apps/web`.
- Creating, rotating or revoking connector keys from the page (stays in the CLI).
- A "replace all" restore, CSV import, or an export button (export is a CLI command only).
- Backup encryption (scope item 5, later).
- Editing configuration from the page. It is display only.

## 4. Architecture

- New package `apps/api/internal/core/olympus`: handlers, `html/template` templates and one embedded stylesheet (`embed`). Standard library only, so no new dependency.
- Started from `main.go` only when `MODE=selfhost`, as a **second `http.Server`** with its own timeouts. Address from `ADMIN_ADDR`, default `127.0.0.1:8082`.
- The public listener (`/mcp/<key>`) registers no UI routes. The admin listener registers no MCP routes. They share the store and nothing else.
- Routes on the admin listener:
  - `GET /` config view
  - `POST /import` multipart upload, then the result view
  - anything else returns `404`
- Depends on self-host core (scope Phase 1) for the owner `users` row and `connector_keys`. This work is its own PR after Phase 1.

## 5. Config view (read-only)

Shows only non-secret facts:

- mode, version, `PUBLIC_URL`
- database status (ping) and transaction count
- connector keys: name, scopes, created, last used, revoked
- the connector URL with the key **masked** (`/mcp/nmk_…`). The full key is stored only as a hash and cannot be shown. The banner and `nemosyne url` remain the places it appears.

Never shown: `HADES_KEY`, the database password, key hashes, any secret. All values are rendered through `html/template` auto-escaping.

## 6. Backup export and import

### 6.1 File format

```json
{
  "format": "nemosyne-backup",
  "version": 1,
  "exported_at": "2026-10-30T12:00:00Z",
  "transactions": [
    { "amount_minor": 12900, "currency": "THB", "merchant": "...", "category": "...",
      "occurred_on": "2026-10-29", "note": "", "reference": "" }
  ]
}
```

- No user id and no database ids. The file is portable across installs.
- Unknown `format` or a `version` newer than the binary supports is rejected with a clear message.

### 6.2 Export (CLI)

`nemosyne export > backup.json` reads the owner's transactions and writes the format above. This is a small addition to the scope doc, which lists export as "later". It is plain JSON; encryption stays a later item. It runs inside the container like the other CLI commands (`docker compose exec api nemosyne export`).

### 6.3 Import (page)

- The owner's user id comes from the server (the single owner row), never from the file.
- Upload size is capped (default 10 MB) with `http.MaxBytesReader`; the JSON decoder rejects unknown fields.
- Each row is validated with the same rules Plutus applies to `add_transaction` (positive amount, three-letter currency, valid date, field length limits). **One invalid row rejects the whole file** and nothing is written.
- Valid rows are written in **one database transaction**. A row is skipped, not inserted, when:
  - its `reference` already exists for the owner (existing unique index), or
  - a record with the same amount, currency, date and merchant (case-insensitive) already exists.
- The result page shows counts only: added, skipped, rejected (if any), plus a row number and reason for a rejected row. It **never echoes file text** (merchant, note) back into the page or logs.
- Implementation note: the current store methods use the pool, not a transaction. Import needs a store method that takes a transaction. This is a code addition, not a migration.

## 7. Security

A page on `127.0.0.1` can still be attacked by any site open in the owner's browser (cross-site form posts, DNS rebinding). Controls, each with a test:

| Control | Why |
|---|---|
| `Host` header must be `localhost` or `127.0.0.1` with the admin port, else `404` | Stops DNS rebinding |
| `POST /import` requires a per-process CSRF token (random at start, in a hidden field) | Stops cross-site posts |
| `POST /import` requires `Origin` (or `Referer`) to match the admin origin | Second CSRF layer |
| Response headers: `Cache-Control: no-store`, `Content-Security-Policy: default-src 'none'; style-src 'self'; form-action 'self'`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, `X-Content-Type-Options: nosniff` | Hardening |
| No cookies, no sessions | Nothing to steal |
| Logs: method, path, status only. Never file contents or transaction data | `AGENTS.md` rule |
| Admin listener is never added to any tunnel config; docs say so | Keeps the page off the internet |

Other `AGENTS.md` rules still hold: queries filter by the owner's id, user text is never treated as instructions, and security behaviour gets tests.

## 8. Docker and the scope change to approve

A loopback listener inside a container is not reachable from the host. So in self-host compose:

- the app binds `0.0.0.0:8082` **inside the container** (set via `ADMIN_ADDR`), and
- compose publishes `127.0.0.1:${ADMIN_PORT:-8082}:8082` (host loopback only).

This is an **exception to scope section 6 ("no ports are published")**: loopback only, never `0.0.0.0` on the host. Tunnel profiles point only at `api:8081`.

Residual risk: other containers on the compose network could reach `api:8082`. The Host check and CSRF token still apply. Keep the tunnel container on a network that is the only one allowed to reach 8081 if this matters to the owner.

## 9. Testing and verification

Go tests (`go test -race -count=1 ./...`), written first:

- Config view: no secret appears in the rendered page; masked URL; revoked key shown as revoked.
- Import: valid file merges; re-import adds zero; bad row rejects the file and writes nothing; oversize file rejected; unknown version rejected; user id in the file is ignored.
- Security: wrong `Host` gives `404`; missing or wrong CSRF token gives `403`; cross-origin `Origin` gives `403`; headers present; logs contain no transaction text.
- Round trip: `export` then `import` on an empty database reproduces the same records.

Manual: run compose, open the page from the host, import a file, confirm the page is not reachable via the tunnel hostname. `gofmt -l .`, `go vet ./...`, `docker compose config -q` with dummy env. The PR states exactly what was and was not run.

## 10. Open items and approvals

- **`apps/web`:** assumed to stay for hosted mode and be left alone here. Removing it is a separate decision for the owner.
- **Needs owner approval per `AGENTS.md`:** compose change (loopback publish), `.env.example` (`ADMIN_ADDR`, `ADMIN_PORT`). No migrations, no new dependencies, nothing in Cerberus.
- **Needs a doc update:** the self-host scope doc (section 6 "no ports published" becomes "no public ports"; `nemosyne export` moves from "later" to this work) and `docs/design.md`.
- **Order:** after self-host core (Phase 1). One concern per PR, plus an agent log.
