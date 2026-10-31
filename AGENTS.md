# AGENTS.md

Instructions for every AI agent (and human) working in this repo. Read all of it before changing anything. `CLAUDE.md` points here too.

## What this project is

Kuronuma is an open-source personal data hub exposed to AI assistants over MCP. Stack: Go API (`apps/api`), Next.js web (`apps/web`), Postgres, nginx, Docker Compose. Components are named after Greek gods: Cerberus (auth and OAuth server), Hermes (MCP endpoint), Plutus (spending), Olympus (web), Gaia (Postgres), Propylaea (nginx), Hades (field encryption, planned). See `README.md` and `docs/design.md`.

**The project is live: `main` is deployed.** Treat every change as production-affecting.

## Golden rules

1. Never push to `main`. Work on a branch, open a PR, wait for green CI, and let the owner merge.
2. Never force-push, rewrite published history, or delete branches you did not create.
3. Never commit secrets (`.env`, keys, tokens, client secrets) or paste them into logs, PRs, or agent logs.
4. Do not claim something works unless you ran it. Put the command and its result in the PR and your agent log. If you could not run something, say so.
5. One concern per PR. Keep diffs small. No drive-by refactors.
6. If a rule here blocks the task, stop and ask the owner. Do not work around it.

## Needs the owner's explicit approval first

- Anything in `apps/api/internal/core/cerberus` (auth, sessions, OAuth) or that changes scopes or permissions
- New or changed database migrations
- `deploy/`, Dockerfiles, nginx, CI workflows, `.env.example`
- Adding or upgrading dependencies, including major bumps that Dependabot proposes (Go, Node, Actions)
- License, security policy, repository settings

## Database migrations

- Add a new numbered file in `apps/api/internal/store/migrations/`. Never edit or renumber a migration that is already on `main`.
- Migrations run automatically when the API starts, so they must be safe on a live database: additive first (new tables, nullable columns). Never drop or rename something in the same release as the code that stops using it.

## Security rules (this app holds personal financial data)

- Every query on user data filters by the authenticated user's id, taken from the verified token. Never accept a user id from tool arguments or request bodies.
- Any MCP or API output containing user-supplied text (merchant, notes, client names) must be sanitized with `hermes.Untrusted` and treated as data, never as instructions.
- Never log tokens, auth codes, cookies, OAuth query strings, or transaction contents.
- Each MCP tool declares the narrowest scope it needs. Destructive tools are marked destructive.
- Add tests for any security-relevant behavior you add or change.

## Verify what you touched before every commit

| Area | Run |
|---|---|
| API (`apps/api`) | `gofmt -l .` (no output), `go vet ./...`, `go test -race -count=1 ./...` |
| Web (`apps/web`) | `npx tsc --noEmit`, `npm run build` |
| Deploy files | `docker compose -f deploy/docker-compose.yml config -q` (with dummy env), nginx syntax check |
| Docs only | check that linked paths exist |

If `go.mod` changes, run `go mod tidy` and commit `go.sum`. CI enforces this.

## Git workflow

- Branch from the latest `main`: `feat/...`, `fix/...`, `chore/...`, `docs/...`.
- Commit messages and PR titles use Conventional Commits: `type(scope): summary`, where type is one of `feat fix docs chore ci refactor test perf build revert`. Enable the commit template once with `git config commit.template .gitmessage`. A CI check enforces PR titles.
- Commits use the repo owner's configured git identity. Do not add AI `Co-Authored-By` trailers. Attribution lives in agent logs (below).
- Fill in the PR template honestly, including what you did not test.

## Agent logs (required)

Every agent that works in this repo keeps a log in its own folder, so agents never collide and the owner can see who did what and why.

- Folder: `agent-logs/<your-agent-name>/`, lowercase kebab-case, for example `claude` or `hermes`. Create it if it does not exist. Use the name you are known by. If unsure, ask the owner once and keep that name.
- One file per work session: `YYYY-MM-DD-HHMM-short-topic.md`.
- Write only inside your own folder. Never edit or delete another agent's logs.
- Commit your log in the same PR as the work. Keep it short and factual. Template: `agent-logs/README.md`.
- Never put secrets or personal data in logs.

## Definition of done

Code and tests are in, local verification passed (with evidence), docs are updated if behavior changed (`README.md`, `docs/design.md`), the agent log is committed, and the PR is open with green CI.
