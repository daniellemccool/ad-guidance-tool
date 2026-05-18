package decision

import (
	"fmt"
	"strings"
	"testing"

	"adg/internal/adapter/printer/printertest"
	"adg/internal/domain/decision"
)

func TestAdded_OnlySuccesses_IDsOnStdout_StatusOnStderr(t *testing.T) {
	s, out, err := printertest.Capture(false)
	presenter := NewAddPresenter(s)
	successes := []*decision.Decision{
		{ID: "0001", Title: "First"},
		{ID: "0002", Title: "Second"},
	}

	presenter.Added(successes, map[string]error{})

	// stdout: one ID per line, in input order, suitable for capture.
	if got, want := out.String(), "0001\n0002\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	// stderr: status text.
	if !strings.Contains(err.String(), "Decision First (0001) added successfully.") {
		t.Errorf("stderr missing First status: %q", err.String())
	}
	if !strings.Contains(err.String(), "Decision Second (0002) added successfully.") {
		t.Errorf("stderr missing Second status: %q", err.String())
	}
}

func TestAdded_Quiet_SuppressesStatusButNotIDs(t *testing.T) {
	s, out, err := printertest.Capture(true)
	presenter := NewAddPresenter(s)
	successes := []*decision.Decision{{ID: "0001", Title: "First"}}

	presenter.Added(successes, map[string]error{})

	if got, want := out.String(), "0001\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if err.String() != "" {
		t.Errorf("stderr should be empty under --quiet; got %q", err.String())
	}
}

func TestAdded_Failures_GoToStderr_RegardlessOfQuiet(t *testing.T) {
	for _, quiet := range []bool{false, true} {
		t.Run(fmt.Sprintf("quiet=%v", quiet), func(t *testing.T) {
			s, out, err := printertest.Capture(quiet)
			presenter := NewAddPresenter(s)
			failures := map[string]error{"Broken": fmt.Errorf("bad title")}

			presenter.Added(nil, failures)

			if out.String() != "" {
				t.Errorf("stdout should be empty when no successes; got %q", out.String())
			}
			if !strings.Contains(err.String(), `Failed to add decision "Broken": bad title`) {
				t.Errorf("stderr missing failure message under quiet=%v: %q", quiet, err.String())
			}
		})
	}
}
