# Dashboard Money Monitor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline) or superpowers:subagent-driven-development. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Turn the Overview from "charts of what happened" into a monitor: a "Month check" card per currency that says how this month is going compared with the same days last month, where it is heading, what changed, what looks unusual and what looks recurring.

**Architecture:** Pure functions in `lib/monitor.ts` (money in whole minor units, no floating point) feed one new component, `components/dashboard/MonthCheck.tsx`, placed first on the Overview. It uses the tools the dashboard already calls plus two more reads of existing tools (`get_spending_summary` for the same days last month, `list_transactions` for the last four months). No budget table, no migration, no API code.

**Tech Stack:** Next.js 16, React, Tailwind 4, the existing Recharts-based `SpendCharts.tsx`. No new dependency.

**Spec:** `docs/superpowers/specs/2026-10-31-web-dashboard-design.md` (Overview tab). There is no separate spec for this change: the owner asked for "a money monitor, keep the current design, improve it and make it work on mobile", and the Decisions below are my calls. Rulings made without a spec are provisional.

## Global Constraints

- Branch `feat/dashboard-money-monitor`. Web only. No new dependency, no migration, no Go code, no deploy, CI or `.env.example` change. If the work seems to need any of these, stop and ask the owner (AGENTS.md rule 6).
- Money stays a string. Add, subtract and compare in whole minor units with `BigInt`, never through `Number`. Amounts in different currencies are never combined: every figure is per currency.
- Every derived figure is worded as an estimate or a reference ("about", "same days last month"). Nothing reads as advice.
- User text (merchant, category) renders as text and wraps; nothing overflows at 360 px.
- Colours only from the existing tokens. Direction is never shown by colour alone (words carry it). `npm run check:design` must pass.
- No new animation. Charts keep their `sr-only` table.
- Dates are compared as dates with the helpers in `lib/format.ts` (`localISODate`, `addDaysISO`, `monthStartISO`, `lastMonths`). No `Intl`, no `Date` arithmetic on local time.
- English and Thai copy for every string. The PR lists every Thai string for a native speaker to check.
- Fetch budget: the Overview adds exactly two tool calls (listed in Task 4). The previous month's full total comes from the existing 6-month series, not a new call.

## Decisions (my calls)

- **D1 Reference, not budget.** Last month is the reference. Budgets need a table and a migration; that is a later PR the owner must approve.
- **D2 Pace.** Compare this month so far with the same days last month (day of month clamped to last month's length).
- **D3 Projection.** `spent so far x days in month / day of month`, rounded down, shown only from the 5th of the month. Before that: "Too early in the month to estimate."
- **D4 What changed.** Up to 3 categories with the largest rise against the same days last month. A category with nothing last time is marked new. Falls are not listed.
- **D5 Larger than usual.** A record this month whose amount is at least 3 times the median of that category's records in the previous three months, with at least 3 such earlier records. Up to 3, largest first.
- **D6 Looks recurring.** Same merchant (trimmed, case-insensitive) and currency in each of the two most recent full months, the two amounts within 10% of the larger one. Shows the latest amount, the latest day of month and whether it has appeared this month. Up to 5, largest first.
- **D7 Placement.** First block of the Overview, one card per currency. The Spending tab is unchanged. Shown only when this month or the same days last month have spending.
- **D8 Degrading.** An error in the records call shows one retry line under the card and never hides the pace block. A block with nothing to say is not drawn (no empty headings).

## Review Focus

- Day 1 to 4 of a month and a month with no spending: no projection, no division by zero, no "NaN" or "Infinity".
- Month ends and year wrap: on 31 Mar the same period is 1 to 28 Feb; on 15 Jan it is 1 to 15 Dec of the year before.
- Mixed currencies (THB and USD records together): nothing is combined or compared across them, in all four blocks.
- The 200-record cap: the card says "Based on your newest 200 records" and never calls a merchant recurring from a month the list only partly covers.
- Long merchant and category names and Thai text at 360 px: no horizontal scroll, no clipped amounts.

## File Structure

- Modify `apps/web/lib/format.ts`: export three money helpers next to the private `minor()`.
- Create `apps/web/lib/monitor.ts`: pure functions, no React, no `@/` imports (so a plain Node check can load it).
- Create `apps/web/components/dashboard/MonthCheck.tsx`: the card.
- Modify `apps/web/lib/dashboard-copy.ts` (type, `en`, `th`), `apps/web/components/dashboard/Overview.tsx` (two calls, render the card), `docs/design.md` (the dashboard paragraph).
- Create `agent-logs/claude/2026-11-22-HHMM-money-monitor.md`.
- Scratch, outside the repo as in earlier PRs: `monitor-check.mjs` (with a small resolve hook so Node loads extensionless `.ts` imports), `month-check-check.py` (Playwright with `dash_mock.py` fixtures).

### Task 1: Money helpers in `lib/format.ts`

**Files:** Modify `apps/web/lib/format.ts`. Test: scratch `monitor-check.mjs`.

**Interfaces:**
- Produces: `diffAmount(a: string, b: string): string` (a minus b, two decimals, signed); `percentOf(part: string, whole: string): number | null` (whole percent, half away from zero, `null` when whole is zero or either input is unusable); `scaleAmount(a: string, num: number, den: number): string` (a x num / den rounded down in minor units; `den <= 0` gives `"0.00"`).

- [ ] **Step 1: Write the failing test.** `diffAmount("120.50","100.00") === "20.50"`; `diffAmount("100.00","120.50") === "-20.50"`; `diffAmount("5","5.00") === "0.00"`; `diffAmount("x","1.00") === "-1.00"`. `percentOf("25.00","100.00") === 25`; `percentOf("1.00","3.00") === 33`; `percentOf("2.00","3.00") === 67`; `percentOf("-25.00","100.00") === -25`; `percentOf("1.00","0.00") === null`. `scaleAmount("2200.00",30,22) === "3000.00"`; `scaleAmount("10.00",1,3) === "3.33"`; `scaleAmount("10.00",1,0) === "0.00"`.
- [ ] **Step 2: Run it, expect FAIL** ("diffAmount is not a function" and the two others).
- [ ] **Step 3: Implement the three functions** with `minor()` and `BigInt`; reuse the existing formatting of `sumAmount` for output.
- [ ] **Step 4: Run it, expect PASS** (all assertions above).
- [ ] **Step 5: Commit** `feat(web): money helpers for the month check`.

### Task 2: Period and pace in `lib/monitor.ts`

**Files:** Create `apps/web/lib/monitor.ts`. Test: scratch `monitor-check.mjs`.

**Interfaces:**
- Consumes: Task 1 helpers; `addDaysISO`, `monthStartISO`, `lastMonths` from `./format`.
- Produces: `daysInMonth(ym: string): number`; `samePeriod(today: string): { prevFrom: string; prevTo: string }`; `type Pace = { spent: string; prevSame: string; prevFull: string | null; delta: string; percent: number | null; direction: "up" | "down" | "same"; projected: string | null }`; `pace(i: { today: string; spent: string; prevSame: string; prevFull: string | null }): Pace`.

- [ ] **Step 1: Write the failing tests.** `daysInMonth("2026-02") === 28`, `("2028-02") === 29`, `("2026-11") === 30`. `samePeriod("2026-11-22")` is `{prevFrom:"2026-10-01", prevTo:"2026-10-22"}`; `samePeriod("2026-03-31")` is `{"2026-02-01","2026-02-28"}`; `samePeriod("2026-01-15")` is `{"2025-12-01","2025-12-15"}`. `pace({today:"2026-11-22", spent:"2200.00", prevSame:"2000.00", prevFull:"2800.00"})` gives `delta "200.00"`, `percent 10`, `direction "up"`, `projected "3000.00"`. On `2026-11-04` `projected` is `null`. `prevSame "0.00"` with spent above zero gives `percent null`, `direction "up"`. Equal amounts give `direction "same"`, `percent 0`. `prevFull: null` is accepted and still projects.
- [ ] **Step 2: Run it, expect FAIL** (module or function missing).
- [ ] **Step 3: Implement** the four functions in `monitor.ts` (no React, imports from `./format` only for what it needs).
- [ ] **Step 4: Run it, expect PASS.**
- [ ] **Step 5: Commit** `feat(web): month period and pace for the month check`.

### Task 3: Movers, unusual and recurring in `lib/monitor.ts`

**Files:** Modify `apps/web/lib/monitor.ts`. Test: scratch `monitor-check.mjs`.

**Interfaces:**
- Consumes: Task 1 helpers, Task 2 `daysInMonth`.
- Produces: `type Group = { key: string; currency: string; total: string; count: number }`; `type Tx = { id: string; amount: string; currency: string; merchant: string; category: string; occurred_on: string }`; `type Mover = { category: string; now: string; before: string; delta: string; isNew: boolean }`; `movers(now: Group[], before: Group[], currency: string, limit?: number): Mover[]` (default limit 3); `unusual(txs: Tx[], today: string, limit?: number): Tx[]` (default 3); `type Recurring = { merchant: string; currency: string; amount: string; day: number; seenThisMonth: boolean }`; `recurring(txs: Tx[], today: string, limit?: number): Recurring[]` (default 5). `unusual` and `recurring` take mixed currencies and never compare across them.

- [ ] **Step 1: Write the failing tests.**
  - movers: now `[food 900.00, transport 300.00, fun 50.00]`, before `[food 500.00, transport 400.00]` (THB) plus a USD group that must be ignored: result is `[food delta "400.00" isNew false, fun delta "50.00" isNew true]` (transport fell, not listed); `limit 1` returns only food.
  - unusual, today `2026-11-22`: earlier food records `100.00, 120.00, 110.00, 90.00` in Aug to Oct (median 105.00, threshold 315.00). A November food `400.00` is returned; `300.00` is not; exactly `315.00` is returned. A category with only two earlier records never returns. A USD food `400.00` is not compared with the THB history. Records before November are never returned. Result is largest first and at most 3.
  - recurring, today `2026-11-22`: Netflix USD `89.90` on `2026-09-05` and `2026-10-05`, none in November gives `{merchant:"Netflix", currency:"USD", amount:"89.90", day:5, seenThisMonth:false}`. `"netflix "` at `89.90` (Oct) and `"Netflix"` at `90.50` (Sep) counts (within 10%) and the amount is the latest, `89.90`. `50.00` then `100.00` does not count. One earlier month only does not count. Records in Aug and Oct but not Sep do not count. Adding a `2026-11-05` Netflix record gives `seenThisMonth: true`.
- [ ] **Step 2: Run it, expect FAIL.**
- [ ] **Step 3: Implement** the three functions. Median of an even count is the whole-minor-unit floor of the two middle values' average.
- [ ] **Step 4: Run it, expect PASS.**
- [ ] **Step 5: Commit** `feat(web): movers, unusual and recurring for the month check`.

### Task 4: Copy, `MonthCheck` and the Overview

**Files:** Modify `apps/web/lib/dashboard-copy.ts`, `apps/web/components/dashboard/Overview.tsx`. Create `apps/web/components/dashboard/MonthCheck.tsx`. Test: scratch `month-check-check.py`.

**Interfaces:**
- Consumes: Tasks 1 to 3; `Donut`/`Kpi` stay untouched; `useTool`, `BlockSkeleton`, `BlockError`, `Card`.
- Produces: `MonthCheck(props: { currency: string; today: string; lang: Lang; month: Group[]; monthTotal: string; prev: Group[]; prevTotal: string; prevFull: string | null; recent: Tx[] | null; recentCapped: boolean; onRetryRecent?: () => void; recentError?: string })`. Copy under `copy.overview.check`.

Calls added to `Overview` (and nothing else): `get_spending_summary` with `{ group_by: "category", from: prevFrom, to: prevTo }` from `samePeriod(today)`, and `list_transactions` with `{ from: lastMonths(today, 4)[0] + "-01", to: today, limit: 200 }`. `prevFull` is the previous month's entry of the existing 6-month series.

Copy (English / Thai), `{a}` an amount with currency, `{p}` a whole percent, `{c}` a category, `{n}` a number:
- title: "Month check" / "ตรวจเงินประจำเดือน"
- soFar: "{a} so far" / "ใช้ไปแล้ว {a}"
- up / down: "{p}% more than the same days last month" / "{p}% less than the same days last month"; Thai "มากกว่าช่วงเดียวกันของเดือนก่อน {p}%" / "น้อยกว่าช่วงเดียวกันของเดือนก่อน {p}%"
- same: "The same as the same days last month" / "เท่ากับช่วงเดียวกันของเดือนก่อน"
- noLast: "No spending in these days last month" / "ช่วงเดียวกันของเดือนก่อนไม่มีรายจ่าย"
- projected: "About {a} by the end of the month" / "ประมาณ {a} เมื่อสิ้นเดือน"; projectedNote: "An estimate from your pace so far." / "ประมาณการจากอัตราการใช้จ่ายถึงตอนนี้"; tooEarly: "Too early in the month to estimate." / "ยังเร็วเกินไปที่จะประมาณการ"
- lastMonthTotal: "Last month: {a}" / "เดือนก่อน: {a}"
- changedTitle: "What changed" / "สิ่งที่เปลี่ยนไป"; new: "New this month" / "ใหม่เดือนนี้"; changedLine: "{c}: {a} more" / "{c}: เพิ่มขึ้น {a}"
- unusualTitle: "Larger than usual" / "สูงกว่าปกติ"; unusualHint: "Usually smaller in {c}" / "ปกติต่ำกว่านี้ในหมวด {c}"
- recurringTitle: "Looks recurring" / "ดูเหมือนรายการประจำ"; notYet: "Not yet this month" / "เดือนนี้ยังไม่มี"; recurringTotal: "About {a} a month" / "ประมาณ {a} ต่อเดือน"; usuallyDay: "usually around day {n}" / "มักประมาณวันที่ {n}"
- capped: "Based on your newest {n} records." / "อ้างอิงจาก {n} รายการล่าสุด"

Layout: mobile first, one column. Top: the "so far" figure, then two thin comparison bars ("This month so far", "Same days last month") scaled to the larger, using `share()` and the chart tokens, each with its amount and a `sr-only` sentence that carries the same comparison. Then the projection line, then the three lists as plain rows (title, amount right-aligned, hint under it). Direction is words ("more", "less"), not colour.

- [ ] **Step 1: Write the failing browser check** (scratch, mocked API, time frozen to `2026-11-22`). With fixtures for 2200.00 THB so far, 2000.00 THB the same days last month, a 6-month series with October 2800.00, and records giving one unusual and one recurring record: at 390 and 1024 px the card shows "Month check", "2,200.00 THB so far", "10% more than the same days last month", "About 3,000.00 THB by the end of the month", "Last month: 2,800.00 THB", one row under each of the three lists; `scrollWidth` equals the viewport width; the calls made are exactly the existing five plus the two above. At `2026-11-03` there is no projection and "Too early" shows. With the records call failing, the pace block still shows and one retry line appears. With 200 records returned the capped note shows. With a 90-character merchant name at 360 px there is no overflow. In Thai (`/th/dashboard`) the Thai titles show.
- [ ] **Step 2: Run it, expect FAIL** ("Month check" not found).
- [ ] **Step 3: Add the copy** to the `DashboardCopy` type and to `en` and `th`.
- [ ] **Step 4: Implement `MonthCheck.tsx`** and wire the two calls and the card into `Overview.tsx` as its first section; hide the card when there is no spending in either window.
- [ ] **Step 5: Run the browser check and `npx tsc --noEmit`, expect PASS and no errors.**
- [ ] **Step 6: Commit** `feat(web): month check card on the overview`.

### Task 5: Docs, log, gates, PR

**Files:** Modify `docs/design.md` (the dashboard paragraph: add the Month check, say it needs no new API). Create `agent-logs/claude/2026-11-22-HHMM-money-monitor.md`.

- [ ] **Step 1: Update `docs/design.md`.** Check that every linked path exists.
- [ ] **Step 2: Run the gates:** `npx tsc --noEmit`, `npm run check:design`, `npm run build` in `apps/web`, then the scratch `v1-gates.sh` and `monitor-check.mjs`. Expected: all pass, the earlier dashboard checks still pass.
- [ ] **Step 3: Write the agent log** (what was done, the commands and results, the Thai strings to check, what was not tested: real data, the live API).
- [ ] **Step 4: Fresh Opus whole-branch review, one fix pass for Critical and Important, each fix test-first; list minors.**
- [ ] **Step 5: Open the PR** `feat(web): month check on the dashboard overview`, wait for green CI, update the body with the result.

## Not in this PR

Budgets and limits (needs a table and the owner's approval), alerts and notifications, reminders for recurring items (Horae), a server tool for daily totals or longer history, and any change to the Spending tab. Mobile layout of the rest of the dashboard is the next plan (`2026-11-22-mobile-dashboard.md`).
