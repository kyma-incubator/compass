package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/require"
)

const runtimeEventURLFormat = "https://%s"
const appEventURLFormat = "https://%s/%s/v1/events"

func TestGetDefaultRuntimeForEventingForApplication_DefaultBehaviourWhenNoEventingAssigned(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	runtimeEventingURLLabelKey := "runtime_eventServiceUrl"
	runtime1Eventing := "eventing.runtime1.local"
	runtime1EventingURL := fmt.Sprintf(runtimeEventURLFormat, runtime1Eventing)
	runtime2EventingURL := "https://eventing.runtime2.local"
	defaultScenarios := []string{"DEFAULT"}

	appName := "app-test-eventing"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	fixtures.SetApplicationLabel(t, ctx, certSecuredGraphQLClient, application.ID, ScenariosLabel, defaultScenarios)

	input1 := fixRuntimeInput("runtime-1-eventing")
	runtime1 := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, input1, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime1)

	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime1.ID, ScenariosLabel, defaultScenarios)
	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime1.ID, runtimeEventingURLLabelKey, runtime1EventingURL)
	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime1.ID, IsNormalizedLabel, "false")

	input2 := fixRuntimeInput("runtime-2-eventing")
	runtime2 := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, input2, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime2)

	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime2.ID, ScenariosLabel, defaultScenarios)
	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime2.ID, runtimeEventingURLLabelKey, runtime2EventingURL)

	// WHEN
	testApp := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, application.ID)

	// THEN
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime1Eventing, appName), testApp.EventingConfiguration.DefaultURL)
}

func TestGetEventingConfigurationForRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	runtimeEventingURLLabelKey := "runtime_eventServiceUrl"
	runtimeEventingURL := "http://eventing.runtime.local"

	input := fixRuntimeInput("runtime-eventing")

	runtime := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, input, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)

	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID, runtimeEventingURLLabelKey, runtimeEventingURL)

	// WHEN
	testRuntime := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID)

	// THEN
	require.Equal(t, runtimeEventingURL, testRuntime.EventingConfiguration.DefaultURL)
}

func TestSetDefaultEventingForApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	runtimeEventingURLLabelKey := "runtime_eventServiceUrl"
	runtime1Eventing := "eventing.runtime1.local"
	runtime1EventingURL := fmt.Sprintf(runtimeEventURLFormat, runtime1Eventing)
	runtime2Eventing := "eventing.runtime2.local"
	runtime2EventingURL := fmt.Sprintf(runtimeEventURLFormat, runtime2Eventing)
	defaultScenarios := []string{"DEFAULT"}

	appName := "app-test-eventing"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	fixtures.SetApplicationLabel(t, ctx, certSecuredGraphQLClient, application.ID, ScenariosLabel, defaultScenarios)

	input1 := fixRuntimeInput("runtime-1-eventing")
	runtime1 := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, input1, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime1)

	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime1.ID, ScenariosLabel, defaultScenarios)
	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime1.ID, runtimeEventingURLLabelKey, runtime1EventingURL)
	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime1.ID, IsNormalizedLabel, "false")

	input2 := fixRuntimeInput("runtime-2-eventing")

	runtime2 := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, input2, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime2)
	require.NoError(t, err)
	require.NotEmpty(t, runtime2.ID)

	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime2.ID, ScenariosLabel, defaultScenarios)
	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime2.ID, runtimeEventingURLLabelKey, runtime2EventingURL)
	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime2.ID, IsNormalizedLabel, "true")

	// WHEN
	testApp := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime1Eventing, appName), testApp.EventingConfiguration.DefaultURL)

	actualEventingCfg := graphql.ApplicationEventingConfiguration{}
	request := fixtures.FixSetDefaultEventingForApplication(application.ID, runtime2.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualEventingCfg)

	// THEN
	defaultAppNameNormalizer := &normalizer.DefaultNormalizator{}
	normalizedAppName := defaultAppNameNormalizer.Normalize(appName)
	saveExampleInCustomDir(t, request.Query(), eventingCategory, "set default eventing for application")
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime2Eventing, normalizedAppName), actualEventingCfg.DefaultURL)

	testApp = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime2Eventing, normalizedAppName), testApp.EventingConfiguration.DefaultURL)
}

func TestEmptyEventConfigurationForApp(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "app-test-eventing", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	input1 := fixRuntimeInput("runtime-1-eventing")

	runtime1 := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, input1, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime1)

	//WHEN
	app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, application.ID)

	//THEN
	assert.Equal(t, "", app.EventingConfiguration.DefaultURL)
}

func TestDeleteDefaultEventingForApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	runtimeEventingURLLabelKey := "runtime_eventServiceUrl"
	runtime1Eventing := "eventing.runtime1.local"
	runtime1EventingURL := fmt.Sprintf(runtimeEventURLFormat, runtime1Eventing)
	runtime2Eventing := "eventing.runtime2.local"
	runtime2EventingURL := fmt.Sprintf(runtimeEventURLFormat, runtime2Eventing)
	defaultScenarios := []string{"DEFAULT"}

	appName := "app-test-eventing"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	fixtures.SetApplicationLabel(t, ctx, certSecuredGraphQLClient, application.ID, ScenariosLabel, defaultScenarios)

	input1 := fixRuntimeInput("runtime-1-eventing")
	runtime1 := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, input1, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime1)

	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime1.ID, ScenariosLabel, defaultScenarios)
	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime1.ID, runtimeEventingURLLabelKey, runtime1EventingURL)
	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime1.ID, IsNormalizedLabel, "false")

	input2 := fixRuntimeInput("runtime-2-eventing")
	runtime2 := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, input2, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime2)

	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime2.ID, ScenariosLabel, defaultScenarios)
	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime2.ID, runtimeEventingURLLabelKey, runtime2EventingURL)
	fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime2.ID, IsNormalizedLabel, "false")

	testApp := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime1Eventing, appName), testApp.EventingConfiguration.DefaultURL)

	fixtures.SetDefaultEventingForApplication(t, ctx, certSecuredGraphQLClient, application.ID, runtime2.ID)

	testApp = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime2Eventing, appName), testApp.EventingConfiguration.DefaultURL)

	// WHEN
	actualEventingCfg := graphql.ApplicationEventingConfiguration{}
	request := fixtures.FixDeleteDefaultEventingForApplication(application.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualEventingCfg)

	// THEN
	saveExampleInCustomDir(t, request.Query(), eventingCategory, "delete default eventing for application")
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime2Eventing, appName), actualEventingCfg.DefaultURL)

	testApp = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime1Eventing, appName), testApp.EventingConfiguration.DefaultURL)
}
