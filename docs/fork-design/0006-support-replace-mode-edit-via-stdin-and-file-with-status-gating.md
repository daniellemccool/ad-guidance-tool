---
status: accepted
date: "2026-05-18"
comments:
    - author: Danielle McCool
      date: "2026-05-18 18:09:57"
      text: marked decision as decided
---

# Support replace-mode edit via stdin and file with status gating

## Context and Problem Statement

LLM-driven ADR authoring is a real workflow: an assistant drafts a complete MADR body, and the user wants to install it as-is. The upstream `adg edit` only supported per-section append flags (`--question`, `--criteria`, ...). Replacing an entire body required hand-copying section by section. At the same time, decided/accepted/superseded ADRs are part of the historical record — replacing their content should be deliberate, not the default. And a malformed body (missing required sections) should be rejected before write, not produce a file that breaks `adg decide` or `adg validate` later.

## Decision Drivers

* LLM-generated full bodies are a first-class authoring workflow
* Historical record (status != proposed) deserves a guard against accidental overwrite
* Input validation should fail-fast at the write boundary, not later at parse time
* The body's H1 is authoritative — if the input renames the decision, the file should rename too

## Considered Options

* Keep append-only; document a "delete file then `adg add` again" workflow for wholesale rewrites
* Add `--from-stdin`/`--from-file` flags that overwrite; always allowed regardless of status
* Add replace mode with status gating: `proposed` is free, anything else requires `--force`; refuse input that doesn't parse as MADR with the three required sections

## Decision Outcome

Chosen option: "Add replace mode with status gating: `proposed` is free, anything else requires `--force`; refuse input that doesn't parse as MADR with the three required sections", because status gating preserves the audit trail by default while --force keeps it a one-flag escape; pre-parse validation prevents writing bodies that downstream commands can't reason about.

## Comments

* **2026-05-18 18:09:57 — @Danielle McCool:** marked decision as decided
