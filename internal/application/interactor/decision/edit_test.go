package decision

import (
	"adg/internal/domain/decision"
	out_mocks "adg/mocks/outputport"
	svc_mocks "adg/mocks/service"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEdit_ByID_Success(t *testing.T) {
	mockSvc := new(svc_mocks.DecisionService)
	mockOut := new(out_mocks.DecisionEdit)

	d := &decision.Decision{ID: "0010"}
	ctx := "What is the impact?"
	opts := []string{"Option 1", "Option 2"}
	drv := "Speed, Cost"

	mockSvc.On("GetDecisionByID", "model", "0010").Return(d, nil)
	mockSvc.On("Edit", "model", d, &ctx, &opts, &drv).Return(nil)
	mockOut.On("Edited", "0010").Return(nil)

	interactor := NewEditDecisionInteractor(mockSvc, mockOut)
	err := interactor.Edit("model", "0010", "", &ctx, &opts, &drv)

	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
	mockOut.AssertExpectations(t)
}

func TestEdit_ByTitle_Success(t *testing.T) {
	mockSvc := new(svc_mocks.DecisionService)
	mockOut := new(out_mocks.DecisionEdit)

	d := &decision.Decision{ID: "0020"}
	ctx := "Change context?"
	opts := []string{"Yes", "No"}

	mockSvc.On("GetDecisionByTitle", "model", "Decide Feature").Return(d, nil)
	mockSvc.On("Edit", "model", d, &ctx, &opts, (*string)(nil)).Return(nil)
	mockOut.On("Edited", "0020").Return(nil)

	interactor := NewEditDecisionInteractor(mockSvc, mockOut)
	err := interactor.Edit("model", "", "Decide Feature", &ctx, &opts, nil)

	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
	mockOut.AssertExpectations(t)
}

func TestEdit_GetDecisionFails(t *testing.T) {
	mockSvc := new(svc_mocks.DecisionService)
	mockOut := new(out_mocks.DecisionEdit)

	mockSvc.On("GetDecisionByID", "model", "9999").Return(nil, errors.New("not found"))

	interactor := NewEditDecisionInteractor(mockSvc, mockOut)
	err := interactor.Edit("model", "9999", "", nil, nil, nil)

	assert.ErrorContains(t, err, "not found")
	mockSvc.AssertExpectations(t)
}

func TestReplaceBody_ByID_Success(t *testing.T) {
	mockSvc := new(svc_mocks.DecisionService)
	mockOut := new(out_mocks.DecisionEdit)

	d := &decision.Decision{ID: "0010", Status: "proposed"}
	body := "# T\n\n## Context and Problem Statement\n\nC\n"

	mockSvc.On("GetDecisionByID", "model", "0010").Return(d, nil)
	mockSvc.On("ReplaceBody", "model", d, body, false).Return(nil)
	mockOut.On("Edited", "0010").Return(nil)

	interactor := NewEditDecisionInteractor(mockSvc, mockOut)
	err := interactor.ReplaceBody("model", "0010", "", body, false)

	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
	mockOut.AssertExpectations(t)
}

func TestReplaceBody_PropagatesServiceError(t *testing.T) {
	mockSvc := new(svc_mocks.DecisionService)
	mockOut := new(out_mocks.DecisionEdit)

	d := &decision.Decision{ID: "0020", Status: "accepted"}
	body := "# T\n"

	mockSvc.On("GetDecisionByID", "model", "0020").Return(d, nil)
	mockSvc.On("ReplaceBody", "model", d, body, false).Return(errors.New("use --force"))

	interactor := NewEditDecisionInteractor(mockSvc, mockOut)
	err := interactor.ReplaceBody("model", "0020", "", body, false)

	assert.ErrorContains(t, err, "use --force")
	mockOut.AssertNotCalled(t, "Edited")
}

func TestEdit_EditFails(t *testing.T) {
	mockSvc := new(svc_mocks.DecisionService)
	mockOut := new(out_mocks.DecisionEdit)

	d := &decision.Decision{ID: "0055"}
	ctx := "Broken update"

	mockSvc.On("GetDecisionByID", "model", "0055").Return(d, nil)
	mockSvc.On("Edit", "model", d, &ctx, (*[]string)(nil), (*string)(nil)).Return(errors.New("write error"))

	interactor := NewEditDecisionInteractor(mockSvc, mockOut)
	err := interactor.Edit("model", "0055", "", &ctx, nil, nil)

	assert.ErrorContains(t, err, "write error")
	mockSvc.AssertExpectations(t)
}
