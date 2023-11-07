package processor_test

import (
	"context"
	"testing"

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
	integrationDependencyID     = "integration-dependency-id"
	integrationDependencyORDID  = "ord:integrationDependency"
	integrationDependencyORDID2 = "ord:integrationDependency2"
	aspectID                    = "aspect-id"
)

func TestIntegrationDependencyProcessor_Process(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	integrationDependencyModel := []*model.IntegrationDependency{
		fixIntegrationDependencyModel(integrationDependencyID, integrationDependencyORDID),
	}

	hashIntegrationDependency, _ := ord.HashObject(integrationDependencyModel)
	resourceHashes := map[string]uint64{
		integrationDependencyORDID: hashIntegrationDependency,
	}

	integrationDependencyInputs := []*model.IntegrationDependencyInput{
		fixIntegrationDependencyInputModel(integrationDependencyORDID),
	}

	integrationDependencyCreateInputs := []*model.IntegrationDependencyInput{
		fixIntegrationDependencyInputModel(integrationDependencyORDID2),
	}

	testCases := []struct {
		Name                         string
		TransactionerFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		IntegrationDependencySvcFn   func() *automock.IntegrationDependencyService
		AspectSvcFn                  func() *automock.AspectService
		InputResource                resource.Type
		InputResourceID              string
		InputIntegrationDependencies []*model.IntegrationDependencyInput
		InputResourceHashes          map[string]uint64
		ExpectedOutput               []*model.IntegrationDependency
		ExpectedErr                  error
	}{
		{
			Name: "Success for application resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(integrationDependencyModel, nil).Twice()
				integrationDependencySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyModel[0].ID, *integrationDependencyInputs[0], hashIntegrationDependency).Return(nil).Once()
				return integrationDependencySvc
			},
			AspectSvcFn: func() *automock.AspectService {
				aspectSvc := &automock.AspectService{}
				aspectSvc.On("DeleteByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(nil).Once()
				aspectSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *integrationDependencyInputs[0].Aspects[0]).Return(aspectID, nil).Once()
				return aspectSvc
			},
			InputResource:                resource.Application,
			InputResourceID:              appID,
			InputIntegrationDependencies: integrationDependencyInputs,
			InputResourceHashes:          resourceHashes,
			ExpectedOutput:               integrationDependencyModel,
		},
		{
			Name: "Success for application template version resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(integrationDependencyModel, nil).Twice()
				integrationDependencySvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, integrationDependencyModel[0].ID, *integrationDependencyInputs[0], hashIntegrationDependency).Return(nil).Once()
				return integrationDependencySvc
			},
			AspectSvcFn: func() *automock.AspectService {
				aspectSvc := &automock.AspectService{}
				aspectSvc.On("DeleteByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(nil).Once()
				aspectSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, integrationDependencyID, *integrationDependencyInputs[0].Aspects[0]).Return(aspectID, nil).Once()
				return aspectSvc
			},
			InputResource:                resource.ApplicationTemplateVersion,
			InputResourceID:              appTemplateVersionID,
			InputIntegrationDependencies: integrationDependencyInputs,
			InputResourceHashes:          resourceHashes,
			ExpectedOutput:               integrationDependencyModel,
		},
		{
			Name: "Success when creating integration dependency for application resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.IntegrationDependency{}, nil).Once()
				integrationDependencySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *integrationDependencyInputs[0], mock.Anything).Return(integrationDependencyID, nil).Once()
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(integrationDependencyModel, nil).Once()
				return integrationDependencySvc
			},
			AspectSvcFn: func() *automock.AspectService {
				aspectSvc := &automock.AspectService{}
				aspectSvc.On("DeleteByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(nil).Once()
				aspectSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *integrationDependencyInputs[0].Aspects[0]).Return(aspectID, nil).Once()
				return aspectSvc
			},
			InputResource:                resource.Application,
			InputResourceID:              appID,
			InputIntegrationDependencies: integrationDependencyInputs,
			InputResourceHashes:          resourceHashes,
			ExpectedOutput:               integrationDependencyModel,
		},
		{
			Name: "Fail on begin transaction while listing integration dependencies",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				return &automock.IntegrationDependencyService{}
			},
			AspectSvcFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			InputResource:                resource.Application,
			InputResourceID:              "",
			InputIntegrationDependencies: []*model.IntegrationDependencyInput{},
			InputResourceHashes:          resourceHashes,
			ExpectedErr:                  testErr,
		},
		{
			Name: "Fail while listing integration dependencies by application id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return integrationDependencySvc
			},
			AspectSvcFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			InputResource:                resource.Application,
			InputResourceID:              appID,
			InputIntegrationDependencies: []*model.IntegrationDependencyInput{},
			InputResourceHashes:          resourceHashes,
			ExpectedErr:                  testErr,
		},
		{
			Name: "Fail while listing integration dependencies by application template version id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, testErr).Once()
				return integrationDependencySvc
			},
			AspectSvcFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			InputResource:                resource.ApplicationTemplateVersion,
			InputResourceID:              appTemplateVersionID,
			InputIntegrationDependencies: []*model.IntegrationDependencyInput{},
			InputResourceHashes:          resourceHashes,
			ExpectedErr:                  testErr,
		},
		{
			Name: "Fail on begin transaction while re-syncing integration dependencies",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, testErr).Once()
				return persistTx, transact
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.IntegrationDependency{}, nil).Once()
				return integrationDependencySvc
			},
			AspectSvcFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			InputResource:                resource.Application,
			InputResourceID:              appID,
			InputIntegrationDependencies: integrationDependencyInputs,
			InputResourceHashes:          resourceHashes,
			ExpectedErr:                  testErr,
		},
		{
			Name: "Fail while updating integration dependency",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(integrationDependencyModel, nil).Once()
				integrationDependencySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyModel[0].ID, *integrationDependencyInputs[0], hashIntegrationDependency).Return(testErr).Once()
				return integrationDependencySvc
			},
			AspectSvcFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			InputResource:                resource.Application,
			InputResourceID:              appID,
			InputIntegrationDependencies: integrationDependencyInputs,
			InputResourceHashes:          resourceHashes,
			ExpectedErr:                  testErr,
		},
		{
			Name: "Fail while deleting integration dependency aspects",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(integrationDependencyModel, nil).Once()
				integrationDependencySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyModel[0].ID, *integrationDependencyInputs[0], hashIntegrationDependency).Return(nil).Once()
				return integrationDependencySvc
			},
			AspectSvcFn: func() *automock.AspectService {
				aspectSvc := &automock.AspectService{}
				aspectSvc.On("DeleteByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(testErr).Once()
				return aspectSvc
			},
			InputResource:                resource.Application,
			InputResourceID:              appID,
			InputIntegrationDependencies: integrationDependencyInputs,
			InputResourceHashes:          resourceHashes,
			ExpectedErr:                  testErr,
		},
		{
			Name: "Fail while creating integration dependency",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(integrationDependencyModel, nil).Once()
				integrationDependencySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *integrationDependencyCreateInputs[0], mock.Anything).Return("", testErr).Once()
				return integrationDependencySvc
			},
			AspectSvcFn: func() *automock.AspectService {
				return &automock.AspectService{}
			},
			InputResource:                resource.Application,
			InputResourceID:              appID,
			InputIntegrationDependencies: integrationDependencyCreateInputs,
			InputResourceHashes:          resourceHashes,
			ExpectedErr:                  testErr,
		},
		{
			Name: "Fail while creating integration dependency aspects for application resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return([]*model.IntegrationDependency{}, nil).Once()
				integrationDependencySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, str.Ptr(packageID), *integrationDependencyCreateInputs[0], mock.Anything).Return(integrationDependencyID, nil).Once()
				return integrationDependencySvc
			},
			AspectSvcFn: func() *automock.AspectService {
				aspectSvc := &automock.AspectService{}
				aspectSvc.On("DeleteByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(nil).Once()
				aspectSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *integrationDependencyCreateInputs[0].Aspects[0]).Return("", testErr).Once()
				return aspectSvc
			},
			InputResource:                resource.Application,
			InputResourceID:              appID,
			InputIntegrationDependencies: integrationDependencyCreateInputs,
			InputResourceHashes:          resourceHashes,
			ExpectedErr:                  testErr,
		},
		{
			Name: "Fail while listing resources after update",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsTwice()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(integrationDependencyModel, nil).Once()
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()

				integrationDependencySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyModel[0].ID, *integrationDependencyInputs[0], hashIntegrationDependency).Return(nil).Once()
				return integrationDependencySvc
			},
			AspectSvcFn: func() *automock.AspectService {
				aspectSvc := &automock.AspectService{}
				aspectSvc.On("DeleteByIntegrationDependencyID", txtest.CtxWithDBMatcher(), integrationDependencyID).Return(nil).Once()
				aspectSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, integrationDependencyID, *integrationDependencyInputs[0].Aspects[0]).Return(aspectID, nil).Once()
				return aspectSvc
			},
			InputResource:                resource.Application,
			InputResourceID:              appID,
			InputIntegrationDependencies: integrationDependencyInputs,
			InputResourceHashes:          resourceHashes,
			ExpectedErr:                  testErr,
		},
	}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()
			integrationDependencySvc := test.IntegrationDependencySvcFn()
			aspectSvc := test.AspectSvcFn()

			integrationDependencyProcessor := processor.NewIntegrationDependencyProcessor(tx, integrationDependencySvc, aspectSvc)
			result, err := integrationDependencyProcessor.Process(context.TODO(), test.InputResource, test.InputResourceID, fixPackages(), test.InputIntegrationDependencies, test.InputResourceHashes)
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.ExpectedOutput, result)
			}

			mock.AssertExpectationsForObjects(t, tx, integrationDependencySvc)
		})
	}
}
