---
name: write-lean-adr
description: >
  Work with lean Architectural Decision Records — compact Decision/Guidance records
  plus routing frontmatter (applies_to / excludes / forbids / companions, priority)
  that `adg` compiles into per-file briefs and a PreToolUse hook. Covers both
  authoring lean records and using them as active governance: run `adg lean brief` before
  an edit, respect `forbids`, treat `companions` as non-governing related files, and
  validate/route with `adg lean index`. Use in repos whose ADRs are lean (Decision +
  Guidance with routing frontmatter), not full MADR. For durable MADR records with
  Context / Considered Options / Decision Outcome and the decide / supersede / revise
  loop, use the write-madr-adr skill instead.
---

# write-lean-adr — author and consume lean decisions with `adg`

A lean ADR is a one-screen record that answers, before an edit: *what rule governs
this file, and how do I know if I've violated it?* The skill has two jobs:

1. **Consume** — before changing code, compile the brief for the paths you're
   touching so the governing ADRs are in context (the PreToolUse hook automates this).
2. **Author** — write compact `Decision` + `Guidance` records with routing
   frontmatter, validated by `adg lean index`.

Full schema (body, frontmatter, glob rules, what each command enforces):
`references/lean-format.md`.

**Which model:** operate on the repo's *active* lean model — `docs/decisions/` unless it
configures another (`adg set-config`). Other ADR-like files — a historical MADR record under
`docs/fork-design/`, seed or template models — are **not** your model; don't route, validate, or
author against them. Pick this skill vs write-madr-adr by the active model's format (lean
Decision/Guidance + routing frontmatter), not by every `NNNN-*.md` in the tree.

## Consume: governance before an edit

```bash
# Which rules govern these paths? Deterministic, no LLM.
adg lean brief --model docs/decisions path/to/file.py another/file.py
```

- **applies_to** routes an ADR to the files it governs; the brief lists invariants
  (hard constraints) before defaults.
- **excludes** carves a sanctioned or out-of-scope path out of a broad `applies_to` —
  the rule deliberately does not fire there.
- **forbids** marks negative space. A brief showing **"⚠ Forbidden scope matched"**
  means the edit touches a path the ADR says must not exist — treat it as a violation
  to stop and reconsider, not as guidance to follow.
- **companions** are expected partner edits the ADR does **not** govern (e.g. the TS
  side of a prop). They appear as "related files to consider," never as a rule — make
  the companion change, but don't treat the ADR as governing it.
- **Routing is advisory.** "No brief appeared" never means "no rule applies": the hook
  fires only on Claude edits and is fail-open. Real enforcement is `adg lean index` / CI /
  review.

### Pre-edit hook (recommended)

`adg lean brief --hook` implements the Claude Code PreToolUse contract — it injects the
brief for the file being edited as `additionalContext`, fail-open. Setup and the
contract are documented with the tool (`docs/lean-prototype/hook-setup.md`).

## Author: a lean record

Lean records are hand-authored markdown (there is no `adg add` for the lean form) at
`docs/decisions/NNNN-slug.md`:

```markdown
---
status: accepted          # proposed | accepted | rejected | deprecated | superseded by ADR-NNNN | amended by ADR-NNNN
category: Architecture    # groups the generated index (not the directory layout)
priority: invariant       # invariant | default — force in the brief
applies_to:
    - port/**/*.py
excludes:
    - "**/port_helpers.py"
---

# <the decision, as a statement>

## Decision

One to three sentences: what was decided.

## Guidance

- What new code must do, what review rejects, the fix path.

## Why            # optional; expected for invariants — the rationale that lets an agent reason about an override
## Checks         # optional; grep targets / invariants, rolled up into the brief
```

- **IDs are a flat global `NNNN`** across the whole model; `category` (not a
  subfolder) groups the index. Duplicate IDs hard-fail.
- **Globs** are forward-slash, repo-root-relative, doublestar (`**`). **Brace globs
  `{a,b}` are rejected** — write separate globs, one per alternative. A single-star
  segment under a nestable directory (`platforms/*.py`) is flagged; prefer
  `platforms/**/*.py`.
- Keep it to one screen. If it runs longer, it is probably two ADRs.

## Validate and index

```bash
adg lean index --model docs/decisions            # validate + print the grouped README
adg lean index --model docs/decisions --write    # write docs/decisions/README.md
adg lean index --model docs/decisions --root .   # also scope-lint globs against the tree
```

`adg lean index` hard-fails on a duplicate ID or a brace glob and warns on glob hygiene,
over-length bodies, and (with `--root`) stale `applies_to`/`excludes`, `forbids` that
now match a file, and default-vs-default scope overlap. Wire `adg lean index --root` into
CI for real enforcement — the hook only routes; the index gates.

## Reference files

- `references/lean-format.md` — the lean format spec: body, frontmatter, glob rules,
  and the "what enforces what" matrix.
