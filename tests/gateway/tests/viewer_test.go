package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViewerQuery(t *testing.T) {
	ctx := context.Background()

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, testScenario)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, testScenario)

	t.Run("Test viewer as Integration System", func(t *testing.T) {
		t.Log("Register Integration System via Certificate Secured Client")
		var intSys graphql.IntegrationSystemExt // needed so the 'defer' can be above the integration system registration
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &intSys)
		intSys = fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, "integration-system")

		t.Logf("Registered Integration System with [id=%s]", intSys.ID)

		t.Log("Request Client Credentials for Integration System")
		var intSystemAuth graphql.IntSysSystemAuth // needed so the 'defer' can be above the integration system auth creation
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, &intSystemAuth)
		intSystemAuth = fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, intSys.ID)
		require.NotEmpty(t, intSystemAuth)

		intSysOauthCredentialData, ok := intSystemAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, "")
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, testConfig.DirectorURL)

		t.Log("Requesting Viewer as Integration System")
		viewer := graphql.Viewer{}
		req := fixtures.FixGetViewerRequest()

		err := testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, req, &viewer)
		require.NoError(t, err)
		assert.Equal(t, intSys.ID, viewer.ID)
		assert.Equal(t, graphql.ViewerTypeIntegrationSystem, viewer.Type)
	})

	t.Run("Test viewer as Application", func(t *testing.T) {
		appInput := graphql.ApplicationRegisterInput{
			Name: "test-app",
			Labels: graphql.Labels{
				testConfig.ApplicationTypeLabelKey: string(util.ApplicationTypeC4C),
			},
		}

		t.Log("Register Application via Certificate Secured Client")
		app, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, appInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &app)
		require.NoError(t, err)
		t.Logf("Registered Application with [id=%s]", app.ID)

		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: testScenario}, app.ID, testConfig.DefaultTestTenant)
		fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: testScenario}, app.ID, testConfig.DefaultTestTenant)

		t.Log("Request Client Credentials for Application")
		appAuth := fixtures.RequestClientCredentialsForApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, app.ID)
		appOauthCredentialData, ok := appAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, appOauthCredentialData.ClientSecret)
		require.NotEmpty(t, appOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, appOauthCredentialData, "")
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", testConfig.Domain))

		t.Log("Requesting Viewer as Application")
		viewer := graphql.Viewer{}
		req := fixtures.FixGetViewerRequest()

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, req, &viewer)
		require.NoError(t, err)
		assert.Equal(t, app.ID, viewer.ID)
		assert.Equal(t, graphql.ViewerTypeApplication, viewer.Type)
	})

	t.Run("Test viewer as Runtime", func(t *testing.T) {
		runtimeInput := graphql.RuntimeRegisterInput{
			Name: "test-runtime",
		}

		t.Log("Register Runtime via Certificate Secured Client")
		var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &runtime)
		runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, runtimeInput, testConfig.GatewayOauth)
		t.Logf("Registered Runtime with [id=%s]", runtime.ID)

		defer fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: testScenario}, runtime.ID, testConfig.DefaultTestTenant)
		fixtures.AssignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: testScenario}, runtime.ID, testConfig.DefaultTestTenant)

		t.Log("Request Client Credentials for Runtime")
		rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, runtime.ID)
		rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
		require.NotEmpty(t, rtmOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, rtmOauthCredentialData, "")
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", testConfig.Domain))

		t.Log("Requesting Viewer as Runtime")
		viewer := graphql.Viewer{}
		req := fixtures.FixGetViewerRequest()

		err := testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, req, &viewer)
		require.NoError(t, err)
		assert.Equal(t, runtime.ID, viewer.ID)
		assert.Equal(t, graphql.ViewerTypeRuntime, viewer.Type)
	})

}
