---
status: accepted
date: "2026-08-03"
category: Decision model
applies_to:
    - internal/domain/decision/lean/validate.go
    - internal/domain/decision/lean/template.go
    - tools/adr-plugin/skills/write-lean-adr/references/lean-rubric.md
priority: default
---

# Write ADR titles as short imperative instructions

## Decision

A record's title is a short imperative instruction — commit-message mood, at most `MaxTitleRunes` runes — because on the session digest the title is the only thing an agent sees. The validator warns (never blocks) on an over-long title of an in-force record; mood is judged by the authoring rubric.

## Guidance

- Lead with the verb and state the rule as an order: "Derive failure classes from evidence, never message text", not "Failure classes are evidence-derived; message text lies about causes". Mechanism and rationale belong in Decision/Guidance/Why, not the title.
- Aphoristic, not vague: a title that names a topic instead of a rule wastes its digest slot. No colon-separated enumerations.
- The length check is the warning-tier const `MaxTitleRunes` in `template.go`, equal to the digest's truncation cap so a compliant title is never clipped; it is not configurable.
- Retitling a record never renames its `NNNN-slug.md` file — filenames are load-bearing (record pointers, `applies_to` globs on other records, index links). Retitle opportunistically when a record is next revised; never retroactively in bulk.

## Why

The digest renders roughly 12 bytes of overhead per record; the title is the rest of the line, so title bytes are the brief's real estate. An imperative title also collides head-on with the action an agent is about to take, which is what makes a one-line tripwire work.
