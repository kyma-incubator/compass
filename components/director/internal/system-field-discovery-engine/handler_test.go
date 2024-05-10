package systemfielddiscoveryengine_test

import (
	"bytes"
	"encoding/json"
	"errors"
	systemfielddiscoveryengine "github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine"
	"github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/data"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_ScheduleAggregationForORDData(t *testing.T) {
	apiPath := "/aggregate"
	applicationID := "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	tenantID := "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	operationID := "ccccccccc-cccc-cccc-cccc-cccccccccccc"
	operation := &model.Operation{ID: operationID}

	testErr := errors.New("test error")

	testCases := []struct {
		Name                string
		OperationManagerFn  func() *automock.OperationsManager
		RequestBody         systemfielddiscoveryengine.SystemFieldDiscoveryResources
		ExpectedErrorOutput string
		ExpectedStatusCode  int
	}{
		{
			Name: "Success - operation already exists",
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, data.NewSystemFieldDiscoveryOperationData(applicationID, tenantID)).Return(operation, nil).Once()
				opManager.On("RescheduleOperation", mock.Anything, operationID).Return(nil).Once()
				return opManager
			},
			RequestBody: systemfielddiscoveryengine.SystemFieldDiscoveryResources{
				ApplicationID: applicationID,
				TenantID:      tenantID,
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name: "Success - operation does not exist, create new operation",
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, data.NewSystemFieldDiscoveryOperationData(applicationID, tenantID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				opManager.On("CreateOperation", mock.Anything, mock.Anything).Return(operationID, nil).Once()
				return opManager
			},
			RequestBody: systemfielddiscoveryengine.SystemFieldDiscoveryResources{
				ApplicationID: applicationID,
				TenantID:      tenantID,
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name: "InternalServerError - error while checking if operation exists",
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, data.NewSystemFieldDiscoveryOperationData(applicationID, tenantID)).Return(nil, testErr).Once()
				return opManager
			},
			RequestBody: systemfielddiscoveryengine.SystemFieldDiscoveryResources{
				ApplicationID: applicationID,
				TenantID:      tenantID,
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Loading Operation for system field discovery data aggregation failed",
		},
		{
			Name: "BadRequest - empty application id",
			OperationManagerFn: func() *automock.OperationsManager {
				return &automock.OperationsManager{}
			},
			RequestBody: systemfielddiscoveryengine.SystemFieldDiscoveryResources{
				ApplicationID: "",
				TenantID:      tenantID,
			},
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "Invalid payload, Application ID or Tenant ID are not provided.",
		},
		{
			Name: "BadRequest - empty tenant id",
			OperationManagerFn: func() *automock.OperationsManager {
				return &automock.OperationsManager{}
			},
			RequestBody: systemfielddiscoveryengine.SystemFieldDiscoveryResources{
				ApplicationID: applicationID,
				TenantID:      "",
			},
			ExpectedStatusCode:  http.StatusBadRequest,
			ExpectedErrorOutput: "Invalid payload, Application ID or Tenant ID are not provided.",
		},
		{
			Name: "InternalServerError - create operation fail",
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, data.NewSystemFieldDiscoveryOperationData(applicationID, tenantID)).Return(nil, apperrors.NewNotFoundError(resource.Operation, operationID)).Once()
				opManager.On("CreateOperation", mock.Anything, mock.Anything).Return("", testErr).Once()
				return opManager
			},
			RequestBody: systemfielddiscoveryengine.SystemFieldDiscoveryResources{
				ApplicationID: applicationID,
				TenantID:      tenantID,
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Creating Operation for system field discovery data aggregation failed",
		},
		{
			Name: "InternalServerError - error while rescheduling operation",
			OperationManagerFn: func() *automock.OperationsManager {
				opManager := &automock.OperationsManager{}
				opManager.On("FindOperationByData", mock.Anything, data.NewSystemFieldDiscoveryOperationData(applicationID, tenantID)).Return(operation, nil).Once()
				opManager.On("RescheduleOperation", mock.Anything, operationID).Return(testErr).Once()
				return opManager
			},
			RequestBody: systemfielddiscoveryengine.SystemFieldDiscoveryResources{
				ApplicationID: applicationID,
				TenantID:      tenantID,
			},
			ExpectedStatusCode:  http.StatusInternalServerError,
			ExpectedErrorOutput: "Scheduling Operation for system field discovery data aggregation failed",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			operationManager := testCase.OperationManagerFn()
			defer mock.AssertExpectationsForObjects(t, operationManager)

			onDemandChannel := make(chan string, 2)

			handler := systemfielddiscoveryengine.NewSystemFieldDiscoveryHTTPHandler(operationManager, onDemandChannel)

			requestBody, err := json.Marshal(testCase.RequestBody)
			assert.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, apiPath, bytes.NewBuffer(requestBody))
			writer := httptest.NewRecorder()

			// WHEN
			handler.ScheduleSaaSRegistryDiscoveryForSystemFieldDiscoveryData(writer, request)

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
