package processor_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	dataProductID     = "data-product-id"
	dataProductORDID1 = "sap.foo.bar:dataProduct:CustomerOrder:v1"
	dataProductORDID2 = "sap.foo.bar:dataProduct:CustomerOrder2:v2"
)

func TestDataProductProcessor_Process(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	dataProductModel := []*model.DataProduct{
		fixDataProductModel(dataProductID, dataProductORDID1),
	}

	updatedDataProductModel := []*model.DataProduct{
		fixDataProductModel(dataProductID, dataProductORDID1),
	}

	hashDataProduct, _ := ord.HashObject(dataProductModel)
	resourceHashes := map[string]uint64{
		dataProductORDID1: hashDataProduct,
	}

	uintHashDataProduct := strconv.FormatUint(hashDataProduct, 10)
	dataProductModel[0].ResourceHash = &uintHashDataProduct

	dataProductInputs := []*model.DataProductInput{
		fixDataProductInputModel(dataProductORDID1),
	}

	dataProductCreateInputs := []*model.DataProductInput{
		fixDataProductInputModel(dataProductORDID2),
	}

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		DataProductSvcFn    func() *automock.DataProductService
		InputResource       resource.Type
		InputResourceID     string
		InputDataProducts   []*model.DataProductInput
		InputResourceHashes map[string]uint64
		ExpectedOutput      []*model.DataProduct
		ExpectedErr         error
	}{
		{
			Name: "Successful Process for application resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			DataProductSvcFn: func() *automock.DataProductService {
				dataProductSvc := &automock.DataProductService{}
				dataProductSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(dataProductModel, nil).Twice()
				dataProductSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, appID, dataProductModel[0].ID, str.Ptr(packageID1), *dataProductInputs[0], hashDataProduct).Return(nil).Once()
				return dataProductSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputDataProducts:   dataProductInputs,
			InputResourceHashes: resourceHashes,
			ExpectedOutput:      dataProductModel,
		},
		{
			Name: "Successful Process for application template version resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			DataProductSvcFn: func() *automock.DataProductService {
				dataProductSvc := &automock.DataProductService{}
				dataProductSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(dataProductModel, nil).Twice()
				dataProductSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, dataProductModel[0].ID, str.Ptr(packageID1), *dataProductInputs[0], hashDataProduct).Return(nil).Once()
				return dataProductSvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputDataProducts:   dataProductInputs,
			InputResourceHashes: resourceHashes,
			ExpectedOutput:      dataProductModel,
		},
		{
			Name: "Success when creating Data Product for application resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			DataProductSvcFn: func() *automock.DataProductService {
				dataProductSvc := &automock.DataProductService{}
				dataProductSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.DataProduct{}, nil).Once()

				// set to time.Now, because on Create the lastUpdate is set to current time
				currentTime := time.Now().Format(time.RFC3339)
				dataProductInputs[0].LastUpdate = &currentTime

				dataProductSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID1), *dataProductInputs[0], mock.Anything).Return(dataProductID, nil).Once()
				dataProductSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(dataProductModel, nil).Once()
				return dataProductSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputDataProducts:   dataProductInputs,
			InputResourceHashes: resourceHashes,
			ExpectedOutput:      dataProductModel,
		},
		{
			Name: "Fail on begin transaction while listing Data Products",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			DataProductSvcFn: func() *automock.DataProductService {
				return &automock.DataProductService{}
			},
			InputResource:       resource.Application,
			InputResourceID:     "",
			InputDataProducts:   []*model.DataProductInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while listing Data Products by application id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			DataProductSvcFn: func() *automock.DataProductService {
				dataProductSvc := &automock.DataProductService{}
				dataProductSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return dataProductSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputDataProducts:   []*model.DataProductInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while listing Data Products by application template version id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			DataProductSvcFn: func() *automock.DataProductService {
				dataProductSvc := &automock.DataProductService{}
				dataProductSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, testErr).Once()
				return dataProductSvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputDataProducts:   []*model.DataProductInput{},
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail on begin transaction on resync in transaction for Data Product",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, testErr).Once()
				return persistTx, transact
			},
			DataProductSvcFn: func() *automock.DataProductService {
				dataProductSvc := &automock.DataProductService{}
				dataProductSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.DataProduct{}, nil).Once()
				return dataProductSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputDataProducts:   dataProductInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while creating Data Product",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			DataProductSvcFn: func() *automock.DataProductService {
				dataProductSvc := &automock.DataProductService{}
				dataProductSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(dataProductModel, nil).Once()

				// set to time.Now, because on Create the lastUpdate is set to current time
				currentTime := time.Now().Format(time.RFC3339)
				dataProductCreateInputs[0].LastUpdate = &currentTime

				dataProductSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID1), *dataProductCreateInputs[0], mock.Anything).Return("", testErr).Once()
				return dataProductSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputDataProducts:   dataProductCreateInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while updating Data Product",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			DataProductSvcFn: func() *automock.DataProductService {
				dataProductSvc := &automock.DataProductService{}
				dataProductSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(dataProductModel, nil).Once()
				dataProductSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, appID, dataProductModel[0].ID, str.Ptr(packageID1), *dataProductInputs[0], hashDataProduct).Return(testErr).Once()
				return dataProductSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputDataProducts:   dataProductInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail while listing resources after resync of Data Product",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsTwice()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			DataProductSvcFn: func() *automock.DataProductService {
				dataProductSvc := &automock.DataProductService{}
				dataProductSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(dataProductModel, nil).Once()
				dataProductSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()

				dataProductSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, appID, dataProductModel[0].ID, str.Ptr(packageID1), *dataProductInputs[0], hashDataProduct).Return(nil).Once()
				return dataProductSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputDataProducts:   dataProductInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Success when updating package id for Data Product",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			DataProductSvcFn: func() *automock.DataProductService {
				dataProductSvc := &automock.DataProductService{}
				dataProductInputs[0].OrdPackageID = str.Ptr(packageORDID2)
				dataProductSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(dataProductModel, nil).Once()
				dataProductSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(updatedDataProductModel, nil).Once()
				dataProductSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, appID, dataProductModel[0].ID, str.Ptr(packageID2), *dataProductInputs[0], hashDataProduct).Return(nil).Once()
				return dataProductSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputDataProducts:   dataProductInputs,
			InputResourceHashes: resourceHashes,
			ExpectedOutput:      updatedDataProductModel,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()
			dataProductSvc := test.DataProductSvcFn()

			dataProductProcessor := processor.NewDataProductProcessor(tx, dataProductSvc)
			result, err := dataProductProcessor.Process(context.TODO(), test.InputResource, test.InputResourceID, fixPackages(), test.InputDataProducts, test.InputResourceHashes)
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.ExpectedOutput, result)
			}

			mock.AssertExpectationsForObjects(t, tx, dataProductSvc)
		})
	}
}
