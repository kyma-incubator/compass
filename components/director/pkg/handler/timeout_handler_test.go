package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/stretchr/testify/require"
)

func TestHandlerWithTimeout_ReturnsTimeoutMessage(t *testing.T) {
	timeout := time.Millisecond * 100
	h, wait := getStubHandleFunc(t, timeout)
	defer wait()

	handlerWithTimeout, err := handler.WithTimeout(h, timeout)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", &bytes.Buffer{})
	w := httptest.NewRecorder()

	handlerWithTimeout.ServeHTTP(w, req)

	resp := w.Result()
	require.NotNil(t, resp)

	require.Equal(t, handler.HeaderContentTypeValue, resp.Header.Get(handler.HeaderContentTypeKey))

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	actualError := getErrorMessage(t, respBody)

	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	require.Contains(t, actualError, "timed out")
}

func TestHandlerWithTimeout_LogCorrelationID(t *testing.T) {
	timeout := time.Millisecond * 100
	h, wait := getStubHandleFunc(t, timeout)
	defer wait()

	handlerWithTimeout, err := handler.WithTimeout(h, timeout)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", &bytes.Buffer{})
	w := httptest.NewRecorder()

	logger, hook := test.NewNullLogger()

	ctx := log.ContextWithLogger(req.Context(), logrus.NewEntry(logger))
	req = req.WithContext(ctx)

	handlerWithTimeout.ServeHTTP(w, req)

	reqID, ok := hook.LastEntry().Data[correlation.RequestIDHeaderKey]
	require.True(t, ok)

	assert.NotEqual(t, log.Configuration().BootstrapCorrelationID, reqID)
}

func getStubHandleFunc(t *testing.T, timeout time.Duration) (http.HandlerFunc, func()) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	sleepTime := timeout + (time.Millisecond * 10)

	h := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer wg.Done()
		time.Sleep(sleepTime)
		_, err := writer.Write([]byte("test"))
		require.Equal(t, http.ErrHandlerTimeout, err)
	})

	wait := func() {
		wg.Wait()
	}

	return h, wait
}

func getErrorMessage(t *testing.T, data []byte) string {
	var body apperrors.Error
	err := json.Unmarshal(data, &body)
	require.NoError(t, err)
	return body.Message
}
