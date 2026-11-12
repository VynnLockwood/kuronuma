# Fixed Monthly Bills Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The user can record fixed monthly bills (subscriptions, rent, mobile), see each month's planned and paid totals per currency, and confirm a month's payment, which creates a linked spending record. Every action is an MCP tool, and the dashboard calls the same operations.

**Architecture:** Bills are a second kind of record inside `plutus`. One new table holds the bills. A payment is an ordinary transaction whose `reference` is `bill:<bill_id>:<YYYY-MM>`, so the existing unique index blocks double confirmation and "paid" means that transaction exists. Pure rules (counting, references, totals) live in `bills.go` with no database; the store gains bill methods; seven tools are registered beside the existing spending tools.

**Tech Stack:** Go standard library, existing `pgx` and `goose`, existing Next.js dashboard. No new dependencies, no new scopes.

**Spec:** `docs/superpowers/specs/2026-11-12-fixed-bills-design.md` (approved 2026-11-12; still uncommitted in the working tree, so it is the first commit on the feature branch).

## Global Constraints

- Branch `feat/fixed-bills` from the latest `main`. Never push to `main`. Open a PR and wait for green CI (`AGENTS.md`).
- Conventional Commits: `feat(plutus): ...`, `feat(web): ...`, `docs: ...`. Author as the repo owner. **No AI `Co-Authored-By` trailers** (`AGENTS.md` overrides the generic reminder). Agent log committed in the same PR.
- Scopes: `plutus:read` for `list_bills` and `monthly_bills_summary`; `plutus:write` for the other five. `delete_bill` is marked `Destructive`. No new scope constants.
- Amounts: `bigint` in hundredths, parsed with the existing `parseMinor` (plain positive number, at most 2 decimals, refused rather than rounded). Currency matches `reCurrency` and is fixed at creation (`update_bill` cannot change it).
- Names: free text, cleaned with `hermes.Untrusted(s, 100)`; empty after cleaning is refused. Categories go through `themis` (`NormalizeName`, `Ensure`, `Check`).
- Months: tools take `YYYY-MM`; the store and the pure code use the first day of the month in UTC. Storage column `start_month date`, `ended_from date NULL`.
- Counting rule: a bill counts in month M when `start_month <= M` and (`ended_from` is NULL or `M < ended_from`).
- Payment reference: exactly `bill:<bill_id>:<YYYY-MM>`. Paid means a transaction with that reference exists for the user.
- Cap: 100 bills per user. Message: `Sorry, the user already has %d bills, the most Kuronuma keeps. End or delete one first.`
- Every query filters by the verified user id. Text returned goes through `hermes.Untrusted`, with `hermes.UntrustedNote` on the response, as the other tools do.
- Verify before every commit (`AGENTS.md`), from `apps/api`: `gofmt -l .` prints nothing, `go vet ./...`, `go test -race -count=1 ./...`. Store database tests run with `TEST_DATABASE_URL` set against a scratch Postgres. Web: type-check, `npm run check:design`, browser smoke test (Task 5).

**Decisions beyond the approved spec.** These fill gaps the spec leaves. Confirm them before implementation:

- **D1.** `confirm_bill_payment` with no date uses today if today is in the confirmed month. If it isn't, the call is refused and asks for a date. This resolves the spec's clash between "default today" and "date must fall in the month."
- **D2.** Confirming a month the bill does not count in (before `start_month`, or at/after `ended_from`) is refused with a message saying so.
- **D3.** `end_bill` refuses an `ended_from` earlier than `start_month`. `update_bill` refuses a new `start_month` later than an existing `ended_from`.
- **D4.** `confirm_bill_payment` applies the per-user transaction cap (`maxTransactions`, 50,000), as `add_transaction` does.
- **D5.** `update_bill` takes `clear_ended` (boolean) to undo an ending; `end_bill` takes `ended_from`. The two are separate tools, as the spec's table shows.

## Review Focus

- A second confirmation for the same bill and month (double click, retry) must not record twice. Pinned in Task 4 (`TestConfirmTwiceRefused`).
- Deleting the spending record for a payment makes the month unpaid again, and confirming then works. Pinned in Task 4 (`TestDeletedPaymentMakesMonthUnpaid`).
- Amounts such as `12.345` or `-5` on a payment are refused, not rounded or stored. Pinned in Task 4 (`TestConfirmAmountTooPreciseRefused`).
- Month edges: a bill ending in January is counted through December; December to January rolls the year; February in a leap year. Pinned in Task 1 (`TestCountsMonthEdges`).
- A category renamed or merged after a bill is saved: the bill and its confirmed payments follow the new name, and a payment confirmed after the rename uses it. Pinned in Task 2 (`TestPostgresBillsFollowCategoryRename`).

## File Structure

- Create `apps/api/internal/modules/plutus/bills.go`: types, `MonthStart`, `ParseMonth`, `Counts`, `BillReference`, `Totals`. Pure; no I/O.
- Create `apps/api/internal/modules/plutus/bills_test.go`: unit tests for the pure rules.
- Create `apps/api/internal/modules/plutus/bill_tools.go`: `registerBillTools` and the seven handlers, plus output shaping.
- Create `apps/api/internal/modules/plutus/bill_tools_test.go`: tool tests against the in-memory fake.
- Modify `apps/api/internal/modules/plutus/plutus.go`: extend `Store` with bill methods; call `m.registerBillTools(s)` in `Register`.
- Modify `apps/api/internal/modules/plutus/plutus_test.go`: extend `memStore` so it satisfies the new `Store` methods.
- Create `apps/api/internal/store/migrations/00007_bills.sql`.
- Create `apps/api/internal/store/bills.go`: the Postgres implementation.
- Create `apps/api/internal/store/bills_test.go`: database tests, skipped without `TEST_DATABASE_URL`.
- Modify `apps/api/internal/store/categories.go`: add `bills` to `categoryRefs`.
- Create `apps/web/components/dashboard/Bills.tsx`; modify `apps/web/components/dashboard/Spending.tsx` (add the tab) and `apps/web/lib/dashboard-copy.ts` (English and Thai strings).
- Modify `README.md`, `docs/design.md`; create `agent-logs/claude/2026-11-12-fixed-bills.md`.

---

### Task 1: Pure bill rules

**Files:**
- Create: `apps/api/internal/modules/plutus/bills.go`
- Test: `apps/api/internal/modules/plutus/bills_test.go`

**Interfaces:**
- Produces:
  - `type Bill struct { ID, Name string; AmountMinor int64; Currency, Category string; StartMonth time.Time; EndedFrom *time.Time }`
  - `type NewBill struct { Name string; AmountMinor int64; Currency, Category string; StartMonth time.Time }`
  - `type BillPatch struct { Name, Category *string; AmountMinor *int64; StartMonth, EndedFrom *time.Time; ClearEnded bool }`
  - `type BillTotal struct { Currency string; PlannedMinor, PaidMinor int64; Paid, Unpaid int }`
  - `func MonthStart(t time.Time) time.Time`: first day of the month, UTC.
  - `func ParseMonth(s string) (time.Time, error)`: accepts only `YYYY-MM`.
  - `func Counts(b Bill, month time.Time) bool`
  - `func BillReference(billID string, month time.Time) string`: `bill:<id>:<YYYY-MM>`.
  - `func Totals(bills []Bill, month time.Time, payments map[string]Transaction) []BillTotal`: `payments` is keyed by reference; result sorted by currency; only bills that count in `month`; a bill is paid when its reference is present, and its paid amount is that transaction's amount.

- [ ] **Step 1: Write the failing tests**

```go
func TestCountsMonthEdges(t *testing.T) {
	// Starts in Nov 2026, ends from Jan 2027 (so counts Nov, Dec; not Jan).
	b := Bill{StartMonth: m(2026, 11), EndedFrom: ptr(m(2027, 1))}
	// assert: Counts(b, m(2026,10)) == false; Nov true; Dec true; Jan false
	// Leap year: a bill starting 2028-02 counts in 2028-02 and not 2028-01.
	// Year roll: ended 2027-01 counted in 2026-12.
}

func TestParseMonthRejectsOtherForms(t *testing.T) {
	// accepts "2026-11" -> 2026-11-01 UTC
	// refuses "2026-11-05", "11/2026", "2026-13", "" with an error
}

func TestBillReference(t *testing.T) {
	// BillReference("3f2a...", m(2026,11)) == "bill:3f2a...:2026-11"
}

func TestTotalsPerCurrencyAndPaid(t *testing.T) {
	// Two THB bills (rent 12000.00 and mobile 399.00), one USD bill (9.99), one THB bill ended before Nov (ignored).
	// payments: reference of the mobile bill for Nov has amount 350.00.
	// want: THB {Planned 1239900, Paid 35000, Paid 1, Unpaid 1}; USD {Planned 999, Paid 0, Unpaid 1}, sorted THB before USD.
}

func TestTotalsEmpty(t *testing.T) {
	// no bills -> empty slice, not an error
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd apps/api && go test ./internal/modules/plutus/ -run 'TestCounts|TestParseMonth|TestBillReference|TestTotals' -v`
Expected: FAIL, undefined `Counts`, `ParseMonth`, `BillReference`, `Totals`.

- [ ] **Step 3: Implement the types and pure functions in `bills.go`**

Use `time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)` for `MonthStart`. `ParseMonth` uses `time.Parse("2006-01", s)` and also checks the string length is 7. `Totals` builds a map keyed by currency, then sorts. `Counts` compares month starts with `Before`/`Equal`. No database calls.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd apps/api && go test ./internal/modules/plutus/ -run 'TestCounts|TestParseMonth|TestBillReference|TestTotals' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/modules/plutus/bills.go apps/api/internal/modules/plutus/bills_test.go
git commit -m "feat(plutus): add the fixed-bill counting and totals rules"
```

---

### Task 2: Storage for bills

**Files:**
- Create: `apps/api/internal/store/migrations/00007_bills.sql`
- Create: `apps/api/internal/store/bills.go`
- Test: `apps/api/internal/store/bills_test.go`
- Modify: `apps/api/internal/modules/plutus/plutus.go` (Store interface), `apps/api/internal/modules/plutus/plutus_test.go` (memStore), `apps/api/internal/store/categories.go` (`categoryRefs`)

**Interfaces:**
- Consumes: `Bill`, `NewBill`, `BillPatch`, `Transaction` from Task 1.
- Produces, added to `plutus.Store`:
  - `AddBill(ctx context.Context, userID string, b NewBill) (Bill, error)`
  - `GetBill(ctx context.Context, userID, id string) (Bill, bool, error)`
  - `ListBills(ctx context.Context, userID string) ([]Bill, error)`: all of the user's bills, ordered by name, then id.
  - `UpdateBill(ctx context.Context, userID, id string, p BillPatch) (Bill, bool, error)`
  - `DeleteBill(ctx context.Context, userID, id string) (bool, error)`
  - `CountBills(ctx context.Context, userID string) (int, error)`
  - `HasPayments(ctx context.Context, userID, billID string) (bool, error)`: true when a transaction with reference prefix `bill:<billID>:` exists.
  - `PaymentsByReference(ctx context.Context, userID string, refs []string) (map[string]Transaction, error)`: only the listed references; an empty `refs` returns an empty map without querying.

- [ ] **Step 1: Write the migration**

`00007_bills.sql`, following `00005_payee_rules.sql`:

```sql
-- +goose Up
CREATE TABLE bills (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         text NOT NULL,
    amount_minor bigint NOT NULL CHECK (amount_minor > 0),
    currency     text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    category     text NOT NULL,
    start_month  date NOT NULL,
    ended_from   date,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CHECK (ended_from IS NULL OR ended_from >= start_month)
);
CREATE INDEX bills_user_idx ON bills (user_id);

-- +goose Down
DROP TABLE bills;
```

- [ ] **Step 2: Write the failing database test**

`store/bills_test.go`, skipped without `TEST_DATABASE_URL`, following `payee_rules_test.go`'s helper pattern (create two users, clean up with `t.Cleanup`):
- `TestPostgresBills`: add a bill; get it; list it; update its amount (past payments untouched); `UpdateBill` with `ClearEnded` after an end; delete it when it has no payments.
- Payments: after `AddTransaction` with `plutus.BillReference(id, month)` as the reference, `HasPayments` is true, `PaymentsByReference` returns that record only for requested references, and deleting the bill is refused by the caller (the store just reports `HasPayments`; the refusal lives in Task 3).
- `alice` cannot get, update or delete `bob`'s bill (returns not found / false).
- `TestPostgresBillsFollowCategoryRename`: create bill in category `food`, confirm a payment, rename `food` with the existing rename method in `categories.go`, then `GetBill` shows the new name and the transaction's category matches.

- [ ] **Step 3: Run the test to verify it fails**

Run: `cd apps/api && TEST_DATABASE_URL=<scratch db> go test ./internal/store/ -run TestPostgresBills -v`
Expected: FAIL to compile (methods undefined). Without `TEST_DATABASE_URL` the tests skip, which is not a pass; the database run is required.

- [ ] **Step 4: Extend the `Store` interface, the test fake, and implement Postgres**

- In `plutus.go`, add the eight methods above to `Store`.
- In `plutus_test.go`, extend `memStore` with a `bills` map and matching methods so the existing tests keep compiling. Its `PaymentsByReference` reads the existing `txs` records by reference.
- In `store/bills.go`, implement them on `*Postgres`. Use `$1::uuid`, scan `start_month` and `ended_from` as dates, and return `(Bill{}, false, nil)` on `pgx.ErrNoRows`. `UpdateBill` builds `UPDATE ... SET updated_at = now()` with only the provided columns, and returns `found=false` when no row matches. `HasPayments` uses `EXISTS (SELECT 1 FROM transactions WHERE user_id = $1::uuid AND reference LIKE 'bill:' || $2 || ':%')`.
- `var _ plutus.Store = (*Postgres)(nil)` must still compile.
- Add `bills` with label `bills` to `categoryRefs` in `categories.go`, so rename and merge update it.

- [ ] **Step 5: Run everything, with the database**

Run: `cd apps/api && gofmt -l . && go vet ./... && TEST_DATABASE_URL=<scratch db> go test -race -count=1 ./...`
Expected: `gofmt` prints nothing; all tests PASS, including `TestPostgresBills` and `TestPostgresBillsFollowCategoryRename`.

Also check the migration both ways: run goose down one step, then up again against the scratch database. Expected: both succeed.

- [ ] **Step 6: Commit**

```bash
git add apps/api/internal/store/migrations/00007_bills.sql apps/api/internal/store/bills.go apps/api/internal/store/bills_test.go apps/api/internal/store/categories.go apps/api/internal/modules/plutus/plutus.go apps/api/internal/modules/plutus/plutus_test.go
git commit -m "feat(plutus): store fixed monthly bills"
```

---

### Task 3: Bill management tools

**Files:**
- Create: `apps/api/internal/modules/plutus/bill_tools.go`
- Test: `apps/api/internal/modules/plutus/bill_tools_test.go`
- Modify: `apps/api/internal/modules/plutus/plutus.go` (`Register` calls `m.registerBillTools(s)`)

**Interfaces:**
- Consumes: Store bill methods from Task 2; `Counts`, `ParseMonth`, `MonthStart` from Task 1; `parseMinor`, `formatMinor`, `parseDate`, `reCurrency`, `reUUID`, `m.today()`, `themis.NormalizeName`, `themis.Ensure`, `themis.Check`.
- Produces: `func (m *Module) registerBillTools(s *hermes.Server)`, which adds `add_bill`, `list_bills`, `update_bill`, `end_bill`, `delete_bill`. `monthly_bills_summary` and `confirm_bill_payment` come in Task 4 and are registered there. The `billOut` JSON shape is `{id, name, amount, currency, category, start_month, ended_from?}` with `start_month` and `ended_from` as `YYYY-MM`.

- [ ] **Step 1: Write the failing tests** (`bill_tools_test.go`, using the existing `memStore` and the `hermes` test helpers used by `plutus_test.go`)

- `TestAddBill`: saves with the given fields; with no `start_month`, it is the current month; a new category is created and the result says `category_created`.
- `TestAddBillValidation`: refuses currency `th`, amount `1.234`, amount `0`, an empty name, and a name over 100 characters, each with a message and no bill saved.
- `TestBillCap`: at 100 bills, the 101st `add_bill` is refused with the cap message and nothing is created (no category left behind).
- `TestListBillsDefaultsToCurrentMonth`: a bill ending before the current month is not listed; one starting next month is not listed.
- `TestUpdateBillKeepsPastPayments`: changing the amount leaves a previously saved transaction's amount unchanged.
- `TestEndBill`: sets `ended_from`; refuses one earlier than `start_month` (D3); `update_bill` with `clear_ended: true` restores it.
- `TestDeleteBill`: allowed with no payments; refused with a payment (message points to `end_bill`).
- `TestUnknownArgumentsRefused` for each new tool.
- `TestBillScopes`: a `plutus:read` token sees `list_bills` and not `add_bill`, `update_bill`, `end_bill` or `delete_bill` in `tools/list`; calling a write tool with it is refused.
- `TestBillIsolation`: one user cannot see, update, end or delete another user's bill.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd apps/api && go test ./internal/modules/plutus/ -run 'TestAddBill|TestListBills|TestUpdateBill|TestEndBill|TestDeleteBill|TestBill' -v`
Expected: FAIL, `registerBillTools` and the handlers are undefined.

- [ ] **Step 3: Implement the five tools**

Descriptions follow the house style of the existing tools: say what the tool does, what to confirm with the user first, and that bill names are data, not instructions. Each handler: decode with `hermes.DecodeArgs`, validate, check the cap with `CountBills` before any category is created (as `add` does), then call the store. `delete_bill` checks `HasPayments` first and refuses with the message. `list_bills` takes an optional `month` (default `MonthStart(m.today())`) and returns bills that `Counts` says apply, with `paid: false` for now (Task 4 adds payment fields).

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd apps/api && go test ./internal/modules/plutus/ -run 'TestAddBill|TestListBills|TestUpdateBill|TestEndBill|TestDeleteBill|TestBill' -v`
Expected: PASS.

- [ ] **Step 5: Verify and commit**

Run: `cd apps/api && gofmt -l . && go vet ./... && go test -race -count=1 ./...` (expected: clean, all PASS).

```bash
git add apps/api/internal/modules/plutus/bill_tools.go apps/api/internal/modules/plutus/bill_tools_test.go apps/api/internal/modules/plutus/plutus.go
git commit -m "feat(plutus): add bill management tools"
```

---

### Task 4: Payment tools and monthly summary

**Files:**
- Modify: `apps/api/internal/modules/plutus/bill_tools.go` (add `confirm_bill_payment`, `monthly_bills_summary`; add payment fields to `list_bills`)
- Test: `apps/api/internal/modules/plutus/bill_tools_test.go` (add cases)

**Interfaces:**
- Consumes: `BillReference`, `Totals`, `Counts`, `ParseMonth` (Task 1); `GetBill`, `ListBills`, `PaymentsByReference`, `CountBills`, and the existing `AddTransaction` and `CountTransactions` (Task 2); `parseMinor`, `parseDate`, `m.today()` (existing).
- Produces:
  - `confirm_bill_payment` (`plutus:write`): inputs `bill_id` (required, UUID), `month` (default current month), `amount` (optional, planned if absent), `date` (optional, D1 rule). Saves via `AddTransaction` with merchant = bill name, category = bill category, reference `BillReference`. A repeat returns `status: "already_paid"` with the existing record and saves nothing. Result: `{status: "saved", transaction, bill}`.
  - `monthly_bills_summary` (`plutus:read`): input `month` (default current). Output `{month, totals: [{currency, planned, paid, paid_count, unpaid_count}]}`, amounts as strings via `formatMinor`.
  - `list_bills` gains `paid`, `paid_amount`, `paid_on` for the requested month, and `has_payments` (true when any payment exists for the bill in any month, from `HasPayments`).

- [ ] **Step 1: Write the failing tests** (add to `bill_tools_test.go`)

- `TestConfirmBillPayment`: default amount is the planned amount; merchant is the bill name; category is the bill's category; reference is `BillReference(id, month)`; with no date and today in the month, the date is today.
- `TestConfirmTwiceRefused`: the second confirmation returns `already_paid` and no second transaction exists (`TestConfirmTwiceRefused` is one of the Review Focus pins).
- `TestConfirmOutsideCountingMonthRefused` (D2): before `start_month` and at/after `ended_from`.
- `TestConfirmDateOutsideMonthRefused`: a date in another month is refused.
- `TestConfirmNoDateOutsideCurrentMonthAsksForDate` (D1): confirming a past month with no date is refused and asks for a date.
- `TestConfirmAmountTooPreciseRefused`: `12.345` and `-5` are refused; nothing saved (Review Focus pin).
- `TestConfirmAppliesTransactionCap` (D4).
- `TestDeletedPaymentMakesMonthUnpaid`: delete the transaction with `delete_transaction`; `list_bills` shows `paid: false`; a new confirmation succeeds (Review Focus pin).
- `TestMonthlySummaryPerCurrency`: two currencies; planned and paid totals and counts match the spec; no mixing.
- `TestConfirmUsersIsolation`: another user cannot confirm or see the payment.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd apps/api && go test ./internal/modules/plutus/ -run 'TestConfirm|TestMonthlySummary|TestDeletedPayment' -v`
Expected: FAIL, handlers undefined.

- [ ] **Step 3: Implement the two tools and the `list_bills` payment fields**

Confirm order: decode; resolve the bill (not found → refusal); `Counts` check (D2); cap check (D4); amount via `parseMinor` or planned; date rules (D1); `AddTransaction`; `created == false` → `already_paid`. Do not call `FindPossibleDuplicate`: the reference already prevents double recording. `monthly_bills_summary` uses `ListBills`, `PaymentsByReference` for the month's references, and `Totals`.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd apps/api && go test ./internal/modules/plutus/ -v -run 'TestConfirm|TestMonthlySummary|TestDeletedPayment|TestListBills'`
Expected: PASS.

- [ ] **Step 5: Verify and commit**

Run: `cd apps/api && gofmt -l . && go vet ./... && go test -race -count=1 ./...`

```bash
git add apps/api/internal/modules/plutus/bill_tools.go apps/api/internal/modules/plutus/bill_tools_test.go
git commit -m "feat(plutus): confirm bill payments and summarise monthly bills"
```

---

### Task 5: Dashboard Bills tab

**Files:**
- Create: `apps/web/components/dashboard/Bills.tsx`
- Modify: `apps/web/components/dashboard/Spending.tsx` (add the tab), `apps/web/lib/dashboard-copy.ts` (English and Thai strings)

**Interfaces:**
- Consumes: the web bridge (`/api/web/call`, via `apps/web/lib/web.ts` and the `useTool` hook) calling `list_bills`, `monthly_bills_summary`, `confirm_bill_payment`, `add_bill`, `update_bill`, `end_bill`, `delete_bill`.
- Produces: a Bills tab with a month selector, per-currency totals at the top, a list of bills with Paid/Unpaid status, a Mark paid button, and add, edit, end and delete actions. Delete appears only for bills whose `list_bills` entry has `has_payments: false` (Task 4 adds that field).

- [ ] **Step 1: Read the existing pattern**

Read `components/dashboard/Spending.tsx` and `useTool.ts` and follow the same tab, state and error-message pattern. Do not add a new fetch layer.

- [ ] **Step 2: Build the tab**

`Bills.tsx` uses existing components (`Card`, `Button`, `Badge`, `Input`, `Alert`), design tokens only, no gradients, no decorative motifs. Strings go in `dashboard-copy.ts` in both languages. Server refusals are shown as returned; they are already written for people.

- [ ] **Step 3: Verify**

Run from `apps/web`: the type-check script in `package.json`, then `npm run check:design`. Expected: both pass.

Browser smoke test against a running stack (`docker compose -f deploy/docker-compose.yml --env-file .env up --build`, sign in): add a bill, mark it paid, confirm the spending record appears on the Spending tab with the bill's name and category, mark it again and see "already paid", end the bill and see it leave the next month's list. Expected: each step behaves as described.

- [ ] **Step 4: Commit**

```bash
git add apps/web/components/dashboard/Bills.tsx apps/web/components/dashboard/Spending.tsx apps/web/lib/dashboard-copy.ts
git commit -m "feat(web): add the Bills tab to the spending dashboard"
```

---

### Task 6: Documentation and agent log

**Files:**
- Modify: `README.md` (the "Working now" list: spending gains fixed monthly bills; the MCP tool count goes from 15 to 22)
- Modify: `docs/design.md` (a bills subsection under Plutus; the build status)
- Create: `agent-logs/claude/2026-11-12-fixed-bills.md`

- [ ] **Step 1: Update the docs** so they describe only what is built. Count the tools in the MCP registration to get the number, rather than trusting the arithmetic.

- [ ] **Step 2: Write the agent log**, following `agent-logs/README.md`: goal, what was done, what was verified (with the commands and results), gaps (e.g. database tests need a scratch Postgres), decisions D1 to D5 and the user's answers.

- [ ] **Step 3: Commit**

```bash
git add README.md docs/design.md agent-logs/claude/2026-11-12-fixed-bills.md
git commit -m "docs: describe fixed monthly bills and log the work"
```

---

## Self-Review

- **Spec coverage:** data model (Task 2), counting and totals (Task 1), payments as linked transactions (Task 4), seven tools with their scopes (Tasks 3 and 4), amount change and ending rules (Tasks 3 and 4), refusal of delete with payments (Task 3), cap of 100 (Task 3), dashboard tab (Task 5), tests listed in the spec (Tasks 1 to 4), docs and log (Task 6). The spec's three "choices made" are carried as checks in Task 4 (date in month: Task 4 `TestConfirmDateOutsideMonthRefused`; `uncategorized` default: Task 3; first-day storage: Global Constraints).
- **Type consistency:** `Bill`, `NewBill`, `BillPatch`, `BillTotal` are defined in Task 1 and used unchanged in Tasks 2 to 4. The store method names are fixed in Task 2 and used by name in Tasks 3 and 4. `registerBillTools` is defined in Task 3 and called from `Register`.
- **Gaps to confirm:** D1 to D5 above. The spec's "date defaults to today" needs the D1 rule to be workable.

**Execution handoff.** The plan is saved at `docs/superpowers/plans/2026-11-12-fixed-bills.md`. Please review it. Confirm D1 to D5, or tell me what to change. Then choose how to build it:

- **Subagent-driven**: a fresh subagent implements each task, and a fresh reviewer checks it before the next starts, with a whole-branch review at the end.
- **Native**: I implement every task myself in this session, then one fresh reviewer checks the whole branch.

I recommend **subagent-driven**. The tasks depend on each other's interfaces (Task 2 feeds Tasks 3 and 4), and this code runs on a live app that holds personal financial data, so an independent review at each task is worth the extra cost.

