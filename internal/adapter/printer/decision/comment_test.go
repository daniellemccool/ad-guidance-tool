package decision

import (
	"strings"
	"testing"

	"adg/internal/adapter/printer/printertest"
)

func TestCommented_StatusOnStderr(t *testing.T) {
	s, out, err := printertest.Capture(false)
	presenter := NewCommentPresenter(s)

	presenter.Commented("0001", "alice", "Great idea!")

	if out.String() != "" {
		t.Errorf("stdout should be empty; got %q", out.String())
	}
	if !strings.Contains(err.String(), `Comment added by alice to decision 0001: "Great idea!"`) {
		t.Errorf("stderr missing status: %q", err.String())
	}
}

func TestCommented_Quiet_Suppresses(t *testing.T) {
	s, _, err := printertest.Capture(true)
	presenter := NewCommentPresenter(s)
	presenter.Commented("0001", "alice", "x")
	if err.String() != "" {
		t.Errorf("stderr should be empty under --quiet; got %q", err.String())
	}
}
