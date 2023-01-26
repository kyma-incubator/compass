package formationtemplateconstraintreferences_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	testCases := []struct {
		Name                                   string
		Context                                context.Context
		Input                                  *model.FormationTemplateConstraintReference
		FormationConstraintReferenceRepository func() *automock.FormationTemplateConstraintReferenceRepository
		ExpectedErrorMsg                       string
	}{
		{
			Name:    "Success",
			Context: ctx,
			Input:   constraintReference,
			FormationConstraintReferenceRepository: func() *automock.FormationTemplateConstraintReferenceRepository {
				repo := &automock.FormationTemplateConstraintReferenceRepository{}
				repo.On("Create", ctx, constraintReference).Return(nil).Once()
				return repo
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:    "Error when creating formation constraint",
			Context: ctx,
			Input:   constraintReference,
			FormationConstraintReferenceRepository: func() *automock.FormationTemplateConstraintReferenceRepository {
				repo := &automock.FormationTemplateConstraintReferenceRepository{}
				repo.On("Create", ctx, constraintReference).Return(testErr).Once()
				return repo
			},
			ExpectedErrorMsg: "while creating Formation Template Constraint Reference",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			constraintReferenceRepo := testCase.FormationConstraintReferenceRepository()

			svc := formationtemplateconstraintreferences.NewService(constraintReferenceRepo, nil)

			// WHEN
			err := svc.Create(testCase.Context, testCase.Input)

			// THEN
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, constraintReferenceRepo)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	testCases := []struct {
		Name                                   string
		Context                                context.Context
		FormationConstraintReferenceRepository func() *automock.FormationTemplateConstraintReferenceRepository
		ExpectedErrorMsg                       string
	}{
		{
			Name: "Success",
			FormationConstraintReferenceRepository: func() *automock.FormationTemplateConstraintReferenceRepository {
				repo := &automock.FormationTemplateConstraintReferenceRepository{}
				repo.On("Delete", ctx, constraintID, templateID).Return(nil).Once()
				return repo
			},
			ExpectedErrorMsg: "",
		},
		{
			Name: "Error when creating formation constraint",
			FormationConstraintReferenceRepository: func() *automock.FormationTemplateConstraintReferenceRepository {
				repo := &automock.FormationTemplateConstraintReferenceRepository{}
				repo.On("Delete", ctx, constraintID, templateID).Return(testErr).Once()
				return repo
			},
			ExpectedErrorMsg: "while deleting Formation Template Constraint Reference",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			constraintReferenceRepo := testCase.FormationConstraintReferenceRepository()

			svc := formationtemplateconstraintreferences.NewService(constraintReferenceRepo, nil)

			// WHEN
			err := svc.Delete(ctx, constraintID, templateID)

			// THEN
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, constraintReferenceRepo)
		})
	}
}
