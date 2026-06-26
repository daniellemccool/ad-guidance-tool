package lean

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Brief compiles an agent-facing guidance packet for a set of changed paths: the
// ADRs whose applies_to globs match, grouped by force (invariants first, then
// defaults), each reduced to its Decision, Guidance bullets, and a traceability
// pointer back to the full record — plus a consolidated list of any Checks. This
// is the deterministic core of "before I touch this code, what rules constrain
// me?": pure path routing, no LLM, suitable for a pre-edit hook or CI.
func Brief(records []Record, changedPaths []string) string {
	type hit struct {
		rec     Record
		matched []string
	}
	var invariants, defaults []hit
	for _, r := range records {
		matched := matchedPatterns(r, changedPaths)
		if len(matched) == 0 {
			continue
		}
		h := hit{rec: r, matched: matched}
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
			b.WriteString(briefEntry(h.rec, h.matched, true))
		}
	}
	if len(defaults) > 0 {
		b.WriteString("\n## Defaults & conventions\n\n")
		for _, h := range defaults {
			b.WriteString(briefEntry(h.rec, h.matched, false))
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

// Matches reports whether any record's applies_to globs match a changed path —
// the cheap gate the hook uses to decide whether to inject anything at all.
func Matches(records []Record, changedPaths []string) bool {
	for _, r := range records {
		if len(matchedPatterns(r, changedPaths)) > 0 {
			return true
		}
	}
	return false
}

// matchedPatterns returns the record's applies_to globs that match at least one
// changed path. A record with no applies_to never matches (it does not auto-route).
func matchedPatterns(r Record, changedPaths []string) []string {
	var matched []string
	for _, pat := range r.D.AppliesTo {
		re, err := regexp.Compile(globToRegexp(pat))
		if err != nil {
			continue
		}
		for _, p := range changedPaths {
			if re.MatchString(p) {
				matched = append(matched, pat)
				break
			}
		}
	}
	return matched
}

func briefEntry(r Record, matched []string, includeWhy bool) string {
	p := ParseBody(r.Body)
	guidance := guidanceSection(p)

	var b strings.Builder
	fmt.Fprintf(&b, "### ADR-%s — %s\n", r.ID, p.Title)
	fmt.Fprintf(&b, "_scope: %s · matched: %s_\n\n", strings.Join(r.D.AppliesTo, ", "), strings.Join(matched, ", "))
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
