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

	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/gql"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestSelfRegisterFlow(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("TestSelfRegisterFlow flow: label definitions of the parent tenant are not overwritten", func(t *testing.T) {
		ctx := context.Background()
		distinguishLblValue := "test-distinguish-value"

		// defaultTenantId is the parent of the subaccountID
		defaultTenantId := tenant.TestTenants.GetDefaultTenantID()
		subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

		// Build graphql director client configured with certificate
		clientKey, rawCertChain := certs.IssueExternalIssuerCertificate(t, testConfig.CA.Certificate, testConfig.CA.Key, subaccountID)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, clientKey, rawCertChain)

		// Register application
		app, err := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "testingApp", defaultTenantId)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, defaultTenantId, &app)
		require.NoError(t, err)
		require.NotEmpty(t, app.ID)

		// Create label definition
		scenarios := []string{"DEFAULT", "sr-test-scenario"}
		fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, defaultTenantId, scenarios)
		defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, defaultTenantId, scenarios[:1])

		// Assign application to scenario
		appLabelRequest := fixtures.FixSetApplicationLabelRequest(app.ID, scenariosLabel, scenarios[1:])
		require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, defaultTenantId, appLabelRequest, nil))
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, defaultTenantId, app.ID, testConfig.DefaultScenarioEnabled)

		// Self register runtime
		runtimeInput := graphql.RuntimeInput{
			Name:        "selfRegisterRuntime",
			Description: ptr.String("selfRegisterRuntime-description"),
			Labels:      graphql.Labels{testConfig.SelfRegisterDistinguishLabelKey: distinguishLblValue},
		}
		runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorCertSecuredClient, defaultTenantId, &runtimeInput)
		defer fixtures.CleanupRuntime(t, ctx, directorCertSecuredClient, defaultTenantId, &runtime)
		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)
		strLbl, ok := runtime.Labels[testConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, strLbl, distinguishLblValue)

		labelDefinitions, err := fixtures.ListLabelDefinitionsWithinTenant(t, ctx, dexGraphQLClient, defaultTenantId)
		require.NoError(t, err)
		numOfScenarioLabelDefinitions := 0
		for _, ld := range labelDefinitions {
			if ld.Key == scenariosLabel {
				numOfScenarioLabelDefinitions++
			}
		}
		// the parent tenant should not see child label definitions
		require.Equal(t, 1, numOfScenarioLabelDefinitions)
	})
}

func TestConsumerProviderFlow(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("ConsumerProvider flow: calls with provider certificate and consumer token are successful when valid subscription exists", func(t *testing.T) {
		ctx := context.Background()
		defaultTenantId := tenant.TestTenants.GetDefaultTenantID()
		secondaryTenant := tenant.TestTenants.GetIDByName(t, tenant.ApplicationsForRuntimeTenantName)
		subscriptionProviderID := "xs-app-name"
		subscriptionProviderSubaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)
		subscriptionConsumerSubaccountID := "1f538f34-30bf-4d3d-aeaa-02e69eef84ae"

		// Build graphql director client configured with certificate
		clientKey, rawCertChain := certs.IssueExternalIssuerCertificate(t, testConfig.CA.Certificate, testConfig.CA.Key, subscriptionProviderSubaccountID)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, clientKey, rawCertChain)

		runtimeInput := graphql.RuntimeInput{
			Name:        "providerRuntime",
			Description: ptr.String("providerRuntime-description"),
			Labels:      graphql.Labels{testConfig.SubscriptionProviderLabelKey: subscriptionProviderID, tenantfetcher.RegionKey: tenantfetcher.RegionPathParamValue, selectorKey: subscriptionProviderSubaccountID},
		}

		// Register provider runtime with the necessary label
		runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorCertSecuredClient, defaultTenantId, &runtimeInput)
		defer fixtures.CleanupRuntime(t, ctx, directorCertSecuredClient, defaultTenantId, &runtime)
		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)

		// Register application
		app, err := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "testingApp", secondaryTenant)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, secondaryTenant, &app)
		require.NoError(t, err)
		require.NotEmpty(t, app.ID)

		// Register consumer application
		consumerApp, err := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "consumerApp", secondaryTenant)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, secondaryTenant, &consumerApp)
		require.NoError(t, err)
		require.NotEmpty(t, consumerApp.ID)
		require.NotEmpty(t, consumerApp.Name)

		// Create label definition
		scenarios := []string{"DEFAULT", "consumer-test-scenario"}
		fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios)
		defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios[:1])

		// Assign consumer application to scenario
		appLabelRequest := fixtures.FixSetApplicationLabelRequest(consumerApp.ID, scenariosLabel, scenarios[1:])
		require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, secondaryTenant, appLabelRequest, nil))
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, secondaryTenant, consumerApp.ID, testConfig.DefaultScenarioEnabled)

		providedTenantIDs := tenantfetcher.Tenant{
			TenantID:               secondaryTenant,
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

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		// Build a request for consumer subscription
		subscribeReq := tenantfetcher.CreateTenantRequest(t, providedTenantIDs, tenantProperties, http.MethodPut, testConfig.TenantFetcherFullRegionalURL, testConfig.ExternalServicesMockURL, testConfig.ClientID, testConfig.ClientSecret)

		t.Log(fmt.Sprintf("Creating a subscription between consumer with subaccount id: %s and provider with name: %s and subaccount id: %s", tenantfetcher.ActualTenantID(providedTenantIDs), runtime.Name, subscriptionProviderSubaccountID))
		subscribeResp, err := httpClient.Do(subscribeReq)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, subscribeResp.StatusCode)

		defer func() {
			subscribeReq := tenantfetcher.CreateTenantRequest(t, providedTenantIDs, tenantProperties, http.MethodDelete, testConfig.TenantFetcherFullRegionalURL, testConfig.ExternalServicesMockURL, testConfig.ClientID, testConfig.ClientSecret)

			t.Log(fmt.Sprintf("Deleting a subscription between consumer with subaccount id: %s and provider with name: %s and subaccount id: %s", tenantfetcher.ActualTenantID(providedTenantIDs), runtime.Name, subscriptionProviderSubaccountID))
			subscribeResp, err := httpClient.Do(subscribeReq)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, subscribeResp.StatusCode)
		}()

		// Create automatic scenario assigment for consumer subaccount
		asaInput := fixtures.FixAutomaticScenarioAssigmentInput(scenarios[1], selectorKey, subscriptionConsumerSubaccountID)
		fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, secondaryTenant)
		defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios[1])

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
		headers := map[string][]string{"Authorization": {fmt.Sprintf("Bearer %s", token.FromExternalServicesMock(t, testConfig.ExternalServicesMockURL, testConfig.ClientID, testConfig.ClientSecret, claims))}}

		// Make a request to the ORD service with http client containing certificate with provider information and token with the consumer data.
		respBody := makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)

		require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
		require.Equal(t, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())

		// Build unsubscribe request
		unsubscribeReq := tenantfetcher.CreateTenantRequest(t, providedTenantIDs, tenantProperties, http.MethodDelete, testConfig.TenantFetcherFullRegionalURL, testConfig.ExternalServicesMockURL, testConfig.ClientID, testConfig.ClientSecret)

		t.Log(fmt.Sprintf("Remove a subscription between consumer with subaccount id: %s and provider with name: %s and subaccount id: %s", tenantfetcher.ActualTenantID(providedTenantIDs), runtime.Name, subscriptionProviderSubaccountID))
		unsubscribeResp, err := httpClient.Do(unsubscribeReq)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, unsubscribeResp.StatusCode)
		respBody = makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)

		require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
	})
}
