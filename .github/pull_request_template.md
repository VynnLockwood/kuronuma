## Summary
<!-- What changed and why, in one or two sentences. -->

## Type
- [ ] feat
- [ ] fix
- [ ] docs
- [ ] chore / ci
- [ ] refactor
- [ ] test

## Changes
-

## How I verified
<!-- Paste commands and their results. -->
- [ ] API changed: `go vet ./... && go test -race -count=1 ./...`
- [ ] Web changed: `npx tsc --noEmit && npm run build`
- [ ] Ran it locally or in Docker

## Not tested / risks
<!-- Be honest about what you could not run. -->

## Needs owner approval (see AGENTS.md)
- [ ] Auth / OAuth
- [ ] Database migration
- [ ] Deploy, CI, or nginx
- [ ] Dependencies
- [ ] None of the above

## Checklist
- [ ] PR title follows Conventional Commits, e.g. `feat(plutus): add spending summary`
- [ ] No secrets in the diff
- [ ] Agent log added in `agent-logs/<agent-name>/` (if an agent did the work)
- [ ] Docs updated if behavior changed
