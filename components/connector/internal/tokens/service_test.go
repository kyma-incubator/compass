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
)

func TestTokenService(t *testing.T) {

	t.Run("should return the token", func(t *testing.T) {

		gcliMock := &gcliMocks.GraphQLClient{}
		expected := make(map[string]map[string]interface{})
		expected["generateCSRToken"] = make(map[string]interface{})
		expected["generateCSRToken"]["token"] = "tokenValue"
		gcliMock.On("Run", context.Background(), mock.Anything, mock.Anything).Run(generateToken(t, expected)).Return(nil).Once()
		tokenService := NewTokenService(gcliMock)

		token, appError := tokenService.GetToken(context.Background(), clientId)
		assert.Equal(t, "tokenValue", token)
		require.NoError(t, appError)

	})

	t.Run("should return error when token not found", func(t *testing.T) {

		gcliMock := &gcliMocks.GraphQLClient{}
		err := errors.New("could not get the token")
		gcliMock.On("Run", context.Background(), mock.Anything, mock.Anything).Return(err)
		tokenService := NewTokenService(gcliMock)

		_, appError := tokenService.GetToken(context.Background(), clientId)
		require.Error(t, appError)
	})
}

func generateToken(t *testing.T, generated map[string]map[string]interface{}) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		arg, ok := args.Get(2).(*map[string]map[string]interface{})
		require.True(t, ok)
		*arg = generated
	}
}
