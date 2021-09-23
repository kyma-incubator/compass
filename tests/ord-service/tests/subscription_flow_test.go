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
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/tidwall/sjson"
	"net/http"
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

const (
	tenantPathParamValue       = "tenant"
	regionPathParamValue       = "eu-1"
	defaultSubdomain           = "default-subdomain"
)

type Tenant struct {
	TenantID               string
	SubaccountID           string
	CustomerID             string
	Subdomain              string
	SubscriptionProviderID string
}

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
	asaInput := fixtures.FixAutomaticScenarioAssigmentInput(scenariosLabel, selectorKey, subscriptionConsumerID)
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, secondaryTenant)
	defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarioName)

	// Set application scenarios label
	fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, consumerApp.ID, scenariosLabel, scenarios[1:])
	defer fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, consumerApp.ID, scenariosLabel, scenarios[:1])

	//// TODO: Adjust once the subscription labeling task is done;
	//fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, defaultTenantId, runtime.ID, testConfig.ConsumerSubaccountIdsLabelKey, []string{subscriptionConsumerID})
	//defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, defaultTenantId, runtime.ID, testConfig.ConsumerSubaccountIdsLabelKey, []string{""})

	providedTenant := Tenant{
		TenantID:               accountID,
		SubaccountID:           subscriptionConsumerID,
		Subdomain:              defaultSubdomain,
		SubscriptionProviderID: subscriptionProviderID,
	}

	// Build a request for consumer subscription
	request := createTenantRequest(t, providedTenant, http.MethodPut, testConfig.TenantFetcherFullRegionalURL)

	t.Log(fmt.Sprintf("Provisioning tenant with ID %s", actualTenantID(providedTenant)))
	response, err := httpClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)

	// HTTP client configured with manually signed client certificate
	extIssuerCertHttpClient := extIssuerCertClient(t, defaultTenantId)

	// Create a token with the necessary consumer claims and add it in authorization header
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

	// Make a request to the ORD service with http client containing certificate with provider information and token with the consumer data.
	respBody := makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)

	require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
}

// createTenantRequest returns a prepared tenant request with token in the header with the necessary claims for tenant-fetcher
func createTenantRequest(t *testing.T, tenant Tenant, httpMethod string, url string) *http.Request {
	var (
		body = "{}"
		err  error
	)

	if len(tenant.TenantID) > 0 {
		body, err = sjson.Set(body, testConfig.TenantIDProperty, tenant.TenantID)
		require.NoError(t, err)
	}
	if len(tenant.SubaccountID) > 0 {
		body, err = sjson.Set(body, testConfig.SubaccountTenantIDProperty, tenant.SubaccountID)
		require.NoError(t, err)
	}
	if len(tenant.CustomerID) > 0 {
		body, err = sjson.Set(body, testConfig.CustomerIDProperty, tenant.CustomerID)
		require.NoError(t, err)
	}
	if len(tenant.Subdomain) > 0 {
		body, err = sjson.Set(body, testConfig.SubdomainProperty, tenant.Subdomain)
		require.NoError(t, err)
	}
	if len(tenant.SubscriptionProviderID) > 0 {
		body, err = sjson.Set(body, testConfig.SubscriptionProviderIDProperty, tenant.SubscriptionProviderID)
		require.NoError(t, err)
	}

	request, err := http.NewRequest(httpMethod, url, bytes.NewBuffer([]byte(body)))
	require.NoError(t, err)
	claims := map[string]interface{}{
		"test": "tenant-fetcher",
		"scope": []string{
			"prefix.Callback",
		},
		"tenant":   "tenant",
		"identity": "tenant-fetcher-tests",
		"iss":      testConfig.ExternalServicesMockURL,
		"exp":      time.Now().Unix() + int64(time.Minute.Seconds()),
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.FetchTokenFromExternalServicesMock(t, testConfig.ExternalServicesMockURL, claims)))

	return request
}

func actualTenantID(tenant Tenant) string {
	if len(tenant.SubaccountID) > 0 {
		return tenant.SubaccountID
	}

	return tenant.TenantID
}
