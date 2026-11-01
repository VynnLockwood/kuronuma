# Dashboard: Tasks and notes (PR I)

- Agent: claude
- Date: 2026-11-02
- Branch / PR: feat/dashboard-tasks-notes

## Goal
A Tasks and notes tab where the user sees and manages everything the assistants saved as tasks, to-dos and notes: filter, add, edit, complete, reopen, archive and delete, in English and Thai.

## What I did
- `components/dashboard/Memories.tsx` and a new tab in the shell; copy for both languages. It uses only the existing tools through the web bridge (`list_memories`, `add_memory`, `update_memory`, `delete_memory`, `list_categories`).
- Filters (type, status, category, tag, search) are kept in the query string; items are grouped open, done and archived; open items past their due date show the word Overdue; complete, reopen and archive; an add form; an edit form that sends only what changed (clearing the due date sends an empty `due_on`); delete with a confirmation, archive offered first.
- A stateful fake of the Atlas tools in the shared test harness (filters like the real tool, add/update/delete change its list), built on real fixtures captured from the Go handlers.
- Small extras found by this tab's checks: the Overview's setup-guide link got a 40 px tap target, and the shared tab list now fits three tabs at 360 px.

## Verified
- Every unit had a failing test first. Playwright against `next start` with faked responses (English and Thai; 1280, 390 and 360 px): the list and groups, overdue marker, filters and query string and reload, status actions with focus and announcements, empty and no-match states, error and retry, hostile text, add (exact arguments, empty fields left out, refusal keeps the typing, new category named), edit (prefilled, unchanged sends nothing, only changed fields, clearing the date, focus return), delete (focus on the safe button, Escape, focus afterwards, an already-removed item, a failed delete), double submit (mutation-checked: removing the guard sends two adds), keyboard-only add, the visible label inside every accessible name, phone widths. All earlier checks still pass; `tsc`, `check:design`, `build` pass; Go suite ok.
- A fresh whole-branch review found 4 Important issues (status changes lost focus and were not announced; an unchanged edit could rewrite tags the user never touched; a failed delete was labelled a failed save; the tests missed several focus areas) and I raised two Minor findings by effect (accessible names without the visible label; the new tab clipped at 360 px). All fixed with failing tests first, except the broader test gaps, which are listed as deferred.

## Not verified / known gaps
- Nothing here talks to the real API: the browser tests use faked responses (real fixtures). The real bridge, session and Postgres are exercised only by CI and after deploy.
- The Thai copy is mine, not a native speaker's: every new string is listed in the PR.
- Not tried in Safari or Firefox; no real phone.

## Decisions and assumptions
- Filters live in the query string with `replaceState`; switching tab or language drops them.
- Search and tag boxes wait 350 ms before asking the server.
- In edit mode the type is read-only (the tool cannot change a kind); saving an unchanged edit says so and closes the form.
- After a successful add or edit focus goes to the heading; after Cancel or an unchanged save it returns to the button that opened the form.
- The Thai Mark done button is "ทำเครื่องหมายว่าเสร็จ" so the visible text is inside its accessible name.
- Deferred minors are in the PR (Clear filters flicker, a partly typed date, overlapping status actions, stale "today", URL filter values no control offers, a docs comma, test gaps).
- Owner's call: search words and tags are in the query string, so they reach the server on reload and can end up in browser history and the nginx access log.

## Next steps
- PR J: Categories and Payee rules.
