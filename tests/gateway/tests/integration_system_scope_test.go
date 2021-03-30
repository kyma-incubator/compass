package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	token2 "github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/tests/pkg/ptr"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSystemScenario(t *testing.T) {
	ctx := context.Background()

	t.Log("Register Integration System with Dex id token")
	intSys := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, "integration-system")

	t.Log("Request Client Credentials for Integration System")
	intSystemAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)

	intSysOauthCredentialData, ok := intSystemAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a token with Client Credentials")
	token := token2.GetAccessToken(t, intSysOauthCredentialData, token2.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(token, testConfig.DirectorURL)
	t.Run("Test application scopes", func(t *testing.T) {
		t.Log("Register an application")
		appInput := graphql.ApplicationRegisterInput{
			Name:                "app",
			ProviderName:        ptr.String("compass"),
			IntegrationSystemID: &intSys.ID,
		}
		appByIntSys, err := fixtures.RegisterApplicationFromInput(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appInput)
		require.NoError(t, err)
		require.NotEmpty(t, appByIntSys.ID)

		t.Log("Get application")
		app := fixtures.GetApplication(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appByIntSys.ID)

		require.NotEmpty(t, app.ID)
		require.Equal(t, appByIntSys.ID, app.ID)

		t.Log("Unregister application")
		fixtures.UnregisterApplication(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appByIntSys.ID)

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
		appTpl := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appTplInput)
		require.NotEmpty(t, appTpl.ID)

		t.Log("Get application template")
		gqlAppTpl := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appTpl.ID)
		require.NotEmpty(t, gqlAppTpl.ID)
		require.Equal(t, appTpl.ID, gqlAppTpl.ID)

		t.Log("Delete application template")
		fixtures.DeleteApplicationTemplate(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appTpl.ID)

	})

	t.Run("Test runtime scopes", func(t *testing.T) {
		t.Log("Register runtime")
		runtimeInput := graphql.RuntimeInput{
			Name: "test",
		}
		runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, &runtimeInput)
		require.NotEmpty(t, runtime.ID)

		t.Log("Get runtime")
		gqlRuntime := fixtures.GetRuntime(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, runtime.ID)

		require.NotEmpty(t, gqlRuntime.ID)
		require.Equal(t, runtime.ID, gqlRuntime.ID)

		t.Log("Unregister runtime")
		fixtures.UnregisterRuntime(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, runtime.ID)

	})

	t.Log("Unregister Integration System")
	fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)
}
