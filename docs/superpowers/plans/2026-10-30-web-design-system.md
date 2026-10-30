# Web Design System Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restyle every page of `apps/web` to the owner's nemosyne Design System (dark, monochrome, square, hairline) without changing copy, routes or behaviour.

**Architecture:** Add `app/tokens.css` with the system's dark tokens. Rewrite `app/globals.css` against those tokens, keeping existing class names so JSX barely changes. Small TSX edits add section indexes, cap the hero logo at 256 px and add crosshair ticks. A node script enforces the "no hex / shadow / large radius outside tokens" rules.

**Tech Stack:** Next.js 16, React 19, TypeScript, plain CSS, `@fontsource` fonts. Verification with Playwright (preinstalled, not added to the repo).

**Spec:** `docs/superpowers/specs/2026-10-30-web-design-system-design.md`

## Global Constraints

- Light only, using the design system's light token values. No dark theme, no `prefers-color-scheme`. `themeColor` is `#ffffff`. (Revision 2)
- Copy is memory-first (Revision 2): Nemo is the memory for your AI; Spending is the first module. Describe only what is built; unbuilt modules say "Coming soon"; no invented stats. No route, API or sign-in behaviour changes.
- Fonts: `@fontsource/barlow` (500, 600), `@fontsource-variable/geist`, `@fontsource-variable/geist-mono`. Keep `@fontsource/prompt` Thai subsets only (400, 500, 600, 700); drop its Latin imports.
- No hex colours in `globals.css` (only in `tokens.css`). No `box-shadow` except `var(--shadow-overlay)`. No `border-radius` above `var(--radius-md)` except `var(--radius-full)` on dots. No `gradient` except `.hero::before` and `.hero-field` rules.
- Display type: Barlow 600 uppercase, tracking 0.02-0.04em. Controls: Barlow 500 uppercase, tracking 0.14em. Labels: Geist Mono uppercase, tracking 0.12em. Thai (`[lang="th"]`): Prompt, no uppercase, no tracking.
- Logo: never rendered larger than 256 CSS px. Use the original black-ink `nemo-logo.png` and ink `nemo-mark.svg`.
- Motion: 200 ms colour, 400 ms width/transform, `cubic-bezier(0.2, 0, 0, 1)`; respect `prefers-reduced-motion`. Landing page only: cursor-reactive hero field and scroll reveals (owner override of the system's flat-motion rule). No new dependencies.
- Payee-category MCP tools are out of scope (separate plan/PR).
- Do not touch `deploy/`, Dockerfiles, CI, `.env.example`, favicons, `public/og.png`.
- Work on branch `feat/web-design-system`; never push to `main`; no AI `Co-Authored-By` trailers; Conventional Commits.

## Review Focus

Failure modes the spec implies but nothing else exercises, most likely first:

1. Reduced motion, touch and no-JS: with `prefers-reduced-motion: reduce` sprites are static and reveals do not hide content; with JS disabled all content is visible; on touch (no fine pointer) nothing throws and nothing is stuck mid-fade. Pinned in Task 5.
2. Thai pages: headings, buttons, nav, badges and eyebrows are not uppercased or tracked and use Prompt. Pinned in Task 6.
3. Narrow phones (320 px): sprites, the memory panel and tracked buttons must not cause horizontal scroll. Pinned in Task 6.
4. Performance: the sprite loop stops when the hero is off-screen or the tab is hidden. Pinned in Task 5.
5. Contrast on white: `--ink-muted`/`--ink-subtle` on `--surface-0..2` >= 4.5:1, `--line-strong` >= 3:1, status colours >= 4.5:1 on `--surface-1`. Pinned in Task 6.

---

## File Structure

- Create `apps/web/app/tokens.css`: the system's dark custom properties only.
- Rewrite `apps/web/app/globals.css`: all component and page styles, tokens only.
- Create `apps/web/scripts/check-design.mjs`: static conformance check, run as `npm run check:design`.
- Modify `apps/web/app/layout.tsx`: font imports, `themeColor`.
- Modify `apps/web/components/landing/Landing.tsx` and `copy.ts`: memory-first content, mission section, section indexes, ticks.
- Create `apps/web/components/landing/HeroField.tsx`, `Reveal.tsx`, `ScrollProgress.tsx` (client components, no dependencies).
- Modify `apps/web/public/brand/nemo-mark.svg` (back to ink fill); delete `apps/web/public/brand/nemo-logo-white.png`.
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

### Task 3: Light tokens and memory-first content

**Files:**
- Modify: `apps/web/app/tokens.css`, `apps/web/app/layout.tsx`, `apps/web/public/brand/nemo-mark.svg`, `apps/web/components/landing/copy.ts`, `apps/web/components/landing/Landing.tsx`
- Delete: `apps/web/public/brand/nemo-logo-white.png`

**Interfaces:**
- Consumes: Task 1 token names (unchanged); Task 2 classes.
- Produces: `Copy` type changes in `copy.ts`: remove `mock.month`, `mock.categories`, `mock.total`; add `hero.panelTitle: string`, `hero.panelNote: string`, and `mission: { eyebrow: string; title: string; sub: string; items: { title: string; body: string }[]; note: string }`. Add `{ id: "mission", label: "Mission" }` to `nav` (after Modules). `Eyebrow({ index, children })` and `SectionHead({ index, eyebrow, title, sub })` in `Landing.tsx`. Section indexes: demo 01, how 02, features 03, modules 04, mission 05, privacy 06, faq 07.

- [ ] **Step 1: Write the failing test** `/tmp/verify/t3.cjs` (Playwright, outside the repo) asserting on `/` and `/th`: `.eyebrow-index` texts equal `["01".."07"]`; `#mission` exists; body computed background is `rgb(255, 255, 255)` and `h1` colour `rgb(0, 0, 0)`; hero logo `src` ends with `/brand/nemo-logo.png` and renders at most 256 px wide; four `.tick-corner` elements, all inside `.module-card.live`; hero contains all four module names (Plutus, Atlas, Horae, Hephaestus) in the memory panel; `/` hero text matches `/memory/i` and the hero contains no `฿` amount.
- [ ] **Step 2: Run it to confirm it fails** (build, start `next start --port 3100`, run). Expected: FAIL on theme, mission, panel and indexes.
- [ ] **Step 3: Flip `tokens.css` to the light values** from the artifact (`--surface-0: #ffffff; --surface-1: #f5f5f5; --surface-2: #ebebeb; --surface-3: #dedede; --ink: #000000; --ink-muted: #474747; --ink-subtle: #5e5e5e; --line: #d6d6d6; --line-strong: #767676; --action: #000000; --on-action: #ffffff; --action-hover: #333333; --focus: #000000; --success: #0d7a4a; --warning: #8a5a00; --danger: #c4281c; --info: #1558b0; --shadow-overlay: 0 24px 64px #00000040`), `color-scheme: light`; `layout.tsx` `themeColor` `#ffffff`; `nemo-mark.svg` fill `#000000`; delete the white logo.
- [ ] **Step 4: Rewrite `copy.ts` EN and TH** to the new `Copy` shape, memory-first. EN hero title stays "Your assistant\u2019s memory, owned by you."; `sub` says Nemo gives your AI a memory it can read and write through MCP, and that spending is the first thing it remembers; `chips` drop "Slip photos never stored" in favour of "Memory you own"; `panelTitle` "Memory"; features reframed around memory (connect once, remember across assistants, you approve, open source), with the slip features moved under the Spending module wording; FAQ and footer lead with memory; mission copy per spec. TH mirrors EN meaning.
- [ ] **Step 5: Update `Landing.tsx`.** Hero visual: replace the money summary card with a memory panel (`.memory-panel`) listing `c.modules.items` with status badges (Spending marked available, others "Coming soon"), headed by `c.hero.panelTitle`; keep the "Saved to Nemo" toast. Demo section heading framed as a Spending example (copy only). Add the mission section (`id="mission"`, `SectionHead index={5}`, a list of `c.mission.items`, and `c.mission.note`). Hero logo `src` `/brand/nemo-logo.png`, 256x256. Keep the Eyebrow/ticks edits already in the working tree.
- [ ] **Step 6: Run the test to confirm it passes.** Expected: all PASS.
- [ ] **Step 7: Verify and commit**
  Run: `cd apps/web && npm run check:design && npx tsc --noEmit && npm run build`
  Expected: all succeed.
  ```bash
  git add -A apps/web && git commit -m "refactor(web): light theme and memory-first landing content"
  ```

---

### Task 4: Layout restyle

**Files:**
- Modify: `apps/web/app/globals.css`

**Interfaces:**
- Consumes: Task 3 markup (`.memory-panel`, mission section, `.eyebrow-index`).
- Produces: class contracts for Task 5: `.hero` is `position: relative`; `.hero-field` is an absolutely positioned layer inside `.hero` with `pointer-events: none`.

- [ ] **Step 1: Write the failing check.** Extend `/tmp/verify/t3.cjs` with: at 1440 px the hero text column is left-aligned (h1 left edge equals the container's left edge), `.section-head` and section content sit in two columns (head width <= 40% of container) for sections 03-07, and no horizontal overflow at 320, 360 and 390 px for `/` and `/th`. Run it; expected FAIL on the two-column rhythm.
- [ ] **Step 2: Restyle.** Asymmetric hero (copy column wider, memory panel offset down); sections 03-07 use a `label | content` two-column grid (sticky head on desktop, stacked below 960 px); `.memory-panel` flat hairline panel with mono module rows; mission section as a hairline-ruled list; hero h1 `clamp(2.5rem, 5.4vw, 4.5rem)` with `.accent` as a 2px underline (`text-decoration`), memory panel and toast positioned so nothing overlaps the 256 px emblem; no gradients except `.hero::before` and `.hero-field`.
- [ ] **Step 3: Run the check, `npm run check:design`, tsc and build; take screenshots at 1440/820/390 and read them.** Expected: PASS and no overlaps or overflow.
- [ ] **Step 4: Commit**
  ```bash
  git add apps/web && git commit -m "refactor(web): asymmetric layout and ruled sections"
  ```

---

### Task 5: Cursor field and scroll animation

**Files:**
- Create: `apps/web/components/landing/HeroField.tsx`, `apps/web/components/landing/Reveal.tsx`, `apps/web/components/landing/ScrollProgress.tsx`
- Modify: `apps/web/components/landing/Landing.tsx`, `apps/web/app/globals.css`, `apps/web/scripts/check-design.mjs`

**Interfaces:**
- Produces: `HeroField(): JSX.Element` (client component rendering `<div className="hero-field" aria-hidden="true">` with a grid-light layer, one `.field-tick`, five `.sprite` SVGs; sets CSS variables `--mx`, `--my` (px) on its parent `.hero`; sets `data-active="true|false"` on itself). `Reveal({ children, delay }: { children: React.ReactNode; delay?: number }): JSX.Element` (client; wraps in `<div data-reveal data-in?>`; adds class `reveal-ready` to `document.documentElement` on mount; sets `data-in` when intersecting; applies `transition-delay` of `delay` ms). `ScrollProgress(): JSX.Element` (client; renders `<div className="scroll-progress" role="presentation">` whose `--p` custom property is scroll fraction 0..1).
- Consumes: Task 4 `.hero` / `.hero-field` contract.

- [ ] **Step 1: Write the failing tests** (Playwright, `/tmp/verify/t5.cjs`): (a) moving the mouse to (900, 300) over the hero sets non-empty `--mx`/`--my` on `.hero` and moves `.field-tick`\u2019s bounding box to within 40 px of that point; (b) five `.sprite` elements exist; one sprite\u2019s bounding box changes over 600 ms; with the mouse placed on a sprite, its distance from the pointer is larger 600 ms later; (c) after scrolling the hero out of view, `.hero-field[data-active]` becomes `"false"` and sprite boxes stop changing; (d) with `reducedMotion: "reduce"`, sprite boxes do not change over 600 ms and every `[data-reveal]` has computed opacity `1`; (e) with `javaScriptEnabled: false`, `[data-reveal]` has computed opacity `1`; (f) at 1440 px, a `[data-reveal]` below the fold has opacity `0` before scrolling and `1` (with `data-in`) after scrolling to it; (g) `.scroll-progress` `--p` is `0` at top and `> 0.9` at the bottom; (h) with a touch context (`hasTouch: true`, no mouse) no console errors occur. Run: expected FAIL (nothing exists).
- [ ] **Step 2: Implement `HeroField`.** Pointer tracking on the parent `.hero` via `pointermove`/`pointerleave` (only when `matchMedia("(pointer: fine)")` matches), throttled with one `requestAnimationFrame`. Tick position snaps to the nearest grid node (grid pitch `--space-8`) with a 200 ms transition. Sprites: five inline-SVG dragonflies built from the `nemo-mark.svg` polygons (stroke-only, `currentColor` ink, 28-44 px), each with position/velocity state; each frame: wander (small random steering), repel from the pointer within 140 px, soft-bound to the hero box. An `IntersectionObserver` on `.hero` and `document.visibilitychange` stop the loop and set `data-active="false"`. With `prefers-reduced-motion: reduce` or no fine pointer, render sprites at fixed seeded positions and never start the loop.
- [ ] **Step 3: Implement `Reveal` and `ScrollProgress`; wire into `Landing.tsx`.** Wrap each section head, card grid item and the mission list items with `Reveal` (stagger `delay = index * 60`). Render `ScrollProgress` inside the sticky header. CSS: `.reveal-ready [data-reveal]:not([data-in]) { opacity: 0; transform: translateY(12px); }`, transition 400 ms `--ease`; `.scroll-progress` 1 px, `transform: scaleX(var(--p))`, origin left, `--ink`; `.hero-field` grid-light via a darker grid masked by `radial-gradient(circle 160px at var(--mx) var(--my), ...)`; `.sprite` and `.field-tick` styles; all disabled under `prefers-reduced-motion: reduce`. Update `check-design.mjs` allowlist to `.hero::before` and selectors starting with `.hero-field`.
- [ ] **Step 4: Run the tests and `npm run check:design`, tsc, build.** Expected: all PASS.
- [ ] **Step 5: Commit**
  ```bash
  git add apps/web && git commit -m "feat(web): cursor-reactive hero field and scroll reveals"
  ```

---

### Task 6: Verification, docs, agent log and PR

**Files:**
- Create: `/tmp/verify/check.mjs` (not committed), `agent-logs/claude/<date>-<HHMM>-web-design-system.md`
- Modify: `docs/design.md`, `README.md` only if it mentions the old theme

- [ ] **Step 1: Write `check.mjs`** with: Thai typography (on `/th`, `h1`, `.btn-primary`, `.nav a`, `.badge`, `.eyebrow` have `text-transform: none`, `letter-spacing` `normal` or `0px`, `font-family` starting `Prompt`); English typography (`h1` uppercase and `Barlow`; `.eyebrow` `Geist Mono`); overflow (scrollWidth <= innerWidth at 320 and 360, on `/`, `/th`, `/dashboard`, `/connect`, signed out and in); long MCP URL on `/connect` stays inside its box at 320; every focused element while tabbing has a 2px solid outline; reduced motion shows `document.getAnimations().length === 0` and `transition-duration` `0s` on `.btn` hover; logos at most 256 px; behaviour checks (sign-in links to `/api/auth/google/login?next=%2Fdashboard`, "Open dashboard" when signed in, language switch both ways, `#how` anchor, skip link first Tab stop, FAQ `details` opens, `/?auth_error=access_denied` forwards to `/dashboard`, one `/api/auth/me` request per load); contrast ratios from computed token values (see Review Focus 5); static check (`npm run check:design` exit 0, no hex outside `tokens.css` and `themeColor`).
- [ ] **Step 2: Run everything.** `cd apps/web && npx tsc --noEmit && npm run build && npm run check:design && node /tmp/verify/t3.cjs && node /tmp/verify/t5.cjs && node /tmp/verify/check.mjs` Expected: all exit 0. Fix failures in the owning file, re-run, commit fixes as `fix(web): ...`.
- [ ] **Step 3: Docs.** `docs/design.md`: replace the Theme bullet with the light design system (tokens in `app/tokens.css`, Barlow/Geist/Geist Mono plus Prompt for Thai, logo cap 256 px), add the memory-first positioning, the cursor field and scroll reveals (landing only, reduced-motion safe) and a **Mission** section (Nemo gets better at more kinds of data: tasks, routines, work data; none promised by date). Check every linked path exists.
- [ ] **Step 4: Agent log** from `agent-logs/README.md` template: goal, changes, every command and result from Step 2, not-verified items (Docker build, real API/Google sign-in, Thai copy not proofread natively, favicons/`og.png` unchanged, new dependencies need owner approval, payee tools not part of this PR).
- [ ] **Step 5: Commit, push and open the PR**
  ```bash
  git add docs agent-logs README.md && git commit -m "docs(web): document light design system, memory positioning and mission"
  git push -u origin feat/web-design-system
  gh pr create --title "refactor(web): light design system, memory-first landing, cursor field" --body-file <filled .github/pull_request_template.md>
  ```
  Expected: PR opens against `main`; body lists the three new font dependencies for owner approval and the Step 2 results; do not merge. Wait for CI and report.
