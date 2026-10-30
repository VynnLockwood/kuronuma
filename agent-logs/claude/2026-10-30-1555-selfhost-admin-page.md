# Self-host local admin page (Tasks 1-5 of the plan)

- Agent: claude
- Date: 2026-10-30
- Branch / PR: feat/selfhost-local-admin-page (local, not pushed, no PR yet)

## Goal
Owner asked to move to the single-owner self-host scope and replace the web frontend, in self-host mode, with a simple Go page that shows config and has a backup import button. Spec: `docs/superpowers/specs/2026-10-30-selfhost-local-admin-page-design.md`. Plan: `docs/superpowers/plans/2026-10-30-selfhost-local-admin-page.md`.

## What I did
- Shared the `add_transaction` validation as `buildTransaction` (plutus), no behaviour change.
- Backup file format (`nemosyne-backup` v1), `WriteBackup`, strict `parseBackup` (plutus).
- Merge import: `PlanImport` + `ExistingIndex`, `ImportBackup`, `Store.ImportTransactions` and `ExportTransactions` (in-memory test store and Postgres).
- `olympus` package: localhost guard (Host allow-list, CSRF token, Origin check, headers) and the config/import pages with embedded templates and CSS.
- Not done: Tasks 6-7 (main.go wiring, `nemosyne export` CLI, compose, `.env.example`, README/design docs). They need the self-host core PR (Phase 1) and owner approval for deploy files.

## Verified
- `/home/claude/scratch/gotest.sh` (scratch copy of plutus, hermes, olympus with the `go` directive lowered to 1.24, `-race`): all packages ok, 71 tests passed, 0 failed.
- `gotest.sh vet` clean; `gofmt -l apps/api` prints nothing.
- Real Chromium (Playwright) against the page with sample data: owner import shows the result page; a cross-site form post returns 403; a spoofed `Host` returns 404; a post without the token returns 403. Screenshots of both pages reviewed.
- Import SQL (reference conflict, existing-records query, `pg_advisory_xact_lock`) run in psql on a temporary Postgres 16 with the real migrations: same-user sessions serialised (2.3 s wait), different users did not wait; a CHECK violation raises as expected.
- Final whole-branch review by a fresh Opus reviewer: no Critical findings; two Important findings fixed with tests that failed first.

## Not verified / known gaps
- Go 1.26 and CI: the sandbox has Go 1.24 and cannot download the 1.26 toolchain or modules (egress policy). Nothing here was run with the real toolchain.
- `apps/api/internal/store/plutus.go` (Export/ImportTransactions) and `plutus_import_test.go` were never compiled or run (pgx unavailable); only gofmt-checked and reviewed by eye. First CI run is the compile check. `TestPostgresImport` skips in CI (no database); run it with `TEST_DATABASE_URL` locally.
- No Docker or compose run. Tasks 6-7 untested because not written.

## Decisions and assumptions
- Merge rule is a multiset merge, and records with and without a reference never match each other (spec text updated). A literal "skip same amount/date/merchant" would drop real repeated payments on restore.
- `Referrer-Policy: same-origin` instead of `no-referrer`: with `no-referrer` Chrome sends `Origin: null` on the form post and the guard locks the owner out (found in the browser run). Spec updated.
- Guard takes a body cap (`newGuard(addr, maxBody)`) and parses the multipart form fully in memory so no upload touches disk.
- `apps/web` assumed to stay for hosted mode, untouched.
- Commits were first made with the sandbox git identity, then re-authored as the repo owner (`VynnLockwood` noreply identity) at the owner's request, with the owner's approval to force-push this branch (AGENTS.md rule 2). Content, messages and dates unchanged. No AI trailers, per AGENTS.md.
- Deferred minors from review: blank page on 403/413 after a restart; extra file copies in memory; row-by-row inserts; `frame-ancestors 'none'` in the CSP.

## Next steps
- Merge self-host core (Phase 1), then do Tasks 6-7. Compose must publish the same port inside and outside the container (the Host allow-list uses the listen port) and the admin server `WriteTimeout` must allow a 10 MB import.
- Owner approval needed for `deploy/docker-compose.yml` and `.env.example`.
- Run the Go tests and `TEST_DATABASE_URL` test with Go 1.26; open the PR once Tasks 6-7 are in.
