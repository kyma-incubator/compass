package director

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultRuntimeForEventingForApplication(t *testing.T) {
	ctx := context.Background()
	runtimeEventingURLLabelKey := "runtime/event_service_url"
	runtime1EventingURL := "http://eventing.runtime1.local"
	runtime2EventingURL := "http://eventing.runtime2.local"
	defaultScenarios := []string{"DEFAULT"}

	application := registerApplication(t, ctx, "app-test-eventing")
	defer unregisterApplication(t, application.ID)

	runtime1 := registerRuntime(t, ctx, "runtime-1-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime1.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime1.ID, scenariosLabel, defaultScenarios)
	setRuntimeLabel(t, ctx, runtime1.ID, runtimeEventingURLLabelKey, runtime1EventingURL)

	runtime2 := registerRuntime(t, ctx, "runtime-2-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime2.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime2.ID, scenariosLabel, defaultScenarios)
	setRuntimeLabel(t, ctx, runtime2.ID, runtimeEventingURLLabelKey, runtime2EventingURL)

	testApp := getApp(ctx, t, application.ID)

	require.Equal(t, runtime1EventingURL, testApp.EventingConfiguration.DefaultURL)
}

func TestGetEventingConfigurationForRuntime(t *testing.T) {
	ctx := context.Background()
	runtimeEventingURLLabelKey := "runtime/event_service_url"
	runtimeEventingURL := "http://eventing.runtime.local"

	runtime := registerRuntime(t, ctx, "runtime-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime.ID, runtimeEventingURLLabelKey, runtimeEventingURL)

	testRuntime := getRuntime(t, ctx, runtime.ID)

	require.Equal(t, runtimeEventingURL, testRuntime.EventingConfiguration.DefaultURL)
}

func TestSetDefaultEventingForApplication(t *testing.T) {
	ctx := context.Background()
	runtimeEventingURLLabelKey := "runtime/event_service_url"
	runtime1EventingURL := "http://eventing.runtime1.local"
	runtime2EventingURL := "http://eventing.runtime2.local"
	defaultScenarios := []string{"DEFAULT"}

	application := registerApplication(t, ctx, "app-test-eventing")
	defer unregisterApplication(t, application.ID)

	runtime1 := registerRuntime(t, ctx, "runtime-1-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime1.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime1.ID, scenariosLabel, defaultScenarios)
	setRuntimeLabel(t, ctx, runtime1.ID, runtimeEventingURLLabelKey, runtime1EventingURL)

	runtime2 := registerRuntime(t, ctx, "runtime-2-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime2.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime2.ID, scenariosLabel, defaultScenarios)
	setRuntimeLabel(t, ctx, runtime2.ID, runtimeEventingURLLabelKey, runtime2EventingURL)

	testApp := getApp(ctx, t, application.ID)
	require.Equal(t, runtime1EventingURL, testApp.EventingConfiguration.DefaultURL)

	actualEventingCfg := graphql.ApplicationEventingConfiguration{}
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: setDefaultEventingForApplication(appID: "%s", runtimeID: "%s") {
						%s
					}
				}`,
			application.ID, runtime2.ID, tc.gqlFieldsProvider.ForEventingConfiguration()))
	err := tc.RunOperation(ctx, request, &actualEventingCfg)

	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "set default eventing for application")
	require.NoError(t, err)
	require.Equal(t, runtime2EventingURL, actualEventingCfg.DefaultURL)

	testApp = getApp(ctx, t, application.ID)
	require.Equal(t, runtime2EventingURL, testApp.EventingConfiguration.DefaultURL)
}

func TestDeleteDefaultEventingForApplication(t *testing.T) {
	ctx := context.Background()
	runtimeEventingURLLabelKey := "runtime/event_service_url"
	runtime1EventingURL := "http://eventing.runtime1.local"
	runtime2EventingURL := "http://eventing.runtime2.local"
	defaultScenarios := []string{"DEFAULT"}

	application := registerApplication(t, ctx, "app-test-eventing")
	defer unregisterApplication(t, application.ID)

	runtime1 := registerRuntime(t, ctx, "runtime-1-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime1.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime1.ID, scenariosLabel, defaultScenarios)
	setRuntimeLabel(t, ctx, runtime1.ID, runtimeEventingURLLabelKey, runtime1EventingURL)

	runtime2 := registerRuntime(t, ctx, "runtime-2-eventing")
	defer unregisterRuntimeWithinTenant(t, runtime2.ID, defaultTenant)

	setRuntimeLabel(t, ctx, runtime2.ID, scenariosLabel, defaultScenarios)
	setRuntimeLabel(t, ctx, runtime2.ID, runtimeEventingURLLabelKey, runtime2EventingURL)

	testApp := getApp(ctx, t, application.ID)
	require.Equal(t, runtime1EventingURL, testApp.EventingConfiguration.DefaultURL)

	setDefaultEventingForApplication(t, ctx, application.ID, runtime2.ID)

	testApp = getApp(ctx, t, application.ID)
	require.Equal(t, runtime2EventingURL, testApp.EventingConfiguration.DefaultURL)

	actualEventingCfg := graphql.ApplicationEventingConfiguration{}
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: deleteDefaultEventingForApplication(appID: "%s") {
						%s
					}
				}`,
			application.ID, tc.gqlFieldsProvider.ForEventingConfiguration()))
	err := tc.RunOperation(ctx, request, &actualEventingCfg)

	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "delete default eventing for application")
	require.NoError(t, err)
	require.Equal(t, runtime2EventingURL, actualEventingCfg.DefaultURL)

	testApp = getApp(ctx, t, application.ID)
	require.Equal(t, runtime1EventingURL, testApp.EventingConfiguration.DefaultURL)
}
