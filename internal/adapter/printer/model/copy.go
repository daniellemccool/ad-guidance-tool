package model

import (
	printer "adg/internal/adapter/printer"
)

type CopyModelPresenter struct {
	s printer.Streams
}

func NewCopyPresenter(s printer.Streams) *CopyModelPresenter {
	return &CopyModelPresenter{s: s}
}

func (p *CopyModelPresenter) Copied(source, target string, copiedDecisions int) {
	p.s.Status("Successfully copied %d decisions from model %s to new model %s\n", copiedDecisions, source, target)
}
