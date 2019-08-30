package authentication_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	token    = "abcd-efgh"
	clientId = "client-id"
	certHash = "qwertyuiop"
)

func TestAuthenticator_AuthenticateToken(t *testing.T) {

	t.Run("should authenticate with token", func(t *testing.T) {
		// given
		ctx := authentication.PutInContext(context.Background(), authentication.ClientIdFromTokenKey, clientId)

		authenticator := authentication.NewAuthenticator()

		// when
		receivedId, err := authenticator.AuthenticateToken(ctx)

		// then
		require.NoError(t, err)
		assert.Equal(t, clientId, receivedId)
	})

	t.Run("should return error if token not found in context", func(t *testing.T) {
		// given
		authenticator := authentication.NewAuthenticator()

		// when
		data, err := authenticator.AuthenticateToken(context.Background())

		// then
		require.Error(t, err)
		assert.Empty(t, data)
	})

}
