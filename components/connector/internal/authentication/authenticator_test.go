package authentication

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens/mocks"
)

const (
	token    = "abcd-efgh"
	clientId = "client-id"
	certHash = "qwertyuiop"
)

func TestAuthenticator_AuthenticateToken(t *testing.T) {

	tokenData := tokens.TokenData{ClientId: clientId, Type: tokens.ApplicationToken}

	t.Run("should authenticate with token", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), ConnectorTokenKey, token)

		tokenSvc := &mocks.Service{}
		tokenSvc.On("Resolve", token).Return(tokenData, nil)

		authenticator := NewAuthenticator(tokenSvc)

		// when
		data, err := authenticator.AuthenticateToken(ctx)

		// then
		require.NoError(t, err)
		assert.Equal(t, tokenData, data)
	})

	t.Run("should return error if token not found in cache", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), ConnectorTokenKey, token)

		tokenSvc := &mocks.Service{}
		tokenSvc.On("Resolve", token).Return(tokens.TokenData{}, apperrors.NotFound("error"))

		authenticator := NewAuthenticator(tokenSvc)

		// when
		data, err := authenticator.AuthenticateToken(ctx)

		// then
		require.Error(t, err)
		assert.Empty(t, data)
	})

	t.Run("should return error if token not found in context", func(t *testing.T) {
		// given
		authenticator := NewAuthenticator(nil)

		// when
		data, err := authenticator.AuthenticateToken(context.Background())

		// then
		require.Error(t, err)
		assert.Empty(t, data)
	})

}

func TestAuthenticator_AuthenticateCertificate(t *testing.T) {

	certificateData := CertificateData{CommonName: clientId, Hash: certHash}

	t.Run("should authenticate with certificate", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), CertificateCommonNameKey, clientId)
		ctx = PutInContext(ctx, CertificateHashKey, certHash)

		authenticator := NewAuthenticator(nil)

		// when
		data, err := authenticator.AuthenticateCertificate(ctx)

		// then
		require.NoError(t, err)
		assert.Equal(t, certificateData, data)
	})

	t.Run("should return error if hash not in context", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), CertificateCommonNameKey, clientId)

		authenticator := NewAuthenticator(nil)

		// when
		data, err := authenticator.AuthenticateCertificate(ctx)

		// then
		require.Error(t, err)
		assert.Empty(t, data)
	})

	t.Run("should return error if common name not in context", func(t *testing.T) {
		// given
		authenticator := NewAuthenticator(nil)

		// when
		data, err := authenticator.AuthenticateCertificate(context.Background())

		// then
		require.Error(t, err)
		assert.Empty(t, data)
	})

}

func TestAuthenticator_AuthenticateTokenOrCertificate(t *testing.T) {

	tokenData := tokens.TokenData{ClientId: clientId, Type: tokens.ApplicationToken}

	t.Run("should authenticate with token", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), ConnectorTokenKey, token)

		tokenSvc := &mocks.Service{}
		tokenSvc.On("Resolve", token).Return(tokenData, nil)

		authenticator := NewAuthenticator(tokenSvc)

		// when
		id, err := authenticator.AuthenticateTokenOrCertificate(ctx)

		// then
		require.NoError(t, err)
		assert.Equal(t, clientId, id)
	})

	t.Run("should authenticate with certificate if no token in context", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), CertificateCommonNameKey, clientId)
		ctx = PutInContext(ctx, CertificateHashKey, certHash)

		authenticator := NewAuthenticator(nil)

		// when
		id, err := authenticator.AuthenticateTokenOrCertificate(ctx)

		// then
		require.NoError(t, err)
		assert.Equal(t, clientId, id)
	})

	t.Run("should authenticate with certificate if token is invalid", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), ConnectorTokenKey, token)
		ctx = PutInContext(ctx, CertificateCommonNameKey, clientId)
		ctx = PutInContext(ctx, CertificateHashKey, certHash)

		tokenSvc := &mocks.Service{}
		tokenSvc.On("Resolve", token).Return(tokens.TokenData{}, apperrors.NotFound("error"))

		authenticator := NewAuthenticator(tokenSvc)

		// when
		id, err := authenticator.AuthenticateTokenOrCertificate(ctx)

		// then
		require.NoError(t, err)
		assert.Equal(t, clientId, id)
	})

	t.Run("should return error if token and cert not provided", func(t *testing.T) {
		// given
		authenticator := NewAuthenticator(nil)

		// when
		data, err := authenticator.AuthenticateTokenOrCertificate(context.Background())

		// then
		require.Error(t, err)
		assert.Empty(t, data)
	})
}
