package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/jwtbuilder"

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

	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	authCtx, inputParams := pkg.FixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := pkg.FixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := pkg.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := pkg.FixRequestBundleInstanceAuthCreationRequest(bndl.ID, bndlInstanceAuthRequestInputStr)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth creation")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, bndlInstanceAuthCreationRequestReq, &output)

	// THEN
	require.NoError(t, err)
	assertBundleInstanceAuthInput(t, bndlInstanceAuthRequestInput, output)

	saveExample(t, bndlInstanceAuthCreationRequestReq.Query(), "request bundle instance auth creation")

	// Fetch Application with bundles
	bundlesForApplicationReq := pkg.FixGetBundlesRequest(application.ID)
	bndlFromAPI := graphql.ApplicationExt{}

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, bundlesForApplicationReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data))
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data[0].InstanceAuths))

	// Fetch Application with bundle
	instanceAuthID := bndlFromAPI.Bundles.Data[0].InstanceAuths[0].ID
	bundlesForApplicationWithInstanceAuthReq := pkg.FixGetBundleWithInstanceAuthRequest(application.ID, bndl.ID, instanceAuthID)

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, bundlesForApplicationWithInstanceAuthReq, &bndlFromAPI)
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

	input := pkg.FixRuntimeInput("runtime-test")

	runtime := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)

	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	authCtx, inputParams := pkg.FixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := pkg.FixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := pkg.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := pkg.FixRequestBundleInstanceAuthCreationRequest(bndl.ID, bndlInstanceAuthRequestInputStr)
	output := graphql.BundleInstanceAuth{}

	scenarios := []string{defaultScenario, "test-scenario"}
	pkg.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant, scenarios)
	defer pkg.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant, scenarios[:1])

	runtimeConsumer := pkg.Tc.NewOperation(ctx).WithConsumer(&jwtbuilder.Consumer{
		ID:   runtime.ID,
		Type: jwtbuilder.RuntimeConsumer,
	})

	t.Run("When runtime is in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		pkg.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[1:])
		defer pkg.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[1:])
		defer pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[:1])

		t.Log("Request bundle instance auth creation")
		err = runtimeConsumer.Run(bndlInstanceAuthCreationRequestReq, dexGraphQLClient, &output)

		// THEN
		require.NoError(t, err)
		assertBundleInstanceAuthInput(t, bndlInstanceAuthRequestInput, output)

		// Fetch Application with bundles
		bundlesForApplicationReq := pkg.FixGetBundlesRequest(application.ID)
		bndlFromAPI := graphql.ApplicationExt{}

		err = runtimeConsumer.Run(bundlesForApplicationReq, dexGraphQLClient, &bndlFromAPI)
		require.NoError(t, err)

		// Assert the bundle instance auths exists
		require.Equal(t, 1, len(bndlFromAPI.Bundles.Data))
		require.Equal(t, 1, len(bndlFromAPI.Bundles.Data[0].InstanceAuths))

		// Fetch Application with bundle instance auth
		instanceAuthID := bndlFromAPI.Bundles.Data[0].InstanceAuths[0].ID
		bundlesForApplicationWithInstanceAuthReq := pkg.FixGetBundleWithInstanceAuthRequest(application.ID, bndl.ID, instanceAuthID)

		err = runtimeConsumer.Run(bundlesForApplicationWithInstanceAuthReq, dexGraphQLClient, &bndlFromAPI)
		require.NoError(t, err)

		// Assert the bundle instance auth exist
		require.Equal(t, instanceAuthID, bndlFromAPI.Bundle.InstanceAuth.ID)
		require.Equal(t, graphql.BundleInstanceAuthStatusConditionPending, bndlFromAPI.Bundle.InstanceAuth.Status.Condition)
	})

	t.Run("When runtime is NOT in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		pkg.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[1:])
		defer pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[:1])

		t.Log("Request bundle instance auth creation")
		err = runtimeConsumer.Run(bndlInstanceAuthCreationRequestReq, dexGraphQLClient, &output)

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

	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	authInput := pkg.FixBasicAuth(t)

	bndlInput := pkg.FixBundleCreateInputWithDefaultAuth("bndl-app-1", authInput)
	bndl, err := pkg.Tc.Graphqlizer.BundleCreateInputToGQL(bndlInput)
	require.NoError(t, err)

	addBndlRequest := pkg.FixAddBundleRequest(application.ID, bndl)
	bndlAddOutput := graphql.Bundle{}

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, addBndlRequest, &bndlAddOutput)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndlAddOutput.ID)
	require.NoError(t, err)

	bndlInstanceAuthRequestInput := pkg.FixBundleInstanceAuthRequestInput(nil, nil)
	bndlInstanceAuthRequestInputStr, err := pkg.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := pkg.FixRequestBundleInstanceAuthCreationRequest(bndlAddOutput.ID, bndlInstanceAuthRequestInputStr)
	authOutput := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth creation")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, bndlInstanceAuthCreationRequestReq, &authOutput)

	// THEN
	require.NoError(t, err)

	// Fetch Application with bundles
	bundlesForApplicationReq := pkg.FixGetBundlesRequest(application.ID)
	bndlFromAPI := graphql.ApplicationExt{}

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, bundlesForApplicationReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data))
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data[0].InstanceAuths))

	// Fetch Application with bundle
	instanceAuthID := bndlFromAPI.Bundles.Data[0].InstanceAuths[0].ID
	bundlesForApplicationWithInstanceAuthReq := pkg.FixGetBundleWithInstanceAuthRequest(application.ID, bndlAddOutput.ID, instanceAuthID)

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, bundlesForApplicationWithInstanceAuthReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auth exist
	require.Equal(t, instanceAuthID, bndlFromAPI.Bundle.InstanceAuth.ID)

	require.Equal(t, graphql.BundleInstanceAuthStatusConditionSucceeded, bndlFromAPI.Bundle.InstanceAuth.Status.Condition)
	assertAuth(t, authInput, bndlFromAPI.Bundle.InstanceAuth.Auth)
}

func TestRequestBundleInstanceAuthDeletion(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	bndlInstanceAuth := pkg.CreateBundleInstanceAuth(t, ctx, dexGraphQLClient, bndl.ID)

	bndlInstanceAuthDeletionRequestReq := pkg.FixRequestBundleInstanceAuthDeletionRequest(bndlInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth deletion")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, bndlInstanceAuthDeletionRequestReq, &output)

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

	input := pkg.FixRuntimeInput("runtime-test")

	runtime := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)

	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	bndlInstanceAuth := pkg.CreateBundleInstanceAuth(t, ctx, dexGraphQLClient, bndl.ID)

	bndlInstanceAuthDeletionRequestReq := pkg.FixRequestBundleInstanceAuthDeletionRequest(bndlInstanceAuth.ID)

	scenarios := []string{defaultScenario, "test-scenario"}
	pkg.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant, scenarios)
	defer pkg.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant, scenarios[:1])

	runtimeConsumer := pkg.Tc.NewOperation(ctx).WithConsumer(&jwtbuilder.Consumer{
		ID:   runtime.ID,
		Type: jwtbuilder.RuntimeConsumer,
	})

	t.Run("When runtime is in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		pkg.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[1:])
		defer pkg.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[1:])
		defer pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[:1])

		// WHEN
		t.Log("Request bundle instance auth deletion")
		output := graphql.BundleInstanceAuth{}
		err := runtimeConsumer.Run(bndlInstanceAuthDeletionRequestReq, dexGraphQLClient, &output)

		// THEN
		require.NoError(t, err)
	})

	t.Run("When runtime is NOT in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		pkg.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[1:])
		defer pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[:1])

		// WHEN
		t.Log("Request bundle instance auth deletion")
		output := graphql.BundleInstanceAuth{}
		err := runtimeConsumer.Run(bndlInstanceAuthDeletionRequestReq, dexGraphQLClient, &output)

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

	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	bndlInstanceAuth := pkg.CreateBundleInstanceAuth(t, ctx, dexGraphQLClient, bndl.ID)

	authInput := pkg.FixBasicAuth(t)
	bndlInstanceAuthSetInput := pkg.FixBundleInstanceAuthSetInputSucceeded(authInput)
	bndlInstanceAuthSetInputStr, err := pkg.Tc.Graphqlizer.BundleInstanceAuthSetInputToGQL(bndlInstanceAuthSetInput)
	require.NoError(t, err)

	setBundleInstanceAuthReq := pkg.FixSetBundleInstanceAuthRequest(bndlInstanceAuth.ID, bndlInstanceAuthSetInputStr)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Set bundle instance auth")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, setBundleInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)
	require.Equal(t, graphql.BundleInstanceAuthStatusConditionSucceeded, output.Status.Condition)
	assertAuth(t, authInput, output.Auth)

	saveExample(t, setBundleInstanceAuthReq.Query(), "set bundle instance auth")
}

func TestDeleteBundleInstanceAuth(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	bndlInstanceAuth := pkg.CreateBundleInstanceAuth(t, ctx, dexGraphQLClient, bndl.ID)

	deleteBundleInstanceAuthReq := pkg.FixDeleteBundleInstanceAuthRequest(bndlInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Delete bundle instance auth")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, deleteBundleInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)

	saveExample(t, deleteBundleInstanceAuthReq.Query(), "delete bundle instance auth")
}
