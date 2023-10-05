package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/uid"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Get(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	name := "foo"
	desc := "bar"

	apiDefinition := fixAPIDefinitionModel(id, name, desc)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.APIRepository
		Input              model.APIDefinitionInput
		InputID            string
		ExpectedDocument   *model.APIDefinition
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinition, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedDocument:   apiDefinition,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when APIDefinition retrieval failed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedDocument:   apiDefinition,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := api.NewService(repo, nil, nil, nil)

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
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.Get(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetForBundle(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	bndlID := "foobar"
	name := "foo"
	desc := "bar"

	apiDefinition := fixAPIDefinitionModel(id, name, desc)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.APIRepository
		Input              model.APIDefinitionInput
		InputID            string
		BundleID           string
		ExpectedAPI        *model.APIDefinition
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetForBundle", ctx, tenantID, id, bndlID).Return(apiDefinition, nil).Once()
				return repo
			},
			InputID:            id,
			BundleID:           bndlID,
			ExpectedAPI:        apiDefinition,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when APIDefinition retrieval failed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetForBundle", ctx, tenantID, id, bndlID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			BundleID:           bndlID,
			ExpectedAPI:        nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := api.NewService(repo, nil, nil, nil)

			// WHEN
			api, err := svc.GetForBundle(ctx, testCase.InputID, testCase.BundleID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedAPI, api)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.GetForBundle(context.TODO(), "", "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetForApplication(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	name := "foo"
	desc := "bar"

	apiDefinition := fixAPIDefinitionModel(id, name, desc)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.APIRepository
		Input              model.APIDefinitionInput
		InputID            string
		AppID              string
		ExpectedAPI        *model.APIDefinition
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByApplicationID", ctx, tenantID, id, appID).Return(apiDefinition, nil).Once()
				return repo
			},
			InputID:            id,
			AppID:              appID,
			ExpectedAPI:        apiDefinition,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when APIDefinition retrieval failed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByApplicationID", ctx, tenantID, id, appID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			AppID:              appID,
			ExpectedAPI:        nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := api.NewService(repo, nil, nil, nil)

			// WHEN
			api, err := svc.GetForApplication(ctx, testCase.InputID, testCase.AppID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedAPI, api)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.GetForApplication(context.TODO(), "", "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByBundleIDs(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	firstAPIDefID := "foo"
	secondAPIDefID := "foo2"
	firstBundleID := "bar"
	secondBundleID := "bar2"
	name := "foo"
	targetURL := "https://test.com"
	numberOfAPIsInFirstBundle := 1
	numberOfAPIsInSecondBundle := 1
	bundleIDs := []string{firstBundleID, secondBundleID}

	apiDefFirstBundle := fixAPIDefinitionModel(firstAPIDefID, name, targetURL)
	apiDefSecondBundle := fixAPIDefinitionModel(secondAPIDefID, name, targetURL)

	apiDefFirstBundleReference := fixModelBundleReference(firstBundleID, firstAPIDefID)
	apiDefSecondBundleReference := fixModelBundleReference(secondBundleID, secondAPIDefID)
	bundleRefs := []*model.BundleReference{apiDefFirstBundleReference, apiDefSecondBundleReference}
	totalCounts := map[string]int{firstBundleID: numberOfAPIsInFirstBundle, secondBundleID: numberOfAPIsInSecondBundle}

	apiDefsFirstBundle := []*model.APIDefinition{apiDefFirstBundle}
	apiDefsSecondBundle := []*model.APIDefinition{apiDefSecondBundle}

	apiDefPageFirstBundle := &model.APIDefinitionPage{
		Data:       apiDefsFirstBundle,
		TotalCount: len(apiDefsFirstBundle),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}
	apiDefPageSecondBundle := &model.APIDefinitionPage{
		Data:       apiDefsSecondBundle,
		TotalCount: len(apiDefsSecondBundle),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	apiDefPages := []*model.APIDefinitionPage{apiDefPageFirstBundle, apiDefPageSecondBundle}

	after := "test"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.APIRepository
		BundleRefSvcFn     func() *automock.BundleReferenceService
		ExpectedResult     []*model.APIDefinitionPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			BundleRefSvcFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", ctx, model.BundleAPIReference, bundleIDs, 2, after).Return(bundleRefs, totalCounts, nil).Once()
				return svc
			},
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("ListByBundleIDs", ctx, tenantID, bundleIDs, bundleRefs, totalCounts, 2, after).Return(apiDefPages, nil).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     apiDefPages,
			ExpectedErrMessage: "",
		},
		{
			Name: "Return error when page size is less than 1",
			BundleRefSvcFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				return svc
			},
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			PageSize:           0,
			ExpectedResult:     apiDefPages,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Return error when page size is bigger than 200",
			BundleRefSvcFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				return svc
			},
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			PageSize:           201,
			ExpectedResult:     apiDefPages,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when APIDefinition BundleReferences listing failed",
			BundleRefSvcFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", ctx, model.BundleAPIReference, bundleIDs, 2, after).Return(nil, nil, testErr).Once()
				return svc
			},
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			PageSize:           2,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when APIDefinition listing failed",
			BundleRefSvcFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("ListByBundleIDs", ctx, model.BundleAPIReference, bundleIDs, 2, after).Return(bundleRefs, totalCounts, nil).Once()
				return svc
			},
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("ListByBundleIDs", ctx, tenantID, bundleIDs, bundleRefs, totalCounts, 2, after).Return(nil, testErr).Once()
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
			bndlRefSvc := testCase.BundleRefSvcFn()

			svc := api.NewService(repo, nil, nil, bndlRefSvc)

			// WHEN
			apiDefs, err := svc.ListByBundleIDs(ctx, bundleIDs, testCase.PageSize, after)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, apiDefs)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListByBundleIDs(context.TODO(), nil, 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByApplicationID(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	name := "foo"
	desc := "bar"

	apiDefinitions := []*model.APIDefinition{
		fixAPIDefinitionModel(id, name, desc),
		fixAPIDefinitionModel(id, name, desc),
		fixAPIDefinitionModel(id, name, desc),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.APIRepository
		ExpectedResult     []*model.APIDefinition
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("ListByResourceID", ctx, tenantID, resource.Application, appID).Return(apiDefinitions, nil).Once()
				return repo
			},
			ExpectedResult:     apiDefinitions,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when APIDefinition listing failed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
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

			svc := api.NewService(repo, nil, nil, nil)

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
		svc := api.NewService(nil, nil, nil, nil)
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
	desc := "bar"

	apiDefinitions := []*model.APIDefinition{
		fixAPIDefinitionModel(id, name, desc),
		fixAPIDefinitionModel(id, name, desc),
		fixAPIDefinitionModel(id, name, desc),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.APIRepository
		ExpectedResult     []*model.APIDefinition
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("ListByResourceID", ctx, "", resource.ApplicationTemplateVersion, appTemplateVersionID).Return(apiDefinitions, nil).Once()
				return repo
			},
			ExpectedResult:     apiDefinitions,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when APIDefinition listing failed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
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

			svc := api.NewService(repo, nil, nil, nil)

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
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListByApplicationID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByApplicationIDPage(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	name := "foo"
	desc := "bar"
	pageSize := 1
	invalidPageSize := 0
	cursor := ""

	apiDefinitions := []*model.APIDefinition{
		fixAPIDefinitionModel(id, name, desc),
		fixAPIDefinitionModel(id, name, desc),
		fixAPIDefinitionModel(id, name, desc),
	}

	apiDefinitionsPage := &model.APIDefinitionPage{
		Data: apiDefinitions,
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
		TotalCount: len(apiDefinitions),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.APIRepository
		ExpectedResult     *model.APIDefinitionPage
		ExpectedErrMessage string
	}{
		{
			Name:     "Success",
			PageSize: pageSize,
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("ListByApplicationIDPage", ctx, tenantID, appID, pageSize, cursor).Return(apiDefinitionsPage, nil).Once()
				return repo
			},
			ExpectedResult:     apiDefinitionsPage,
			ExpectedErrMessage: "",
		},
		{
			Name:     "Returns error when APIDefinition paging failed",
			PageSize: pageSize,
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("ListByApplicationIDPage", ctx, tenantID, appID, pageSize, cursor).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:     "Returns error when pageSize is out of the range 1-200",
			PageSize: invalidPageSize,
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := api.NewService(repo, nil, nil, nil)

			// WHEN
			docs, err := svc.ListByApplicationIDPage(ctx, appID, testCase.PageSize, cursor)

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
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListByApplicationID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Create(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	bundleID := "bndlid"
	bundleID2 := "bndlid2"
	name := "Foo"
	targetURL2 := "https://test2-url.com"

	fixedTimestamp := time.Now()
	frURL := "foo.bar"
	spec := "test"
	spec2 := "test2"
	isDefaultBundle := true

	modelInput := model.APIDefinitionInput{
		Name:         name,
		TargetURLs:   api.ConvertTargetURLToJSONArray(targetURL),
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

	modelAPIDefinitionForApp := fixAPIDefinitionWithPackageModel(id, name)
	modelAPIDefinitionForApp.ApplicationID = &appID

	modelAPIDefinitionForAppTemplateVersion := fixAPIDefinitionWithPackageModel(id, name)
	modelAPIDefinitionForAppTemplateVersion.ApplicationTemplateVersionID = &appTemplateVersionID

	bundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: str.Ptr(api.ExtractTargetURLFromJSONArray(modelAPIDefinitionForApp.TargetURLs)),
	}
	secondBundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: &targetURL2,
	}

	bundleReferenceInputWithDefaultBundle := &model.BundleReferenceInput{
		APIDefaultTargetURL: str.Ptr(api.ExtractTargetURLFromJSONArray(modelAPIDefinitionForApp.TargetURLs)),
		IsDefaultBundle:     &isDefaultBundle,
	}

	singleDefaultTargetURLPerBundle := map[string]string{bundleID: targetURL}
	defaultTargetURLPerBundle := map[string]string{bundleID: targetURL, bundleID2: targetURL2}

	ctx := context.TODO()
	ctxWithTenant := tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name                      string
		RepositoryFn              func() *automock.APIRepository
		UIDServiceFn              func() *automock.UIDService
		SpecServiceFn             func() *automock.SpecService
		BundleReferenceFn         func() *automock.BundleReferenceService
		Input                     model.APIDefinitionInput
		SpecsInput                []*model.SpecInput
		ResourceType              resource.Type
		ResourceID                string
		DefaultTargetURLPerBundle map[string]string
		DefaultBundleID           string
		Ctx                       context.Context
		ExpectedErr               error
	}{
		{
			Name: "Success for application",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctxWithTenant, tenantID, modelAPIDefinitionForApp).Return(nil).Once()
				return repo
			},
			UIDServiceFn: fixUIDService(id),
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *modelSpecsInput[0], resource.Application, model.APISpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *modelSpecsInput[1], resource.Application, model.APISpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), str.Ptr(bundleID)).Return(nil).Once()
				return svc
			},
			ResourceType: resource.Application,
			ResourceID:   appID,
			Input:        modelInput,
			SpecsInput:   modelSpecsInput,
		},
		{
			Name: "Success for application template version",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("CreateGlobal", ctx, modelAPIDefinitionForAppTemplateVersion).Return(nil).Once()
				return repo
			},
			UIDServiceFn: fixUIDService(id),
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], resource.ApplicationTemplateVersion, model.APISpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[1], resource.ApplicationTemplateVersion, model.APISpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), str.Ptr(bundleID)).Return(nil).Once()
				return svc
			},
			ResourceType: resource.ApplicationTemplateVersion,
			ResourceID:   appTemplateVersionID,
			Input:        modelInput,
			Ctx:          ctx,
			SpecsInput:   modelSpecsInput,
		},
		{
			Name: "Success in ORD scenario where defaultTargetURLPerBundle map is passed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctxWithTenant, tenantID, modelAPIDefinitionForApp).Return(nil).Once()
				return repo
			},
			UIDServiceFn: fixUIDService(id),
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *modelSpecsInput[0], resource.Application, model.APISpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *modelSpecsInput[1], resource.Application, model.APISpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *bundleReferenceInputWithDefaultBundle, model.BundleAPIReference, str.Ptr(id), str.Ptr(bundleID)).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *secondBundleReferenceInput, model.BundleAPIReference, str.Ptr(id), str.Ptr(bundleID2)).Return(nil).Once()
				return svc
			},
			ResourceType:              resource.Application,
			ResourceID:                appID,
			Input:                     modelInput,
			SpecsInput:                modelSpecsInput,
			DefaultTargetURLPerBundle: defaultTargetURLPerBundle,
			DefaultBundleID:           bundleID,
		},
		{
			Name: "Error - API Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctxWithTenant, tenantID, modelAPIDefinitionForApp).Return(testErr).Once()
				return repo
			},
			UIDServiceFn:  fixUIDService(id),
			SpecServiceFn: emptySpecService,
			BundleReferenceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ResourceType: resource.Application,
			ResourceID:   appID,
			Input:        modelInput,
			SpecsInput:   modelSpecsInput,
			ExpectedErr:  testErr,
		},
		{
			Name: "Error - API Creation for Application Template Version",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("CreateGlobal", ctx, modelAPIDefinitionForAppTemplateVersion).Return(testErr).Once()
				return repo
			},
			UIDServiceFn:  fixUIDService(id),
			SpecServiceFn: emptySpecService,
			BundleReferenceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ResourceType: resource.ApplicationTemplateVersion,
			ResourceID:   appTemplateVersionID,
			Input:        modelInput,
			Ctx:          ctx,
			SpecsInput:   modelSpecsInput,
			ExpectedErr:  testErr,
		},
		{
			Name: "Error - Spec Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctxWithTenant, tenantID, modelAPIDefinitionForApp).Return(nil).Once()
				return repo
			},
			UIDServiceFn: fixUIDService(id),
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *modelSpecsInput[0], resource.Application, model.APISpecReference, id).Return("", testErr).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			ResourceType: resource.Application,
			ResourceID:   appID,
			Input:        modelInput,
			SpecsInput:   modelSpecsInput,
			ExpectedErr:  testErr,
		},
		{
			Name: "Error - BundleReference API Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctxWithTenant, tenantID, modelAPIDefinitionForApp).Return(nil).Once()
				return repo
			},
			UIDServiceFn: fixUIDService(id),
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *modelSpecsInput[0], resource.Application, model.APISpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *modelSpecsInput[1], resource.Application, model.APISpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), str.Ptr(bundleID)).Return(testErr).Once()
				return svc
			},
			ResourceType: resource.Application,
			ResourceID:   appID,
			Input:        modelInput,
			SpecsInput:   modelSpecsInput,
			ExpectedErr:  testErr,
		},
		{
			Name: "Error in ORD scenario - BundleReference API Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctxWithTenant, tenantID, modelAPIDefinitionForApp).Return(nil).Once()
				return repo
			},
			UIDServiceFn: fixUIDService(id),
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *modelSpecsInput[0], resource.Application, model.APISpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *modelSpecsInput[1], resource.Application, model.APISpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), str.Ptr(bundleID)).Return(testErr).Once()
				return svc
			},
			ResourceType:              resource.Application,
			ResourceID:                appID,
			Input:                     modelInput,
			SpecsInput:                modelSpecsInput,
			DefaultTargetURLPerBundle: singleDefaultTargetURLPerBundle,
			ExpectedErr:               testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			uidService := testCase.UIDServiceFn()
			specService := testCase.SpecServiceFn()
			bundleReferenceService := testCase.BundleReferenceFn()

			svc := api.NewService(repo, uidService, specService, bundleReferenceService)
			svc.SetTimestampGen(func() time.Time { return fixedTimestamp })

			defaultCtx := ctxWithTenant
			if testCase.Ctx != nil {
				defaultCtx = testCase.Ctx
			}

			// WHEN
			result, err := svc.Create(defaultCtx, testCase.ResourceType, testCase.ResourceID, &bundleID, str.Ptr(packageID), testCase.Input, testCase.SpecsInput, testCase.DefaultTargetURLPerBundle, 0, testCase.DefaultBundleID)

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
			bundleReferenceService.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := api.NewService(nil, fixUIDService(id)(), nil, nil)
		// WHEN
		_, err := svc.Create(context.TODO(), "", "", nil, nil, model.APIDefinitionInput{}, []*model.SpecInput{}, nil, 0, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_CreateInApplication(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	name := "Foo"
	targetURL := "https://test-url.com"

	timestamp := time.Now()
	frURL := "foo.bar"
	spec := "test"

	modelInput := model.APIDefinitionInput{
		Name:         name,
		TargetURLs:   api.ConvertTargetURLToJSONArray(targetURL),
		VersionInput: &model.VersionInput{},
	}

	modelSpecInput := &model.SpecInput{
		Data: &spec,
		FetchRequest: &model.FetchRequestInput{
			URL: frURL,
		},
	}

	modelSpecsInput := []*model.SpecInput{
		modelSpecInput,
	}

	modelAPIDefinition := &model.APIDefinition{
		ApplicationID: &appID,
		Name:          name,
		TargetURLs:    api.ConvertTargetURLToJSONArray(targetURL),
		Version:       &model.Version{},
		BaseEntity: &model.BaseEntity{
			ID:    id,
			Ready: true,
		},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name                      string
		RepositoryFn              func() *automock.APIRepository
		UIDServiceFn              func() *automock.UIDService
		SpecServiceFn             func() *automock.SpecService
		Input                     model.APIDefinitionInput
		SpecsInput                *model.SpecInput
		DefaultTargetURLPerBundle map[string]string
		DefaultBundleID           string
		ExpectedErr               error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, tenantID, modelAPIDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], resource.Application, model.APISpecReference, id).Return("id", nil).Once()
				return svc
			},
			Input:      modelInput,
			SpecsInput: modelSpecInput,
		},
		{
			Name: "Error - API Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, tenantID, modelAPIDefinition).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: emptySpecService,
			Input:         modelInput,
			SpecsInput:    modelSpecInput,
			ExpectedErr:   testErr,
		},
		{
			Name: "Error - Spec Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, tenantID, modelAPIDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], resource.Application, model.APISpecReference, id).Return("", testErr).Once()
				return svc
			},
			Input:       modelInput,
			SpecsInput:  modelSpecInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			uidService := testCase.UIDServiceFn()
			specService := testCase.SpecServiceFn()
			defer mock.AssertExpectationsForObjects(t, repo, specService, uidService)

			svc := api.NewService(repo, uidService, specService, nil)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// WHEN
			result, err := svc.CreateInApplication(ctx, appID, testCase.Input, testCase.SpecsInput)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.IsType(t, "string", result)
			}
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := api.NewService(nil, uid.NewService(), nil, nil)
		// WHEN
		_, err := svc.CreateInApplication(context.TODO(), "", model.APIDefinitionInput{}, &model.SpecInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	var bundleID *string
	fixedTimestamp := time.Now()
	frURL := "foo.bar"
	spec := "spec"

	modelInput := model.APIDefinitionInput{
		Name:         "Foo",
		TargetURLs:   api.ConvertTargetURLToJSONArray("https://test-url.com"),
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
		ObjectType: model.APISpecReference,
		ObjectID:   id,
		Data:       &spec,
	}

	inputAPIDefinitionModel := mock.MatchedBy(func(api *model.APIDefinition) bool {
		return api.Name == modelInput.Name
	})

	apiDefinitionModel := &model.APIDefinition{
		Name:       "Bar",
		TargetURLs: api.ConvertTargetURLToJSONArray("https://test-url-updated.com"),
		Version:    &model.Version{},
	}

	bundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: str.Ptr(api.ExtractTargetURLFromJSONArray(modelInput.TargetURLs)),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name              string
		RepositoryFn      func() *automock.APIRepository
		SpecServiceFn     func() *automock.SpecService
		BundleReferenceFn func() *automock.BundleReferenceService
		Input             model.APIDefinitionInput
		InputID           string
		SpecInput         *model.SpecInput
		ResourceType      resource.Type
		ExpectedErr       error
	}{
		{
			Name: "Success When Spec is Found should update it",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, resource.Application, model.APISpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctx, id, modelSpecInput, resource.Application, model.APISpecReference, id).Return(nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), bundleID).Return(nil).Once()
				return svc
			},
			InputID:      "foo",
			Input:        modelInput,
			SpecInput:    &modelSpecInput,
			ResourceType: resource.Application,
			ExpectedErr:  nil,
		},
		{
			Name: "Success When Spec is not found should create in",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, resource.Application, model.APISpecReference, id).Return(nil, nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, modelSpecInput, resource.Application, model.APISpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), bundleID).Return(nil).Once()
				return svc
			},
			InputID:      "foo",
			Input:        modelInput,
			SpecInput:    &modelSpecInput,
			ResourceType: resource.Application,
			ExpectedErr:  nil,
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, tenantID, inputAPIDefinitionModel).Return(testErr).Once()
				return repo
			},
			SpecServiceFn: emptySpecService,
			BundleReferenceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			InputID:      "foo",
			Input:        modelInput,
			SpecInput:    &modelSpecInput,
			ResourceType: resource.Application,
			ExpectedErr:  testErr,
		},
		{
			Name: "Get Spec Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, resource.Application, model.APISpecReference, id).Return(nil, testErr).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), bundleID).Return(nil).Once()
				return svc
			},
			InputID:      "foo",
			Input:        modelInput,
			SpecInput:    &modelSpecInput,
			ResourceType: resource.Application,
			ExpectedErr:  testErr,
		},
		{
			Name: "Spec Creation Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, resource.Application, model.APISpecReference, id).Return(nil, nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, modelSpecInput, resource.Application, model.APISpecReference, id).Return("", testErr).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), bundleID).Return(nil).Once()
				return svc
			},
			InputID:      "foo",
			Input:        modelInput,
			SpecInput:    &modelSpecInput,
			ResourceType: resource.Application,
			ExpectedErr:  testErr,
		},
		{
			Name: "Spec Update Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, resource.Application, model.APISpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctx, id, modelSpecInput, resource.Application, model.APISpecReference, id).Return(testErr).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), bundleID).Return(nil).Once()
				return svc
			},
			InputID:      "foo",
			Input:        modelInput,
			SpecInput:    &modelSpecInput,
			ResourceType: resource.Application,
			ExpectedErr:  testErr,
		},
		{
			Name: "BundleReference API Update Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: emptySpecService,
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), bundleID).Return(testErr).Once()
				return svc
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
			bundleReferenceSvc := testCase.BundleReferenceFn()

			svc := api.NewService(repo, nil, specSvc, bundleReferenceSvc)
			svc.SetTimestampGen(func() time.Time { return fixedTimestamp })

			// WHEN
			err := svc.Update(ctx, testCase.ResourceType, testCase.InputID, testCase.Input, testCase.SpecInput)

			// THEN
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
			specSvc.AssertExpectations(t)
			bundleReferenceSvc.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.Update(context.TODO(), "", "", model.APIDefinitionInput{}, &model.SpecInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_UpdateInManyBundles(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	isDefaultBundle := true

	modelInput := model.APIDefinitionInput{
		Name:         "Foo",
		TargetURLs:   api.ConvertTargetURLToJSONArray(firstTargetURL),
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
		ObjectType: model.APISpecReference,
		ObjectID:   id,
		Data:       &spec,
	}

	inputAPIDefinitionModel := mock.MatchedBy(func(api *model.APIDefinition) bool {
		return api.Name == modelInput.Name
	})

	apiDefinitionModelForApp := &model.APIDefinition{
		Name:          "Bar",
		TargetURLs:    api.ConvertTargetURLToJSONArray("https://test-url-updated.com"),
		Version:       &model.Version{},
		ApplicationID: &appID,
	}

	apiDefinitionModelForAppTemplateVersion := &model.APIDefinition{
		Name:                         "Bar",
		TargetURLs:                   api.ConvertTargetURLToJSONArray("https://test-url-updated.com"),
		Version:                      &model.Version{},
		ApplicationTemplateVersionID: &appTemplateVersionID,
	}

	bundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: str.Ptr(api.ExtractTargetURLFromJSONArray(modelInput.TargetURLs)),
	}
	secondBundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: str.Ptr(secondTargetURL),
	}

	bundleReferenceInputWithDefaultBundle := &model.BundleReferenceInput{
		APIDefaultTargetURL: str.Ptr(api.ExtractTargetURLFromJSONArray(modelInput.TargetURLs)),
		IsDefaultBundle:     &isDefaultBundle,
	}

	secondBundleReferenceInputWithDefaultBundle := &model.BundleReferenceInput{
		APIDefaultTargetURL: str.Ptr(secondTargetURL),
		IsDefaultBundle:     &isDefaultBundle,
	}

	defaultTargetURLPerBundleForUpdate := map[string]string{firstBundleID: firstTargetURL}
	defaultTargetURLPerBundleForCreation := map[string]string{secondBundleID: secondTargetURL}
	bundleIDsForDeletion := []string{thirdBundleID}

	ctx := context.TODO()
	ctxWithTenant := tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name                                 string
		RepositoryFn                         func() *automock.APIRepository
		SpecServiceFn                        func() *automock.SpecService
		BundleReferenceFn                    func() *automock.BundleReferenceService
		Input                                model.APIDefinitionInput
		InputID                              string
		SpecInput                            *model.SpecInput
		DefaultTargetURLPerBundleForUpdate   map[string]string
		DefaultTargetURLPerBundleForCreation map[string]string
		BundleIDsForDeletion                 []string
		DefaultBundleID                      string
		ResourceType                         resource.Type
		Ctx                                  context.Context
		ExpectedErr                          error
	}{
		{
			Name: "Success in ORD case for Application",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctxWithTenant, tenantID, id).Return(apiDefinitionModelForApp, nil).Once()
				repo.On("Update", ctxWithTenant, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctxWithTenant, resource.Application, model.APISpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctxWithTenant, id, modelSpecInput, resource.Application, model.APISpecReference, id).Return(nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctxWithTenant, *bundleReferenceInputWithDefaultBundle, model.BundleAPIReference, str.Ptr(id), &firstBundleID).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *secondBundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &secondBundleID).Return(nil).Once()
				svc.On("DeleteByReferenceObjectID", ctxWithTenant, model.BundleAPIReference, str.Ptr(id), &thirdBundleID).Return(nil).Once()
				return svc
			},
			InputID:                              "foo",
			Input:                                modelInput,
			SpecInput:                            &modelSpecInput,
			DefaultTargetURLPerBundleForUpdate:   defaultTargetURLPerBundleForUpdate,
			DefaultTargetURLPerBundleForCreation: defaultTargetURLPerBundleForCreation,
			BundleIDsForDeletion:                 bundleIDsForDeletion,
			DefaultBundleID:                      firstBundleID,
			ResourceType:                         resource.Application,
			ExpectedErr:                          nil,
		},
		{
			Name: "Success in ORD case for Application Template Version",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByIDGlobal", ctx, id).Return(apiDefinitionModelForAppTemplateVersion, nil).Once()
				repo.On("UpdateGlobal", ctx, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, resource.ApplicationTemplateVersion, model.APISpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctx, id, modelSpecInput, resource.ApplicationTemplateVersion, model.APISpecReference, id).Return(nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInputWithDefaultBundle, model.BundleAPIReference, str.Ptr(id), &firstBundleID).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *secondBundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &secondBundleID).Return(nil).Once()
				svc.On("DeleteByReferenceObjectID", ctx, model.BundleAPIReference, str.Ptr(id), &thirdBundleID).Return(nil).Once()
				return svc
			},
			InputID:                              "foo",
			Input:                                modelInput,
			SpecInput:                            &modelSpecInput,
			DefaultTargetURLPerBundleForUpdate:   defaultTargetURLPerBundleForUpdate,
			DefaultTargetURLPerBundleForCreation: defaultTargetURLPerBundleForCreation,
			BundleIDsForDeletion:                 bundleIDsForDeletion,
			DefaultBundleID:                      firstBundleID,
			ResourceType:                         resource.ApplicationTemplateVersion,
			Ctx:                                  ctx,
			ExpectedErr:                          nil,
		},
		{
			Name: "Success in ORD case when there is defaultBundle for BundleReference that has to be created",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctxWithTenant, tenantID, id).Return(apiDefinitionModelForApp, nil).Once()
				repo.On("Update", ctxWithTenant, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctxWithTenant, resource.Application, model.APISpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctxWithTenant, id, modelSpecInput, resource.Application, model.APISpecReference, id).Return(nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctxWithTenant, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &firstBundleID).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *secondBundleReferenceInputWithDefaultBundle, model.BundleAPIReference, str.Ptr(id), &secondBundleID).Return(nil).Once()
				svc.On("DeleteByReferenceObjectID", ctxWithTenant, model.BundleAPIReference, str.Ptr(id), &thirdBundleID).Return(nil).Once()
				return svc
			},
			InputID:                              "foo",
			Input:                                modelInput,
			SpecInput:                            &modelSpecInput,
			DefaultTargetURLPerBundleForUpdate:   defaultTargetURLPerBundleForUpdate,
			DefaultTargetURLPerBundleForCreation: defaultTargetURLPerBundleForCreation,
			BundleIDsForDeletion:                 bundleIDsForDeletion,
			DefaultBundleID:                      secondBundleID,
			ResourceType:                         resource.Application,
			ExpectedErr:                          nil,
		},
		{
			Name: "Error on BundleReference Update",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctxWithTenant, tenantID, id).Return(apiDefinitionModelForApp, nil).Once()
				repo.On("Update", ctxWithTenant, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: emptySpecService,
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctxWithTenant, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &firstBundleID).Return(testErr).Once()
				return svc
			},
			InputID:                              "foo",
			Input:                                modelInput,
			SpecInput:                            &modelSpecInput,
			DefaultTargetURLPerBundleForUpdate:   defaultTargetURLPerBundleForUpdate,
			DefaultTargetURLPerBundleForCreation: defaultTargetURLPerBundleForCreation,
			BundleIDsForDeletion:                 bundleIDsForDeletion,
			ResourceType:                         resource.Application,
			ExpectedErr:                          testErr,
		},
		{
			Name: "Error on BundleReference Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctxWithTenant, tenantID, id).Return(apiDefinitionModelForApp, nil).Once()
				repo.On("Update", ctxWithTenant, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: emptySpecService,
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctxWithTenant, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &firstBundleID).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *secondBundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &secondBundleID).Return(testErr).Once()
				return svc
			},
			InputID:                              "foo",
			Input:                                modelInput,
			SpecInput:                            &modelSpecInput,
			DefaultTargetURLPerBundleForUpdate:   defaultTargetURLPerBundleForUpdate,
			DefaultTargetURLPerBundleForCreation: defaultTargetURLPerBundleForCreation,
			BundleIDsForDeletion:                 bundleIDsForDeletion,
			ResourceType:                         resource.Application,
			ExpectedErr:                          testErr,
		},
		{
			Name: "Error on BundleReference Deletion",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctxWithTenant, tenantID, id).Return(apiDefinitionModelForApp, nil).Once()
				repo.On("Update", ctxWithTenant, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: emptySpecService,
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctxWithTenant, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &firstBundleID).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctxWithTenant, *secondBundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &secondBundleID).Return(nil).Once()
				svc.On("DeleteByReferenceObjectID", ctxWithTenant, model.BundleAPIReference, str.Ptr(id), &thirdBundleID).Return(testErr).Once()
				return svc
			},
			InputID:                              "foo",
			Input:                                modelInput,
			SpecInput:                            &modelSpecInput,
			DefaultTargetURLPerBundleForUpdate:   defaultTargetURLPerBundleForUpdate,
			DefaultTargetURLPerBundleForCreation: defaultTargetURLPerBundleForCreation,
			BundleIDsForDeletion:                 bundleIDsForDeletion,
			ResourceType:                         resource.Application,
			ExpectedErr:                          testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			specSvc := testCase.SpecServiceFn()
			bundleReferenceSvc := testCase.BundleReferenceFn()

			svc := api.NewService(repo, nil, specSvc, bundleReferenceSvc)
			svc.SetTimestampGen(func() time.Time { return fixedTimestamp })

			defaultCtx := ctxWithTenant
			if testCase.Ctx != nil {
				defaultCtx = testCase.Ctx
			}

			// WHEN
			err := svc.UpdateInManyBundles(defaultCtx, testCase.ResourceType, testCase.InputID, testCase.Input, testCase.SpecInput, testCase.DefaultTargetURLPerBundleForUpdate, testCase.DefaultTargetURLPerBundleForCreation, testCase.BundleIDsForDeletion, 0, testCase.DefaultBundleID)

			// then
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
			specSvc.AssertExpectations(t)
			bundleReferenceSvc.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.UpdateInManyBundles(context.TODO(), resource.Application, "", model.APIDefinitionInput{}, &model.SpecInput{}, nil, nil, nil, 0, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_UpdateForApplication(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"
	firstTargetURL := "https://test-url.com"
	timestamp := time.Now()
	frURL := "foo.bar"
	spec := "spec"

	modelInput := model.APIDefinitionInput{
		Name:         "Foo",
		TargetURLs:   api.ConvertTargetURLToJSONArray(firstTargetURL),
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
		ObjectType: model.APISpecReference,
		ObjectID:   id,
		Data:       &spec,
	}

	inputAPIDefinitionModel := mock.MatchedBy(func(api *model.APIDefinition) bool {
		return api.Name == modelInput.Name
	})

	apiDefinitionModel := &model.APIDefinition{
		Name:       "Bar",
		TargetURLs: api.ConvertTargetURLToJSONArray("https://test-url-updated.com"),
		Version:    &model.Version{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name              string
		RepositoryFn      func() *automock.APIRepository
		SpecServiceFn     func() *automock.SpecService
		BundleReferenceFn func() *automock.BundleReferenceService
		SpecInput         *model.SpecInput
		Input             model.APIDefinitionInput
		InputID           string
		ExpectedErr       error
	}{
		{
			Name: "Success with spec update",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, resource.Application, model.APISpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctx, id, modelSpecInput, resource.Application, model.APISpecReference, id).Return(nil).Once()
				return svc
			},
			SpecInput:   &modelSpecInput,
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Success with spec create",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, resource.Application, model.APISpecReference, id).Return(nil, nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, modelSpecInput, resource.Application, model.APISpecReference, id).Return(specID, nil).Once()
				return svc
			},
			SpecInput:   &modelSpecInput,
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Error when getting API",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(nil, testErr).Once()
				return repo
			},
			SpecServiceFn: emptySpecService,
			SpecInput:     &modelSpecInput,
			InputID:       "foo",
			Input:         modelInput,
			ExpectedErr:   testErr,
		},
		{
			Name: "Error when updating API",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, tenantID, inputAPIDefinitionModel).Return(testErr).Once()
				return repo
			},
			SpecServiceFn: emptySpecService,
			SpecInput:     &modelSpecInput,
			InputID:       "foo",
			Input:         modelInput,
			ExpectedErr:   testErr,
		},
		{
			Name: "Error when getting specs after API update",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, resource.Application, model.APISpecReference, id).Return(nil, testErr).Once()
				return svc
			},
			SpecInput:   &modelSpecInput,
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error when creating specs after API update",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, resource.Application, model.APISpecReference, id).Return(nil, nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, modelSpecInput, resource.Application, model.APISpecReference, id).Return("", testErr).Once()
				return svc
			},
			SpecInput:   &modelSpecInput,
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error when updating specs after API update",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, tenantID, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, resource.Application, model.APISpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctx, id, modelSpecInput, resource.Application, model.APISpecReference, id).Return(testErr).Once()
				return svc
			},
			SpecInput:   &modelSpecInput,
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			specSvc := testCase.SpecServiceFn()
			defer mock.AssertExpectationsForObjects(t, repo, specSvc)

			svc := api.NewService(repo, nil, specSvc, nil)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// WHEN
			err := svc.UpdateForApplication(ctx, testCase.InputID, testCase.Input, testCase.SpecInput)

			// then
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.UpdateForApplication(context.TODO(), "", model.APIDefinitionInput{}, &model.SpecInput{})
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
		RepositoryFn func() *automock.APIRepository
		InputID      string
		ResourceType resource.Type
		ExpectedErr  error
	}{
		{
			Name: "Success for Application",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Delete", ctx, tenantID, id).Return(nil).Once()
				return repo
			},
			InputID:      id,
			ResourceType: resource.Application,
			ExpectedErr:  nil,
		},
		{
			Name: "Success for Application Template Version",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("DeleteGlobal", ctx, id).Return(nil).Once()
				return repo
			},
			InputID:      id,
			ResourceType: resource.ApplicationTemplateVersion,
			ExpectedErr:  nil,
		},
		{
			Name: "Delete Error for Application",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Delete", ctx, tenantID, id).Return(testErr).Once()
				return repo
			},
			InputID:      id,
			ResourceType: resource.Application,
			ExpectedErr:  testErr,
		},
		{
			Name: "Delete Error for Application Tempalate Version",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
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

			svc := api.NewService(repo, nil, nil, nil)

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
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.Delete(context.TODO(), resource.Application, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_DeleteAllByBundleID(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	id := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.APIRepository
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("DeleteAllByBundleID", ctx, tenantID, id).Return(nil).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("DeleteAllByBundleID", ctx, tenantID, id).Return(testErr).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := api.NewService(repo, nil, nil, nil)

			// WHEN
			err := svc.DeleteAllByBundleID(ctx, testCase.InputID)

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
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.DeleteAllByBundleID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListFetchRequests(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testErr := errors.New("Test error")

	frURL := "foo.bar"
	firstFRID := "frID"
	secondFRID := "frID2"
	firstSpecID := "specID"
	secondSpecID := "specID2"
	specIDs := []string{firstSpecID, secondSpecID}
	fixedTimestamp := time.Now()

	firstFetchRequest := fixModelFetchRequest(firstFRID, frURL, fixedTimestamp)
	secondFetchRequest := fixModelFetchRequest(secondFRID, frURL, fixedTimestamp)
	fetchRequests := []*model.FetchRequest{firstFetchRequest, secondFetchRequest}

	testCases := []struct {
		Name                  string
		SpecServiceFn         func() *automock.SpecService
		ExpectedFetchRequests []*model.FetchRequest
		ExpectedErrMessage    string
	}{
		{
			Name: "Success",
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("ListFetchRequestsByReferenceObjectIDs", ctx, tenantID, specIDs, model.APISpecReference).Return(fetchRequests, nil).Once()
				return svc
			},
			ExpectedFetchRequests: fetchRequests,
		},
		{
			Name: "Success - Fetch Request Not Found",
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("ListFetchRequestsByReferenceObjectIDs", ctx, tenantID, specIDs, model.APISpecReference).Return(nil, apperrors.NewNotFoundError(resource.FetchRequest, "")).Once()
				return svc
			},
			ExpectedFetchRequests: nil,
		},
		{
			Name: "Error while listing Fetch Requests",
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("ListFetchRequestsByReferenceObjectIDs", ctx, tenantID, specIDs, model.APISpecReference).Return(nil, testErr).Once()
				return svc
			},
			ExpectedFetchRequests: nil,
			ExpectedErrMessage:    testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			specService := testCase.SpecServiceFn()

			svc := api.NewService(nil, nil, specService, nil)

			// WHEN
			frs, err := svc.ListFetchRequests(ctx, specIDs)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, frs, testCase.ExpectedFetchRequests)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			specService.AssertExpectations(t)
		})
	}
	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListFetchRequests(context.TODO(), nil)
		assert.True(t, apperrors.IsCannotReadTenant(err))
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

func emptySpecConverter() *automock.SpecConverter {
	conv := &automock.SpecConverter{}
	return conv
}
