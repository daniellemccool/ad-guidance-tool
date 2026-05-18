package model

import (
	"errors"
	"testing"

	in_mocks "adg/mocks/inputport"
	svc_mocks "adg/mocks/service"

	"github.com/stretchr/testify/assert"
)

// hadIssuesStub returns a function suitable for the NewValidateCommand's
// hadIssues callback. It returns the boolean value directly so tests can
// simulate "presenter saw issues" without a real presenter.
func hadIssuesStub(v bool) func() bool {
	return func() bool { return v }
}

func TestNewValidateCommand_NoIssues_NoError(t *testing.T) {
	mockInput := new(in_mocks.ModelValidate)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedModelPath")
	mockInput.On("Validate", "resolvedModelPath").Return(nil)

	cmd := NewValidateCommand(mockInput, mockConfig, hadIssuesStub(false))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
	mockInput.AssertCalled(t, "Validate", "resolvedModelPath")
}

func TestNewValidateCommand_WithIssues_ReturnsSentinelError(t *testing.T) {
	mockInput := new(in_mocks.ModelValidate)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedModelPath")
	mockInput.On("Validate", "resolvedModelPath").Return(nil)

	cmd := NewValidateCommand(mockInput, mockConfig, hadIssuesStub(true))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.ErrorIs(t, err, ErrValidationIssues)
}

func TestNewValidateCommand_ConfigError(t *testing.T) {
	mockInput := new(in_mocks.ModelValidate)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("").Once()

	cmd := NewValidateCommand(mockInput, mockConfig, hadIssuesStub(false))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestNewValidateCommand_InputReturnsError(t *testing.T) {
	mockInput := new(in_mocks.ModelValidate)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedModelPath")
	mockInput.On("Validate", "resolvedModelPath").Return(errors.New("validation failed"))

	cmd := NewValidateCommand(mockInput, mockConfig, hadIssuesStub(false))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.EqualError(t, err, "validation failed")
}
