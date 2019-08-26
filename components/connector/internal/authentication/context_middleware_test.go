package authentication_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestAuthContextMiddleware_PropagateAuthentication(t *testing.T) {

	connectorToken := "connector-token"

	t.Run("should put authentication to context", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := authentication.GetStringFromContext(r.Context(), authentication.ConnectorTokenKey)
			require.NoError(t, err)

			assert.Equal(t, connectorToken, token)

			w.WriteHeader(http.StatusOK)
		})

		request, err := http.NewRequest(http.MethodGet, "", nil)
		require.NoError(t, err)
		request.Header.Add(authentication.ClientCertHeader, "CommonName=certificateCommonName,Hash=qwertyuiop")
		request.Header.Add(authentication.ConnectorTokenHeader, connectorToken)
		rr := httptest.NewRecorder()

		authContextMiddleware := authentication.NewAuthenticationContextMiddleware()

		// when
		handlerWithMiddleware := authContextMiddleware.PropagateAuthentication(handler)
		handlerWithMiddleware.ServeHTTP(rr, request)
	})

}
