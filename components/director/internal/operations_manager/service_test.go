package operations_manager_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/internal/operations_manager/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	ordOpType         = "ORD_AGGREGATION"
	completedOpStatus = "COMPLETED"
	failedOpStatus    = "FAILED"
)

func TestService_CreateORDOperations(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	testErr := errors.New("Test error")

	testCases := []struct {
		Name        string
		OpCreatorFn func() *automock.OperationCreator
		ExpectedErr error
	}{
		{
			Name: "Success",
			OpCreatorFn: func() *automock.OperationCreator {
				creator := &automock.OperationCreator{}
				creator.On("Create", ctx).Return(nil).Once()
				return creator
			},
		},
		{
			Name: "Error while creating ord operations",
			OpCreatorFn: func() *automock.OperationCreator {
				creator := &automock.OperationCreator{}
				creator.On("Create", ctx).Return(testErr).Once()
				return creator
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			opCreator := testCase.OpCreatorFn()
			svc := operations_manager.NewOperationService(nil, nil, opCreator)

			// WHEN
			err := svc.CreateORDOperations(ctx)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, opCreator)
		})
	}
}

func TestService_DeleteOldOperations(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	completedOpDays := 1
	failedOpDays := 1
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testErr)
	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		OpSvcFn         func() *automock.OperationService
		ExpectedErr     error
	}{
		{
			Name: "Success",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceeds()
			},
			OpSvcFn: func() *automock.OperationService {
				repo := &automock.OperationService{}
				repo.On("DeleteOlderThan", txtest.CtxWithDBMatcher(), ordOpType, completedOpStatus, completedOpDays).Return(nil).Once()
				repo.On("DeleteOlderThan", txtest.CtxWithDBMatcher(), ordOpType, failedOpStatus, failedOpDays).Return(nil).Once()
				return repo
			},
		},
		{
			Name: "Error while deleting old completed operations",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceeds()
			},
			OpSvcFn: func() *automock.OperationService {
				repo := &automock.OperationService{}
				repo.On("DeleteOlderThan", txtest.CtxWithDBMatcher(), ordOpType, completedOpStatus, completedOpDays).Return(testErr).Once()
				return repo
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error while deleting old failed operations",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceeds()
			},
			OpSvcFn: func() *automock.OperationService {
				repo := &automock.OperationService{}
				repo.On("DeleteOlderThan", txtest.CtxWithDBMatcher(), ordOpType, completedOpStatus, completedOpDays).Return(nil).Once()
				repo.On("DeleteOlderThan", txtest.CtxWithDBMatcher(), ordOpType, failedOpStatus, failedOpDays).Return(testErr).Once()

				return repo
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Error while beginning transaction",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			OpSvcFn: func() *automock.OperationService {
				return &automock.OperationService{}
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			_, tx := testCase.TransactionerFn()
			opSvc := testCase.OpSvcFn()

			svc := operations_manager.NewOperationService(tx, opSvc, nil)

			// WHEN
			err := svc.DeleteOldOperations(ctx, ordOpType, completedOpDays, failedOpDays)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, tx, opSvc)
		})
	}
}
