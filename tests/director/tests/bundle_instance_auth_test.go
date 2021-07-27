package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func TestRequestBundleInstanceAuthCreation(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenantId, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	authCtx, inputParams := fixtures.FixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := fixtures.FixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixtures.FixRequestBundleInstanceAuthCreationRequest(bndl.ID, bndlInstanceAuthRequestInputStr)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth creation")
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, bndlInstanceAuthCreationRequestReq, &output)

	// THEN
	require.NoError(t, err)
	require.Nil(t, output.RuntimeID)
	require.Nil(t, output.RuntimeContextID)
	assertions.AssertBundleInstanceAuthInput(t, bndlInstanceAuthRequestInput, output)

	saveExample(t, bndlInstanceAuthCreationRequestReq.Query(), "request bundle instance auth creation")

	// Fetch Application with bundles
	bundlesForApplicationReq := fixtures.FixGetBundlesRequest(application.ID)
	bndlFromAPI := graphql.ApplicationExt{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, bundlesForApplicationReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data))
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data[0].InstanceAuths))

	// Fetch Application with bundle
	instanceAuthID := bndlFromAPI.Bundles.Data[0].InstanceAuths[0].ID
	bundlesForApplicationWithInstanceAuthReq := fixtures.FixGetBundleWithInstanceAuthRequest(application.ID, bndl.ID, instanceAuthID)

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, bundlesForApplicationWithInstanceAuthReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auth exist
	require.Equal(t, instanceAuthID, bndlFromAPI.Bundle.InstanceAuth.ID)
	require.Equal(t, graphql.BundleInstanceAuthStatusConditionPending, bndlFromAPI.Bundle.InstanceAuth.Status.Condition)
}

func TestRequestBundleInstanceAuthCreationAsRuntimeConsumer(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixtures.FixRuntimeInput("runtime-test")

	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenantId, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	authCtx, inputParams := fixtures.FixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := fixtures.FixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixtures.FixRequestBundleInstanceAuthCreationRequest(bndl.ID, bndlInstanceAuthRequestInputStr)
	output := graphql.BundleInstanceAuth{}

	scenarios := []string{conf.DefaultScenario, "test-scenario"}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantId, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantId, scenarios[:1])

	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, tenantId, runtime.ID)
	rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
	require.NotEmpty(t, rtmOauthCredentialData.ClientID)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, rtmOauthCredentialData, token.RuntimeScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	runtimeConsumer := testctx.Tc.NewOperation(ctx)

	t.Run("When runtime is in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[:1])

		t.Log("Request bundle instance auth creation")
		err = runtimeConsumer.Run(bndlInstanceAuthCreationRequestReq, oauthGraphQLClient, &output)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, output.RuntimeID)
		require.Equal(t, runtime.ID, *output.RuntimeID)
		require.Nil(t, output.RuntimeContextID)
		assertions.AssertBundleInstanceAuthInput(t, bndlInstanceAuthRequestInput, output)

		// Fetch Application with bundles
		bundlesForApplicationReq := fixtures.FixGetBundlesRequest(application.ID)
		bndlFromAPI := graphql.ApplicationExt{}

		err = runtimeConsumer.Run(bundlesForApplicationReq, oauthGraphQLClient, &bndlFromAPI)
		require.NoError(t, err)

		// Assert the bundle instance auths exists
		require.Equal(t, 1, len(bndlFromAPI.Bundles.Data))
		require.Equal(t, 1, len(bndlFromAPI.Bundles.Data[0].InstanceAuths))

		// Fetch Application with bundle instance auth
		instanceAuthID := bndlFromAPI.Bundles.Data[0].InstanceAuths[0].ID
		bundlesForApplicationWithInstanceAuthReq := fixtures.FixGetBundleWithInstanceAuthRequest(application.ID, bndl.ID, instanceAuthID)

		err = runtimeConsumer.Run(bundlesForApplicationWithInstanceAuthReq, oauthGraphQLClient, &bndlFromAPI)
		require.NoError(t, err)

		// Assert the bundle instance auth exist
		require.Equal(t, instanceAuthID, bndlFromAPI.Bundle.InstanceAuth.ID)
		require.Equal(t, graphql.BundleInstanceAuthStatusConditionPending, bndlFromAPI.Bundle.InstanceAuth.Status.Condition)
	})

	t.Run("When runtime is NOT in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[:1])

		output = graphql.BundleInstanceAuth{}
		t.Log("Request bundle instance auth creation")
		err = runtimeConsumer.Run(bndlInstanceAuthCreationRequestReq, oauthGraphQLClient, &output)

		// THEN
		require.Error(t, err)
		require.Nil(t, output.RuntimeID)
		require.Nil(t, output.RuntimeContextID)
		require.Contains(t, err.Error(), "The operation is not allowed")
	})
}

func TestRuntimeIdInBundleInstanceAuthIsSetToNullWhenDeletingRuntime(t *testing.T) {
	ctx := context.Background()
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixtures.FixRuntimeInput("runtime-test")
	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
	defer fixtures.UnregisterGracefullyRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	authInput := fixtures.FixOauthAuth(t)
	bndlInput := fixtures.FixBundleCreateInputWithDefaultAuth("bndl-app-1", authInput)
	bndl, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(bndlInput)
	require.NoError(t, err)

	addBndlRequest := fixtures.FixAddBundleRequest(application.ID, bndl)
	bndlAddOutput := graphql.Bundle{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, addBndlRequest, &bndlAddOutput)
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndlAddOutput.ID)
	require.NoError(t, err)

	authCtx, inputParams := fixtures.FixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := fixtures.FixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixtures.FixRequestBundleInstanceAuthCreationRequest(bndlAddOutput.ID, bndlInstanceAuthRequestInputStr)
	bndlInstanceAuth := graphql.BundleInstanceAuth{}

	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, tenantId, runtime.ID)
	rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
	require.NotEmpty(t, rtmOauthCredentialData.ClientID)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, rtmOauthCredentialData, token.RuntimeScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	runtimeConsumer := testctx.Tc.NewOperation(ctx)

	scenarios := []string{conf.DefaultScenario, "test-scenario"}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantId, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantId, scenarios[:1])

	// set application scenarios label
	fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[1:])
	defer fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[:1])

	// set runtime scenarios label
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[1:])

	t.Log("Request bundle instance auth creation")
	err = runtimeConsumer.Run(bndlInstanceAuthCreationRequestReq, oauthGraphQLClient, &bndlInstanceAuth)

	// THEN
	require.NoError(t, err)
	require.NotNil(t, bndlInstanceAuth.RuntimeID)
	require.Equal(t, runtime.ID, *bndlInstanceAuth.RuntimeID)
	require.Nil(t, bndlInstanceAuth.RuntimeContextID)
	assertions.AssertBundleInstanceAuthInput(t, bndlInstanceAuthRequestInput, bndlInstanceAuth)
	assertions.AssertAuth(t, authInput, bndlInstanceAuth.Auth)

	// Fetch Application with bundles
	bundlesForApplicationReq := fixtures.FixGetBundlesRequest(application.ID)
	appExt := graphql.ApplicationExt{}

	t.Log("Fetch application with bundles")
	err = runtimeConsumer.Run(bundlesForApplicationReq, oauthGraphQLClient, &appExt)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(appExt.Bundles.Data))
	require.Equal(t, 1, len(appExt.Bundles.Data[0].InstanceAuths))
	require.NotNil(t, runtime.ID, appExt.Bundles.Data[0].InstanceAuths[0].RuntimeID)
	require.Equal(t, runtime.ID, *appExt.Bundles.Data[0].InstanceAuths[0].RuntimeID)
	require.Equal(t, graphql.BundleInstanceAuthStatusConditionSucceeded, appExt.Bundles.Data[0].InstanceAuths[0].Status.Condition)
	assertions.AssertAuth(t, authInput, appExt.Bundles.Data[0].InstanceAuths[0].Auth)

	t.Log("Unregister runtime")
	delReq := fixtures.FixUnregisterRuntimeRequest(runtime.ID)
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, delReq, nil)
	require.NoError(t, err)

	appExt = graphql.ApplicationExt{}
	t.Log("Fetch application with bundles after deleting runtime")
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, bundlesForApplicationReq, &appExt)
	require.NoError(t, err)

	t.Log("Assert that the runtime_id column in the bundle instance auth table is set to null after deleting runtime")
	require.Equal(t, 1, len(appExt.Bundles.Data))
	require.Equal(t, 1, len(appExt.Bundles.Data[0].InstanceAuths))
	require.Nil(t, appExt.Bundles.Data[0].InstanceAuths[0].RuntimeID)
	require.Equal(t, graphql.BundleInstanceAuthStatusConditionUnused, appExt.Bundles.Data[0].InstanceAuths[0].Status.Condition)
	assertions.AssertAuth(t, authInput, appExt.Bundles.Data[0].InstanceAuths[0].Auth)
}

func TestRequestBundleInstanceAuthCreationWithDefaultAuth(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	authInput := fixtures.FixBasicAuth(t)

	bndlInput := fixtures.FixBundleCreateInputWithDefaultAuth("bndl-app-1", authInput)
	bndl, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(bndlInput)
	require.NoError(t, err)

	addBndlRequest := fixtures.FixAddBundleRequest(application.ID, bndl)
	bndlAddOutput := graphql.Bundle{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, addBndlRequest, &bndlAddOutput)
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndlAddOutput.ID)
	require.NoError(t, err)

	bndlInstanceAuthRequestInput := fixtures.FixBundleInstanceAuthRequestInput(nil, nil)
	bndlInstanceAuthRequestInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixtures.FixRequestBundleInstanceAuthCreationRequest(bndlAddOutput.ID, bndlInstanceAuthRequestInputStr)
	authOutput := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth creation")
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, bndlInstanceAuthCreationRequestReq, &authOutput)

	// THEN
	require.NoError(t, err)

	// Fetch Application with bundles
	bundlesForApplicationReq := fixtures.FixGetBundlesRequest(application.ID)
	bndlFromAPI := graphql.ApplicationExt{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, bundlesForApplicationReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data))
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data[0].InstanceAuths))

	// Fetch Application with bundle
	instanceAuthID := bndlFromAPI.Bundles.Data[0].InstanceAuths[0].ID
	bundlesForApplicationWithInstanceAuthReq := fixtures.FixGetBundleWithInstanceAuthRequest(application.ID, bndlAddOutput.ID, instanceAuthID)

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, bundlesForApplicationWithInstanceAuthReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auth exist
	require.Equal(t, instanceAuthID, bndlFromAPI.Bundle.InstanceAuth.ID)

	require.Equal(t, graphql.BundleInstanceAuthStatusConditionSucceeded, bndlFromAPI.Bundle.InstanceAuth.Status.Condition)
	assertions.AssertAuth(t, authInput, bndlFromAPI.Bundle.InstanceAuth.Auth)
}

func TestRequestBundleInstanceAuthDeletion(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenantId, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	bndlInstanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, dexGraphQLClient, bndl.ID)

	bndlInstanceAuthDeletionRequestReq := fixtures.FixRequestBundleInstanceAuthDeletionRequest(bndlInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth deletion")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, bndlInstanceAuthDeletionRequestReq, &output)

	// THEN
	require.NoError(t, err)

	saveExample(t, bndlInstanceAuthDeletionRequestReq.Query(), "request bundle instance auth deletion")
}

func TestRequestBundleInstanceAuthDeletionAsRuntimeConsumer(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixtures.FixRuntimeInput("runtime-test")

	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenantId, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	bndlInstanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, dexGraphQLClient, bndl.ID)

	bndlInstanceAuthDeletionRequestReq := fixtures.FixRequestBundleInstanceAuthDeletionRequest(bndlInstanceAuth.ID)

	scenarios := []string{conf.DefaultScenario, "test-scenario"}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantId, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantId, scenarios[:1])

	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, tenantId, runtime.ID)
	rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
	require.NotEmpty(t, rtmOauthCredentialData.ClientID)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, rtmOauthCredentialData, token.RuntimeScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	runtimeConsumer := testctx.Tc.NewOperation(ctx)

	t.Run("When runtime is in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[:1])

		// WHEN
		t.Log("Request bundle instance auth deletion")
		output := graphql.BundleInstanceAuth{}
		err := runtimeConsumer.Run(bndlInstanceAuthDeletionRequestReq, oauthGraphQLClient, &output)

		// THEN
		require.NoError(t, err)
	})

	t.Run("When runtime is NOT in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[:1])

		// WHEN
		t.Log("Request bundle instance auth deletion")
		output := graphql.BundleInstanceAuth{}
		err := runtimeConsumer.Run(bndlInstanceAuthDeletionRequestReq, oauthGraphQLClient, &output)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "The operation is not allowed")
	})
}

func TestSetBundleInstanceAuth(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenantId, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	bndlInstanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, dexGraphQLClient, bndl.ID)

	authInput := fixtures.FixBasicAuth(t)
	bndlInstanceAuthSetInput := fixtures.FixBundleInstanceAuthSetInputSucceeded(authInput)
	bndlInstanceAuthSetInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthSetInputToGQL(bndlInstanceAuthSetInput)
	require.NoError(t, err)

	setBundleInstanceAuthReq := fixtures.FixSetBundleInstanceAuthRequest(bndlInstanceAuth.ID, bndlInstanceAuthSetInputStr)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Set bundle instance auth")
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, setBundleInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)
	require.Equal(t, graphql.BundleInstanceAuthStatusConditionSucceeded, output.Status.Condition)
	assertions.AssertAuth(t, authInput, output.Auth)

	saveExample(t, setBundleInstanceAuthReq.Query(), "set bundle instance auth")
}

func TestDeleteBundleInstanceAuth(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenantId, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	bndlInstanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, dexGraphQLClient, bndl.ID)

	deleteBundleInstanceAuthReq := fixtures.FixDeleteBundleInstanceAuthRequest(bndlInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Delete bundle instance auth")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteBundleInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)

	saveExample(t, deleteBundleInstanceAuthReq.Query(), "delete bundle instance auth")
}
