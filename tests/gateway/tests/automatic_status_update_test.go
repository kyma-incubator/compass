package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/tests/pkg/ptr"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
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
		app := pkg.RegisterApplicationFromInputWithinTenant(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, appInput)

		defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, app.ID, testConfig.DefaultTenant)

		assert.Equal(t, graphql.ApplicationStatusConditionInitial, app.Status.Condition)

		status := graphql.ApplicationStatusConditionFailed
		t.Log("Update the application")
		appUpdateInput := graphql.ApplicationUpdateInput{
			Description:     str.Ptr("New description"),
			StatusCondition: &status,
		}
		appUpdated, err := pkg.UpdateApplicationWithinTenant(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID, appUpdateInput)
		require.NoError(t, err)

		t.Log("Ensure the status condition")
		assert.Equal(t, graphql.ApplicationStatusConditionFailed, appUpdated.Status.Condition)
	})

	t.Run("Test status update as Integration System", func(t *testing.T) {
		t.Log("Register Integration System with Dex id token")
		intSys := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, "integration-system")

		t.Logf("Registered Integration System with [id=%s]", intSys.ID)
		defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)

		t.Log("Request Client Credentials for Integration System")
		intSystemAuth := pkg.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, intSys.ID)

		intSysOauthCredentialData, ok:= intSystemAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := pkg.GetAccessToken(t, intSysOauthCredentialData, "application:write")
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, testConfig.DirectorURL)

		t.Log("Register application as Integration System")
		appInput := graphql.ApplicationRegisterInput{
			Name:                "app-is",
			ProviderName:        ptr.String("compass"),
			IntegrationSystemID: &intSys.ID,
		}
		app := pkg.RegisterApplicationFromInputWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, appInput)

		defer pkg.UnregisterApplication(t, ctx, oauthGraphQLClient, app.ID, testConfig.DefaultTenant)

		assert.Equal(t, graphql.ApplicationStatusConditionInitial, app.Status.Condition)

		t.Log("Update the application")
		status := graphql.ApplicationStatusConditionFailed
		appUpdateInput := graphql.ApplicationUpdateInput{
			Description:     str.Ptr("New description"),
			StatusCondition: &status,
		}
		appUpdated, err := pkg.UpdateApplicationWithinTenant(t, ctx, oauthGraphQLClient, testConfig.DefaultTenant, app.ID, appUpdateInput)
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
		app := pkg.RegisterApplicationFromInputWithinTenant(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, appInput)

		t.Logf("Registered Application with [id=%s]", app.ID)
		defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, app.ID, testConfig.DefaultTenant)

		t.Log("Request Client Credentials for Application")
		appAuth := pkg.RequestClientCredentialsForApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)
		appOauthCredentialData, ok := appAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, appOauthCredentialData.ClientSecret)
		require.NotEmpty(t, appOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := pkg.GetAccessToken(t, appOauthCredentialData, "application:read")
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, testConfig.DirectorURL)

		t.Log("Get Application as Application")
		actualApp := graphql.ApplicationExt{}
		req := pkg.FixGetApplicationRequest(app.ID)

		assert.Equal(t, graphql.ApplicationStatusConditionFailed, app.Status.Condition)

		err = pkg.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, testConfig.DefaultTenant, req, &actualApp)
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
		runtime := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, &runtimeInput)

		t.Logf("Registered Runtime with [id=%s]", runtime.ID)
		defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, runtime.ID)

		t.Log("Request Client Credentials for Runtime")
		rtmAuth := pkg.RequestClientCredentialsForRuntime(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, runtime.ID)
		rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
		require.NotEmpty(t, rtmOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := pkg.GetAccessToken(t, rtmOauthCredentialData, "runtime:read")
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, testConfig.DirectorURL)

		t.Log("Get Runtime as Runtime")
		actualRuntime := graphql.RuntimeExt{}
		req := pkg.FixGetRuntimeRequest(runtime.ID)

		assert.Equal(t, graphql.RuntimeStatusConditionFailed, runtime.Status.Condition)

		err = pkg.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, testConfig.DefaultTenant, req, &actualRuntime)
		require.NoError(t, err)

		t.Log("Ensure the status condition")
		assert.Equal(t, graphql.RuntimeStatusConditionConnected, actualRuntime.Status.Condition)
	})

}
