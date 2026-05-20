package decision

import (
	printer "adg/internal/adapter/printer"
	domain "adg/internal/domain/decision"
	"fmt"

	"github.com/spf13/cobra"
)

// NewSlugCommand prints the slug a title would produce, without creating
// anything. Plan briefs that need to reference an ADR's eventual filename
// (NNNN-<slug>.md) call this during plan-writing so the predicted slug
// matches what `adg add` will actually emit. Output goes to stdout per
// ADR 0004 (machine values on stdout); errors and any non-empty diagnostics
// go to stderr.
func NewSlugCommand(s printer.Streams) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "slug TITLE",
		Short:        "Print the slug a given title would produce (no ADR is created)",
		Long:         "Slugify a title using the same rules as `adg add` and print the result.\nIntended for plan briefs that reference filenames like NNNN-<slug>.md before the ADR exists.",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, err := domain.Slugify(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(s.Out, slug)
			return nil
		},
	}
	return cmd
}
