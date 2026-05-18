package model

import (
	printer "adg/internal/adapter/printer"
)

type ImportModelPresenter struct {
	s printer.Streams
}

func NewImportPresenter(s printer.Streams) *ImportModelPresenter {
	return &ImportModelPresenter{s: s}
}

func (p *ImportModelPresenter) Imported(sourcePath, targetPath string, importedDecisions int) error {
	p.s.Status("Successfully imported model %s with %d decisions to: %s\n", sourcePath, importedDecisions, targetPath)
	return nil
}
