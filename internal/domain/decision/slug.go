package decision

import (
	"fmt"
	"strings"
)

// Slugify converts a title into a filename-safe slug. Any rune outside
// [a-z0-9-] is replaced with '-' (not stripped) so word boundaries from
// punctuation, underscores, and type parameters survive — e.g.
// `VecDeque<u8>` becomes `vecdeque-u8` rather than `vecdequeu8`. Consecutive
// '-' collapse to one and leading/trailing '-' are trimmed. An empty result
// returns an error so callers surface a clear failure instead of writing
// `NNNN-.md` or producing a filename that won't round-trip.
//
// Exported so the `adg slug` preview command and the file repository can
// share one definition; ADR 0008's plan-paper workflow needs callers to be
// able to predict an ADR's filename without creating it.
func Slugify(title string) (string, error) {
	var b strings.Builder
	prevDash := true // treat start-of-string as already-dashed so we trim leading '-'
	for _, r := range strings.ToLower(title) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	slug := strings.TrimRight(b.String(), "-")
	if slug == "" {
		return "", fmt.Errorf("title %q slugifies to empty; please include at least one letter or digit", title)
	}
	return slug, nil
}
