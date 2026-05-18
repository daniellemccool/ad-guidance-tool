package decision

import (
	"strings"
	"testing"

	"adg/internal/adapter/printer/printertest"
)

func TestLinked_WithReverseTag(t *testing.T) {
	s, out, err := printertest.Capture(false)
	presenter := NewLinkPresenter(s)

	presenter.Linked("001", "002", "depends on", "supports")

	if out.String() != "" {
		t.Errorf("stdout should be empty; got %q", out.String())
	}
	if !strings.Contains(err.String(), "Link added: 001 →[depends on]→ 002") {
		t.Errorf("stderr missing forward link: %q", err.String())
	}
	if !strings.Contains(err.String(), "Reverse link added: 002 →[supports]→ 001") {
		t.Errorf("stderr missing reverse link: %q", err.String())
	}
}

func TestLinked_WithoutReverseTag(t *testing.T) {
	s, _, err := printertest.Capture(false)
	presenter := NewLinkPresenter(s)

	presenter.Linked("003", "004", "blocks", "")

	if !strings.Contains(err.String(), "Link added: 003 →[blocks]→ 004") {
		t.Errorf("stderr missing forward link: %q", err.String())
	}
	if strings.Contains(err.String(), "Reverse link added") {
		t.Errorf("should not have reverse link; got %q", err.String())
	}
}
