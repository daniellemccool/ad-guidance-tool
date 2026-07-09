---
status: accepted
date: "2026-07-09"
source: sub-project B retired the MADR command family and with it the inputport/interactor/presenter stack
category: Architecture
applies_to:
    - cmd/**/*.go
    - internal/adapter/command/**/*.go
priority: invariant
forbids:
    - internal/application/**/*.go
supersedes:
    - "0003"
---

# Commands are thin cobra adapters over the domain

## Decision

An `adg` command is a thin cobra adapter in `internal/adapter/command/<group>` that parses flags, calls the domain directly, and picks the right output stream. The inputport/interactor/outputport/presenter stack is retired and must not reappear; business logic lives in the domain, never in the cobra layer.

## Guidance

- A new command is a thin adapter: flag parsing, a call into the domain, and output routed per the stdout/stderr convention. Review rejects a command that grows business logic; the fix path is moving that logic into the domain package that owns it.
- Do not reintroduce ports, interactors, or per-group presenters (`internal/application/`, `internal/adapter/printer/<group>`). Shared rendering belongs to the domain renderer (`lean.Brief` / `lean.RenderIndex`); adapters call it and never reimplement formatting.
- `internal/adapter/printer` survives only as the IO-streams helper (`streams.go` and its test double); new formatting code does not go there.

## Why

Business logic that grows in a cobra adapter escapes the domain test suites and can only be exercised through cobra execute tests, and a reintroduced port/interactor/presenter layer would recreate a second logic-and-rendering path that drifts from the domain every other consumer shares — rebuilding architecture whose promotion target no longer exists.
