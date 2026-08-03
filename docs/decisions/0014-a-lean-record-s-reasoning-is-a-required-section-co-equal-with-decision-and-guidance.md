---
status: accepted
date: "2026-07-01"
category: ADR formats
applies_to:
    - internal/domain/decision/lean/validate.go
    - internal/domain/decision/lean/template.go
    - internal/domain/decision/lean/brief.go
    - internal/domain/decision/lean/digest.go
priority: invariant
---

# Require reasoning (## Why) in every accepted lean record

## Decision

Every accepted lean record must carry a `## Why`, and that reasoning is record-only: the validator treats `## Why` as a required section co-equal with Decision and Guidance, and no brief ever renders it.

## Guidance

- `validate.go` hard-fails an accepted record with no `## Why` (the same tier as a missing Decision); a not-yet-accepted draft only warns.
- `template.go` scaffolds `## Why` for every new record, so `adg lean new --status accepted` refuses a record whose reasoning is unfilled.
- No brief renderer (`brief.go`, `digest.go`) renders `## Why` in any mode — a brief stays Decision + Guidance only (the digest, titles only), so injection cost is unchanged.

## Why

A rule recorded without its reason can only be obeyed or violated, never reasoned about or generalized — the first failure a growing, non-authoring audience hits. Keeping the reasoning in the record and out of the brief makes the harness teach without inflating every context injection.
