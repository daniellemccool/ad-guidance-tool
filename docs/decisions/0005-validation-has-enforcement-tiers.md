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

Every validation check belongs to exactly one enforcement tier:

1. **Hard failure** — a non-warning issue fails the CLI/CI (`adg lean index` / `adg validate` exit non-zero).
2. **Warning** — advisory only; printed, never blocks (the `Issue.Warning` flag).
3. **Fail-open** — the PreToolUse hook never blocks or errors an edit; it injects guidance or nothing.

## Guidance

- A new check must declare its tier. Make it a hard failure only if a malformed model should stop CI;
  otherwise make it a warning (`Issue{Warning: true}`).
- The hook path (`adg lean brief --hook` → `HookContext`) stays fail-open regardless of any check's tier —
  never make the hook block an edit, even on a hard-failure-tier finding. Enforcement is CI's job.
- `adg lean brief` (non-hook) may surface validation to stderr and exit non-zero on hard failures, but the
  hook contract above is inviolate.

## Why

Without explicit tiers, future checks drift: a check added "to be safe" makes the hook too strict
(blocking edits) or makes CI too soft (warnings that should fail). Pinning each check to a tier keeps
the advisory hook ergonomic and the enforcing gate strict.
