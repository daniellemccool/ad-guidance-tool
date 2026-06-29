# Lean ADR review rubric

Judge a lean ADR against these rules. A lean ADR is a one-screen record optimized for the
compiled brief — it must answer, before an edit: *what rule governs this file, and how do I
know if I've violated it?*

1. Decision is the rule in 1–3 sentences of prose — no lists, no per-case enumeration (that
   belongs in Guidance). A list in Decision is a violation.
2. A pure prohibition is expressed with `forbids` alone and a prohibition title — not an
   `applies_to` re-describing the positive mechanism (which belongs in the mechanism ADR).
3. `applies_to` names the few files that enforce the rule, not the whole neighborhood; an
   over-broad `**` that inflates briefs and overlap is a violation.
4. `Why` is reserved for invariants (and the rare load-bearing default). A default with a
   `Why` that only explains "why it's nice" is padding.
5. Priority is re-judged: `invariant` means a rule that must never be silently simplified or
   breached (a defect if violated), not a convention. Flag a genuine hard constraint marked
   `default`, or a mere convention marked `invariant`.
6. The body names the mechanism/file, not sibling ADR numbers (numbers churn on renumber).
7. The title is a crisp statement or prohibition, not a colon-separated list of mechanisms.
8. Each fact is stated once: Decision = the rule, Guidance = what to do, Why = the rationale.
   No repeated enumeration across sections.
9. A behavioral rule points at its test(s) — as `companions` (partner edits) or in
   `applies_to` (when a test *is* the enforcement). Abstract structural rules need no test.
10. `## Checks` / `checks` are concrete grep/verification targets not already implied by
    Guidance — not a restatement of a Guidance bullet. No checks is often correct.

The compact-mode test: a compact brief renders only the FIRST Guidance bullet of a default.
If that first bullet would not steer the edit correctly on its own, the ADR fails this test.

Verdict: "revise" if any rule is clearly violated in a way a human reviewer would send back;
otherwise "pass". Be specific and actionable; do not invent problems where the record is fine.
