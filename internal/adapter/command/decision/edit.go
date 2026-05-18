package decision

import (
	util "adg/internal/adapter/command"
	"adg/internal/application/inputport"
	domain "adg/internal/domain/config"
	"fmt"

	"github.com/spf13/cobra"
)

func NewEditCommand(input inputport.DecisionEdit, config domain.ConfigService) *cobra.Command {
	var modelPath, idOrTitle, id, title string
	var context, drivers string
	var options []string
	var err error

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a decision file",
		RunE: func(cmd *cobra.Command, args []string) error {
			modelPath, err = util.ResolveModelPathOrDefault(modelPath, config)
			if err != nil {
				return err
			}

			err := util.ResolveIdOrTitle(idOrTitle, &id, &title)
			if err != nil {
				return err
			}

			// validate: must be editing something
			if context == "" && drivers == "" && len(options) == 0 {
				return fmt.Errorf("at least one of --context, --option, or --drivers must be provided")
			}

			// prepare pointer args
			var ctxPtr, drvPtr *string
			var oPtr *[]string

			if context != "" {
				ctxPtr = &context
			}
			if drivers != "" {
				drvPtr = &drivers
			}
			if len(options) > 0 {
				oPtr = &options
			}

			return input.Edit(modelPath, id, title, ctxPtr, oPtr, drvPtr)
		},
	}

	cmd.Flags().StringVar(&modelPath, "model", "", "Path to the decision model (optional if set in config)")
	cmd.Flags().StringVar(&idOrTitle, "id", "", "ID or title of the decision to edit (e.g. 0001, 'my-decision')")
	cmd.Flags().StringVar(&context, "context", "", "Append to the Context and Problem Statement section")
	cmd.Flags().StringArrayVar(&options, "option", nil, "Add one or more considered options (repeat to add multiple)")
	cmd.Flags().StringVar(&drivers, "drivers", "", "Append to the Decision Drivers section")

	return cmd
}
