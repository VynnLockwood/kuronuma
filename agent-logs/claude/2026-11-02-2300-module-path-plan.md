# Go module path rename plan (PR P)

- Agent: claude
- Date: 2026-11-02
- Branch / PR: docs/module-path-plan

## Goal
Plan renaming the Go module from `.../kuronuma-mcp/apps/api` to `.../kuronuma/apps/api` so the import path matches the public repository.

## What I did
- Measured the scope on `main` with `git grep`: 40 lines in 26 files outside history (24 Go files, `go.mod`, one line in `docs/design.md`); 2 files are in cerberus (imports only).
- Checked the Dockerfile and CI: they use directories, not the module path.
- Wrote `docs/superpowers/plans/2026-11-02-module-path.md`.

## Verified
- Docs only; the counts and file lists come from `git grep` on `main`.

## Not verified / known gaps
- Nothing is changed yet. `store` and `cmd` can only be compiled by CI.

## Decisions and assumptions
- New path is `github.com/VynnLockwood/kuronuma/apps/api`.
- History (agent logs, old plans and specs) is left as it was.

## Next steps
- Owner merges the plan; then implement as `refactor/module-path`.
