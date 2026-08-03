---
status: amended by ADR-0021
date: "2026-07-08"
category: Architecture
applies_to:
    - internal/adapter/command/lean/budget.go
    - internal/domain/decision/lean/budget.go
    - internal/domain/decision/lean/validate.go
priority: invariant
---

# Keep project config in .adg.yaml; it never changes a brief

## Decision

Project-scoped `adg` settings live in one per-project file, `<model-root>/.adg.yaml`, read only at the command edge and passed into the domain as a value; the domain never reads config. A setting may relax an authoring-time nudge, but it must never change what a compiled brief injects.

## Guidance

- A new project-scoped setting goes in `<model-root>/.adg.yaml`, is loaded in the command layer (`internal/adapter/command/lean`, e.g. `loadBudget`/`budgetFor`), and is handed to the domain as a value object (e.g. `lean.Budget`); packages under `internal/domain/decision/lean` must not read config files or learn the config's on-disk shape.
- A knob may relax an authoring check (e.g. `body_budget: narrative` relaxes only the whole-body one-screen nudge) but must leave the agent-facing budgets — `MaxDecisionWords`, the brief's `MaxBriefLines`, the digest's `MaxDigestBytes`, and `MaxTitleRunes` — and everything a brief renders unchanged; the injected brief's content and token cost are independent of any project config.
- Config is advisory: an absent, unknown, or malformed setting degrades to the built-in default and warns on stderr (fail-open per ADR-0005, stderr per ADR-0008) — it never hard-fails a command, and a degrade path returns the real default, never a zero value.

## Why

One tool must serve repos with different needs — terse governance (nobody reads the records) versus full ADR narrative (a human wants to read the reasoning) — without a second format or global user state; a per-project file carried in the repo is what makes that possible. Reading config only at the edge preserves the domain's purity (ADR-0003) and the single-renderer guarantee (ADR-0002). The hard line — a knob may change authoring discipline but never the brief — is what protects the property many repos depend on: that the agent-facing brief is low-token and identical across every consumer regardless of local settings. If a setting could reshape the brief, project config would silently change what every governed edit sees, and that guarantee would be gone.
