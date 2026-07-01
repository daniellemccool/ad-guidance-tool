package lean

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runVerify(t *testing.T, dir, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := NewVerifyCommand(nil)
	var out, errb bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errb)
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetArgs(append([]string{"--model", dir}, args...))
	err = cmd.Execute()
	return out.String(), errb.String(), err
}

func writeADR(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLeanVerify_ExplicitPathsRendersFooter(t *testing.T) {
	dir := t.TempDir()
	writeADR(t, dir, "0001-test-rule.md",
		"---\nstatus: accepted\ncategory: Test\napplies_to:\n    - port/**/*.py\n---\n\n# Test rule\n\n## Decision\n\nWe do X.\n\n## Guidance\n\n- Do Y.\n\n## Why\n\nWithout it, later code can't tell a valid change from an invalid one.\n")

	// --root "" skips scope lint (the temp dir is not the tree being checked).
	out, _, err := runVerify(t, dir, "", "--root", "", "port/x.py")
	if err != nil {
		t.Fatalf("verify errored: %v", err)
	}
	if !strings.Contains(out, "## Before you finish") || !strings.Contains(out, "ADR-0001 — Test rule") {
		t.Errorf("expected the brief + footer for the governed path:\n%s", out)
	}
}

func TestLeanVerify_HookFailOpen(t *testing.T) {
	// A nonexistent model must not break the agent's stop: --hook always exits 0.
	_, _, err := runVerify(t, filepath.Join(t.TempDir(), "does-not-exist"), `{"cwd":"."}`, "--hook")
	if err != nil {
		t.Errorf("--hook must be fail-open (exit 0) even with a bad model; got: %v", err)
	}
}
