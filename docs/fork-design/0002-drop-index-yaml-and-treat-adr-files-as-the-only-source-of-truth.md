---
status: accepted
date: "2026-05-18"
comments:
    - author: Danielle McCool
      date: "2026-05-18 18:08:33"
      text: marked decision as decided
---

# Drop index.yaml and treat ADR files as the only source of truth

## Context and Problem Statement

The upstream tool maintained an `index.yaml` sidecar that duplicated each ADR's frontmatter for fast lookup. It served as a read cache for `adg list` and `adg validate`, with `adg rebuild` to refresh it. The cache drifted from the files whenever an edit was made outside the tool, producing false validation failures and confusing the user about which artifact to trust.

## Decision Drivers

* Single source of truth: no drift class of bugs
* Simplicity: one fewer artifact in the repository
* Performance: at typical model sizes (≤ a few hundred ADRs) a directory scan is fast enough
* Tooling: any IDE/editor that touches an ADR shouldn't have to know about the sidecar

## Considered Options

* Keep `index.yaml` as a write-through cache, fail loudly on drift
* Drop `index.yaml`; scan files on every read
* Generate `index.yaml` lazily on first read but never trust it for writes

## Decision Outcome

Chosen option: "Drop `index.yaml`; scan files on every read", because removes an entire bug class (drift) and a whole command (rebuild) at the cost of a directory walk that is below human perception at our model sizes.

## Comments

* **2026-05-18 18:08:33 — @Danielle McCool:** marked decision as decided
