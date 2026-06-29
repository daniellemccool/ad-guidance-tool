package lean

import (
	util "adg/internal/adapter/command"
	domain "adg/internal/domain/config"
	leandomain "adg/internal/domain/decision/lean"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// ErrLeanValidationIssues is returned when the lean validator (or scope lint)
// reports a hard failure, so main.go exits non-zero. Warnings do not trigger it.
var ErrLeanValidationIssues = errors.New("lean validation issues found")

// NewIndexCommand wires `adg index`. It validates every lean ADR and prints (or
// writes) the category-grouped README; with --root it also runs scope lint.
func NewIndexCommand(config domain.ConfigService) *cobra.Command {
	var modelPath, root string
	var write bool

	cmd := &cobra.Command{
		Use:   "index",
		Short: "Validate lean ADRs and generate the grouped README index",
		Long: `Index validates every lean ADR and prints (or writes, with --write) the
category-grouped README index. Validation hard-fails on a duplicate ID or an
unsupported brace glob ({a,b}) and warns on glob-hygiene nits. With --root <tree>
it also runs scope lint: stale applies_to/excludes globs (matching no file under
the tree), forbids globs that now match a file, and default-vs-default scope
overlap (computed on applies_to minus excludes). Warnings are advisory; a hard
validation failure exits non-zero.`,
		SilenceErrors: true, // issues already printed to stderr
		SilenceUsage:  true, // failure is data, not user error
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

			issues := leandomain.Validate(records)
			if root != "" {
				lintIssues, lerr := leandomain.LintTree(records, root)
				if lerr != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: scope lint: %v\n", lerr)
					return lerr
				}
				issues = append(issues, lintIssues...)
			}

			errCount := 0
			for _, is := range issues {
				kind := "FAIL"
				if is.Warning {
					kind = "warn"
				} else {
					errCount++
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "[%s] %s: %s\n", kind, is.ID, is.Message)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "\nvalidated %d ADR(s): %d failure(s), %d warning(s)\n", len(records), errCount, len(issues)-errCount)

			index := leandomain.RenderIndex(records)
			if write {
				out := filepath.Join(resolved, "README.md")
				if werr := os.WriteFile(out, []byte(index), 0o644); werr != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: writing index: %v\n", werr)
					return werr
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "wrote %s\n", out)
			} else {
				fmt.Fprint(cmd.OutOrStdout(), index)
			}

			if errCount > 0 {
				return ErrLeanValidationIssues
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&modelPath, "model", "", "Path to the lean ADR directory (optional if configured)")
	cmd.Flags().StringVar(&root, "root", "", "Source tree root for scope lint (stale applies_to/excludes, forbids-has-files, overlap); skipped if empty")
	cmd.Flags().BoolVar(&write, "write", false, "Write the generated index to <model>/README.md instead of stdout")
	return cmd
}
