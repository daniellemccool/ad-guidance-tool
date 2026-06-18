# ADR form factor

The target is a **small, decision-focused record** — typically 30–60 lines. An ADR captures
*what* was decided and *why*, plus the alternatives weighed. It is not a design doc, a
tutorial, or a changelog. If a section reads like documentation, it belongs in docs, not here.

## Sections

| Section | Required | Content |
|---|---|---|
| `# <title>` (H1) | by save | Encodes the decision. Optional on input — adg prepends it from the stored title. |
| `## Context and Problem Statement` | **yes** | 2–3 sentences: the situation and the question being decided. End on the question. |
| `## Considered Options` | **yes** | A bullet per option — just the titles. Real options, not strawmen. |
| `## Decision Outcome` | **yes** | Leave as `{...}` when authoring; `adg decide` writes `Chosen option: "X", because Y.` |
| `## Decision Drivers` | optional | The forces/constraints that shaped the call. A few bullets, not a rubric. |
| `### Consequences` | optional | Nested under Decision Outcome. Use for load-bearing follow-on facts. |

Omit by default: a full `## Pros and Cons of the Options` matrix, a `### Confirmation`
section, and a `## More Information` dump. Add per-option pros/cons only when the trade-off
is genuinely close and the reasoning needs to be preserved.

## Where load-bearing detail goes

If a decision carries operational detail that must survive (an ordering constraint, a
formula, a default value), put it in `### Consequences` under the outcome — **not** loose in
Decision Outcome. `adg decide` rewrites the outcome line; content in `### Consequences`
survives, and the guard refuses to clobber it without `--force`.

## What the validator enforces (and what it leaves to you)

The validator is the floor, not the whole bar. On an **accepted** ADR it hard-fails:

- **Leftover scaffolding** — any literal template token (`{...}`, `{driver 1}`,
  `{option 1}`, `{option 2}`, `{option title}`, `{justification}`). A token means a section
  was never written.
- **Empty required sections** — Context with no prose, etc.

It does **not** judge concision, option quality, or whether the rationale is sound. Those are
yours. Proposed stubs are exempt from both checks — scaffolding is expected while drafting;
it just must be gone before the ADR is accepted.

## Title discipline

There is no summary field, and `adg list` shows only titles. A good title states the WHAT
and usually the WHY in ~10–15 words, as a statement, not a question:

- Good: `Use a bounded VecDeque for subprocess stdout capture to cap peak memory`
- Weak: `stdout handling` / `How should we capture stdout?`
