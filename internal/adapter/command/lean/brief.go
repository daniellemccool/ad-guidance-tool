// Package lean holds the cobra commands for the lean ADR format (new, brief,
// index, verify, check, review). These are intentionally thin shells over the
// internal/domain/decision/lean package, which already returns finished output:
// the thin-shell shortcut is the named, time-boxed exception ADR-0003 requires,
// and ADR-0002 governs the deferred promotion onto the full inputport/interactor/
// presenter stack — whose presenter must delegate to the shared renderer rather
// than reimplement it.
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
	var hook, full, compact bool

	cmd := &cobra.Command{
		Use:   "brief [changed-path...]",
		Short: "Compile the architecture guidance brief for changed files",
		Long: `Brief routes changed file paths to the ADRs that govern them and prints a
compact guidance packet: the matching ADRs grouped by force (invariants first),
each with its Decision and Guidance, and a "Before you finish" footer that re-runs
the index gate and lists the matched ADRs' Checks and named test files.

Routing is by frontmatter globs: applies_to selects, excludes carves out sanctioned
or out-of-scope paths, and forbids flags edits to negative-space paths; companions
are surfaced as related files. Without --hook, brief also validates the model and
prints any issues (e.g. an unsupported brace glob) to stderr, exiting non-zero on a
hard failure, so a malformed scope is never silently mis-routed.

By default the brief is rendered in auto mode: full entries, but if the brief would
exceed one screen the defaults collapse to a one-line checklist (invariants and
forbidden-path hits always stay full). --full forces every entry full with scope
detail (the debuggable form); --compact forces the condensed form.

With --hook it runs as a Claude Code PreToolUse hook: it reads the hook JSON on
stdin and injects the brief for the edited file as additionalContext. Hook mode is
fail-open — any error injects nothing and never blocks the edit. Hook mode always
uses auto rendering; --full/--compact are invalid with --hook.`,
		SilenceErrors: true, // issues already printed to stderr
		SilenceUsage:  true, // failure is data, not user error
		RunE: func(cmd *cobra.Command, args []string) error {
			if hook {
				if full || compact {
					err := fmt.Errorf("--full/--compact are invalid with --hook (hook mode always uses auto rendering)")
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
					return err
				}
				return runHook(cmd, modelPath, config)
			}

			mode := leandomain.BriefAuto
			if full {
				mode = leandomain.BriefFull
			} else if compact {
				mode = leandomain.BriefCompact
			}

			resolved, err := util.ResolveModelPathOrDefault(modelPath, config)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			if len(args) == 0 {
				err := fmt.Errorf("provide one or more changed paths, or use --hook")
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			records, err := leandomain.LoadDir(resolved)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			// Non-hook brief is fail-closed: surface model problems (e.g. an
			// unsupported brace glob that would silently mis-route) to stderr rather
			// than render a wrong brief in silence. runHook stays fail-open.
			hard := reportLeanIssues(cmd.ErrOrStderr(), leandomain.Validate(records))
			fmt.Fprint(cmd.OutOrStdout(), leandomain.Brief(records, args, mode))
			if hard > 0 {
				return ErrLeanValidationIssues
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&modelPath, "model", "", "Path to the lean ADR directory (optional if configured)")
	cmd.Flags().BoolVar(&hook, "hook", false, "Claude Code PreToolUse hook mode: read hook JSON from stdin and inject the brief for the edited file")
	cmd.Flags().BoolVar(&full, "full", false, "Render every governing ADR in full, with scope/matched detail (debuggable)")
	cmd.Flags().BoolVar(&compact, "compact", false, "Render defaults as a one-line checklist; invariants and forbidden-path hits stay full")
	cmd.MarkFlagsMutuallyExclusive("full", "compact")
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

// reportLeanIssues prints validation issues to w in the same [FAIL]/[warn] form as
// `adg index`, returning the number of hard failures (so a caller can exit non-zero).
func reportLeanIssues(w io.Writer, issues []leandomain.Issue) int {
	hard := 0
	for _, is := range issues {
		kind := "FAIL"
		if is.Warning {
			kind = "warn"
		} else {
			hard++
		}
		fmt.Fprintf(w, "[%s] %s: %s\n", kind, is.ID, is.Message)
	}
	return hard
}
