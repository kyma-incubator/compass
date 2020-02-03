package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

const runtimeEventURLFormat = "https://%s"
const appEventURLFormat = "https://%s/%s/v1/events"

func TestGetDefaultRuntimeForEventingForApplication_DefaultBehaviourWhenNoEventingAssigned(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	runtimeEventingURLLabelKey := "runtime/event_service_url"
	runtime1Eventing := "eventing.runtime1.local"
	runtime1EventingURL := fmt.Sprintf(runtimeEventURLFormat, runtime1Eventing)
	runtime2EventingURL := "https://eventing.runtime2.local"
	defaultScenarios := []string{"DEFAULT"}

	appName := "app-test-eventing"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	runtime1 := registerRuntime(t, ctx, "runtime-1-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime1.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime1.ID, scenariosLabel, defaultScenarios)
	setRuntimeLabel(t, ctx, runtime1.ID, runtimeEventingURLLabelKey, runtime1EventingURL)

	runtime2 := registerRuntime(t, ctx, "runtime-2-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime2.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime2.ID, scenariosLabel, defaultScenarios)
	setRuntimeLabel(t, ctx, runtime2.ID, runtimeEventingURLLabelKey, runtime2EventingURL)

	// WHEN
	testApp := getApplication(t, ctx, application.ID)

	// THEN
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime1Eventing, appName), testApp.EventingConfiguration.DefaultURL)
}

func TestGetEventingConfigurationForRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	runtimeEventingURLLabelKey := "runtime/event_service_url"
	runtimeEventingURL := "http://eventing.runtime.local"

	runtime := registerRuntime(t, ctx, "runtime-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime.ID, runtimeEventingURLLabelKey, runtimeEventingURL)

	// WHEN
	testRuntime := getRuntime(t, ctx, runtime.ID)

	// THEN
	require.Equal(t, runtimeEventingURL, testRuntime.EventingConfiguration.DefaultURL)
}

func TestSetDefaultEventingForApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	runtimeEventingURLLabelKey := "runtime/event_service_url"
	runtime1Eventing := "eventing.runtime1.local"
	runtime1EventingURL := fmt.Sprintf(runtimeEventURLFormat, runtime1Eventing)
	runtime2Eventing := "eventing.runtime2.local"
	runtime2EventingURL := fmt.Sprintf(runtimeEventURLFormat, runtime2Eventing)
	defaultScenarios := []string{"DEFAULT"}

	appName := "app-test-eventing"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	runtime1 := registerRuntime(t, ctx, "runtime-1-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime1.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime1.ID, scenariosLabel, defaultScenarios)
	setRuntimeLabel(t, ctx, runtime1.ID, runtimeEventingURLLabelKey, runtime1EventingURL)

	runtime2 := registerRuntime(t, ctx, "runtime-2-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime2.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime2.ID, scenariosLabel, defaultScenarios)
	setRuntimeLabel(t, ctx, runtime2.ID, runtimeEventingURLLabelKey, runtime2EventingURL)

	// WHEN
	testApp := getApplication(t, ctx, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime1Eventing, appName), testApp.EventingConfiguration.DefaultURL)

	actualEventingCfg := graphql.ApplicationEventingConfiguration{}
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: setDefaultEventingForApplication(appID: "%s", runtimeID: "%s") {
						%s
					}
				}`,
			application.ID, runtime2.ID, tc.gqlFieldsProvider.ForEventingConfiguration()))
	err := tc.RunOperation(ctx, request, &actualEventingCfg)

	// THEN
	saveExampleInCustomDir(t, request.Query(), eventingCategory, "set default eventing for application")
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime2Eventing, appName), actualEventingCfg.DefaultURL)

	testApp = getApplication(t, ctx, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime2Eventing, appName), testApp.EventingConfiguration.DefaultURL)
}

func TestEmptyEventConfigurationForApp(t *testing.T) {
	//GIVEN
	ctx := context.Background()
	application := registerApplication(t, ctx, "app-test-eventing")
	defer unregisterApplication(t, application.ID)

	runtime1 := registerRuntime(t, ctx, "runtime-1-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime1.ID, defaultTenant)

	//WHEN
	app := getApplication(t, ctx, application.ID)

	//THEN
	assert.Equal(t, "", app.EventingConfiguration.DefaultURL)
}

func TestDeleteDefaultEventingForApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	runtimeEventingURLLabelKey := "runtime/event_service_url"
	runtime1Eventing := "eventing.runtime1.local"
	runtime1EventingURL := fmt.Sprintf(runtimeEventURLFormat, runtime1Eventing)
	runtime2Eventing := "eventing.runtime2.local"
	runtime2EventingURL := fmt.Sprintf(runtimeEventURLFormat, runtime2Eventing)
	defaultScenarios := []string{"DEFAULT"}

	appName := "app-test-eventing"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	runtime1 := registerRuntime(t, ctx, "runtime-1-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime1.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime1.ID, scenariosLabel, defaultScenarios)
	setRuntimeLabel(t, ctx, runtime1.ID, runtimeEventingURLLabelKey, runtime1EventingURL)

	runtime2 := registerRuntime(t, ctx, "runtime-2-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime2.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime2.ID, scenariosLabel, defaultScenarios)
	setRuntimeLabel(t, ctx, runtime2.ID, runtimeEventingURLLabelKey, runtime2EventingURL)

	testApp := getApplication(t, ctx, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime1Eventing, appName), testApp.EventingConfiguration.DefaultURL)

	setDefaultEventingForApplication(t, ctx, application.ID, runtime2.ID)

	testApp = getApplication(t, ctx, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime2Eventing, appName), testApp.EventingConfiguration.DefaultURL)

	// WHEN
	actualEventingCfg := graphql.ApplicationEventingConfiguration{}
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: deleteDefaultEventingForApplication(appID: "%s") {
						%s
					}
				}`,
			application.ID, tc.gqlFieldsProvider.ForEventingConfiguration()))
	err := tc.RunOperation(ctx, request, &actualEventingCfg)

	// THEN
	saveExampleInCustomDir(t, request.Query(), eventingCategory, "delete default eventing for application")
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime2Eventing, appName), actualEventingCfg.DefaultURL)

	testApp = getApplication(t, ctx, application.ID)
	require.Equal(t, fmt.Sprintf(appEventURLFormat, runtime1Eventing, appName), testApp.EventingConfiguration.DefaultURL)
}
