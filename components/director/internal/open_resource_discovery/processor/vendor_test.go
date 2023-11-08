package processor_test

import (
	"context"
	"errors"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVendorProcessor_Process(t *testing.T) {
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	applicationID := "application-id"
	applicationTemplateVersionID := "application-template-version-id"

	vendorModels := []*model.Vendor{
		{
			ID:                           "vendor-id",
			OrdID:                        "ord:vendor",
			ApplicationID:                &applicationID,
			ApplicationTemplateVersionID: &applicationTemplateVersionID,
		},
	}

	vendorInputs := []*model.VendorInput{
		{
			OrdID: "ord:vendor",
		},
	}

	vendorCreateInputs := []*model.VendorInput{
		{
			OrdID: "ord:vendor2",
		},
	}

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		VendorSvcFn     func() *automock.VendorService
		InputResource   resource.Type
		InputResourceID string
		InputVendors    []*model.VendorInput
		ExpectedOutput  []*model.Vendor
		ExpectedErr     error
	}{
		{
			Name: "Success for application resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			VendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(vendorModels, nil).Twice()
				vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, vendorModels[0].ID, *vendorInputs[0]).Return(nil).Once()
				return vendorSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputVendors:    vendorInputs,
			ExpectedOutput:  vendorModels,
		},
		{
			Name: "Success for application template version resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			VendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), applicationTemplateVersionID).Return(vendorModels, nil).Twice()
				vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, vendorModels[0].ID, *vendorInputs[0]).Return(nil).Once()
				return vendorSvc
			},
			InputResource:   resource.ApplicationTemplateVersion,
			InputResourceID: applicationTemplateVersionID,
			InputVendors:    vendorInputs,
			ExpectedOutput:  vendorModels,
		},
		{
			Name: "Fail on begin transaction while listing vendors",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			VendorSvcFn: func() *automock.VendorService {
				return &automock.VendorService{}
			},
			InputResource:   resource.Application,
			InputResourceID: "",
			InputVendors:    []*model.VendorInput{},
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while listing vendors by application id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			VendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(nil, testErr).Once()
				return vendorSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputVendors:    []*model.VendorInput{},
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while listing vendors by application template version id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			VendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), applicationTemplateVersionID).Return(nil, testErr).Once()
				return vendorSvc
			},
			InputResource:   resource.ApplicationTemplateVersion,
			InputResourceID: applicationTemplateVersionID,
			InputVendors:    []*model.VendorInput{},
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail on begin transaction while re-syncing vendors",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, testErr).Once()
				return persistTx, transact
			},
			VendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return([]*model.Vendor{}, nil).Once()
				return vendorSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputVendors:    vendorInputs,
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while updating vendor",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			VendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(vendorModels, nil).Once()
				vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, vendorModels[0].ID, *vendorInputs[0]).Return(testErr).Once()
				return vendorSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputVendors:    vendorInputs,
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while creating vendor",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			VendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(vendorModels, nil).Once()
				vendorSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, applicationID, *vendorCreateInputs[0]).Return("", testErr).Once()
				return vendorSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputVendors:    vendorCreateInputs,
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
			VendorSvcFn: func() *automock.VendorService {
				vendorSvc := &automock.VendorService{}
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(vendorModels, nil).Once()
				vendorSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(nil, testErr).Once()

				vendorSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, vendorModels[0].ID, *vendorInputs[0]).Return(nil).Once()
				return vendorSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputVendors:    vendorInputs,
			ExpectedErr:     testErr,
		},
	}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

			vendorSvc := test.VendorSvcFn()

			vendorProcessor := processor.NewVendorProcessor(tx, vendorSvc)
			result, err := vendorProcessor.Process(context.TODO(), test.InputResource, test.InputResourceID, test.InputVendors)
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.ExpectedOutput, result)
			}

			mock.AssertExpectationsForObjects(t, tx, vendorSvc)
		})
	}
}
