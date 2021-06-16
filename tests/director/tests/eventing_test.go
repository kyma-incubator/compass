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
	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, defaultScenarios)

	input1 := fixtures.FixRuntimeInput("runtime-1-eventing")

	runtime1 := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input1)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime1.ID)

	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime1.ID, ScenariosLabel, defaultScenarios)
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime1.ID, runtimeEventingURLLabelKey, runtime1EventingURL)
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime1.ID, IsNormalizedLabel, "false")

	input2 := fixtures.FixRuntimeInput("runtime-2-eventing")

	runtime2 := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input2)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime2.ID)

	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime2.ID, ScenariosLabel, defaultScenarios)
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime2.ID, runtimeEventingURLLabelKey, runtime2EventingURL)

	// WHEN
	testApp := fixtures.GetApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	// THEN
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime1Eventing, appName), testApp.EventingConfiguration.DefaultURL)
}

func TestGetEventingConfigurationForRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	runtimeEventingURLLabelKey := "runtime_eventServiceUrl"
	runtimeEventingURL := "http://eventing.runtime.local"

	input := fixtures.FixRuntimeInput("runtime-eventing")

	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, runtimeEventingURLLabelKey, runtimeEventingURL)

	// WHEN
	testRuntime := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

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
	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, defaultScenarios)

	input1 := fixtures.FixRuntimeInput("runtime-1-eventing")

	runtime1 := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input1)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime1.ID)

	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime1.ID, ScenariosLabel, defaultScenarios)
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime1.ID, runtimeEventingURLLabelKey, runtime1EventingURL)
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime1.ID, IsNormalizedLabel, "false")

	input2 := fixtures.FixRuntimeInput("runtime-2-eventing")

	runtime2 := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input2)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime2.ID)

	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime2.ID, ScenariosLabel, defaultScenarios)
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime2.ID, runtimeEventingURLLabelKey, runtime2EventingURL)
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime2.ID, IsNormalizedLabel, "true")

	// WHEN
	testApp := fixtures.GetApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime1Eventing, appName), testApp.EventingConfiguration.DefaultURL)

	actualEventingCfg := graphql.ApplicationEventingConfiguration{}
	request := fixtures.FixSetDefaultEventingForApplication(application.ID, runtime2.ID)
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualEventingCfg)

	// THEN
	defaultAppNameNormalizer := &normalizer.DefaultNormalizator{}
	normalizedAppName := defaultAppNameNormalizer.Normalize(appName)
	saveExampleInCustomDir(t, request.Query(), eventingCategory, "set default eventing for application")
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime2Eventing, normalizedAppName), actualEventingCfg.DefaultURL)

	testApp = fixtures.GetApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime2Eventing, normalizedAppName), testApp.EventingConfiguration.DefaultURL)
}

func TestEmptyEventConfigurationForApp(t *testing.T) {
	//GIVEN
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-eventing", tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	input1 := fixtures.FixRuntimeInput("runtime-1-eventing")

	runtime1 := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input1)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime1.ID)

	//WHEN
	app := fixtures.GetApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

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
	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, application.ID, ScenariosLabel, defaultScenarios)

	input1 := fixtures.FixRuntimeInput("runtime-1-eventing")

	runtime1 := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input1)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime1.ID)

	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime1.ID, ScenariosLabel, defaultScenarios)
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime1.ID, runtimeEventingURLLabelKey, runtime1EventingURL)
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime1.ID, IsNormalizedLabel, "false")

	input2 := fixtures.FixRuntimeInput("runtime-2-eventing")

	runtime2 := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input2)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime2.ID)

	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime2.ID, ScenariosLabel, defaultScenarios)
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime2.ID, runtimeEventingURLLabelKey, runtime2EventingURL)
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime2.ID, IsNormalizedLabel, "false")

	testApp := fixtures.GetApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime1Eventing, appName), testApp.EventingConfiguration.DefaultURL)

	fixtures.SetDefaultEventingForApplication(t, ctx, dexGraphQLClient, application.ID, runtime2.ID)

	testApp = fixtures.GetApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime2Eventing, appName), testApp.EventingConfiguration.DefaultURL)

	// WHEN
	actualEventingCfg := graphql.ApplicationEventingConfiguration{}
	request := fixtures.FixDeleteDefaultEventingForApplication(application.ID)
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualEventingCfg)

	// THEN
	saveExampleInCustomDir(t, request.Query(), eventingCategory, "delete default eventing for application")
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime2Eventing, appName), actualEventingCfg.DefaultURL)

	testApp = fixtures.GetApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime1Eventing, appName), testApp.EventingConfiguration.DefaultURL)
}
