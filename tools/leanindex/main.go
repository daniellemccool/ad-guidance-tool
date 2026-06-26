// Command leanindex is a prototype of a future `adg index` / `adg validate`
// for the lean ADR format (see internal/domain/decision/lean). It loads a model
// directory, validates each lean ADR, and writes a generated, grouped README
// index. It is intentionally a thin shell over the lean package so the package —
// not this command — carries the logic that would move into adg proper.
//
// Usage:
//
//	go run ./tools/leanindex --model docs/lean-prototype            # validate + print index
//	go run ./tools/leanindex --model docs/lean-prototype --write    # also write README.md
package main

import (
	"adg/internal/domain/decision/lean"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	modelPath := flag.String("model", "", "path to the lean ADR directory")
	write := flag.Bool("write", false, "write the generated index to <model>/README.md")
	root := flag.String("root", "", "source tree root for scope lint (stale applies_to + overlap); skipped if empty")
	flag.Parse()

	if *modelPath == "" {
		fmt.Fprintln(os.Stderr, "error: --model is required")
		os.Exit(2)
	}

	records, err := lean.LoadDir(*modelPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	issues := lean.Validate(records)
	if *root != "" {
		lintIssues, lerr := lean.LintTree(records, *root)
		if lerr != nil {
			fmt.Fprintf(os.Stderr, "error: scope lint: %v\n", lerr)
			os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "[%s] %s: %s\n", kind, is.ID, is.Message)
	}
	fmt.Fprintf(os.Stderr, "\nvalidated %d ADR(s): %d failure(s), %d warning(s)\n", len(records), errCount, len(issues)-errCount)

	index := lean.RenderIndex(records)
	if *write {
		out := filepath.Join(*modelPath, "README.md")
		if err := os.WriteFile(out, []byte(index), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing index: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "wrote %s\n", out)
	} else {
		fmt.Print(index)
	}

	if errCount > 0 {
		os.Exit(1)
	}
}
