package handler_test

import (
	"bytes"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/handler"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/res"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
	require.Equal(t, res.HeaderContentTypeValue, resp.Header.Get(res.HeaderContentTypeKey))

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	actualError := getErrorMessage(t, respBody)

	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	require.Contains(t, actualError, "timed out")
}

func getErrorMessage(t *testing.T, data []byte) string {
	var body res.ErrorResponse
	err := json.Unmarshal(data, &body)
	require.NoError(t, err)
	return body.Error
}
