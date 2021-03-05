package tokens

import (
	"context"
	"errors"
	"testing"

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
		expected := NewCSRTokenResponse(token)
		gcliMock.On("Run", context.Background(), mock.Anything, mock.Anything).Run(GenerateTestToken(expected)).Return(nil).Once()
		tokenService := NewTokenService(gcliMock)
		// WHEN
		actualToken, appError := tokenService.GetToken(context.Background(), clientId)
		// THEN
		assert.Equal(t, token, actualToken)
		require.NoError(t, appError)
	})

	t.Run("should return error when token not found", func(t *testing.T) {
		// GIVEN
		gcliMock := &gcliMocks.GraphQLClient{}
		err := errors.New("could not get the token")
		gcliMock.On("Run", context.Background(), mock.Anything, mock.Anything).Return(err)
		tokenService := NewTokenService(gcliMock)
		// WHEN
		actualToken, appError := tokenService.GetToken(context.Background(), clientId)
		// THEN
		require.Error(t, appError)
		assert.Equal(t, "", actualToken)
	})
}
