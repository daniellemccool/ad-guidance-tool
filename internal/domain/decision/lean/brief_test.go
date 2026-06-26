package lean

import (
	"adg/internal/domain/decision/madr"
	"strings"
	"testing"
)

func briefRec(id, filename, priority string, appliesTo []string, body string) Record {
	return Record{
		ID:       id,
		Filename: filename,
		Body:     body,
		D:        madr.Decision{Status: "accepted", Priority: priority, AppliesTo: appliesTo},
	}
}

func TestBrief_RoutesGroupsAndChecks(t *testing.T) {
	records := []Record{
		briefRec("0001", "0001-meta.md", "default", nil,
			"# Meta\n\n## Decision\n\nProcess only.\n\n## Implication\n\nN/A.\n"),
		briefRec("0002", "0002-imports.md", "default", []string{"port/**/*.py"},
			"# No cross-layer private imports\n\n## Decision\n\nNo underscore imports across layers.\n\n## Implication\n\n- Review rejects `from x import _y` across layers.\n"),
		briefRec("0003", "0003-uploads.md", "default", []string{"**/flow_builder.py"},
			"# Reject unsafe uploads\n\n## Decision\n\nSafety before validation.\n\n## Guidance\n\n- Call check_file_safety first.\n\n## Checks\n\n- Confirm safety runs before extraction.\n"),
		briefRec("0004", "0004-pii.md", "invariant", []string{"**/*.py"},
			"# PII safety boundary\n\n## Decision\n\nCatch all exceptions at the boundary.\n\n## Implication\n\n- Do not remove the handler.\n\n## Checks\n\n- grep for new bridge.sendLogs paths.\n"),
	}

	out := Brief(records, []string{"port/helpers/flow_builder.py"})

	// 0001 (no applies_to) must not route.
	if strings.Contains(out, "ADR-0001") {
		t.Errorf("meta ADR with no applies_to should not route:\n%s", out)
	}
	// Invariant section precedes defaults; 0004 is the invariant.
	iHard := strings.Index(out, "## Hard constraints")
	iDef := strings.Index(out, "## Defaults & conventions")
	i4 := strings.Index(out, "ADR-0004")
	i2 := strings.Index(out, "ADR-0002")
	if !(iHard >= 0 && iHard < iDef) {
		t.Errorf("invariants section should precede defaults:\n%s", out)
	}
	if !(iHard < i4 && i4 < iDef && iDef < i2) {
		t.Errorf("0004 should be under invariants, 0002 under defaults:\n%s", out)
	}
	// Matched-pattern explainability.
	if !strings.Contains(out, "matched: **/flow_builder.py") {
		t.Errorf("expected matched-pattern annotation for 0003:\n%s", out)
	}
	// Guidance synonym: 0002 uses ## Implication, surfaced as Guidance bullets.
	if !strings.Contains(out, "Review rejects") {
		t.Errorf("expected 0002 Implication surfaced as guidance:\n%s", out)
	}
	// Consolidated checks from 0003 and 0004.
	if !strings.Contains(out, "## Checks to run") ||
		!strings.Contains(out, "(ADR-0003) Confirm safety runs before extraction.") ||
		!strings.Contains(out, "(ADR-0004) grep for new bridge.sendLogs paths.") {
		t.Errorf("expected consolidated checks from matched ADRs:\n%s", out)
	}
}

func TestBrief_PathFiltersOutNonMatches(t *testing.T) {
	records := []Record{
		briefRec("0003", "0003-uploads.md", "default", []string{"**/flow_builder.py"},
			"# Uploads\n\n## Decision\n\nx\n\n## Implication\n\n- y\n"),
		briefRec("0004", "0004-pii.md", "invariant", []string{"**/*.py"},
			"# PII\n\n## Decision\n\nx\n\n## Implication\n\n- y\n"),
	}
	out := Brief(records, []string{"port/script.py"})
	if strings.Contains(out, "ADR-0003") {
		t.Errorf("0003 should not match port/script.py:\n%s", out)
	}
	if !strings.Contains(out, "ADR-0004") {
		t.Errorf("0004 (**/*.py) should match port/script.py:\n%s", out)
	}
}

func TestBulletLines_FoldsWrappedContinuations(t *testing.T) {
	section := "## Implication\n\n- first bullet that wraps\n  onto a second line\n- second bullet\n"
	got := bulletLines(section)
	want := []string{"first bullet that wraps onto a second line", "second bullet"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("bulletLines folding wrong:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestBrief_NoMatches(t *testing.T) {
	records := []Record{
		briefRec("0004", "0004-pii.md", "invariant", []string{"**/*.py"},
			"# PII\n\n## Decision\n\nx\n\n## Implication\n\n- y\n"),
	}
	out := Brief(records, []string{"docs/notes.md"})
	if !strings.Contains(out, "No ADRs match these paths") {
		t.Errorf("expected no-match message:\n%s", out)
	}
}
