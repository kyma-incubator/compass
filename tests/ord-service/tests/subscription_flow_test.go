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
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestSubscriptionFlow(t *testing.T) {
	ctx := context.Background()
	defaultTenantId := tenant.TestTenants.GetDefaultTenantID()
	secondaryTenant := tenant.TestTenants.GetIDByName(t, tenant.ApplicationsForRuntimeTenantName)
	accountID := uuid.New().String()
	subscriptionProviderID := "xs-app-name"
	subscriptionProviderSubaccountID := "f8075207-1478-4a80-bd26-24a4785a2bfd"
	subscriptionConsumerSubaccountID := "1f538f34-30bf-4d3d-aeaa-02e69eef84ae"

	runtimeInput := graphql.RuntimeInput{
		Name:        "providerRuntime",
		Description: ptr.String("providerRuntime-description"),
		Labels:      graphql.Labels{testConfig.SubscriptionProviderLabelKey: subscriptionProviderID, tenantfetcher.RegionKey: tenantfetcher.RegionPathParamValue, selectorKey: subscriptionProviderSubaccountID},
	}

	// Register provider runtime with the necessary label
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
	require.NotEmpty(t, consumerApp.Name)

	// Create label definition
	scenarios := []string{"DEFAULT", "consumer-test-scenario"}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios[:1])

	// Create automatic scenario assigment for consumer subaccount
	asaInput := fixtures.FixAutomaticScenarioAssigmentInput(scenarios[1], selectorKey, subscriptionConsumerSubaccountID)
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, secondaryTenant)
	defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios[1])

	// Assign application to scenario
	appLabelRequest := fixtures.FixSetApplicationLabelRequest(consumerApp.ID, scenariosLabel, scenarios[1:])
	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, secondaryTenant, appLabelRequest, nil))
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, secondaryTenant, consumerApp.ID, testConfig.DefaultScenarioEnabled)

	providedTenantIDs := tenantfetcher.Tenant{
		TenantID:               accountID,
		SubaccountID:           subscriptionConsumerSubaccountID,
		Subdomain:              tenantfetcher.DefaultSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	}

	tenantProperties := tenantfetcher.TenantIDProperties{
		TenantIDProperty:               testConfig.TenantIDProperty,
		SubaccountTenantIDProperty:     testConfig.SubaccountTenantIDProperty,
		CustomerIDProperty:             testConfig.CustomerIDProperty,
		SubdomainProperty:              testConfig.SubdomainProperty,
		SubscriptionProviderIDProperty: testConfig.SubscriptionProviderIDProperty,
	}

	// Build a request for consumer subscription
	request := tenantfetcher.CreateTenantRequest(t, providedTenantIDs, tenantProperties, http.MethodPut, testConfig.TenantFetcherFullRegionalURL, testConfig.ExternalServicesMockURL)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	t.Log(fmt.Sprintf("Creating a subscription between consumer with subaccount id: %s and provider with name: %s and subaccount id: %s", tenantfetcher.ActualTenantID(providedTenantIDs), runtime.Name, subscriptionProviderSubaccountID))
	response, err := httpClient.Do(request)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	tenantfetcher.AssertRuntimeSubscription(t, ctx, runtime.ID, providedTenantIDs, dexGraphQLClient, testConfig.ConsumerSubaccountIdsLabelKey)

	// HTTP client configured with manually signed client certificate
	extIssuerCertHttpClient := extIssuerCertClient(t, subscriptionProviderSubaccountID)

	// Create a token with the necessary consumer claims and add it in authorization header
	claims := map[string]interface{}{
		"subsc-key-test": "subscription-flow",
		"scope":          []string{},
		"tenant":         subscriptionConsumerSubaccountID,
		"identity":       "subscription-flow-identity",
		"iss":            testConfig.ExternalServicesMockURL,
		"exp":            time.Now().Unix() + int64(time.Minute.Seconds()),
	}
	headers := map[string][]string{"Authorization": {fmt.Sprintf("Bearer %s", token.FromExternalServicesMock(t, testConfig.ExternalServicesMockURL, claims))}}

	// Make a request to the ORD service with http client containing certificate with provider information and token with the consumer data.
	respBody := makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)

	require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
	require.Equal(t, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())
}
