package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/tests/pkg/ptr"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSystemScenario(t *testing.T) {
	ctx := context.Background()

	t.Log("Register Integration System via Certificate Secured Client")
	var intSys graphql.IntegrationSystemExt // needed so the 'defer' can be above the integration system registration
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &intSys)
	intSys = fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, "integration-system")

	t.Log("Request Client Credentials for Integration System")
	var intSystemAuth graphql.IntSysSystemAuth // needed so the 'defer' can be above the integration system auth creation
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, &intSystemAuth)
	intSystemAuth = fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, intSys.ID)
	require.NotEmpty(t, intSystemAuth)

	intSysOauthCredentialData, ok := intSystemAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a token with Client Credentials")
	token := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(token, testConfig.DirectorURL)
	t.Run("Test application scopes", func(t *testing.T) {
		t.Log("Register an application")
		appInput := graphql.ApplicationRegisterInput{
			Name:                "app",
			ProviderName:        ptr.String("compass"),
			IntegrationSystemID: &intSys.ID,
		}
		appByIntSys, err := fixtures.RegisterApplicationFromInput(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, appInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &appByIntSys)
		require.NoError(t, err)
		require.NotEmpty(t, appByIntSys.ID)

		t.Log("Get application")
		app := fixtures.GetApplication(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, appByIntSys.ID)

		require.NotEmpty(t, app.ID)
		require.Equal(t, appByIntSys.ID, app.ID)

		t.Log("Unregister application")
		fixtures.UnregisterApplication(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, appByIntSys.ID)

	})
	t.Run("Test application template scopes", func(t *testing.T) {
		t.Log("Create an application template")
		appTplInput := graphql.ApplicationTemplateInput{
			Name:        "test",
			Description: nil,
			ApplicationInput: &graphql.ApplicationJSONInput{
				Name:         "test",
				ProviderName: ptr.String("test"),
			},
			Placeholders: nil,
			AccessLevel:  "GLOBAL",
		}
		appTpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, appTplInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, appTpl)
		require.NoError(t, err)
		require.NotEmpty(t, appTpl.ID)

		t.Log("Get application template")
		gqlAppTpl := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, appTpl.ID)
		require.NotEmpty(t, gqlAppTpl.ID)
		require.Equal(t, appTpl.ID, gqlAppTpl.ID)

		t.Log("Delete application template")
		fixtures.DeleteApplicationTemplate(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, appTpl.ID)

	})

	t.Run("Test runtime scopes", func(t *testing.T) {
		t.Log("Register runtime")
		runtimeInput := graphql.RuntimeRegisterInput{
			Name: "test",
		}
		var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &runtime)
		runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, &runtimeInput)
		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)

		t.Log("Get runtime")
		gqlRuntime := fixtures.GetRuntime(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, runtime.ID)

		require.NotEmpty(t, gqlRuntime.ID)
		require.Equal(t, runtime.ID, gqlRuntime.ID)

		t.Log("Unregister runtime")
		fixtures.UnregisterRuntime(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, runtime.ID)

	})
}
