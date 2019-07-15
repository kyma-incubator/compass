package document_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestResolver_AddDocument(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	id := "bar"
	modelDocument := fixModelDocument(id, applicationID)
	gqlDocument := fixGQLDocument(id, applicationID)
	gqlInput := fixGQLDocumentInput(id)
	modelInput := fixModelDocumentInput(id)

	testCases := []struct {
		Name             string
		ServiceFn        func() *automock.DocumentService
		AppServiceFn     func() *automock.ApplicationService
		ConverterFn      func() *automock.DocumentConverter
		ExpectedDocument *graphql.Document
		ExpectedErr      error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Create", context.TODO(), applicationID, *modelInput).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(modelDocument, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), applicationID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelDocument).Return(gqlDocument).Once()
				return conv
			},
			ExpectedDocument: gqlDocument,
			ExpectedErr:      nil,
		},
		{
			Name: "Returns error when application not exits",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), applicationID).Return(false, nil)
				return appSvc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},

			ExpectedDocument: nil,
			ExpectedErr:      errors.New("Cannot add Document to not existing Application"),
		},
		{
			Name: "Returns error when application existence check failed",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), applicationID).Return(false, testErr)
				return appSvc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},

			ExpectedDocument: nil,
			ExpectedErr:      testErr,
		},
		{
			Name: "Returns error when document creation failed",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Create", context.TODO(), applicationID, *modelInput).Return("", testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), applicationID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			ExpectedDocument: nil,
			ExpectedErr:      testErr,
		},
		{
			Name: "Returns error when document retrieval failed",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Create", context.TODO(), applicationID, *modelInput).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), applicationID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			ExpectedDocument: nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			appSvc := testCase.AppServiceFn()
			converter := testCase.ConverterFn()

			resolver := document.NewResolver(svc, appSvc, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.AddDocument(context.TODO(), applicationID, *gqlInput)

			// then
			assert.Equal(t, testCase.ExpectedDocument, result)
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			svc.AssertExpectations(t)
			appSvc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteDocument(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	id := "bar"
	modelDocument := fixModelDocument(id, applicationID)
	gqlDocument := fixGQLDocument(id, applicationID)

	testCases := []struct {
		Name             string
		ServiceFn        func() *automock.DocumentService
		ConverterFn      func() *automock.DocumentConverter
		ExpectedDocument *graphql.Document
		ExpectedErr      error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Get", context.TODO(), id).Return(modelDocument, nil).Once()
				svc.On("Delete", context.TODO(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("ToGraphQL", modelDocument).Return(gqlDocument).Once()
				return conv
			},
			ExpectedDocument: gqlDocument,
			ExpectedErr:      nil,
		},
		{
			Name: "Returns error when document retrieval failed",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			ExpectedDocument: nil,
			ExpectedErr:      testErr,
		},
		{
			Name: "Returns error when document deletion failed",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Get", context.TODO(), id).Return(modelDocument, nil).Once()
				svc.On("Delete", context.TODO(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("ToGraphQL", modelDocument).Return(gqlDocument).Once()
				return conv
			},
			ExpectedDocument: nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := document.NewResolver(svc, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.DeleteDocument(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedDocument, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}
