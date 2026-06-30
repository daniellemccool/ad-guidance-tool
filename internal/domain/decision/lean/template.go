// Package lean implements a streamlined, LLM-first ADR format as a parallel
// alternative to the MADR shape in the madr package. A lean ADR is a one-screen
// record with two core sections — Decision (what was decided) and Guidance (what
// the next contributor must do; "Implication" is an accepted alias) — plus an
// optional Why (expected for invariants) and Checks when load-bearing. Provenance,
// grouping, and supersession/amendment relations live in machine-readable
// frontmatter (madr.Frontmatter), so the generated index and the integrity checks
// stay automatable.
//
// This package deliberately reuses madr for frontmatter, file splitting, and
// filename parsing; only the body shape, validation rules, and index generation
// differ.
package lean

import (
	"fmt"
	"strings"
)

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

// RenderNewBodyFor emits the lean body scaffold, adding a `## Why` stub only for
// invariants. Rationale is expected on an invariant (it is what lets an agent
// reason about an override instead of silently weakening the rule), so the
// scaffold prompts for it rather than leaving the author to remember.
func RenderNewBodyFor(title, priority string) string {
	body := fmt.Sprintf(LeanBodyTemplate, title)
	if strings.EqualFold(strings.TrimSpace(priority), "invariant") {
		body += "\n## Why\n\n{...}\n"
	}
	return body
}

// PlaceholderTokens are the literal scaffolding strings emitted by the lean
// template. Like MADR's, they are expected in a `proposed` draft but must be
// gone before an ADR is accepted.
var PlaceholderTokens = []string{"{...}"}

// MaxBodyLines is the soft one-screen ceiling. Bodies longer than this earn a
// validator warning (not a hard failure) — the "if it runs past one screen,
// it's probably two ADRs" forcing function, enforced gently.
const MaxBodyLines = 60

// MaxBriefLines is the soft ceiling for a compiled brief in BriefAuto mode. A brief
// that renders longer than this re-renders its defaults compactly, so a hub file —
// one that many ADRs govern — does not inject a wall of full entries as
// additionalContext on every edit. Mirrors MaxBodyLines: a one-screen budget,
// applied to the brief rather than the record.
const MaxBriefLines = 60

// MaxDecisionWords is the soft ceiling for a Decision section (inline code spans
// excluded). A lean Decision states the rule in one to three sentences; a longer
// one is usually a paragraph that has absorbed per-case detail belonging in
// Guidance. Over the ceiling earns an advisory warning, not a failure.
const MaxDecisionWords = 60
