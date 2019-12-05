package apptemplate_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)
	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID).Once()
		return uidSvc
	}
	modelAppTemplate := fixModelAppTemplate(testID, testName)

	testCases := []struct {
		Name              string
		Input             *model.ApplicationTemplateInput
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		ExpectedError     error
		ExpectedOutput    string
	}{
		{
			Name:  "Success",
			Input: fixModelAppTemplateInput(testName, appInputJSONString),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Create", ctx, *modelAppTemplate).Return(nil).Once()
				return appTemplateRepo
			},
			ExpectedOutput: testID,
		},
		{
			Name:  "Error when creating application template",
			Input: fixModelAppTemplateInput(testName, appInputJSONString),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Create", ctx, *modelAppTemplate).Return(testError).Once()
				return appTemplateRepo
			},
			ExpectedError:  testError,
			ExpectedOutput: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			idSvc := uidSvcFn()
			svc := apptemplate.NewService(appTemplateRepo, idSvc)

			// WHEN
			result, err := svc.Create(ctx, *testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			appTemplateRepo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
		})
	}
}

func TestService_Get(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)
	modelAppTemplate := fixModelAppTemplate(testID, testName)

	testCases := []struct {
		Name              string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		ExpectedError     error
		ExpectedOutput    *model.ApplicationTemplate
	}{
		{
			Name: "Success",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, testID).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			ExpectedOutput: modelAppTemplate,
		},
		{
			Name: "Error when getting application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Get", ctx, testID).Return(nil, testError).Once()
				return appTemplateRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, nil)

			// WHEN
			result, err := svc.Get(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			appTemplateRepo.AssertExpectations(t)
		})
	}
}

func TestService_Exists(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)

	testCases := []struct {
		Name              string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		ExpectedError     error
		ExpectedOutput    bool
	}{
		{
			Name: "Success",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", ctx, testID).Return(true, nil).Once()
				return appTemplateRepo
			},
			ExpectedOutput: true,
		},
		{
			Name: "Error when getting application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", ctx, testID).Return(false, testError).Once()
				return appTemplateRepo
			},
			ExpectedError:  testError,
			ExpectedOutput: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, nil)

			// WHEN
			result, err := svc.Exists(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			appTemplateRepo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)
	modelAppTemplate := fixModelAppTemplatePage([]*model.ApplicationTemplate{
		fixModelAppTemplate("foo1", "bar1"),
		fixModelAppTemplate("foo2", "bar2"),
	})

	testCases := []struct {
		Name              string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		InputPageSize     int
		ExpectedError     error
		ExpectedOutput    model.ApplicationTemplatePage
	}{
		{
			Name: "Success",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("List", ctx, 50, testCursor).Return(modelAppTemplate, nil).Once()
				return appTemplateRepo
			},
			InputPageSize:  50,
			ExpectedOutput: modelAppTemplate,
		},
		{
			Name: "Error when listing application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("List", ctx, 50, testCursor).Return(model.ApplicationTemplatePage{}, testError).Once()
				return appTemplateRepo
			},
			InputPageSize:  50,
			ExpectedError:  testError,
			ExpectedOutput: model.ApplicationTemplatePage{},
		},
		{
			Name: "Error when page size too small",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				return appTemplateRepo
			},
			InputPageSize:  0,
			ExpectedError:  errors.New("page size must be between 1 and 100"),
			ExpectedOutput: model.ApplicationTemplatePage{},
		},
		{
			Name: "Error when page size too big",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				return appTemplateRepo
			},
			InputPageSize:  101,
			ExpectedError:  errors.New("page size must be between 1 and 100"),
			ExpectedOutput: model.ApplicationTemplatePage{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, nil)

			// WHEN
			result, err := svc.List(ctx, testCase.InputPageSize, testCursor)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			appTemplateRepo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)
	modelAppTemplate := fixModelAppTemplate(testID, testName)

	testCases := []struct {
		Name              string
		Input             *model.ApplicationTemplateInput
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		ExpectedError     error
	}{
		{
			Name:  "Success",
			Input: fixModelAppTemplateInput(testName, appInputJSONString),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Update", ctx, *modelAppTemplate).Return(nil).Once()
				return appTemplateRepo
			},
		},
		{
			Name:  "Error when updating application template",
			Input: fixModelAppTemplateInput(testName, appInputJSONString),
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Update", ctx, *modelAppTemplate).Return(testError).Once()
				return appTemplateRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, nil)

			// WHEN
			err := svc.Update(ctx, testID, *testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			appTemplateRepo.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant)

	testCases := []struct {
		Name              string
		AppTemplateRepoFn func() *automock.ApplicationTemplateRepository
		ExpectedError     error
	}{
		{
			Name: "Success",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Delete", ctx, testID).Return(nil).Once()
				return appTemplateRepo
			},
		},
		{
			Name: "Error when deleting application template",
			AppTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Delete", ctx, testID).Return(testError).Once()
				return appTemplateRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateRepo := testCase.AppTemplateRepoFn()
			svc := apptemplate.NewService(appTemplateRepo, nil)

			// WHEN
			err := svc.Delete(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			appTemplateRepo.AssertExpectations(t)
		})
	}
}
