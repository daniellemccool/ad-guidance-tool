package lean

import (
	"adg/internal/domain/decision/madr"
	"strings"
	"testing"
)

func rec(id, filename, status, category string, body string, supersedes, amends []string) Record {
	return Record{
		ID:       id,
		Filename: filename,
		Body:     body,
		D:        madr.Decision{Status: status, Category: category, Supersedes: supersedes, Amends: amends},
	}
}

func acceptedBody(title string) string {
	return "# " + title + "\n\n## Decision\n\nWe do X.\n\n## Implication\n\nNew code must do Y.\n"
}

func TestParseBody_SectionsAndTitle(t *testing.T) {
	p := ParseBody(acceptedBody("Use X"))
	if p.Title != "Use X" {
		t.Errorf("title = %q, want %q", p.Title, "Use X")
	}
	if got := p.Sections["decision"]; got != "We do X." {
		t.Errorf("decision = %q", got)
	}
	if got := p.Sections["implication"]; got != "New code must do Y." {
		t.Errorf("implication = %q", got)
	}
}

func TestValidate_AcceptedHappyPath(t *testing.T) {
	issues := Validate([]Record{rec("0001", "0001-x.md", "accepted", "Meta", acceptedBody("Use X"), nil, nil)})
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got: %+v", issues)
	}
}

func TestValidate_AcceptedMissingImplicationAndToken(t *testing.T) {
	body := "# T\n\n## Decision\n\n{...}\n"
	issues := Validate([]Record{rec("0001", "0001-t.md", "accepted", "Meta", body, nil, nil)})
	var msgs []string
	for _, i := range issues {
		msgs = append(msgs, i.Message)
	}
	joined := strings.Join(msgs, "\n")
	for _, want := range []string{"Guidance", "template placeholder"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected an issue mentioning %q; got:\n%s", want, joined)
		}
	}
}

func TestValidate_ProposedDraftExemptFromBodyChecks(t *testing.T) {
	body := "# T\n\n## Decision\n\n{...}\n\n## Implication\n\n{...}\n"
	issues := Validate([]Record{rec("0001", "0001-t.md", "proposed", "", body, nil, nil)})
	for _, i := range issues {
		if !i.Warning {
			t.Errorf("proposed draft should not hard-fail; got: %+v", i)
		}
	}
}

func TestValidate_BadStatusVocabulary(t *testing.T) {
	issues := Validate([]Record{rec("0001", "0001-t.md", "Superseded by ADR 0002", "Meta", acceptedBody("T"), nil, nil)})
	found := false
	for _, i := range issues {
		if strings.Contains(i.Message, "not valid lean vocabulary") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected invalid-vocabulary issue; got: %+v", issues)
	}
}

func TestValidate_SupersessionIntegrity(t *testing.T) {
	// 0001 says it is superseded by 0002, and 0002 correctly lists 0001.
	old := rec("0001", "0001-old.md", "superseded by ADR-0002", "Meta", acceptedBody("Old"), nil, nil)
	newOK := rec("0002", "0002-new.md", "accepted", "Meta", acceptedBody("New"), []string{"0001"}, nil)
	if issues := Validate([]Record{old, newOK}); len(issues) != 0 {
		t.Fatalf("consistent supersession should pass; got: %+v", issues)
	}

	// Now break the reverse link: 0002 forgets to list 0001.
	newBad := rec("0002", "0002-new.md", "accepted", "Meta", acceptedBody("New"), nil, nil)
	issues := Validate([]Record{old, newBad})
	found := false
	for _, i := range issues {
		if strings.Contains(i.Message, "supersedes list does not include") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected supersession reverse-link issue; got: %+v", issues)
	}
}

func TestValidate_AmendmentIntegrity(t *testing.T) {
	base := rec("0001", "0001-base.md", "amended by ADR-0002", "Meta", acceptedBody("Base"), nil, nil)
	amender := rec("0002", "0002-amend.md", "accepted", "Meta", acceptedBody("Amend"), nil, []string{"0001"})
	if issues := Validate([]Record{base, amender}); len(issues) != 0 {
		t.Fatalf("consistent amendment should pass; got: %+v", issues)
	}

	amenderBad := rec("0002", "0002-amend.md", "accepted", "Meta", acceptedBody("Amend"), nil, nil)
	issues := Validate([]Record{base, amenderBad})
	found := false
	for _, i := range issues {
		if strings.Contains(i.Message, "amends list does not include") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected amendment reverse-link issue; got: %+v", issues)
	}
}

func TestValidate_LengthWarning(t *testing.T) {
	var b strings.Builder
	b.WriteString("# T\n\n## Decision\n\nx\n\n## Implication\n\n")
	for i := 0; i < MaxBodyLines+5; i++ {
		b.WriteString("* line\n")
	}
	issues := Validate([]Record{rec("0001", "0001-t.md", "accepted", "Meta", b.String(), nil, nil)})
	found := false
	for _, i := range issues {
		if i.Warning && strings.Contains(i.Message, "one screen") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected one-screen length warning; got: %+v", issues)
	}
}

func TestRenderIndex_GroupingAndAnnotations(t *testing.T) {
	records := []Record{
		rec("0001", "0001-framework.md", "accepted", "Meta", acceptedBody("ADR framework"), nil, nil),
		rec("0009", "0009-cqrs.md", "accepted", "CQRS + event sourcing", acceptedBody("CQRS + event sourcing"), nil, nil),
		rec("0014", "0014-audit.md", "superseded by ADR-0039", "CQRS + event sourcing", acceptedBody("Audit: projection-only"), nil, nil),
		rec("0039", "0039-audit2.md", "accepted", "Data layer", acceptedBody("Audit two-channel"), []string{"0014"}, nil),
		rec("0002", "0002-privacy.md", "amended by ADR-0054", "Invariants", acceptedBody("Privacy invariant"), nil, nil),
		rec("0054", "0054-direct.md", "accepted", "Implementation strategy", acceptedBody("Direct delivery"), nil, []string{"0002"}),
	}
	out := RenderIndex(records)

	// Meta (0001) must precede CQRS (min 0009) which must precede Data layer (0039).
	iMeta := strings.Index(out, "### Meta")
	iCQRS := strings.Index(out, "### CQRS + event sourcing")
	iData := strings.Index(out, "### Data layer")
	if !(iMeta >= 0 && iMeta < iCQRS && iCQRS < iData) {
		t.Errorf("group order wrong: meta=%d cqrs=%d data=%d\n%s", iMeta, iCQRS, iData, out)
	}
	if !strings.Contains(out, "~~[0014 — Audit: projection-only](./0014-audit.md)~~ — *superseded by ADR 0039*") {
		t.Errorf("superseded entry not struck through/annotated:\n%s", out)
	}
	if !strings.Contains(out, "*(amended by ADR 0054)*") {
		t.Errorf("amended entry not annotated:\n%s", out)
	}
}

// LoadDir + Validate end-to-end against the relocated broken example: an
// accepted record missing Guidance, carrying a leftover {...} token, and with no
// category must produce the corresponding failures and warning.
func TestLoadDir_BrokenExampleFailsValidation(t *testing.T) {
	records, err := LoadDir("testdata/broken")
	if err != nil {
		t.Fatalf("LoadDir errored: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	var msgs []string
	for _, i := range Validate(records) {
		msgs = append(msgs, i.Message)
	}
	joined := strings.Join(msgs, "\n")
	for _, want := range []string{"required section: Guidance", "template placeholder", "no category"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected a finding mentioning %q; got:\n%s", want, joined)
		}
	}
}

func TestRenderNewBody_ContainsCoreSections(t *testing.T) {
	body := RenderNewBody("Pick a thing")
	for _, want := range []string{"# Pick a thing", "## Decision", "## Guidance", "{...}"} {
		if !strings.Contains(body, want) {
			t.Errorf("template missing %q:\n%s", want, body)
		}
	}
}
