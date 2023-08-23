package operationsmanager

import (
	"context"
	"errors"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/operations_manager/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
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

			opManager := NewOperationsManager(transact, operationSvc, model.OrdAggregationOpType, OperationsManagerConfig{})

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

			opManager := NewOperationsManager(transact, operationSvc, model.OrdAggregationOpType, OperationsManagerConfig{})

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
				operationSvc.On("RescheduleOperation", txtest.CtxWithDBMatcher(), operationID, int(highOperationPriority)).Return(nil).Once()
				return operationSvc
			},
			Input: operationID,
		},
		{
			Name: "Error while rescheduling operation",
			TxFn: txGen.ThatDoesntExpectCommit,
			OperationSvcFn: func() *automock.OperationService {
				operationSvc := &automock.OperationService{}
				operationSvc.On("RescheduleOperation", txtest.CtxWithDBMatcher(), operationID, int(highOperationPriority)).Return(testError).Once()
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
				operationSvc.On("RescheduleOperation", txtest.CtxWithDBMatcher(), operationID, int(highOperationPriority)).Return(nil).Once()
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

			opManager := NewOperationsManager(transact, operationSvc, model.OrdAggregationOpType, OperationsManagerConfig{})

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
