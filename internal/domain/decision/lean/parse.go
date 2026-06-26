package lean

import (
	"regexp"
	"strings"
)

// Parsed is the result of parsing a lean ADR body.
type Parsed struct {
	Title    string
	Sections map[string]string // lowercased H2 header -> section body (header line stripped, trimmed)
	Order    []string          // H2 headers in document order, original casing
}

var (
	h1Re = regexp.MustCompile(`(?m)^# +(.+?)\s*$`)
	h2Re = regexp.MustCompile(`(?m)^## +(.+?)\s*$`)
)

// ParseBody extracts the H1 title and the H2 sections of a lean ADR body. The
// returned Sections map is keyed by the lowercased header text; values are the
// section body with the header line removed and surrounding whitespace trimmed.
func ParseBody(body string) Parsed {
	p := Parsed{Sections: map[string]string{}}

	if m := h1Re.FindStringSubmatch(body); m != nil {
		p.Title = strings.TrimSpace(m[1])
	}

	idxs := h2Re.FindAllStringSubmatchIndex(body, -1)
	for i, idx := range idxs {
		header := strings.TrimSpace(body[idx[2]:idx[3]])
		contentStart := idx[1]
		contentEnd := len(body)
		if i+1 < len(idxs) {
			contentEnd = idxs[i+1][0]
		}
		content := strings.TrimSpace(body[contentStart:contentEnd])
		p.Sections[strings.ToLower(header)] = content
		p.Order = append(p.Order, header)
	}
	return p
}
