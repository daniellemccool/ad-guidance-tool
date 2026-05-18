package decision

import (
	"strings"
	"testing"

	"adg/internal/adapter/printer/printertest"
	"adg/internal/domain/decision"
)

func TestListed_JSON(t *testing.T) {
	s, out, _ := printertest.Capture(false)
	presenter := NewListPresenter(s)
	decisions := []decision.Decision{
		{ID: "0002", Title: "Decision B", Status: "proposed"},
		{ID: "0001", Title: "Decision A", Status: "accepted"},
	}

	presenter.Listed(decisions, "json")

	if !strings.Contains(out.String(), `"Title": "Decision A"`) || !strings.Contains(out.String(), `"Title": "Decision B"`) {
		t.Errorf("stdout JSON missing expected content:\n%s", out.String())
	}
}

func TestListed_YAML(t *testing.T) {
	s, out, _ := printertest.Capture(false)
	presenter := NewListPresenter(s)
	decisions := []decision.Decision{
		{ID: "0002", Title: "Decision B", Status: "proposed"},
	}

	presenter.Listed(decisions, "yaml")

	if !strings.Contains(out.String(), "title: Decision B") {
		t.Errorf("stdout YAML missing expected content:\n%s", out.String())
	}
}

func TestListed_Markdown(t *testing.T) {
	s, out, _ := printertest.Capture(false)
	presenter := NewListPresenter(s)
	decisions := []decision.Decision{
		{
			ID: "0001", Title: "Decision A", Status: "proposed",
			Tags:       []string{"tag1"},
			Supersedes: []string{"0002"},
			Links: map[string][]string{
				"related": {"0004"},
			},
		},
	}

	presenter.Listed(decisions, "md")

	if !strings.Contains(out.String(), "### 0001 - Decision A") || !strings.Contains(out.String(), "- **Supersedes:** 0002") || !strings.Contains(out.String(), "**Related:** 0004") {
		t.Errorf("stdout markdown missing expected content:\n%s", out.String())
	}
}

func TestListed_Simple(t *testing.T) {
	s, out, _ := printertest.Capture(false)
	presenter := NewListPresenter(s)
	decisions := []decision.Decision{
		{ID: "0001", Title: "Simple Decision", Status: "proposed", Tags: []string{"alpha", "beta"}},
	}

	presenter.Listed(decisions, "simple")

	if !strings.Contains(out.String(), "0001 [proposed] - Simple Decision : [alpha beta]") {
		t.Errorf("stdout simple format incorrect:\n%s", out.String())
	}
}

func TestListed_EmptyModel_NoticeOnStderr(t *testing.T) {
	s, out, err := printertest.Capture(false)
	presenter := NewListPresenter(s)

	presenter.Listed([]decision.Decision{}, "json")

	if out.String() != "" {
		t.Errorf("stdout should be empty on empty model; got %q", out.String())
	}
	if !strings.Contains(err.String(), "Model is empty") {
		t.Errorf("stderr missing empty-model notice; got %q", err.String())
	}
}
