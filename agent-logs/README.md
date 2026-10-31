# Agent logs

Each agent that works on Kuronuma writes its own session logs here, in its own folder:

```
agent-logs/
  claude/
    2026-10-30-1530-hermes-mcp.md
  hermes/
    2026-10-31-0900-fix-ci.md
```

Rules (full version in `/AGENTS.md`):

- Folder name is your agent name in lowercase kebab-case. Create it if missing.
- One file per work session, named `YYYY-MM-DD-HHMM-short-topic.md`.
- Write only in your own folder. Never edit or delete another agent's logs.
- Commit the log in the same PR as the work.
- No secrets, tokens, or personal data.

## Template

```markdown
# <short title>

- Agent: <name>
- Date: YYYY-MM-DD
- Branch / PR: <branch>, <PR link if any>

## Goal
What the owner asked for, in one or two sentences.

## What I did
- Bullet list of changes (files or areas, and why).

## Verified
- `<command>`: <result>

## Not verified / known gaps
- Anything I could not run or test, and why.

## Decisions and assumptions
- Choices I made that the owner may want to revisit.

## Next steps
- Suggested follow-ups.
```
