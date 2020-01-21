package middlewares

import (
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_ClientFromTokenMiddleware(t *testing.T) {

	t.Run("Should extract Client-Id-From-Token header", func(t *testing.T) {
		// given
		clientIdFromToken := "myapp"

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value, err := GetStringFromContext(r.Context(), ClientIdFromTokenKey)
			require.NoError(t, err)

			assert.Equal(t, value, clientIdFromToken)
		})

		r := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://www.someurl.com/get", strings.NewReader(""))
		req.Header.Set(oathkeeper.ClientIdFromTokenHeader, clientIdFromToken)

		middleware := NewClientFromTokenMiddleware()
		handlerWithMiddleware := middleware.GetClientIdFromToken(handler)

		// when
		handlerWithMiddleware.ServeHTTP(r, req)

		// then
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("Should return Forbidden when Client-Id-From-Token header is missing", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Fail(t, "Handler must not be called")
		})

		r := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://www.someurl.com/get", strings.NewReader(""))

		middleware := NewClientFromTokenMiddleware()
		handlerWithMiddleware := middleware.GetClientIdFromToken(handler)

		// when
		handlerWithMiddleware.ServeHTTP(r, req)

		// then
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
