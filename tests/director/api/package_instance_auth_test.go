package api

import (
	"context"
	"github.com/kyma-incubator/compass/tests/director/pkg/jwtbuilder"
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

func TestRequestPackageInstanceAuthCreationAsRuntimeConsumer(t *testing.T) {
	ctx := context.Background()

	runtime := registerRuntime(t, ctx, "runtime-test")
	defer unregisterRuntime(t, runtime.ID)

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

		t.Log("Request package instance auth creation")
		err = runtimeConsumer.Run(pkgInstanceAuthCreationRequestReq, &output)

		// THEN
		require.NoError(t, err)
		assertPackageInstanceAuth(t, pkgInstanceAuthRequestInput, output)

		// Fetch Application with packages
		packagesForApplicationReq := fixPackagesRequest(application.ID)
		pkgFromAPI := graphql.ApplicationExt{}

		err = runtimeConsumer.Run(packagesForApplicationReq, &pkgFromAPI)
		require.NoError(t, err)

		// Assert the package instance auths exists
		require.Equal(t, 1, len(pkgFromAPI.Packages.Data))
		require.Equal(t, 1, len(pkgFromAPI.Packages.Data[0].InstanceAuths))

		// Fetch Application with package instance auth
		instanceAuthID := pkgFromAPI.Packages.Data[0].InstanceAuths[0].ID
		packagesForApplicationWithInstanceAuthReq := fixPackageWithInstanceAuthRequest(application.ID, pkg.ID, instanceAuthID)

		err = runtimeConsumer.Run(packagesForApplicationWithInstanceAuthReq, &pkgFromAPI)
		require.NoError(t, err)

		// Assert the package instance auth exist
		require.Equal(t, instanceAuthID, pkgFromAPI.Package.InstanceAuth.ID)
		require.Equal(t, graphql.PackageInstanceAuthStatusConditionPending, pkgFromAPI.Package.InstanceAuth.Status.Condition)
	})

	t.Run("When runtime is NOT in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		setApplicationLabel(t, ctx, application.ID, scenariosLabel, scenarios[:1])

		// set runtime scenarios label
		setRuntimeLabel(t, ctx, runtime.ID, scenariosLabel, scenarios[1:])
		defer setRuntimeLabel(t, ctx, runtime.ID, scenariosLabel, scenarios[:1])

		t.Log("Request package instance auth creation")
		err = runtimeConsumer.Run(pkgInstanceAuthCreationRequestReq, &output)

		// THEN
		require.Error(t, err)
	})
}

func TestRequestPackageInstanceAuthCreationWithDefaultAuth(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-package")
	defer unregisterApplication(t, application.ID)

	authInput := fixBasicAuth(t)

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

func TestRequestPackageInstanceAuthDeletionAsRuntimeConsumer(t *testing.T) {
	ctx := context.Background()

	runtime := registerRuntime(t, ctx, "runtime-test")
	defer unregisterRuntime(t, runtime.ID)

	application := registerApplication(t, ctx, "app-test-package")
	defer unregisterApplication(t, application.ID)

	pkg := createPackage(t, ctx, application.ID, "pkg-app-1")
	defer deletePackage(t, ctx, pkg.ID)

	pkgInstanceAuth := createPackageInstanceAuth(t, ctx, pkg.ID)

	pkgInstanceAuthDeletionRequestReq := fixRequestPackageInstanceAuthDeletionRequest(pkgInstanceAuth.ID)

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
		t.Log("Request package instance auth deletion")
		output := graphql.PackageInstanceAuth{}
		err := runtimeConsumer.Run(pkgInstanceAuthDeletionRequestReq, &output)

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
		t.Log("Request package instance auth deletion")
		output := graphql.PackageInstanceAuth{}
		err := runtimeConsumer.Run(pkgInstanceAuthDeletionRequestReq, &output)

		// THEN
		require.Error(t, err)
	})
}

func TestSetPackageInstanceAuth(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-package")
	defer unregisterApplication(t, application.ID)

	pkg := createPackage(t, ctx, application.ID, "pkg-app-1")
	defer deletePackage(t, ctx, pkg.ID)

	pkgInstanceAuth := createPackageInstanceAuth(t, ctx, pkg.ID)

	authInput := fixBasicAuth(t)
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
