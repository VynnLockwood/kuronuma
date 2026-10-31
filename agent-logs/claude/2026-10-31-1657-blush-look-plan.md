# Blush look and motion plan (plan only)

- Agent: claude
- Date: 2026-10-31
- Branch / PR: docs/kuronuma-v1-plan (PR link on GitHub)

## Goal
The Kuronuma rename (PR #29) was merged. Next in the spec is PR V1: the Blush palette, original SVG motifs and gentle motion. This PR is the implementation plan for V1 only; no code.

## What I did
- Read the redesigned web app (theme tokens, shadcn-style components, landing structure, the design check script) and the Go consent page styles so the plan names real files and values.
- Checked every Blush colour pair with a contrast script before writing the values into the plan (all pass WCAG AA; the muted-on-surface and subtle-on-surface pairs are close at 5.3 and 3.0, so the plan makes the contrast check part of `check-design.mjs`).
- Wrote `docs/superpowers/plans/2026-10-31-kuronuma-blush-look.md`: five tasks (tokens and the contrast guard; motifs and where they sit; motion; Go consent and error pages; share image, docs, review and PR) with exact palette and motif values, interfaces, tests and review focus.

## Ruling made while planning
- The spec says links' underline "slides in". A slide needs a gradient or a background-size trick, and `check:design` forbids gradients, so the plan fades the underline colour in instead. Cost if wrong: a different effect on links; the owner can ask for a slide and relax the rule.

## Verified
- Docs only: no placeholder text; each item of spec section 2 maps to a task.

## Not verified / known gaps
- Nothing was built or run. The look is only approved as a sketch; the real result is judged from screenshots in the V1 PR.

## Next steps
- Owner reviews the plan and picks an execution method; then V1 is built. V2 (the cursor-reactive letters) follows.
