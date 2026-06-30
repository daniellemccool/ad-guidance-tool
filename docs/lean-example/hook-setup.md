# Pre-edit hook: inject the architecture brief automatically

This wires the lean `brief` into Claude Code as a **PreToolUse** hook. Before the agent edits a
file, the hook compiles the brief for *that file* and injects it as `additionalContext` — so the
ADRs governing the file are in context at edit time, at the cost of ~20–40 lines (the brief), not the
whole ADR corpus. It is **fail-open**: if no ADR governs the file, or anything errors, the hook emits
nothing and the edit proceeds. It **dedupes per session**: each ADR is injected at most once per
Claude Code session (keyed by the payload's `session_id`), so repeated edits to broadly-scoped files
don't re-pay for the same brief — a forbids violation always re-surfaces, and with no `session_id` it
injects every time.

## 1. Install adg

These hooks invoke `adg` as a bare command, so it must be on your `PATH`. Install the prebuilt binary
(no Go toolchain needed):

```
curl -fsSL https://raw.githubusercontent.com/daniellemccool/ad-guidance-tool/main/install.sh | sh
```

Or from source: `go install .` (or `go build -o ~/.local/bin/adg .`). `adg lean brief --hook` is the
hook entry point — no separate binary needed.

## 2. Register the hook

> **Using the `write-adr` plugin (v1.2.0+)?** Skip this step. The plugin bundles this exact
> PreToolUse hook in `hooks/hooks.json`, so enabling the plugin registers it automatically — you
> still need step 1 (system `adg` on `PATH`). The manual registration below is for repos that wire
> the hook in directly without the plugin.

In the target repo's `.claude/settings.json` (project scope), point `--model` at that repo's lean ADR
directory:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "adg lean brief --hook --model docs/decisions",
            "timeout": 30
          }
        ]
      }
    ]
  }
}
```

The hook runs from the project root, so a relative `--model` resolves correctly. The edited file's
absolute path is relativized against the project root before matching, so `applies_to` globs are
written repo-root-relative (e.g. `port/**/*.py`).

## 3. Verify

Run `/hooks` in Claude Code to confirm the `PreToolUse` hook is registered, then edit a file that an
ADR's `applies_to` matches — the brief appears in context. To test the binary directly without a live
session, feed it the same JSON Claude Code sends:

```
printf '{"cwd":"%s","tool_name":"Edit","tool_input":{"file_path":"%s/port/helpers/flow_builder.py"}}' "$PWD" "$PWD" \
  | adg lean brief --hook --model docs/lean-example
```

A governed file prints a `hookSpecificOutput.additionalContext` JSON object; an ungoverned file
prints nothing.

## Golden path: the always-on convention

The hook injects the brief at *edit* time — a safety net once the agent touches a governed file. But
the higher-value moment is *planning*: the agent should pull the brief for the paths it expects to
touch **before** it designs a change, so it doesn't commit to (say) a forbidden second pipeline and
only get told at write time. Don't rely on the agent *auto-discovering* the `write-adr` skills to do
this — skill auto-discovery on phrasings like "add an ADR" or "comply with our decisions" is
unreliable in practice; the brief reaches the model dependably only through the hook (at edit time)
and an always-loaded convention (at planning time).

So put a short convention in the **consuming repo's `CLAUDE.md`**, which the agent reads every session:

```markdown
## Architecture decisions (adg)

Working agreements live as lean ADRs in `docs/decisions/`, compiled into a brief.
**Consult them while planning a change, not just at write time** — a PreToolUse hook injects the
brief on edits, but pull it yourself *before* you design the change:

    adg lean brief --model docs/decisions <paths you expect to touch>

Treat the brief as constraints:
- **Invariant** → a hard rule; read the full ADR before planning.
- **Forbidden scope matched** → stop and surface the conflict; don't build it.
- **Companions** → check whether the related files also need edits.
- **No brief appeared** → never assume no rule applies.

After editing, run `adg lean index --model docs/decisions --root .` and the tests.
If the change establishes a reusable pattern, record it with `adg lean new`.
```

The two layers are complementary: the **convention** gets the agent to consult governance up front —
it catches feature work like "create a new extractor," which the hook only sees once files are being
written — and the **hook** is the backstop for edits the agent makes without consulting it first.

## Notes

- **Advisory routing, not comprehensive enforcement.** This hook fires only on Claude's
  `Edit`/`Write`/`MultiEdit`. It will **not** catch shell edits, formatters, generated rewrites, manual
  human edits, or other agents/tools, and it is fail-open — so **"no brief appeared" never means "no rule
  applies."** Use `adg lean index --root` in CI, code review, or executable checks for actual enforcement.
- **Token cost** is the point: only matching ADRs are injected. An edit to an ungoverned file costs
  zero tokens. A file under a broad invariant plus a couple of path-scoped defaults is ~30–40 lines.
- **Prereq:** the repo's ADRs must be in the lean format with routing frontmatter — `applies_to`
  (and optionally `excludes` / `forbids` / `companions` and `priority`). The hook routes on
  `applies_to` and `forbids` (an edit to a forbidden path injects the rule as a violation) and honors
  `excludes`; until a repo's records are migrated/annotated it simply no-ops there. See the lean-format
  reference (`tools/adr-plugin/skills/write-lean-adr/references/lean-format.md`) for the full schema.
- **Contract:** `PreToolUse` + exit 0 + `{"hookSpecificOutput":{"hookEventName":"PreToolUse","additionalContext":"…"}}`
  injects to the model without blocking the edit. The logic lives in `lean.HookContext`
  (`internal/domain/decision/lean/hook.go`); `adg lean brief --hook` is a thin stdin/stdout shell.

## Post-edit hook (optional): re-check on stop

The PreToolUse hook prevents mistakes *before* an edit; a **Stop** hook catches misses *after*.
`adg lean verify --hook` re-runs validation (and `--root .` scope lint) and re-renders the brief —
with its "Before you finish" footer (checks + named tests to run) — for the files changed this
session (derived from git: working tree vs HEAD, plus untracked). It is **advisory and
non-blocking**: output goes to stderr and it always exits 0, so it never prevents stopping.

> **Note:** the `write-adr` plugin does **not** bundle this Stop hook — it fires on every turn-end,
> which is usually too frequent. The plugin bundles only the per-session-deduped PreToolUse hook. For
> periodic enforcement, prefer the commit-time gate (next section) over a Stop hook. The snippet below
> is for repos that still want the Stop re-check wired in manually.

```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "adg lean verify --hook --model docs/decisions",
            "timeout": 60
          }
        ]
      }
    ]
  }
}
```

What it proves and doesn't: re-running the index validates the **model and scope routing**, not that
the edited code obeys the prose guidance — so the footer is a prompt to verify, not a proof of
compliance. Code-level enforcement comes from executable `checks` (`adg lean check`) and CI. Blocking
behavior is deliberately *not* wired here yet; the Stop hook only warns.

## Commit-time enforcement gate (recommended)

Enforcement belongs at commit time, not on every edit or every turn. The lean `pre-commit` template
(`tools/adr-plugin/skills/write-lean-adr/assets/githooks/pre-commit`) runs once per commit:
`adg lean index --root .` (validate the model — hard-fails on duplicate IDs or bad globs) and
`adg lean check` on the staged files (the executable grep rules). It is **graceful**: if `adg` is not
on `PATH` it warns and lets the commit through, so contributors without `adg` are never blocked — use
CI for comprehensive enforcement.

Install it directly (`cp …/pre-commit .git/hooks/pre-commit`) or, in a repo using Husky, drop the
same logic into `.husky/pre-commit`. Override the model dir with `ADR_MODEL` (default `docs/decisions`).
