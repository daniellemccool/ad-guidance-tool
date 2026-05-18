package model

import (
	util "adg/internal/adapter/command"
	"adg/internal/application/inputport"
	domain "adg/internal/domain/config"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// ErrValidationIssues is returned by the validate cobra command when the
// presenter reports any issue. It carries no message because the issues
// have already been printed to stderr; main.go just needs to exit
// non-zero. The validate command sets SilenceErrors and SilenceUsage so
// cobra doesn't double-print or show usage on a validation-failed exit.
var ErrValidationIssues = errors.New("validation issues found")

// NewValidateCommand wires `adg validate`. The hadIssues callback is the
// validate presenter's HadIssues method; passing a closure keeps the
// cobra command independent of the concrete presenter type.
func NewValidateCommand(input inputport.ModelValidate, config domain.ConfigService, hadIssues func() bool) *cobra.Command {
	var modelPath string

	cmd := &cobra.Command{
		Use:           "validate",
		Short:         "Validate the model's ADRs against MADR shape and integrity rules",
		SilenceErrors: true, // issues already printed to stderr by the presenter
		SilenceUsage:  true, // failure is data, not user error
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedPath, err := util.ResolveModelPathOrDefault(modelPath, config)
			if err != nil {
				// SilenceErrors swallows this otherwise, so write it ourselves.
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}

			if err := input.Validate(resolvedPath); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			if hadIssues() {
				return ErrValidationIssues
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&modelPath, "model", "", "Path to the decision model directory (optional if configured)")

	return cmd
}
