package lean

import (
	util "adg/internal/adapter/command"
	domain "adg/internal/domain/config"
	leandomain "adg/internal/domain/decision/lean"
	"fmt"

	"github.com/spf13/cobra"
)

// NewCheckCommand wires `adg lean check`: it runs the executable checks declared in
// the model's ADRs (frontmatter `checks` — grep assertions) against the source tree.
// With path arguments it searches only those files (check what changed); without, the
// whole tree under --root. A failing check exits non-zero so CI can gate on it.
func NewCheckCommand(config domain.ConfigService) *cobra.Command {
	var modelPath, root string

	cmd := &cobra.Command{
		Use:   "check [changed-path...]",
		Short: "Run the executable grep-assertion checks declared in the model's ADRs",
		Long: `Check runs the frontmatter "checks" of every lean ADR against the source tree.
Each check is a grep assertion: a regexp that must be absent (the default — a violation if
it matches anywhere in scope) or present, within the check's in/except globs. With path
arguments it searches only those files (the "check what I changed" lens); without, the whole
tree under --root. Exit is non-zero if any check fails.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := util.ResolveModelPathOrDefault(modelPath, config)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			records, err := leandomain.LoadDir(resolved)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			results, err := leandomain.RunChecks(records, root, args)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}

			failed := 0
			for _, c := range results {
				if c.Failed {
					failed++
					fmt.Fprintf(cmd.ErrOrStderr(), "[FAIL] ADR-%s: %s — %s\n", c.ID, c.Desc, c.Detail)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "[pass] ADR-%s: %s\n", c.ID, c.Desc)
				}
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "\nran %d check(s): %d failed\n", len(results), failed)
			if failed > 0 {
				return ErrLeanValidationIssues
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&modelPath, "model", "", "Path to the lean ADR directory (optional if configured)")
	cmd.Flags().StringVar(&root, "root", ".", "Source tree root to search")
	return cmd
}
