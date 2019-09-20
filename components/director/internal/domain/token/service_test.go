package token_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/token"
	"github.com/kyma-incubator/compass/components/director/internal/domain/token/automock"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var URL = `http://localhost:3001/graphql`

func TestTokenService_GetOneTimeTokenForRuntime(t *testing.T) {
	runtimeID := "98cb3b05-0f27-43ea-9249-605ac74a6cf0"
	expectedRequest := gcli.NewRequest(fmt.Sprintf(`
		mutation { generateRuntimeToken (runtimeID:"%s")
		  {
			token
		  }
		}`, runtimeID))

	t.Run("Success runtime token", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		cli := &automock.GraphQLClient{}
		expectedToken := "token"

		expected := token.ExternalTokenModel{RuntimeToken: token.ExternalRuntimeToken{Token: expectedToken}}
		input := token.ExternalTokenModel{}
		cli.On("Run", ctx, expectedRequest, &token.ExternalTokenModel{}).
			Run(mockReturnFilledToken(t, input, expected)).Return(nil).Once()
		svc := token.NewTokenService(cli, URL)

		//WHEN
		token, err := svc.GenerateOneTimeToken(ctx, runtimeID, token.RuntimeToken)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, expectedToken, token.Token)
		cli.AssertExpectations(t)
	})

	t.Run("Error - generating failed", func(t *testing.T) {
		ctx := context.TODO()
		cli := &automock.GraphQLClient{}
		testErr := errors.New("test error")
		cli.On("Run", ctx, expectedRequest, &token.ExternalTokenModel{}).
			Return(testErr).Once()
		svc := token.NewTokenService(cli, URL)

		//WHEN
		_, err := svc.GenerateOneTimeToken(ctx, runtimeID, token.RuntimeToken)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		cli.AssertExpectations(t)
	})
}

func TestTokenService_GetOneTimeTokenForApp(t *testing.T) {
	appID := "98cb3b05-0f27-43ea-9249-605ac74a6cf0"
	expectedRequest := gcli.NewRequest(fmt.Sprintf(`
		mutation { generateApplicationToken (appID:"%s")
		  {
			token
		  }
		}`, appID))

	t.Run("Success runtime token", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		cli := &automock.GraphQLClient{}
		expectedToken := "token"

		expected := token.ExternalTokenModel{AppToken: token.ExternalRuntimeToken{Token: expectedToken}}
		input := token.ExternalTokenModel{}
		cli.On("Run", ctx, expectedRequest, &token.ExternalTokenModel{}).
			Run(mockReturnFilledToken(t, input, expected)).Return(nil).Once()
		svc := token.NewTokenService(cli, URL)

		//WHEN
		token, err := svc.GenerateOneTimeToken(ctx, appID, token.ApplicationToken)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, expectedToken, token.Token)
		cli.AssertExpectations(t)
	})

	t.Run("Error - generating failed", func(t *testing.T) {
		ctx := context.TODO()
		cli := &automock.GraphQLClient{}
		testErr := errors.New("test error")
		cli.On("Run", ctx, expectedRequest, &token.ExternalTokenModel{}).
			Return(testErr).Once()
		svc := token.NewTokenService(cli, URL)

		//WHEN
		_, err := svc.GenerateOneTimeToken(ctx, appID, token.ApplicationToken)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		cli.AssertExpectations(t)
	})
}

func mockReturnFilledToken(t *testing.T, input, output token.ExternalTokenModel) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		arg, ok := args.Get(2).(*token.ExternalTokenModel)
		require.True(t, ok)
		require.NotNil(t, arg)
		require.Equal(t, input, *arg)
		*arg = output
	}
}
