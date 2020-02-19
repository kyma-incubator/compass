package api

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func TestRequestPackageInstanceAuthCreation(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-package")
	defer unregisterApplication(t, application.ID)

	pkg := createPackage(t, ctx, application.ID, "pkg-app-1")
	defer deletePackage(t, ctx, pkg.ID)

	authCtx, inputParams := fixPackageInstanceAuthContextAndInputParams(t)
	pkgInstanceAuthRequestInput := fixPackageInstanceAuthRequestInput(authCtx, inputParams)
	pkgInstanceAuthRequestInputStr, err := tc.graphqlizer.PackageInstanceAuthRequestInputToGQL(pkgInstanceAuthRequestInput)
	require.NoError(t, err)

	pkgInstanceAuthCreationRequestReq := fixRequestPackageInstanceAuthCreationRequest(pkg.ID, pkgInstanceAuthRequestInputStr)
	output := graphql.PackageInstanceAuth{}

	// WHEN
	t.Log("Request package instance auth creation")
	err = tc.RunOperation(ctx, pkgInstanceAuthCreationRequestReq, &output)

	// THEN
	require.NoError(t, err)
	assertPackageInstanceAuth(t, pkgInstanceAuthRequestInput, output)

	saveExample(t, pkgInstanceAuthCreationRequestReq.Query(), "request package instance auth creation")

	// Fetch Application with packages
	packagesForApplicationReq := fixPackagesRequest(application.ID)
	pkgFromAPI := graphql.ApplicationExt{}

	err = tc.RunOperation(ctx, packagesForApplicationReq, &pkgFromAPI)
	require.NoError(t, err)

	// Assert the package instance auths exists
	require.Equal(t, 1, len(pkgFromAPI.Packages.Data))
	require.Equal(t, 1, len(pkgFromAPI.Packages.Data[0].InstanceAuths))

	// Fetch Application with package
	instanceAuthID := pkgFromAPI.Packages.Data[0].InstanceAuths[0].ID
	packagesForApplicationWithInstanceAuthReq := fixPackageWithInstanceAuthRequest(application.ID, pkg.ID, instanceAuthID)

	err = tc.RunOperation(ctx, packagesForApplicationWithInstanceAuthReq, &pkgFromAPI)
	require.NoError(t, err)

	// Assert the package instance auth exist
	require.Equal(t, instanceAuthID, pkgFromAPI.Package.InstanceAuth.ID)
	require.Equal(t, graphql.PackageInstanceAuthStatusConditionPending, pkgFromAPI.Package.InstanceAuth.Status.Condition)
}

func TestRequestPackageInstanceAuthCreationWithDefaultAuth(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-package")
	defer unregisterApplication(t, application.ID)

	authInput := fixBasicAuth()

	pkgInput := fixPackageCreateInputWithDefaultAuth("pkg-app-1", authInput)
	pkg, err := tc.graphqlizer.PackageCreateInputToGQL(pkgInput)
	require.NoError(t, err)

	addPkgRequest := fixAddPackageRequest(application.ID, pkg)
	pkgAddOutput := graphql.Package{}

	err = tc.RunOperation(ctx, addPkgRequest, &pkgAddOutput)
	defer deletePackage(t, ctx, pkgAddOutput.ID)
	require.NoError(t, err)

	pkgInstanceAuthRequestInput := fixPackageInstanceAuthRequestInput(nil, nil)
	pkgInstanceAuthRequestInputStr, err := tc.graphqlizer.PackageInstanceAuthRequestInputToGQL(pkgInstanceAuthRequestInput)
	require.NoError(t, err)

	pkgInstanceAuthCreationRequestReq := fixRequestPackageInstanceAuthCreationRequest(pkgAddOutput.ID, pkgInstanceAuthRequestInputStr)
	authOutput := graphql.PackageInstanceAuth{}

	// WHEN
	t.Log("Request package instance auth creation")
	err = tc.RunOperation(ctx, pkgInstanceAuthCreationRequestReq, &authOutput)

	// THEN
	require.NoError(t, err)

	// Fetch Application with packages
	packagesForApplicationReq := fixPackagesRequest(application.ID)
	pkgFromAPI := graphql.ApplicationExt{}

	err = tc.RunOperation(ctx, packagesForApplicationReq, &pkgFromAPI)
	require.NoError(t, err)

	// Assert the package instance auths exists
	require.Equal(t, 1, len(pkgFromAPI.Packages.Data))
	require.Equal(t, 1, len(pkgFromAPI.Packages.Data[0].InstanceAuths))

	// Fetch Application with package
	instanceAuthID := pkgFromAPI.Packages.Data[0].InstanceAuths[0].ID
	packagesForApplicationWithInstanceAuthReq := fixPackageWithInstanceAuthRequest(application.ID, pkgAddOutput.ID, instanceAuthID)

	err = tc.RunOperation(ctx, packagesForApplicationWithInstanceAuthReq, &pkgFromAPI)
	require.NoError(t, err)

	// Assert the package instance auth exist
	require.Equal(t, instanceAuthID, pkgFromAPI.Package.InstanceAuth.ID)

	require.Equal(t, graphql.PackageInstanceAuthStatusConditionSucceeded, pkgFromAPI.Package.InstanceAuth.Status.Condition)
	assertAuth(t, authInput, pkgFromAPI.Package.InstanceAuth.Auth)
}

func TestRequestPackageInstanceAuthDeletion(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-package")
	defer unregisterApplication(t, application.ID)

	pkg := createPackage(t, ctx, application.ID, "pkg-app-1")
	defer deletePackage(t, ctx, pkg.ID)

	pkgInstanceAuth := createPackageInstanceAuth(t, ctx, pkg.ID)

	pkgInstanceAuthDeletionRequestReq := fixRequestPackageInstanceAuthDeletionRequest(pkgInstanceAuth.ID)
	output := graphql.PackageInstanceAuth{}

	// WHEN
	t.Log("Request package instance auth deletion")
	err := tc.RunOperation(ctx, pkgInstanceAuthDeletionRequestReq, &output)

	// THEN
	require.NoError(t, err)

	saveExample(t, pkgInstanceAuthDeletionRequestReq.Query(), "request package instance auth deletion")
}

func TestSetPackageInstanceAuth(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-package")
	defer unregisterApplication(t, application.ID)

	pkg := createPackage(t, ctx, application.ID, "pkg-app-1")
	defer deletePackage(t, ctx, pkg.ID)

	pkgInstanceAuth := createPackageInstanceAuth(t, ctx, pkg.ID)

	authInput := fixBasicAuth()
	pkgInstanceAuthSetInput := fixPackageInstanceAuthSetInputSucceeded(authInput)
	pkgInstanceAuthSetInputStr, err := tc.graphqlizer.PackageInstanceAuthSetInputToGQL(pkgInstanceAuthSetInput)
	require.NoError(t, err)

	setPackageInstanceAuthReq := fixSetPackageInstanceAuthRequest(pkgInstanceAuth.ID, pkgInstanceAuthSetInputStr)
	output := graphql.PackageInstanceAuth{}

	// WHEN
	t.Log("Set package instance auth")
	err = tc.RunOperation(ctx, setPackageInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)
	require.Equal(t, graphql.PackageInstanceAuthStatusConditionSucceeded, output.Status.Condition)
	assertAuth(t, authInput, output.Auth)

	saveExample(t, setPackageInstanceAuthReq.Query(), "set package instance auth")
}

func TestDeletePackageInstanceAuth(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-package")
	defer unregisterApplication(t, application.ID)

	pkg := createPackage(t, ctx, application.ID, "pkg-app-1")
	defer deletePackage(t, ctx, pkg.ID)

	pkgInstanceAuth := createPackageInstanceAuth(t, ctx, pkg.ID)

	deletePackageInstanceAuthReq := fixDeletePackageInstanceAuthRequest(pkgInstanceAuth.ID)
	output := graphql.PackageInstanceAuth{}

	// WHEN
	t.Log("Delete package instance auth")
	err := tc.RunOperation(ctx, deletePackageInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)

	saveExample(t, deletePackageInstanceAuthReq.Query(), "delete package instance auth")
}
