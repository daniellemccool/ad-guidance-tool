package decision

import (
	"adg/internal/adapter/printer/printertest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlugCommand_PrintsSlugToStdout(t *testing.T) {
	s, out, _ := printertest.Capture(false)
	cmd := NewSlugCommand(s)
	cmd.SetArgs([]string{"My Decision Title"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "my-decision-title\n", out.String())
}

// Per ADR 0004, machine values go on stdout (so plan briefs can capture
// the slug via $() in shell). Stderr stays clean on success.
func TestSlugCommand_StderrEmptyOnSuccess(t *testing.T) {
	s, _, errBuf := printertest.Capture(false)
	cmd := NewSlugCommand(s)
	cmd.SetArgs([]string{"Anything"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Empty(t, errBuf.String())
}

// The Epic 2 feedback example: the predicted slug should be exactly what
// `adg add` would write, including the small words "is" and "via" that
// some stopword-stripping conventions drop.
func TestSlugCommand_MatchesAddCommandSlugForReportedRegression(t *testing.T) {
	s, out, _ := printertest.Capture(false)
	cmd := NewSlugCommand(s)
	cmd.SetArgs([]string{"Bug class supervision JoinSet CancellationToken shutdown order is load-bearing"})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t,
		"bug-class-supervision-joinset-cancellationtoken-shutdown-order-is-load-bearing\n",
		out.String())
}

func TestSlugCommand_RejectsEmptyTitle(t *testing.T) {
	s, out, _ := printertest.Capture(false)
	cmd := NewSlugCommand(s)
	cmd.SetArgs([]string{"???"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "slugifies to empty"))
	assert.Empty(t, out.String())
}

func TestSlugCommand_RequiresExactlyOneArg(t *testing.T) {
	s, _, _ := printertest.Capture(false)
	cmd := NewSlugCommand(s)
	cmd.SetArgs([]string{})
	cmd.SilenceErrors = true
	assert.Error(t, cmd.Execute())

	cmd = NewSlugCommand(s)
	cmd.SetArgs([]string{"a", "b"})
	cmd.SilenceErrors = true
	assert.Error(t, cmd.Execute())
}
