package dataproduct_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/dataproduct"
	"github.com/kyma-incubator/compass/components/director/internal/domain/dataproduct/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_ListByApplicationID(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	dataProducts := []*model.DataProduct{fixDataProductModel(dataProductID)}
	applicationID := "application-id"
	testCases := []struct {
		Name              string
		InputID           string
		DataProductRepoFn func() *automock.DataProductRepository
		ExpectedOutput    []*model.DataProduct
		ExpectedError     error
	}{
		{
			Name:    "Success",
			InputID: applicationID,
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("ListByResourceID", ctx, tenantID, resource.Application, applicationID).Return(dataProducts, nil).Once()
				return dataProductRepo
			},
			ExpectedOutput: dataProducts,
		},
		{
			Name:    "Fail while listing by resource id",
			InputID: applicationID,
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("ListByResourceID", ctx, tenantID, resource.Application, applicationID).Return(nil, testErr).Once()
				return dataProductRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			dataProductRepo := testCase.DataProductRepoFn()
			svc := dataproduct.NewService(dataProductRepo, nil)

			// WHEN
			dataProducts, err := svc.ListByApplicationID(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, dataProducts)
			}

			dataProductRepo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := dataproduct.NewService(nil, uid.NewService())
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
	dataProducts := []*model.DataProduct{fixDataProductModel(dataProductID)}
	applicationTemplateVersionID := "application-template-version-id"
	testCases := []struct {
		Name              string
		InputID           string
		DataProductRepoFn func() *automock.DataProductRepository
		ExpectedOutput    []*model.DataProduct
		ExpectedError     error
	}{
		{
			Name:    "Success",
			InputID: applicationTemplateVersionID,
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("ListByResourceID", ctx, "", resource.ApplicationTemplateVersion, applicationTemplateVersionID).Return(dataProducts, nil).Once()
				return dataProductRepo
			},
			ExpectedOutput: dataProducts,
		},
		{
			Name:    "Fail while listing by resource id",
			InputID: applicationTemplateVersionID,
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("ListByResourceID", ctx, "", resource.ApplicationTemplateVersion, applicationTemplateVersionID).Return(nil, testErr).Once()
				return dataProductRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			dataProductRepo := testCase.DataProductRepoFn()
			svc := dataproduct.NewService(dataProductRepo, nil)

			// WHEN
			dataProducts, err := svc.ListByApplicationTemplateVersionID(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, dataProducts)
			}

			dataProductRepo.AssertExpectations(t)
		})
	}
}

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(dataProductID)
		return uidSvc
	}

	testCases := []struct {
		Name              string
		InputResourceType resource.Type
		InputResourceID   string
		DataProductInput  model.DataProductInput
		DataProductRepoFn func() *automock.DataProductRepository
		UIDServiceFn      func() *automock.UIDService
		ExpectedError     error
		ExpectedOutput    string
	}{
		{
			Name:              "Success with resource type Application",
			InputResourceType: resource.Application,
			InputResourceID:   "application-id",
			DataProductInput:  fixDataProductInputModelWithPackageOrdID(packageID),
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("Create", ctx, tenantID, mock.Anything).Return(nil).Once()
				return dataProductRepo
			},
			UIDServiceFn:   uidSvcFn,
			ExpectedOutput: dataProductID,
		},
		{
			Name:              "Success with resource type ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputResourceID:   "application-template-version-id",
			DataProductInput:  fixDataProductInputModelWithPackageOrdID(packageID),
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("CreateGlobal", ctx, mock.Anything).Return(nil).Once()
				return dataProductRepo
			},
			UIDServiceFn:   uidSvcFn,
			ExpectedOutput: dataProductID,
		},
		{
			Name:              "Fail while creating Data Product for Application",
			InputResourceType: resource.Application,
			InputResourceID:   "application-id",
			DataProductInput:  fixDataProductInputModelWithPackageOrdID(packageID),
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("Create", ctx, tenantID, mock.Anything).Return(testErr).Once()
				return dataProductRepo
			},
			UIDServiceFn:  uidSvcFn,
			ExpectedError: testErr,
		},
		{
			Name:              "Fail while creating Data Product for ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputResourceID:   "application-template-version-id",
			DataProductInput:  fixDataProductInputModelWithPackageOrdID(packageID),
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("CreateGlobal", ctx, mock.Anything).Return(testErr).Once()
				return dataProductRepo
			},
			UIDServiceFn:  uidSvcFn,
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			dataProductRepo := testCase.DataProductRepoFn()
			idSvc := testCase.UIDServiceFn()
			svc := dataproduct.NewService(dataProductRepo, idSvc)

			// WHEN
			result, err := svc.Create(ctx, testCase.InputResourceType, testCase.InputResourceID, str.Ptr(packageID), testCase.DataProductInput, 123)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, result)
			}

			dataProductRepo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := dataproduct.NewService(nil, uid.NewService())
		// WHEN
		_, err := svc.Create(context.TODO(), "", "", nil, model.DataProductInput{}, 123)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)

	testCases := []struct {
		Name              string
		InputResourceType resource.Type
		InputID           string
		InputPackageID    *string
		DataProductInput  model.DataProductInput
		DataProductRepoFn func() *automock.DataProductRepository
		ExpectedError     error
		ExpectedOutput    string
	}{
		{
			Name:              "Success with resource type Application",
			InputResourceType: resource.Application,
			InputID:           appID,
			DataProductInput:  fixDataProductInputModelWithPackageOrdID(packageID),
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("Update", ctx, tenantID, mock.Anything).Return(nil).Once()
				return dataProductRepo
			},
			ExpectedOutput: dataProductID,
		},
		{
			Name:              "Success with resource type ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputID:           appID,
			DataProductInput:  fixDataProductInputModelWithPackageOrdID(packageID),
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("UpdateGlobal", ctx, mock.Anything).Return(nil).Once()
				return dataProductRepo
			},
			ExpectedOutput: dataProductID,
		},
		{
			Name:              "Fail while updating Data Product for Application",
			InputResourceType: resource.Application,
			InputID:           appID,
			DataProductInput:  fixDataProductInputModelWithPackageOrdID(packageID),
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("Update", ctx, tenantID, mock.Anything).Return(testErr).Once()

				return dataProductRepo
			},
			ExpectedError: testErr,
		},
		{
			Name:              "Fail while updating Data Product for ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputID:           appID,
			DataProductInput:  fixDataProductInputModelWithPackageOrdID(packageID),
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("UpdateGlobal", ctx, mock.Anything).Return(testErr).Once()

				return dataProductRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			dataProductRepo := testCase.DataProductRepoFn()
			svc := dataproduct.NewService(dataProductRepo, nil)
			testCase.InputPackageID = str.Ptr(packageID)
			// WHEN
			err := svc.Update(ctx, testCase.InputResourceType, testCase.InputID, dataProductID, testCase.InputPackageID, testCase.DataProductInput, 123)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			dataProductRepo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := dataproduct.NewService(nil, uid.NewService())
		// WHEN
		err := svc.Update(context.TODO(), "", "", "", nil, model.DataProductInput{}, 123)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Get(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)

	testCases := []struct {
		Name              string
		InputID           string
		DataProductRepoFn func() *automock.DataProductRepository
		ExpectedOutput    *model.DataProduct
		ExpectedError     error
	}{
		{
			Name:    "Success",
			InputID: dataProductID,
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("GetByID", ctx, tenantID, dataProductID).Return(fixDataProductModel(dataProductID), nil).Once()
				return dataProductRepo
			},
			ExpectedOutput: fixDataProductModel(dataProductID),
		},
		{
			Name:    "Fail while getting Data Product",
			InputID: dataProductID,
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("GetByID", ctx, tenantID, dataProductID).Return(nil, testErr).Once()
				return dataProductRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			dataProductRepo := testCase.DataProductRepoFn()
			svc := dataproduct.NewService(dataProductRepo, nil)

			// WHEN
			dataProduct, err := svc.Get(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, dataProduct)
			}

			dataProductRepo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := dataproduct.NewService(nil, uid.NewService())
		// WHEN
		_, err := svc.Get(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)

	testCases := []struct {
		Name              string
		InputResourceType resource.Type
		InputID           string
		DataProductRepoFn func() *automock.DataProductRepository
		ExpectedError     error
	}{
		{
			Name:              "Success with resource type Application",
			InputResourceType: resource.Application,
			InputID:           dataProductID,
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("Delete", ctx, tenantID, dataProductID).Return(nil).Once()
				return dataProductRepo
			},
		},
		{
			Name:              "Success with resource type ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputID:           dataProductID,
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("DeleteGlobal", ctx, dataProductID).Return(nil).Once()
				return dataProductRepo
			},
		},
		{
			Name:              "Fail while deleting Data Product for Application",
			InputResourceType: resource.Application,
			InputID:           dataProductID,
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("Delete", ctx, tenantID, dataProductID).Return(testErr).Once()
				return dataProductRepo
			},
			ExpectedError: testErr,
		},
		{
			Name:              "Fail while deleting Data Product for ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputID:           dataProductID,
			DataProductRepoFn: func() *automock.DataProductRepository {
				dataProductRepo := &automock.DataProductRepository{}
				dataProductRepo.On("DeleteGlobal", ctx, dataProductID).Return(testErr).Once()
				return dataProductRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			dataProductRepo := testCase.DataProductRepoFn()
			svc := dataproduct.NewService(dataProductRepo, nil)

			// WHEN
			err := svc.Delete(ctx, testCase.InputResourceType, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			dataProductRepo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := dataproduct.NewService(nil, uid.NewService())
		// WHEN
		err := svc.Delete(context.TODO(), "", "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}
