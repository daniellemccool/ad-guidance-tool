package decision

import (
	"strings"
	"testing"

	"adg/internal/adapter/printer/printertest"
)

func TestEdited_StatusOnStderr(t *testing.T) {
	s, out, err := printertest.Capture(false)
	presenter := NewEditPresenter(s)

	presenter.Edited("0077")

	if out.String() != "" {
		t.Errorf("stdout should be empty; got %q", out.String())
	}
	if !strings.Contains(err.String(), "Decision 0077 updated successfully.") {
		t.Errorf("stderr missing status: %q", err.String())
	}
}
