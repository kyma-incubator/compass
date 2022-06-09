package formationtemplate_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}

	testCases := []struct {
		Name                        string
		Input                       *model.FormationTemplateInput
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		FormationTemplateConverter  func() *automock.FormationTemplateConverter
		ExpectedOutput              string
		ExpectedError               error
	}{
		{
			Name:  "Success",
			Input: &inputFormationTemplateModel,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelInputToModel", &inputFormationTemplateModel, testID).Return(&formationTemplateModel).Once()
				return converter
			},
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Create", ctx, &formationTemplateModel).Return(nil).Once()
				return repo

			},
			ExpectedOutput: testID,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when creating formation template",
			Input: &inputFormationTemplateModel,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelInputToModel", &inputFormationTemplateModel, testID).Return(&formationTemplateModel).Once()
				return converter
			},
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Create", ctx, &formationTemplateModel).Return(testErr).Once()
				return repo

			},
			ExpectedOutput: "",
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()
			formationTemplateConv := testCase.FormationTemplateConverter()
			idSvc := uidSvcFn()

			svc := formationtemplate.NewService(formationTemplateRepo, idSvc, formationTemplateConv)

			// WHEN
			result, err := svc.Create(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, formationTemplateRepo, idSvc, formationTemplateConv)
		})
	}
}

func TestService_Get(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	testCases := []struct {
		Name                        string
		Input                       string
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		ExpectedOutput              *model.FormationTemplate
		ExpectedError               error
	}{
		{
			Name:  "Success",
			Input: testID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, testID).Return(&formationTemplateModel, nil).Once()
				return repo

			},
			ExpectedOutput: &formationTemplateModel,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when getting formation template",
			Input: testID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, testID).Return(nil, testErr).Once()
				return repo

			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil)

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

			mock.AssertExpectationsForObjects(t, formationTemplateRepo)
		})
	}
}

func TestService_List(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")
	pageSize := 20
	invalidPageSize := -100

	testCases := []struct {
		Name                        string
		PageSize                    int
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		ExpectedOutput              *model.FormationTemplatePage
		ExpectedError               error
	}{
		{
			Name:     "Success",
			PageSize: pageSize,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("List", ctx, pageSize, mock.Anything).Return(&formationTemplateModelPage, nil).Once()
				return repo

			},
			ExpectedOutput: &formationTemplateModelPage,
			ExpectedError:  nil,
		},
		{
			Name:     "Error when listing formation templates",
			PageSize: pageSize,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("List", ctx, pageSize, mock.Anything).Return(nil, testErr).Once()
				return repo

			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:                        "Error when invalid page size is given",
			PageSize:                    invalidPageSize,
			FormationTemplateRepository: UnusedFormationTemplateRepository,
			ExpectedOutput:              nil,
			ExpectedError:               errors.New("page size must be between 1 and 200"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil)

			// WHEN
			result, err := svc.List(ctx, testCase.PageSize, "")

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, formationTemplateRepo)
		})
	}
}

func TestService_Update(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}

	testCases := []struct {
		Name                        string
		Input                       string
		InputFormationTemplate      *model.FormationTemplateInput
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		FormationTemplateConverter  func() *automock.FormationTemplateConverter
		ExpectedError               error
	}{
		{
			Name:                   "Success",
			Input:                  testID,
			InputFormationTemplate: &inputFormationTemplateModel,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(true, nil).Once()
				repo.On("Update", ctx, &formationTemplateModel).Return(nil).Once()
				return repo

			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelInputToModel", &inputFormationTemplateModel, testID).Return(&formationTemplateModel).Once()

				return converter
			},
			ExpectedError: nil,
		},
		{
			Name:                   "Error when formation template does not exist",
			Input:                  testID,
			InputFormationTemplate: &inputFormationTemplateModel,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(false, nil).Once()
				return repo

			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedError:              apperrors.NewNotFoundError(resource.FormationTemplate, testID),
		},
		{
			Name:                   "Error when formation existence check failed",
			Input:                  testID,
			InputFormationTemplate: &inputFormationTemplateModel,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(false, testErr).Once()
				return repo

			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedError:              testErr,
		},
		{
			Name:                   "Error when updating formation template fails",
			Input:                  testID,
			InputFormationTemplate: &inputFormationTemplateModel,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(true, nil).Once()
				repo.On("Update", ctx, &formationTemplateModel).Return(testErr).Once()
				return repo

			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelInputToModel", &inputFormationTemplateModel, testID).Return(&formationTemplateModel).Once()

				return converter
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()
			formationTemplateConverter := testCase.FormationTemplateConverter()

			svc := formationtemplate.NewService(formationTemplateRepo, uidSvcFn(), formationTemplateConverter)

			// WHEN
			err := svc.Update(ctx, testCase.Input, testCase.InputFormationTemplate)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, formationTemplateRepo)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	testCases := []struct {
		Name                        string
		Input                       string
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		ExpectedError               error
	}{
		{
			Name:  "Success",
			Input: testID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Delete", ctx, testID).Return(nil).Once()
				return repo

			},
			ExpectedError: nil,
		},
		{
			Name:  "Error when deleting formation template",
			Input: testID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Delete", ctx, testID).Return(testErr).Once()
				return repo

			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil)

			// WHEN
			err := svc.Delete(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, formationTemplateRepo)
		})
	}
}
