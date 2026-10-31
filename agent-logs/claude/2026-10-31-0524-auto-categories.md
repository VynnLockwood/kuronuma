# Auto-create categories and describe Nemo as general memory (PR E)

- Agent: claude
- Date: 2026-10-31
- Branch / PR: feat/auto-categories (PR link on GitHub)

## Goal
Owner reported two things from real use: the assistant could only recommend an existing category instead of making a new one, and it told them Nemo only handles money, not tasks or work. Approved design (in chat): the service creates categories itself when a record is saved, and describes itself as general memory.

## Diagnosis
- The owner's connector lists only the four original spending tools. Assistants keep the tool list from when they connected, so it has no `create_category`, `add_memory` and so on. The server refused unknown categories and told the assistant to use `create_category`, which it could not see, so it could only suggest an existing name. Re-adding the connector fixes the missing tools; this PR fixes the behavior even for a connector that has not been re-added, because `add_transaction` is still there.
- The server's built-in instructions said only "personal data", so assistants treated it as a money app.

## What I did
- `themis.Ensure` (new): makes sure a category is in the user's list and creates it if not; `uncategorized` needs nothing; at `themis.MaxCategories` (100 per user, new) a new name is refused with a message to pick an existing category or merge.
- `add_transaction`, `add_memory` and `update_memory` call `Ensure` instead of refusing; the result carries `category_created: true` when a category was made. Payee rules (`set_payee_rule`) keep the old refusal on purpose (approved scope), and a payee rule whose category left the list is still skipped.
- `plutus.Store` and `atlas.Store` gain `CreateCategory` (the Postgres store already has it).
- New `hermes.DefaultInstructions` (used in `main.go`): says Nemosyne is the user's memory for spending, tasks, to-dos and notes, names the tools, how categories work, and keeps the data-not-instructions and confirm-before-saving rules.
- Tool wording: category parameters, `list_categories` and `create_category` no longer say "ask first / never invent / unknown names are refused".
- `docs/design.md` Plutus and Themis sections.

## Verified
- `/home/claude/scratch/gotest.sh` (scratch copy of hermes, themis, plutus, atlas; Go 1.24 directive; `-race`): all ok; `gotest.sh vet` clean; `gofmt -l apps/api` empty. Tests written first and seen failing (compile errors for the missing symbols), then passing.
- A fresh Opus reviewer (read-only, own checkout) found nothing Critical. One Important finding fixed test-first (see Review). After the fix: all ok again, vet clean, gofmt empty.
- New tests: `TestEnsure` (9 cases: uncategorized, listed, new, first ever, lost race, 99 and 100 at the cap, store failure passes through untouched), `TestUnknownCategoryIsCreated` (plutus and atlas, including update and repeat saves, per-user lists, Thai names), cap tests for both modules, payee-rule refusal kept (with the 50-name cap), instructions test.

## Not verified / known gaps
- Go 1.26 and CI: `internal/store` and `cmd/` are not compiled in my sandbox. The Postgres store is unchanged and already satisfies both interfaces (`var _ plutus.Store`, `var _ atlas.Store`).
- Nothing deployed or tried with a real assistant. After deploy a connector that has not been re-added still has only the old tools; it gets auto-created categories on `add_transaction` but not the task tools until re-added.

## Review
- **Important, fixed:** the category is created before the duplicate checks, so `already_saved`, `possible_duplicate` (spending) and `not_found` (`update_memory` with an unknown id) could leave a new category behind without saying so, breaking the "result says category_created" promise. Now every outcome after the creation reports `category_created`. Three new tests, seen failing first.
- **Minor, fixed:** the instructions said "you do not need to ask first" next to "get their confirmation", which could read as skipping confirmation of the record. Now: "a new category needs no separate question, but still confirm the record itself." The instructions test no longer checks for a phrase that never existed.
- **Minor, fixed:** the earlier spec still said unknown categories are refused; added a "superseded by PR E" line. The plan files are history and were left alone.
- **Minor, not changed (flagged in the PR):** the consent screen text for `plutus:write` ("Add and delete your spending records") and `atlas:write` does not mention that they can now create categories. Changing scope text needs the owner's sign-off.
- **Minor, deferred:** the 100 cap can be overshot by parallel saves (count and insert are separate steps; a single `INSERT ... SELECT ... WHERE count < 100` would fix it). Look-alike names (Unicode forms, Cyrillic letters) become separate categories; a real fix needs `x/text`, a new dependency that needs the owner's approval. Pre-existing behavior, but auto-creation makes it more likely.

## Decisions and assumptions
- **Permission change (owner approved in chat):** `plutus:write` and `atlas:write` can now also create a category as a side effect of saving. No scope was added or changed and Cerberus is untouched.
- Cap of 100 categories per user is my choice; the owner can change the constant.
- Payee rules were left refusing unknown categories (not in the approved design). The owner may want them to create too.

## Deferred minors
- (see Review for the cap race and look-alike names)
- A category made by a typo is only fixed by `merge_categories`; there is no fuzzy "did you mean" hint.
- Earlier spec (`docs/superpowers/specs/2026-10-31-memory-and-categories-design.md`) still describes the refusal; it is a historical record and was not edited.
