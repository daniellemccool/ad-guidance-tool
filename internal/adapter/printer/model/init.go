package model

import (
	printer "adg/internal/adapter/printer"
)

type InitModelPresenter struct {
	s printer.Streams
}

func NewInitPresenter(s printer.Streams) *InitModelPresenter {
	return &InitModelPresenter{s: s}
}

func (p *InitModelPresenter) Initialized(modelPath string) {
	p.s.Status("Successfully created model directory: %s\n", modelPath)
}
