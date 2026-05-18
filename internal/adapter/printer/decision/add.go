package decision

import (
	"fmt"

	printer "adg/internal/adapter/printer"
	domain "adg/internal/domain/decision"
)

type AddDecisionsPresenter struct {
	s printer.Streams
}

func NewAddPresenter(s printer.Streams) *AddDecisionsPresenter {
	return &AddDecisionsPresenter{s: s}
}

// Added writes one ID per success to stdout (so callers can pipe or capture)
// and the human-readable status to stderr. Per-title failures print to
// stderr regardless of --quiet because they are errors.
func (p *AddDecisionsPresenter) Added(successes []*domain.Decision, failures map[string]error) {
	for _, decision := range successes {
		fmt.Fprintln(p.s.Out, decision.ID)
		p.s.Status("Decision %s (%s) added successfully.\n", decision.Title, decision.ID)
	}
	for title, err := range failures {
		p.s.Errf("Failed to add decision %q: %v\n", title, err)
	}
}
