package cmd

import (
	leancmd "adg/internal/adapter/command/lean"

	"github.com/spf13/cobra"
)

// Lean-format commands (prototype). Thin shells over the lean domain package;
// see internal/adapter/command/lean for the promotion-to-full-stack note.
//
// All three live under the `lean` parent: `adg lean new` (author), `adg lean brief`
// (consume), `adg lean index` (validate + generate). No top-level aliases.
func init() {
	leanCmd := &cobra.Command{
		Use:   "lean",
		Short: "Lean-format ADR commands (authoring, brief, index)",
	}
	leanCmd.AddCommand(
		leancmd.NewLeanNewCommand(configSvc),
		leancmd.NewBriefCommand(configSvc),
		leancmd.NewIndexCommand(configSvc),
		leancmd.NewVerifyCommand(configSvc),
		leancmd.NewCheckCommand(configSvc),
		leancmd.NewReviewCommand(configSvc),
	)
	rootCmd.AddCommand(leanCmd)
}
