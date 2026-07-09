package commands

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveModelPath_FlagWins(t *testing.T) {
	assert.Equal(t, "/explicit/path", ResolveModelPath("/explicit/path"))
}

func TestResolveModelPath_DefaultsToDocsDecisions(t *testing.T) {
	assert.Equal(t, "docs/decisions", ResolveModelPath(""))
}

func TestModelLoadHint_NamesPathAndFlag(t *testing.T) {
	err := ModelLoadHint("docs/decisions", errors.New("open docs/decisions: no such file or directory"))
	assert.ErrorContains(t, err, `"docs/decisions"`)
	assert.ErrorContains(t, err, "--model")
}

