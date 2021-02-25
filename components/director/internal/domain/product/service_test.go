package product_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/product"
	"github.com/kyma-incubator/compass/components/director/internal/domain/product/automock"
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

	modelProduct := fixProductModel()
	modelInput := *fixProductModelInput()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.ProductRepository
		Input        model.ProductInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("Create", ctx, modelProduct).Return(nil).Once()
				return repo
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Error - Product creation",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("Create", ctx, modelProduct).Return(testErr).Once()
				return repo
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := product.NewService(repo)

			// when
			result, err := svc.Create(ctx, appID, testCase.Input)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.IsType(t, "string", result)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := product.NewService(nil)
		// WHEN
		_, err := svc.Create(context.TODO(), "", model.ProductInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelProduct := fixProductModel()
	modelInput := *fixProductModelInput()

	inputProductModel := mock.MatchedBy(func(pkg *model.Product) bool {
		return pkg.Title == modelInput.Title
	})

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.ProductRepository
		Input        model.ProductInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByID", ctx, tenantID, ordID).Return(modelProduct, nil).Once()
				repo.On("Update", ctx, inputProductModel).Return(nil).Once()
				return repo
			},
			InputID:     ordID,
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByID", ctx, tenantID, ordID).Return(modelProduct, nil).Once()
				repo.On("Update", ctx, inputProductModel).Return(testErr).Once()
				return repo
			},
			InputID:     ordID,
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByID", ctx, tenantID, ordID).Return(nil, testErr).Once()
				return repo
			},
			InputID:     ordID,
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := product.NewService(repo)

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
		svc := product.NewService(nil)
		// WHEN
		err := svc.Update(context.TODO(), "", model.ProductInput{})
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
		RepositoryFn func() *automock.ProductRepository
		Input        model.ProductInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("Delete", ctx, tenantID, ordID).Return(nil).Once()
				return repo
			},
			InputID:     ordID,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("Delete", ctx, tenantID, ordID).Return(testErr).Once()
				return repo
			},
			InputID:     ordID,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := product.NewService(repo)

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
		svc := product.NewService(nil)
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

	testCases := []struct {
		Name           string
		RepoFn         func() *automock.ProductRepository
		ExpectedError  error
		ExpectedOutput bool
	}{
		{
			Name: "Success",
			RepoFn: func() *automock.ProductRepository {
				pkgRepo := &automock.ProductRepository{}
				pkgRepo.On("Exists", ctx, tenantID, ordID).Return(true, nil).Once()
				return pkgRepo
			},
			ExpectedOutput: true,
		},
		{
			Name: "Error when getting Product",
			RepoFn: func() *automock.ProductRepository {
				pkgRepo := &automock.ProductRepository{}
				pkgRepo.On("Exists", ctx, tenantID, ordID).Return(false, testErr).Once()
				return pkgRepo
			},
			ExpectedError:  testErr,
			ExpectedOutput: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			pkgRepo := testCase.RepoFn()
			svc := product.NewService(pkgRepo)

			// WHEN
			result, err := svc.Exist(ctx, ordID)

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
		svc := product.NewService(nil)
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

	pkg := fixProductModel()

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ProductRepository
		Input              model.ProductInput
		InputID            string
		ExpectedProduct    *model.Product
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByID", ctx, tenantID, ordID).Return(pkg, nil).Once()
				return repo
			},
			InputID:            ordID,
			ExpectedProduct:    pkg,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Product retrieval failed",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByID", ctx, tenantID, ordID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            ordID,
			ExpectedProduct:    pkg,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := product.NewService(repo)

			// when
			pkg, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedProduct, pkg)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := product.NewService(nil)
		// WHEN
		_, err := svc.Get(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}
