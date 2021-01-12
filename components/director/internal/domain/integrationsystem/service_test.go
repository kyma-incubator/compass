package integrationsystem_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem/automock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID).Once()
		return uidSvc
	}
	modelIntSysInput := fixModelIntegrationSystemInput(testName)
	modelIntSys := fixModelIntegrationSystem(testID, testName)

	testCases := []struct {
		Name           string
		IntSysRepoFn   func() *automock.IntegrationSystemRepository
		ExpectedError  error
		ExpectedOutput string
	}{
		{
			Name: "Success",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				intSysRepo.On("Create", ctx, *modelIntSys).Return(nil).Once()
				return intSysRepo
			},
			ExpectedOutput: testID,
		},
		{
			Name: "Error when creating Integration System",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				intSysRepo.On("Create", ctx, *modelIntSys).Return(testError).Once()
				return intSysRepo
			},
			ExpectedError:  testError,
			ExpectedOutput: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			intSysRepo := testCase.IntSysRepoFn()
			idSvc := uidSvcFn()
			svc := integrationsystem.NewService(intSysRepo, idSvc)

			// WHEN
			result, err := svc.Create(ctx, modelIntSysInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			intSysRepo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
		})
	}
}

func TestService_Get(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	modelIntSys := fixModelIntegrationSystem(testID, testName)

	testCases := []struct {
		Name           string
		IntSysRepoFn   func() *automock.IntegrationSystemRepository
		ExpectedError  error
		ExpectedOutput *model.IntegrationSystem
	}{
		{
			Name: "Success",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				intSysRepo.On("Get", ctx, testID).Return(modelIntSys, nil).Once()
				return intSysRepo
			},
			ExpectedOutput: modelIntSys,
		},
		{
			Name: "Error when getting integration system",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				intSysRepo.On("Get", ctx, testID).Return(nil, testError).Once()
				return intSysRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			intSysRepo := testCase.IntSysRepoFn()
			svc := integrationsystem.NewService(intSysRepo, nil)

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

			intSysRepo.AssertExpectations(t)
		})
	}
}

func TestService_Exists(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	testCases := []struct {
		Name           string
		IntSysRepoFn   func() *automock.IntegrationSystemRepository
		ExpectedError  error
		ExpectedOutput bool
	}{
		{
			Name: "Success",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				intSysRepo.On("Exists", ctx, testID).Return(true, nil).Once()
				return intSysRepo
			},
			ExpectedOutput: true,
		},
		{
			Name: "Error when getting integration system",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				intSysRepo.On("Exists", ctx, testID).Return(false, testError).Once()
				return intSysRepo
			},
			ExpectedError:  testError,
			ExpectedOutput: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			intSysRepo := testCase.IntSysRepoFn()
			svc := integrationsystem.NewService(intSysRepo, nil)

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

			intSysRepo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	modelIntSys := fixModelIntegrationSystemPage([]*model.IntegrationSystem{
		fixModelIntegrationSystem("foo1", "bar1"),
		fixModelIntegrationSystem("foo2", "bar2"),
	})

	testCases := []struct {
		Name           string
		IntSysRepoFn   func() *automock.IntegrationSystemRepository
		InputPageSize  int
		ExpectedError  error
		ExpectedOutput model.IntegrationSystemPage
	}{
		{
			Name: "Success",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				intSysRepo.On("List", ctx, 50, testCursor).Return(modelIntSys, nil).Once()
				return intSysRepo
			},
			InputPageSize:  50,
			ExpectedOutput: modelIntSys,
		},
		{
			Name: "Error when listing integration system",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				intSysRepo.On("List", ctx, 50, testCursor).Return(model.IntegrationSystemPage{}, testError).Once()
				return intSysRepo
			},
			InputPageSize:  50,
			ExpectedError:  testError,
			ExpectedOutput: model.IntegrationSystemPage{},
		},
		{
			Name: "Error when page size too small",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				return intSysRepo
			},
			InputPageSize:  0,
			ExpectedError:  errors.New("page size must be between 1 and 200"),
			ExpectedOutput: model.IntegrationSystemPage{},
		},
		{
			Name: "Error when page size too big",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				return intSysRepo
			},
			InputPageSize:  201,
			ExpectedError:  errors.New("page size must be between 1 and 200"),
			ExpectedOutput: model.IntegrationSystemPage{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			intSysRepo := testCase.IntSysRepoFn()
			svc := integrationsystem.NewService(intSysRepo, nil)

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

			intSysRepo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	modelIntSys := fixModelIntegrationSystem(testID, testName)
	modelIntSysInput := fixModelIntegrationSystemInput(testName)

	testCases := []struct {
		Name          string
		IntSysRepoFn  func() *automock.IntegrationSystemRepository
		ExpectedError error
	}{
		{
			Name: "Success",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				intSysRepo.On("Update", ctx, *modelIntSys).Return(nil).Once()
				return intSysRepo
			},
		},
		{
			Name: "Error when updating integration system",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				intSysRepo.On("Update", ctx, *modelIntSys).Return(testError).Once()
				return intSysRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			intSysRepo := testCase.IntSysRepoFn()
			svc := integrationsystem.NewService(intSysRepo, nil)

			// WHEN
			err := svc.Update(ctx, testID, modelIntSysInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			intSysRepo.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	testCases := []struct {
		Name          string
		IntSysRepoFn  func() *automock.IntegrationSystemRepository
		ExpectedError error
	}{
		{
			Name: "Success",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				intSysRepo.On("Delete", ctx, testID).Return(nil).Once()
				return intSysRepo
			},
		},
		{
			Name: "Error when deleting integration system",
			IntSysRepoFn: func() *automock.IntegrationSystemRepository {
				intSysRepo := &automock.IntegrationSystemRepository{}
				intSysRepo.On("Delete", ctx, testID).Return(testError).Once()
				return intSysRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			intSysRepo := testCase.IntSysRepoFn()
			svc := integrationsystem.NewService(intSysRepo, nil)

			// WHEN
			err := svc.Delete(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			intSysRepo.AssertExpectations(t)
		})
	}
}
