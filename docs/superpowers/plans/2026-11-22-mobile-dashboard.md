# Dashboard Mobile Layout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline) or superpowers:subagent-driven-development. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Make the signed-in dashboard comfortable on a phone: a much shorter Spending tab where the records come first, compact stat tiles, filters and the add form out of the way until needed, and header links that are easy to tap. Desktop layout does not change.

**Architecture:** Class and state changes in `Spending.tsx`, `Overview.tsx` and the shared header, with a few new copy strings. No new API calls, no new dependency.

**Tech Stack:** Next.js 16, React, Tailwind 4 (`max-md:` and `md:` variants; 768 px is the line, as elsewhere in the dashboard).

**Spec:** `docs/superpowers/specs/2026-10-31-web-dashboard-design.md`. No separate spec: the owner asked for "work well on mobile layout". Decisions below are my calls and provisional.

## Global Constraints

- Branch `feat/dashboard-mobile`, started after `feat/dashboard-money-monitor` is merged so the Month check card is audited too. Web only. No new dependency, migration, Go, deploy or CI change; if the work needs one, stop and ask the owner.
- At 768 px and wider nothing moves or changes (the same elements in the same order). Every change below is `max-md:` or behind a control that only exists below 768 px.
- Nothing may overflow horizontally at 360 px, in English and Thai (`/th/dashboard`), with a 90-character merchant name.
- Interactive controls are at least 40 px high where they are controls on their own line; the four header links get a 44 x 44 px tappable area without changing how the header looks or its width.
- Hiding is by state and CSS, not by unmounting: typed text, filters and the open or closed state survive rotating the phone or resizing across 768 px.
- A hidden control never hides its consequence: if filters are collapsed the list says how many are active; a save made from the collapsed add form shows its status message.
- `npm run check:design` passes; colours from tokens; reduced motion respected; copy in English and Thai, Thai strings listed in the PR.

## Decisions (my calls)

- **D1 Order on phones.** Totals, then Records, then Add a record (CSS `order`, so desktop keeps Totals, Add, Records).
- **D2 Add form.** Below 768 px it sits behind an "Add a record" button (`aria-expanded`, `aria-controls`); opening moves focus to Amount; a successful save closes it and the status line stays visible above the button. From 768 px it is always shown and the button is hidden.
- **D3 Filters.** Period chips stay visible and scroll sideways in one row. Search and Sort stay visible. Date range and category go into a native `<details>` "More filters" below 768 px, whose summary shows the count of active filters; from 768 px the same fields show as today.
- **D4 Tiles.** Spending tab: the first KPI spans the full row, the other two sit side by side. Overview tasks tiles sit three across with smaller figures. Currency total tiles stay as they are.
- **D5 Header links.** Hit-slop with a pseudo-element, not padding, so the 360 px header does not overflow again.
- **D6 Not done:** a sticky tab bar and a bottom navigation (both can cover a focused field or the keyboard), swipe between currencies, and any chart change (Recharts tooltips already work on tap).

## Review Focus

- Open the collapsed add form, type, rotate or resize across 768 px, rotate back: the text is still there.
- Collapsed filters with a date range set: the summary says "1" (or the count) and the list text still says it is filtered; "Clear filters" is reachable.
- Save from the collapsed form on a 360 x 640 screen with the on-screen keyboard up: the status message and the new record are findable without hunting.
- The delete confirmation dialog fits 360 x 640 and its buttons are tappable.
- Thai text and a 90-character merchant name at 360 px: no horizontal scroll on any tab.

## File Structure

- Modify `apps/web/components/dashboard/Spending.tsx`, `apps/web/components/dashboard/Overview.tsx`, `apps/web/components/SiteHeader.tsx`, `apps/web/components/AppNav.tsx`, `apps/web/components/RepoLink.tsx`, `apps/web/lib/dashboard-copy.ts`, `docs/design.md`.
- Create `agent-logs/claude/2026-11-22-HHMM-mobile-dashboard.md`.
- Scratch (outside the repo): `mobile-audit.py` (already written for this plan; extended with the checks below).

### Task 1: Spending tab on phones

**Files:** Modify `Spending.tsx`, `dashboard-copy.ts`. Test: scratch `mobile-audit.py`.

**Interfaces:**
- Produces: copy `copy.spending.addToggle` ("Add a record" / "เพิ่มรายการ"), `copy.spending.moreFilters` ("More filters" / "ตัวกรองเพิ่มเติม"), `copy.spending.activeFilters(n: number)` ("{n} on" / "เปิดอยู่ {n}"). Test ids `add-toggle` and `more-filters` on the two controls.

- [ ] **Step 1: Write the failing check** at 390 x 844 (touch) with the standard mock: no horizontal overflow; the Records heading's top is within 1700 px of the page top (today about 6000 px); the Add form's Amount field is not visible and `add-toggle` has `aria-expanded="false"`; activating `add-toggle` shows it, sets `aria-expanded="true"` and focuses Amount; From, To and category are not visible until `more-filters` is opened; with From set the summary shows "1 on"; at 1024 x 800 the add form is visible without a click, `add-toggle` is not visible, and the order is Totals, Add, Records; the full save flow at 390 (open, fill, save) closes the form, shows the "Saved" status and refetches the list; typed text in Amount survives a resize 390 to 1024 to 390.
- [ ] **Step 2: Run it, expect FAIL** (records heading far below the fold; toggle missing).
- [ ] **Step 3: Implement** the order classes (`max-md:order-*` on the three sections), the add toggle and `addOpen` state, the period chips as a single scrolling row (`max-md:flex-nowrap max-md:overflow-x-auto`, no page overflow), and the `<details>` around From, To and category with the active count.
- [ ] **Step 4: Run it and `npx tsc --noEmit`, expect PASS and no errors.**
- [ ] **Step 5: Commit** `feat(web): shorter spending tab on phones`.

### Task 2: Stat tiles, header links, whole-dashboard audit

**Files:** Modify `Spending.tsx` (KPI grid), `Overview.tsx` (task tiles), `SiteHeader.tsx`, `AppNav.tsx`, `RepoLink.tsx`. Test: scratch `mobile-audit.py`.

- [ ] **Step 1: Write the failing check** at 360 and 390 px, English and Thai, all three tabs: no overflow with a 90-character merchant in the records; the Spending KPI grid has the first tile spanning the row and the other two side by side (their tops are equal); the three task tiles share one row; each of the four header links has `elementFromPoint` hits across a 44 x 44 px box around its centre; header width and `scrollWidth` unchanged at 360; every `button`, `select`, `input`, `[role=tab]` on the dashboard is at least 40 px high (header links excluded by the hit-slop rule); desktop screenshot at 1280 is byte-identical to the one on `main` for the Overview and Spending tabs (the "nothing moves" rule).
- [ ] **Step 2: Run it, expect FAIL** (KPI tiles stacked, link hit areas small).
- [ ] **Step 3: Implement** the KPI and tile grids and the hit-slop pseudo-elements (`relative` plus `before:absolute before:-inset-2.5`).
- [ ] **Step 4: Run it, `npx tsc --noEmit`, `npm run check:design`, expect PASS.**
- [ ] **Step 5: Commit** `feat(web): compact tiles and easier header taps on phones`.

### Task 3: Docs, log, gates, PR

**Files:** Modify `docs/design.md` (a short "Mobile" bullet in the dashboard section). Create the agent log.

- [ ] **Step 1: Update `docs/design.md`;** check linked paths exist.
- [ ] **Step 2: Run the gates:** `npx tsc --noEmit`, `npm run check:design`, `npm run build`, scratch `v1-gates.sh` and `mobile-audit.py`. Expected: all pass, earlier dashboard checks still pass.
- [ ] **Step 3: Write the agent log** (commands, results, Thai strings, not tested: a real phone, a real on-screen keyboard, real data).
- [ ] **Step 4: Fresh Opus whole-branch review, one fix pass for Critical and Important, each fix test-first; list minors.**
- [ ] **Step 5: Open the PR** `feat(web): dashboard layout for phones`, wait for green CI, update the body with the result.

## Not in this PR

Sticky tab bar, bottom navigation, a swipe between currencies, chart changes, and any change to the landing page or the connect page.
