package madr

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// LegacyFrontmatter is the shape of the upstream ADG frontmatter we accept
// during migration. Fields we don't carry forward (adr_id) parse but get
// dropped. The `comment` field on LegacyComment is the §A.1 bug surface —
// upstream wrote a placeholder index there; the real text lived in the body
// under a `<a name="comment-N"></a>` anchor.
type LegacyFrontmatter struct {
	ADRID    string              `yaml:"adr_id"`
	Title    string              `yaml:"title"`
	Status   string              `yaml:"status"`
	Tags     []string            `yaml:"tags"`
	Links    map[string][]string `yaml:"links"`
	Comments []LegacyComment     `yaml:"comments"`
}

// LegacyComment mirrors the upstream ADG comment shape with the buggy
// `comment` field (a placeholder index, not the prose).
type LegacyComment struct {
	Author  string `yaml:"author"`
	Date    string `yaml:"date"`
	Comment string `yaml:"comment"`
}

// Section header replacements applied to legacy bodies. Order matters less
// than completeness — every entry is `## ` exact-prefix replacement.
var legacyHeaderMap = map[string]string{
	"Question": "Context and Problem Statement",
	"Options":  "Considered Options",
	"Criteria": "Decision Drivers",
	"Outcome":  "Decision Outcome",
}

var (
	anchorTagRe        = regexp.MustCompile(`<a name="[^"]*"></a>\s*`)
	legacyH2Re         = regexp.MustCompile(`(?m)^## +(.+?)\s*$`)
	commentAnchorStart = regexp.MustCompile(`<a name="comment-(\d+)"></a>`)
	nextH2LineRe       = regexp.MustCompile(`(?m)^## `)
	numberedItem       = regexp.MustCompile(`(?m)^(\s*)\d+\.\s+`)
)

// MigrateLegacy parses upstream ADG bytes and returns a MADR-shaped Decision
// plus body. The Decision is missing ID and Slug — callers (the migrate
// interactor) derive those from the filename and the new H1, then pass both
// to repo.Save which renders the file in MADR shape with the regenerated
// `## Comments` section.
//
// Comment recovery is best-effort: a frontmatter comment whose body anchor
// can be found gets its real text; one that can't gets a placeholder
// "(unrecoverable: <legacy value>)" so the validator flags it.
func MigrateLegacy(content []byte) (Decision, string, error) {
	fmText, body, err := SplitFile(content)
	if err != nil {
		return Decision{}, "", err
	}

	legacyFM, err := parseLegacyFrontmatter(fmText)
	if err != nil {
		return Decision{}, "", err
	}

	d := Decision{
		Tags:  legacyFM.Tags,
		Links: legacyFM.Links,
	}

	switch legacyFM.Status {
	case "open", "":
		d.Status = "proposed"
	case "decided":
		d.Status = "accepted"
		d.LegacyOutcome = true
	default:
		d.Status = legacyFM.Status
	}

	humanTitle := DeslugifyTitle(legacyFM.Title)
	d.Title = humanTitle

	bodyComments := extractLegacyComments(body)
	for i, lc := range legacyFM.Comments {
		anchor := fmt.Sprintf("%d", i+1)
		text := strings.TrimSpace(bodyComments[anchor])
		if text == "" {
			text = fmt.Sprintf("(unrecoverable: legacy comment placeholder %q)", lc.Comment)
		}
		d.Comments = append(d.Comments, Comment{
			Author: lc.Author,
			Date:   lc.Date,
			Text:   text,
		})
	}

	newBody := transformLegacyBody(body, humanTitle)
	return d, newBody, nil
}

func parseLegacyFrontmatter(text string) (LegacyFrontmatter, error) {
	var lfm LegacyFrontmatter
	if strings.TrimSpace(text) == "" {
		return lfm, nil
	}
	if err := yaml.Unmarshal([]byte(text), &lfm); err != nil {
		return lfm, fmt.Errorf("failed to parse legacy frontmatter: %w", err)
	}
	return lfm, nil
}

// DeslugifyTitle turns "define-architecture-layout" into "Define
// architecture layout". This is best-effort — the upstream `title` field was
// always slug-shaped, so there's no perfect recovery of original casing.
// Capitalize the first letter and replace dashes with spaces.
func DeslugifyTitle(slug string) string {
	if slug == "" {
		return ""
	}
	spaced := strings.ReplaceAll(slug, "-", " ")
	return strings.ToUpper(spaced[:1]) + spaced[1:]
}

// extractLegacyComments walks the body looking for `<a name="comment-N"></a>`
// anchors and captures the prose between each anchor and the next one (or
// the next H2, or EOF). Returns map[N]→prose with each value stripped of
// surrounding whitespace and leading H3 lines (the legacy body used
// `### date — author` headers between each comment).
//
// Implemented index-based rather than as one big regex because Go's regexp
// package lacks lookahead — using the next anchor as a non-consuming
// terminator can't be expressed in a single pattern.
func extractLegacyComments(body string) map[string]string {
	result := map[string]string{}
	matches := commentAnchorStart.FindAllStringSubmatchIndex(body, -1)
	for i, m := range matches {
		idx := body[m[2]:m[3]]
		proseStart := m[1]
		proseEnd := len(body)
		if i+1 < len(matches) {
			proseEnd = matches[i+1][0]
		} else if h2 := nextH2LineRe.FindStringIndex(body[proseStart:]); h2 != nil {
			proseEnd = proseStart + h2[0]
		}
		prose := strings.TrimSpace(body[proseStart:proseEnd])
		// Strip the H3 date-author header if present.
		lines := strings.Split(prose, "\n")
		var out []string
		seenHeader := false
		for _, l := range lines {
			if !seenHeader && strings.HasPrefix(strings.TrimSpace(l), "### ") {
				seenHeader = true
				continue
			}
			out = append(out, l)
		}
		result[idx] = strings.TrimSpace(strings.Join(out, "\n"))
	}
	return result
}

// transformLegacyBody applies the body-level migration: strip HTML anchors,
// rename section headers, bulletize numbered options, prepend H1, and
// remove the entire `## Comments` H2 (it gets regenerated from frontmatter
// at save time).
func transformLegacyBody(body, humanTitle string) string {
	// Strip every <a name="X"></a> anywhere in the body.
	body = anchorTagRe.ReplaceAllString(body, "")

	// Rename H2 headers via the legacyHeaderMap.
	body = legacyH2Re.ReplaceAllStringFunc(body, func(line string) string {
		m := legacyH2Re.FindStringSubmatch(line)
		if m == nil {
			return line
		}
		header := strings.TrimSpace(m[1])
		if replacement, ok := legacyHeaderMap[header]; ok {
			return "## " + replacement
		}
		return line
	})

	// Strip the `## Comments` section (regenerated by the renderer on save).
	body = stripCommentsSection(body)

	// Bulletize numbered items inside `## Considered Options`.
	body = bulletizeOptions(body)

	// Prepend H1.
	body = strings.TrimLeft(body, "\n")
	return "# " + humanTitle + "\n\n" + body
}

// bulletizeOptions finds the Considered Options section and converts
// `1. item` to `* item` for each numbered line in that section only.
func bulletizeOptions(body string) string {
	idx := strings.Index(body, "## Considered Options")
	if idx == -1 {
		return body
	}
	end := len(body)
	if next := strings.Index(body[idx+1:], "\n## "); next != -1 {
		end = idx + 1 + next
	}
	section := body[idx:end]
	rewritten := numberedItem.ReplaceAllString(section, "$1* ")
	return body[:idx] + rewritten + body[end:]
}
