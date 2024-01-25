package systemfetcher_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	systemfetcher "github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_ScheduleAggregationForSystemFetcherData(t *testing.T) {
	apiPath := "/sync"
	operationID := "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	tenantID := "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	globalAccountID := "ccccccccc-cccc-cccc-cccc-cccccccccccc"

	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                       string
		TransactionerFn            func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		OperationManagerFn         func() *automock.OperationsManager
		BusinessTenantMappingSvcFn func() *automock.BusinessTenantMappingService
		RequestBody                systemfetcher.AggregationResources
		ExpectedErrorOutput        string
		ExpectedStatusCode         int
	}{
		{
			Name:            "Success - operation already exists",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(tenantID)).Return(&model.Operation{ID: operationID}, nil).Once()
				opManager.On("RescheduleOperation", mock.Anything, operationID).Return(nil).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				return &automock.BusinessTenantMappingService{}
			},
			RequestBody: systemfetcher.AggregationResources{
				TenantID: tenantID,
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:            "Success - operation with Tenant does not exist, create new operation",
			TransactionerFn: txGen.ThatSucceeds,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(tenantID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				opManager.On("CreateOperation", mock.Anything, mock.Anything).Return(operationID, nil).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				businessTenantMappingSvc := &automock.BusinessTenantMappingService{}
				businessTenantMappingSvc.On("Exists", txtest.CtxWithDBMatcher(), tenantID).Return(nil).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResources{
				TenantID: tenantID,
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name: "Success - operation with Tenant does not exist and ID is from external tenant, create new operation",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceedsTwice()
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
				return persistTx, transact
			},
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(globalAccountID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				opManager.On("CreateOperation", mock.Anything, mock.Anything).Return(operationID, nil).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				businessTenantMappingSvc := &automock.BusinessTenantMappingService{}
				businessTenantMappingSvc.On("Exists", txtest.CtxWithDBMatcher(), globalAccountID).Return(apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				businessTenantMappingSvc.On("ExistsByExternalTenant", txtest.CtxWithDBMatcher(), globalAccountID).Return(nil).Once()
				businessTenantMappingSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), globalAccountID).Return(tenantID, nil).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResources{
				TenantID: globalAccountID,
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:            "InternalServerError - when RescheduleOperation fails",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(tenantID)).Return(&model.Operation{ID: operationID}, nil).Once()
				opManager.On("RescheduleOperation", mock.Anything, operationID).Return(testErr).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				return &automock.BusinessTenantMappingService{}
			},
			RequestBody: systemfetcher.AggregationResources{
				TenantID: tenantID,
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Scheduling Operation for System Fetcher data aggregation failed",
		},
		{
			Name:            "InternalServerError - error while checking if operation exists",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(tenantID)).Return(nil, testErr).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				return &automock.BusinessTenantMappingService{}
			},
			RequestBody: systemfetcher.AggregationResources{
				TenantID: tenantID,
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Loading Operation for System Fetcher data aggregation failed",
		},
		{
			Name: "NotFound - operation with Tenant ID - the tenant does not exist",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatDoesntStartTransaction()
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Twice()
				return persistTx, transact
			},
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(globalAccountID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				businessTenantMappingSvc := &automock.BusinessTenantMappingService{}
				businessTenantMappingSvc.On("Exists", txtest.CtxWithDBMatcher(), globalAccountID).Return(apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				businessTenantMappingSvc.On("ExistsByExternalTenant", txtest.CtxWithDBMatcher(), globalAccountID).Return(apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResources{
				TenantID: globalAccountID,
			},
			ExpectedStatusCode:  http.StatusNotFound,
			ExpectedErrorOutput: "External Tenant not found",
		},
		{
			Name: "InternalServerError - operation with Tenant ID - error when look for tenant",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatDoesntStartTransaction()
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Twice()
				return persistTx, transact
			},
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(globalAccountID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				businessTenantMappingSvc := &automock.BusinessTenantMappingService{}
				businessTenantMappingSvc.On("Exists", txtest.CtxWithDBMatcher(), globalAccountID).Return(apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				businessTenantMappingSvc.On("ExistsByExternalTenant", txtest.CtxWithDBMatcher(), globalAccountID).Return(testErr).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResources{
				TenantID: globalAccountID,
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Check for External Tenant failed",
		},
		{
			Name: "InternalServerError - operation with Tenant does not exist and ID is from external tenant, create new operation",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Twice()
				return persistTx, transact
			},
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(globalAccountID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				businessTenantMappingSvc := &automock.BusinessTenantMappingService{}
				businessTenantMappingSvc.On("Exists", txtest.CtxWithDBMatcher(), globalAccountID).Return(apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				businessTenantMappingSvc.On("ExistsByExternalTenant", txtest.CtxWithDBMatcher(), globalAccountID).Return(nil).Once()
				businessTenantMappingSvc.On("GetInternalTenant", txtest.CtxWithDBMatcher(), globalAccountID).Return("", testErr).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResources{
				TenantID: globalAccountID,
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Getting External Tenant failed",
		},
		{
			Name:            "InternalServerError - check tenant existence fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(globalAccountID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				businessTenantMappingSvc := &automock.BusinessTenantMappingService{}
				businessTenantMappingSvc.On("Exists", txtest.CtxWithDBMatcher(), globalAccountID).Return(testErr).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResources{
				TenantID: globalAccountID,
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Getting Tenant failed",
		},
		{
			Name:            "InternalServerError - create operation fail",
			TransactionerFn: txGen.ThatSucceeds,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(tenantID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				opManager.On("CreateOperation", mock.Anything, mock.Anything).Return("", testErr).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				businessTenantMappingSvc := &automock.BusinessTenantMappingService{}
				businessTenantMappingSvc.On("Exists", txtest.CtxWithDBMatcher(), tenantID).Return(nil).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResources{
				TenantID: tenantID,
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Creating Operation for System Fetcher data aggregation failed",
		},
		{
			Name:            "BadRequest - invalid payload",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			OperationManagerFn: func() *automock.OperationsManager {
				return &automock.OperationsManager{}
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				return &automock.BusinessTenantMappingService{}
			},
			RequestBody: systemfetcher.AggregationResources{
				TenantID: "",
			},
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "Invalid payload",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, tx := testCase.TransactionerFn()
			operationManager := testCase.OperationManagerFn()
			businessTenantMappingSvc := testCase.BusinessTenantMappingSvcFn()
			defer mock.AssertExpectationsForObjects(t, persist, tx, operationManager, businessTenantMappingSvc)

			onDemandChannel := make(chan string, 2)

			handler := systemfetcher.NewSystemFetcherAggregatorHTTPHandler(operationManager, businessTenantMappingSvc, tx, onDemandChannel)

			requestBody, err := json.Marshal(testCase.RequestBody)
			assert.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, apiPath, bytes.NewBuffer(requestBody))
			writer := httptest.NewRecorder()

			// WHEN
			handler.ScheduleAggregationForSystemFetcherData(writer, request)

			// THEN
			resp := writer.Result()
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			if len(testCase.ExpectedErrorOutput) > 0 {
				assert.Contains(t, string(body), testCase.ExpectedErrorOutput)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.ExpectedStatusCode, resp.StatusCode)
		})
	}
}
