package authentication

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestAuthContextMiddleware_PropagateAuthentication(t *testing.T) {

	t.Run("should put authentication to context", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idFromToken, err := GetStringFromContext(r.Context(), ClientIdFromTokenKey)
			require.NoError(t, err)
			assert.Equal(t, clientId, idFromToken)

			idFromCert, err := GetStringFromContext(r.Context(), ClientIdFromCertificateKey)
			require.NoError(t, err)
			assert.Equal(t, clientId, idFromCert)

			hash, err := GetStringFromContext(r.Context(), ClientCertificateHashKey)
			require.NoError(t, err)
			assert.Equal(t, certHash, hash)

			w.WriteHeader(http.StatusOK)
		})

		request, err := http.NewRequest(http.MethodGet, "", nil)
		require.NoError(t, err)

		request.Header.Add(oathkeeper.ClientIdFromTokenHeader, clientId)
		request.Header.Add(oathkeeper.ClientIdFromCertificateHeader, clientId)
		request.Header.Add(oathkeeper.ClientCertificateHashHeader, certHash)
		rr := httptest.NewRecorder()

		authContextMiddleware := NewAuthenticationContextMiddleware()

		// when
		handlerWithMiddleware := authContextMiddleware.PropagateAuthentication(handler)
		handlerWithMiddleware.ServeHTTP(rr, request)
	})

	t.Run("should read id token if such is provided", func(t *testing.T) {
		expectedTenant := "tenant"
		expectedConsumerType := "consumer"
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, err := GetStringFromContext(r.Context(), TenantKey)
			require.NoError(t, err)
			assert.Equal(t, expectedTenant, tenant)

			consumerType, err := GetStringFromContext(r.Context(), ConsumerType)
			require.NoError(t, err)
			assert.Equal(t, expectedConsumerType, consumerType)

			w.WriteHeader(http.StatusOK)
		})

		request, err := http.NewRequest(http.MethodGet, "", nil)
		require.NoError(t, err)
		token := `eyJubyI6Im5vIiwiYWxnIjoiUlMyNTYifQ==.eyJ0ZW5hbnQiOiJ7XCJjb25zdW1lclRlbmFudFwiOlwidGVuYW50XCJ9IiwgImNvbnN1bWVyVHlwZSI6ImNvbnN1bWVyIn0=.eyJubyI6Im5vIn0=`
		request.Header.Add("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		authContextMiddleware := NewAuthenticationContextMiddleware()

		// when
		handlerWithMiddleware := authContextMiddleware.PropagateAuthentication(handler)
		handlerWithMiddleware.ServeHTTP(rr, request)
		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	})
}
