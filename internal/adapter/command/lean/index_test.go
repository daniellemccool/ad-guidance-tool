package lean

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// runIndex executes `adg lean index` against a temp model dir. config is nil —
// safe because --model is always set, so ResolveModelPathOrDefault never touches it
// (and the flag-validation errors under test fire before the model is even loaded).
func runIndex(t *testing.T, dir string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := NewIndexCommand(nil)
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
