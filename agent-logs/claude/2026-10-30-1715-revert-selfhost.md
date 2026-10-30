# Revert the self-host local admin page (#12)

- Agent: claude
- Date: 2026-10-31
- Branch / PR: revert/selfhost-pr-12

## Goal
Owner asked to go back to the hosted version. Scope chosen by the owner: revert all of PR #12, keep its agent log.

## What I did
- `git revert -m 1 671d90a` on a fresh branch from `main`, then restored `agent-logs/claude/2026-10-30-1555-selfhost-admin-page.md`.
- Removed with the revert: `apps/api/internal/core/olympus`, Plutus backup export/import (`backup.go`, `import.go`, `buildTransaction`, the two new `Store` methods), the Postgres import code and test, and the self-host spec and plan.

## Verified
- `git diff 62a9e15 HEAD` (62a9e15 = `main` before #12): the only difference is the kept agent log. `git diff --quiet` on `apps docs deploy .github README.md AGENTS.md` reports identical.
- Scratch run (Go 1.24 directive, `-race`) of `hermes` and `plutus`: ok. CI results are in the PR.

## Not verified / known gaps
- Nothing was run with Go 1.26 locally; CI is the check.
- Hosted deployment itself was not touched or re-deployed.

## Decisions and assumptions
- Kept the old agent log instead of deleting it (it records the work and what went wrong and right). Its text still describes the self-host direction; that is history, not current state.
- The backup export/import code is gone. `docs/design.md` still promises "full export and delete-everything" for hosted users, so that will need to be rebuilt for hosted mode if wanted; it is in git history (commits in #12).

## Next steps
- Separate PR: remove the hero dragonfly sprites.
- Then design flexible memory (tasks, to-dos, notes), managed categories and payee rules.
