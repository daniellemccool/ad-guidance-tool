---
status: accepted
date: "2026-06-29"
category: Architecture
source: architecture of the initial adg implementation; models/clean is template seed data, not this decision
priority: invariant
applies_to:
    - cmd/**/*.go
    - internal/adapter/command/**/*.go
    - internal/application/**/*.go
    - internal/adapter/printer/**/*.go
---

# Stable commands run through the Clean Architecture stack

## Decision

A graduated (stable) `adg` command runs the full Clean Architecture stack: a cobra adapter in
`internal/adapter/command/<group>` calls an `inputport` interface, implemented by an `interactor` in
`internal/application/interactor/<group>`, which talks to the domain and returns through an `outputport`
to a presenter in `internal/adapter/printer/<group>`. The cobra layer parses flags and wires; it holds
no business logic, and the domain does not format output.

## Guidance

- A new stable command adds an `inputport` interface, an `interactor`, an `outputport` + presenter, and
  a thin cobra adapter that depends only on the inputport — see `internal/adapter/command/model/
  validate.go` for the canonical shape.
- A command may be a **thin shell that calls the domain directly only when an ADR names it as an
  explicit, time-boxed exception with a promotion path** (the lean `brief`/`index` commands, per ADR
  0002). Absent such an ADR, the thin-shell shortcut is a violation, not a style choice.
- Keep business logic out of `cmd/` and `internal/adapter/command/`; keep presentation out of the domain.

## Why

The stack keeps the CLI testable and the domain reusable (the same interactors back any future surface).
Allowing thin shells only behind a named, promotion-tracked ADR is what stops a "temporary" shortcut
from becoming permanent architectural debt — it forces ADR 0002 to resolve rather than linger.

## Checks

- A new command under `internal/adapter/command/` has a matching interactor (`internal/application/
  interactor/`) and presenter (`internal/adapter/printer/`); a command that imports the domain directly
  and formats its own output is a thin shell and must cite an exception ADR.
