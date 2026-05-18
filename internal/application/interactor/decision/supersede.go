package decision

import (
	"adg/internal/application/inputport"
	util "adg/internal/application/interactor"
	"adg/internal/application/outputport"
	domain "adg/internal/domain/decision"
)

type SupersedeInteractor struct {
	service domain.DecisionService
	output  outputport.DecisionSupersede
}

func NewSupersedeInteractor(service domain.DecisionService, output outputport.DecisionSupersede) inputport.DecisionSupersede {
	return &SupersedeInteractor{
		service: service,
		output:  output,
	}
}

func (i *SupersedeInteractor) Supersede(modelPath, newID, newTitle, oldID, oldTitle, rationale string) error {
	newD, err := util.ResolveDecisionByIdOrTitle(modelPath, newID, newTitle, i.service)
	if err != nil {
		return err
	}
	oldD, err := util.ResolveDecisionByIdOrTitle(modelPath, oldID, oldTitle, i.service)
	if err != nil {
		return err
	}

	if err := i.service.Supersede(modelPath, newD, oldD, rationale); err != nil {
		return err
	}

	i.output.Superseded(newD.ID, oldD.ID)
	return nil
}
