package decision

import (
	util "adg/internal/adapter/command"
	"adg/internal/application/inputport"
	domain "adg/internal/domain/config"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func NewEditCommand(input inputport.DecisionEdit, config domain.ConfigService) *cobra.Command {
	var modelPath, idOrTitle, id, title string
	var context, drivers, fromFile string
	var options []string
	var fromStdin, force bool

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a decision file",
		RunE: func(cmd *cobra.Command, args []string) error {
			modelPath, err := util.ResolveModelPathOrDefault(modelPath, config)
			if err != nil {
				return err
			}
			if err := util.ResolveIdOrTitle(idOrTitle, &id, &title); err != nil {
				return err
			}

			appendModeUsed := context != "" || drivers != "" || len(options) > 0
			replaceModeUsed := fromStdin || fromFile != ""

			if fromStdin && fromFile != "" {
				return fmt.Errorf("--from-stdin and --from-file are mutually exclusive")
			}
			if replaceModeUsed && appendModeUsed {
				return fmt.Errorf("replace mode (--from-stdin / --from-file) cannot be combined with append flags (--context / --drivers / --option)")
			}

			if replaceModeUsed {
				body, err := readReplacementBody(fromStdin, fromFile, cmd.InOrStdin())
				if err != nil {
					return err
				}
				return input.ReplaceBody(modelPath, id, title, body, force)
			}

			if !appendModeUsed {
				return fmt.Errorf("at least one of --context, --option, --drivers, --from-stdin, or --from-file must be provided")
			}

			// Append mode (existing behavior).
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
	cmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Replace the decision body with MADR-shaped markdown read from stdin")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Replace the decision body with MADR-shaped markdown read from the given file")
	cmd.Flags().BoolVar(&force, "force", false, "Allow replace mode on a non-proposed decision")

	return cmd
}

func readReplacementBody(fromStdin bool, fromFile string, stdin io.Reader) (string, error) {
	if fromStdin {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}
		return string(data), nil
	}
	data, err := os.ReadFile(fromFile)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", fromFile, err)
	}
	return string(data), nil
}
