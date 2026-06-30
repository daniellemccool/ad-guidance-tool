package lean

import (
	"bytes"
	"strings"
	"testing"
)

func runReview(t *testing.T, dir string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := NewReviewCommand(nil)
	var out, errb bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errb)
	cmd.SetArgs(append([]string{"--model", dir}, args...))
	err = cmd.Execute()
	return out.String(), errb.String(), err
}

func TestLeanReview_EmitsPacketWithFindings(t *testing.T) {
	dir := t.TempDir()
	// A Decision that is a list trips the deterministic leanness lint, so the packet
	// should carry that finding alongside the record.
	writeADR(t, dir, "0001-test.md",
		"---\nstatus: accepted\ncategory: T\napplies_to:\n    - port/**/*.py\n---\n\n# Test rule\n\n## Decision\n\n- one\n- two\n\n## Guidance\n\n- do x\n")

	out, _, err := runReview(t, dir, "0001-test.md")
	if err != nil {
		t.Fatalf("review errored: %v", err)
	}
	for _, want := range []string{
		"Lean ADR review packet",
		"references/lean-rubric.md",
		"## ADR-0001 — Test rule",
		"Decision contains a list", // the deterministic finding is included
	} {
		if !strings.Contains(out, want) {
			t.Errorf("packet missing %q:\n%s", want, out)
		}
	}
}
