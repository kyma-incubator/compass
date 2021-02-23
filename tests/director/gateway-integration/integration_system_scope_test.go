package gateway_integration

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"

	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSystemScenario(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Log("Register Integration System with Dex id token")
	intSys := registerIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, "integration-system")

	t.Log("Request Client Credentials for Integration System")
	intSysOauthCredentialData := requestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)

	t.Log("Issue a token with Client Credentials")
	token := getAccessToken(t, intSysOauthCredentialData, integrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(token, testConfig.DirectorURL)
	t.Run("Test application scopes", func(t *testing.T) {
		t.Log("Register an application")
		appInput := graphql.ApplicationRegisterInput{
			Name:                "app",
			ProviderName:        ptr.String("compass"),
			IntegrationSystemID: &intSys.ID,
		}
		appByIntSys := registerApplicationFromInputWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appInput)
		require.NotEmpty(t, appByIntSys.ID)

		t.Log("Get application")
		app := getApplication(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appByIntSys.ID)
		require.NotEmpty(t, app.ID)
		require.Equal(t, appByIntSys.ID, app.ID)

		t.Log("Unregister application")
		UnregisterApplication(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appByIntSys.ID)

	})
	t.Run("Test application template scopes", func(t *testing.T) {
		t.Log("Create an application template")
		appTplInput := graphql.ApplicationTemplateInput{
			Name:        "test",
			Description: nil,
			ApplicationInput: &graphql.ApplicationRegisterInput{
				Name:         "test",
				ProviderName: ptr.String("test"),
			},
			Placeholders: nil,
			AccessLevel:  "GLOBAL",
		}
		appTpl := createApplicationTemplate(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appTplInput)
		require.NotEmpty(t, appTpl.ID)

		t.Log("Get application template")
		gqlAppTpl := getApplicationTemplate(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appTpl.ID)
		require.NotEmpty(t, gqlAppTpl.ID)
		require.Equal(t, appTpl.ID, gqlAppTpl.ID)

		t.Log("Delete application template")
		deleteApplicationTemplate(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appTpl.ID)

	})

	t.Run("Test runtime scopes", func(t *testing.T) {
		t.Log("Register runtime")
		runtimeInput := graphql.RuntimeInput{
			Name: "test",
		}
		runtime := RegisterRuntimeFromInputWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, &runtimeInput)
		require.NotEmpty(t, runtime.ID)

		t.Log("Get runtime")
		gqlRuntime := getRuntime(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, runtime.ID)
		require.NotEmpty(t, gqlRuntime.ID)
		require.Equal(t, runtime.ID, gqlRuntime.ID)

		t.Log("Unregister runtime")
		UnregisterRuntimeWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, runtime.ID)

	})

	t.Log("Unregister Integration System")
	unregisterIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)
}
