---
status: accepted
date: "2026-06-29"
category: Decision model
source: docs/fork-design/0005-make-supersede-a-first-class-bidirectional-command.md
priority: default
applies_to:
    - internal/domain/decision/service.go
    - internal/domain/decision/lean/validate.go
---

# Supersedes, amends, and links are distinct relationships

## Decision

`supersedes`, `amends`, and `links` are three distinct relationships with different precedence, status,
and force semantics — not interchangeable labels. A relationship-bearing change must use the one that
matches the author's actual intent.

## Guidance

- Pick by intent and never overload one for another: **supersede** when a new record replaces and
  retires an old one (bidirectional — the old record's status flips to "superseded by ADR-NNNN");
  **amend** for an additive correction that leaves the base record in force; **link** for a loose
  cross-reference with no precedence, status, or force.
- Keep both ends consistent: `lean/validate.go` verifies forward + reverse integrity for supersedes and
  amends, so a relationship-bearing change must keep both ends in sync.
- `links` must not carry supersede/amend semantics (the `Link` path already refuses supersede tags).
