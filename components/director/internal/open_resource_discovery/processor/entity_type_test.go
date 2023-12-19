package processor_test

import (
	"context"
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

func TestEntityTypeProcessor_Process(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)

	entityTypeModels := []*model.EntityType{
		fixEntityTypeModel(entityTypeID),
	}

	updatedEntityTypeModels := []*model.EntityType{
		fixEntityTypeModel(entityTypeID),
	}

	updatedEntityTypeModels[0].PackageID = packageID2

	emptyEntityTypeModels := []*model.EntityType{}

	entityTypeInputs := []*model.EntityTypeInput{
		fixEntityTypeInputModel(),
	}

	resourceHashes := map[string]uint64{ordID: uint64ResourceHash}

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		EntityTypeSvcFn     func() *automock.EntityTypeService
		InputResource       resource.Type
		InputResourceID     string
		InputEntityTypes    []*model.EntityTypeInput
		InputPackagesFromDB []*model.Package
		InputResourceHashes map[string]uint64
		ExpectedOutput      []*model.EntityType
		ExpectedErr         error
	}{
		{
			Name: "Success for application resource when resource exists",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			EntityTypeSvcFn: func() *automock.EntityTypeService {
				entityTypeSvc := &automock.EntityTypeService{}
				entityTypeSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(entityTypeModels, nil).Twice()
				entityTypeSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, entityTypeModels[0].ID, packageID1, *entityTypeInputs[0], uint64ResourceHash).Return(nil).Once()
				return entityTypeSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedOutput:      entityTypeModels,
		},
		{
			Name: "Success for application resource when resource does not exists",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			EntityTypeSvcFn: func() *automock.EntityTypeService {
				entityTypeSvc := &automock.EntityTypeService{}
				entityTypeSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(emptyEntityTypeModels, nil).Once()
				entityTypeSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(entityTypeModels, nil).Once()
				entityTypeSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, packageID1, *entityTypeInputs[0], uint64ResourceHash).Return(entityTypeID, nil).Once()
				return entityTypeSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedOutput:      entityTypeModels,
		},
		{
			Name: "Success for application template version resource when resource exists",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			EntityTypeSvcFn: func() *automock.EntityTypeService {
				entityTypeSvc := &automock.EntityTypeService{}
				entityTypeSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(entityTypeModels, nil).Twice()
				entityTypeSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, entityTypeModels[0].ID, packageID1, *entityTypeInputs[0], uint64ResourceHash).Return(nil).Once()
				return entityTypeSvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedOutput:      entityTypeModels,
		},
		{
			Name: "Success for application template version resource when resource does not exists",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			EntityTypeSvcFn: func() *automock.EntityTypeService {
				entityTypeSvc := &automock.EntityTypeService{}
				entityTypeSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(emptyEntityTypeModels, nil).Once()
				entityTypeSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(entityTypeModels, nil).Once()
				entityTypeSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, packageID1, *entityTypeInputs[0], uint64ResourceHash).Return(entityTypeID, nil).Once()
				return entityTypeSvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedOutput:      entityTypeModels,
		},
		{
			Name: "Fails when listing entity types by application ID",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			EntityTypeSvcFn: func() *automock.EntityTypeService {
				entityTypeSvc := &automock.EntityTypeService{}
				entityTypeSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, testErr).Once()
				return entityTypeSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fails when listing entity types by application template version ID",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			EntityTypeSvcFn: func() *automock.EntityTypeService {
				entityTypeSvc := &automock.EntityTypeService{}
				entityTypeSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, testErr).Once()
				return entityTypeSvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fails when update entity type for application resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			EntityTypeSvcFn: func() *automock.EntityTypeService {
				entityTypeSvc := &automock.EntityTypeService{}
				entityTypeSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(entityTypeModels, nil).Once()
				entityTypeSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, entityTypeModels[0].ID, packageID1, *entityTypeInputs[0], uint64ResourceHash).Return(testErr).Once()
				return entityTypeSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fails when create entity type for application resource",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			EntityTypeSvcFn: func() *automock.EntityTypeService {
				entityTypeSvc := &automock.EntityTypeService{}
				entityTypeSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(emptyEntityTypeModels, nil).Once()
				entityTypeSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, packageID1, *entityTypeInputs[0], uint64ResourceHash).Return("", testErr).Once()
				return entityTypeSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Fail when second list of entity types fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsTwice()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			EntityTypeSvcFn: func() *automock.EntityTypeService {
				entityTypeSvc := &automock.EntityTypeService{}
				entityTypeSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(emptyEntityTypeModels, nil).Once()
				entityTypeSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, testErr).Once()
				entityTypeSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, packageID1, *entityTypeInputs[0], uint64ResourceHash).Return(entityTypeID, nil).Once()
				return entityTypeSvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         testErr,
		},
		{
			Name: "Success when updating package id for Entity Type",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			EntityTypeSvcFn: func() *automock.EntityTypeService {
				entityTypeSvc := &automock.EntityTypeService{}
				entityTypeInputs[0].OrdPackageID = packageORDID2
				entityTypeSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(entityTypeModels, nil).Once()
				entityTypeSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(updatedEntityTypeModels, nil).Once()
				entityTypeSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, entityTypeModels[0].ID, packageID2, *entityTypeInputs[0], uint64ResourceHash).Return(nil).Once()
				return entityTypeSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedOutput:      updatedEntityTypeModels,
		},
	}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

			entityTypeSvc := test.EntityTypeSvcFn()

			entityTypeProcessor := processor.NewEntityTypeProcessor(tx, entityTypeSvc)
			result, err := entityTypeProcessor.Process(context.TODO(), test.InputResource, test.InputResourceID, test.InputEntityTypes, fixPackages(), test.InputResourceHashes)

			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.ExpectedOutput, result)
			}

			mock.AssertExpectationsForObjects(t, tx, entityTypeSvc)
		})
	}
}
