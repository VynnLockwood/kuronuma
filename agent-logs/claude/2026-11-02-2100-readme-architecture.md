# README rewrite and architecture diagram (PR R)

- Agent: claude
- Date: 2026-11-02
- Branch / PR: docs/readme-architecture

## Goal
Rewrite the README in plain words without the Greek god names, with the current progress and an architecture picture, as in the merged plan `docs/superpowers/plans/2026-11-02-skills-readme-nav.md` (PR R).

## What I did
- Read the code for every claim: tool names, dashboard pages, routes, nginx locations, compose services, `go.sum`.
- Drew the architecture diagram with the vendored `diagram-design` skill and the Kuronuma profile (Architecture type: 9 nodes, 9 arrows, 2 focal nodes). Files in `docs/assets/`: `architecture.html` (source), `.svg` (system fonts), `.png` (rendered with Inter and Geist Mono embedded in a temporary copy).
- Rewrote `README.md`: what it is, the picture, Status (Working now, In progress, Planned), getting started (the outdated go.sum step removed), Google setup, connect an assistant, develop, CI, how it is built, agents and diagrams, security, licence.

## Verified
- `readme-check.py` (scratch, outside the repo): PASS. It fails on any god name (case-insensitive), a broken relative link, an image without real alt text, a PNG under 1600 px or over 400 KB, a missing status heading, stale text, an unrecognised bash command, a TLS label on the proxy, a missing sign-in-to-database edge, an overstated CI claim, or a Planned item not in `docs/design.md`. Each assertion was seen failing first.
- `python3 -I skills/diagram-design/scripts/self_check.py --offline docs/assets/architecture.html`: OK.
- Looked at the PNG at full size and at 600 px wide; no clipped or overlapping text.
- A fresh Opus review found one Critical (the diagram said nginx does TLS; the bundled nginx listens on port 80 only) and three Important findings; all fixed. Tools (15) and dashboard features were confirmed against the code by the reviewer.

## Not verified / known gaps
- I did not start the Compose stack; the README commands are checked against files, not run. Whether `go test ./...` needs a live Postgres was not checked.
- The SVG was checked only on a machine with Inter and Geist Mono installed; the fallback is untested.
- The Claude connector menu names and `claude mcp add` syntax are copied from the old README.

## Decisions and assumptions
- The README points to `docs/design.md` for the code names instead of naming parts.
- PNG and SVG use slightly different typefaces on purpose (PNG matches the site; SVG makes no font request).
- `docs/design.md` and the code keep their names.

## Next steps
- PR W (repo link in the nav). Owner may want to fix the compose and design-doc wording that the database is on an "internal network only".
