package tenant_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetExternalTenant(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test", "external-test")
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
			svc := tenant.NewService(tenantMappingRepoFn, nil, nil)

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
	ctx := tenant.SaveToContext(context.TODO(), "test", "external-test")
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
			svc := tenant.NewService(tenantMappingRepoFn, nil, nil)

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

func TestService_ExtractTenantIDForTenantScopedFormationTemplates(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testID, testExternal)
	ctxWithEmptyTenants := tenant.SaveToContext(context.TODO(), "", "")

	testCases := []struct {
		Name                string
		Context             context.Context
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		ExpectedError       string
		ExpectedOutput      string
	}{
		{
			Name:    "Success when tenant is GA",
			Context: ctx,
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("Get", ctx, testID).Return(newModelBusinessTenantMappingWithType(testID, testName, "", tenantEntity.Account), nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testID,
		},
		{
			Name:    "Success when tenant is SA",
			Context: ctx,
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("Get", ctx, testID).Return(newModelBusinessTenantMappingWithType(testID, testName, testParentID, tenantEntity.Subaccount), nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testParentID,
		},
		{
			Name:    "Success when empty tenant",
			Context: ctxWithEmptyTenants,
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				return &automock.TenantMappingRepository{}
			},
			ExpectedOutput: "",
		},
		{
			Name:    "Error when getting the internal tenant",
			Context: ctx,
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("Get", ctx, testID).Return(nil, testError).Once()
				return tenantMappingRepo
			},
			ExpectedError:  testError.Error(),
			ExpectedOutput: "",
		},
		{
			Name:    "Error when tenant is not in context",
			Context: context.TODO(),
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				return &automock.TenantMappingRepository{}
			},
			ExpectedError:  "cannot read tenant from context",
			ExpectedOutput: "",
		},
		{
			Name:    "Error when there is only internalID in context",
			Context: tenant.SaveToContext(context.TODO(), testID, ""),
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				return &automock.TenantMappingRepository{}
			},
			ExpectedError:  apperrors.NewTenantNotFoundError("").Error(),
			ExpectedOutput: "",
		},
		{
			Name:    "Error when there is only externalID in context",
			Context: tenant.SaveToContext(context.TODO(), "", testID),
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				return &automock.TenantMappingRepository{}
			},
			ExpectedError:  apperrors.NewTenantNotFoundError(testID).Error(),
			ExpectedOutput: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepoFn := testCase.TenantMappingRepoFn()
			svc := tenant.NewService(tenantMappingRepoFn, nil, nil)

			// WHEN
			result, err := svc.ExtractTenantIDForTenantScopedFormationTemplates(testCase.Context)

			// THEN
			if len(testCase.ExpectedError) > 0 {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError)
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
	ctx := tenant.SaveToContext(context.TODO(), "test", "external-test")
	modelTenantMappings := []*model.BusinessTenantMapping{
		newModelBusinessTenantMapping("foo1", "bar1"),
		newModelBusinessTenantMapping("foo2", "bar2"),
	}

	testCases := []struct {
		Name                string
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		ExpectedError       error
		ExpectedOutput      []*model.BusinessTenantMapping
	}{
		{
			Name: "Success",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("List", ctx).Return(modelTenantMappings, nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: modelTenantMappings,
		},
		{
			Name: "Error when listing tenants",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("List", ctx).Return([]*model.BusinessTenantMapping{}, testError).Once()
				return tenantMappingRepo
			},
			ExpectedError:  testError,
			ExpectedOutput: []*model.BusinessTenantMapping{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepo := testCase.TenantMappingRepoFn()
			svc := tenant.NewService(tenantMappingRepo, nil, nil)

			// WHEN
			result, err := svc.List(ctx)

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

func TestService_ListPageBySearchTerm(t *testing.T) {
	// GIVEN
	searchTerm := ""
	first := 100
	endCursor := ""
	ctx := tenant.SaveToContext(context.TODO(), "test", "external-test")
	modelTenantMappingPage := &model.BusinessTenantMappingPage{
		Data: []*model.BusinessTenantMapping{
			newModelBusinessTenantMapping("foo1", "bar1"),
			newModelBusinessTenantMapping("foo2", "bar2"),
		},
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
		TotalCount: 2,
	}

	testCases := []struct {
		Name                string
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		ExpectedError       error
		ExpectedOutput      *model.BusinessTenantMappingPage
	}{
		{
			Name: "Success",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("ListPageBySearchTerm", ctx, searchTerm, first, endCursor).Return(modelTenantMappingPage, nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: modelTenantMappingPage,
		},
		{
			Name: "Error when listing tenants",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("ListPageBySearchTerm", ctx, searchTerm, first, endCursor).Return(&model.BusinessTenantMappingPage{}, testError).Once()
				return tenantMappingRepo
			},
			ExpectedError:  testError,
			ExpectedOutput: &model.BusinessTenantMappingPage{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepo := testCase.TenantMappingRepoFn()
			svc := tenant.NewService(tenantMappingRepo, nil, nil)

			// WHEN
			result, err := svc.ListPageBySearchTerm(ctx, searchTerm, first, endCursor)

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
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test", "external-test")
	tenantInput := newModelBusinessTenantMappingInput(testName, "", "")
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
				tenantMappingRepo.On("DeleteByExternalTenant", ctx, tenantInput.ExternalTenant).Return(nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: nil,
		},
		{
			Name: "Error while deleting the tenant mapping",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("DeleteByExternalTenant", ctx, tenantInput.ExternalTenant).Return(testErr).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepo := testCase.TenantMappingRepoFn()
			svc := tenant.NewService(tenantMappingRepo, nil, nil)

			// WHEN
			err := svc.DeleteMany(ctx, []string{tenantInput.ExternalTenant})

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

func TestService_CreateManyIfNotExists(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test", "external-test")

	tenantInputs := []model.BusinessTenantMappingInput{newModelBusinessTenantMappingInput("test1", "", ""),
		newModelBusinessTenantMappingInput("test2", "", "").WithExternalTenant("external2")}
	tenantInputsWithSubdomains := []model.BusinessTenantMappingInput{newModelBusinessTenantMappingInput("test1", testSubdomain, ""),
		newModelBusinessTenantMappingInput("test2", "", "").WithExternalTenant("external2")}
	tenantInputsWithRegions := []model.BusinessTenantMappingInput{newModelBusinessTenantMappingInput("test1", "", testRegion),
		newModelBusinessTenantMappingInput("test2", "", testRegion).WithExternalTenant("external2")}
	tenantModelInputsWithParent := []model.BusinessTenantMappingInput{newModelBusinessTenantMappingInputWithType(testID, "test1", testParentID, "", "", tenantEntity.Account),
		newModelBusinessTenantMappingInputWithType(testParentID, "test2", "", "", "", tenantEntity.Customer)}
	tenantWithSubdomainAndRegion := newModelBusinessTenantMappingInput("test1", testSubdomain, testRegion)
	tenantModelInputsWithParentOrganization := []model.BusinessTenantMappingInput{newModelBusinessTenantMappingInputWithType(testID, "test1", testParentID, "", "", tenantEntity.Organization),
		newModelBusinessTenantMappingInputWithType(testParentID, "test2", "", "", "", tenantEntity.Folder)}

	tenantModels := []model.BusinessTenantMapping{*newModelBusinessTenantMapping(testID, "test1"),
		newModelBusinessTenantMapping(testID, "test2").WithExternalTenant("external2")}

	expectedResult := []string{testID, testID}
	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}
	noopLabelRepo := func() *automock.LabelRepository {
		return &automock.LabelRepository{}
	}
	noopLabelUpsertSvc := func() *automock.LabelUpsertService {
		return &automock.LabelUpsertService{}
	}
	testErr := errors.New("test")
	type testCase struct {
		Name                string
		tenantInputs        []model.BusinessTenantMappingInput
		TenantMappingRepoFn func(string) *automock.TenantMappingRepository
		LabelRepoFn         func() *automock.LabelRepository
		LabelUpsertSvcFn    func() *automock.LabelUpsertService
		UIDSvcFn            func() *automock.UIDService
		ExpectedError       error
		ExpectedResult      []string
	}

	testCases := []testCase{
		{
			Name:         "Success",
			tenantInputs: tenantInputs,
			TenantMappingRepoFn: func(createRepoFunc string) *automock.TenantMappingRepository {
				return createRepoSvc(ctx, createRepoFunc, tenantModels[0], tenantModels[1])
			},
			UIDSvcFn:         uidSvcFn,
			LabelRepoFn:      noopLabelRepo,
			LabelUpsertSvcFn: noopLabelUpsertSvc,
			ExpectedError:    nil,
			ExpectedResult:   expectedResult,
		},
		{
			Name:         "Success when parent tenant exists with another ID",
			tenantInputs: tenantModelInputsWithParent,
			TenantMappingRepoFn: func(createFunc string) *automock.TenantMappingRepository {
				parent := tenantModelInputsWithParent[1]
				modifiedTenant := tenantModelInputsWithParent[0]
				modifiedTenant.Parent = testInternalParentID

				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On(createFunc, ctx, *parent.ToBusinessTenantMapping(testTemporaryInternalParentID)).Return(nil).Once()
				tenantMappingRepo.On("GetByExternalTenant", ctx, parent.ExternalTenant).Return(parent.ToBusinessTenantMapping(testInternalParentID), nil).Once()
				tenantMappingRepo.On(createFunc, ctx, *modifiedTenant.ToBusinessTenantMapping(testID)).Return(nil).Once()
				tenantMappingRepo.On("GetByExternalTenant", ctx, modifiedTenant.ExternalTenant).Return(modifiedTenant.ToBusinessTenantMapping(testID), nil).Once()
				return tenantMappingRepo
			},
			UIDSvcFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(testID).Once()
				uidSvc.On("Generate").Return(testTemporaryInternalParentID).Once()
				return uidSvc
			},
			LabelRepoFn:      noopLabelRepo,
			LabelUpsertSvcFn: noopLabelUpsertSvc,
			ExpectedError:    nil,
			ExpectedResult:   []string{testInternalParentID, testID},
		},
		{
			Name:         "Success when parent tenant organization exists with another ID",
			tenantInputs: tenantModelInputsWithParentOrganization,
			TenantMappingRepoFn: func(createFunc string) *automock.TenantMappingRepository {
				parent := tenantModelInputsWithParentOrganization[1]
				modifiedTenant := tenantModelInputsWithParentOrganization[0]
				modifiedTenant.Parent = testInternalParentID

				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On(createFunc, ctx, *parent.ToBusinessTenantMapping(testTemporaryInternalParentID)).Return(nil).Once()
				tenantMappingRepo.On("GetByExternalTenant", ctx, parent.ExternalTenant).Return(parent.ToBusinessTenantMapping(testInternalParentID), nil).Once()
				tenantMappingRepo.On(createFunc, ctx, *modifiedTenant.ToBusinessTenantMapping(testID)).Return(nil).Once()
				tenantMappingRepo.On("GetByExternalTenant", ctx, modifiedTenant.ExternalTenant).Return(modifiedTenant.ToBusinessTenantMapping(testID), nil).Once()
				return tenantMappingRepo
			},
			UIDSvcFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(testID).Once()
				uidSvc.On("Generate").Return(testTemporaryInternalParentID).Once()
				return uidSvc
			},
			LabelRepoFn:      noopLabelRepo,
			LabelUpsertSvcFn: noopLabelUpsertSvc,
			ExpectedError:    nil,
			ExpectedResult:   []string{testInternalParentID, testID},
		},
		{
			Name:         "Success when subdomain should be added",
			tenantInputs: tenantInputsWithSubdomains,
			TenantMappingRepoFn: func(createFuncName string) *automock.TenantMappingRepository {
				return createRepoSvc(ctx, createFuncName, *tenantInputsWithSubdomains[0].ToBusinessTenantMapping(testID), *tenantInputsWithSubdomains[1].ToBusinessTenantMapping(testID))
			},
			UIDSvcFn:    uidSvcFn,
			LabelRepoFn: noopLabelRepo,
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				label := &model.LabelInput{
					Key:        "subdomain",
					Value:      testSubdomain,
					ObjectID:   testID,
					ObjectType: model.TenantLabelableObject,
				}
				svc.On("UpsertLabel", ctx, testID, label).Return(nil).Once()
				return svc
			},
			ExpectedError:  nil,
			ExpectedResult: expectedResult,
		},
		{
			Name:         "Success when region should be added",
			tenantInputs: tenantInputsWithRegions,
			TenantMappingRepoFn: func(createFuncName string) *automock.TenantMappingRepository {
				return createRepoSvc(ctx, createFuncName, *tenantInputsWithRegions[0].ToBusinessTenantMapping(testID), *tenantInputsWithRegions[1].ToBusinessTenantMapping(testID))
			},
			UIDSvcFn:    uidSvcFn,
			LabelRepoFn: noopLabelRepo,
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				regionLabel := &model.LabelInput{
					Key:        "region",
					Value:      testRegion,
					ObjectID:   tenantModels[1].ID,
					ObjectType: model.TenantLabelableObject,
				}
				svc.On("UpsertLabel", ctx, testID, regionLabel).Return(nil).Twice()
				return svc
			},
			ExpectedError:  nil,
			ExpectedResult: expectedResult,
		},
		{
			Name:         "Error when checking the existence of tenant",
			tenantInputs: []model.BusinessTenantMappingInput{tenantWithSubdomainAndRegion},
			TenantMappingRepoFn: func(createFuncName string) *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On(createFuncName, ctx, *tenantWithSubdomainAndRegion.ToBusinessTenantMapping(testID)).Return(nil).Once()
				tenantMappingRepo.On("GetByExternalTenant", ctx, tenantWithSubdomainAndRegion.ExternalTenant).Return(nil, testErr)
				return tenantMappingRepo
			},
			UIDSvcFn:         uidSvcFn,
			LabelRepoFn:      noopLabelRepo,
			LabelUpsertSvcFn: noopLabelUpsertSvc,
			ExpectedError:    testErr,
			ExpectedResult:   nil,
		},
		{
			Name:         "Error when subdomain label setting fails",
			tenantInputs: tenantInputsWithSubdomains,
			TenantMappingRepoFn: func(createFuncName string) *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On(createFuncName, ctx, tenantModels[0]).Return(nil).Once()
				tenantMappingRepo.On("GetByExternalTenant", ctx, tenantInputsWithSubdomains[0].ExternalTenant).Return(&tenantModels[0], nil)
				return tenantMappingRepo
			},
			UIDSvcFn:    uidSvcFn,
			LabelRepoFn: noopLabelRepo,
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				label := &model.LabelInput{
					Key:        "subdomain",
					Value:      testSubdomain,
					ObjectID:   testID,
					ObjectType: model.TenantLabelableObject,
				}
				svc.On("UpsertLabel", ctx, testID, label).Return(testErr).Once()
				return svc
			},
			ExpectedError:  testErr,
			ExpectedResult: nil,
		},
		{
			Name:         "Error when region label setting fails",
			tenantInputs: tenantInputsWithRegions,
			TenantMappingRepoFn: func(createFuncName string) *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On(createFuncName, ctx, tenantModels[0]).Return(nil).Once()
				tenantMappingRepo.On("GetByExternalTenant", ctx, tenantInputsWithRegions[0].ExternalTenant).Return(&tenantModels[0], nil)
				return tenantMappingRepo
			},
			UIDSvcFn:    uidSvcFn,
			LabelRepoFn: noopLabelRepo,
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				label := &model.LabelInput{
					Key:        "region",
					Value:      testRegion,
					ObjectID:   testID,
					ObjectType: model.TenantLabelableObject,
				}
				svc.On("UpsertLabel", ctx, testID, label).Return(testErr).Once()
				return svc
			},
			ExpectedError:  testErr,
			ExpectedResult: nil,
		},
		{
			Name:         "Error when creating the tenant",
			tenantInputs: tenantInputs,
			TenantMappingRepoFn: func(createFuncName string) *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On(createFuncName, ctx, tenantModels[0]).Return(testErr).Once()
				return tenantMappingRepo
			},
			UIDSvcFn:         uidSvcFn,
			LabelRepoFn:      noopLabelRepo,
			LabelUpsertSvcFn: noopLabelUpsertSvc,
			ExpectedError:    testErr,
			ExpectedResult:   nil,
		},
	}

	t.Run("CreateManyIfNotExists", func(t *testing.T) {
		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				uidSvc := testCase.UIDSvcFn()
				tenantMappingRepo := testCase.TenantMappingRepoFn("UnsafeCreate")
				labelRepo := testCase.LabelRepoFn()
				labelUpsertSvc := testCase.LabelUpsertSvcFn()
				defer mock.AssertExpectationsForObjects(t, tenantMappingRepo, uidSvc, labelRepo, labelUpsertSvc)

				svc := tenant.NewServiceWithLabels(tenantMappingRepo, uidSvc, labelRepo, labelUpsertSvc, nil)

				// WHEN
				res, err := svc.CreateManyIfNotExists(ctx, testCase.tenantInputs...)

				// THEN
				if testCase.ExpectedError != nil {
					require.Error(t, err)
					assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, testCase.ExpectedResult, res)
				}
			})
		}
	})

	t.Run("UpsertMany", func(t *testing.T) {
		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				uidSvc := testCase.UIDSvcFn()
				tenantMappingRepo := testCase.TenantMappingRepoFn("Upsert")
				labelRepo := testCase.LabelRepoFn()
				labelUpsertSvc := testCase.LabelUpsertSvcFn()
				defer mock.AssertExpectationsForObjects(t, tenantMappingRepo, uidSvc, labelRepo, labelUpsertSvc)

				svc := tenant.NewServiceWithLabels(tenantMappingRepo, uidSvc, labelRepo, labelUpsertSvc, nil)

				// WHEN
				res, err := svc.UpsertMany(ctx, testCase.tenantInputs...)

				// THEN
				if testCase.ExpectedError != nil {
					require.Error(t, err)
					assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, testCase.ExpectedResult, res)
				}
			})
		}
	})
}

func Test_UpsertSingle(t *testing.T) {
	ctx := tenant.SaveToContext(context.TODO(), "test", "external-test")

	tenantInput := newModelBusinessTenantMappingInput("test1", "", "")
	tenantInputWithSubdomain := newModelBusinessTenantMappingInput("test1", testSubdomain, "")
	tenantInputWithRegion := newModelBusinessTenantMappingInput("test1", "", testRegion)

	tenantModel := newModelBusinessTenantMapping(testID, "test1")

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}

	noopLabelRepo := func() *automock.LabelRepository {
		return &automock.LabelRepository{}
	}
	noopLabelUpsertSvc := func() *automock.LabelUpsertService {
		return &automock.LabelUpsertService{}
	}

	testCases := []struct {
		Name                string
		tenantInput         model.BusinessTenantMappingInput
		TenantMappingRepoFn func(string) *automock.TenantMappingRepository
		LabelRepoFn         func() *automock.LabelRepository
		LabelUpsertSvcFn    func() *automock.LabelUpsertService
		UIDSvcFn            func() *automock.UIDService
		ExpectedError       error
		ExpectedResult      string
	}{
		{
			Name:        "Success",
			tenantInput: tenantInput,
			TenantMappingRepoFn: func(createRepoFunc string) *automock.TenantMappingRepository {
				return createRepoSvc(ctx, createRepoFunc, *tenantModel)
			},
			UIDSvcFn:         uidSvcFn,
			LabelRepoFn:      noopLabelRepo,
			LabelUpsertSvcFn: noopLabelUpsertSvc,
			ExpectedError:    nil,
			ExpectedResult:   testID,
		},
		{
			Name:        "Success when subdomain should be added",
			tenantInput: tenantInputWithSubdomain,
			TenantMappingRepoFn: func(createFuncName string) *automock.TenantMappingRepository {
				return createRepoSvc(ctx, createFuncName, *tenantInputWithSubdomain.ToBusinessTenantMapping(testID))
			},
			UIDSvcFn:    uidSvcFn,
			LabelRepoFn: noopLabelRepo,
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				label := &model.LabelInput{
					Key:        "subdomain",
					Value:      testSubdomain,
					ObjectID:   testID,
					ObjectType: model.TenantLabelableObject,
				}
				svc.On("UpsertLabel", ctx, testID, label).Return(nil).Once()
				return svc
			},
			ExpectedError:  nil,
			ExpectedResult: testID,
		},
		{
			Name:        "Success when region should be added",
			tenantInput: tenantInputWithRegion,
			TenantMappingRepoFn: func(createFuncName string) *automock.TenantMappingRepository {
				return createRepoSvc(ctx, createFuncName, *tenantInputWithRegion.ToBusinessTenantMapping(testID))
			},
			UIDSvcFn:    uidSvcFn,
			LabelRepoFn: noopLabelRepo,
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				label := &model.LabelInput{
					Key:        "region",
					Value:      testRegion,
					ObjectID:   testID,
					ObjectType: model.TenantLabelableObject,
				}
				svc.On("UpsertLabel", ctx, testID, label).Return(nil).Once()
				return svc
			},
			ExpectedError:  nil,
			ExpectedResult: testID,
		},
		{
			Name:        "Error when checking the existence of tenant",
			tenantInput: tenantInput,
			TenantMappingRepoFn: func(createFuncName string) *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On(createFuncName, ctx, *tenantInput.ToBusinessTenantMapping(testID)).Return(nil).Once()
				tenantMappingRepo.On("GetByExternalTenant", ctx, tenantInput.ExternalTenant).Return(nil, testError)
				return tenantMappingRepo
			},
			UIDSvcFn:         uidSvcFn,
			LabelRepoFn:      noopLabelRepo,
			LabelUpsertSvcFn: noopLabelUpsertSvc,
			ExpectedError:    testError,
			ExpectedResult:   "",
		},
		{
			Name:        "Error when subdomain label setting fails",
			tenantInput: tenantInputWithSubdomain,
			TenantMappingRepoFn: func(createFuncName string) *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On(createFuncName, ctx, *tenantInputWithSubdomain.ToBusinessTenantMapping(testID)).Return(nil).Once()
				tenantMappingRepo.On("GetByExternalTenant", ctx, tenantInputWithSubdomain.ExternalTenant).Return(tenantModel, nil)
				return tenantMappingRepo
			},
			UIDSvcFn:    uidSvcFn,
			LabelRepoFn: noopLabelRepo,
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				label := &model.LabelInput{
					Key:        "subdomain",
					Value:      testSubdomain,
					ObjectID:   testID,
					ObjectType: model.TenantLabelableObject,
				}
				svc.On("UpsertLabel", ctx, testID, label).Return(testError).Once()
				return svc
			},
			ExpectedError:  testError,
			ExpectedResult: "",
		},
		{
			Name:        "Error when region label setting fails",
			tenantInput: tenantInputWithRegion,
			TenantMappingRepoFn: func(createFuncName string) *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On(createFuncName, ctx, *tenantInputWithRegion.ToBusinessTenantMapping(testID)).Return(nil).Once()
				tenantMappingRepo.On("GetByExternalTenant", ctx, tenantInputWithRegion.ExternalTenant).Return(tenantModel, nil)
				return tenantMappingRepo
			},
			UIDSvcFn:    uidSvcFn,
			LabelRepoFn: noopLabelRepo,
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				label := &model.LabelInput{
					Key:        "region",
					Value:      testRegion,
					ObjectID:   testID,
					ObjectType: model.TenantLabelableObject,
				}
				svc.On("UpsertLabel", ctx, testID, label).Return(testError).Once()
				return svc
			},
			ExpectedError:  testError,
			ExpectedResult: "",
		},
		{
			Name:        "Error when creating the tenant",
			tenantInput: tenantInput,
			TenantMappingRepoFn: func(createFuncName string) *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On(createFuncName, ctx, *tenantModel).Return(testError).Once()
				return tenantMappingRepo
			},
			UIDSvcFn:         uidSvcFn,
			LabelRepoFn:      noopLabelRepo,
			LabelUpsertSvcFn: noopLabelUpsertSvc,
			ExpectedError:    testError,
			ExpectedResult:   "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			uidSvc := testCase.UIDSvcFn()
			tenantMappingRepo := testCase.TenantMappingRepoFn("Upsert")
			labelRepo := testCase.LabelRepoFn()
			labelUpsertSvc := testCase.LabelUpsertSvcFn()
			defer mock.AssertExpectationsForObjects(t, tenantMappingRepo, uidSvc, labelRepo, labelUpsertSvc)

			svc := tenant.NewServiceWithLabels(tenantMappingRepo, uidSvc, labelRepo, labelUpsertSvc, nil)

			// WHEN
			result, err := svc.UpsertSingle(ctx, testCase.tenantInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				require.Equal(t, testCase.ExpectedResult, result)
			}
		})
	}
}

func Test_MultipleToTenantMapping(t *testing.T) {
	testCases := []struct {
		Name          string
		InputSlice    []model.BusinessTenantMappingInput
		ExpectedSlice []model.BusinessTenantMapping
	}{
		{
			Name: "success with more than one parent chain",
			InputSlice: []model.BusinessTenantMappingInput{
				{
					Name:           "acc1",
					ExternalTenant: "0",
				},
				{
					Name:           "acc2",
					ExternalTenant: "1",
					Parent:         "2",
				},
				{
					Name:           "customer1",
					ExternalTenant: "2",
					Parent:         "4",
				},
				{
					Name:           "acc3",
					ExternalTenant: "3",
				},
				{
					Name:           "x1",
					ExternalTenant: "4",
				},
			},
			ExpectedSlice: []model.BusinessTenantMapping{
				{
					ID:             "0",
					Name:           "acc1",
					ExternalTenant: "0",
					Status:         tenantEntity.Active,
					Type:           tenantEntity.Unknown,
				},
				{
					ID:             "4",
					Name:           "x1",
					ExternalTenant: "4",
					Status:         tenantEntity.Active,
					Type:           tenantEntity.Unknown,
				},
				{
					ID:             "2",
					Name:           "customer1",
					ExternalTenant: "2",
					Parent:         "4",
					Status:         tenantEntity.Active,
					Type:           tenantEntity.Unknown,
				},
				{
					ID:             "1",
					Name:           "acc2",
					ExternalTenant: "1",
					Parent:         "2",
					Status:         tenantEntity.Active,
					Type:           tenantEntity.Unknown,
				},
				{
					ID:             "3",
					Name:           "acc3",
					ExternalTenant: "3",
					Status:         tenantEntity.Active,
					Type:           tenantEntity.Unknown,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := tenant.NewService(nil, &serialUUIDService{}, nil)
			require.Equal(t, testCase.ExpectedSlice, svc.MultipleToTenantMapping(testCase.InputSlice))
		})
	}
}

func Test_Update(t *testing.T) {
	tnt := model.BusinessTenantMappingInput{
		Name:           testName,
		ExternalTenant: testExternal,
		Parent:         testParentID,
		Subdomain:      testSubdomain,
		Region:         testRegion,
		Type:           string(tenantEntity.Account),
		Provider:       testProvider,
	}
	tntToBusinessTenantMapping := &model.BusinessTenantMapping{
		ID:             testID,
		Name:           testName,
		ExternalTenant: testExternal,
		Parent:         testParentID,
		Type:           tenantEntity.Account,
		Provider:       testProvider,
		Status:         tenantEntity.Active,
		Initialized:    nil,
	}

	testCases := []struct {
		Name                      string
		InputID                   string
		InputTenant               model.BusinessTenantMappingInput
		TenantMappingRepositoryFn func() *automock.TenantMappingRepository
		ExpectedErr               error
	}{
		{
			Name:        "Success",
			InputID:     testID,
			InputTenant: tnt,
			TenantMappingRepositoryFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("Update", mock.Anything, tntToBusinessTenantMapping).Return(nil)
				return tenantMappingRepo
			},
			ExpectedErr: nil,
		},
		{
			Name:        "Returns error when can't update the tenant",
			InputID:     testID,
			InputTenant: tnt,
			TenantMappingRepositoryFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("Update", mock.Anything, tntToBusinessTenantMapping).Return(testError)
				return tenantMappingRepo
			},
			ExpectedErr: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			tenantMappingRepo := testCase.TenantMappingRepositoryFn()
			serialUUIDService := &serialUUIDService{}
			svc := tenant.NewService(tenantMappingRepo, serialUUIDService, nil)
			err := svc.Update(ctx, testCase.InputID, testCase.InputTenant)

			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_MoveBeforeIndex(t *testing.T) {
	testCases := []struct {
		Name           string
		InputSlice     []model.BusinessTenantMapping
		TargetTenantID string
		TargetIndex    int
		ExpectedSlice  []model.BusinessTenantMapping
		ShouldMove     bool
	}{
		{
			Name: "success",
			InputSlice: []model.BusinessTenantMapping{
				{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"},
			},
			TargetTenantID: "4",
			TargetIndex:    1,
			ExpectedSlice: []model.BusinessTenantMapping{
				{ID: "1"}, {ID: "4"}, {ID: "2"}, {ID: "3"}, {ID: "5"},
			},
			ShouldMove: true,
		},
		{
			Name: "move before first element",
			InputSlice: []model.BusinessTenantMapping{
				{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"},
			},
			TargetTenantID: "3",
			TargetIndex:    0,
			ExpectedSlice: []model.BusinessTenantMapping{
				{ID: "3"}, {ID: "1"}, {ID: "2"}, {ID: "4"}, {ID: "5"},
			},
			ShouldMove: true,
		},
		{
			Name: "move before last element",
			InputSlice: []model.BusinessTenantMapping{
				{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"},
			},
			TargetTenantID: "3",
			TargetIndex:    4,
			ExpectedSlice: []model.BusinessTenantMapping{
				{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result, moved := tenant.MoveBeforeIfShould(testCase.InputSlice, testCase.TargetTenantID, testCase.TargetIndex)
			require.Equal(t, testCase.ShouldMove, moved)
			require.Equal(t, testCase.ExpectedSlice, result)
		})
	}
}

func Test_ListLabels(t *testing.T) {
	const tenantID = "edc6857b-b0c7-49e6-9f0a-e87a9c2a4eb8"

	ctx := context.TODO()
	testErr := errors.New("failed to list labels")

	t.Run("Success", func(t *testing.T) {
		labels := map[string]*model.Label{
			"label-key": {
				ID:         "5ef5ebd0-987d-4cb6-a3c1-7d710de259a2",
				Tenant:     str.Ptr(tenantID),
				Key:        "label-key",
				Value:      "value",
				ObjectID:   tenantID,
				ObjectType: model.TenantLabelableObject,
			},
		}

		uidSvc := &automock.UIDService{}
		labelUpsertSvc := &automock.LabelUpsertService{}

		tenantRepo := &automock.TenantMappingRepository{}
		tenantRepo.On("Exists", ctx, tenantID).Return(true, nil)

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("ListForObject", ctx, tenantID, model.TenantLabelableObject, tenantID).Return(labels, nil)

		defer mock.AssertExpectationsForObjects(t, tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		svc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelUpsertSvc, nil)

		actualLabels, err := svc.ListLabels(ctx, tenantID)
		assert.NoError(t, err)
		assert.Equal(t, labels, actualLabels)
	})

	t.Run("Error when tenant existence cannot be ensured", func(t *testing.T) {
		uidSvc := &automock.UIDService{}
		labelRepo := &automock.LabelRepository{}
		labelUpsertSvc := &automock.LabelUpsertService{}

		tenantRepo := &automock.TenantMappingRepository{}
		tenantRepo.On("Exists", ctx, tenantID).Return(false, testErr)

		defer mock.AssertExpectationsForObjects(t, tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		svc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelUpsertSvc, nil)

		_, err := svc.ListLabels(ctx, tenantID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("while checking if tenant with ID %s exists", tenantID))
	})

	t.Run("Error when tenant does not exist", func(t *testing.T) {
		uidSvc := &automock.UIDService{}
		labelRepo := &automock.LabelRepository{}
		labelUpsertSvc := &automock.LabelUpsertService{}

		tenantRepo := &automock.TenantMappingRepository{}
		tenantRepo.On("Exists", ctx, tenantID).Return(false, nil)

		defer mock.AssertExpectationsForObjects(t, tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		svc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelUpsertSvc, nil)

		_, err := svc.ListLabels(ctx, tenantID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFoundError(err))
	})

	t.Run("Error when fails to list labels from repo", func(t *testing.T) {
		uidSvc := &automock.UIDService{}
		labelUpsertSvc := &automock.LabelUpsertService{}

		tenantRepo := &automock.TenantMappingRepository{}
		tenantRepo.On("Exists", ctx, tenantID).Return(true, nil)

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("ListForObject", ctx, tenantID, model.TenantLabelableObject, tenantID).Return(nil, testErr)

		defer mock.AssertExpectationsForObjects(t, tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		svc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelUpsertSvc, nil)

		_, err := svc.ListLabels(ctx, tenantID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("whilie listing labels for tenant with ID %s", tenantID))
	})
}

func Test_GetTenantByExternalID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		expected := &model.BusinessTenantMapping{
			ID:             testID,
			Name:           testName,
			ExternalTenant: testExternal,
			Status:         tenantEntity.Active,
			Type:           tenantEntity.Account,
		}

		uidSvc := &automock.UIDService{}
		labelUpsertSvc := &automock.LabelUpsertService{}
		labelRepo := &automock.LabelRepository{}

		tenantRepo := &automock.TenantMappingRepository{}
		tenantRepo.On("GetByExternalTenant", ctx, testID).Return(expected, nil)

		defer mock.AssertExpectationsForObjects(t, tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		svc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelUpsertSvc, nil)

		// WHEN
		actual, err := svc.GetTenantByExternalID(ctx, testID)

		// THEN
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("Returns error when retrieval from DB fails", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()

		uidSvc := &automock.UIDService{}
		labelUpsertSvc := &automock.LabelUpsertService{}
		labelRepo := &automock.LabelRepository{}

		tenantRepo := &automock.TenantMappingRepository{}
		tenantRepo.On("GetByExternalTenant", ctx, testID).Return(nil, testError)

		defer mock.AssertExpectationsForObjects(t, tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		svc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelUpsertSvc, nil)

		// WHEN
		actual, err := svc.GetTenantByExternalID(ctx, testID)

		// THEN
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Equal(t, testError, err)
	})
}

func TestService_CreateTenantAccessForResource(t *testing.T) {
	testCases := []struct {
		Name             string
		ConverterFn      func() *automock.BusinessTenantMappingConverter
		PersistenceFn    func() (*sqlx.DB, testdb.DBMock)
		Input            *model.TenantAccess
		ExpectedErrorMsg string
	}{
		{
			Name: "Success",
			ConverterFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessToEntity", tenantAccessModel).Return(tenantAccessEntity).Once()
				return conv
			},
			PersistenceFn: func() (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_applications ( tenant_id, id, owner ) VALUES ( ?, ?, ? ) ON CONFLICT ON CONSTRAINT tenant_applications_pkey DO NOTHING`)).
					WithArgs(testInternal, testID, true).
					WillReturnResult(sqlmock.NewResult(1, 1))
				return db, dbMock
			},
			Input: tenantAccessModel,
		},
		{
			Name:             "Error when resource does not have access table",
			Input:            invalidTenantAccessModel,
			ExpectedErrorMsg: fmt.Sprintf("entity %q does not have access table", invalidResourceType),
		},
		{
			Name: "Error while creating tenant access",
			ConverterFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessToEntity", tenantAccessModel).Return(tenantAccessEntity).Once()
				return conv
			},
			Input:            tenantAccessModel,
			ExpectedErrorMsg: fmt.Sprintf("while creating tenant acccess for resource type %q with ID %q for tenant %q", tenantAccessModel.ResourceType, tenantAccessModel.ResourceID, tenantAccessModel.InternalTenantID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			converter := unusedConverter()
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := unusedDBMock(t)
			if testCase.PersistenceFn != nil {
				db, dbMock = testCase.PersistenceFn()
			}
			ctx = persistence.SaveToContext(ctx, db)

			svc := tenant.NewService(nil, nil, converter)

			// WHEN
			err := svc.CreateTenantAccessForResource(ctx, testCase.Input)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, converter)
			dbMock.AssertExpectations(t)
		})
	}
}

func TestService_CreateTenantAccessForResourceRecursively(t *testing.T) {
	testCases := []struct {
		Name             string
		ConverterFn      func() *automock.BusinessTenantMappingConverter
		PersistenceFn    func() (*sqlx.DB, testdb.DBMock)
		Input            *model.TenantAccess
		ExpectedErrorMsg string
	}{
		{
			Name: "Success",
			ConverterFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessToEntity", tenantAccessModel).Return(tenantAccessEntity).Once()
				return conv
			},
			PersistenceFn: func() (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				dbMock.ExpectExec(regexp.QuoteMeta(`WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO tenant_applications ( tenant_id, id, owner ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents)`)).
					WithArgs(testInternal, testID, true).
					WillReturnResult(sqlmock.NewResult(1, 1))
				return db, dbMock
			},
			Input: tenantAccessModel,
		},
		{
			Name:             "Error when resource does not have access table",
			Input:            invalidTenantAccessModel,
			ExpectedErrorMsg: fmt.Sprintf("entity %q does not have access table", invalidResourceType),
		},
		{
			Name: "Error while creating tenant access",
			ConverterFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessToEntity", tenantAccessModel).Return(tenantAccessEntity).Once()
				return conv
			},
			Input:            tenantAccessModel,
			ExpectedErrorMsg: fmt.Sprintf("while creating tenant acccess for resource type %q with ID %q for tenant %q", tenantAccessModel.ResourceType, tenantAccessModel.ResourceID, tenantAccessModel.InternalTenantID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			converter := unusedConverter()
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := unusedDBMock(t)
			if testCase.PersistenceFn != nil {
				db, dbMock = testCase.PersistenceFn()
			}
			ctx = persistence.SaveToContext(ctx, db)

			svc := tenant.NewService(nil, nil, converter)

			// WHEN
			err := svc.CreateTenantAccessForResourceRecursively(ctx, testCase.Input)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, converter)
			dbMock.AssertExpectations(t)
		})
	}
}

func TestService_DeleteTenantAccessForResource(t *testing.T) {
	testCases := []struct {
		Name             string
		ConverterFn      func() *automock.BusinessTenantMappingConverter
		PersistenceFn    func() (*sqlx.DB, testdb.DBMock)
		Input            *model.TenantAccess
		ExpectedErrorMsg string
	}{
		{
			Name: "Success",
			ConverterFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessToEntity", tenantAccessModel).Return(tenantAccessEntity).Once()
				return conv
			},
			PersistenceFn: func() (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				dbMock.ExpectExec(regexp.QuoteMeta(`WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = $1 UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) DELETE FROM tenant_applications WHERE id IN ($2) AND tenant_id IN (SELECT id FROM parents)`)).
					WithArgs(testInternal, testID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				return db, dbMock
			},
			Input: tenantAccessModel,
		},
		{
			Name:             "Error when resource does not have access table",
			Input:            invalidTenantAccessModel,
			ExpectedErrorMsg: fmt.Sprintf("entity %q does not have access table", invalidResourceType),
		},
		{
			Name: "Error while deleting tenant access",
			ConverterFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessToEntity", tenantAccessModel).Return(tenantAccessEntity).Once()
				return conv
			},
			Input:            tenantAccessModel,
			ExpectedErrorMsg: fmt.Sprintf("while deleting tenant acccess for resource type %q with ID %q for tenant %q", tenantAccessModel.ResourceType, tenantAccessModel.ResourceID, tenantAccessModel.InternalTenantID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			converter := unusedConverter()
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := unusedDBMock(t)
			if testCase.PersistenceFn != nil {
				db, dbMock = testCase.PersistenceFn()
			}
			ctx = persistence.SaveToContext(ctx, db)

			svc := tenant.NewService(nil, nil, converter)

			// WHEN
			err := svc.DeleteTenantAccessForResourceRecursively(ctx, testCase.Input)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, converter)
			dbMock.AssertExpectations(t)
		})
	}
}

func TestService_GetTenantAccessForResource(t *testing.T) {
	testCases := []struct {
		Name             string
		ConverterFn      func() *automock.BusinessTenantMappingConverter
		PersistenceFn    func() (*sqlx.DB, testdb.DBMock)
		Input            *model.TenantAccess
		ExpectedErrorMsg string
		ExpectedOutput   *model.TenantAccess
	}{
		{
			Name: "Success",
			ConverterFn: func() *automock.BusinessTenantMappingConverter {
				conv := &automock.BusinessTenantMappingConverter{}
				conv.On("TenantAccessFromEntity", tenantAccessEntity).Return(tenantAccessModelWithoutExternalTenant).Once()
				return conv
			},
			PersistenceFn: func() (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				rows := sqlmock.NewRows(tenantAccessTestTableColumns)
				rows.AddRow(testInternal, testID, true)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, id, owner FROM tenant_applications WHERE tenant_id = $1 AND id = $2`)).
					WithArgs(testInternal, testID).WillReturnRows(rows)
				return db, dbMock
			},
			Input:          tenantAccessModel,
			ExpectedOutput: tenantAccessModelWithoutExternalTenant,
		},
		{
			Name:             "Error when resource does not have access table",
			Input:            invalidTenantAccessModel,
			ExpectedErrorMsg: fmt.Sprintf("entity %q does not have access table", invalidResourceType),
		},
		{
			Name: "Error while getting tenant access",
			PersistenceFn: func() (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, id, owner FROM tenant_applications WHERE tenant_id = $1 AND id = $2`)).
					WithArgs(testInternal, testID).WillReturnError(testError)
				return db, dbMock
			},
			Input:            tenantAccessModel,
			ExpectedErrorMsg: "Unexpected error while executing SQL query",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			converter := unusedConverter()
			if testCase.ConverterFn != nil {
				converter = testCase.ConverterFn()
			}
			db, dbMock := unusedDBMock(t)
			if testCase.PersistenceFn != nil {
				db, dbMock = testCase.PersistenceFn()
			}
			ctx = persistence.SaveToContext(ctx, db)

			svc := tenant.NewService(nil, nil, converter)

			// WHEN
			result, err := svc.GetTenantAccessForResource(ctx, testCase.Input.InternalTenantID, testCase.Input.ResourceID, testCase.Input.ResourceType)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedOutput, result)
			}

			mock.AssertExpectationsForObjects(t, converter)
			dbMock.AssertExpectations(t)
		})
	}
}

type serialUUIDService struct {
	i int
}

func (s *serialUUIDService) Generate() string {
	result := s.i
	s.i++
	return fmt.Sprintf("%d", result)
}

func createRepoSvc(ctx context.Context, createFuncName string, tenants ...model.BusinessTenantMapping) *automock.TenantMappingRepository {
	tenantMappingRepo := &automock.TenantMappingRepository{}
	for _, t := range tenants {
		tenantMappingRepo.On(createFuncName, ctx, t).Return(nil).Once()
		tenantMappingRepo.On("GetByExternalTenant", ctx, t.ExternalTenant).Return(&t, nil).Once()
	}
	return tenantMappingRepo
}
