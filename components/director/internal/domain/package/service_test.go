package mp_package_test

import (
	"context"
	"fmt"
	"testing"

	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/kyma-incubator/compass/components/director/internal/domain/package/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	modelPackage := fixPackageModel()
	modelInput := *fixPackageModelInput()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.PackageRepository
		UIDServiceFn func() *automock.UIDService
		Input        model.PackageInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("Create", ctx, modelPackage).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(packageID)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Error - Package creation",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("Create", ctx, modelPackage).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(packageID).Once()
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			upackageIDService := testCase.UIDServiceFn()

			svc := mp_package.NewService(repo, upackageIDService)

			// when
			result, err := svc.Create(ctx, appID, testCase.Input)

			// then
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
		svc := mp_package.NewService(nil, nil)
		// WHEN
		_, err := svc.Create(context.TODO(), "", model.PackageInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelPackage := fixPackageModel()
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
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByID", ctx, tenantID, packageID).Return(modelPackage, nil).Once()
				repo.On("Update", ctx, inputPackageModel).Return(nil).Once()
				return repo
			},
			InputID:     packageID,
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByID", ctx, tenantID, packageID).Return(modelPackage, nil).Once()
				repo.On("Update", ctx, inputPackageModel).Return(testErr).Once()
				return repo
			},
			InputID:     packageID,
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByID", ctx, tenantID, packageID).Return(nil, testErr).Once()
				return repo
			},
			InputID:     packageID,
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := mp_package.NewService(repo, nil)

			// when
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

			// then
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
		svc := mp_package.NewService(nil, nil)
		// WHEN
		err := svc.Update(context.TODO(), "", model.PackageInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// given
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
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := mp_package.NewService(repo, nil)

			// when
			err := svc.Delete(ctx, testCase.InputID)

			// then
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
		svc := mp_package.NewService(nil, nil)
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
			svc := mp_package.NewService(pkgRepo, nil)

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
		svc := mp_package.NewService(nil, nil)
		// WHEN
		_, err := svc.Exist(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	pkg := fixPackageModel()

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
			svc := mp_package.NewService(repo, nil)

			// when
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
		svc := mp_package.NewService(nil, nil)
		// WHEN
		_, err := svc.Get(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}
