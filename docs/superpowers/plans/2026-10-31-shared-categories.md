# Shared Categories (PR A) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** One managed category list per user, enforced when spending is saved, with tools to list, create, rename and merge categories.

**Architecture:** A new small package `core/themis` owns category names, rules and the four MCP tools, over a `Store` interface that the Postgres store implements. `plutus` validates `add_transaction` categories against the list through one new `Store` method. Rename and merge touch every category-bearing table through a single list in the Postgres store, which PR B (payee rules) and PR C (Atlas) extend. Only `transactions` exists today.

**Tech Stack:** Go standard library, existing `pgx` and `goose`. No new dependencies.

**Spec:** `docs/superpowers/specs/2026-10-31-memory-and-categories-design.md` (sections 4, 8 and 9). This is plan A of four; B (payee rules), C (Atlas) and D (copy and docs) get their own plans after A merges, because they consume A's interfaces.

## Global Constraints

- Category names: lowercase, trimmed, whitespace collapsed, at most 40 characters, via `strings.ToLower(hermes.Untrusted(s, 40))`; empty after normalising is invalid.
- `uncategorized` is always valid and implicit (no row); it cannot be created, renamed or merged away; it may be a merge target.
- Records keep their category as a name; the `transactions` table does not change.
- An unknown category is refused, not created. The error names the category, lists the user's categories (at most 50, then "and N more"), says to ask the user whether to create it with `create_category` or choose one, and says the category can be omitted to file under `uncategorized`.
- Rename `A` to `B` fails if `B` exists; merge target must exist or be `uncategorized`; both run in one database transaction.
- `list_categories` also returns in-use names missing from the list, flagged `unlisted`.
- Merge: 1 to 10 sources, de-duplicated, none equal to the target or `uncategorized`.
- Scopes: `categories:read` ("Read your category list") for `list_categories` (read-only); `categories:write` ("Create, rename and merge your categories") for `create_category`, `rename_category`, `merge_categories` (destructive).
- Migration `00004_categories.sql`, additive only: table `categories(id, user_id, name, created_at)`, unique `(user_id, name)`, no CHECK on name, backfilled with `SELECT DISTINCT user_id, category FROM transactions WHERE category <> 'uncategorized'`.
- Every query filters by the verified user id; no user id from arguments. Text returned to the assistant goes through `hermes.Untrusted` and the response carries `hermes.UntrustedNote`. Nothing is logged except method, path and status.
- Verify per `AGENTS.md`: `gofmt -l .` empty, `go vet ./...`, `go test -race -count=1 ./...` from `apps/api`. Commits: Conventional Commits, authored as the repo owner (already configured), no AI trailers. Never push to `main`.

## Review Focus

- A name differing only by case, spacing or invisible characters (`Food`, ` food `, `fo​od`) is the same category. Pinned in Task 2 and Task 4.
- Thai names (`อาหาร`) are created, listed, matched and counted by characters, not bytes. Pinned in Task 2 and Task 4.
- An assistant holding only `plutus:write` that hits an unknown category can recover (the message says to omit it), and a user with hundreds of categories does not get a giant error. Pinned in Task 4.
- Merge and rename edge cases: source equal to target, duplicate sources, `uncategorized` as a source, an empty or all-invisible name, the same name twice in one call. Pinned in Task 2.
- A name in use but missing from the list (`unlisted`) can be merged away and shows up in `list_categories`. Pinned in Task 3 (database test and psql) and Task 2 (tool output).

## Prerequisites (gate, not tasks)

- **Owner approval before Task 3 and Task 5:** new migration `00004_categories.sql`, and new scopes `categories:read` and `categories:write` in `main.go`. Ask in the plan review.
- **Go 1.26:** this sandbox has Go 1.24 and cannot fetch 1.26 or modules. Tasks 1, 2 and 4 run in the scratch harness `/home/claude/scratch/gotest.sh` (a stdlib-only copy with the `go` directive lowered). Add `internal/core/themis` to its copy list when Task 2 creates it. `internal/store` and `cmd/` import `pgx` and cannot be compiled here: gofmt only, CI is the compile check. Say so in the PR; do not claim they were run.
- Work on branch `feat/shared-categories` from the branch holding this plan and the spec.

---

### Task 1: Move the tool helpers into `hermes`

`themis` (and later Atlas and payee rules) need the same helpers `plutus` keeps private. Move them once; no behaviour change.

**Files:**
- Create: `apps/api/internal/core/hermes/toolkit.go`, `apps/api/internal/core/hermes/toolkit_test.go`
- Modify: `apps/api/internal/modules/plutus/plutus.go` (delete `schema`, `prop`, `decodeArgs`, `respond`; call the `hermes` versions)

**Interfaces:**
- Produces: `hermes.Schema(props map[string]any, required ...string) map[string]any`; `hermes.Prop(typ, desc string) map[string]any`; `hermes.DecodeArgs(raw json.RawMessage, dst any) error` (strict, unknown fields rejected, returns a `ToolError` starting `Invalid arguments:`); `hermes.Respond(fields map[string]any) (string, error)` (adds `_note` = `UntrustedNote`, marshals JSON).

- [ ] **Step 1: Write failing tests** in `toolkit_test.go`: `TestSchema` (object type, `additionalProperties` false, `required` present only when given); `TestDecodeArgsStrict` (unknown field and wrong type both return `ToolError` containing `Invalid arguments`; valid input fills the struct); `TestRespondAddsNote` (output JSON has `_note` equal to `UntrustedNote` and keeps the given fields).
- [ ] **Step 2: Run** `go test ./internal/core/hermes -run 'Schema|DecodeArgs|Respond' -v` (scratch harness). Expected: FAIL, undefined.
- [ ] **Step 3: Implement** the four functions in `toolkit.go` by moving the bodies from `plutus.go`; replace every call in `plutus.go` and delete the private copies.
- [ ] **Step 4: Run** the whole scratch suite. Expected: all PASS, including every existing `plutus` test unchanged.
- [ ] **Step 5: Commit** `refactor(hermes): share tool helpers with modules`.

### Task 2: The `themis` package: names, rules and the four tools

**Files:**
- Create: `apps/api/internal/core/themis/themis.go`, `apps/api/internal/core/themis/themis_test.go`

**Interfaces:**
- Consumes: Task 1 helpers; `hermes.Tool`, `hermes.Caller`, `hermes.Errorf`, `hermes.Untrusted`.
- Produces:
  - `const ScopeRead = "categories:read"`, `ScopeWrite = "categories:write"`, `Uncategorized = "uncategorized"`
  - `func NormalizeName(s string) string`
  - `var ErrExists`, `ErrNotFound` (sentinel errors the store returns)
  - `type Category struct{ Name string; Listed bool; Uses map[string]int }` (`Uses` keyed by module label; this PR fills only `"spending"`)
  - `type Store interface`:
    - `ListCategories(ctx context.Context, userID string) ([]Category, error)` (every listed name plus every in-use name missing from the list, `Listed` false for those; `Uncategorized` need not appear)
    - `CreateCategory(ctx context.Context, userID, name string) (created bool, err error)`
    - `RenameCategory(ctx context.Context, userID, from, to string) (updated map[string]int, err error)` (`ErrNotFound` if `from` not listed, `ErrExists` if `to` listed)
    - `MergeCategories(ctx context.Context, userID string, sources []string, target string) (moved map[string]int, err error)` (`ErrNotFound` if the target is neither listed nor `Uncategorized`, or a source is neither listed nor in use)
  - `func New(store Store) *Module`; `func (m *Module) Register(s *hermes.Server)`
  - Tools: `list_categories`, `create_category` (`{name}`), `rename_category` (`{from, to}`), `merge_categories` (`{sources: [string], target}`), with the scopes and flags in Global Constraints.

- [ ] **Step 1: Write failing tests** in `themis_test.go`, using an in-memory `Store` fake (per-user lists and per-name `spending` counts, returning the sentinel errors like the real one) behind the real MCP server, in the style of `plutus_test.go` (`newEnv`, `call`). Names and assertions:
  - `TestNormalizeName`: `"  Food "`→`food`; `"fo​od"`→`food`; `"อาหาร"` unchanged; 50 Thai characters → exactly 40 characters (count runes); `"​ "`→`""`.
  - `TestListAlwaysIncludesUncategorized`: empty store still returns `uncategorized` (listed); an in-use name missing from the list returns with `"unlisted": true`; order is total uses descending then name; response has `_note`.
  - `TestCreate`: `"  Food "` creates `food`; second create returns `status` `already_exists` (not an error); `create_category` `uncategorized` returns `already_exists`; empty or all-invisible name is an error; Thai name round-trips through `list_categories`.
  - `TestRename`: success returns `updated.spending` and the list shows the new name; target exists → error containing `merge_categories`; unknown → error; `from` or `to` `uncategorized` → error; `from == to` after normalising (`Food` vs `food`) → error.
  - `TestMerge`: success returns `moved`; target `uncategorized` allowed; duplicate sources (`a`, `A`, `a`) merged once; target among sources → error; `uncategorized` as a source → error; 11 sources → error; zero sources → error; unknown source or target → error naming `list_categories`; an `unlisted` source merges away.
  - `TestScopesAndFlags`: a token with only `categories:read` sees only `list_categories` in `tools/list` (read-only true); `merge_categories` is destructive; a read-only token calling `create_category` is refused.
  - `TestUsersAreIsolated`: user B never sees or changes user A's categories.
  - `TestTextIsData`: a name with control and zero-width characters comes back cleaned in every response.
- [ ] **Step 2: Run** `/home/claude/scratch/gotest.sh ./internal/core/themis`. Expected: FAIL, undefined.
- [ ] **Step 3: Implement** per the interfaces. `NormalizeName` is `strings.ToLower(hermes.Untrusted(s, 40))`. The module adds `Uncategorized` to the list result and sorts it; rename and merge validate and normalise names, de-duplicate sources, and map `ErrExists` and `ErrNotFound` to `hermes.Errorf` messages written for the assistant (static text plus sanitised names). Responses go through `hermes.Respond`. Tool descriptions tell the assistant to confirm with the user before `rename_category` and `merge_categories`.
- [ ] **Step 4: Run** the whole scratch suite and `gotest.sh vet`. Expected: all PASS.
- [ ] **Step 5: Commit** `feat(themis): managed categories and their tools`.

### Task 3: Migration and Postgres store (needs owner approval first)

**Files:**
- Create: `apps/api/internal/store/migrations/00004_categories.sql`, `apps/api/internal/store/categories.go`, `apps/api/internal/store/categories_test.go`

**Interfaces:**
- Consumes: Task 2 `themis.Store`, `themis.ErrExists`, `themis.ErrNotFound`, `themis.Uncategorized`.
- Produces: `*Postgres` implements `themis.Store` (add `var _ themis.Store = (*Postgres)(nil)`); package variable `categoryRefs` in `categories.go`: a fixed slice of `{table, label string}` for every table with a `category` column and `user_id`, currently `{"transactions", "spending"}`. PR B and C append to it. Table names come only from this slice, never from input.

- [ ] **Step 1: Write the gated test** `TestPostgresCategories` (skips unless `TEST_DATABASE_URL`; creates and removes its own user, as `TestPostgresImport` did in the reverted PR): seed transactions in categories `food`, `travel`, `uncategorized`; `ListCategories` shows `travel` unlisted before it is created and with the right `spending` counts; create, then rename `food`→`meals` updates the transactions and the list; rename onto an existing name returns `ErrExists`; merge `travel` plus an unlisted source into `meals` moves all rows and removes the source rows; merge into an unknown target returns `ErrNotFound` and changes nothing; another user's categories and transactions are untouched throughout.
- [ ] **Step 2: Write the migration** per Global Constraints, with `-- +goose Up` and `Down` (`DROP TABLE categories`).
- [ ] **Step 3: Implement** the four methods in `categories.go`. Rename and merge run in one pgx transaction and loop over `categoryRefs`. List is one query: the user's listed names full-joined with in-use counts, which marks `unlisted` rows.
- [ ] **Step 4: Validate the SQL and the backfill on a throwaway Postgres 16** (`/usr/lib/postgresql/16/bin`, as the postgres user, port 54329): apply migrations 00001-00004 with the goose `Up` sections, having first inserted a user and transactions in `food`, `travel` and `uncategorized` before running 00004's statements. Expected: `categories` holds `food` and `travel` for that user and no `uncategorized`; running the backfill twice adds nothing; the rename, merge and list statements from `categories.go` return the expected rows and counts.
- [ ] **Step 5: Run** `gofmt -l apps/api` (empty) and `gofmt -e` on the new files. Expected: ok. The package cannot be compiled here; CI will.
- [ ] **Step 6: Commit** `feat(store): categories table, backfill and store`.

### Task 4: `add_transaction` checks the category list

**Files:**
- Modify: `apps/api/internal/modules/plutus/plutus.go` (`Store` interface, `add`, `add_transaction` description, replace `normCategory` with `themis.NormalizeName`), `apps/api/internal/modules/plutus/plutus_test.go`
- Modify: `apps/api/internal/store/plutus.go` (Postgres `CategoryNames`)

**Interfaces:**
- Consumes: Task 2 `themis.NormalizeName`, `themis.Uncategorized`.
- Produces: `plutus.Store` gains `CategoryNames(ctx context.Context, userID string) ([]string, error)` (the user's listed names, without `uncategorized`).

- [ ] **Step 1: Write failing tests** and make the existing ones keep passing. First run `grep -n '"category"' apps/api/internal/modules/plutus/plutus_test.go` and seed `memStore` (new `names map[string][]string`, settable per user, `CategoryNames` reads it) in `newEnv` with every category the existing tests use, for both `alice` and `bob`. New tests:
  - `TestCategoryMustBeInTheList`: list `food`; `food` and `Food` and ` fo​od ` save as `food`; `travel` is refused and nothing is stored; the message contains `"travel"`, `food`, `create_category`, `omit` and `uncategorized`; a user with no categories gets a message saying they have none yet; omitting the category and passing `uncategorized` both save.
  - `TestThaiCategory`: with list `อาหาร`, a Thai category saves and lists back unchanged.
  - `TestRefusalListIsCapped`: 120 listed names produce a message naming at most 50 and ending `and 70 more`.
  - `TestCategoryListIsPerUser`: a name listed for `bob` is refused for `alice`.
- [ ] **Step 2: Run** the new tests (scratch harness). Expected: FAIL (`travel` is accepted today).
- [ ] **Step 3: Implement.** In `add`, after normalising: if the category is empty it becomes `Uncategorized`; if it is not `Uncategorized`, call `m.store.CategoryNames` and, when absent, return `hermes.Errorf` with the message above (names sanitised, list capped at 50). Update the tool description for `category` ("a name from the user's list; unknown names are refused; optional"). Add the Postgres `CategoryNames` (`SELECT name FROM categories WHERE user_id = $1::uuid ORDER BY name`) in `store/plutus.go`.
- [ ] **Step 4: Run** the whole scratch suite. Expected: all PASS, existing tests included.
- [ ] **Step 5: Commit** `feat(plutus): refuse categories that are not in the user's list`.

### Task 5: Wiring, docs and agent log (needs owner approval for the scopes)

**Files:**
- Modify: `apps/api/cmd/nemosyne-api/main.go` (the `Scopes` list near line 94, and `themis.New(db).Register(mcp)` next to `plutus.New(db, time.Now).Register(mcp)`), `README.md`, `docs/design.md`
- Create: `agent-logs/claude/<date>-shared-categories.md` (template in `agent-logs/README.md`)

- [ ] **Step 1: Get owner approval** for the two new scopes. Stop here if not given.
- [ ] **Step 2: Edit `main.go`:** add `{Name: themis.ScopeRead, Description: "Read your category list"}` and `{Name: themis.ScopeWrite, Description: "Create, rename and merge your categories"}`; register the module. Check with `gofmt -e` (cannot compile here).
- [ ] **Step 3: Docs.** `README.md`: the four new tools, the two scopes, and the note that existing connectors must be re-authorised to see them. `docs/design.md`: a Themis entry (categories and the managed-list rule) and the changed `add_transaction` behaviour.
- [ ] **Step 4: Agent log** with commands run and results, what could not be run (Go 1.26, the `store` and `cmd` packages, the database test), and the rulings.
- [ ] **Step 5: Verify:** `gofmt -l apps/api` empty; scratch suite and vet pass; links in edited docs exist. Commit `feat(themis): register category tools and scopes`; open the PR from the template, wait for CI.

---

## Self-review

- **Spec coverage (sections 4, 8, 9 for this PR):** list and rules and refusal (Tasks 2, 4); migration and backfill (3); rename/merge atomic (3, with the `categoryRefs` extension point); `unlisted` (2, 3); scopes and flags (2, 5); tests list items for categories (2, 3, 4). Payee rules and memory items are in plans B and C, which is why this plan updates only `transactions`.
- **Type consistency:** `Store`, `ErrExists`, `ErrNotFound`, `Category`, `NormalizeName` and `categoryRefs` are defined once and used under the same names; `Uses` is keyed by label (`spending` now, `memory` in plan C).
- **Proportion:** signatures, test names and assertions only; bodies are left to the implementer.
