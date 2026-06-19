# adr-plugin — the `write-adr` Claude Code plugin

This is a [Claude Code](https://code.claude.com) plugin that ships *with* `adg` so its
guidance tracks the CLI in lockstep — the reference in `skills/write-adr/references/` is
updated in the same change that updates the CLI, which is the whole reason it lives here
rather than in a separate skills repo.

## Layout

```
tools/adr-plugin/
  .claude-plugin/plugin.json     plugin manifest
  skills/write-adr/              the skill (SKILL.md, references, assets, evals)
```

## Install

Either add this repo as a marketplace directly:

```
/plugin marketplace add daniellemccool/ad-guidance-tool
```

…or install via a marketplace that references it with a `git-subdir` source pointing at
`tools/adr-plugin`.

`adg` itself must be on `PATH` for the skill's commands to run.
