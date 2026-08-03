---
status: accepted
date: "2026-08-03"
category: Architecture
applies_to:
    - internal/domain/decision/lean/digest.go
    - internal/domain/decision/lean/hookcorpus.go
    - internal/domain/decision/lean/template.go
    - tools/adr-plugin/hooks/hooks.json
priority: invariant
companions:
    - internal/domain/decision/lean/digest_test.go
amends:
    - "0015"
---

# Inject the session-open digest, capped at 2 KB

## Decision

Session open injects a distinct brief class — the digest, a grouped titles-only tripwire index — hard-capped at `MaxDigestBytes` (2048 bytes) by an internal degradation ladder. The digest is fail-open, never configurable, and additive: no other brief class changes shape to serve session start.

## Guidance

- Titles are the payload: bare `NNNN` IDs, `!` invariant markers, category groups ordered like the README index, per-group scope hints — never per-record summaries, filenames, or section prose. The digest's job is recall, not instruction; full guidance stays on demand via `adg lean brief <paths>`.
- The ladder in `digest.go` tries: grouped digest → flat digest (no group headers or scope hints — grouping is a bonus the budget may not afford) → invariant-only digest plus a defaults count → a fixed floor (counts + pointers). The first render within `MaxDigestBytes` wins; the floor always fits by construction, so the path never errors (fail-open per the enforcement-tiers rule).
- Corpus tuning measures against the real renderer: `adg lean brief --digest` (non-hook) prints the digest and reports every rung's size (`DigestReport`) — never re-implement the renderer to predict fit (the one-renderer rule).
- The ceiling is the exported const `MaxDigestBytes` in `template.go`; no project config may change it or anything the digest renders. Digest lines truncate titles at `MaxTitleRunes` runes.
- Only the SessionStart hook consumes the digest (`adg lean brief --hook --digest`). Plan subagents keep the whole-corpus brief, and the edit/commit paths keep their routed briefs — do not point another moment at the digest without its own record.

## Why

A 44 KB session-open whole-brief exceeded Claude Code's 10,000-character hook cap, was spooled to a file, and reached the agent as a ~2 KB preview: the governance silently vanished, and the one record that would have prevented a real mistake never entered context. A digest that always fits inline guarantees the recall layer actually arrives.
