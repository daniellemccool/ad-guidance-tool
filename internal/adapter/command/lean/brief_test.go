package lean

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runBrief executes `adg lean brief` against a temp model dir.
func runBrief(t *testing.T, dir, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := NewBriefCommand()
	var out, errb bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errb)
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetArgs(append([]string{"--model", dir}, args...))
	err = cmd.Execute()
	return out.String(), errb.String(), err
}

func TestLeanBrief_ModeFlagsInvalidWithHook(t *testing.T) {
	dir := t.TempDir()
	for _, flag := range []string{"--full", "--compact"} {
		_, errb, err := runBrief(t, dir, "{}", "--hook", flag)
		if err == nil {
			t.Errorf("%s with --hook should error; stderr:\n%s", flag, errb)
			continue
		}
		if !strings.Contains(errb, "invalid with --hook") {
			t.Errorf("%s with --hook should explain the conflict; got stderr:\n%s", flag, errb)
		}
	}
}

func TestLeanBrief_DigestDiagnosticMode(t *testing.T) {
	// --digest without --hook renders the digest to stdout and the per-rung size
	// report to stderr (the corpus-tuning diagnostic).
	dir := t.TempDir()
	record := "---\nstatus: accepted\ndate: \"2026-08-03\"\ncategory: Meta\npriority: invariant\napplies_to:\n    - src/**/*.go\n---\n\n# Keep it simple\n\n## Decision\n\nRule.\n\n## Guidance\n\n- Do it.\n\n## Why\n\nBecause drift.\n"
	if err := os.WriteFile(filepath.Join(dir, "0001-keep-it-simple.md"), []byte(record), 0o644); err != nil {
		t.Fatal(err)
	}
	out, errb, err := runBrief(t, dir, "", "--digest")
	if err != nil {
		t.Fatalf("--digest diagnostic mode should succeed; stderr:\n%s", errb)
	}
	if !strings.Contains(out, "ADR digest") || !strings.Contains(out, "!0001 Keep it simple") {
		t.Errorf("stdout should carry the digest; got:\n%s", out)
	}
	if !strings.Contains(errb, "selected") || !strings.Contains(errb, "rung") {
		t.Errorf("stderr should carry the rung report; got:\n%s", errb)
	}
}

func TestLeanBrief_DigestMutuallyExclusiveWithOtherHookModes(t *testing.T) {
	dir := t.TempDir()
	for _, flag := range []string{"--whole", "--invariants", "--staged", "--guard"} {
		_, errb, err := runBrief(t, dir, "{}", "--hook", "--digest", flag)
		if err == nil {
			t.Errorf("--digest with %s should error", flag)
			continue
		}
		if !strings.Contains(errb, "mutually exclusive") {
			t.Errorf("--digest with %s should explain the conflict; got stderr:\n%s", flag, errb)
		}
	}
}

func TestLeanBrief_FullAndCompactMutuallyExclusive(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := runBrief(t, dir, "", "--full", "--compact", "x.py"); err == nil {
		t.Errorf("--full and --compact together should error")
	}
}
