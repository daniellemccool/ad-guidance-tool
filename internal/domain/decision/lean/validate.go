package lean

import (
	"adg/internal/domain/decision/madr"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Record bundles everything the validator and index generator need about one
// lean ADR: its ID, on-disk filename (for index links), parsed frontmatter, and
// raw body.
type Record struct {
	ID       string
	Filename string
	D        madr.Decision
	Body     string
}

// Issue is one validation finding. Warning issues are advisory (e.g. the
// one-screen length nudge); non-warning issues are failures.
type Issue struct {
	ID      string
	Message string
	Warning bool
}

var (
	statusRe       = regexp.MustCompile(`^(proposed|accepted|rejected|deprecated|superseded by ADR-\d{4}|amended by ADR-\d{4})$`)
	supersededByRe = regexp.MustCompile(`^superseded by ADR-(\d{4})$`)
	amendedByRe    = regexp.MustCompile(`^amended by ADR-(\d{4})$`)
)

// Validate runs lean-shape and integrity checks across a set of ADRs. Body
// checks (required sections, leftover scaffolding) apply only to accepted
// records — proposed drafts are work-in-progress. Relationship integrity
// (supersession, amendment) is checked across the whole set.
func Validate(records []Record) []Issue {
	var issues []Issue
	issues = append(issues, duplicateIDIssues(records)...)

	byID := make(map[string]Record, len(records))
	for _, r := range records {
		byID[r.ID] = r
	}
	for _, r := range records {
		issues = append(issues, validateOne(r, byID)...)
	}
	return issues
}

// duplicateIDIssues flags ADRs that resolve to the same NNNN. The ID model is a
// flat global NNNN across the whole model — `category` frontmatter (not a
// subfolder) does the grouping — so a collision is always a hard error. The
// relationship checks below read a last-wins byID map and so would mis-report
// under duplicates, but the run is already non-zero and the author must fix the
// ID before those findings can be trusted.
func duplicateIDIssues(records []Record) []Issue {
	files := map[string][]string{}
	var order []string
	for _, r := range records {
		if _, seen := files[r.ID]; !seen {
			order = append(order, r.ID)
		}
		files[r.ID] = append(files[r.ID], r.Filename)
	}
	var issues []Issue
	for _, id := range order {
		if fs := files[id]; len(fs) > 1 {
			issues = append(issues, Issue{
				ID:      id,
				Message: fmt.Sprintf("duplicate ID %s used by %d files (%s); IDs must be a unique flat global NNNN across the whole model", id, len(fs), strings.Join(fs, ", ")),
			})
		}
	}
	return issues
}

func validateOne(r Record, byID map[string]Record) []Issue {
	var issues []Issue
	add := func(msg string) { issues = append(issues, Issue{ID: r.ID, Message: msg}) }
	warn := func(msg string) { issues = append(issues, Issue{ID: r.ID, Message: msg, Warning: true}) }

	status := strings.TrimSpace(r.D.Status)
	if status != "" && !statusRe.MatchString(status) {
		add(fmt.Sprintf("status %q is not valid lean vocabulary (proposed, accepted, rejected, deprecated, superseded by ADR-NNNN, amended by ADR-NNNN)", status))
	}

	switch strings.TrimSpace(r.D.Priority) {
	case "", "invariant", "default":
	default:
		add(fmt.Sprintf("priority %q is not valid (invariant, default, or unset)", r.D.Priority))
	}

	// Glob syntax + hygiene across every routing field. Brace expansion is a hard
	// failure (the matcher treats `{` `}` as literals, so a brace glob silently
	// routes to nothing) with a concrete fix; a single-star segment under a
	// nestable directory is an advisory nudge.
	for _, field := range []struct {
		name  string
		globs []string
	}{
		{"applies_to", r.D.AppliesTo},
		{"excludes", r.D.Excludes},
		{"forbids", r.D.Forbids},
		{"companions", r.D.Companions},
	} {
		for _, g := range field.globs {
			if strings.ContainsAny(g, "{}") {
				add(fmt.Sprintf("%s glob %q uses brace expansion, which is not supported; write separate globs (one per alternative)", field.name, g))
			}
			if sug, ok := singleStarNested(g); ok {
				warn(fmt.Sprintf("%s glob %q uses a single-star segment under a directory; it won't match nested paths — did you mean %q?", field.name, g, sug))
			}
		}
	}

	// Companions surface only when the record also routes; companions with no
	// applies_to and no forbids can never appear in a brief.
	if len(r.D.Companions) > 0 && len(r.D.AppliesTo) == 0 && len(r.D.Forbids) == 0 {
		warn("companions set but the record has no applies_to or forbids, so it never routes and the companions never surface")
	}

	p := ParseBody(r.Body)
	if strings.TrimSpace(p.Title) == "" {
		add("H1 title is missing or empty")
	}

	// An accepted ADR is a finished record: both core sections written, no
	// leftover scaffolding. Proposed drafts are exempt.
	if status == "accepted" {
		if sectionEmpty(p, "decision") {
			add("missing or empty required section: Decision")
		}
		if guidanceEmpty(p) {
			add("missing or empty required section: Guidance")
		}
		for _, tok := range PlaceholderTokens {
			if strings.Contains(r.Body, tok) {
				add(fmt.Sprintf("body still contains template placeholder %q; fill it in before accepting", tok))
			}
		}
		if strings.TrimSpace(r.D.Category) == "" {
			warn("no category set; the ADR will be grouped under \"Uncategorized\" in the index")
		}
	}

	if n := bodyLineCount(r.Body); n > MaxBodyLines {
		warn(fmt.Sprintf("body is %d lines (> %d); a lean ADR should fit one screen — consider splitting", n, MaxBodyLines))
	}

	// Supersession integrity (forward + reverse), mirroring adg's MADR validator.
	if m := supersededByRe.FindStringSubmatch(status); m != nil {
		succID := m[1]
		succ, ok := byID[succID]
		if !ok {
			add(fmt.Sprintf("status references ADR-%s but no such ADR exists", succID))
		} else if !slices.Contains(succ.D.Supersedes, r.ID) {
			add(fmt.Sprintf("status says superseded by ADR-%s, but ADR-%s's supersedes list does not include %s", succID, succID, r.ID))
		}
	}
	for _, predID := range r.D.Supersedes {
		pred, ok := byID[predID]
		if !ok {
			add(fmt.Sprintf("supersedes %s but no such ADR exists", predID))
			continue
		}
		want := fmt.Sprintf("superseded by ADR-%s", r.ID)
		if strings.TrimSpace(pred.D.Status) != want {
			add(fmt.Sprintf("supersedes %s but ADR-%s status is %q, not %q", predID, predID, pred.D.Status, want))
		}
	}

	// Amendment integrity (forward + reverse) — the machine-checkable version of
	// Ferry's prose "Amended by" annotation.
	if m := amendedByRe.FindStringSubmatch(status); m != nil {
		amID := m[1]
		am, ok := byID[amID]
		if !ok {
			add(fmt.Sprintf("status references ADR-%s but no such ADR exists", amID))
		} else if !slices.Contains(am.D.Amends, r.ID) {
			add(fmt.Sprintf("status says amended by ADR-%s, but ADR-%s's amends list does not include %s", amID, amID, r.ID))
		}
	}
	for _, baseID := range r.D.Amends {
		if _, ok := byID[baseID]; !ok {
			add(fmt.Sprintf("amends %s but no such ADR exists", baseID))
		}
	}

	return issues
}

func sectionEmpty(p Parsed, key string) bool {
	body, ok := p.Sections[key]
	return !ok || strings.TrimSpace(body) == ""
}

// guidanceSection returns the Guidance section body, accepting the legacy
// "Implication" header as an alias. Guidance is canonical; the two are synonyms.
func guidanceSection(p Parsed) string {
	if g, ok := p.Sections["guidance"]; ok {
		return g
	}
	return p.Sections["implication"]
}

func guidanceEmpty(p Parsed) bool {
	return strings.TrimSpace(guidanceSection(p)) == ""
}

func bodyLineCount(body string) int {
	return len(strings.Split(strings.TrimRight(body, "\n"), "\n"))
}

// singleStarNested reports whether a glob uses a single-star final segment under a
// directory (e.g. platforms/*.py), which silently misses nested paths, and returns
// the recursive form (platforms/**/*.py) to suggest. It deliberately inspects only
// the final segment, so a root-level glob (*.py, main.py) and a mid-path literal
// tail (a/*/b.py) are not flagged.
func singleStarNested(pat string) (string, bool) {
	if strings.Contains(pat, "**") {
		return "", false
	}
	i := strings.LastIndex(pat, "/")
	if i < 0 {
		return "", false
	}
	if !strings.Contains(pat[i+1:], "*") {
		return "", false
	}
	return pat[:i] + "/**/" + pat[i+1:], true
}
