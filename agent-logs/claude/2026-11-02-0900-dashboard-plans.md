# Dashboard plans (G to K)

- Agent: claude
- Date: 2026-11-02
- Branch / PR: docs/dashboard-plans

## Goal
Turn the merged dashboard spec into implementation plans for its five PRs.

## What I did
- Wrote one plan per PR in `docs/superpowers/plans/`: G (the bridge, with a new `webguard` package), H (shell, Overview, Spending), I (Tasks and notes), J (Categories and Payee rules), K (Connected assistants). Separate files so each PR has its own ledger and its own final review.
- Checked against the code: tools reject unknown arguments (so the bridge cannot be used to choose a user), `list_categories` already returns usage counts and an `unlisted` marker, the merge limit is 10, `due_on: ""` clears a date.

## Verified
- Docs only: files and code paths the plans name exist.

## Not verified / known gaps
- Nothing is implemented. Real tool result shapes are captured as fixtures during PR H, not assumed here.

## Decisions and assumptions
- The bridge header is `X-Kuronuma-Web` (the spec said `X-Nemosyne-Web`; the product is now Kuronuma and the header is new).
- The Sec-Fetch-Site rule also refuses `same-site`, stricter than the spec.
- The browser-request checks live in a leaf package `webguard`, so PR K can reuse them without editing `hermes`.
- Dates use the Gregorian year in Thai too, pending a native speaker.

## Next steps
- Owner reviews and merges; then implement G first.
