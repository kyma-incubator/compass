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
	t.Run("should return 423 by default", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/webhook/delete", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(webhook.NewDeleteHTTPHandler())
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusLocked, rr.Code)
	})

	t.Run("should return locked after lock", func(t *testing.T) {
		// Lock
		okRequestDataFalse := webhook.OKRequestData{
			OK: false,
		}
		jsonValueoOKRequestDataFalse, err := json.Marshal(okRequestDataFalse)
		require.NoError(t, err)
		reqPostFalse, err := http.NewRequest(http.MethodPost, "/webhook/delete/operation", bytes.NewBuffer(jsonValueoOKRequestDataFalse))
		require.NoError(t, err)
		rrPostFalse := httptest.NewRecorder()
		handlerPostFalse := http.HandlerFunc(webhook.NewWebHookOperationHTTPHandler())
		handlerPostFalse.ServeHTTP(rrPostFalse, reqPostFalse)
		require.Equal(t, http.StatusOK, rrPostFalse.Code)

		// Verify operation is in progress
		req, err := http.NewRequest(http.MethodGet, "/webhook/delete/operation", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(webhook.NewWebHookOperationHTTPHandler())
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Contains(t, rr.Body.String(), webhook.OperationResponseStatusINProgress)

		// Verify is locked
		req, err = http.NewRequest(http.MethodGet, "/webhook/delete", nil)
		require.NoError(t, err)
		rr = httptest.NewRecorder()
		handler = http.HandlerFunc(webhook.NewDeleteHTTPHandler())
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusLocked, rr.Code)
	})

	t.Run("should return unlocked after unlock", func(t *testing.T) {
		// Unlock
		okRequestDataTrue := webhook.OKRequestData{
			OK: true,
		}
		jsonValueoOKRequestDataTrue, err := json.Marshal(okRequestDataTrue)
		require.NoError(t, err)
		reqPostTrue, err := http.NewRequest(http.MethodPost, "/webhook/delete/operation", bytes.NewBuffer(jsonValueoOKRequestDataTrue))
		require.NoError(t, err)
		rrPostTrue := httptest.NewRecorder()
		handlerPostTrue := http.HandlerFunc(webhook.NewWebHookOperationHTTPHandler())
		handlerPostTrue.ServeHTTP(rrPostTrue, reqPostTrue)
		require.Equal(t, http.StatusOK, rrPostTrue.Code)

		// Verify operation is succeeded
		req, err := http.NewRequest(http.MethodGet, "/webhook/delete/operation", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(webhook.NewWebHookOperationHTTPHandler())
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Contains(t, rr.Body.String(), webhook.OperationResponseStatusOK)

		// Verify is unlocked
		req, err = http.NewRequest(http.MethodGet, "/webhook/delete", nil)
		require.NoError(t, err)
		rr = httptest.NewRecorder()
		handler = http.HandlerFunc(webhook.NewDeleteHTTPHandler())
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
	})
}