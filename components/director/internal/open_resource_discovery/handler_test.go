package ord_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_ScheduleAggregationForORDData(t *testing.T) {
	apiPath := "/aggregate"
	applicationID := "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	appTemplateID := "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	operationID := "ccccccccc-cccc-cccc-cccc-cccccccccccc"
	webhookID := "ddddddddd-dddd-dddd-dddd-dddddddddddd"
	operation := &model.Operation{ID: operationID}
	webhook := &model.Webhook{ID: webhookID}

	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		OperationManagerFn  func() *automock.OperationsManager
		WebhookSvcFn        func() *automock.WebhookService
		RequestBody         ord.AggregationResources
		ExpectedErrorOutput string
		ExpectedStatusCode  int
	}{
		{
			Name:            "Success - operation already exists",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, ord.NewOrdOperationData(applicationID, appTemplateID)).Return(&model.Operation{ID: operationID}, nil).Once()
				opManager.On("RescheduleOperation", mock.Anything, operationID).Return(nil).Once()
				return opManager
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         applicationID,
				ApplicationTemplateID: appTemplateID,
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:            "Success - operation with Application and ApplicationTemplate does not exist, create new operation",
			TransactionerFn: txGen.ThatSucceeds,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, ord.NewOrdOperationData(applicationID, appTemplateID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				opManager.On("CreateOperation", mock.Anything, mock.Anything).Return(operationID, nil).Once()
				return opManager
			},
			WebhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("GetByIDAndWebhookTypeGlobal", txtest.CtxWithDBMatcher(), appTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeOpenResourceDiscovery).Return(webhook, nil).Once()
				return whSvc
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         applicationID,
				ApplicationTemplateID: appTemplateID,
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:            "Success - operation with Application does not exist, create new operation",
			TransactionerFn: txGen.ThatSucceeds,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, ord.NewOrdOperationData(applicationID, "")).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				opManager.On("CreateOperation", mock.Anything, mock.Anything).Return(operationID, nil).Once()
				return opManager
			},
			WebhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("GetByIDAndWebhookTypeGlobal", txtest.CtxWithDBMatcher(), applicationID, model.ApplicationWebhookReference, model.WebhookTypeOpenResourceDiscovery).Return(webhook, nil).Once()
				return whSvc
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         applicationID,
				ApplicationTemplateID: "",
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:            "Success - operation with ApplicationTemplate does not exist, create new operation",
			TransactionerFn: txGen.ThatSucceeds,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, ord.NewOrdOperationData("", appTemplateID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				opManager.On("CreateOperation", mock.Anything, mock.Anything).Return(operationID, nil).Once()
				return opManager
			},
			WebhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("GetByIDAndWebhookTypeGlobal", txtest.CtxWithDBMatcher(), appTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeOpenResourceDiscovery).Return(webhook, nil).Once()
				return whSvc
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         "",
				ApplicationTemplateID: appTemplateID,
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:            "InternalServerError - error while checking if operation exists",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, ord.NewOrdOperationData(applicationID, "")).Return(nil, testErr).Once()
				return opManager
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         applicationID,
				ApplicationTemplateID: "",
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Loading Operation for ORD data aggregation failed",
		},
		{
			Name:            "BadRequest - provided application does not have ord webhook",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, ord.NewOrdOperationData(applicationID, "")).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			WebhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("GetByIDAndWebhookTypeGlobal", txtest.CtxWithDBMatcher(), applicationID, model.ApplicationWebhookReference, model.WebhookTypeOpenResourceDiscovery).Return(nil, apperrors.NewNotFoundError(resource.Webhook, "")).Once()
				return whSvc
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         applicationID,
				ApplicationTemplateID: "",
			},
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "The provided Application does not have ORD webhook",
		},
		{
			Name:            "InternalServerError - error while checking if the provided application has ord webhook",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, ord.NewOrdOperationData(applicationID, "")).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			WebhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("GetByIDAndWebhookTypeGlobal", txtest.CtxWithDBMatcher(), applicationID, model.ApplicationWebhookReference, model.WebhookTypeOpenResourceDiscovery).Return(nil, testErr).Once()
				return whSvc
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         applicationID,
				ApplicationTemplateID: "",
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Loading ORD webhooks of Application for ORD data aggregation failed",
		},
		{
			Name:            "BadRequest - provided application template does not have ord webhook",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, ord.NewOrdOperationData("", appTemplateID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			WebhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("GetByIDAndWebhookTypeGlobal", txtest.CtxWithDBMatcher(), appTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeOpenResourceDiscovery).Return(nil, apperrors.NewNotFoundError(resource.Webhook, "")).Once()
				return whSvc
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         "",
				ApplicationTemplateID: appTemplateID,
			},
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "The provided ApplicationTemplate does not have ORD webhook",
		},
		{
			Name:            "InternalServerError - error while checking if the provided application template has ord webhook",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, ord.NewOrdOperationData("", appTemplateID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			WebhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("GetByIDAndWebhookTypeGlobal", txtest.CtxWithDBMatcher(), appTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeOpenResourceDiscovery).Return(nil, testErr).Once()
				return whSvc
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         "",
				ApplicationTemplateID: appTemplateID,
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Loading ORD webhooks of Application Template for ORD data aggregation failed",
		},
		{
			Name:            "InternalServerError - error while checking if the provided application and application template have ord webhook",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, ord.NewOrdOperationData(applicationID, appTemplateID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			WebhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("GetByIDAndWebhookTypeGlobal", txtest.CtxWithDBMatcher(), appTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeOpenResourceDiscovery).Return(nil, testErr).Once()
				return whSvc
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         applicationID,
				ApplicationTemplateID: appTemplateID,
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Loading ORD webhooks of Application Template for ORD data aggregation failed",
		},
		{
			Name: "BadRequest - provided application and application template do not have ord webhook",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, ord.NewOrdOperationData(applicationID, appTemplateID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				return opManager
			},
			WebhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("GetByIDAndWebhookTypeGlobal", txtest.CtxWithDBMatcher(), appTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeOpenResourceDiscovery).Return(nil, apperrors.NewNotFoundError(resource.Webhook, webhookID)).Once()
				return whSvc
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         applicationID,
				ApplicationTemplateID: appTemplateID,
			},
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "The provided ApplicationTemplate does not have ORD webhook",
		},
		{
			Name:            "InternalServerError - create operation fail",
			TransactionerFn: txGen.ThatSucceeds,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, ord.NewOrdOperationData("", appTemplateID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				opManager.On("CreateOperation", mock.Anything, mock.Anything).Return("", testErr).Once()
				return opManager
			},
			WebhookSvcFn: func() *automock.WebhookService {
				whSvc := &automock.WebhookService{}
				whSvc.On("GetByIDAndWebhookTypeGlobal", txtest.CtxWithDBMatcher(), appTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeOpenResourceDiscovery).Return(webhook, nil).Once()
				return whSvc
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         "",
				ApplicationTemplateID: appTemplateID,
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Creating Operation for ORD data aggregation failed",
		},
		{
			Name:            "InternalServerError - error while rescheduling operation",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, ord.NewOrdOperationData(applicationID, appTemplateID)).Return(operation, nil).Once()
				opManager.On("RescheduleOperation", mock.Anything, operationID).Return(testErr).Once()
				return opManager
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         applicationID,
				ApplicationTemplateID: appTemplateID,
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Scheduling Operation for ORD data aggregation failed",
		},
		{
			Name:            "BadRequest - invalid payload",
			TransactionerFn: txGen.ThatDoesntStartTransaction,
			OperationManagerFn: func() *automock.OperationsManager {
				return &automock.OperationsManager{}
			},
			WebhookSvcFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			RequestBody: ord.AggregationResources{
				ApplicationID:         "",
				ApplicationTemplateID: "",
			},
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "Invalid payload",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, tx := testCase.TransactionerFn()
			operationManager := testCase.OperationManagerFn()
			webhookSvc := testCase.WebhookSvcFn()
			defer mock.AssertExpectationsForObjects(t, persist, tx, operationManager, webhookSvc)

			onDemandChannel := make(chan string, 2)

			handler := ord.NewORDAggregatorHTTPHandler(operationManager, webhookSvc, tx, onDemandChannel)

			requestBody, err := json.Marshal(testCase.RequestBody)
			assert.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, apiPath, bytes.NewBuffer(requestBody))
			writer := httptest.NewRecorder()

			// WHEN
			handler.ScheduleAggregationForORDData(writer, request)

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
