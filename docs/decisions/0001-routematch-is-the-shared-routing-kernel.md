---
status: accepted
date: "2026-06-29"
category: Architecture
source: feat/lean-tool-pass — extracted while adding excludes/forbids/companions
priority: invariant
applies_to:
    - internal/domain/decision/lean/route.go
    - internal/domain/decision/lean/brief.go
    - internal/domain/decision/lean/lint.go
    - internal/domain/decision/lean/hook.go
tags:
    - routing
    - lean-format
---

# Route matching is the single shared routing kernel

## Decision

All lean routing — "given a record and a set of changed paths, what does the record govern?" — goes
through one function, `routeMatch` in `route.go`. The compiled brief, the `Matches` hook gate, and
`LintTree`'s scope-overlap calculation all derive their answer from it. `routeMatch` returns the
governed paths plus the matched / excluded / forbidden / companion globs; no other code re-derives
"does this rule apply here?".

## Guidance

- New routing behavior (`applies_to`, `excludes`, `forbids`, `companions`, future scope keys) is added
  to `routeMatch` and consumed from its result — never reimplemented in `brief.go`, `lint.go`, or the
  hook.
- `Brief`, `Matches`, and the `LintTree` overlap pass must keep obtaining governed paths / matched
  globs from `routeMatch`; a second path-matching loop is the thing this ADR exists to prevent.
- Rendering (the brief packet, the index) stays in `brief.go` / `index.go`; routing stays in
  `route.go`. "Route decides, brief renders."

## Why

The brief, the PreToolUse hook, and CI overlap detection answer the same question. When that logic was
duplicated they could disagree — a file governed in the brief but not counted for overlap, or an
`excludes` honored on one path and not another. A single kernel makes "what does this rule govern?"
answerable in exactly one place, so the advisory hook and the enforcing index never diverge.

## Checks

- Confirm `Brief`, `Matches`, and the overlap pass in `LintTree` all obtain governed paths / matched
  globs from `routeMatch`, not from a private glob loop. (Per-pattern stale / forbids-has-files checks
  in `lint.go` calling the glob engine directly are fine — that is existence testing, not routing.)
