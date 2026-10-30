# Web design system refactor (Olympus)

Status: draft for owner review. Date: 2026-10-30. Branch: `feat/web-design-system`.

## Goal

Restyle every page of `apps/web` to match the owner's **nemosyne Design System** (Claude artifact "nemosyne", type Design System). Today the site uses a white theme with a blue accent. After this change it is dark, monochrome, square and hairline-based, as the system specifies.

## Decisions (agreed with the owner)

| Topic | Decision |
|---|---|
| Theme | Dark only. No light theme, no `prefers-color-scheme` switch. |
| Fonts | Barlow 500/600, Geist (variable), Geist Mono (variable) via self-hosted `@fontsource` packages. Prompt stays for Thai only. New dependencies approved by the owner. |
| Copy | Restyle only. No wording changes in EN or TH. |
| Scope | Landing (`/`, `/th`), `/dashboard`, `/connect`, header, footer. |
| Approach | Tokens first, restyle in place. Existing class names and JSX are kept. |

## Out of scope

- Copy, routes, API calls, behaviour, auth flow.
- `deploy/`, Dockerfiles, CI, `.env.example`.
- Favicons (`app/icon.png`, `app/apple-icon.png`) and `public/og.png`. They keep the old look; the PR notes this.
- A light theme.

## Source of truth

The artifact's `project/tokens.json` (dark values), `project/README.md` (visual foundations) and `project/components/bundle.css` (component patterns). Where this spec and the artifact disagree, the artifact wins, except for the Thai exception below.

## Design

### Tokens
- New `apps/web/app/tokens.css` holds the system's dark values as CSS custom properties, using the system's names (`--surface-0..3`, `--ink`, `--ink-muted`, `--ink-subtle`, `--line`, `--line-strong`, `--action`, `--on-action`, `--action-hover`, `--focus`, `--scrim`, `--success`, `--warning`, `--danger`, `--info`, `--space-1..10`, `--radius-sm|md|full`, `--shadow-overlay`, `--z-*`, font families).
- `globals.css` imports nothing else for colour. It contains no hex literals except inside `tokens.css`.
- Removed: the blue accent, gradients, soft shadows, 20px radii, pill buttons, the float/rise animations and the halo behind the logo.

### Typography
- Display (Barlow 600, uppercase, tracking 0.02-0.04em): hero and section headlines, card titles.
- Controls (Barlow 500, uppercase, tracking 0.14em): buttons, nav links, tabs, accordion triggers.
- Body (Geist): paragraphs, in `--ink-muted`. Headlines in `--ink`.
- Labels (Geist Mono, uppercase, tracking 0.12em): eyebrows, badges, table headers, field labels, numbers.
- **Thai exception:** Barlow and Geist have no Thai glyphs. In `.landing[lang="th"]` and any Thai text, headings and controls use Prompt, with no uppercase and no tracking (as today). Latin fallbacks are listed after Prompt so mixed text still renders.
- `layout.tsx` imports Barlow 500/600, Geist variable, Geist Mono variable, and only the Thai subsets of Prompt (weights 400/500/600/700 as used). Prompt Latin subsets are dropped.

### Components (restyle in place)
- **Buttons:** 2px radius, uppercase Barlow. Primary is `--action` on `--on-action`. Secondary is a 1px `--ink` outline that inverts on hover. Ghost is text only. Sizes 32/44/56 px. Arrows nudge 3px on hover.
- **Cards, inputs, dashboard table, URL box, auth card:** `--surface-1` fill, 1px `--line` border, no shadow. Hover brightens the border to `--line-strong`. Inputs use `--line-strong` borders.
- **Chips and badges:** mono, uppercase, 1px `--line-strong` border, 2px radius. Status colours only for feedback and always next to a word or icon.
- **Links:** `--ink` with a `--line-strong` underline.
- **Focus:** 2px solid `--focus`, 3px offset, on every interactive element.
- **Header:** 72px, `--surface-0`, bottom hairline, sticky. Nav links are Barlow, uppercase, tracked. Language switch becomes a square segmented control.
- **FAQ:** accordion with a plus/minus drawn by two hairlines.
- **Footer:** hairline columns, mono bottom row.
- **Motion:** 200 ms colour changes, 400 ms width/transform, `cubic-bezier(0.2, 0, 0, 1)`. No bounce, no parallax. All motion respects `prefers-reduced-motion`.

### Landing page
- Section eyebrows get zero-padded indexes (`01 / <existing eyebrow text>`). This is the only markup addition. The index is generated in `Landing.tsx` from section order, so `copy.ts` is unchanged.
- Hero: black ground with the faint hairline grid (system `Hero` pattern) instead of gradients. No halo, no float animation.
- Mock chat and summary cards: flat, hairline borders, no shadow, mono labels.
- Crosshair tick ornaments on the live module card only (system allows one to three ornaments per view).
- The `Sparkle` ornament is not used.

### Logo
- Generate `public/brand/nemo-logo-white.png` from `nemo-logo.png` by setting the ink to white and keeping alpha (the same recolour the artifact's `emblem-white.png` uses).
- Hero shows it at **256 px or smaller** (the system forbids larger). The 512 px file is used so 256 px renders sharp on 2x screens.
- Header mark: `nemo-mark.svg` switched to `currentColor` or a white variant so it reads on `--surface-0`.
- `<meta name="theme-color">` becomes `#000000`.

### Dashboard and Connect
Same tokens and components. Signed-in/out states, buttons and links are unchanged. The warning notice uses `--warning` with its label text.

## Verification (must run before the PR)

- `cd apps/web && npx tsc --noEmit` and `npm run build`.
- Playwright against `next start` with a stand-in API: screenshots of `/`, `/th`, `/dashboard`, `/connect` at 1440, 820 and 390 px.
- Computed contrast for every text/background pair used (target AA: 4.5:1 text, 3:1 controls).
- No horizontal overflow at 320 and 360 px.
- Behaviour checks re-run: sign-in buttons and links, signed-in "Open dashboard", language switch both ways, nav anchor, FAQ opens, skip link is the first Tab stop, `?auth_error` forwarding, one session request per load.
- Confirm no hex colours outside `tokens.css`, no `box-shadow` except the modal token, no `border-radius` above `--radius-md` except `--radius-full` on dots.

## Delivery

- Branch `feat/web-design-system`; PR title `refactor(web): apply nemosyne design system`; never push to `main`.
- Commits follow Conventional Commits. No AI co-author trailers (per `AGENTS.md`).
- Update `docs/design.md` (Branding and web section) and `README.md` if it mentions the theme.
- Add `agent-logs/claude/<date>-<time>-web-design-system.md` with commands run and results, plus what was not tested.
- PR lists: new dependencies (`@fontsource/barlow`, `@fontsource-variable/geist`, `@fontsource-variable/geist-mono`), the lockfile change, favicons and `og.png` left unchanged, and Thai typography not proofread natively.

## Risks and open items

- Barlow ships only 500 and 600 in the system; any heavier weight must fall back to 600.
- The source emblem is 256 px, so large uses stay capped; a vector original would remove that limit.
- Favicons and the social preview stay on the old style until the owner asks for a dark version.
- Thai display type uses Prompt, so it will not look identical to the Latin Barlow headlines.
