package gateway_integration

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutomaticStatusUpdate(t *testing.T) {

	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Run("Test status update as static user", func(t *testing.T) {

		t.Log("Register application as Static User")
		appInput := graphql.ApplicationRegisterInput{
			Name:         "app-static-user",
			ProviderName: ptr.String("compass"),
		}
		app, err := registerApplicationWithinTenant(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, appInput)
		require.NoError(t, err)

		defer unregisterApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)

		assert.Equal(t, graphql.ApplicationStatusConditionInitial, app.Status.Condition)

		status := graphql.ApplicationStatusConditionFailed
		t.Log("Update the application")
		appUpdateInput := graphql.ApplicationUpdateInput{
			Description:     str.Ptr("New description"),
			StatusCondition: &status,
		}
		appUpdated, err := updateApplicationWithinTenant(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID, appUpdateInput)
		require.NoError(t, err)

		t.Log("Ensure the status condition")
		assert.Equal(t, graphql.ApplicationStatusConditionFailed, appUpdated.Status.Condition)
	})

	t.Run("Test status update as Integration System", func(t *testing.T) {
		t.Log("Register Integration System with Dex id token")
		intSys := registerIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, "integration-system")

		t.Logf("Registered Integration System with [id=%s]", intSys.ID)
		defer unregisterIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)

		t.Log("Request Client Credentials for Integration System")
		intSysOauthCredentialData := requestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := getAccessToken(t, intSysOauthCredentialData, "application:write")
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, testConfig.DirectorURL)

		t.Log("Register application as Integration System")
		appInput := graphql.ApplicationRegisterInput{
			Name:                "app-is",
			ProviderName:        ptr.String("compass"),
			IntegrationSystemID: &intSys.ID,
		}
		app, err := registerApplicationWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appInput)
		require.NoError(t, err)

		defer unregisterApplication(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, app.ID)

		assert.Equal(t, graphql.ApplicationStatusConditionInitial, app.Status.Condition)

		t.Log("Update the application")
		status := graphql.ApplicationStatusConditionFailed
		appUpdateInput := graphql.ApplicationUpdateInput{
			Description:     str.Ptr("New description"),
			StatusCondition: &status,
		}
		appUpdated, err := updateApplicationWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, app.ID, appUpdateInput)
		require.NoError(t, err)

		t.Log("Ensure the status condition")
		assert.Equal(t, graphql.ApplicationStatusConditionFailed, appUpdated.Status.Condition)
	})

	t.Run("Test automatic status update as Application", func(t *testing.T) {
		status := graphql.ApplicationStatusConditionFailed
		appInput := graphql.ApplicationRegisterInput{
			Name: "test-app",
			Labels: &graphql.Labels{
				"scenarios": []interface{}{"DEFAULT"},
			},
			StatusCondition: &status,
		}

		t.Log("Register Application with Dex id token")
		app := registerApplicationFromInputWithinTenant(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, appInput)

		t.Logf("Registered Application with [id=%s]", app.ID)
		defer unregisterApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)

		t.Log("Request Client Credentials for Application")
		appAuth := requestClientCredentialsForApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)
		appOauthCredentialData, ok := appAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, appOauthCredentialData.ClientSecret)
		require.NotEmpty(t, appOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := getAccessToken(t, appOauthCredentialData, "application:read")
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, testConfig.DirectorURL)

		t.Log("Get Application as Application")
		actualApp := graphql.ApplicationExt{}
		req := fixGetApplicationRequest(app.ID)

		assert.Equal(t, graphql.ApplicationStatusConditionFailed, app.Status.Condition)

		err = tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, testConfig.DefaultTenant, req, &actualApp)
		require.NoError(t, err)

		assert.Equal(t, graphql.ApplicationStatusConditionConnected, actualApp.Status.Condition)
	})

	t.Run("Test automatic status update as Runtime", func(t *testing.T) {
		status := graphql.RuntimeStatusConditionFailed
		runtimeInput := graphql.RuntimeInput{
			Name: "test-runtime",
			Labels: &graphql.Labels{
				"scenarios": []interface{}{"DEFAULT"},
			},
			StatusCondition: &status,
		}

		t.Log("Register Runtime with Dex id token")
		runtime := registerRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, &runtimeInput)

		t.Logf("Registered Runtime with [id=%s]", runtime.ID)
		defer unregisterRuntimeWithinTenant(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, runtime.ID)

		t.Log("Request Client Credentials for Runtime")
		rtmAuth := requestClientCredentialsForRuntime(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, runtime.ID)
		rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
		require.NotEmpty(t, rtmOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := getAccessToken(t, rtmOauthCredentialData, "runtime:read")
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, testConfig.DirectorURL)

		t.Log("Get Runtime as Runtime")
		actualRuntime := graphql.RuntimeExt{}
		req := fixRuntimeRequest(runtime.ID)

		assert.Equal(t, graphql.RuntimeStatusConditionFailed, runtime.Status.Condition)

		err = tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, testConfig.DefaultTenant, req, &actualRuntime)
		require.NoError(t, err)

		t.Log("Ensure the status condition")
		assert.Equal(t, graphql.RuntimeStatusConditionConnected, actualRuntime.Status.Condition)
	})

}
