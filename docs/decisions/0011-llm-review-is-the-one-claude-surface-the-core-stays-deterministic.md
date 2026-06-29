---
status: accepted
date: "2026-06-29"
category: Architecture
applies_to:
    - internal/adapter/leanreview/review.go
    - internal/adapter/command/lean/review.go
priority: invariant
checks:
    - desc: the deterministic lean core does not import the Anthropic SDK
      grep: anthropics/anthropic-sdk-go
      in:
          - internal/domain/decision/lean/*.go
      expect: absent
---

# LLM review is the one Claude surface; the core stays deterministic

## Decision

`adg lean review` is the only part of the lean toolchain that calls an LLM. Routing, brief compilation, indexing, and grep checks are deterministic and never import the Anthropic SDK.

## Guidance

- Keep LLM calls inside `internal/adapter/leanreview`; the deterministic core (`internal/domain/decision/lean`) must not import the SDK.
- Review is advisory — judge against the rubric and report findings; it blocks only when run with `--fail-on-revise`.
- Default to a cheap, reliable model (Sonnet); escalate with `--reviewer` for contentious cases.

## Why

If routing or the brief depended on an LLM, the same edit could get different governance on different runs — the advisory hook and the enforcing index would diverge and CI would be nondeterministic. Confining the LLM to review keeps "route decides, brief renders" reproducible and keeps the hot path free of network calls and cost.
