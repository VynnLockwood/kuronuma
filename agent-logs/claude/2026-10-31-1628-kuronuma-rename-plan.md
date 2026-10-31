# Kuronuma rename plan (plan only)

- Agent: claude
- Date: 2026-10-31
- Branch / PR: docs/kuronuma-rename-plan (PR link on GitHub)

## Goal
The Kuronuma design spec was merged. The owner said to go ahead with the design and that the address, Cloudflare and Google sign-in steps will be done later by another agent. This PR is the implementation plan for the first PR, the rename (PR R). No code.

## What I did
- Listed every place the old names appear outside the historical folders (about 55 files) and the exact English and Thai copy lines, so the plan can fix the wording instead of leaving it to guesswork.
- Wrote `docs/superpowers/plans/2026-10-31-kuronuma-rename.md`: six tasks (a brand check as the failing test, Go module and build files, assistant-facing text and identifiers, the web app, deploy files and docs, log and PR) with the exact new copy, the names that stay, the tests and the verification commands.
- Left out of the plan on purpose: the new address, Cloudflare, the Google redirect URI and `PUBLIC_URL` (the owner's, done later by another agent).

## Verified
- Docs only: no placeholder text; the plan's references match the spec.

## Not verified / known gaps
- Nothing was built or run. The English FAQ wording about the name and all Thai text are mine and need the owner's and a native speaker's read.

## Next steps
- Owner reviews the plan and picks an execution method (native or subagent-driven); then PR R is built.
