package decision

import (
	util "adg/internal/adapter/command"
	"adg/internal/application/inputport"
	domain "adg/internal/domain/config"
	"fmt"

	"github.com/spf13/cobra"
)

func NewDecideCommand(input inputport.DecisionDecide, config domain.ConfigService) *cobra.Command {
	var modelPath, idOrTitle, id, title, option, reason, author string
	var force bool
	var err error

	cmd := &cobra.Command{
		Use:   "decide",
		Short: "Marks a decision as accepted by selecting one of its options",
		Long: `Decide finalizes a decision by selecting a specific option from Considered Options
and marking the decision as accepted. The option must already exist in the decision's
Considered Options list — use 'adg edit --option' first if you need to add a new one.

--force bypasses two safety guards: re-deciding an already-accepted ADR, and
overwriting a Decision Outcome section that contains author-written content
(non-placeholder text, including any nested ### Consequences subsection).`,
		// Errors describe model state (already-accepted, authored outcome,
		// unknown option) rather than CLI misuse, so don't dump Usage on failure.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			modelPath, err = util.ResolveModelPathOrDefault(modelPath, config)
			if err != nil {
				return err
			}

			err := util.ResolveIdOrTitle(idOrTitle, &id, &title)
			if err != nil {
				return err
			}

			if author == "" {
				author = config.GetAuthor()
			}

			if option == "" {
				return fmt.Errorf("--option must be provided (either its name or a positive integer (1-based index)")
			}

			return input.Decide(modelPath, id, title, option, reason, author, force)
		},
	}

	cmd.Flags().StringVar(&modelPath, "model", "", "Path to the decision model (optional if configured)")
	cmd.Flags().StringVar(&idOrTitle, "id", "", "ID or title of the decision to decide, e.g., 0001, 'my-decision'")
	cmd.Flags().StringVar(&option, "option", "", "Name or the number of the option being selected, e.g., 'first-option' or '1' (required)")
	cmd.Flags().StringVar(&reason, "rationale", "", "Optional rationale or explanation for the selected option")
	cmd.Flags().StringVar(&author, "author", "", "Name of the person deciding (overrides config)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Bypass safety guards: allow re-deciding an already-accepted ADR, and allow overwriting a Decision Outcome section that contains author-written content (anything other than an empty section, `{...}`, or the unedited `adg add` template).")

	return cmd
}
