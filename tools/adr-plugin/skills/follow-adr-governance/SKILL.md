---
name: follow-adr-governance
description: >
  Use when editing code in a repository governed by `adg` lean ADR briefs — especially when
  an architecture brief is injected as context, or the user asks to comply with ADR
  governance. A behavior primer for obeying the brief while changing code; the PreToolUse
  hook and the compiled brief do the real work. To create or rewrite ADRs, use write-lean-adr.
---

# follow-adr-governance — obey the lean ADR brief while changing code

In an `adg`-governed repo, a deterministic **architecture brief** is compiled for the files
you touch and usually injected automatically by the PreToolUse hook as `additionalContext`.
The brief is self-contained and authoritative — **follow it**. You do not need to relearn how
ADRs are authored; the rules are already in the brief.

## While changing code

- **Treat an injected brief as authoritative.** It lists the ADRs that govern the changed
  files, grouped by force, plus the checks and tests to run.
- **Invariants are hard constraints.** Everything under "Hard constraints (invariants)" must
  hold. If an invariant's guidance is ambiguous, open the full record before planning.
- **Defaults are conventions** — follow them unless one conflicts with an invariant, in which
  case the invariant wins.
- **`⚠ Forbidden scope matched`** means the edit touches a path an ADR says must not exist:
  **stop and surface the conflict**, don't proceed.
- **Companions** are related files to consider (e.g. the TS side of a prop), *not* governed —
  make the partner edit, but the ADR doesn't govern it.
- **If a requested change violates an invariant, stop and surface the conflict** rather than
  silently complying or silently breaching the rule.

## If no brief was injected

The hook fires only on Claude `Edit`/`Write`/`MultiEdit` and is fail-open — "no brief" never
means "no rule applies." Compile it yourself for the paths you're about to touch:

```bash
adg lean brief --model docs/decisions <changed-path> [more-paths]
```

## Before you finish

Do what the brief's **"Before you finish"** footer says:

- Run the **checks** it lists and the **tests** the matched ADRs name.
- Re-run `adg lean index --model docs/decisions --root .` — this validates the model and scope
  routing, not that the code obeys the prose guidance, so still confirm compliance yourself.

This can be automated with a Stop hook (`adg lean verify --hook`) that re-runs the gate and
re-shows the footer for changed files — advisory, non-blocking. See the tool's hook-setup doc.
