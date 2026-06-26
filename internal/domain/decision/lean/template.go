// Package lean implements a streamlined, LLM-first ADR format as a parallel
// alternative to the MADR shape in the madr package. A lean ADR is a one-screen
// record with two core sections — Decision (what was decided) and Implication
// (what the next contributor must do) — plus optional Why/Context/Alternatives
// sections when load-bearing. Provenance, grouping, and supersession/amendment
// relations live in machine-readable frontmatter (madr.Frontmatter), so the
// generated index and the integrity checks stay automatable.
//
// This package deliberately reuses madr for frontmatter, file splitting, and
// filename parsing; only the body shape, validation rules, and index generation
// differ.
package lean

import "fmt"

// LeanBodyTemplate is the body scaffold a new lean ADR is created with.
// Frontmatter (status, category, source) is rendered separately by
// madr.RenderFile from the Decision struct, so it is not part of this string.
//
// The shape is intentionally minimal: a one-to-three-sentence Decision and
// Guidance stating the operational consequence (what new code must do, what
// review should reject, the fix path). Optional `## Why`, `## Checks`,
// `## Context`, and `## Alternatives` are added only when load-bearing — the
// goal is brevity. If a lean ADR runs past one screen, it is probably two ADRs.
const LeanBodyTemplate = `# %s

## Decision

{...}

## Guidance

{...}
`

// RenderNewBody emits the lean body scaffold for a freshly-created ADR.
func RenderNewBody(title string) string {
	return fmt.Sprintf(LeanBodyTemplate, title)
}

// PlaceholderTokens are the literal scaffolding strings emitted by the lean
// template. Like MADR's, they are expected in a `proposed` draft but must be
// gone before an ADR is accepted.
var PlaceholderTokens = []string{"{...}"}

// MaxBodyLines is the soft one-screen ceiling. Bodies longer than this earn a
// validator warning (not a hard failure) — the "if it runs past one screen,
// it's probably two ADRs" forcing function, enforced gently.
const MaxBodyLines = 60
