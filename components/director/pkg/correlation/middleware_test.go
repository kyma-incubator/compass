package correlation_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/stretchr/testify/assert"
)

func TestContextEnrichMiddleware_AttachCorrelationIDToContext(t *testing.T) {
	// given
	handler := correlation.AttachCorrelationIDToContext()

	t.Run("when x-request-id header is present it's added as correlation header to the request context and headers", func(t *testing.T) {
		expectedCorrelationID := "123"

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headersFromContext := r.Context().Value(correlation.HeadersContextKey).(correlation.Headers)
			actual, ok := headersFromContext[correlation.RequestIDHeaderKey]
			assert.True(t, ok)
			assert.Equal(t, actual, expectedCorrelationID)

			headerFromRequest := r.Header.Get(correlation.RequestIDHeaderKey)
			assert.Equal(t, headerFromRequest, expectedCorrelationID)
		})

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(correlation.RequestIDHeaderKey, expectedCorrelationID)

		handler(nextHandler).ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("when no identifying headers are present a new correlation header is added to the request context and headers", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers := correlation.HeadersForRequest(r)
			requestIDHeader, ok := headers[correlation.RequestIDHeaderKey]
			assert.True(t, ok)
			assert.NotEmpty(t, requestIDHeader)

			headerFromRequest := r.Header.Get(correlation.RequestIDHeaderKey)
			assert.NotEmpty(t, headerFromRequest)
		})

		req := httptest.NewRequest("GET", "/", nil)
		handler(nextHandler).ServeHTTP(httptest.NewRecorder(), req)
	})
}
