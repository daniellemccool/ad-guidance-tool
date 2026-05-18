package decision

import (
	in_mocks "adg/mocks/inputport"
	svc_mocks "adg/mocks/service"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAddCommand_Success(t *testing.T) {
	mockInput := new(in_mocks.DecisionAdd)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")
	mockInput.On("Add", "resolvedPath", []string{"Test Decision"}, "").Return(nil)

	cmd := NewAddCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--title", "Test Decision"})

	err := cmd.Execute()
	assert.NoError(t, err)

	mockInput.AssertCalled(t, "Add", "resolvedPath", []string{"Test Decision"}, "")
}

func TestNewAddCommand_NoTitles(t *testing.T) {
	mockInput := new(in_mocks.DecisionAdd)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")

	cmd := NewAddCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.EqualError(t, err, "at least one --title must be provided")
}

func TestNewAddCommand_InputReturnsError(t *testing.T) {
	mockInput := new(in_mocks.DecisionAdd)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")
	mockInput.On("Add", "resolvedPath", []string{"Fail Decision"}, "").Return(errors.New("add failed"))

	cmd := NewAddCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--title", "Fail Decision"})

	err := cmd.Execute()
	assert.EqualError(t, err, "add failed")
}

func TestNewAddCommand_IDFlag_PadsAndPasses(t *testing.T) {
	mockInput := new(in_mocks.DecisionAdd)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")
	mockInput.On("Add", "resolvedPath", []string{"With ID"}, "0022").Return(nil)

	cmd := NewAddCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--title", "With ID", "--id", "22"})

	err := cmd.Execute()
	assert.NoError(t, err)
	mockInput.AssertCalled(t, "Add", "resolvedPath", []string{"With ID"}, "0022")
}

func TestNewAddCommand_IDFlag_AcceptsAlreadyPadded(t *testing.T) {
	mockInput := new(in_mocks.DecisionAdd)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")
	mockInput.On("Add", "resolvedPath", []string{"With ID"}, "0042").Return(nil)

	cmd := NewAddCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--title", "With ID", "--id", "0042"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewAddCommand_IDFlag_RejectsMultipleTitles(t *testing.T) {
	mockInput := new(in_mocks.DecisionAdd)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")

	cmd := NewAddCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--title", "A", "--title", "B", "--id", "22"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--id can only be used with a single --title")
	mockInput.AssertNotCalled(t, "Add")
}

func TestNewAddCommand_IDFlag_RejectsNonNumeric(t *testing.T) {
	mockInput := new(in_mocks.DecisionAdd)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")

	cmd := NewAddCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--title", "X", "--id", "abc"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --id")
	mockInput.AssertNotCalled(t, "Add")
}

func TestNewAddCommand_IDFlag_RejectsOutOfRange(t *testing.T) {
	mockInput := new(in_mocks.DecisionAdd)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")

	cmd := NewAddCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--title", "X", "--id", "0"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "range 1-9999")
	mockInput.AssertNotCalled(t, "Add")
}

func TestNewAddCommand_ConfigFails(t *testing.T) {
	mockInput := new(in_mocks.DecisionAdd)
	mockConfig := new(svc_mocks.ConfigService)

	// Simulate empty model path resolution
	mockConfig.On("IsLoaded").Return(false)
	mockConfig.On("GetDefaultModelPath").Return("")

	cmd := NewAddCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--title", "Should Fail"})

	err := cmd.Execute()
	assert.Error(t, err)
}
