# Pre-edit hook: inject the architecture brief automatically

This wires the lean `brief` into Claude Code as a **PreToolUse** hook. Before the agent edits a
file, the hook compiles the brief for *that file* and injects it as `additionalContext` — so the
ADRs governing the file are in context at edit time, at the cost of ~20–40 lines (the brief), not the
whole ADR corpus. It is **fail-open**: if no ADR governs the file, or anything errors, the hook emits
nothing and the edit proceeds.

## 1. Install adg

```
go install .            # or: go build -o ~/.local/bin/adg .
```

`adg lean brief --hook` is the hook entry point (no separate binary needed).

## 2. Register the hook

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
  | adg lean brief --hook --model docs/lean-prototype
```

A governed file prints a `hookSpecificOutput.additionalContext` JSON object; an ungoverned file
prints nothing.

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
