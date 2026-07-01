package lean

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// BriefMode selects how Brief renders the defaults it routes to.
//
//   - BriefFull always renders every governing ADR in full (scope/matched detail,
//     Decision, full Guidance, inline companions) — the debuggable form.
//   - BriefCompact keeps invariants and forbidden-path hits full but collapses each
//     default to a one-line checklist item and aggregates companions once.
//   - BriefAuto renders full, then re-renders defaults compact if the brief exceeds
//     MaxBriefLines — so small briefs stay rich and hub files stay readable.
type BriefMode int

const (
	BriefAuto BriefMode = iota
	BriefFull
	BriefCompact
)

// maxCompactSummary caps the one-line summary on a compact default entry. Some
// Decisions collapse into a whole paragraph (layered-architecture rules especially);
// truncating keeps a hub brief's checklist scannable.
const maxCompactSummary = 200

// briefHit is one record that routed into the brief, with its routing result.
type briefHit struct {
	rec   Record
	route routeResult
}

// Brief compiles an agent-facing guidance packet for a set of changed paths: the
// ADRs whose applies_to globs match (via the routeMatch kernel in route.go),
// grouped by force (invariants first, then defaults), plus a consolidated list of
// any Checks. This is the deterministic core of "before I touch this code, what
// rules constrain me?": pure path routing, no LLM, suitable for a pre-edit hook or
// CI. mode controls how defaults render (see BriefMode). Brief renders; route.go
// decides.
func Brief(records []Record, changedPaths []string, mode BriefMode) string {
	var invariants, defaults []briefHit
	for _, r := range records {
		route := routeMatch(r, changedPaths)
		// A record enters the brief when it governs a changed path or when a
		// forbids glob is violated; companions alone never route.
		if len(route.matched) == 0 && len(route.forbidden) == 0 {
			continue
		}
		h := briefHit{rec: r, route: route}
		if strings.EqualFold(strings.TrimSpace(r.D.Priority), "invariant") {
			invariants = append(invariants, h)
		} else {
			defaults = append(defaults, h)
		}
	}
	byID := func(hs []briefHit) { sort.Slice(hs, func(i, j int) bool { return hs[i].rec.ID < hs[j].rec.ID }) }
	byID(invariants)
	byID(defaults)

	switch mode {
	case BriefFull:
		return renderBrief(invariants, defaults, false)
	case BriefCompact:
		return renderBrief(invariants, defaults, true)
	default: // BriefAuto: full unless it blows the one-screen budget.
		full := renderBrief(invariants, defaults, false)
		if briefLineCount(full) <= MaxBriefLines {
			return full
		}
		return renderBrief(invariants, defaults, true)
	}
}

// conventionPreamble frames the corpus briefs (SessionStart / SubagentStart) with the
// same "how to treat these" contract the golden-path CLAUDE.md convention carries, so
// the guidance reaches the model through the hook even when the convention text is not
// loaded. It ends with a blank line; the sections that follow attach directly.
const conventionPreamble = `# Architecture brief

These lean ADRs are the working agreements that govern this codebase. Consult them while you plan a change, not only when you edit:
- **Invariant** — a hard rule; open the linked record before planning a change that touches it.
- **Forbidden scope** — if a rule marks paths off-limits, stop and surface the conflict; do not build it.
- **Related files** — when a rule names companions, check whether they also need edits.
- **No rule shown never means no rule applies** — routing is advisory and fail-open.

Before you design a change, pull the file-scoped brief for the paths you expect to touch:
    adg lean brief <paths>

`

// BriefWhole compiles the whole-corpus brief: every in-force ADR, invariants in full
// and defaults condensed, with no path routing, no companions, and no footer. It is the
// SessionStart injection — the agent sees the full set of working agreements once per
// session, before the first prompt. Rendering reuses the same entry helpers as the
// routed Brief (ADR-0002: one renderer). Empty when no in-force ADRs exist.
func BriefWhole(records []Record) string {
	inv, def := corpusSplit(records)
	if len(inv)+len(def) == 0 {
		return ""
	}
	return renderCorpus(inv, def, true)
}

// BriefInvariants compiles the invariants-only brief: every in-force invariant ADR in
// full, no defaults, no companions, no footer. It is the SubagentStart (Plan) injection
// — the architect starts with the hard constraints in view. The SubagentStart payload
// carries no file paths, so this cannot be path-scoped; the invariants are the
// always-relevant floor. Empty when the model declares no invariants.
func BriefInvariants(records []Record) string {
	inv, _ := corpusSplit(records)
	if len(inv) == 0 {
		return ""
	}
	return renderCorpus(inv, nil, false)
}

// corpusSplit partitions the in-force records into invariants and defaults by priority,
// each sorted by ID. Terminal records are dropped via inForce — the same lifecycle gate
// routeMatch applies — so the corpus briefs never surface a retired rule. The briefHits
// carry a zero route: they are unrouted entries, rendered from their declared scope.
func corpusSplit(records []Record) (invariants, defaults []briefHit) {
	for _, r := range records {
		if !inForce(r.D.Status) {
			continue
		}
		h := briefHit{rec: r}
		if strings.EqualFold(strings.TrimSpace(r.D.Priority), "invariant") {
			invariants = append(invariants, h)
		} else {
			defaults = append(defaults, h)
		}
	}
	byID := func(hs []briefHit) { sort.Slice(hs, func(i, j int) bool { return hs[i].rec.ID < hs[j].rec.ID }) }
	byID(invariants)
	byID(defaults)
	return invariants, defaults
}

// renderCorpus renders a corpus brief: the convention preamble, invariants in full (no
// companions), and — when includeDefaults — defaults as a condensed checklist. No
// related-files section and no "Before you finish" footer: those are edit/commit-time
// concerns, not session-start framing.
func renderCorpus(invariants, defaults []briefHit, includeDefaults bool) string {
	var b strings.Builder
	b.WriteString(conventionPreamble)

	if len(invariants) > 0 {
		b.WriteString("## Hard constraints (invariants)\n\n")
		for _, h := range invariants {
			b.WriteString(briefEntry(h.rec, h.route, true, false))
		}
	}
	if includeDefaults && len(defaults) > 0 {
		b.WriteString("## Defaults & conventions (condensed — open the record for full guidance)\n\n")
		for _, h := range defaults {
			b.WriteString(compactDefaultLine(h.rec))
		}
	}
	return b.String()
}

// renderBrief writes the brief. With compactDefaults, each plain default collapses
// to a checklist line and companions are aggregated once at the end; invariants and
// forbidden-path hits always render in full.
func renderBrief(invariants, defaults []briefHit, compactDefaults bool) string {
	var b strings.Builder
	b.WriteString("# Architecture brief\n\n")

	total := len(invariants) + len(defaults)
	if total == 0 {
		b.WriteString("No ADRs match these paths. (Routing is by `applies_to` globs; an ADR with no globs never auto-routes.)\n")
		return b.String()
	}
	fmt.Fprintf(&b, "%d ADR(s) govern this change.\n", total)
	if len(invariants) > 0 {
		b.WriteString("An invariant applies: treat the invariant entries as hard constraints and load those records in full before planning.\n")
	} else {
		b.WriteString("This brief is authoritative for routine changes; load a full record only if its guidance is ambiguous or two rules conflict.\n")
	}

	if len(invariants) > 0 {
		b.WriteString("\n## Hard constraints (invariants)\n\n")
		for _, h := range invariants {
			b.WriteString(briefEntry(h.rec, h.route, true, !compactDefaults))
		}
	}
	if len(defaults) > 0 {
		if compactDefaults {
			b.WriteString("\n## Defaults & conventions (condensed — open the record for full guidance)\n\n")
		} else {
			b.WriteString("\n## Defaults & conventions\n\n")
		}
		for _, h := range defaults {
			// A forbids violation is always rendered full, even for a default: a
			// negative-space hit must stay prominent regardless of the line budget.
			if compactDefaults && len(h.route.forbidden) == 0 {
				b.WriteString(compactDefaultLine(h.rec))
			} else {
				b.WriteString(briefEntry(h.rec, h.route, false, !compactDefaults))
			}
		}
	}

	// In compact mode companions are pulled out of the per-entry rendering and
	// listed once here as context, rather than repeated inside every entry.
	if compactDefaults {
		b.WriteString(relatedFilesSection(invariants, defaults))
	}

	// The post-edit half of the contract: what to run before finishing. Always
	// present when the brief routed at all, so the hot path closes the loop.
	b.WriteString(beforeYouFinish(append(append([]briefHit{}, invariants...), defaults...)))
	return b.String()
}

// beforeYouFinish renders the post-edit action footer: re-run the model gate, the
// matched ADRs' Checks, and the test files those ADRs name (via companions or
// applies_to). It always tells the agent to re-run the index; the Checks and tests
// lines appear only when the matched ADRs supply them.
func beforeYouFinish(hits []briefHit) string {
	var checks []string
	testSet := map[string]bool{}
	execChecks := 0
	for _, h := range hits {
		execChecks += len(h.rec.D.Checks)
		if c, ok := ParseBody(h.rec.Body).Sections["checks"]; ok {
			for _, line := range bulletLines(c) {
				checks = append(checks, fmt.Sprintf("(ADR-%s) %s", h.rec.ID, line))
			}
		}
		for _, pat := range append(append([]string{}, h.rec.D.AppliesTo...), h.rec.D.Companions...) {
			if looksLikeTest(pat) {
				testSet[pat] = true
			}
		}
	}
	tests := make([]string, 0, len(testSet))
	for t := range testSet {
		tests = append(tests, t)
	}
	sort.Strings(tests)

	var b strings.Builder
	b.WriteString("\n## Before you finish\n\n")
	b.WriteString("- Re-run `adg lean index --root .` — this validates the model and scope routing, not that the code obeys the prose guidance above.\n")
	if execChecks > 0 {
		fmt.Fprintf(&b, "- Run `adg lean check` — %d executable check(s) on the matched ADRs.\n", execChecks)
	}
	if len(checks) > 0 {
		b.WriteString("- Review the matched ADR checks:\n")
		for _, c := range checks {
			fmt.Fprintf(&b, "    - %s\n", c)
		}
	}
	if len(tests) > 0 {
		b.WriteString("- Run the tests these ADRs name:\n")
		for _, t := range tests {
			fmt.Fprintf(&b, "    - %s\n", t)
		}
	}
	return b.String()
}

// testPathRe matches a path/glob that looks like a test: a tests/ or test/ segment,
// a test_ prefix, or a _test. / .test. / .spec. infix (Python, Go, JS conventions).
var testPathRe = regexp.MustCompile(`(?i)(^|/)tests?/|(^|/)test_|_test\.|\.test\.|\.spec\.`)

func looksLikeTest(pat string) bool { return testPathRe.MatchString(pat) }

// briefEntry renders one full entry. includeWhy surfaces the Why (invariants only);
// includeCompanions inlines the companion list (full mode) — compact mode passes
// false and aggregates companions via relatedFilesSection instead.
func briefEntry(r Record, route routeResult, includeWhy, includeCompanions bool) string {
	p := ParseBody(r.Body)
	guidance := guidanceSection(p)

	var b strings.Builder
	fmt.Fprintf(&b, "### ADR-%s — %s\n", r.ID, p.Title)

	// Scope line. A governed record shows which applies_to globs matched (plus any
	// excludes that suppressed a sibling path). A forbids-only hit has no matched
	// glob, so it shows the forbids scope rather than an empty `matched:`.
	if len(route.matched) > 0 {
		fmt.Fprintf(&b, "_scope: %s · matched: %s", strings.Join(r.D.AppliesTo, ", "), strings.Join(route.matched, ", "))
		if len(route.excluded) > 0 {
			fmt.Fprintf(&b, " · excluded: %s", strings.Join(route.excluded, ", "))
		}
		b.WriteString("_\n\n")
	} else if len(route.forbidden) > 0 {
		fmt.Fprintf(&b, "_scope: forbids %s_\n\n", strings.Join(route.forbidden, ", "))
	} else if len(r.D.AppliesTo) > 0 {
		// Unrouted corpus entry (no changed path drove it in): show the declared scope.
		fmt.Fprintf(&b, "_scope: %s_\n\n", strings.Join(r.D.AppliesTo, ", "))
	} else {
		b.WriteString("_scope: (no applies_to globs — see the record)_\n\n")
	}

	// A forbids hit is a violation — the change touches negative-space the ADR marks
	// off-limits. We store the matched globs (not the offending path), so this is
	// phrased as a scope match, not a single "forbidden path".
	if len(route.forbidden) > 0 {
		fmt.Fprintf(&b, "**⚠ Forbidden scope matched:** `%s` — ADR-%s marks these paths off-limits; do not add files here.\n\n", strings.Join(route.forbidden, ", "), r.ID)
	}

	if d := p.Sections["decision"]; d != "" {
		fmt.Fprintf(&b, "**Decision:** %s\n\n", oneLine(d))
	}
	if guidance != "" {
		b.WriteString("**Guidance:**\n")
		for _, line := range bulletLines(guidance) {
			fmt.Fprintf(&b, "- %s\n", line)
		}
		b.WriteString("\n")
	}
	// Companions are expected partner edits this ADR does not govern; list them so
	// the agent considers them, with a soft note when the current change touches
	// none of them.
	if includeCompanions && len(r.D.Companions) > 0 {
		fmt.Fprintf(&b, "**Related files to consider:** %s\n", strings.Join(r.D.Companions, ", "))
		if len(route.companionsHit) == 0 {
			b.WriteString("_(none of the changed paths are among these — listed for context)_\n")
		}
		b.WriteString("\n")
	}
	// Rationale is what lets an agent avoid unsafe "simplifications" of an
	// invariant, so surface Why for invariants (only) when present.
	if includeWhy {
		if why := strings.TrimSpace(p.Sections["why"]); why != "" {
			fmt.Fprintf(&b, "**Why:** %s\n\n", oneLine(why))
		}
	}
	fmt.Fprintf(&b, "→ %s\n\n", r.Filename)
	return b.String()
}

// compactDefaultLine collapses a default to one checklist item:
// `- ADR-NNNN <title>: <summary> → <file>`. The summary prefers the first Guidance
// bullet (the actionable rule) over the Decision, which can be a long paragraph;
// it falls back to the Decision and is capped at maxCompactSummary.
func compactDefaultLine(r Record) string {
	p := ParseBody(r.Body)
	summary := firstGuidanceBullet(p)
	if summary == "" {
		summary = oneLine(p.Sections["decision"])
	}
	summary = truncateRunes(summary, maxCompactSummary)
	if summary == "" {
		return fmt.Sprintf("- ADR-%s %s → %s\n", r.ID, p.Title, r.Filename)
	}
	return fmt.Sprintf("- ADR-%s %s: %s → %s\n", r.ID, p.Title, summary, r.Filename)
}

// relatedFilesSection aggregates the declared companions of every routed record
// into a single section (compact mode). It returns "" when no routed record
// declares a companion.
func relatedFilesSection(groups ...[]briefHit) string {
	var all []briefHit
	for _, g := range groups {
		all = append(all, g...)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].rec.ID < all[j].rec.ID })

	var lines strings.Builder
	for _, h := range all {
		if len(h.rec.D.Companions) > 0 {
			fmt.Fprintf(&lines, "- ADR-%s: %s\n", h.rec.ID, strings.Join(h.rec.D.Companions, ", "))
		}
	}
	if lines.Len() == 0 {
		return ""
	}
	return "\n## Related files to consider\n\n" + lines.String()
}

// firstGuidanceBullet returns the first Guidance bullet (the leading actionable
// rule), or "" when there is no Guidance section.
func firstGuidanceBullet(p Parsed) string {
	g := strings.TrimSpace(guidanceSection(p))
	if g == "" {
		return ""
	}
	lines := bulletLines(g)
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}

// truncateRunes trims s to at most max runes, appending an ellipsis when it cuts.
func truncateRunes(s string, max int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return strings.TrimRight(string(r[:max-1]), " ") + "…"
}

// briefLineCount counts the rendered lines of a brief (for the BriefAuto budget).
func briefLineCount(s string) int {
	return strings.Count(s, "\n")
}

var briefBulletRe = regexp.MustCompile(`^\s*[-*]\s+(.*)$`)

// bulletLines returns the text of each bullet in s, folding hard-wrapped
// continuation lines into their bullet so nothing is truncated. If s has no
// bullets, it returns the whole block collapsed to one line.
func bulletLines(s string) []string {
	var out []string
	cur := -1
	for _, ln := range strings.Split(s, "\n") {
		if m := briefBulletRe.FindStringSubmatch(ln); m != nil {
			out = append(out, strings.TrimSpace(m[1]))
			cur = len(out) - 1
			continue
		}
		trimmed := strings.TrimSpace(ln)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			cur = -1 // blank line or heading ends the current bullet
			continue
		}
		if cur >= 0 {
			out[cur] += " " + trimmed // fold a wrapped continuation line
		}
	}
	if len(out) == 0 {
		return []string{oneLine(s)}
	}
	return out
}

// oneLine collapses all runs of whitespace (including newlines) to single spaces.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
