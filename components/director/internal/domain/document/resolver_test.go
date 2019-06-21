package document_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document/automock"
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/stretchr/testify/assert"
)

func TestResolver_AddDocument(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	id := "bar"
	modelDocument := fixModelDocument(applicationID, id)
	gqlDocument := fixGQLDocument(id)
	gqlInput := fixGQLDocumentInput(id)
	modelInput := fixModelDocumentInput(id)

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
				svc.On("Create", context.TODO(), *modelInput, applicationID).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(modelDocument, nil).Once()
				return svc
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
			Name: "Create Error",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Create", context.TODO(), *modelInput, applicationID).Return("", testErr).Once()
				return svc
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
			Name: "Get Error",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Create", context.TODO(), *modelInput, applicationID).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
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

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := document.NewResolver(svc, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.AddDocument(context.TODO(), applicationID, *gqlInput)

			// then
			assert.Equal(t, testCase.ExpectedDocument, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteDocument(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	id := "bar"
	modelDocument := fixModelDocument(applicationID, id)
	gqlDocument := fixGQLDocument(id)

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
			Name: "Not Found",
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Get", context.TODO(), id).Return(nil, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			ExpectedDocument: nil,
			ExpectedErr:      nil,
		},
		{
			Name: "Get Error",
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
			Name: "Delete Error",
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

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := document.NewResolver(svc, nil)
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
