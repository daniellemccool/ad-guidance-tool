package decision

import (
	"fmt"

	printer "adg/internal/adapter/printer"
)

type ReviseDecisionPresenter struct {
	s printer.Streams
}

func NewRevisePresenter(s printer.Streams) *ReviseDecisionPresenter {
	return &ReviseDecisionPresenter{s: s}
}

// Revised writes the new revised-decision ID to stdout and the status to
// stderr, mirroring `add`'s output discipline so callers can do
// `NEW=$(adg revise --id 0001)`.
func (p *ReviseDecisionPresenter) Revised(originalID, revisedID string) {
	fmt.Fprintln(p.s.Out, revisedID)
	p.s.Status("Successfully revised decision %s → new decision %s\n", originalID, revisedID)
}
