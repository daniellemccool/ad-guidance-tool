package decision

import (
	printer "adg/internal/adapter/printer"
)

type EditDecisionPresenter struct {
	s printer.Streams
}

func NewEditPresenter(s printer.Streams) *EditDecisionPresenter {
	return &EditDecisionPresenter{s: s}
}

func (p *EditDecisionPresenter) Edited(decisionID string) {
	p.s.Status("Decision %s updated successfully.\n", decisionID)
}
