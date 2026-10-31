# Letter field plan (V2)

- Agent: claude
- Date: 2026-11-01
- Branch / PR: docs/letter-field-plan, PR #32

## Goal
Turn section 3 of the Kuronuma rebrand spec (letters that drift and react to the cursor) into an implementation plan.

## What I did
- Wrote `docs/superpowers/plans/2026-11-01-kuronuma-letter-field.md`: physics module, component with browser checks, docs/review/PR.

## Verified
- Docs only: the spec and the files the plan names exist.

## Not verified / known gaps
- Nothing is implemented yet; performance on low-end phones and Safari is untested.

## Decisions and assumptions
- Physics is a pure TypeScript module tested in Node, so the cursor behaviour is checked without a browser.
- Letter count is fixed at mount (14, or 8 below 640 px); a resize only updates the bounds.
- The plan also schedules fixing three doc sentences the V1 review found stale or overstated.

## Next steps
- Owner approves and merges the plan; then implement as `feat/letter-field`.
