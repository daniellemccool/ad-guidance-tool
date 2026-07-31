---
status: accepted
date: "2026-07-31"
category: Architecture
applies_to:
    - internal/domain/decision/lean/hook.go
    - internal/domain/decision/lean/hooksession.go
    - tools/adr-plugin/hooks/hooks.json
priority: invariant
companions:
    - internal/domain/decision/lean/hook_test.go
---

# A fresh context always receives its first brief

## Decision

Hook-brief dedup is scoped to the injection context — the session, or the subagent within it (`session_id` + `agent_id`) — never the session alone. Subagent briefs are two-tier: Plan subagents get the whole-corpus brief at start; implementers get the file-scoped brief at first edit. A fresh context is never starved by another context's dedup state.

## Guidance

- `HookContext` composes its dedup key from `session_id` plus `agent_id` when the payload carries one (`hook.go`); never key dedup state on the session alone — a subagent shares the parent's `session_id` but is a fresh context that needs its own first injection.
- The SubagentStart hook injects the whole-corpus brief (`--whole`) for Plan agents; implementer subagents are served by the PreToolUse edit-time brief under the per-context key — do not add a start-time file-scoped subagent brief (SubagentStart payloads carry no task paths).
- Within one context the existing behavior holds: repeated edits dedup, a forbids hit always re-emits, and no `session_id` means no dedup at all.

## Why

Subagents share the parent's session id, so session-keyed dedup silently muted every brief to fresh-context implementer subagents — governance went dark exactly where review needs it most, and no test failed while it happened.
