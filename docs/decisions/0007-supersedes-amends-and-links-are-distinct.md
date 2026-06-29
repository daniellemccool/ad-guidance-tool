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

`supersedes`, `amends`, and `links` are three separate relationships, not interchangeable labels:

- **supersedes** — replacement: the new record takes over and the old is retired (bidirectional; the old
  record's status flips to "superseded by ADR-NNNN").
- **amends** — additive correction: the base record stays in force; the amendment adjusts or extends it.
- **links** — loose association: a non-precedence cross-reference, with no status or force implication.

## Guidance

- Use supersession for replacement, amendment for additive correction, and a link for mere association;
  never overload one for another (don't "supersede" when you only meant to relate).
- Keep both ends consistent: the integrity checks in `lean/validate.go` verify forward + reverse for
  supersedes and amends — a relationship-bearing change must preserve that bidirectional invariant.
- `links` must not carry supersede/amend semantics (the `Link` path already refuses supersede tags).
