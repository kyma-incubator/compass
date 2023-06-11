package ordpackage_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	ordpackage "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/kyma-incubator/compass/components/director/internal/domain/package/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	modelPackageForApp := fixPackageModelForApp()
	modelPackageForAppTemplateVersion := fixPackageModelForAppTemplateVersion()
	modelInput := *fixPackageModelInput()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.PackageRepository
		UIDServiceFn func() *automock.UIDService
		Input        model.PackageInput
		ResourceType resource.Type
		ResourceID   string
		ExpectedErr  error
	}{
		{
			Name: "Success for Application",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("Create", ctx, tenantID, modelPackageForApp).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(packageID)
				return svc
			},
			Input:        modelInput,
			ResourceType: resource.Application,
			ResourceID:   appID,
			ExpectedErr:  nil,
		},
		{
			Name: "Success for Application Template Version",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("CreateGlobal", ctx, modelPackageForAppTemplateVersion).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(packageID)
				return svc
			},
			Input:        modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ResourceID:   appTemplateVersionID,
			ExpectedErr:  nil,
		},
		{
			Name: "Error - Package creation for Application",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("Create", ctx, tenantID, modelPackageForApp).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(packageID).Once()
				return svc
			},
			Input:        modelInput,
			ResourceType: resource.Application,
			ResourceID:   appID,
			ExpectedErr:  testErr,
		},
		{
			Name: "Error - Package creation for Application Template Version",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("CreateGlobal", ctx, modelPackageForAppTemplateVersion).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(packageID).Once()
				return svc
			},
			Input:        modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ResourceID:   appTemplateVersionID,
			ExpectedErr:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			upackageIDService := testCase.UIDServiceFn()

			svc := ordpackage.NewService(repo, upackageIDService)

			// WHEN
			result, err := svc.Create(ctx, testCase.ResourceType, testCase.ResourceID, testCase.Input, uint64(123456))

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.IsType(t, "string", result)
			}

			mock.AssertExpectationsForObjects(t, repo, upackageIDService)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := ordpackage.NewService(nil, fixUIDService())
		// WHEN
		_, err := svc.Create(context.TODO(), resource.Application, "", model.PackageInput{}, 0)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	modelPackageForApp := fixPackageModelForApp()
	modelPackageForAppTemplateVersion := fixPackageModelForAppTemplateVersion()
	modelInput := *fixPackageModelInput()

	inputPackageModel := mock.MatchedBy(func(pkg *model.Package) bool {
		return pkg.Title == modelInput.Title
	})

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.PackageRepository
		Input        model.PackageInput
		InputID      string
		ResourceType resource.Type
		ExpectedErr  error
	}{
		{
			Name: "Success for Application",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByID", ctx, tenantID, packageID).Return(modelPackageForApp, nil).Once()
				repo.On("Update", ctx, tenantID, inputPackageModel).Return(nil).Once()
				return repo
			},
			InputID:      packageID,
			Input:        modelInput,
			ResourceType: resource.Application,
			ExpectedErr:  nil,
		},
		{
			Name: "Success for Application Template Version",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByIDGlobal", ctx, packageID).Return(modelPackageForAppTemplateVersion, nil).Once()
				repo.On("UpdateGlobal", ctx, inputPackageModel).Return(nil).Once()
				return repo
			},
			InputID:      packageID,
			Input:        modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ExpectedErr:  nil,
		},
		{
			Name: "Update Error for Application",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByID", ctx, tenantID, packageID).Return(modelPackageForApp, nil).Once()
				repo.On("Update", ctx, tenantID, inputPackageModel).Return(testErr).Once()
				return repo
			},
			InputID:      packageID,
			Input:        modelInput,
			ResourceType: resource.Application,
			ExpectedErr:  testErr,
		},
		{
			Name: "Update Error for Application Template Version",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByIDGlobal", ctx, packageID).Return(modelPackageForAppTemplateVersion, nil).Once()
				repo.On("UpdateGlobal", ctx, inputPackageModel).Return(testErr).Once()
				return repo
			},
			InputID:      packageID,
			Input:        modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ExpectedErr:  testErr,
		},
		{
			Name: "Get Error for Application",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByID", ctx, tenantID, packageID).Return(nil, testErr).Once()
				return repo
			},
			InputID:      packageID,
			Input:        modelInput,
			ResourceType: resource.Application,
			ExpectedErr:  testErr,
		},
		{
			Name: "Get Error for Application Template Version",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByIDGlobal", ctx, packageID).Return(nil, testErr).Once()
				return repo
			},
			InputID:      packageID,
			Input:        modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ExpectedErr:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := ordpackage.NewService(repo, nil)

			// WHEN
			err := svc.Update(ctx, testCase.ResourceType, testCase.InputID, testCase.Input, 0)

			// THEN
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := ordpackage.NewService(nil, fixUIDService())
		// WHEN
		err := svc.Update(context.TODO(), resource.Application, "", model.PackageInput{}, 0)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.PackageRepository
		Input        model.PackageInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("Delete", ctx, tenantID, packageID).Return(nil).Once()
				return repo
			},
			InputID:     packageID,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("Delete", ctx, tenantID, packageID).Return(testErr).Once()
				return repo
			},
			InputID:     packageID,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := ordpackage.NewService(repo, nil)

			// WHEN
			err := svc.Delete(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := ordpackage.NewService(nil, nil)
		// WHEN
		err := svc.Delete(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Exist(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	packageID := packageID

	testCases := []struct {
		Name           string
		RepoFn         func() *automock.PackageRepository
		ExpectedError  error
		ExpectedOutput bool
	}{
		{
			Name: "Success",
			RepoFn: func() *automock.PackageRepository {
				pkgRepo := &automock.PackageRepository{}
				pkgRepo.On("Exists", ctx, tenantID, packageID).Return(true, nil).Once()
				return pkgRepo
			},
			ExpectedOutput: true,
		},
		{
			Name: "Error when getting Package",
			RepoFn: func() *automock.PackageRepository {
				pkgRepo := &automock.PackageRepository{}
				pkgRepo.On("Exists", ctx, tenantID, packageID).Return(false, testErr).Once()
				return pkgRepo
			},
			ExpectedError:  testErr,
			ExpectedOutput: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			pkgRepo := testCase.RepoFn()
			svc := ordpackage.NewService(pkgRepo, nil)

			// WHEN
			result, err := svc.Exist(ctx, packageID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			pkgRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := ordpackage.NewService(nil, nil)
		// WHEN
		_, err := svc.Exist(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Get(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	pkg := fixPackageModelForApp()

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.PackageRepository
		Input              model.PackageInput
		InputID            string
		ExpectedPackage    *model.Package
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByID", ctx, tenantID, packageID).Return(pkg, nil).Once()
				return repo
			},
			InputID:            packageID,
			ExpectedPackage:    pkg,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Package retrieval failed",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByID", ctx, tenantID, packageID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            packageID,
			ExpectedPackage:    pkg,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := ordpackage.NewService(repo, nil)

			// WHEN
			pkg, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedPackage, pkg)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := ordpackage.NewService(nil, nil)
		// WHEN
		_, err := svc.Get(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByApplicationID(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	pkgs := []*model.Package{
		fixPackageModelForApp(),
		fixPackageModelForApp(),
		fixPackageModelForApp(),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.PackageRepository
		ExpectedResult     []*model.Package
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("ListByResourceID", ctx, tenantID, appID, resource.Application).Return(pkgs, nil).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     pkgs,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Package listing failed",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("ListByResourceID", ctx, tenantID, appID, resource.Application).Return(nil, testErr).Once()
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

			svc := ordpackage.NewService(repo, nil)

			// WHEN
			docs, err := svc.ListByApplicationID(ctx, appID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, docs)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := ordpackage.NewService(nil, nil)
		// WHEN
		_, err := svc.ListByApplicationID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByApplicationTemplateVersionID(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	pkgs := []*model.Package{
		fixPackageModelForApp(),
		fixPackageModelForApp(),
		fixPackageModelForApp(),
	}

	ctx := context.TODO()

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.PackageRepository
		ExpectedResult     []*model.Package
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("ListByResourceID", ctx, "", appTemplateVersionID, resource.ApplicationTemplateVersion).Return(pkgs, nil).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     pkgs,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Package listing failed",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("ListByResourceID", ctx, "", appTemplateVersionID, resource.ApplicationTemplateVersion).Return(nil, testErr).Once()
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

			svc := ordpackage.NewService(repo, nil)

			// WHEN
			docs, err := svc.ListByApplicationTemplateVersionID(ctx, appTemplateVersionID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, docs)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func fixUIDService() *automock.UIDService {
	uidSvc := &automock.UIDService{}
	uidSvc.On("Generate").Return(packageID)
	return uidSvc
}
