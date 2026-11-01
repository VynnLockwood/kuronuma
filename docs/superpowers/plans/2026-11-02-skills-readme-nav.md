# Skills folder, a plain README with a diagram, and a repo link in the nav (PRs S, R, W) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline) or superpowers:subagent-driven-development. Steps use checkbox (`- [ ]`) syntax. Three PRs, one concern each, in this order: S (skills), R (README and diagram), W (nav link). Each branches from the latest `main`; each gets its own review and PR.

**Goal:** (S) add the `diagram-design` skill to the project under one main `skills/` folder; (R) rewrite the README without the Greek god names, with the current progress and an architecture picture made with that skill; (W) put a link to the public repository in the top navigation of every page.

## Findings that shape the plan

- `npx skills add cathrynlavery/diagram-design` was run in an empty scratch repo with telemetry off. The CLI installs to `.claude/skills/diagram-design` and writes `skills-lock.json` (source, path, SHA-256 `60a5683c...`). It copies 5.8 MB: `SKILL.md`, 40-odd `references/*.md`, `scripts/` and 220 example HTML files. Upstream is MIT, version 2.6, commit `0cf73c3`.
- Reviewed before adopting: the scripts import only the Python standard library (plus Playwright for `export_png.py`); no network calls, no subprocesses, no `eval`/`exec`; the markdown contains no install or shell commands. The example HTML files load Google Fonts (`fonts.googleapis.com`) when opened; they are examples, not served by our site, but this is noted in `skills/THIRD_PARTY.md`.
- The skill's first-run gate asks to customise its style guide. We use a saved profile with the Blush tokens (paper `#fffafb`, ink `#2a2024`, accent `#c23b69`) so the vendored files stay identical to upstream and diagrams match the site.
- The repository is now public, so the site can link to it without a build-time variable.
- Skills are third-party instructions that agents follow. Adding one is a supply-chain decision, so it needs the owner's approval: merging this plan is that approval, for this skill at this commit only.

## PR S: the skills folder

**Files:** Create `skills/diagram-design/` (verbatim copy of the upstream `skills/diagram-design/`), `skills/README.md`, `skills/THIRD_PARTY.md`, `skills-lock.json` (as written by the CLI), `.claude/skills` (a symlink to `../skills`), a Kuronuma profile for the skill. Modify `AGENTS.md` (a short "Skills" section).

1. **One folder.** `skills/` at the repo root is the only place skills live. `.claude/skills` is a symlink to it so Claude Code finds them; other tools can point at `skills/` too.
2. **Pinned and attributable.** `THIRD_PARTY.md` lists name, source URL, commit, licence (MIT, copyright holder, full licence text) and the Google Fonts note. `skills-lock.json` keeps the hash so `npx skills experimental_install` can restore it.
3. **Unmodified upstream.** Nothing inside `skills/diagram-design/` is edited except the saved profile and project marker the skill itself creates for customisation; `skills/README.md` says how to update (`npx skills update`, re-review the diff, bump the commit in `THIRD_PARTY.md`).
4. **Rule for every future skill:** read its scripts and instructions first, record the commit, get the owner's approval. `AGENTS.md` says so in three lines.

- [ ] **Step 1: Write the failing check** `skills-check.sh` (outside the repo): `skills/diagram-design/SKILL.md` exists with `name: diagram-design`; `skills/THIRD_PARTY.md` names the commit and licence; the hash in `skills-lock.json` equals the hash of the vendored folder as the CLI computes it; `.claude/skills` resolves to `skills/`; no file under `skills/` is outside the expected tree; no script under `skills/` imports `socket`, `urllib`, `requests`, `subprocess` or calls `eval`/`exec`; `AGENTS.md` has a Skills section; the Blush profile exists and the skill's own `scripts/self_check.py` (if it has a profile check) passes.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement.** Install in a scratch repo, copy into `skills/`, write the docs, create the symlink, create the profile with the Blush tokens.
- [ ] **Step 4: Verify.** The check passes; Claude Code lists the skill; generate one throwaway diagram into `/tmp` to prove the profile is picked up with no first-run prompt; CI-equivalent checks (docs link check) pass.
- [ ] **Step 5: Commit** `feat(skills): add the diagram-design skill under skills/`. Review by a fresh Opus reviewer (Review Focus below), one fix pass, agent log, PR with CI result.

## PR R: README and architecture picture

**Files:** Modify `README.md`; create `docs/assets/architecture.png` and `docs/assets/architecture.svg`, plus the source HTML in `docs/assets/architecture.html` and the generation note.

1. **No god names in the README.** Describe by role: web dashboard, MCP endpoint, sign-in and OAuth server, database, reverse proxy, spending, tasks and notes, categories. Code identifiers (package names such as `internal/core/hermes`) stay in the code and in `docs/design.md`; the README links to the design doc for the mapping. A check fails if the README contains any of: Olympus, Propylaea, Hermes, Cerberus, Hades, Gaia, Plutus, Themis, Atlas, Horae, Hephaestus.
2. **Current progress, honestly.** A table of what works (Google sign-in, web dashboard with overview, spending and tasks and notes in English and Thai, MCP tools for spending, tasks, to-dos, notes, categories and payee rules, assistant sign-in over OAuth, self-hosting with Docker Compose, the Kuronuma voice), what is in progress (categories and payee rules and connected assistants in the dashboard), and what is planned (encrypted fields, routines, work data). Facts come from merged PRs and the code, not memory; every claim has a path or PR behind it in the agent log.
3. **Architecture picture.** Generated with the `diagram-design` skill using the Kuronuma profile: browser and assistant, reverse proxy, web app, API with its parts, database. Exported to PNG (README uses it; GitHub renders it with any fonts) and SVG, with the alt text written out. Labels use the plain role names.
4. **Everything else kept:** getting started, Google sign-in setup, connect an assistant, licence. Paths and commands are checked against the repo.

- [ ] **Step 1: Write the failing check** (`readme-check.py`): none of the forbidden names; every relative link and image path exists; the image has alt text; the diagram PNG is 1600 px wide or more and under 400 KB; the status table has the three headings; every command in a fenced `bash` block starts with a command that exists in the repo (`docker compose -f deploy/docker-compose.yml`, `cp .env.example .env`, ...).
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement** the diagram first (view it), then the README.
- [ ] **Step 4: Verify.** The check passes; render the README preview and look at it; the picture is readable at 400 px wide.
- [ ] **Step 5: Commit** `docs: rewrite the README in plain words with an architecture diagram`. Review, fix pass, agent log, PR.

## PR W: repo link in the nav

**Files:** Modify `apps/web/lib/brand.ts`, `components/SiteHeader.tsx` (and the icon set if needed), `components/landing/Landing.tsx` (footer already uses `REPO_URL`), README note on `NEXT_PUBLIC_REPO_URL`. No Docker or deploy file changes.

1. **Always present.** `REPO_URL` defaults to `https://github.com/VynnLockwood/kuronuma`; `NEXT_PUBLIC_REPO_URL` still overrides it (self-hosters can point it at their fork or set it empty to hide the link).
2. **In the header on every page** that uses `SiteHeader` (landing English and Thai, connect, dashboard in every state, 404): a GitHub icon button with the visible label "GitHub" from `sm` up and the icon alone below, right of the main nav and before the language switch and sign-in button. Opens in the same tab? No: a new tab with `rel="noopener noreferrer"`.
3. **Quiet.** It is an outline or ghost button in existing tokens (no new colours), touch target at least 40 px on mobile, visible focus ring, accessible name "Kuronuma on GitHub" (Thai: "Kuronuma บน GitHub"), `aria-hidden` icon. The header must not wrap or overflow at 360 px in English or Thai, and the signed-in header (avatar, tabs) still fits.
4. **No new decoration.** The motif budget and the placement checks are unaffected.

- [ ] **Step 1: Write the failing test** `nav-repo-check.py` (Playwright): the link exists on `/`, `/th`, `/connect`, `/dashboard` (signed out, signed in, unreachable), `/th/dashboard`, `/does-not-exist`; `href` is the repo URL; `target=_blank`, `rel` has `noopener` and `noreferrer`; accessible name correct in both languages; label visible at 1280 and 768, icon only at 390 and 360; target at least 40 by 40 px at 390; header height unchanged and no horizontal scroll at 1280, 1024, 768, 390, 360; reachable by keyboard Tab in a sensible order; no overlap with the language switch or the sign-in button. A second case builds with `NEXT_PUBLIC_REPO_URL=` empty and expects no link.
- [ ] **Step 2: Run it.** Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Verify.** `tsc --noEmit`, `check:design`, `next build`, every existing browser check PASS, the new one PASS; screenshots of the header at 1280 and 390 in both languages.
- [ ] **Step 5: Commit** `feat(web): link to the repository from the top navigation`. Review, fix pass, agent log, PR.

## Review Focus

- **S:** the vendored folder is byte-identical to upstream at the pinned commit (diff against a fresh clone); no script gains network or process access; the symlink is relative and works on a fresh clone; `THIRD_PARTY.md` is complete (licence text, commit, fonts note); the AGENTS.md section does not weaken any existing rule; repository size grew by about 6 MB and nothing else large was added.
- **R:** no god name anywhere in the README; every claim in the status table is true today (spot-check five against the code); the diagram labels match the README words and the real request flow (browser to proxy to web or API; assistant to MCP endpoint; database internal only); links and images resolve; alt text is meaningful.
- **W:** the link is on every header state listed; accessible names in both languages; no wrap or overflow at 360 px; no layout shift on first paint; the env override and the empty case both behave; no hex colour added; no deploy file touched.

## Not included

- Customising the skill itself or writing new skills (a later PR, if wanted).
- Removing the god names from `docs/design.md`, code identifiers or the database (they stay; the README links to the mapping).
- A star count or any third-party widget on the site (no remote requests).
- The Go module path rename and dashboard PRs J and K (separate plans, already queued).
