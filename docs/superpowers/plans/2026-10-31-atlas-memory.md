# Atlas Memory Items (PR C) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Nemo can remember tasks, to-dos and notes: one flexible memory type with status, priority, due date, category and tags, found again by keyword search plus filters.

**Architecture:** A new module `modules/atlas` (types, pure helpers, four MCP tools) over a `Store` interface that the Postgres store implements first, so every commit compiles. Categories are the shared list from PR A, checked through one helper that PR C moves into `themis` so `plutus` and `atlas` share the same refusal. The new table joins `categoryRefs`, so renaming or merging a category moves memories with it.

**Tech Stack:** Go standard library, existing `pgx` and `goose`. No new dependencies; no Postgres extension.

**Spec:** `docs/superpowers/specs/2026-10-31-memory-and-categories-design.md` (sections 3, 6, 8, 9). Plan C of four; A (categories) and B (payee rules) are merged; D (dashboard, landing and `/th` copy, docs) follows.

## Global Constraints

- Table `memories`: `kind` in (`task`, `todo`, `note`); `title` 1-200 characters; `body` up to 4000, default empty; `status` in (`open`, `done`, `archived`) default `open`; `priority` in (`low`, `normal`, `high`) default `normal`; optional `due_on` date; `category` a name from the user's list, default `uncategorized`; `tags` up to 10, each 1-30 characters, lowercase, de-duplicated; `created_at`, `updated_at`, `completed_at`.
- `completed_at` is set when status becomes `done` and cleared when it leaves `done`; editing other fields of a done item keeps it.
- Migration `00006_memories.sql`, additive: indexes on `(user_id, status, due_on)` and `(user_id, kind)`.
- Scopes (new, need owner approval): `atlas:read` ("Read your tasks, to-dos and notes") for `list_memories` (read-only); `atlas:write` ("Add, change and delete your tasks, to-dos and notes") for `add_memory`, `update_memory`, `delete_memory` (destructive).
- **Recall (`list_memories`):** filters `kind`, `status`, `category`, `tag`, `due_before`, `due_after` (both inclusive; undated items never match a due filter), `limit` (default 50, at most 200). `query`: split on whitespace into at most 5 terms of at most 50 characters; every term must appear, case-insensitively, as a **substring** of the title, body or any tag. Substring matching, not full-text search (it does not split Thai). `%`, `_` and `\` in a term are escaped and match literally.
- **Order:** open items first, then `due_on` ascending (no date last), then priority (high first), then newest.
- Category is checked against the user's list with the same refusal message as `add_transaction`; an unknown one is refused, not created.
- No duplicate detection (the spec has none): the tool descriptions tell the assistant to list before adding.
- Every query filters by the verified user id; no user id from arguments. Text returned goes through `hermes.Untrusted` and the response carries `hermes.UntrustedNote`. Nothing is logged except method, path and status.
- Verify per `AGENTS.md`: `gofmt -l .` empty, `go vet ./...`, `go test -race -count=1 ./...` from `apps/api`. Commits: Conventional Commits, authored as the repo owner (already configured), no AI trailers. Never push to `main`.

## Review Focus

- Searching in Thai finds a Thai substring, and `%`, `_`, `\` in a query match themselves rather than acting as wildcards. Pinned in Task 3 (real Postgres) and Task 4.
- Completing and reopening: `completed_at` set once, cleared on reopen, kept when a done item is edited; `due_on` can be changed, cleared (`""`) or left alone, and `2026-02-31` is refused. Pinned in Task 4.
- An update that changes nothing, a malformed or unknown id, and an item deleted between list and update behave sensibly (clear error or `not_found`, no crash). Pinned in Task 4.
- Tags: more than 10, over-long, duplicate-by-case and empty tags. Pinned in Task 2 and Task 4.
- A double-submitted `add_memory` and simultaneous updates: no error, no corrupt row (duplicates are allowed), clean under `-race`. Pinned in Task 4. A category renamed or merged later moves its memories (`uses.memory`). Pinned in Task 3.

## Prerequisites (gate, not tasks)

- **Owner approval before Task 3 and Task 5:** migration `00006_memories.sql`, and new scopes `atlas:read` and `atlas:write` in `main.go`. Ask in the plan review.
- **Go 1.26:** unreachable in this sandbox. Tasks 1, 2 and 4 run in `/home/claude/scratch/gotest.sh`; add `internal/modules/atlas` to its copy list when Task 2 creates it. `internal/store` and `cmd/` import pgx: gofmt only locally, and CI is the compile check (it compiled PR A and B's equivalent code).
- Work on branch `feat/atlas-memory` from the branch holding this plan.

---

### Task 1: One shared category check in `themis`

**Files:**
- Modify: `apps/api/internal/core/themis/themis.go`, `apps/api/internal/core/themis/themis_test.go`, `apps/api/internal/modules/plutus/plutus.go`

**Interfaces:**
- Produces: `func Check(category string, names []string) error` in `themis`: `nil` when `category` is `Uncategorized` or in `names`, otherwise the `hermes.Errorf` refusal that `plutus.checkCategory` builds today (same text, list capped at 50 then `and N more`, names sanitised).
- `plutus.checkCategory` keeps its signature and becomes: return nil for `Uncategorized`, load `m.store.CategoryNames`, return `themis.Check(category, names)`.

- [ ] **Step 1: Write failing tests** `TestCheck` in `themis_test.go`: `uncategorized` and a listed name pass; an unknown name returns a `hermes.ToolError` containing `"travel"`, the listed names, `create_category`, `reconnect` and `omit`; an empty list says `no categories yet`; 120 names show the first 50 and `and 70 more`; a Thai name passes when listed and its refusal shows the Thai names intact; names with control characters come back cleaned.
- [ ] **Step 2: Run** `/home/claude/scratch/gotest.sh ./internal/core/themis -run TestCheck`. Expected: FAIL, undefined.
- [ ] **Step 3: Implement** `Check` by moving the message-building code out of `plutus.checkCategory`; reduce `checkCategory` as above and delete the now-unused constant and imports there.
- [ ] **Step 4: Run** the whole scratch suite. Expected: PASS, every existing `plutus` category test unchanged.
- [ ] **Step 5: Commit** `refactor(themis): one category check for every module`.

### Task 2: Atlas types, store interface and pure helpers

**Files:**
- Create: `apps/api/internal/modules/atlas/atlas.go`, `apps/api/internal/modules/atlas/helpers.go`, `apps/api/internal/modules/atlas/helpers_test.go`

**Interfaces:**
- Produces in `atlas`:
  - `const ScopeRead = "atlas:read"`, `ScopeWrite = "atlas:write"`; kind, status and priority constants matching Global Constraints.
  - `type Memory struct{ ID, Kind, Title, Body, Status, Priority, Category string; Tags []string; DueOn, CompletedAt *time.Time; CreatedAt, UpdatedAt time.Time }`
  - `type NewMemory struct{ Kind, Title, Body, Priority, Category string; Tags []string; DueOn *time.Time }`
  - `type Update struct{ Title, Body, Status, Priority, Category *string; Tags *[]string; SetDue bool; DueOn *time.Time }` (`SetDue` true with nil `DueOn` clears the date)
  - `type Filter struct{ Terms []string; Kind, Status, Category, Tag string; DueBefore, DueAfter *time.Time; Limit int }`
  - `type Store interface`: `AddMemory(ctx, userID string, n NewMemory) (Memory, error)`; `ListMemories(ctx, userID string, f Filter) ([]Memory, error)`; `UpdateMemory(ctx, userID, id string, u Update) (m Memory, found bool, err error)` (sets `completed_at` per Global Constraints); `DeleteMemory(ctx, userID, id string) (bool, error)`; `CategoryNames(ctx, userID string) ([]string, error)`
  - `func NormalizeTags(in []string) ([]string, error)` (lowercase via `hermes.Untrusted(t, 30)`, empty dropped, de-duplicated in order; more than 10 after that is an error)
  - `func ParseTerms(q string) ([]string, error)` (whitespace split, each term cleaned and cut to 50 characters; empty query gives none; more than 5 terms is an error)
  - `func LikePattern(term string) string` (`%` + the term with `\`, `%`, `_` escaped with `\` + `%`)
  - `func Less(a, b Memory) bool` (the list order from Global Constraints; used by the in-memory test store, mirrored by SQL in Task 3)

- [ ] **Step 1: Write failing tests** `helpers_test.go`: `TestNormalizeTags` (`Work`/`work`/` WORK ` give one tag; an empty and an invisible-only tag are dropped; a 40-character tag is cut to 30 characters; Thai kept; 11 distinct tags is an error, 10 plus duplicates is fine; nil gives an empty list); `TestParseTerms` (`  buy   Milk `→`buy`,`Milk`; empty→none; 6 terms error; a 80-character term cut to 50; a Thai term kept); `TestLikePattern` (`50%_off\x` becomes `%50\%\_off\\x%`; plain text unchanged inside `%`); `TestLess` (open before done and archived; earlier due first; undated last; high before normal before low; newest first as the last tie-break).
- [ ] **Step 2: Run** `/home/claude/scratch/gotest.sh ./internal/modules/atlas`. Expected: FAIL, undefined.
- [ ] **Step 3: Implement** the types and helpers; no tools yet. Add `internal/modules/atlas` to the scratch harness copy list.
- [ ] **Step 4: Run** the whole scratch suite and `gotest.sh vet`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(atlas): memory types, tag and search helpers`.

### Task 3: Migration and Postgres store (needs owner approval first)

**Files:**
- Create: `apps/api/internal/store/migrations/00006_memories.sql`, `apps/api/internal/store/memories.go`, `apps/api/internal/store/memories_test.go`
- Modify: `apps/api/internal/store/categories.go` (append `{"memories", "memory"}` to `categoryRefs`; make `ListCategories` start every category with a zero for each label in `categoryRefs`), `apps/api/internal/core/themis/themis.go` (fix the stale comment on `Category.Uses`), `apps/api/internal/core/themis/themis_test.go` (the fake store zero-fills `payee_rules` and `memory` like the real one)

**Interfaces:**
- Consumes: Task 2 `atlas.Store` and types.
- Produces: `*Postgres` implements `atlas.Store` (add `var _ atlas.Store = (*Postgres)(nil)`). `CategoryNames` already exists on `*Postgres` from PR A.

- [ ] **Step 1: Write the gated test** `TestPostgresMemories` (skips unless `TEST_DATABASE_URL`; own users, cleaned up): add and list round trip including tags and a null `due_on`; each filter; `Terms` AND-ed across title, body and tags; a Thai substring found; `100%`, `a_b` and `c\d` found only by themselves; the list order from Task 2 on seeded rows; `limit`; update sets `completed_at` on `done`, clears it on `open`, keeps it when an unrelated field changes, and `SetDue` with nil clears the date; update and delete of another user's id return not found; deleting twice; rename and merge of a category move memories and `ListCategories` reports `Uses["memory"]`; every category carries all three `Uses` labels (zeros included); two goroutines adding at once both succeed.
- [ ] **Step 2: Write the migration** per Global Constraints (`CHECK`s on `kind`, `status`, `priority`; `tags text[] NOT NULL DEFAULT '{}'`), with `Up` and `Down` (`DROP TABLE memories`).
- [ ] **Step 3: Implement** the methods in `memories.go`. List builds its `WHERE` from fixed fragments with bound parameters; each term adds `(title ILIKE $n OR body ILIKE $n OR EXISTS (SELECT 1 FROM unnest(tags) t WHERE t ILIKE $n))` using `atlas.LikePattern` and `ESCAPE '\'`; `ORDER BY (status <> 'open'), due_on ASC NULLS LAST, CASE priority WHEN 'high' THEN 0 WHEN 'normal' THEN 1 ELSE 2 END, created_at DESC`. Update builds `SET` from the non-nil fields and handles `completed_at` in SQL (`CASE` on the new status).
- [ ] **Step 4: Validate the SQL on a throwaway Postgres 16** (`/usr/lib/postgresql/16/bin`, postgres user, port 54329): migrations 00001-00006 applied; the list statement against Thai text (`ร้านอาหาร`), `%`/`_`/`\` terms, `unnest(tags)`, the order clause and the `completed_at` `CASE`; the check constraints reject a bad kind, status and priority; Down then Up. Expected: results as the gated test describes.
- [ ] **Step 5: Run** `gofmt -l apps/api` (empty), `gofmt -e` on the new files, and the scratch suite. Cannot compile `store` here; CI will.
- [ ] **Step 6: Commit** `feat(store): memories table and store`.

### Task 4: The Atlas tools

**Files:**
- Create: `apps/api/internal/modules/atlas/tools.go`, `apps/api/internal/modules/atlas/tools_test.go`

**Interfaces:**
- Consumes: Task 1 `themis.Check`, `themis.NormalizeName`, `themis.Uncategorized`; Task 2 everything.
- Produces: `func New(store Store, now func() time.Time) *Module` and `func (m *Module) Register(s *hermes.Server)` registering, in this order: `add_memory`, `list_memories`, `update_memory`, `delete_memory`. Arguments: `add_memory {kind, title, body?, due_on?, priority?, category?, tags?}`; `list_memories {query?, kind?, status?, category?, tag?, due_before?, due_after?, limit?}`; `update_memory {id, title?, body?, status?, priority?, due_on?, category?, tags?}` (`due_on` `""` clears; at least one field required); `delete_memory {id}`. Results: `add_memory` and `update_memory` return `status` (`saved` or `updated` or `not_found`) and the item; `list_memories` returns `memories` and `count`; `delete_memory` returns `deleted` or `not_found`.

- [ ] **Step 1: Write failing tests** in `tools_test.go` through the real MCP server, with an in-memory store (per user, `UpdatedAt` and `CreatedAt` from a counter, list order via `atlas.Less`, substring matching on lower-cased text, `completed_at` handled like the real one):
  - `TestAddMemory`: defaults (`open`, `normal`, `uncategorized`); invalid kind, missing or invisible-only title refused; a 300-character title stored as 200 and a 5000-character body as 4000; tags normalised and capped (11 distinct refused); bad `due_on` (`28/10/2026`, `2026-02-31`, `1980-01-01`, `2101-01-01`) refused; an unknown category refused with the shared message (`create_category`, the list); a Thai title round-trips.
  - `TestListFiltersAndOrder`: each filter alone and combined; invalid `kind`/`status` refused; `limit` default 50, 500 clamped to 200; due filters inclusive and exclude undated items; the default order from Global Constraints.
  - `TestListQuery`: terms AND-ed across title, body and tags; case-insensitive; Thai substring; `100%`, `a_b`, `c\d` match literally; 6 terms refused; empty query lists everything.
  - `TestUpdateMemory`: single-field updates; `done` sets `completed_at`, reopening clears it, editing a done item keeps it; `due_on` changed, cleared with `""`, untouched when absent, invalid refused; tags replaced and `[]` clears; an unknown category refused; no fields refused; a well-formed unknown id returns `not_found`; a malformed id refused; another user's id returns `not_found` and changes nothing.
  - `TestDeleteMemory`: `deleted` then `not_found`; another user cannot delete.
  - `TestScopesAndFlags`: a token with only `atlas:read` sees only `list_memories` (read-only); `delete_memory` is destructive; a read-only token calling `add_memory` is refused naming `atlas:write`.
  - `TestUsersAreIsolated`, `TestTextIsData` (control and zero-width characters cleaned in every field and error; the untrusted-text note present), `TestConcurrentUse` (two goroutines each add the same item and update and delete others; no error, no panic, clean under `-race`, duplicates allowed).
- [ ] **Step 2: Run** `/home/claude/scratch/gotest.sh ./internal/modules/atlas`. Expected: FAIL, undefined.
- [ ] **Step 3: Implement** per the interfaces. Tool descriptions: ask the user before saving or deleting; list before adding to avoid duplicates; say what `kind` means (`task` work item, `todo` personal checklist item, `note` anything to remember); `update_memory` is how an item is completed (`status: done`) or archived; due filters are inclusive.
- [ ] **Step 4: Run** the whole scratch suite with `-race` and `gotest.sh vet`. Expected: all PASS.
- [ ] **Step 5: Commit** `feat(atlas): add, list, update and delete memories`.

### Task 5: Wiring, docs, review and PR (needs owner approval for the scopes)

**Files:**
- Modify: `apps/api/cmd/nemosyne-api/main.go` (two scopes in the `Scopes` list, `atlas.New(db, time.Now).Register(mcp)`), `README.md`, `docs/design.md`
- Create: `agent-logs/claude/<date>-atlas-memory.md`

- [ ] **Step 1: Get owner approval** for `atlas:read` and `atlas:write`. Stop here if not given.
- [ ] **Step 2: Edit `main.go`** (consent text from Global Constraints) and check with `gofmt -e`; it cannot be compiled here.
- [ ] **Step 3: Docs.** `README.md`: the four tools, the two scopes, and that existing connectors must be re-authorised to see them. `docs/design.md`: an Atlas section (kinds, status, due date, tags, search rule and why it is substring, order, categories shared) and the module table. Do not change the landing or dashboard copy (that is PR D).
- [ ] **Step 4: Verify** `gofmt -l apps/api` empty; scratch suite and vet pass; links in edited docs exist. Commit `feat(atlas): register memory tools and scopes`.
- [ ] **Step 5: Final whole-branch review** by a fresh reviewer on the most capable available model; fix Critical and Important findings test-first; ledger the rest as deferred minors.
- [ ] **Step 6: Agent log**, commit, push, open the PR from the template with the migration and scopes marked for owner approval, wait for CI, correct the PR text with the real CI result.

---

## Self-review

- **Spec coverage (sections 6, 8, 9 for this PR):** table, constraints and migration (Task 3); kinds, status, priority, due date, category, tags, `completed_at` rule (Tasks 2-4); four tools and flags (Task 4); recall, substring search, literal wildcards, order, limit (Tasks 2-4); shared category check and refusal (Task 1, Task 4); scopes and consent text (Task 5); the memory tests in section 9 including Thai, wildcards, concurrency, isolation and `Untrusted` (Tasks 3-4); category rename and merge moving memories (Task 3). Web and copy are PR D, as the spec says.
- **Type consistency:** `Memory`, `NewMemory`, `Update`, `Filter`, `Store`, `NormalizeTags`, `ParseTerms`, `LikePattern`, `Less` are defined in Task 2 and used under the same names in Tasks 3 and 4; `themis.Check` is defined in Task 1 and used in Task 4.
- **Proportion:** signatures, test names and assertions only; bodies are left to the implementer.
