package integrationdependency_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationdependency"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationdependency/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(integrationDependencyID)
		return uidSvc
	}

	testCases := []struct {
		Name                        string
		InputResourceType           resource.Type
		InputResourceID             string
		IntegrationDependencyInput  model.IntegrationDependencyInput
		IntegrationDependencyRepoFn func() *automock.IntegrationDependencyRepository
		UIDServiceFn                func() *automock.UIDService
		ExpectedError               error
		ExpectedOutput              string
	}{
		{
			Name:                       "Success with resource type Application",
			InputResourceType:          resource.Application,
			InputResourceID:            "application-id",
			IntegrationDependencyInput: fixIntegrationDependencyInputModelWithPackageOrdID(packageID),
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("Create", ctx, tenantID, mock.Anything).Return(nil).Once()
				return IntegrationDependencyRepo
			},
			UIDServiceFn:   uidSvcFn,
			ExpectedOutput: integrationDependencyID,
		},
		{
			Name:                       "Success with resource type ApplicationTemplateVersion",
			InputResourceType:          resource.ApplicationTemplateVersion,
			InputResourceID:            "application-template-version-id",
			IntegrationDependencyInput: fixIntegrationDependencyInputModelWithPackageOrdID(packageID),
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("CreateGlobal", ctx, mock.Anything).Return(nil).Once()
				return IntegrationDependencyRepo
			},
			UIDServiceFn:   uidSvcFn,
			ExpectedOutput: integrationDependencyID,
		},
		{
			Name:                       "Fail while creating integration dependency for Application",
			InputResourceType:          resource.Application,
			InputResourceID:            "application-id",
			IntegrationDependencyInput: fixIntegrationDependencyInputModelWithPackageOrdID(packageID),
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("Create", ctx, tenantID, mock.Anything).Return(testErr).Once()
				return IntegrationDependencyRepo
			},
			UIDServiceFn:  uidSvcFn,
			ExpectedError: testErr,
		},
		{
			Name:                       "Fail while creating integration dependency for ApplicationTemplateVersion",
			InputResourceType:          resource.ApplicationTemplateVersion,
			InputResourceID:            "application-template-version-id",
			IntegrationDependencyInput: fixIntegrationDependencyInputModelWithPackageOrdID(packageID),
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("CreateGlobal", ctx, mock.Anything).Return(testErr).Once()
				return IntegrationDependencyRepo
			},
			UIDServiceFn:  uidSvcFn,
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			integrationDependencyRepo := testCase.IntegrationDependencyRepoFn()
			idSvc := testCase.UIDServiceFn()
			svc := integrationdependency.NewService(integrationDependencyRepo, idSvc)

			// WHEN
			result, err := svc.Create(ctx, testCase.InputResourceType, testCase.InputResourceID, str.Ptr(packageID), testCase.IntegrationDependencyInput, 123)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, result)
			}

			integrationDependencyRepo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := integrationdependency.NewService(nil, uid.NewService())
		// WHEN
		_, err := svc.Create(context.TODO(), "", "", nil, model.IntegrationDependencyInput{}, 123)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	testCases := []struct {
		Name                        string
		InputResourceType           resource.Type
		InputID                     string
		IntegrationDependencyInput  model.IntegrationDependencyInput
		IntegrationDependencyRepoFn func() *automock.IntegrationDependencyRepository
		ExpectedError               error
		ExpectedOutput              string
	}{
		{
			Name:                       "Success with resource type Application",
			InputResourceType:          resource.Application,
			InputID:                    integrationDependencyID,
			IntegrationDependencyInput: fixIntegrationDependencyInputModelWithPackageOrdID(packageID),
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("GetByID", ctx, tenantID, integrationDependencyID).Return(fixIntegrationDependencyModel(integrationDependencyID), nil).Once()
				IntegrationDependencyRepo.On("Update", ctx, tenantID, mock.Anything).Return(nil).Once()
				return IntegrationDependencyRepo
			},
			ExpectedOutput: integrationDependencyID,
		},
		{
			Name:                       "Success with resource type ApplicationTemplateVersion",
			InputResourceType:          resource.ApplicationTemplateVersion,
			InputID:                    integrationDependencyID,
			IntegrationDependencyInput: fixIntegrationDependencyInputModelWithPackageOrdID(packageID),
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("GetByIDGlobal", ctx, integrationDependencyID).Return(fixIntegrationDependencyModel(integrationDependencyID), nil).Once()
				IntegrationDependencyRepo.On("UpdateGlobal", ctx, mock.Anything).Return(nil).Once()
				return IntegrationDependencyRepo
			},
			ExpectedOutput: integrationDependencyID,
		},
		{
			Name:                       "Fail while getting integration dependency by id for Application",
			InputResourceType:          resource.Application,
			InputID:                    integrationDependencyID,
			IntegrationDependencyInput: fixIntegrationDependencyInputModelWithPackageOrdID(packageID),
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("GetByID", ctx, tenantID, integrationDependencyID).Return(nil, testErr).Once()
				return IntegrationDependencyRepo
			},
			ExpectedError: testErr,
		},
		{
			Name:                       "Fail while getting integration dependency by id for ApplicationTemplateVersion",
			InputResourceType:          resource.ApplicationTemplateVersion,
			InputID:                    integrationDependencyID,
			IntegrationDependencyInput: fixIntegrationDependencyInputModelWithPackageOrdID(packageID),
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("GetByIDGlobal", ctx, integrationDependencyID).Return(nil, testErr).Once()
				return IntegrationDependencyRepo
			},
			ExpectedError: testErr,
		},
		{
			Name:                       "Fail while updating integration dependency for Application",
			InputResourceType:          resource.Application,
			InputID:                    integrationDependencyID,
			IntegrationDependencyInput: fixIntegrationDependencyInputModelWithPackageOrdID(packageID),
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("GetByID", ctx, tenantID, integrationDependencyID).Return(fixIntegrationDependencyModel(integrationDependencyID), nil).Once()
				IntegrationDependencyRepo.On("Update", ctx, tenantID, mock.Anything).Return(testErr).Once()

				return IntegrationDependencyRepo
			},
			ExpectedError: testErr,
		},
		{
			Name:                       "Fail while updating integration dependency for ApplicationTemplateVersion",
			InputResourceType:          resource.ApplicationTemplateVersion,
			InputID:                    integrationDependencyID,
			IntegrationDependencyInput: fixIntegrationDependencyInputModelWithPackageOrdID(packageID),
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("GetByIDGlobal", ctx, integrationDependencyID).Return(fixIntegrationDependencyModel(integrationDependencyID), nil).Once()
				IntegrationDependencyRepo.On("UpdateGlobal", ctx, mock.Anything).Return(testErr).Once()

				return IntegrationDependencyRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			integrationDependencyRepo := testCase.IntegrationDependencyRepoFn()
			svc := integrationdependency.NewService(integrationDependencyRepo, nil)

			// WHEN
			err := svc.Update(ctx, testCase.InputResourceType, testCase.InputID, integrationDependencyID, testCase.IntegrationDependencyInput, 123)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			integrationDependencyRepo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := integrationdependency.NewService(nil, uid.NewService())
		// WHEN
		err := svc.Update(context.TODO(), "", "", "", model.IntegrationDependencyInput{}, 123)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)

	testCases := []struct {
		Name                        string
		InputResourceType           resource.Type
		InputID                     string
		IntegrationDependencyRepoFn func() *automock.IntegrationDependencyRepository
		ExpectedError               error
	}{
		{
			Name:              "Success with resource type Application",
			InputResourceType: resource.Application,
			InputID:           integrationDependencyID,
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("Delete", ctx, tenantID, integrationDependencyID).Return(nil).Once()
				return IntegrationDependencyRepo
			},
		},
		{
			Name:              "Success with resource type ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputID:           integrationDependencyID,
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("DeleteGlobal", ctx, integrationDependencyID).Return(nil).Once()
				return IntegrationDependencyRepo
			},
		},
		{
			Name:              "Fail while deleting integration dependency for Application",
			InputResourceType: resource.Application,
			InputID:           integrationDependencyID,
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("Delete", ctx, tenantID, integrationDependencyID).Return(testErr).Once()
				return IntegrationDependencyRepo
			},
			ExpectedError: testErr,
		},
		{
			Name:              "Fail while deleting integration dependency for ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputID:           integrationDependencyID,
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("DeleteGlobal", ctx, integrationDependencyID).Return(testErr).Once()
				return IntegrationDependencyRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			IntegrationDependencyRepo := testCase.IntegrationDependencyRepoFn()
			svc := integrationdependency.NewService(IntegrationDependencyRepo, nil)

			// WHEN
			err := svc.Delete(ctx, testCase.InputResourceType, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			IntegrationDependencyRepo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := integrationdependency.NewService(nil, uid.NewService())
		// WHEN
		err := svc.Delete(context.TODO(), "", "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Get(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)

	testCases := []struct {
		Name                        string
		InputID                     string
		IntegrationDependencyRepoFn func() *automock.IntegrationDependencyRepository
		ExpectedOutput              *model.IntegrationDependency
		ExpectedError               error
	}{
		{
			Name:    "Success",
			InputID: integrationDependencyID,
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("GetByID", ctx, tenantID, integrationDependencyID).Return(fixIntegrationDependencyModel(integrationDependencyID), nil).Once()
				return IntegrationDependencyRepo
			},
			ExpectedOutput: fixIntegrationDependencyModel(integrationDependencyID),
		},
		{
			Name:    "Fail while getting integration dependency",
			InputID: integrationDependencyID,
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("GetByID", ctx, tenantID, integrationDependencyID).Return(nil, testErr).Once()
				return IntegrationDependencyRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			IntegrationDependencyRepo := testCase.IntegrationDependencyRepoFn()
			svc := integrationdependency.NewService(IntegrationDependencyRepo, nil)

			// WHEN
			integrationDependency, err := svc.Get(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, integrationDependency)
			}

			IntegrationDependencyRepo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := integrationdependency.NewService(nil, uid.NewService())
		// WHEN
		_, err := svc.Get(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByApplicationID(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	integrationDependencies := []*model.IntegrationDependency{fixIntegrationDependencyModel(integrationDependencyID)}
	applicationID := "application-id"
	testCases := []struct {
		Name                        string
		InputID                     string
		IntegrationDependencyRepoFn func() *automock.IntegrationDependencyRepository
		ExpectedOutput              []*model.IntegrationDependency
		ExpectedError               error
	}{
		{
			Name:    "Success",
			InputID: applicationID,
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("ListByResourceID", ctx, tenantID, resource.Application, applicationID).Return(integrationDependencies, nil).Once()
				return IntegrationDependencyRepo
			},
			ExpectedOutput: integrationDependencies,
		},
		{
			Name:    "Fail while listing by resource id",
			InputID: applicationID,
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("ListByResourceID", ctx, tenantID, resource.Application, applicationID).Return(nil, testErr).Once()
				return IntegrationDependencyRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			IntegrationDependencyRepo := testCase.IntegrationDependencyRepoFn()
			svc := integrationdependency.NewService(IntegrationDependencyRepo, nil)

			// WHEN
			integrationDependencies, err := svc.ListByApplicationID(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, integrationDependencies)
			}

			IntegrationDependencyRepo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := integrationdependency.NewService(nil, uid.NewService())
		// WHEN
		_, err := svc.ListByApplicationID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByApplicationTemplateVersionID(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	integrationDependencies := []*model.IntegrationDependency{fixIntegrationDependencyModel(integrationDependencyID)}
	applicationTemplateVersionID := "application-template-version-id"
	testCases := []struct {
		Name                        string
		InputID                     string
		IntegrationDependencyRepoFn func() *automock.IntegrationDependencyRepository
		ExpectedOutput              []*model.IntegrationDependency
		ExpectedError               error
	}{
		{
			Name:    "Success",
			InputID: applicationTemplateVersionID,
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("ListByResourceID", ctx, "", resource.ApplicationTemplateVersion, applicationTemplateVersionID).Return(integrationDependencies, nil).Once()
				return IntegrationDependencyRepo
			},
			ExpectedOutput: integrationDependencies,
		},
		{
			Name:    "Fail while listing by resource id",
			InputID: applicationTemplateVersionID,
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("ListByResourceID", ctx, "", resource.ApplicationTemplateVersion, applicationTemplateVersionID).Return(nil, testErr).Once()
				return IntegrationDependencyRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			IntegrationDependencyRepo := testCase.IntegrationDependencyRepoFn()
			svc := integrationdependency.NewService(IntegrationDependencyRepo, nil)

			// WHEN
			integrationDependencies, err := svc.ListByApplicationTemplateVersionID(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, integrationDependencies)
			}

			IntegrationDependencyRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListByApplicationIDs(t *testing.T) {
	//GIVEN
	firstIntDepID := "intDepID1"
	secondIntDepID := "intDepID2"
	firstAppID := "appID1"
	secondAppID := "appID2"
	appIDs := []string{firstAppID, secondAppID}

	intDepFirstApp := fixIntegrationDependencyModel(firstIntDepID)
	intDepFirstApp.ApplicationID = &firstAppID
	intDepSecondApp := fixIntegrationDependencyModel(secondIntDepID)
	intDepSecondApp.ApplicationID = &secondAppID

	intDeps := []*model.IntegrationDependencyPage{
		{
			Data: []*model.IntegrationDependency{intDepFirstApp, intDepSecondApp},
			PageInfo: &pagination.Page{
				StartCursor: "",
				EndCursor:   "",
				HasNextPage: false,
			},
			TotalCount: 1,
		},
	}

	after := "test"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.IntegrationDependencyRepository
		ExpectedResult     []*model.IntegrationDependencyPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.IntegrationDependencyRepository {
				repo := &automock.IntegrationDependencyRepository{}
				repo.On("ListByApplicationIDs", ctx, tenantID, appIDs, 2, after).Return(intDeps, nil).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     intDeps,
			ExpectedErrMessage: "",
		},
		{
			Name: "Return error when page size is less than 1",
			RepositoryFn: func() *automock.IntegrationDependencyRepository {
				repo := &automock.IntegrationDependencyRepository{}
				return repo
			},
			PageSize:           0,
			ExpectedResult:     intDeps,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Return error when page size is bigger than 200",
			RepositoryFn: func() *automock.IntegrationDependencyRepository {
				repo := &automock.IntegrationDependencyRepository{}
				return repo
			},
			PageSize:           201,
			ExpectedResult:     intDeps,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when Integration Dependencies listing failed",
			RepositoryFn: func() *automock.IntegrationDependencyRepository {
				repo := &automock.IntegrationDependencyRepository{}
				repo.On("ListByApplicationIDs", ctx, tenantID, appIDs, 2, after).Return(nil, testErr).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := integrationdependency.NewService(repo, nil)

			// WHEN
			intDeps, err := svc.ListByApplicationIDs(ctx, appIDs, testCase.PageSize, after)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, intDeps)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := integrationdependency.NewService(nil, nil)
		// WHEN
		_, err := svc.ListByApplicationIDs(context.TODO(), nil, 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByPackageID(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	integrationDependencies := []*model.IntegrationDependency{fixIntegrationDependencyModel(integrationDependencyID)}
	packageID := "package-id"
	testCases := []struct {
		Name                        string
		InputID                     string
		IntegrationDependencyRepoFn func() *automock.IntegrationDependencyRepository
		ExpectedOutput              []*model.IntegrationDependency
		ExpectedError               error
	}{
		{
			Name:    "Success",
			InputID: packageID,
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("ListByResourceID", ctx, tenantID, resource.Package, packageID).Return(integrationDependencies, nil).Once()
				return IntegrationDependencyRepo
			},
			ExpectedOutput: integrationDependencies,
		},
		{
			Name:    "Fail while listing by resource id",
			InputID: packageID,
			IntegrationDependencyRepoFn: func() *automock.IntegrationDependencyRepository {
				IntegrationDependencyRepo := &automock.IntegrationDependencyRepository{}
				IntegrationDependencyRepo.On("ListByResourceID", ctx, tenantID, resource.Package, packageID).Return(nil, testErr).Once()
				return IntegrationDependencyRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			IntegrationDependencyRepo := testCase.IntegrationDependencyRepoFn()
			svc := integrationdependency.NewService(IntegrationDependencyRepo, nil)

			// WHEN
			integrationDependencies, err := svc.ListByPackageID(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, integrationDependencies)
			}

			IntegrationDependencyRepo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := integrationdependency.NewService(nil, uid.NewService())
		// WHEN
		_, err := svc.ListByPackageID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}
