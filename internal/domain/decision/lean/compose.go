package lean

import (
	"fmt"
	"strconv"
	"strings"
)

// NextID returns the next free flat-global NNNN for a model given its current
// records — the maximum numeric ID plus one, zero-padded. An empty model starts
// at 0001. (IDs are a flat global sequence; see LoadDir.)
func NextID(records []Record) string {
	max := 0
	for _, r := range records {
		if n, err := strconv.Atoi(strings.TrimSpace(r.ID)); err == nil && n > max {
			max = n
		}
	}
	return fmt.Sprintf("%04d", max+1)
}

// EnsureTitle returns body carrying a leading `# <title>` H1. If body has no H1,
// one is prepended; if it already has an H1 equal to title, body is returned
// unchanged; a differing H1 is an error — the author should set the title through
// the command, not hand-write a conflicting H1 in the body.
func EnsureTitle(body, title string) (string, error) {
	switch h1 := strings.TrimSpace(ParseBody(body).Title); h1 {
	case "":
		return "# " + title + "\n\n" + strings.TrimLeft(body, "\n"), nil
	case strings.TrimSpace(title):
		return body, nil
	default:
		return "", fmt.Errorf("body H1 %q differs from title %q; omit the H1 and let the command set it from --title", h1, title)
	}
}
