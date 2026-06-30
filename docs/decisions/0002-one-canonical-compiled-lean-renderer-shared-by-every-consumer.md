---
status: accepted
date: "2026-06-29"
source: lean tool pass — the shared-renderer boundary, made explicit
category: Architecture
applies_to:
    - internal/domain/decision/lean/brief.go
    - internal/domain/decision/lean/index.go
    - internal/domain/decision/lean/hook.go
    - internal/adapter/command/lean/**/*.go
    - cmd/lean.go
priority: invariant
---

# One canonical compiled lean renderer shared by every consumer

## Decision

There is exactly one compiled lean brief/index renderer — `lean.Brief`, `lean.RenderIndex`, and
`lean.HookContext` in `internal/domain/decision/lean` — shared, unchanged, by every consumer: the `adg`
CLI (the `adg lean` commands), the PreToolUse hook, CI, and any other tool that renders a lean brief or
index. Duplicating the renderer, or moving/refactoring it so the hook and CI would render differently
from the CLI, is forbidden.

## Guidance

- Keep the renderer in the domain (`internal/domain/decision/lean`); every consumer calls it. Do not
  copy its logic into an adapter, a presenter, a tool, or the hook.
- The deferred promotion of the lean commands onto the Clean Architecture stack must have its presenter
  delegate to this renderer, never reimplement formatting — this is the named thin-shell exception and
  its promotion path.
- A change to the renderer changes it for every consumer at once (by construction); never add a code
  path that renders for one consumer differently from another.

## Why

The hook advises and CI enforces off the same compiled bytes; a second or diverged renderer would let
the hook show one thing while `adg lean index`/CI checks another — the guidance an agent sees at edit time
would no longer match what the gate enforces. One renderer is what keeps advisory routing and
enforcement honest about the same output.
