# adr-plugin — ADR skills for Claude Code

This is a [Claude Code](https://code.claude.com) plugin that ships *with* `adg` so its
guidance tracks the CLI in lockstep — the references in `skills/*/references/` are updated
in the same change that updates the CLI, which is the whole reason they live here rather
than in a separate skills repo.

It ships two format-specific skills — pick the one matching a repo's ADRs:

- **write-madr-adr** — durable MADR records (Context / Considered Options / Decision
  Outcome) with the `decide` / `supersede` / `revise` lifecycle.
- **write-lean-adr** — compact lean Decision/Guidance records with routing frontmatter
  (`applies_to` / `excludes` / `forbids` / `companions`), plus the agent-consumption rules
  (`adg lean brief`, the PreToolUse hook).

## Layout

```
tools/adr-plugin/
  .claude-plugin/plugin.json     plugin manifest
  skills/write-madr-adr/         MADR skill (SKILL.md, references, assets, evals)
  skills/write-lean-adr/         lean skill (SKILL.md, references)
```

## Install

Either add this repo as a marketplace directly:

```
/plugin marketplace add daniellemccool/ad-guidance-tool
```

…or install via a marketplace that references it with a `git-subdir` source pointing at
`tools/adr-plugin`.

`adg` itself must be on `PATH` for the skill's commands to run.
