# Kuronuma blush look and motion (PR V1) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give the site the Blush palette, six original SVG motifs and gentle motion, and give the Go-rendered consent and error pages the same look.

**Architecture:** The look lives in the existing shadcn theme tokens (`:root` in `globals.css`), so every page changes at once. Motifs are small React SVG components that take their colours from new `--motif-*` tokens. Motion is CSS keyframes and transitions only, off under `prefers-reduced-motion`. A contrast check inside `check-design.mjs` keeps the palette accessible.

**Tech Stack:** Next.js 16, Tailwind 4, shadcn-style components, Go `html/template`. No new dependencies.

**Spec:** `docs/superpowers/specs/2026-10-31-kuronuma-rebrand-design.md` (section 2). The letter field is PR V2, not here.

## Global Constraints

- Palette B, Blush. Final values (checked: every pair below passes WCAG AA, large text 3:1):
  `--background #fffafb`, `--foreground #2a2024`, `--card #ffffff`, `--popover #ffffff`, `--primary #c23b69`, `--primary-foreground #ffffff`, `--secondary`, `--muted`, `--accent all #fdf0f4`, their foregrounds `#2a2024`, `--muted-foreground #7a5c68`, `--border` and `--input #efd3dd`, `--hairline #f6e3ea`, `--ring #c23b69`, `--link #b03361`, `--link-hover #8f274d`, `--subtle #a08592` (large text only), `--ink-2 #5c4650`, `--live #34c759`, `--success #1a7f37` on `--success-muted #e8f5ec`, `--info #5646c4` on `--info-muted #efeaff`, `--destructive #b42318` on `--destructive-muted #fff4f2`; `themeColor` `#fffafb`.
- Motif tokens (hex only in `:root`): `--motif-gold #ffd66b` and `--motif-gold-line #c99a1d`; `--motif-leaf #9ad8ae` and `--motif-leaf-line #4f9a6a`; `--motif-rose #fbd3e0` and `--motif-rose-line #d98aa6`; `--motif-lilac #e5defa` and `--motif-lilac-line #8d7fd6`.
- `npm run check:design` stays the rule: hex colours only in the `:root` block, no gradients anywhere. No new dependencies.
- Motion: durations 150 to 600 ms from shared tokens (`--dur-quick 180ms`, `--dur-base 320ms`, `--dur-slow 600ms`) and one easing (`--ease-soft cubic-bezier(0.2, 0, 0, 1)`); animate only `transform` and `opacity`; hover-only effects sit inside `@media (hover: hover)`; everything continuous is off under `prefers-reduced-motion: reduce`.
- Thai text is never tracked tighter; no motion splits Thai words (the hero title is animated as its two existing parts).
- Motifs are `aria-hidden`, not focusable, and the layer that holds them has `pointer-events: none`.
- Consent and error pages: still no scripts and no external images apart from the logo; inline SVG only.
- Commits: `VynnLockwood <120536136+VynnLockwood@users.noreply.github.com>`, no AI trailers. Conventional Commits. Branch `feat/blush-look`.
- Owner approval: the consent page is sign-in code (Task 4). The spec lists this PR as needing it; approving the plan gives it.
- Spec deviation (ruling made here): the spec says "underline slides in". A slide needs a gradient or a background-size trick, which `check:design` forbids, so links fade their underline colour in instead.

## Review Focus

- Every text and background pair in the theme, the badges and the Go pages meets AA (the `subtle` colour only for large text); the focus ring is at least 3:1 on the page.
- Under reduced motion nothing animates and no text stays invisible waiting for an animation (hero text, reveal content).
- Motifs never sit over text, buttons or chips at 1280, 768, 390 and 360 px, and never cause horizontal scroll; the decorative layer cannot take a click.
- The Thai page keeps working: no negative letter-spacing, no split Thai words, the same motif placement.
- Hover lift and sparkle appear only where hover exists; keyboard focus keeps a visible ring on cards, buttons and tabs.

---

### Task 1: Blush tokens and the contrast guard

**Files:**
- Modify: `apps/web/app/globals.css` (`:root` values from Global Constraints; the comment block), `apps/web/app/layout.tsx` (`themeColor`), `apps/web/scripts/check-design.mjs`
- Test (outside the repo): `/home/claude/scratch/palette-check.py`

**Interfaces:**
- Produces: the theme tokens above, used by every later task; `check:design` also fails when a listed pair is below its ratio.

- [ ] **Step 1: Write the failing tests.** In `check-design.mjs`, parse the `:root` hex tokens and add: `--background` must equal `#fffafb` and `--primary` `#c23b69`; and a `contrast(a, b)` function (WCAG relative luminance) with these minimums: `foreground` on `background`, `muted`, `card` 7; `muted-foreground`, `ink-2`, `link`, `link-hover` on `background` and `muted` 4.5; `primary-foreground` on `primary` 4.5; `success` on `success-muted`, `info` on `info-muted`, `destructive` on `destructive-muted` 4.5; `subtle` on `background` and `muted` 3; `ring` and `primary` on `background` 3. `palette-check.py` (Playwright, `/` and `/th`, 1280 px): `body` background is `rgb(255, 250, 251)`; the first primary button's background is `rgb(194, 59, 105)`; a `.text-link` element's colour is `rgb(176, 51, 97)`; no element's computed colour or background is the old blue (`rgb(0, 113, 227)` or `rgb(0, 102, 204)`).
- [ ] **Step 2: Run both.** Expected: `check:design` and `palette-check.py` FAIL on the old values.
- [ ] **Step 3: Implement.** Set the token values; remove nothing else. Update the file's comment to "Kuronuma theme, Blush".
- [ ] **Step 4: Verify.** `npm run check:design` prints `check:design ok`; `palette-check.py` PASS; `tsc` and `build` pass; screenshots of `/`, `/th`, `/connect` and the signed-in dashboard at 1280 and 390 px show readable text, visible badges and a visible focus ring.
- [ ] **Step 5: Commit** `feat(web): switch the theme to the Blush palette`.

### Task 2: Motifs and where they sit

**Files:**
- Create: `apps/web/components/motifs/index.tsx`
- Modify: `apps/web/app/globals.css` (the `--motif-*` tokens), `apps/web/components/landing/Landing.tsx`
- Test (outside the repo): `/home/claude/scratch/motif-check.py`

**Interfaces:**
- Produces: `type MotifProps = { size?: number; className?: string; style?: React.CSSProperties }`; `Sparkle`, `Sprout`, `Cloud`, `Letter`, `Petal`, `Star` (each `(props: MotifProps) => JSX.Element`, viewBox `0 0 48 48`, default size 24); `MotifDivider(): JSX.Element` (Sparkle, Sprout, Sparkle at size 18, centred, `aria-hidden`). PR V2 reuses `Letter`.
- Shapes are the approved sketch's: sparkle `M24 4 C25.5 16 32 22.5 44 24 C32 25.5 25.5 32 24 44 C22.5 32 16 25.5 4 24 C16 22.5 22.5 16 24 4Z` (gold); sprout stem `M24 44 V22`, leaves `M24 26 C12 26 8 16 10 10 C20 10 25 16 24 26Z` and `M24 22 C24 14 30 8 40 8 C41 18 35 24 24 22Z` (leaf); cloud `M13 34 C6 34 5 24 12 23 C12 15 24 12 28 19 C35 15 43 21 39 28 C45 30 43 38 36 36 L14 36Z` (lilac); letter `rect x6 y12 w36 h25 rx4`, flap `M7 14 L24 27 L41 14`, heart `M24 35 C20 31 19 28 21.5 27 C23 26.5 24 28 24 28 C24 28 25 26.5 26.5 27 C29 28 28 31 24 35Z` moved up 3 (rose); petal `M24 40 C10 36 6 22 14 12 C22 14 28 20 24 40Z` with vein `M24 40 C30 28 30 20 24 10` (rose); star `M24 7 L28.5 18 L40.5 19 L31.3 26.8 L34.2 38.5 L24 32.2 L13.8 38.5 L16.7 26.8 L7.5 19 L19.5 18Z` (lilac). Fills and strokes use the `--motif-*` tokens via `var(...)`, stroke width 1.5 (2 for the petal).

- [ ] **Step 1: Write the failing test** `motif-check.py` (`/` and `/th`; 1280, 768, 390, 360 px): every one of the six motifs appears at least once (mark each with `data-motif="sparkle|sprout|cloud|letter|petal|star"`); every motif is an `svg` with `aria-hidden="true"`, `focusable="false"`, and is not reachable by Tab; the decorative layers have computed `pointer-events: none`; no motif's bounding box intersects the box of the `h1`, the hero paragraph, the hero buttons or the chips; `document.documentElement.scrollWidth <= innerWidth`.
- [ ] **Step 2: Run it.** Expected: FAIL, no motifs exist.
- [ ] **Step 3: Implement** the six components and `MotifDivider`. Placement in `Landing.tsx`: the hero `<section>` becomes `relative` with `overflow-x-clip`, holding one `aria-hidden` layer (`pointer-events-none absolute inset-0`, hidden below `sm`, with two corner motifs only below `md`) with Sparkle (left, top 18%), Star (right, top 12%), Petal (left, top 52%), Sprout (right, top 46%), kept outside the text column; `MotifDivider` above the FAQ and above the final call to action; Cloud in the footer's top right (size 56, low opacity); Letter and Sparkle beside the logo in the closing band.
- [ ] **Step 4: Verify.** `motif-check.py` PASS; `tsc`, `check:design`, `build` pass; screenshots of `/` and `/th` at 1280 and 390 px, looked at by eye.
- [ ] **Step 5: Commit** `feat(web): add original SVG motifs to the landing page`.

### Task 3: Motion

**Files:**
- Modify: `apps/web/app/globals.css`, `apps/web/components/ui/card.tsx`, `apps/web/components/ui/button.tsx`, `apps/web/components/ui/tabs.tsx`, `apps/web/components/landing/Landing.tsx`
- Test (outside the repo): `/home/claude/scratch/motion-check.py`

**Interfaces:**
- Consumes: Task 2's `Sparkle` and `data-motif` markers.
- Produces: `Card` gains `cute?: boolean` (default false): adds `group`, a hover lift (`-translate-y-1`) and a `Sparkle` in the top right that fades in on hover; CSS classes `.motif-twinkle` and `.motif-float` (custom properties `--delay`, `--dur`), `.hero-in` (custom property `--i`), `.btn-shimmer`.

- [ ] **Step 1: Write the failing test** `motion-check.py`: (a) motifs have `animation-name` `twinkle` or `float`, and a different `animation-delay` across motifs; (b) hero eyebrow, both h1 parts, paragraph and button row have `animation-name: rise-in` with increasing `animation-delay`; (c) hovering a landing card moves it up (transform changes) and its sparkle goes from opacity 0 to above 0.9, and neither happens when emulating no hover; (d) hovering the default button changes its `::after` animation to `shimmer`; (e) clicking an accordion question changes the `details` height gradually (a value strictly between closed and open at 80 ms) ; (f) switching a tab on `/connect` plays `fade-in` on the new panel; (g) the served CSS contains `@view-transition`; (h) with `reduced_motion='reduce'`: every motif, hero element and button has `animation-name: none` or a duration of at most 0.01 ms, and the h1, hero text and a `[data-reveal]` heading are fully opaque.
- [ ] **Step 2: Run it.** Expected: FAIL on (a) to (g); (h) passes already for reveal.
- [ ] **Step 3: Implement.** Tokens and `@keyframes twinkle`, `float`, `rise-in`, `shimmer`, `fade-in` in `globals.css`; `.hero-in` applied to the hero eyebrow, the two title parts, paragraph and button row with `--i` 0 to 5 (80 ms steps, `both` fill); reveal tuned to a 16 px rise and 640 ms; the shimmer is a flat skewed band as the button's `::after` (no gradient); link underline colour fades in on hover; the accordion opens smoothly with `interpolate-size: allow-keywords` and a `details::details-content` height transition (ignored where unsupported); `@view-transition { navigation: auto }` with a 240 ms cross-fade; the `(hover: hover)` guard on lift, sparkle and shimmer; a `prefers-reduced-motion` block that sets `animation: none` on the motion classes and `@view-transition { navigation: none }`. Pass `cute` to the landing's feature, module and step cards.
- [ ] **Step 4: Verify.** `motion-check.py` PASS; `tsc`, `check:design`, `build` pass; screenshots and a short look at a recorded frame sequence of the hero entrance.
- [ ] **Step 5: Commit** `feat(web): add gentle motion and a reduced-motion fallback`.

### Task 4: Blush consent and error pages (Go)

**Files:**
- Modify: `apps/api/internal/core/cerberus/consent.go`, `apps/api/internal/core/cerberus/consent_test.go`

**Interfaces:**
- Produces: colour constants in `consent.go` (`colBackground "#fffafb"`, `colInk "#2a2024"`, `colMuted "#7a5c68"`, `colPrimary "#c23b69"`, `colLink "#b03361"`, `colBorder "#efd3dd"`, `colSurface "#fdf0f4"`) used by `pageStyle`; two inline SVG motifs (Sparkle and Sprout, `aria-hidden="true"`) in the consent card and one in the error card.

- [ ] **Step 1: Write the failing tests:** `TestConsentPageHasMotifsAndNoScripts` (two `<svg` elements with `aria-hidden="true"`, no `<script`, the CSP unchanged); `TestErrorPageHasAMotif` (one svg); `TestPageColoursAreAccessible` (a WCAG `contrast` helper in the test; ink, muted and link on background and surface at least 4.5, white on primary at least 4.5, and the constants equal the values above).
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement** the constants, restyle `pageStyle` with them (flat, no gradients), and add the inline SVGs with the same paths as Task 2.
- [ ] **Step 4: Verify.** `/home/claude/scratch/gotest.sh` and `vet` pass, `gofmt -l apps/api` is empty; screenshots of the rendered consent (English, Thai) and error pages at 1280 and 390 px.
- [ ] **Step 5: Commit** `feat(cerberus): restyle the consent and error pages in Blush`.

### Task 5: Share image, docs, log and PR

**Files:**
- Modify: `apps/web/public/og.png` (regenerate, 1200 by 630: Blush background, logo, "Kuronuma", คุโรนุมะ, the tagline, two motifs), `docs/design.md` (Theme and Motion bullets)
- Create: `agent-logs/claude/2026-10-31-HHMM-blush-look.md`

- [ ] **Step 1: Regenerate `og.png`** with a Playwright screenshot script outside the repo; check the size and look at it.
- [ ] **Step 2: Update `docs/design.md`:** the Theme bullet lists the Blush tokens and the contrast rule; the Motion bullet lists the effects, the reduced-motion rule and the underline ruling.
- [ ] **Step 3: Final checks:** the web gates, the Go suite, `check:design`, the three scratch Playwright tests, the brand check from PR R still passing.
- [ ] **Step 4: Whole-branch review** by a fresh reviewer with this plan's Review Focus; one fix pass for Critical and Important findings, each with a failing test first.
- [ ] **Step 5: Write the log, open the PR** `feat(web): Blush look, original motifs and gentle motion` (template ticked: sign-in code approved; list every new English and Thai string, none expected; screenshots described; the CI result), wait for CI, put the real result in the body.
