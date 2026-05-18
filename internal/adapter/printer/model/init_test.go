package model

import (
	"testing"

	"adg/internal/adapter/printer/printertest"
)

func TestInitModelPresenter_StatusOnStderr(t *testing.T) {
	s, out, err := printertest.Capture(false)
	presenter := NewInitPresenter(s)

	presenter.Initialized("test/model")

	if out.String() != "" {
		t.Errorf("stdout should be empty; got %q", out.String())
	}
	expected := "Successfully created model directory: test/model\n"
	if got := err.String(); got != expected {
		t.Errorf("stderr = %q, want %q", got, expected)
	}
}

func TestInitModelPresenter_Quiet_Suppresses(t *testing.T) {
	s, _, err := printertest.Capture(true)
	presenter := NewInitPresenter(s)
	presenter.Initialized("anywhere")
	if err.String() != "" {
		t.Errorf("stderr should be empty under --quiet; got %q", err.String())
	}
}
