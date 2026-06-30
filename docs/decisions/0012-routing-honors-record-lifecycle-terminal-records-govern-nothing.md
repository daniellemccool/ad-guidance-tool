---
status: accepted
date: "2026-06-30"
category: Architecture
applies_to:
    - internal/domain/decision/lean/route.go
priority: invariant
checks:
    - desc: the routing kernel gates on record lifecycle (inForce)
      grep: inForce
      in:
          - internal/domain/decision/lean/route.go
      expect: present
---

# Routing honors record lifecycle; terminal records govern nothing

## Decision

`routeMatch` returns nothing for a terminal record — one whose status is rejected, deprecated, or superseded. Retiring a record removes it from every form of governance: briefs, the PreToolUse hook gate, scope lint, and the leanness nudges.

## Guidance

- To retire a rule, set its status (`superseded by ADR-NNNN`, `deprecated`, or `rejected`); it stops routing automatically — do not also strip its `applies_to`/`forbids`.
- The in-force record (the replacement) carries the routing; the retired one keeps its body as history and still appears in the index README.
- Lifecycle gating lives in one place, `inForce`, called by the routing kernel; never add a second status check in a brief/hook/lint path.

## Why

Previously a superseded record kept routing as long as it carried `applies_to`, so a retired rule lingered in every brief and the hook until someone manually stripped its globs — easy to forget, and it silently mis-governed edits. Gating routing on lifecycle in the kernel makes "flip the status" the complete retire action, and keeps governance showing only rules that are actually in force.
