package document_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"

	documentModel := fixModelDocument("1", id)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

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
				repo.On("GetByID", id).Return(documentModel, nil).Once()
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
				repo.On("GetByID", id).Return(nil, testErr).Once()
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

			svc := document.NewService(repo, nil)

			// when
			document, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedDocument, document)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"

	modelDocuments := []*model.Document{
		fixModelDocument(applicationID, "foo"),
		fixModelDocument(applicationID, "bar"),
		fixModelDocument("baz", "bar"),
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

	first := 2
	after := "test"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.DocumentRepository
		InputPageSize      *int
		InputCursor        *string
		ExpectedResult     *model.DocumentPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("ListByApplicationID", applicationID, &first, &after).Return(documentPage, nil).Once()
				return repo
			},
			InputPageSize:      &first,
			InputCursor:        &after,
			ExpectedResult:     documentPage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when document listing failed",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("ListByApplicationID", applicationID, &first, &after).Return(nil, testErr).Once()
				return repo
			},
			InputPageSize:      &first,
			InputCursor:        &after,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := document.NewService(repo, nil)

			// when
			docs, err := svc.List(ctx, applicationID, testCase.InputPageSize, testCase.InputCursor)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, docs)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	modelInput := fixModelDocumentInput("foo")
	id := "foo"
	applicationID := "foo"
	modelDoc := modelInput.ToDocument(id, tnt, applicationID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.DocumentRepository
		UIDServiceFn func() *automock.UIDService
		Input        model.DocumentInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Create", modelDoc).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			Input:       *modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when document creation failed",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Create", modelDoc).Return(testErr).Once()
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			idSvc := testCase.UIDServiceFn()
			svc := document.NewService(repo, idSvc)

			// when
			result, err := svc.Create(ctx, applicationID, testCase.Input)

			// then
			assert.IsType(t, "string", result)
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	id := "bar"
	documentModel := fixModelDocument(applicationID, id)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

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
				repo.On("GetByID", id).Return(documentModel, nil).Once()
				repo.On("Delete", documentModel).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when document deletion failed",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("GetByID", id).Return(documentModel, nil).Once()
				repo.On("Delete", documentModel).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when document retrieval failed",
			RepositoryFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("GetByID", id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := document.NewService(repo, nil)

			// when
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}
