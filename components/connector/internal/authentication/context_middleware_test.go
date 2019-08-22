package authentication_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication/mocks"
)

func TestAuthContextMiddleware_PropagateAuthentication(t *testing.T) {

	certificateCommonName := "certificateCommonName"
	certificateHash := "qwertyuiop"
	connectorToken := "connector-token"

	t.Run("should put authentication to context", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := authentication.GetStringFromContext(r.Context(), authentication.ConnectorTokenKey)
			require.NoError(t, err)
			commonName, err := authentication.GetStringFromContext(r.Context(), authentication.CertificateCommonNameKey)
			require.NoError(t, err)
			certHash, err := authentication.GetStringFromContext(r.Context(), authentication.CertificateHashKey)
			require.NoError(t, err)

			assert.Equal(t, connectorToken, token)
			assert.Equal(t, certificateCommonName, commonName)
			assert.Equal(t, certificateHash, certHash)

			w.WriteHeader(http.StatusOK)
		})

		request, err := http.NewRequest(http.MethodGet, "", nil)
		require.NoError(t, err)
		request.Header.Add(authentication.ClientCertHeader, "CommonName=certificateCommonName,Hash=qwertyuiop")
		request.Header.Add(authentication.ConnectorTokenHeader, connectorToken)
		rr := httptest.NewRecorder()

		headerParser := &mocks.CertificateHeaderParser{}
		headerParser.On("GetCertificateData", mock.AnythingOfType("*http.Request")).Return(certificateCommonName, certificateHash, true)

		authContextMiddleware := authentication.NewAuthenticationContextMiddleware(headerParser)

		// when
		handlerWithMiddleware := authContextMiddleware.PropagateAuthentication(handler)
		handlerWithMiddleware.ServeHTTP(rr, request)
	})

	t.Run("should not put cert context if client certificate header is empty", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := authentication.GetStringFromContext(r.Context(), authentication.ConnectorTokenKey)
			require.NoError(t, err)
			commonName, err := authentication.GetStringFromContext(r.Context(), authentication.CertificateCommonNameKey)
			require.Error(t, err)
			certHash, err := authentication.GetStringFromContext(r.Context(), authentication.CertificateHashKey)
			require.Error(t, err)

			assert.Equal(t, connectorToken, token)
			assert.Empty(t, commonName)
			assert.Empty(t, certHash)

			w.WriteHeader(http.StatusOK)
		})

		request, err := http.NewRequest(http.MethodGet, "", nil)
		require.NoError(t, err)
		request.Header.Add(authentication.ConnectorTokenHeader, connectorToken)
		rr := httptest.NewRecorder()

		headerParser := &mocks.CertificateHeaderParser{}
		headerParser.On("GetCertificateData", mock.AnythingOfType("*http.Request")).Return("", "", false)

		authContextMiddleware := authentication.NewAuthenticationContextMiddleware(headerParser)

		// when
		handlerWithMiddleware := authContextMiddleware.PropagateAuthentication(handler)
		handlerWithMiddleware.ServeHTTP(rr, request)
	})

}
