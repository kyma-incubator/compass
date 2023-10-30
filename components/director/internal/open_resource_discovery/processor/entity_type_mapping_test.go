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

func TestEntityTypeMappingProcessor_Process(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(errTest)

	entityTypeMappingModels := []*model.EntityTypeMapping{
		fixEntityTypeMappingModel(ID1),
	}

	emptyEntityTypeMappingModels := []*model.EntityTypeMapping{}

	entityTypeMappingInputs := []*model.EntityTypeMappingInput{
		fixEntityTypeMappingInputModel(),
	}

	testCases := []struct {
		Name                    string
		TransactionerFn         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		EntityTypeMappingSvcFn  func() *automock.EntityTypeMappingService
		InputResource           resource.Type
		InputResourceID         string
		InputEntityTypeMappings []*model.EntityTypeMappingInput
		ExpectedOutput          []*model.EntityTypeMapping
		ExpectedErr             error
	}{
		{
			Name: "Success for API resource when resource exists",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByAPIDefinitionID", txtest.CtxWithDBMatcher(), apiID).Return(entityTypeMappingModels, nil).Twice()
				entityTypeMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.API, ID1).Return(nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.API, apiID, *entityTypeMappingInputs[0]).Return(mock.Anything, nil).Once()
				return entityTypeMappingSvc
			},
			InputResource:           resource.API,
			InputResourceID:         apiID,
			InputEntityTypeMappings: entityTypeMappingInputs,
			ExpectedOutput:          entityTypeMappingModels,
		},
		{
			Name: "Success for API resource when resource does not exists",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByAPIDefinitionID", txtest.CtxWithDBMatcher(), apiID).Return(emptyEntityTypeMappingModels, nil).Once()
				entityTypeMappingSvc.On("ListByAPIDefinitionID", txtest.CtxWithDBMatcher(), apiID).Return(entityTypeMappingModels, nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.API, apiID, *entityTypeMappingInputs[0]).Return(mock.Anything, nil).Once()
				return entityTypeMappingSvc
			},
			InputResource:           resource.API,
			InputResourceID:         apiID,
			InputEntityTypeMappings: entityTypeMappingInputs,
			ExpectedOutput:          entityTypeMappingModels,
		},
		{
			Name: "Success for Event resource when resource exists",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByEventDefinitionID", txtest.CtxWithDBMatcher(), eventID).Return(entityTypeMappingModels, nil).Twice()
				entityTypeMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.EventDefinition, ID1).Return(nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.EventDefinition, eventID, *entityTypeMappingInputs[0]).Return(mock.Anything, nil).Once()
				return entityTypeMappingSvc
			},
			InputResource:           resource.EventDefinition,
			InputResourceID:         eventID,
			InputEntityTypeMappings: entityTypeMappingInputs,
			ExpectedOutput:          entityTypeMappingModels,
		},
		{
			Name: "Success for Event resource when resource does not exists",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByEventDefinitionID", txtest.CtxWithDBMatcher(), eventID).Return(emptyEntityTypeMappingModels, nil).Once()
				entityTypeMappingSvc.On("ListByEventDefinitionID", txtest.CtxWithDBMatcher(), eventID).Return(entityTypeMappingModels, nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.EventDefinition, eventID, *entityTypeMappingInputs[0]).Return(mock.Anything, nil).Once()
				return entityTypeMappingSvc
			},
			InputResource:           resource.EventDefinition,
			InputResourceID:         eventID,
			InputEntityTypeMappings: entityTypeMappingInputs,
			ExpectedOutput:          entityTypeMappingModels,
		},
		{
			Name: "Fails when listing entity type mappings by API ID",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByAPIDefinitionID", txtest.CtxWithDBMatcher(), apiID).Return(nil, errTest).Once()
				return entityTypeMappingSvc
			},
			InputResource:           resource.API,
			InputResourceID:         apiID,
			InputEntityTypeMappings: entityTypeMappingInputs,
			ExpectedErr:             errTest,
		},
		{
			Name: "Fails when listing entity type mappings by Event ID",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByEventDefinitionID", txtest.CtxWithDBMatcher(), eventID).Return(nil, errTest).Once()
				return entityTypeMappingSvc
			},
			InputResource:           resource.EventDefinition,
			InputResourceID:         eventID,
			InputEntityTypeMappings: entityTypeMappingInputs,
			ExpectedErr:             errTest,
		},
		{
			Name: "Fails for API resource when resource deletion fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByAPIDefinitionID", txtest.CtxWithDBMatcher(), apiID).Return(entityTypeMappingModels, nil).Once()
				entityTypeMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.API, ID1).Return(errTest).Once()
				return entityTypeMappingSvc
			},
			InputResource:           resource.API,
			InputResourceID:         apiID,
			InputEntityTypeMappings: entityTypeMappingInputs,
			ExpectedErr:             errTest,
		},
		{
			Name: "Fails for Event resource when resource deletion fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByEventDefinitionID", txtest.CtxWithDBMatcher(), eventID).Return(entityTypeMappingModels, nil).Once()
				entityTypeMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.EventDefinition, ID1).Return(errTest).Once()
				return entityTypeMappingSvc
			},
			InputResource:           resource.EventDefinition,
			InputResourceID:         eventID,
			InputEntityTypeMappings: entityTypeMappingInputs,
			ExpectedErr:             errTest,
		},
		{
			Name: "Fails for API resource when resource creation fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByAPIDefinitionID", txtest.CtxWithDBMatcher(), apiID).Return(entityTypeMappingModels, nil).Once()
				entityTypeMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.API, ID1).Return(nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.API, apiID, *entityTypeMappingInputs[0]).Return("", errTest).Once()
				return entityTypeMappingSvc
			},
			InputResource:           resource.API,
			InputResourceID:         apiID,
			InputEntityTypeMappings: entityTypeMappingInputs,
			ExpectedErr:             errTest,
		},
		{
			Name: "Fails for Event resource when resource creation fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByEventDefinitionID", txtest.CtxWithDBMatcher(), eventID).Return(entityTypeMappingModels, nil).Once()
				entityTypeMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.EventDefinition, ID1).Return(nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.EventDefinition, eventID, *entityTypeMappingInputs[0]).Return("", errTest).Once()
				return entityTypeMappingSvc
			},
			InputResource:           resource.EventDefinition,
			InputResourceID:         eventID,
			InputEntityTypeMappings: entityTypeMappingInputs,
			ExpectedErr:             errTest,
		},
		{
			Name: "Fails for API resource when second listing fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsTwice()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByAPIDefinitionID", txtest.CtxWithDBMatcher(), apiID).Return(entityTypeMappingModels, nil).Once()
				entityTypeMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.API, ID1).Return(nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.API, apiID, *entityTypeMappingInputs[0]).Return(mock.Anything, nil).Once()
				entityTypeMappingSvc.On("ListByAPIDefinitionID", txtest.CtxWithDBMatcher(), apiID).Return(nil, errTest).Once()
				return entityTypeMappingSvc
			},
			InputResource:           resource.API,
			InputResourceID:         apiID,
			InputEntityTypeMappings: entityTypeMappingInputs,
			ExpectedErr:             errTest,
		},
		{
			Name: "Fails for Event resource when second listing fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsTwice()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			EntityTypeMappingSvcFn: func() *automock.EntityTypeMappingService {
				entityTypeMappingSvc := &automock.EntityTypeMappingService{}
				entityTypeMappingSvc.On("ListByEventDefinitionID", txtest.CtxWithDBMatcher(), eventID).Return(entityTypeMappingModels, nil).Once()
				entityTypeMappingSvc.On("Delete", txtest.CtxWithDBMatcher(), resource.EventDefinition, ID1).Return(nil).Once()
				entityTypeMappingSvc.On("Create", txtest.CtxWithDBMatcher(), resource.EventDefinition, eventID, *entityTypeMappingInputs[0]).Return(mock.Anything, nil).Once()
				entityTypeMappingSvc.On("ListByEventDefinitionID", txtest.CtxWithDBMatcher(), eventID).Return(nil, errTest).Once()
				return entityTypeMappingSvc
			},
			InputResource:           resource.EventDefinition,
			InputResourceID:         eventID,
			InputEntityTypeMappings: entityTypeMappingInputs,
			ExpectedErr:             errTest,
		},
	}
	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			_, tx := test.TransactionerFn()

			entityTypeMappingSvc := test.EntityTypeMappingSvcFn()

			entityTypeMappingProcessor := processor.NewEntityTypeMappingProcessor(tx, entityTypeMappingSvc)
			result, err := entityTypeMappingProcessor.Process(context.TODO(), test.InputResource, test.InputResourceID, test.InputEntityTypeMappings)

			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.ExpectedOutput, result)
			}

			mock.AssertExpectationsForObjects(t, tx, entityTypeMappingSvc)
		})
	}
}
