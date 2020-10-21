package handler_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/stretchr/testify/require"
)

func TestHandlerWithTimeout_ReturnsTimeoutMessage(t *testing.T) {
	timeout := time.Millisecond * 100
	h := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(time.Second)
		writer.WriteHeader(http.StatusOK)
	})

	handlerWithTimeout, err := handler.WithTimeout(h, timeout)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", &bytes.Buffer{})
	w := httptest.NewRecorder()

	handlerWithTimeout.ServeHTTP(w, req)

	resp := w.Result()

	require.NotNil(t, resp)

	require.NotNil(t, resp)
	require.Equal(t, handler.HeaderContentTypeValue, resp.Header.Get(handler.HeaderContentTypeKey))

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	actualError := getErrorMessage(t, respBody)

	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	require.Contains(t, actualError, "timed out")
}

func getErrorMessage(t *testing.T, data []byte) string {
	var body apperrors.Error
	err := json.Unmarshal(data, &body)
	require.NoError(t, err)
	return body.Message
}
