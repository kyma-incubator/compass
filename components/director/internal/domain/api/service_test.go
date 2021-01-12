package api_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	packageID := "foobar"
	name := "foo"
	desc := "bar"

	apiDefinition := fixAPIDefinitionModel(id, packageID, name, desc)

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

func TestService_GetForPackage(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	pkgID := "foobar"
	name := "foo"
	desc := "bar"

	apiDefinition := fixAPIDefinitionModel(id, pkgID, name, desc)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.APIRepository
		Input              model.APIDefinitionInput
		InputID            string
		PackageID          string
		ExpectedAPI        *model.APIDefinition
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetForPackage", ctx, tenantID, id, pkgID).Return(apiDefinition, nil).Once()
				return repo
			},
			InputID:            id,
			PackageID:          pkgID,
			ExpectedAPI:        apiDefinition,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when APIDefinition retrieval failed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetForPackage", ctx, tenantID, id, pkgID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			PackageID:          pkgID,
			ExpectedAPI:        nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := api.NewService(repo, nil, nil, nil)

			// when
			api, err := svc.GetForPackage(ctx, testCase.InputID, testCase.PackageID)

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
		_, err := svc.GetForPackage(context.TODO(), "", "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListForPackage(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	pkgID := "foobar"
	name := "foo"
	desc := "bar"

	apiDefinitions := []*model.APIDefinition{
		fixAPIDefinitionModel(id, pkgID, name, desc),
		fixAPIDefinitionModel(id, pkgID, name, desc),
		fixAPIDefinitionModel(id, pkgID, name, desc),
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
				repo.On("ListForPackage", ctx, tenantID, packageID, 2, after).Return(apiDefinitionPage, nil).Once()
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
				repo.On("ListForPackage", ctx, tenantID, packageID, 2, after).Return(nil, testErr).Once()
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
			docs, err := svc.ListForPackage(ctx, packageID, testCase.PageSize, after)

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
		_, err := svc.ListForPackage(context.TODO(), "", 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_CreateToPackage(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	packageID := "pkgid"
	name := "Foo"
	targetUrl := "https://test-url.com"

	timestamp := time.Now()
	frID := "fr-id"
	frURL := "foo.bar"
	spec := "test"

	modelFr := fixModelFetchRequest(frID, frURL, timestamp)

	modelInput := model.APIDefinitionInput{
		Name:      name,
		TargetURL: targetUrl,
		Spec: &model.APISpecInput{
			FetchRequest: &model.FetchRequestInput{
				URL: frURL,
			},
		},
		Version: &model.VersionInput{},
	}

	modelAPIDefinition := &model.APIDefinition{
		ID:        id,
		PackageID: packageID,
		Tenant:    tenantID,
		Name:      name,
		TargetURL: targetUrl,
		Spec:      &model.APISpec{},
		Version:   &model.Version{},
	}

	modelAPIDefinitionWithSpec := &model.APIDefinition{
		ID:        id,
		PackageID: packageID,
		Tenant:    tenantID,
		Name:      name,
		TargetURL: targetUrl,
		Spec:      &model.APISpec{Data: &spec},
		Version:   &model.Version{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name                  string
		RepositoryFn          func() *automock.APIRepository
		FetchRequestRepoFn    func() *automock.FetchRequestRepository
		UIDServiceFn          func() *automock.UIDService
		FetchRequestServiceFn func() *automock.FetchRequestService
		Input                 model.APIDefinitionInput
		ExpectedErr           error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, modelAPIDefinition).Return(nil).Once()
				repo.On("Update", ctx, modelAPIDefinition).Return(nil).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, modelFr).Return(nil).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, fixModelFetchRequest(frID, frURL, timestamp)).Return(nil)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Success fetched API Spec",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, modelAPIDefinition).Return(nil).Once()
				repo.On("Update", ctx, modelAPIDefinitionWithSpec).Return(nil).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, modelFr).Return(nil).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, fixModelFetchRequest(frID, frURL, timestamp)).Return(&spec)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Error - API Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, modelAPIDefinition).Return(testErr).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - Fetch Request Creation",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, modelAPIDefinition).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, modelFr).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - API Update",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, modelAPIDefinition).Return(nil).Once()
				repo.On("Update", ctx, modelAPIDefinition).Return(testErr).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, modelFr).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, modelFr).Return(nil)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Success when fetching API Spec failed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, modelAPIDefinition).Return(nil).Once()
				repo.On("Update", ctx, modelAPIDefinition).Return(nil).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, modelFr).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, fixModelFetchRequest(frID, frURL, timestamp)).Return(nil)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			uidService := testCase.UIDServiceFn()
			fetchRequestService := testCase.FetchRequestServiceFn()

			svc := api.NewService(repo, fetchRequestRepo, uidService, fetchRequestService)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			result, err := svc.CreateInPackage(ctx, packageID, testCase.Input)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.IsType(t, "string", result)
			}

			repo.AssertExpectations(t)
			fetchRequestRepo.AssertExpectations(t)
			uidService.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.CreateInPackage(context.TODO(), "", model.APIDefinitionInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	timestamp := time.Now()
	frID := "fr-id"
	frURL := "foo.bar"

	modelInput := model.APIDefinitionInput{
		Name:      "Foo",
		TargetURL: "https://test-url.com",
		Spec: &model.APISpecInput{
			FetchRequest: &model.FetchRequestInput{
				URL: frURL,
			},
		},
		Version: &model.VersionInput{},
	}

	inputAPIDefinitionModel := mock.MatchedBy(func(api *model.APIDefinition) bool {
		return api.Name == modelInput.Name
	})

	apiDefinitionModel := &model.APIDefinition{
		Name:      "Bar",
		TargetURL: "https://test-url-updated.com",
		Spec:      &model.APISpec{},
		Version:   &model.Version{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	modelFr := fixModelFetchRequest(frID, frURL, timestamp)

	testCases := []struct {
		Name                  string
		RepositoryFn          func() *automock.APIRepository
		FetchRequestRepoFn    func() *automock.FetchRequestRepository
		UIDServiceFn          func() *automock.UIDService
		FetchRequestServiceFn func() *automock.FetchRequestService
		Input                 model.APIDefinitionInput
		InputID               string
		ExpectedErr           error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenantID, model.APIFetchRequestReference, id).Return(nil).Once()
				repo.On("Create", ctx, modelFr).Return(nil).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, modelFr).Return(nil)
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
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
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenantID, model.APIFetchRequestReference, id).Return(nil).Once()
				repo.On("Create", ctx, modelFr).Return(nil).Once()

				return repo
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, modelFr).Return(nil)
				return svc
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Delete FetchRequest by reference Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(apiDefinitionModel, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenantID, model.APIFetchRequestReference, id).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Fetch Request Creation Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(apiDefinitionModel, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenantID, model.APIFetchRequestReference, id).Return(nil).Once()
				repo.On("Create", ctx, modelFr).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(nil, testErr).Once()
				return repo
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Success when fetch request failed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(apiDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenantID, model.APIFetchRequestReference, id).Return(nil).Once()
				repo.On("Create", ctx, modelFr).Return(nil).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, modelFr).Return(nil)
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			uidSvc := testCase.UIDServiceFn()
			fetchRequestSvc := testCase.FetchRequestServiceFn()

			svc := api.NewService(repo, fetchRequestRepo, uidSvc, fetchRequestSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

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
			fetchRequestRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.Update(context.TODO(), "", model.APIDefinitionInput{})
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
		Input        model.APIDefinitionInput
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

func TestService_RefetchAPISpec(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	apiID := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	dataBytes := "data"
	modelAPISpec := &model.APISpec{
		Data: &dataBytes,
	}

	modelAPIDefinition := &model.APIDefinition{
		Spec: modelAPISpec,
	}

	timestamp := time.Now()
	fr := &model.FetchRequest{
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
	}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.APIRepository
		FetchRequestRepoFn func() *automock.FetchRequestRepository
		FetchRequestSvcFn  func() *automock.FetchRequestService
		ExpectedAPISpec    *model.APISpec
		ExpectedErr        error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, apiID).Return(modelAPIDefinition, nil).Once()
				repo.On("Update", ctx, modelAPIDefinition).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenantID, model.APIFetchRequestReference, apiID).Return(nil, nil)
				return repo
			},
			FetchRequestSvcFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			ExpectedAPISpec: modelAPISpec,
			ExpectedErr:     nil,
		},
		{
			Name: "Success - fetched API Spec",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, apiID).Return(modelAPIDefinition, nil).Once()
				repo.On("Update", ctx, modelAPIDefinition).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenantID, model.APIFetchRequestReference, apiID).Return(fr, nil)
				return repo
			},
			FetchRequestSvcFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, fr).Return(&dataBytes)
				return svc
			},
			ExpectedAPISpec: modelAPISpec,
			ExpectedErr:     nil,
		},
		{
			Name: "Get from repository error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, apiID).Return(nil, testErr).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			FetchRequestSvcFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Get fetch request error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, apiID).Return(modelAPIDefinition, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenantID, model.APIFetchRequestReference, apiID).Return(nil, testErr)
				return repo
			},
			FetchRequestSvcFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     errors.Wrapf(testErr, "while getting FetchRequest by API Definition ID %s", apiID),
		},
		{
			Name: "Error when updating API Definition failed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", ctx, tenantID, apiID).Return(modelAPIDefinition, nil).Once()
				repo.On("Update", ctx, modelAPIDefinition).Return(testErr).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenantID, model.APIFetchRequestReference, apiID).Return(nil, nil)
				return repo
			},
			FetchRequestSvcFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     errors.Wrap(testErr, "while updating api with api spec"),
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			frRepo := testCase.FetchRequestRepoFn()
			frSvc := testCase.FetchRequestSvcFn()

			svc := api.NewService(repo, frRepo, nil, frSvc)

			// when
			result, err := svc.RefetchAPISpec(ctx, apiID)

			// then
			assert.Equal(t, testCase.ExpectedAPISpec, result)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := api.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.RefetchAPISpec(context.TODO(), "")
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

	refID := "doc-id"
	frURL := "foo.bar"
	timestamp := time.Now()

	fetchRequestModel := fixModelFetchRequest("foo", frURL, timestamp)

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.APIRepository
		FetchRequestRepoFn   func() *automock.FetchRequestRepository
		InputAPIDefID        string
		ExpectedFetchRequest *model.FetchRequest
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Exists", ctx, tenantID, refID).Return(true, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenantID, model.APIFetchRequestReference, refID).Return(fetchRequestModel, nil).Once()
				return repo
			},
			InputAPIDefID:        refID,
			ExpectedFetchRequest: fetchRequestModel,
			ExpectedErrMessage:   "",
		},
		{
			Name: "Error - API Definition Not Exist",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Exists", ctx, tenantID, refID).Return(false, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			InputAPIDefID:        refID,
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   fmt.Sprintf("API Definition with ID %s doesn't exist", refID),
		},
		{
			Name: "Success - Not Found",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Exists", ctx, tenantID, refID).Return(true, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenantID, model.APIFetchRequestReference, refID).Return(nil, apperrors.NewNotFoundError(resource.API, "")).Once()
				return repo
			},
			InputAPIDefID:        refID,
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   "",
		},
		{
			Name: "Error - Get FetchRequest",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Exists", ctx, tenantID, refID).Return(true, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenantID, model.APIFetchRequestReference, refID).Return(nil, testErr).Once()
				return repo
			},
			InputAPIDefID:        refID,
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name: "Error - API doesn't exist",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Exists", ctx, tenantID, refID).Return(false, testErr).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			InputAPIDefID:      refID,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			svc := api.NewService(repo, fetchRequestRepo, nil, nil)

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
			fetchRequestRepo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := api.NewService(nil, nil, nil, nil)
		// when
		_, err := svc.GetFetchRequest(context.TODO(), "dd")
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}
