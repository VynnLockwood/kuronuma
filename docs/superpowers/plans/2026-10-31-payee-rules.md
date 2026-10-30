# Payee Rules (PR B) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The user can teach Nemo what a payee is ("7-Eleven is a convenience store, category food"), and `add_transaction` files new spending under the matching rule's category when no category is given.

**Architecture:** Rules live in `plutus` (types, name normalisation, a pure matcher) over three new `Store` methods that the Postgres store implements first, so every commit compiles. `add_transaction` resolves its category as explicit, then rule, then `uncategorized`, and reports which. The new table is added to `categoryRefs`, so renaming or merging a category (PR A) moves the rules with it.

**Tech Stack:** Go standard library, existing `pgx` and `goose`. No new dependencies, no new scopes.

**Spec:** `docs/superpowers/specs/2026-10-31-memory-and-categories-design.md` (sections 5, 8 and 9). Plan B of four; A (shared categories) is merged, C (Atlas) and D (copy and docs) follow.

## Global Constraints

- Rules use the existing scopes: `set_payee_rule` and `delete_payee_rule` need `plutus:write` (delete is destructive); `list_payee_rules` needs `plutus:read` (read-only). No new scopes.
- `payee_norm` is the payee lowercased, trimmed and with runs of spaces collapsed (`strings.ToLower(hermes.Untrusted(s, 100))`), at least 2 characters; `payee` max 100, `description` max 200 characters; at most 500 rules per user.
- A rule's category must be in the user's list or be `uncategorized` when it is set (same refusal message as `add_transaction`).
- **Matching:** a rule applies when the normalised merchant **contains** `payee_norm`. The longest `payee_norm` (counted in characters, not bytes) wins; ties go to the most recently updated rule.
- **`add_transaction`:** an explicit category (validated against the list) always wins, including an explicit `uncategorized`; with none given, the matching rule's category; otherwise `uncategorized`. The response carries `category_source` (`given`, `payee_rule` or `default`) and `matched_payee` (the rule's payee) when a rule applied.
- Rules apply to new transactions only; saved records are never rewritten.
- Migration `00005_payee_rules.sql`, additive: table `payee_rules(id, user_id, payee, payee_norm, category, description, created_at, updated_at)`, unique `(user_id, payee_norm)`.
- Every query filters by the verified user id. Text returned goes through `hermes.Untrusted` and the response carries `hermes.UntrustedNote`. Nothing is logged except method, path and status.
- Verify per `AGENTS.md`: `gofmt -l .` empty, `go vet ./...`, `go test -race -count=1 ./...` from `apps/api`. Commits: Conventional Commits, authored as the repo owner (already configured), no AI trailers. Never push to `main`.

## Review Focus

- A short rule matches unrelated merchants (`go` matches `Google`): that is what "contains" means, so the response must name the rule that matched, and a 1-character payee is refused. Pinned in Task 3.
- An assistant that passes `uncategorized` "to be safe" switches rules off for that save. Explicit wins by spec; the tool description tells the assistant to omit the category when unsure. Pinned in Task 3.
- A category renamed or merged later: the rule follows it; if a rule's category has vanished anyway, saving falls back to `uncategorized` instead of failing. Pinned in Task 2 (database test and psql) and Task 3.
- Thai payees, and the longest-match rule counting characters rather than bytes (a 5-character Thai rule must lose to an 8-character ASCII rule even though it has more bytes). Pinned in Task 1.
- A runaway assistant creating rules: the 500 cap keeps every `add_transaction` cheap (all rules load on each save); updating an existing rule at the cap still works. Pinned in Task 3.

## Prerequisites (gate, not tasks)

- **Owner approval before Task 2:** new migration `00005_payee_rules.sql` (`AGENTS.md`). Ask in the plan review.
- **Go 1.26:** unreachable in this sandbox. Tasks 1 and 3 run in the scratch harness `/home/claude/scratch/gotest.sh` (already copies `hermes`, `themis`, `plutus`). `internal/store` and `cmd/` import pgx: gofmt only locally; CI compiled PR A's equivalent code, and is the compile check again. Say so in the PR.
- Work on branch `feat/payee-rules` from the branch holding this plan.

---

### Task 1: Payee names and the matcher (pure logic)

**Files:**
- Create: `apps/api/internal/modules/plutus/payee.go`, `apps/api/internal/modules/plutus/payee_test.go`

**Interfaces:**
- Consumes: `hermes.Untrusted`.
- Produces: `const maxPayeeLen = 100`, `minPayeeLen = 2`, `maxPayeeDescription = 200`, `maxPayeeRules = 500`; `type PayeeRule struct{ ID, Payee, PayeeNorm, Category, Description string; UpdatedAt time.Time }`; `func NormalizePayee(s string) string`; `func matchRule(rules []PayeeRule, merchant string) (PayeeRule, bool)`.

- [ ] **Step 1: Write failing tests** in `payee_test.go`:
  - `TestNormalizePayee`: `"  7-ELEVEN  Silom "`→`7-eleven silom`; `"7​-eleven"`→`7-eleven`; Thai unchanged; 150 Thai characters → exactly 100 characters (count runes); `"​ "`→`""`.
  - `TestMatchRule` (table, rules built with `PayeeNorm` already normalised): single rule matches `"7-ELEVEN  Silom"` (case and spacing ignored); no rule matches `"Grab"` → `ok` false; `7-eleven` and `7-eleven silom` both match, the longer wins; two rules of equal length that both match (`ab`, `bc` on `abc`) → the one with the later `UpdatedAt` wins; Thai rule `ร้านอาหาร` matches `ร้านอาหารครัวไทย`; **rune counting:** a 5-character Thai rule and an 8-character ASCII rule both contained in one merchant → the 8-character one wins; empty rule list → `ok` false; `go` matches `Google` (documented behaviour).
- [ ] **Step 2: Run** `/home/claude/scratch/gotest.sh ./internal/modules/plutus -run 'NormalizePayee|MatchRule'`. Expected: FAIL, undefined.
- [ ] **Step 3: Implement** in `payee.go`. `NormalizePayee` is `strings.ToLower(hermes.Untrusted(s, maxPayeeLen))`. `matchRule` normalises the merchant the same way, keeps the best containing rule by `utf8.RuneCountInString(PayeeNorm)` then `UpdatedAt` then `PayeeNorm` (so the result is deterministic).
- [ ] **Step 4: Run** the whole scratch suite. Expected: all PASS.
- [ ] **Step 5: Commit** `feat(plutus): payee names and rule matching`.

### Task 2: Migration and Postgres store (needs owner approval first)

**Files:**
- Create: `apps/api/internal/store/migrations/00005_payee_rules.sql`, `apps/api/internal/store/payee_rules.go`, `apps/api/internal/store/payee_rules_test.go`
- Modify: `apps/api/internal/store/categories.go` (append `{"payee_rules", "payee_rules"}` to `categoryRefs`)

**Interfaces:**
- Consumes: Task 1 `plutus.PayeeRule`.
- Produces: `*Postgres` methods (the `plutus.Store` interface grows in Task 3, so this task compiles on its own):
  - `PayeeRules(ctx context.Context, userID string) ([]plutus.PayeeRule, error)`
  - `SetPayeeRule(ctx context.Context, userID string, r plutus.PayeeRule) (rule plutus.PayeeRule, created bool, err error)` (upsert on `(user_id, payee_norm)`; refreshes `payee`, `category`, `description` and `updated_at`; `created` from `RETURNING ..., (xmax = 0)`)
  - `DeletePayeeRule(ctx context.Context, userID, payeeNorm string) (bool, error)`

- [ ] **Step 1: Write the gated test** `TestPostgresPayeeRules` (skips unless `TEST_DATABASE_URL`; own users, cleaned up, as `TestPostgresCategories` does): set returns `created` true, a second set with the same `payee_norm` returns `created` false with the new category and description and a later `updated_at`; `PayeeRules` returns it; delete returns true then false; another user's rules never appear or change; `themis` rename of the rule's category through `db.RenameCategory` moves the rule, merge likewise, and `ListCategories` reports `Uses["payee_rules"]` for it.
- [ ] **Step 2: Write the migration** per Global Constraints, with `Up` and `Down` (`DROP TABLE payee_rules`).
- [ ] **Step 3: Implement** the three methods and the `categoryRefs` entry. Table and column names are fixed strings; values are bound parameters.
- [ ] **Step 4: Validate the SQL on a throwaway Postgres 16** (`/usr/lib/postgresql/16/bin`, as the postgres user, port 54329): apply migrations 00001-00005; run the upsert twice and confirm `(xmax = 0)` is true then false and `updated_at` advances; confirm the unique constraint rejects a duplicate `(user_id, payee_norm)` insert outside the upsert; run the rename and merge statements from `categories.go` with `payee_rules` included and confirm the rule's category follows; confirm Down then Up works. Expected: all as described.
- [ ] **Step 5: Run** `gofmt -l apps/api` (empty) and `gofmt -e` on the new files. Cannot compile here; CI will.
- [ ] **Step 6: Commit** `feat(store): payee rules table and store`.

### Task 3: Tools and `add_transaction` (rules at work)

**Files:**
- Modify: `apps/api/internal/modules/plutus/plutus.go` (`Store` interface, three tools registered after `delete_transaction` in this order: `set_payee_rule`, `list_payee_rules`, `delete_payee_rule`; `add`; `category` property text), `apps/api/internal/modules/plutus/plutus_test.go` (`memStore`, `TestScopes`)
- Create: `apps/api/internal/modules/plutus/payee_tools_test.go`

**Interfaces:**
- Consumes: Task 1 types and `matchRule`; Task 2's Postgres methods (the `plutus.Store` interface gains exactly their signatures).
- Produces: `plutus.Store` gains `PayeeRules`, `SetPayeeRule`, `DeletePayeeRule` (signatures as in Task 2). Tools: `set_payee_rule {payee, category, description?}`, `list_payee_rules {}`, `delete_payee_rule {payee}`.

- [ ] **Step 1: Write failing tests** in `payee_tools_test.go`; extend `memStore` (per-user rules map; `UpdatedAt` from an increasing counter so ties are controllable) and update `TestScopes` (new expected tool lists and annotation rows: `list_payee_rules` read-only, `delete_payee_rule` destructive, `set_payee_rule` neither). Tests, through the real MCP server:
  - `TestPayeeRuleAppliesWhenNoCategory`: rule `7-Eleven`→`food`; saving merchant `7-ELEVEN Silom` with no category stores `food`, `category_source` is `payee_rule`, `matched_payee` is `7-Eleven`.
  - `TestExplicitCategoryBeatsRule`: explicit `transport` wins (`given`); explicit `uncategorized` also wins; an explicit category not in the list is refused as before.
  - `TestNoMatchIsDefault`: no rule → `uncategorized`, `category_source` `default`, no `matched_payee`.
  - `TestShortRuleMatchesAndNamesItself`: rule `go`→`bills`; merchant `Google` is filed under `bills` and the response names `go`.
  - `TestLongestRuleAndTies`: longer rule wins; equal-length matches go to the most recently set.
  - `TestSetPayeeRuleValidation`: 1-character and invisible-only payee refused; empty category refused; category not in the list refused with the list and `create_category`; `uncategorized` accepted; a 300-character description stored as 200; a Thai payee accepted.
  - `TestSetPayeeRuleUpdatesInPlace`: setting `7-ELEVEN` after `7-eleven` returns `status` `updated` and leaves one rule.
  - `TestRuleLimit`: 500 rules fill the user's quota, a 501st new payee is refused, updating an existing one still works.
  - `TestListAndDeletePayeeRules`: list is sorted by payee and shows payee, category, description; delete returns `deleted` then `not_found`; user B can neither see nor delete user A's rules.
  - `TestVanishedRuleCategoryFallsBack`: after a rule is set, the user's list loses that category; saving a matching transaction succeeds as `uncategorized`, `category_source` `default`.
  - `TestPayeeTextIsData`: control and zero-width characters in payee, category and description come back cleaned; responses carry the untrusted-text note.
- [ ] **Step 2: Run** the new tests (scratch harness). Expected: FAIL (tools and fields do not exist).
- [ ] **Step 3: Implement.** In `add`: if the cleaned category is non-empty it is validated and is `given`; otherwise load the user's rules, `matchRule` on the merchant, and when it matches and `checkCategory` accepts the rule's category use it (`payee_rule`, `matched_payee`), falling back to `uncategorized` (`default`) when the rule's category is refused with a `hermes.ToolError` (any other error is returned). Add the two response fields to the `saved` result only. Tool descriptions: `set_payee_rule` says to ask the user first and that rules affect only future transactions; `add_transaction`'s `category` text says that omitting it lets the user's payee rules choose and that an explicit `uncategorized` switches rules off. Cap and validation per Global Constraints; names normalised with `NormalizePayee` and `themis.NormalizeName`.
- [ ] **Step 4: Run** the whole scratch suite and `gotest.sh vet`. Expected: all PASS, existing tests included.
- [ ] **Step 5: Commit** `feat(plutus): payee rules tools and automatic categories`.

### Task 4: Docs, agent log, review and PR

**Files:**
- Modify: `README.md` (tool list), `docs/design.md` (Plutus section)
- Create: `agent-logs/claude/<date>-payee-rules.md` (template in `agent-logs/README.md`)

- [ ] **Step 1: Docs.** `README.md`: add `set_payee_rule`, `list_payee_rules`, `delete_payee_rule` to the tool list (no new scopes). `docs/design.md`: payee rules under Plutus (matching rule, precedence, new transactions only, 500 cap).
- [ ] **Step 2: Verify** `gofmt -l apps/api` empty; scratch suite and vet pass; links in edited docs exist. Commit `docs: describe payee rules`.
- [ ] **Step 3: Final whole-branch review** by a fresh reviewer on the most capable available model; fix Critical and Important findings test-first; ledger the rest as deferred minors.
- [ ] **Step 4: Agent log** (commands run and results, what could not be run, rulings, deferred minors), commit, push the branch and open the PR from the template, mark the migration for owner approval, wait for CI, correct the PR text with the real CI result.

---

## Self-review

- **Spec coverage (section 5, plus 8 and 9 for this PR):** table and migration (Task 2); normalisation, minimum length and caps (Tasks 1, 3); contains-matching, longest wins, ties (Task 1); precedence, `category_source`, `matched_payee` (Task 3); rules apply to new transactions only (Task 3 description and test of default); tools and scopes (Task 3 including `TestScopes`); rename and merge follow the rule's category (Task 2); the tests in section 9 for payee rules (Tasks 1 to 3).
- **Type consistency:** `PayeeRule`, `NormalizePayee`, `matchRule` and the three store method signatures are defined in Tasks 1 and 2 and used under the same names in Task 3.
- **Proportion:** signatures, test names and assertions only; bodies are left to the implementer.
