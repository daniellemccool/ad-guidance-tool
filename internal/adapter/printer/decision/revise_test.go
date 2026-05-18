package decision

import (
	"strings"
	"testing"

	"adg/internal/adapter/printer/printertest"
)

func TestRevised_NewIDOnStdout_StatusOnStderr(t *testing.T) {
	s, out, err := printertest.Capture(false)
	presenter := NewRevisePresenter(s)

	presenter.Revised("0001", "0002")

	if got, want := out.String(), "0002\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if !strings.Contains(err.String(), "Successfully revised decision 0001 → new decision 0002") {
		t.Errorf("stderr missing status: %q", err.String())
	}
}

func TestRevised_Quiet_SuppressesStatusButNotID(t *testing.T) {
	s, out, err := printertest.Capture(true)
	presenter := NewRevisePresenter(s)

	presenter.Revised("0001", "0002")

	if out.String() != "0002\n" {
		t.Errorf("stdout = %q, want 0002\\n", out.String())
	}
	if err.String() != "" {
		t.Errorf("stderr should be empty under --quiet; got %q", err.String())
	}
}
