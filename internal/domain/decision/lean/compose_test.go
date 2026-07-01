package lean

import (
	"strings"
	"testing"
)

func TestNextID(t *testing.T) {
	if got := NextID(nil); got != "0001" {
		t.Errorf("empty model NextID = %q, want 0001", got)
	}
	recs := []Record{{ID: "0001"}, {ID: "0003"}, {ID: "notanid"}}
	if got := NextID(recs); got != "0004" {
		t.Errorf("NextID = %q, want 0004 (max numeric + 1, non-numeric ignored)", got)
	}
}

func TestEnsureTitle(t *testing.T) {
	// No H1 -> prepended.
	got, err := EnsureTitle("## Decision\n\nx\n", "My Title")
	if err != nil || !strings.HasPrefix(got, "# My Title\n\n## Decision") {
		t.Errorf("prepend failed: %q (err %v)", got, err)
	}
	// Matching H1 -> unchanged.
	body := "# My Title\n\n## Decision\n\nx\n"
	if got, err := EnsureTitle(body, "My Title"); err != nil || got != body {
		t.Errorf("matching H1 should be unchanged: %q (err %v)", got, err)
	}
	// Differing H1 -> error, no body.
	if _, err := EnsureTitle("# Other\n\nx\n", "My Title"); err == nil {
		t.Errorf("expected error on conflicting H1")
	}
}

func TestRenderNewBodyFor_AlwaysScaffoldsWhy(t *testing.T) {
	// Every routed record carries its rationale, so the scaffold prompts for a Why
	// regardless of priority (the invariant-vs-default distinction is in the validator).
	for _, priority := range []string{"invariant", "default", ""} {
		if got := RenderNewBodyFor("T", priority); !strings.Contains(got, "## Why") {
			t.Errorf("scaffold for priority %q should include a Why stub:\n%s", priority, got)
		}
	}
}
