package model

import (
	"testing"

	"adg/internal/adapter/printer/printertest"
)

func TestImportModelPresenter_StatusOnStderr(t *testing.T) {
	s, out, err := printertest.Capture(false)
	presenter := NewImportPresenter(s)

	if e := presenter.Imported("source-model", "target-model", 5); e != nil {
		t.Errorf("Imported returned error: %v", e)
	}

	if out.String() != "" {
		t.Errorf("stdout should be empty; got %q", out.String())
	}
	expected := "Successfully imported model source-model with 5 decisions to: target-model\n"
	if got := err.String(); got != expected {
		t.Errorf("stderr = %q, want %q", got, expected)
	}
}
