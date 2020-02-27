package mp_package_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"

	"github.com/kyma-incubator/compass/components/director/internal/domain/package/automock"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	applicationID := "appid"
	name := "foo"
	desc := "bar"

	modelInput := model.PackageCreateInput{
		Name:                           name,
		Description:                    &desc,
		InstanceAuthRequestInputSchema: fixBasicSchema(),
		DefaultInstanceAuth:            &model.AuthInput{},
	}

	modelPackage := &model.Package{
		ID:                             id,
		TenantID:                       tenantID,
		ApplicationID:                  applicationID,
		Name:                           name,
		Description:                    &desc,
		InstanceAuthRequestInputSchema: fixBasicSchema(),
		DefaultInstanceAuth:            &model.Auth{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.PackageRepository
		UIDServiceFn func() *automock.UIDService
		Input        model.PackageCreateInput
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
				svc.On("Generate").Return(id).Once()
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
				svc.On("Generate").Return(id).Once()
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
			uidService := testCase.UIDServiceFn()

			svc := mp_package.NewService(repo, uidService)

			// when
			result, err := svc.Create(ctx, applicationID, testCase.Input)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.IsType(t, "string", result)
			}

			repo.AssertExpectations(t)
			uidService.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := mp_package.NewService(nil, nil)
		// WHEN
		_, err := svc.Create(context.TODO(), "", model.PackageCreateInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	name := "bar"
	desc := "baz"

	modelInput := model.PackageUpdateInput{
		Name:                           name,
		Description:                    &desc,
		InstanceAuthRequestInputSchema: fixBasicSchema(),
		DefaultInstanceAuth:            &model.AuthInput{},
	}

	inputPackageModel := mock.MatchedBy(func(pkg *model.Package) bool {
		return pkg.Name == modelInput.Name
	})

	packageModel := &model.Package{
		ID:                             id,
		TenantID:                       tenantID,
		ApplicationID:                  "id",
		Name:                           name,
		Description:                    &desc,
		InstanceAuthRequestInputSchema: fixBasicSchema(),
		DefaultInstanceAuth:            &model.Auth{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.PackageRepository
		Input        model.PackageUpdateInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(packageModel, nil).Once()
				repo.On("Update", ctx, inputPackageModel).Return(nil).Once()
				return repo
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(packageModel, nil).Once()
				repo.On("Update", ctx, inputPackageModel).Return(testErr).Once()
				return repo
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(nil, testErr).Once()
				return repo
			},
			InputID:     "foo",
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
		err := svc.Update(context.TODO(), "", model.PackageUpdateInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.PackageRepository
		Input        model.PackageCreateInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("Delete", ctx, tenantID, id).Return(nil).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("Delete", ctx, tenantID, id).Return(testErr).Once()
				return repo
			},
			InputID:     id,
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
	ctx := tenant.SaveToContext(context.TODO(), tenantID)
	id := "foo"

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
				pkgRepo.On("Exists", ctx, tenantID, id).Return(true, nil).Once()
				return pkgRepo
			},
			ExpectedOutput: true,
		},
		{
			Name: "Error when getting Package",
			RepoFn: func() *automock.PackageRepository {
				pkgRepo := &automock.PackageRepository{}
				pkgRepo.On("Exists", ctx, tenantID, id).Return(false, testErr).Once()
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
			result, err := svc.Exist(ctx, id)

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

	id := "foo"
	name := "foo"
	desc := "bar"

	pkg := fixPackageModel(t, name, desc)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.PackageRepository
		Input              model.PackageCreateInput
		InputID            string
		ExpectedPackage    *model.Package
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(pkg, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedPackage:    pkg,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Package retrieval failed",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
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

func TestService_GetForApplication(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	appID := "bar"
	name := "foo"
	desc := "bar"

	pkg := fixPackageModel(t, name, desc)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.PackageRepository
		Input              model.PackageCreateInput
		InputID            string
		ApplicationID      string
		ExpectedPackage    *model.Package
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetForApplication", ctx, tenantID, id, appID).Return(pkg, nil).Once()
				return repo
			},
			InputID:            id,
			ApplicationID:      appID,
			ExpectedPackage:    pkg,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Package retrieval failed",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetForApplication", ctx, tenantID, id, appID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ApplicationID:      appID,
			ExpectedPackage:    nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := mp_package.NewService(repo, nil)

			// when
			document, err := svc.GetForApplication(ctx, testCase.InputID, testCase.ApplicationID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedPackage, document)
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
		_, err := svc.GetForApplication(context.TODO(), "", "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetByInstanceAuthID(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	appID := "bar"
	name := "foo"
	desc := "bar"

	pkg := fixPackageModel(t, name, desc)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.PackageRepository
		Input              model.PackageCreateInput
		InstanceAuthID     string
		ExpectedPackage    *model.Package
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByInstanceAuthID", ctx, tenantID, appID).Return(pkg, nil).Once()
				return repo
			},
			InstanceAuthID:     appID,
			ExpectedPackage:    pkg,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Package retrieval failed",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("GetByInstanceAuthID", ctx, tenantID, appID).Return(nil, testErr).Once()
				return repo
			},
			InstanceAuthID:     appID,
			ExpectedPackage:    nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := mp_package.NewService(repo, nil)

			// when
			document, err := svc.GetByInstanceAuthID(ctx, testCase.InstanceAuthID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedPackage, document)
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
		_, err := svc.GetForApplication(context.TODO(), "", "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByApplicationID(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "bar"
	name := "foo"
	desc := "bar"

	packages := []*model.Package{
		fixPackageModel(t, name, desc),
		fixPackageModel(t, name, desc),
		fixPackageModel(t, name, desc),
	}
	packagePage := &model.PackagePage{
		Data:       packages,
		TotalCount: len(packages),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	after := "test"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.PackageRepository
		ExpectedResult     *model.PackagePage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("ListByApplicationID", ctx, tenantID, applicationID, 2, after).Return(packagePage, nil).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     packagePage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Return error when page size is less than 1",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				return repo
			},
			PageSize:           0,
			ExpectedResult:     packagePage,
			ExpectedErrMessage: "page size must be between 1 and 100",
		},
		{
			Name: "Return error when page size is bigger than 100",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				return repo
			},
			PageSize:           101,
			ExpectedResult:     packagePage,
			ExpectedErrMessage: "page size must be between 1 and 100",
		},
		{
			Name: "Returns error when Package listing failed",
			RepositoryFn: func() *automock.PackageRepository {
				repo := &automock.PackageRepository{}
				repo.On("ListByApplicationID", ctx, tenantID, applicationID, 2, after).Return(nil, testErr).Once()
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

			svc := mp_package.NewService(repo, nil)

			// when
			docs, err := svc.ListByApplicationID(ctx, applicationID, testCase.PageSize, after)

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
		svc := mp_package.NewService(nil, nil)
		// WHEN
		_, err := svc.ListByApplicationID(context.TODO(), "", 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}
