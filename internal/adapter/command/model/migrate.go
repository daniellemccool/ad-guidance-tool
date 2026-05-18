package model

import (
	util "adg/internal/adapter/command"
	"adg/internal/application/inputport"
	domain "adg/internal/domain/config"

	"github.com/spf13/cobra"
)

// NewMigrateCommand wires `adg migrate --model <dir> [--dry-run]`. It
// converts every file in the model dir that matches the upstream ADG
// format into MADR shape, renaming `AD0001-x.md` to `0001-x.md` and
// rewriting the body. Files already in MADR shape are skipped.
func NewMigrateCommand(input inputport.ModelMigrate, config domain.ConfigService) *cobra.Command {
	var modelPath string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Convert legacy ADG files in a model to MADR 4.0 shape",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := util.ResolveModelPathOrDefault(modelPath, config)
			if err != nil {
				return err
			}
			return input.Migrate(resolved, dryRun)
		},
	}

	cmd.Flags().StringVar(&modelPath, "model", "", "Path to the decision model (optional if set in config)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the planned rewrites without touching files")

	return cmd
}
