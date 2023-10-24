package processor_test

import (
	"context"
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
	"testing"
)

const (
	integrationDependencyID    = "integration-dependency-id"
	integrationDependencyID2   = "integration-dependency-id2"
	integrationDependencyORDID = "ord:integrationDependency"
	packageORDID               = "ns:package:PACKAGE_ID:v1"
	packageID                  = "package-id"
)

func TestIntegrationDependencyProcessor_Process(t *testing.T) {
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	applicationID := "application-id"
	applicationTemplateVersionID := "application-template-version-id"

	integrationDependencyModel := []*model.IntegrationDependency{
		{
			BaseEntity: &model.BaseEntity{
				ID:    integrationDependencyID,
				Ready: true,
			},
			OrdID:                        str.Ptr(integrationDependencyORDID),
			ApplicationID:                &applicationID,
			ApplicationTemplateVersionID: &applicationTemplateVersionID,
			PackageID:                    str.Ptr(packageORDID),
		},
	}

	hashIntegrationDependency, _ := ord.HashObject(integrationDependencyModel)

	integrationDependencyInputs := []*model.IntegrationDependencyInput{
		{
			OrdID:        str.Ptr(integrationDependencyORDID),
			OrdPackageID: str.Ptr(packageORDID),
		},
	}

	integrationDependencyCreateInputs := []*model.IntegrationDependencyInput{
		{
			OrdID:        str.Ptr(integrationDependencyID2),
			OrdPackageID: str.Ptr(packageORDID),
		},
	}

	testCases := []struct {
		Name                         string
		TransactionerFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		IntegrationDependencySvcFn   func() *automock.IntegrationDependencyService
		InputResource                resource.Type
		InputResourceID              string
		InputIntegrationDependencies []*model.IntegrationDependencyInput
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
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(integrationDependencyModel, nil).Twice()
				integrationDependencySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, applicationID, integrationDependencyModel[0].ID, *integrationDependencyInputs[0], hashIntegrationDependency).Return(nil).Once()
				return integrationDependencySvc
			},
			InputResource:                resource.Application,
			InputResourceID:              applicationID,
			InputIntegrationDependencies: integrationDependencyInputs,
			ExpectedOutput:               integrationDependencyModel,
		},
		{
			Name: "Success for application template version resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), applicationTemplateVersionID).Return(integrationDependencyModel, nil).Twice()
				integrationDependencySvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, applicationTemplateVersionID, integrationDependencyModel[0].ID, *integrationDependencyInputs[0], hashIntegrationDependency).Return(nil).Once()
				return integrationDependencySvc
			},
			InputResource:                resource.ApplicationTemplateVersion,
			InputResourceID:              applicationTemplateVersionID,
			InputIntegrationDependencies: integrationDependencyInputs,
			ExpectedOutput:               integrationDependencyModel,
		},
		{
			Name: "Success when creating integration dependency for application resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return([]*model.IntegrationDependency{}, nil).Once()
				integrationDependencySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, applicationID, str.Ptr(packageID), *integrationDependencyCreateInputs[0], mock.Anything).Return(integrationDependencyID, nil).Once()
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(integrationDependencyModel, nil).Once()
				return integrationDependencySvc
			},
			InputResource:                resource.Application,
			InputResourceID:              applicationID,
			InputIntegrationDependencies: integrationDependencyCreateInputs,
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
			InputResource:                resource.Application,
			InputResourceID:              "",
			InputIntegrationDependencies: []*model.IntegrationDependencyInput{},
			ExpectedErr:                  testErr,
		},
		{
			Name: "Fail while listing integration dependencies by application id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(nil, testErr).Once()
				return integrationDependencySvc
			},
			InputResource:                resource.Application,
			InputResourceID:              applicationID,
			InputIntegrationDependencies: []*model.IntegrationDependencyInput{},
			ExpectedErr:                  testErr,
		},
		{
			Name: "Fail while listing integration dependencies by application template version id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			IntegrationDependencySvcFn: func() *automock.IntegrationDependencyService {
				integrationDependencySvc := &automock.IntegrationDependencyService{}
				integrationDependencySvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), applicationTemplateVersionID).Return(nil, testErr).Once()
				return integrationDependencySvc
			},
			InputResource:                resource.ApplicationTemplateVersion,
			InputResourceID:              applicationTemplateVersionID,
			InputIntegrationDependencies: []*model.IntegrationDependencyInput{},
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
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return([]*model.IntegrationDependency{}, nil).Once()
				return integrationDependencySvc
			},
			InputResource:                resource.Application,
			InputResourceID:              applicationID,
			InputIntegrationDependencies: integrationDependencyInputs,
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
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(integrationDependencyModel, nil).Once()
				integrationDependencySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, applicationID, integrationDependencyModel[0].ID, *integrationDependencyInputs[0], hashIntegrationDependency).Return(testErr).Once()
				return integrationDependencySvc
			},
			InputResource:                resource.Application,
			InputResourceID:              applicationID,
			InputIntegrationDependencies: integrationDependencyInputs,
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
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(integrationDependencyModel, nil).Once()
				integrationDependencySvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, applicationID, str.Ptr(packageID), *integrationDependencyCreateInputs[0], mock.Anything).Return("", testErr).Once()
				return integrationDependencySvc
			},
			InputResource:                resource.Application,
			InputResourceID:              applicationID,
			InputIntegrationDependencies: integrationDependencyCreateInputs,
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
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(integrationDependencyModel, nil).Once()
				integrationDependencySvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(nil, testErr).Once()

				integrationDependencySvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, applicationID, integrationDependencyModel[0].ID, *integrationDependencyInputs[0], hashIntegrationDependency).Return(nil).Once()
				return integrationDependencySvc
			},
			InputResource:                resource.Application,
			InputResourceID:              applicationID,
			InputIntegrationDependencies: integrationDependencyInputs,
			ExpectedErr:                  testErr,
		},
	}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()
			integrationDependencySvc := test.IntegrationDependencySvcFn()

			resourceHashes := map[string]uint64{
				integrationDependencyORDID: hashIntegrationDependency,
			}
			packagesFromDB := []*model.Package{
				{
					ID:    packageID,
					OrdID: packageORDID},
			}

			integrationDependencyProcessor := processor.NewIntegrationDependencyProcessor(tx, integrationDependencySvc)
			result, err := integrationDependencyProcessor.Process(context.TODO(), test.InputResource, test.InputResourceID, packagesFromDB, test.InputIntegrationDependencies, resourceHashes)
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
