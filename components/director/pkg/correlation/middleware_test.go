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
	mid := correlation.NewContextEnrichMiddleware()

	t.Run("when x-request-id header is present it's added as correlation ID to the request context", func(t *testing.T) {
		expectedCorrelationID := "123"

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actual := correlation.IDFromContext(r.Context())
			assert.Equal(t,actual, expectedCorrelationID)
		})

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(correlation.RequestIDHeaderKey, expectedCorrelationID)

		mid.AttachCorrelationIDToContext(nextHandler).ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("when no identifying headers are present a new correlation ID is added to the request context", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actual := correlation.IDFromContext(r.Context())
			assert.NotEmpty(t, actual)
		})

		req := httptest.NewRequest("GET", "/", nil)
		mid.AttachCorrelationIDToContext(nextHandler).ServeHTTP(httptest.NewRecorder(), req)
	})
}
