package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func TestRequestBundleInstanceAuthCreation(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	input := fixtures.FixRuntimeInput("runtime-test")

	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	authCtx, inputParams := fixtures.FixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := fixtures.FixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixtures.FixRequestBundleInstanceAuthCreationRequest(bndl.ID, bndlInstanceAuthRequestInputStr)
	output := graphql.BundleInstanceAuth{}

	scenarios := []string{conf.DefaultScenario, "test-scenario"}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant, scenarios[:1])

	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, tenant, runtime.ID)
	rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
	require.NotEmpty(t, rtmOauthCredentialData.ClientID)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := pkg.GetAccessToken(t, rtmOauthCredentialData, "runtime:write application:read")
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	runtimeConsumer := testctx.Tc.NewOperation(ctx)

	t.Run("When runtime is in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[:1])

		t.Log("Request bundle instance auth creation")
		err = runtimeConsumer.Run(bndlInstanceAuthCreationRequestReq, oauthGraphQLClient, &output)

		// THEN
		require.NoError(t, err)
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
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[:1])

		t.Log("Request bundle instance auth creation")
		err = runtimeConsumer.Run(bndlInstanceAuthCreationRequestReq, oauthGraphQLClient, &output)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "The operation is not allowed")
	})
}

func TestRequestBundleInstanceAuthCreationWithDefaultAuth(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	authInput := fixtures.FixBasicAuth(t)

	bndlInput := fixtures.FixBundleCreateInputWithDefaultAuth("bndl-app-1", authInput)
	bndl, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(bndlInput)
	require.NoError(t, err)

	addBndlRequest := fixtures.FixAddBundleRequest(application.ID, bndl)
	bndlAddOutput := graphql.Bundle{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, addBndlRequest, &bndlAddOutput)
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndlAddOutput.ID)
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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	bndlInstanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, dexGraphQLClient, bndl.ID)

	bndlInstanceAuthDeletionRequestReq := fixtures.FixRequestBundleInstanceAuthDeletionRequest(bndlInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth deletion")
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, bndlInstanceAuthDeletionRequestReq, &output)

	// THEN
	require.NoError(t, err)

	saveExample(t, bndlInstanceAuthDeletionRequestReq.Query(), "request bundle instance auth deletion")
}

func TestRequestBundleInstanceAuthDeletionAsRuntimeConsumer(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	input := fixtures.FixRuntimeInput("runtime-test")

	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	bndlInstanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, dexGraphQLClient, bndl.ID)

	bndlInstanceAuthDeletionRequestReq := fixtures.FixRequestBundleInstanceAuthDeletionRequest(bndlInstanceAuth.ID)

	scenarios := []string{conf.DefaultScenario, "test-scenario"}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant, scenarios[:1])

	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, tenant, runtime.ID)
	rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
	require.NotEmpty(t, rtmOauthCredentialData.ClientID)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := pkg.GetAccessToken(t, rtmOauthCredentialData, "runtime:write")
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	runtimeConsumer := testctx.Tc.NewOperation(ctx)

	t.Run("When runtime is in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[:1])

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
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[:1])

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	bndlInstanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, dexGraphQLClient, bndl.ID)

	deleteBundleInstanceAuthReq := fixtures.FixDeleteBundleInstanceAuthRequest(bndlInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Delete bundle instance auth")
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteBundleInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)

	saveExample(t, deleteBundleInstanceAuthReq.Query(), "delete bundle instance auth")
}
