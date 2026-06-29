package lean

import (
	"bytes"
	"strings"
	"testing"
)

// runBrief executes `adg lean brief` against a temp model dir. config is nil —
// safe because --model is always set, and the flag-conflict errors under test fire
// before the model is loaded.
func runBrief(t *testing.T, dir, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := NewBriefCommand(nil)
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

func TestLeanBrief_FullAndCompactMutuallyExclusive(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := runBrief(t, dir, "", "--full", "--compact", "x.py"); err == nil {
		t.Errorf("--full and --compact together should error")
	}
}
