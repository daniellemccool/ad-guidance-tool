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

	presenter.ModelValidated("test-model", nil)

	if out.String() != "" {
		t.Errorf("stdout should be empty; got %q", out.String())
	}
	if got, want := err.String(), "test-model model is valid\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
	if presenter.HadIssues() {
		t.Error("HadIssues should be false on clean validation")
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
			presenter.ModelValidated("test-model", issues)

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
