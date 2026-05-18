package cmd

import (
	cmd "adg/internal/adapter/command/model"
	print "adg/internal/adapter/printer/model"
	interactor "adg/internal/application/interactor/model"
)

func init() {
	s := streams()
	validatePresenter := print.NewModelValidatePresenter(s)
	rootCmd.AddCommand(
		cmd.NewCopyCommand(interactor.NewCopyModelInteractor(modelSvc, decisionSvc, print.NewCopyPresenter(s)), configSvc),
		cmd.NewImportCommand(interactor.NewImportModelInteractor(modelSvc, decisionSvc, print.NewImportPresenter(s)), configSvc),
		cmd.NewInitCommand(interactor.NewInitModelInteractor(modelSvc, print.NewInitPresenter(s))),
		cmd.NewMergeModelsCommand(interactor.NewMergeModelsInteractor(modelSvc, decisionSvc, print.NewMergePresenter(s))),
		cmd.NewMigrateCommand(interactor.NewMigrateInteractor(modelSvc, print.NewMigratePresenter(s)), configSvc),
		cmd.NewValidateCommand(
			interactor.NewModelValidateInteractor(modelSvc, validatePresenter),
			configSvc,
			validatePresenter.HadIssues,
		),
	)
}
