package decision

import (
	"fmt"

	util "adg/internal/adapter/command"
	"adg/internal/application/inputport"
	domain "adg/internal/domain/config"

	"github.com/spf13/cobra"
)

// NewSupersedeCommand wires `adg supersede --new <id> --old <id> [--rationale ...]`.
// First-class bidirectional supersession: the new decision's Supersedes list
// gains the old ID and (if not already there) its status is promoted to
// accepted; the old decision's status becomes "superseded by ADR-<new>".
// The `link` command refuses tag=supersedes/superseded-by and points users
// here.
func NewSupersedeCommand(input inputport.DecisionSupersede, config domain.ConfigService) *cobra.Command {
	var modelPath, newRef, oldRef, rationale string
	var newID, newTitle, oldID, oldTitle string

	cmd := &cobra.Command{
		Use:   "supersede",
		Short: "Mark a decision as superseded by a newer one (bidirectional)",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := util.ResolveModelPathOrDefault(modelPath, config)
			if err != nil {
				return err
			}
			if newRef == "" || oldRef == "" {
				return fmt.Errorf("both --new and --old are required")
			}
			if err := util.ResolveIdOrTitle(newRef, &newID, &newTitle); err != nil {
				return err
			}
			if err := util.ResolveIdOrTitle(oldRef, &oldID, &oldTitle); err != nil {
				return err
			}
			return input.Supersede(resolved, newID, newTitle, oldID, oldTitle, rationale)
		},
	}

	cmd.Flags().StringVar(&modelPath, "model", "", "Path to the decision model (optional if set in config)")
	cmd.Flags().StringVar(&newRef, "new", "", "ID or title of the new (replacement) decision")
	cmd.Flags().StringVar(&oldRef, "old", "", "ID or title of the old (replaced) decision")
	cmd.Flags().StringVar(&rationale, "rationale", "", "Optional rationale; appended to the new decision's Decision Outcome section")

	return cmd
}
