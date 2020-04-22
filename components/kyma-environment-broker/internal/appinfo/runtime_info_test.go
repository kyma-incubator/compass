package appinfo_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/appinfo"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/appinfo/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/httputil"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/logger"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sebdah/goldie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRuntimeInfoHandlerSuccess(t *testing.T) {
	tests := map[string]struct {
		instances     []internal.Instance
		provisionOp   []internal.ProvisioningOperation
		deprovisionOp []internal.DeprovisioningOperation
	}{
		"no instances": {
			instances: []internal.Instance{},
		},
		"instances without operations": {
			instances: []internal.Instance{
				fixInstance(1), fixInstance(2), fixInstance(2),
			},
		},
		"instances without service and plan name should have defaults": {
			instances: func() []internal.Instance {
				i := fixInstance(1)
				i.ServicePlanName = ""
				i.ServiceName = ""
				// selecting servicePlanName based on existing real planID
				i.ServicePlanID = broker.GCPPlanID
				return []internal.Instance{i}
			}(),
		},
		"instances with provision operation": {
			instances: []internal.Instance{
				fixInstance(1), fixInstance(2), fixInstance(3),
			},
			provisionOp: []internal.ProvisioningOperation{
				fixProvisionOperation(1), fixProvisionOperation(2),
			},
		},
		"instances with deprovision operation": {
			instances: []internal.Instance{
				fixInstance(1), fixInstance(2), fixInstance(3),
			},
			deprovisionOp: []internal.DeprovisioningOperation{
				fixDeprovisionOperation(1), fixDeprovisionOperation(2),
			},
		},
		"instances with provision and deprovision operations": {
			instances: []internal.Instance{
				fixInstance(1), fixInstance(2), fixInstance(3),
			},
			provisionOp: []internal.ProvisioningOperation{
				fixProvisionOperation(1), fixProvisionOperation(2),
			},
			deprovisionOp: []internal.DeprovisioningOperation{
				fixDeprovisionOperation(1), fixDeprovisionOperation(2),
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			var (
				fixReq     = httptest.NewRequest("GET", "http://example.com/foo", nil)
				respSpy    = httptest.NewRecorder()
				writer     = httputil.NewResponseWriter(logger.NewLogDummy(), true)
				memStorage = newInMemoryStorage(t, tc.instances, tc.provisionOp, tc.deprovisionOp)
			)

			handler := appinfo.NewRuntimeInfoHandler(memStorage.Instances(), writer)

			// when
			handler.ServeHTTP(respSpy, fixReq)

			// then
			assert.Equal(t, http.StatusOK, respSpy.Result().StatusCode)
			assert.Equal(t, "application/json", respSpy.Result().Header.Get("Content-Type"))

			assertJSONWithGoldenFile(t, respSpy.Body.Bytes())
		})
	}
}

func TestRuntimeInfoHandlerFailures(t *testing.T) {
	// given
	var (
		fixReq  = httptest.NewRequest("GET", "http://example.com/foo", nil)
		respSpy = httptest.NewRecorder()
		writer  = httputil.NewResponseWriter(logger.NewLogDummy(), true)
		expBody = `{
				  "status": 500,
				  "requestId": "",
				  "message": "Something went very wrong. Please try again.",
				  "details": "while fetching all instances: ups.. internal info"
				}`
	)

	storageMock := &automock.InstanceFinder{}
	defer storageMock.AssertExpectations(t)
	storageMock.On("FindAllJoinedWithOperations", mock.Anything).Return(nil, errors.New("ups.. internal info"))
	handler := appinfo.NewRuntimeInfoHandler(storageMock, writer)

	// when
	handler.ServeHTTP(respSpy, fixReq)

	// then
	assert.Equal(t, http.StatusInternalServerError, respSpy.Result().StatusCode)
	assert.Equal(t, "application/json", respSpy.Result().Header.Get("Content-Type"))

	assert.JSONEq(t, expBody, respSpy.Body.String())
}

func assertJSONWithGoldenFile(t *testing.T, gotRawJSON []byte) {
	t.Helper()
	g := goldie.New(t, goldie.WithNameSuffix(".golden.json"))

	var jsonGoType interface{}
	require.NoError(t, json.Unmarshal(gotRawJSON, &jsonGoType))
	g.AssertJson(t, t.Name(), jsonGoType)
}

func fixTime() time.Time {
	return time.Date(2020, 04, 21, 0, 0, 23, 42, time.UTC)
}

func fixInstance(idx int) internal.Instance {
	return internal.Instance{
		InstanceID:             fmt.Sprintf("InstanceID field. IDX: %d", idx),
		RuntimeID:              fmt.Sprintf("RuntimeID field. IDX: %d", idx),
		GlobalAccountID:        fmt.Sprintf("GlobalAccountID field. IDX: %d", idx),
		SubAccountID:           fmt.Sprintf("SubAccountID field. IDX: %d", idx),
		ServiceID:              fmt.Sprintf("ServiceID field. IDX: %d", idx),
		ServiceName:            fmt.Sprintf("ServiceName field. IDX: %d", idx),
		ServicePlanID:          fmt.Sprintf("ServicePlanID field. IDX: %d", idx),
		ServicePlanName:        fmt.Sprintf("ServicePlanName field. IDX: %d", idx),
		DashboardURL:           fmt.Sprintf("DashboardURL field. IDX: %d", idx),
		ProvisioningParameters: fmt.Sprintf("ProvisioningParameters field. IDX: %d", idx),
		CreatedAt:              fixTime().Add(time.Duration(idx) * time.Second),
		UpdatedAt:              fixTime().Add(time.Duration(idx) * time.Minute),
		DelatedAt:              fixTime().Add(time.Duration(idx) * time.Hour),
	}
}

func newInMemoryStorage(t *testing.T,
	instances []internal.Instance,
	provisionOp []internal.ProvisioningOperation,
	deprovisionOp []internal.DeprovisioningOperation) storage.BrokerStorage {

	t.Helper()
	memStorage := storage.NewMemoryStorage()
	for _, i := range instances {
		require.NoError(t, memStorage.Instances().Insert(i))
	}
	for _, op := range provisionOp {
		require.NoError(t, memStorage.Operations().InsertProvisioningOperation(op))
	}
	for _, op := range deprovisionOp {
		require.NoError(t, memStorage.Operations().InsertDeprovisioningOperation(op))
	}

	return memStorage
}

func fixProvisionOperation(idx int) internal.ProvisioningOperation {
	return internal.ProvisioningOperation{
		Operation: fixSucceededOperation(idx),
	}
}
func fixDeprovisionOperation(idx int) internal.DeprovisioningOperation {
	return internal.DeprovisioningOperation{
		Operation: fixSucceededOperation(idx),
	}
}

func fixSucceededOperation(idx int) internal.Operation {
	return internal.Operation{
		ID:                     fmt.Sprintf("Operation ID field. IDX: %d", idx),
		Version:                0,
		CreatedAt:              fixTime().Add(time.Duration(idx) * 24 * time.Hour),
		UpdatedAt:              fixTime().Add(time.Duration(idx) * 48 * time.Hour),
		InstanceID:             fmt.Sprintf("InstanceID field. IDX: %d", idx),
		ProvisionerOperationID: fmt.Sprintf("ProvisionerOperationID field. IDX: %d", idx),
		State:                  domain.Succeeded,
		Description:            fmt.Sprintf("esc for succeeded op.. IDX: %d", idx),
	}
}
