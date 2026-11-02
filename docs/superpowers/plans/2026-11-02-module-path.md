# Go module path rename (PR P) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline) or superpowers:subagent-driven-development. Steps use checkbox (`- [ ]`) syntax.

**Goal:** The repository is now `github.com/VynnLockwood/kuronuma` and public, but the Go module is still `github.com/VynnLockwood/kuronuma-mcp/apps/api`. Rename the module so the import path matches the repository and `go get github.com/VynnLockwood/kuronuma/apps/api/...` works for anyone.

**Decision:** the new module path is `github.com/VynnLockwood/kuronuma/apps/api` (the module lives in the `apps/api` subdirectory, so the path ends with it).

## Scope (measured on `main`)

- 40 lines in 26 files outside history: `apps/api/go.mod` (the `module` line), 24 Go files (imports) across `cmd/kuronuma-api`, `internal/core/{cerberus,cerberus/google,hermes,themis}`, `internal/modules/{atlas,plutus}` and `internal/store`, and one line in `docs/design.md` (the repository name).
- **Not changed, on purpose:** `agent-logs/` and `docs/superpowers/plans/` and `docs/superpowers/specs/` (7 lines in 4 files plus one spec; they record what was true when they were written), `go.sum` (the main module's own path is not in it), the Dockerfile and CI (they refer to directories, not the module path; the build context is `apps/api`).
- **Needs the owner's approval (AGENTS.md):** two files are in `internal/core/cerberus` (import lines only; no auth, session, OAuth, scope or permission logic changes). Merging this plan is that approval.
- **Cannot be fully run here:** the `store` and `cmd` packages need third-party dependencies, so only CI compiles them. The plan compensates with a script that proves every import of the new path points at a real directory, and CI is the final proof.

## Rules

1. Mechanical only: replace the exact string `github.com/VynnLockwood/kuronuma-mcp/apps/api` with `github.com/VynnLockwood/kuronuma/apps/api` in `apps/api/go.mod` and the Go files. No other edit in those files.
2. `gofmt -l` stays empty (import order is unchanged because every changed path keeps the same neighbours).
3. `go.sum` is not touched; `go mod tidy` must be a no-op (CI already enforces it).
4. One concern: no refactor, no rename of directories or binaries (`cmd/kuronuma-api` stays).

## Review Focus

- No occurrence of the old path remains outside the excluded history folders (`git grep`).
- Every import beginning with the new module path resolves to a directory that exists and contains Go files, in every Go file including test files and the `store` and `cmd` packages the scratch harness cannot compile.
- `go.mod` changed by exactly one line; `go.sum`, Dockerfiles, nginx and workflows are untouched; no cerberus line other than an import changed.
- `docs/design.md` says the repository is `kuronuma` and the sentence still reads correctly.
- The scratch Go harness is updated for the new path so its results are real (it writes its own go.mod).

---

### Task 1: Rename the module path

**Files:** Modify `apps/api/go.mod`, the 24 Go files that import the module, `docs/design.md` line 55. Test (outside the repo): `/home/claude/scratch/module-path-check.py`.

- [ ] **Step 1: Write the failing check.** It fails if: `apps/api/go.mod` does not start with `module github.com/VynnLockwood/kuronuma/apps/api`; any file outside `agent-logs/`, `docs/superpowers/` contains `VynnLockwood/kuronuma-mcp`; any Go import of the new prefix lacks a matching directory with `.go` files; `docs/design.md` still calls the repository `kuronuma-mcp`; `go.sum` differs from `main`; `git diff --stat` touches a file outside the expected list.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement** with one scripted replacement; read the diff.
- [ ] **Step 4: Verify.** The check passes; update the scratch harness module line; `gofmt -l apps/api` empty; `bash gotest.sh` (cerberus, hermes, themis, webguard, atlas, plutus) and `gotest.sh vet` pass; CI compiles everything.
- [ ] **Step 5: Commit** `refactor(api): rename the Go module to match the repository`. Fresh Opus review with the Review Focus, one fix pass, agent log, PR with the real CI result.

## Not included

- Renaming `cmd/kuronuma-api`, packages, the binary or the Docker image.
- Editing history (agent logs, old plans and specs).
- Dashboard PRs J and K (separate plans, already merged).
