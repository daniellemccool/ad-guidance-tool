---
status: accepted
date: "2026-06-29"
category: Validation
priority: invariant
applies_to:
    - internal/domain/decision/lean/validate.go
    - internal/domain/decision/lean/lint.go
    - internal/domain/decision/lean/hook.go
    - internal/adapter/command/lean/index.go
    - internal/adapter/command/lean/brief.go
    - internal/application/interactor/model/validate.go
---

# Validation has enforcement tiers

## Decision

Every validation check is assigned exactly one enforcement tier — hard failure, warning, or fail-open —
and the choice is deliberate, not incidental. The PreToolUse hook is always fail-open, regardless of any
individual check's tier.

## Guidance

- A new check must declare its tier: a **hard failure** (a non-warning `Issue` that exits `adg lean
  index` / `adg validate` non-zero) only if a malformed model should stop CI; otherwise a **warning**
  (`Issue{Warning: true}`, printed but never blocking).
- The hook path (`adg lean brief --hook` → `HookContext`) is **fail-open**: it injects guidance or
  nothing and never blocks or errors an edit, regardless of a check's tier — even a hard-failure-tier
  finding. Enforcement is CI's job.
- `adg lean brief` (non-hook) may surface validation to stderr and exit non-zero on hard failures.

## Why

Without explicit tiers, future checks drift: a check added "to be safe" makes the hook too strict
(blocking edits) or makes CI too soft (warnings that should fail). Pinning each check to a tier keeps
the advisory hook ergonomic and the enforcing gate strict.
