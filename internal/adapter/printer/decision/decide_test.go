package decision

import (
	"strings"
	"testing"

	"adg/internal/adapter/printer/printertest"
)

func TestDecided_StatusOnStderr(t *testing.T) {
	s, out, err := printertest.Capture(false)
	presenter := NewDecidePresenter(s)

	presenter.Decided("0010")

	if out.String() != "" {
		t.Errorf("stdout should be empty; got %q", out.String())
	}
	if !strings.Contains(err.String(), "Decision 0010 has been marked as decided.") {
		t.Errorf("stderr missing status: %q", err.String())
	}
}
