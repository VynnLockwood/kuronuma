# Skills

This is the one place agent skills live in this repository. A skill is a folder with a `SKILL.md` that tells an AI agent how to do one kind of job well.

| Skill | What for | Source |
|---|---|---|
| `diagram-design` | Architecture and other diagrams as HTML, SVG or PNG | third party, see [THIRD_PARTY.md](THIRD_PARTY.md) |

`.claude/skills` is a symlink to this folder, so Claude Code finds the skills here. Other tools can be pointed at `skills/` as well.

## Using diagram-design with the Kuronuma look

The skill reads its colours and fonts from a profile. Ours is `profiles/kuronuma.md` (the Blush palette, Inter and Geist Mono, no Google Fonts), and the file `.diagram-design` at the repository root selects it. The skill looks for profiles in your home folder, so copy ours there once:

```bash
mkdir -p ~/.diagram-design/profiles
cp skills/profiles/kuronuma.md ~/.diagram-design/profiles/kuronuma.md
```

Without the copy the skill says the profile is missing and asks which one to use; it never guesses.

PNG export needs Playwright (`pip install playwright && playwright install chromium`). SVG and HTML export need nothing.

## Do not let the skill edit itself

`.claude/skills` points here, so the vendored `skills/diagram-design/` is the skill's "installed copy". The skill's own onboarding, `save` and markerless `load` commands write into that copy. **Never run the skill's onboarding, `save` or `load` in this repository.** If it says the profile is missing, copy ours (above) instead of choosing one of its setup options. After any diagram session, `git status` must show nothing under `skills/diagram-design/`. If it does, discard those changes.

## Adding or updating a skill

1. Read the skill first: its scripts and its instructions. A skill is third-party text an agent will follow with full permissions, so treat it like a dependency.
2. Get the owner's approval (see [AGENTS.md](../AGENTS.md)).
3. Put it in `skills/<name>/` unmodified, record the source, commit, licence, bundled third-party content and what you checked in [THIRD_PARTY.md](THIRD_PARTY.md), and keep `skills-lock.json` current.
4. Updating: do not run the installer inside this repository. Run it in a scratch repo (`mkdir /tmp/scratch && cd /tmp/scratch && git init && DISABLE_TELEMETRY=1 npx skills@1.7.4 add <source> -y --copy -a claude-code`), copy the result over `skills/<name>/`, check `diff -r` against the upstream commit, review the diff, then update the commit in `THIRD_PARTY.md` and `skills-lock.json`. The installer replaces folders it manages, and through the `.claude/skills` link that would be the tracked copy.
5. Restoring from the lock file: do the same in a scratch repo with `npx skills@1.7.4 experimental_install` and copy the restored skill into `skills/`.
6. The installer sends an install telemetry event unless `DISABLE_TELEMETRY` or `DO_NOT_TRACK` is set.
