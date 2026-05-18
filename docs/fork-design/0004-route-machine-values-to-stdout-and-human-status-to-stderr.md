---
status: accepted
date: "2026-05-18"
comments:
    - author: Danielle McCool
      date: "2026-05-18 18:09:22"
      text: marked decision as decided
---

# Route machine values to stdout and human status to stderr

## Context and Problem Statement

The upstream tool wrote all output to stdout, mixing machine-readable identifiers with English status text. Capturing the ID of a newly-added decision in a shell script required parsing prose — `awk` against "Decision X (0001) added successfully." — which is brittle, untranslatable, and impossible to compose with `xargs`. Validation errors and success messages also shared the same stream, so `adg validate || handler` couldn't reliably detect failure without parsing exit codes that didn't differentiate.

## Decision Drivers

* UNIX CLI convention: stdout is data, stderr is diagnostics, exit code is the failure signal
* Composability: `$(adg ...)` should yield clean machine values
* Errors must always be visible — even when the user silences status
* Validation failures need a non-zero exit code so CI scripts can gate on them

## Considered Options

* Keep upstream behavior: everything on stdout
* Move machine values to stdout; status to stderr; errors to stderr regardless of `--quiet`; non-zero exit on validation issues
* Add a `--format=json` mode and emit only structured output

## Decision Outcome

Chosen option: "Move machine values to stdout; status to stderr; errors to stderr regardless of `--quiet`; non-zero exit on validation issues", because matches every standard Unix tool and unlocks ID=$(adg add) without a JSON layer; the --quiet flag suppresses status without affecting machine values or errors.

## Comments

* **2026-05-18 18:09:22 — @Danielle McCool:** marked decision as decided
