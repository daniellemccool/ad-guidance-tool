package model

import (
	"testing"

	"adg/internal/adapter/printer/printertest"
)

func TestMergeModelsPresenter_StatusOnStderr(t *testing.T) {
	s, out, err := printertest.Capture(false)
	presenter := NewMergePresenter(s)

	if e := presenter.Merged("modelA", "modelB", "target", 5); e != nil {
		t.Errorf("Merged returned error: %v", e)
	}

	if out.String() != "" {
		t.Errorf("stdout should be empty; got %q", out.String())
	}
	expected := "Successfully merged 5 decisions from models modelA and modelB to new directory: target\n"
	if got := err.String(); got != expected {
		t.Errorf("stderr = %q, want %q", got, expected)
	}
}
