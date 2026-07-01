package lean

import (
	"strings"
	"testing"
)

func TestIsRecordFile(t *testing.T) {
	yes := []string{"0005-x.md", "0012-routing-honors-lifecycle.md", "00123-y.md"}
	no := []string{"README.md", "notes.md", "005-x.md", "0005-x.txt", "0005.md"}
	for _, f := range yes {
		if !IsRecordFile(f) {
			t.Errorf("%q should be a record file", f)
		}
	}
	for _, f := range no {
		if IsRecordFile(f) {
			t.Errorf("%q should not be a record file", f)
		}
	}
}

func TestModelGuard_BlocksCreationWarnsEditIgnoresRest(t *testing.T) {
	// Hand-creating a record (Write to a not-yet-existing NNNN-*.md) → block.
	blk := ModelGuard("Write", true, false)
	if !strings.Contains(blk, `"permissionDecision":"deny"`) || !strings.Contains(blk, "adg lean new") {
		t.Errorf("creating a record by hand should be blocked with a use-adg-lean-new reason:\n%s", blk)
	}
	// Overwriting an existing record, or Edit/MultiEdit → warn, never block.
	for _, c := range []struct {
		tool   string
		exists bool
	}{{"Write", true}, {"Edit", true}, {"MultiEdit", true}} {
		out := ModelGuard(c.tool, true, c.exists)
		if strings.Contains(out, "permissionDecision") {
			t.Errorf("%s on an existing record must warn, not block:\n%s", c.tool, out)
		}
		if !strings.Contains(out, "additionalContext") || !strings.Contains(out, "adg lean review") {
			t.Errorf("%s on a record should inject an advisory to use the review flow:\n%s", c.tool, out)
		}
	}
	// A non-record path is ignored entirely.
	if out := ModelGuard("Write", false, false); out != "" {
		t.Errorf("a non-record path should yield nothing, got:\n%s", out)
	}
}
