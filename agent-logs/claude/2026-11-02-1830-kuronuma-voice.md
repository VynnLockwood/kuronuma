# Kuronuma voice for connected assistants

- Agent: claude
- Date: 2026-11-02
- Branch / PR: feat/kuronuma-voice

## Goal
Make connected assistants sound like Kuronuma (quiet, sincere, careful) when they talk about the user's memory, following the merged plan `docs/superpowers/plans/2026-11-02-kuronuma-voice.md`.

## What I did
- `hermes.VoiceGuide` appended to `DefaultInstructions` (after the safety rules); scoped to talk about what Kuronuma keeps, normal voice elsewhere.
- Fixed Plutus and Atlas sentences moved into named `msg*` constants with a gentle opener; facts and instructions to the assistant kept.
- Tests: voice present, after the rules, rules intact, short and plain, names nobody, not "become Kuronuma"; message table tests.
- `docs/design.md` Voice paragraph; one README line.

## Verified
- `gotest.sh` (cerberus, hermes, themis, webguard, atlas, plutus): pass. `go vet`: clean. `gofmt -l apps/api`: no output.
- Each new test was seen failing first; the review fix (voice scope) was RED then GREEN.
- Fresh Opus review: no Critical; one Important (voice not limited to this topic), fixed.

## Not verified / known gaps
- How a real assistant renders the voice (Claude, Claude Code, others; English and Thai) is untested. Wording is a first draft.
- `store` and `cmd` packages are compiled only by CI; `cmd/kuronuma-api/main.go` still passes `hermes.DefaultInstructions`.

## Decisions and assumptions
- Raised the old 1,200-character test cap to 1,800, as the plan says.
- Only fixed sentences and limit errors were changed; validation errors stay exact.
- No on/off setting (would need `.env.example`).

## Next steps
- Try it in a real client and tune the wording; deferred minors listed in the PR.
