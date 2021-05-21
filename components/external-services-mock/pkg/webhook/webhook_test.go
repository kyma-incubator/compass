package webhook_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/webhook"
	"github.com/stretchr/testify/require"
)

func TestWebhook(t *testing.T) {
	var handler http.HandlerFunc

	t.Run("should return 423 by default", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "/webhook/delete", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		handler = webhook.NewDeleteHTTPHandler()
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusLocked, rr.Code)
	})

	t.Run("should return locked after lock", func(t *testing.T) {
		// Lock
		requestData := webhook.OperationStatusRequestData{
			InProgress: true,
		}
		jsonRequestData, err := json.Marshal(requestData)
		require.NoError(t, err)
		reqPostFalse, err := http.NewRequest(http.MethodPost, "/webhook/delete/operation", bytes.NewBuffer(jsonRequestData))
		require.NoError(t, err)
		rrPostFalse := httptest.NewRecorder()
		handler = webhook.NewWebHookOperationPostHTTPHandler()
		handler.ServeHTTP(rrPostFalse, reqPostFalse)
		require.Equal(t, http.StatusOK, rrPostFalse.Code)

		// Verify operation is in progress
		req, err := http.NewRequest(http.MethodGet, "/webhook/delete/operation", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		handler = webhook.NewWebHookOperationGetHTTPHandler()
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Contains(t, rr.Body.String(), webhook.OperationResponseStatusINProgress)

		// Verify is locked
		req, err = http.NewRequest(http.MethodDelete, "/webhook/delete", nil)
		require.NoError(t, err)
		rr = httptest.NewRecorder()
		handler = webhook.NewDeleteHTTPHandler()
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusLocked, rr.Code)
	})

	t.Run("should return unlocked after unlock", func(t *testing.T) {
		// Unlock
		requestData := webhook.OperationStatusRequestData{
			InProgress: false,
		}
		jsonRequestData, err := json.Marshal(requestData)
		require.NoError(t, err)
		reqPostTrue, err := http.NewRequest(http.MethodPost, "/webhook/delete/operation", bytes.NewBuffer(jsonRequestData))
		require.NoError(t, err)
		rrPostTrue := httptest.NewRecorder()
		handler = webhook.NewWebHookOperationPostHTTPHandler()
		handler.ServeHTTP(rrPostTrue, reqPostTrue)
		require.Equal(t, http.StatusOK, rrPostTrue.Code)

		// Verify operation is succeeded
		req, err := http.NewRequest(http.MethodGet, "/webhook/delete/operation", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		handler = webhook.NewWebHookOperationGetHTTPHandler()
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Contains(t, rr.Body.String(), webhook.OperationResponseStatusOK)

		// Verify is unlocked
		req, err = http.NewRequest(http.MethodDelete, "/webhook/delete", nil)
		require.NoError(t, err)
		rr = httptest.NewRecorder()
		handler = webhook.NewDeleteHTTPHandler()
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should fail when call POST handler with GET", func(t *testing.T) {
		// Unlock
		requestData := webhook.OperationStatusRequestData{
			InProgress: true,
		}
		jsonRequestData, err := json.Marshal(requestData)
		require.NoError(t, err)
		reqPostTrue, err := http.NewRequest(http.MethodGet, "/webhook/delete/operation", bytes.NewBuffer(jsonRequestData))
		require.NoError(t, err)
		rrPostTrue := httptest.NewRecorder()
		handler = webhook.NewWebHookOperationPostHTTPHandler()
		handler.ServeHTTP(rrPostTrue, reqPostTrue)
		require.Equal(t, http.StatusMethodNotAllowed, rrPostTrue.Code)

	})

	t.Run("should fail when call GET handler with POST", func(t *testing.T) {
		// Unlock
		requestData := webhook.OperationStatusRequestData{
			InProgress: true,
		}
		jsonRequestData, err := json.Marshal(requestData)
		require.NoError(t, err)
		reqPostTrue, err := http.NewRequest(http.MethodPost, "/webhook/delete/operation", bytes.NewBuffer(jsonRequestData))
		require.NoError(t, err)
		rrPostTrue := httptest.NewRecorder()
		handler = webhook.NewWebHookOperationGetHTTPHandler()
		handler.ServeHTTP(rrPostTrue, reqPostTrue)
		require.Equal(t, http.StatusMethodNotAllowed, rrPostTrue.Code)

	})

}
