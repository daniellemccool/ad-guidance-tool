# Lean ADR authoring rubric

A lean ADR is optimized for the **compiled brief**, not for historical completeness.
The question every record must answer crisply is:

> Before editing this file, what rule do I need in my head — and how do I know if I've
> violated it?

If that answer isn't crisp, the ADR is doing too much. This rubric is the standard the
`write-lean-adr` skill authors to; `adg lean review` (when present) judges against it; and
its mechanically-checkable parts are enforced as advisory warnings by `adg lean index`
(see "What the tool checks" below).

> Provenance: distilled from `~/notes/lean-adr-brief-rewrite-guidance.md`. That note is the
> origin; **this file is the source of truth** — edit here, not there.

## Section shapes

- **Decision** — the rule, in **one to three sentences of prose**. Nothing else.
- **Guidance** — reviewable consequences as bullets: what new code must do, what review
  rejects, where the fix belongs.
- **Why** — the reasoning: why the rule exists / what it protects, so a reader can generalize.
  Required on every finished (accepted) record; keep it to the reason in one line, not a design
  essay. Record-only — never rendered into a brief.
- **Checks** — only concrete grep/test/manual checks.

## Authoring discipline

1. **Decision is the rule in 1–3 sentences — no lists.** A Decision that contains a bullet
   list or a per-case/per-file enumeration is a smell: the enumeration belongs in Guidance.
   *(Lint: a list in Decision, or a Decision over ~60 words, warns.)*
2. **Model a prohibition as a `forbids`-only record.** If the essence is negative space
   ("don't build X"), express it with `forbids` alone and a prohibition title — do **not**
   also bolt an `applies_to` describing the positive mechanism onto it. The positive
   mechanism lives in the mechanism ADR(s); duplicating it here creates overlap and drift.
3. **Scope `applies_to` to the enforcement points, not the neighborhood.** Match the few
   files that implement or could violate the rule; prefer naming them over a broad `**`.
   Over-broad scope inflates every brief it touches and manufactures overlap. (A boundary
   that lives in one file should not route into every sibling.)
4. **Every finished record carries its `Why`.** State why the rule exists in one line — do not
   omit it as "padding" and do not pad it into an essay. For an invariant, the `Why` is what
   breaks if the rule is weakened. Required once the record is `accepted`, and record-only (it
   is never injected into a brief, so it costs nothing at edit time).
5. **Re-judge priority; don't inherit it.** `invariant` = a rule an agent must never
   silently simplify or breach — a *defect* if violated, not a convention. Mark a genuine
   hard constraint `invariant` even if its source record treated it softly.
6. **Name the mechanism or file, not the ADR number.** Write "route donate payloads through
   `handle_donate_result()`", not "see ADR-0012". Routing surfaces the related rule, and
   numbers churn on renumber — a body that cites numbers rots.
7. **Titles are crisp statements or prohibitions, not enumerations.** "No parallel
   extraction architecture" — not "Single extraction: FlowBuilder + curated extraction +
   DDP_CATEGORIES". A colon-separated list of mechanisms in the title means the ADR is
   doing too much.
8. **State each fact once.** Decision = the rule, Guidance = what to do, Why = the reasoning
   (required). Don't repeat the same enumeration across Decision and Guidance.
9. **For a behavioral rule, point at its test(s).** List the test that exercises the rule as
   a `companion` when it's an expected partner edit; put it in `applies_to` instead when a
   test *is* the rule's executable enforcement (editing the test means editing the rule).
   Abstract structural rules with no one-to-one test skip this.
10. **Add a check only when it earns its place.** A Check is a concrete grep/verification
    target *not already implied by Guidance* — "grep for `CommandUIRender(` outside
    `port_helpers.py`," not "confirm safety runs first" (which restates a Guidance bullet).
    No checks is the common, correct case. If a check is automatable, make it a **frontmatter
    `checks` grep assertion** (`adg lean check` runs it); keep only non-automatable checks as
    prose in `## Checks`.

## When is it an invariant?

Mark `invariant` when violation crosses a boundary that must stay intact even if a local
implementation seems to work:

- PII / privacy boundaries; consent boundaries
- security or data-minimization constraints
- protocol/compatibility with an external engine or host
- a single source of truth needed to prevent divergent behavior
- a safety check that must precede risky processing
- any rule where tests can pass while the system becomes unsafe

Everything else is a `default` (a convention).

## The `Why` (required on every accepted record)

A `## Why` is required on every accepted record — it is a section co-equal with Decision and
Guidance, and `adg lean index` hard-fails an accepted record without one. It answers **"why
does this rule exist / what breaks without it?"** — not "why is this a nice design?". For an
invariant, center it on what breaks if the rule is breached or weakened. Keep it to the reason,
one line. It is **record-only**: it lives in the record for a human (or an LLM that deliberately
opens the file) and is never rendered into a brief, so it costs nothing at injection time.

## When to combine vs split

Combine two ADRs only when they always route together **and** an editor cannot reasonably
obey one without knowing the other. Keep them separate when they have different failure
modes, different checks, can change independently, or one is an invariant and the other a
convention.

## The compact-mode test (do this before you finish)

A compact brief renders **only the first Guidance bullet** of a default. So:

> If only the first Guidance bullet appeared, would it still steer the edit correctly?

If not, rewrite the first bullet — it is now load-bearing.

## What the tool checks (deterministic)

`adg lean index` **hard-fails** an accepted record that is missing a required section —
Decision, Guidance, or `## Why` (a heading alone doesn't count). It also **warns** (never
fails) on the mechanical leanness subset of this rubric, skipped on terminal records and on
unfilled scaffold placeholders:

- Decision contains a list, or runs over `MaxDecisionWords` (~60).
- Guidance has no list item (lead with reviewable bullets).

The judgment-level rules (2, 3, 5, 7, 8 and "is the first bullet load-bearing?") are not
mechanically checkable — they are what a human reviewer or `adg lean review` weighs.
