# adg direction: from ADR management to agent-operational governance

This note records where `adg` is heading and why. It summarizes an exploration that compared the
current MADR-based format against leaner styles (the Ferry `decisions/architecture/` corpus and the
d3i `data-donation-task` ADRs), prototyped a synthesis, and incorporated external review.

## The reframe

The governing idea:

> An ADR should not primarily answer **"what happened?"** It should answer **"what rule governs my
> next edit, and how do I know whether I've violated it?"**

Historically ADRs are decision *logs* — write-once, read for archaeology. In an agent-driven
codebase the high-value use is as *active constraints* consulted before editing. So `adg` should move
from "better MADR management" toward being a **compiler for architecture context**: it helps an agent
discover the rules relevant to a change, understand their force, and verify compliance.

## What we've done

### 1. Fixed three `adg` UX bugs (branch `fix/adg-cli-snags`)
- **`--id` short form** accepted everywhere — `edit`/`decide`/etc. take `1` and zero-pad to `0001`,
  matching `add` (centralized in one resolver).
- **`decide` preserves a pre-authored `### Consequences`** — the placeholder check and rewrite are
  scoped to the outcome prose above the first nested H3, aligning the code to the documented workflow.
  (`### Consequences` is the full MADR template's optional H3, so this is MADR-correct.)
- **Legacy detection enumerates the whole tree** — `validate`/`list` collect every legacy file,
  group by subfolder, and name the "multiple colliding sub-models" case instead of bailing on the first.

### 2. Format evaluation + lean prototype (branch `prototype/lean-adr`)
- A parallel `lean` package (`internal/domain/decision/lean/`): a `Decision` + `Implication`
  template, parser, validator, and a **generated grouped README index** (removes the hand-maintained
  index burden). MADR, `decide`, and existing tests are untouched.
- Added `source` / `category` / `amends` frontmatter (omitempty, MADR-safe).
- Converted real Ferry and d3i ADRs as samples (~halving size while keeping operational substance).
- Design calls: `source` stays **optional** (only worth it with a durable deliberation artifact, not a
  branch name); do **not** add `author` — use the existing `decision-makers` field plus git blame.

### 3. Agent-governance core: `applies_to` routing + `brief --paths`
- `applies_to` globs + `priority` (`invariant` | `default`) frontmatter.
- A deterministic **`brief --paths`** compiler (`internal/domain/decision/lean/brief.go`, demoed via
  `tools/leanbrief`): changed files → the ADRs that govern them, grouped by force, each with Decision +
  Guidance + a consolidated "Checks to run" roll-up + a pointer back to the full record. No LLM,
  CI-friendly. Zero-dep doublestar glob matcher in `glob.go`.

## The target model

- **Record shape:** small, one screen. Required = `Decision` + `Guidance` (canonical; `Implication`
  is an accepted alias). Optional = `Why` (expected for invariants — rationale is what lets an agent
  reason about overrides) and `Checks`. Drop mandatory MADR Context / Considered Options / Decision
  Drivers as the default.
- **Routing metadata:** `applies_to` globs are the flagship (mechanical, verifiable). `priority`
  separates hard invariants from defaults. Cross-cutting rules (not path-scoped) use a small
  routing-tag vocabulary rather than free-text "triggers".
- **The compiler is the product:** the compiled `brief` matters more than the raw format —
  `adg brief --paths` / `--task`, a generated index, and (later) `adg check --diff`.
- **Governance layer is the differentiator** over plain scoped-rules systems (Cursor rules,
  CODEOWNERS, editorconfig): supersession/amendment integrity, a validated status vocabulary, and
  validations such as stale-`applies_to` (globs matching nothing) and overlapping-scope conflict
  detection.

## Status & open decisions

- `brief --paths` proves the thesis; the index generator and lean validator work.
- **Naming resolved:** `Guidance` is canonical; `Implication` is an accepted alias (template, brief
  output, and validator standardize on Guidance).
- **Scope lint shipped:** `LintTree` (in `lean/lint.go`, wired via `leanindex --root`) reports stale
  `applies_to` globs and default-vs-default scope overlap.
- The brief is now authoritative for routine changes and only tells the agent to load full records
  when an invariant applies (or guidance is ambiguous / rules conflict); `Why` is surfaced for invariants.
- **Cobra wiring done:** `adg brief` (with `--hook`) and `adg index` (with `--root`/`--write`) are
  registered as *thin* commands over the lean package (`internal/adapter/command/lean`). Follow-up:
  promote them to the full inputport/interactor/presenter stack used by the MADR commands once lean
  graduates from prototype.
- **Pre-edit hook shipped:** `adg brief --hook` implements the Claude Code PreToolUse contract —
  injects the brief for the file being edited as `additionalContext`, fail-open. See `hook-setup.md`.
- **Open:** whether NL `Checks` should become *executable* (real `check --diff` gating — a bigger,
  separate bet); whether to retire the now-orthogonal `decide`/options machinery.

## Next steps (recommended order)

1. **Author `applies_to` on real ADRs** (e.g. the ~40 d3i records) so the hook routes live in a repo.
2. **Promote the lean commands** to the full interactor/presenter stack (tracked follow-up).
3. **Executable checks** — turn `## Checks` into runnable patterns so `check --diff` can gate (the
   larger investment).

Explicit non-goals for now: `triggers` / `brief --task` semantic routing, cobra wiring, and the
`Implication`→`Guidance` rename — none block proving the deterministic core first.
