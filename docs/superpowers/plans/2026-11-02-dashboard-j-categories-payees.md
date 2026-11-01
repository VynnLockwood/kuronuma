# Dashboard PR J: Categories and Payee rules Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline) or superpowers:subagent-driven-development. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Two tabs: Categories (the one list shared by spending, notes and payee rules, with how much each category is used, create, rename and merge) and Payee rules (list, set, delete).

**Spec:** `docs/superpowers/specs/2026-10-31-web-dashboard-design.md` (section 2, tabs Categories and Payee rules). PR J of five; needs G and H merged (I is independent).

**Architecture:** `components/dashboard/Categories.tsx` and `PayeeRules.tsx`, registered as tabs; copy in `lib/dashboard-copy.ts`. Tools: `list_categories` (verified: it returns the user's categories with how many records use each, and marks a category `unlisted` when records use a name that is not in the list; read the exact field names from `themis.go` and its tests), `create_category`, `rename_category`, `merge_categories`, `list_payee_rules`, `set_payee_rule`, `delete_payee_rule`. No new API code. If a per-module usage split (spending, memory, payee rules) is wanted and `list_categories` only has a total, show the total; do not add a tool in this PR (it would need its own approval).

## Global Constraints

All Global Constraints of the PR H plan apply. Branch `feat/dashboard-categories-payees`. No owner approval needed.

- `uncategorized` is built in: it is shown, never offered for rename, merge-away or delete, and is a valid merge target only if the tool allows it (follow the tool's refusal text).
- Limits: at most 100 categories per user (the tool refuses more; the form shows the tool's text); merge accepts a bounded number of sources (read `maxMergeSources` from `themis.go`); payee rules at most 500; use the same numbers in hints, the server stays the authority.
- A category marked `unlisted` is shown with a plain-words explanation and a one-click "merge into a listed category" that opens the merge dialog with it preselected.
- Rename and merge move records: after either, show the count the tool reports (`updated` / `moved`) in plain words, then refetch everything the tab shows.
- Merge is destructive for the source names: a confirm dialog lists the sources, the target and the warning that the sources disappear; pick sources by checkbox, target by select.
- Payee rules apply to new records only; say so beside the form (the tool's own message).
- A payee rule's category must already exist (the tool refuses otherwise): the form offers the user's names; a refusal shows the tool's text, which lists their categories.

## Review Focus

- Rename and merge cannot be sent with an empty, identical or built-in name; the dialogs show exactly what will happen and the counts afterwards match the tool's.
- After a merge or rename every list on screen (including the payee rules' categories and the form's suggestions) refetches, so no stale name remains.
- User text (category names, payees, descriptions) renders as text; long names do not break the layout at 390 px.
- 422 texts are shown unchanged next to the control that caused them; nothing is retried automatically.
- Keyboard and screen-reader operation of the checkbox list and dialogs; focus restored.
- Thai: terms consistent with the other tabs.

---

### Task 1: Categories tab

**Files:** Create `components/dashboard/Categories.tsx`; shell registry; copy. Tests: `/home/claude/scratch/dash-categories-check.py`.

- [ ] **Step 1: Write the failing test** (faked responses, EN and TH, 1280 and 390): the list with the usage counts; create calls `create_category` and refetches; rename (inline form) calls `rename_category` with `from` and `to` and shows the moved count; merge (checkbox sources, target select, confirm dialog) calls `merge_categories` and shows the moved count; `uncategorized` has no rename or merge-away controls; refused inputs (duplicate name, too many categories, same name) show the tool's text; loading, empty and error states; a hostile category name shows literally; no horizontal scroll at 390.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Verify.** Test PASS, `tsc`, `check:design`, `build`, earlier scratch checks, screenshots reviewed.
- [ ] **Step 5: Commit** `feat(web): add the Categories tab`.

### Task 2: Payee rules tab

**Files:** Create `components/dashboard/PayeeRules.tsx`; shell registry; copy. Tests: `/home/claude/scratch/dash-payees-check.py`.

- [ ] **Step 1: Write the failing test:** the list (payee, category, description); set (payee, category from the user's names, description) calls `set_payee_rule` and shows created vs updated; delete asks first, then `delete_payee_rule`; the "applies to new records only" note; a refusal for a category that does not exist shows the tool's text; empty state explains what a rule does with one example in the page language; EN and TH at 1280 and 390; hostile text renders literally.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Verify** as above.
- [ ] **Step 5: Commit** `feat(web): add the Payee rules tab`.

### Task 3: Docs, review and PR

- [ ] `docs/design.md`, Thai strings list, whole-branch review by a fresh Opus reviewer with this Review Focus (one fix pass, tests first), agent log, PR `feat(web): add the Categories and Payee rules tabs` with the template, the real CI result, "Rulings I made" and "Deferred minors".
