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
	"crypto/tls"
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	tenantPathParamValue = "tenant"
	regionPathParamValue = "eu-1"
	defaultSubdomain     = "default-subdomain"
)

func TestSubscriptionFlow(t *testing.T) {
	ctx := context.Background()
	defaultTenantId := tenant.TestTenants.GetDefaultTenantID()
	secondaryTenant := tenant.TestTenants.GetIDByName(t, tenant.ApplicationsForRuntimeTenantName)
	subscriptionProviderID := "xs-app-name"
	subscriptionConsumerID := "1f538f34-30bf-4d3d-aeaa-02e69eef84ae"
	accountID := uuid.New().String()

	runtimeInput := graphql.RuntimeInput{
		Name:        "testingRuntime",
		Description: ptr.String("testingRuntime-description"),
		Labels:      graphql.Labels{testConfig.SubscriptionProviderLabelKey: subscriptionProviderID, "region": regionPathParamValue},
	}

	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, defaultTenantId, &runtimeInput)
	defer fixtures.CleanupRuntime(t, ctx, dexGraphQLClient, defaultTenantId, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	app, err := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "testingApp", secondaryTenant)
	defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, secondaryTenant, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	consumerApp, err := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "consumerApp", secondaryTenant)
	defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, secondaryTenant, &consumerApp)
	require.NoError(t, err)
	require.NotEmpty(t, consumerApp.ID)

	// Create label definition
	scenarios := []string{"DEFAULT", "consumer-test-scenario"}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios[:1])

	// Create automatic scenario assigment for consumer subaccount
	asaInput := fixtures.FixAutomaticScenarioAssigmentInput(scenarios[1], selectorKey, subscriptionConsumerID)
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, secondaryTenant)
	defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarioName)

	// Assign application to scenario
	appLabelRequest := fixtures.FixSetApplicationLabelRequest(consumerApp.ID, scenariosLabel, scenarios[1:])
	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, secondaryTenant, appLabelRequest, nil))
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, secondaryTenant, consumerApp.ID, testConfig.DefaultScenarioEnabled)

	providedTenantIDs := tenant.TenantIDs{
		TenantID:               accountID,
		SubaccountID:           subscriptionConsumerID,
		Subdomain:              defaultSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	}

	tenantProperties := tenant.TenantIDProperties{
		TenantIDProperty:               testConfig.TenantIDProperty,
		SubaccountTenantIDProperty:     testConfig.SubaccountTenantIDProperty,
		CustomerIDProperty:             testConfig.CustomerIDProperty,
		SubdomainProperty:              testConfig.SubdomainProperty,
		SubscriptionProviderIDProperty: testConfig.SubscriptionProviderIDProperty,
	}

	// Build a request for consumer subscription
	request := tenant.CreateTenantRequest(t, providedTenantIDs, tenantProperties, http.MethodPut, testConfig.TenantFetcherFullRegionalURL, testConfig.ExternalServicesMockURL)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	t.Log(fmt.Sprintf("Creating a subscription between consumer with id %s and provider with name %s", tenant.ActualTenantID(providedTenantIDs), runtime.Name))
	response, err := httpClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)

	// HTTP client configured with manually signed client certificate
	extIssuerCertHttpClient := extIssuerCertClient(t, defaultTenantId)

	// Create a token with the necessary consumer claims and add it in authorization header
	claims := map[string]interface{}{
		"test":     "bas-flow",
		"scope":    []string{},
		"tenant":   subscriptionConsumerID,
		"identity": "subscription-flow",
		"iss":      testConfig.ExternalServicesMockURL,
		"exp":      time.Now().Unix() + int64(time.Minute.Seconds()),
	}
	headers := map[string][]string{"Authorization": {fmt.Sprintf("Bearer %s", token.FetchTokenFromExternalServicesMock(t, testConfig.ExternalServicesMockURL, claims))}}

	// Make a request to the ORD service with http client containing certificate with provider information and token with the consumer data.
	respBody := makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)

	require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
	require.Equal(t, "consumerApp", gjson.Get(respBody, "value.0.title").String())
}
