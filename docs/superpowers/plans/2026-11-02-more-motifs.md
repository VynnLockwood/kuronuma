# More motifs, placed with rules (PR M) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline) or superpowers:subagent-driven-development. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Add 16 new original motifs (teddy bear, paper plane, heart, moon, bunny, cat, cupcake, strawberry, balloon, rainbow, daisy, songbird, teacup, mushroom, bow, pencil) and place them on every page so the site looks cute and still calm: a few well-chosen spots, never a scatter.

**The drawings:** `docs/superpowers/plans/assets/new-motifs-preview.png` shows all 16 large and at three small sizes. `assets/new-motifs-source.py` is the SVG source (48 by 48, flat fills, 1.5 round outline, colours from tokens). Port the paths verbatim into `components/motifs/`.

**Architecture:** the new motifs join `components/motifs/index.tsx` (same `Frame`, `data-motif`, `aria-hidden`, `pointer-events-none` as the six existing ones). Two new decorative colour pairs, `--motif-honey` and `--motif-sky`, join the `:root` block. Placement uses two small layout pieces so every page follows the same grammar: `Spot` (one motif centred above a heading or a block, in flow, so it can never overlap text) and `Accents` (a mirrored pair of small motifs in the outer margins, from `lg` up only). The Go consent and error pages get inline SVG versions of one motif each.

## Placement grammar (the rules that keep it tidy)

1. **Three sizes only.** Spot 64 px (96 px in an empty state); accent 32 px; micro 20 px (dividers). Nothing else.
2. **One spot per section.** A section has at most one `Spot`. Accents come in mirrored pairs, only at `lg` and wider, and only in margins that are empty at that width.
3. **In flow, not over.** A spot is a block in the layout (it pushes content down); it is never absolutely positioned over text. Accents sit in the page margin outside the content column, never over a card or text.
4. **Meaning first.** Each motif is chosen for what the section says (map below). A motif is not repeated on the same page except in dividers.
5. **Budget.** At most 5 decorative motifs are in the viewport at once on any page at any scroll position (the letter field is excluded), measured by a test.
6. **Calm where data lives.** Lists, tables and forms in the dashboard get none. Only the sign-in card, the unreachable notice, empty states and the 404 page get motifs.
7. **Motion.** Only the landing page motifs animate (twinkle or float, existing classes, at most four at once); the dashboard, connect and 404 pages are still. Reduced motion stops all of it.
8. **No new colour outside tokens.** `check:design` fails on a hex colour or gradient in any motif.

## Placement map

| Page | Where | Motif |
|---|---|---|
| Landing hero | stays as built (sparkle, star in the top band; petal, sprout on the product shot) | existing |
| Landing, "Ask in plain words" | under the left column's paragraph, `md` and up | teddy (96) with a small plane flying toward the cards |
| Landing, How it works | spot above the heading | plane |
| Landing, Features | spot above the heading | heart |
| Landing, Modules | one 36 px motif top-right of each module card: Plutus strawberry, Atlas pencil, Horae moon, Hephaestus mushroom (the hover sparkle moves to the bottom-right on these cards) | per card |
| Landing, Mission | spot above the heading, left column | rainbow |
| Landing, Privacy | spot above the heading | cat |
| Landing, FAQ | spot above the heading | teacup |
| Landing, closing band | bunny and balloon accents at the band's lower corners, `lg` and up; the existing letter, logo, sparkle row stays | bunny, balloon |
| Dividers | the divider trio becomes a rotating set of three micro motifs | bow, daisy, bird, and the existing |
| `/connect` | spot above the h1; spot above "Then just ask." | plane; teacup is used on the landing so use bird here |
| Dashboard sign-in card | bow and sparkle flank the logo (like the closing band) | bow, sparkle |
| Dashboard unreachable | spot above the title | plane (lost, with its dashed trail) |
| Overview empty state | spot (96) above the title | bunny |
| Spending empty state | spot (96) | cupcake |
| Tasks and notes empty state | spot (96) | pencil |
| Tasks and notes "no match" and Spending "no match" | spot (64) | mushroom |
| 404 page (new `app/not-found.tsx`, English and Thai, `noindex`) | spot (96) above the title | bunny peeking, with a link home |
| Go error page | replaces the sparkle | plane |
| Go consent page | unchanged (sparkle and sprout) | existing |
| Share image `og.png` | add the teddy, the plane and the heart beside the logo | regenerate |

Later dashboard PRs (J, K) take their empty states from this set: Categories empty: balloon; Payee rules empty: heart; Connected assistants empty: plane. They are not built here.

## Global Constraints

- Every motif: `aria-hidden="true"`, `focusable="false"`, no `tabindex`, `pointer-events: none`, `data-motif` set, colours only through `var(--...)`.
- No new dependency. `npm run check:design` stays green and now also scans `components/motifs/` for hex colours and gradients.
- Thai pages get the same placement, no layout shift between languages.
- No existing character, logo or artwork is reproduced; every motif here is an original drawing.
- Commits: `VynnLockwood <120536136+VynnLockwood@users.noreply.github.com>`, no AI trailers. Conventional Commits. Branch `feat/more-motifs`.
- Owner approval needed (AGENTS.md): the Go consent and error pages are sign-in code (markup only, one inline SVG changed). Approving this plan gives it.

## Review Focus

- Messiness: the viewport budget (5) holds on every page at 1280, 1024, 768, 390 and 360 px, in English and Thai, at every scroll position; sizes are only the three tiers; one spot per section; accents only in empty margins; nothing overlaps text, buttons, chips, inputs or card content (the existing overlap check, now on all pages and states).
- Calm: the populated dashboard (lists, tables, forms) has no motif; nothing animates there; reduced motion stops all motion on the landing page.
- Accessibility: all motifs hidden from assistive technology, never focusable, never take a click; spots do not change heading order, tab order or reading order; contrast of text is unaffected (nothing sits behind text).
- Layout: no horizontal scroll; no layout shift when the language changes; the closing band's accents never touch the text column; module cards keep their badge and title readable; the 404 page has a working link home.
- Code: motifs are plain presentational components with no state; no hex or gradient in any motif; the Go pages still run no script and the CSP is unchanged; bundle growth is small (report the size).
- The drifting letter layer, the Blush palette, the existing six motifs and all earlier checks are unaffected.

---

### Task 1: The 16 motifs and the colours

**Files:** Modify `apps/web/components/motifs/index.tsx`, `apps/web/app/globals.css` (`--motif-honey`, `--motif-honey-line` `#b9824f`, `--motif-sky`, `--motif-sky-line` `#6da9d1`; honey fill `#f2cfa0`, sky fill `#cfe8f8`), `apps/web/scripts/check-design.mjs`. Test (outside the repo): `/home/claude/scratch/motif-catalogue-check.py`.

**Interfaces:** Produces `Teddy, Plane, Heart, Moon, Bunny, Cat, Cupcake, Strawberry, Balloon, Rainbow, Daisy, Bird, Teacup, Mushroom, Bow, Pencil` (each `(props: MotifProps)`), the new tokens, and the `check:design` rule for motifs.

- [ ] **Step 1: Write the failing tests.** `check-design.mjs`: fail when any file under `components/motifs/` contains a hex colour or `gradient`, and when `--motif-honey` or `--motif-sky` (and their `-line`) are missing from `:root`. `motif-catalogue-check.py`: a temporary page is not allowed in the repo, so the check reads the built component source instead: each of the 22 motifs is exported, has a unique `data-motif` name, a `viewBox="0 0 48 48"` and no hex colour; the test is run through `node --experimental-strip-types` against the file.
- [ ] **Step 2: Run them.** Expected: FAIL (tokens and components missing).
- [ ] **Step 3: Implement.** Port the SVG paths from `assets/new-motifs-source.py` verbatim; use `paint()` and the token names; the teddy's `translate(0 -1.2)` and the rainbow's `scale(.9)` groups stay.
- [ ] **Step 4: Verify.** `tsc`, `check:design`, `build`; the catalogue check PASS; render the 22 motifs once in a scratch page (outside the repo) at 132, 40, 28 and 20 px and look at it.
- [ ] **Step 5: Commit** `feat(web): add sixteen original motifs`.

### Task 2: `Spot`, `Accents` and the landing page

**Files:** Create `apps/web/components/motifs/Spot.tsx` (exports `Spot` and `Accents`); modify `components/landing/Landing.tsx`, `components/ui/card.tsx` only if the module cards need the moved hover sparkle. Tests: `/home/claude/scratch/placement-check.py` (new, covers every page in later tasks too); update `motif-check.py` and `motion-check.py` to tolerate the new motifs.

**Interfaces:** `Spot({ motif, size?: 64 | 96, className? })` renders a centred in-flow block (`aria-hidden`, `pointer-events-none`); `Accents({ left, right })` renders the mirrored pair, hidden below `lg`, positioned in the margin outside the content column.

- [ ] **Step 1: Write the failing test** `placement-check.py` (Playwright, `/` and `/th`; 1280, 1024, 768, 390, 360 px): every motif outside the letter field is `aria-hidden`, not focusable, `pointer-events: none`; each is one of the three sizes (width within 1 px of 20, 32, 36 (module cards), 64 or 96); no motif overlaps any text element, button, link, input or chip (rects intersect test, as `motif-check.py`); at most one spot per section and accents only at 1024 px and wider; the viewport budget: scrolling in 300 px steps, at most 5 decorative motifs intersect the viewport at once; the map above is followed (each section's motif has the expected `data-motif`); no horizontal scroll; the English and Thai pages place the same motifs in the same sections; reduced motion: no motif has a running animation.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement** `Spot`, `Accents` and the landing placements from the map, including the module card motifs and the closing band accents.
- [ ] **Step 4: Verify.** `placement-check.py` PASS for `/` and `/th`; all earlier checks pass; screenshots of every section at 1280 and 390 px reviewed by eye, English and Thai.
- [ ] **Step 5: Commit** `feat(web): place the new motifs on the landing page`.

### Task 3: Connect, dashboard states and the 404 page

**Files:** Modify `app/connect/page.tsx`, `components/dashboard/Shell.tsx` (sign-in card, unreachable), `Overview.tsx`, `Spending.tsx`, `Memories.tsx` (empty and no-match states); create `app/not-found.tsx` and the Thai text for it. Extend `placement-check.py` and the dashboard checks.

- [ ] **Step 1: Write the failing tests.** `placement-check.py` adds `/connect`; the dashboard sign-in card, the unreachable notice, the three empty states and the two no-match states (English and Thai, 1280 and 390, faked API as in the dashboard checks): each has exactly the expected motif, sizes follow the grammar, no overlap, budget holds; the populated Overview, Spending and Tasks tabs contain **no** motif; `/does-not-exist` returns the 404 page with the bunny, a link home (`/`) and Thai text, `noindex`, no horizontal scroll.
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Verify.** All dashboard, placement and earlier checks PASS; screenshots reviewed.
- [ ] **Step 5: Commit** `feat(web): add motifs to the connect page, dashboard states and a 404 page`.

### Task 4: Go pages, share image, docs

**Files:** Modify `apps/api/internal/core/cerberus/consent.go` (error page uses the plane), `consent_test.go`; regenerate `apps/web/public/og.png` (script outside the repo); `docs/design.md` (a "Motifs and where they go" paragraph: the grammar and the map).

- [ ] **Step 1: Write the failing test.** `TestErrorPageHasAMotif` expects the plane (a `<svg aria-hidden="true"` with the plane's marker) instead of the sparkle; the consent page keeps its two.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement** the inline plane SVG (same paths, literal colours as in the other inline motifs), the new `og.png` (1200 by 630), the docs paragraph.
- [ ] **Step 4: Verify.** `gotest.sh`, vet, gofmt; look at the rendered error page and the new share image.
- [ ] **Step 5: Commit** `feat(cerberus): use the paper plane on the error page and refresh the share image`.

### Task 5: Review and PR

- [ ] Final checks (web gates, Go suite, every scratch Playwright test), whole-branch review by a fresh Opus reviewer with this Review Focus (one fix pass, tests first), agent log, PR `feat(web): sixteen new motifs, placed with rules` with the template (check Auth / OAuth for the error page markup), the real CI result, bundle size, "Rulings I made" and "Deferred minors" in the final message.

## Not included

- A picture of Kuronuma Sawako or any other existing character (see the PR description of this plan).
- More themes (paper crane, sakura branch, seed packet and so on): a second batch, if wanted.
- Motifs in the drifting letter layer (it stays envelopes only, to stay calm).
- The module path rename (`kuronuma-mcp` to `kuronuma`): a separate PR.
