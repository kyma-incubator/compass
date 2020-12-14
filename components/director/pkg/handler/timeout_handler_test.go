package handler_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/stretchr/testify/require"
)

func TestHandlerWithTimeout_ReturnsTimeoutMessage(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	timeout := time.Millisecond * 100
	h := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer wg.Done()
		time.Sleep(time.Millisecond * 110)
		_, err := writer.Write([]byte("test"))
		require.Equal(t, http.ErrHandlerTimeout, err)
	})

	handlerWithTimeout, err := handler.WithTimeout(h, timeout)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", &bytes.Buffer{})
	w := httptest.NewRecorder()

	handlerWithTimeout.ServeHTTP(w, req)

	resp := w.Result()
	require.NotNil(t, resp)

	require.Equal(t, handler.HeaderContentTypeValue, resp.Header.Get(handler.HeaderContentTypeKey))

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	actualError := getErrorMessage(t, respBody)

	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	require.Contains(t, actualError, "timed out")

	wg.Wait()
}

func getErrorMessage(t *testing.T, data []byte) string {
	var body apperrors.Error
	err := json.Unmarshal(data, &body)
	require.NoError(t, err)
	return body.Message
}
