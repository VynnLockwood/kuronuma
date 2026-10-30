# Fix: Claude connector could not register with Nemosyne

- Agent: claude
- Date: 2026-10-30
- Branch / PR: `fix/oauth-registration-lenient` (no PR opened yet)

## Goal
The owner connected Nemosyne from Claude and got "Couldn't register with nemosyne's sign-in service" (dynamic client registration).

## What I found
- PR #8 is merged and CI on `main` passed. I cannot reach the deployed server, so I could not see the actual request or response.
- A likely code cause: registration rejected clients that asked for a token auth method other than `none`, or for extra grant/response types. RFC 7591 lets the server substitute supported values instead of failing.
- A likely deployment cause: `deploy/nginx/nginx.conf` is a bind mount, so `docker compose up --build -d` does not reload nginx. With the old config, `/.well-known/*` goes to the web app (404) and clients then fall back to `/register` on the web app.

## What I did
- `oauth.go`: registration now accepts and ignores other auth methods, grant types and response types, and always answers as a public client (`token_endpoint_auth_method: none`). Redirect URIs are still strictly validated.
- Added logs for registration (success and rejection, with reason and redirect hosts, no secrets).
- Tests: former strict cases now expect 201; new test covers realistic client payloads.
- README: update procedure (restart nginx) and a connection troubleshooting section with curl checks.

## Verified
- `gofmt -l .`: no output. `go vet` and `go test -race -count=1` for config, hermes, cerberus, plutus: pass.
- Mutation check: re-adding the strict auth-method check made the new test fail; code restored.
- Whole API builds against stub dependencies.

## Not verified / known gaps
- I do not know the real cause on the deployed server. The leniency fix is a hardening that may not be the cause; the curl checks and API logs will show which layer fails.
- Not tested against a real Claude connector.

## Next steps
- Owner: run the troubleshooting checks, redeploy with an nginx restart, retry, and share the failing output if it persists.
