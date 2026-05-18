package decision

import (
	"adg/internal/application/inputport"
	util "adg/internal/application/interactor"
	"adg/internal/application/outputport"
	domain "adg/internal/domain/decision"
)

type EditDecisionInteractor struct {
	service domain.DecisionService
	output  outputport.DecisionEdit
}

func NewEditDecisionInteractor(service domain.DecisionService, output outputport.DecisionEdit) inputport.DecisionEdit {
	return &EditDecisionInteractor{
		service: service,
		output:  output,
	}
}

func (i *EditDecisionInteractor) Edit(modelPath, id, title string, context *string, options *[]string, drivers *string) error {
	decision, err := util.ResolveDecisionByIdOrTitle(modelPath, id, title, i.service)
	if err != nil {
		return err
	}

	if err := i.service.Edit(modelPath, decision, context, options, drivers); err != nil {
		return err
	}

	i.output.Edited(decision.ID)
	return nil
}

// ReplaceBody is the replace-mode entry point for `adg edit
// --from-stdin/--from-file`. The service enforces status gating, shape
// checks, and the optional title rewrite; we just resolve the decision
// and route the output event.
func (i *EditDecisionInteractor) ReplaceBody(modelPath, id, title, newBody string, force bool) error {
	decision, err := util.ResolveDecisionByIdOrTitle(modelPath, id, title, i.service)
	if err != nil {
		return err
	}

	if err := i.service.ReplaceBody(modelPath, decision, newBody, force); err != nil {
		return err
	}

	i.output.Edited(decision.ID)
	return nil
}
