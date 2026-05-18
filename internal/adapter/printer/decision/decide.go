package decision

import (
	printer "adg/internal/adapter/printer"
)

type DecidePresenter struct {
	s printer.Streams
}

func NewDecidePresenter(s printer.Streams) *DecidePresenter {
	return &DecidePresenter{s: s}
}

func (p *DecidePresenter) Decided(decisionID string) {
	p.s.Status("Decision %s has been marked as decided.\n", decisionID)
}
