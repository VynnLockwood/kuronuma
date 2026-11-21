# Money monitor and mobile dashboard plans

- Agent: claude
- Date: 2026-11-22
- Branch / PR: docs/money-monitor-plan

## Goal
The owner merged the chart-based spending dashboard (PR #59) and asked for it to work "like a money monitor, not just a simple visual", keeping the current design, with a layout that works well on a phone. Plan only; no product code.

## What I did
- Read what #59 added (`SpendCharts.tsx`, Overview, Spending, the agent log of that session) and what the API can already return (`get_spending_summary` by category, month or merchant; `list_transactions` up to 200).
- Wrote `docs/superpowers/plans/2026-11-22-money-monitor.md`: a "Month check" card on the Overview (pace against the same days last month, a projection, what changed, larger than usual, looks recurring). It uses existing tools only, so no migration, no dependency and no API code.
- Wrote `docs/superpowers/plans/2026-11-22-mobile-dashboard.md`: records first on phones, add form and extra filters behind controls, compact tiles, larger tap areas in the header. Desktop is unchanged.
- Re-set the local `origin` to `https://github.com/VynnLockwood/kuronuma` (it had gone back to the old name).

## Verified
- `npm ci` then `npm run build` in `apps/web` on `main` (79a7219): build passes. The first build failed only because `recharts` was not installed locally yet.
- Playwright against the built app with mocked API data at 360 and 390 px, Overview and Spending: no horizontal overflow. The Spending tab is about 8,800 px tall at 390 px, with the add form above the records, seven filter fields stacked and three full-width KPI tiles.
- Linked paths in both plans exist, except the files the plans say to create.

## Not verified / known gaps
- Nothing was run against real data or the live API.
- The projection and the recurring and unusual rules are my choices (see Decisions in the plan); they have not been shown to the owner.
- Thai copy in the plans is written by me and needs a native check.

## Decisions and assumptions
- "This old design" read as "keep the current design and build on it".
- Last month is the reference line instead of a budget, because budgets need a table and the owner's approval for a migration.
- Two PRs, monitor first, then mobile, so each stays one concern.

## Next steps
- Owner merges this plan PR (the approval), then I implement the monitor, then the mobile layout.
- Later, if wanted: budgets (needs a migration), recurring reminders (Horae).
