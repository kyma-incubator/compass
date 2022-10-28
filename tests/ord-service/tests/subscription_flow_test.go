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
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/subscription"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestSelfRegisterFlow(t *testing.T) {
	ctx := context.Background()
	accountTenantID := conf.AccountTenantID // accountTenantID is parent of the tenant/subaccountID of the configured certificate client's tenant below

	// Register application
	app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "testingApp", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), accountTenantID)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, accountTenantID, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	formationName := "sr-test-scenario"
	t.Logf("Creating formation with name %s...", formationName)
	createFormationReq := fixtures.FixCreateFormationRequest(formationName)
	executeGQLRequest(t, ctx, createFormationReq, formationName, accountTenantID)
	t.Logf("Successfully created formation: %s", formationName)

	defer func() {
		t.Logf("Deleting formation with name: %s...", formationName)
		deleteRequest := fixtures.FixDeleteFormationRequest(formationName)
		executeGQLRequest(t, ctx, deleteRequest, formationName, accountTenantID)
		t.Logf("Successfully deleted formation with name: %s...", formationName)
	}()

	t.Logf("Assign application to formation %s", formationName)
	assignToFormation(t, ctx, app.ID, string(graphql.FormationObjectTypeApplication), formationName, accountTenantID)
	t.Logf("Successfully assigned application to formation %s", formationName)

	defer func() {
		t.Logf("Unassign application from formation %s", formationName)
		unassignFromFormation(t, ctx, app.ID, string(graphql.FormationObjectTypeApplication), formationName, accountTenantID)
		t.Logf("Successfully unassigned application from formation %s", formationName)
	}()

	// Self register runtime
	runtimeInput := graphql.RuntimeRegisterInput{
		Name:        "selfRegisterRuntime",
		Description: ptr.String("selfRegisterRuntime-description"),
		Labels:      graphql.Labels{conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue},
	}
	runtime := fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, certSecuredGraphQLClient, &runtimeInput)
	defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, certSecuredGraphQLClient, &runtime)
	require.NotEmpty(t, runtime.ID)
	strLbl, ok := runtime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
	require.True(t, ok)
	require.Contains(t, strLbl, runtime.ID)

	saasAppLbl, ok := runtime.Labels[conf.SaaSAppNameLabelKey].(string)
	require.True(t, ok)
	require.NotEmpty(t, saasAppLbl)

	regionLbl, ok := runtime.Labels[tenantfetcher.RegionKey].(string)
	require.True(t, ok)
	require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, regionLbl)

	// Verify that the label returned cannot be modified
	setLabelRequest := fixtures.FixSetRuntimeLabelRequest(runtime.ID, conf.SubscriptionConfig.SelfRegisterLabelKey, "value")
	label := graphql.Label{}
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, setLabelRequest, &label)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("could not set unmodifiable label with key %s", conf.SubscriptionConfig.SelfRegisterLabelKey))

	labelDefinitions, err := fixtures.ListLabelDefinitionsWithinTenant(t, ctx, certSecuredGraphQLClient, accountTenantID)
	require.NoError(t, err)
	numOfScenarioLabelDefinitions := 0
	for _, ld := range labelDefinitions {
		if ld.Key == scenariosLabel {
			numOfScenarioLabelDefinitions++
		}
	}
	// the parent tenant should not see child label definitions
	require.Equal(t, 1, numOfScenarioLabelDefinitions)
}

func TestConsumerProviderFlow(stdT *testing.T) {
	t := testingx.NewT(stdT)

	ctx := context.Background()
	secondaryTenant := conf.TestConsumerAccountID
	subscriptionProviderSubaccountID := conf.TestProviderSubaccountID
	subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID
	subscriptionConsumerTenantID := conf.TestConsumerTenantID

	// We need an externally issued cert with a subject that is not part of the access level mappings
	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestSecretName:         conf.CertSvcInstanceTestSecretName,
		ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestExternalCertSubject:               strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.ExternalCertProviderConfig.TestExternalCertCN, "ord-service-subscription-cn", -1),
		ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
		ExternalCertProvider:                  certprovider.CertificateService,
	}

	// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(stdT, ctx, externalCertProviderConfig)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

	t.Run("ConsumerProvider flow", func(stdT *testing.T) {
		runtimeInput := graphql.RuntimeRegisterInput{
			Name:        "providerRuntime",
			Description: ptr.String("providerRuntime-description"),
			Labels: graphql.Labels{
				conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue},
		}

		runtime := fixtures.RegisterRuntimeFromInputWithoutTenant(stdT, ctx, directorCertSecuredClient, &runtimeInput)
		defer fixtures.CleanupRuntimeWithoutTenant(stdT, ctx, directorCertSecuredClient, &runtime)
		require.NotEmpty(stdT, runtime.ID)

		regionLbl, ok := runtime.Labels[tenantfetcher.RegionKey].(string)
		require.True(t, ok)
		require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, regionLbl)

		saasAppLbl, ok := runtime.Labels[conf.SaaSAppNameLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, saasAppLbl)

		// Register application
		app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "testingApp", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), secondaryTenant)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, &app)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, app.ID)

		// Register consumer application
		const localTenantID = "localTenantID"
		consumerApp, err := fixtures.RegisterApplicationWithTypeAndLocalTenantID(t, ctx, certSecuredGraphQLClient, "consumerApp", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), localTenantID, secondaryTenant)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, &consumerApp)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, consumerApp.ID)
		require.NotEmpty(stdT, consumerApp.Name)

		const correlationID = "correlationID"
		bndlInput := graphql.BundleCreateInput{
			Name:           "test-bundle",
			CorrelationIDs: []string{correlationID},
		}
		bundle := fixtures.CreateBundleWithInput(t, ctx, certSecuredGraphQLClient, secondaryTenant, consumerApp.ID, bndlInput)
		require.NotEmpty(stdT, bundle.ID)

		consumerFormationName := "consumer-test-scenario"
		stdT.Logf("Creating formation with name %s...", consumerFormationName)
		createFormationReq := fixtures.FixCreateFormationRequest(consumerFormationName)
		executeGQLRequest(stdT, ctx, createFormationReq, consumerFormationName, secondaryTenant)
		stdT.Logf("Successfully created formation: %s", consumerFormationName)

		defer func() {
			stdT.Logf("Deleting formation with name: %s...", consumerFormationName)
			deleteRequest := fixtures.FixDeleteFormationRequest(consumerFormationName)
			executeGQLRequest(stdT, ctx, deleteRequest, consumerFormationName, secondaryTenant)
			stdT.Logf("Successfully deleted formation with name: %s...", consumerFormationName)
		}()

		stdT.Logf("Assign application to formation %s", consumerFormationName)
		assignToFormation(stdT, ctx, consumerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned application to formation %s", consumerFormationName)

		defer func() {
			stdT.Logf("Unassign application from formation %s", consumerFormationName)
			unassignFromFormation(stdT, ctx, consumerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
			stdT.Logf("Successfully unassigned application from formation %s", consumerFormationName)
		}()

		stdT.Logf("Assign tenant %s to formation %s...", subscriptionConsumerSubaccountID, consumerFormationName)
		assignToFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", consumerFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned tenant %s to formation %s", subscriptionConsumerSubaccountID, consumerFormationName)

		defer func() {
			stdT.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, consumerFormationName)
			unassignFromFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", consumerFormationName, secondaryTenant)
			stdT.Logf("Successfully unassigned tenant %s to formation %s", subscriptionConsumerSubaccountID, consumerFormationName)
		}()

		selfRegLabelValue, ok := runtime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(stdT, ok)
		require.Contains(stdT, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+runtime.ID)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(stdT, err)
		response, err := httpClient.Do(depConfigureReq)
		require.NoError(stdT, err)
		defer func() {
			if err := response.Body.Close(); err != nil {
				stdT.Logf("Could not close response body %s", err)
			}
		}()
		require.Equal(stdT, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/application/tenants/%s/subscriptions", subscriptionConsumerTenantID)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(stdT, err)
		subscriptionToken := token.GetClientCredentialsToken(stdT, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(stdT, runtime.ID, runtime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		stdT.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, runtime.Name, runtime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(stdT, runtime.ID, runtime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
		require.NoError(stdT, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				stdT.Logf("Could not close response body %s", err)
			}
		}()
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(stdT, err)
		require.Equal(stdT, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(stdT, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(stdT, func() bool {
			return subscription.GetSubscriptionJobStatus(stdT, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		stdT.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, runtime.Name, runtime.ID, subscriptionProviderSubaccountID)

		// After successful subscription from above we call the director component with "double authentication(token + certificate)" in order to test claims validation is successful
		consumerToken := token.GetUserToken(stdT, ctx, conf.ConsumerTokenURL+conf.TokenPath, conf.ProviderClientID, conf.ProviderClientSecret, conf.BasicUsername, conf.BasicPassword, "subscriptionClaims")
		headers := map[string][]string{util.AuthorizationHeader: {fmt.Sprintf("Bearer %s", consumerToken)}}

		stdT.Log("Calling director to verify claims validation is successful...")
		getRtmReq := fixtures.FixGetRuntimeRequest(runtime.ID)
		getRtmReq.Header = headers
		rtmExt := graphql.RuntimeExt{}

		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, getRtmReq, &rtmExt)
		require.NoError(stdT, err)
		require.Equal(stdT, runtime.ID, rtmExt.ID)
		require.Equal(stdT, runtimeInput.Name, rtmExt.Name)
		stdT.Log("Director claims validation was successful")

		// Create destination that matches to the created bundle
		region := conf.SubscriptionConfig.SelfRegRegion
		instance, ok := conf.DestinationsConfig.RegionToInstanceConfig[region]
		require.True(t, ok)

		subdomain := conf.DestinationConsumerSubdomain
		client, err := clients.NewDestinationClient(instance, conf.DestinationAPIConfig, subdomain)
		require.NoError(stdT, err)

		destination := clients.Destination{
			Name:            "test",
			Type:            "HTTP",
			URL:             "http://localhost",
			Authentication:  "BasicAuthentication",
			XCorrelationID:  correlationID,
			XSystemTenantID: localTenantID,
			XSystemType:     string(util.ApplicationTypeC4C),
		}

		client.CreateDestination(stdT, destination)
		defer client.DeleteDestination(stdT, destination.Name)
		// After successful subscription from above, the part of the code below prepare and execute a request to the ord service

		// HTTP client configured with certificate with patched subject, issued from cert-rotation job
		certHttpClient := CreateHttpClientWithCert(providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		// Make a request to the ORD service with http client containing certificate with provider information and token with the consumer data.
		stdT.Log("Getting consumer application using both provider and consumer credentials...")
		respBody := makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)
		require.Len(stdT, gjson.Get(respBody, "value").Array(), 1)
		require.Equal(stdT, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())
		stdT.Log("Successfully fetched consumer application using both provider and consumer credentials")

		// Make a request to the ORD service expanding bundles and destinations.
		// With no destinations
		respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+
			"/systemInstances?$expand=consumptionBundles($expand=destinations)&$format=json", headers)

		require.Len(stdT, gjson.Get(respBody, "value").Array(), 1)
		require.Equal(stdT, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())
		require.NotEmpty(stdT, gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Raw)
		require.Empty(stdT, gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Array())
		stdT.Log("Successfully fetched system with bundles with no destinations")

		// With destinations
		respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+
			"/systemInstances?$expand=consumptionBundles($expand=destinations)&$format=json&reload=true", headers)
		require.Equal(stdT, 1, len(gjson.Get(respBody, "value").Array()))
		require.Equal(stdT, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())
		require.NotEmpty(stdT, gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Raw)
		destinationsFromResponse := gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Array()
		require.Len(stdT, destinationsFromResponse, 1)
		require.Equal(stdT, destination.Name, destinationsFromResponse[0].Get("sensitiveData.destinationConfiguration.Name").String())
		stdT.Log("Successfully fetched system with bundles and destinations")

		subscription.BuildAndExecuteUnsubscribeRequest(stdT, runtime.ID, runtime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		stdT.Log("Validating no application is returned after successful unsubscription request...")
		respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)
		require.Empty(stdT, gjson.Get(respBody, "value").Array())
		stdT.Log("Successfully validated no application is returned after successful unsubscription request")

		stdT.Log("Validating no destination is returned after successful unsubscription request...")
		respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/destinations?$format=json", headers)
		require.Empty(stdT, gjson.Get(respBody, "value").Array())
		stdT.Log("Successfully validated no destination is returned after successful unsubscription request")

		stdT.Log("Validating director returns error during claims validation after unsubscribe request is successfully executed...")
		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, getRtmReq, &rtmExt)
		require.Error(stdT, err)
		require.Contains(stdT, err.Error(), fmt.Sprintf("Consumer's external tenant %s was not found as subscription record in the runtime context table for the runtime in the provider tenant", subscriptionConsumerSubaccountID))
		stdT.Log("Successfully validated an error is returned during claims validation after unsubscribe request")
	})

	t.Run("Consumer provider with user context header", func(t *testing.T) {
		ctx = context.Background()

		runtimeInput := graphql.RuntimeRegisterInput{
			Name:        "providerRuntime-with-user-context-header",
			Description: ptr.String("providerRuntime-with-user-context-header-description"),
			Labels: graphql.Labels{
				conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue,
			},
		}

		runtime := fixtures.RegisterRuntimeFromInputWithoutTenant(stdT, ctx, directorCertSecuredClient, &runtimeInput)
		defer fixtures.CleanupRuntimeWithoutTenant(stdT, ctx, directorCertSecuredClient, &runtime)
		require.NotEmpty(stdT, runtime.ID)

		regionLbl, ok := runtime.Labels[tenantfetcher.RegionKey].(string)
		require.True(t, ok)
		require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, regionLbl)

		saasAppLbl, ok := runtime.Labels[conf.SaaSAppNameLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, saasAppLbl)

		// Register application
		app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "testingApp", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), secondaryTenant)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, &app)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, app.ID)

		// Register consumer application
		const localTenantID = "localTenantID"
		consumerApp, err := fixtures.RegisterApplicationWithTypeAndLocalTenantID(t, ctx, certSecuredGraphQLClient, "consumerApp", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), localTenantID, secondaryTenant)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, &consumerApp)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, consumerApp.ID)
		require.NotEmpty(stdT, consumerApp.Name)

		const correlationID = "correlationID"
		bndlInput := graphql.BundleCreateInput{
			Name:           "test-bundle",
			CorrelationIDs: []string{correlationID},
		}
		bundle := fixtures.CreateBundleWithInput(t, ctx, certSecuredGraphQLClient, secondaryTenant, consumerApp.ID, bndlInput)
		require.NotEmpty(stdT, bundle.ID)

		consumerFormationName := "consumer-test-scenario"
		stdT.Logf("Creating formation with name %s...", consumerFormationName)
		createFormationReq := fixtures.FixCreateFormationRequest(consumerFormationName)
		executeGQLRequest(stdT, ctx, createFormationReq, consumerFormationName, secondaryTenant)
		stdT.Logf("Successfully created formation: %s", consumerFormationName)

		defer func() {
			stdT.Logf("Deleting formation with name: %s...", consumerFormationName)
			deleteRequest := fixtures.FixDeleteFormationRequest(consumerFormationName)
			executeGQLRequest(stdT, ctx, deleteRequest, consumerFormationName, secondaryTenant)
			stdT.Logf("Successfully deleted formation with name: %s...", consumerFormationName)
		}()

		stdT.Logf("Assign application to formation %s", consumerFormationName)
		assignToFormation(stdT, ctx, consumerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned application to formation %s", consumerFormationName)

		defer func() {
			stdT.Logf("Unassign application from formation %s", consumerFormationName)
			unassignFromFormation(stdT, ctx, consumerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
			stdT.Logf("Successfully unassigned application from formation %s", consumerFormationName)
		}()

		stdT.Logf("Assign tenant %s to formation %s...", subscriptionConsumerSubaccountID, consumerFormationName)
		assignToFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", consumerFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned tenant %s to formation %s", subscriptionConsumerSubaccountID, consumerFormationName)

		defer func() {
			stdT.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, consumerFormationName)
			unassignFromFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", consumerFormationName, secondaryTenant)
			stdT.Logf("Successfully unassigned tenant %s to formation %s", subscriptionConsumerSubaccountID, consumerFormationName)
		}()

		selfRegLabelValue, ok := runtime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(stdT, ok)
		require.Contains(stdT, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+runtime.ID)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(stdT, err)
		response, err := httpClient.Do(depConfigureReq)
		require.NoError(stdT, err)
		defer func() {
			if err := response.Body.Close(); err != nil {
				stdT.Logf("Could not close response body %s", err)
			}
		}()
		require.Equal(stdT, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/application/tenants/%s/subscriptions", subscriptionConsumerTenantID)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(stdT, err)
		subscriptionToken := token.GetClientCredentialsToken(stdT, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(stdT, runtime.ID, runtime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		stdT.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, runtime.Name, runtime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		require.NoError(stdT, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				stdT.Logf("Could not close response body %s", err)
			}
		}()
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(stdT, err)
		require.Equal(stdT, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		defer subscription.BuildAndExecuteUnsubscribeRequest(stdT, runtime.ID, runtime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(stdT, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(stdT, func() bool {
			return subscription.GetSubscriptionJobStatus(stdT, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		stdT.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, runtime.Name, runtime.ID, subscriptionProviderSubaccountID)

		// After successful subscription from above we call the director component with "double authentication(token + user_context header)" in order to test claims validation is successful
		consumerToken := token.GetUserToken(stdT, ctx, conf.ConsumerTokenURL+conf.TokenPath, conf.ProviderClientID, conf.ProviderClientSecret, conf.BasicUsername, conf.BasicPassword, "subscriptionClaims")
		consumerClaims := token.FlattenTokenClaims(stdT, consumerToken)
		headers := map[string][]string{subscription.UserContextHeader: {consumerClaims}}

		stdT.Log("Calling director to verify claims validation is successful...")
		getRtmReq := fixtures.FixGetRuntimeRequest(runtime.ID)
		getRtmReq.Header = headers
		rtmExt := graphql.RuntimeExt{}

		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, getRtmReq, &rtmExt)
		require.NoError(stdT, err)
		require.Equal(stdT, runtime.ID, rtmExt.ID)
		require.Equal(stdT, runtimeInput.Name, rtmExt.Name)
		stdT.Log("Director claims validation was successful")

		// Create destination that matches to the created bundle
		region := conf.SubscriptionConfig.SelfRegRegion
		instance, ok := conf.DestinationsConfig.RegionToInstanceConfig[region]
		require.True(t, ok)

		subdomain := conf.DestinationConsumerSubdomain
		client, err := clients.NewDestinationClient(instance, conf.DestinationAPIConfig, subdomain)
		require.NoError(stdT, err)

		destination := clients.Destination{
			Name:            "test",
			Type:            "HTTP",
			URL:             "http://localhost",
			Authentication:  "BasicAuthentication",
			XCorrelationID:  correlationID,
			XSystemTenantID: localTenantID,
			XSystemType:     string(util.ApplicationTypeC4C),
		}

		client.CreateDestination(stdT, destination)
		defer client.DeleteDestination(stdT, destination.Name)
		// After successful subscription from above, the part of the code below prepare and execute a request to the ord service

		// HTTP client configured with certificate with patched subject, issued from cert-rotation job
		certHttpClient := CreateHttpClientWithCert(providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		// Make a request to the ORD service with http client containing certificate with provider information and token with the consumer data.
		stdT.Log("Getting consumer application using both provider and consumer credentials...")
		respBody := makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)
		require.Len(stdT, gjson.Get(respBody, "value").Array(), 1)
		require.Equal(stdT, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())
		stdT.Log("Successfully fetched consumer application using both provider and consumer credentials")

		// Make a request to the ORD service expanding bundles and destinations.
		// With no destinations
		respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+
			"/systemInstances?$expand=consumptionBundles($expand=destinations)&$format=json", headers)

		require.Len(stdT, gjson.Get(respBody, "value").Array(), 1)
		require.Equal(stdT, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())
		require.NotEmpty(stdT, gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Raw)
		require.Empty(stdT, gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Array())
		stdT.Log("Successfully fetched system with bundles with no destinations")

		// With destinations
		respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+
			"/systemInstances?$expand=consumptionBundles($expand=destinations)&$format=json&reload=true", headers)
		require.Equal(stdT, 1, len(gjson.Get(respBody, "value").Array()))
		require.Equal(stdT, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())
		require.NotEmpty(stdT, gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Raw)
		destinationsFromResponse := gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Array()
		require.Len(stdT, destinationsFromResponse, 1)
		require.Equal(stdT, destination.Name, destinationsFromResponse[0].Get("sensitiveData.destinationConfiguration.Name").String())
		stdT.Log("Successfully fetched system with bundles and destinations")

		subscription.BuildAndExecuteUnsubscribeRequest(stdT, runtime.ID, runtime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		stdT.Log("Validating no application is returned after successful unsubscription request...")
		respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)
		require.Empty(stdT, gjson.Get(respBody, "value").Array())
		stdT.Log("Successfully validated no application is returned after successful unsubscription request")

		stdT.Log("Validating no destination is returned after successful unsubscription request...")
		respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/destinations?$format=json", headers)
		require.Empty(stdT, gjson.Get(respBody, "value").Array())
		stdT.Log("Successfully validated no destination is returned after successful unsubscription request")

		stdT.Log("Validating director returns error during claims validation after unsubscribe request is successfully executed...")
		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, getRtmReq, &rtmExt)
		require.Error(stdT, err)
		require.Contains(stdT, err.Error(), fmt.Sprintf("Consumer's external tenant %s was not found as subscription record in the runtime context table for the runtime in the provider tenant", subscriptionConsumerSubaccountID))
		stdT.Log("Successfully validated an error is returned during claims validation after unsubscribe request")
	})
}

func assignToFormation(t *testing.T, ctx context.Context, objectID, objectType, formationName, tenantID string) {
	assignReq := fixtures.FixAssignFormationRequest(objectID, objectType, formationName)
	executeGQLRequest(t, ctx, assignReq, formationName, tenantID)
}

func unassignFromFormation(t *testing.T, ctx context.Context, objectID, objectType, formationName, tenantID string) {
	unassignReq := fixtures.FixUnassignFormationRequest(objectID, objectType, formationName)
	executeGQLRequest(t, ctx, unassignReq, formationName, tenantID)
}

func executeGQLRequest(t *testing.T, ctx context.Context, gqlRequest *gcli.Request, formationName, tenantID string) {
	var formation graphql.Formation
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, gqlRequest, &formation)
	require.NoError(t, err)
	require.Equal(t, formationName, formation.Name)
}
