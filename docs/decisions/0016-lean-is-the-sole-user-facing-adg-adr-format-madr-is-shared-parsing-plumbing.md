---
status: accepted
date: "2026-07-08"
category: ADR formats
applies_to:
    - internal/domain/decision/madr/**/*.go
    - tools/adr-plugin/skills/**/SKILL.md
supersedes:
    - "0004"
priority: default
---

# Lean is the sole user-facing adg ADR format; madr is shared parsing plumbing

## Decision

`adg` has one user-facing ADR format: lean. There is no MADR authoring/decide lifecycle and no second authoring skill. The `madr` package survives only as shared, format-neutral parsing plumbing (frontmatter, file-split, `RenderFile`) that the lean package reuses.

## Guidance

- Do not reintroduce a second user-facing format or a `decide`/options authoring lifecycle; new authoring behavior extends the lean commands (`adg lean …`) and the write-lean-adr skill.
- The `madr` package is plumbing, not a format: keep only frontmatter/file-split/`RenderFile` primitives there and do not fork them; a lean record is authored and validated only through the lean path.
- Migrate an existing MADR-format record by authoring a lean record for it (`adg lean new --from-stdin --date <original>`); there is no auto-converter.

## Why

Maintaining two user-facing flavors was the real cost, and the MADR `decide` lifecycle was the sole source of the field-reported bugs; collapsing to lean removes both. The shared parsing core is retained (the surviving half of ADR-0004) so the plumbing is never re-forked — the split that mattered was at the user-facing layer, and now there is only one side of it.
