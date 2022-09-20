package correlation_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/stretchr/testify/assert"
)

const expectedRequestID = "123"

func TestContextEnrichMiddleware_AttachCorrelationIDToContext(t *testing.T) {
	// given
	handler := correlation.AttachCorrelationIDToContext()

	t.Run("when x-request-id header is present it's added as correlation header to the request context and headers", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headersFromContext, ok := r.Context().Value(correlation.HeadersContextKey).(correlation.Headers)
			assert.True(t, ok)

			actual, ok := headersFromContext[correlation.RequestIDHeaderKey]
			assert.True(t, ok)
			assert.Equal(t, actual, expectedRequestID)

			headerFromRequest := r.Header.Get(correlation.RequestIDHeaderKey)
			assert.Equal(t, headerFromRequest, expectedRequestID)

			correlationID := correlation.CorrelationIDFromContext(r.Context())
			assert.Equal(t, correlationID, expectedRequestID)
		})

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(correlation.RequestIDHeaderKey, expectedRequestID)

		handler(nextHandler).ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("when no identifying headers are present a new correlation header is added to the request context and headers", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headersFromContext, ok := r.Context().Value(correlation.HeadersContextKey).(correlation.Headers)
			assert.True(t, ok)

			requestIDHeader, ok := headersFromContext[correlation.RequestIDHeaderKey]
			assert.True(t, ok)
			assert.NotEmpty(t, requestIDHeader)

			headerFromRequest := r.Header.Get(correlation.RequestIDHeaderKey)
			assert.NotEmpty(t, headerFromRequest)

			correlationID := correlation.CorrelationIDFromContext(r.Context())
			assert.NotEmpty(t, correlationID)
		})

		req := httptest.NewRequest("GET", "/", nil)
		handler(nextHandler).ServeHTTP(httptest.NewRecorder(), req)
	})
}

func TestContextEnrichMiddleware_HeadersForRequest(t *testing.T) {
	// given
	headerKeys := []string{"x-request-id", "x-b3-traceid", "x-b3-spanid", "x-b3-parentspanid", "x-b3-sampled", "x-b3-flags", "b3"}

	for _, header := range headerKeys {
		t.Run(fmt.Sprintf("returns %s when %s header is present", header, header), func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set(correlation.RequestIDHeaderKey, expectedRequestID)

			headers := correlation.HeadersForRequest(req)
			actualRequestID, ok := headers[correlation.RequestIDHeaderKey]
			assert.True(t, ok)
			assert.Equal(t, expectedRequestID, actualRequestID)
		})
	}
}

func TestContextEnrichMiddleware_HeadersForRequest_WellKnownCorrelationIDsAreAddedToRequestIfPresentInContextHeaders(t *testing.T) {
	wellKnownHeaderKey := "x-b3-traceid"
	wellKnownHeaderValue := "35b74672-9f48-4361-8f47-408832bd5a25"

	ctx := correlation.SaveCorrelationIDHeaderToContext(context.Background(), &wellKnownHeaderKey, &wellKnownHeaderValue)

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(ctx)

	headers := correlation.HeadersForRequest(req)
	actualRequestID, ok := headers[correlation.RequestIDHeaderKey]
	assert.True(t, ok)
	assert.NotEmpty(t, actualRequestID)

	actualWellKnownRequestIDFromMap, ok := headers[wellKnownHeaderKey]
	assert.True(t, ok)
	assert.Equal(t, wellKnownHeaderValue, actualWellKnownRequestIDFromMap)

	wellKnownHeaderKeyTitleCase := "X-B3-Traceid"
	actualWellKnownRequestIDFromRequest := req.Header[wellKnownHeaderKeyTitleCase][0]
	assert.True(t, ok)
	assert.Equal(t, wellKnownHeaderValue, actualWellKnownRequestIDFromRequest)
}

func TestContextEnrichMiddleware_HeadersForRequest_AdditionalContextHeadersAreAddedToRequest(t *testing.T) {
	headerKey := "X-Additional-Request-Id"
	headerValue := "35b74672-9f48-4361-8f47-408832bd5a25"

	ctx := correlation.SaveCorrelationIDHeaderToContext(context.Background(), &headerKey, &headerValue)

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(ctx)

	headers := correlation.HeadersForRequest(req)
	actualRequestID, ok := headers[correlation.RequestIDHeaderKey]
	assert.True(t, ok)
	assert.NotEmpty(t, actualRequestID)

	actualAdditionalRequestID, ok := headers[headerKey]
	assert.True(t, ok)
	assert.Equal(t, headerValue, actualAdditionalRequestID)

	actualAdditionalRequestIDFromRequest := req.Header[headerKey][0]
	assert.True(t, ok)
	assert.Equal(t, headerValue, actualAdditionalRequestIDFromRequest)
}
