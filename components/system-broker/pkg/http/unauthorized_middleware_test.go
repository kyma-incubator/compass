package http_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	httputil "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/stretchr/testify/require"
)

func TestUnauthorizedMiddlewareWhenErrorOccursShouldReturnInternalServerError(t *testing.T) {
	const authorizedString = "insufficient scopes provided"

	testUrl, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	request := &http.Request{
		Method: http.MethodPost,
		URL:    testUrl,
		Header: map[string][]string{},
	}

	preMiddleware := preMiddleware(t, http.StatusInternalServerError, "test")
	unauthorizedMiddleware := httputil.UnauthorizedMiddleware(authorizedString)
	postMiddleware := postMiddleware(t, http.StatusInternalServerError, "test")

	preMiddleware(unauthorizedMiddleware(postMiddleware(nil))).ServeHTTP(nil, request)
}

func TestUnauthorizedMiddlewareWhenUnauthorizedShouldReturnUnauthorized(t *testing.T) {
	const directorUnauthorizedErrorString = "insufficient scopes provided"

	testUrl, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)
	request := &http.Request{
		Method: http.MethodPost,
		URL:    testUrl,
		Header: map[string][]string{},
	}

	preMiddleware := preMiddleware(t, http.StatusUnauthorized, "unauthorized: insufficient scopes")
	unauthorizedMiddleware := httputil.UnauthorizedMiddleware(directorUnauthorizedErrorString)
	postMiddleware := postMiddleware(t, http.StatusInternalServerError, fmt.Sprintf(`{"description":"%s"}`, directorUnauthorizedErrorString))

	preMiddleware(unauthorizedMiddleware(postMiddleware(nil))).ServeHTTP(nil, request)
}

func preMiddleware(t *testing.T, statusCode int, body string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			recorder := httptest.NewRecorder()

			next.ServeHTTP(recorder, r)

			require.Equal(t, statusCode, recorder.Code)
			require.Equal(t, recorder.Header().Get("key"), "val")
			require.Contains(t, recorder.Body.String(), body)
		})
	}
}

func postMiddleware(t *testing.T, statusCode int, body string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(statusCode)
			rw.Header().Add("key", "val")
			_, err := rw.Write([]byte(body))

			require.NoError(t, err)
		})
	}
}
