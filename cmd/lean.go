package cmd

import (
	leancmd "adg/internal/adapter/command/lean"
)

// Lean-format commands (prototype). Thin shells over the lean domain package;
// see internal/adapter/command/lean for the promotion-to-full-stack note.
func init() {
	rootCmd.AddCommand(
		leancmd.NewBriefCommand(configSvc),
		leancmd.NewIndexCommand(configSvc),
	)
}
