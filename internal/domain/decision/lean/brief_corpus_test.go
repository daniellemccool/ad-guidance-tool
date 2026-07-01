package lean

import (
	"strings"
	"testing"
)

func superseded(r Record) Record { r.D.Status = "superseded by ADR-0099"; return r }

func TestBrief_OmitsChangedPathsHeader(t *testing.T) {
	r := briefRec("0004", "0004-pii.md", "invariant", []string{"**/*.py"},
		"# PII\n\n## Decision\n\nx\n\n## Guidance\n\n- y\n")
	out := Brief([]Record{r}, []string{"port/x.py"}, BriefFull)
	if strings.Contains(out, "Changed paths") {
		t.Errorf("the routed brief should no longer print a changed-paths header:\n%s", out)
	}
}

func TestBriefWhole_InvariantsFullDefaultsCondensedNoHeaderNoFooter(t *testing.T) {
	records := []Record{
		briefRec("0001", "0001-x.md", "invariant", []string{"port/**/*.py"},
			"# Streaming\n\n## Decision\n\nStream uploads.\n\n## Guidance\n\n- Pass the adapter.\n"),
		briefRec("0002", "0002-x.md", "default", []string{"api/**/*.py"},
			"# Naming\n\n## Decision\n\nsnake_case.\n\n## Guidance\n\n- Use snake_case.\n"),
		superseded(briefRec("0003", "0003-x.md", "default", []string{"**/*.py"}, "# Old\n\n## Decision\n\nx\n")),
	}
	out := BriefWhole(records)

	for _, banned := range []string{"Changed paths", "## Before you finish", "Related files to consider", "matched:"} {
		if strings.Contains(out, banned) {
			t.Errorf("whole brief must not contain %q:\n%s", banned, out)
		}
	}
	if !strings.Contains(out, "working agreements that govern this codebase") {
		t.Errorf("expected the convention preamble:\n%s", out)
	}
	// Invariant rendered full; default collapsed to a checklist line.
	if !strings.Contains(out, "### ADR-0001") || !strings.Contains(out, "**Decision:**") {
		t.Errorf("the invariant should render in full:\n%s", out)
	}
	if !strings.Contains(out, "- ADR-0002 Naming: Use snake_case. → 0002-x.md") {
		t.Errorf("a default should render as a condensed checklist line:\n%s", out)
	}
	// Unrouted entries show their declared applies_to scope.
	if !strings.Contains(out, "_scope: port/**/*.py_") {
		t.Errorf("expected an unrouted scope line for the invariant:\n%s", out)
	}
	// Terminal record excluded.
	if strings.Contains(out, "ADR-0003") {
		t.Errorf("a terminal record must not appear in the whole brief:\n%s", out)
	}
}

func TestBriefWhole_EmptyWhenNoInForce(t *testing.T) {
	r := briefRec("0001", "0001-x.md", "default", []string{"**/*.py"}, "# X\n\n## Decision\n\nx\n")
	r.D.Status = "deprecated"
	if out := BriefWhole([]Record{r}); out != "" {
		t.Errorf("whole brief with no in-force records should be empty, got:\n%s", out)
	}
}

func TestBriefInvariants_OnlyInvariants(t *testing.T) {
	records := []Record{
		briefRec("0001", "0001-x.md", "invariant", []string{"port/**/*.py"},
			"# Streaming\n\n## Decision\n\nStream.\n\n## Guidance\n\n- Pass the adapter.\n"),
		briefRec("0002", "0002-x.md", "default", []string{"api/**/*.py"},
			"# Naming\n\n## Decision\n\nsnake_case.\n\n## Guidance\n\n- snake.\n"),
	}
	out := BriefInvariants(records)
	if !strings.Contains(out, "### ADR-0001") || !strings.Contains(out, "**Decision:**") {
		t.Errorf("expected the invariant rendered full:\n%s", out)
	}
	for _, banned := range []string{"ADR-0002", "## Defaults & conventions", "## Before you finish", "Related files to consider"} {
		if strings.Contains(out, banned) {
			t.Errorf("invariants-only brief must not contain %q:\n%s", banned, out)
		}
	}
}

func TestBriefInvariants_EmptyWhenNoInvariants(t *testing.T) {
	r := briefRec("0002", "0002-x.md", "default", []string{"**/*.py"}, "# X\n\n## Decision\n\nx\n")
	if out := BriefInvariants([]Record{r}); out != "" {
		t.Errorf("no invariants should yield empty, got:\n%s", out)
	}
}

func TestForbidden_ReturnsViolatorsAndHonorsLifecycle(t *testing.T) {
	forbid := briefRecX("0011", "0011-x.md", "invariant", nil, nil, []string{"port/extraction/**/*.py"}, nil, "")
	other := briefRec("0002", "0002-x.md", "default", []string{"port/**/*.py"}, "")
	got := Forbidden([]Record{forbid, other}, []string{"port/extraction/new.py"})
	if len(got) != 1 || got[0].ID != "0011" {
		t.Errorf("Forbidden = %+v, want just 0011", got)
	}
	term := forbid
	term.D.Status = "superseded by ADR-0012"
	if len(Forbidden([]Record{term}, []string{"port/extraction/new.py"})) != 0 {
		t.Error("a superseded forbids record must not be reported as a violation")
	}
}
