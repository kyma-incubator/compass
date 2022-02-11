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
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	gcli "github.com/machinebox/graphql"

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
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	authorizationHeader = "Authorization"
	contentTypeHeader   = "Content-Type"
	locationHeader      = "Location"
	jobSucceededStatus  = "SUCCEEDED"
	eventuallyTimeout   = 5 * time.Second
	eventuallyTick      = time.Second

	contentTypeApplicationJson = "application/json"
)

// TODO:: Consider unskipping this test
func TestSelfRegisterFlow(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("TestSelfRegisterFlow flow: label definitions of the parent tenant are not overwritten", func(t *testing.T) {
		ctx := context.Background()
		// defaultTenantId is the parent of the subaccountID
		defaultTenantId := tenant.TestTenants.GetDefaultTenantID()

		// Build graphql director client configured with certificate
		clientKey, rawCertChain := certs.ClientCertPair(t, testConfig.ExternalCA.Certificate, testConfig.ExternalCA.Key)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, clientKey, rawCertChain, testConfig.SkipSSLValidation)

		// Register application
		app, err := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "testingApp", defaultTenantId)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, defaultTenantId, &app)
		require.NoError(t, err)
		require.NotEmpty(t, app.ID)

		formationName := "sr-test-scenario"
		t.Logf("Creating formation with name %s...", formationName)
		createFormationReq := fixtures.FixCreateFormationRequest(formationName)
		executeGQLRequest(t, ctx, createFormationReq, formationName, defaultTenantId)
		t.Logf("Successfully created formation: %s", formationName)

		defer func() {
			t.Logf("Deleting formation with name: %s...", formationName)
			deleteRequest := fixtures.FixDeleteFormationRequest(formationName)
			executeGQLRequest(t, ctx, deleteRequest, formationName, defaultTenantId)
			t.Logf("Successfully deleted formation with name: %s...", formationName)
		}()

		t.Logf("Assign application to formation %s", formationName)
		assignToFormation(t, ctx, app.ID, "APPLICATION", formationName, defaultTenantId)
		t.Logf("Successfully assigned application to formation %s", formationName)

		defer func() {
			t.Logf("Unassign application from formation %s", formationName)
			unassignFromFormation(t, ctx, app.ID, "APPLICATION", formationName, defaultTenantId)
			t.Logf("Successfully unassigned application from formation %s", formationName)
		}()

		// Self register runtime
		runtimeInput := graphql.RuntimeInput{
			Name:        "selfRegisterRuntime",
			Description: ptr.String("selfRegisterRuntime-description"),
			Labels:      graphql.Labels{testConfig.SubscriptionProviderLabelKey: testConfig.SubscriptionProviderID},
		}
		runtime := fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &runtimeInput)
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &runtime)
		require.NotEmpty(t, runtime.ID)
		strLbl, ok := runtime.Labels[testConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, strLbl, runtime.ID)

		// Verify that the label returned cannot be modified
		setLabelRequest := fixtures.FixSetRuntimeLabelRequest(runtime.ID, testConfig.SelfRegisterLabelKey, "value")
		label := graphql.Label{}
		err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, setLabelRequest, &label)
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

// TODO:: Remove after successful validation on dev-val cluster
//func TestConsumerProviderFlow(t *testing.T) {
//	t.Run("ConsumerProvider flow: calls with provider certificate and consumer token are successful when valid subscription exists", func(t *testing.T) {
//		ctx := context.Background()
//		secondaryTenant := testConfig.TestConsumerAccountID
//		subscriptionProviderSubaccountID := testConfig.TestProviderSubaccountID
//		subscriptionConsumerSubaccountID := testConfig.TestConsumerSubaccountID
//		jobName := "external-certificate-rotation-test-job"
//
//		// Prepare provider external client certificate and secret
//		k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
//		require.NoError(t, err)
//		createExtCertJob(t, ctx, k8sClient, testConfig, jobName) // Create temporary external certificate job which will save the modified client certificate in temporary secret
//		defer func() {
//			k8s.DeleteJob(t, ctx, k8sClient, jobName, testConfig.ExternalClientCertTestSecretNamespace)
//			k8s.DeleteSecret(t, ctx, k8sClient, testConfig.ExternalClientCertTestSecretName, testConfig.ExternalClientCertTestSecretNamespace)
//		}()
//		k8s.WaitForJobToSucceed(t, ctx, k8sClient, jobName, testConfig.ExternalClientCertTestSecretNamespace)
//
//		providerExtCrtTestSecret, err := k8sClient.CoreV1().Secrets(testConfig.ExternalClientCertTestSecretNamespace).Get(ctx, testConfig.ExternalClientCertTestSecretName, metav1.GetOptions{})
//		require.NoError(t, err)
//		providerKeyBytes := providerExtCrtTestSecret.Data[testConfig.ExternalCA.SecretKeyKey]
//		providerCertChainBytes := providerExtCrtTestSecret.Data[testConfig.ExternalCA.SecretCertificateKey]
//
//		// Build graphql director client configured with certificate
//		providerClientKey, providerRawCertChain := certs.ClientCertPair(t, providerCertChainBytes, providerKeyBytes)
//		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, testConfig.SkipSSLValidation)
//
//		runtimeInput := graphql.RuntimeInput{
//			Name:        "providerRuntime",
//			Description: ptr.String("providerRuntime-description"),
//			Labels:      graphql.Labels{testConfig.SubscriptionProviderLabelKey: testConfig.SubscriptionProviderID, tenantfetcher.RegionKey: tenantfetcher.RegionPathParamValue},
//		}
//
//		runtime := fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &runtimeInput)
//		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &runtime)
//		require.NotEmpty(t, runtime.ID)
//
//		// Register application
//		app, err := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "testingApp", secondaryTenant)
//		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, secondaryTenant, &app)
//		require.NoError(t, err)
//		require.NotEmpty(t, app.ID)
//
//		// Register consumer application
//		consumerApp, err := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "consumerApp", secondaryTenant)
//		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, secondaryTenant, &consumerApp)
//		require.NoError(t, err)
//		require.NotEmpty(t, consumerApp.ID)
//		require.NotEmpty(t, consumerApp.Name)
//
//		// Create label definition
//		scenarios := []string{"DEFAULT", "consumer-test-scenario"}
//		fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios)
//		defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios[:1])
//
//		// Assign consumer application to scenario
//		appLabelRequest := fixtures.FixSetApplicationLabelRequest(consumerApp.ID, scenariosLabel, scenarios[1:])
//		require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, secondaryTenant, appLabelRequest, nil))
//		defer fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, secondaryTenant, consumerApp.ID, testConfig.DefaultScenarioEnabled)
//
//		// Create automatic scenario assigment for consumer subaccount
//		asaInput := fixtures.FixAutomaticScenarioAssigmentInput(scenarios[1], selectorKey, subscriptionConsumerSubaccountID)
//		fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, secondaryTenant)
//		defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, secondaryTenant, scenarios[1])
//
//		providedTenants := tenantfetcher.Tenant{
//			TenantID:               secondaryTenant,
//			SubaccountID:           subscriptionConsumerSubaccountID,
//			Subdomain:              tenantfetcher.DefaultSubdomain,
//			SubscriptionProviderID: testConfig.SubscriptionProviderID,
//		}
//
//		tenantProperties := tenantfetcher.TenantIDProperties{
//			TenantIDProperty:               testConfig.TenantIDProperty,
//			SubaccountTenantIDProperty:     testConfig.SubaccountTenantIDProperty,
//			CustomerIDProperty:             testConfig.CustomerIDProperty,
//			SubdomainProperty:              testConfig.SubdomainProperty,
//			SubscriptionProviderIDProperty: testConfig.SubscriptionProviderIDProperty,
//		}
//
//		httpClient := &http.Client{
//			Timeout: 10 * time.Second,
//			Transport: &http.Transport{
//				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
//			},
//		}
//
//		// Build a request for consumer subscription
//		subscribeReq := tenantfetcher.CreateTenantRequest(t, providedTenants, tenantProperties, http.MethodPut, testConfig.TenantFetcherFullRegionalURL, testConfig.ConsumerTokenURL, testConfig.ClientID, testConfig.ClientSecret)
//
//		t.Log(fmt.Sprintf("Creating a subscription between consumer with subaccount id: %s and provider with name: %s and subaccount id: %s", tenantfetcher.ActualTenantID(providedTenants), runtime.Name, subscriptionProviderSubaccountID))
//		subscribeResp, err := httpClient.Do(subscribeReq)
//		require.NoError(t, err)
//		require.Equal(t, http.StatusOK, subscribeResp.StatusCode)
//
//		defer func() {
//			unsubscribeReq := tenantfetcher.CreateTenantRequest(t, providedTenants, tenantProperties, http.MethodDelete, testConfig.TenantFetcherFullRegionalURL, testConfig.ConsumerTokenURL, testConfig.ClientID, testConfig.ClientSecret)
//
//			t.Log(fmt.Sprintf("Deleting a subscription between consumer with subaccount id: %s and provider with name: %s and subaccount id: %s", tenantfetcher.ActualTenantID(providedTenants), runtime.Name, subscriptionProviderSubaccountID))
//			unsubscribeResp, err := httpClient.Do(unsubscribeReq)
//			require.NoError(t, err)
//			require.Equal(t, http.StatusOK, unsubscribeResp.StatusCode)
//		}()
//
//		// HTTP client configured with manually signed client certificate
//		extIssuerCertHttpClient := extIssuerCertClient(providerClientKey, providerRawCertChain, testConfig.SkipSSLValidation)
//
//		// Create a token with the necessary consumer claims and add it in authorization header
//		tkn := token.GetClientCredentialsToken(t, ctx, testConfig.ConsumerTokenURL+testConfig.TokenPath, testConfig.ClientID, testConfig.ClientSecret, "subscriptionClaims")
//		headers := map[string][]string{authorizationHeader: {fmt.Sprintf("Bearer %s", tkn)}}
//
//		// Make a request to the ORD service with http client containing certificate with provider information and token with the consumer data.
//		respBody := makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)
//
//		require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
//		require.Equal(t, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())
//
//		// Build unsubscribe request
//		unsubscribeReq := tenantfetcher.CreateTenantRequest(t, providedTenants, tenantProperties, http.MethodDelete, testConfig.TenantFetcherFullRegionalURL, testConfig.ConsumerTokenURL, testConfig.ClientID, testConfig.ClientSecret)
//
//		t.Log(fmt.Sprintf("Remove a subscription between consumer with subaccount id: %s and provider with name: %s and subaccount id: %s", tenantfetcher.ActualTenantID(providedTenants), runtime.Name, subscriptionProviderSubaccountID))
//		unsubscribeResp, err := httpClient.Do(unsubscribeReq)
//
//		require.NoError(t, err)
//		require.Equal(t, http.StatusOK, unsubscribeResp.StatusCode)
//		respBody = makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)
//
//		require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
//	})
//}

func TestNewChanges(t *testing.T) {
	ctx := context.Background()
	secondaryTenant := testConfig.TestConsumerAccountID
	subscriptionProviderSubaccountID := testConfig.TestProviderSubaccountID
	subscriptionConsumerSubaccountID := testConfig.TestConsumerSubaccountID
	subscriptionConsumerTenantID := testConfig.TestConsumerTenantID
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
		Labels:      graphql.Labels{testConfig.SubscriptionProviderLabelKey: testConfig.SubscriptionProviderID, tenantfetcher.RegionKey: tenantfetcher.RegionPathParamValue},
	}

	runtime := fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &runtimeInput)
	defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &runtime)
	require.NotEmpty(t, runtime.ID)

	// TODO:: Consider adding defer deletion for self register cloning?

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

	consumerFormationName := "consumer-test-scenario"
	t.Logf("Creating formation with name %s...", consumerFormationName)
	createFormationReq := fixtures.FixCreateFormationRequest(consumerFormationName)
	executeGQLRequest(t, ctx, createFormationReq, consumerFormationName, secondaryTenant)
	t.Logf("Successfully created formation: %s", consumerFormationName)

	defer func() {
		t.Logf("Deleting formation with name: %s...", consumerFormationName)
		deleteRequest := fixtures.FixDeleteFormationRequest(consumerFormationName)
		executeGQLRequest(t, ctx, deleteRequest, consumerFormationName, secondaryTenant)
		t.Logf("Successfully deleted formation with name: %s...", consumerFormationName)
	}()

	t.Logf("Assign application to formation %s", consumerFormationName)
	assignToFormation(t, ctx, consumerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
	t.Logf("Successfully assigned application to formation %s", consumerFormationName)

	defer func() {
		t.Logf("Unassign application from formation %s", consumerFormationName)
		unassignFromFormation(t, ctx, consumerApp.ID, "APPLICATION", consumerFormationName, secondaryTenant)
		t.Logf("Successfully unassigned application from formation %s", consumerFormationName)
	}()

	t.Logf("Assign tenant %s to formation %s...", subscriptionConsumerSubaccountID, consumerFormationName)
	assignToFormation(t, ctx, subscriptionConsumerSubaccountID, "TENANT", consumerFormationName, secondaryTenant)
	t.Logf("Successfully assigned tenant %s to formation %s", subscriptionConsumerSubaccountID, consumerFormationName)

	defer func() {
		t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, consumerFormationName)
		unassignFromFormation(t, ctx, subscriptionConsumerSubaccountID, "TENANT", consumerFormationName, secondaryTenant)
		t.Logf("Successfully unassigned tenant %s to formation %s", subscriptionConsumerSubaccountID, consumerFormationName)
	}()

	selfRegLabelValue, ok := runtime.Labels[testConfig.SelfRegisterLabelKey].(string)
	require.True(t, ok)
	require.Contains(t, selfRegLabelValue, testConfig.SelfRegisterLabelValuePrefix+runtime.ID)
	response, err := http.DefaultClient.Post(testConfig.ExternalServicesMockBaseURL+"/v1/dependencies/configure", contentTypeApplicationJson, bytes.NewBuffer([]byte(selfRegLabelValue)))
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

	apiPath := fmt.Sprintf("/saas-manager/v1/application/tenants/%s/subscriptions", subscriptionConsumerTenantID)
	subscribeReq, err := http.NewRequest(http.MethodPost, testConfig.SubscriptionURL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
	require.NoError(t, err)
	subscriptionToken := token.GetClientCredentialsToken(t, ctx, testConfig.SubscriptionTokenURL+testConfig.TokenPath, testConfig.SubscriptionClientID, testConfig.SubscriptionClientSecret, "tenantFetcherClaims")
	subscribeReq.Header.Add(authorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
	subscribeReq.Header.Add(contentTypeHeader, contentTypeApplicationJson)

	t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, runtime.Name, runtime.ID, subscriptionProviderSubaccountID)
	resp, err := httpClient.Do(subscribeReq)
	require.NoError(t, err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, body))

	defer buildAndExecuteUnsubscribeRequest(t, runtime, httpClient, apiPath, subscriptionToken, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

	subJobStatusPath := resp.Header.Get(locationHeader)
	require.NotEmpty(t, subJobStatusPath)
	subJobStatusURL := testConfig.SubscriptionURL + subJobStatusPath
	require.Eventually(t, func() bool {
		return getSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == jobSucceededStatus
	}, eventuallyTimeout, eventuallyTick)
	t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, runtime.Name, runtime.ID, subscriptionProviderSubaccountID)

	// HTTP client configured with certificate with patched subject, issued from cert-rotation job
	extIssuerCertHttpClient := extIssuerCertClient(providerClientKey, providerRawCertChain, testConfig.SkipSSLValidation)

	consumerToken := token.GetUserToken(t, ctx, testConfig.ConsumerTokenURL+testConfig.TokenPath, testConfig.ClientID, testConfig.ClientSecret, testConfig.BasicUsername, testConfig.BasicPassword, "subscriptionClaims")
	headers := map[string][]string{authorizationHeader: {fmt.Sprintf("Bearer %s", consumerToken)}}

	// Make a request to the ORD service with http client containing certificate with provider information and token with the consumer data.
	t.Log("Getting consumer application using both provider and consumer credentials...")
	respBody := makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)
	require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
	require.Equal(t, consumerApp.Name, gjson.Get(respBody, "value.0.title").String())
	t.Log("Successfully fetched consumer application using both provider and consumer credentials")

	buildAndExecuteUnsubscribeRequest(t, runtime, httpClient, apiPath, subscriptionToken, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

	t.Log("Validating no application is returned after successful unsubscription request...")
	respBody = makeRequestWithHeaders(t, extIssuerCertHttpClient, testConfig.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)
	require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
	t.Log("Successfully validated no application is returned after successful unsubscription request")
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
				if env.Name == "CERT_SVC_CSR_ENDPOINT" || env.Name == "CERT_SVC_CLIENT_ID" || env.Name == "CERT_SVC_CLIENT_SECRET" || env.Name == "CERT_SVC_OAUTH_URL" {
					env.ValueFrom.SecretKeyRef.Name = testConfig.CertSvcInstanceTestSecretName // external certificate credentials used to execute consumer-provider test
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

func buildAndExecuteUnsubscribeRequest(t *testing.T, runtime graphql.RuntimeExt, httpClient *http.Client, apiPath, subscriptionToken, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID string) {
	unsubscribeReq, err := http.NewRequest(http.MethodDelete, testConfig.SubscriptionURL+apiPath, bytes.NewBuffer([]byte{}))
	require.NoError(t, err)
	unsubscribeReq.Header.Add(authorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))

	t.Logf("Removing subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, runtime.Name, runtime.ID, subscriptionProviderSubaccountID)
	unsubscribeResp, err := httpClient.Do(unsubscribeReq)
	require.NoError(t, err)
	unsubscribeBody, err := ioutil.ReadAll(unsubscribeResp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, unsubscribeResp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", unsubscribeResp.StatusCode, http.StatusAccepted, unsubscribeBody))
	unsubJobStatusPath := unsubscribeResp.Header.Get(locationHeader)
	require.NotEmpty(t, unsubJobStatusPath)
	unsubJobStatusURL := testConfig.SubscriptionURL + unsubJobStatusPath
	require.Eventually(t, func() bool {
		return getSubscriptionJobStatus(t, httpClient, unsubJobStatusURL, subscriptionToken) == jobSucceededStatus
	}, eventuallyTimeout, eventuallyTick)
	t.Logf("Successfully removed subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q, and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, runtime.Name, runtime.ID, subscriptionProviderSubaccountID)
}

func getSubscriptionJobStatus(t *testing.T, httpClient *http.Client, jobStatusURL, token string) string {
	getJobReq, err := http.NewRequest(http.MethodGet, jobStatusURL, bytes.NewBuffer([]byte{}))
	require.NoError(t, err)
	getJobReq.Header.Add(authorizationHeader, fmt.Sprintf("Bearer %s", token))
	getJobReq.Header.Add(contentTypeHeader, contentTypeApplicationJson)

	resp, err := httpClient.Do(getJobReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	id := gjson.GetBytes(respBody, "id")
	state := gjson.GetBytes(respBody, "state")
	require.True(t, id.Exists())
	require.True(t, state.Exists())
	t.Logf("state of the asynchronous job with id: %s is: %s", id.String(), state.String())

	jobErr := gjson.GetBytes(respBody, "error")
	if jobErr.Exists() {
		t.Errorf("Error occurred while executing asynchronous subscription job: %s", jobErr.String())
	}

	return state.String()
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
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, gqlRequest, &formation)
	require.NoError(t, err)
	require.Equal(t, formationName, formation.Name)
}
