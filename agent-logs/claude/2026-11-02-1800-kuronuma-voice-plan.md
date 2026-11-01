# Kuronuma voice plan (PR N)

- Agent: claude
- Date: 2026-11-02
- Branch / PR: docs/kuronuma-voice-plan

## Goal
Plan how connected assistants get a Kuronuma personality when they reply to the user, and check whether the repository is ready to be made public.

## What I did
- Wrote `docs/superpowers/plans/2026-11-02-kuronuma-voice.md`: voice guide appended to the MCP instructions, gentle fixed messages in Plutus and Atlas, rules (style never authority, short, language-neutral, facts exact), tests, review focus.
- Audited the repository for going public (history, secrets, personal data, settings). Findings are in the PR description.

## Verified
- Docs only. Read `hermes/instructions.go` and the fixed message strings in `plutus.go` and `atlas/tools.go`; the plan matches them.
- Audit: no secret patterns in tracked files or history, no `.env` ever added, no personal email in any author, committer or message across all refs.

## Not verified / known gaps
- Branch protection and the private-vulnerability-reporting setting need the owner's GitHub access (the API token gets 403).
- Nothing implemented yet.

## Decisions and assumptions
- Read "personality" as an original temperament (shy, sincere, careful) rather than copied dialogue or a claim to be an existing character.
- No on/off environment variable, because `.env.example` needs owner approval.

## Next steps
- Owner merges; implement as `feat/kuronuma-voice`.
