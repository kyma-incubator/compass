package log_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestRequestLoggerGeneratesCorrelationIDWhenNotFoundInHeaders(t *testing.T) {
	response := httptest.NewRecorder()

	testURL, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	request := &http.Request{
		Method: http.MethodPost,
		URL:    testURL,
		Header: map[string][]string{},
	}

	handler := log.RequestLogger()
	handler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		entry := log.C(request.Context())

		correlationIDFromLogger, exists := entry.Data[log.FieldRequestID]
		require.True(t, exists)
		require.NotEmpty(t, correlationIDFromLogger)
	})).ServeHTTP(response, request)
}

func TestRequestLoggerUseCorrelationIDFromHeaderIfProvided(t *testing.T) {
	correlationID := "test-correlation-id"
	response := httptest.NewRecorder()

	testURL, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	request := &http.Request{
		Method: http.MethodPost,
		URL:    testURL,
		Header: map[string][]string{},
	}
	request.Header.Set("x-request-id", correlationID)
	request.Header.Set("X-Real-IP", "127.0.0.1")

	handler := log.RequestLogger()
	handler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		entry := log.C(request.Context())

		correlationIDFromLogger, exists := entry.Data[log.FieldRequestID]
		require.True(t, exists)
		require.Equal(t, correlationID, correlationIDFromLogger)
	})).ServeHTTP(response, request)
}

func TestRequestLoggerWithMDC(t *testing.T) {
	response := httptest.NewRecorder()
	testURL, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	request := &http.Request{
		Method: http.MethodPost,
		URL:    testURL,
		Header: map[string][]string{},
	}

	oldLogger := logrus.StandardLogger().Out
	buf := bytes.Buffer{}
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(oldLogger)

	handler := log.RequestLogger()
	handler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		hasMdc := false
		if mdc := log.MdcFromContext(request.Context()); nil != mdc {
			hasMdc = true
			mdc.Set("test", "test")
		}
		require.True(t, hasMdc, "There is no MDC in the request context")

		//remove the "Started handling ..." line
		buf.Reset()
	})).ServeHTTP(response, request)

	logLine := buf.String()
	hasMdcMessage := strings.Contains(logLine, "test=test")
	require.True(t, hasMdcMessage, "The log line does not contain the MDC content: %v", logLine)
}

func TestRequestLoggerDebugPaths(t *testing.T) {
	response := httptest.NewRecorder()
	testURL, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	request := &http.Request{
		Method: http.MethodPost,
		URL:    testURL,
		Header: map[string][]string{},
	}

	oldLogger := logrus.StandardLogger().Out
	buf := bytes.Buffer{}
	logrus.SetOutput(&buf)
	logrus.SetLevel(logrus.DebugLevel)
	defer logrus.SetOutput(oldLogger)

	emptyHandlerFunc := func(writer http.ResponseWriter, request *http.Request) {}

	const debugPath = "/healthz"
	handler := log.RequestLogger(debugPath)
	handler(http.HandlerFunc(emptyHandlerFunc)).ServeHTTP(response, request)

	logs := buf.String()
	require.Contains(t, logs, `level=info msg="Started handling request..."`)
	require.Contains(t, logs, `level=info msg="Finished handling request..."`)

	buf.Reset()
	request.URL.Path = debugPath
	handler(http.HandlerFunc(emptyHandlerFunc)).ServeHTTP(response, request)

	logs = buf.String()
	require.Contains(t, logs, `level=debug msg="Started handling request..."`)
	require.Contains(t, logs, `level=debug msg="Finished handling request..."`)
}
