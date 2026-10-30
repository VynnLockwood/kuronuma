# Self-host Local Admin Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A Go-served, localhost-only page that shows read-only config and imports a Nemosyne JSON backup (merge), plus a `nemosyne export` command that produces that file.

**Architecture:** Tasks 1-5 are self-contained Go code tested with in-memory fakes: shared validation and backup logic in `plutus`, and a new `olympus` HTTP package. Tasks 6-7 wire it into `main.go`, the CLI, compose and docs, and depend on the self-host core PR (Phase 1). The page runs on its own `http.Server`; the public listener never serves it.

**Tech Stack:** Go standard library only (`net/http`, `html/template`, `embed`, `encoding/json`), existing `pgx`. No new dependencies, no migrations.

**Spec:** `docs/superpowers/specs/2026-10-30-selfhost-local-admin-page-design.md`

## Global Constraints

- Admin listener default `ADMIN_ADDR=127.0.0.1:8082`; started only when `MODE=selfhost`; hosted mode, Cerberus and `apps/web` are not touched.
- Backup file: `{"format":"nemosyne-backup","version":1,"exported_at":...,"transactions":[...]}`; no user id, no database ids; unknown `format` or `version` > 1 rejected.
- Upload cap default 10 MB (`http.MaxBytesReader`); JSON decoder rejects unknown fields.
- One invalid row rejects the whole file; valid rows written in one DB transaction; merge only, nothing deleted.
- Skip rule (as implemented, see spec 6.3): a row with a reference is skipped only if that reference exists or repeats in the file; a row without one is matched on amount + currency + date + lower(merchant) with multiset counting (adds `max(0, N - existing)`).
- Owner user id comes from the server, never from the file. Result page shows counts, row numbers and static reasons only; never file text. Logs: method, path, status only.
- Page security: `Host` must be `localhost` or `127.0.0.1` with the admin port (else `404`); `POST /import` needs the per-process CSRF token (else `403`) and matching `Origin`/`Referer` (else `403`); headers `Cache-Control: no-store`, `Content-Security-Policy: default-src 'none'; style-src 'self'; form-action 'self'`, `X-Frame-Options: DENY`, `Referrer-Policy: same-origin`, `X-Content-Type-Options: nosniff`; no cookies.
- Never rendered: `HADES_KEY`, DB password, key hashes, full connector keys.
- Verify per `AGENTS.md`: `gofmt -l .` empty, `go vet ./...`, `go test -race -count=1 ./...` from `apps/api`. No test may depend on a database in CI (CI has none).
- Commits: Conventional Commits, no AI `Co-Authored-By` trailer, owner's git identity. Never push to `main`.

## Review Focus

- Empty or minimal file (`{}` with valid header, `"transactions": []`): accepted, 0 added, no crash. Pinned in Task 2.
- Wrong kind of file (CSV, PNG, truncated JSON, UTF-8 BOM prefix): clear error, nothing written. Pinned in Task 2.
- Same row twice inside one file: with a reference the second is skipped; without one both are kept (two real identical payments) and a re-import adds nothing. Pinned in Task 3.
- HTML or script in `merchant`/`note`: stored sanitised, never rendered into any page. Pinned in Task 5.
- Double-click or re-upload of the same file: second import adds 0. Pinned in Task 3.
- `Host: LOCALHOST:8082` (case) and a wrong port: case-insensitive match accepted, wrong port `404`. Pinned in Task 4.

## Prerequisites (gate, not tasks)

- **Go 1.26 toolchain.** `go.mod` says `go 1.26.0`; this sandbox has 1.24.7 and the toolchain download is blocked. Tasks 1-5 must run where Go 1.26 is available (owner machine or CI). If not, say so in the PR; do not claim tests passed.
- **Self-host core merged (Phase 1)** before Tasks 6-7: it provides `MODE=selfhost`, the owner `users` row, `connector_keys`, and the `nemosyne` CLI dispatcher.
- **Owner approval** before Task 7: `deploy/docker-compose.yml` (loopback publish) and `.env.example`.
- Work on branch `feat/selfhost-local-admin-page` from latest `main`.

---

### Task 1: Share the transaction validation rules

Import must apply exactly the rules `add_transaction` applies. Extract them; no behaviour change.

**Files:**
- Modify: `apps/api/internal/modules/plutus/plutus.go` (the checks inside `add`, lines ~252-300)
- Test: `apps/api/internal/modules/plutus/plutus_test.go`

**Interfaces:**
- Produces: `func (m *Module) buildTransaction(minor int64, currency, merchant, occurredOn, category, note, reference string) (NewTransaction, error)` and `const maxMinor = 999_999_999_999` (largest value `parseMinor` can return). Errors are `hermes.Errorf` values with the existing messages.

- [ ] **Step 1: Write failing tests** `TestBuildTransaction` (table): minor `0` and `-5` and `maxMinor+1` rejected; currency `thb` is upper-cased to `THB`, `TH` and `THBB` rejected; merchant `"  "` rejected; merchant of 150 runes truncated to 100; date `"2026-13-01"` rejected; date 3 days after `m.now()` rejected, 1 day after accepted; date `1989-12-31` rejected; empty category becomes `uncategorized`; category `"Food"` becomes `food`.
- [ ] **Step 2: Run** `go test ./internal/modules/plutus -run TestBuildTransaction -v` from `apps/api`. Expected: FAIL, `buildTransaction` undefined.
- [ ] **Step 3: Implement `buildTransaction`** by moving the currency, merchant, date, future/old-date, category and `Untrusted` field handling out of `add`; `add` now calls `parseMinor` then `buildTransaction`.
- [ ] **Step 4: Run** `go test -race -count=1 ./internal/modules/plutus`. Expected: PASS, including every pre-existing `add` test unchanged.
- [ ] **Step 5: Commit** `refactor(plutus): share transaction validation`.

### Task 2: Backup file format, export and parsing

**Files:**
- Create: `apps/api/internal/modules/plutus/backup.go`
- Test: `apps/api/internal/modules/plutus/backup_test.go`
- Modify: `apps/api/internal/modules/plutus/plutus.go` (`Store` interface gains `ExportTransactions`)
- Modify: `apps/api/internal/modules/plutus/plutus_test.go` (`memStore` implements it)

**Interfaces:**
- Consumes: Task 1 `buildTransaction`.
- Produces:
  - `const BackupFormat = "nemosyne-backup"`, `const BackupVersion = 1`
  - `type BackupRow struct{ AmountMinor int64 "json:amount_minor"; Currency, Merchant, Category string; OccurredOn string "json:occurred_on"; Note, Reference string }` (JSON tags are the snake_case names in the spec)
  - `var ErrBackup = errors.New("not a valid Nemosyne backup")` (wrapped with a static reason)
  - `type RowError struct{ Row int; Reason string }` and `type RowsError struct{ Rows []RowError }` (implements `error`)
  - `func (m *Module) WriteBackup(ctx context.Context, w io.Writer, userID string) error`
  - `func (m *Module) parseBackup(r io.Reader, maxBytes int64) ([]NewTransaction, error)` returns `ErrBackup` or `*RowsError`
  - `Store.ExportTransactions(ctx context.Context, userID string) ([]Transaction, error)` (newest first, all rows)

- [ ] **Step 1: Write failing tests:**
  - `TestWriteBackup_Header`: output decodes to `format == "nemosyne-backup"`, `version == 1`, rows contain no `id` or `user` keys.
  - `TestParseBackup_Valid`: two rows parse to two `NewTransaction`.
  - `TestParseBackup_EmptyList` and `TestParseBackup_EmptyObjectWithHeader`: return zero rows, nil error.
  - `TestParseBackup_Rejects` (table): `not json`, truncated JSON, CSV text, PNG magic bytes, UTF-8 BOM + valid JSON, wrong `format`, `version: 2`, `version: 0`, unknown top-level field, unknown row field, trailing garbage after the object. Each returns an error that `errors.Is(err, ErrBackup)`.
  - `TestParseBackup_TooLarge`: input one byte over `maxBytes` returns `ErrBackup`.
  - `TestParseBackup_BadRowsReportedTogether`: rows 2 and 5 invalid returns `*RowsError` with `Row` 2 and 5 (1-based), `Reason` strings that contain none of the row's input text.
  - `TestParseBackup_IgnoresUserID`: a row containing `"user_id":"x"` is rejected as an unknown field.
- [ ] **Step 2: Run** `go test ./internal/modules/plutus -run 'Backup' -v`. Expected: FAIL (undefined).
- [ ] **Step 3: Implement** with `json.Decoder.DisallowUnknownFields`, `io.LimitReader(r, maxBytes+1)` to detect overflow, and a check that no tokens follow the object. `WriteBackup` reads `ExportTransactions`, maps to `BackupRow` (date as `2006-01-02`), writes indented JSON with `exported_at` in UTC. Add `ExportTransactions` to the Postgres store as `SELECT <txColumns> FROM transactions WHERE user_id = $1::uuid ORDER BY occurred_on DESC, created_at DESC` in `apps/api/internal/store/plutus.go`.
- [ ] **Step 4: Run** `go test -race -count=1 ./internal/modules/plutus ./internal/store` (store has no DB tests; this checks it compiles). Expected: PASS.
- [ ] **Step 5: Commit** `feat(plutus): backup file format and export`.

### Task 3: Merge import

**Files:**
- Modify: `apps/api/internal/modules/plutus/backup.go`, `plutus.go` (`Store` gains `ImportTransactions`), `plutus_test.go` (`memStore`)
- Modify: `apps/api/internal/store/plutus.go` (Postgres `ImportTransactions`)
- Test: `apps/api/internal/modules/plutus/backup_test.go`, `apps/api/internal/store/plutus_import_test.go` (skips unless `TEST_DATABASE_URL` is set)

**Interfaces:**
- Consumes: Task 2 `parseBackup`.
- Produces:
  - `type ImportResult struct{ Added, Skipped int }`
  - `func (m *Module) ImportBackup(ctx context.Context, userID string, r io.Reader, maxBytes int64) (ImportResult, error)`
  - `Store.ImportTransactions(ctx context.Context, userID string, rows []NewTransaction) (ImportResult, error)`: atomic; all rows or none.

- [ ] **Step 1: Write failing tests (memStore):**
  - `TestImport_AddsAndCounts`: three new rows give `Added 3, Skipped 0`.
  - `TestImport_Idempotent`: importing the same file twice gives second result `Added 0, Skipped 3`.
  - `TestImport_SkipsExistingReference` and `TestImport_SkipsSameAmountDateMerchant` (merchant differs only by case).
  - `TestImport_DuplicateRowsInsideFile`: two identical rows give `Added 1, Skipped 1`.
  - `TestImport_BadRowWritesNothing`: one invalid row among valid ones returns `*RowsError` and `ListTransactions` is empty.
  - `TestImport_ScopedToUser`: importing for `u1` leaves `u2` with no rows; two users importing the same file both get all rows.
  - `TestImport_RoundTrip`: `WriteBackup` then `ImportBackup` into an empty store gives equal amount, currency, merchant, category, date, note and reference.
- [ ] **Step 2: Run** `go test ./internal/modules/plutus -run TestImport -v`. Expected: FAIL.
- [ ] **Step 3: Implement.** `ImportBackup` parses, then calls `ImportTransactions` once. Postgres version: `pgx` transaction; for each row, skip if a row with same `reference` exists or `lower(merchant)` + amount + currency + date match (checked inside the transaction so rows earlier in the same file count), else insert; commit at the end, roll back on any error. `memStore` mirrors this under its mutex.
- [ ] **Step 4: Write the gated Postgres test** `TestPostgresImport` (same scenarios as idempotency, in-file duplicates and rollback on a forced failure); `t.Skip` when `TEST_DATABASE_URL` is empty. Run it against a local Postgres if one is available and say in the PR whether it ran.
- [ ] **Step 5: Run** `go test -race -count=1 ./internal/modules/plutus ./internal/store`. Expected: PASS (Postgres test skipped without the env var).
- [ ] **Step 6: Commit** `feat(plutus): merge import for backups`.

### Task 4: Admin request security

**Files:**
- Create: `apps/api/internal/core/olympus/security.go`
- Test: `apps/api/internal/core/olympus/security_test.go`

**Interfaces:**
- Produces: `func newGuard(addr string) (*guard, error)` (parses the port from `addr`; generates a 32-byte random CSRF token); `func (g *guard) Token() string`; `func (g *guard) Wrap(next http.Handler) http.Handler`.

- [ ] **Step 1: Write failing tests** with `addr = "127.0.0.1:8082"` and a stub `next`:
  - `TestHostAllowList`: `127.0.0.1:8082`, `localhost:8082`, `LOCALHOST:8082` reach `next`; `evil.com`, `127.0.0.1:9999`, `127.0.0.1`, empty, and `localhost.evil.com:8082` get `404` and `next` is not called.
  - `TestSecurityHeaders`: every response, including the `404`, has all five headers from Global Constraints with exact values, and no `Set-Cookie`.
  - `TestImportRequiresToken`: `POST /import` without token or with a wrong token gets `403`; correct token passes.
  - `TestImportRequiresSameOrigin`: `Origin: http://evil.com` gets `403`; `Origin: http://127.0.0.1:8082` passes; no `Origin` but `Referer: http://127.0.0.1:8082/` passes; neither header gets `403`.
  - `TestMethodsLimited`: `PUT /` and `DELETE /import` get `405`.
  - `TestTokenDiffersPerGuard`: two guards have different tokens; comparison uses `subtle.ConstantTimeCompare`.
- [ ] **Step 2: Run** `go test ./internal/core/olympus -run 'Host|Header|Import|Method|Token' -v`. Expected: FAIL.
- [ ] **Step 3: Implement.** The token is read from the multipart form field `csrf` (parse with a size-limited reader before checking, so large bodies are not buffered first). Check order: Host, then method, then Origin/Referer, then token.
- [ ] **Step 4: Run** `go test -race -count=1 ./internal/core/olympus`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(olympus): localhost-only request guard`.

### Task 5: Pages and import handler

**Files:**
- Create: `apps/api/internal/core/olympus/server.go`, `templates/index.html.tmpl`, `templates/result.html.tmpl`, `static/olympus.css` (all `embed`ded)
- Test: `apps/api/internal/core/olympus/server_test.go`

**Interfaces:**
- Consumes: Task 3 `plutus.ImportResult`, `plutus.ErrBackup`, `*plutus.RowsError`; Task 4 `newGuard`.
- Produces:
  - `type KeyInfo struct{ Name string; Scopes []string; CreatedAt time.Time; LastUsedAt *time.Time; Revoked bool }`
  - `type Info struct{ Mode, Version, PublicURL string; DatabaseOK bool; TransactionCount int; Keys []KeyInfo }`
  - `type Source interface{ Info(ctx context.Context) (Info, error) }`
  - `type Importer interface{ ImportBackup(ctx context.Context, r io.Reader, maxBytes int64) (plutus.ImportResult, error) }` (owner id is bound inside the implementation, not passed)
  - `type Options struct{ Addr string; Source Source; Importer Importer; MaxUploadBytes int64; Log *slog.Logger }`
  - `func New(o Options) (*Server, error)`; `func (s *Server) Handler() http.Handler` (guard already applied); `GET /static/olympus.css`

- [ ] **Step 1: Write failing tests** (fake `Source` and `Importer`, `httptest`, valid Host and token):
  - `TestIndexShowsConfig`: body contains mode, version, public URL, count, key name, scopes, `revoked` for a revoked key, and the masked connector URL `<PublicURL>/mcp/nmk_…`; contains a `csrf` hidden input and an `enctype="multipart/form-data"` form.
  - `TestIndexNeverShowsSecrets`: with `Info` containing a key named `HADES_KEY=abc` the text is escaped, and the page contains none of `DATABASE_URL`, `postgres://`, `key_hash`, `SESSION_SECRET`.
  - `TestIndexDatabaseDown`: `DatabaseOK=false` shows a visible "database unreachable" notice with status `200`.
  - `TestImportSuccessShowsCountsOnly`: result page shows `Added 2` and `Skipped 1`.
  - `TestImportEscapesAndNeverEchoes`: fake importer returns a `*plutus.RowsError` with a `Reason`; and a separate upload whose merchant is `<script>alert(1)</script>` never appears in the response body at all.
  - `TestImportBadFileMessage`: `ErrBackup` gives status `400` and a short static message; unexpected error gives `500` with a generic message and nothing from the error text.
  - `TestImportTooLarge`: body over `MaxUploadBytes` gives `413`; the fake importer is not called with more than the cap.
  - `TestImportMissingFile`: no `file` part gives `400`.
  - `TestLogsHaveNoBodies`: with a buffer logger, a request log line has only method, path, status; the uploaded content is absent.
- [ ] **Step 2: Run** `go test ./internal/core/olympus -run 'Index|Import|Logs' -v`. Expected: FAIL.
- [ ] **Step 3: Implement** with `html/template` (never `text/template`), templates parsed once at startup from `embed.FS`, no inline styles or scripts (CSP forbids them). Default `MaxUploadBytes` is 10 MiB when zero. Map `ErrBackup` to `400`, `*RowsError` to `422` (list up to 20 rows then "and N more"), `http.MaxBytesError` to `413`, anything else to `500`.
- [ ] **Step 4: Run** `go test -race -count=1 ./internal/core/olympus`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(olympus): config view and backup import page`.

### Task 6: Wire into the binary (blocked on Phase 1)

**Files:**
- Modify: `apps/api/cmd/nemosyne-api/main.go` and the Phase 1 CLI dispatcher file (path set by Phase 1)
- Modify: `apps/api/internal/config/config.go` and `config_test.go` (`AdminAddr`, from `ADMIN_ADDR`, default `127.0.0.1:8082`)
- Test: `apps/api/internal/config/config_test.go`, `apps/api/cmd/nemosyne-api/admin_wiring_test.go`

**Interfaces:**
- Consumes: Task 5 `olympus.New`, `olympus.Source`, `olympus.Importer`; Task 2 `Module.WriteBackup`; from Phase 1: the owner user id lookup and the key-list query.
- Produces: `config.Config.AdminAddr string`; CLI command `nemosyne export` writing the backup JSON to stdout.

- [ ] **Step 1: Write failing tests:** `TestLoad_AdminAddrDefault` (empty gives `127.0.0.1:8082`; custom value kept; value without a port rejected); `TestAdminServerOnlyInSelfhost` (hosted mode starts no admin server); `TestPublicMuxHasNoUIRoutes` (public handler returns `404` for `/` and `/import`); `TestExportCommandOutput` (writes valid JSON to a buffer, nothing else to stdout).
- [ ] **Step 2: Run** `go test ./internal/config ./cmd/...`. Expected: FAIL.
- [ ] **Step 3: Implement.** Build the admin `http.Server` with `ReadHeaderTimeout 10s`, `ReadTimeout 60s`, `WriteTimeout 60s`, `IdleTimeout 2m`; shut it down with the main server. Adapters: an `olympus.Source` backed by the store (ping, count, keys) and an `olympus.Importer` that binds the owner id. Log only the admin address at startup. Do not add the admin address to any tunnel config.
- [ ] **Step 4: Run** `gofmt -l .`, `go vet ./...`, `go test -race -count=1 ./...`. Expected: no gofmt output, all PASS.
- [ ] **Step 5: Manual check** (record exact commands and results): start the API in selfhost mode, `curl -i http://127.0.0.1:8082/` returns `200`; `curl -i -H 'Host: evil.com' http://127.0.0.1:8082/` returns `404`; `curl -i -X POST http://127.0.0.1:8082/import` returns `403`; `curl -i http://127.0.0.1:8081/` returns `404`.
- [ ] **Step 6: Commit** `feat(api): serve local admin page in selfhost mode`.

### Task 7: Compose, env, docs and agent log (needs owner approval first)

**Files:**
- Modify: `deploy/docker-compose.yml`, `.env.example` (after approval)
- Modify: `README.md`, `docs/design.md`; the self-host scope doc if it is on this branch
- Create: `agent-logs/claude/YYYY-MM-DD-HHMM-selfhost-admin-page.md` (template in `agent-logs/README.md`)

- [ ] **Step 1: Get explicit owner approval** for the compose and `.env.example` edits. Stop here if not given.
- [ ] **Step 2: Edit compose:** `ADMIN_ADDR: "0.0.0.0:8082"` in the `api` service environment and `ports: ["127.0.0.1:${ADMIN_PORT:-8082}:8082"]`; add `ADMIN_PORT` to `.env.example`. Leave every other service as is.
- [ ] **Step 3: Verify** `ADMIN_PORT=8082 POSTGRES_PASSWORD=x SESSION_SECRET=$(printf 'a%.0s' {1..40}) docker compose -f deploy/docker-compose.yml config -q`. Expected: exit 0. Confirm the rendered `ports` entry starts with `127.0.0.1:`.
- [ ] **Step 4: Docs.** README: how to open the page, `docker compose exec api nemosyne export > backup.json`, and the warning that the page must never be added to a tunnel. `docs/design.md`: describe Olympus as the local admin page in self-host mode. Update the scope doc: "no public ports" and `nemosyne export` moved into this work.
- [ ] **Step 5: Agent log** with commands run and results, what was not run (Go 1.26 availability, Postgres test, Docker run if skipped), and the `apps/web` assumption.
- [ ] **Step 6: Check links** in the edited docs exist. Commit `docs: document local admin page and export`, open the PR with the template filled in honestly, wait for green CI.

---

## Self-review

- **Spec coverage:** architecture and routes (Task 5, 6); config view and masked URL (5); format, export, import rules (2, 3); security table (4, 5); Docker exception (7); tests and manual checks (each task, 6); approvals and ordering (Prerequisites, 7). No gaps found.
- **Type consistency:** `ImportResult`, `RowsError`, `ErrBackup`, `Importer.ImportBackup` and `Source.Info` are defined once (Tasks 2, 3, 5) and used under the same names in Task 6.
- **Proportion:** signatures, test names and assertions only; function bodies are left to the implementer.
