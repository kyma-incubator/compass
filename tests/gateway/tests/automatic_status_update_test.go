package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutomaticStatusUpdate(t *testing.T) {

	ctx := context.Background()

	t.Run("Test status update as static user", func(t *testing.T) {

		t.Log("Register application as Static User")
		appInput := graphql.ApplicationRegisterInput{
			Name:         "app-static-user",
			ProviderName: ptr.String("compass"),
		}
		app, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, appInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &app)
		require.NoError(t, err)

		assert.Equal(t, graphql.ApplicationStatusConditionInitial, app.Status.Condition)

		status := graphql.ApplicationStatusConditionFailed
		t.Log("Update the application")
		appUpdateInput := graphql.ApplicationUpdateInput{
			Description:     str.Ptr("New description"),
			StatusCondition: &status,
		}
		appUpdated, err := fixtures.UpdateApplicationWithinTenant(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, app.ID, appUpdateInput)
		require.NoError(t, err)

		t.Log("Ensure the status condition")
		assert.Equal(t, graphql.ApplicationStatusConditionFailed, appUpdated.Status.Condition)
	})

	t.Run("Test status update as Integration System", func(t *testing.T) {
		t.Log("Register Integration System via Certificate Secured Client")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, "integration-system")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		t.Logf("Registered Integration System with [id=%s]", intSys.ID)

		t.Log("Request Client Credentials for Integration System")
		intSystemAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, intSys.ID)

		intSysOauthCredentialData, ok := intSystemAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, testConfig.DirectorURL)

		t.Log("Register application as Integration System")
		appInput := graphql.ApplicationRegisterInput{
			Name:                "app-is",
			ProviderName:        ptr.String("compass"),
			IntegrationSystemID: &intSys.ID,
		}
		app, err := fixtures.RegisterApplicationFromInput(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, appInput)
		defer fixtures.CleanupApplication(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, &app)
		require.NoError(t, err)

		assert.Equal(t, graphql.ApplicationStatusConditionInitial, app.Status.Condition)

		t.Log("Update the application")
		status := graphql.ApplicationStatusConditionFailed
		appUpdateInput := graphql.ApplicationUpdateInput{
			Description:     str.Ptr("New description"),
			StatusCondition: &status,
		}
		appUpdated, err := fixtures.UpdateApplicationWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, app.ID, appUpdateInput)
		require.NoError(t, err)

		t.Log("Ensure the status condition")
		assert.Equal(t, graphql.ApplicationStatusConditionFailed, appUpdated.Status.Condition)
	})

	t.Run("Test automatic status update as Application", func(t *testing.T) {
		status := graphql.ApplicationStatusConditionFailed
		appInput := graphql.ApplicationRegisterInput{
			Name: "test-app",
			Labels: graphql.Labels{
				"scenarios": []interface{}{"DEFAULT"},
			},
			StatusCondition: &status,
		}

		t.Log("Register Application via Certificate Secured Client")
		app, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, appInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &app)
		require.NoError(t, err)
		t.Logf("Registered Application with [id=%s]", app.ID)

		t.Log("Request Client Credentials for Application")
		appAuth := fixtures.RequestClientCredentialsForApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, app.ID)
		appOauthCredentialData, ok := appAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, appOauthCredentialData.ClientSecret)
		require.NotEmpty(t, appOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, appOauthCredentialData, token.ApplicationScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, testConfig.DirectorURL)

		t.Log("Get Application as Application")
		actualApp := graphql.ApplicationExt{}
		req := fixtures.FixGetApplicationRequest(app.ID)

		assert.Equal(t, graphql.ApplicationStatusConditionFailed, app.Status.Condition)

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, req, &actualApp)
		require.NoError(t, err)

		assert.Equal(t, graphql.ApplicationStatusConditionConnected, actualApp.Status.Condition)
	})

	t.Run("Test automatic status update as Runtime", func(t *testing.T) {
		status := graphql.RuntimeStatusConditionFailed
		runtimeInput := graphql.RuntimeInput{
			Name: "test-runtime",
			Labels: graphql.Labels{
				"scenarios":                              []interface{}{"DEFAULT"},
				testConfig.AppSelfRegDistinguishLabelKey: []interface{}{testConfig.AppSelfRegDistinguishLabelValue},
				tenantfetcher.RegionKey:                  testConfig.AppSelfRegRegion,
			},
			StatusCondition: &status,
		}

		t.Log("Register Runtime via Certificate Secured Client")
		runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &runtimeInput)
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &runtime)
		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)
		t.Logf("Registered Runtime with [id=%s]", runtime.ID)

		t.Log("Request Client Credentials for Runtime")
		rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, runtime.ID)
		rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
		require.NotEmpty(t, rtmOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, rtmOauthCredentialData, token.RuntimeScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, testConfig.DirectorURL)

		t.Log("Get Runtime as Runtime")
		actualRuntime := graphql.RuntimeExt{}
		req := fixtures.FixGetRuntimeRequest(runtime.ID)

		assert.Equal(t, graphql.RuntimeStatusConditionFailed, runtime.Status.Condition)

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, testConfig.DefaultTestTenant, req, &actualRuntime)
		require.NoError(t, err)

		t.Log("Ensure the status condition")
		assert.Equal(t, graphql.RuntimeStatusConditionConnected, actualRuntime.Status.Condition)
	})

}
