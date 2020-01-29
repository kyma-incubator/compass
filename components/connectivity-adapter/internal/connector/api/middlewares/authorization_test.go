package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware_ExtractHeaders(t *testing.T) {

	getSuccessHandlerFunc := func(headers map[string]string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headersMap, err := GetAuthHeadersFromContext(r.Context(), AuthorizationHeadersKey)
			require.NoError(t, err)

			for key, value := range headers {
				assert.Equal(t, value, headersMap[key])
			}
		})
	}

	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Fail(t, "Handler must not be called")
	})

	headersFromToken := map[string]string{
		oathkeeper.ClientIdFromTokenHeader: "myapp",
	}

	headersFromCertificate := map[string]string{
		oathkeeper.ClientIdFromCertificateHeader: "myapp",
		oathkeeper.ClientCertificateHashHeader:   "certificate hash",
	}

	testCases := []struct {
		description string
		headers     map[string]string
		handler     http.Handler
		status      int
	}{
		{
			description: "Should extract Client-Id-From-Token header",
			headers:     headersFromToken,
			handler:     getSuccessHandlerFunc(headersFromToken),
			status:      http.StatusOK,
		},
		{
			description: "Should extract Client-Id-From-Certificate header",
			headers:     headersFromCertificate,
			handler:     getSuccessHandlerFunc(headersFromCertificate),
			status:      http.StatusOK,
		},
		{
			description: "Should return Forbidden when Client-Id-From-Token, Client-Id-From-Certificate, Client-Certificate-Hash headers are missing",
			headers:     map[string]string{},
			handler:     errorHandler,
			status:      http.StatusForbidden,
		},
		{
			description: "Should return Forbidden when Client-Certificate-Hash header is missing",
			headers: map[string]string{
				oathkeeper.ClientIdFromCertificateHeader: "myapp",
			},
			handler: errorHandler,
			status:  http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// given
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				headersMap, err := GetAuthHeadersFromContext(r.Context(), AuthorizationHeadersKey)
				require.NoError(t, err)

				for key, value := range tc.headers {
					assert.Equal(t, value, headersMap[key])
				}
			})

			r := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "http://www.someurl.com/get", strings.NewReader(""))

			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}

			middleware := NewAuthorizationMiddleware()
			handlerWithMiddleware := middleware.GetAuthorizationHeaders(handler)

			// when
			handlerWithMiddleware.ServeHTTP(r, req)

			// then
			assert.Equal(t, tc.status, r.Code)
		})
	}
}
