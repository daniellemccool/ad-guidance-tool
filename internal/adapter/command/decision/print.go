package decision

import (
	util "adg/internal/adapter/command"
	"adg/internal/application/inputport"
	domain "adg/internal/domain/config"
	"fmt"

	"github.com/spf13/cobra"
)

// NewPrintCommand wires the `adg view` cobra command. Section flag names use
// MADR vocabulary: `--context` for the question section (renamed in MADR to
// "Context and Problem Statement") and `--drivers` for what was "Criteria" in
// the legacy ADG template. With no section flags specified, the full body is
// printed.
//
// todo: rename all related functions and files to view.
func NewPrintCommand(input inputport.DecisionPrint, config domain.ConfigService) *cobra.Command {
	var err error
	var modelPath string
	var idsOrTitles, ids, titles []string
	var printContext, printDrivers, printOptions, printOutcome, printComments bool

	cmd := &cobra.Command{
		Use:   "view",
		Short: "Show the full or partial content of one or more decision files",
		RunE: func(cmd *cobra.Command, args []string) error {
			modelPath, err = util.ResolveModelPathOrDefault(modelPath, config)
			if err != nil {
				return err
			}

			for _, value := range idsOrTitles {
				var id, title string
				if err := util.ResolveIdOrTitle(value, &id, &title); err != nil {
					return err
				}
				if id != "" {
					ids = append(ids, id)
				} else {
					titles = append(titles, title)
				}
			}

			if len(ids) == 0 && len(titles) == 0 {
				return fmt.Errorf("at least one --id or --title must be provided")
			}

			if !printContext && !printDrivers && !printOptions && !printOutcome && !printComments {
				printContext = true
				printDrivers = true
				printOptions = true
				printOutcome = true
				printComments = true
			}

			sections := map[string]bool{
				"context":  printContext,
				"drivers":  printDrivers,
				"options":  printOptions,
				"outcome":  printOutcome,
				"comments": printComments,
			}

			return input.Print(modelPath, ids, titles, sections)
		},
	}

	cmd.Flags().StringVar(&modelPath, "model", "", "Path to the decision model directory")
	cmd.Flags().StringSliceVar(&idsOrTitles, "id", nil, "IDs or titles of the decisions to print (e.g. 0001, 'my-decision') (can be repeated)")

	cmd.Flags().BoolVar(&printContext, "context", false, "Print the Context and Problem Statement section")
	cmd.Flags().BoolVar(&printDrivers, "drivers", false, "Print the Decision Drivers section")
	cmd.Flags().BoolVar(&printOptions, "options", false, "Print the Considered Options section")
	cmd.Flags().BoolVar(&printOutcome, "outcome", false, "Print the Decision Outcome section")
	cmd.Flags().BoolVar(&printComments, "comments", false, "Print the Comments section")

	return cmd
}
