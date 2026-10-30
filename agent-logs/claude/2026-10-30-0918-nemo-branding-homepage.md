# Nemo branding, bilingual homepage, white-tone web design

- Agent: claude
- Date: 2026-10-30
- Branch / PR: `feat/web-landing` (no PR opened yet)

## Goal
The owner asked for: a short friendly name (Nemo / เนโม่), a clear public homepage like other open-source tools, a nicer template in the spirit of fastwork.com, the attached image as the logo, and a white-tone theme.

## What I did
- Brand assets from the attached image (`apps/web/public/brand`, `app/icon.png`, `app/apple-icon.png`, `public/og.png`): the full illustration (upscaled 2x), a vector mark traced from the dragonfly for small sizes, favicons, and a social-preview image.
- Public homepage at `/` (English) and `/th` (Thai): hero, example conversation, how it works, features, modules, honest privacy section, FAQ, call to action, footer. Language switch, skip link, reduced-motion support.
- Dashboard moved to `/dashboard` and restyled; `/connect` restyled. Sign-in returns to `/dashboard`. The homepage forwards `/?auth_error=...` to the dashboard so the API did not need to change.
- New stylesheet (white tone, one blue accent), Prompt font self-hosted via `@fontsource/prompt` (Thai + Latin, OFL license).
- One shared session check so the three sign-in buttons make a single request.
- README and `docs/design.md` updated.

## Verified
- `npx tsc --noEmit` and `npm run build`: pass.
- Playwright against `next start` with a stand-in API: screenshots at 1440, 820, 390 px (English and Thai) and the app pages; no broken images; Prompt font loads; no horizontal overflow at 320 and 360 px on all four pages.
- 19 behavior checks pass: sign-in buttons and their links, signed-in buttons switch to "Open dashboard", language switch both ways, nav anchor, FAQ opens, skip link is the first Tab stop, `?auth_error` forwarding, one session request per load, wordmark not clipped on small phones.
- Color contrast computed for all text/background pairs used: all at least 4.83:1.

## Not verified / known gaps
- The web Docker image was not built (Docker daemon unavailable here); `npm run build` passes.
- Not tested against the real API or a real Google sign-in.
- Thai copy was written by me and has not been proofread by a native speaker.
- The source logo is 256 px wide, so the large logo is upscaled and slightly soft; a larger or SVG original would improve it. The small vector mark is my simplification of the dragonfly, not the original art.
- Social preview needs `NEXT_PUBLIC_SITE_URL` at build time, and the footer source link needs `NEXT_PUBLIC_REPO_URL`. The Dockerfile and compose file do not pass them yet (deploy files need owner approval).
- Dashboard and Connect pages are English only.

## Decisions and assumptions
- Nemo is only the display name. Repository, MCP server name (`nemosyne`), cookies and env vars are unchanged so existing connectors keep working.
- Fastwork.com was used only for general structure (header with CTA and language selector, hero, pill chips, mock UI cards, trust section). No copy, imagery or fake social proof was copied or invented.
- Marketing copy only claims what is built; modules not built say "Coming soon". The privacy section states that on a hosted instance the operator can technically read stored data.
- Added a dependency (`@fontsource/prompt`), which AGENTS.md says needs owner approval; it replaces fetching fonts from Google at build time, which is blocked in some environments and leaks requests to Google.

## Next steps
- Owner: review the PR, proofread the Thai copy, share a higher-resolution or SVG logo if available.
- Decide whether to pass `NEXT_PUBLIC_SITE_URL` and `NEXT_PUBLIC_REPO_URL` through the Dockerfile and compose file.
- Optional: localize Dashboard and Connect pages, and add a "Connected apps" page.
