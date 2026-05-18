package model

import (
	"errors"
	"testing"

	"adg/internal/application/outputport"
	decisiondomain "adg/internal/domain/decision"
	out_mocks "adg/mocks/outputport"
	svc_mocks "adg/mocks/service"

	"github.com/stretchr/testify/assert"
)

func TestMigrateInteractor_MapsStepsToOutputport(t *testing.T) {
	mockSvc := new(svc_mocks.ModelService)
	mockOut := new(out_mocks.ModelMigrate)

	domainSteps := []decisiondomain.MigrationStep{
		{OldPath: "x/AD0001-a.md", NewPath: "x/0001-a.md"},
		{OldPath: "x/AD0002-b.md", Error: errors.New("read failed")},
	}
	expectedOut := []outputport.MigrationStep{
		{OldPath: "x/AD0001-a.md", NewPath: "x/0001-a.md", Error: ""},
		{OldPath: "x/AD0002-b.md", NewPath: "", Error: "read failed"},
	}

	mockSvc.On("Migrate", "x", false).Return(domainSteps, nil)
	mockOut.On("Migrated", expectedOut, false).Return()

	interactor := NewMigrateInteractor(mockSvc, mockOut)
	err := interactor.Migrate("x", false)

	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
	mockOut.AssertExpectations(t)
}

func TestMigrateInteractor_PropagatesError(t *testing.T) {
	mockSvc := new(svc_mocks.ModelService)
	mockOut := new(out_mocks.ModelMigrate)

	mockSvc.On("Migrate", "x", true).Return(nil, errors.New("walk failed"))

	interactor := NewMigrateInteractor(mockSvc, mockOut)
	err := interactor.Migrate("x", true)

	assert.ErrorContains(t, err, "walk failed")
	mockOut.AssertNotCalled(t, "Migrated")
}
