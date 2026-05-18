package decision

import (
	"adg/internal/application/inputport"
	"adg/internal/application/outputport"
	decisiondomain "adg/internal/domain/decision"
	modeldomain "adg/internal/domain/model"
	"fmt"
)

type AddDecisionsInteractor struct {
	modelService    modeldomain.ModelService
	decisionService decisiondomain.DecisionService
	output          outputport.DecisionAdd
}

func NewAddDecisionsInteractor(
	modelService modeldomain.ModelService,
	decisionService decisiondomain.DecisionService,
	output outputport.DecisionAdd,
) inputport.DecisionAdd {
	return &AddDecisionsInteractor{
		modelService:    modelService,
		decisionService: decisionService,
		output:          output,
	}
}

func (i *AddDecisionsInteractor) Add(modelPath string, titles []string, id string) error {
	var successes []*decisiondomain.Decision
	failures := make(map[string]error)

	if !i.modelService.Exists(modelPath) {
		return fmt.Errorf("can not add decisions, model directory %q does not exist (run `adg init %s` to create it)", modelPath, modelPath)
	}

	if id != "" && len(titles) != 1 {
		return fmt.Errorf("--id can only be used with a single --title (got %d titles)", len(titles))
	}

	for idx, title := range titles {
		assignID := ""
		if idx == 0 {
			assignID = id
		}
		decision, err := i.decisionService.AddNew(modelPath, title, assignID)
		if err != nil {
			// With explicit --id, the caller has made a deterministic claim
			// about which ID they want; surface the failure as a command
			// error (non-zero exit) without going through the batch-style
			// failure printer, which would just re-print the same message.
			if id != "" {
				return err
			}
			failures[title] = err
			continue
		}
		successes = append(successes, decision)
	}

	i.output.Added(successes, failures)
	return nil
}
