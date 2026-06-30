package cmd

import (
	leancmd "adg/internal/adapter/command/lean"

	"github.com/spf13/cobra"
)

// Lean-format commands. Thin shells over the lean domain package; see
// internal/adapter/command/lean for the promotion-to-full-stack note (ADR-0003).
//
// All live under the `lean` parent, with no top-level aliases: `adg lean new`
// (author), `adg lean brief` (compile the brief / PreToolUse hook), `adg lean
// index` (validate + generate the README), `adg lean verify` (Stop-hook re-check),
// `adg lean check` (executable grep assertions), and `adg lean review` (emit a
// review packet for a reviewer to judge against the rubric).
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
