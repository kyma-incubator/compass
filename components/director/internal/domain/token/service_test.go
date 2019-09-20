package token

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/token/automock"
	"github.com/kyma-incubator/compass/components/director/internal/graphql_client"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
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

	t.Run("Success", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		cli := &automock.GraphQLClient{}
		expectedToken := "token"

		expected := ExternalTokenModel{GenerateRuntimeToken: ExternalRuntimeToken{Token: expectedToken}}
		input := ExternalTokenModel{}
		cli.On("Run", ctx, expectedRequest, &ExternalTokenModel{}).
			Run(mockReturnFilledToken(t, input, expected)).Return(nil).Once()
		svc := NewTokenService(cli)

		//WHEN
		token, err := svc.getOneTimeToken(ctx, runtimeID)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, expectedToken, token)
		cli.AssertExpectations(t)
	})

	t.Run("Error - generating failed", func(t *testing.T) {
		ctx := context.TODO()
		cli := &automock.GraphQLClient{}
		testErr := errors.New("test error")
		cli.On("Run", ctx, expectedRequest, &ExternalTokenModel{}).
			Return(testErr).Once()
		svc := NewTokenService(cli)

		//WHEN
		_, err := svc.getOneTimeToken(ctx, runtimeID)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		cli.AssertExpectations(t)
	})
}

func mockReturnFilledToken(t *testing.T, input, output ExternalTokenModel) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		arg, ok := args.Get(2).(*ExternalTokenModel)
		require.True(t, ok)
		require.NotNil(t, arg)
		require.Equal(t, input, *arg)
		*arg = output
	}
}

//TODO: remove those things below
func TestTokenService_GetOneTimeTokenForRuntime2(t *testing.T) {
	runtimeID := "runtime_ID"
	ctx := context.TODO()

	cli := graphql_client.NewGraphQLClient(URL)

	svc := NewTokenService(cli)
	token, err := svc.getOneTimeToken(ctx, runtimeID)

	require.NoError(t, err)
	fmt.Println(token)
}
