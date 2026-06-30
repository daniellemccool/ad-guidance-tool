# adr-plugin — ADR skills for Claude Code

This is a [Claude Code](https://code.claude.com) plugin that ships *with* `adg` so its
guidance tracks the CLI in lockstep — the references in `skills/*/references/` are updated
in the same change that updates the CLI, which is why this repo is the plugin's canonical
home. The `d3i-skills` marketplace **lists** this plugin via a `git-subdir` source pinned to a
release tag — a reference to this repo, not a copy — so there is one source of truth and nothing
to sync.

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
`tools/adr-plugin` (this is how the `d3i-skills` marketplace lists it, pinned to a release tag).

The skills call the `adg` CLI, and it **rides along**: the plugin ships a `bin/adg` wrapper that
Claude Code puts on `PATH` while the plugin is enabled, and on first use it downloads the prebuilt
`adg` matching the plugin's version — no Go toolchain, no manual install.

A **system `adg` on `PATH`** is still needed for any `adg` invocation that runs *outside* the skills'
execution context: the copied-out git hook, governance hooks a target repo wires into its own
settings, and — as of v1.2.0 — **this plugin's own bundled hooks** (`hooks/hooks.json`, below).
Install it once with the prebuilt binary:

```
curl -fsSL https://raw.githubusercontent.com/daniellemccool/ad-guidance-tool/main/install.sh | sh
```

## Bundled hook

As of **v1.2.0** the plugin ships `hooks/hooks.json` with one fail-open governance hook, so enabling
the plugin registers it automatically — a repo no longer needs the manual `.claude/settings.json`
snippet from `docs/lean-example/hook-setup.md` (it remains for non-plugin users):

- **PreToolUse** (matcher `Edit|Write|MultiEdit`) → `adg lean brief --hook --model docs/decisions`.
  Injects the governing lean ADRs as context *before* each edit, **deduped per session**: each ADR is
  injected at most once per Claude Code session (keyed by `session_id`), so repeated edits to
  broadly-scoped files don't re-pay for the same brief. A forbids violation always re-surfaces.

It is **fail-open**: it needs system `adg` on `PATH` and a `docs/decisions` lean model in the repo;
absent either, it emits nothing and the edit proceeds — so "no brief appeared" never means "no rule
applies."

The hook is the *edit-time safety net*; it is deliberately not the whole story. The brief earns most
of its value at **planning time** — pull it yourself with `adg lean brief` before designing a change
(see the "Golden path" convention in `docs/lean-example/hook-setup.md`, meant for a consuming repo's
`CLAUDE.md`). **Enforcement** belongs at **commit time**, not on every turn: install the lean
`pre-commit` gate (`skills/write-lean-adr/assets/githooks/pre-commit`) to run `adg lean index` +
`check` once per commit. The authoring-side review phase is *not* a hook either — it is driven by the
`write-lean-adr` skill, which has a subagent judge `adg lean review` output against the rubric.
