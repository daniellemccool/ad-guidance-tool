package cmd

import (
	cmd "adg/internal/adapter/command/decision"
	print "adg/internal/adapter/printer/decision"
	interactor "adg/internal/application/interactor/decision"
)

func init() {
	s := streams()
	rootCmd.AddCommand(
		cmd.NewAddCommand(interactor.NewAddDecisionsInteractor(modelSvc, decisionSvc, print.NewAddPresenter(s)), configSvc),
		cmd.NewCommentCommand(interactor.NewCommentDecisionInteractor(decisionSvc, print.NewCommentPresenter(s)), configSvc),
		cmd.NewDecideCommand(interactor.NewDecideInteractor(decisionSvc, print.NewDecidePresenter(s)), configSvc),
		cmd.NewEditCommand(interactor.NewEditDecisionInteractor(decisionSvc, print.NewEditPresenter(s)), configSvc),
		cmd.NewLinkCommand(interactor.NewLinkDecisionsInteractor(decisionSvc, print.NewLinkPresenter(s)), configSvc),
		cmd.NewListCommand(interactor.NewListDecisionsInteractor(decisionSvc, print.NewListPresenter(s)), configSvc),
		cmd.NewPrintCommand(interactor.NewPrintDecisionsInteractor(decisionSvc, print.NewPrintPresenter(s)), configSvc),
		cmd.NewReviseCommand(interactor.NewReviseDecisionInteractor(decisionSvc, print.NewRevisePresenter(s)), configSvc),
		cmd.NewSlugCommand(s),
		cmd.NewSupersedeCommand(interactor.NewSupersedeInteractor(decisionSvc, print.NewSupersedePresenter(s)), configSvc),
		cmd.NewTagCommand(interactor.NewTagDecisionInteractor(decisionSvc, print.NewTagPresenter(s)), configSvc),
	)
}
