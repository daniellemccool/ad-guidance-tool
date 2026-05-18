package decision

import (
	"testing"

	in_mocks "adg/mocks/inputport"
	svc_mocks "adg/mocks/service"

	"github.com/stretchr/testify/assert"
)

func TestNewSupersedeCommand_Success(t *testing.T) {
	mockInput := new(in_mocks.DecisionSupersede)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")
	mockInput.On("Supersede", "resolvedPath", "0002", "", "0001", "", "rationale").Return(nil)

	cmd := NewSupersedeCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--new", "0002", "--old", "0001", "--rationale", "rationale"})

	err := cmd.Execute()
	assert.NoError(t, err)
	mockInput.AssertExpectations(t)
}

func TestNewSupersedeCommand_MissingNew(t *testing.T) {
	mockInput := new(in_mocks.DecisionSupersede)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")

	cmd := NewSupersedeCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--old", "0001"})

	err := cmd.Execute()
	assert.ErrorContains(t, err, "both --new and --old are required")
}
