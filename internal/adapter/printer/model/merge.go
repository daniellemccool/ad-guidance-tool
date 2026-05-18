package model

import (
	printer "adg/internal/adapter/printer"
)

type MergeModelsPresenter struct {
	s printer.Streams
}

func NewMergePresenter(s printer.Streams) *MergeModelsPresenter {
	return &MergeModelsPresenter{s: s}
}

func (p *MergeModelsPresenter) Merged(modelAPath, modelBPath, targetPath string, mergedDecisions int) error {
	p.s.Status("Successfully merged %d decisions from models %s and %s to new directory: %s\n", mergedDecisions, modelAPath, modelBPath, targetPath)
	return nil
}
