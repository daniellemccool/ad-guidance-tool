# adr-plugin — ADR skills for Claude Code

This is a [Claude Code](https://code.claude.com) plugin that ships *with* `adg` so its
guidance tracks the CLI in lockstep — the references in `skills/*/references/` are updated
in the same change that updates the CLI, which is the whole reason they live here rather
than in a separate skills repo.

It ships three skills — two for *authoring* (pick the one matching a repo's ADR format),
and one for *obeying* lean briefs while changing code:

- **write-madr-adr** — author durable MADR records (Context / Considered Options / Decision
  Outcome) with the `decide` / `supersede` / `revise` lifecycle.
- **write-lean-adr** — author/migrate/rewrite/review compact lean Decision/Guidance records
  with routing frontmatter (`applies_to` / `excludes` / `forbids` / `companions`).
- **follow-adr-governance** — a behavior primer for obeying an injected lean ADR brief while
  editing code. The real work is done by the brief and the PreToolUse hook; this skill is the
  fallback/primer (treat the brief as authoritative, invariants are hard constraints, stop on
  forbidden scope, run the footer's checks/tests).

The split is deliberate: authoring is rare and deliberate (rich skills with a rubric);
consuming is frequent and carried by the hook + the self-contained brief (a tiny skill).

## Layout

```
tools/adr-plugin/
  .claude-plugin/plugin.json     plugin manifest
  skills/write-madr-adr/         MADR authoring skill (SKILL.md, references, assets, evals)
  skills/write-lean-adr/         lean authoring skill (SKILL.md, references, evals)
  skills/follow-adr-governance/  obey-the-brief behavior primer (SKILL.md, evals)
```

## Install

Either add this repo as a marketplace directly:

```
/plugin marketplace add daniellemccool/ad-guidance-tool
```

…or install via a marketplace that references it with a `git-subdir` source pointing at
`tools/adr-plugin`.

`adg` itself must be on `PATH` for the skill's commands to run.
