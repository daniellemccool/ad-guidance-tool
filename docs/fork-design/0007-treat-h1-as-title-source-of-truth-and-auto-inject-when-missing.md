---
status: accepted
date: "2026-05-18"
---

# Treat H1 as title source-of-truth and auto-inject it from existing title when missing

## Context and Problem Statement

ADR 0001 adopted MADR 4.0; MADR requires an H1 (`# {short title}`) at the start of the body, and `madr/types.go` documents the title as derived from the H1 (not from frontmatter). `adg edit --from-stdin` (PR 3) accepts whole-body markdown replacements, and `ReplaceBody` already overwrites `d.Title` from the body's H1 when present.

But until now, when a submitted body had no H1, `ReplaceBody` saved it verbatim. The validator on the next load then reported `"H1 title is missing or empty"` because `LoadById` derives `d.Title` from the parsed H1. Authors had to retype the title — once in the `adg add --title` step (which sets up the frontmatter and initial body) and again as the H1 in the replacement body — and a stale copy was easy to introduce.

The friction is real for LLM-driven authoring: the assistant has the body content, knows the decision's ID, and shouldn't have to maintain a separate H1 string that must match the original `--title` byte-for-byte to avoid a slug rewrite + file rename it didn't intend.

## Decision Drivers

* H1-as-title is already the documented source-of-truth — round-trips through Parse/Load/Save assume it
* LLM authoring should not require restating context the tool already carries
* If an author *does* supply an H1, it must keep behaving as the authoritative title (renames the file, updates the slug)
* The fix must not silently lose information when the author omits the H1 by mistake — the existing title is the obvious fallback
* No new flag or escape hatch: this is plumbing, not policy

## Considered Options

* Reject bodies with no H1; force authors to include one
* Auto-inject `# <d.Title>\n\n` at the top of the body when no H1 is present
* Store the title in frontmatter as the new source-of-truth, deprecate H1-as-title

## Decision Outcome

Chosen option: "Auto-inject `# <d.Title>\n\n` at the top of the body when no H1 is present", because the H1-as-title contract is already wired through Parse/Load/Save and the validator, and reversing it would ripple through every component; rejecting H1-less bodies adds friction without giving authors anything in return; auto-injection preserves the contract while removing the duplicate-restatement burden.

### Consequences

* Good: Authors can submit bodies that start at `## Context...` without restating the title; one source of truth at author-time, one source of truth on disk.
* Good: If the body *does* carry an H1, behavior is unchanged — the body's H1 is the title and triggers a slug rewrite if it differs from the decision's stored title.
* Bad: A typo'd or empty `--title` at `adg add` time still survives into the H1 if the replacement body has no H1. (The same is true today; this change neither helps nor hurts.)
* Neutral: The injection is silent — no stderr note. This matches the canonical `RenderNewBody` behavior, which also writes the H1 invisibly during `adg add`.
