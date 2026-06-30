package cmd

import (
	printer "adg/internal/adapter/printer"
	decisiondomain "adg/internal/domain/decision"
	modeldomain "adg/internal/domain/model"
	configinfra "adg/internal/infrastructure/config"
	decisioninfra "adg/internal/infrastructure/decision"
	modelinfra "adg/internal/infrastructure/model"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "adg",
	Short: "Architectural Decision Guidance CLI",
	Long:  "CLI tool for managing architectural decision records and models",
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true, // hides completion cmd from help text but it is still available
	},
}

// version is the adg version string. It defaults to "dev" for source builds and
// is overridden at release time via -ldflags "-X adg/cmd.version=<tag>" (wired by
// .goreleaser.yaml). init() reads it into rootCmd.Version, so the value injected
// at link time is what `adg --version` prints.
var version = "dev"

// Quiet is bound to the persistent --quiet flag. Presenters read it
// through Streams.Quiet (pointer) so the parsed value is visible at
// write time even though presenters are constructed during init().
var Quiet bool

func init() {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("adg {{.Version}}\n")
	rootCmd.PersistentFlags().BoolVar(&Quiet, "quiet", false,
		"Suppress success status messages on stderr; machine values on stdout and errors still print")
}

// streams returns the shared Streams the cmd-layer hands to presenters.
// It binds Quiet by pointer so the flag value is observed at write time.
func streams() printer.Streams {
	return printer.Streams{Out: os.Stdout, Err: os.Stderr, Quiet: &Quiet}
}

var configSvc, configErr = configinfra.NewConfigService()
var decisionRepo = decisioninfra.NewFileDecisionRepository()
var modelRepo = modelinfra.NewFileModelRepository()
var modelSvc = modeldomain.NewModelService(modelRepo, decisionRepo)
var decisionSvc = decisiondomain.NewDecisionService(decisionRepo)

func Execute() error {
	if configErr != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize config service: %v\n", configErr)
		os.Exit(1)
	}
	return rootCmd.Execute()
}
