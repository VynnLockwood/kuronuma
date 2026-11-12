# Fixed monthly bills

- Agent: claude
- Date: 2026-11-12
- Branch / PR: feat/fixed-bills (local; no push, no PR opened)

## Goal
Let the user record fixed monthly costs (rent, subscriptions, phone bills), see a monthly total per currency, and confirm each month's payment as a normal spending record. Works through MCP tools and a Bills tab on the spending dashboard. Spec: `docs/superpowers/specs/2026-11-12-fixed-bills-design.md`.

## What I did
Commits (`git log --oneline e1cfe16..HEAD`; the spec and plan commit is `e1cfe16`; `git log --oneline d526e2f..HEAD` was empty before this docs commit):
- `c3aa854` feat(plutus): add the fixed-bill counting and totals rules
- `ec37edf` feat(plutus): store fixed monthly bills (migration 00007, Postgres store)
- `d5581b9` feat(plutus): add bill management tools (add, list, update, end, delete)
- `126596c` feat(plutus): confirm bill payments and summarise monthly bills
- `1a507dd` feat(web): add the Bills tab to the spending dashboard
- `d526e2f` fix(web): let the bills tab record payments for other months
- This docs commit: README (tool count 15 to 22, bills in the working list), `docs/design.md` (Bills subsection, build status), this log.

Tool count: `grep -rn 'AddTool' apps/api --include=*.go | grep -v _test` finds 22 registrations (Atlas 4, Plutus 4 + payee rules 3 + bills 7, Themis 4).

## Verified
- `gofmt -l` on the Go changes: printed nothing (reported by each task).
- `npm` type-check and `npm run check:design` for the web: passed (reported by Task 5 and its fix round).
- Docs read against the code: each bills claim in README and `docs/design.md` checked against `bill_tools.go`, `bills.go`, `00007_bills.sql`, `store/categories.go` and `Bills.tsx`.
- Per-task reviews (spec and quality) found no Critical or Important issue after one fix round on Task 5 (Mark paid could not send a date for other months).

## Not verified / known gaps
All of these are UNVERIFIED because the sandbox has no Go 1.26 toolchain, no module proxy, no Docker and no Postgres:
- Go build, `go vet`, `go test -race ./...`: not run. Every test was written as specified but has never executed.
- Postgres database tests (`apps/api/internal/store/bills_test.go`): not run; need a scratch Postgres.
- Migration 00007 goose up, down, up check: not run.
- Browser smoke test of the Bills tab: not run.
- Thai strings were not reviewed by a native speaker.
- Open minor findings carried from the reviews: `HasPayments` uses `LIKE` on the bill id (store/bills.go:115; the id is validated as a UUID first, so no wildcard can reach it from tools); the future-month "ask for a date" path leads to a dead end because any date in a future month is refused (the tab's date picker has no valid day either); the dashboard cannot undo an ending; `billName` refuses over-long names at 101 characters rather than cutting; the bill and transaction caps are checked then inserted, not atomically; `hermes/instructions.go` does not mention bills; some bill messages are not in `voice_test.go`; `Bills.tsx` is one large component.

## Decisions and assumptions
Rulings recorded in the SDD ledger:
- Verification environment: implementers ran only `gofmt -l` and the web checks, wrote all Go tests, and report everything else as UNVERIFIED. CI is the gate before merge.
- Commit identity: kept the configured identity `Claude <noreply@anthropic.com>`; nothing authored as the owner. If the owner wants their identity, rewrite with `git rebase --exec 'git commit --amend --no-edit --reset-author'` before the PR.
- No push and no PR: both publish to a shared remote, so they were left for the owner.
- `has_payments` in `list_bills` costs one query per bill (at most 100 because of the cap). Accepted, no batching.

Decisions made while planning, as built:
- D1: `confirm_bill_payment` with no date uses today only if today is in the month confirmed; otherwise it is refused and asks for a date. A given date must be within the month.
- D2: confirming a month the bill does not count in is refused.
- D3: `end_bill` refuses `ended_from` before `start_month`; `update_bill` refuses a `start_month` after an existing `ended_from` unless `clear_ended` is set.
- D4: `confirm_bill_payment` applies the 50,000 transaction cap.
- D5: `update_bill` takes `clear_ended` to undo an ending; `end_bill` takes `ended_from`.

D1 to D5 were approved with the plan on 2026-11-12.

## Next steps
- Run CI on a PR (Go vet and race tests, Postgres tests, web build) and fix what it finds; run goose up, down, up for 00007 and a browser smoke test of the Bills tab.
- Fix the future-month dead end with its own refusal message and tests.
- Add a way to undo an ending from the tab, and consider batching `has_payments`.
