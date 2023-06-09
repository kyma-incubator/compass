package ordvendor_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor"
	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor/automock"
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

	modelVendorForApp := fixVendorModelForApp()
	modelVendorForAppTemplateVersion := fixVendorModelForAppTemplateVersion()
	modelInput := *fixVendorModelInput()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.VendorRepository
		UIDServiceFn func() *automock.UIDService
		Input        model.VendorInput
		ResourceType resource.Type
		ResourceID   string
		ExpectedErr  error
	}{
		{
			Name: "Success for Application",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("Create", ctx, tenantID, modelVendorForApp).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(vendorID)
				return svc
			},
			Input:        modelInput,
			ResourceType: resource.Application,
			ResourceID:   appID,
			ExpectedErr:  nil,
		},
		{
			Name: "Success for Application Template Version",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("CreateGlobal", ctx, modelVendorForAppTemplateVersion).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(vendorID)
				return svc
			},
			Input:        modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ResourceID:   appTemplateVersionID,
			ExpectedErr:  nil,
		},
		{
			Name: "Error - Vendor creation for Application",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("Create", ctx, tenantID, modelVendorForApp).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(vendorID)
				return svc
			},
			Input:        modelInput,
			ResourceType: resource.Application,
			ResourceID:   appID,
			ExpectedErr:  testErr,
		},
		{
			Name: "Error - Vendor creation for Application Template Version",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("CreateGlobal", ctx, modelVendorForAppTemplateVersion).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(vendorID)
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

			svc := ordvendor.NewService(repo, uidSvc)

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
		svc := ordvendor.NewService(nil, fixUIDService())
		// WHEN
		_, err := svc.Create(context.TODO(), resource.Application, "", model.VendorInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_CreateGlobal(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()

	modelVendor := fixGlobalVendorModel()
	modelInput := *fixVendorModelInput()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.VendorRepository
		UIDServiceFn func() *automock.UIDService
		Input        model.VendorInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("CreateGlobal", ctx, modelVendor).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(vendorID)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Error - Vendor creation",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("CreateGlobal", ctx, modelVendor).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(vendorID)
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

			svc := ordvendor.NewService(repo, uidSvc)

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

	modelVendorForApp := fixVendorModelForApp()
	modelVendorForAppTemplateVersion := fixVendorModelForAppTemplateVersion()
	modelInput := *fixVendorModelInput()

	inputVendorModel := mock.MatchedBy(func(vendor *model.Vendor) bool {
		return vendor.Title == modelInput.Title
	})

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.VendorRepository
		Input        model.VendorInput
		InputID      string
		ResourceType resource.Type
		ExpectedErr  error
	}{
		{
			Name: "Success for Application",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("GetByID", ctx, tenantID, vendorID).Return(modelVendorForApp, nil).Once()
				repo.On("Update", ctx, tenantID, inputVendorModel).Return(nil).Once()
				return repo
			},
			InputID:      vendorID,
			Input:        modelInput,
			ResourceType: resource.Application,
			ExpectedErr:  nil,
		},
		{
			Name: "Success for Application Template Version",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("GetByIDGlobal", ctx, vendorID).Return(modelVendorForAppTemplateVersion, nil).Once()
				repo.On("UpdateGlobal", ctx, inputVendorModel).Return(nil).Once()
				return repo
			},
			InputID:      vendorID,
			Input:        modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ExpectedErr:  nil,
		},
		{
			Name: "Update Error for Application",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("GetByID", ctx, tenantID, vendorID).Return(modelVendorForApp, nil).Once()
				repo.On("Update", ctx, tenantID, inputVendorModel).Return(testErr).Once()
				return repo
			},
			InputID:      vendorID,
			Input:        modelInput,
			ResourceType: resource.Application,
			ExpectedErr:  testErr,
		},
		{
			Name: "Update Error for Application Template Version",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("GetByIDGlobal", ctx, vendorID).Return(modelVendorForApp, nil).Once()
				repo.On("UpdateGlobal", ctx, inputVendorModel).Return(testErr).Once()
				return repo
			},
			InputID:      vendorID,
			Input:        modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ExpectedErr:  testErr,
		},
		{
			Name: "Get Error for Application",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("GetByID", ctx, tenantID, vendorID).Return(nil, testErr).Once()
				return repo
			},
			InputID:      vendorID,
			Input:        modelInput,
			ResourceType: resource.Application,
			ExpectedErr:  testErr,
		},
		{
			Name: "Get Error for Application Template Version",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("GetByIDGlobal", ctx, vendorID).Return(nil, testErr).Once()
				return repo
			},
			InputID:      vendorID,
			Input:        modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ExpectedErr:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := ordvendor.NewService(repo, nil)

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
		svc := ordvendor.NewService(nil, fixUIDService())
		// WHEN
		err := svc.Update(context.TODO(), resource.Application, "", model.VendorInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_UpdateGlobal(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	modelVendor := fixGlobalVendorModel()
	modelInput := *fixVendorModelInput()

	inputVendorModel := mock.MatchedBy(func(vendor *model.Vendor) bool {
		return vendor.Title == modelInput.Title
	})

	ctx := context.TODO()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.VendorRepository
		Input        model.VendorInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("GetByIDGlobal", ctx, vendorID).Return(modelVendor, nil).Once()
				repo.On("UpdateGlobal", ctx, inputVendorModel).Return(nil).Once()
				return repo
			},
			InputID:     vendorID,
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "UpdateGlobal Error",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("GetByIDGlobal", ctx, vendorID).Return(modelVendor, nil).Once()
				repo.On("UpdateGlobal", ctx, inputVendorModel).Return(testErr).Once()
				return repo
			},
			InputID:     vendorID,
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("GetByIDGlobal", ctx, vendorID).Return(nil, testErr).Once()
				return repo
			},
			InputID:     vendorID,
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := ordvendor.NewService(repo, nil)

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
		RepositoryFn func() *automock.VendorRepository
		Input        model.VendorInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("Delete", ctx, tenantID, vendorID).Return(nil).Once()
				return repo
			},
			InputID:     vendorID,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("Delete", ctx, tenantID, vendorID).Return(testErr).Once()
				return repo
			},
			InputID:     vendorID,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := ordvendor.NewService(repo, nil)

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
		svc := ordvendor.NewService(nil, nil)
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
		RepositoryFn func() *automock.VendorRepository
		Input        model.VendorInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("DeleteGlobal", ctx, vendorID).Return(nil).Once()
				return repo
			},
			InputID:     vendorID,
			ExpectedErr: nil,
		},
		{
			Name: "DeleteGlobal Error",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("DeleteGlobal", ctx, vendorID).Return(testErr).Once()
				return repo
			},
			InputID:     vendorID,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := ordvendor.NewService(repo, nil)

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
		RepoFn         func() *automock.VendorRepository
		ExpectedError  error
		ExpectedOutput bool
	}{
		{
			Name: "Success",
			RepoFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("Exists", ctx, tenantID, vendorID).Return(true, nil).Once()
				return repo
			},
			ExpectedOutput: true,
		},
		{
			Name: "Error when getting Vendor",
			RepoFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("Exists", ctx, tenantID, vendorID).Return(false, testErr).Once()
				return repo
			},
			ExpectedError:  testErr,
			ExpectedOutput: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepoFn()
			svc := ordvendor.NewService(repo, nil)

			// WHEN
			result, err := svc.Exist(ctx, vendorID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			repo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := ordvendor.NewService(nil, nil)
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

	vendor := fixVendorModelForApp()

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.VendorRepository
		Input              model.VendorInput
		InputID            string
		ExpectedVendor     *model.Vendor
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("GetByID", ctx, tenantID, vendorID).Return(vendor, nil).Once()
				return repo
			},
			InputID:            vendorID,
			ExpectedVendor:     vendor,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Vendor retrieval failed",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("GetByID", ctx, tenantID, vendorID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            vendorID,
			ExpectedVendor:     vendor,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := ordvendor.NewService(repo, nil)

			// WHEN
			vendor, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedVendor, vendor)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := ordvendor.NewService(nil, nil)
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

	vendors := []*model.Vendor{
		fixVendorModelForApp(),
		fixVendorModelForApp(),
		fixVendorModelForApp(),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.VendorRepository
		ExpectedResult     []*model.Vendor
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("ListByResourceID", ctx, tenantID, appID, resource.Application).Return(vendors, nil).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     vendors,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Vendor listing failed",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
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

			svc := ordvendor.NewService(repo, nil)

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
		svc := ordvendor.NewService(nil, nil)
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

	vendors := []*model.Vendor{
		fixVendorModelForAppTemplateVersion(),
		fixVendorModelForAppTemplateVersion(),
	}

	ctx := context.TODO()

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.VendorRepository
		ExpectedResult     []*model.Vendor
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("ListByResourceID", ctx, "", appTemplateVersionID, resource.ApplicationTemplateVersion).Return(vendors, nil).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     vendors,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Vendor listing failed",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
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

			svc := ordvendor.NewService(repo, nil)

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
		svc := ordvendor.NewService(nil, nil)
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

	vendors := []*model.Vendor{
		fixGlobalVendorModel(),
		fixGlobalVendorModel(),
		fixGlobalVendorModel(),
	}

	ctx := context.TODO()

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.VendorRepository
		ExpectedResult     []*model.Vendor
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
				repo.On("ListGlobal", ctx).Return(vendors, nil).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     vendors,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Vendor listing failed",
			RepositoryFn: func() *automock.VendorRepository {
				repo := &automock.VendorRepository{}
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

			svc := ordvendor.NewService(repo, nil)

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
	svc.On("Generate").Return(vendorID)
	return svc
}
