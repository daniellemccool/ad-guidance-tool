package lean

// Budget governs the whole-body length discipline the validator applies to a lean
// record. It affects ONLY the one-screen "should fit one screen" nudge — never the
// agent-facing Decision-word or brief-line budgets, which stay constant so the
// injected brief remains low-token regardless of how a project sets this.
type Budget struct {
	// WarnWholeBody enables the whole-body one-screen length warning.
	WarnWholeBody bool
	// MaxBodyLines is the line threshold for that warning when WarnWholeBody is set.
	MaxBodyLines int
}

// DefaultBudget is the one-screen "lean" discipline: the whole-body warning is on at
// the package MaxBodyLines ceiling. It reproduces the behavior from before per-project
// budgets existed, so every default caller and existing test is unchanged.
func DefaultBudget() Budget {
	return Budget{WarnWholeBody: true, MaxBodyLines: MaxBodyLines}
}

// NarrativeBudget opts a project out of the one-screen discipline: the whole-body
// warning is suppressed so the record-only Why/Context/Alternatives may run to a full,
// traditional-ADR length. Decision-word and brief-line budgets are unaffected.
func NarrativeBudget() Budget {
	return Budget{WarnWholeBody: false, MaxBodyLines: MaxBodyLines}
}
