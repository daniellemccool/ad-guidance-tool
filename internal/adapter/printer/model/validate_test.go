package model

import (
	"strings"
	"testing"

	"adg/internal/adapter/printer/printertest"
	"adg/internal/application/outputport"
)

func TestModelValidatePresenter_NoIssues_StatusOnStderr_NoHadIssues(t *testing.T) {
	s, out, err := printertest.Capture(false)
	presenter := NewModelValidatePresenter(s)

	presenter.ModelValidated("test-model", 6, nil)

	if out.String() != "" {
		t.Errorf("stdout should be empty; got %q", out.String())
	}
	// Success output is the "model is valid" line plus a per-check breakdown
	// so operators can see what was actually checked rather than guessing.
	stderr := err.String()
	for _, expected := range []string{
		"test-model model is valid",
		"6 ADR(s) scanned",
		"filenames match NNNN-slug.md",
		"H1 titles present",
		"required sections present",
		"status vocabulary",
		"supersession links",
		"comments well-formed",
	} {
		if !strings.Contains(stderr, expected) {
			t.Errorf("stderr missing %q; got %q", expected, stderr)
		}
	}
	if presenter.HadIssues() {
		t.Error("HadIssues should be false on clean validation")
	}
}

// --quiet must suppress the success summary entirely (it's status text).
// This locks ADR 0004's routing intent: --quiet hides status, not data.
func TestModelValidatePresenter_NoIssues_Quiet_NoOutput(t *testing.T) {
	s, out, err := printertest.Capture(true)
	presenter := NewModelValidatePresenter(s)

	presenter.ModelValidated("test-model", 6, nil)

	if out.String() != "" {
		t.Errorf("stdout should be empty; got %q", out.String())
	}
	if err.String() != "" {
		t.Errorf("stderr should be empty under --quiet on clean validation; got %q", err.String())
	}
}

func TestModelValidatePresenter_WithIssues_HadIssues_TrueAndPrintedEvenWhenQuiet(t *testing.T) {
	for _, quiet := range []bool{false, true} {
		t.Run("quiet="+boolStr(quiet), func(t *testing.T) {
			s, out, err := printertest.Capture(quiet)
			presenter := NewModelValidatePresenter(s)

			issues := []outputport.ValidationIssue{
				{ID: "0001", Message: "filename does not match NNNN-slug.md"},
				{ID: "0002", Message: "comment 1 has empty text"},
			}
			presenter.ModelValidated("test-model", 2, issues)

			if out.String() != "" {
				t.Errorf("stdout should be empty; got %q", out.String())
			}
			// Issues are errors and must print even with --quiet.
			for _, expected := range []string{
				"test-model model has 2 validation issue(s):",
				"ID 0001: filename does not match NNNN-slug.md",
				"ID 0002: comment 1 has empty text",
			} {
				if !strings.Contains(err.String(), expected) {
					t.Errorf("stderr missing %q under quiet=%v; got %q", expected, quiet, err.String())
				}
			}
			if !presenter.HadIssues() {
				t.Error("HadIssues should be true after seeing issues")
			}
		})
	}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
