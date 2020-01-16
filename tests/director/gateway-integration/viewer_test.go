package gateway_integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViewerQuery(t *testing.T) {
	domain := os.Getenv("DOMAIN")
	require.NotEmpty(t, domain)
	tenant := os.Getenv("DEFAULT_TENANT")
	require.NotEmpty(t, tenant)
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Run("Test viewer as Integration System", func(t *testing.T) {
		t.Log("Register Integration System with Dex id token")
		intSys := registerIntegrationSystem(t, ctx, dexGraphQLClient, tenant, "integration-system")

		t.Logf("Registered Integration System with [id=%s]", intSys.ID)
		defer unregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)

		t.Log("Request Client Credentials for Integration System")
		intSysOauthCredentialData := requestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := getAccessToken(t, intSysOauthCredentialData, "")
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", domain))

		t.Log("Requesting Viewer as Integration System")
		viewer := graphql.Viewer{}
		req := fixGetViewerRequest()

		err = tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenant, req, &viewer)
		require.NoError(t, err)
		assert.Equal(t, intSys.ID, viewer.ID)
		assert.Equal(t, graphql.ViewerTypeIntegrationSystem, viewer.Type)
	})

	t.Run("Test viewer as Application", func(t *testing.T) {
		appInput := graphql.ApplicationRegisterInput{
			Name: "test-app",
			Labels: &graphql.Labels{
				"scenarios": []interface{}{"DEFAULT"},
			},
		}

		t.Log("Register Application with Dex id token")
		app := registerApplicationFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, appInput)

		t.Logf("Registered Application with [id=%s]", app.ID)
		defer unregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)

		t.Log("Request Client Credentials for Application")
		appAuth := requestClientCredentialsForApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
		appOauthCredentialData, ok := appAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, appOauthCredentialData.ClientSecret)
		require.NotEmpty(t, appOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := getAccessToken(t, appOauthCredentialData, "")
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", domain))

		t.Log("Requesting Viewer as Application")
		viewer := graphql.Viewer{}
		req := fixGetViewerRequest()

		err = tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenant, req, &viewer)
		require.NoError(t, err)
		assert.Equal(t, app.ID, viewer.ID)
		assert.Equal(t, graphql.ViewerTypeApplication, viewer.Type)
	})

	t.Run("Test viewer as Runtime", func(t *testing.T) {
		runtimeInput := graphql.RuntimeInput{
			Name: "test-runtime",
			Labels: &graphql.Labels{
				"scenarios": []interface{}{"DEFAULT"},
			},
		}

		t.Log("Register Runtime with Dex id token")
		runtime := registerRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &runtimeInput)

		t.Logf("Registered Runtime with [id=%s]", runtime.ID)
		defer unregisterRuntimeWithinTenant(t, ctx, dexGraphQLClient, tenant, runtime.ID)

		t.Log("Request Client Credentials for Runtime")
		rtmAuth := requestClientCredentialsForRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)
		rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
		require.NotEmpty(t, rtmOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := getAccessToken(t, rtmOauthCredentialData, "")
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", domain))

		t.Log("Requesting Viewer as Runtime")
		viewer := graphql.Viewer{}
		req := fixGetViewerRequest()

		err = tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenant, req, &viewer)
		require.NoError(t, err)
		assert.Equal(t, runtime.ID, viewer.ID)
		assert.Equal(t, graphql.ViewerTypeRuntime, viewer.Type)
	})

}
