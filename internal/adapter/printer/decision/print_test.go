package decision

import (
	"strings"
	"testing"

	"adg/internal/adapter/printer/printertest"
	"adg/internal/application/outputport"
)

const sampleBody = `# Decision Title

## Context and Problem Statement

What should we do?

## Considered Options

* Option A
* Option B

## Decision Outcome

Chosen option: "Option A", because reasons.

## Comments

### 2026-01-01 12:00:00 — Jane

Looks good.
`

func TestPrinted_NoFilter_PrintsFullBodyToStdout(t *testing.T) {
	s, out, _ := printertest.Capture(false)
	presenter := NewPrintPresenter(s)

	bodies := []outputport.DecisionBody{
		{ID: "0002", Body: sampleBody},
		{ID: "0001", Body: sampleBody},
	}

	presenter.Printed(bodies, nil)

	for _, expected := range []string{
		"===== Decision 0001 =====",
		"===== Decision 0002 =====",
		"## Context and Problem Statement",
		"## Considered Options",
		"## Decision Outcome",
	} {
		if !strings.Contains(out.String(), expected) {
			t.Errorf("expected stdout to contain %q\n%s", expected, out.String())
		}
	}

	// IDs should be sorted ascending.
	if strings.Index(out.String(), "Decision 0001") > strings.Index(out.String(), "Decision 0002") {
		t.Errorf("expected 0001 to print before 0002:\n%s", out.String())
	}
}

func TestPrinted_FiltersBySection(t *testing.T) {
	s, out, _ := printertest.Capture(false)
	presenter := NewPrintPresenter(s)

	bodies := []outputport.DecisionBody{
		{ID: "0001", Body: sampleBody},
	}

	presenter.Printed(bodies, map[string]bool{"context": true, "outcome": true})

	if !strings.Contains(out.String(), "What should we do?") {
		t.Errorf("expected Context section, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Chosen option:") {
		t.Errorf("expected Outcome section, got:\n%s", out.String())
	}
	if strings.Contains(out.String(), "* Option A") {
		t.Errorf("Considered Options should not be printed when filtered out:\n%s", out.String())
	}
}
