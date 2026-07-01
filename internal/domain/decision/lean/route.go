package lean

import (
	"regexp"
	"strings"
)

// inForce reports whether a record with this status still governs edits. Terminal
// records — rejected, deprecated, or superseded — are frozen history: routeMatch
// returns nothing for them, so they never appear in a brief or trip the hook, and
// the scope lint and leanness nudges leave them alone. proposed, accepted, amended,
// and unset records are in force. This is the single lifecycle gate; flipping a
// record's status to a terminal value retires it from governance with no need to
// strip its routing globs.
func inForce(status string) bool {
	s := strings.TrimSpace(status)
	if s == "rejected" || s == "deprecated" || supersededByRe.MatchString(s) {
		return false
	}
	return true
}

// This file is the routing kernel for the lean format: the one place that decides,
// for a record and a set of changed paths, what the record governs. Brief (this
// package), Matches (the hook gate), and LintTree's overlap calc all route through
// routeMatch, so routing semantics cannot drift between the brief, the PreToolUse
// hook, and CI. Rendering lives in brief.go and index.go; routing lives here.

// Matches reports whether any record routes to a changed path — the cheap gate
// the hook uses to decide whether to inject anything at all. It uses the same
// in-brief test as Brief (a governed path or a forbids violation), so the hook
// never injects an empty brief and never misses a forbidden-path edit.
func Matches(records []Record, changedPaths []string) bool {
	for _, r := range records {
		route := routeMatch(r, changedPaths)
		if len(route.matched) > 0 || len(route.forbidden) > 0 {
			return true
		}
	}
	return false
}

// Forbidden returns the in-force records whose forbids globs at least one changed path
// violates. It routes through routeMatch (the kernel), so the commit-time block uses the
// same negative-space test as the brief and the hook gate rather than re-deriving it.
func Forbidden(records []Record, changedPaths []string) []Record {
	var out []Record
	for _, r := range records {
		if len(routeMatch(r, changedPaths).forbidden) > 0 {
			out = append(out, r)
		}
	}
	return out
}

// routeResult is the outcome of routing one record against a set of changed
// paths. It is the single source of routing truth — Brief, Matches, the hook, and
// LintTree's overlap calc all derive from routeMatch, so routing semantics live in
// exactly one place.
//
//   - governed:  changed paths this record governs (some applies_to matches and no
//     excludes does), in input order, deduped.
//   - matched:   the applies_to globs that governed at least one path.
//   - excluded:  the excludes globs that suppressed an otherwise-governed path.
//   - forbidden: the forbids globs a changed path hit — a violation, surfaced even
//     when nothing else matches. Forbids is independent of applies_to/excludes.
//   - companionsHit: the companions globs a changed path matched. Companions never
//     route (they are absent from the in-brief test); this only drives the brief's
//     soft "you didn't touch the companion" note.
//
// Pattern slices are emitted in frontmatter declaration order and deduped, so the
// brief's `matched:` line reads in the same order as its `scope:` line.
type routeResult struct {
	governed      []string
	matched       []string
	excluded      []string
	forbidden     []string
	companionsHit []string
}

// routeMatch routes one record against changedPaths. Globs are compiled once per
// record (not per path); an uncompilable glob is skipped — routing is fail-open,
// and the validator and scope lint hard-fail bad globs at the gate instead.
func routeMatch(r Record, changedPaths []string) routeResult {
	// A retired record governs nothing — it is frozen history, not a live rule.
	if !inForce(r.D.Status) {
		return routeResult{}
	}
	applies := compileGlobs(r.D.AppliesTo)
	excludes := compileGlobs(r.D.Excludes)
	forbids := compileGlobs(r.D.Forbids)
	companions := compileGlobs(r.D.Companions)

	matchedSet := map[string]bool{}
	excludedSet := map[string]bool{}
	forbiddenSet := map[string]bool{}
	companionSet := map[string]bool{}
	governedSet := map[string]bool{}
	var governed []string

	for _, p := range changedPaths {
		// forbids and companions are evaluated independently of applies_to.
		for _, g := range forbids {
			if g.re.MatchString(p) {
				forbiddenSet[g.pat] = true
			}
		}
		for _, g := range companions {
			if g.re.MatchString(p) {
				companionSet[g.pat] = true
			}
		}

		var appliedBy []string
		for _, g := range applies {
			if g.re.MatchString(p) {
				appliedBy = append(appliedBy, g.pat)
			}
		}
		if len(appliedBy) == 0 {
			continue // excludes is only meaningful on a path applies_to governs
		}
		excludedBy := ""
		for _, g := range excludes {
			if g.re.MatchString(p) {
				excludedBy = g.pat // first matching excludes wins
				break
			}
		}
		if excludedBy != "" {
			excludedSet[excludedBy] = true
			continue
		}
		for _, pat := range appliedBy {
			matchedSet[pat] = true
		}
		if !governedSet[p] {
			governedSet[p] = true
			governed = append(governed, p)
		}
	}

	// Emit in declaration order, deduping (a pattern listed twice emits once).
	res := routeResult{governed: governed}
	res.matched = emitInOrder(r.D.AppliesTo, matchedSet)
	res.excluded = emitInOrder(r.D.Excludes, excludedSet)
	res.forbidden = emitInOrder(r.D.Forbids, forbiddenSet)
	res.companionsHit = emitInOrder(r.D.Companions, companionSet)
	return res
}

type compiledGlob struct {
	pat string
	re  *regexp.Regexp
}

// compileGlobs compiles each glob once, dropping any that fail to compile (routing
// is fail-open; bad globs are reported by Validate/LintTree, not here).
func compileGlobs(pats []string) []compiledGlob {
	var out []compiledGlob
	for _, pat := range pats {
		re, err := regexp.Compile(globToRegexp(pat))
		if err != nil {
			continue
		}
		out = append(out, compiledGlob{pat: pat, re: re})
	}
	return out
}

// emitInOrder returns the members of set in pats' declaration order, each at most
// once even if pats lists a pattern more than once.
func emitInOrder(pats []string, set map[string]bool) []string {
	var out []string
	for _, pat := range pats {
		if set[pat] {
			out = append(out, pat)
			set[pat] = false
		}
	}
	return out
}
