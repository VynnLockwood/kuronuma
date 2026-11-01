# Dashboard PR H: shell, Overview and Spending Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline) or superpowers:subagent-driven-development. Steps use checkbox (`- [ ]`) syntax.

**Goal:** A signed-in user opens `/dashboard` (or `/th/dashboard`) and sees an Overview of their spending and memory, and a Spending tab where they can filter, add and delete records, in English or Thai.

**Spec:** `docs/superpowers/specs/2026-10-31-web-dashboard-design.md` (section 2, tabs Overview and Spending). PR H of five; needs PR G (the bridge) merged. Visual language: the Blush theme and shared components already in `apps/web` (`docs/design.md` Theme bullet); the dashboard adds no motion beyond the tab fade.

**Architecture:** a typed client for the bridge (`lib/web.ts`), copy for both languages (`lib/dashboard-copy.ts`), small shared pieces (`components/dashboard/Confirm.tsx` on native `<dialog>`, `charts.tsx` with `BarList` and `Columns`, `format.ts` for money and dates), a `Shell` that handles the session states, language switch and tabs (query string `?tab=`), and one component per tab. `/dashboard` and `/th/dashboard` render the same shell with a `lang` prop. Tab components are written so PRs I to K add tabs without touching the shell beyond one registry entry.

## Global Constraints

- Every piece of user text is a React text node: no `dangerouslySetInnerHTML`, no user-controlled URLs. Money amounts are shown from the strings the tools return; no floating-point arithmetic on money.
- `lib/web.ts`: `callTool(name, args)` sends `POST /api/web/call` with `X-Kuronuma-Web: 1`, `Content-Type: application/json`, same-origin credentials; returns a discriminated result `{ok: true, data} | {ok: false, kind: "signed-out" | "refused" | "unavailable", message}`; never throws raw HTML or a stack; 422 gives the tool's text as `message`, 401 `signed-out`, anything else `unavailable` with a generic message in the page language.
- States on every tab: loading (skeletons), empty (with a link to `/connect` and one sentence), error (the tool's own message or the generic one, with a retry button), disabled buttons while a request runs; after a change the tab refetches (no optimistic updates).
- Accessibility: labelled controls, keyboard operation, focus returned to the trigger after a dialog closes, a table under every chart for screen readers, contrast per `check:design`.
- `npm run check:design`: tokens only (no hex outside `:root`), no gradients. No new dependency, no chart library.
- Not signed in: the existing sign-in card (translated). API unreachable: the existing notice (translated).
- Thai copy is written natively, not word for word; every Thai string is listed in the PR for a native speaker. Thai headings are never tracked tighter (existing rule).
- Both pages are `noindex` (existing dashboard layout); the Thai route needs the same.
- Tests (outside the repo, written first and seen failing): Playwright with faked `/api/auth/me` and `/api/web/call` responses. Fixtures are real: capture one real result per tool from the Go handlers (a throwaway test in the scratch copy that prints each tool's JSON for a seeded fake store) and save them under `/home/claude/scratch/dash-fixtures/`. Shared helper `/home/claude/scratch/dash_mock.py`.
- Commits: `VynnLockwood <120536136+VynnLockwood@users.noreply.github.com>`, no AI trailers. Conventional Commits. Branch `feat/dashboard-shell-spending`. No owner approval needed.

## Review Focus

- A request is never sent before the session is known to be signed in; a 401 mid-session shows the sign-in card, not a broken tab.
- User text is rendered as text everywhere (merchant, note, category, tags); a merchant like `<img src=x onerror=...>` shows literally.
- Money: amounts shown exactly as returned, per currency (never summed across currencies), negative or zero cases, `possible_duplicate` is asked about and retried with `allow_duplicate`.
- The 200-record limit is stated on screen when the list is full.
- Thai: every visible string is translated, no English leaks, layouts survive longer Thai text at 390 px, no negative tracking.
- Phone width (390 and 360): no horizontal scroll, tap targets at least 40 px, charts readable.
- Focus: dialogs trap and restore focus, Escape closes, the destructive button is not the default.
- The letter field and existing pages are unaffected; existing checks (palette, motif, motion, focus, brand, letter) still pass.

---

### Task 1: Client, copy skeleton, shared pieces

**Files:** Create `apps/web/lib/web.ts`, `lib/format.ts`, `lib/dashboard-copy.ts` (skeleton with the shell strings), `components/dashboard/Confirm.tsx`, `components/dashboard/charts.tsx`. Tests: `/home/claude/scratch/dash-lib-check.mjs` (Node with `--experimental-strip-types`) for `callTool` against a stub `fetch` and for `format.ts`.

- [ ] **Step 1: Write the failing tests:** `callTool` sends the method, URL, headers and JSON body; maps 200, 401, 403, 404, 422 (message from `error`), 429, 500 and a network failure to the right result; an HTML error page body never reaches `message`; `format` shows `450.50 THB` style amounts from strings, formats dates and month labels in `en` and `th` with the Gregorian year in both (`1 Nov 2026` and `1 พ.ย. 2026`; the Thai Buddhist year is not used, and the PR asks the native speaker to confirm); per-currency grouping never adds across currencies.
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement** the files. `BarList` (label, value as string, share 0 to 1) and `Columns` (label, value, share) draw with CSS widths/heights and tokens only and render a visually hidden table; `Confirm` wraps `<dialog>` with a title, text, confirm and cancel buttons, focus restore.
- [ ] **Step 4: Verify.** Node test PASS; `tsc`, `check:design`, `build`.
- [ ] **Step 5: Commit** `feat(web): add the dashboard client library and shared pieces`.

### Task 2: Shell and Overview

**Files:** Create `components/dashboard/Shell.tsx`, `Overview.tsx`, `app/dashboard/page.tsx` (replace the current hub page, keeping its sign-in card, loading skeleton and notice as translated components), `app/th/dashboard/page.tsx` and layout metadata; extend `dashboard-copy.ts`. Tests: `/home/claude/scratch/dash-shell-check.py`.

- [ ] **Step 1: Write the failing test:** signed out shows the sign-in card in EN and TH; unreachable shows the notice; signed in shows the tab list (Overview, Spending; later PRs add more) with `?tab=` selecting it, a language switch that keeps the tab, and no request to `/api/web/call` while signed out; Overview shows this month's spending per currency, spending by category (bars), the last 6 months (columns) and the counts of open, overdue and due-this-week tasks and the five newest notes, built from `get_spending_summary` (by category and by month), `list_memories` and `list_categories`; the empty account shows the empty states with links to `/connect`; an error in one block does not blank the others; a 401 mid-session shows the sign-in card; phone width has no horizontal scroll.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement.** Keep what the current dashboard shows that is still useful (user chip, sign out, the module list as the Overview's footer or removed if redundant; decide and record a ruling).
- [ ] **Step 4: Verify.** The test PASS in EN and TH at 1280 and 390; screenshots reviewed; `tsc`, `check:design`, `build`; earlier scratch checks still pass.
- [ ] **Step 5: Commit** `feat(web): add the dashboard shell and Overview`.

### Task 3: Spending tab

**Files:** Create `components/dashboard/Spending.tsx`; extend copy. Tests: `/home/claude/scratch/dash-spending-check.py`.

- [ ] **Step 1: Write the failing test:** date and category filters call `list_transactions` with the right arguments; the list is newest first and shows a note when 200 records came back; the summary switch (category, month, merchant) calls `get_spending_summary` with `group_by` and draws bars with a table under them; add form (amount, currency, merchant, date, category with the user's names suggested from `list_categories`, note) calls `add_transaction`, shows validation text from a 422 inline, refetches on success and shows a toast-free confirmation line; `possible_duplicate` shows a question and retrying sends `allow_duplicate: true`; a new category name in the form is created by the tool and the page says so (`category_created`); delete opens the confirm dialog, calls `delete_transaction` only after confirming, and refetches; a merchant containing markup shows literally; EN and TH at 1280 and 390.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Verify** as above plus keyboard-only walk through add and delete (focus order and restore).
- [ ] **Step 5: Commit** `feat(web): add the Spending tab`.

### Task 4: Docs, review and PR

- [ ] `docs/design.md` (dashboard section: routes, the client, states, charts, language), the Thai strings list for the PR, whole-branch review by a fresh Opus reviewer with this Review Focus (one fix pass, tests first), agent log, PR `feat(web): add the dashboard shell, Overview and Spending` with the template, the real CI result, "Rulings I made" and "Deferred minors".
