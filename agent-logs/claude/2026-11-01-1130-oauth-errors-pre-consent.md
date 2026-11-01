# OAuth errors on our own page before sign-in (S3)

- Agent: claude
- Date: 2026-11-01
- Branch / PR: fix/oauth-errors-pre-consent

## Goal
Stop the authorize endpoint from redirecting to a client's `redirect_uri` for errors found before the user has signed in and approved.

## What I did
- `parseAuthz` errors no longer carry a redirect flag; `fail` always renders our error page. The deny path after a CSRF-checked POST still redirects.
- Tests: every pre-consent error, signed in and signed out, must be a 400 HTML page with no `Location`; an escaping test; the deny test still passes.
- `docs/design.md` authorize bullet updated.

## Verified
- Before the fix the new tests failed (303 redirects to the client's URI, including when signed out). After: `gotest.sh` ok across cerberus, hermes, themis, atlas, plutus; `gotest.sh vet` ok; `gofmt -l apps/api` empty.
- A fresh whole-branch review found no Critical or Important issues and confirmed every remaining redirect needs a verified, CSRF-checked POST.

## Not verified / known gaps
- The cerberus integration test and the `google` subpackage need dependencies the harness lacks; CI runs them. Nothing in them referenced the removed field (grep).
- Not tried against live Claude, Gemini or ChatGPT.

## Decisions and assumptions
- Deferred minors: loopback CLI clients now wait on their callback instead of failing fast, and the page does not say to go back to the assistant; unused `*http.Request` in `fail`; a stronger escaping test; a POST-path test; header assertions on the error path; a doc sentence on same-origin POSTs.

## Next steps
- S2 PR (caps, unused client cleanup, quoted names).
