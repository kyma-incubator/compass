package authentication_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"

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
		ctx := authentication.PutInContext(context.Background(), authentication.ConnectorTokenKey, token)

		tokenSvc := &mocks.Service{}
		tokenSvc.On("Resolve", token).Return(tokenData, nil)

		authenticator := authentication.NewAuthenticator(tokenSvc)

		// when
		data, err := authenticator.AuthenticateToken(ctx)

		// then
		require.NoError(t, err)
		assert.Equal(t, tokenData, data)
	})

	t.Run("should return error if token not found in cache", func(t *testing.T) {
		// given
		ctx := authentication.PutInContext(context.Background(), authentication.ConnectorTokenKey, token)

		tokenSvc := &mocks.Service{}
		tokenSvc.On("Resolve", token).Return(tokens.TokenData{}, apperrors.NotFound("error"))

		authenticator := authentication.NewAuthenticator(tokenSvc)

		// when
		data, err := authenticator.AuthenticateToken(ctx)

		// then
		require.Error(t, err)
		assert.Empty(t, data)
	})

	t.Run("should return error if token not found in context", func(t *testing.T) {
		// given
		authenticator := authentication.NewAuthenticator(nil)

		// when
		data, err := authenticator.AuthenticateToken(context.Background())

		// then
		require.Error(t, err)
		assert.Empty(t, data)
	})

}
