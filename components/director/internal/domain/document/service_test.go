package document_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

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
	// given
	testErr := errors.New("Test error")

	id := "foo"

	documentModel := fixModelDocument("1", id)
	tnt := documentModel.Tenant
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, documentModel.Tenant, externalTnt)

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

			// when
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
	// given
	testErr := errors.New("Test error")
	id := "foo"
	bndlID := bndlID()
	tenantID := "bar"
	externalTenantID := "external-tenant"

	bundleID := "test"
	doc := fixModelDocument(id, bndlID)

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

			// when
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

func TestService_ListForBundle(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	bundleID := "bar"
	modelDocuments := []*model.Document{
		fixModelDocument("foo", bundleID),
		fixModelDocument("bar", bundleID),
		fixModelDocument("baz", bundleID),
	}
	documentPage := &model.DocumentPage{
		Data:       modelDocuments,
		TotalCount: len(modelDocuments),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	tnt := modelDocuments[0].Tenant

	first := 2
	after := "test"
	externalTenantID := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, modelDocuments[0].Tenant, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.DocumentRepository
		ExpectedResult     *model.DocumentPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("ListForBundle", ctx, tnt, bundleID, first, after).Return(documentPage, nil).Once()
				return repo
			},
			ExpectedResult:     documentPage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when document listing failed",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("ListForBundle", ctx, tnt, bundleID, first, after).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := document.NewService(repo, nil, nil)

			// when
			docs, err := svc.ListForBundle(ctx, bundleID, first, after)

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
func TestService_CreateToBundle(t *testing.T) {
	// given
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
	modelDoc := modelInput.ToDocumentWithinBundle(id, tnt, bundleID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.DocumentRepository
		FetchRequestRepoFn func() *automock.FetchRequestRepository
		UIDServiceFn       func() *automock.UIDService
		Input              model.DocumentInput
		ExpectedErr        error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Create", ctx, modelDoc).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, fixModelFetchRequest(frID, frURL, timestamp)).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			Input:       *modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when document creation failed",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Create", ctx, modelDoc).Return(testErr).Once()
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
			Input:       *modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - Fetch Request Creation",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Create", ctx, modelDoc).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, fixModelFetchRequest(frID, frURL, timestamp)).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			Input:       *modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			idSvc := testCase.UIDServiceFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			svc := document.NewService(repo, fetchRequestRepo, idSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			result, err := svc.CreateInBundle(ctx, bundleID, testCase.Input)

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
		svc := document.NewService(nil, nil, nil)
		// when
		_, err := svc.CreateInBundle(context.TODO(), "Dd", model.DocumentInput{})
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}
func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	bundleID := "foobar"
	documentModel := fixModelDocument(id, bundleID)

	tnt := documentModel.Tenant
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, documentModel.Tenant, externalTnt)

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

			// when
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

func TestService_GetFetchRequest(t *testing.T) {
	// given
	tnt := "tenant"
	externalTnt := "external-tenant"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	refID := "doc-id"
	frURL := "foo.bar"
	timestamp := time.Now()

	fetchRequestModel := fixModelFetchRequest("foo", frURL, timestamp)

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.DocumentRepository
		FetchRequestRepoFn   func() *automock.FetchRequestRepository
		ExpectedFetchRequest *model.FetchRequest
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Exists", ctx, tnt, refID).Return(true, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tnt, model.DocumentFetchRequestReference, refID).Return(fetchRequestModel, nil).Once()
				return repo
			},
			ExpectedFetchRequest: fetchRequestModel,
			ExpectedErrMessage:   "",
		},
		{
			Name: "Success - Not Found",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Exists", ctx, tnt, refID).Return(true, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tnt, model.DocumentFetchRequestReference, refID).Return(nil, apperrors.NewNotFoundError(resource.Document, "")).Once()
				return repo
			},
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   "",
		},
		{
			Name: "Error - Get FetchRequest",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Exists", ctx, tnt, refID).Return(true, nil).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tnt, model.DocumentFetchRequestReference, refID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name: "Error - Document doesn't exist",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Exists", ctx, tnt, refID).Return(false, nil).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			ExpectedErrMessage:   "Document with ID doc-id doesn't exist",
			ExpectedFetchRequest: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			svc := document.NewService(repo, fetchRequestRepo, nil)

			// when
			l, err := svc.GetFetchRequest(ctx, refID)

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
}

func TestService_ListAllByBundleIDs(t *testing.T) {
	// given
	testErr := errors.New("test error")

	tnt := "tenant"
	externalTnt := "external-tenant"
	firstDocID := "foo"
	secondDocID := "foo2"
	firstBundleID := "bar"
	secondBundleID := "bar2"
	bundleIDs := []string{firstBundleID, secondBundleID}

	docFirstBundle := fixModelDocument(firstDocID, firstBundleID)
	docSecondBundle := fixModelDocument(secondDocID, secondBundleID)

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
				repo.On("ListAllForBundle", ctx, tnt, bundleIDs, 2, after).Return(docPages, nil).Once()
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
				repo.On("ListAllForBundle", ctx, tnt, bundleIDs, 2, after).Return(nil, testErr).Once()
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

			// when
			docs, err := svc.ListAllByBundleIDs(ctx, bundleIDs, testCase.PageSize, after)

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
		_, err := svc.ListAllByBundleIDs(context.TODO(), nil, 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListFetchRequests(t *testing.T) {
	// given
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

			// when
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
