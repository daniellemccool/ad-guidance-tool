---
status: accepted
date: "2026-05-18"
comments:
    - author: Danielle McCool
      date: "2026-05-18 18:09:03"
      text: marked decision as decided
---

# Store comments in frontmatter and regenerate the body section on save

## Context and Problem Statement

The upstream tool's comment system had a data-loss bug (informally §A.1): the frontmatter stored a placeholder index in `Comment.Comment` ("1", "2", ...) while the real prose lived in the body under `<a name="comment-N"></a>` anchor blocks. Any tool operation that rewrote the body — including `adg edit`, `adg decide`, and `adg revise` — wiped the prose, leaving the placeholders alone in frontmatter. Users discovered this on every "I edited my ADR and my comments turned into numbers" surprise.

## Decision Drivers

* Eliminate the data-loss class permanently, not patch one path at a time
* Comment text should round-trip cleanly through `parse → render → parse`
* The body's `## Comments` H2 should be machine-regenerable so hand-edits are not silently lost in confusing ways
* The architectural anchor of the fork's domain refactor — the bug whose existence motivated the rewrite

## Considered Options

* Keep comments inline in the body with stricter anchor parsing; teach every body-rewrite path to preserve them
* Move comments to frontmatter as `{author, date, text}` entries; regenerate the body's `## Comments` H2 from frontmatter on every save
* Store comments in a sidecar file (one comments.yaml per ADR or one per model)

## Decision Outcome

Chosen option: "Move comments to frontmatter as `{author, date, text}` entries; regenerate the body's `## Comments` H2 from frontmatter on every save", because eliminates the bug by construction; the renderer strips and regenerates the section on every save so no rewrite path can drop the text.

## Comments

* **2026-05-18 18:09:03 — @Danielle McCool:** marked decision as decided
