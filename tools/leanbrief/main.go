// Command leanbrief is a prototype of a future `adg brief`. Given a set of
// changed file paths, it compiles the architecture guidance packet for them —
// the ADRs whose applies_to globs match, grouped by force, with Decision,
// Guidance, and Checks. Pure path routing, no LLM: suitable for a pre-edit hook
// or CI step. The logic lives in the lean package; this is a thin shell.
//
// Usage:
//
//	# ad-hoc: compile the brief for one or more changed paths
//	go run ./tools/leanbrief --model docs/lean-prototype port/helpers/flow_builder.py
//
//	# Claude Code PreToolUse hook: read the hook JSON on stdin and inject the
//	# brief for the edited file as additionalContext (fail-open, exit 0).
//	leanbrief --hook --model docs/lean-prototype
package main

import (
	"adg/internal/domain/decision/lean"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	modelPath := flag.String("model", "", "path to the lean ADR directory")
	hook := flag.Bool("hook", false, "Claude Code PreToolUse hook mode: read the hook JSON from stdin and inject the brief for the edited file")
	flag.Parse()

	if *modelPath == "" {
		fmt.Fprintln(os.Stderr, "error: --model is required")
		os.Exit(2)
	}

	// Hook mode is fail-open: any error means inject nothing and exit 0 so an
	// edit is never blocked by this hook.
	if *hook {
		records, err := lean.LoadDir(*modelPath)
		if err != nil {
			return
		}
		payload, err := io.ReadAll(os.Stdin)
		if err != nil {
			return
		}
		if out := lean.HookContext(records, payload); out != "" {
			fmt.Println(out)
		}
		return
	}

	changed := flag.Args()
	if len(changed) == 0 {
		fmt.Fprintln(os.Stderr, "error: provide one or more changed paths as arguments")
		os.Exit(2)
	}

	records, err := lean.LoadDir(*modelPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(lean.Brief(records, changed, lean.BriefAuto))
}
