package decision

import (
	"errors"
	"testing"

	"adg/internal/domain/decision"
	out_mocks "adg/mocks/outputport"
	svc_mocks "adg/mocks/service"

	"github.com/stretchr/testify/assert"
)

func TestSupersedeInteractor_ByID_Success(t *testing.T) {
	mockSvc := new(svc_mocks.DecisionService)
	mockOut := new(out_mocks.DecisionSupersede)

	newD := &decision.Decision{ID: "0002"}
	oldD := &decision.Decision{ID: "0001"}

	mockSvc.On("GetDecisionByID", "model", "0002").Return(newD, nil)
	mockSvc.On("GetDecisionByID", "model", "0001").Return(oldD, nil)
	mockSvc.On("Supersede", "model", newD, oldD, "because").Return(nil)
	mockOut.On("Superseded", "0002", "0001").Return(nil)

	interactor := NewSupersedeInteractor(mockSvc, mockOut)
	err := interactor.Supersede("model", "0002", "", "0001", "", "because")

	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
	mockOut.AssertExpectations(t)
}

func TestSupersedeInteractor_NewNotFound(t *testing.T) {
	mockSvc := new(svc_mocks.DecisionService)
	mockOut := new(out_mocks.DecisionSupersede)

	mockSvc.On("GetDecisionByID", "model", "0099").Return(nil, errors.New("not found"))

	interactor := NewSupersedeInteractor(mockSvc, mockOut)
	err := interactor.Supersede("model", "0099", "", "0001", "", "")

	assert.ErrorContains(t, err, "not found")
	mockOut.AssertNotCalled(t, "Superseded")
}

func TestSupersedeInteractor_ServiceErrorPropagates(t *testing.T) {
	mockSvc := new(svc_mocks.DecisionService)
	mockOut := new(out_mocks.DecisionSupersede)

	newD := &decision.Decision{ID: "0002"}
	oldD := &decision.Decision{ID: "0001"}

	mockSvc.On("GetDecisionByID", "model", "0002").Return(newD, nil)
	mockSvc.On("GetDecisionByID", "model", "0001").Return(oldD, nil)
	mockSvc.On("Supersede", "model", newD, oldD, "").Return(errors.New("save failed"))

	interactor := NewSupersedeInteractor(mockSvc, mockOut)
	err := interactor.Supersede("model", "0002", "", "0001", "", "")

	assert.ErrorContains(t, err, "save failed")
	mockOut.AssertNotCalled(t, "Superseded")
}
