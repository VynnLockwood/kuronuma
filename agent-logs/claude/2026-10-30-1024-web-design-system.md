# Web: nemosyne design system (light), memory-first landing, cursor field

- Agent: claude
- Date: 2026-10-30
- Branch / PR: `feat/web-design-system` (PR link in the PR description)

## Goal
The owner asked to refactor all web design to match their "nemosyne" Design System artifact, then redirected: white (light) theme, Nemo represents memory (spending is one module), a cursor-reactive background with sprites, scroll animation, and a mission section about getting better at other kinds of data. The payee-category MCP tools were requested too and are deliberately **not** in this PR (separate API/migration work).

## What I did
- `apps/web/app/tokens.css`: the design system's light tokens. `globals.css` rewritten to use only those tokens (square corners, hairlines, no shadows, one ink-inverted action colour).
- Fonts via `@fontsource`: Barlow, Geist, Geist Mono (new dependencies). Prompt kept for Thai only (Thai text is never uppercased or tracked).
- Landing, dashboard, connect, header, footer restyled. Memory-first copy (EN rewritten, TH updated to match). New memory panel in the hero, new Mission section, `01 / ...` section indexes, asymmetric hero, label | content section layout.
- `HeroField.tsx` (cursor-lit grid, snapping crosshair tick, 5 dragonfly sprites), `Reveal.tsx` (scroll reveal via `data-reveal`), `ScrollProgress.tsx`. No animation library.
- `scripts/check-design.mjs` (`npm run check:design`): no hex colours, shadows, gradients or large radii outside the allowed rules.
- Logo shown at 256 px max (the design system's cap). `docs/design.md` updated (theme, positioning, motion exception, Mission). Spec and plan under `docs/superpowers/`.

## Verified
Run in `apps/web` against `next start` on port 3100 with `/api/auth/me` stubbed by Playwright (no real API):
- `npx tsc --noEmit`: ok
- `npm run build`: ok (all 7 routes prerendered)
- `npm run check:design`: ok
- Landing content/theme checks (`/` and `/th`): 15 pass, 0 fail
- Layout checks (two-column rhythm, no overlaps, no overflow at 320/360/390/820 px): 24 pass, 0 fail
- Interaction checks (pointer vars, tick snap, sprites wander and veer away, loop pauses off-screen/hidden, reveal on scroll, scroll progress, reduced motion, no JS, touch, no console errors): 15 pass, 0 fail
- Typography (Thai uses Prompt, no uppercase/tracking; EN Barlow/Geist/Geist Mono), WCAG contrast from the live tokens, 2px focus ring on every tab stop, skip link first, sign-in links, signed-in "Open dashboard", language switch both ways, `#how` anchor, FAQ opens, `?auth_error` forwards to `/dashboard`, one `/api/auth/me` request per load, no overflow on all four pages signed in/out at 320 and 360 px, long MCP URL stays in its box: 55 pass, 0 fail
- Thai font fallback (found in review): on `/th`, Latin letters in headings fell back to the OS font (Inter); fixed with `--font-thai` = Prompt, Geist. Test via Chrome DevTools `getPlatformFontsForNode`: RED (Inter used) then GREEN (Geist used); whole suite re-run green after the fix (15, 24, 15 and 55 checks, 0 failures; `tsc` and `check:design` ok).
- Stray hex search: only `themeColor` in `layout.tsx`.
- The Playwright scripts are in `/tmp` and are not committed (Playwright is not a repo dependency).

## Not verified / known gaps
- Web Docker image not built; real API and real Google sign-in not tested.
- Thai copy is not proofread by a native speaker.
- Favicons (`app/icon.png`, `app/apple-icon.png`) and `public/og.png` keep the earlier art.
- Not tested on real touch hardware or Safari/Firefox (Chromium only).
- Cursor sprites and scroll animation deliberately go beyond the design system's flat-motion rule (owner request); landing page only.

## Review
One fresh-context whole-branch review (Opus). No Critical findings. Important: Thai Latin fallback (fixed, above); commit author identity is the checkout's `Claude <noreply@anthropic.com>`, not the owner's (left for the owner to re-author or accept at squash). Deferred minors: footer text tracked on `/th`; page prints blank below the fold (reveal has no print override); sprites never settle back and move 2x faster on 120 Hz; media queries read once at mount; memory panel `role="img"` hides module status from screen readers; some dead code/markup; `getBoundingClientRect` every frame.

## Decisions and assumptions
- Needs owner approval (AGENTS.md): new dependencies `@fontsource/barlow`, `@fontsource-variable/geist`, `@fontsource-variable/geist-mono` (owner approved in chat).
- Dark-only was first agreed, then the owner asked for white; the design system's own light values are used.
- `RevealController` observes `[data-reveal]` attributes instead of wrapping elements, to keep list/grid markup intact.
- `.btn-ghost` maps to the system's outline "secondary" button.

## Next steps
- Owner: review PR, proofread Thai, decide on favicons and `og.png`.
- Separate PR: payee-category MCP tools (needs a migration, so owner approval first).
