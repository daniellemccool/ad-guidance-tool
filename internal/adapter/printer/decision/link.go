package decision

import (
	printer "adg/internal/adapter/printer"
)

type LinkPresenter struct {
	s printer.Streams
}

func NewLinkPresenter(s printer.Streams) *LinkPresenter {
	return &LinkPresenter{s: s}
}

func (p *LinkPresenter) Linked(sourceID, targetID, tag, reverseTag string) {
	p.s.Status("Link added: %s →[%s]→ %s\n", sourceID, tag, targetID)
	if reverseTag != "" {
		p.s.Status("Reverse link added: %s →[%s]→ %s\n", targetID, reverseTag, sourceID)
	}
}
