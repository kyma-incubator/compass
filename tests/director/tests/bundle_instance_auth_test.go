package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/jwtbuilder"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func TestRequestBundleInstanceAuthCreation(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bndl := createBundle(t, ctx, application.ID, "bndl-app-1")
	defer deleteBundle(t, ctx, bndl.ID)

	authCtx, inputParams := fixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := fixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := tc.graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixRequestBundleInstanceAuthCreationRequest(bndl.ID, bndlInstanceAuthRequestInputStr)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth creation")
	err = tc.RunOperation(ctx, bndlInstanceAuthCreationRequestReq, &output)

	// THEN
	require.NoError(t, err)
	assertBundleInstanceAuthInput(t, bndlInstanceAuthRequestInput, output)

	saveExample(t, bndlInstanceAuthCreationRequestReq.Query(), "request bundle instance auth creation")

	// Fetch Application with bundles
	bundlesForApplicationReq := fixBundlesRequest(application.ID)
	bndlFromAPI := graphql.ApplicationExt{}

	err = tc.RunOperation(ctx, bundlesForApplicationReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data))
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data[0].InstanceAuths))

	// Fetch Application with bundle
	instanceAuthID := bndlFromAPI.Bundles.Data[0].InstanceAuths[0].ID
	bundlesForApplicationWithInstanceAuthReq := fixBundleWithInstanceAuthRequest(application.ID, bndl.ID, instanceAuthID)

	err = tc.RunOperation(ctx, bundlesForApplicationWithInstanceAuthReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auth exist
	require.Equal(t, instanceAuthID, bndlFromAPI.Bundle.InstanceAuth.ID)
	require.Equal(t, graphql.BundleInstanceAuthStatusConditionPending, bndlFromAPI.Bundle.InstanceAuth.Status.Condition)
}

func TestRequestBundleInstanceAuthCreationAsRuntimeConsumer(t *testing.T) {
	ctx := context.Background()

	runtime := registerRuntime(t, ctx, "runtime-test")
	defer unregisterRuntime(t, runtime.ID)

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bndl := createBundle(t, ctx, application.ID, "bndl-app-1")
	defer deleteBundle(t, ctx, bndl.ID)

	authCtx, inputParams := fixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := fixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := tc.graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixRequestBundleInstanceAuthCreationRequest(bndl.ID, bndlInstanceAuthRequestInputStr)
	output := graphql.BundleInstanceAuth{}

	scenarios := []string{defaultScenario, "test-scenario"}
	updateScenariosLabelDefinitionWithinTenant(t, ctx, testTenants.GetDefaultTenantID(), scenarios)
	defer updateScenariosLabelDefinitionWithinTenant(t, ctx, testTenants.GetDefaultTenantID(), scenarios[:1])

	runtimeConsumer := tc.NewOperation(ctx).WithConsumer(&jwtbuilder.Consumer{
		ID:   runtime.ID,
		Type: jwtbuilder.RuntimeConsumer,
	})

	t.Run("When runtime is in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		setApplicationLabel(t, ctx, application.ID, scenariosLabel, scenarios[1:])
		defer setApplicationLabel(t, ctx, application.ID, scenariosLabel, scenarios[:1])

		// set runtime scenarios label
		setRuntimeLabel(t, ctx, runtime.ID, scenariosLabel, scenarios[1:])
		defer setRuntimeLabel(t, ctx, runtime.ID, scenariosLabel, scenarios[:1])

		t.Log("Request bundle instance auth creation")
		err = runtimeConsumer.Run(bndlInstanceAuthCreationRequestReq, &output)

		// THEN
		require.NoError(t, err)
		assertBundleInstanceAuthInput(t, bndlInstanceAuthRequestInput, output)

		// Fetch Application with bundles
		bundlesForApplicationReq := fixBundlesRequest(application.ID)
		bndlFromAPI := graphql.ApplicationExt{}

		err = runtimeConsumer.Run(bundlesForApplicationReq, &bndlFromAPI)
		require.NoError(t, err)

		// Assert the bundle instance auths exists
		require.Equal(t, 1, len(bndlFromAPI.Bundles.Data))
		require.Equal(t, 1, len(bndlFromAPI.Bundles.Data[0].InstanceAuths))

		// Fetch Application with bundle instance auth
		instanceAuthID := bndlFromAPI.Bundles.Data[0].InstanceAuths[0].ID
		bundlesForApplicationWithInstanceAuthReq := fixBundleWithInstanceAuthRequest(application.ID, bndl.ID, instanceAuthID)

		err = runtimeConsumer.Run(bundlesForApplicationWithInstanceAuthReq, &bndlFromAPI)
		require.NoError(t, err)

		// Assert the bundle instance auth exist
		require.Equal(t, instanceAuthID, bndlFromAPI.Bundle.InstanceAuth.ID)
		require.Equal(t, graphql.BundleInstanceAuthStatusConditionPending, bndlFromAPI.Bundle.InstanceAuth.Status.Condition)
	})

	t.Run("When runtime is NOT in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		setApplicationLabel(t, ctx, application.ID, scenariosLabel, scenarios[:1])

		// set runtime scenarios label
		setRuntimeLabel(t, ctx, runtime.ID, scenariosLabel, scenarios[1:])
		defer setRuntimeLabel(t, ctx, runtime.ID, scenariosLabel, scenarios[:1])

		t.Log("Request bundle instance auth creation")
		err = runtimeConsumer.Run(bndlInstanceAuthCreationRequestReq, &output)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "The operation is not allowed")
	})
}

func TestRequestBundleInstanceAuthCreationWithDefaultAuth(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	authInput := fixBasicAuth(t)

	bndlInput := fixBundleCreateInputWithDefaultAuth("bndl-app-1", authInput)
	bndl, err := tc.graphqlizer.BundleCreateInputToGQL(bndlInput)
	require.NoError(t, err)

	addBndlRequest := fixAddBundleRequest(application.ID, bndl)
	bndlAddOutput := graphql.Bundle{}

	err = tc.RunOperation(ctx, addBndlRequest, &bndlAddOutput)
	defer deleteBundle(t, ctx, bndlAddOutput.ID)
	require.NoError(t, err)

	bndlInstanceAuthRequestInput := fixBundleInstanceAuthRequestInput(nil, nil)
	bndlInstanceAuthRequestInputStr, err := tc.graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixRequestBundleInstanceAuthCreationRequest(bndlAddOutput.ID, bndlInstanceAuthRequestInputStr)
	authOutput := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth creation")
	err = tc.RunOperation(ctx, bndlInstanceAuthCreationRequestReq, &authOutput)

	// THEN
	require.NoError(t, err)

	// Fetch Application with bundles
	bundlesForApplicationReq := fixBundlesRequest(application.ID)
	bndlFromAPI := graphql.ApplicationExt{}

	err = tc.RunOperation(ctx, bundlesForApplicationReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data))
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data[0].InstanceAuths))

	// Fetch Application with bundle
	instanceAuthID := bndlFromAPI.Bundles.Data[0].InstanceAuths[0].ID
	bundlesForApplicationWithInstanceAuthReq := fixBundleWithInstanceAuthRequest(application.ID, bndlAddOutput.ID, instanceAuthID)

	err = tc.RunOperation(ctx, bundlesForApplicationWithInstanceAuthReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auth exist
	require.Equal(t, instanceAuthID, bndlFromAPI.Bundle.InstanceAuth.ID)

	require.Equal(t, graphql.BundleInstanceAuthStatusConditionSucceeded, bndlFromAPI.Bundle.InstanceAuth.Status.Condition)
	assertAuth(t, authInput, bndlFromAPI.Bundle.InstanceAuth.Auth)
}

func TestRequestBundleInstanceAuthDeletion(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bndl := createBundle(t, ctx, application.ID, "bndl-app-1")
	defer deleteBundle(t, ctx, bndl.ID)

	bndlInstanceAuth := createBundleInstanceAuth(t, ctx, bndl.ID)

	bndlInstanceAuthDeletionRequestReq := fixRequestBundleInstanceAuthDeletionRequest(bndlInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth deletion")
	err := tc.RunOperation(ctx, bndlInstanceAuthDeletionRequestReq, &output)

	// THEN
	require.NoError(t, err)

	saveExample(t, bndlInstanceAuthDeletionRequestReq.Query(), "request bundle instance auth deletion")
}

func TestRequestBundleInstanceAuthDeletionAsRuntimeConsumer(t *testing.T) {
	ctx := context.Background()

	runtime := registerRuntime(t, ctx, "runtime-test")
	defer unregisterRuntime(t, runtime.ID)

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bndl := createBundle(t, ctx, application.ID, "bndl-app-1")
	defer deleteBundle(t, ctx, bndl.ID)

	bndlInstanceAuth := createBundleInstanceAuth(t, ctx, bndl.ID)

	bndlInstanceAuthDeletionRequestReq := fixRequestBundleInstanceAuthDeletionRequest(bndlInstanceAuth.ID)

	scenarios := []string{defaultScenario, "test-scenario"}
	updateScenariosLabelDefinitionWithinTenant(t, ctx, testTenants.GetDefaultTenantID(), scenarios)
	defer updateScenariosLabelDefinitionWithinTenant(t, ctx, testTenants.GetDefaultTenantID(), scenarios[:1])

	runtimeConsumer := tc.NewOperation(ctx).WithConsumer(&jwtbuilder.Consumer{
		ID:   runtime.ID,
		Type: jwtbuilder.RuntimeConsumer,
	})

	t.Run("When runtime is in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		setApplicationLabel(t, ctx, application.ID, scenariosLabel, scenarios[1:])
		defer setApplicationLabel(t, ctx, application.ID, scenariosLabel, scenarios[:1])

		// set runtime scenarios label
		setRuntimeLabel(t, ctx, runtime.ID, scenariosLabel, scenarios[1:])
		defer setRuntimeLabel(t, ctx, runtime.ID, scenariosLabel, scenarios[:1])

		// WHEN
		t.Log("Request bundle instance auth deletion")
		output := graphql.BundleInstanceAuth{}
		err := runtimeConsumer.Run(bndlInstanceAuthDeletionRequestReq, &output)

		// THEN
		require.NoError(t, err)
	})

	t.Run("When runtime is NOT in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		setApplicationLabel(t, ctx, application.ID, scenariosLabel, scenarios[:1])

		// set runtime scenarios label
		setRuntimeLabel(t, ctx, runtime.ID, scenariosLabel, scenarios[1:])
		defer setRuntimeLabel(t, ctx, runtime.ID, scenariosLabel, scenarios[:1])

		// WHEN
		t.Log("Request bundle instance auth deletion")
		output := graphql.BundleInstanceAuth{}
		err := runtimeConsumer.Run(bndlInstanceAuthDeletionRequestReq, &output)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "The operation is not allowed")
	})
}

func TestSetBundleInstanceAuth(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bndl := createBundle(t, ctx, application.ID, "bndl-app-1")
	defer deleteBundle(t, ctx, bndl.ID)

	bndlInstanceAuth := createBundleInstanceAuth(t, ctx, bndl.ID)

	authInput := fixBasicAuth(t)
	bndlInstanceAuthSetInput := fixBundleInstanceAuthSetInputSucceeded(authInput)
	bndlInstanceAuthSetInputStr, err := tc.graphqlizer.BundleInstanceAuthSetInputToGQL(bndlInstanceAuthSetInput)
	require.NoError(t, err)

	setBundleInstanceAuthReq := fixSetBundleInstanceAuthRequest(bndlInstanceAuth.ID, bndlInstanceAuthSetInputStr)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Set bundle instance auth")
	err = tc.RunOperation(ctx, setBundleInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)
	require.Equal(t, graphql.BundleInstanceAuthStatusConditionSucceeded, output.Status.Condition)
	assertAuth(t, authInput, output.Auth)

	saveExample(t, setBundleInstanceAuthReq.Query(), "set bundle instance auth")
}

func TestDeleteBundleInstanceAuth(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bndl := createBundle(t, ctx, application.ID, "bndl-app-1")
	defer deleteBundle(t, ctx, bndl.ID)

	bndlInstanceAuth := createBundleInstanceAuth(t, ctx, bndl.ID)

	deleteBundleInstanceAuthReq := fixDeleteBundleInstanceAuthRequest(bndlInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Delete bundle instance auth")
	err := tc.RunOperation(ctx, deleteBundleInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)

	saveExample(t, deleteBundleInstanceAuthReq.Query(), "delete bundle instance auth")
}
