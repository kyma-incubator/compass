package api_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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
	// given
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

			// when
			document, err := svc.Get(ctx, testCase.InputID)

			// then
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
	// given
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

			// when
			api, err := svc.GetForBundle(ctx, testCase.InputID, testCase.BundleID)

			// then
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

func TestService_ListForBundle(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	name := "foo"
	desc := "bar"

	apiDefinitions := []*model.APIDefinition{
		fixAPIDefinitionModel(id, name, desc),
		fixAPIDefinitionModel(id, name, desc),
		fixAPIDefinitionModel(id, name, desc),
	}
	apiDefinitionPage := &model.APIDefinitionPage{
		Data:       apiDefinitions,
		TotalCount: len(apiDefinitions),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	after := "test"

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
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("ListForBundle", ctx, tenantID, bundleID, 2, after).Return(apiDefinitionPage, nil).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     apiDefinitionPage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Return error when page size is less than 1",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			PageSize:           0,
			ExpectedResult:     apiDefinitionPage,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Return error when page size is bigger than 200",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			PageSize:           201,
			ExpectedResult:     apiDefinitionPage,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when APIDefinition listing failed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("ListForBundle", ctx, tenantID, bundleID, 2, after).Return(nil, testErr).Once()
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

			svc := api.NewService(repo, nil, nil, nil)

			// when
			docs, err := svc.ListForBundle(ctx, bundleID, testCase.PageSize, after)

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
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListForBundle(context.TODO(), "", 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByApplicationID(t *testing.T) {
	// given
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
				repo.On("ListByApplicationID", ctx, tenantID, appID).Return(apiDefinitions, nil).Once()
				return repo
			},
			ExpectedResult:     apiDefinitions,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when APIDefinition listing failed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("ListByApplicationID", ctx, tenantID, appID).Return(nil, testErr).Once()
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
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListByApplicationID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	bundleID := "bndlid"
	bundleID2 := "bndlid2"
	packageID := packageID
	name := "Foo"
	targetUrl := "https://test-url.com"
	targetUrl2 := "https://test2-url.com"

	timestamp := time.Now()
	frURL := "foo.bar"
	spec := "test"
	spec2 := "test2"

	modelInput := model.APIDefinitionInput{
		Name:         name,
		TargetURLs:   api.ConvertTargetUrlToJsonArray(targetUrl),
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

	modelAPIDefinition := &model.APIDefinition{
		PackageID:     &packageID,
		ApplicationID: appID,
		Tenant:        tenantID,
		Name:          name,
		TargetURLs:    api.ConvertTargetUrlToJsonArray(targetUrl),
		Version:       &model.Version{},
		BaseEntity: &model.BaseEntity{
			ID:    id,
			Ready: true,
		},
	}

	bundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: str.Ptr(api.ExtractTargetUrlFromJsonArray(modelAPIDefinition.TargetURLs)),
	}
	secondBundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: &targetUrl2,
	}

	singleDefaultTargetURLPerBundle := map[string]string{bundleID: targetUrl}
	defaultTargetURLPerBundle := map[string]string{bundleID: targetUrl, bundleID2: targetUrl2}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name                      string
		RepositoryFn              func() *automock.APIRepository
		UIDServiceFn              func() *automock.UIDService
		SpecServiceFn             func() *automock.SpecService
		BundleReferenceFn         func() *automock.BundleReferenceService
		Input                     model.APIDefinitionInput
		SpecsInput                []*model.SpecInput
		DefaultTargetURLPerBundle map[string]string
		ExpectedErr               error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, modelAPIDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], model.APISpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[1], model.APISpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), str.Ptr(bundleID)).Return(nil).Once()
				return svc
			},
			Input:      modelInput,
			SpecsInput: modelSpecsInput,
		},
		{
			Name: "Success in ORD scenario where defaultTargetURLPerBundle map is passed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, modelAPIDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], model.APISpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[1], model.APISpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), str.Ptr(bundleID)).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *secondBundleReferenceInput, model.BundleAPIReference, str.Ptr(id), str.Ptr(bundleID2)).Return(nil).Once()
				return svc
			},
			Input:                     modelInput,
			SpecsInput:                modelSpecsInput,
			DefaultTargetURLPerBundle: defaultTargetURLPerBundle,
		},
		{
			Name: "Error - API Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, modelAPIDefinition).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			Input:       modelInput,
			SpecsInput:  modelSpecsInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - Spec Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, modelAPIDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], model.APISpecReference, id).Return("", testErr).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			Input:       modelInput,
			SpecsInput:  modelSpecsInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - BundleReference API Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, modelAPIDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], model.APISpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[1], model.APISpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), str.Ptr(bundleID)).Return(testErr).Once()
				return svc
			},
			Input:       modelInput,
			SpecsInput:  modelSpecsInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error in ORD scenario - BundleReference API Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, modelAPIDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[0], model.APISpecReference, id).Return("id", nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *modelSpecsInput[1], model.APISpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("CreateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), str.Ptr(bundleID)).Return(testErr).Once()
				return svc
			},
			Input:                     modelInput,
			SpecsInput:                modelSpecsInput,
			DefaultTargetURLPerBundle: singleDefaultTargetURLPerBundle,
			ExpectedErr:               testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			uidService := testCase.UIDServiceFn()
			specService := testCase.SpecServiceFn()
			bundleReferenceService := testCase.BundleReferenceFn()

			svc := api.NewService(repo, uidService, specService, bundleReferenceService)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			result, err := svc.Create(ctx, appID, &bundleID, &packageID, testCase.Input, testCase.SpecsInput, testCase.DefaultTargetURLPerBundle)

			// then
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
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.Create(context.TODO(), "", nil, nil, model.APIDefinitionInput{}, []*model.SpecInput{}, nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	var bundleID *string
	timestamp := time.Now()
	frURL := "foo.bar"
	spec := "spec"

	modelInput := model.APIDefinitionInput{
		Name:         "Foo",
		TargetURLs:   api.ConvertTargetUrlToJsonArray("https://test-url.com"),
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
		Tenant:     tenantID,
		ObjectType: model.APISpecReference,
		ObjectID:   id,
		Data:       &spec,
	}

	inputAPIDefinitionModel := mock.MatchedBy(func(api *model.APIDefinition) bool {
		return api.Name == modelInput.Name
	})

	apiDefinitionModel := &model.APIDefinition{
		Name:       "Bar",
		TargetURLs: api.ConvertTargetUrlToJsonArray("https://test-url-updated.com"),
		Version:    &model.Version{},
	}

	bundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: str.Ptr(api.ExtractTargetUrlFromJsonArray(modelInput.TargetURLs)),
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
		ExpectedErr       error
	}{
		{
			Name: "Success When Spec is Found should update it",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.APISpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctx, id, modelSpecInput, model.APISpecReference, id).Return(nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), bundleID).Return(nil).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			SpecInput:   &modelSpecInput,
			ExpectedErr: nil,
		},
		{
			Name: "Success When Spec is not found should create in",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.APISpecReference, id).Return(nil, nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, modelSpecInput, model.APISpecReference, id).Return("id", nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), bundleID).Return(nil).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			SpecInput:   &modelSpecInput,
			ExpectedErr: nil,
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputAPIDefinitionModel).Return(testErr).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				return &automock.BundleReferenceService{}
			},
			InputID:     "foo",
			Input:       modelInput,
			SpecInput:   &modelSpecInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Spec Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.APISpecReference, id).Return(nil, testErr).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), bundleID).Return(nil).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			SpecInput:   &modelSpecInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Spec Creation Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.APISpecReference, id).Return(nil, nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, modelSpecInput, model.APISpecReference, id).Return("", testErr).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), bundleID).Return(nil).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			SpecInput:   &modelSpecInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Spec Update Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.APISpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctx, id, modelSpecInput, model.APISpecReference, id).Return(testErr).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), bundleID).Return(nil).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			SpecInput:   &modelSpecInput,
			ExpectedErr: testErr,
		},
		{
			Name: "BundleReference API Update Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), bundleID).Return(testErr).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			SpecInput:   &modelSpecInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			specSvc := testCase.SpecServiceFn()
			bundleReferenceSvc := testCase.BundleReferenceFn()

			svc := api.NewService(repo, nil, specSvc, bundleReferenceSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			err := svc.Update(ctx, testCase.InputID, testCase.Input, testCase.SpecInput)

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
		err := svc.Update(context.TODO(), "", model.APIDefinitionInput{}, &model.SpecInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_UpdateInManyBundles(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	firstBndlID := "id1"
	secondBndlID := "id2"
	thirdBndlID := "id3"
	firstTargetURL := "https://test-url.com"
	secondTargetURL := "https://test2-url.com"
	timestamp := time.Now()
	frURL := "foo.bar"
	spec := "spec"

	modelInput := model.APIDefinitionInput{
		Name:         "Foo",
		TargetURLs:   api.ConvertTargetUrlToJsonArray(firstTargetURL),
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
		Tenant:     tenantID,
		ObjectType: model.APISpecReference,
		ObjectID:   id,
		Data:       &spec,
	}

	inputAPIDefinitionModel := mock.MatchedBy(func(api *model.APIDefinition) bool {
		return api.Name == modelInput.Name
	})

	apiDefinitionModel := &model.APIDefinition{
		Name:       "Bar",
		TargetURLs: api.ConvertTargetUrlToJsonArray("https://test-url-updated.com"),
		Version:    &model.Version{},
	}

	bundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: str.Ptr(api.ExtractTargetUrlFromJsonArray(modelInput.TargetURLs)),
	}
	secondBundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: str.Ptr(secondTargetURL),
	}

	defaultTargetURLPerBundleForUpdate := map[string]string{firstBndlID: firstTargetURL}
	defaultTargetURLPerBundleForCreation := map[string]string{secondBndlID: secondTargetURL}
	bundleIDsForDeletion := []string{thirdBndlID}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

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
		ExpectedErr                          error
	}{
		{
			Name: "Success in ORD case",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.APISpecReference, id).Return(modelSpec, nil).Once()
				svc.On("UpdateByReferenceObjectID", ctx, id, modelSpecInput, model.APISpecReference, id).Return(nil).Once()
				return svc
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &firstBndlID).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *secondBundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &secondBndlID).Return(nil).Once()
				svc.On("DeleteByReferenceObjectID", ctx, model.BundleAPIReference, str.Ptr(id), &thirdBndlID).Return(nil).Once()
				return svc
			},
			InputID:                              "foo",
			Input:                                modelInput,
			SpecInput:                            &modelSpecInput,
			DefaultTargetURLPerBundleForUpdate:   defaultTargetURLPerBundleForUpdate,
			DefaultTargetURLPerBundleForCreation: defaultTargetURLPerBundleForCreation,
			BundleIDsForDeletion:                 bundleIDsForDeletion,
			ExpectedErr:                          nil,
		},
		{
			Name: "Error on BundleReference Update",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &firstBndlID).Return(testErr).Once()
				return svc
			},
			InputID:                              "foo",
			Input:                                modelInput,
			SpecInput:                            &modelSpecInput,
			DefaultTargetURLPerBundleForUpdate:   defaultTargetURLPerBundleForUpdate,
			DefaultTargetURLPerBundleForCreation: defaultTargetURLPerBundleForCreation,
			BundleIDsForDeletion:                 bundleIDsForDeletion,
			ExpectedErr:                          testErr,
		},
		{
			Name: "Error on BundleReference Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &firstBndlID).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *secondBundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &secondBndlID).Return(testErr).Once()
				return svc
			},
			InputID:                              "foo",
			Input:                                modelInput,
			SpecInput:                            &modelSpecInput,
			DefaultTargetURLPerBundleForUpdate:   defaultTargetURLPerBundleForUpdate,
			DefaultTargetURLPerBundleForCreation: defaultTargetURLPerBundleForCreation,
			BundleIDsForDeletion:                 bundleIDsForDeletion,
			ExpectedErr:                          testErr,
		},
		{
			Name: "Error on BundleReference Deletion",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			BundleReferenceFn: func() *automock.BundleReferenceService {
				svc := &automock.BundleReferenceService{}
				svc.On("UpdateByReferenceObjectID", ctx, *bundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &firstBndlID).Return(nil).Once()
				svc.On("CreateByReferenceObjectID", ctx, *secondBundleReferenceInput, model.BundleAPIReference, str.Ptr(id), &secondBndlID).Return(nil).Once()
				svc.On("DeleteByReferenceObjectID", ctx, model.BundleAPIReference, str.Ptr(id), &thirdBndlID).Return(testErr).Once()
				return svc
			},
			InputID:                              "foo",
			Input:                                modelInput,
			SpecInput:                            &modelSpecInput,
			DefaultTargetURLPerBundleForUpdate:   defaultTargetURLPerBundleForUpdate,
			DefaultTargetURLPerBundleForCreation: defaultTargetURLPerBundleForCreation,
			BundleIDsForDeletion:                 bundleIDsForDeletion,
			ExpectedErr:                          testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			specSvc := testCase.SpecServiceFn()
			bundleReferenceSvc := testCase.BundleReferenceFn()

			svc := api.NewService(repo, nil, specSvc, bundleReferenceSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			err := svc.UpdateInManyBundles(ctx, testCase.InputID, testCase.Input, testCase.SpecInput, testCase.DefaultTargetURLPerBundleForUpdate, testCase.DefaultTargetURLPerBundleForCreation, testCase.BundleIDsForDeletion)

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
		err := svc.UpdateInManyBundles(context.TODO(), "", model.APIDefinitionInput{}, &model.SpecInput{}, nil, nil, nil)
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
				repo.On("Delete", ctx, tenantID, id).Return(nil).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
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

			svc := api.NewService(repo, nil, nil, nil)

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
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.Delete(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_DeleteAllByBundleID(t *testing.T) {
	// given
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
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := api.NewService(repo, nil, nil, nil)

			// when
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
		err := svc.Delete(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetFetchRequest(t *testing.T) {
	// given
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testErr := errors.New("Test error")

	apiID := "api-id"
	refID := "doc-id"
	frURL := "foo.bar"

	spec := "spec"

	timestamp := time.Now()

	modelSpec := &model.Spec{
		ID:         refID,
		Tenant:     tenantID,
		ObjectType: model.APISpecReference,
		ObjectID:   apiID,
		Data:       &spec,
	}

	fetchRequestModel := fixModelFetchRequest("foo", frURL, timestamp)

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.APIRepository
		SpecServiceFn        func() *automock.SpecService
		InputAPIDefID        string
		ExpectedFetchRequest *model.FetchRequest
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Exists", ctx, tenantID, apiID).Return(true, nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.APISpecReference, apiID).Return(modelSpec, nil).Once()
				svc.On("GetFetchRequest", ctx, refID).Return(fetchRequestModel, nil).Once()
				return svc
			},
			InputAPIDefID:        apiID,
			ExpectedFetchRequest: fetchRequestModel,
			ExpectedErrMessage:   "",
		},
		{
			Name: "Error - API Definition Not Exist",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Exists", ctx, tenantID, apiID).Return(false, nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			InputAPIDefID:        apiID,
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   fmt.Sprintf("API Definition with id %s doesn't exist", apiID),
		},
		{
			Name: "Success - Spec Not Found",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Exists", ctx, tenantID, apiID).Return(true, nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.APISpecReference, apiID).Return(nil, nil).Once()
				return svc
			},
			InputAPIDefID:        apiID,
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   "",
		},
		{
			Name: "Success - Fetch Request Not Found",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Exists", ctx, tenantID, apiID).Return(true, nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.APISpecReference, apiID).Return(modelSpec, nil).Once()
				svc.On("GetFetchRequest", ctx, refID).Return(nil, apperrors.NewNotFoundError(resource.FetchRequest, "")).Once()
				return svc
			},
			InputAPIDefID:        apiID,
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   "",
		},
		{
			Name: "Error - Get Spec",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Exists", ctx, tenantID, apiID).Return(true, nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.APISpecReference, apiID).Return(nil, testErr).Once()
				return svc
			},
			InputAPIDefID:        apiID,
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name: "Error - Get FetchRequest",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Exists", ctx, tenantID, apiID).Return(true, nil).Once()
				return repo
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", ctx, model.APISpecReference, apiID).Return(modelSpec, nil).Once()
				svc.On("GetFetchRequest", ctx, refID).Return(nil, testErr).Once()
				return svc
			},
			InputAPIDefID:        apiID,
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			specService := testCase.SpecServiceFn()

			svc := api.NewService(repo, nil, specService, nil)

			// when
			l, err := svc.GetFetchRequest(ctx, testCase.InputAPIDefID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, l, testCase.ExpectedFetchRequest)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := api.NewService(nil, nil, nil, nil)
		// when
		_, err := svc.GetFetchRequest(context.TODO(), "dd")
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}
