package authentication

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/oathkeeper"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestAuthContextMiddleware_PropagateAuthentication(t *testing.T) {

	tokenType := "Application"

	t.Run("should put authentication to context", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idFromToken, err := GetStringFromContext(r.Context(), ClientIdFromTokenKey)
			require.NoError(t, err)
			assert.Equal(t, clientId, idFromToken)

			idFromCert, err := GetStringFromContext(r.Context(), ClientIdFromCertificateKey)
			require.NoError(t, err)
			assert.Equal(t, clientId, idFromCert)

			hash, err := GetStringFromContext(r.Context(), ClientCertificateHash)
			require.NoError(t, err)
			assert.Equal(t, certHash, hash)

			tokenT, err := GetStringFromContext(r.Context(), TokenTypeKey)
			require.NoError(t, err)
			assert.Equal(t, tokenType, tokenT)

			w.WriteHeader(http.StatusOK)
		})

		request, err := http.NewRequest(http.MethodGet, "", nil)
		require.NoError(t, err)

		request.Header.Add(oathkeeper.ClientIdFromTokenHeader, clientId)
		request.Header.Add(oathkeeper.ClientIdFromCertificateHeader, clientId)
		request.Header.Add(oathkeeper.TokenTypeHeader, tokenType)
		request.Header.Add(oathkeeper.ClientCertificateHashHeader, certHash)
		rr := httptest.NewRecorder()

		authContextMiddleware := NewAuthenticationContextMiddleware()

		// when
		handlerWithMiddleware := authContextMiddleware.PropagateAuthentication(handler)
		handlerWithMiddleware.ServeHTTP(rr, request)
	})
}
