package tenant_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetExternalTenant(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test")
	tenantMappingModel := newModelBusinessTenantMapping(testID, testName)

	testCases := []struct {
		Name                string
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		ExpectedError       error
		ExpectedOutput      string
	}{
		{
			Name: "Success",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("Get", ctx, testID).Return(tenantMappingModel, nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testExternal,
		},
		{
			Name: "Error when getting the internal tenant",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("Get", ctx, testID).Return(nil, testError).Once()
				return tenantMappingRepo
			},
			ExpectedError:  testError,
			ExpectedOutput: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepoFn := testCase.TenantMappingRepoFn()
			svc := tenant.NewService(tenantMappingRepoFn, nil)

			// WHEN
			result, err := svc.GetExternalTenant(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			tenantMappingRepoFn.AssertExpectations(t)
		})
	}
}

func TestService_GetInternalTenant(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test")
	tenantMappingModel := newModelBusinessTenantMapping(testID, testName)

	testCases := []struct {
		Name                string
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		ExpectedError       error
		ExpectedOutput      string
	}{
		{
			Name: "Success",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("GetByExternalTenant", ctx, testExternal).Return(tenantMappingModel, nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testID,
		},
		{
			Name: "Error when getting the internal tenant",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("GetByExternalTenant", ctx, testExternal).Return(nil, testError).Once()
				return tenantMappingRepo
			},
			ExpectedError:  testError,
			ExpectedOutput: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepoFn := testCase.TenantMappingRepoFn()
			svc := tenant.NewService(tenantMappingRepoFn, nil)

			// WHEN
			result, err := svc.GetInternalTenant(ctx, testExternal)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			tenantMappingRepoFn.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test")
	modelTenantMappingPage := newModelBusinessTenantMapingPage([]*model.BusinessTenantMapping{
		newModelBusinessTenantMapping("foo1", "bar1"),
		newModelBusinessTenantMapping("foo2", "bar2"),
	})

	testCases := []struct {
		Name                string
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		InputPageSize       int
		ExpectedError       error
		ExpectedOutput      *model.BusinessTenantMappingPage
	}{
		{
			Name: "Success",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("List", ctx, 50, testCursor).Return(&modelTenantMappingPage, nil).Once()
				return tenantMappingRepo
			},
			InputPageSize:  50,
			ExpectedOutput: &modelTenantMappingPage,
		},
		{
			Name: "Error when listing integration system",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("List", ctx, 50, testCursor).Return(&model.BusinessTenantMappingPage{}, testError).Once()
				return tenantMappingRepo
			},
			InputPageSize:  50,
			ExpectedError:  testError,
			ExpectedOutput: &model.BusinessTenantMappingPage{},
		},
		{
			Name: "Error when page size too small",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				return tenantMappingRepo
			},
			InputPageSize:  0,
			ExpectedError:  errors.New("page size must be between 1 and 100"),
			ExpectedOutput: nil,
		},
		{
			Name: "Error when page size too big",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				intSysRepo := &automock.TenantMappingRepository{}
				return intSysRepo
			},
			InputPageSize:  101,
			ExpectedError:  errors.New("page size must be between 1 and 100"),
			ExpectedOutput: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepo := testCase.TenantMappingRepoFn()
			svc := tenant.NewService(tenantMappingRepo, nil)

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

			tenantMappingRepo.AssertExpectations(t)
		})
	}
}

func TestService_DeleteMany(t *testing.T) {
	//GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test")
	tenantInput := newModelBusinessTenantMappingInput(testName)
	tenantModel := newModelBusinessTenantMapping(testID, testName)
	tenantModelInactive := newModelBusinessTenantMapping(testID, testName).WithStatus(model.Inactive)
	testErr := errors.New("test")
	testCases := []struct {
		Name                string
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		ExpectedOutput      error
	}{
		{
			Name: "Success",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("GetByExternalTenant", ctx, tenantInput.ExternalTenant).Return(tenantModel, nil).Once()
				tenantMappingRepo.On("Update", ctx, &tenantModelInactive).Return(nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: nil,
		},
		{
			Name: "Success when tenant not found",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("GetByExternalTenant", ctx, tenantInput.ExternalTenant).
					Return(tenantModel, apperrors.NewNotFoundError("test")).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: nil,
		},
		{
			Name: "Error while getting the tenant mapping",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("GetByExternalTenant", ctx, tenantInput.ExternalTenant).Return(nil, testErr).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testErr,
		},
		{
			Name: "Error while marking as inactive",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("GetByExternalTenant", ctx, tenantInput.ExternalTenant).Return(tenantModel, nil).Once()
				tenantMappingRepo.On("Update", ctx, &tenantModelInactive).Return(testErr).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepo := testCase.TenantMappingRepoFn()
			svc := tenant.NewService(tenantMappingRepo, nil)

			// WHEN
			err := svc.DeleteMany(ctx, []model.BusinessTenantMappingInput{tenantInput})

			// THEN
			if testCase.ExpectedOutput != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedOutput.Error())
			} else {
				assert.NoError(t, err)
			}

			tenantMappingRepo.AssertExpectations(t)
		})
	}

}

func TestService_AbsolutelyNotUpsert(t *testing.T) {
	//GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test")

	tenantInputs := []model.BusinessTenantMappingInput{newModelBusinessTenantMappingInput("test1"),
		newModelBusinessTenantMappingInput("test1").WithExternalTenant("external2")}

	tenantModels := []model.BusinessTenantMapping{*newModelBusinessTenantMapping(testID, "test1"),
		newModelBusinessTenantMapping(testID, "test2").WithExternalTenant("external2")}

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}
	testErr := errors.New("test")
	testCases := []struct {
		Name                string
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		ExpectedOutput      error
	}{
		{
			Name: "Succes",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantModels[0].ExternalTenant).Return(false, nil)
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantModels[1].ExternalTenant).Return(true, nil)
				tenantMappingRepo.On("Create", ctx, tenantModels[0]).Return(nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: nil,
		},
		{
			Name: "Error when checking the existence of tenant",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantModels[0].ExternalTenant).Return(false, testErr)
				return tenantMappingRepo
			},
			ExpectedOutput: testErr,
		},
		{
			Name: "Error when creating the tenant",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantModels[0].ExternalTenant).Return(false, nil)
				tenantMappingRepo.On("Create", ctx, tenantModels[0]).Return(testErr).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepo := testCase.TenantMappingRepoFn()
			uidSvc := uidSvcFn()
			svc := tenant.NewService(tenantMappingRepo, uidSvc)

			// WHEN
			err := svc.Create(ctx, tenantInputs)

			// THEN
			if testCase.ExpectedOutput != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedOutput.Error())
			} else {
				assert.NoError(t, err)
			}

			tenantMappingRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}

}

func TestService_Sync(t *testing.T) {
	//GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test")

	tenantInputs := []model.BusinessTenantMappingInput{newModelBusinessTenantMappingInput("test1"),
		newModelBusinessTenantMappingInput("test2").WithExternalTenant("external2")}

	tenantModels := []model.BusinessTenantMapping{*newModelBusinessTenantMapping(testID, "test1"),
		newModelBusinessTenantMapping(testID, "test2").WithExternalTenant("external2")}

	tenantFromDb := newModelBusinessTenantMapping(testID, "test3").WithExternalTenant("external3")

	tenantsFromDb := newModelBusinessTenantMapingPage(
		[]*model.BusinessTenantMapping{&tenantFromDb})

	tenantToDelete := newModelBusinessTenantMapping(testID, "test3").WithStatus(model.Inactive).WithExternalTenant("external3")

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}
	testErr := errors.New("test")
	testCases := []struct {
		Name                string
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		ExpectedOutput      error
	}{
		{
			Name: "Success",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("List", ctx, 100, "").Return(&tenantsFromDb, nil).Once()
				tenantMappingRepo.On("Update", ctx, &tenantToDelete).Return(nil).Once()
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantModels[0].ExternalTenant).Return(false, nil).Once()
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantModels[1].ExternalTenant).Return(true, nil).Once()
				tenantMappingRepo.On("Create", ctx, tenantModels[0]).Return(nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: nil,
		},
		{
			Name: "Error when listing",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("List", ctx, 100, "").Return(nil, testErr).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testErr,
		},
		{
			Name: "Error when deleting",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("List", ctx, 100, "").Return(&tenantsFromDb, nil).Once()
				tenantMappingRepo.On("Update", ctx, &tenantToDelete).Return(testErr).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testErr,
		},
		{
			Name: "Error when checking the existence",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("List", ctx, 100, "").Return(&tenantsFromDb, nil).Once()
				tenantMappingRepo.On("Update", ctx, &tenantToDelete).Return(nil).Once()
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantModels[0].ExternalTenant).Return(false, testErr).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testErr,
		},
		{
			Name: "Error when creating tenant",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("List", ctx, 100, "").Return(&tenantsFromDb, nil).Once()
				tenantMappingRepo.On("Update", ctx, &tenantToDelete).Return(nil).Once()
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantModels[0].ExternalTenant).Return(false, nil).Once()
				tenantMappingRepo.On("Create", ctx, tenantModels[0]).Return(testErr).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepo := testCase.TenantMappingRepoFn()
			uidSvc := uidSvcFn()
			svc := tenant.NewService(tenantMappingRepo, uidSvc)

			// WHEN
			err := svc.Sync(ctx, tenantInputs)

			// THEN
			if testCase.ExpectedOutput != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedOutput.Error())
			} else {
				assert.NoError(t, err)
			}

			tenantMappingRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}

}
