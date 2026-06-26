// Package lean holds the cobra commands for the lean ADR format (brief, index).
// These are intentionally thin shells over the internal/domain/decision/lean
// package, which already returns finished output; they will be promoted to the
// full inputport/interactor/presenter stack used by the MADR commands once the
// lean format graduates from prototype.
package lean

import (
	util "adg/internal/adapter/command"
	domain "adg/internal/domain/config"
	leandomain "adg/internal/domain/decision/lean"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// NewBriefCommand wires `adg brief`. It routes changed paths to the ADRs that
// govern them (by applies_to) and prints the compiled guidance packet; with
// --hook it runs as a Claude Code PreToolUse hook over stdin.
func NewBriefCommand(config domain.ConfigService) *cobra.Command {
	var modelPath string
	var hook bool

	cmd := &cobra.Command{
		Use:   "brief [changed-path...]",
		Short: "Compile the architecture guidance brief for changed files",
		Long: `Brief routes changed file paths to the ADRs that govern them (by their
applies_to globs) and prints a compact guidance packet: the matching ADRs grouped
by force (invariants first), each with its Decision, Guidance, and Checks.

With --hook it runs as a Claude Code PreToolUse hook: it reads the hook JSON on
stdin and injects the brief for the edited file as additionalContext. Hook mode is
fail-open — any error injects nothing and never blocks the edit.`,
		// Errors describe model state, not CLI misuse.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if hook {
				return runHook(cmd, modelPath, config)
			}

			resolved, err := util.ResolveModelPathOrDefault(modelPath, config)
			if err != nil {
				return err
			}
			if len(args) == 0 {
				return fmt.Errorf("provide one or more changed paths, or use --hook")
			}
			records, err := leandomain.LoadDir(resolved)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), leandomain.Brief(records, args))
			return nil
		},
	}

	cmd.Flags().StringVar(&modelPath, "model", "", "Path to the lean ADR directory (optional if configured)")
	cmd.Flags().BoolVar(&hook, "hook", false, "Claude Code PreToolUse hook mode: read hook JSON from stdin and inject the brief for the edited file")
	return cmd
}

// runHook is fully fail-open: a misconfigured model, an empty directory, or a
// malformed payload all inject nothing and exit 0, so the hook never breaks an edit.
func runHook(cmd *cobra.Command, modelPath string, config domain.ConfigService) error {
	resolved, err := util.ResolveModelPathOrDefault(modelPath, config)
	if err != nil {
		return nil
	}
	records, err := leandomain.LoadDir(resolved)
	if err != nil {
		return nil
	}
	payload, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return nil
	}
	if out := leandomain.HookContext(records, payload); out != "" {
		fmt.Fprintln(cmd.OutOrStdout(), out)
	}
	return nil
}
