package document_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestResolver_AddDocument(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	packageID := "bar"
	id := "bar"
	modelDocument := fixModelDocument(id, applicationID, packageID)
	gqlDocument := fixGQLDocument(id, applicationID, packageID)
	gqlInput := fixGQLDocumentInput(id)
	modelInput := fixModelDocumentInput(id)

	testCases := []struct {
		Name             string
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn        func() *automock.DocumentService
		AppServiceFn     func() *automock.ApplicationService
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
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Create", contextParam, applicationID, *modelInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(modelDocument, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, applicationID).Return(true, nil)
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
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, applicationID).Return(false, nil)
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
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, applicationID).Return(false, testErr)
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
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Create", contextParam, applicationID, *modelInput).Return("", testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, applicationID).Return(true, nil)
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
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("Create", contextParam, applicationID, *modelInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, applicationID).Return(true, nil)
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
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			appSvc := testCase.AppServiceFn()
			converter := testCase.ConverterFn()

			resolver := document.NewResolver(transact, svc, appSvc, nil, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.AddDocument(context.TODO(), applicationID, *gqlInput)

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
			appSvc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_AddDocumentToPackage(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	packageID := "bar"
	id := "bar"
	modelDocument := fixModelDocument(id, applicationID, packageID)
	gqlDocument := fixGQLDocument(id, applicationID, packageID)
	gqlInput := fixGQLDocumentInput(id)
	modelInput := fixModelDocumentInput(id)

	testCases := []struct {
		Name             string
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn        func() *automock.DocumentService
		PkgServiceFn     func() *automock.PackageService
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
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("CreateInPackage", contextParam, packageID, *modelInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(modelDocument, nil).Once()
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				appSvc := &automock.PackageService{}
				appSvc.On("Exist", contextParam, packageID).Return(true, nil)
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
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				appSvc := &automock.PackageService{}
				appSvc.On("Exist", contextParam, packageID).Return(false, nil)
				return appSvc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},

			ExpectedDocument: nil,
			ExpectedErr:      errors.New("Cannot add Document to not existing Package"),
		},
		{
			Name: "Returns error when application existence check failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				appSvc := &automock.PackageService{}
				appSvc.On("Exist", contextParam, packageID).Return(false, testErr)
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
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("CreateInPackage", contextParam, packageID, *modelInput).Return("", testErr).Once()
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				appSvc := &automock.PackageService{}
				appSvc.On("Exist", contextParam, packageID).Return(true, nil)
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
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("CreateInPackage", contextParam, packageID, *modelInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			PkgServiceFn: func() *automock.PackageService {
				appSvc := &automock.PackageService{}
				appSvc.On("Exist", contextParam, packageID).Return(true, nil)
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
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			pkgSvc := testCase.PkgServiceFn()
			converter := testCase.ConverterFn()

			resolver := document.NewResolver(transact, svc, nil, pkgSvc, nil)
			resolver.SetConverter(converter)

			// when
			result, err := resolver.AddDocumentToPackage(context.TODO(), packageID, *gqlInput)

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
			pkgSvc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteDocument(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	id := "bar"
	packageID := "bar"
	modelDocument := fixModelDocument(id, applicationID, packageID)
	gqlDocument := fixGQLDocument(id, applicationID, packageID)

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
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
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
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
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
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()

				return transact
			},
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

			// when
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
	// given
	testErr := errors.New("Test error")

	id := "bar"
	url := "foo.bar"

	timestamp := time.Now()
	frModel := fixModelFetchRequest("foo", url, timestamp)
	frGQL := fixGQLFetchRequest(url, timestamp)
	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.DocumentService
		ConverterFn     func() *automock.FetchRequestConverter
		ExpectedResult  *graphql.FetchRequest
		ExpectedErr     error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetFetchRequest", contextParam, id).Return(frModel, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("ToGraphQL", frModel).Return(frGQL).Once()
				return conv
			},
			ExpectedResult: frGQL,
			ExpectedErr:    nil,
		},
		{
			Name: "Doesn't exist",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetFetchRequest", contextParam, id).Return(nil, nil).Once()
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
			Name: "Error",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetFetchRequest", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := document.NewResolver(transact, svc, nil, nil, converter)

			// when
			result, err := resolver.FetchRequest(context.TODO(), &graphql.Document{ID: id})

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persistTx.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}
