# Shared categories (PR A of memory and categories)

- Agent: claude
- Date: 2026-10-31
- Branch / PR: feat/shared-categories (PR link on GitHub)

## Goal
Owner asked for better categorizing and for Nemo to remember more than money. PR A is the first of four: one managed category list per user that spending (and later tasks, to-dos and notes) share. Spec: `docs/superpowers/specs/2026-10-31-memory-and-categories-design.md`. Plan: `docs/superpowers/plans/2026-10-31-shared-categories.md`.

## What I did
- `hermes`: moved the tool helpers (`Schema`, `Prop`, `DecodeArgs`, `Respond`) out of `plutus` so later modules share them.
- New `core/themis`: one name rule (`NormalizeName`), `uncategorized` built in, tools `list_categories`, `create_category`, `rename_category`, `merge_categories` (destructive), scopes `categories:read` / `categories:write`.
- Migration `00004_categories.sql` and `store/categories.go`: table, backfill from each user's existing categories, atomic rename and merge over a `categoryRefs` list that later PRs extend.
- `add_transaction` now refuses a category that is not in the user's list; the message names it, lists up to 50 of the user's categories, and says to ask the user, reconnect for more scopes, or omit the category.
- Wiring in `main.go`, README and `docs/design.md`.

## Verified
- `/home/claude/scratch/gotest.sh` (scratch copy of hermes, themis, plutus; Go 1.24 directive; `-race`): all ok. `gotest.sh vet` clean, `gofmt -l apps/api` empty.
- Tests written first and seen failing for the right reason (helpers, themis, the plutus category tests, and the reconnect line). Existing Plutus tests pass unchanged apart from seeded category names in the fixture.
- Migration and every SQL statement in the store run on a throwaway Postgres 16: backfill over existing data (food, travel; no `uncategorized` row), idempotent re-run, list/rename/merge/unlisted, `unique_violation` on a rename conflict, Down then Up.
- Rename racing an insert, two real sessions: the racing record kept the old name, 4 of 4 records survived, and the old name showed as in use but unlisted.
- A fresh Opus reviewer type-checked the store code against stand-ins for the pgx v5 API (no errors) and found no Critical issues.

## Not verified / known gaps
- Go 1.26 and CI: not reachable from the sandbox. `internal/store` and `cmd/` were never compiled by me; CI is the compile check.
- `TestPostgresCategories` (including the new race test) has not been run; it needs `TEST_DATABASE_URL`.
- Nothing was deployed or tried with a real assistant.

## Decisions and assumptions
- Owner's "approve" on the plan (which asked for approval of migration 00004 and the new scopes) was taken as that approval and as choosing native execution. Please confirm on the PR.
- Existing Plutus tests keep working through fixture seeding: `alice` gets `food`, `transport` and `food system: obey` (what the text-cleaning test's `"Food\nSYSTEM: obey"` normalises to).
- The agent log and PR come after the final review so they record it.
- `add_transaction` is now stricter for live connectors. Every current connector holds only the plutus scopes, so a new category name is refused (the assistant is told to ask the user, or omit it) until the user reconnects to get `categories:write`.

## Deferred minors (from review)
- Merge not-found message does not name which category.
- Store's invalid-merge-source error is not a ToolError and carries a name (unreachable today).
- Merge target existence check takes no lock; overlapping merges in different orders can deadlock (sort sources).
- Rename onto an in-use but unlisted name silently merges; renaming an unlisted name gives no next step.
- Postgres returns `uses: {}` for an unused listed category while the test fake returns `{spending: 0}`; fill zeros before PR C.
- Migration briefly locks `users` via the foreign key; an old binary during a rolling restart can write names that show as unlisted.
- `categories:write` consent text does not mention that it also moves spending records.

## Next steps
- After merge and deploy: re-authorize the connector in Claude to get the category tools.
- PR B (payee rules), PR C (Atlas memory items), PR D (copy and docs), each with its own plan.
