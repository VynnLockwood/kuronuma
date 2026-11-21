# Spending dashboard charts

- Agent: claude
- Date: 2026-11-21
- Branch / PR: claude/sleepy-cannon-hfgb2g, no PR yet

## Goal
Redesign the spending views of the dashboard (Overview and Spending tabs) with more visual charts using a free component library. Landing page unchanged.

## What I did
- Added `recharts` (MIT) with the owner's approval; `package.json` and `package-lock.json` updated.
- New `components/dashboard/SpendCharts.tsx`: donut (category share), area trend (months), ranked bars (merchants), KPI tiles. Hover tooltips, reduced-motion respected, hidden from screen readers with the existing sr-only table kept.
- Spending tab Totals: KPI row plus a chart that follows the Category / Month / Merchant switch; the existing table stays.
- Overview tab: category donut and a 6-month trend replace the CSS bars.
- Chart colours are `--chart-1..6` tokens in `globals.css` (no hex in components, no gradients).
- `sumAmount` in `lib/format.ts` adds money strings with BigInt for the folded "Other" slice.
- Copy keys `other`, `topLabel`, `shareHead` in English and Thai.
- Follow-up: Spending tab filters and sorting. Quick period presets (this month, last 30 days, last 3 months, this year) fill the existing from/to dates. A search box matches merchant or note, and a sort menu orders records (newest, oldest, highest, lowest, merchant). Totals have a sort menu (largest first, name). Search and sort run in the browser on the fetched records only. Copy added in English and Thai.

## Verified
- `npx tsc --noEmit`: no errors.
- `npm run check:design`: ok.
- `npm run build`: compiled, all routes generated.
- Playwright with mocked API: sort orders, search ("cafe" showed 2 of 4) and the "This month" preset (set the From date, re-fetched with it) behaved as expected.
- Rendered both tabs with Playwright against mocked API responses (screenshots checked by eye).

## Not verified / known gaps
- Not run against real data or the live API. Thai layout and the Merchant chart were not screenshotted.
- No unit tests for `sumAmount` or the sort and search logic.
- Search and sort only cover the 200 most recent records the list loads.
- `npm audit` reports an existing Next.js advisory (not from this change); a Next upgrade needs owner approval.
