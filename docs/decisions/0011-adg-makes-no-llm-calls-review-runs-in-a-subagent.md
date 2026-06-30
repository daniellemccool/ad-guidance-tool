---
status: accepted
date: "2026-06-29"
category: Architecture
applies_to:
    - internal/adapter/command/lean/review.go
priority: invariant
checks:
    - desc: adg has no LLM dependency — nothing imports the Anthropic SDK
      grep: anthropic-sdk-go
      in:
          - "**/*.go"
          - go.mod
      expect: absent
---

# adg makes no LLM calls; ADR review runs in a Claude Code subagent

## Decision

`adg` is fully deterministic and depends on no LLM. ADR review is performed by a Claude Code subagent (driven by the write-lean-adr skill) that judges the deterministic `adg lean review` packet against the rubric, using the session's own model access — no Anthropic SDK, no API key.

## Guidance

- `adg lean review` emits a deterministic packet (the target ADRs plus their lint findings); the judging happens in the agent, never in the tool.
- Do not add an LLM client or the Anthropic SDK to `adg`; review uses the session's model access, so it needs no `ANTHROPIC_API_KEY`.
- Gate CI with the deterministic checks (`adg lean index --root`, `adg lean check`); LLM review is an interactive, advisory aid, not a build gate.

## Why

Keeping every governance decision the tool makes deterministic is what lets the advisory hook and the enforcing index agree run-to-run — an LLM in routing or the brief would make the same edit get different governance on different runs, and CI nondeterministic. It also keeps `adg` free of network calls, per-token cost, and an API-key requirement. The judgment that genuinely benefits from a model runs where a model already is — the Claude Code session — not bolted into the CLI.
