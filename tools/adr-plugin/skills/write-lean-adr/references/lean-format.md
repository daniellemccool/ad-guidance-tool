# Lean ADR format

A parallel, agent-first ADR format (alongside the MADR format handled by the write-madr-adr skill). A lean ADR is a
**one-screen record** whose job is to answer, before an edit: *what rule governs this file, and how do
I know if I've violated it?* Routing metadata in the frontmatter lets `adg` compile a per-file brief
(and a PreToolUse hook) deterministically — no LLM. Lean ADRs are hand-authored markdown; `adg lean index`
validates them and regenerates the grouped README.

## Body

| Section | Required | Content |
|---|---|---|
| `# <title>` (H1) | yes | The decision as a statement. |
| `## Decision` | yes | 1–3 sentences: what was decided. |
| `## Guidance` | yes | What the next contributor must do — what new code must do, what review rejects, the fix path. (`## Implication` is an accepted alias.) |
| `## Why` | optional | Rationale. Expected for invariants — it is what lets an agent reason about an override instead of silently "simplifying" the rule. |
| `## Checks` | optional | Concrete things to confirm (grep targets, invariants). Rolled up into the brief's "Checks to run". |
| `## Context` / `## Alternatives` | optional | Only when load-bearing. |

If it runs past ~one screen (`MaxBodyLines`), it is probably two ADRs (the validator warns).

## Identity

**Flat global `NNNN`** across the whole model — the number is unique model-wide; files are
`NNNN-slug.md`. Index grouping comes from the `category` frontmatter, **not** the directory layout.
(The legacy per-subfolder `AD####` scheme collided once a number repeated across folders.) Duplicate
IDs hard-fail `adg lean index`.

## Frontmatter

| Key | Type | Values / form | Purpose |
|---|---|---|---|
| `status` | string | `proposed` \| `accepted` \| `rejected` \| `deprecated` \| `superseded by ADR-NNNN` \| `amended by ADR-NNNN` | Lifecycle. Accepted records are validated as finished. |
| `category` | string | free text | Index grouping. |
| `priority` | string | `invariant` \| `default` (or unset) | Force in the brief: invariants are hard constraints, surfaced with their `Why`. |
| `applies_to` | []glob | repo-root-relative | Routes the ADR to changed files. |
| `excludes` | []glob | repo-root-relative | Carves paths out of `applies_to`: a path is governed iff some `applies_to` matches **and** no `excludes` does. Use for a rule's sanctioned home or out-of-scope subpaths. |
| `forbids` | []glob | repo-root-relative | Negative-space scope — paths that should not exist. Routes the brief as a **violation** when matched, is exempt from the stale lint, and warns when it matches a real file. |
| `companions` | []glob | repo-root-relative | Expected partner edits this ADR does **not** govern. Surfaced as "related files"; never routes. |
| `source` | string | free text | Provenance (a durable deliberation artifact, not a branch name). |
| `supersedes` / `amends` | []NNNN | | Relationship integrity (forward + reverse links checked). |
| `tags` | []string | | Free metadata. |
| `date` | string | `YYYY-MM-DD` | |

All routing keys are `omitempty` — omit what you do not use.

## Glob rules

A zero-dependency doublestar matcher (`glob.go`); forward slashes, repo-root-relative:

- `**` — zero or more path segments (`port/**/*.py` matches both `port/x.py` and `port/a/b/x.py`).
- `*` — any run within one segment (does not cross `/`).
- `?` — one non-`/` character.
- **Brace expansion `{a,b}` is NOT supported** — it hard-fails validation. Write separate globs, one
  per alternative (`port/extraction/**`, `port/flows/**`, …).
- **Single-star under a nestable directory** (`platforms/*.py`) matches only one level — the validator
  warns and suggests the recursive form (`platforms/**/*.py`).

## What enforces what

Routing is **advisory**; enforcement is the index/lint/checks layer. "No brief appeared" never means
"no rule applies." The brief, the hook gate, and the index overlap check all route through one kernel
(`route.go`), so they cannot disagree about what a rule governs.

| Surface | Mode | Enforces |
|---|---|---|
| `adg lean brief --hook` (PreToolUse) | advisory, **fail-open** | Injects the matching ADRs for the edited file. Fires only on Claude `Edit`/`Write`/`MultiEdit`; misses shell/formatter/human/other-agent edits. |
| `adg lean brief <paths>` | advisory, **fail-closed on a bad model** | Compiles the brief; also runs validation and prints issues (e.g. a brace glob) to stderr, exiting non-zero on a hard failure. |
| `adg lean index` | **hard gate** | Duplicate ID and brace glob (hard); glob-hygiene, over-length body, missing category (warn); status vocabulary and supersede/amend integrity. |
| `adg lean index --root <tree>` | **hard gate + scope lint** | All of the above, plus: stale `applies_to`/`excludes` (match nothing), `forbids` that now matches a file, and default-vs-default scope overlap (computed on `applies_to` minus `excludes`). |
| `## Checks` | manual / future `check --diff` | Prose today, rolled into the brief; executable checks are a separate, later bet. |
