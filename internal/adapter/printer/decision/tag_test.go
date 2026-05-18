package decision

import (
	"strings"
	"testing"

	"adg/internal/adapter/printer/printertest"
)

func TestTagged_StatusOnStderr(t *testing.T) {
	s, out, err := printertest.Capture(false)
	presenter := NewTagPresenter(s)

	presenter.Tagged("0001", []string{"critical", "UI"})

	if out.String() != "" {
		t.Errorf("stdout should be empty; got %q", out.String())
	}
	if !strings.Contains(err.String(), "Tags [critical, UI] added to decision 0001") {
		t.Errorf("stderr missing status: %q", err.String())
	}
}
