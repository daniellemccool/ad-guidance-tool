package decision

import (
	in_mocks "adg/mocks/inputport"
	svc_mocks "adg/mocks/service"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Regression: --id accepts titles too (ResolveIdOrTitle routes non-NNNN
// values to titles), and titles can contain commas. StringSliceVar would
// CSV-split "Store::open, migrate" into two lookups, both of which would
// then fail to resolve. StringArrayVar takes the value verbatim.
func TestNewPrintCommand_TitleWithCommaIsSingleLookup(t *testing.T) {
	mockInput := new(in_mocks.DecisionPrint)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")
	mockInput.On("Print", "resolvedPath", []string(nil), []string{"Store::open, migrate"}, mock.Anything).Return(nil)

	cmd := NewPrintCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--id", "Store::open, migrate"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewPrintCommand_AllSectionsDefaulted(t *testing.T) {
	mockInput := new(in_mocks.DecisionPrint)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")
	mockInput.On("Print", "resolvedPath", []string{"0001"}, []string{"My Decision"}, map[string]bool{
		"context":  true,
		"drivers":  true,
		"options":  true,
		"outcome":  true,
		"comments": true,
	}).Return(nil)

	cmd := NewPrintCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{
		"--id", "0001", "--id", "My Decision",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewPrintCommand_CustomSectionsSelected(t *testing.T) {
	mockInput := new(in_mocks.DecisionPrint)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")
	mockInput.On("Print", "resolvedPath", []string{"0001"}, []string(nil), map[string]bool{
		"context":  true,
		"drivers":  false,
		"options":  false,
		"outcome":  false,
		"comments": false,
	}).Return(nil)

	cmd := NewPrintCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{
		"--id", "0001", "--context",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewPrintCommand_ErrorWhenNoIdsProvided(t *testing.T) {
	mockInput := new(in_mocks.DecisionPrint)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("resolvedPath")

	cmd := NewPrintCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.ErrorContains(t, err, "at least one --id or --title must be provided")
}

func TestNewPrintCommand_ModelPathResolutionFails(t *testing.T) {
	mockInput := new(in_mocks.DecisionPrint)
	mockConfig := new(svc_mocks.ConfigService)

	mockConfig.On("IsLoaded").Return(true)
	mockConfig.On("GetDefaultModelPath").Return("").Once()

	cmd := NewPrintCommand(mockInput, mockConfig)
	cmd.SetArgs([]string{"--id", "0001"})

	err := cmd.Execute()
	assert.Error(t, err)
}
