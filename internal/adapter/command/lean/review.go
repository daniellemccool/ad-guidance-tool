package lean

import (
	util "adg/internal/adapter/command"
	"adg/internal/adapter/leanreview"
	domain "adg/internal/domain/config"
	leandomain "adg/internal/domain/decision/lean"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// defaultReviewModel is the model adg lean review uses unless --reviewer overrides it.
// A rubric judge runs well (and cheaply) on Sonnet; escalate to opus for contentious
// migrations or a release gate.
const defaultReviewModel = "claude-sonnet-4-6"

// NewReviewCommand wires `adg lean review`: an LLM judge of lean ADRs against the
// authoring rubric. It is the one LLM surface; routing/brief/index/check stay
// deterministic. Advisory by default (exit 0); --fail-on-revise gates.
func NewReviewCommand(config domain.ConfigService) *cobra.Command {
	var modelPath, reviewer, since string
	var failOnRevise bool

	cmd := &cobra.Command{
		Use:   "review [adr-file...]",
		Short: "LLM review of lean ADRs against the authoring rubric",
		Long: `Review judges lean ADRs against the lean authoring rubric using Claude and prints a
pass/revise verdict per ADR with specific, rubric-anchored findings. The default model is
claude-sonnet-4-6; pass --reviewer to escalate (e.g. --reviewer claude-opus-4-8). Requires
ANTHROPIC_API_KEY.

Targets: explicit ADR file paths, or --since <git-ref> (records changed since that ref), or all
records in the model when neither is given. Advisory by default (exit 0); --fail-on-revise exits
non-zero if any ADR needs revision, for a release gate.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")) == "" {
				err := fmt.Errorf("ANTHROPIC_API_KEY is not set; `adg lean review` calls the Anthropic API")
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
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
			targets := selectReviewTargets(records, args, since)
			if len(targets) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "no ADRs to review")
				return nil
			}

			// Per-ADR deterministic findings, passed to the model as context.
			byID := map[string][]leandomain.Issue{}
			for _, is := range leandomain.Validate(records) {
				byID[is.ID] = append(byID[is.ID], is)
			}

			r := leanreview.NewReviewer(reviewer)
			revise := 0
			for _, rec := range targets {
				rv, rerr := r.ReviewOne(cmd.Context(), rec, byID[rec.ID])
				if rerr != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error reviewing ADR-%s: %v\n", rec.ID, rerr)
					continue
				}
				printReview(cmd.OutOrStdout(), rv)
				if rv.Verdict == "revise" {
					revise++
				}
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "\nreviewed %d ADR(s): %d need revision\n", len(targets), revise)
			if failOnRevise && revise > 0 {
				return ErrLeanValidationIssues
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&modelPath, "model", "", "Path to the lean ADR directory (optional if configured)")
	cmd.Flags().StringVar(&reviewer, "reviewer", defaultReviewModel, "Claude model for the review (e.g. claude-opus-4-8 to escalate)")
	cmd.Flags().StringVar(&since, "since", "", "Review only the ADR files changed since this git ref")
	cmd.Flags().BoolVar(&failOnRevise, "fail-on-revise", false, "Exit non-zero if any ADR's verdict is \"revise\" (release gate)")
	return cmd
}

func printReview(w io.Writer, rv leanreview.ADRReview) {
	fmt.Fprintf(w, "ADR-%s — %s\n", rv.ADR, strings.ToUpper(rv.Verdict))
	for _, f := range rv.Findings {
		fmt.Fprintf(w, "  - [%s] %s: %s\n", f.Severity, f.RubricRule, f.SuggestedFix)
	}
}

// selectReviewTargets resolves which records to review: explicit ADR file paths, the
// records changed --since a git ref, or all records when neither is given. Paths are
// matched to records by filename.
func selectReviewTargets(records []leandomain.Record, args []string, since string) []leandomain.Record {
	if len(args) == 0 && since == "" {
		return records
	}
	want := map[string]bool{}
	if len(args) > 0 {
		for _, a := range args {
			want[filepath.Base(a)] = true
		}
	} else {
		for _, p := range gitDiffNames(".", since) {
			want[filepath.Base(p)] = true
		}
	}
	var out []leandomain.Record
	for _, r := range records {
		if want[r.Filename] {
			out = append(out, r)
		}
	}
	return out
}

// gitDiffNames returns the paths changed since ref (fail-open to nil).
func gitDiffNames(dir, ref string) []string {
	c := exec.Command("git", "diff", "--name-only", ref)
	c.Dir = dir
	b, err := c.Output()
	if err != nil {
		return nil
	}
	var out []string
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if p := strings.TrimSpace(line); p != "" {
			out = append(out, p)
		}
	}
	return out
}
