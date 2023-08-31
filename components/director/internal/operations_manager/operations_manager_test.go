package operationsmanager

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/operations_manager/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOperationsManager_MarkOperationCompleted(t *testing.T) {
	// GIVEN
	testError := errors.New("test error")
	operationID := "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	txGen := txtest.NewTransactionContextGenerator(testError)
	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		OperationSvcFn func() *automock.OperationService
		Input          string
		ExpectedError  error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			OperationSvcFn: func() *automock.OperationService {
				operationSvc := &automock.OperationService{}
				operationSvc.On("MarkAsCompleted", txtest.CtxWithDBMatcher(), operationID).Return(nil).Once()
				return operationSvc
			},
			Input: operationID,
		},
		{
			Name: "Error while marking as completed",
			TxFn: txGen.ThatDoesntExpectCommit,
			OperationSvcFn: func() *automock.OperationService {
				operationSvc := &automock.OperationService{}
				operationSvc.On("MarkAsCompleted", txtest.CtxWithDBMatcher(), operationID).Return(testError).Once()
				return operationSvc
			},
			Input:         operationID,
			ExpectedError: testError,
		},
		{
			Name: "Error while beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			OperationSvcFn: func() *automock.OperationService {
				return &automock.OperationService{}
			},
			Input:         operationID,
			ExpectedError: testError,
		},
		{
			Name: "Error while committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			OperationSvcFn: func() *automock.OperationService {
				operationSvc := &automock.OperationService{}
				operationSvc.On("MarkAsCompleted", txtest.CtxWithDBMatcher(), operationID).Return(nil).Once()
				return operationSvc
			},
			Input:         operationID,
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			operationSvc := testCase.OperationSvcFn()

			opManager := NewOperationsManager(transact, operationSvc, model.OperationTypeOrdAggregation, OperationsManagerConfig{})

			// WHEN
			err := opManager.MarkOperationCompleted(context.TODO(), testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, operationSvc)
		})
	}
}

func TestOperationsManager_MarkOperationFailed(t *testing.T) {
	// GIVEN
	testError := errors.New("test error")
	operationID := "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	txGen := txtest.NewTransactionContextGenerator(testError)
	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		OperationSvcFn func() *automock.OperationService
		Input          string
		ExpectedError  error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			OperationSvcFn: func() *automock.OperationService {
				operationSvc := &automock.OperationService{}
				operationSvc.On("MarkAsFailed", txtest.CtxWithDBMatcher(), operationID, testError.Error()).Return(nil).Once()
				return operationSvc
			},
			Input: operationID,
		},
		{
			Name: "Error while marking as failed",
			TxFn: txGen.ThatDoesntExpectCommit,
			OperationSvcFn: func() *automock.OperationService {
				operationSvc := &automock.OperationService{}
				operationSvc.On("MarkAsFailed", txtest.CtxWithDBMatcher(), operationID, testError.Error()).Return(testError).Once()
				return operationSvc
			},
			Input:         operationID,
			ExpectedError: testError,
		},
		{
			Name: "Error while beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			OperationSvcFn: func() *automock.OperationService {
				return &automock.OperationService{}
			},
			Input:         operationID,
			ExpectedError: testError,
		},
		{
			Name: "Error while committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			OperationSvcFn: func() *automock.OperationService {
				operationSvc := &automock.OperationService{}
				operationSvc.On("MarkAsFailed", txtest.CtxWithDBMatcher(), operationID, testError.Error()).Return(nil).Once()
				return operationSvc
			},
			Input:         operationID,
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			operationSvc := testCase.OperationSvcFn()

			opManager := NewOperationsManager(transact, operationSvc, model.OperationTypeOrdAggregation, OperationsManagerConfig{})

			// WHEN
			err := opManager.MarkOperationFailed(context.TODO(), testCase.Input, testError.Error())

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, operationSvc)
		})
	}
}

func TestOperationsManager_RescheduleOperation(t *testing.T) {
	// GIVEN
	testError := errors.New("test error")
	operationID := "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	txGen := txtest.NewTransactionContextGenerator(testError)
	testCases := []struct {
		Name           string
		TxFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		OperationSvcFn func() *automock.OperationService
		Input          string
		ExpectedError  error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			OperationSvcFn: func() *automock.OperationService {
				operationSvc := &automock.OperationService{}
				operationSvc.On("RescheduleOperation", txtest.CtxWithDBMatcher(), operationID, int(HighOperationPriority)).Return(nil).Once()
				return operationSvc
			},
			Input: operationID,
		},
		{
			Name: "Error while rescheduling operation",
			TxFn: txGen.ThatDoesntExpectCommit,
			OperationSvcFn: func() *automock.OperationService {
				operationSvc := &automock.OperationService{}
				operationSvc.On("RescheduleOperation", txtest.CtxWithDBMatcher(), operationID, int(HighOperationPriority)).Return(testError).Once()
				return operationSvc
			},
			Input:         operationID,
			ExpectedError: testError,
		},
		{
			Name: "Error while beginning transaction",
			TxFn: txGen.ThatFailsOnBegin,
			OperationSvcFn: func() *automock.OperationService {
				return &automock.OperationService{}
			},
			Input:         operationID,
			ExpectedError: testError,
		},
		{
			Name: "Error while committing transaction",
			TxFn: txGen.ThatFailsOnCommit,
			OperationSvcFn: func() *automock.OperationService {
				operationSvc := &automock.OperationService{}
				operationSvc.On("RescheduleOperation", txtest.CtxWithDBMatcher(), operationID, int(HighOperationPriority)).Return(nil).Once()
				return operationSvc
			},
			Input:         operationID,
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			operationSvc := testCase.OperationSvcFn()

			opManager := NewOperationsManager(transact, operationSvc, model.OperationTypeOrdAggregation, OperationsManagerConfig{})

			// WHEN
			err := opManager.RescheduleOperation(context.TODO(), testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, operationSvc)
		})
	}
}

func TestOperationsManager_GetOperation(t *testing.T) {
	// GIVEN
	testError := errors.New("test error")
	operation1InProgress := fixOperationModel("aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", model.OperationTypeOrdAggregation, model.OperationStatusInProgress, int(LowOperationPriority))
	operation2InProgress := fixOperationModel("bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", model.OperationTypeOrdAggregation, model.OperationStatusInProgress, int(LowOperationPriority))
	operation3InProgress := fixOperationModel("ccccccccc-cccc-cccc-cccc-cccccccccccc", model.OperationTypeOrdAggregation, model.OperationStatusInProgress, int(LowOperationPriority))

	txGen := txtest.NewTransactionContextGenerator(testError)
	testCases := []struct {
		Name              string
		TxFn              func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		OperationSvcFn    func() *automock.OperationService
		ExpectedError     error
		ExpectedOperation *model.Operation
	}{
		{
			Name: "Success with one in the queue",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(2)
			},
			OperationSvcFn: func() *automock.OperationService {
				operation1 := fixOperationModel("aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", model.OperationTypeOrdAggregation, model.OperationStatusScheduled, int(LowOperationPriority))
				operationSvc := &automock.OperationService{}
				operationSvc.On("ListPriorityQueue", txtest.CtxWithDBMatcher(), mock.Anything, model.OperationTypeOrdAggregation).Return([]*model.Operation{operation1}, nil).Once()
				operationSvc.On("LockOperation", txtest.CtxWithDBMatcher(), operation1.ID).Return(true, nil).Once()
				operationSvc.On("Get", txtest.CtxWithDBMatcher(), operation1.ID).Return(operation1, nil).Once()
				operationSvc.On("Update", txtest.CtxWithDBMatcher(), operation1InProgress).Return(nil).Once()
				return operationSvc
			},
			ExpectedOperation: operation1InProgress,
		},
		{
			Name: "Success with two in the queue",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			OperationSvcFn: func() *automock.OperationService {
				operation1 := fixOperationModel("aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", model.OperationTypeOrdAggregation, model.OperationStatusScheduled, int(LowOperationPriority))
				operation2 := fixOperationModel("bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", model.OperationTypeOrdAggregation, model.OperationStatusScheduled, int(LowOperationPriority))
				operationSvc := &automock.OperationService{}
				operationSvc.On("ListPriorityQueue", txtest.CtxWithDBMatcher(), mock.Anything, model.OperationTypeOrdAggregation).Return([]*model.Operation{operation1, operation2}, nil).Once()
				operationSvc.On("LockOperation", txtest.CtxWithDBMatcher(), operation1.ID).Return(false, nil).Once()
				operationSvc.On("LockOperation", txtest.CtxWithDBMatcher(), operation2.ID).Return(true, nil).Once()
				operationSvc.On("Get", txtest.CtxWithDBMatcher(), operation2.ID).Return(operation2, nil).Once()
				operationSvc.On("Update", txtest.CtxWithDBMatcher(), operation2InProgress).Return(nil).Once()
				return operationSvc
			},
			ExpectedOperation: operation2InProgress,
		},
		{
			Name: "Error when first transaction fails",
			TxFn: txGen.ThatFailsOnBegin,
			OperationSvcFn: func() *automock.OperationService {
				return &automock.OperationService{}
			},
			ExpectedError: testError,
		},
		{
			Name: "Error when list priority queue fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			OperationSvcFn: func() *automock.OperationService {
				operationSvc := &automock.OperationService{}
				operationSvc.On("ListPriorityQueue", txtest.CtxWithDBMatcher(), mock.Anything, model.OperationTypeOrdAggregation).Return(nil, testError).Once()
				return operationSvc
			},
			ExpectedError: testError,
		},
		{
			Name: "Error when transaction after list priority queue fails",
			TxFn: txGen.ThatFailsOnCommit,
			OperationSvcFn: func() *automock.OperationService {
				operation1 := fixOperationModel("aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", model.OperationTypeOrdAggregation, model.OperationStatusScheduled, int(LowOperationPriority))
				operationSvc := &automock.OperationService{}
				operationSvc.On("ListPriorityQueue", txtest.CtxWithDBMatcher(), mock.Anything, model.OperationTypeOrdAggregation).Return([]*model.Operation{operation1}, nil).Once()
				return operationSvc
			},
			ExpectedError: testError,
		},
		{
			Name: "Error when committing first transaction in tryToGet fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, testError).Times(1)
				return persistTx, transact
			},
			OperationSvcFn: func() *automock.OperationService {
				operation1 := fixOperationModel("aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", model.OperationTypeOrdAggregation, model.OperationStatusScheduled, int(LowOperationPriority))
				operationSvc := &automock.OperationService{}
				operationSvc.On("ListPriorityQueue", txtest.CtxWithDBMatcher(), mock.Anything, model.OperationTypeOrdAggregation).Return([]*model.Operation{operation1}, nil).Once()
				return operationSvc
			},
			ExpectedError: testError,
		},
		{
			Name: "Error when lock operation in tryToGet fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Times(1)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(1)
				return persistTx, transact
			},
			OperationSvcFn: func() *automock.OperationService {
				operation1 := fixOperationModel("aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", model.OperationTypeOrdAggregation, model.OperationStatusScheduled, int(LowOperationPriority))
				operationSvc := &automock.OperationService{}
				operationSvc.On("ListPriorityQueue", txtest.CtxWithDBMatcher(), mock.Anything, model.OperationTypeOrdAggregation).Return([]*model.Operation{operation1}, nil).Once()
				operationSvc.On("LockOperation", txtest.CtxWithDBMatcher(), operation1.ID).Return(false, testError).Once()
				return operationSvc
			},
			ExpectedError: testError,
		},
		{
			Name: "Error when lock operation in tryToGet returns nil",
			TxFn: txGen.ThatSucceedsTwice,
			OperationSvcFn: func() *automock.OperationService {
				operation1 := fixOperationModel("aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", model.OperationTypeOrdAggregation, model.OperationStatusScheduled, int(LowOperationPriority))
				operationSvc := &automock.OperationService{}
				operationSvc.On("ListPriorityQueue", txtest.CtxWithDBMatcher(), mock.Anything, model.OperationTypeOrdAggregation).Return([]*model.Operation{operation1}, nil).Once()
				operationSvc.On("LockOperation", txtest.CtxWithDBMatcher(), operation1.ID).Return(false, nil).Once()
				return operationSvc
			},
			ExpectedError: apperrors.NewNoScheduledOperationsError(),
		},
		{
			Name: "Error when Get operation in tryToGet fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Times(1)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(1)
				return persistTx, transact
			},
			OperationSvcFn: func() *automock.OperationService {
				operation1 := fixOperationModel("aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", model.OperationTypeOrdAggregation, model.OperationStatusScheduled, int(LowOperationPriority))
				operationSvc := &automock.OperationService{}
				operationSvc.On("ListPriorityQueue", txtest.CtxWithDBMatcher(), mock.Anything, model.OperationTypeOrdAggregation).Return([]*model.Operation{operation1}, nil).Once()
				operationSvc.On("LockOperation", txtest.CtxWithDBMatcher(), operation1.ID).Return(true, nil).Once()
				operationSvc.On("Get", txtest.CtxWithDBMatcher(), operation1.ID).Return(nil, testError).Once()
				return operationSvc
			},
			ExpectedError: testError,
		},
		{
			Name: "Error when Update operation in tryToGet fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Times(1)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(1)
				return persistTx, transact
			},
			OperationSvcFn: func() *automock.OperationService {
				operation3 := fixOperationModel("ccccccccc-cccc-cccc-cccc-cccccccccccc", model.OperationTypeOrdAggregation, model.OperationStatusScheduled, int(LowOperationPriority))
				operationSvc := &automock.OperationService{}
				operationSvc.On("ListPriorityQueue", txtest.CtxWithDBMatcher(), mock.Anything, model.OperationTypeOrdAggregation).Return([]*model.Operation{operation3}, nil).Once()
				operationSvc.On("LockOperation", txtest.CtxWithDBMatcher(), operation3.ID).Return(true, nil).Once()
				operationSvc.On("Get", txtest.CtxWithDBMatcher(), operation3.ID).Return(operation3, nil).Once()
				operationSvc.On("Update", txtest.CtxWithDBMatcher(), operation3InProgress).Return(testError).Once()
				return operationSvc
			},
			ExpectedOperation: operation3InProgress,
			ExpectedError:     testError,
		},
		{
			Name: "Error when committing transaction after Update operation in tryToGet fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				persistTx.On("Commit").Return(testError).Once()
				transact.On("Begin").Return(persistTx, nil).Times(1)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(1)
				return persistTx, transact
			},
			OperationSvcFn: func() *automock.OperationService {
				operation3 := fixOperationModel("ccccccccc-cccc-cccc-cccc-cccccccccccc", model.OperationTypeOrdAggregation, model.OperationStatusScheduled, int(LowOperationPriority))
				operationSvc := &automock.OperationService{}
				operationSvc.On("ListPriorityQueue", txtest.CtxWithDBMatcher(), mock.Anything, model.OperationTypeOrdAggregation).Return([]*model.Operation{operation3}, nil).Once()
				operationSvc.On("LockOperation", txtest.CtxWithDBMatcher(), operation3.ID).Return(true, nil).Once()
				operationSvc.On("Get", txtest.CtxWithDBMatcher(), operation3.ID).Return(operation3, nil).Once()
				operationSvc.On("Update", txtest.CtxWithDBMatcher(), operation3InProgress).Return(nil).Once()
				return operationSvc
			},
			ExpectedOperation: operation3InProgress,
			ExpectedError:     testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			operationSvc := testCase.OperationSvcFn()

			// Adjust the updated_at timestamp of expected operation if any
			if testCase.ExpectedOperation != nil {
				currentTime := time.Now()
				now = func() time.Time { return currentTime }
				testCase.ExpectedOperation.UpdatedAt = &currentTime
			}

			opManager := NewOperationsManager(transact, operationSvc, model.OperationTypeOrdAggregation, OperationsManagerConfig{})
			// WHEN
			actualOperation, err := opManager.GetOperation(context.TODO())

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOperation, actualOperation)
			}

			mock.AssertExpectationsForObjects(t, persist, transact, operationSvc)
		})
	}
}

// TODO test for FindOperationByData
// TODO test for CreateOperation

func fixOperationModel(id string, opType model.OperationType, opStatus model.OperationStatus, priority int) *model.Operation {
	return &model.Operation{
		ID:        id,
		OpType:    opType,
		Status:    opStatus,
		Data:      json.RawMessage("[]"),
		Error:     json.RawMessage("[]"),
		Priority:  priority,
		CreatedAt: &time.Time{},
		UpdatedAt: &time.Time{},
	}
}
