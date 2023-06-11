package product_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

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
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	modelProductForApp := fixProductModelForApp()
	modelProductForAppTemplateVersion := fixProductModelForAppTemplateVersion()
	modelInput := *fixProductModelInput()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.ProductRepository
		UIDServiceFn func() *automock.UIDService
		Input        model.ProductInput
		ResourceType resource.Type
		ResourceID   string
		ExpectedErr  error
	}{
		{
			Name: "Success for Application",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("Create", ctx, tenantID, modelProductForApp).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(productID)
				return svc
			},
			Input:        modelInput,
			ResourceType: resource.Application,
			ResourceID:   appID,
			ExpectedErr:  nil,
		},
		{
			Name: "Success for Application Template Version",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("CreateGlobal", ctx, modelProductForAppTemplateVersion).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(productID)
				return svc
			},
			Input:        modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ResourceID:   appTemplateVersionID,
			ExpectedErr:  nil,
		},
		{
			Name: "Error - Product creation for Application",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("Create", ctx, tenantID, modelProductForApp).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(productID)
				return svc
			},
			Input:        modelInput,
			ResourceType: resource.Application,
			ResourceID:   appID,
			ExpectedErr:  testErr,
		},
		{
			Name: "Error - Product creation for Application Template Version",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("CreateGlobal", ctx, modelProductForAppTemplateVersion).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(productID)
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
			uidSvc := testCase.UIDServiceFn()

			svc := product.NewService(repo, uidSvc)

			// WHEN
			result, err := svc.Create(ctx, testCase.ResourceType, testCase.ResourceID, testCase.Input)

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
		svc := product.NewService(nil, fixUIDService())
		// WHEN
		_, err := svc.Create(context.TODO(), resource.Application, "", model.ProductInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_CreateGlobal(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()

	modelProduct := fixGlobalProductModel()
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
				repo.On("CreateGlobal", ctx, modelProduct).Return(nil).Once()
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
				repo.On("CreateGlobal", ctx, modelProduct).Return(testErr).Once()
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
			// GIVEN
			repo := testCase.RepositoryFn()
			uidSvc := testCase.UIDServiceFn()

			svc := product.NewService(repo, uidSvc)

			// WHEN
			result, err := svc.CreateGlobal(ctx, testCase.Input)

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
}

func TestService_Update(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	modelProductForApp := fixProductModelForApp()
	modelProductForAppTemplateVersion := fixProductModelForApp()
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
		ResourceType resource.Type
		ExpectedErr  error
	}{
		{
			Name: "Success for Application",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByID", ctx, tenantID, productID).Return(modelProductForApp, nil).Once()
				repo.On("Update", ctx, tenantID, inputProductModel).Return(nil).Once()
				return repo
			},
			InputID:      productID,
			Input:        modelInput,
			ResourceType: resource.Application,
			ExpectedErr:  nil,
		},
		{
			Name: "Success for Application Template Version",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByIDGlobal", ctx, productID).Return(modelProductForAppTemplateVersion, nil).Once()
				repo.On("UpdateGlobal", ctx, inputProductModel).Return(nil).Once()
				return repo
			},
			InputID:      productID,
			Input:        modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ExpectedErr:  nil,
		},
		{
			Name: "Update Error for Application",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByID", ctx, tenantID, productID).Return(modelProductForApp, nil).Once()
				repo.On("Update", ctx, tenantID, inputProductModel).Return(testErr).Once()
				return repo
			},
			InputID:      productID,
			Input:        modelInput,
			ResourceType: resource.Application,
			ExpectedErr:  testErr,
		},
		{
			Name: "Update Error for Application Template Version",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByIDGlobal", ctx, productID).Return(modelProductForAppTemplateVersion, nil).Once()
				repo.On("UpdateGlobal", ctx, inputProductModel).Return(testErr).Once()
				return repo
			},
			InputID:      productID,
			Input:        modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ExpectedErr:  testErr,
		},
		{
			Name: "Get Error for Application",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByID", ctx, tenantID, productID).Return(nil, testErr).Once()
				return repo
			},
			InputID:      productID,
			Input:        modelInput,
			ResourceType: resource.Application,
			ExpectedErr:  testErr,
		},
		{
			Name: "Get Error for Application Template Version",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByIDGlobal", ctx, productID).Return(nil, testErr).Once()
				return repo
			},
			InputID:      productID,
			Input:        modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ExpectedErr:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := product.NewService(repo, nil)

			// WHEN
			err := svc.Update(ctx, testCase.ResourceType, testCase.InputID, testCase.Input)

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
		err := svc.Update(context.TODO(), resource.Application, "", model.ProductInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_UpdateGlobal(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	modelProduct := fixGlobalProductModel()
	modelInput := *fixProductModelInput()

	inputProductModel := mock.MatchedBy(func(prod *model.Product) bool {
		return prod.Title == modelInput.Title
	})

	ctx := context.TODO()

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
				repo.On("GetByIDGlobal", ctx, productID).Return(modelProduct, nil).Once()
				repo.On("UpdateGlobal", ctx, inputProductModel).Return(nil).Once()
				return repo
			},
			InputID:     productID,
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "UpdateGlobal Error",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("GetByIDGlobal", ctx, productID).Return(modelProduct, nil).Once()
				repo.On("UpdateGlobal", ctx, inputProductModel).Return(testErr).Once()
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
				repo.On("GetByIDGlobal", ctx, productID).Return(nil, testErr).Once()
				return repo
			},
			InputID:     productID,
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := product.NewService(repo, nil)

			// WHEN
			err := svc.UpdateGlobal(ctx, testCase.InputID, testCase.Input)

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
}

func TestService_Delete(t *testing.T) {
	// GIVEN
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
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := product.NewService(repo, nil)

			// WHEN
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

func TestService_DeleteGlobal(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()

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
				repo.On("DeleteGlobal", ctx, productID).Return(nil).Once()
				return repo
			},
			InputID:     productID,
			ExpectedErr: nil,
		},
		{
			Name: "DeleteGlobal Error",
			RepositoryFn: func() *automock.ProductRepository {
				repo := &automock.ProductRepository{}
				repo.On("DeleteGlobal", ctx, productID).Return(testErr).Once()
				return repo
			},
			InputID:     productID,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := product.NewService(repo, nil)

			// WHEN
			err := svc.DeleteGlobal(ctx, testCase.InputID)

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
	// GIVEN
	testErr := errors.New("Test error")

	productModel := fixProductModelForApp()

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

			// WHEN
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
	// GIVEN
	testErr := errors.New("Test error")

	products := []*model.Product{
		fixProductModelForApp(),
		fixProductModelForApp(),
		fixProductModelForApp(),
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
				repo.On("ListByResourceID", ctx, tenantID, appID, resource.Application).Return(products, nil).Once()
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

			svc := product.NewService(repo, nil)

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
		svc := product.NewService(nil, nil)
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

	products := []*model.Product{
		fixProductModelForAppTemplateVersion(),
		fixProductModelForAppTemplateVersion(),
	}

	ctx := context.TODO()

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
				repo.On("ListByResourceID", ctx, "", appTemplateVersionID, resource.ApplicationTemplateVersion).Return(products, nil).Once()
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

			svc := product.NewService(repo, nil)

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
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := product.NewService(nil, nil)
		// WHEN
		_, err := svc.ListByApplicationID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListGlobal(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	products := []*model.Product{
		fixGlobalProductModel(),
		fixGlobalProductModel(),
		fixGlobalProductModel(),
	}

	ctx := context.TODO()

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
				repo.On("ListGlobal", ctx).Return(products, nil).Once()
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
				repo.On("ListGlobal", ctx).Return(nil, testErr).Once()
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

			// WHEN
			docs, err := svc.ListGlobal(ctx)

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
	svc := &automock.UIDService{}
	svc.On("Generate").Return(productID).Once()
	return svc
}
