package decision

import (
	util "adg/internal/adapter/command"
	"adg/internal/application/inputport"
	domain "adg/internal/domain/config"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func NewAddCommand(input inputport.DecisionAdd, config domain.ConfigService) *cobra.Command {
	var titles []string
	var modelPath, id string
	var err error

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Adds one or more decision points to a model",
		// Errors here describe model state (collision, missing dir, bad title)
		// rather than CLI misuse, so don't dump Usage on failure.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			modelPath, err = util.ResolveModelPathOrDefault(modelPath, config)
			if err != nil {
				return err
			}

			if len(titles) == 0 {
				return fmt.Errorf("at least one --title must be provided")
			}

			normalizedID := ""
			if id != "" {
				if len(titles) != 1 {
					return fmt.Errorf("--id can only be used with a single --title (got %d titles)", len(titles))
				}
				normalizedID, err = normalizeID(id)
				if err != nil {
					return err
				}
			}

			return input.Add(modelPath, titles, normalizedID)
		},
	}

	cmd.Flags().StringVar(&modelPath, "model", "", "Path to the decision model (optional if configured)")
	// StringArrayVar (not StringSliceVar) so a single --title is taken verbatim;
	// pflag's StringSlice splits on commas, which silently turned a title like
	// "Store::open, migrate" into two ADRs. Multiple titles still work by
	// repeating the flag: --title A --title B.
	cmd.Flags().StringArrayVar(&titles, "title", nil, "One or more titles for new decisions (repeat the flag to add multiple)")
	cmd.Flags().StringVar(&id, "id", "", "Optional explicit ID (1-9999, zero-padded to 4 digits). Fails if the ID is already taken.")

	return cmd
}

// normalizeID accepts "22" or "0022" and returns "0022". Rejects values outside
// 1..9999 and non-numeric input. 0000 is reserved.
func normalizeID(input string) (string, error) {
	n, err := strconv.Atoi(input)
	if err != nil {
		return "", fmt.Errorf("invalid --id %q: must be a number 1-9999", input)
	}
	if n < 1 || n > 9999 {
		return "", fmt.Errorf("invalid --id %q: must be in range 1-9999", input)
	}
	return fmt.Sprintf("%04d", n), nil
}
