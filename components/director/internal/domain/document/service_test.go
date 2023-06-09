package document_test

import (
	"context"
	"errors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Get(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"

	documentModel := fixModelDocumentForApp("1", id)
	tnt := givenTenant()
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.DocumentRepository
		Input              model.DocumentInput
		InputID            string
		ExpectedDocument   *model.Document
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(documentModel, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedDocument:   documentModel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when document retrieval failed",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedDocument:   documentModel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := document.NewService(repo, nil, nil)

			// WHEN
			doc, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedDocument, doc)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_GetForBundle(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	id := "foo"
	bndlID := bndlID()
	tenantID := "bar"
	externalTenantID := "external-tenant"

	bundleID := "test"
	doc := fixModelDocumentForApp(id, bndlID)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.DocumentRepository
		Input              model.DocumentInput
		InputID            string
		BundleID           string
		ExpectedDocument   *model.Document
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("GetForBundle", ctx, tenantID, id, bundleID).Return(doc, nil).Once()
				return repo
			},
			InputID:            id,
			BundleID:           bundleID,
			ExpectedDocument:   doc,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Event Definition retrieval failed",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("GetForBundle", ctx, tenantID, id, bundleID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			BundleID:           bundleID,
			ExpectedDocument:   doc,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := document.NewService(repo, nil, nil)

			// WHEN
			eventAPIDefinition, err := svc.GetForBundle(ctx, testCase.InputID, testCase.BundleID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedDocument, eventAPIDefinition)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := document.NewService(nil, nil, nil)
		// WHEN
		_, err := svc.GetForBundle(context.TODO(), "", "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_CreateToBundle(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	tnt := "tenant"
	externalTnt := "external-tenant"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	id := "foo"
	bundleID := "foo"
	frURL := "foo.bar"
	frID := "fr-id"
	timestamp := time.Now()
	modelInput := fixModelDocumentInputWithFetchRequest(frURL)
	modelAppDoc := modelInput.ToDocumentWithinBundle(id, bundleID, resource.Application, appID)
	modelAppTemplateVersionDoc := modelInput.ToDocumentWithinBundle(id, bundleID, resource.ApplicationTemplateVersion, appTemplateVersionID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.DocumentRepository
		FetchRequestRepoFn func() *automock.FetchRequestRepository
		UIDServiceFn       func() *automock.UIDService
		Input              model.DocumentInput
		ResourceType       resource.Type
		ResourceID         string
		ExpectedErr        error
	}{
		{
			Name: "Success for Application",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Create", ctx, tnt, modelAppDoc).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, tnt, fixModelFetchRequest(frID, frURL, timestamp)).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			Input:        *modelInput,
			ResourceType: resource.Application,
			ResourceID:   appID,
			ExpectedErr:  nil,
		},
		{
			Name: "Success for Application Template Version",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("CreateGlobal", ctx, modelAppTemplateVersionDoc).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("CreateGlobal", ctx, fixModelFetchRequest(frID, frURL, timestamp)).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			Input:        *modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ResourceID:   appTemplateVersionID,
			ExpectedErr:  nil,
		},
		{
			Name: "Returns error when document creation failed for Application",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Create", ctx, tnt, modelAppDoc).Return(testErr).Once()
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
			Input:        *modelInput,
			ResourceType: resource.Application,
			ResourceID:   appID,
			ExpectedErr:  testErr,
		},
		{
			Name: "Returns error when document creation failed for Application Template Version",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("CreateGlobal", ctx, modelAppTemplateVersionDoc).Return(testErr).Once()
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
			Input:        *modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ResourceID:   appTemplateVersionID,
			ExpectedErr:  testErr,
		},
		{
			Name: "Error - Fetch Request Creation for Application",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Create", ctx, tnt, modelAppDoc).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, tnt, fixModelFetchRequest(frID, frURL, timestamp)).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			Input:        *modelInput,
			ResourceType: resource.Application,
			ResourceID:   appID,
			ExpectedErr:  testErr,
		},
		{
			Name: "Error - Fetch Request Creation for Application Template Version",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("CreateGlobal", ctx, modelAppTemplateVersionDoc).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("CreateGlobal", ctx, fixModelFetchRequest(frID, frURL, timestamp)).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			Input:        *modelInput,
			ResourceType: resource.ApplicationTemplateVersion,
			ResourceID:   appTemplateVersionID,
			ExpectedErr:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			idSvc := testCase.UIDServiceFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			svc := document.NewService(repo, fetchRequestRepo, idSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// WHEN
			result, err := svc.CreateInBundle(ctx, testCase.ResourceType, testCase.ResourceID, bundleID, testCase.Input)

			// then
			assert.IsType(t, "string", result)
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
			fetchRequestRepo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := document.NewService(nil, nil, fixUIDSvc())
		// WHEN
		_, err := svc.CreateInBundle(context.TODO(), resource.Application, "appID", "bndlID", model.DocumentInput{})
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}
func TestService_Delete(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "bar"

	tnt := givenTenant()
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.DocumentRepository
		Input              model.DocumentInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Delete", ctx, tnt, id).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when document deletion failed",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Delete", ctx, tnt, id).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := document.NewService(repo, nil, nil)

			// WHEN
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_ListByBundleIDs(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")

	tnt := "tenant"
	externalTnt := "external-tenant"
	firstDocID := "foo"
	secondDocID := "foo2"
	firstBundleID := "bar"
	secondBundleID := "bar2"
	bundleIDs := []string{firstBundleID, secondBundleID}

	docFirstBundle := fixModelDocumentForApp(firstDocID, firstBundleID)
	docSecondBundle := fixModelDocumentForApp(secondDocID, secondBundleID)

	docsFirstBundle := []*model.Document{docFirstBundle}
	docsSecondBundle := []*model.Document{docSecondBundle}

	docPageFirstBundle := &model.DocumentPage{
		Data:       docsFirstBundle,
		TotalCount: len(docsFirstBundle),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}
	docPageSecondBundle := &model.DocumentPage{
		Data:       docsSecondBundle,
		TotalCount: len(docsSecondBundle),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	docPages := []*model.DocumentPage{docPageFirstBundle, docPageSecondBundle}

	after := "test"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.DocumentRepository
		ExpectedResult     []*model.DocumentPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("ListByBundleIDs", ctx, tnt, bundleIDs, 2, after).Return(docPages, nil).Once()
				return repo
			},
			PageSize:       2,
			ExpectedResult: docPages,
		},
		{
			Name: "Return error when page size is less than 1",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			PageSize:           0,
			ExpectedResult:     docPages,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Return error when page size is bigger than 200",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			PageSize:           201,
			ExpectedResult:     docPages,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when Documents listing failed",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("ListByBundleIDs", ctx, tnt, bundleIDs, 2, after).Return(nil, testErr).Once()
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

			svc := document.NewService(repo, nil, nil)

			// WHEN
			docs, err := svc.ListByBundleIDs(ctx, bundleIDs, testCase.PageSize, after)

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
		svc := document.NewService(nil, nil, nil)
		// WHEN
		_, err := svc.ListByBundleIDs(context.TODO(), nil, 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListFetchRequests(t *testing.T) {
	// GIVEN
	tnt := "tenant"
	externalTnt := "external-tenant"
	testErr := errors.New("test error")
	frURL := "foo.bar"
	firstFRID := "frID"
	secondFRID := "frID2"
	firstDocID := "docID"
	secondDocID := "docID2"
	docIDs := []string{firstDocID, secondDocID}
	timestamp := time.Now()

	firstFetchRequest := fixModelFetchRequest(firstFRID, frURL, timestamp)
	secondFetchRequest := fixModelFetchRequest(secondFRID, frURL, timestamp)
	fetchRequests := []*model.FetchRequest{firstFetchRequest, secondFetchRequest}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.FetchRequestRepository
		ExpectedResult     []*model.FetchRequest
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("ListByReferenceObjectIDs", ctx, tnt, model.DocumentFetchRequestReference, docIDs).Return(fetchRequests, nil).Once()
				return repo
			},
			ExpectedResult: fetchRequests,
		},
		{
			Name: "Returns error when Fetch Requests listing failed",
			RepositoryFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("ListByReferenceObjectIDs", ctx, tnt, model.DocumentFetchRequestReference, docIDs).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := document.NewService(nil, repo, nil)

			// WHEN
			frs, err := svc.ListFetchRequests(ctx, docIDs)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, frs)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := document.NewService(nil, nil, nil)
		// WHEN
		_, err := svc.ListFetchRequests(context.TODO(), nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func fixUIDSvc() *automock.UIDService {
	svc := &automock.UIDService{}
	svc.On("Generate").Return(docID)
	return svc
}
