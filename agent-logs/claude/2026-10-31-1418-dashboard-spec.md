# Web dashboard design (spec only)

- Agent: claude
- Date: 2026-10-31
- Branch / PR: docs/web-dashboard-spec (PR link on GitHub)

## Goal
Owner asked for a web dashboard that shows all of a signed-in user's stored data and lets them manage it, plus a security check before the repository goes public. This PR is the written design for the dashboard only; no code.

## What I did
- Explored the sign-in session handling, the store methods, the tool layer and the web app to see what a browser can reuse.
- Asked the owner three questions and recorded the answers in the spec: scope (view and manage everything, spending add and delete only), language (English and Thai), approach (a bridge to the existing tool handlers).
- Wrote `docs/superpowers/specs/2026-10-31-web-dashboard-design.md`: the bridge endpoint and its CSRF rules, the five tabs, a Connected assistants tab, testing, and five PRs (G to K) with their approvals.
- Separately, an independent reviewer did a read-only security review and I checked the repository settings I can see. The findings went to the owner privately; they are not written down here or in the spec, and fixes will be separate PRs.

## Verified
- Docs only: the spec's referenced paths exist (`docs/design.md`); no placeholder text (`grep` for TBD, TODO, FIXME is empty).

## Not verified / known gaps
- Nothing was built or run. The spec's behavior claims (for example that `oauth_tokens` has the `user_id` and `client_id` needed to revoke a connection) were read from the migrations, not tested.

## Next steps
- Owner reviews the spec; then the implementation plan for PR G (bridge) is written. The security fixes agreed with the owner come first.
