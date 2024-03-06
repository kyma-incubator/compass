package systemfetcher_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

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
	foundAccount := &model.BusinessTenantMapping{
		ID:   tenantID,
		Type: tenant.Account,
	}
	foundFolder := &model.BusinessTenantMapping{
		ID:   tenantID,
		Type: tenant.Folder,
	}

	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                       string
		TransactionerFn            func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		OperationManagerFn         func() *automock.OperationsManager
		BusinessTenantMappingSvcFn func() *automock.BusinessTenantMappingService
		RequestBody                systemfetcher.AggregationResource
		ExpectedErrorOutput        string
		ExpectedStatusCode         int
	}{
		{
			Name:            "Success - operation already exists, reschedule",
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
			RequestBody: systemfetcher.AggregationResource{
				TenantIDs: []string{tenantID},
			},
			ExpectedStatusCode: http.StatusAccepted,
		},
		{
			Name:            "Success - operation already exists, skip reschedule",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(tenantID)).Return(&model.Operation{ID: operationID}, nil).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				return &automock.BusinessTenantMappingService{}
			},
			RequestBody: systemfetcher.AggregationResource{
				TenantIDs:      []string{tenantID},
				SkipReschedule: true,
			},
			ExpectedStatusCode: http.StatusAccepted,
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
				businessTenantMappingSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(foundAccount, nil).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResource{
				TenantIDs: []string{tenantID},
			},
			ExpectedStatusCode: http.StatusAccepted,
		},
		{
			Name: "Success - operation with Tenant does not exist and ID is from external tenant, create new operation",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx, transact := txGen.ThatSucceeds()
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
				businessTenantMappingSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), globalAccountID).Return(nil, apperrors.NewNotFoundError(resource.Tenant, tenantID)).Once()
				businessTenantMappingSvc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), globalAccountID).Return(foundAccount, nil).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResource{
				TenantIDs: []string{globalAccountID},
			},
			ExpectedStatusCode: http.StatusAccepted,
		},
		{
			Name:            "RescheduleOperation returns error",
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
			RequestBody: systemfetcher.AggregationResource{
				TenantIDs: []string{tenantID},
			},
			ExpectedStatusCode: http.StatusAccepted,
		},
		{
			Name:            "FindOperationByData returns error",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(tenantID)).Return(nil, testErr).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				return &automock.BusinessTenantMappingService{}
			},
			RequestBody: systemfetcher.AggregationResource{
				TenantIDs: []string{tenantID},
			},
			ExpectedStatusCode: http.StatusAccepted,
		},
		{
			Name: "Getting tenant by internal id returns error",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(tenantID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				businessTenantMappingSvc := &automock.BusinessTenantMappingService{}
				businessTenantMappingSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(nil, testErr).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResource{
				TenantIDs: []string{tenantID},
			},
			ExpectedStatusCode: http.StatusAccepted,
		},
		{
			Name: "Getting tenant by external id returns error",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persist, tx := txGen.ThatDoesntExpectCommit()
				tx.On("Begin").Return(persist, nil).Once()
				tx.On("RollbackUnlessCommitted", mock.Anything, persist).Return(true).Once()

				return persist, tx
			},
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(tenantID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				businessTenantMappingSvc := &automock.BusinessTenantMappingService{}
				businessTenantMappingSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(nil, apperrors.NewNotFoundError(resource.Tenant, tenantID)).Once()
				businessTenantMappingSvc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), tenantID).Return(nil, testErr).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResource{
				TenantIDs: []string{tenantID},
			},
			ExpectedStatusCode: http.StatusAccepted,
		},
		{
			Name: "Getting tenant by external id returns not found error",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persist, tx := txGen.ThatDoesntExpectCommit()
				tx.On("Begin").Return(persist, nil).Once()
				tx.On("RollbackUnlessCommitted", mock.Anything, persist).Return(true).Once()

				return persist, tx
			},
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(tenantID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				businessTenantMappingSvc := &automock.BusinessTenantMappingService{}
				businessTenantMappingSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(nil, apperrors.NewNotFoundError(resource.Tenant, tenantID)).Once()
				businessTenantMappingSvc.On("GetTenantByExternalID", txtest.CtxWithDBMatcher(), tenantID).Return(nil, apperrors.NewNotFoundError(resource.Tenant, tenantID)).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResource{
				TenantIDs: []string{tenantID},
			},
			ExpectedStatusCode: http.StatusAccepted,
		},
		{
			Name:            "Nil business tenant mapping response",
			TransactionerFn: txGen.ThatSucceeds,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(tenantID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				businessTenantMappingSvc := &automock.BusinessTenantMappingService{}
				businessTenantMappingSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(nil, nil).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResource{
				TenantIDs: []string{tenantID},
			},
			ExpectedStatusCode: http.StatusAccepted,
		},
		{
			Name:            "Success - business tenant mapping not of type account or customer",
			TransactionerFn: txGen.ThatSucceeds,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(tenantID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				businessTenantMappingSvc := &automock.BusinessTenantMappingService{}
				businessTenantMappingSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(foundFolder, nil).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResource{
				TenantIDs: []string{tenantID},
			},
			ExpectedStatusCode: http.StatusAccepted,
		},
		{
			Name:            "Create operation returns errror",
			TransactionerFn: txGen.ThatSucceeds,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, systemfetcher.NewSystemFetcherOperationData(tenantID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				opManager.On("CreateOperation", mock.Anything, mock.Anything).Return("", testErr).Once()
				return opManager
			},
			BusinessTenantMappingSvcFn: func() *automock.BusinessTenantMappingService {
				businessTenantMappingSvc := &automock.BusinessTenantMappingService{}
				businessTenantMappingSvc.On("GetTenantByID", txtest.CtxWithDBMatcher(), tenantID).Return(foundAccount, nil).Once()
				return businessTenantMappingSvc
			},
			RequestBody: systemfetcher.AggregationResource{
				TenantIDs: []string{tenantID},
			},
			ExpectedStatusCode: http.StatusAccepted,
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
			RequestBody: systemfetcher.AggregationResource{
				TenantIDs: []string{},
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
			workersChannel := make(chan struct{}, 10)
			handler := systemfetcher.NewSystemFetcherAggregatorHTTPHandler(operationManager, businessTenantMappingSvc, tx, onDemandChannel, workersChannel)

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
			assert.Eventually(t, func() bool {
				return len(workersChannel) == 0
			}, 3*time.Second, 100*time.Millisecond)
		})
	}
}
