---
status: accepted
date: "2026-06-29"
category: Validation
priority: invariant
applies_to:
    - internal/domain/decision/lean/validate.go
    - internal/domain/decision/lean/lint.go
    - internal/domain/decision/lean/hook.go
    - internal/domain/decision/lean/hookcorpus.go
    - internal/adapter/command/lean/index.go
    - internal/adapter/command/lean/brief.go
    - internal/application/interactor/model/validate.go
---

# Validation has enforcement tiers

## Decision

Every validation check is assigned exactly one enforcement tier — hard failure, warning, or fail-open —
deliberately, not incidentally. Context-injection hooks are always fail-open; the commit-time advisor is
the sole exception, and it blocks only on a forbidden-scope violation.

## Guidance

- A new check must declare its tier: a **hard failure** (a non-warning `Issue` that exits `adg lean
  index` / `adg validate` non-zero) only if a malformed model should stop CI; otherwise a **warning**
  (`Issue{Warning: true}`, printed but never blocking).
- The injection paths (`adg lean brief --hook` → `HookContext`, `--whole` → `SessionBrief`,
  `--invariants` → `SubagentBrief`) are **fail-open**: they inject guidance or nothing and never block or
  error an edit/session/dispatch, regardless of a check's tier — even a hard-failure-tier finding.
- The commit advisor (`--staged` → `CommitAdvisory`) is the **one deliberate block**: a staged path that
  hits a `forbids` glob returns a PreToolUse `permissionDecision: deny` carrying the brief; any other
  governed commit is advisory (`additionalContext`), and any parse/git error injects nothing. A block
  belongs at commit time (a whole change), never on a single edit. Enforcement otherwise is CI's job.
- `adg lean brief` (non-hook) may surface validation to stderr and exit non-zero on hard failures.

## Why

Without explicit tiers, future checks drift: a check added "to be safe" makes the hook too strict
(blocking edits) or makes CI too soft (warnings that should fail). Pinning each check to a tier keeps
the advisory hook ergonomic and the enforcing gate strict.
