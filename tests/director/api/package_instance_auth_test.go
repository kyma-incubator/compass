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

	pkg := createBundle(t, ctx, application.ID, "pkg-app-1")
	defer deleteBundle(t, ctx, pkg.ID)

	authCtx, inputParams := fixBundleInstanceAuthContextAndInputParams(t)
	pkgInstanceAuthRequestInput := fixBundleInstanceAuthRequestInput(authCtx, inputParams)
	pkgInstanceAuthRequestInputStr, err := tc.graphqlizer.BundleInstanceAuthRequestInputToGQL(pkgInstanceAuthRequestInput)
	require.NoError(t, err)

	pkgInstanceAuthCreationRequestReq := fixRequestBundleInstanceAuthCreationRequest(pkg.ID, pkgInstanceAuthRequestInputStr)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth creation")
	err = tc.RunOperation(ctx, pkgInstanceAuthCreationRequestReq, &output)

	// THEN
	require.NoError(t, err)
	assertBundleInstanceAuth(t, pkgInstanceAuthRequestInput, output)

	saveExample(t, pkgInstanceAuthCreationRequestReq.Query(), "request bundle instance auth creation")

	// Fetch Application with bundles
	bundlesForApplicationReq := fixBundlesRequest(application.ID)
	pkgFromAPI := graphql.ApplicationExt{}

	err = tc.RunOperation(ctx, bundlesForApplicationReq, &pkgFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(pkgFromAPI.Bundles.Data))
	require.Equal(t, 1, len(pkgFromAPI.Bundles.Data[0].InstanceAuths))

	// Fetch Application with bundle
	instanceAuthID := pkgFromAPI.Bundles.Data[0].InstanceAuths[0].ID
	bundlesForApplicationWithInstanceAuthReq := fixBundleWithInstanceAuthRequest(application.ID, pkg.ID, instanceAuthID)

	err = tc.RunOperation(ctx, bundlesForApplicationWithInstanceAuthReq, &pkgFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auth exist
	require.Equal(t, instanceAuthID, pkgFromAPI.Bundle.InstanceAuth.ID)
	require.Equal(t, graphql.BundleInstanceAuthStatusConditionPending, pkgFromAPI.Bundle.InstanceAuth.Status.Condition)
}

func TestRequestBundleInstanceAuthCreationWithDefaultAuth(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	authInput := fixBasicAuth(t)

	pkgInput := fixBundleCreateInputWithDefaultAuth("pkg-app-1", authInput)
	pkg, err := tc.graphqlizer.BundleCreateInputToGQL(pkgInput)
	require.NoError(t, err)

	addPkgRequest := fixAddBundleRequest(application.ID, pkg)
	pkgAddOutput := graphql.Bundle{}

	err = tc.RunOperation(ctx, addPkgRequest, &pkgAddOutput)
	defer deleteBundle(t, ctx, pkgAddOutput.ID)
	require.NoError(t, err)

	pkgInstanceAuthRequestInput := fixBundleInstanceAuthRequestInput(nil, nil)
	pkgInstanceAuthRequestInputStr, err := tc.graphqlizer.BundleInstanceAuthRequestInputToGQL(pkgInstanceAuthRequestInput)
	require.NoError(t, err)

	pkgInstanceAuthCreationRequestReq := fixRequestBundleInstanceAuthCreationRequest(pkgAddOutput.ID, pkgInstanceAuthRequestInputStr)
	authOutput := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth creation")
	err = tc.RunOperation(ctx, pkgInstanceAuthCreationRequestReq, &authOutput)

	// THEN
	require.NoError(t, err)

	// Fetch Application with bundles
	bundlesForApplicationReq := fixBundlesRequest(application.ID)
	pkgFromAPI := graphql.ApplicationExt{}

	err = tc.RunOperation(ctx, bundlesForApplicationReq, &pkgFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(pkgFromAPI.Bundles.Data))
	require.Equal(t, 1, len(pkgFromAPI.Bundles.Data[0].InstanceAuths))

	// Fetch Application with bundle
	instanceAuthID := pkgFromAPI.Bundles.Data[0].InstanceAuths[0].ID
	bundlesForApplicationWithInstanceAuthReq := fixBundleWithInstanceAuthRequest(application.ID, pkgAddOutput.ID, instanceAuthID)

	err = tc.RunOperation(ctx, bundlesForApplicationWithInstanceAuthReq, &pkgFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auth exist
	require.Equal(t, instanceAuthID, pkgFromAPI.Bundle.InstanceAuth.ID)

	require.Equal(t, graphql.BundleInstanceAuthStatusConditionSucceeded, pkgFromAPI.Bundle.InstanceAuth.Status.Condition)
	assertAuth(t, authInput, pkgFromAPI.Bundle.InstanceAuth.Auth)
}

func TestRequestBundleInstanceAuthDeletion(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	pkg := createBundle(t, ctx, application.ID, "pkg-app-1")
	defer deleteBundle(t, ctx, pkg.ID)

	pkgInstanceAuth := createBundleInstanceAuth(t, ctx, pkg.ID)

	pkgInstanceAuthDeletionRequestReq := fixRequestBundleInstanceAuthDeletionRequest(pkgInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth deletion")
	err := tc.RunOperation(ctx, pkgInstanceAuthDeletionRequestReq, &output)

	// THEN
	require.NoError(t, err)

	saveExample(t, pkgInstanceAuthDeletionRequestReq.Query(), "request bundle instance auth deletion")
}

func TestSetBundleInstanceAuth(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	pkg := createBundle(t, ctx, application.ID, "pkg-app-1")
	defer deleteBundle(t, ctx, pkg.ID)

	pkgInstanceAuth := createBundleInstanceAuth(t, ctx, pkg.ID)

	authInput := fixBasicAuth(t)
	pkgInstanceAuthSetInput := fixBundleInstanceAuthSetInputSucceeded(authInput)
	pkgInstanceAuthSetInputStr, err := tc.graphqlizer.BundleInstanceAuthSetInputToGQL(pkgInstanceAuthSetInput)
	require.NoError(t, err)

	setBundleInstanceAuthReq := fixSetBundleInstanceAuthRequest(pkgInstanceAuth.ID, pkgInstanceAuthSetInputStr)
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

	pkg := createBundle(t, ctx, application.ID, "pkg-app-1")
	defer deleteBundle(t, ctx, pkg.ID)

	pkgInstanceAuth := createBundleInstanceAuth(t, ctx, pkg.ID)

	deleteBundleInstanceAuthReq := fixDeleteBundleInstanceAuthRequest(pkgInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Delete bundle instance auth")
	err := tc.RunOperation(ctx, deleteBundleInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)

	saveExample(t, deleteBundleInstanceAuthReq.Query(), "delete bundle instance auth")
}
