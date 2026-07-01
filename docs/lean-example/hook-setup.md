# Hooks: inject the brief, guard the model

The `write-adr` plugin bundles a suite of Claude Code hooks that route the compiled lean brief into the
model across the change lifecycle, and guard the ADR model itself. Every hook is **fail-open** (any error,
or no governing ADR, injects nothing and the action proceeds) except two deliberate blocks, noted below.

| When | Event · matcher | What it does |
|---|---|---|
| Session start | `SessionStart` · startup/clear/compact | inject the **whole-corpus brief** (all in-force ADRs; invariants full, defaults condensed) once |
| Plan dispatch | `SubagentStart` · `Plan` | inject the **invariants** into the planning subagent |
| Before an edit | `PreToolUse` · Edit/Write/MultiEdit | inject the **file-scoped brief** (deduped per session) |
| Before a commit | `PreToolUse` · Bash | brief the **staged** files; **block** on a `forbids` hit (deterministic) |
| ADR write/edit | `PreToolUse` · Edit/Write/MultiEdit | **guard**: **block** hand-creating a record, warn on editing one |
| Before a commit | `PreToolUse` · Bash + `if: git commit *` | **agent**: review whether the staged code **complies** with the governing ADRs (advisory) |
| ADR changed on disk | `FileChanged` · ADR glob | **agent**: review the changed **record** against the lean rubric |

The two blocks are the commit `forbids` gate and the ADR-creation guard ([ADR-0005](../decisions/0005-validation-has-enforcement-tiers.md)); everything else advises.

## 1. Install adg

The hooks invoke `adg` as a bare command, so it must be on your `PATH` (the plugin's `bin/adg` rides along
for the *skills*, but the hooks run outside that PATH). Prebuilt binary, no Go toolchain:

```
curl -fsSL https://raw.githubusercontent.com/daniellemccool/ad-guidance-tool/main/install.sh | sh
```

From source: `go install .`. The hook entry point is `adg lean brief --hook` (with `--whole` / `--invariants`
/ `--staged` / `--guard` selecting the mode); the two agent hooks need no `adg` change — they orchestrate
`git diff` + `adg lean brief`.

## 2. Enable

**With the plugin (recommended):** enabling `write-adr` registers all of the above from
`tools/adr-plugin/hooks/hooks.json` — you only need step 1. Skip to Verify.

**Without the plugin:** copy that `hooks.json` into the target repo's `.claude/settings.json` and point each
`--model` at the repo's lean ADR directory. The hooks run from the project root, so a relative `--model`
resolves, and the edited path is relativized before matching (write `applies_to` globs repo-root-relative,
e.g. `port/**/*.py`). The two agent hooks (`type: agent`) and `FileChanged` are **experimental** — confirm
they fire in a live session before relying on them.

## 3. Verify

Run `/hooks` in Claude Code to confirm registration. Test a `command` hook directly with the JSON Claude
Code sends:

```
printf '{"cwd":"%s","tool_name":"Edit","tool_input":{"file_path":"%s/port/helpers/flow_builder.py"}}' "$PWD" "$PWD" \
  | adg lean brief --hook --model docs/lean-example
```

A governed file prints a `hookSpecificOutput.additionalContext` object; an ungoverned file prints nothing.
`--whole`/`--invariants` take a `SessionStart`/`SubagentStart` payload; `--guard`/`--staged` take the
Edit/Bash payload. The `agent`/`FileChanged` hooks can only be exercised in a live session.

## Notes

- **Advisory routing, not comprehensive enforcement.** The hooks fire on Claude's tool calls only — not on
  shell edits, formatters, or other tools — and are fail-open, so **"no brief appeared" never means "no rule
  applies."** Real enforcement is `adg lean index --root` in CI, the commit gate below, and the executable
  `## Checks`.
- **Token cost is the point:** only matching ADRs are injected; an edit to an ungoverned file costs zero.
  The file-scoped brief dedupes per session (each ADR at most once; a `forbids` hit always re-surfaces); the
  whole-corpus brief fires once per session; the guard never dedupes.
- **Prereq:** the model must be lean records with routing frontmatter (`applies_to`, optional
  `excludes`/`forbids`/`companions`/`priority`). See
  [`lean-format.md`](../../tools/adr-plugin/skills/write-lean-adr/references/lean-format.md).
- **Where the logic lives:** `lean.HookContext` / `SessionBrief` / `SubagentBrief` / `CommitAdvisory` /
  `ModelGuard` in `internal/domain/decision/lean/`; `adg lean brief --hook …` is a thin stdin/stdout shell.

## Commit-time enforcement gate (recommended)

The bundled commit hook advises; hard enforcement belongs in a git `pre-commit`. The lean template
(`tools/adr-plugin/skills/write-lean-adr/assets/githooks/pre-commit`) runs once per commit: `adg lean index
--root .` (validates the model) and `adg lean check` on staged files (the executable grep rules). It is
**graceful** — if `adg` is absent it warns and lets the commit through — so CI is the comprehensive gate.
Install with `cp …/pre-commit .git/hooks/pre-commit` (or the same logic in `.husky/pre-commit`); override the
model dir with `ADR_MODEL` (default `docs/decisions`).

A **`Stop`** hook (`adg lean verify --hook`) that re-shows the brief for changed files at turn-end is
available for manual wiring but **not bundled** — it fires every turn, too often; prefer the commit gate.
