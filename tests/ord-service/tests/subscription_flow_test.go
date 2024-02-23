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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	directordestinationcreator "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	esmdestinationcreator "github.com/kyma-incubator/compass/components/external-services-mock/pkg/destinationcreator"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/claims"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/tidwall/sjson"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
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

		// Register application directly in the tenant and don't add it to any formations and validate that this system won't be visible for the SaaS app calling as part of the formation.
		app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "testingApp", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), secondaryTenant)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, &app)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, app.ID)

		// Register application directly in subscriptionConsumerSubaccountID (that is in the formation) and validate that this system won't be visible for the SaaS app calling as part of the formation. That way we test the ORD filtering based on formations.
		subaccountApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "e2e-test-subaccount-app", subscriptionConsumerSubaccountID)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, &subaccountApp)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, subaccountApp.ID)

		// Register application directly in the tenant and add it to a formation that the caller is not discover consumer in and validate that this system won't be visible for the SaaS app calling as part of the formation.
		appInAnotherFormation, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "testingApp2", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), secondaryTenant)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, &appInAnotherFormation)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, appInAnotherFormation.ID)

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
		artifactKind := graphql.ArtifactTypeSubscription

		stdT.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, formationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplate(stdT, ctx, certSecuredGraphQLClient, graphql.FormationTemplateInput{
			Name:                   formationTmplName,
			ApplicationTypes:       []string{string(applicationType)},
			RuntimeTypes:           []string{conf.SubscriptionProviderAppNameValue},
			RuntimeTypeDisplayName: &formationTmplName,
			RuntimeArtifactKind:    &artifactKind,
			DiscoveryConsumers:     []string{string(applicationType), conf.SubscriptionProviderAppNameValue},
		})

		secondFormationTmplName := "e2e-test-formation-template-without-discovery-consumer"
		stdT.Logf("Creating formation template with name %q, without discovery consumers configured", secondFormationTmplName)
		var ftWithoutDiscoveryConsumers graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ftWithoutDiscoveryConsumers)
		ftWithoutDiscoveryConsumers = fixtures.CreateFormationTemplate(stdT, ctx, certSecuredGraphQLClient, graphql.FormationTemplateInput{
			Name:                   secondFormationTmplName,
			ApplicationTypes:       []string{string(applicationType)},
			RuntimeTypes:           []string{conf.SubscriptionProviderAppNameValue},
			RuntimeTypeDisplayName: &secondFormationTmplName,
			RuntimeArtifactKind:    &artifactKind,
		})

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

		noDiscoveryConsumptionFormationName := "no-discovery-consumption-scenario"
		stdT.Logf("Creating formation with name: %q from template with name: %q", noDiscoveryConsumptionFormationName, secondFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, noDiscoveryConsumptionFormationName)
		noDiscoveryConsumptionFormation := fixtures.CreateFormationFromTemplateWithinTenant(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, noDiscoveryConsumptionFormationName, &secondFormationTmplName)
		require.NotEmpty(t, noDiscoveryConsumptionFormation.ID)
		stdT.Logf("Successfully created formation: %s", noDiscoveryConsumptionFormationName)

		stdT.Logf("Assign application to formation %s", noDiscoveryConsumptionFormationName)
		assignToFormation(stdT, ctx, appInAnotherFormation.ID, "APPLICATION", noDiscoveryConsumptionFormationName, secondaryTenant)
		defer unassignFromFormation(stdT, ctx, appInAnotherFormation.ID, "APPLICATION", noDiscoveryConsumptionFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned application to formation %s", noDiscoveryConsumptionFormationName)

		stdT.Logf("Assign tenant %s to formation %s...", subscriptionConsumerSubaccountID, noDiscoveryConsumptionFormationName)
		assignToFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", noDiscoveryConsumptionFormationName, secondaryTenant)
		defer unassignFromFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", noDiscoveryConsumptionFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned tenant %s to formation %s", subscriptionConsumerSubaccountID, noDiscoveryConsumptionFormationName)

		selfRegLabelValue, ok := runtime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(stdT, ok)
		require.Contains(stdT, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+runtime.ID)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		deps, err := json.Marshal([]string{selfRegLabelValue})
		require.NoError(stdT, err)
		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer(deps))
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
		subscriptionToken := token.GetClientCredentialsToken(stdT, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, claims.TenantFetcherClaimKey)
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
		body, err := io.ReadAll(resp.Body)
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
		consumerToken := token.GetUserToken(stdT, ctx, conf.ConsumerTokenURL+conf.TokenPath, conf.ProviderClientID, conf.ProviderClientSecret, conf.BasicUsername, conf.BasicPassword, claims.SubscriptionClaimKey)
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

		subdomain := conf.DestinationConsumerSubdomainMtls
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
		appTemplateName := fixtures.CreateAppTemplateName("provider-app-template")
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

		// Register application directly in the tenant and don't add it to any formations and validate that this system won't be visible for the SaaS app calling as part of the formation.
		app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "testingApp", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), secondaryTenant)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, &app)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, app.ID)

		// Register application directly in subscriptionConsumerSubaccountID (that is in the formation) and validate that this system won't be visible for the SaaS app calling as part of the formation. That way we test the ORD filtering based on formations.
		subaccountApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "e2e-test-subaccount-app", subscriptionConsumerSubaccountID)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, &subaccountApp)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, subaccountApp.ID)

		// Register application directly in the tenant and add it to a formation that the caller is not discover consumer in and validate that this system won't be visible for the SaaS app calling as part of the formation.
		appInAnotherFormation, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "testingApp2", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), secondaryTenant)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, &appInAnotherFormation)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, appInAnotherFormation.ID)

		// Register consumer application
		const localTenantID = "localTenantID"
		consumerApp, err := fixtures.RegisterApplicationWithTypeAndLocalTenantID(t, ctx, certSecuredGraphQLClient, "consumerApp", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), localTenantID, secondaryTenant)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, &consumerApp)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, consumerApp.ID)
		require.NotEmpty(stdT, consumerApp.Name)

		const correlationID = "correlationID"
		const correlationIDSecond = "correlationID-second"
		bndlInput1 := graphql.BundleCreateInput{
			Name:           "test-bundle-1",
			CorrelationIDs: []string{correlationID},
		}
		bndlInput2 := graphql.BundleCreateInput{
			Name:           "test-bundle-2",
			CorrelationIDs: []string{correlationIDSecond},
		}

		bundle1 := fixtures.CreateBundleWithInput(t, ctx, certSecuredGraphQLClient, secondaryTenant, consumerApp.ID, bndlInput1)
		require.NotEmpty(stdT, bundle1.ID)
		bundle2 := fixtures.CreateBundleWithInput(t, ctx, certSecuredGraphQLClient, secondaryTenant, consumerApp.ID, bndlInput2)
		require.NotEmpty(stdT, bundle2.ID)

		formationTmplName := "e2e-test-formation-template-name"
		applicationType := util.ApplicationTypeC4C
		artifactKind := graphql.ArtifactTypeSubscription

		stdT.Logf("Creating formation template for the provider application tempalаte type %q with name %q", conf.SubscriptionProviderAppNameValue, formationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplate(stdT, ctx, certSecuredGraphQLClient, graphql.FormationTemplateInput{
			Name:                   formationTmplName,
			ApplicationTypes:       []string{string(applicationType), "SAP provider-app-template"},
			RuntimeTypes:           []string{conf.SubscriptionProviderAppNameValue},
			RuntimeTypeDisplayName: &formationTmplName,
			RuntimeArtifactKind:    &artifactKind,
			DiscoveryConsumers:     []string{string(applicationType), "SAP provider-app-template", conf.SubscriptionProviderAppNameValue},
		})

		secondFormationTmplName := "e2e-test-formation-template-without-discovery-consumer"
		stdT.Logf("Creating formation template with name %q, without discovery consumers configured", secondFormationTmplName)
		var ftWithoutDiscoveryConsumers graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ftWithoutDiscoveryConsumers)
		ftWithoutDiscoveryConsumers = fixtures.CreateFormationTemplate(stdT, ctx, certSecuredGraphQLClient, graphql.FormationTemplateInput{
			Name:                   secondFormationTmplName,
			ApplicationTypes:       []string{string(applicationType), "SAP provider-app-template"},
			RuntimeTypes:           []string{conf.SubscriptionProviderAppNameValue},
			RuntimeTypeDisplayName: &secondFormationTmplName,
			RuntimeArtifactKind:    &artifactKind,
		})

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

		noDiscoveryConsumptionFormationName := "no-discovery-consumption-scenario"
		stdT.Logf("Creating formation with name: %q from template with name: %q", noDiscoveryConsumptionFormationName, secondFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, noDiscoveryConsumptionFormationName)
		noDiscoveryConsumptionFormation := fixtures.CreateFormationFromTemplateWithinTenant(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, noDiscoveryConsumptionFormationName, &secondFormationTmplName)
		require.NotEmpty(t, noDiscoveryConsumptionFormation.ID)
		stdT.Logf("Successfully created formation: %s", noDiscoveryConsumptionFormationName)

		stdT.Logf("Assign application to formation %s", noDiscoveryConsumptionFormationName)
		assignToFormation(stdT, ctx, appInAnotherFormation.ID, "APPLICATION", noDiscoveryConsumptionFormationName, secondaryTenant)
		defer unassignFromFormation(stdT, ctx, appInAnotherFormation.ID, "APPLICATION", noDiscoveryConsumptionFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned application to formation %s", noDiscoveryConsumptionFormationName)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		deps, err := json.Marshal([]string{selfRegLabelValue})
		require.NoError(stdT, err)
		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer(deps))
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
		subscriptionToken := token.GetClientCredentialsToken(stdT, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, claims.TenantFetcherClaimKey)
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
		body, err := io.ReadAll(resp.Body)
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

		stdT.Logf("Assign provider application with id %q to formation %s", providerApp.ID, noDiscoveryConsumptionFormationName)
		assignToFormation(stdT, ctx, providerApp.ID, "APPLICATION", noDiscoveryConsumptionFormationName, secondaryTenant)
		defer unassignFromFormation(stdT, ctx, providerApp.ID, "APPLICATION", noDiscoveryConsumptionFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned application to formation %s", noDiscoveryConsumptionFormationName)

		// After successful subscription from above we call the director component with "double authentication(token + certificate)" in order to test claims validation is successful
		consumerToken := token.GetUserToken(stdT, ctx, conf.ConsumerTokenURL+conf.TokenPath, conf.ProviderClientID, conf.ProviderClientSecret, conf.BasicUsername, conf.BasicPassword, claims.SubscriptionClaimKey)
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

		subdomain := conf.DestinationConsumerSubdomainMtls
		client, err := clients.NewDestinationClient(instance, conf.DestinationAPIConfig, subdomain)
		require.NoError(stdT, err)

		destination := clients.Destination{
			Name:            "test",
			Type:            "HTTP",
			URL:             "http://localhost",
			Authentication:  "BasicAuthentication",
			XCorrelationID:  fmt.Sprintf("%s,%s-new", correlationID, correlationID),
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

			appDestinations1Raw := gjson.Get(respBody, "consumptionBundles.0.destinations").Raw
			if appDestinations1Raw == "" {
				return false
			}
			require.NotEmpty(stdT, appDestinations1Raw)

			appDestinations2Raw := gjson.Get(respBody, "consumptionBundles.1.destinations").Raw
			if appDestinations2Raw == "" {
				return false
			}
			require.NotEmpty(stdT, appDestinations2Raw)

			appDestinationsForBundle1 := gjson.Get(respBody, "consumptionBundles.0.destinations").Array()
			if len(appDestinationsForBundle1) != 1 {
				return false
			}
			require.Len(stdT, appDestinationsForBundle1, 1)

			appDestinationsForBundle2 := gjson.Get(respBody, "consumptionBundles.1.destinations").Array()
			if len(appDestinationsForBundle2) != 1 {
				return false
			}
			require.Len(stdT, appDestinationsForBundle2, 1)

			appDestination1Name := appDestinationsForBundle1[0].Get("sensitiveData.destinationConfiguration.Name").String()
			if appDestination1Name != destination.Name {
				return false
			}
			require.Equal(stdT, destination.Name, appDestination1Name)

			appDestination2Name := appDestinationsForBundle2[0].Get("sensitiveData.destinationConfiguration.Name").String()
			if appDestination2Name != destination.Name {
				return false
			}
			require.Equal(stdT, destination.Name, appDestination2Name)

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

		unassignFromFormation(stdT, ctx, providerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
		unassignFromFormation(stdT, ctx, providerApp.ID, "APPLICATION", noDiscoveryConsumptionFormationName, secondaryTenant)
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

		// Register application directly in the tenant and don't add it to any formations and validate that this system won't be visible for the SaaS app calling as part of the formation.
		app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "testingApp", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), secondaryTenant)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, &app)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, app.ID)

		// Register application directly in subscriptionConsumerSubaccountID (that is in the formation) and validate that this system won't be visible for the SaaS app calling as part of the formation. That way we test the ORD filtering based on formations.
		subaccountApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "e2e-test-subaccount-app", subscriptionConsumerSubaccountID)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, &subaccountApp)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, subaccountApp.ID)

		// Register application directly in the tenant and add it to a formation that the caller is not discover consumer in and validate that this system won't be visible for the SaaS app calling as part of the formation.
		appInAnotherFormation, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "testingApp2", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), secondaryTenant)
		defer fixtures.CleanupApplication(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, &appInAnotherFormation)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, appInAnotherFormation.ID)

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
		artifactKind := graphql.ArtifactTypeSubscription

		stdT.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, formationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplate(stdT, ctx, certSecuredGraphQLClient, graphql.FormationTemplateInput{
			Name:                   formationTmplName,
			ApplicationTypes:       []string{string(applicationType)},
			RuntimeTypes:           []string{conf.SubscriptionProviderAppNameValue},
			RuntimeTypeDisplayName: &formationTmplName,
			RuntimeArtifactKind:    &artifactKind,
			DiscoveryConsumers:     []string{string(applicationType), conf.SubscriptionProviderAppNameValue},
		})

		secondFormationTmplName := "e2e-test-formation-template-without-discovery-consumer"
		stdT.Logf("Creating formation template with name %q, without discovery consumers configured", secondFormationTmplName)
		var ftWithoutDiscoveryConsumers graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ftWithoutDiscoveryConsumers)
		ftWithoutDiscoveryConsumers = fixtures.CreateFormationTemplate(stdT, ctx, certSecuredGraphQLClient, graphql.FormationTemplateInput{
			Name:                   secondFormationTmplName,
			ApplicationTypes:       []string{string(applicationType)},
			RuntimeTypes:           []string{conf.SubscriptionProviderAppNameValue},
			RuntimeTypeDisplayName: &secondFormationTmplName,
			RuntimeArtifactKind:    &artifactKind,
		})

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

		noDiscoveryConsumptionFormationName := "no-discovery-consumption-scenario"
		stdT.Logf("Creating formation with name: %q from template with name: %q", noDiscoveryConsumptionFormationName, secondFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, noDiscoveryConsumptionFormationName)
		noDiscoveryConsumptionFormation := fixtures.CreateFormationFromTemplateWithinTenant(stdT, ctx, certSecuredGraphQLClient, secondaryTenant, noDiscoveryConsumptionFormationName, &secondFormationTmplName)
		require.NotEmpty(t, noDiscoveryConsumptionFormation.ID)
		stdT.Logf("Successfully created formation: %s", noDiscoveryConsumptionFormationName)

		stdT.Logf("Assign application to formation %s", noDiscoveryConsumptionFormationName)
		assignToFormation(stdT, ctx, appInAnotherFormation.ID, "APPLICATION", noDiscoveryConsumptionFormationName, secondaryTenant)
		defer unassignFromFormation(stdT, ctx, appInAnotherFormation.ID, "APPLICATION", noDiscoveryConsumptionFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned application to formation %s", noDiscoveryConsumptionFormationName)

		stdT.Logf("Assign tenant %s to formation %s...", subscriptionConsumerSubaccountID, noDiscoveryConsumptionFormationName)
		assignToFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", noDiscoveryConsumptionFormationName, secondaryTenant)
		defer unassignFromFormation(stdT, ctx, subscriptionConsumerSubaccountID, "TENANT", noDiscoveryConsumptionFormationName, secondaryTenant)
		stdT.Logf("Successfully assigned tenant %s to formation %s", subscriptionConsumerSubaccountID, noDiscoveryConsumptionFormationName)

		selfRegLabelValue, ok := runtime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(stdT, ok)
		require.Contains(stdT, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+runtime.ID)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		deps, err := json.Marshal([]string{selfRegLabelValue})
		require.NoError(stdT, err)
		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer(deps))
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
		subscriptionToken := token.GetClientCredentialsToken(stdT, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, claims.TenantFetcherClaimKey)
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
		body, err := io.ReadAll(resp.Body)
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
		consumerToken := token.GetUserToken(stdT, ctx, conf.ConsumerTokenURL+conf.TokenPath, conf.ProviderClientID, conf.ProviderClientSecret, conf.BasicUsername, conf.BasicPassword, claims.SubscriptionClaimKey)
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

		subdomain := conf.DestinationConsumerSubdomainMtls
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

func TestDestinationFetcher(stdT *testing.T) {
	t := testingx.NewT(stdT)
	certSecuredHTTPClient := fixtures.FixCertSecuredHTTPClient(certCache, conf.ExternalClientCertSecretName, conf.SkipSSLValidation)

	subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID
	subscriptionConsumerTenantID := conf.TestConsumerTenantID
	subscriptionProviderSubaccountID := conf.TestProviderSubaccountID
	subscriptionConsumerAccountID := conf.TestConsumerAccountID
	assignOperation := "assign"
	certSubject := strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.ExternalCertProviderConfig.TestExternalCertCN, "app-template-subscription-cn", -1)
	consumerType := "Integration System"                // should be a valid consumer type
	tenantAccessLevels := []string{"account", "global"} // should be a valid tenant access level

	apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)

	ctx := context.Background()

	// We need an externally issued cert with a subject that is not part of the access level mappings
	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestSecretName:         conf.CertSvcInstanceTestSecretName,
		ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestExternalCertSubject:               certSubject,
		ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
		ExternalCertProvider:                  certprovider.CertificateService,
	}

	// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(stdT, ctx, externalCertProviderConfig, false)
	appProviderDirectorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

	// The external cert secret created by the NewExternalCertFromConfig above is used by the external-services-mock for the async formation status API call,
	// that's why in the function above there is a false parameter that don't delete it and an explicit defer deletion func is added here
	// so, the secret could be deleted at the end of the test. Otherwise, it will remain as leftover resource in the cluster
	defer func() {
		k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
		require.NoError(t, err)
		k8s.DeleteSecret(stdT, ctx, k8sClient, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace)
	}()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
		},
	}

	t.Run("Associate a destination with bundles using an existing destination with formation assignment reference created by destination creator", func(t *testing.T) {
		// Prepare resources that are needed so we end up with 2 application templates and 2 applications for them, one bundle for the first application, a formation template, a formation with the two systems
		// Create Application Template
		namePlaceholder := "name"
		displayNamePlaceholder := "display-name"
		appRegion := "test-app-region"
		appNamespace := "compass.test"
		localTenantID := "local-tenant-id"
		localTenantID2 := "local-tenant-id2"
		app1BaseURL := "http://e2e-test-app1-base-url"
		app2BaseURL := "http://e2e-test-app2-base-url"

		applicationType1 := "app-subscription-type-1"
		applicationType2 := "app-subscription-type-2"
		providerFormationTmplName := "app-subscription-template-name"
		formationName := "app-subscription-formation-name"

		appTemplateInput := fixtures.FixApplicationTemplateWithoutWebhook(applicationType1, localTenantID, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder)
		appTemplateInput.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = conf.SubscriptionConfig.SelfRegDistinguishLabelValue
		appTemplateInput.ApplicationInput.Labels[conf.GlobalSubaccountIDLabelKey] = subscriptionConsumerSubaccountID
		appTemplateInput.ApplicationInput.BaseURL = &app1BaseURL
		appTemplateInput.ApplicationInput.LocalTenantID = nil
		for i := range appTemplateInput.Placeholders {
			appTemplateInput.Placeholders[i].JSONPath = str.Ptr(fmt.Sprintf("$.%s", conf.SubscriptionProviderAppNameProperty))
		}

		t.Logf("Create application template for type %q", applicationType1)
		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(stdT, ctx, appProviderDirectorCertSecuredClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(stdT, ctx, appProviderDirectorCertSecuredClient, tenant.TestTenants.GetDefaultTenantID(), appTmpl)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, appTmpl.ID)
		require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, appTmpl.Labels[tenantfetcher.RegionKey])

		selfRegLabelValue, ok := appTmpl.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(stdT, ok)
		require.Contains(stdT, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+appTmpl.ID)

		deps, err := json.Marshal([]string{selfRegLabelValue, conf.ProviderDestinationConfig.Dependency})
		require.NoError(stdT, err)
		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer(deps))
		require.NoError(stdT, err)
		response, err := httpClient.Do(depConfigureReq)
		defer func() {
			if err := response.Body.Close(); err != nil {
				stdT.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(stdT, err)
		require.Equal(stdT, http.StatusOK, response.StatusCode)

		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, claims.TenantFetcherClaimKey)

		defer subscription.BuildAndExecuteUnsubscribeRequest(t, appTmpl.ID, appTmpl.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)
		subscription.CreateSubscription(t, conf.SubscriptionConfig, httpClient, appTmpl, apiPath, subscriptionToken, subscriptionConsumerTenantID, subscriptionConsumerSubaccountID, subscriptionProviderSubaccountID, conf.SubscriptionProviderAppNameValue, true, true, conf.SubscriptionConfig.StandardFlow)

		actualAppPage := graphql.ApplicationPage{}
		getSrcAppReq := fixtures.FixGetApplicationsRequestWithPagination()
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, getSrcAppReq, &actualAppPage)
		require.NoError(t, err)

		require.Len(t, actualAppPage.Data, 1)
		require.Equal(t, appTmpl.ID, *actualAppPage.Data[0].ApplicationTemplateID)
		app1 := *actualAppPage.Data[0]
		t.Logf("app1 ID: %q", app1.ID)

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, "app-subscription-notifications")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		t.Logf("Create application template for type %q", applicationType2)
		appTemplateInput = fixtures.FixApplicationTemplateWithoutWebhook(applicationType2, localTenantID2, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder)
		appTemplateInput.ApplicationInput.Labels[conf.GlobalSubaccountIDLabelKey] = subscriptionConsumerSubaccountID
		appTemplateInput.ApplicationInput.BaseURL = &app2BaseURL
		appTmpl2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl2)
		require.NoError(t, err)

		// Create certificate subject mapping with custom subject that was used to create a certificate for the graphql client above
		internalConsumerID := appTmpl2.ID // add application templated ID as certificate subject mapping internal consumer to satisfy the authorization checks in the formation assignment status API
		certSubjectMappingCustomSubjectWithCommaSeparator := strings.ReplaceAll(strings.TrimLeft(certSubject, "/"), "/", ",")
		csmInput := fixtures.FixCertificateSubjectMappingInput(certSubjectMappingCustomSubjectWithCommaSeparator, consumerType, &internalConsumerID, tenantAccessLevels)
		t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", certSubjectMappingCustomSubjectWithCommaSeparator, consumerType, tenantAccessLevels)

		var csmCreate graphql.CertificateSubjectMapping // needed so the 'defer' can be above the cert subject mapping creation
		defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreate)
		csmCreate = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput)

		t.Logf("Sleeping for %s, so the hydrator component could update the certificate subject mapping cache with the new data", conf.CertSubjectMappingResyncInterval.String())
		time.Sleep(conf.CertSubjectMappingResyncInterval)

		t.Logf("Create application 2 from template %q", applicationType2)
		appFromTmplSrc2 := fixtures.FixApplicationFromTemplateInput(applicationType2, namePlaceholder, "app2-subscription-description-tests", displayNamePlaceholder, "App 2 Display Name")
		appFromTmplSrc2GQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc2)
		require.NoError(t, err)
		createAppFromTmplSecondRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrc2GQL)
		app2 := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, createAppFromTmplSecondRequest, &app2)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &app2)
		require.NoError(t, err)
		require.NotEmpty(t, app2.ID)
		t.Logf("app2 ID: %q", app2.ID)

		t.Logf("Create bundle for application with ID: %q", app1.ID)
		const correlationID = "correlationID"
		bndlInput := graphql.BundleCreateInput{
			Name:           "test-bundle-1",
			CorrelationIDs: []string{correlationID},
		}
		bundle := fixtures.CreateBundleWithInput(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, app1.ID, bndlInput)
		require.NotEmpty(stdT, bundle.ID)

		t.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, providerFormationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(stdT, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(stdT, ctx, certSecuredGraphQLClient, providerFormationTmplName, conf.SubscriptionProviderAppNameValue, []string{applicationType1, applicationType2}, graphql.ArtifactTypeSubscription)

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		cleanupDestinationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupDestinationsFromExternalSvcMock(t, certSecuredHTTPClient)

		cleanupDestnationCertificatesFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupDestnationCertificatesFromExternalSvcMock(t, certSecuredHTTPClient)

		// first app
		applicationTntMappingWebhookType := graphql.WebhookTypeApplicationTenantMapping
		asyncCallbackWebhookMode := graphql.WebhookModeAsyncCallback
		urlTemplateAsyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async/destinations/{{.TargetApplication.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplateApplication := "{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}"
		outputTemplateAsyncApplication := "{\\\"config\\\":\\\"{{.Body.configuration}}\\\",\\\"state\\\":\\\"{{.Body.state}}\\\",\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\":\\\"{{.Body.error}}\\\",\\\"success_status_code\\\":202}"

		applicationAsyncWebhookInput := fixtures.FixFormationNotificationWebhookInput(applicationTntMappingWebhookType, asyncCallbackWebhookMode, urlTemplateAsyncApplication, inputTemplateApplication, outputTemplateAsyncApplication)

		t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", applicationTntMappingWebhookType, asyncCallbackWebhookMode, app2.ID)
		actualApplicationAsyncWebhookInput := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationAsyncWebhookInput, subscriptionConsumerAccountID, app2.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApplicationAsyncWebhookInput.ID)

		// second app
		syncWebhookMode := graphql.WebhookModeSync
		urlTemplateSyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/destinations/configuration/{{.TargetApplication.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		outputTemplateSyncApplication := "{\\\"config\\\":\\\"{{.Body.configuration}}\\\",\\\"state\\\":\\\"{{.Body.state}}\\\",\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\":200}"

		applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(applicationTntMappingWebhookType, syncWebhookMode, urlTemplateSyncApplication, inputTemplateApplication, outputTemplateSyncApplication)

		t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", applicationTntMappingWebhookType, syncWebhookMode, app1.ID)
		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, subscriptionConsumerAccountID, app1.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApplicationWebhook.ID)

		// create formation constraints and attach them to formation template
		firstConstraintInput := graphql.FormationConstraintInput{
			Name:            "e2e-destination-creator-notification-status-returned",
			ConstraintType:  graphql.ConstraintTypePre,
			TargetOperation: graphql.TargetOperationNotificationStatusReturned,
			Operator:        formationconstraintpkg.DestinationCreator,
			ResourceType:    graphql.ResourceTypeApplication,
			ResourceSubtype: applicationType1,
			InputTemplate:   "{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",{{ if .NotificationStatusReport }}\\\"notification_status_report_memory_address\\\":{{ .NotificationStatusReport.GetAddress }},{{ end }}{{ if .FormationAssignment }}\\\"formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}",
			ConstraintScope: graphql.ConstraintScopeFormationType,
		}

		t.Logf("Creating formation constraint with name: %q", firstConstraintInput.Name)
		firstConstraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, firstConstraintInput)
		defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, firstConstraint.ID)
		require.NotEmpty(t, firstConstraint.ID)

		fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, firstConstraint.ID, firstConstraint.Name, ft.ID, ft.Name)

		// second constraint
		secondConstraintInput := graphql.FormationConstraintInput{
			Name:            "e2e-destination-creator-send-notification",
			ConstraintType:  graphql.ConstraintTypePre,
			TargetOperation: graphql.TargetOperationSendNotification,
			Operator:        formationconstraintpkg.DestinationCreator,
			ResourceType:    graphql.ResourceTypeApplication,
			ResourceSubtype: applicationType1,
			InputTemplate:   "{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",{{ if .FormationAssignment }}\\\"formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}",
			ConstraintScope: graphql.ConstraintScopeFormationType,
		}

		t.Logf("Creating formation constraint with name: %q", secondConstraintInput.Name)
		secondConstraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, secondConstraintInput)
		defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, secondConstraint.ID)
		require.NotEmpty(t, secondConstraint.ID)

		fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, secondConstraint.ID, secondConstraint.Name, ft.ID, ft.Name)

		// create formation
		t.Logf("Creating formation with name: %q from template with name: %q", formationName, providerFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName, &providerFormationTmplName)
		require.NotEmpty(t, formation.ID)

		formationInput := graphql.FormationInput{Name: formationName}
		t.Logf("Assign application 2 with ID: %s to formation: %q", app2.ID, formationName)
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app2.ID, subscriptionConsumerAccountID)
		assignedFormation := fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app2.ID, subscriptionConsumerAccountID)
		require.Equal(t, formation.ID, assignedFormation.ID)
		require.Equal(t, formation.State, assignedFormation.State)

		t.Logf("Assign application 1 with ID: %s to formation %s", app1.ID, formationName)
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app1.ID, subscriptionConsumerAccountID)
		assignedFormation = fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app1.ID, subscriptionConsumerAccountID)
		require.Equal(t, formationName, assignedFormation.Name)

		assignmentWithDestDetails := fixtures.GetFormationAssignmentsBySourceAndTarget(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formation.ID, app2.ID, app1.ID)
		require.NotEmpty(t, assignmentWithDestDetails)

		t.Logf("Assert formation assignments during %s operation...", assignOperation)
		noAuthDestinationName := "e2e-design-time-destination-name"
		noAuthDestinationURL := "http://e2e-design-time-url-example.com"

		// Configure destination service client
		region := conf.SubscriptionConfig.SelfRegRegion
		instance, ok := conf.DestinationsConfig.RegionToInstanceConfig[region]
		require.True(t, ok)
		destinationClient, err := clients.NewDestinationClient(instance, conf.DestinationAPIConfig, conf.DestinationConsumerSubdomainMtls)
		require.NoError(t, err)

		destinationProviderToken := token.GetClientCredentialsToken(t, ctx, conf.ProviderDestinationConfig.TokenURL+conf.ProviderDestinationConfig.TokenPath, conf.ProviderDestinationConfig.ClientID, conf.ProviderDestinationConfig.ClientSecret, claims.DestinationProviderClaimKey)

		t.Log("Assert destinations and destination certificates are created...")
		assertNoAuthDestination(t, destinationClient, conf.ProviderDestinationConfig.ServiceURL, noAuthDestinationName, noAuthDestinationURL, "", conf.TestProviderSubaccountID, destinationProviderToken)
		t.Log("Destinations and destination certificates have been successfully created")

		// Create destination that matches to the created bundle
		subdomain := conf.DestinationConsumerSubdomainMtls
		client, err := clients.NewDestinationClient(instance, conf.DestinationAPIConfig, subdomain)
		require.NoError(stdT, err)

		destination := clients.Destination{
			Name:           "e2e-client-cert-auth-destination-name",
			Type:           "HTTP",
			URL:            "http://e2e-client-cert-auth-url-example.com",
			Authentication: "ClientCertificateAuthentication",
			XCorrelationID: correlationID,
		}

		client.CreateDestination(stdT, destination)
		defer client.DeleteDestination(stdT, destination.Name)

		stdT.Log("Getting consumer application using both provider and consumer credentials...")

		cfgWithInternalVisibilityScope := &clientcredentials.Config{
			ClientID:     intSysOauthCredentialData.ClientID,
			ClientSecret: intSysOauthCredentialData.ClientSecret,
			TokenURL:     intSysOauthCredentialData.URL,
			Scopes:       []string{"internal_visibility:read"},
		}
		unsecuredHttpClient := http.DefaultClient
		unsecuredHttpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, unsecuredHttpClient)
		httpClient1 := cfgWithInternalVisibilityScope.Client(ctx)
		httpClient1.Timeout = 20 * time.Second

		// Make a request to the ORD service with http client.
		respBody := makeRequestWithHeaders(stdT, httpClient1, conf.ORDServiceURL+fmt.Sprintf("/systemInstances(%s)?$format=json&reload=true", app1.ID), map[string][]string{tenantHeader: {conf.TestConsumerSubaccountID}})

		require.Equal(stdT, app1.Name, gjson.Get(respBody, "title").String())
		stdT.Log("Successfully fetched consumer application using both provider and consumer credentials")

		// Make a request to the ORD service expanding bundles and destinations.
		// With destinations - waiting for the synchronization job
		stdT.Log("Getting system with bundles and destinations - waiting for the synchronization job")
		require.Eventually(stdT, func() bool {
			respBody = makeRequestWithHeaders(stdT, httpClient1, conf.ORDServiceURL+fmt.Sprintf("/systemInstances(%s)?$expand=consumptionBundles($expand=destinations)&$format=json&reload=true", app1.ID), map[string][]string{tenantHeader: {conf.TestConsumerSubaccountID}})

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

			return true
		}, time.Second*30, time.Second)
		stdT.Log("Successfully fetched system with bundles and destinations while waiting for the synchronization job")

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	})
}

func cleanupNotificationsFromExternalSvcMock(t *testing.T, client *http.Client) {
	req, err := http.NewRequest(http.MethodDelete, conf.ExternalServicesMockMtlsSecuredURL+"/formation-callback/cleanup", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func cleanupDestinationsFromExternalSvcMock(t *testing.T, client *http.Client) {
	req, err := http.NewRequest(http.MethodDelete, conf.ExternalServicesMockMtlsSecuredURL+"/destinations/cleanup", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func cleanupDestnationCertificatesFromExternalSvcMock(t *testing.T, client *http.Client) {
	req, err := http.NewRequest(http.MethodDelete, conf.ExternalServicesMockMtlsSecuredURL+"/destination-certificates/cleanup", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
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

func fixAppTemplateInputWithDefaultDistinguishLabelAndSubdomainRegion(name string) graphql.ApplicationTemplateInput {
	input := fixtures.FixApplicationTemplate(name)
	input.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = conf.SubscriptionConfig.SelfRegDistinguishLabelValue
	input.ApplicationInput.BaseURL = str.Ptr(fmt.Sprintf(baseURLTemplate, "{{subdomain}}", "{{region}}"))
	input.Placeholders = append(input.Placeholders, &graphql.PlaceholderDefinitionInput{Name: "subdomain"}, &graphql.PlaceholderDefinitionInput{Name: "region"})
	return input
}

func assertNoAuthDestination(t *testing.T, client *clients.DestinationClient, serviceURL, noAuthDestinationName, noAuthDestinationURL, instanceID, ownerSubaccountID, authToken string) {
	noAuthDestBytes := client.FindDestinationByName(t, serviceURL, noAuthDestinationName, authToken, "", http.StatusOK)
	var noAuthDest esmdestinationcreator.DestinationSvcNoAuthenticationDestResponse
	err := json.Unmarshal(noAuthDestBytes, &noAuthDest)
	require.NoError(t, err)
	require.Equal(t, ownerSubaccountID, noAuthDest.Owner.SubaccountID)
	require.Equal(t, instanceID, noAuthDest.Owner.InstanceID)
	require.Equal(t, noAuthDestinationName, noAuthDest.DestinationConfiguration.Name)
	require.Equal(t, directordestinationcreator.TypeHTTP, noAuthDest.DestinationConfiguration.Type)
	require.Equal(t, noAuthDestinationURL, noAuthDest.DestinationConfiguration.URL)
	require.Equal(t, directordestinationcreator.AuthTypeNoAuth, noAuthDest.DestinationConfiguration.Authentication)
	require.Equal(t, directordestinationcreator.ProxyTypeInternet, noAuthDest.DestinationConfiguration.ProxyType)
}
