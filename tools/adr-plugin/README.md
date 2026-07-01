# adr-plugin — ADR skills for Claude Code

This is a [Claude Code](https://code.claude.com) plugin that ships *with* `adg` so its
guidance tracks the CLI in lockstep — the references in `skills/*/references/` are updated
in the same change that updates the CLI, which is why this repo is the plugin's canonical
home. The `d3i-skills` marketplace **lists** this plugin via a `git-subdir` source pinned to a
release tag — a reference to this repo, not a copy — so there is one source of truth and nothing
to sync.

It ships four skills — a *gateway* that routes any ADR task, two for *authoring* (pick the one
matching a repo's ADR format), and one for *obeying* lean briefs while changing code:

- **using-write-adr** — the gateway: broadly discoverable ("Use when ADRs / decisions / adg come up
  in any way"), it routes ADR work to `adg` + the specific skill below rather than hand-rolling it. It
  is the auto-discovery front door; the UserPromptSubmit router hook is its deterministic backstop.
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
settings, and — as of v1.3.0 — **this plugin's own bundled hooks** (`hooks/hooks.json`, below).
Install it once with the prebuilt binary:

```
curl -fsSL https://raw.githubusercontent.com/daniellemccool/ad-guidance-tool/main/install.sh | sh
```

## Bundled hooks

As of **v1.3.0** the plugin ships `hooks/hooks.json` with a suite of fail-open governance hooks, so
enabling the plugin registers them automatically — a repo no longer needs the manual
`.claude/settings.json` snippets from `docs/lean-example/hook-setup.md` (they remain for non-plugin
users). Every hook routes off the same compiled brief and needs system `adg` on `PATH` plus a
`docs/decisions` lean model:

- **SessionStart** (all sources) → `bin/adg-session-start.sh`. In a governed repo (has `docs/decisions/`)
  it **greets** every session — announcing that the write-adr governance is active and its entry points,
  *even when the lean model is empty* (a read-only or mid-migration session meets no other hook) — and
  **version-checks**: when the system `adg` is missing or older than the plugin, tells the agent to prompt
  the user to run `install.sh`. Silent in ungoverned repos.
- **UserPromptSubmit** (fires on every prompt; the script keyword-filters) → `bin/adr-router.sh`. When a
  prompt mentions ADRs / `docs/decisions` / `adg`, injects a pointer telling the agent to do ADR work
  through the write-adr skills + `adg` (author/migrate with `adg lean new`, review, or obey a brief via
  follow-adr-governance) rather than reinventing it. This is the **deterministic** backstop for skill
  auto-discovery, which is unreliable — a grep fires it, not the model's discretion. Silent otherwise.
- **SessionStart** (matcher `^(startup|clear|compact)$`) → `adg lean brief --hook --whole`. Injects the
  **whole-corpus brief** — every in-force ADR, invariants full and defaults condensed — once at session
  start, so the working agreements are in context before the first prompt. (Not on `resume`: the earlier
  injection is already restored.)
- **SubagentStart** (matcher `^Plan$`) → `adg lean brief --hook --invariants`. Injects the **invariants**
  into a `Plan` subagent as it starts designing a change. The dispatch payload carries no paths, so this
  is the always-relevant floor, not a file-scoped brief.
- **PreToolUse** (matcher `^(Edit|Write|MultiEdit|NotebookEdit)$`) → `adg lean brief --hook`. Injects the
  governing ADRs for the file *about to be edited*, **deduped per session** (each ADR at most once per
  session; a forbids violation always re-surfaces).
- **PreToolUse** (matcher `^Bash$`) → `adg lean brief --hook --staged`. The *deterministic* commit layer:
  on a `git commit` call, briefs the **staged** files and **blocks the commit** (`permissionDecision:
  deny`) when a staged path hits a `forbids` glob — a deliberate block
  ([ADR-0005](../../docs/decisions/0005-validation-has-enforcement-tiers.md)).
- **PreToolUse** (matcher `^Bash$`, `if: "Bash(git commit *)"`) → a **`type: agent`** hook: the
  *code-compliance* reviewer (distinct from the ADR-quality reviewer below). Before a commit lands, it
  assesses whether the **staged diff obeys** the ADRs governing the touched files (`git diff --cached` +
  `adg lean brief`) and reports violations. **Advisory** — the deterministic `forbids` block above is the
  only hard stop. Pinned to Sonnet; the commit pauses while it reviews (`if` keeps it off every Bash call).
- **PreToolUse** (matcher `^(Edit|Write|MultiEdit)$`) → `adg lean brief --hook --guard`. Guards the ADR
  model itself: **blocks** a hand-authored *new* record (a `Write` to a not-yet-existing `NNNN-*.md`) so
  records go through `adg lean new`, and **warns** (advisory, no block) on an *edit* to an existing record
  so the write-lean-adr revise/review flow isn't deadlocked. Not deduped.
- **FileChanged** (matcher `docs/decisions/[0-9][0-9][0-9][0-9]-*.md`) → a **`type: agent`** hook that runs
  the ADR review agent (`adg lean review` + the lean rubric) when a record changes *on disk* — the backstop
  for changes made **outside** Claude's Edit/Write (an external editor, git, tooling). ⚠️ **Experimental:**
  the file watcher appears to be interactive-only (it did not fire in headless `-p` runs) and the matcher
  glob support was not verified — confirm it in a live session via `/hooks` before relying on it.

The injection hooks are **fail-open**: no system `adg`, no lean model, or any error means nothing is
injected and the edit proceeds — so "no brief appeared" never means "no rule applies." The two blocking
hooks (commit `forbids`, ADR-record creation) are the deliberate exceptions.

Together these cover the lifecycle: **SessionStart/SubagentStart** put the rules in view at *planning*
time, **PreToolUse** is the *edit-time* safety net, and the **commit advisor** is the last gate before a
change lands. The always-loaded `CLAUDE.md` convention (the "Golden path" in
`docs/lean-example/hook-setup.md`) is now complementary rather than the only planning-time channel.
**Comprehensive enforcement** still belongs at **commit time / CI**: install the lean `pre-commit` gate
(`skills/write-lean-adr/assets/githooks/pre-commit`) to run `adg lean index` + `check` once per commit.
The authoring-side review phase is *not* a hook either — it is driven by the `write-lean-adr` skill,
which has a subagent judge `adg lean review` output against the rubric.
