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

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/tidwall/sjson"

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

var baseURLTemplate = "http://%s.%s.subscription.com"

func TestSelfRegisterFlow(t *testing.T) {
	ctx := context.Background()

	// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

	accountTenantID := conf.AccountTenantID // accountTenantID is parent of the tenant/subaccountID of the configured certificate client's tenant below

	// Register application
	app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "testingApp", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), accountTenantID)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, accountTenantID, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	formationName := "sr-test-scenario"
	t.Logf("Creating formation with name %s...", formationName)
	createFormationReq := fixtures.FixCreateFormationRequest(formationName)
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, accountTenantID, formationName)
	executeGQLRequest(t, ctx, createFormationReq, formationName, accountTenantID)
	t.Logf("Successfully created formation: %s", formationName)

	t.Logf("Assign application to formation %s", formationName)
	assignToFormation(t, ctx, app.ID, string(graphql.FormationObjectTypeApplication), formationName, accountTenantID)
	defer unassignFromFormation(t, ctx, app.ID, string(graphql.FormationObjectTypeApplication), formationName, accountTenantID)
	t.Logf("Successfully assigned application to formation %s", formationName)

	// Self register runtime
	runtimeInput := graphql.RuntimeRegisterInput{
		Name:        "selfRegisterRuntime",
		Description: ptr.String("selfRegisterRuntime-description"),
		Labels:      graphql.Labels{conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue},
	}

	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &runtime)
	runtime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &runtimeInput)
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
	err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, setLabelRequest, &label)
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
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(stdT, ctx, externalCertProviderConfig, true)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

	t.Run("ConsumerProvider flow with runtime as provider", func(stdT *testing.T) {
		runtimeInput := graphql.RuntimeRegisterInput{
			Name:        "providerRuntime",
			Description: ptr.String("providerRuntime-description"),
			Labels: graphql.Labels{
				conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue},
		}

		var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntimeWithoutTenant(stdT, ctx, directorCertSecuredClient, &runtime)
		runtime = fixtures.RegisterRuntimeFromInputWithoutTenant(stdT, ctx, directorCertSecuredClient, &runtimeInput)
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

		// Register application directly in subscriptionConsumerSubaccountID (that is in the formation) and validate that this system won't be visible for the SaaS app calling as part of the formation. That way we test the ORD filtering based on formations.
		subaccountApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "e2e-test-subaccount-app", subscriptionConsumerSubaccountID)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, &subaccountApp)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, subaccountApp.ID)

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

		formationTmplName := "e2e-test-formation-template-name"
		applicationType := util.ApplicationTypeC4C

		stdT.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, formationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(stdT, ctx, certSecuredGraphQLClient, formationTmplName, conf.SubscriptionProviderAppNameValue, []string{string(applicationType)}, graphql.ArtifactTypeSubscription)

		consumerFormationName := "consumer-test-scenario"
		stdT.Logf("Creating formation with name: %q from template with name: %q", consumerFormationName, formationTmplName)
		defer fixtures.DeleteFormationWithinTenant(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, consumerFormationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, consumerFormationName, &formationTmplName)
		require.NotEmpty(t, formation.ID)
		stdT.Logf("Successfully created formation: %s", consumerFormationName)

		stdT.Logf("Assign application to formation %s", consumerFormationName)
		assignToFormation(stdT, ctx, consumerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
		defer unassignFromFormation(stdT, ctx, consumerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned application to formation %s", consumerFormationName)

		stdT.Logf("Assign tenant %s to formation %s...", subscriptionConsumerSubaccountID, consumerFormationName)
		assignToFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", consumerFormationName, secondaryTenant)
		defer unassignFromFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", consumerFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned tenant %s to formation %s", subscriptionConsumerSubaccountID, consumerFormationName)

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
		defer func() {
			if err := response.Body.Close(); err != nil {
				stdT.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(stdT, err)
		require.Equal(stdT, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(stdT, err)
		subscriptionToken := token.GetClientCredentialsToken(stdT, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)
		subscribeReq.Header.Add(conf.SubscriptionConfig.SubscriptionFlowHeaderKey, conf.SubscriptionConfig.StandardFlow)
		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(stdT, runtime.ID, runtime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, "", subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)

		stdT.Logf("Creating a subscription between consumer with subaccount id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, runtime.Name, runtime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(stdT, runtime.ID, runtime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, "", subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				stdT.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(stdT, err)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(stdT, err)
		require.Equal(stdT, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(stdT, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(stdT, func() bool {
			return subscription.GetSubscriptionJobStatus(stdT, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		stdT.Logf("Successfully created subscription between consumer with subaccount id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, runtime.Name, runtime.ID, subscriptionProviderSubaccountID)

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
		// With destinations - waiting for the synchronization job
		stdT.Log("Getting system with bundles and destinations - waiting for the synchronization job")
		require.Eventually(stdT, func() bool {
			respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/systemInstances?$expand=consumptionBundles($expand=destinations)&$format=json", headers)

			appsLen := len(gjson.Get(respBody, "value").Array())
			if appsLen != 1 {
				return false
			}
			require.Equal(stdT, 1, appsLen)

			appName := gjson.Get(respBody, "value.0.title").String()
			if appName != consumerApp.Name {
				return false
			}
			require.Equal(stdT, consumerApp.Name, appName)

			appDestinationsRaw := gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Raw
			if appDestinationsRaw == "" {
				return false
			}
			require.NotEmpty(stdT, appDestinationsRaw)

			appDestinations := gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Array()
			if len(appDestinations) != 1 {
				return false
			}
			require.Len(stdT, appDestinations, 1)

			appDestinationName := appDestinations[0].Get("sensitiveData.destinationConfiguration.Name").String()
			if appDestinationName != destination.Name {
				return false
			}
			require.Equal(stdT, destination.Name, appDestinationName)

			return true
		}, time.Second*30, time.Second)
		stdT.Log("Successfully fetched system with bundles and destinations while waiting for the synchronization job")

		// Create second destination
		destinationSecond := destination
		destinationSecond.Name = "test-second"
		client.CreateDestination(stdT, destinationSecond)
		defer client.DeleteDestination(stdT, destinationSecond.Name)

		// With destinations - reload
		stdT.Log("Getting system with bundles and destinations - reloading the destination")
		respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+
			"/systemInstances?$expand=consumptionBundles($expand=destinations)&$format=json&reload=true", headers)
		require.Equal(stdT, 1, len(gjson.Get(respBody, "value").Array()))
		require.Equal(stdT, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())
		require.NotEmpty(stdT, gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Raw)
		destinationsFromResponse := gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Array()
		require.Len(stdT, destinationsFromResponse, 2)
		require.ElementsMatch(stdT, []string{destination.Name, destinationSecond.Name}, []string{destinationsFromResponse[0].Get("sensitiveData.destinationConfiguration.Name").String(), destinationsFromResponse[1].Get("sensitiveData.destinationConfiguration.Name").String()})
		stdT.Log("Successfully fetched system with bundles and destinations")

		subscription.BuildAndExecuteUnsubscribeRequest(stdT, runtime.ID, runtime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, "", subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)

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

	t.Run("ConsumerProvider flow with application template as provider", func(stdT *testing.T) {
		// Create Application Template
		appTemplateName := createAppTemplateName("provider-app-template")
		appTemplateInput := fixAppTemplateInputWithDefaultDistinguishLabelAndSubdomainRegion(appTemplateName)
		for i := range appTemplateInput.Placeholders {
			appTemplateInput.Placeholders[i].JSONPath = str.Ptr(fmt.Sprintf("$.%s", conf.SubscriptionProviderAppNameProperty))
		}

		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(stdT, ctx, directorCertSecuredClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(stdT, ctx, directorCertSecuredClient, tenant.TestTenants.GetDefaultTenantID(), appTmpl)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, appTmpl.ID)
		require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, appTmpl.Labels[tenantfetcher.RegionKey])

		selfRegLabelValue, ok := appTmpl.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(stdT, ok)
		require.Contains(stdT, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+appTmpl.ID)

		selfRegDistinguishValue, ok := appTmpl.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, selfRegDistinguishValue)

		// Register application
		app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "testingApp", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), secondaryTenant)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, &app)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, app.ID)

		// Register application directly in subscriptionConsumerSubaccountID (that is in the formation) and validate that this system won't be visible for the SaaS app calling as part of the formation. That way we test the ORD filtering based on formations.
		subaccountApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "e2e-test-subaccount-app", subscriptionConsumerSubaccountID)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, &subaccountApp)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, subaccountApp.ID)

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

		formationTmplName := "e2e-test-formation-template-name"
		applicationType := util.ApplicationTypeC4C

		stdT.Logf("Creating formation template for the provider application tempalаte type %q with name %q", conf.SubscriptionProviderAppNameValue, formationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(stdT, ctx, certSecuredGraphQLClient, formationTmplName, conf.SubscriptionProviderAppNameValue, []string{string(applicationType), "SAP provider-app-template"}, graphql.ArtifactTypeSubscription)

		consumerFormationName := "consumer-test-scenario"
		stdT.Logf("Creating formation with name: %q from template with name: %q", consumerFormationName, formationTmplName)
		defer fixtures.DeleteFormationWithinTenant(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, consumerFormationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, consumerFormationName, &formationTmplName)
		require.NotEmpty(t, formation.ID)
		stdT.Logf("Successfully created formation: %s", consumerFormationName)

		stdT.Logf("Assign application to formation %s", consumerFormationName)
		assignToFormation(stdT, ctx, consumerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
		defer unassignFromFormation(stdT, ctx, consumerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned application to formation %s", consumerFormationName)

		stdT.Logf("Assign tenant %s to formation %s...", subscriptionConsumerSubaccountID, consumerFormationName)
		assignToFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", consumerFormationName, secondaryTenant)
		defer unassignFromFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", consumerFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned tenant %s to formation %s", subscriptionConsumerSubaccountID, consumerFormationName)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(stdT, err)
		response, err := httpClient.Do(depConfigureReq)
		defer func() {
			if err := response.Body.Close(); err != nil {
				stdT.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(stdT, err)
		require.Equal(stdT, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(stdT, err)
		subscriptionToken := token.GetClientCredentialsToken(stdT, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)
		subscribeReq.Header.Add(conf.SubscriptionConfig.SubscriptionFlowHeaderKey, conf.SubscriptionConfig.StandardFlow)
		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(stdT, appTmpl.ID, appTmpl.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, "", subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)

		stdT.Logf("Creating a subscription between consumer with subaccount id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, appTmpl.Name, appTmpl.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(stdT, appTmpl.ID, appTmpl.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, "", subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				stdT.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(stdT, err)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(stdT, err)
		require.Equal(stdT, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(stdT, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(stdT, func() bool {
			return subscription.GetSubscriptionJobStatus(stdT, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		stdT.Logf("Successfully created subscription between consumer with subaccount id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, appTmpl.Name, appTmpl.ID, subscriptionProviderSubaccountID)

		// Find the provider application ID
		stdT.Logf("List applications with tenant %q", secondaryTenant)
		appPage := fixtures.GetApplicationPageMinimal(stdT, ctx, certSecuredGraphQLClient, secondaryTenant)
		var providerApp graphql.Application
		for _, app := range appPage.Data {
			if app.Name == conf.SubscriptionProviderAppNameValue {
				providerApp = *app
				break
			}
		}

		// Assign the provider application to the formation
		stdT.Logf("Assign provider application with id %q to formation %s", providerApp.ID, consumerFormationName)
		assignToFormation(stdT, ctx, providerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
		defer unassignFromFormation(stdT, ctx, providerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned application to formation %s", consumerFormationName)

		// After successful subscription from above we call the director component with "double authentication(token + certificate)" in order to test claims validation is successful
		consumerToken := token.GetUserToken(stdT, ctx, conf.ConsumerTokenURL+conf.TokenPath, conf.ProviderClientID, conf.ProviderClientSecret, conf.BasicUsername, conf.BasicPassword, "subscriptionClaims")
		headers := map[string][]string{util.AuthorizationHeader: {fmt.Sprintf("Bearer %s", consumerToken)}}

		stdT.Log("Calling director to verify claims validation is successful...")
		getAppTmplReq := fixtures.FixApplicationTemplateRequest(appTmpl.ID)
		getAppTmplReq.Header = headers
		getAppTmpl := graphql.ApplicationTemplate{}

		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, getAppTmplReq, &getAppTmpl)
		require.NoError(stdT, err)
		require.Equal(stdT, appTmpl.ID, getAppTmpl.ID)
		require.Equal(stdT, appTmpl.Name, getAppTmpl.Name)
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
		respBody := makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+fmt.Sprintf("/systemInstances(%s)?$format=json", consumerApp.ID), headers)
		require.Equal(stdT, consumerApp.Name, gjson.Get(respBody, "title").String())
		stdT.Log("Successfully fetched consumer application using both provider and consumer credentials")

		// Make a request to the ORD service expanding bundles and destinations.
		// With destinations - waiting for the synchronization job
		stdT.Log("Getting system with bundles and destinations - waiting for the synchronization job")
		require.Eventually(stdT, func() bool {
			respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+fmt.Sprintf("/systemInstances(%s)?$expand=consumptionBundles($expand=destinations)&$format=json", consumerApp.ID), headers)

			appName := gjson.Get(respBody, "title").String()
			if appName != consumerApp.Name {
				return false
			}
			require.Equal(stdT, consumerApp.Name, appName)

			appDestinationsRaw := gjson.Get(respBody, "consumptionBundles.0.destinations").Raw
			if appDestinationsRaw == "" {
				return false
			}
			require.NotEmpty(stdT, appDestinationsRaw)

			appDestinations := gjson.Get(respBody, "consumptionBundles.0.destinations").Array()
			if len(appDestinations) != 1 {
				return false
			}
			require.Len(stdT, appDestinations, 1)

			appDestinationName := appDestinations[0].Get("sensitiveData.destinationConfiguration.Name").String()
			if appDestinationName != destination.Name {
				return false
			}
			require.Equal(stdT, destination.Name, appDestinationName)

			return true
		}, time.Second*30, time.Second)
		stdT.Log("Successfully fetched system with bundles and destinations while waiting for the synchronization job")

		// Create second destination
		destinationSecond := destination
		destinationSecond.Name = "test-second"
		client.CreateDestination(stdT, destinationSecond)
		defer client.DeleteDestination(stdT, destinationSecond.Name)

		// With destinations - reload
		respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+fmt.Sprintf("/systemInstances(%s)?$expand=consumptionBundles($expand=destinations)&$format=json&reload=true", consumerApp.ID), headers)
		require.Equal(stdT, consumerApp.Name, gjson.Get(respBody, "title").String())
		require.NotEmpty(stdT, gjson.Get(respBody, "consumptionBundles.0.destinations").Raw)
		destinationsFromResponse := gjson.Get(respBody, "consumptionBundles.0.destinations").Array()
		require.Len(stdT, destinationsFromResponse, 2)
		require.ElementsMatch(stdT, []string{destination.Name, destinationSecond.Name}, []string{destinationsFromResponse[0].Get("sensitiveData.destinationConfiguration.Name").String(), destinationsFromResponse[1].Get("sensitiveData.destinationConfiguration.Name").String()})
		stdT.Log("Successfully fetched system with bundles and destinations")

		subscription.BuildAndExecuteUnsubscribeRequest(stdT, appTmpl.ID, appTmpl.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, "", subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)

		stdT.Log("Validating no application is returned after successful unsubscription request...")
		respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)
		require.Empty(stdT, gjson.Get(respBody, "value").Array())
		stdT.Log("Successfully validated no application is returned after successful unsubscription request")

		stdT.Log("Validating no destination is returned after successful unsubscription request...")
		respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/destinations?$format=json", headers)
		require.Empty(stdT, gjson.Get(respBody, "value").Array())
		stdT.Log("Successfully validated no destination is returned after successful unsubscription request")

		stdT.Log("Validating director returns error during claims validation after unsubscribe request is successfully executed...")
		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, getAppTmplReq, &getAppTmpl)
		require.Error(stdT, err)
		require.Contains(stdT, err.Error(), fmt.Sprintf("Consumer's external tenant %s was not found as subscription record in the applications table for any application templates in the provider tenant", subscriptionConsumerSubaccountID))
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

		var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntimeWithoutTenant(stdT, ctx, directorCertSecuredClient, &runtime)
		runtime = fixtures.RegisterRuntimeFromInputWithoutTenant(stdT, ctx, directorCertSecuredClient, &runtimeInput)
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

		// Register application directly in subscriptionConsumerSubaccountID (that is in the formation) and validate that this system won't be visible for the SaaS app calling as part of the formation. That way we test the ORD filtering based on formations.
		subaccountApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "e2e-test-subaccount-app", subscriptionConsumerSubaccountID)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, &subaccountApp)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, subaccountApp.ID)

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

		formationTmplName := "e2e-test-formation-template-name"
		applicationType := util.ApplicationTypeC4C

		stdT.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, formationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(stdT, ctx, certSecuredGraphQLClient, formationTmplName, conf.SubscriptionProviderAppNameValue, []string{string(applicationType)}, graphql.ArtifactTypeSubscription)

		consumerFormationName := "consumer-test-scenario"
		stdT.Logf("Creating formation with name: %q from template with name: %q", consumerFormationName, formationTmplName)
		defer fixtures.DeleteFormationWithinTenant(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, consumerFormationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, consumerFormationName, &formationTmplName)
		require.NotEmpty(t, formation.ID)
		stdT.Logf("Successfully created formation: %s", consumerFormationName)

		stdT.Logf("Assign application to formation %s", consumerFormationName)
		assignToFormation(stdT, ctx, consumerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
		defer unassignFromFormation(stdT, ctx, consumerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned application to formation %s", consumerFormationName)

		stdT.Logf("Assign tenant %s to formation %s...", subscriptionConsumerSubaccountID, consumerFormationName)
		assignToFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", consumerFormationName, secondaryTenant)
		defer unassignFromFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", consumerFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned tenant %s to formation %s", subscriptionConsumerSubaccountID, consumerFormationName)

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
		defer func() {
			if err := response.Body.Close(); err != nil {
				stdT.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(stdT, err)
		require.Equal(stdT, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(stdT, err)
		subscriptionToken := token.GetClientCredentialsToken(stdT, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)
		subscribeReq.Header.Add(conf.SubscriptionConfig.SubscriptionFlowHeaderKey, conf.SubscriptionConfig.StandardFlow)
		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(stdT, runtime.ID, runtime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, "", subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)

		stdT.Logf("Creating a subscription between consumer with subaccount id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, runtime.Name, runtime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				stdT.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(stdT, err)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(stdT, err)
		require.Equal(stdT, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		defer subscription.BuildAndExecuteUnsubscribeRequest(stdT, runtime.ID, runtime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, "", subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(stdT, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(stdT, func() bool {
			return subscription.GetSubscriptionJobStatus(stdT, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		stdT.Logf("Successfully created subscription between consumer with subaccount id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, runtime.Name, runtime.ID, subscriptionProviderSubaccountID)

		// After successful subscription from above we call the director component with "double authentication(token + user_context header)" in order to test claims validation is successful
		consumerToken := token.GetUserToken(stdT, ctx, conf.ConsumerTokenURL+conf.TokenPath, conf.ProviderClientID, conf.ProviderClientSecret, conf.BasicUsername, conf.BasicPassword, "subscriptionClaims")
		consumerClaims := token.FlattenTokenClaims(stdT, consumerToken)
		consumerClaimsWithEncodedValue, err := sjson.Set(consumerClaims, "encodedValue", "test+n%C3%B8n+as%C3%A7ii+ch%C3%A5%C2%AEacte%C2%AE")
		require.NoError(t, err)
		headers := map[string][]string{subscription.UserContextHeader: {consumerClaimsWithEncodedValue}}

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
		// With destinations - waiting for the synchronization job
		stdT.Log("Getting system with bundles and destinations - waiting for the synchronization job")
		require.Eventually(stdT, func() bool {
			respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/systemInstances?$expand=consumptionBundles($expand=destinations)&$format=json", headers)

			appsLen := len(gjson.Get(respBody, "value").Array())
			if appsLen != 1 {
				return false
			}
			require.Equal(stdT, 1, appsLen)

			appName := gjson.Get(respBody, "value.0.title").String()
			if appName != consumerApp.Name {
				return false
			}
			require.Equal(stdT, consumerApp.Name, appName)

			appDestinationsRaw := gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Raw
			if appDestinationsRaw == "" {
				return false
			}
			require.NotEmpty(stdT, appDestinationsRaw)

			appDestinations := gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Array()
			if len(appDestinations) != 1 {
				return false
			}
			require.Len(stdT, appDestinations, 1)

			appDestinationName := appDestinations[0].Get("sensitiveData.destinationConfiguration.Name").String()
			if appDestinationName != destination.Name {
				return false
			}
			require.Equal(stdT, destination.Name, appDestinationName)

			return true
		}, time.Second*30, time.Second)
		stdT.Log("Successfully fetched system with bundles and destinations while waiting for the synchronization job")

		// Create second destination
		destinationSecond := destination
		destinationSecond.Name = "test-second"
		client.CreateDestination(stdT, destinationSecond)
		defer client.DeleteDestination(stdT, destinationSecond.Name)

		// With destinations - reload
		stdT.Log("Getting system with bundles and destinations - reloading the destination")
		respBody = makeRequestWithHeaders(stdT, certHttpClient, conf.ORDExternalCertSecuredServiceURL+
			"/systemInstances?$expand=consumptionBundles($expand=destinations)&$format=json&reload=true", headers)
		require.Equal(stdT, 1, len(gjson.Get(respBody, "value").Array()))
		require.Equal(stdT, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())
		require.NotEmpty(stdT, gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Raw)
		destinationsFromResponse := gjson.Get(respBody, "value.0.consumptionBundles.0.destinations").Array()
		require.Len(stdT, destinationsFromResponse, 2)
		require.ElementsMatch(stdT, []string{destination.Name, destinationSecond.Name}, []string{destinationsFromResponse[0].Get("sensitiveData.destinationConfiguration.Name").String(), destinationsFromResponse[1].Get("sensitiveData.destinationConfiguration.Name").String()})
		stdT.Log("Successfully fetched system with bundles and destinations")

		subscription.BuildAndExecuteUnsubscribeRequest(stdT, runtime.ID, runtime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, "", subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)

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
	t.Logf("Unassign object with type %s and id %s from formation %s", objectType, objectID, formationName)
	unassignReq := fixtures.FixUnassignFormationRequest(objectID, objectType, formationName)
	executeGQLRequest(t, ctx, unassignReq, formationName, tenantID)
	t.Logf("Successfully unassigned object with type %s and id %s from formation %s", objectType, objectID, formationName)
}

func executeGQLRequest(t *testing.T, ctx context.Context, gqlRequest *gcli.Request, formationName, tenantID string) {
	var formation graphql.Formation
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, gqlRequest, &formation)
	require.NoError(t, err)
	require.Equal(t, formationName, formation.Name)
}

func createAppTemplateName(name string) string {
	return fmt.Sprintf("SAP %s", name)
}

func fixAppTemplateInputWithDefaultDistinguishLabelAndSubdomainRegion(name string) graphql.ApplicationTemplateInput {
	input := fixtures.FixApplicationTemplate(name)
	input.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = conf.SubscriptionConfig.SelfRegDistinguishLabelValue
	input.ApplicationInput.BaseURL = str.Ptr(fmt.Sprintf(baseURLTemplate, "{{subdomain}}", "{{region}}"))
	input.Placeholders = append(input.Placeholders, &graphql.PlaceholderDefinitionInput{Name: "subdomain"}, &graphql.PlaceholderDefinitionInput{Name: "region"})
	return input
}
