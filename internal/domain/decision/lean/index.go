package lean

import (
	"fmt"
	"sort"
	"strings"
)

const uncategorized = "Uncategorized"

// RenderIndex generates the grouped README index from a set of lean ADRs —
// removing the hand-maintenance burden that an index-as-prose-file otherwise
// carries. ADRs are grouped by their frontmatter Category; groups are ordered by
// the lowest ADR ID they contain (so the Meta/0001 group leads), and entries
// within a group are ID-ordered. Superseded entries are struck through and
// annotated with their successor; amended entries get an inline note. Both
// annotations are derived from machine-readable frontmatter, so they cannot
// drift from the records themselves.
func RenderIndex(records []Record) string {
	type group struct {
		name  string
		recs  []Record
		minID string
	}
	groups := map[string]*group{}
	for _, r := range records {
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
	b.WriteString("# Architectural decisions\n\n")
	b.WriteString("This index is generated from the ADR frontmatter — do not edit by hand.\n")
	b.WriteString("Load the ADR(s) whose filename matches the area you are touching.\n\n")
	b.WriteString("## Index\n")

	for _, g := range ordered {
		sort.Slice(g.recs, func(i, j int) bool { return g.recs[i].ID < g.recs[j].ID })
		b.WriteString("\n### ")
		b.WriteString(g.name)
		b.WriteString("\n\n")
		for _, r := range g.recs {
			b.WriteString(indexEntry(r))
			b.WriteString("\n")
		}
	}
	return b.String()
}

func indexEntry(r Record) string {
	title := ParseBody(r.Body).Title
	if title == "" {
		title = "(untitled)"
	}
	label := fmt.Sprintf("%s — %s", r.ID, title)
	link := fmt.Sprintf("[%s](./%s)", label, r.Filename)

	status := strings.TrimSpace(r.D.Status)
	if m := supersededByRe.FindStringSubmatch(status); m != nil {
		// Struck through, with the successor called out.
		return fmt.Sprintf("- ~~%s~~ — *superseded by ADR %s*", link, m[1])
	}

	entry := "- " + link
	if m := amendedByRe.FindStringSubmatch(status); m != nil {
		entry += fmt.Sprintf(" *(amended by ADR %s)*", m[1])
	}
	return entry
}
