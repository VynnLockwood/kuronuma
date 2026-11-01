# More motifs plan (PR M)

- Agent: claude
- Date: 2026-11-02
- Branch / PR: docs/more-motifs-plan

## Goal
Plan how to add 16 new original motifs to the site and place them on every page without making it messy.

## What I did
- Drew 16 motifs (SVG, in the existing style) and rendered a preview sheet; saved the sheet and its source under `docs/superpowers/plans/assets/`.
- Wrote `docs/superpowers/plans/2026-11-02-more-motifs.md`: a placement grammar (three sizes, one spot per section, in flow never over text, a viewport budget of five, calm where data lives), a page-by-page map, five tasks and a review focus.
- Pointed this checkout at the renamed repository (`VynnLockwood/kuronuma`); the old name redirects.

## Verified
- Docs and images only. The pages and sections in the map exist in the code (read from `Landing.tsx`, `connect/page.tsx`, the dashboard components, and the Go consent file).

## Not verified / known gaps
- Nothing is implemented. Placement quality is judged by the tests described in the plan and by eye in the implementing PR.

## Decisions and assumptions
- A character picture of Kuronuma Sawako was asked for; I did not do it (copyrighted artwork in a public MIT repository). The plan keeps to original drawings.
- The module path rename is a separate PR.

## Next steps
- Owner reviews and merges; then implement as `feat/more-motifs`.
