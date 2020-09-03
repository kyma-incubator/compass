package api

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func TestRequestBundleInstanceAuthCreation(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bundle := createBundle(t, ctx, application.ID, "bundle-app-1")
	defer deleteBundle(t, ctx, bundle.ID)

	authCtx, inputParams := fixBundleInstanceAuthContextAndInputParams(t)
	bundleInstanceAuthRequestInput := fixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bundleInstanceAuthRequestInputStr, err := tc.graphqlizer.BundleInstanceAuthRequestInputToGQL(bundleInstanceAuthRequestInput)
	require.NoError(t, err)

	bundleInstanceAuthCreationRequestReq := fixRequestBundleInstanceAuthCreationRequest(bundle.ID, bundleInstanceAuthRequestInputStr)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth creation")
	err = tc.RunOperation(ctx, bundleInstanceAuthCreationRequestReq, &output)

	// THEN
	require.NoError(t, err)
	assertBundleInstanceAuth(t, bundleInstanceAuthRequestInput, output)

	saveExample(t, bundleInstanceAuthCreationRequestReq.Query(), "request bundle instance auth creation")

	// Fetch Application with bundles
	bundlesForApplicationReq := fixBundlesRequest(application.ID)
	bundleFromAPI := graphql.ApplicationExt{}

	err = tc.RunOperation(ctx, bundlesForApplicationReq, &bundleFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(bundleFromAPI.Bundles.Data))
	require.Equal(t, 1, len(bundleFromAPI.Bundles.Data[0].InstanceAuths))

	// Fetch Application with bundle
	instanceAuthID := bundleFromAPI.Bundles.Data[0].InstanceAuths[0].ID
	bundlesForApplicationWithInstanceAuthReq := fixBundleWithInstanceAuthRequest(application.ID, bundle.ID, instanceAuthID)

	err = tc.RunOperation(ctx, bundlesForApplicationWithInstanceAuthReq, &bundleFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auth exist
	require.Equal(t, instanceAuthID, bundleFromAPI.Bundle.InstanceAuth.ID)
	require.Equal(t, graphql.BundleInstanceAuthStatusConditionPending, bundleFromAPI.Bundle.InstanceAuth.Status.Condition)
}

func TestRequestBundleInstanceAuthCreationWithDefaultAuth(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	authInput := fixBasicAuth(t)

	bundleInput := fixBundleCreateInputWithDefaultAuth("bundle-app-1", authInput)
	bundle, err := tc.graphqlizer.BundleCreateInputToGQL(bundleInput)
	require.NoError(t, err)

	addBundleRequest := fixAddBundleRequest(application.ID, bundle)
	bundleAddOutput := graphql.Bundle{}

	err = tc.RunOperation(ctx, addBundleRequest, &bundleAddOutput)
	defer deleteBundle(t, ctx, bundleAddOutput.ID)
	require.NoError(t, err)

	bundleInstanceAuthRequestInput := fixBundleInstanceAuthRequestInput(nil, nil)
	bundleInstanceAuthRequestInputStr, err := tc.graphqlizer.BundleInstanceAuthRequestInputToGQL(bundleInstanceAuthRequestInput)
	require.NoError(t, err)

	bundleInstanceAuthCreationRequestReq := fixRequestBundleInstanceAuthCreationRequest(bundleAddOutput.ID, bundleInstanceAuthRequestInputStr)
	authOutput := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth creation")
	err = tc.RunOperation(ctx, bundleInstanceAuthCreationRequestReq, &authOutput)

	// THEN
	require.NoError(t, err)

	// Fetch Application with bundles
	bundlesForApplicationReq := fixBundlesRequest(application.ID)
	bundleFromAPI := graphql.ApplicationExt{}

	err = tc.RunOperation(ctx, bundlesForApplicationReq, &bundleFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(bundleFromAPI.Bundles.Data))
	require.Equal(t, 1, len(bundleFromAPI.Bundles.Data[0].InstanceAuths))

	// Fetch Application with bundle
	instanceAuthID := bundleFromAPI.Bundles.Data[0].InstanceAuths[0].ID
	bundlesForApplicationWithInstanceAuthReq := fixBundleWithInstanceAuthRequest(application.ID, bundleAddOutput.ID, instanceAuthID)

	err = tc.RunOperation(ctx, bundlesForApplicationWithInstanceAuthReq, &bundleFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auth exist
	require.Equal(t, instanceAuthID, bundleFromAPI.Bundle.InstanceAuth.ID)

	require.Equal(t, graphql.BundleInstanceAuthStatusConditionSucceeded, bundleFromAPI.Bundle.InstanceAuth.Status.Condition)
	assertAuth(t, authInput, bundleFromAPI.Bundle.InstanceAuth.Auth)
}

func TestRequestBundleInstanceAuthDeletion(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bundle := createBundle(t, ctx, application.ID, "bundle-app-1")
	defer deleteBundle(t, ctx, bundle.ID)

	bundleInstanceAuth := createBundleInstanceAuth(t, ctx, bundle.ID)

	bundleInstanceAuthDeletionRequestReq := fixRequestBundleInstanceAuthDeletionRequest(bundleInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth deletion")
	err := tc.RunOperation(ctx, bundleInstanceAuthDeletionRequestReq, &output)

	// THEN
	require.NoError(t, err)

	saveExample(t, bundleInstanceAuthDeletionRequestReq.Query(), "request bundle instance auth deletion")
}

func TestSetBundleInstanceAuth(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bundle := createBundle(t, ctx, application.ID, "bundle-app-1")
	defer deleteBundle(t, ctx, bundle.ID)

	bundleInstanceAuth := createBundleInstanceAuth(t, ctx, bundle.ID)

	authInput := fixBasicAuth(t)
	bundleInstanceAuthSetInput := fixBundleInstanceAuthSetInputSucceeded(authInput)
	bundleInstanceAuthSetInputStr, err := tc.graphqlizer.BundleInstanceAuthSetInputToGQL(bundleInstanceAuthSetInput)
	require.NoError(t, err)

	setBundleInstanceAuthReq := fixSetBundleInstanceAuthRequest(bundleInstanceAuth.ID, bundleInstanceAuthSetInputStr)
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

	bundle := createBundle(t, ctx, application.ID, "bundle-app-1")
	defer deleteBundle(t, ctx, bundle.ID)

	bundleInstanceAuth := createBundleInstanceAuth(t, ctx, bundle.ID)

	deleteBundleInstanceAuthReq := fixDeleteBundleInstanceAuthRequest(bundleInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Delete bundle instance auth")
	err := tc.RunOperation(ctx, deleteBundleInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)

	saveExample(t, deleteBundleInstanceAuthReq.Query(), "delete bundle instance auth")
}
