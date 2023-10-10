package capability_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/capability"
	"github.com/kyma-incubator/compass/components/director/internal/domain/capability/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestService_ListByApplicationID(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	name := "foo"

	capabilities := []*model.Capability{
		fixCapabilityModel(id, name),
		fixCapabilityModel(id, name),
		fixCapabilityModel(id, name),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.CapabilityRepository
		ExpectedResult     []*model.Capability
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("ListByResourceID", ctx, tenantID, resource.Application, appID).Return(capabilities, nil).Once()
				return repo
			},
			ExpectedResult:     capabilities,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Capability listing failed",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("ListByResourceID", ctx, tenantID, resource.Application, appID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := capability.NewService(repo, nil, nil)

			// WHEN
			docs, err := svc.ListByApplicationID(ctx, appID)

			// THEN
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
		svc := capability.NewService(nil, nil, nil)
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

	id := "foo"
	name := "foo"

	capabilities := []*model.Capability{
		fixCapabilityModel(id, name),
		fixCapabilityModel(id, name),
		fixCapabilityModel(id, name),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.CapabilityRepository
		ExpectedResult     []*model.Capability
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("ListByResourceID", ctx, "", resource.ApplicationTemplateVersion, appTemplateVersionID).Return(capabilities, nil).Once()
				return repo
			},
			ExpectedResult:     capabilities,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Capability listing failed",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("ListByResourceID", ctx, "", resource.ApplicationTemplateVersion, appTemplateVersionID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := capability.NewService(repo, nil, nil)

			// WHEN
			docs, err := svc.ListByApplicationTemplateVersionID(ctx, appTemplateVersionID)

			// THEN
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
		svc := capability.NewService(nil, nil, nil)
		// WHEN
		_, err := svc.ListByApplicationID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Get(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	name := "foo"

	capabilityModel := fixCapabilityModel(id, name)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.CapabilityRepository
		Input              model.CapabilityInput
		InputID            string
		ExpectedDocument   *model.Capability
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(capabilityModel, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedDocument:   capabilityModel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Capability retrieval failed",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedDocument:   capabilityModel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := capability.NewService(repo, nil, nil)

			// WHEN
			document, err := svc.Get(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedDocument, document)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := capability.NewService(nil, nil, nil)
		// WHEN
		_, err := svc.Get(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Create(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	name := "Foo"

	fixedTimestamp := time.Now()
	frURL := "foo.bar"
	spec := "test"
	spec2 := "test2"

	modelInput := model.CapabilityInput{
		Name:         name,
		Type:         CapabilityTypeMDICapabilityDefinitionV1,
		VersionInput: &model.VersionInput{},
	}

	modelSpecsInput := []*model.SpecInput{
		{
			Data: &spec,
			FetchRequest: &model.FetchRequestInput{
				URL: frURL,
			},
		},
		{
			Data: &spec2,
			FetchRequest: &model.FetchRequestInput{
				URL: frURL,
			},
		},
	}

	modelCapabilityForApp := fixCapabilityWithPackageModel(id, name)
	modelCapabilityForApp.ApplicationID = &appID

	modelCapabilityForAppTemplateVersion := fixCapabilityWithPackageModel(id, name)
	modelCapabilityForAppTemplateVersion.ApplicationTemplateVersionID = &appTemplateVersionID

	ctx := context.TODO()
	ctxWithTenant := tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name          string
		RepositoryFn  func() *automock.CapabilityRepository
		UIDServiceFn  func() *automock.UIDService
		SpecServiceFn func() *automock.SpecService
		Input         model.CapabilityInput
		SpecsInput    []*model.SpecInput
		ResourceType  resource.Type
		ResourceID    string
		Ctx           context.Context
		ExpectedErr   error
	}{
		{
			Name: "Success for application",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("Create", ctxWithTenant, tenantID, modelCapabilityForApp).Return(nil).Once()
				return repo
			},
			UIDServiceFn: fixUIDService(id),
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *modelSpecsInput[0], resource.Application, model.CapabilitySpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *modelSpecsInput[1], resource.Application, model.CapabilitySpecReference, id).Return("id", nil).Once()
				return svc
			},
			ResourceType: resource.Application,
			ResourceID:   appID,
			Input:        modelInput,
			SpecsInput:   modelSpecsInput,
		},
		{
			Name: "Success for application template version",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("CreateGlobal", ctx, modelCapabilityForAppTemplateVersion).Return(nil).Once()
				return repo
			},
			UIDServiceFn: fixUIDService(id),
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], resource.ApplicationTemplateVersion, model.CapabilitySpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[1], resource.ApplicationTemplateVersion, model.CapabilitySpecReference, id).Return("id", nil).Once()
				return svc
			},
			ResourceType: resource.ApplicationTemplateVersion,
			ResourceID:   appTemplateVersionID,
			Input:        modelInput,
			Ctx:          ctx,
			SpecsInput:   modelSpecsInput,
		},
		{
			Name: "Error creating Capability",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("Create", ctxWithTenant, tenantID, modelCapabilityForApp).Return(testErr).Once()
				return repo
			},
			UIDServiceFn:  fixUIDService(id),
			SpecServiceFn: emptySpecService,
			ResourceType:  resource.Application,
			ResourceID:    appID,
			Input:         modelInput,
			SpecsInput:    modelSpecsInput,
			ExpectedErr:   testErr,
		},
		{
			Name: "Error creating Capability for Application Template Version",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("CreateGlobal", ctx, modelCapabilityForAppTemplateVersion).Return(testErr).Once()
				return repo
			},
			UIDServiceFn:  fixUIDService(id),
			SpecServiceFn: emptySpecService,
			ResourceType:  resource.ApplicationTemplateVersion,
			ResourceID:    appTemplateVersionID,
			Input:         modelInput,
			Ctx:           ctx,
			SpecsInput:    modelSpecsInput,
			ExpectedErr:   testErr,
		},
		{
			Name: "Error creating Spec",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("Create", ctxWithTenant, tenantID, modelCapabilityForApp).Return(nil).Once()
				return repo
			},
			UIDServiceFn: fixUIDService(id),
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *modelSpecsInput[0], resource.Application, model.CapabilitySpecReference, id).Return("", testErr).Once()
				return svc
			},
			ResourceType: resource.Application,
			ResourceID:   appID,
			Input:        modelInput,
			SpecsInput:   modelSpecsInput,
			ExpectedErr:  testErr,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			uidService := testCase.UIDServiceFn()
			specService := testCase.SpecServiceFn()

			svc := capability.NewService(repo, uidService, specService)
			svc.SetTimestampGen(func() time.Time { return fixedTimestamp })

			defaultCtx := ctxWithTenant
			if testCase.Ctx != nil {
				defaultCtx = testCase.Ctx
			}

			// WHEN
			result, err := svc.Create(defaultCtx, testCase.ResourceType, testCase.ResourceID, str.Ptr(packageID), testCase.Input, testCase.SpecsInput, 0)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.IsType(t, "string", result)
			}

			repo.AssertExpectations(t)
			specService.AssertExpectations(t)
			uidService.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := capability.NewService(nil, fixUIDService(id)(), nil)
		// WHEN
		_, err := svc.Create(context.TODO(), "", "", nil, model.CapabilityInput{}, []*model.SpecInput{}, 0)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	spec := "spec"
	frURL := "foo.bar"

	modelInput := model.CapabilityInput{
		Name:         "foo",
		Type:         CapabilityTypeMDICapabilityDefinitionV1,
		VersionInput: &model.VersionInput{},
	}

	modelSpecInput := model.SpecInput{
		Data: &spec,
		FetchRequest: &model.FetchRequestInput{
			URL: frURL,
		},
	}

	modelSpec := &model.Spec{
		ID:         id,
		ObjectType: model.CapabilitySpecReference,
		ObjectID:   id,
		Data:       &spec,
	}

	inputCapabilityModel := mock.MatchedBy(func(capability *model.Capability) bool {
		return capability.Name == modelInput.Name
	})

	capabilityModelForApp := &model.Capability{
		Name:          "Bar",
		Version:       &model.Version{},
		ApplicationID: &appID,
	}

	capabilityModelForAppTemplateVersion := &model.Capability{
		Name:                         "Bar",
		Version:                      &model.Version{},
		ApplicationTemplateVersionID: &appTemplateVersionID,
	}

	ctx := context.TODO()
	ctxWithTenant := tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name            string
		RepositoryFn    func() *automock.CapabilityRepository
		SpecServiceFn   func() *automock.SpecService
		Input           model.CapabilityInput
		InputID         string
		SpecInput       *model.SpecInput
		DefaultBundleID string
		ResourceType    resource.Type
		Ctx             context.Context
		ExpectedErr     error
	}{
		{
			Name: "Success for Application",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("GetByID", ctxWithTenant, tenantID, id).Return(capabilityModelForApp, nil).Once()
				repo.On("Update", ctxWithTenant, tenantID, inputCapabilityModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctxWithTenant, resource.Application, model.CapabilitySpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctxWithTenant, id, modelSpecInput, resource.Application, model.CapabilitySpecReference, id).Return(nil).Once()
				return svc
			},
			InputID:      "foo",
			Input:        modelInput,
			SpecInput:    &modelSpecInput,
			ResourceType: resource.Application,
			ExpectedErr:  nil,
		},
		{
			Name: "Success for Application Template Version",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("GetByIDGlobal", ctx, id).Return(capabilityModelForAppTemplateVersion, nil).Once()
				repo.On("UpdateGlobal", ctx, inputCapabilityModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, resource.ApplicationTemplateVersion, model.CapabilitySpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctx, id, modelSpecInput, resource.ApplicationTemplateVersion, model.CapabilitySpecReference, id).Return(nil).Once()
				return svc
			},
			InputID:      "foo",
			Input:        modelInput,
			SpecInput:    &modelSpecInput,
			ResourceType: resource.ApplicationTemplateVersion,
			Ctx:          ctx,
			ExpectedErr:  nil,
		},
		{
			Name: "Error while updating Capability",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("GetByID", ctxWithTenant, tenantID, id).Return(capabilityModelForApp, nil).Once()
				repo.On("Update", ctxWithTenant, tenantID, inputCapabilityModel).Return(testErr).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			InputID:      "foo",
			Input:        modelInput,
			SpecInput:    &modelSpecInput,
			ResourceType: resource.Application,
			ExpectedErr:  testErr,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			specSvc := testCase.SpecServiceFn()

			svc := capability.NewService(repo, nil, specSvc)
			svc.SetTimestampGen(func() time.Time { return fixedTimestamp })

			defaultCtx := ctxWithTenant
			if testCase.Ctx != nil {
				defaultCtx = testCase.Ctx
			}

			// WHEN
			err := svc.Update(defaultCtx, testCase.ResourceType, testCase.InputID, testCase.Input, testCase.SpecInput, 0)

			// then
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
			specSvc.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := capability.NewService(nil, nil, nil)
		// WHEN
		err := svc.Update(context.TODO(), resource.Application, "", model.CapabilityInput{}, &model.SpecInput{}, 0)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.CapabilityRepository
		InputID      string
		ResourceType resource.Type
		ExpectedErr  error
	}{
		{
			Name: "Success for Application",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("Delete", ctx, tenantID, id).Return(nil).Once()
				return repo
			},
			InputID:      id,
			ResourceType: resource.Application,
			ExpectedErr:  nil,
		},
		{
			Name: "Success for Application Template Version",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("DeleteGlobal", ctx, id).Return(nil).Once()
				return repo
			},
			InputID:      id,
			ResourceType: resource.ApplicationTemplateVersion,
			ExpectedErr:  nil,
		},
		{
			Name: "Error deleting for Application",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("Delete", ctx, tenantID, id).Return(testErr).Once()
				return repo
			},
			InputID:      id,
			ResourceType: resource.Application,
			ExpectedErr:  testErr,
		},
		{
			Name: "Error deleting for Application Template Version",
			RepositoryFn: func() *automock.CapabilityRepository {
				repo := &automock.CapabilityRepository{}
				repo.On("DeleteGlobal", ctx, id).Return(testErr).Once()
				return repo
			},
			InputID:      id,
			ResourceType: resource.ApplicationTemplateVersion,
			ExpectedErr:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := capability.NewService(repo, nil, nil)

			// WHEN
			err := svc.Delete(ctx, testCase.ResourceType, testCase.InputID)

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
		svc := capability.NewService(nil, nil, nil)
		// WHEN
		err := svc.Delete(context.TODO(), resource.Application, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func fixUIDService(id string) func() *automock.UIDService {
	return func() *automock.UIDService {
		svc := &automock.UIDService{}
		svc.On("Generate").Return(id).Once()
		return svc
	}
}

func emptySpecService() *automock.SpecService {
	svc := &automock.SpecService{}
	return svc
}
