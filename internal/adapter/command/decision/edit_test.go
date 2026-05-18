package decision

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	in_mocks "adg/mocks/inputport"
	svc_mocks "adg/mocks/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewEditCommand_AppendMode_Success(t *testing.T) {
	mockInput := new(in_mocks.DecisionEdit)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")
	mockInput.On("Edit", "resolvedPath", "0001", "", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	cmd := NewEditCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{
		"--id", "0001",
		"--context", "updated context paragraph",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewEditCommand_MissingFields(t *testing.T) {
	mockInput := new(in_mocks.DecisionEdit)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")

	cmd := NewEditCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{
		"--id", "0001",
	})

	err := cmd.Execute()
	assert.ErrorContains(t, err, "at least one of --context, --option, --drivers, --from-stdin, or --from-file must be provided")
}

func TestNewEditCommand_EditFails(t *testing.T) {
	mockInput := new(in_mocks.DecisionEdit)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")
	mockInput.On("Edit", "resolvedPath", "0001", "", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("edit failed"))

	cmd := NewEditCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{
		"--id", "0001",
		"--context", "some context",
	})

	err := cmd.Execute()
	assert.EqualError(t, err, "edit failed")
}

func TestNewEditCommand_FromStdin_Replace(t *testing.T) {
	mockInput := new(in_mocks.DecisionEdit)
	mockConfig := new(svc_mocks.ConfigService)

	body := "# New title\n\n## Context and Problem Statement\n\nNew content.\n"

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")
	mockInput.On("ReplaceBody", "resolvedPath", "0001", "", body, false).Return(nil)

	cmd := NewEditCommand(mockInput, mockConfig)
	cmd.SetIn(bytes.NewBufferString(body))
	cmd.SetArgs([]string{"--id", "0001", "--from-stdin"})

	err := cmd.Execute()
	assert.NoError(t, err)
	mockInput.AssertExpectations(t)
}

func TestNewEditCommand_FromFile_Replace(t *testing.T) {
	mockInput := new(in_mocks.DecisionEdit)
	mockConfig := new(svc_mocks.ConfigService)

	body := "# F\n\n## Context and Problem Statement\n\nX\n"
	tmp := filepath.Join(t.TempDir(), "body.md")
	if err := writeTempFile(tmp, body); err != nil {
		t.Fatalf("writeTempFile errored: %v", err)
	}

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")
	mockInput.On("ReplaceBody", "resolvedPath", "0001", "", body, true).Return(nil)

	cmd := NewEditCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--id", "0001", "--from-file", tmp, "--force"})

	err := cmd.Execute()
	assert.NoError(t, err)
	mockInput.AssertExpectations(t)
}

func TestNewEditCommand_StdinAndFile_Conflict(t *testing.T) {
	mockInput := new(in_mocks.DecisionEdit)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")

	cmd := NewEditCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--id", "0001", "--from-stdin", "--from-file", "/tmp/x"})

	err := cmd.Execute()
	assert.ErrorContains(t, err, "mutually exclusive")
	mockInput.AssertNotCalled(t, "ReplaceBody", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestNewEditCommand_AppendAndReplace_Conflict(t *testing.T) {
	mockInput := new(in_mocks.DecisionEdit)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")

	cmd := NewEditCommand(mockInput, mockConfig)
	cmd.SetIn(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"--id", "0001", "--from-stdin", "--context", "x"})

	err := cmd.Execute()
	assert.ErrorContains(t, err, "replace mode (--from-stdin / --from-file) cannot be combined with append flags")
}

func writeTempFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
