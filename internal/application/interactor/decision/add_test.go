package decision

import (
	decisiondomain "adg/internal/domain/decision"
	out_mocks "adg/mocks/outputport"
	svc_mocks "adg/mocks/service"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdd_ModelDoesNotExist(t *testing.T) {
	mockModelSvc := new(svc_mocks.ModelService)
	mockDecisionSvc := new(svc_mocks.DecisionService)
	mockOutput := new(out_mocks.DecisionAdd)

	modelPath := "nonexistent"
	mockModelSvc.On("Exists", modelPath).Return(false)

	interactor := NewAddDecisionsInteractor(mockModelSvc, mockDecisionSvc, mockOutput)

	// Act/When
	err := interactor.Add(modelPath, []string{"Decision A"}, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can not add decisions")
	mockModelSvc.AssertExpectations(t)
	mockOutput.AssertNotCalled(t, "Added", mock.Anything, mock.Anything)
}

func TestAdd_AllSuccess(t *testing.T) {
	mockModelSvc := new(svc_mocks.ModelService)
	mockDecisionSvc := new(svc_mocks.DecisionService)
	mockOutput := new(out_mocks.DecisionAdd)

	modelPath := "model"
	titles := []string{"Decision A", "Decision B"}

	mockModelSvc.On("Exists", modelPath).Return(true)
	mockDecisionSvc.On("AddNew", modelPath, "Decision A", "").Return(&decisiondomain.Decision{Title: "Decision A"}, nil)
	mockDecisionSvc.On("AddNew", modelPath, "Decision B", "").Return(&decisiondomain.Decision{Title: "Decision B"}, nil)
	mockOutput.On("Added", mock.AnythingOfType("[]*madr.Decision"), mock.AnythingOfType("map[string]error")).Return()

	interactor := NewAddDecisionsInteractor(mockModelSvc, mockDecisionSvc, mockOutput)

	err := interactor.Add(modelPath, titles, "")

	assert.NoError(t, err)
	mockModelSvc.AssertExpectations(t)
	mockDecisionSvc.AssertExpectations(t)
	mockOutput.AssertExpectations(t)
}

func TestAdd_PartialFailures(t *testing.T) {
	mockModelSvc := new(svc_mocks.ModelService)
	mockDecisionSvc := new(svc_mocks.DecisionService)
	mockOutput := new(out_mocks.DecisionAdd)

	modelPath := "model"
	titles := []string{"Decision A", "Decision B"}
	addErr := errors.New("invalid title")

	mockModelSvc.On("Exists", modelPath).Return(true)
	mockDecisionSvc.On("AddNew", modelPath, "Decision A", "").Return(nil, addErr)
	mockDecisionSvc.On("AddNew", modelPath, "Decision B", "").Return(&decisiondomain.Decision{Title: "Decision B"}, nil)
	mockOutput.On("Added", mock.Anything, mock.Anything).Return()

	interactor := NewAddDecisionsInteractor(mockModelSvc, mockDecisionSvc, mockOutput)

	err := interactor.Add(modelPath, titles, "")

	assert.NoError(t, err)
	mockModelSvc.AssertExpectations(t)
	mockDecisionSvc.AssertExpectations(t)
	mockOutput.AssertExpectations(t)
}

func TestAdd_WithIDPassesThrough(t *testing.T) {
	mockModelSvc := new(svc_mocks.ModelService)
	mockDecisionSvc := new(svc_mocks.DecisionService)
	mockOutput := new(out_mocks.DecisionAdd)

	modelPath := "model"
	mockModelSvc.On("Exists", modelPath).Return(true)
	mockDecisionSvc.On("AddNew", modelPath, "With ID", "0022").Return(&decisiondomain.Decision{ID: "0022", Title: "With ID"}, nil)
	mockOutput.On("Added", mock.Anything, mock.Anything).Return()

	interactor := NewAddDecisionsInteractor(mockModelSvc, mockDecisionSvc, mockOutput)
	err := interactor.Add(modelPath, []string{"With ID"}, "0022")

	assert.NoError(t, err)
	mockDecisionSvc.AssertExpectations(t)
}

func TestAdd_WithIDPropagatesFailureAsError(t *testing.T) {
	mockModelSvc := new(svc_mocks.ModelService)
	mockDecisionSvc := new(svc_mocks.DecisionService)
	mockOutput := new(out_mocks.DecisionAdd)

	modelPath := "model"
	collisionErr := errors.New("ADR 0022 already exists at model/0022-other.md")

	mockModelSvc.On("Exists", modelPath).Return(true)
	mockDecisionSvc.On("AddNew", modelPath, "Plan-paper task", "0022").Return(nil, collisionErr)

	interactor := NewAddDecisionsInteractor(mockModelSvc, mockDecisionSvc, mockOutput)
	err := interactor.Add(modelPath, []string{"Plan-paper task"}, "0022")

	// --id failures must surface as command errors so plan-paper executors
	// can detect collisions via exit code. The batch-style failure printer
	// must NOT be called — its output would duplicate the same message
	// cobra prints from the returned error.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "0022 already exists")
	mockDecisionSvc.AssertExpectations(t)
	mockOutput.AssertNotCalled(t, "Added")
}

func TestAdd_BatchFailuresStillReturnNil(t *testing.T) {
	// Without --id, the legacy batch contract holds: per-title failures
	// are reported via output but the command returns nil. Don't regress
	// this when adding the --id-aware fail-fast behavior.
	mockModelSvc := new(svc_mocks.ModelService)
	mockDecisionSvc := new(svc_mocks.DecisionService)
	mockOutput := new(out_mocks.DecisionAdd)

	modelPath := "model"
	mockModelSvc.On("Exists", modelPath).Return(true)
	mockDecisionSvc.On("AddNew", modelPath, "A", "").Return(nil, errors.New("bad title"))
	mockDecisionSvc.On("AddNew", modelPath, "B", "").Return(&decisiondomain.Decision{Title: "B"}, nil)
	mockOutput.On("Added", mock.Anything, mock.Anything).Return()

	interactor := NewAddDecisionsInteractor(mockModelSvc, mockDecisionSvc, mockOutput)
	err := interactor.Add(modelPath, []string{"A", "B"}, "")
	assert.NoError(t, err)
}

func TestAdd_WithIDAndMultipleTitlesRejected(t *testing.T) {
	mockModelSvc := new(svc_mocks.ModelService)
	mockDecisionSvc := new(svc_mocks.DecisionService)
	mockOutput := new(out_mocks.DecisionAdd)

	modelPath := "model"
	mockModelSvc.On("Exists", modelPath).Return(true)

	interactor := NewAddDecisionsInteractor(mockModelSvc, mockDecisionSvc, mockOutput)
	err := interactor.Add(modelPath, []string{"A", "B"}, "0022")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--id can only be used with a single --title")
	mockDecisionSvc.AssertNotCalled(t, "AddNew")
	mockOutput.AssertNotCalled(t, "Added")
}
