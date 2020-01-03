package e2e

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/gql"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/idtokenprovider"
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
	config, err := idtokenprovider.LoadConfig()
	require.NoError(t, err)
	dexToken, err := idtokenprovider.Authenticate(config.IdProviderConfig)
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Run("Test viewer as Integration System", func(t *testing.T) {
		t.Log("Create Integration System with Dex id token")
		intSys := registerIntegrationSystem(t, ctx, dexGraphQLClient, tenant, "integration-system")

		t.Logf("Registered Integration System with [id=%s]", intSys.ID)
		defer unregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)

		t.Log("Generate Client Credentials for Integration System")
		intSysAuth := generateClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, tenant, intSys.ID)
		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, intSysOauthCredentialData.ClientSecret)
		require.NotEmpty(t, intSysOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		oauthCredentials := fmt.Sprintf("%s:%s", intSysOauthCredentialData.ClientID, intSysOauthCredentialData.ClientSecret)
		encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
		hydraToken, err := fetchHydraAccessTokenWithScopesForApplication(t, encodedCredentials, intSysOauthCredentialData.URL)
		require.NoError(t, err)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(hydraToken.AccessToken, fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", domain))

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
			Name:                "test-app",
			ProviderDisplayName: "compass",
			Labels: &graphql.Labels{
				"scenarios": []interface{}{"DEFAULT"},
			},
		}

		t.Log("Create Application with Dex id token")
		app := registerApplicationFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, appInput)

		t.Logf("Registered Application with [id=%s]", app.ID)
		defer deleteApplication(t, ctx, dexGraphQLClient, tenant, app.ID)

		t.Log("Generate Client Credentials for Application")
		appAuth := generateClientCredentialsForApplication(t, ctx, dexGraphQLClient, tenant, app.ID)
		appOauthCredentialData, ok := appAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, appOauthCredentialData.ClientSecret)
		require.NotEmpty(t, appOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		oauthCredentials := fmt.Sprintf("%s:%s", appOauthCredentialData.ClientID, appOauthCredentialData.ClientSecret)
		encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
		hydraToken, err := fetchHydraAccessTokenWithScopesForApplication(t, encodedCredentials, appOauthCredentialData.URL)
		require.NoError(t, err)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(hydraToken.AccessToken, fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", domain))

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
			Name: "test-runtime513251",
			Labels: &graphql.Labels{
				"scenarios": []interface{}{"DEFAULT"},
			},
		}

		t.Log("Create Runtime with Dex id token")
		runtime := registerRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &runtimeInput)

		t.Logf("Registered Runtime with [id=%s]", runtime.ID)
		defer unregisterRuntimeWithinTenant(t, ctx, dexGraphQLClient, tenant, runtime.ID)

		t.Log("Generate Client Credentials for Runtime")
		rtmAuth := generateClientCredentialsForRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)
		rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
		require.NotEmpty(t, rtmOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		oauthCredentials := fmt.Sprintf("%s:%s", rtmOauthCredentialData.ClientID, rtmOauthCredentialData.ClientSecret)
		encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
		hydraToken, err := fetchHydraAccessTokenWithScopesForRuntime(t, encodedCredentials, rtmOauthCredentialData.URL)
		require.NoError(t, err)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(hydraToken.AccessToken, fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", domain))

		t.Log("Requesting Viewer as Runtime")
		viewer := graphql.Viewer{}
		req := fixGetViewerRequest()

		err = tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenant, req, &viewer)
		require.NoError(t, err)
		assert.Equal(t, runtime.ID, viewer.ID)
		assert.Equal(t, graphql.ViewerTypeRuntime, viewer.Type)
	})

}
