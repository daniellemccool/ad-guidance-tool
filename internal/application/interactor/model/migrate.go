package model

import (
	"adg/internal/application/inputport"
	"adg/internal/application/outputport"
	modeldomain "adg/internal/domain/model"
)

type MigrateInteractor struct {
	service modeldomain.ModelService
	output  outputport.ModelMigrate
}

func NewMigrateInteractor(service modeldomain.ModelService, output outputport.ModelMigrate) inputport.ModelMigrate {
	return &MigrateInteractor{
		service: service,
		output:  output,
	}
}

func (i *MigrateInteractor) Migrate(modelPath string, dryRun bool) error {
	steps, err := i.service.Migrate(modelPath, dryRun)
	if err != nil {
		return err
	}

	out := make([]outputport.MigrationStep, 0, len(steps))
	for _, s := range steps {
		var errMsg string
		if s.Error != nil {
			errMsg = s.Error.Error()
		}
		out = append(out, outputport.MigrationStep{
			OldPath: s.OldPath,
			NewPath: s.NewPath,
			Error:   errMsg,
		})
	}
	i.output.Migrated(out, dryRun)
	return nil
}
