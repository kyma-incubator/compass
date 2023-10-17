package processor_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTombstoneProcessor_Process(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	applicationID := "application-id"
	applicationTemplateVersionID := "application-template-version-id"

	tombstonesModel := []*model.Tombstone{
		{
			ID:                           "tombstone-id",
			OrdID:                        "ord:tombstone",
			ApplicationID:                &applicationID,
			ApplicationTemplateVersionID: &applicationTemplateVersionID,
		},
	}

	tombstoneInputs := []*model.TombstoneInput{
		{
			OrdID: "ord:tombstone",
		},
	}

	tombstoneCreateInputs := []*model.TombstoneInput{
		{
			OrdID: "ord:tombstone2",
		},
	}

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		TombstoneSvcFn  func() *automock.TombstoneService
		InputResource   resource.Type
		InputResourceID string
		InputTombstones []*model.TombstoneInput
		ExpectedOutput  []*model.Tombstone
		ExpectedErr     error
	}{
		{
			Name: "Success for application resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			TombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(tombstonesModel, nil).Twice()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, tombstonesModel[0].ID, *tombstoneInputs[0]).Return(nil).Once()
				return tombstoneSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputTombstones: tombstoneInputs,
			ExpectedOutput:  tombstonesModel,
		},
		{
			Name: "Success for application template version resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			TombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), applicationTemplateVersionID).Return(tombstonesModel, nil).Twice()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, tombstonesModel[0].ID, *tombstoneInputs[0]).Return(nil).Once()
				return tombstoneSvc
			},
			InputResource:   resource.ApplicationTemplateVersion,
			InputResourceID: applicationTemplateVersionID,
			InputTombstones: tombstoneInputs,
			ExpectedOutput:  tombstonesModel,
		},
		{
			Name: "Fail on begin transaction while listing tombstones",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			TombstoneSvcFn: func() *automock.TombstoneService {
				return &automock.TombstoneService{}
			},
			InputResource:   resource.Application,
			InputResourceID: "",
			InputTombstones: []*model.TombstoneInput{},
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while listing tombstones by application id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			TombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(nil, testErr).Once()
				return tombstoneSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputTombstones: []*model.TombstoneInput{},
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while listing tombstones by application template version id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			TombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), applicationTemplateVersionID).Return(nil, testErr).Once()
				return tombstoneSvc
			},
			InputResource:   resource.ApplicationTemplateVersion,
			InputResourceID: applicationTemplateVersionID,
			InputTombstones: []*model.TombstoneInput{},
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail on begin transaction while re-syncing tombstones",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, testErr).Once()
				return persistTx, transact
			},
			TombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return([]*model.Tombstone{}, nil).Once()
				return tombstoneSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputTombstones: tombstoneInputs,
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while updating tombstone",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			TombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(tombstonesModel, nil).Once()
				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, tombstonesModel[0].ID, *tombstoneInputs[0]).Return(testErr).Once()
				return tombstoneSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputTombstones: tombstoneInputs,
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while creating tombstone",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			TombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(tombstonesModel, nil).Once()
				tombstoneSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, applicationID, *tombstoneCreateInputs[0]).Return("", testErr).Once()
				return tombstoneSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputTombstones: tombstoneCreateInputs,
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
			TombstoneSvcFn: func() *automock.TombstoneService {
				tombstoneSvc := &automock.TombstoneService{}
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(tombstonesModel, nil).Once()
				tombstoneSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(nil, testErr).Once()

				tombstoneSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, tombstonesModel[0].ID, *tombstoneInputs[0]).Return(nil).Once()
				return tombstoneSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputTombstones: tombstoneInputs,
			ExpectedErr:     testErr,
		},
	}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

			tombstoneSvc := test.TombstoneSvcFn()

			tombstoneProcessor := processor.NewTombstoneProcessor(tx, tombstoneSvc)
			result, err := tombstoneProcessor.Process(context.TODO(), test.InputResource, test.InputResourceID, test.InputTombstones)
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.ExpectedOutput, result)
			}

			mock.AssertExpectationsForObjects(t, tx, tombstoneSvc)
		})
	}
}
