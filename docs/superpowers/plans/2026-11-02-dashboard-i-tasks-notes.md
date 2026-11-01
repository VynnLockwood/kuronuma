# Dashboard PR I: Tasks and notes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline) or superpowers:subagent-driven-development. Steps use checkbox (`- [ ]`) syntax.

**Goal:** A Tasks and notes tab where the user sees and manages everything the assistants saved as tasks, to-dos and notes: filter, add, edit, complete, reopen, archive and delete.

**Spec:** `docs/superpowers/specs/2026-10-31-web-dashboard-design.md` (section 2, tab Tasks and notes). PR I of five; needs PRs G and H merged (the bridge, `lib/web.ts`, the shell, copy file, `Confirm`, shared states).

**Architecture:** one component `components/dashboard/Memories.tsx` registered as a tab in the shell; copy added to `lib/dashboard-copy.ts` (EN and TH). It uses only the existing tools through `callTool`: `list_memories`, `add_memory`, `update_memory`, `delete_memory`, `list_categories`.

## Global Constraints

All Global Constraints of the PR H plan apply (text nodes only, `callTool` result handling, states, accessibility, `check:design`, Thai written natively and listed in the PR, fixtures captured from real handler output, Playwright with faked responses, commits as the owner without AI trailers). Branch `feat/dashboard-tasks-notes`. No owner approval needed.

- Kinds: `task`, `todo`, `note`; status `open`, `done`, `archived`; priority `low`, `normal`, `high`; due date `YYYY-MM-DD`; tags up to 10 of at most 30 characters; titles up to 200 and details up to 4,000 characters (read the exact limits and argument names from `apps/api/internal/modules/atlas/tools.go` and use the same constants in client-side hints; the server remains the authority and its 422 text is shown).
- Order: open items first (overdue, then due soonest, then priority, then newest), as `list_memories` returns them; the tab does not re-sort except to group by status.
- Overdue means open with a due date before today in the user's browser date; "due this week" matches the Overview's rule from PR H (share one helper).
- Completing sets `status: done`; reopening sets `status: open`; archiving sets `archived`; the tool handles `completed_at`.
- The details text loses line breaks (the server's sanitiser); the form says so once, near the field, in plain words (known limit, owner decision pending).
- Delete is permanent and uses the confirm dialog; archive is the gentle option and is offered first.

## Review Focus

- Every edit sends only the changed fields; clearing the due date sends `due_on: ""` and clearing the details sends `body: ""` (the tool's documented way), tags are replaced as a whole list; an unchanged form sends nothing.
- A 422 (a refused input) keeps the user's typing and shows the tool's text next to the form.
- Overdue and due-this-week are correct around midnight and in the user's time zone (dates compared as dates, not instants).
- Filters (kind, status, category, tag, search) combine, are reflected in the query string so a reload keeps them, and an empty result says why (filter vs nothing saved yet).
- User text (title, details, tags, category) renders as text; a title like `<script>` shows literally; tag chips cannot overflow the card at 390 px.
- Keyboard: add, edit, complete, delete are all reachable; focus restored after dialogs; screen readers hear the status changes (live region).
- Thai: every string translated; the status and kind words match across the Overview and this tab.

---

### Task 1: List, filters and the status actions

**Files:** Create `components/dashboard/Memories.tsx` (list, filters, status actions); modify the shell tab registry; extend copy. Tests: `/home/claude/scratch/dash-memories-check.py`.

- [ ] **Step 1: Write the failing test** (faked responses, EN and TH, 1280 and 390): the tab loads `list_memories`; kind, status, category, tag and search filters call it with the right arguments and update the query string (a reload keeps them); open first, then done, then archived; overdue items are marked (text, not colour alone); complete, reopen and archive call `update_memory` with `id` and `status` only and refetch; loading, empty (nothing saved yet vs nothing matches the filters) and error states; a title with markup shows literally; no horizontal scroll at 390.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Verify.** Test PASS, `tsc`, `check:design`, `build`, earlier scratch checks, screenshots reviewed.
- [ ] **Step 5: Commit** `feat(web): list and filter tasks and notes on the dashboard`.

### Task 2: Add, edit and delete

**Files:** Modify `Memories.tsx`; extend copy; extend the check file.

- [ ] **Step 1: Write the failing test:** the add form (kind, title, details, due date, priority, category with the user's names suggested, tags) calls `add_memory`, shows `category_created` when a new category came from it, and refetches; a 422 shows the tool's text and keeps the typing; edit opens the same form filled, sends only the changed fields to `update_memory` (clearing the due date sends `due_on: ""`), and an unchanged form sends nothing; delete asks first, then calls `delete_memory`, then refetches; double clicks do not send twice (button disabled while running); keyboard-only walk through add, edit and delete; the details field explains that line breaks are not kept.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Verify** as above.
- [ ] **Step 5: Commit** `feat(web): add, edit and delete tasks and notes on the dashboard`.

### Task 3: Docs, review and PR

- [ ] `docs/design.md` (dashboard section), Thai strings list, whole-branch review by a fresh Opus reviewer with this Review Focus (one fix pass, tests first), agent log, PR `feat(web): add the Tasks and notes tab to the dashboard` with the template, the real CI result, "Rulings I made" and "Deferred minors".
