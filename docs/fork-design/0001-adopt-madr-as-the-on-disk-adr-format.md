---
status: accepted
date: "2026-05-18"
comments:
    - author: Danielle McCool
      date: "2026-05-18 18:07:55"
      text: marked decision as decided
---

# Adopt MADR as the on-disk ADR format

## Context and Problem Statement

The upstream `adg` tool uses a custom Markdown layout: section headers wrapped in `<a name="..."></a>` HTML anchors, a sidecar `index.yaml` duplicating ADR metadata, and frontmatter fields like `adr_id` that mirror what the filename already encodes. The format is not interoperable with other tools in the ADR ecosystem (adr-tools, MADR-aware static-site generators, IDE plugins), and the anchor-based body shape makes ordinary Markdown editors mangle the file.

## Decision Drivers

* Interoperability with the broader MADR ecosystem
* Round-trip stability under `parse → render` (no silent reflows)
* MADR's principle of preferring body content over metadata fields
* Existing upstream user base needs a migration path, not a flag day

## Considered Options

* Keep the upstream format and add downstream converters at use sites
* Adopt MADR 4.0 as the only on-disk format; provide `adg migrate` for legacy files
* Support both formats with a `--format` flag, dispatching at read time

## Decision Outcome

Chosen option: "Adopt MADR 4.0 as the only on-disk format; provide `adg migrate` for legacy files", because single format means one round-trip property to maintain and one parser to test; the migrate command bridges legacy users without forcing them to swap tools.

## Comments

* **2026-05-18 18:07:55 — @Danielle McCool:** marked decision as decided
