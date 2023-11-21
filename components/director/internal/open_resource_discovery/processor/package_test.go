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

func TestPackageProcessor_Process(t *testing.T) {
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	applicationID := "application-id"
	applicationTemplateVersionID := "application-template-version-id"

	packageModels := []*model.Package{
		{
			ID:                           "package-id",
			OrdID:                        "ord:package",
			ApplicationID:                &applicationID,
			ApplicationTemplateVersionID: &applicationTemplateVersionID,
		},
	}

	packageInputs := []*model.PackageInput{
		{
			OrdID: "ord:package",
		},
	}

	packageCreateInputs := []*model.PackageInput{
		{
			OrdID: "ord:package2",
		},
	}

	packageHashes := map[string]uint64{
		"ord:package":  1,
		"ord:package2": 2,
	}

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		PackageSvcFn    func() *automock.PackageService
		InputResource   resource.Type
		InputResourceID string
		InputPackages   []*model.PackageInput
		ExpectedOutput  []*model.Package
		ExpectedErr     error
	}{
		{
			Name: "Success for application resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			PackageSvcFn: func() *automock.PackageService {
				packageSvc := &automock.PackageService{}
				packageSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(packageModels, nil).Twice()
				packageSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, packageModels[0].ID, *packageInputs[0], packageHashes[packageInputs[0].OrdID]).Return(nil).Once()
				return packageSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputPackages:   packageInputs,
			ExpectedOutput:  packageModels,
		},
		{
			Name: "Success for application template version resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			PackageSvcFn: func() *automock.PackageService {
				packageSvc := &automock.PackageService{}
				packageSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), applicationTemplateVersionID).Return(packageModels, nil).Twice()
				packageSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, packageModels[0].ID, *packageInputs[0], packageHashes[packageInputs[0].OrdID]).Return(nil).Once()
				return packageSvc
			},
			InputResource:   resource.ApplicationTemplateVersion,
			InputResourceID: applicationTemplateVersionID,
			InputPackages:   packageInputs,
			ExpectedOutput:  packageModels,
		},
		{
			Name: "Fail on begin transaction while listing packages",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			PackageSvcFn: func() *automock.PackageService {
				return &automock.PackageService{}
			},
			InputResource:   resource.Application,
			InputResourceID: "",
			InputPackages:   []*model.PackageInput{},
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while listing packages by application id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			PackageSvcFn: func() *automock.PackageService {
				packageSvc := &automock.PackageService{}
				packageSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(nil, testErr).Once()
				return packageSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputPackages:   []*model.PackageInput{},
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while listing packages by application template version id",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			PackageSvcFn: func() *automock.PackageService {
				packageSvc := &automock.PackageService{}
				packageSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), applicationTemplateVersionID).Return(nil, testErr).Once()
				return packageSvc
			},
			InputResource:   resource.ApplicationTemplateVersion,
			InputResourceID: applicationTemplateVersionID,
			InputPackages:   []*model.PackageInput{},
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail on begin transaction while re-syncing packages",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, testErr).Once()
				return persistTx, transact
			},
			PackageSvcFn: func() *automock.PackageService {
				packageSvc := &automock.PackageService{}
				packageSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return([]*model.Package{}, nil).Once()
				return packageSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputPackages:   packageInputs,
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while updating package",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			PackageSvcFn: func() *automock.PackageService {
				packageSvc := &automock.PackageService{}
				packageSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(packageModels, nil).Once()
				packageSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, packageModels[0].ID, *packageInputs[0], packageHashes[packageInputs[0].OrdID]).Return(testErr).Once()
				return packageSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputPackages:   packageInputs,
			ExpectedErr:     testErr,
		},
		{
			Name: "Fail while creating package",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			PackageSvcFn: func() *automock.PackageService {
				packageSvc := &automock.PackageService{}
				packageSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(packageModels, nil).Once()
				packageSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, applicationID, *packageCreateInputs[0], packageHashes[packageCreateInputs[0].OrdID]).Return("", testErr).Once()
				return packageSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputPackages:   packageCreateInputs,
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
			PackageSvcFn: func() *automock.PackageService {
				packageSvc := &automock.PackageService{}
				packageSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(packageModels, nil).Once()
				packageSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), applicationID).Return(nil, testErr).Once()

				packageSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, packageModels[0].ID, *packageInputs[0], packageHashes[packageInputs[0].OrdID]).Return(nil).Once()
				return packageSvc
			},
			InputResource:   resource.Application,
			InputResourceID: applicationID,
			InputPackages:   packageInputs,
			ExpectedErr:     testErr,
		},
	}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

			packageSvc := test.PackageSvcFn()

			packageProcessor := processor.NewPackageProcessor(tx, packageSvc)
			result, err := packageProcessor.Process(context.TODO(), test.InputResource, test.InputResourceID, test.InputPackages, packageHashes)
			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.ExpectedOutput, result)
			}

			mock.AssertExpectationsForObjects(t, tx, packageSvc)
		})
	}
}
