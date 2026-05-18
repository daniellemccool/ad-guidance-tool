package decision

import (
	printer "adg/internal/adapter/printer"
)

type SupersedePresenter struct {
	s printer.Streams
}

func NewSupersedePresenter(s printer.Streams) *SupersedePresenter {
	return &SupersedePresenter{s: s}
}

func (p *SupersedePresenter) Superseded(newID, oldID string) {
	p.s.Status("ADR-%s now supersedes ADR-%s; ADR-%s status set to \"superseded by ADR-%s\"\n", newID, oldID, oldID, newID)
}
