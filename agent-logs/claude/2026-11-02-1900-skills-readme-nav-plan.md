# Skills, README and nav plan (PRs S, R, W)

- Agent: claude
- Date: 2026-11-02
- Branch / PR: docs/skills-readme-nav-plan

## Goal
Plan three changes the owner asked for: add the `diagram-design` skill under one `skills/` folder, rewrite the README without the Greek names (current progress, architecture picture), and link to the public repository from the top navigation.

## What I did
- Ran `npx skills add cathrynlavery/diagram-design` in an empty scratch repo (telemetry off) to see what it installs, and read the skill: stdlib-only scripts, no network or subprocess calls, MIT, 5.8 MB.
- Read `lib/brand.ts`, `SiteHeader`, the README and `docs/design.md` to ground the plan; confirmed the repo is public.
- Wrote `docs/superpowers/plans/2026-11-02-skills-readme-nav.md`.

## Verified
- Docs only. The install result, hash and script scan are from the scratch run; nothing was added to the repository yet.

## Not verified / known gaps
- The skill's behaviour with a saved profile is untested until PR S. The GitHub icon and header layout are untested until PR W.

## Decisions and assumptions
- One `skills/` folder with `.claude/skills` as a symlink, rather than the CLI's default location.
- Vendored unmodified; Kuronuma styling through the skill's profile feature.
- `REPO_URL` defaults to the public repo (no Docker or deploy change).
- README loses the names, not the code or `docs/design.md`.

## Next steps
- Owner merges the plan; then PR S, R, W in order.
