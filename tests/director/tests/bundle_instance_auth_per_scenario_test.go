package tests

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
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
	data, cleanup, ok := setupAppAndRuntimeWithCommonScenario(t, ctx, commonScenario)
	defer cleanup()
	require.True(t, ok)

	t.Run("Remove Application from scenario by setApplicationLabel mutation", func(t *testing.T) {
		request := fixtures.FixSetApplicationLabelRequest(data.appId, ScenariosLabel, []interface{}{otherAppScenario})
		label := graphql.Label{}
		err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &label)
		require.Error(t, err)
	})

	t.Run("Remove Runtime from scenario by setRuntimeLabel mutation", func(t *testing.T) {
		request := fixtures.FixSetRuntimeLabelRequest(data.runtimeId, ScenariosLabel, []interface{}{otherRuntimeScenario})
		label := graphql.Label{}
		err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &label)
		require.Error(t, err)
	})

	t.Run("Remove Application from scenario by deleteApplicationLabel mutation", func(t *testing.T) {
		request := fixtures.FixDeleteApplicationLabelRequest(data.appId, ScenariosLabel)
		label := graphql.Label{}
		err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &label)
		require.Error(t, err)
	})

	t.Run("Remove Runtime from scenario by deleteRuntimeLabel mutation", func(t *testing.T) {
		request := fixtures.FixDeleteRuntimeLabelRequest(data.runtimeId, ScenariosLabel)
		label := graphql.Label{}
		err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &label)
		require.Error(t, err)
	})

	t.Run("Remove Runtime from scenario by updateRuntime mutation", func(t *testing.T) {
		updateInput := graphql.RuntimeInput{
			Name:   data.runtimeName,
			Labels: graphql.Labels{ScenariosLabel: "other-runtime-scenario"},
		}
		updateInputGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(updateInput)
		require.NoError(t, err)

		request := fixtures.FixUpdateRuntimeRequest(data.runtimeId, updateInputGQL)
		input := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &input)
		require.Error(t, err)
	})
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

func setupAppAndRuntimeWithCommonScenario(t *testing.T, ctx context.Context, scenario string) (*data, func(), bool) {
	uniqueStr := strconv.FormatInt(time.Now().UnixNano(), 10)
	log.Infof("Unique string %s", uniqueStr)

	defaultTenant := tenant.TestTenants.GetDefaultTenantID()

	var cleanups []func()
	executeCleanUp := func() {
		for _, cleanup := range cleanups {
			cleanup()
		}
	}
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
	if !assert.NoError(t, err) {
		return nil, executeCleanUp, false
	}

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
		//fixtures.DeleteApplicationLabel(t, ctx, dexGraphQLClient, actualApp.ID, ScenariosLabel)
		fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, actualApp.ID, ScenariosLabel, []interface{}{conf.DefaultScenario})
		fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, defaultTenant, actualApp.ID)
	}}, cleanups...)
	if !assert.NoError(t, err) {
		return nil, executeCleanUp, false
	}

	//Register Bundle
	bndlName := "test-bundle" + uniqueStr
	authInput := fixtures.FixBasicAuth(t)

	bndlInput := fixtures.FixBundleCreateInputWithDefaultAuth(bndlName, authInput)
	bndl, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(bndlInput)
	require.NoError(t, err)

	addBndlRequest := fixtures.FixAddBundleRequest(actualApp.ID, bndl)
	bndlAddOutput := graphql.Bundle{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, addBndlRequest, &bndlAddOutput)
	cleanups = append([]func(){func() { fixtures.DeleteBundle(t, ctx, dexGraphQLClient, defaultTenant, bndlAddOutput.ID) }}, cleanups...)
	if !assert.NoError(t, err) {
		return nil, executeCleanUp, false
	}

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
	if !assert.NoError(t, err) {
		return nil, executeCleanUp, false
	}

	return &data{
		appId:       actualApp.ID,
		runtimeId:   actualRuntime.ID,
		scenarios:   []string{scenario},
		runtimeName: actualRuntime.Name,
	}, executeCleanUp, true
}

type data struct {
	appId       string
	runtimeId   string
	runtimeName string
	scenarios   []string
}
