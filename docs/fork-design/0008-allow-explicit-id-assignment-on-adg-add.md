---
status: accepted
date: "2026-05-18"
---

# Allow explicit ID assignment on `adg add` with fail-fast collision

## Context and Problem Statement

`adg add` has historically been the single source of ID assignment: the repository scans the model directory for the highest existing NNNN and assigns max+1. That's the right default — humans creating ADRs interactively shouldn't have to think about IDs — but it makes plan-paper authoring fragile.

A plan-paper says "Task T1 will land AD0022, AD0023, AD0024." The executor reads the plan, calls `adg add`, and *hopes* the auto-assignment matches. In practice they have to wrap each call in defensive shell:

```bash
ID22=$(scripts/adr new "...")
[[ "$ID22" == "0022" ]] || { echo "expected 0022, got $ID22"; exit 1; }
```

This is paper-over-the-cracks. The plan is deterministic; the tool should let the executor commit to that determinism. Without it, any concurrent change to the model directory (a teammate landing a different ADR, a forgotten `proposed` from yesterday) silently shifts the assignment and the plan-paper's cross-references drift.

## Decision Drivers

* Plan-paper authoring needs deterministic ID assignment; the IDs are part of the plan's contract, not an output of the tool
* Auto-assignment should remain the default — interactive use shouldn't get harder
* Collisions must be loud: the worst outcome is silently shifting to the next free slot, because that defeats the whole point of asking for a specific ID
* The repository owns file-naming; ID validation belongs at that layer, not in the CLI
* Keep the wrapper simple — agents call `adr new <title> [<id>]`, no separate verb

## Considered Options

* Keep auto-only; document the "capture-and-verify" defensive shell as the canonical pattern
* Add `--id NNNN` with fail-fast on collision: explicit assignment errors loudly if the ID is taken
* Add `--id NNNN` with auto-skip on collision: warns to stderr and assigns next free
* Add `--id NNNN --force` to allow overwriting an existing ADR with the new content

## Decision Outcome

Chosen option: "Add `--id NNNN` with fail-fast on collision: explicit assignment errors loudly if the ID is taken", because the entire reason to specify an ID is that the caller has external commitments to it (a plan-paper, a cross-reference, a teammate's draft); auto-skip would silently break those commitments, and `--force` invites overwriting real work. Fail-fast is the only behavior consistent with "the caller knows what ID they want."

### Consequences

* Good: Plan-paper executors can write `adr new "Title" 0022` and trust the assignment or get a clear error — no more capture-and-verify.
* Good: Auto-assignment is unchanged; `adr new <title>` still picks max+1. No behavior change for interactive use.
* Good: Auto-assignment continues to pick from the highest existing ID, so pre-placing `0042` explicitly and then auto-creating yields `0043` (the gap from `0001` to `0041` is preserved as a deliberate choice by whoever set 0042).
* Neutral: The CLI accepts both `--id 22` (zero-pads to `0022`) and `--id 0022`. The repository requires the 4-digit form; normalization happens at the CLI boundary so service and infrastructure layers stay strict.
* Neutral: `0000` is rejected as reserved (auto-assignment starts at `0001`, so it would never appear naturally; allowing it would create a special case for no benefit).
* Bad: `--id` can only be paired with a single `--title`. Multiple titles with one `--id` is ambiguous (does it apply to the first? all? consecutively?); refusing the combination at parse time is the simplest rule.
