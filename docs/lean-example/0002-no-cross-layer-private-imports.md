---
status: accepted
date: "2026-03-13"
category: Python architecture
source: feat/facebook-ddp-error-handling (the violation that prompted it)
priority: default
applies_to:
    - "port/**/*.py"
    - "platforms/**/*.py"
tags:
    - imports
    - encapsulation
---

# No cross-layer private imports

## Decision

A private function (leading underscore) must not be imported across architectural layers. The
underscore is the author's contract that the function is internal; importing it from another layer
violates that contract and creates hidden coupling between layers that should be independent.

## Guidance

- Any `from port.helpers.flow_builder import _something` in `script.py` or `platforms/` is a review violation.
- The fix is always the same: if a function is genuinely needed across layers, rename it without the
  underscore and move it to the appropriate shared layer (`helpers/`) — a deliberate decision, not a reach-in.

## Why

A private cross-layer import couples layers that must evolve independently; once one exists the
boundary the layering exists to protect is silently gone, and every later import copies the mistake
until refactoring any one layer means untangling all of them.
