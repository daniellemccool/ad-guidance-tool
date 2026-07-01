---
status: accepted
date: "2026-03-13"
category: Fork governance
tags:
    - process
    - documentation
---

# Use MADR for architectural decision records

## Decision

Architectural decisions are recorded as ADRs under `docs/decisions/`, managed with `adg`, and
created at decision time. The structured format captures the *why* behind a rule so contributors
can apply it in novel situations, not just obey it.

## Guidance

- When a significant architectural choice is made (new package boundary, new pattern, divergence
  from upstream), create an ADR before or alongside the implementing PR.
- Code review may ask "is this decision documented?" and block on a missing ADR.
- ADRs live in the repo, visible in PRs — rationale survives contributor turnover.

## Why

A rule recorded without its reasoning gets obeyed until it is inconvenient, then quietly broken;
capturing the why is what lets a contributor apply a decision to a case it never foresaw — and what
keeps the decision from eroding as the people who made it move on.

## Alternatives

Prose in `ARCHITECTURE.md` (captures rules but not rationale), inline code comments (not
discoverable as decisions), and no formal records (the status quo that was already causing drift)
were rejected.
