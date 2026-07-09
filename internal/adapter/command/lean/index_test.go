package lean

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runIndex executes `adg lean index` against a temp model dir.
func runIndex(t *testing.T, dir string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := NewIndexCommand()
	var out, errb bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errb)
	cmd.SetArgs(append([]string{"--model", dir}, args...))
	err = cmd.Execute()
	return out.String(), errb.String(), err
}

func TestLeanIndex_OverlapsRequiresRoot(t *testing.T) {
	dir := t.TempDir()
	_, errb, err := runIndex(t, dir, "--overlaps")
	if err == nil {
		t.Fatalf("expected an error when --overlaps is used without --root; stderr:\n%s", errb)
	}
	if !strings.Contains(errb, "--overlaps requires --root") {
		t.Errorf("expected a clear '--overlaps requires --root' error; got stderr:\n%s", errb)
	}
}

func TestLeanIndex_OverlapsInvalidValue(t *testing.T) {
	dir := t.TempDir()
	_, errb, err := runIndex(t, dir, "--overlaps=bogus", "--root", ".")
	if err == nil {
		t.Fatalf("expected an error for an invalid --overlaps value; stderr:\n%s", errb)
	}
	if !strings.Contains(errb, "invalid --overlaps") {
		t.Errorf("expected an 'invalid --overlaps' error; got stderr:\n%s", errb)
	}
}

// TestLeanIndex_BodyBudgetHonorsProjectConfig is a command-level regression test
// that `.adg.yaml`'s body_budget is actually wired into `adg index`: with no config
// present, an over-length accepted record earns the one-screen nudge; once the
// model root carries `body_budget: narrative`, the same record no longer does.
// This closes the gap where the loader (budget_test.go) and the domain gate
// (validate_test.go) are each unit-tested but nothing asserts a real command
// connects them — reverting the ValidateWithBudget wiring would pass the suite
// undetected without a test like this one.
func TestLeanIndex_BodyBudgetHonorsProjectConfig(t *testing.T) {
	overLong := "---\nstatus: accepted\ncategory: Test\n---\n\n# Long rule\n\n## Decision\n\nWe do X.\n\n## Guidance\n\n- Do Y.\n\n## Why\n\nWithout it, later code can't tell a valid change from an invalid one.\n"
	for i := 0; i < 70; i++ {
		overLong += fmt.Sprintf("Filler line %d to push the body well past the one-screen ceiling.\n", i)
	}

	t.Run("no config warns", func(t *testing.T) {
		dir := t.TempDir()
		writeADR(t, dir, "0001-long-rule.md", overLong)

		_, errb, err := runIndex(t, dir)
		if err != nil {
			t.Fatalf("index errored: %v; stderr:\n%s", err, errb)
		}
		if !strings.Contains(errb, "should fit one screen") {
			t.Errorf("expected the one-screen nudge with no .adg.yaml present; got stderr:\n%s", errb)
		}
	})

	t.Run("narrative budget suppresses warning", func(t *testing.T) {
		dir := t.TempDir()
		writeADR(t, dir, "0001-long-rule.md", overLong)
		writeConfig(t, dir, "body_budget: narrative\n")

		_, errb, err := runIndex(t, dir)
		if err != nil {
			t.Fatalf("index errored: %v; stderr:\n%s", err, errb)
		}
		if strings.Contains(errb, "should fit one screen") {
			t.Errorf("expected body_budget: narrative to suppress the one-screen nudge; got stderr:\n%s", errb)
		}
	})
}

func TestIndexCommand_DefaultsToDocsDecisions(t *testing.T) {
	dir := t.TempDir()
	model := filepath.Join(dir, "docs", "decisions")
	if err := os.MkdirAll(model, 0o755); err != nil {
		t.Fatal(err)
	}
	record := `---
status: accepted
date: "2026-07-09"
category: Test
priority: default
applies_to:
    - src/**/*.go
---

# Bare invocation resolves the conventional model

## Decision

Test decision.

## Guidance

- Test guidance bullet.

## Why

Test why.
`
	if err := os.WriteFile(filepath.Join(model, "0001-bare-invocation-resolves-the-conventional-model.md"), []byte(record), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	cmd := NewIndexCommand()
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{}) // no --model: must resolve to docs/decisions

	if err := cmd.Execute(); err != nil {
		t.Fatalf("bare `lean index` in a repo with docs/decisions must succeed; err=%v stderr=%s", err, errOut.String())
	}
	if !strings.Contains(out.String(), "0001") {
		t.Fatalf("index output must list the record; got: %s", out.String())
	}
}

func TestIndexCommand_MissingDefaultModelNamesPathAndFlag(t *testing.T) {
	t.Chdir(t.TempDir()) // no docs/decisions here

	cmd := NewIndexCommand()
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("bare `lean index` without docs/decisions must fail")
	}
	if !strings.Contains(errOut.String(), `"docs/decisions"`) || !strings.Contains(errOut.String(), "--model") {
		t.Fatalf("error must name the resolved path and suggest --model; got: %s", errOut.String())
	}
}
