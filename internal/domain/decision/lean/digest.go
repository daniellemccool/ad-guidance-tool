package lean

import (
	"fmt"
	"sort"
	"strings"
)

// BriefDigest compiles the session-open digest (ADR-0021): a grouped, titles-only
// tripwire index of every in-force record, hard-capped at MaxDigestBytes. The digest
// is a recall layer, not an instruction layer — the title is the payload, and full
// guidance stays on demand (`adg lean brief <paths>`, the edit-time hooks). A
// degradation ladder guarantees the cap: full digest, then invariant-only digest
// plus a defaults count, then a fixed floor of counts and pointers. Fail-open —
// empty when no in-force records exist, and never an error. Pure rendering: no
// path routing happens here (routing stays in route.go per ADR-0001).
func BriefDigest(records []Record) string {
	var live, invs []Record
	for _, r := range records {
		if !inForce(r.D.Status) {
			continue
		}
		live = append(live, r)
		if digestInvariant(r) {
			invs = append(invs, r)
		}
	}
	if len(live) == 0 {
		return ""
	}
	total, invCount := len(live), len(invs)

	if out := renderDigest(live, total, invCount, false); len(out) <= MaxDigestBytes {
		return out
	}
	if invCount > 0 {
		if out := renderDigest(invs, total, invCount, true); len(out) <= MaxDigestBytes {
			return out
		}
	}
	return digestFloor(total, invCount)
}

func digestInvariant(r Record) bool {
	return strings.EqualFold(strings.TrimSpace(r.D.Priority), "invariant")
}

// digestContract is the one framing line every digest rung opens with: what the
// list is, what the marker means, and where the full text lives.
func digestContract(total, invariants int) string {
	return fmt.Sprintf("ADR digest — %d records (%d invariants). Titles ARE the decisions. "+
		"! = invariant: open the record before planning changes in its scope. "+
		"Full text: `adg lean brief <paths>`; index: docs/decisions/README.md.\n", total, invariants)
}

// renderDigest renders one digest rung: the contract line, then category groups
// (ordering mirrors RenderIndex: groups by lowest contained ADR ID, Uncategorized
// last, entries ID-ordered). invariantOnly appends the defaults-count closing line
// so the dropped tier stays visible.
func renderDigest(recs []Record, total, invariants int, invariantOnly bool) string {
	type group struct {
		name  string
		recs  []Record
		minID string
	}
	groups := map[string]*group{}
	for _, r := range recs {
		cat := strings.TrimSpace(r.D.Category)
		if cat == "" {
			cat = uncategorized
		}
		g, ok := groups[cat]
		if !ok {
			g = &group{name: cat, minID: r.ID}
			groups[cat] = g
		}
		g.recs = append(g.recs, r)
		if r.ID < g.minID {
			g.minID = r.ID
		}
	}
	ordered := make([]*group, 0, len(groups))
	for _, g := range groups {
		sort.Slice(g.recs, func(i, j int) bool { return g.recs[i].ID < g.recs[j].ID })
		ordered = append(ordered, g)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].name == uncategorized {
			return false
		}
		if ordered[j].name == uncategorized {
			return true
		}
		return ordered[i].minID < ordered[j].minID
	})

	var b strings.Builder
	b.WriteString(digestContract(total, invariants))
	for _, g := range ordered {
		b.WriteString("\n")
		b.WriteString(strings.ToUpper(g.name))
		if hint := scopeHint(g.recs); hint != "" {
			fmt.Fprintf(&b, "  (%s)", hint)
		}
		b.WriteString("\n")
		for _, r := range g.recs {
			b.WriteString(digestLine(r))
		}
	}
	if invariantOnly && total > invariants {
		fmt.Fprintf(&b, "\nPlus %d defaults & conventions — see docs/decisions/README.md.\n", total-invariants)
	}
	return b.String()
}

// digestLine renders one record: ` [!]NNNN <title>`. Bare ID, no filename — the ID
// locates docs/decisions/NNNN-*.md. Titles truncate at MaxTitleRunes, the same cap
// the title-length nudge warns at, so a compliant title is never clipped.
func digestLine(r Record) string {
	marker := ""
	if digestInvariant(r) {
		marker = "!"
	}
	title := truncateRunes(ParseBody(r.Body).Title, MaxTitleRunes)
	if title == "" {
		return fmt.Sprintf(" %s%s\n", marker, r.ID)
	}
	return fmt.Sprintf(" %s%s %s\n", marker, r.ID, title)
}

// scopeHint derives a group's path hint from its members' applies_to globs: the
// literal directory prefix of each glob, deduped, prefix-subsumed, at most the
// three shortest. Empty when no member declares a scope.
func scopeHint(recs []Record) string {
	seen := map[string]bool{}
	var dirs []string
	for _, r := range recs {
		for _, pat := range r.D.AppliesTo {
			d := literalDir(pat)
			if d != "" && !seen[d] {
				seen[d] = true
				dirs = append(dirs, d)
			}
		}
	}
	if len(dirs) == 0 {
		return ""
	}
	sort.Slice(dirs, func(i, j int) bool {
		if len(dirs[i]) != len(dirs[j]) {
			return len(dirs[i]) < len(dirs[j])
		}
		return dirs[i] < dirs[j]
	})
	var kept []string
	for _, d := range dirs {
		subsumed := false
		for _, k := range kept {
			if strings.HasPrefix(d, k) {
				subsumed = true
				break
			}
		}
		if !subsumed {
			kept = append(kept, d)
		}
		if len(kept) == 3 {
			break
		}
	}
	return strings.Join(kept, ", ")
}

// literalDir returns the literal directory prefix of a glob: everything before the
// first metacharacter, trimmed back to the last path separator. A bare filename
// (or a glob with no literal directory) yields "".
func literalDir(pat string) string {
	end := len(pat)
	if i := strings.IndexAny(pat, "*?[{"); i >= 0 {
		end = i
	}
	prefix := pat[:end]
	i := strings.LastIndex(prefix, "/")
	if i < 0 {
		return ""
	}
	return prefix[:i+1]
}

// digestFloor is the always-fits bottom rung: counts and pointers, no per-record
// lines. Emitted when even the invariant-only digest busts MaxDigestBytes.
func digestFloor(total, invariants int) string {
	return fmt.Sprintf("ADR digest — %d lean records (%d invariants) govern this codebase; "+
		"too many to list inline. Read the index (docs/decisions/README.md), then pull the "+
		"rules for the paths you will touch: `adg lean brief <paths>`.\n", total, invariants)
}
