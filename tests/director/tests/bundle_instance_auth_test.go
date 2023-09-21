package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	json2 "github.com/kyma-incubator/compass/tests/pkg/json"
	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

var (
	appName    = "app-test-bundle"
	bundleName = "bndl-app-1"
	rtmName    = "rtm"
)

func TestRequestBundleInstanceAuthCreation(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bundleName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	authCtx, inputParams := fixtures.FixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := fixtures.FixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixtures.FixRequestBundleInstanceAuthCreationRequest(bndl.ID, bndlInstanceAuthRequestInputStr)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth creation")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, bndlInstanceAuthCreationRequestReq, &output)

	// THEN
	require.NoError(t, err)
	require.Nil(t, output.RuntimeID)
	require.Nil(t, output.RuntimeContextID)
	assertions.AssertBundleInstanceAuthInput(t, bndlInstanceAuthRequestInput, output)

	SaveExample(t, bndlInstanceAuthCreationRequestReq.Query(), "request bundle instance auth creation")

	// Fetch Application with bundles
	bundlesForApplicationReq := fixtures.FixGetBundlesRequest(application.ID)
	bndlFromAPI := graphql.ApplicationExt{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, bundlesForApplicationReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data))
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data[0].InstanceAuths))

	// Fetch Application with bundle
	instanceAuthID := bndlFromAPI.Bundles.Data[0].InstanceAuths[0].ID
	bundlesForApplicationWithInstanceAuthReq := fixtures.FixGetBundleWithInstanceAuthRequest(application.ID, bndl.ID, instanceAuthID)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, bundlesForApplicationWithInstanceAuthReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auth exist
	require.Equal(t, instanceAuthID, bndlFromAPI.Bundle.InstanceAuth.ID)
	require.Equal(t, graphql.BundleInstanceAuthStatusConditionPending, bndlFromAPI.Bundle.InstanceAuth.Status.Condition)
}

func TestRequestBundleInstanceAuthCreationAsRuntimeConsumer(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixRuntimeInput("runtime-test")

	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, input, conf.GatewayOauth)

	application, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, appName, conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bundleName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	authCtx, inputParams := fixtures.FixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := fixtures.FixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixtures.FixRequestBundleInstanceAuthCreationRequest(bndl.ID, bndlInstanceAuthRequestInputStr)
	output := graphql.BundleInstanceAuth{}

	testScenarioSecond := "test-scenario-2"

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenario)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenario)

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenarioSecond)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenarioSecond)

	formation1 := graphql.FormationInput{Name: testScenario}
	formation2 := graphql.FormationInput{Name: testScenarioSecond}

	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), certSecuredGraphQLClient, tenantId, runtime.ID)
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
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formation1, application.ID, tenantId)
		fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formation1, application.ID, tenantId)

		// set runtime scenarios label
		defer fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, formation1, runtime.ID, tenantId)
		fixtures.AssignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, formation1, runtime.ID, tenantId)

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
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formation2, application.ID, tenantId)
		fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formation2, application.ID, tenantId)

		// set runtime scenarios label
		defer fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, formation1, runtime.ID, tenantId)
		fixtures.AssignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, formation1, runtime.ID, tenantId)

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

	input := fixRuntimeInput("runtime-test")

	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, input, conf.GatewayOauth)

	application, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, appName, conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	authInput := fixtures.FixOauthAuth(t)
	bndlInput := fixtures.FixBundleCreateInputWithDefaultAuth(bundleName, authInput)
	bndl, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(bndlInput)
	require.NoError(t, err)

	addBndlRequest := fixtures.FixAddBundleRequest(application.ID, bndl)
	bndlAddOutput := graphql.Bundle{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addBndlRequest, &bndlAddOutput)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndlAddOutput.ID)
	require.NoError(t, err)

	authCtx, inputParams := fixtures.FixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := fixtures.FixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixtures.FixRequestBundleInstanceAuthCreationRequest(bndlAddOutput.ID, bndlInstanceAuthRequestInputStr)
	bndlInstanceAuth := graphql.BundleInstanceAuth{}

	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), certSecuredGraphQLClient, tenantId, runtime.ID)
	rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
	require.NotEmpty(t, rtmOauthCredentialData.ClientID)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, rtmOauthCredentialData, token.RuntimeScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	runtimeConsumer := testctx.Tc.NewOperation(ctx)

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenario)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenario)

	// set application scenarios label
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: testScenario}, application.ID, tenantId)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: testScenario}, application.ID, tenantId)

	// set runtime scenarios label
	defer fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: testScenario}, runtime.ID, tenantId)
	fixtures.AssignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: testScenario}, runtime.ID, tenantId)

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
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, delReq, nil)
	require.NoError(t, err)

	appExt = graphql.ApplicationExt{}
	t.Log("Fetch application with bundles after deleting runtime")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, bundlesForApplicationReq, &appExt)
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

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	authInput := fixtures.FixBasicAuth(t)

	bndlInput := fixtures.FixBundleCreateInputWithDefaultAuth(bundleName, authInput)
	bndl, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(bndlInput)
	require.NoError(t, err)

	addBndlRequest := fixtures.FixAddBundleRequest(application.ID, bndl)
	bndlAddOutput := graphql.Bundle{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addBndlRequest, &bndlAddOutput)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndlAddOutput.ID)
	require.NoError(t, err)

	bndlInstanceAuthRequestInput := fixtures.FixBundleInstanceAuthRequestInput(nil, nil)
	bndlInstanceAuthRequestInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixtures.FixRequestBundleInstanceAuthCreationRequest(bndlAddOutput.ID, bndlInstanceAuthRequestInputStr)
	authOutput := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth creation")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, bndlInstanceAuthCreationRequestReq, &authOutput)

	// THEN
	require.NoError(t, err)

	// Fetch Application with bundles
	bundlesForApplicationReq := fixtures.FixGetBundlesRequest(application.ID)
	bndlFromAPI := graphql.ApplicationExt{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, bundlesForApplicationReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data))
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data[0].InstanceAuths))

	// Fetch Application with bundle
	instanceAuthID := bndlFromAPI.Bundles.Data[0].InstanceAuths[0].ID
	bundlesForApplicationWithInstanceAuthReq := fixtures.FixGetBundleWithInstanceAuthRequest(application.ID, bndlAddOutput.ID, instanceAuthID)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, bundlesForApplicationWithInstanceAuthReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auth exist
	require.Equal(t, instanceAuthID, bndlFromAPI.Bundle.InstanceAuth.ID)

	require.Equal(t, graphql.BundleInstanceAuthStatusConditionSucceeded, bndlFromAPI.Bundle.InstanceAuth.Status.Condition)
	assertions.AssertAuth(t, authInput, bndlFromAPI.Bundle.InstanceAuth.Auth)
}

func TestRequestBundleInstanceAuthCreationWithDefaultAuthForCertificateOAuth(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "app-test-bundle-cert-oauth", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	authInput := fixtures.FixCertificateOauthAuth(t)

	bndlInput := fixtures.FixBundleCreateInputWithDefaultAuth("bndl-app-1-cert-oauth", authInput)
	bndl, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(bndlInput)
	require.NoError(t, err)

	addBndlRequest := fixtures.FixAddBundleRequest(application.ID, bndl)
	bndlAddOutput := graphql.Bundle{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addBndlRequest, &bndlAddOutput)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndlAddOutput.ID)
	require.NoError(t, err)

	bndlInstanceAuthRequestInput := fixtures.FixBundleInstanceAuthRequestInput(nil, nil)
	bndlInstanceAuthRequestInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixtures.FixRequestBundleInstanceAuthCreationRequest(bndlAddOutput.ID, bndlInstanceAuthRequestInputStr)
	authOutput := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth creation")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, bndlInstanceAuthCreationRequestReq, &authOutput)

	// THEN
	require.NoError(t, err)

	// Fetch Application with bundles
	bundlesForApplicationReq := fixtures.FixGetBundlesRequest(application.ID)
	bndlFromAPI := graphql.ApplicationExt{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, bundlesForApplicationReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data))
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data[0].InstanceAuths))

	// Fetch Application with bundle
	instanceAuthID := bndlFromAPI.Bundles.Data[0].InstanceAuths[0].ID
	bundlesForApplicationWithInstanceAuthReq := fixtures.FixGetBundleWithInstanceAuthRequest(application.ID, bndlAddOutput.ID, instanceAuthID)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, bundlesForApplicationWithInstanceAuthReq, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auth exist
	require.Equal(t, instanceAuthID, bndlFromAPI.Bundle.InstanceAuth.ID)

	require.Equal(t, graphql.BundleInstanceAuthStatusConditionSucceeded, bndlFromAPI.Bundle.InstanceAuth.Status.Condition)
	assertions.AssertAuth(t, authInput, bndlFromAPI.Bundle.InstanceAuth.Auth)
}

func TestRequestBundleInstanceAuthDeletion(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bundleName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	bndlInstanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, certSecuredGraphQLClient, bndl.ID)

	bndlInstanceAuthDeletionRequestReq := fixtures.FixRequestBundleInstanceAuthDeletionRequest(bndlInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Request bundle instance auth deletion")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, bndlInstanceAuthDeletionRequestReq, &output)

	// THEN
	require.NoError(t, err)

	SaveExample(t, bndlInstanceAuthDeletionRequestReq.Query(), "request bundle instance auth deletion")
}

func TestRequestBundleInstanceAuthDeletionAsRuntimeConsumer(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	input := fixRuntimeInput("runtime-test")

	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, input, conf.GatewayOauth)

	application, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, appName, conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bundleName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	bndlInstanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, certSecuredGraphQLClient, bndl.ID)

	bndlInstanceAuthDeletionRequestReq := fixtures.FixRequestBundleInstanceAuthDeletionRequest(bndlInstanceAuth.ID)

	testScenarioSecond := "test-scenario-2"

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenario)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenario)

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenarioSecond)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, testScenarioSecond)

	formation1 := graphql.FormationInput{Name: testScenario}
	formation2 := graphql.FormationInput{Name: testScenarioSecond}

	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), certSecuredGraphQLClient, tenantId, runtime.ID)
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
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formation1, application.ID, tenantId)
		fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formation1, application.ID, tenantId)

		// set runtime scenarios label
		defer fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, formation1, runtime.ID, tenantId)
		fixtures.AssignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, formation1, runtime.ID, tenantId)

		// WHEN
		t.Log("Request bundle instance auth deletion")
		output := graphql.BundleInstanceAuth{}
		err := runtimeConsumer.Run(bndlInstanceAuthDeletionRequestReq, oauthGraphQLClient, &output)

		// THEN
		require.NoError(t, err)
	})

	t.Run("When runtime is NOT in the same scenario as application", func(t *testing.T) {
		// set application scenarios label
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formation2, application.ID, tenantId)
		fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formation2, application.ID, tenantId)

		// set runtime scenarios label
		defer fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, formation1, runtime.ID, tenantId)
		fixtures.AssignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, formation1, runtime.ID, tenantId)

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

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bundleName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	bndlInstanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, certSecuredGraphQLClient, bndl.ID)

	authInput := fixtures.FixBasicAuth(t)
	bndlInstanceAuthSetInput := fixtures.FixBundleInstanceAuthSetInputSucceeded(authInput)
	bndlInstanceAuthSetInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthSetInputToGQL(bndlInstanceAuthSetInput)
	require.NoError(t, err)

	setBundleInstanceAuthReq := fixtures.FixSetBundleInstanceAuthRequest(bndlInstanceAuth.ID, bndlInstanceAuthSetInputStr)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Set bundle instance auth")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, setBundleInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)
	require.Equal(t, graphql.BundleInstanceAuthStatusConditionSucceeded, output.Status.Condition)
	assertions.AssertAuth(t, authInput, output.Auth)

	SaveExample(t, setBundleInstanceAuthReq.Query(), "set bundle instance auth")
}

func TestSetBundleInstanceAuthWithCertificateOAuthCredentials(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "app-test-bundle-cert-oauth", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, "bndl-app-1-cert-oauth")
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	bndlInstanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, certSecuredGraphQLClient, bndl.ID)

	authInput := fixtures.FixCertificateOauthAuth(t)
	bndlInstanceAuthSetInput := fixtures.FixBundleInstanceAuthSetInputSucceeded(authInput)
	bndlInstanceAuthSetInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthSetInputToGQL(bndlInstanceAuthSetInput)
	require.NoError(t, err)

	setBundleInstanceAuthReq := fixtures.FixSetBundleInstanceAuthRequest(bndlInstanceAuth.ID, bndlInstanceAuthSetInputStr)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Set bundle instance auth")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, setBundleInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)
	require.Equal(t, graphql.BundleInstanceAuthStatusConditionSucceeded, output.Status.Condition)
	assertions.AssertAuth(t, authInput, output.Auth)
}

func TestDeleteBundleInstanceAuth(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bundleName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	bndlInstanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, certSecuredGraphQLClient, bndl.ID)

	deleteBundleInstanceAuthReq := fixtures.FixDeleteBundleInstanceAuthRequest(bndlInstanceAuth.ID)
	output := graphql.BundleInstanceAuth{}

	// WHEN
	t.Log("Delete bundle instance auth")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, deleteBundleInstanceAuthReq, &output)

	// THEN
	require.NoError(t, err)

	SaveExample(t, deleteBundleInstanceAuthReq.Query(), "delete bundle instance auth")
}

func TestCreateUpdateBundleInstanceAuth(t *testing.T) {
	ctx = context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Logf("Registering application with name %q in tenant %q...", appName, tenantId)
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	t.Logf("Creating bundle with name %q for application with ID %q in tenant %q...", bundleName, application.ID, tenantId)
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bundleName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	t.Logf("Registering runtime with name %q in tenant %q...", rtmName, tenantId)
	rtmIn := fixRuntimeInput(rtmName)
	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, &rtmIn)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	// Create the bundle instance auth
	authInput := fixtures.FixBasicAuth(t)
	authCtx, inputParams := fixtures.FixBundleInstanceAuthContextAndInputParams(t)
	in, err := testctx.Tc.Graphqlizer.BundleInstanceAuthCreateInputToGQL(fixtures.FixBundleInstanceAuthCreateInput(authCtx, inputParams, authInput, &runtime.ID, nil))
	require.NoError(t, err)

	fieldProvider := graphqlizer.GqlFieldsProvider{}

	createBundleInstanceAuthReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {result: createBundleInstanceAuth(bundleID: "%s", in: %s) {%s}}`, bndl.ID, in, fieldProvider.ForBundleInstanceAuth()))

	var instanceAuth graphql.BundleInstanceAuth

	t.Logf("Creating bundle instance auth for bundle with id %q...", bndl.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createBundleInstanceAuthReq, &instanceAuth)
	defer fixtures.DeleteBundleInstanceAuth(t, ctx, certSecuredGraphQLClient, instanceAuth.ID)
	require.NoError(t, err)
	require.EqualValues(t, authInput.Credential.Basic, instanceAuth.Auth.Credential.(*graphql.BasicCredentialData))

	SaveExample(t, createBundleInstanceAuthReq.Query(), "create bundle instance auth")

	// Update the bundle instance auth
	updatedAuthInput := fixtures.FixOauthAuth(t)

	var authCtxData interface{} = map[string]interface{}{"ContextData": "ContextValue"}
	var inputParamsData interface{} = map[string]interface{}{"InKey": "InValue"}
	updatedAuthCtx := json2.MarshalJSON(t, authCtxData)
	updatedInputParams := json2.MarshalJSON(t, inputParamsData)

	in, err = testctx.Tc.Graphqlizer.BundleInstanceAuthUpdateInputToGQL(fixtures.FixBundleInstanceAuthUpdateInput(updatedAuthCtx, updatedInputParams, updatedAuthInput))
	require.NoError(t, err)

	updateBundleInstanceAuthReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {result: updateBundleInstanceAuth(id: "%s", bundleID: "%s", in: %s) {%s}}`, instanceAuth.ID, bndl.ID, in, fieldProvider.ForBundleInstanceAuth()))

	var updatedInstanceAuth graphql.BundleInstanceAuth

	t.Logf("Updating bundle instance auth for bundle with id %q...", bndl.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateBundleInstanceAuthReq, &updatedInstanceAuth)
	require.NoError(t, err)
	require.EqualValues(t, updatedAuthInput.Credential.Oauth, updatedInstanceAuth.Auth.Credential.(*graphql.OAuthCredentialData))

	SaveExample(t, updateBundleInstanceAuthReq.Query(), "update bundle instance auth")
}
