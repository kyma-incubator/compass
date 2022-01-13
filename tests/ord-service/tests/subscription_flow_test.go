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
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/gql"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	contentTypeHeader                = "Content-Type"
	contentTypeApplicationURLEncoded = "application/x-www-form-urlencoded"
	contentTypeApplicationJson       = "application/json"

	grantTypeFieldName = "grant_type"
	passwordGrantType  = "password"
	claimsKey          = "claims_key"

	clientIDKey = "client_id"
	userNameKey = "username"
	passwordKey = "password"
)

func TestSelfRegisterFlow(t *testing.T) {
	t.Run("TestSelfRegisterFlow flow: label definitions of the parent tenant are not overwritten", func(t *testing.T) {
		ctx := context.Background()
		distinguishLblValue := "test-distinguish-value"

		// defaultTenantId is the parent of the subaccountID
		defaultTenantId := tenant.TestTenants.GetDefaultTenantID()
		subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

		// Build graphql director client configured with certificate
		clientKey, rawCertChain := certs.ClientCertPair(t, testConfig.ExternalCA.Certificate, testConfig.ExternalCA.Key)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, clientKey, rawCertChain, testConfig.SkipSSLValidation)

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
		runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorCertSecuredClient, subaccountID, &runtimeInput)
		defer fixtures.CleanupRuntime(t, ctx, directorCertSecuredClient, subaccountID, &runtime)
		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)
		strLbl, ok := runtime.Labels[testConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, strLbl, runtime.ID)

		// Verify that the label returned cannot be modified
		setLabelRequest := fixtures.FixSetRuntimeLabelRequest(runtime.ID, testConfig.SelfRegisterLabelKey, "value")
		label := graphql.Label{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, directorCertSecuredClient, subaccountID, setLabelRequest, &label)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("could not set unmodifiable label with key %s", testConfig.SelfRegisterLabelKey))

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

func TestConsumerProviderFlow(t *testing.T) {
	t.Run("ConsumerProvider flow: calls with provider certificate and consumer token are successful when valid subscription exists", func(t *testing.T) {
		ctx := context.Background()
		//defaultTenantId := testConfig.TestProviderAccountID TODO::: Check why it's not used anymore
		secondaryTenant := testConfig.TestConsumerAccountID
		subscriptionProviderID := "xsappname-value"
		//reuseServiceSubaccountID := testConfig.TestReuseSvcSubaccountID
		subscriptionProviderSubaccountID := testConfig.TestProviderSubaccountID
		subscriptionConsumerSubaccountID := testConfig.TestConsumerSubaccountID
		jobName := "external-certificate-rotation-test-job"

		// Prepare provider external client certificate and secret
		k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
		require.NoError(t, err)
		createExtCertJob(t, ctx, k8sClient, testConfig, jobName) // Create temporary external certificate job which will save the modified client certificate in temporary secret
		defer func() {
			k8s.DeleteJob(t, ctx, k8sClient, jobName, testConfig.ExternalClientCertTestSecretNamespace)
			k8s.DeleteSecret(t, ctx, k8sClient, testConfig.ExternalClientCertTestSecretName, testConfig.ExternalClientCertTestSecretNamespace)
		}()
		k8s.WaitForJobToSucceed(t, ctx, k8sClient, jobName, testConfig.ExternalClientCertTestSecretNamespace)

		providerExtCrtTestSecret, err := k8sClient.CoreV1().Secrets(testConfig.ExternalClientCertTestSecretNamespace).Get(ctx, testConfig.ExternalClientCertTestSecretName, metav1.GetOptions{})
		require.NoError(t, err)
		providerKeyBytes := providerExtCrtTestSecret.Data[testConfig.ExternalCA.SecretKeyKey]
		providerCertChainBytes := providerExtCrtTestSecret.Data[testConfig.ExternalCA.SecretCertificateKey]

		// Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certs.ClientCertPair(t, providerCertChainBytes, providerKeyBytes)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, testConfig.SkipSSLValidation)

		runtimeInput := graphql.RuntimeInput{
			Name:        "providerRuntime",
			Description: ptr.String("providerRuntime-description"),
			Labels:      graphql.Labels{testConfig.SubscriptionProviderLabelKey: subscriptionProviderID, tenantfetcher.RegionKey: tenantfetcher.RegionPathParamValue, selectorKey: subscriptionProviderSubaccountID},
		}

		// Register provider runtime with the necessary label
		runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorCertSecuredClient, subscriptionProviderSubaccountID, &runtimeInput) // TODO:: Change runtime registration to be without tenant header(subscriptionProviderSubaccountID)
		defer fixtures.CleanupRuntime(t, ctx, directorCertSecuredClient, subscriptionProviderSubaccountID, &runtime)                                      // TODO:: Change runtime registration to be without tenant header(subscriptionProviderSubaccountID)
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

		// Create automatic scenario assigment for consumer subaccount
		asaInput := fixtures.FixAutomaticScenarioAssigmentInput(scenarios[1], selectorKey, subscriptionConsumerSubaccountID)
		fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, secondaryTenant)
		defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios[1])

		providedTenants := tenantfetcher.Tenant{
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
		subscribeReq := tenantfetcher.CreateTenantRequest(t, providedTenants, tenantProperties, http.MethodPut, testConfig.TenantFetcherFullRegionalURL, testConfig.TokenURL, testConfig.ClientID, testConfig.ClientSecret)

		t.Log(fmt.Sprintf("Creating a subscription between consumer with subaccount id: %s and provider with name: %s and subaccount id: %s", tenantfetcher.ActualTenantID(providedTenants), runtime.Name, subscriptionProviderSubaccountID))
		subscribeResp, err := httpClient.Do(subscribeReq)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, subscribeResp.StatusCode)

		defer func() {
			unsubscribeReq := tenantfetcher.CreateTenantRequest(t, providedTenants, tenantProperties, http.MethodDelete, testConfig.TenantFetcherFullRegionalURL, testConfig.TokenURL, testConfig.ClientID, testConfig.ClientSecret)

			t.Log(fmt.Sprintf("Deleting a subscription between consumer with subaccount id: %s and provider with name: %s and subaccount id: %s", tenantfetcher.ActualTenantID(providedTenants), runtime.Name, subscriptionProviderSubaccountID))
			unsubscribeResp, err := httpClient.Do(unsubscribeReq)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, unsubscribeResp.StatusCode)
		}()

		// HTTP client configured with manually signed client certificate
		extIssuerCertHttpClient := extIssuerCertClient(providerClientKey, providerRawCertChain, testConfig.SkipSSLValidation)

		// Create a token with the necessary consumer claims and add it in authorization header
		tkn := token.GetClientCredentialsToken(t, ctx, testConfig.TokenURL+testConfig.TokenPath, testConfig.ClientID, testConfig.ClientSecret, "subscriptionClaims")
		headers := map[string][]string{"Authorization": {fmt.Sprintf("Bearer %s", tkn)}}

		// Make a request to the ORD service with http client containing certificate with provider information and token with the consumer data.
		respBody := makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)

		require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
		require.Equal(t, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())

		// Build unsubscribe request
		unsubscribeReq := tenantfetcher.CreateTenantRequest(t, providedTenants, tenantProperties, http.MethodDelete, testConfig.TenantFetcherFullRegionalURL, testConfig.TokenURL, testConfig.ClientID, testConfig.ClientSecret)

		t.Log(fmt.Sprintf("Remove a subscription between consumer with subaccount id: %s and provider with name: %s and subaccount id: %s", tenantfetcher.ActualTenantID(providedTenants), runtime.Name, subscriptionProviderSubaccountID))
		unsubscribeResp, err := httpClient.Do(unsubscribeReq)

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, unsubscribeResp.StatusCode)
		respBody = makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)

		require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
	})
}

func TestNewChanges(t *testing.T) {
	t.Run("TestNewChanges", func(t *testing.T) {
		ctx := context.Background()
		//defaultTenantId := testConfig.TestProviderAccountID // TODO:: In dev env this should be our CMP GA in correct feature set(B)
		secondaryTenant := testConfig.TestConsumerAccountID // TODO:: In real env can be also our CMP GA in correct feature set
		subscriptionProviderID := "xsappname-value"
		//reuseServiceSubaccountID := testConfig.TestReuseSvcSubaccountID // TODO:: "reuse svc subaacount" var name or something similar, that is our current CMP SA in CMP GA. Check correct Feature Set.
		subscriptionProviderSubaccountID := testConfig.TestProviderSubaccountID // TODO:: TestProviderSubaccount on local setup and new SA (multi tenant app), that is BAS representative. I need to create xsuaa instance of application plan(check describe json properties in local files). Also, create saas-registry instance(plan application) with xsappname to MTA(multitenant application or something similar), onSubscription: callback (maybe it's mandatory) with external svc mock address that return empty body. And getDependency(check what is the correct property) on which we need to attached also external svc mock that return xsappnameClone label value. The third instance should be cert-svc instance, maybe nothing special here, just standard instance.
		subscriptionConsumerSubaccountID := testConfig.TestConsumerSubaccountID // TODO:: make it configurable with different values for local and dev setup. On dev create new "consumer" SA(in correct feature set) that will be used to subscribe to provider SA
		jobName := "external-certificate-rotation-test-job"

		// Prepare provider external client certificate and secret
		k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
		require.NoError(t, err)
		createExtCertJob(t, ctx, k8sClient, testConfig, jobName) // Create temporary external certificate job which will save the modified client certificate in temporary secret
		defer func() {
			k8s.DeleteJob(t, ctx, k8sClient, jobName, testConfig.ExternalClientCertTestSecretNamespace)
			k8s.DeleteSecret(t, ctx, k8sClient, testConfig.ExternalClientCertTestSecretName, testConfig.ExternalClientCertTestSecretNamespace)
		}()
		k8s.WaitForJobToSucceed(t, ctx, k8sClient, jobName, testConfig.ExternalClientCertTestSecretNamespace)

		providerExtCrtTestSecret, err := k8sClient.CoreV1().Secrets(testConfig.ExternalClientCertTestSecretNamespace).Get(ctx, testConfig.ExternalClientCertTestSecretName, metav1.GetOptions{})
		require.NoError(t, err)
		providerKeyBytes := providerExtCrtTestSecret.Data[testConfig.ExternalCA.SecretKeyKey]
		require.NotEmpty(t, providerKeyBytes)
		providerCertChainBytes := providerExtCrtTestSecret.Data[testConfig.ExternalCA.SecretCertificateKey]
		require.NotEmpty(t, providerCertChainBytes)

		// Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certs.ClientCertPair(t, providerCertChainBytes, providerKeyBytes)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, testConfig.SkipSSLValidation)

		runtimeInput := graphql.RuntimeInput{
			Name:        "providerRuntime",
			Description: ptr.String("providerRuntime-description"),
			Labels:      graphql.Labels{testConfig.SubscriptionProviderLabelKey: subscriptionProviderID, tenantfetcher.RegionKey: tenantfetcher.RegionPathParamValue, selectorKey: subscriptionProviderSubaccountID},
		}

		// Register provider runtime with the necessary label
		runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorCertSecuredClient, subscriptionProviderSubaccountID, &runtimeInput) // TODO:: Change runtime registration to be without tenant header(subscriptionProviderSubaccountID)
		defer fixtures.CleanupRuntime(t, ctx, directorCertSecuredClient, subscriptionProviderSubaccountID, &runtime)                                      // TODO:: Change runtime registration to be without tenant header(subscriptionProviderSubaccountID)
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

		// Create automatic scenario assigment for consumer subaccount
		asaInput := fixtures.FixAutomaticScenarioAssigmentInput(scenarios[1], selectorKey, subscriptionConsumerSubaccountID)
		fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, secondaryTenant)
		defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios[1])

		strLbl, ok := runtime.Labels[testConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		response, err := http.DefaultClient.Post(testConfig.SubscriptionURL+"/v1/dependencies/configure", contentTypeApplicationJson, bytes.NewBuffer([]byte(strLbl)))
		require.NoError(t, err)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.Equal(t, http.StatusOK, response.StatusCode)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: testConfig.SkipSSLValidation},
			},
		}

		apiPath := fmt.Sprintf("/saas-manager/v1/application/tenants/%s/subscriptions", subscriptionConsumerSubaccountID)
		subscribeReq, err := http.NewRequest(http.MethodPost, testConfig.SubscriptionURL+apiPath, bytes.NewBuffer([]byte{}))
		require.NoError(t, err)
		tenantFetcherToken := token.GetClientCredentialsToken(t, ctx, testConfig.TokenURL+testConfig.TokenPath, testConfig.ClientID, testConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tenantFetcherToken))

		t.Log(fmt.Sprintf("Creating a subscription between consumer with subaccount id: %q and provider with name: %q and subaccount id: %q", subscriptionConsumerSubaccountID, runtime.Name, subscriptionProviderSubaccountID))
		resp, err := httpClient.Do(subscribeReq)
		require.NoError(t, err)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer func() {
			unsubscribeReq, err := http.NewRequest(http.MethodDelete, testConfig.SubscriptionURL+apiPath, bytes.NewBuffer([]byte{}))
			unsubscribeReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tenantFetcherToken))

			t.Log(fmt.Sprintf("Remove a subscription between consumer with subaccount id: %s and provider with name: %s and subaccount id: %s", subscriptionConsumerSubaccountID, runtime.Name, subscriptionProviderSubaccountID))
			unsubscribeResp, err := httpClient.Do(unsubscribeReq)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, unsubscribeResp.StatusCode)
		}()

		// HTTP client configured with manually signed client certificate
		extIssuerCertHttpClient := extIssuerCertClient(providerClientKey, providerRawCertChain, testConfig.SkipSSLValidation)

		subscriptionToken := token.GetUserToken(t, ctx, testConfig.TokenURL+testConfig.TokenPath, testConfig.ClientID, testConfig.ClientSecret, testConfig.BasicUsername, testConfig.BasicPassword, "subscriptionClaims")
		headers := map[string][]string{"Authorization": {fmt.Sprintf("Bearer %s", subscriptionToken)}}

		// Make a request to the ORD service with http client containing certificate with provider information and token with the consumer data.
		respBody := makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)
		require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
		require.Equal(t, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())

		// Build unsubscribe request
		unsubscribeReq, err := http.NewRequest(http.MethodDelete, testConfig.SubscriptionURL+apiPath, bytes.NewBuffer([]byte{}))
		unsubscribeReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tenantFetcherToken))

		t.Log(fmt.Sprintf("Remove a subscription between consumer with subaccount id: %s and provider with name: %s and subaccount id: %s", subscriptionConsumerSubaccountID, runtime.Name, subscriptionProviderSubaccountID))
		unsubscribeResp, err := httpClient.Do(unsubscribeReq)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, unsubscribeResp.StatusCode)

		respBody = makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)
		require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
	})
}

// createExtCertJob will schedule a temporary kubernetes job from director-external-certificate-rotation-job cronjob
// with replaced certificate subject and secret name so the tests can be executed on real environment with the correct values.
func createExtCertJob(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, testConfig config, jobName string) {
	cronjobName := "director-external-certificate-rotation-job"

	cronjob := k8s.GetCronJob(t, ctx, k8sClient, cronjobName, testConfig.ExternalClientCertTestSecretNamespace)

	// change the secret name and certificate subject
	podContainers := &cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers
	for cIndex := range *podContainers {
		container := &(*podContainers)[cIndex]
		if container.Name == testConfig.ExternalCertCronjobContainerName {
			for eIndex := range container.Env {
				env := &container.Env[eIndex]
				if env.Name == "CLIENT_CERT_SECRET_NAME" {
					env.Value = testConfig.ExternalClientCertTestSecretName
				}
				if env.Name == "CERT_SUBJECT_PATTERN" {
					env.Value = testConfig.TestExternalCertSubject
				}
			}
			break
		}
	}

	jobDef := &v1.Job{
		Spec: v1.JobSpec{
			Parallelism:             cronjob.Spec.JobTemplate.Spec.Parallelism,
			Completions:             cronjob.Spec.JobTemplate.Spec.Completions,
			ActiveDeadlineSeconds:   cronjob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds,
			BackoffLimit:            cronjob.Spec.JobTemplate.Spec.BackoffLimit,
			Selector:                cronjob.Spec.JobTemplate.Spec.Selector,
			ManualSelector:          cronjob.Spec.JobTemplate.Spec.ManualSelector,
			Template:                cronjob.Spec.JobTemplate.Spec.Template,
			TTLSecondsAfterFinished: cronjob.Spec.JobTemplate.Spec.TTLSecondsAfterFinished,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
	}

	k8s.CreateJobByGivenJobDefinition(t, ctx, k8sClient, jobName, testConfig.ExternalClientCertTestSecretNamespace, jobDef)
}
