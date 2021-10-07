package tokens

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	gcliMocks "github.com/kyma-incubator/compass/components/connector/internal/tokens/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	clientId = "client-id"
	token    = "tokenValue"
)

func TestTokenService(t *testing.T) {

	t.Run("should return the token", func(t *testing.T) {
		// GIVEN
		gcliMock := &gcliMocks.GraphQLClient{}
		defer gcliMock.AssertExpectations(t)
		expected := NewTokenResponse(token)
		ctx := context.WithValue(context.Background(), authentication.TenantKey, "tenant")
		gcliMock.On("Run", ctx, mock.Anything, mock.Anything).Run(GenerateTestToken(t, expected)).Return(nil).Once()
		tokenService := NewTokenService(gcliMock)
		// WHEN
		actualToken, appError := tokenService.GetToken(ctx, clientId, "Application")
		// THEN
		require.NoError(t, appError)
		assert.Equal(t, token, actualToken)
	})

	t.Run("should return error when token not found", func(t *testing.T) {
		// GIVEN
		gcliMock := &gcliMocks.GraphQLClient{}
		defer gcliMock.AssertExpectations(t)
		err := errors.New("could not get the token")
		ctx := context.WithValue(context.Background(), authentication.TenantKey, "tenant")
		gcliMock.On("Run", ctx, mock.Anything, mock.Anything).Return(err)
		tokenService := NewTokenService(gcliMock)
		// WHEN
		actualToken, appError := tokenService.GetToken(ctx, clientId, "Application")
		// THEN
		require.Error(t, appError)
		assert.Equal(t, "", actualToken)
	})
}

func GenerateTestToken(t *testing.T, generated TokenResponse) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		arg, ok := args.Get(2).(*TokenResponse)
		require.True(t, ok)
		*arg = generated
	}
}
