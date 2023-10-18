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
	txGen := txtest.NewTransactionContextGenerator(errTest)

	entityTypeModels := []*model.EntityType{
		fixEntityTypeModel(ID1),
	}

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
				entityTypeSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, appID, *entityTypeInputs[0], uint64ResourceHash).Return(nil).Once()
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
				entityTypeSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *entityTypeInputs[0], uint64ResourceHash).Return(ID1, nil).Once()
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
				entityTypeSvc.On("Update", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, *entityTypeInputs[0], uint64ResourceHash).Return(nil).Once()
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
				entityTypeSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, *entityTypeInputs[0], uint64ResourceHash).Return(ID1, nil).Once()
				return entityTypeSvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedOutput:      entityTypeModels,
		},
		{
			Name: "Fails when listing entity types by applicaiton ID",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			EntityTypeSvcFn: func() *automock.EntityTypeService {
				entityTypeSvc := &automock.EntityTypeService{}
				entityTypeSvc.On("ListByApplicationID", txtest.CtxWithDBMatcher(), appID).Return(nil, errTest).Once()
				return entityTypeSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
		},
		{
			Name: "Fails when listing entity types by applicaiton template version ID",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			EntityTypeSvcFn: func() *automock.EntityTypeService {
				entityTypeSvc := &automock.EntityTypeService{}
				entityTypeSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, errTest).Once()
				return entityTypeSvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
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
				entityTypeSvc.On("Update", txtest.CtxWithDBMatcher(), resource.Application, appID, *entityTypeInputs[0], uint64ResourceHash).Return(errTest).Once()
				return entityTypeSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
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
				entityTypeSvc.On("Create", txtest.CtxWithDBMatcher(), resource.Application, appID, *entityTypeInputs[0], uint64ResourceHash).Return("", errTest).Once()
				return entityTypeSvc
			},
			InputResource:       resource.Application,
			InputResourceID:     appID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
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
				entityTypeSvc.On("ListByApplicationTemplateVersionID", txtest.CtxWithDBMatcher(), appTemplateVersionID).Return(nil, errTest).Once()
				entityTypeSvc.On("Create", txtest.CtxWithDBMatcher(), resource.ApplicationTemplateVersion, appTemplateVersionID, *entityTypeInputs[0], uint64ResourceHash).Return(ID1, nil).Once()
				return entityTypeSvc
			},
			InputResource:       resource.ApplicationTemplateVersion,
			InputResourceID:     appTemplateVersionID,
			InputEntityTypes:    entityTypeInputs,
			InputResourceHashes: resourceHashes,
			ExpectedErr:         errTest,
		},
	}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

			entityTypeSvc := test.EntityTypeSvcFn()

			entityTypeProcessor := processor.NewEntityTypeProcessor(tx, entityTypeSvc)
			result, err := entityTypeProcessor.Process(context.TODO(), test.InputResource, test.InputResourceID, test.InputPackagesFromDB, test.InputEntityTypes, test.InputResourceHashes)

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
