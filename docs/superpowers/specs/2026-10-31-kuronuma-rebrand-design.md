# Kuronuma: rebrand, blush look and motion (design)

Date: 2026-10-31. Status: for owner review. Builds on `docs/design.md` and the web redesign (shadcn/ui, white store style).

## Goal

Turn Nemosyne ("Nemo") into **Kuronuma**: a new public name (site, assistants' connector, repository), a softer blush look with original cute SVG motifs, and gentle effects and transitions. The tone of the site stays the same: calm, plain, honest about data.

## Decisions (from the owner)

| Question | Decision |
|---|---|
| Names | Repository `kuronuma-mcp`; address `kuronuma.lunariankids.com`; MCP server name `Kuronuma`; Thai name คุโรนุมะ. |
| Palette | B, "Blush": pink-tinted white, rose action colour. |
| Extra effect | Paper letters drifting in the background that react to the cursor. |
| Cute elements | Original motifs only (sparkle, sprout, cloud, letter, petal, star). No mascot. |
| Greek module names | Kept (Plutus, Themis, Atlas, Cerberus, Hermes, ...). The landing section about them is reworded as a plain "modules" section. |

Boundaries: the art is original. It does not depict the character that inspired the name and does not reuse fan art. The character's name, face and series branding stay out of the public repository and site.

Not in this work: a new logo (the current illustration stays for now), a mascot, a dark theme, the dashboard features (separate spec), the security fixes S2 and S3 (separate PRs).

## 1. Rename (PR R, appearance unchanged)

Changes:

- **Copy:** every English and Thai string (`components/landing/copy.ts`, pages, `lib/brand.ts`): "Nemo" and "Nemosyne" become "Kuronuma", "เนโม่" becomes "คุโรนุมะ". `BRAND` loses the "short for" split: `name` and `full` are both "Kuronuma". The FAQ "Why 'Nemosyne'?" becomes "Why 'Kuronuma'?" with neutral wording (a gentle, quiet name that suits a memory); the "Nemo is the friendly name of Nemosyne" footer line goes. The "pantheon of modules" section becomes "Modules", keeping the module names as labels.
- **Assistant-facing:** MCP server `Name: "kuronuma"`, `Title: "Kuronuma"`; `hermes.DefaultInstructions` and tool texts say Kuronuma; the Cerberus consent and error pages say Kuronuma (English and Thai); tool names are unchanged.
- **Identifiers:** session and flow cookies `nem_session` and `nem_flow` become `kuro_session` and `kuro_flow` (everyone signs in again); OAuth client id prefix `nem_c_` becomes `kuro_c_` for new clients (stored ids keep working); the HMAC domain strings (`nemosyne-oauth-flow-v1:`, the consent domain) become `kuronuma-...` (flows in progress are invalidated).
- **Code and build:** Go module `github.com/VynnLockwood/kuronuma-mcp/apps/api` (go.mod and every import); `cmd/nemosyne-api` becomes `cmd/kuronuma-api` (Dockerfile, CI); web package `kuronuma-web`; nginx include `nemosyne-proxy.conf` becomes `kuronuma-proxy.conf`; CI image names; logo file `nemo-logo.png` becomes `kuronuma-logo.png`; the unused `nemo-mark.*` files are removed; `og.png` is regenerated with the new name.
- **Docs:** README, AGENTS.md, `docs/design.md`.

Deliberately unchanged:

- **Database name and user (`nemosyne`) and Docker Compose's project name (`name: nemosyne`).** The Postgres data volume is named after the Compose project. Renaming either would create a new empty volume and leave the data behind. Users never see these names. A comment in `deploy/docker-compose.yml` says why.
- Old agent logs and `docs/superpowers/**` (history).
- If the final public repository lives under another owner, the module path changes in one scripted step (go.mod and imports).

Check: `git grep -I -i -E 'nemosyne|\bnemo\b'` outside the historical folders may match only the allowed lines (the database and Compose names and the comment explaining them). The script's output goes in the PR.

## 2. Blush look and motion (PR V1)

- **Palette:** new values for the existing shadcn tokens in `globals.css` (the only place colours are defined; `check:design` keeps enforcing it). Starting point: background `#fffafb`, ink `#2a2024`, soft surface `#fdf0f4`, borders `#efd3dd`, action rose about `#c23b69` (the lighter `#d6527c` in the sketch fails contrast for white text), muted text `#7a5c68`. Final values are chosen with a contrast script (WCAG AA for every text and background pair, large text 3:1); its output goes in the PR. Status colours (success, info, destructive) are kept readable on the blush background.
- **Motifs:** `components/motifs/` holds Sparkle, Sprout, Cloud, Letter, Petal and Star as small React SVG components, drawn for this project. Colours come from CSS variables, they are `aria-hidden`, and none carry meaning. Uses: floating accents around the hero, small dividers between sections, a tiny sparkle on card hover, a cloud in the footer.
- **Motion** (CSS and SVG only):

| Where | Effect |
|---|---|
| Hero | Words fade and rise in with a short stagger |
| Sparkles and stars | Slow twinkle (opacity and scale), 3 to 5 s loops |
| Motifs | Gentle float, a few pixels, 6 to 9 s loops, out of phase |
| Scroll reveal | Existing reveal gets a soft rise and fade with a stagger |
| Cards | Lift and a tiny sparkle on hover, only where hover exists |
| Buttons | Soft scale and a quick shimmer sweep on hover and focus |
| Links | Underline slides in |
| Tabs and accordion | Cross-fade for tab content, smooth height for the accordion |
| Navigation | `@view-transition { navigation: auto }`: pages cross-fade where supported, ignored elsewhere |

- **Rules:** durations 150 to 600 ms with shared easing tokens; animate only transform and opacity; every effect is off under `prefers-reduced-motion: reduce`; hover-only effects are guarded by `(hover: hover)`.
- **Consent and error pages** (Go-rendered) get the blush palette and two inline SVG motifs. They still allow no scripts and no external images apart from the logo.
- **No new dependencies.**

## 3. Letters that react to the cursor (PR V2)

- **What:** a background layer of about 14 small paper letters (the Letter motif, rose outline, opacity 0.35 to 0.7, 22 to 40 px) that drift slowly across the page. Fewer on narrow screens (about 8).
- **Motion:** each letter has a slow base drift (roughly 6 to 16 px per second, mostly up and to the side) and a small tilt, and wraps around the edges. With a mouse, a letter within about 120 px of the pointer is eased away from it (force grows as the pointer gets closer, motion is damped so it stays smooth) and tilts a little. Touch screens get drift only.
- **Implementation:** one client component `LetterField` mounted once in the root layout, a fixed layer behind the content (`pointer-events: none`), a single `requestAnimationFrame` loop that moves the elements with transforms only, time step clamped, and the loop stopped while the tab is hidden. A pure function steps the physics, so it can be checked without a browser.
- **Accessibility and safety:** `aria-hidden`, not focusable, never intercepts clicks, content contrast is unaffected (letters sit under the content and stay faint). Under `prefers-reduced-motion` the letters are drawn once at fixed positions and never move.

## 4. New address and cutover (owner actions)

Order matters; the old setup is kept until the new one is verified.

1. **Cloudflare:** add DNS (or the tunnel hostname) for `kuronuma.lunariankids.com` pointing at the same origin.
2. **Google Cloud:** add `https://kuronuma.lunariankids.com/api/auth/google/callback` to the OAuth client's authorised redirect URIs. Keep the old one for now.
3. **Deploy:** merge PR R, set `PUBLIC_URL=https://kuronuma.lunariankids.com` in `.env`, redeploy. Check `/api/healthz`, sign in, and the Connect page.
4. **Assistants:** re-add the connector in Claude and in Gemini with the new URL, and approve again.
5. **Old address:** add a Cloudflare redirect rule from `nemosyne.lunariankids.com` to the new host for web pages (the redirect lives in Cloudflare so the public repository never contains the domain). Remove the old Google redirect URI and DNS once nothing uses them.

Rollback: redeploy the previous image with the old `PUBLIC_URL`; nothing in step 1 or 2 has to be undone.

What users notice: they sign in again (new cookie names and address), and connected assistants must be re-added.

## 5. Delivery

Each is its own PR, merged one at a time, each with an agent log.

| PR | Content | Needs owner approval |
|---|---|---|
| R | Rename (section 1) | Yes: `deploy/`, Dockerfiles, CI, `.env.example`, sign-in code (cookie names, consent text) |
| V1 | Blush look, motifs, motion, consent page restyle, new og image | Yes: consent page is sign-in code |
| V2 | Letter field | No |
| S3, S2 | Security fixes already agreed | As already approved |
| Dashboard | Plan and PRs from the dashboard spec, on the new brand | As in that spec |
| Public repo | New `kuronuma-mcp` repository from the finished code with a clean history | Yes: repository settings |

## 6. Testing

- **R:** the grep check; Go tests, `gofmt`, `go vet` and the Go 1.26 build in CI; web `tsc`, `check:design` and build; Playwright screenshots of `/`, `/th`, `/connect`, `/dashboard` and the consent page; a check that no page text says Nemo.
- **V1:** the contrast script; screenshots at desktop and phone width in English and Thai; a reduced-motion run showing no running animations or transitions; `check:design` updated for the new tokens.
- **V2:** a browser test that moves the mouse next to a letter and asserts that it moves away; that nothing moves under reduced motion; that a click on a button under a letter reaches the button; the physics function on its own.
- Real sign-in, the real API and real Postgres are exercised only by CI and after deploy, as before.

## 7. Risks and open points

- **Logo:** unchanged for now; a new mark is a later step.
- **FAQ naming text and all Thai text** are mine and need the owner's and a native speaker's read; they are listed in the PRs.
- **Cutover:** connections and sessions end; the sign-in redirect URI must be added in Google before deploying or sign-in fails.
- **Motion budget:** the letter loop is the only continuous script animation; it is small and pauses when hidden, but low-end phones will show it first.
- **Public repo:** created last, so the module path may need the one-step change above.
