package model

import (
	"testing"

	"adg/internal/adapter/printer/printertest"
)

func TestCopyModelPresenter_StatusOnStderr(t *testing.T) {
	s, out, err := printertest.Capture(false)
	presenter := NewCopyPresenter(s)

	presenter.Copied("source-model", "target-model", 3)

	if out.String() != "" {
		t.Errorf("stdout should be empty; got %q", out.String())
	}
	expected := "Successfully copied 3 decisions from model source-model to new model target-model\n"
	if got := err.String(); got != expected {
		t.Errorf("stderr = %q, want %q", got, expected)
	}
}
