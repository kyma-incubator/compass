package product_test

import (
	"context"
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
		UIDServiceFn func() *automock.UIDService
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
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(productID)
				return svc
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
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(productID)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			uidSvc := testCase.UIDServiceFn()

			svc := product.NewService(repo, uidSvc)

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
		svc := product.NewService(nil, nil)
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

	inputProductModel := mock.MatchedBy(func(prod *model.Product) bool {
		return prod.Title == modelInput.Title
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
				repo.On("GetByID", ctx, tenantID, productID).Return(modelProduct, nil).Once()
				repo.On("Update", ctx, inputProductModel).Return(nil).Once()
				return repo
			},
			InputID:     productID,
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByID", ctx, tenantID, productID).Return(modelProduct, nil).Once()
				repo.On("Update", ctx, inputProductModel).Return(testErr).Once()
				return repo
			},
			InputID:     productID,
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByID", ctx, tenantID, productID).Return(nil, testErr).Once()
				return repo
			},
			InputID:     productID,
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := product.NewService(repo, nil)

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
		svc := product.NewService(nil, nil)
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
				repo.On("Delete", ctx, tenantID, productID).Return(nil).Once()
				return repo
			},
			InputID:     productID,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("Delete", ctx, tenantID, productID).Return(testErr).Once()
				return repo
			},
			InputID:     productID,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := product.NewService(repo, nil)

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
		svc := product.NewService(nil, nil)
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
				productRepo := &automock.ProductRepository{}
				productRepo.On("Exists", ctx, tenantID, productID).Return(true, nil).Once()
				return productRepo
			},
			ExpectedOutput: true,
		},
		{
			Name: "Error when getting Product",
			RepoFn: func() *automock.ProductRepository {
				productRepo := &automock.ProductRepository{}
				productRepo.On("Exists", ctx, tenantID, productID).Return(false, testErr).Once()
				return productRepo
			},
			ExpectedError:  testErr,
			ExpectedOutput: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			productRepo := testCase.RepoFn()
			svc := product.NewService(productRepo, nil)

			// WHEN
			result, err := svc.Exist(ctx, productID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			productRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := product.NewService(nil, nil)
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

	productModel := fixProductModel()

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
				repo.On("GetByID", ctx, tenantID, productID).Return(productModel, nil).Once()
				return repo
			},
			InputID:            productID,
			ExpectedProduct:    productModel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Product retrieval failed",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByID", ctx, tenantID, productID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            productID,
			ExpectedProduct:    productModel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := product.NewService(repo, nil)

			// when
			prod, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedProduct, prod)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := product.NewService(nil, nil)
		// WHEN
		_, err := svc.Get(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByApplicationID(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	products := []*model.Product{
		fixProductModel(),
		fixProductModel(),
		fixProductModel(),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.ProductRepository
		ExpectedResult     []*model.Product
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("ListByApplicationID", ctx, tenantID, appID).Return(products, nil).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     products,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when APIDefinition listing failed",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("ListByApplicationID", ctx, tenantID, appID).Return(nil, testErr).Once()
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

			svc := product.NewService(repo, nil)

			// when
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
		svc := product.NewService(nil, nil)
		// WHEN
		_, err := svc.ListByApplicationID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}
