package formationconstraint_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/automock"
	tnt "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := tnt.SaveToContext(context.TODO(), testTenantID, testTenantID)

	testErr := errors.New("test error")

	uidSvcFn := func() *automock.UidService {
		uidSvc := &automock.UidService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}

	testCases := []struct {
		Name                          string
		Context                       context.Context
		Input                         *model.FormationConstraintInput
		FormationConstraintRepository func() *automock.FormationConstraintRepository
		FormationConstraintConverter  func() *automock.FormationConstraintConverter
		ExpectedOutput                string
		ExpectedError                 error
	}{
		{
			Name:    "Success",
			Context: ctx,
			Input:   formationConstraintModelInput,
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromModelInputToModel", formationConstraintModelInput, testID).Return(formationConstraintModel).Once()
				return converter
			},
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("Create", ctx, formationConstraintModel).Return(nil).Once()
				return repo
			},
			ExpectedOutput: testID,
			ExpectedError:  nil,
		},
		{
			Name:    "Error when creating formation constraint",
			Context: ctx,
			Input:   formationConstraintModelInput,
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromModelInputToModel", formationConstraintModelInput, testID).Return(formationConstraintModel).Once()
				return converter
			},
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("Create", ctx, formationConstraintModel).Return(testErr).Once()
				return repo
			},
			ExpectedOutput: "",
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationConstraintRepo := testCase.FormationConstraintRepository()
			formationConstraintConv := testCase.FormationConstraintConverter()
			idSvc := uidSvcFn()

			svc := formationconstraint.NewService(formationConstraintRepo, nil, idSvc, formationConstraintConv)

			// WHEN
			result, err := svc.Create(testCase.Context, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, formationConstraintRepo, idSvc, formationConstraintConv)
		})
	}
}

func TestService_Get(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	testCases := []struct {
		Name                          string
		Input                         string
		FormationConstraintRepository func() *automock.FormationConstraintRepository
		ExpectedOutput                *model.FormationConstraint
		ExpectedError                 error
	}{
		{
			Name:  "Success",
			Input: testID,
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("Get", ctx, testID).Return(formationConstraintModel, nil).Once()
				return repo
			},
			ExpectedOutput: formationConstraintModel,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when getting formation template",
			Input: testID,
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("Get", ctx, testID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationConstraintRepo := testCase.FormationConstraintRepository()

			svc := formationconstraint.NewService(formationConstraintRepo, nil, nil, nil)

			// WHEN
			result, err := svc.Get(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, formationConstraintRepo)
		})
	}
}

func TestService_List(t *testing.T) {
	// GIVEN
	ctx := tnt.SaveToContext(context.TODO(), testTenantID, testTenantID)

	testErr := errors.New("test error")

	testCases := []struct {
		Name                          string
		Context                       context.Context
		FormationConstraintRepository func() *automock.FormationConstraintRepository
		ExpectedOutput                []*model.FormationConstraint
		ExpectedError                 error
	}{
		{
			Name:    "Success",
			Context: ctx,
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("ListAll", ctx).Return([]*model.FormationConstraint{formationConstraintModel}, nil).Once()
				return repo
			},
			ExpectedOutput: []*model.FormationConstraint{formationConstraintModel},
			ExpectedError:  nil,
		},
		{
			Name:    "Error when listing formation constraints",
			Context: ctx,
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("ListAll", ctx).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationConstraintRepo := testCase.FormationConstraintRepository()

			svc := formationconstraint.NewService(formationConstraintRepo, nil, nil, nil)

			// WHEN
			result, err := svc.List(testCase.Context)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, formationConstraintRepo)
		})
	}
}

func TestService_ListByFormationTemplateID(t *testing.T) {
	// GIVEN
	ctx := tnt.SaveToContext(context.TODO(), testTenantID, testTenantID)

	testErr := errors.New("test error")

	testCases := []struct {
		Name                                   string
		Context                                context.Context
		FormationConstraintRepository          func() *automock.FormationConstraintRepository
		FormationConstraintReferenceRepository func() *automock.FormationTemplateConstraintReferenceRepository
		ExpectedOutput                         []*model.FormationConstraint
		ExpectedError                          error
	}{
		{
			Name:    "Success",
			Context: ctx,
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("ListByIDs", ctx, []string{testID}).Return([]*model.FormationConstraint{formationConstraintModel}, nil).Once()
				return repo
			},
			FormationConstraintReferenceRepository: func() *automock.FormationTemplateConstraintReferenceRepository {
				repo := &automock.FormationTemplateConstraintReferenceRepository{}
				repo.On("ListByFormationTemplateID", ctx, formationTemplateID).Return([]*model.FormationTemplateConstraintReference{formationConstraintReference}, nil).Once()
				return repo
			},
			ExpectedOutput: []*model.FormationConstraint{formationConstraintModel},
			ExpectedError:  nil,
		},
		{
			Name:    "Error when listing formation constraints",
			Context: ctx,
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("ListByIDs", ctx, []string{testID}).Return(nil, testErr).Once()
				return repo
			},
			FormationConstraintReferenceRepository: func() *automock.FormationTemplateConstraintReferenceRepository {
				repo := &automock.FormationTemplateConstraintReferenceRepository{}
				repo.On("ListByFormationTemplateID", ctx, formationTemplateID).Return([]*model.FormationTemplateConstraintReference{formationConstraintReference}, nil).Once()
				return repo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:                          "Error when listing constraint references",
			Context:                       ctx,
			FormationConstraintRepository: UnusedFormationConstraintRepository,
			FormationConstraintReferenceRepository: func() *automock.FormationTemplateConstraintReferenceRepository {
				repo := &automock.FormationTemplateConstraintReferenceRepository{}
				repo.On("ListByFormationTemplateID", ctx, formationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationConstraintRepo := testCase.FormationConstraintRepository()
			formationConstraintReferenceRepo := testCase.FormationConstraintReferenceRepository()
			svc := formationconstraint.NewService(formationConstraintRepo, formationConstraintReferenceRepo, nil, nil)

			// WHEN
			result, err := svc.ListByFormationTemplateID(testCase.Context, formationTemplateID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, formationConstraintRepo, formationConstraintReferenceRepo)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	ctx := tnt.SaveToContext(context.TODO(), testTenantID, testTenantID)
	testErr := errors.New("test error")

	testCases := []struct {
		Name                          string
		Context                       context.Context
		Input                         string
		FormationConstraintRepository func() *automock.FormationConstraintRepository
		ExpectedError                 error
	}{
		{
			Name:    "Success",
			Context: ctx,
			Input:   testID,
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("Delete", ctx, testID).Return(nil).Once()
				return repo
			},
			ExpectedError: nil,
		},
		{
			Name:    "Error when deleting formation constraint",
			Context: ctx,
			Input:   testID,
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("Delete", ctx, testID).Return(testErr).Once()
				return repo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationConstraintRepo := testCase.FormationConstraintRepository()

			svc := formationconstraint.NewService(formationConstraintRepo, nil, nil, nil)

			// WHEN
			err := svc.Delete(testCase.Context, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, formationConstraintRepo)
		})
	}
}

func TestService_ListMatchingConstraints(t *testing.T) {
	// GIVEN
	ctx := tnt.SaveToContext(context.TODO(), testTenantID, testTenantID)

	testErr := errors.New("test error")

	testCases := []struct {
		Name                                   string
		Context                                context.Context
		FormationConstraintRepository          func() *automock.FormationConstraintRepository
		FormationConstraintReferenceRepository func() *automock.FormationTemplateConstraintReferenceRepository
		ExpectedOutput                         []*model.FormationConstraint
		ExpectedError                          error
	}{
		{
			Name:    "Success",
			Context: ctx,
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("ListMatchingFormationConstraints", ctx, []string{testID}, location, matchingDetails).Return([]*model.FormationConstraint{formationConstraintModel}, nil).Once()
				return repo
			},
			FormationConstraintReferenceRepository: func() *automock.FormationTemplateConstraintReferenceRepository {
				repo := &automock.FormationTemplateConstraintReferenceRepository{}
				repo.On("ListByFormationTemplateID", ctx, formationTemplateID).Return([]*model.FormationTemplateConstraintReference{formationConstraintReference}, nil).Once()
				return repo
			},
			ExpectedOutput: []*model.FormationConstraint{formationConstraintModel},
			ExpectedError:  nil,
		},
		{
			Name:    "Error when listing formation constraints",
			Context: ctx,
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("ListMatchingFormationConstraints", ctx, []string{testID}, location, matchingDetails).Return(nil, testErr).Once()
				return repo
			},
			FormationConstraintReferenceRepository: func() *automock.FormationTemplateConstraintReferenceRepository {
				repo := &automock.FormationTemplateConstraintReferenceRepository{}
				repo.On("ListByFormationTemplateID", ctx, formationTemplateID).Return([]*model.FormationTemplateConstraintReference{formationConstraintReference}, nil).Once()
				return repo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:                          "Error when listing constraint references",
			Context:                       ctx,
			FormationConstraintRepository: UnusedFormationConstraintRepository,
			FormationConstraintReferenceRepository: func() *automock.FormationTemplateConstraintReferenceRepository {
				repo := &automock.FormationTemplateConstraintReferenceRepository{}
				repo.On("ListByFormationTemplateID", ctx, formationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationConstraintRepo := testCase.FormationConstraintRepository()
			formationConstraintReferenceRepo := testCase.FormationConstraintReferenceRepository()
			svc := formationconstraint.NewService(formationConstraintRepo, formationConstraintReferenceRepo, nil, nil)

			// WHEN
			result, err := svc.ListMatchingConstraints(testCase.Context, formationTemplateID, location, matchingDetails)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, formationConstraintRepo, formationConstraintReferenceRepo)
		})
	}
}

func TestService_Update(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	testCases := []struct {
		Name                          string
		Context                       context.Context
		InputConstraintTemplate       *model.FormationConstraintInput
		FormationConstraintRepository func() *automock.FormationConstraintRepository
		FormationConstraintConverter  func() *automock.FormationConstraintConverter
		ExpectedErrorMessage          string
	}{
		{
			Name:                    "Success",
			Context:                 ctx,
			InputConstraintTemplate: formationConstraintModelInput,
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("Update", ctx, formationConstraintModel).Return(nil).Once()
				return repo
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromModelInputToModel", formationConstraintModelInput, testID).Return(formationConstraintModel).Once()

				return converter
			},
			ExpectedErrorMessage: "",
		},
		{
			Name:                    "Error when updating fails",
			Context:                 ctx,
			InputConstraintTemplate: formationConstraintModelInput,
			FormationConstraintRepository: func() *automock.FormationConstraintRepository {
				repo := &automock.FormationConstraintRepository{}
				repo.On("Update", ctx, formationConstraintModel).Return(testErr).Once()
				return repo
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromModelInputToModel", formationConstraintModelInput, testID).Return(formationConstraintModel).Once()

				return converter
			},
			ExpectedErrorMessage: "while updating Formation Constraint with ID",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := UnusedFormationConstraintRepository()
			if testCase.FormationConstraintRepository != nil {
				repo = testCase.FormationConstraintRepository()
			}
			conv := UnusedFormationConstraintConverter()
			if testCase.FormationConstraintConverter != nil {
				conv = testCase.FormationConstraintConverter()
			}
			svc := formationconstraint.NewService(repo, nil, nil, conv)

			// WHEN
			err := svc.Update(testCase.Context, testID, testCase.InputConstraintTemplate)

			// THEN
			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo, conv)
		})
	}
}

func TestService_ListByFormationTemplateIDs(t *testing.T) {
	ctx := context.TODO()

	testErr := errors.New("test err")
	formationTemplateIDs := []string{formationTemplateID1, formationTemplateID2, formationTemplateID3}
	constraintIDs := []string{constraintID1, constraintID2}

	constraintRefs := []*model.FormationTemplateConstraintReference{
		{
			ConstraintID:        constraintIDs[0],
			FormationTemplateID: formationTemplateIDs[0],
		},
		{
			ConstraintID:        constraintIDs[1],
			FormationTemplateID: formationTemplateIDs[1],
		},
		{
			ConstraintID:        constraintIDs[1],
			FormationTemplateID: formationTemplateIDs[2],
		},
	}

	constraints := []*model.FormationConstraint{formationConstraint1, formationConstraint2, globalConstraint}

	constraintsPerFormationTemplate := [][]*model.FormationConstraint{
		{
			formationConstraint1,
			formationConstraint2,
			globalConstraint,
		},
		{
			formationConstraint1,
			globalConstraint,
		},
	}

	testCases := []struct {
		Name                                     string
		FormationTemplateConstraintReferenceRepo func() *automock.FormationTemplateConstraintReferenceRepository
		FormationConstraintRepo                  func() *automock.FormationConstraintRepository
		Input                                    []string
		ExpectedConstraints                      [][]*model.FormationConstraint
		ExpectedError                            error
	}{
		{
			Name: "Success",
			FormationTemplateConstraintReferenceRepo: func() *automock.FormationTemplateConstraintReferenceRepository {
				refRepo := &automock.FormationTemplateConstraintReferenceRepository{}
				refRepo.On("ListByFormationTemplateIDs", ctx, formationTemplateIDs).Return(constraintRefs, nil)
				return refRepo
			},
			FormationConstraintRepo: func() *automock.FormationConstraintRepository {
				constraintRepo := &automock.FormationConstraintRepository{}
				constraintRepo.On("ListByIDsAndGlobal", ctx, append(constraintIDs, constraintIDs[1])).Return(constraints, nil)
				return constraintRepo
			},
			Input:               formationTemplateIDs,
			ExpectedConstraints: constraintsPerFormationTemplate,
			ExpectedError:       nil,
		},
		{
			Name: "Returns error when listing constraints fails",
			FormationTemplateConstraintReferenceRepo: func() *automock.FormationTemplateConstraintReferenceRepository {
				refRepo := &automock.FormationTemplateConstraintReferenceRepository{}
				refRepo.On("ListByFormationTemplateIDs", ctx, formationTemplateIDs).Return(constraintRefs, nil)
				return refRepo
			},
			FormationConstraintRepo: func() *automock.FormationConstraintRepository {
				constraintRepo := &automock.FormationConstraintRepository{}
				constraintRepo.On("ListByIDsAndGlobal", ctx, append(constraintIDs, constraintIDs[1])).Return(nil, testErr)
				return constraintRepo
			},
			Input:               formationTemplateIDs,
			ExpectedConstraints: nil,
			ExpectedError:       testErr,
		},
		{
			Name: "Returns error when listing constraints refs fails",
			FormationTemplateConstraintReferenceRepo: func() *automock.FormationTemplateConstraintReferenceRepository {
				refRepo := &automock.FormationTemplateConstraintReferenceRepository{}
				refRepo.On("ListByFormationTemplateIDs", ctx, formationTemplateIDs).Return(nil, testErr)
				return refRepo
			},
			Input:               formationTemplateIDs,
			ExpectedConstraints: nil,
			ExpectedError:       testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			constraintRefRepo := testCase.FormationTemplateConstraintReferenceRepo()
			constraintRepo := UnusedFormationConstraintRepository()
			if testCase.FormationConstraintRepo != nil {
				constraintRepo = testCase.FormationConstraintRepo()
			}
			svc := formationconstraint.NewService(constraintRepo, constraintRefRepo, nil, nil)

			res, err := svc.ListByFormationTemplateIDs(ctx, testCase.Input)

			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Nil(t, res)
			} else {
				require.Nil(t, err)
				reflect.DeepEqual(res, testCase.ExpectedConstraints)
			}

			mock.AssertExpectationsForObjects(t, constraintRefRepo, constraintRepo)
		})
	}
}
