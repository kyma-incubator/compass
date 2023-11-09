package processor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProductProcessor_Process(t *testing.T) {
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	applicationID := "application-id"
	applicationTemplateVersionID := "application-template-version-id"

	productModels := []*model.Product{
		{
			ID:                           "product-id",
			OrdID:                        "ord:product",
			ApplicationID:                &applicationID,
			ApplicationTemplateVersionID: &applicationTemplateVersionID,
		},
	}

	productInputs := []*model.ProductInput{
		{
			OrdID: "ord:product",
		},
	}

	productCreateInputs := []*model.ProductInput{
		{
			OrdID: "ord:product2",
		},
	}

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ProductSvcFn    func() *automock.ProductService
		InputResource   resource.Type
		InputResourceID string
		InputProducts   []*model.ProductInput
		ExpectedOutput  []*model.Product
		ExpectedErr     error
	}{
		{
			Name: "Success for application resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			ProductSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(productModels, nil).Twice()
				productSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, productModels[0].ID, *productInputs[0]).Return(nil).Once()
				return productSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputProducts:   productInputs,
			ExpectedOutput:  productModels,
		},
		{
			Name: "Success for application template version resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			ProductSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), applicationTemplateVersionID).Return(productModels, nil).Twice()
				productSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, productModels[0].ID, *productInputs[0]).Return(nil).Once()
				return productSvc
			},
			InputResource:   resource.ApplicationTemplateVersion,
			InputResourceID: applicationTemplateVersionID,
			InputProducts:   productInputs,
			ExpectedOutput:  productModels,
		},
		{
			Name: "Fail on begin transaction while listing products",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			ProductSvcFn: func() *automock.ProductService {
				return &automock.ProductService{}
			},
			InputResource:   resource.Application,
			InputResourceID: "",
			InputProducts:   []*model.ProductInput{},
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while listing products by application id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			ProductSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(nil, testErr).Once()
				return productSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputProducts:   []*model.ProductInput{},
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while listing products by application template version id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			ProductSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), applicationTemplateVersionID).Return(nil, testErr).Once()
				return productSvc
			},
			InputResource:   resource.ApplicationTemplateVersion,
			InputResourceID: applicationTemplateVersionID,
			InputProducts:   []*model.ProductInput{},
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail on begin transaction while re-syncing products",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, testErr).Once()
				return persistTx, transact
			},
			ProductSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return([]*model.Product{}, nil).Once()
				return productSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputProducts:   productInputs,
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while updating product",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			ProductSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(productModels, nil).Once()
				productSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, productModels[0].ID, *productInputs[0]).Return(testErr).Once()
				return productSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputProducts:   productInputs,
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while creating product",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			ProductSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(productModels, nil).Once()
				productSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, applicationID, *productCreateInputs[0]).Return("", testErr).Once()
				return productSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputProducts:   productCreateInputs,
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while listing resources after update",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsTwice()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			ProductSvcFn: func() *automock.ProductService {
				productSvc := &automock.ProductService{}
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(productModels, nil).Once()
				productSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(nil, testErr).Once()

				productSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, productModels[0].ID, *productInputs[0]).Return(nil).Once()
				return productSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputProducts:   productInputs,
			ExpectedErr:     testErr,
		},
	}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

			productSvc := test.ProductSvcFn()

			productProcessor := processor.NewProductProcessor(tx, productSvc)
			result, err := productProcessor.Process(context.TODO(), test.InputResource, test.InputResourceID, test.InputProducts)
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.ExpectedOutput, result)
			}

			mock.AssertExpectationsForObjects(t, tx, productSvc)
		})
	}
}
