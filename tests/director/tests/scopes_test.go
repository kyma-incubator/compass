package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScopesAuthorization(t *testing.T) {
	// given
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixtures.FixRuntimeInput("runtime-test")
	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

	id := uuid.New().String()

	testCases := []struct {
		Name                 string
		UseDefaultScopes     bool
		Scopes               string
		ExpectedErrorMessage string
	}{
		{Name: "Different Scopes", Scopes: "runtime:read", ExpectedErrorMessage: "insufficient scopes provided"},
		{Name: "No scopes", Scopes: "", ExpectedErrorMessage: "insufficient scopes provided"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, tenantId, runtime.ID)
			rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
			require.True(t, ok)
			require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
			require.NotEmpty(t, rtmOauthCredentialData.ClientID)

			t.Log("Issue a Hydra token with Client Credentials")
			accessToken := token.GetAccessToken(t, rtmOauthCredentialData, testCase.Scopes)
			oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

			request := fixtures.FixApplicationForRuntimeRequest(id)
			response := graphql.ApplicationPage{}

			// when
			err := testctx.Tc.RunOperation(ctx, oauthGraphQLClient, request, &response)

			// then
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
		})
	}
}
