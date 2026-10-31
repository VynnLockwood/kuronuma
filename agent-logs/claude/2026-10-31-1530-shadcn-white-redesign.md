# Web redesign: white, store-style UI on shadcn/ui

- Agent: claude
- Date: 2026-10-31
- Branch / PR: claude/cool-hamilton-tbfyqx, PR not opened yet

## Goal
The owner asked for a rework of all web pages on shadcn/ui, with a white tone in the spirit of a product store page. They reviewed and approved design mockups of every page (Landing, Sign in, Dashboard, Connect, page states, UI kit) before the code change.

## What I did
- Added Tailwind CSS v4 (`tailwindcss`, `@tailwindcss/postcss`, `postcss`) and shadcn/ui building blocks (`class-variance-authority`, `clsx`, `tailwind-merge`, `@radix-ui/react-slot`, `@radix-ui/react-tabs`). Added `@fontsource-variable/inter`; removed `@fontsource/barlow` and `@fontsource-variable/geist`, which are no longer used. The owner approved the dependency change together with the redesign.
- `app/globals.css`: replaced the old 1.5k-line stylesheet with shadcn theme tokens (white, `#F5F5F7` grey, blue action colour, large radius), the Tailwind theme mapping, the scroll-reveal rules and the reduced-motion rules. Removed `app/tokens.css`.
- `components/ui/`: Button, Card, Badge, Input, Tabs (Radix), Alert, Skeleton, Accordion (native `<details>`, no JS). `lib/utils.ts` adds `cn`.
- Rebuilt `Landing.tsx` (`/` and `/th`; copy unchanged, still from `copy.ts`), `dashboard/page.tsx` (sign-in card, loading skeleton, offline state with retry, signed-in view), `connect/page.tsx` (copy button, client tabs). Same data flow, auth calls and `?auth_error` handling as before.
- Removed `HeroField.tsx` (cursor grid) and `ScrollProgress.tsx`; kept the scroll reveal. New `AppNav.tsx`; `SiteHeader`, `Logo`, `AuthButton` and `icons.tsx` (new clock, copy, alert, unplug and Google icons) updated.
- `scripts/check-design.mjs`: rules now match the new theme (hex colours only in the `:root` block of globals.css, none in TSX, no gradients).
- `docs/design.md`: Logo, Theme and Landing motion bullets updated.

## Verified
- `npx tsc --noEmit`: clean.
- `npm run check:design`: ok.
- `npm run build`: succeeds, all 8 routes are static.
- Browser smoke test (Playwright, `next start`, no API running): `/`, `/th`, `/connect`, `/dashboard` at 1440 px and 390 px have no horizontal overflow and no page errors. The only console errors are 500s from `/api/auth/me`, because the Go API was not running. The Connect tab switch works and the MCP URL field shows `<origin>/mcp`. The dashboard was also checked with a mocked `/api/auth/me`: 401 plus `?auth_error=access_denied` shows the sign-in card with the error and cleans the URL; 200 shows the signed-in view.

## Not verified / known gaps
- Not tested against the real API, real Google sign-in or the Docker image build.
- No dark theme.
- The Go-rendered consent page (Cerberus) still has the old look; changing it needs separate approval.

## Decisions and assumptions
- Commits use the owner's git identity with no AI co-author trailer, as the owner asked and as AGENTS.md requires.
- Kept every line of the existing English and Thai copy; only the layout and styling changed.
- shadcn components were written by hand to match the shadcn source, not generated with the shadcn CLI, so there is no `components.json` yet.

## Next steps
- Open a PR and run CI, including the Docker image build.
- Optionally restyle the Cerberus consent page to match.
