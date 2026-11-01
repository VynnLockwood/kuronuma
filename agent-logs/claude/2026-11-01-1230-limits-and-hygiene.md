# Per-user caps, unused client cleanup, quoted names (S2)

- Agent: claude
- Date: 2026-11-01
- Branch / PR: fix/limits-and-hygiene

## Goal
The remaining small security fixes before open sourcing: per-user caps, cleanup of unused registered clients, quoted category names in tool text.

## What I did
- Caps: 50,000 transactions and 5,000 tasks, to-dos and notes per user, checked before anything is created; updating and deleting never blocked; a repeated reference at the cap is answered as already saved. New `CountTransactions`, `CountMemories` and `TransactionByReference` store methods. Payee rules were already capped at 500, so nothing was added there.
- Cleanup: `CleanupExpired` also removes registered clients over 180 days old that hold no code and no token (`DeleteUnusedClients`).
- Every pre-consent OAuth error page now tells the person how to recover (go back to the assistant, or remove and re-add the connector).
- Quoting: the "Their categories" list in the not-in-list error now quotes each name; the audit found no other unquoted path.
- `docs/design.md` updated.

## Verified
- Each change had a failing test first (compile or behaviour): caps, cleanup timing, error-page hint, quoting, repeated reference. `gotest.sh` ok (cerberus top level, hermes, themis, atlas, plutus), vet ok, `gofmt -l apps/api` empty.
- A fresh whole-branch review found 2 Important issues (30-day cleanup stranded idle users; repeated reference refused at the cap); both fixed with failing tests first.

## Not verified / known gaps
- The `store` package (real SQL, a new DB-gated count test) cannot be built here (no Go 1.26 toolchain access); CI compiles it. The DB-gated tests are skipped in CI (no database), so the SQL itself has not run anywhere; I read it against the migrations.
- Not tried against live assistants.

## Decisions and assumptions
- The plan proposed 30 days for client cleanup; I use 180 (ruling in the PR).
- Counting then inserting can overshoot a cap by one row under a race.
- Deferred minors: mid-flow client deletion at the 180-day edge; no DB test for the cleanup SQL; stale log label in `main.go`; no test that a refused add creates no category; a docs sentence order; JSON decode errors unquoted in `hermes/toolkit.go`.

## Next steps
- The dashboard plan and PRs.
