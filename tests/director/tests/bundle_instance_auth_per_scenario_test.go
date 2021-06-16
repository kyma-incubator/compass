package tests

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnassignResourceFromScenarioWhenThereIsExistingBundleInstanceAuth(t *testing.T) {
	ctx := context.Background()
	createScenarioLabelDefIfNotExist(t, ctx)

	commonScenario := "common-scenario"
	otherAppScenario := "other-app-scenario"
	otherRuntimeScenario := "other-runtime-scenario"

	scenarios := []string{conf.DefaultScenario, commonScenario, otherAppScenario, otherRuntimeScenario}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), scenarios[:1])

	// GIVEN
	data, cleanup := setupAppAndRuntimeWithCommonScenario(t, ctx, commonScenario)
	defer cleanup()

	t.Run("Remove Application from scenario by setApplicationLabel mutation", func(t *testing.T) {
		request := fixtures.FixSetApplicationLabelRequest(data.appId, ScenariosLabel, []interface{}{otherAppScenario})
		requireExistingBundleInstanceAuthsError(t, data.appName, resource.Application, testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &graphql.Label{}))
	})

	t.Run("Remove Runtime from scenario by setRuntimeLabel mutation", func(t *testing.T) {
		request := fixtures.FixSetRuntimeLabelRequest(data.runtimeId, ScenariosLabel, []interface{}{otherRuntimeScenario})
		requireExistingBundleInstanceAuthsError(t, data.runtimeName, resource.Runtime, testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &graphql.Label{}))
	})

	t.Run("Remove Application from scenario by deleteApplicationLabel mutation", func(t *testing.T) {
		request := fixtures.FixDeleteApplicationLabelRequest(data.appId, ScenariosLabel)
		requireExistingBundleInstanceAuthsError(t, data.appName, resource.Application, testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &graphql.Label{}))
	})

	t.Run("Remove Runtime from scenario by deleteRuntimeLabel mutation", func(t *testing.T) {
		request := fixtures.FixDeleteRuntimeLabelRequest(data.runtimeId, ScenariosLabel)
		requireExistingBundleInstanceAuthsError(t, data.runtimeName, resource.Runtime, testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &graphql.Label{}))
	})

	t.Run("Remove Runtime from scenario by updateRuntime mutation", func(t *testing.T) {
		updateInput := graphql.RuntimeInput{
			Name:   data.runtimeName,
			Labels: graphql.Labels{ScenariosLabel: []interface{}{otherRuntimeScenario}},
		}
		updateInputGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(updateInput)
		require.NoError(t, err)

		request := fixtures.FixUpdateRuntimeRequest(data.runtimeId, updateInputGQL)
		requireExistingBundleInstanceAuthsError(t, data.runtimeName, resource.Runtime, testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &graphql.RuntimeExt{}))
	})
}

func TestUnassignResourceFromScenarioAfterBundleInstanceAuthWasRemoved(t *testing.T) {
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetDefaultTenantID()
	createScenarioLabelDefIfNotExist(t, ctx)

	scenario := "scenario-1"
	scenarios := []string{conf.DefaultScenario, scenario}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), scenarios[:1])

	// GIVEN
	data, cleanup := setupAppAndRuntimeWithCommonScenario(t, ctx, scenario)
	defer cleanup()

	fixtures.DeleteBundleInstanceAuth(t, ctx, dexGraphQLClient, tenantID, data.bundleInstanceAuthID)

	setAppLabelReq := fixtures.FixSetApplicationLabelRequest(data.appId, ScenariosLabel, []interface{}{conf.DefaultScenario})
	require.NoError(t, testctx.Tc.RunOperation(ctx, dexGraphQLClient, setAppLabelReq, &graphql.Label{}))
}

func TestUnassignResourceFromScenarioUsingAutomaticScenarioAssignment(t *testing.T) {
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetDefaultTenantID()
	createScenarioLabelDefIfNotExist(t, ctx)

	commonScenario := "common-scenario"
	otherScenario := "other-common-scenario"

	scenarios := []string{conf.DefaultScenario, commonScenario, otherScenario}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), scenarios[:1])

	// GIVEN
	data, cleanup := setupAppAndRuntimeWithCommonScenario(t, ctx, commonScenario)
	defer cleanup()

	// Label selector for automatic scenario assignment
	labelSelector := graphql.LabelSelectorInput{
		Key:   "keyA",
		Value: "valueA",
	}

	// Update application scenarios
	setAppLabelRequest := fixtures.FixSetApplicationLabelRequest(data.appId, ScenariosLabel, []interface{}{commonScenario, otherScenario})
	appLabel := graphql.Label{}
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, setAppLabelRequest, &appLabel)
	require.NoError(t, err)

	// Set runtime label
	setRuntimeLabelRequest := fixtures.FixSetRuntimeLabelRequest(data.runtimeId, labelSelector.Key, labelSelector.Value)
	runtimeLabel := graphql.Label{}
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, setRuntimeLabelRequest, &runtimeLabel)
	require.NoError(t, err)

	// Create automatic scenario assignment
	asaInput := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: otherScenario,
		Selector:     &labelSelector,
	}
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, tenantID)

	// Cannot delete automatic scenario assignment when there are existing bundle instance auths
	assignment := graphql.AutomaticScenarioAssignment{}
	req := fixtures.FixDeleteAutomaticScenarioAssignmentForScenarioRequest(otherScenario)
	requireExistingBundleInstanceAuthsError(t, data.runtimeName, resource.Runtime, testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, req, &assignment))

	// Cannot delete label which corresponds for the automatic scenario assignment
	request := fixtures.FixDeleteRuntimeLabelRequest(data.runtimeId, labelSelector.Key)
	requireExistingBundleInstanceAuthsError(t, data.runtimeName, resource.Runtime, testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &graphql.Label{}))

	// Successfully dismiss scenario for which there is automatic scenario assignment
	updateInput := graphql.RuntimeInput{
		Name: data.runtimeName,
		Labels: graphql.Labels{
			ScenariosLabel:    []string{commonScenario},
			labelSelector.Key: labelSelector.Value,
		},
	}
	updateInputGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(updateInput)
	require.NoError(t, err)
	request = fixtures.FixUpdateRuntimeRequest(data.runtimeId, updateInputGQL)
	require.NoError(t, testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &graphql.RuntimeExt{}))

	// Successfully remove ASA Label from runtime while explicitly assigning the scenario
	updateInput = graphql.RuntimeInput{
		Name: data.runtimeName,
		Labels: graphql.Labels{
			ScenariosLabel: []string{commonScenario, otherScenario},
		},
	}
	updateInputGQL, err = testctx.Tc.Graphqlizer.RuntimeInputToGQL(updateInput)
	require.NoError(t, err)
	request = fixtures.FixUpdateRuntimeRequest(data.runtimeId, updateInputGQL)
	require.NoError(t, testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &graphql.RuntimeExt{}))

	// Deleting bundle instance auth
	fixtures.DeleteBundleInstanceAuth(t, ctx, dexGraphQLClient, tenantID, data.bundleInstanceAuthID)

	// Now automatic scenario assignment deletion should be successful
	req = fixtures.FixDeleteAutomaticScenarioAssignmentForScenarioRequest(otherScenario)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, req, &assignment)
	assert.NoError(t, err)
}

func createScenarioLabelDefIfNotExist(t *testing.T, ctx context.Context) {
	defs, err := fixtures.ListLabelDefinitionsWithinTenant(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID())
	require.NoError(t, err)

	for _, def := range defs {
		if def.Key == ScenariosLabel {
			return
		}
	}

	fixtures.CreateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), []string{conf.DefaultScenario})
}

func setupAppAndRuntimeWithCommonScenario(t *testing.T, ctx context.Context, scenario string) (result *data, cleanup func()) {
	uniqueStr := strconv.FormatInt(time.Now().UnixNano(), 10)
	defaultTenant := tenant.TestTenants.GetDefaultTenantID()

	var cleanups []func()
	cleanup = func() {
		for _, cleanup := range cleanups {
			cleanup()
		}
	}

	setupOk := false
	defer func() {
		if setupOk {
			return
		}

		cleanupRef := cleanup
		cleanup = func() {
			//No cleanup remains for the caller
		}
		cleanupRef()

	}()

	// Register Runtime
	runtimeInput := graphql.RuntimeInput{
		Name:        "runtime-" + uniqueStr,
		Description: ptr.String("runtime-1-description"),
		Labels:      graphql.Labels{ScenariosLabel: []interface{}{scenario}},
	}

	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(runtimeInput)
	require.NoError(t, err)
	actualRuntime := graphql.RuntimeExt{}

	registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &actualRuntime)
	cleanups = append([]func(){func() { fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, defaultTenant, actualRuntime.ID) }}, cleanups...)
	require.NoError(t, err)

	// Register Application
	appInput := graphql.ApplicationRegisterInput{
		Name:        "app" + uniqueStr,
		Description: ptr.String("my first wordpress application"),
		Labels:      graphql.Labels{ScenariosLabel: []interface{}{scenario}},
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(appInput)
	require.NoError(t, err)

	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	actualApp := graphql.ApplicationExt{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualApp)
	cleanups = append([]func(){func() {
		fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, actualApp.ID, ScenariosLabel, []interface{}{conf.DefaultScenario})
		fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, defaultTenant, actualApp.ID)
	}}, cleanups...)
	require.NoError(t, err)

	//Register Bundle
	bndlName := "test-bundle" + uniqueStr
	authInput := fixtures.FixBasicAuth(t)

	bndlInput := fixtures.FixBundleCreateInputWithDefaultAuth(bndlName, authInput)
	bndl, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(bndlInput)
	require.NoError(t, err)

	addBndlRequest := fixtures.FixAddBundleRequest(actualApp.ID, bndl)
	bndlAddOutput := graphql.Bundle{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, addBndlRequest, &bndlAddOutput)
	require.NoError(t, err)

	// Request bundle instance auth
	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, defaultTenant, actualRuntime.ID)
	rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
	require.NotEmpty(t, rtmOauthCredentialData.ClientID)

	accessToken := token.GetAccessToken(t, rtmOauthCredentialData, token.RuntimeScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	authCtx, inputParams := fixtures.FixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := fixtures.FixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixtures.FixRequestBundleInstanceAuthCreationRequest(bndlAddOutput.ID, bndlInstanceAuthRequestInputStr)
	actualBundleInstanceAuth := graphql.BundleInstanceAuth{}

	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, bndlInstanceAuthCreationRequestReq, &actualBundleInstanceAuth)
	cleanups = append([]func(){func() {
		fixtures.DeleteBundleInstanceAuth(t, ctx, dexGraphQLClient, defaultTenant, actualBundleInstanceAuth.ID)
	}}, cleanups...)
	require.NoError(t, err)

	result = &data{
		appId:                actualApp.ID,
		runtimeId:            actualRuntime.ID,
		scenarios:            []string{scenario},
		appName:              actualApp.Name,
		runtimeName:          actualRuntime.Name,
		bundleInstanceAuthID: actualBundleInstanceAuth.ID,
	}
	setupOk = true
	return
}

type data struct {
	appId                string
	runtimeId            string
	appName              string
	runtimeName          string
	bundleInstanceAuthID string
	scenarios            []string
}

func requireExistingBundleInstanceAuthsError(t *testing.T, rName string, rType resource.Type, err error) {
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("%s %s is still used and cannot be removed from formation. Contact you system administrator.", strings.Title(string(rType)), rName))
}
