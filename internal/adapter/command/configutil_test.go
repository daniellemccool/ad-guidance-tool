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

func TestResolveIdOrTitle_ValidID(t *testing.T) {
	var id, title string
	err := ResolveIdOrTitle("0001", &id, &title)
	assert.NoError(t, err)
	assert.Equal(t, "0001", id)
	assert.Empty(t, title)
}

func TestResolveIdOrTitle_ValidTitle(t *testing.T) {
	var id, title string
	err := ResolveIdOrTitle("my-decision", &id, &title)
	assert.NoError(t, err)
	assert.Equal(t, "my-decision", title)
	assert.Empty(t, id)
}

func TestResolveIdOrTitle_EmptyInput(t *testing.T) {
	var id, title string
	err := ResolveIdOrTitle("", &id, &title)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you must specify the decisions via --id")
}

func TestResolveIdOrTitle_ShortIDZeroPadded(t *testing.T) {
	var id, title string
	err := ResolveIdOrTitle("1", &id, &title)
	assert.NoError(t, err)
	assert.Equal(t, "0001", id)
	assert.Empty(t, title)
}

func TestResolveIdOrTitle_OutOfRangeID(t *testing.T) {
	var id, title string
	err := ResolveIdOrTitle("0", &id, &title)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "range 1-9999")
}

func TestResolveIdOrTitle_InvalidFormat(t *testing.T) {
	var id, title string
	err := ResolveIdOrTitle("####", &id, &title)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input must be either an ID")
}

func TestGetTemplateSections_Nygard(t *testing.T) {
	sections, err := GetTemplateSections("Nygard")
	assert.NoError(t, err)
	assert.Equal(t, "Context", sections["question"])
	assert.Equal(t, "Consequences", sections["criteria"])
	assert.Equal(t, "Decision", sections["outcome"])
}

func TestGetTemplateSections_MADR(t *testing.T) {
	sections, err := GetTemplateSections("madr")
	assert.NoError(t, err)
	assert.Equal(t, "Context and Problem Statement", sections["question"])
	assert.Equal(t, "Considered Options", sections["options"])
	assert.Equal(t, "Decision Drivers", sections["criteria"])
	assert.Equal(t, "Decision Outcome", sections["outcome"])
}

func TestGetTemplateSections_UnknownTemplate(t *testing.T) {
	sections, err := GetTemplateSections("random")
	assert.Error(t, err)
	assert.Nil(t, sections)
	assert.Contains(t, err.Error(), "unknown template")
}
