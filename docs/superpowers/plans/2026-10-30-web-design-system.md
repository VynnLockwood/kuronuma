# Web Design System Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restyle every page of `apps/web` to the owner's nemosyne Design System (dark, monochrome, square, hairline) without changing copy, routes or behaviour.

**Architecture:** Add `app/tokens.css` with the system's dark tokens. Rewrite `app/globals.css` against those tokens, keeping existing class names so JSX barely changes. Small TSX edits add section indexes, cap the hero logo at 256 px and add crosshair ticks. A node script enforces the "no hex / shadow / large radius outside tokens" rules.

**Tech Stack:** Next.js 16, React 19, TypeScript, plain CSS, `@fontsource` fonts. Verification with Playwright (preinstalled, not added to the repo).

**Spec:** `docs/superpowers/specs/2026-10-30-web-design-system-design.md`

## Global Constraints

- Dark only. No light theme, no `prefers-color-scheme`. `themeColor` is `#000000`.
- No copy changes in `components/landing/copy.ts` or any page text. No route, API or behaviour changes.
- Fonts: `@fontsource/barlow` (500, 600), `@fontsource-variable/geist`, `@fontsource-variable/geist-mono`. Keep `@fontsource/prompt` Thai subsets only (400, 500, 600, 700); drop its Latin imports.
- No hex colours in `globals.css` (only in `tokens.css`). No `box-shadow` except `var(--shadow-overlay)`. No `border-radius` above `var(--radius-md)` except `var(--radius-full)` on dots. No `gradient` except the Hero hairline grid.
- Display type: Barlow 600 uppercase, tracking 0.02-0.04em. Controls: Barlow 500 uppercase, tracking 0.14em. Labels: Geist Mono uppercase, tracking 0.12em. Thai (`[lang="th"]`): Prompt, no uppercase, no tracking.
- Logo: never rendered larger than 256 CSS px.
- Motion: 200 ms colour, 400 ms width/transform, `cubic-bezier(0.2, 0, 0, 1)`; respect `prefers-reduced-motion`.
- Do not touch `deploy/`, Dockerfiles, CI, `.env.example`, favicons, `public/og.png`.
- Work on branch `feat/web-design-system`; never push to `main`; no AI `Co-Authored-By` trailers; Conventional Commits.

## Review Focus

Failure modes the spec implies but nothing else exercises, most likely first:

1. Thai pages: headings, buttons, nav, badges and eyebrows must not be uppercased or tracked, and must use Prompt (Barlow/Geist have no Thai glyphs). Pinned in Task 4.
2. Narrow phones (320 px): uppercase tracked `nowrap` buttons and the language switch must not cause horizontal scroll. Pinned in Task 4.
3. Long unbroken strings (the MCP URL in `/connect`) must wrap or scroll inside their box, not widen the page. Pinned in Task 4.
4. Keyboard focus: every link, button, `summary` and language-switch link shows a visible 2px focus ring on black. Pinned in Task 4.
5. `prefers-reduced-motion: reduce`: no running animations or transitions. Pinned in Task 4.

---

## File Structure

- Create `apps/web/app/tokens.css`: the system's dark custom properties only.
- Rewrite `apps/web/app/globals.css`: all component and page styles, tokens only.
- Create `apps/web/scripts/check-design.mjs`: static conformance check, run as `npm run check:design`.
- Modify `apps/web/app/layout.tsx`: font imports, `themeColor`.
- Modify `apps/web/components/landing/Landing.tsx`: section indexes, hero logo, ticks.
- Modify `apps/web/public/brand/nemo-mark.svg` (white fill); create `apps/web/public/brand/nemo-logo-white.png`.
- Modify `apps/web/package.json`, `package-lock.json`: three font packages, `check:design` script.
- Modify `docs/design.md`; create `agent-logs/claude/<date>-<time>-web-design-system.md`.

---

### Task 1: Foundation (tokens, fonts, assets, conformance check)

**Files:**
- Create: `apps/web/app/tokens.css`, `apps/web/scripts/check-design.mjs`, `apps/web/public/brand/nemo-logo-white.png`
- Modify: `apps/web/app/layout.tsx`, `apps/web/package.json`, `apps/web/package-lock.json`, `apps/web/public/brand/nemo-mark.svg`

**Interfaces:**
- Produces: CSS custom properties named exactly as in the artifact's `tokens.json`: `--surface-0..3`, `--ink`, `--ink-muted`, `--ink-subtle`, `--line`, `--line-strong`, `--action`, `--on-action`, `--action-hover`, `--focus`, `--scrim`, `--success`, `--warning`, `--danger`, `--info`, `--space-1..10`, `--radius-none|sm|md|full`, `--shadow-overlay`, `--z-nav|overlay|modal`, `--font-display`, `--font-sans`, `--font-mono`, `--font-thai`. Also `npm run check:design` (exit 1 with a list of violations).
- Consumes: nothing.

- [ ] **Step 1: Write the failing conformance check** in `scripts/check-design.mjs`. It reads `app/globals.css` and fails (exit 1, one line per violation with line number) on: a hex colour literal; `box-shadow` other than `none` or `var(--shadow-overlay)`; `linear-gradient`/`radial-gradient` outside a rule whose selector is `.bx-hero__grid` or `.hero::before`; a `border-radius` value that is a number of px/rem above 4px (allow `0`, `1px`, `2px`, `4px`, `var(--radius-*)`, `50%` only on selectors containing `dot`/`avatar`). Add `"check:design": "node scripts/check-design.mjs"` to `package.json` scripts.
- [ ] **Step 2: Run it against the current CSS to confirm it fails**
  Run: `cd apps/web && npm run check:design`
  Expected: exit 1, many violations (hex colours, shadows, 20px radii).
- [ ] **Step 3: Install fonts**
  Run: `cd apps/web && npm install @fontsource/barlow @fontsource-variable/geist @fontsource-variable/geist-mono`
  Expected: `package.json` and `package-lock.json` change; no other dependency changes.
- [ ] **Step 4: Create `app/tokens.css`** with a single `:root` block holding every token from the artifact's `project/tokens.json`, using the **dark** values verbatim (e.g. `--surface-0: #000000; --surface-1: #0a0a0a; --surface-2: #141414; --surface-3: #1f1f1f; --ink: #ffffff; --ink-muted: #b3b3b3; --ink-subtle: #8f8f8f; --line: #262626; --line-strong: #707070; --action: #ffffff; --on-action: #000000; --action-hover: #d4d4d4; --focus: #ffffff; --scrim: #000000b3; --success: #5fd39a; --warning: #f2b84b; --danger: #ff6a5c; --info: #6db6ff; --shadow-overlay: 0 24px 64px #000000cc`), spacing 4/8/12/16/24/32/48/64/96/128 px, radii 0/2px/4px/9999px, z-index 50/100/110. Font stacks: `--font-display: "Barlow", "Prompt", "Arial Narrow", system-ui, sans-serif`; `--font-sans: "Geist Variable", "Prompt", system-ui, -apple-system, sans-serif`; `--font-mono: "Geist Mono Variable", "Prompt", ui-monospace, "SF Mono", Menlo, monospace`; `--font-thai: "Prompt", system-ui, sans-serif`. Also `color-scheme: dark`.
- [ ] **Step 5: Update `app/layout.tsx`**: import `@fontsource/barlow/latin-500.css`, `latin-600.css`, `@fontsource-variable/geist/index.css`, `@fontsource-variable/geist-mono/index.css`, keep only the four `@fontsource/prompt/thai-*.css` imports, import `./tokens.css` before `./globals.css`, set `viewport.themeColor` to `"#000000"`.
- [ ] **Step 6: Generate the white logo and mark.** With Pillow, open `public/brand/nemo-logo.png` (RGBA 512x512), set every pixel's RGB to 255 and keep alpha, save as `public/brand/nemo-logo-white.png`. In `nemo-mark.svg` change `fill="#0b1220"` to `fill="#ffffff"`.
  Run: `python3 -c "from PIL import Image; im=Image.open('public/brand/nemo-logo-white.png'); print(im.size, im.getextrema())"`
  Expected: `(512, 512)` and RGB extrema `(255, 255)` with alpha spanning 0..255.
- [ ] **Step 7: Verify and commit**
  Run: `cd apps/web && npx tsc --noEmit && npm run build`
  Expected: both succeed (CSS still old, fonts now load).
  ```bash
  git add apps/web && git commit -m "build(web): add design-system tokens, fonts and conformance check"
  ```

---

### Task 2: Shell, controls and app pages in `globals.css`

**Files:**
- Modify: `apps/web/app/globals.css` (replace the old `:root` block and the base, header, nav, language switch, buttons, footer, `.page`, `.auth-card`, `.who`, `.connect-card`, `.notice`, `.url-card`, `.urlrow`, `code`, `.guide`, `.steps` rules)

**Interfaces:**
- Consumes: Task 1 tokens. Class names stay exactly as they are today (`.btn`, `.btn-primary`, `.btn-ghost`, `.btn-sm`, `.btn-lg`, `.site-header`, `.header-row`, `.logo`, `.logo-word`, `.nav`, `.header-actions`, `.lang`, `.site-footer`, `.footer-grid`, `.footer-bottom`, `.page`, `.page-h1`, `.page-h2`, `.auth-card`, `.who`, `.avatar-fallback`, `.connect-card`, `.notice`, `.url-card`, `.urlrow`, `.code`, `.guide-grid`, `.steps`, `.card`, `.muted`, `.skip`, `.container`).
- Produces: base typography and controls that Task 3 builds on.

- [ ] **Step 1: Write the failing screenshot check.** Create `/tmp/verify/shot.mjs` (outside the repo) that starts from `npx next start --port 3100` in `apps/web`, intercepts `**/api/auth/me` with 401 (signed-out) or a stub user JSON (signed-in), and saves full-page screenshots of `/dashboard` and `/connect` at 1440, 820 and 390 px into `/tmp/verify/out`. Run it now and look at one image.
  Expected: still the white theme (this is the "before").
- [ ] **Step 2: Rewrite the base layer.** `body`: `--surface-0` ground, `--ink-muted` text, `--font-sans`, 16/24. Headings `--ink`. `.display` rules: `h1`, `h2`, `.card h3` use `--font-display`, 600, uppercase, tracking 0.02em (h1) to 0.04em (card titles). Links `--ink` with `text-decoration-color: var(--line-strong)`, underline offset 4px. `:focus-visible { outline: 2px solid var(--focus); outline-offset: 3px; }`. Keep `.skip`, `.container`, `.muted`. Add `html[lang="th"], [lang="th"]` overrides: headings, buttons, nav, badges, eyebrows use `--font-thai`, `text-transform: none`, `letter-spacing: 0`.
- [ ] **Step 3: Header, nav, language switch, buttons, footer.** Header 72 px, `--surface-0`, bottom `--line`, `position: sticky`, `z-index: var(--z-nav)`, no blur. Nav links and buttons per the Global Constraints (Barlow 500, uppercase, 0.14em, 13-14 px). Buttons: radius `--radius-sm`, heights 32/44/56 px; primary `--action` on `--on-action` with `--action-hover`; ghost is text-only; add `.btn-ghost` secondary outline `1px solid var(--ink)` that inverts on hover; `.btn:disabled` opacity 0.35; arrow icon inside buttons nudges 3px on hover. `.lang`: square segmented control with `--line-strong` border, current item `--action`/`--on-action`. Footer: `--surface-0`, top `--line`, mono bottom row. Buttons use `white-space: normal` below 480 px so labels can wrap.
- [ ] **Step 4: App pages.** `.card`, `.auth-card`, `.connect-card`, `.url-card`: `--surface-1`, 1px `--line`, `--radius-sm`, no shadow; hover on link-cards brightens border to `--line-strong`. `.notice` uses `--warning` text and border with `--surface-1` fill. `.urlrow code` and `.code` use `--font-mono`, `--surface-2`, `overflow-wrap: anywhere`. `.avatar-fallback` square-edged with `--radius-full` allowed. `.steps` numerals in `--font-mono`.
- [ ] **Step 5: Verify**
  Run: `cd apps/web && npx tsc --noEmit && npm run build && node /tmp/verify/shot.mjs`
  Expected: build passes; dashboard and connect screenshots show black ground, white primary buttons, square hairline cards, no blue.
- [ ] **Step 6: Commit**
  ```bash
  git add apps/web/app/globals.css && git commit -m "refactor(web): restyle shell, controls and app pages with design tokens"
  ```

---

### Task 3: Landing page styles and markup

**Files:**
- Modify: `apps/web/app/globals.css` (hero, mocks, chips, sections, cards, steps, features, modules, badges, privacy, FAQ, CTA band, responsive `@media`, reduced-motion), `apps/web/components/landing/Landing.tsx`, `apps/web/components/Logo.tsx`

**Interfaces:**
- Consumes: Task 2 base. `SectionHead({ eyebrow, title, sub }: { eyebrow: string; title: string; sub?: string })` is extended with `index: number`.
- Produces: `SectionHead({ index, eyebrow, title, sub })` rendering `<p class="eyebrow"><span class="eyebrow-index">0{index}</span> / {eyebrow}</p>`; `demo-copy` eyebrow gets index 1 the same way. Indexes follow section order: demo 01, how 02, features 03, modules 04, privacy 05, faq 06.

- [ ] **Step 1: Run the conformance check and confirm it still fails** (landing rules still old).
  Run: `cd apps/web && npm run check:design`
  Expected: exit 1 listing violations in the landing rules only.
- [ ] **Step 2: Markup edits in `Landing.tsx`.**
  - Add the `index` prop to `SectionHead`; pass 2-6 from the existing call sites; render the demo eyebrow with index 1 via a small shared `Eyebrow({ index, children })` component in the same file.
  - Hero eyebrow keeps its sparkle icon (a stroke icon in `currentColor`) and drops the `pill` class; style it as a plain mono eyebrow.
  - Hero logo `<img>`: `src="/brand/nemo-logo-white.png"`, `width={256} height={256}`, same `alt`.
  - Add `<span className="tick tl" aria-hidden />` and `tr`/`bl`/`br` siblings inside the `module-card live` item only (class names `.bx-tick`-style: `.tick-corner`, `.tick-tl`, `.tick-tr`, `.tick-bl`, `.tick-br`; do not reuse the existing `.tick` check-mark class).
  - Keep the `accent` span on the second hero title word; style `.accent` as `--ink` with a 2px `--ink` bottom border (no colour accent exists).
  - `Logo.tsx` stays on `nemo-mark.svg` (now white); no change unless the build shows a problem.
- [ ] **Step 3: Landing CSS.** Hero: `--surface-0`, hairline grid via `.hero::before` (two `linear-gradient`s of `var(--line)` at `var(--space-8)` pitch, opacity 0.35, masked to fade at top and bottom), no halo, no `float`/`rise` keyframes, `.hero-logo` max width 256 px. `.mock`, `.float-toast`, `.mock-summary`, `.bubble-*`, `.facts`, `.slip`: `--surface-1`, 1px `--line`, no shadow; toast keeps its position. `.chips li` and `.badge`: mono 11 px uppercase, `--line-strong` border, `--radius-sm`; `.badge-ok` uses `--success` text and border. `.section` padding `var(--space-9)` desktop and `var(--space-8)` mobile; `.section.surface` uses `--surface-1` with top/bottom `--line`. `.section-head` left-aligned, h2 per display style, `.eyebrow` mono 12 px `--ink-muted`, `.eyebrow-index` `--ink`. Cards flat; `.feature:hover, .module-card:hover` only brighten the border, no lift. `.module-card.live`: `--line-strong` border, `position: relative`, ticks 11 px `--ink` crosshairs offset -6 px at the corners. `.step-num` mono, square `--radius-sm`, `--action` fill. `.faq details`: no box, top `--line-strong` on the group, `--line` between items, plus/minus drawn with two 1px hairlines (replace the rotating chevron rule). `.cta-band`: `--surface-1`, 1px `--line-strong`, no gradient, no large radius. `.note` uses `--warning`. Drop every `@keyframes`; keep one `transition` set per control at 200 ms.
- [ ] **Step 4: Responsive and motion.** Rewrite the `@media (max-width: 960px)` and `(max-width: 640px)` blocks for the new sizes (grids collapse to one column, header nav hides as today, hero padding `var(--space-7) var(--space-4)`). Add `@media (prefers-reduced-motion: reduce) { *, *::before, *::after { animation: none !important; transition: none !important; scroll-behavior: auto !important; } }`.
- [ ] **Step 5: Run the conformance check**
  Run: `cd apps/web && npm run check:design`
  Expected: exit 0, no output.
- [ ] **Step 6: Verify build and screenshots**
  Run: `cd apps/web && npx tsc --noEmit && npm run build && node /tmp/verify/shot.mjs`
  Expected: build passes; `/` and `/th` screenshots at 1440/820/390 show black ground, uppercase Barlow headlines (Thai in Prompt, sentence case), `01 / …` eyebrows, hero emblem at most 256 px, four crosshair ticks only on the live module card.
- [ ] **Step 7: Commit**
  ```bash
  git add apps/web && git commit -m "refactor(web): restyle landing page with design system"
  ```

---

### Task 4: Behaviour and design checks, docs, log, PR

**Files:**
- Create: `/tmp/verify/check.mjs` (not committed), `agent-logs/claude/2026-10-30-<HHMM>-web-design-system.md`
- Modify: `docs/design.md` (Branding and web section), `README.md` only if it mentions the white theme or Prompt

**Interfaces:**
- Consumes: the running site on port 3100 with `/api/auth/me` stubbed by Playwright route interception.

- [ ] **Step 1: Write `check.mjs` with these assertions (all must pass).** Pinned to Review Focus:
  - Thai typography: on `/th`, `h1`, `.btn-primary`, `.nav a`, `.badge`, `.eyebrow` have computed `text-transform: none`, `letter-spacing` `normal` or `0px`, and `font-family` starting with `Prompt`.
  - English typography: on `/`, `h1` computed `text-transform: uppercase` and `font-family` starts with `Barlow`; `.eyebrow` starts with `Geist Mono`.
  - Overflow: at viewport widths 320 and 360, for `/`, `/th`, `/dashboard`, `/connect`, both signed-out and signed-in, `document.documentElement.scrollWidth <= innerWidth`.
  - Long string: on `/connect`, the MCP URL element's `scrollWidth <= its container's clientWidth + 1` or it has `overflow-x: auto`, and page scrollWidth stays within the viewport at 320.
  - Focus: tabbing through `/` and `/connect`, every focused element has a computed `outline-style: solid` and `outline-width: 2px` (includes `summary` and the language links).
  - Reduced motion: with `reducedMotion: 'reduce'`, `document.getAnimations().length === 0` after load and hovering a `.btn` shows computed `transition-duration` of `0s`.
  - Logo cap: every `img[src*="nemo-logo"]` has rendered width <= 256.
  - Behaviour (from the earlier branding work): sign-in buttons link to `/api/auth/google/login?next=%2Fdashboard`; signed-in state shows "Open dashboard"; language switch EN<->TH works; nav anchor `#how` scrolls; first Tab stop is the skip link; a FAQ `details` opens on click; `/?auth_error=access_denied` forwards to `/dashboard`; exactly one `/api/auth/me` request per page load.
  - Contrast: compute WCAG ratios for `--ink-muted` and `--ink-subtle` on `--surface-0..2`, and each status colour on `--surface-1`; all >= 4.5:1; `--line-strong` on `--surface-0..2` >= 3:1.
  - Static: `npm run check:design` exits 0; `rg -n '#[0-9a-fA-F]{3,8}' apps/web/app apps/web/components` finds hex values only in `tokens.css` and `layout.tsx`'s `themeColor`.
- [ ] **Step 2: Run everything**
  Run: `cd apps/web && npx tsc --noEmit && npm run build && npm run check:design && node /tmp/verify/check.mjs`
  Expected: all exit 0; the script prints each assertion with PASS. Fix any failure in the owning CSS/TSX, re-run, and commit fixes as `fix(web): ...`.
- [ ] **Step 3: Update docs.** In `docs/design.md` replace the "Theme" bullet (white tone, blue accent, Prompt only) with: dark-only nemosyne design system, tokens in `app/tokens.css`, fonts Barlow/Geist/Geist Mono plus Prompt for Thai, logo cap 256 px, and list favicons/`og.png` as still on the old light-theme art. Update the Known gaps line if needed. Check every linked path exists.
- [ ] **Step 4: Write the agent log** from the template in `agent-logs/README.md`: goal, what changed, every command run with its result from Step 2, and not-verified items (Docker image build, real API/Google sign-in, Thai typography not proofread natively, favicons/`og.png` unchanged, new dependencies need owner approval).
- [ ] **Step 5: Commit, push the branch and open the PR**
  ```bash
  git add docs agent-logs README.md && git commit -m "docs(web): document design system and add agent log"
  git push -u origin feat/web-design-system
  gh pr create --title "refactor(web): apply nemosyne design system" --body-file <filled .github/pull_request_template.md>
  ```
  Expected: PR opens against `main`; body lists new dependencies for owner approval and attaches the Step 2 results; do not merge. Wait for CI and report its result.
