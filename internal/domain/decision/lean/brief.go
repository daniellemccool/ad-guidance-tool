package lean

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Brief compiles an agent-facing guidance packet for a set of changed paths: the
// ADRs whose applies_to globs match (via the routeMatch kernel in route.go),
// grouped by force (invariants first, then defaults), each reduced to its
// Decision, Guidance bullets, and a traceability pointer back to the full record —
// plus a consolidated list of any Checks. This is the deterministic core of
// "before I touch this code, what rules constrain me?": pure path routing, no LLM,
// suitable for a pre-edit hook or CI. Brief renders; route.go decides.
func Brief(records []Record, changedPaths []string) string {
	type hit struct {
		rec   Record
		route routeResult
	}
	var invariants, defaults []hit
	for _, r := range records {
		route := routeMatch(r, changedPaths)
		// A record enters the brief when it governs a changed path or when a
		// forbids glob is violated; companions alone never route.
		if len(route.matched) == 0 && len(route.forbidden) == 0 {
			continue
		}
		h := hit{rec: r, route: route}
		if strings.EqualFold(strings.TrimSpace(r.D.Priority), "invariant") {
			invariants = append(invariants, h)
		} else {
			defaults = append(defaults, h)
		}
	}
	byID := func(hs []hit) { sort.Slice(hs, func(i, j int) bool { return hs[i].rec.ID < hs[j].rec.ID }) }
	byID(invariants)
	byID(defaults)

	var b strings.Builder
	b.WriteString("# Architecture brief\n\n")
	fmt.Fprintf(&b, "Changed paths (%d):\n", len(changedPaths))
	for _, p := range changedPaths {
		fmt.Fprintf(&b, "- %s\n", p)
	}
	b.WriteString("\n")

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
			b.WriteString(briefEntry(h.rec, h.route, true))
		}
	}
	if len(defaults) > 0 {
		b.WriteString("\n## Defaults & conventions\n\n")
		for _, h := range defaults {
			b.WriteString(briefEntry(h.rec, h.route, false))
		}
	}

	var checks []string
	for _, h := range append(append([]hit{}, invariants...), defaults...) {
		if c, ok := ParseBody(h.rec.Body).Sections["checks"]; ok {
			for _, line := range bulletLines(c) {
				checks = append(checks, fmt.Sprintf("(ADR-%s) %s", h.rec.ID, line))
			}
		}
	}
	if len(checks) > 0 {
		b.WriteString("\n## Checks to run\n\n")
		for _, c := range checks {
			fmt.Fprintf(&b, "- %s\n", c)
		}
	}
	return b.String()
}

func briefEntry(r Record, route routeResult, includeWhy bool) string {
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
	} else {
		fmt.Fprintf(&b, "_scope: forbids %s_\n\n", strings.Join(route.forbidden, ", "))
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
	if len(r.D.Companions) > 0 {
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
