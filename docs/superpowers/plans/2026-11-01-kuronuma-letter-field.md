# Kuronuma letter field (PR V2) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline) or superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Paper letters drift slowly in the page background and ease away from a mouse pointer. Calm, faint, never in the way.

**Architecture:** A pure physics module (`lib/letter-physics.ts`) steps one letter at a time and builds the starting set, so it can be tested in Node without a browser. One client component, `LetterField`, mounted once in the root layout, draws a fixed, click-through layer behind the page and runs a single `requestAnimationFrame` loop that writes only `transform`. Under reduced motion it draws the letters once and never moves them.

**Tech Stack:** Next.js 16, React 19, the existing `Letter` motif. No new dependencies.

**Spec:** `docs/superpowers/specs/2026-10-31-kuronuma-rebrand-design.md` (section 3).

## Global Constraints

- About 14 letters on screens 640 px wide and up, 8 below. Size 22 to 40 px, opacity 0.35 to 0.7, the existing `Letter` motif (rose outline, white fill).
- Base drift 6 to 16 px per second, mostly up and sideways, plus a small tilt; letters wrap at the edges (margin = their own size).
- Pointer (mouse or pen only, never touch): a letter whose centre is within 120 px is pushed away; the push grows as the pointer gets closer (quadratic) and the letter tilts a little away from it. Velocity relaxes back to the base drift (damping), so the motion is smooth.
- Time step clamped to 50 ms. The loop stops while the tab is hidden and restarts on return. Only `transform` is written in the loop.
- The layer: `position: fixed; inset: 0; overflow: hidden; pointer-events: none`, behind all content (negative z-index), `aria-hidden="true"`, letters not focusable. It never causes horizontal scroll and never takes a click.
- Reduced motion: letters drawn once at their starting positions; no loop, no pointer handling. Listen for the setting changing.
- Server render outputs only the empty layer; letters are created after mount, so there is no hydration mismatch. Starting positions come from a seeded generator (same seed, same layout).
- Letter count is decided at mount; a resize only updates the bounds.
- Commits: `VynnLockwood <120536136+VynnLockwood@users.noreply.github.com>`, no AI trailers. Conventional Commits. Branch `feat/letter-field`. No new dependencies; `npm run check:design` stays green (no hex outside `:root`, no gradients).
- Letters never replace content: nothing in the page depends on them, and they are faint enough that text contrast is unchanged in practice (the letters sit under the content and most sections have their own background).

## Review Focus

- Click-through: no letter can ever be the target of a click, tap or focus; the layer never changes tab order or screen-reader output.
- Reduced motion: nothing moves, not even from the pointer; changing the setting while the page is open takes effect.
- Pointer handling: touch never pushes; leaving the window or blurring clears the pointer so letters do not stay pushed.
- Performance: one loop, no allocation per frame, transform only, stops when hidden, no listener or loop left behind when the component unmounts (route changes).
- Letters stay inside the viewport layer at 1280, 768, 390 and 360 px with no horizontal scroll, on `/`, `/th`, `/connect` and the dashboard.
- Existing checks that count or measure motifs ignore the letter field's letters.

---

### Task 1: Physics module

**Files:**
- Create: `apps/web/lib/letter-physics.ts`
- Test (outside the repo): `/home/claude/scratch/letter-physics.test.mjs`, run with `node --experimental-strip-types`

**Interfaces:**
- Produces: `type LetterState = { x: number; y: number; vx: number; vy: number; bx: number; by: number; rot: number; size: number; opacity: number }` (x, y = top-left in px); `type Pointer = { x: number; y: number } | null`; constants `RADIUS = 120`, `MAX_DT = 0.05`; `makeLetters(count: number, width: number, height: number, seed: number): LetterState[]`; `stepLetter(l: LetterState, dt: number, pointer: Pointer, width: number, height: number): LetterState` (returns a new object, never mutates).

- [ ] **Step 1: Write the failing test** `letter-physics.test.mjs` (plain `node:assert`): (a) `makeLetters(14, 1280, 800, 7)` returns 14 letters, each with size 22 to 40, opacity 0.35 to 0.7, speed `hypot(bx, by)` 6 to 16, `by <= 0` for at least 10 of 14 (mostly up), position inside the box; the same seed gives deep-equal output and a different seed does not; (b) with no pointer, after 10 s of 16 ms steps the velocity equals the base drift within 0.5 px/s; (c) a letter whose centre is 60 px to the right of the pointer ends up further right (x larger) after 0.5 s than the same letter with no pointer, and rotates away (rot differs); (d) a pointer 200 px away changes nothing (output deep-equals the no-pointer run); (e) the push is stronger when closer: displacement at 30 px is greater than at 90 px; (f) `stepLetter(l, 5, ...)` equals `stepLetter(l, 0.05, ...)` (clamp); (g) a letter that leaves the left edge reappears on the right (x > width), and the same for top and bottom; (h) the input object is unchanged after a step; (i) a pointer exactly on a letter's centre does not produce NaN (pushes up).
- [ ] **Step 2: Run it.** Expected: FAIL (module missing).
- [ ] **Step 3: Implement** `letter-physics.ts` with no imports: seeded generator (mulberry32), `makeLetters`, `stepLetter` (velocity relaxes toward the base with rate 2.5 per second; push acceleration `900 * (1 - d / RADIUS)^2` px per second squared along the away direction; rotation eases toward `±14°` away from the pointer scaled by `1 - d / RADIUS`, else toward 0, at rate 4 per second; position integrates; wrap with margin = size).
- [ ] **Step 4: Verify.** Test PASS; `npx tsc --noEmit` passes.
- [ ] **Step 5: Commit** `feat(web): add the letter physics module`.

### Task 2: The component and the browser checks

**Files:**
- Create: `apps/web/components/LetterField.tsx`
- Modify: `apps/web/app/layout.tsx` (mount once, inside `<body>` before `{children}`)
- Test (outside the repo): `/home/claude/scratch/letter-check.py`; adjust `/home/claude/scratch/motif-check.py` and `/home/claude/scratch/motion-check.py` to ignore `[data-letter-field] svg`.

**Interfaces:**
- Consumes: Task 1's functions, and `Letter` from `components/motifs`.
- Produces: `LetterField` (client component, no props); markers `data-letter-field` on the layer and `data-letter` on each letter wrapper.

- [ ] **Step 1: Write the failing test** `letter-check.py` (Playwright, against `next dev` or `next start`): (a) at 1280 px the layer has 14 `[data-letter]`, at 390 px 8; the layer is `position: fixed`, `pointer-events: none`, `aria-hidden="true"`, z-index below 0; each letter's size is 22 to 40 px and opacity 0.35 to 0.7; (b) `document.elementFromPoint` at the centre of every letter is not inside the layer, and the document is not horizontally scrollable at 1280, 390 and 360; (c) pointer: pick the letter nearest the middle of the viewport, move the mouse to a point 30 px to its left, wait 700 ms; the distance between its centre and the pointer is at least 30 px larger than at the start; moving the mouse far away (300 px) and waiting 700 ms moves the letter by no more than 25 px; (d) reduced motion (`reduced_motion="reduce"`): 14 letters present, and every letter's position is identical (within 0.5 px) at 0 ms and 1500 ms, also after moving the mouse next to one; (e) touch (`has_touch`, `is_mobile`, 390 px): letters drift (at least one moves 3 px or more in 1500 ms) and a tap next to one does not push it more than 25 px; (f) hidden tab: set `document.hidden` to true and dispatch `visibilitychange`; positions are identical (within 1 px) 1000 ms later; (g) the first server-rendered HTML contains the layer with no letters in it (fetch the page without JavaScript and look for `data-letter-field` and the absence of `data-letter=`).
- [ ] **Step 2: Run it.** Expected: FAIL on everything (no component).
- [ ] **Step 3: Implement** `LetterField.tsx` per the constraints: state `letters` set in an effect after measuring the window; one effect owns the loop (rAF, `performance.now()` deltas, refs to the letter elements, `el.style.transform = translate3d(x, y, 0) rotate(rotdeg)`); `pointermove` (mouse and pen only) sets the pointer, `pointerleave` on the document element and window `blur` clear it; `visibilitychange` cancels or restarts the loop; a `resize` listener updates width and height; `matchMedia("(prefers-reduced-motion: reduce)")` read at mount and on change (reduced: draw once, cancel the loop, clear the pointer); clean up every listener and the frame on unmount. Mount in `layout.tsx`.
- [ ] **Step 4: Verify.** `letter-check.py` PASS; `motif-check.py`, `motion-check.py`, `palette-check.py`, `review-fixes-check.py`, `brand-check.py` still PASS; `tsc`, `check:design`, `build` pass; screenshots of `/` at 1280 and 390 px (letters visible, faint, behind content) and a short frame sequence of a letter moving away from the pointer.
- [ ] **Step 5: Commit** `feat(web): add drifting letters that react to the cursor`.

### Task 3: Docs, review and PR

**Files:**
- Modify: `docs/design.md` (Motion bullet: the letter field, reduced motion, the motion-budget note), and fix the two stale sentences found in the V1 review (the `og.png` Known-gaps line; "Dashboard and Connect add no motion beyond the tab fade"), and state that `@view-transition` only affects full page loads.
- Create: `agent-logs/claude/2026-11-01-HHMM-letter-field.md`

- [ ] **Step 1: Docs.** Edit as above.
- [ ] **Step 2: Final checks:** web gates, Go suite (nothing in Go changes; run it anyway), all scratch Playwright tests.
- [ ] **Step 3: Whole-branch review** by a fresh Opus reviewer with this plan's Review Focus; one fix pass for Critical and Important findings, each with a failing test first; minors to the log.
- [ ] **Step 4: Log and PR** `feat(web): drifting letters that react to the cursor` with the template, the real CI result, and "Rulings I made" and "Deferred minors" in the final message.
