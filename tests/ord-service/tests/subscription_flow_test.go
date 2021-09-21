/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestSubscriptionFlow(t *testing.T) {
	ctx := context.Background()
	defaultTenantId := tenant.TestTenants.GetDefaultTenantID()
	secondaryTenant := tenant.TestTenants.GetIDByName(t, tenant.ApplicationsForRuntimeTenantName)
	subscriptionConsumerID := "1f538f34-30bf-4d3d-aeaa-02e69eef84ae"

	runtimeInput := graphql.RuntimeInput{
		Name:        "testingRuntime",
		Description: ptr.String("testingRuntime-description"),
		Labels:      graphql.Labels{testConfig.SubscriptionProviderLabelKey: "xs-app-name", "region": "region-name"},
	}

	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, defaultTenantId, &runtimeInput)
	defer fixtures.CleanupRuntime(t, ctx, dexGraphQLClient, defaultTenantId, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	app, err := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "testingApp", defaultTenantId)
	defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, defaultTenantId, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	consumerApp, err := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "consumerApp", secondaryTenant)
	defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, secondaryTenant, &consumerApp)
	require.NoError(t, err)
	require.NotEmpty(t, consumerApp.ID)

	// create label definition
	scenarios := []string{"DEFAULT", "consumer-test-scenario"}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios[:1])

	// create automatic scenario assigment for consumer subaccount
	asaInput := fixtures.FixAutomaticScenarioAssigmentInput(scenariosLabel, selectorKey, subscriptionConsumerID)
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, secondaryTenant)
	defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarioName)

	// set application scenarios label
	fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, consumerApp.ID, scenariosLabel, scenarios[1:])
	defer fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, consumerApp.ID, scenariosLabel, scenarios[:1])

	// TODO: Adjust once the subscription labeling task is done
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, defaultTenantId, runtime.ID, testConfig.ConsumerSubaccountIdsLabelKey, []string{subscriptionConsumerID})
	defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, defaultTenantId, runtime.ID, testConfig.ConsumerSubaccountIdsLabelKey, []string{""})

	extIssuerCertHttpClient := extIssuerCertClient(t, defaultTenantId)

	claims := map[string]interface{}{
		"test": "bas-flow",
		"scope": []string{
			"prefix.Callback",
		},
		"tenant":   subscriptionConsumerID,
		"identity": "subscription-flow",
		"iss":      testConfig.ExternalServicesMockURL,
		"exp":      time.Now().Unix() + int64(time.Minute.Seconds()),
	}
	headers := map[string][]string{"Authorization": {fmt.Sprintf("Bearer %s", token.FetchTokenFromExternalServicesMock(t, testConfig.ExternalServicesMockURL, claims))}}
	respBody := makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)

	require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
}
