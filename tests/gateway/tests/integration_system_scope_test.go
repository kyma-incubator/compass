package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/ptr"

	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSystemScenario(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Log("Register Integration System with Dex id token")
	intSys := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, "integration-system")

	t.Log("Request Client Credentials for Integration System")
	intSystemAuth := pkg.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)

	intSysOauthCredentialData, ok:= intSystemAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a token with Client Credentials")
	token := pkg.GetAccessToken(t, intSysOauthCredentialData, pkg.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(token, testConfig.DirectorURL)
	t.Run("Test application scopes", func(t *testing.T) {
		t.Log("Register an application")
		appInput := graphql.ApplicationRegisterInput{
			Name:                "app",
			ProviderName:        ptr.String("compass"),
			IntegrationSystemID: &intSys.ID,
		}
		appByIntSys := pkg.RegisterApplicationFromInputWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appInput)
		require.NotEmpty(t, appByIntSys.ID)

		t.Log("Get application")
		app := pkg.GetApplication(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appByIntSys.ID)
		require.NotEmpty(t, app.ID)
		require.Equal(t, appByIntSys.ID, app.ID)

		t.Log("Unregister application")
		pkg.UnregisterApplication(t, ctx, oauthGraphQLClient, appByIntSys.ID, testConfig.DefaultTenant)

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
		appTpl := pkg.CreateApplicationTemplate(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appTplInput)
		require.NotEmpty(t, appTpl.ID)

		t.Log("Get application template")
		gqlAppTpl := pkg.GetApplicationTemplate(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appTpl.ID)
		require.NotEmpty(t, gqlAppTpl.ID)
		require.Equal(t, appTpl.ID, gqlAppTpl.ID)

		t.Log("Delete application template")
		pkg.DeleteApplicationTemplate(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appTpl.ID)

	})

	t.Run("Test runtime scopes", func(t *testing.T) {
		t.Log("Register runtime")
		runtimeInput := graphql.RuntimeInput{
			Name: "test",
		}
		runtime := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, &runtimeInput)
		require.NotEmpty(t, runtime.ID)

		t.Log("Get runtime")
		gqlRuntime := pkg.GetRuntime(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, runtime.ID)
		require.NotEmpty(t, gqlRuntime.ID)
		require.Equal(t, runtime.ID, gqlRuntime.ID)

		t.Log("Unregister runtime")
		pkg.UnregisterRuntime(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, runtime.ID)

	})

	t.Log("Unregister Integration System")
	pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)
}
