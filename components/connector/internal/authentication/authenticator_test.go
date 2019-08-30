package authentication

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	clientId = "client-id"
	certHash = "qwertyuiop"
)

func TestAuthenticator_AuthenticateToken(t *testing.T) {

	t.Run("should authenticate with token", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), ClientIdFromTokenKey, clientId)

		authenticator := NewAuthenticator()

		// when
		receivedId, err := authenticator.AuthenticateToken(ctx)

		// then
		require.NoError(t, err)
		assert.Equal(t, clientId, receivedId)
	})

	t.Run("should return error if client id is empty", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), ClientIdFromTokenKey, "")

		authenticator := NewAuthenticator()

		// when
		receivedId, err := authenticator.AuthenticateToken(ctx)

		// then
		require.Error(t, err)
		require.Empty(t, receivedId)
	})

	t.Run("should return error if token not found in context", func(t *testing.T) {
		// given
		authenticator := NewAuthenticator()

		// when
		data, err := authenticator.AuthenticateToken(context.Background())

		// then
		require.Error(t, err)
		assert.Empty(t, data)
	})

}

func TestAuthenticator_AuthenticateCertificate(t *testing.T) {

	t.Run("should authenticate with certificate", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), ClientIdFromCertificateKey, clientId)
		ctx = PutInContext(ctx, ClientCertificateHash, certHash)

		authenticator := NewAuthenticator()

		// when
		id, err := authenticator.AuthenticateCertificate(ctx)

		// then
		require.NoError(t, err)
		assert.Equal(t, clientId, id)
	})

	t.Run("should return error if client id is empty", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), ClientIdFromCertificateKey, "")
		ctx = PutInContext(ctx, ClientCertificateHash, certHash)

		authenticator := NewAuthenticator()

		// when
		id, err := authenticator.AuthenticateCertificate(ctx)

		// then
		require.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("should return error if hash not in context", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), ClientIdFromCertificateKey, clientId)

		authenticator := NewAuthenticator()

		// when
		id, err := authenticator.AuthenticateCertificate(ctx)

		// then
		require.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("should return error if client id not found in context", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), ClientCertificateHash, certHash)

		authenticator := NewAuthenticator()

		// when
		id, err := authenticator.AuthenticateCertificate(ctx)

		// then
		require.Error(t, err)
		assert.Empty(t, id)
	})

}

func TestAuthenticator_AuthenticateTokenOrCertificate(t *testing.T) {

	t.Run("should authenticate with token", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), ClientIdFromTokenKey, clientId)

		authenticator := NewAuthenticator()

		// when
		receivedId, err := authenticator.AuthenticateTokenOrCertificate(ctx)

		// then
		require.NoError(t, err)
		assert.Equal(t, clientId, receivedId)
	})

	t.Run("should authenticate with certificate", func(t *testing.T) {
		// given
		ctx := PutInContext(context.Background(), ClientIdFromCertificateKey, clientId)
		ctx = PutInContext(ctx, ClientCertificateHash, certHash)

		authenticator := NewAuthenticator()

		// when
		id, err := authenticator.AuthenticateTokenOrCertificate(ctx)

		// then
		require.NoError(t, err)
		assert.Equal(t, clientId, id)
	})

	t.Run("should return error if failed to authenticate with token and cert", func(t *testing.T) {
		// given
		authenticator := NewAuthenticator()

		// when
		id, err := authenticator.AuthenticateTokenOrCertificate(context.Background())

		// then
		require.Error(t, err)
		assert.Empty(t, id)
	})
}
