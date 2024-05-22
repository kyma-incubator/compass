package operation_test

import (
	"context"
	"errors"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestResolver_Operation(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")
	now := time.Now()
	testID := "8ee7ef81-ca8e-4399-a5d2-3a5f96ecc4c8"
	modelOperation := fixOperationModelWithID(testID, model.OperationTypeOrdAggregation, model.OperationStatusFailed, 1)
	graphqlOperation := fixOperationGraphqlWithIDAndTimestamp(testID, graphql.ScheduledOperationTypeOrdAggregation, graphql.OperationStatusFailed, "error message", &now)
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn         func() *automock.OperationService
		ConverterFn       func() *automock.OperationConverter
		ExpectedOperation *graphql.Operation
		ExpectedErr       error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelOperation, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.OperationConverter {
				conv := &automock.OperationConverter{}
				conv.On("ToGraphQL", modelOperation).Return(graphqlOperation, nil).Once()
				return conv
			},
			ExpectedOperation: graphqlOperation,
			ExpectedErr:       nil,
		},
		{
			Name:            "Returns error when getting operation fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: emptyOperationConverter,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when converting operation fails",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelOperation, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.OperationConverter {
				conv := &automock.OperationConverter{}
				conv.On("ToGraphQL", modelOperation).Return(nil, testErr).Once()
				return conv
			},
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when starting transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn:       emptyOperationService,
			ConverterFn:     emptyOperationConverter,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelOperation, nil).Once()
				return svc
			},
			ConverterFn: emptyOperationConverter,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			resolver := operation.NewResolver(transact, svc, converter)
			defer mock.AssertExpectationsForObjects(t, transact, persistTx, svc, converter)

			// WHEN
			result, err := resolver.Operation(context.TODO(), testID)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, testCase.ExpectedOperation, result)
			}
		})
	}
}

func TestResolver_Schedule(t *testing.T) {
	// GIVEN
	testErr := errors.New("test error")
	now := time.Now()
	testID := "8ee7ef81-ca8e-4399-a5d2-3a5f96ecc4c8"
	modelOperation := fixOperationModelWithID(testID, model.OperationTypeOrdAggregation, model.OperationStatusFailed, 1)
	graphqlOperation := fixOperationGraphqlWithIDAndTimestamp(testID, graphql.ScheduledOperationTypeOrdAggregation, graphql.OperationStatusFailed, "error message", &now)
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn         func() *automock.OperationService
		ConverterFn       func() *automock.OperationConverter
		ExpectedOperation *graphql.Operation
		ExpectedErr       error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("RescheduleOperation", txtest.CtxWithDBMatcher(), testID, int(operationsmanager.HighOperationPriority)).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelOperation, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.OperationConverter {
				conv := &automock.OperationConverter{}
				conv.On("ToGraphQL", modelOperation).Return(graphqlOperation, nil).Once()
				return conv
			},
			ExpectedOperation: graphqlOperation,
			ExpectedErr:       nil,
		},
		{
			Name:            "Returns error when rescheduling operation fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("RescheduleOperation", txtest.CtxWithDBMatcher(), testID, int(operationsmanager.HighOperationPriority)).Return(testErr).Once()
				return svc
			},
			ConverterFn: emptyOperationConverter,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when getting operation fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("RescheduleOperation", txtest.CtxWithDBMatcher(), testID, int(operationsmanager.HighOperationPriority)).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: emptyOperationConverter,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when converting operation fails",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("RescheduleOperation", txtest.CtxWithDBMatcher(), testID, int(operationsmanager.HighOperationPriority)).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelOperation, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.OperationConverter {
				conv := &automock.OperationConverter{}
				conv.On("ToGraphQL", modelOperation).Return(nil, testErr).Once()
				return conv
			},
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when starting transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn:       emptyOperationService,
			ConverterFn:     emptyOperationConverter,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.OperationService {
				svc := &automock.OperationService{}
				svc.On("RescheduleOperation", txtest.CtxWithDBMatcher(), testID, int(operationsmanager.HighOperationPriority)).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(modelOperation, nil).Once()
				return svc
			},
			ConverterFn: emptyOperationConverter,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			resolver := operation.NewResolver(transact, svc, converter)
			defer mock.AssertExpectationsForObjects(t, transact, persistTx, svc, converter)

			// WHEN
			result, err := resolver.Schedule(context.TODO(), testID)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, testCase.ExpectedOperation, result)
			}
		})
	}
}

func emptyOperationService() *automock.OperationService {
	return &automock.OperationService{}
}

func emptyOperationConverter() *automock.OperationConverter {
	return &automock.OperationConverter{}
}
