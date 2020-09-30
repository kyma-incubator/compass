package oauth_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/oauth"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTokenProviderFromValue_GetAuthorizationToken(t *testing.T) {
	tokenValue := "token"

	testProvider := oauth.NewTokenProviderFromValue(tokenValue)
	token, err := testProvider.GetAuthorizationToken(context.TODO())

	require.NoError(t, err)
	require.Equal(t, tokenValue, token.AccessToken)
}
