# Dashboard shell, Overview and Spending (PR H)

- Agent: claude
- Date: 2026-11-02
- Branch / PR: feat/dashboard-shell-spending

## Goal
A signed-in user opens `/dashboard` or `/th/dashboard` and sees an Overview of their spending and memory and a Spending tab to filter, add and delete records, in English or Thai, using the web bridge from PR G.

## What I did
- `lib/web.ts` (`callTool`, `webRequest`): never throws, sends the `X-Kuronuma-Web` header, shows only a tool's own 422 text, maps everything else (including proxy HTML) to one generic message in the page language. `lib/format.ts`: money kept as strings (grouping, `share`, `compareAmount`, `compactAmount`), dates with the Gregorian year in both languages, `dueBucket`. `lib/dashboard-copy.ts`: all copy in English and Thai.
- Shell (session states, tabs in `?tab=`, language switch, `noindex`), Overview, Spending, `Confirm` on native `<dialog>`, hand-built `BarList` and `Columns` with tables for screen readers. Routes `/dashboard` and `/th/dashboard`; `AppNav` takes a language.
- The shared `Input` no longer hides its outline (the same faint-ring flaw fixed for buttons and tabs in the Blush review).
- Real tool output was captured from the Go handlers as test fixtures, so the UI is built on real shapes.

## Verified
- Every unit had a failing test first. `dash-lib-check.mjs` (Node, 18 checks); Playwright `dash-shell-check.py` and `dash-spending-check.py` with faked `/api/auth/me` and `/api/web/call` (English and Thai, 1280/390/360 px): signed out, unreachable, no request before the session is known, tabs and query string, language switch, empty, populated, one failing block, a 401 mid-session, hostile text, add, validation refusal, duplicate ask and retry, new category, already saved, delete with focus handling, phone widths, focus ring. All earlier checks (palette, motif, motion, focus/contrast/Thai, letter field, brand) still pass; `tsc`, `check:design`, `build` pass; Go suite ok.
- Looked at screenshots of both tabs, both languages, desktop and phone.
- A fresh whole-branch review found 2 Important issues (focus lost after a confirmed delete; totals on a different period than the list with no explanation) and I raised a third from Minor by effect (amounts cut with an ellipsis on phones); all fixed with failing tests first.

## Not verified / known gaps
- Nothing here talks to the real API: the browser tests use faked responses (real fixtures). The real bridge, session and Postgres are exercised only by CI and after deploy.
- The Thai copy is mine, not a native speaker's: every string is listed in the PR.
- Not tried in Safari or Firefox; no real phone.

## Decisions and assumptions
- Gregorian year in Thai dates (the Thai Buddhist year is not used); the PR asks the native speaker.
- The old "Your modules" grid is gone (tabs replace it); the three example prompts moved into the empty state; the connect card stays at the foot of Overview.
- Spending layout is Totals, Add a record, Records (so adding does not need scrolling past up to 200 rows).
- After a save the amount, merchant, category and note clear; currency and date stay.
- Column-chart labels above 1,000 are shortened by cutting digits (2.5k); exact values stay in the table and the tooltip.
- Deferred minors are listed in the PR (dialog click and Escape edge cases, stale `today` across midnight, English refusal texts on the Thai page, nav/tab desync, loading announcement, test gaps, two stale doc lines).

## Next steps
- PR I: Tasks and notes.
