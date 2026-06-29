package lean

import (
	"bytes"
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
