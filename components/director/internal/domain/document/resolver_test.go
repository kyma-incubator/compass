package document_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/stretchr/testify/assert"
)

var contextParam = mock.MatchedBy(func(ctx context.Context) bool {
	persistenceOp, err := persistence.FromCtx(ctx)
	return err == nil && persistenceOp != nil
})

func TestResolver_AddDocumentToBundle(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	bundleID := "bar"
	id := "bar"
	modelBundle := fixModelBundle(bundleID)
	modelDocument := fixModelDocumentForApp(id, bundleID)
	gqlDocument := fixGQLDocument(id, bundleID)
	gqlInput := fixGQLDocumentInput(id)
	modelInput := fixModelDocumentInput(id)

	testCases := []struct {
		Name             string
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn        func() *automock.DocumentService
		BndlServiceFn    func() *automock.BundleService
		ConverterFn      func() *automock.DocumentConverter
		ExpectedDocument *graphql.Document
		ExpectedErr      error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("CreateInBundle", contextParam, resource.Application, appID, bundleID, *modelInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(modelDocument, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", contextParam, bundleID).Return(modelBundle, nil)
				return appSvc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput, nil).Once()
				conv.On("ToGraphQL", modelDocument).Return(gqlDocument).Once()
				return conv
			},
			ExpectedDocument: gqlDocument,
			ExpectedErr:      nil,
		},
		{
			Name: "Returns error when bundle does not exits",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", contextParam, bundleID).Return(nil, apperrors.NewNotFoundError(resource.Bundle, bundleID))
				return appSvc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput, nil).Once()
				return conv
			},

			ExpectedDocument: nil,
			ExpectedErr:      errors.New("cannot add Document to not existing Bundle"),
		},
		{
			Name: "Returns error when application existence check failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", contextParam, bundleID).Return(modelBundle, testErr)
				return appSvc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput, nil).Once()
				return conv
			},

			ExpectedDocument: nil,
			ExpectedErr:      testErr,
		},
		{
			Name: "Returns error when document creation failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("CreateInBundle", contextParam, resource.Application, appID, bundleID, *modelInput).Return("", testErr).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", contextParam, bundleID).Return(modelBundle, nil)
				return appSvc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput, nil).Once()
				return conv
			},
			ExpectedDocument: nil,
			ExpectedErr:      testErr,
		},
		{
			Name: "Returns error when document retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("CreateInBundle", contextParam, resource.Application, appID, bundleID, *modelInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Get", contextParam, bundleID).Return(modelBundle, nil)
				return appSvc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput, nil).Once()
				return conv
			},
			ExpectedDocument: nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			bndlSvc := testCase.BndlServiceFn()
			converter := testCase.ConverterFn()

			resolver := document.NewResolver(transact, svc, nil, bndlSvc, nil)
			resolver.SetConverter(converter)

			// WHEN
			result, err := resolver.AddDocumentToBundle(context.TODO(), bundleID, *gqlInput)

			// then
			assert.Equal(t, testCase.ExpectedDocument, result)
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			persistTx.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			bndlSvc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteDocument(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "bar"
	bundleID := "bar"
	modelDocument := fixModelDocumentForApp(id, bundleID)
	gqlDocument := fixGQLDocument(id, bundleID)

	testCases := []struct {
		Name             string
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn        func() *automock.DocumentService
		ConverterFn      func() *automock.DocumentConverter
		ExpectedDocument *graphql.Document
		ExpectedErr      error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Get", contextParam, id).Return(modelDocument, nil).Once()
				svc.On("Delete", contextParam, id).Return(nil).Once()
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
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Get", contextParam, id).Return(nil, testErr).Once()
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
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Get", contextParam, id).Return(modelDocument, nil).Once()
				svc.On("Delete", contextParam, id).Return(testErr).Once()
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
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := document.NewResolver(transact, svc, nil, nil, nil)
			resolver.SetConverter(converter)

			// WHEN
			result, err := resolver.DeleteDocument(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedDocument, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persistTx.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_FetchRequest(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")

	firstDocID := "docID"
	secondDocID := "docID2"
	docIDs := []string{firstDocID, secondDocID}
	firstFRID := "frID"
	secondFRID := "frID2"
	frURL := "foo.bar"
	timestamp := time.Now()

	frFirstDoc := fixModelFetchRequest(firstFRID, frURL, timestamp)
	frSecondDoc := fixModelFetchRequest(secondFRID, frURL, timestamp)
	fetchRequests := []*model.FetchRequest{frFirstDoc, frSecondDoc}

	gqlFRFirstDoc := fixGQLFetchRequest(frURL, timestamp)
	gqlFRSecondDoc := fixGQLFetchRequest(frURL, timestamp)
	gqlFetchRequests := []*graphql.FetchRequest{gqlFRFirstDoc, gqlFRSecondDoc}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.DocumentService
		ConverterFn     func() *automock.FetchRequestConverter
		ExpectedResult  []*graphql.FetchRequest
		ExpectedErr     []error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("ListFetchRequests", txtest.CtxWithDBMatcher(), docIDs).Return(fetchRequests, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("ToGraphQL", frFirstDoc).Return(gqlFRFirstDoc, nil).Once()
				conv.On("ToGraphQL", frSecondDoc).Return(gqlFRSecondDoc, nil).Once()
				return conv
			},
			ExpectedResult: gqlFetchRequests,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "FetchRequest doesn't exist",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("ListFetchRequests", txtest.CtxWithDBMatcher(), docIDs).Return(nil, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    nil,
		},
		{
			Name:            "Error when listing Document FetchRequests",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("ListFetchRequests", txtest.CtxWithDBMatcher(), docIDs).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Error when converting FetchRequest to graphql",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("ListFetchRequests", txtest.CtxWithDBMatcher(), docIDs).Return(fetchRequests, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("ToGraphQL", frFirstDoc).Return(nil, testErr).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("ListFetchRequests", txtest.CtxWithDBMatcher(), docIDs).Return(fetchRequests, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("ToGraphQL", frFirstDoc).Return(gqlFRFirstDoc, nil).Once()
				conv.On("ToGraphQL", frSecondDoc).Return(gqlFRSecondDoc, nil).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			firstFRParams := dataloader.ParamFetchRequestDocument{ID: firstDocID, Ctx: context.TODO()}
			secondFRParams := dataloader.ParamFetchRequestDocument{ID: secondDocID, Ctx: context.TODO()}
			keys := []dataloader.ParamFetchRequestDocument{firstFRParams, secondFRParams}
			resolver := document.NewResolver(transact, svc, nil, nil, converter)

			// WHEN
			result, err := resolver.FetchRequestDocumentDataLoader(keys)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
	t.Run("Returns error when there are no Docs", func(t *testing.T) {
		resolver := document.NewResolver(nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.FetchRequestDocumentDataLoader([]dataloader.ParamFetchRequestDocument{})
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInternalError("No Documents found").Error())
	})

	t.Run("Returns error when Document ID is empty", func(t *testing.T) {
		params := dataloader.ParamFetchRequestDocument{ID: "", Ctx: context.TODO()}
		keys := []dataloader.ParamFetchRequestDocument{params}

		resolver := document.NewResolver(nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.FetchRequestDocumentDataLoader(keys)
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInternalError("Cannot fetch FetchRequest. Document ID is empty").Error())
	})
}
