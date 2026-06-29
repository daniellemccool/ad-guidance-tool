package lean

import (
	"adg/internal/domain/decision/madr"
	"strings"
	"testing"
)

func briefRec(id, filename, priority string, appliesTo []string, body string) Record {
	return briefRecX(id, filename, priority, appliesTo, nil, nil, nil, body)
}

func briefRecX(id, filename, priority string, appliesTo, excludes, forbids, companions []string, body string) Record {
	return Record{
		ID:       id,
		Filename: filename,
		Body:     body,
		D: madr.Decision{
			Status: "accepted", Priority: priority,
			AppliesTo: appliesTo, Excludes: excludes, Forbids: forbids, Companions: companions,
		},
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

func TestRouteMatch_ExcludesSuppressesGovernedPath(t *testing.T) {
	// 0008: govern helpers under port, but the sanctioned construction home
	// port_helpers.py must not fire (the real worked-around case).
	r := briefRecX("0008", "0008-x.md", "default",
		[]string{"port/**/*.py"}, []string{"**/port_helpers.py"}, nil, nil, "")
	got := routeMatch(r, []string{"port/helpers/flow_builder.py", "port/helpers/port_helpers.py"})
	if len(got.matched) != 1 || got.matched[0] != "port/**/*.py" {
		t.Errorf("matched = %v, want [port/**/*.py]", got.matched)
	}
	if len(got.excluded) != 1 || got.excluded[0] != "**/port_helpers.py" {
		t.Errorf("excluded = %v, want [**/port_helpers.py]", got.excluded)
	}
	if len(got.governed) != 1 || got.governed[0] != "port/helpers/flow_builder.py" {
		t.Errorf("governed = %v, want [port/helpers/flow_builder.py]", got.governed)
	}
}

func TestRouteMatch_DedupAcrossPaths(t *testing.T) {
	r := briefRecX("0002", "0002-x.md", "default", []string{"port/**/*.py"}, nil, nil, nil, "")
	got := routeMatch(r, []string{"port/a.py", "port/b.py"})
	if len(got.matched) != 1 {
		t.Errorf("matched should dedup to one pattern; got %v", got.matched)
	}
	if len(got.governed) != 2 {
		t.Errorf("expected both paths governed; got %v", got.governed)
	}
}

func TestRouteMatch_ExcludesOnNonGovernedPathNotReported(t *testing.T) {
	r := briefRecX("0008", "0008-x.md", "default",
		[]string{"port/**/*.py"}, []string{"docs/**"}, nil, nil, "")
	got := routeMatch(r, []string{"docs/x.md"})
	if len(got.matched) != 0 || len(got.excluded) != 0 || len(got.governed) != 0 {
		t.Errorf("excludes must not be reported on a path applies_to never governs; got %+v", got)
	}
}

func TestRouteMatch_ForbidsIndependentOfApplies(t *testing.T) {
	r := briefRecX("0011", "0011-x.md", "invariant",
		nil, nil, []string{"port/extraction/**/*.py"}, nil, "")
	got := routeMatch(r, []string{"port/extraction/new.py"})
	if len(got.forbidden) != 1 || got.forbidden[0] != "port/extraction/**/*.py" {
		t.Errorf("forbidden = %v, want [port/extraction/**/*.py]", got.forbidden)
	}
	if len(got.matched) != 0 {
		t.Errorf("no applies_to, so matched should be empty; got %v", got.matched)
	}
}

func TestBrief_ExcludedAnnotation(t *testing.T) {
	r := briefRecX("0008", "0008-x.md", "default",
		[]string{"port/**/*.py"}, []string{"**/port_helpers.py"}, nil, nil,
		"# Pages\n\n## Decision\n\nBuild via helpers.\n\n## Guidance\n\n- x\n")
	out := Brief([]Record{r}, []string{"port/helpers/flow_builder.py", "port/helpers/port_helpers.py"})
	if !strings.Contains(out, "excluded: **/port_helpers.py") {
		t.Errorf("expected excluded annotation in scope line:\n%s", out)
	}
}

func TestBrief_ForbiddenWarningRendered(t *testing.T) {
	r := briefRecX("0011", "0011-x.md", "invariant",
		nil, nil, []string{"port/extraction/**/*.py"}, nil,
		"# Single architecture\n\n## Decision\n\nOne extraction path.\n\n## Guidance\n\n- x\n")
	out := Brief([]Record{r}, []string{"port/extraction/new.py"})
	if !strings.Contains(out, "Forbidden scope matched") || !strings.Contains(out, "ADR-0011") {
		t.Errorf("expected forbidden-scope warning:\n%s", out)
	}
}

func TestBrief_ForbiddenOnlyNoEmptyMatched(t *testing.T) {
	r := briefRecX("0011", "0011-x.md", "invariant",
		nil, nil, []string{"port/extraction/**/*.py"}, nil,
		"# Single architecture\n\n## Decision\n\nx\n\n## Guidance\n\n- x\n")
	out := Brief([]Record{r}, []string{"port/extraction/new.py"})
	// The scope-line annotation (`· matched: …`) must be absent for a forbids-only
	// hit; "Forbidden scope matched:" is the separate warning line and is expected.
	if strings.Contains(out, "· matched:") {
		t.Errorf("forbidden-only entry must not render an empty matched: annotation:\n%s", out)
	}
	if !strings.Contains(out, "scope: forbids port/extraction/**/*.py") {
		t.Errorf("expected a forbids scope line:\n%s", out)
	}
}

func TestBrief_CompanionsRenderedWithSoftNote(t *testing.T) {
	r := briefRecX("0009", "0009-x.md", "default",
		[]string{"api/d3i_props.py"}, nil, nil, []string{"packages/data-collector/src/App.tsx"},
		"# Props\n\n## Decision\n\nx\n\n## Guidance\n\n- x\n")
	// Touch the governed file but not the companion -> soft note appears.
	out := Brief([]Record{r}, []string{"api/d3i_props.py"})
	if !strings.Contains(out, "Related files to consider:") ||
		!strings.Contains(out, "packages/data-collector/src/App.tsx") {
		t.Errorf("expected companions line:\n%s", out)
	}
	if !strings.Contains(out, "none of the changed paths are among these") {
		t.Errorf("expected soft note when the companion is untouched:\n%s", out)
	}
	// Touch the companion too -> no soft note.
	out2 := Brief([]Record{r}, []string{"api/d3i_props.py", "packages/data-collector/src/App.tsx"})
	if strings.Contains(out2, "none of the changed paths are among these") {
		t.Errorf("did not expect the soft note when the companion is touched:\n%s", out2)
	}
}

func TestBrief_CompanionsOnlyDoesNotRoute(t *testing.T) {
	r := briefRecX("0009", "0009-x.md", "default",
		nil, nil, nil, []string{"src/App.tsx"},
		"# Props\n\n## Decision\n\nx\n\n## Guidance\n\n- x\n")
	out := Brief([]Record{r}, []string{"src/App.tsx"})
	if !strings.Contains(out, "No ADRs match these paths") {
		t.Errorf("a companions-only record must never route:\n%s", out)
	}
}

func TestMatches_ForbiddenTriggersGate(t *testing.T) {
	r := briefRecX("0011", "0011-x.md", "invariant",
		nil, nil, []string{"port/extraction/**/*.py"}, nil, "")
	if !Matches([]Record{r}, []string{"port/extraction/new.py"}) {
		t.Errorf("a forbidden-path edit should trigger the hook gate")
	}
	if Matches([]Record{r}, []string{"port/other.py"}) {
		t.Errorf("a non-matching path should not trigger the gate")
	}
}
