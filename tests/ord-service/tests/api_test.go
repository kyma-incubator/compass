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
	"crypto"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	urlpkg "net/url"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/tests/pkg/gql"

	"github.com/kyma-incubator/compass/tests/pkg/k8s"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/request"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	acceptHeader                 = "Accept"
	tenantHeader                 = "Tenant"
	scenarioName                 = "formation-filtering"
	scenariosLabel               = "scenarios"
	restAPIProtocol              = "rest"
	odataV2APIProtocol           = "odata-v2"
	applicationTenantIDHeaderKey = "applicationTenantId"
)

func TestORDService(t *testing.T) {
	ctx := context.Background()

	defaultTestTenant := tenant.TestTenants.GetIDByName(t, tenant.TenantSeparationTenantName)
	secondaryTenant := tenant.TestTenants.GetIDByName(t, tenant.ApplicationsForRuntimeTenantName)

	tenantFilteringTenant := conf.AccountTenantID
	subTenantID := conf.SubaccountTenantID

	tenantAPIProtocolFiltering := tenant.TestTenants.GetIDByName(t, tenant.ListLabelDefinitionsTenantName)

	// Cannot use tenant constants as the names become too long and cannot be inserted
	appInput := fixtures.CreateApp("tenant1")
	appInput2 := fixtures.CreateApp("tenant2")
	appInputInScenario := fixtures.CreateApp("tenant3-in-scenario")
	appInputInScenario.Labels = map[string]interface{}{
		conf.ApplicationTypeLabelKey: string(util.ApplicationTypeC4C),
	}
	appInputNotInScenario := fixtures.CreateApp("tenant3-no-scenario")
	appInputAPIProtocolFiltering := fixtures.CreateApp("tenant4")
	appInputAPIProtocolFiltering.Bundles = append(appInputAPIProtocolFiltering.Bundles, fixtures.FixBundleWithOnlyOdataAPIs())

	apisMap := make(map[string]directorSchema.APIDefinitionInput, 0)
	for _, apiDefinition := range appInput.Bundles[0].APIDefinitions {
		apisMap[apiDefinition.Name] = *apiDefinition
	}

	eventsMap := make(map[string]directorSchema.EventDefinitionInput, 0)
	for _, eventDefinition := range appInput.Bundles[0].EventDefinitions {
		eventsMap[eventDefinition.Name] = *eventDefinition
	}

	apisMap2 := make(map[string]directorSchema.APIDefinitionInput, 0)
	for _, apiDefinition := range appInput2.Bundles[0].APIDefinitions {
		apisMap2[apiDefinition.Name] = *apiDefinition
	}

	eventsMap2 := make(map[string]directorSchema.EventDefinitionInput, 0)
	for _, eventDefinition := range appInput2.Bundles[0].EventDefinitions {
		eventsMap2[eventDefinition.Name] = *eventDefinition
	}

	apisMapInScenario := make(map[string]directorSchema.APIDefinitionInput, 0)
	for _, apiDefinition := range appInputInScenario.Bundles[0].APIDefinitions {
		apisMapInScenario[apiDefinition.Name] = *apiDefinition
	}

	eventsMapInScenario := make(map[string]directorSchema.EventDefinitionInput, 0)
	for _, eventDefinition := range appInputInScenario.Bundles[0].EventDefinitions {
		eventsMapInScenario[eventDefinition.Name] = *eventDefinition
	}

	app, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, defaultTestTenant, appInput)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, defaultTestTenant, &app)
	require.NoError(t, err)

	app2, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, secondaryTenant, appInput2)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, secondaryTenant, &app2)
	require.NoError(t, err)

	appInScenario, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, tenantFilteringTenant, appInputInScenario)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantFilteringTenant, &appInScenario)
	require.NoError(t, err)

	appNotInScenario, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, tenantFilteringTenant, appInputNotInScenario)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantFilteringTenant, &appNotInScenario)
	require.NoError(t, err)

	appAPIProtocolFiltering, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, tenantAPIProtocolFiltering, appInputAPIProtocolFiltering)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantAPIProtocolFiltering, &appAPIProtocolFiltering)
	require.NoError(t, err)

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", "test-int-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSystemCredentials := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSys.ID)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSystemCredentials.ID)

	unsecuredHttpClient := http.DefaultClient
	unsecuredHttpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	intSystemHttpClient, err := clients.NewIntegrationSystemClient(ctx, intSystemCredentials)
	require.NoError(t, err)

	commonName := "anotherCommonName"
	replacer := strings.NewReplacer(conf.TestProviderSubaccountID, subTenantID, conf.TestExternalCertCN, commonName)
	externalCertProviderConfig := createExternalConfigProvider(replacer.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject))

	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig, true)
	extIssuerCertHttpClient := CreateHttpClientWithCert(providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

	t.Run("401 when requests to ORD Service are unsecured", func(t *testing.T) {
		params := url.Values{}
		params.Add("$format", "json")

		serviceURL := conf.ORDServiceURL + "/$metadata?" + params.Encode()
		makeRequestWithStatusExpect(t, unsecuredHttpClient, serviceURL, http.StatusUnauthorized)
	})

	t.Run("400 when requests to ORD Service do not have tenant header", func(t *testing.T) {
		params := url.Values{}
		params.Add("$select", "id,title")
		params.Add("$format", "json")

		serviceURL := conf.ORDServiceURL + "/consumptionBundles?" + params.Encode()
		makeRequestWithStatusExpect(t, intSystemHttpClient, serviceURL, http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service have wrong tenant header", func(t *testing.T) {
		params := url.Values{}
		params.Add("$select", "id,title")
		params.Add("$format", "json")

		serviceURL := conf.ORDServiceURL + "/consumptionBundles?" + params.Encode()
		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, serviceURL, map[string][]string{tenantHeader: {" "}}, http.StatusBadRequest, conf.ORDServiceDefaultResponseType)
	})

	t.Run("400 when requests to ORD Service api specification do not have tenant header", func(t *testing.T) {
		params := url.Values{}
		params.Add("$select", "id,resourceDefinitions")
		params.Add("$format", "json")

		serviceURL := conf.ORDServiceURL + "/apis?" + params.Encode()
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, serviceURL, map[string][]string{tenantHeader: {defaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].APIDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()
		makeRequestWithStatusExpect(t, intSystemHttpClient, specURL, http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service event specification do not have tenant header", func(t *testing.T) {
		params := url.Values{}
		params.Add("$select", "id,resourceDefinitions")
		params.Add("$format", "json")

		serviceURL := conf.ORDServiceURL + "/events?" + params.Encode()
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, serviceURL, map[string][]string{tenantHeader: {defaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].EventDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()
		makeRequestWithStatusExpect(t, intSystemHttpClient, specURL, http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service api specification have wrong tenant header", func(t *testing.T) {
		params := url.Values{}
		params.Add("$select", "id,resourceDefinitions")
		params.Add("$format", "json")

		serviceURL := conf.ORDServiceURL + "/apis?" + params.Encode()
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, serviceURL, map[string][]string{tenantHeader: {defaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].APIDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()
		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, specURL, map[string][]string{tenantHeader: {" "}}, http.StatusBadRequest, conf.ORDServiceDefaultResponseType)
	})

	t.Run("400 when requests to ORD Service event specification have wrong tenant header", func(t *testing.T) {
		params := url.Values{}
		params.Add("$select", "id,resourceDefinitions")
		params.Add("$format", "json")

		serviceURL := conf.ORDServiceURL + "/events?" + params.Encode()
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, serviceURL, map[string][]string{tenantHeader: {defaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].EventDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()
		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, specURL, map[string][]string{tenantHeader: {" "}}, http.StatusBadRequest, conf.ORDServiceDefaultResponseType)
	})

	t.Run("Requesting entities without specifying response format falls back to configured default response type when Accept header allows everything", func(t *testing.T) {
		params := url.Values{}
		params.Add("$select", "id,title")
		serviceURL := conf.ORDServiceURL + "/consumptionBundles?" + params.Encode()

		makeRequestWithHeaders(t, intSystemHttpClient, serviceURL, map[string][]string{acceptHeader: {"*/*"}, tenantHeader: {defaultTestTenant}})
	})

	t.Run("Requesting entities without specifying response format falls back to response type specified by Accept header when it provides a specific type", func(t *testing.T) {
		params := url.Values{}
		params.Add("$select", "id,title")
		serviceURL := conf.ORDServiceURL + "/consumptionBundles?" + params.Encode()

		makeRequestWithHeaders(t, intSystemHttpClient, serviceURL, map[string][]string{acceptHeader: {"application/json"}, tenantHeader: {defaultTestTenant}})
	})

	t.Run("Requesting Packages returns empty", func(t *testing.T) {
		params := url.Values{}
		params.Add("$expand", "apis,events")
		params.Add("$format", "json")

		serviceURL := conf.ORDServiceURL + "/packages?" + params.Encode()
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, serviceURL, map[string][]string{tenantHeader: {defaultTestTenant}})
		require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
	})

	t.Run("Requesting filtering of Bundles that do not have only ODATA APIs", func(t *testing.T) {
		params := url.Values{}
		params.Add("$filter", "apis/any(d:d/apiProtocol ne 'odata-v2')")
		params.Add("$select", "title,description")
		params.Add("$expand", "apis($select=apiProtocol)")
		params.Add("$format", "json")

		serviceURL := conf.ORDServiceURL + "/consumptionBundles?" + strings.ReplaceAll(params.Encode(), "+", "%20")
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, serviceURL, map[string][]string{tenantHeader: {tenantAPIProtocolFiltering}})

		require.Equal(t, len(appInputAPIProtocolFiltering.Bundles)-1, len(gjson.Get(respBody, "value").Array()))
		require.Equal(t, appInputAPIProtocolFiltering.Bundles[0].Name, gjson.Get(respBody, "value.0.title").String())
		require.Equal(t, *appInputAPIProtocolFiltering.Bundles[0].Description, gjson.Get(respBody, "value.0.description").String())

		// validate that among the returned APIs there is one with apiProtocol == rest
		require.Equal(t, len(appInputAPIProtocolFiltering.Bundles[0].APIDefinitions), len(gjson.Get(respBody, "value.0.apis").Array()))

		isRestAPIFound := false
		for i := range appInputAPIProtocolFiltering.Bundles[0].APIDefinitions {
			apiProtocol := gjson.Get(respBody, fmt.Sprintf("value.0.apis.%d.apiProtocol", i)).String()
			require.NotEmpty(t, apiProtocol)

			if apiProtocol == restAPIProtocol {
				isRestAPIFound = true
			}
		}
		require.Equal(t, true, isRestAPIFound)
	})

	t.Run("Requesting filtering of Bundles that have only ODATA APIs", func(t *testing.T) {
		params := url.Values{}
		params.Add("$filter", "apis/all(d:d/apiProtocol eq 'odata-v2')")
		params.Add("$select", "title,description")
		params.Add("$expand", "apis($select=apiProtocol)")
		params.Add("$format", "json")

		serviceURL := conf.ORDServiceURL + "/consumptionBundles?" + strings.ReplaceAll(params.Encode(), "+", "%20")
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, serviceURL, map[string][]string{tenantHeader: {tenantAPIProtocolFiltering}})

		require.Equal(t, len(appInputAPIProtocolFiltering.Bundles)-1, len(gjson.Get(respBody, "value").Array()))
		require.Equal(t, appInputAPIProtocolFiltering.Bundles[1].Name, gjson.Get(respBody, "value.0.title").String())
		require.Equal(t, *appInputAPIProtocolFiltering.Bundles[1].Description, gjson.Get(respBody, "value.0.description").String())

		// validate that among the returned APIs all are with apiProtocol == odata-v2
		require.Equal(t, len(appInputAPIProtocolFiltering.Bundles[1].APIDefinitions), len(gjson.Get(respBody, "value.0.apis").Array()))

		areAllAPIsODATA := true
		for i := range appInputAPIProtocolFiltering.Bundles[1].APIDefinitions {
			apiProtocol := gjson.Get(respBody, fmt.Sprintf("value.0.apis.%d.apiProtocol", i)).String()
			require.NotEmpty(t, apiProtocol)

			if apiProtocol != odataV2APIProtocol {
				areAllAPIsODATA = false
			}
		}
		require.Equal(t, true, areAllAPIsODATA)
	})

	for _, resource := range []string{"vendors", "tombstones", "products"} { // This tests assert integrity between ORD Service JPA model and our Database model
		t.Run(fmt.Sprintf("Requesting %s", resource), func(t *testing.T) {
			params := url.Values{}
			params.Add("$format", "json")

			serviceURL := fmt.Sprintf("%s/%s?", conf.ORDServiceURL, resource) + params.Encode()
			respBody := makeRequestWithHeaders(t, intSystemHttpClient, serviceURL, map[string][]string{tenantHeader: {defaultTestTenant}})
			require.True(t, gjson.Get(respBody, "value").Exists())
		})
	}

	// create label definition
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantFilteringTenant, scenarioName)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantFilteringTenant, scenarioName)

	// create automatic scenario assigment for subTenant
	formationInput := directorSchema.FormationInput{Name: scenarioName}
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput.Name, subTenantID, tenantFilteringTenant)
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subTenantID, tenantFilteringTenant)

	// assert no system instances are visible without formation
	params := url.Values{}
	params.Add("$format", "json")

	respBody := makeRequestWithHeadersAndQueryParams(t, intSystemHttpClient, conf.ORDServiceURL+"/systemInstances?", map[string][]string{tenantHeader: {subTenantID}}, params)
	require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))

	// assign application to scenario
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, directorSchema.FormationInput{Name: scenarioName}, appInScenario.ID, tenantFilteringTenant)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, directorSchema.FormationInput{Name: scenarioName}, appInScenario.ID, tenantFilteringTenant)

	for _, testData := range []struct {
		msg       string
		headers   map[string][]string
		appInput  directorSchema.ApplicationRegisterInput
		apisMap   map[string]directorSchema.APIDefinitionInput
		eventsMap map[string]directorSchema.EventDefinitionInput
		client    *http.Client
		url       string
	}{
		{
			msg:       defaultTestTenant + " as integration system",
			headers:   map[string][]string{tenantHeader: {defaultTestTenant}},
			appInput:  appInput,
			apisMap:   apisMap,
			eventsMap: eventsMap,
			client:    intSystemHttpClient,
			url:       conf.ORDServiceURL,
		},
		{
			msg:       secondaryTenant + " as integration system",
			headers:   map[string][]string{tenantHeader: {secondaryTenant}},
			appInput:  appInput2,
			apisMap:   apisMap2,
			eventsMap: eventsMap2,
			client:    intSystemHttpClient,
			url:       conf.ORDServiceURL,
		},
		{
			msg:       subTenantID + " as integration system",
			headers:   map[string][]string{tenantHeader: {subTenantID}},
			appInput:  appInputInScenario,
			apisMap:   apisMapInScenario,
			eventsMap: eventsMapInScenario,
			client:    intSystemHttpClient,
			url:       conf.ORDServiceURL,
		},
		{
			msg:       subTenantID + " as Runtime using externally issued certificate",
			headers:   map[string][]string{}, // The tenant comes from the certificate
			appInput:  appInputInScenario,
			apisMap:   apisMapInScenario,
			eventsMap: eventsMapInScenario,
			client:    extIssuerCertHttpClient,
			url:       conf.ORDExternalCertSecuredServiceURL,
		},
	} {

		t.Run(fmt.Sprintf("Requesting System Instances for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			params := url.Values{}
			params.Add("$format", "json")

			respBody := makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/systemInstances?", testData.headers, params)

			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, *testData.appInput.Description, gjson.Get(respBody, "value.0.description").String())
		})

		t.Run(fmt.Sprintf("Requesting System Instances with apis for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			params := url.Values{}
			params.Add("$expand", "apis($select=id,title,description,entryPoints,partOfConsumptionBundles,releaseStatus,apiProtocol,resourceDefinitions)")
			params.Add("$format", "json")

			respBody := makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/systemInstances?", testData.headers, params)
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, *testData.appInput.Description, gjson.Get(respBody, "value.0.description").String())

			assertEqualAPIDefinitions(t, testData.appInput.Bundles[0].APIDefinitions, gjson.Get(respBody, "value.0.apis").String(), testData.apisMap, testData.client, testData.headers)
		})

		t.Run(fmt.Sprintf("Requesting System Instances with events for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			params := url.Values{}
			params.Add("$expand", "events($select=id,title,description,partOfConsumptionBundles,releaseStatus,resourceDefinitions)")
			params.Add("$format", "json")

			respBody := makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/systemInstances?", testData.headers, params)
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, *testData.appInput.Description, gjson.Get(respBody, "value.0.description").String())

			assertEqualEventDefinitions(t, testData.appInput.Bundles[0].EventDefinitions, gjson.Get(respBody, "value.0.events").String(), testData.eventsMap, testData.client, testData.headers)
		})

		t.Run(fmt.Sprintf("Requesting Bundles for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			params := url.Values{}
			params.Add("$select", "title,description")
			params.Add("$format", "json")

			respBody := makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/consumptionBundles?", testData.headers, params)

			require.Equal(t, len(testData.appInput.Bundles), len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Bundles[0].Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, *testData.appInput.Bundles[0].Description, gjson.Get(respBody, "value.0.description").String())
		})

		t.Run(fmt.Sprintf("Requesting APIs and their specs for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			params := url.Values{}
			params.Add("$select", "id,title,description,entryPoints,partOfConsumptionBundles,releaseStatus,apiProtocol,resourceDefinitions")
			params.Add("$format", "json")

			respBody := makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/apis?", testData.headers, params)

			assertEqualAPIDefinitions(t, testData.appInput.Bundles[0].APIDefinitions, gjson.Get(respBody, "value").String(), testData.apisMap, testData.client, testData.headers)
		})

		t.Run(fmt.Sprintf("Requesting Events and their specs for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			params := url.Values{}
			params.Add("$select", "id,title,description,partOfConsumptionBundles,releaseStatus,resourceDefinitions")
			params.Add("$format", "json")

			respBody := makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/events?", testData.headers, params)

			assertEqualEventDefinitions(t, testData.appInput.Bundles[0].EventDefinitions, gjson.Get(respBody, "value").String(), testData.eventsMap, testData.client, testData.headers)
		})

		// Paging:
		t.Run(fmt.Sprintf("Requesting paging of Bundles for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles)
			params := url.Values{}
			params.Add("$top", "10")
			params.Add("$skip", "0")
			params.Add("$select", "id,title")
			params.Add("$format", "json")

			respBody := makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/consumptionBundles?", testData.headers, params)
			require.Equal(t, totalCount, len(gjson.Get(respBody, "value").Array()))

			params = url.Values{}
			params.Add("$top", "10")
			params.Add("$skip", fmt.Sprintf("%d", totalCount))
			params.Add("$select", "id,title")
			params.Add("$format", "json")

			respBody = makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/consumptionBundles?", testData.headers, params)
			require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
		})

		t.Run(fmt.Sprintf("Requesting paging of Bundle APIs for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].APIDefinitions)
			params := urlpkg.Values{}

			params.Add("$select", "id,title")
			params.Add("$expand", "apis($top=10;$select=id,title)")
			params.Add("$format", "json")
			respBody := makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/consumptionBundles?", testData.headers, params)
			require.Equal(t, totalCount, len(gjson.Get(respBody, "value.0.apis").Array()))

			expectedItemCount := 1
			params.Set("$expand", fmt.Sprintf("apis($top=10;$skip=%d;$select=id,title)", totalCount-expectedItemCount))
			respBody = makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/consumptionBundles?", testData.headers, params)
			require.Equal(t, expectedItemCount, len(gjson.Get(respBody, "value").Array()))
		})

		t.Run(fmt.Sprintf("Requesting paging of Bundle Events for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].EventDefinitions)
			params := urlpkg.Values{}

			params.Add("$select", "id,title")
			params.Add("$expand", "events($top=10;$select=id,title)")
			params.Add("$format", "json")
			respBody := makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/consumptionBundles?", testData.headers, params)
			require.Equal(t, totalCount, len(gjson.Get(respBody, "value.0.events").Array()))

			expectedItemCount := 1
			params.Set("$expand", fmt.Sprintf("events($top=10;$skip=%d;$select=id,title)", totalCount-expectedItemCount))
			respBody = makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/consumptionBundles?", testData.headers, params)
			require.Equal(t, expectedItemCount, len(gjson.Get(respBody, "value").Array()))
		})

		// Filtering:
		t.Run(fmt.Sprintf("Requesting filtering of Bundles for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			bndlName := testData.appInput.Bundles[0].Name

			params := urlpkg.Values{}
			params.Add("$filter", fmt.Sprintf("(title eq '%s')", bndlName))
			params.Add("$select", "id,title")
			params.Add("$format", "json")
			serviceURL := testData.url + "/consumptionBundles?" + strings.ReplaceAll(params.Encode(), "+", "%20")
			respBody := makeRequestWithHeaders(t, testData.client, serviceURL, testData.headers)
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

			params.Set("$filter", fmt.Sprintf("(title ne '%s')", bndlName))
			serviceURL = testData.url + "/consumptionBundles?" + strings.ReplaceAll(params.Encode(), "+", "%20")
			respBody = makeRequestWithHeaders(t, testData.client, serviceURL, testData.headers)
			require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
		})

		t.Run(fmt.Sprintf("Requesting filtering of Bundle APIs for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].APIDefinitions)
			apiName := testData.appInput.Bundles[0].APIDefinitions[0].Name

			params := urlpkg.Values{}

			params.Add("$expand", fmt.Sprintf("apis($filter=(title eq '%s');$select=id,title)", apiName))
			params.Add("$select", "id,title")
			params.Add("$format", "json")
			serviceURL := testData.url + "/consumptionBundles?" + strings.ReplaceAll(params.Encode(), "+", "%20")
			respBody := makeRequestWithHeaders(t, testData.client, serviceURL, testData.headers)
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

			params.Set("$expand", fmt.Sprintf("apis($filter=(title ne '%s');$select=id,title)", apiName))
			serviceURL = testData.url + "/consumptionBundles?" + strings.ReplaceAll(params.Encode(), "+", "%20")
			respBody = makeRequestWithHeaders(t, testData.client, serviceURL, testData.headers)
			require.Equal(t, totalCount-1, len(gjson.Get(respBody, "value.0.apis").Array()))
		})

		t.Run(fmt.Sprintf("Requesting filtering of Bundle Events for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].EventDefinitions)
			eventName := testData.appInput.Bundles[0].EventDefinitions[0].Name

			params := urlpkg.Values{}
			params.Add("$expand", fmt.Sprintf("events($filter=(title eq '%s');$select=id,title)", eventName))
			params.Add("$select", "id,title")
			params.Add("$format", "json")
			serviceURL := testData.url + "/consumptionBundles?" + strings.ReplaceAll(params.Encode(), "+", "%20")
			respBody := makeRequestWithHeaders(t, testData.client, serviceURL, testData.headers)
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

			params.Set("$expand", fmt.Sprintf("events($filter=(title ne '%s');$select=id,title)", eventName))
			serviceURL = testData.url + "/consumptionBundles?" + strings.ReplaceAll(params.Encode(), "+", "%20")
			respBody = makeRequestWithHeaders(t, testData.client, serviceURL, testData.headers)
			require.Equal(t, totalCount-1, len(gjson.Get(respBody, "value.0.events").Array()))
		})

		// Projection:
		t.Run(fmt.Sprintf("Requesting projection of Bundles for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			params := urlpkg.Values{}

			params.Add("$select", "title")
			params.Add("$format", "json")
			respBody := makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/consumptionBundles?", testData.headers, params)
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Bundles[0].Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, false, gjson.Get(respBody, "value.0.description").Exists())
		})

		t.Run(fmt.Sprintf("Requesting projection of Bundle APIs for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			params := urlpkg.Values{}

			params.Add("$select", "id,title")
			params.Add("$expand", "apis($select=title)")
			params.Add("$format", "json")
			respBody := makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/consumptionBundles?", testData.headers, params)

			apis := gjson.Get(respBody, "value.0.apis").Array()
			require.Len(t, apis, len(testData.appInput.Bundles[0].APIDefinitions))

			for i := range testData.appInput.Bundles[0].APIDefinitions {
				name := apis[i].Get("title").String()
				require.NotEmpty(t, name)

				_, exists := testData.apisMap[name]
				require.True(t, exists)
				require.False(t, apis[i].Get("description").Exists())
			}
		})

		t.Run(fmt.Sprintf("Requesting projection of Bundle Events for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			params := urlpkg.Values{}

			params.Add("$select", "id,title")
			params.Add("$expand", "events($select=title)")
			params.Add("$format", "json")
			respBody := makeRequestWithHeadersAndQueryParams(t, testData.client, testData.url+"/consumptionBundles?", testData.headers, params)

			events := gjson.Get(respBody, "value.0.events").Array()
			require.Len(t, events, len(testData.appInput.Bundles[0].EventDefinitions))

			for i := range testData.appInput.Bundles[0].EventDefinitions {
				name := events[i].Get("title").String()
				require.NotEmpty(t, name)

				_, exists := testData.eventsMap[name]
				require.True(t, exists)
				require.False(t, events[i].Get("description").Exists())
			}
		})

		//Ordering:
		t.Run(fmt.Sprintf("Requesting ordering of Bundles for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			params := urlpkg.Values{}
			params.Add("$select", "id,title,description")
			params.Add("$orderby", "title asc,description desc")
			params.Add("$format", "json")
			serviceURL := testData.url + "/consumptionBundles?" + strings.ReplaceAll(params.Encode(), "+", "%20")
			respBody := makeRequestWithHeaders(t, testData.client, serviceURL, testData.headers)
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Bundles[0].Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, *testData.appInput.Bundles[0].Description, gjson.Get(respBody, "value.0.description").String())
		})

		t.Run(fmt.Sprintf("Requesting ordering of Bundle APIs for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			params := urlpkg.Values{}
			params.Add("$select", "id,title,description")
			params.Add("$expand", fmt.Sprintf("apis($orderby=%s;$select=id,title,description)", "title asc,description desc"))
			params.Add("$format", "json")
			serviceURL := testData.url + "/consumptionBundles?" + strings.ReplaceAll(params.Encode(), "+", "%20")
			respBody := makeRequestWithHeaders(t, testData.client, serviceURL, testData.headers)

			apis := gjson.Get(respBody, "value.0.apis").Array()
			require.Len(t, apis, len(testData.appInput.Bundles[0].APIDefinitions))

			for i := range testData.appInput.Bundles[0].APIDefinitions {
				name := apis[i].Get("title").String()
				require.NotEmpty(t, name)

				expectedAPI, exists := testData.apisMap[name]
				require.True(t, exists)

				require.Equal(t, *expectedAPI.Description, apis[i].Get("description").String())
			}
		})

		t.Run(fmt.Sprintf("Requesting ordering of Bundle Events for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			params := urlpkg.Values{}
			params.Add("$select", "id,title,description")
			params.Add("$expand", fmt.Sprintf("events($orderby=%s;$select=id,title,description)", "title asc,description desc"))
			params.Add("$format", "json")
			serviceURL := testData.url + "/consumptionBundles?" + strings.ReplaceAll(params.Encode(), "+", "%20")
			respBody := makeRequestWithHeaders(t, testData.client, serviceURL, testData.headers)

			events := gjson.Get(respBody, "value.0.events").Array()
			require.Len(t, events, len(testData.appInput.Bundles[0].EventDefinitions))

			for i := range testData.appInput.Bundles[0].EventDefinitions {
				name := events[i].Get("title").String()
				require.NotEmpty(t, name)

				expectedEvent, exists := testData.eventsMap[name]
				require.True(t, exists)

				require.Equal(t, *expectedEvent.Description, events[i].Get("description").String())
			}
		})
	}

	t.Run("404 when request to ORD Service for api spec have another tenant header value", func(t *testing.T) {
		params := urlpkg.Values{}

		params.Add("$select", "id,resourceDefinitions")
		params.Add("$format", "json")
		respBody := makeRequestWithHeadersAndQueryParams(t, intSystemHttpClient, conf.ORDServiceURL+"/apis?", map[string][]string{tenantHeader: {defaultTestTenant}}, params)
		require.Equal(t, len(appInput.Bundles[0].APIDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()

		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, specURL, map[string][]string{tenantHeader: {secondaryTenant}}, http.StatusNotFound, conf.ORDServiceDefaultResponseType)
		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, specURL, map[string][]string{tenantHeader: {defaultTestTenant}}, http.StatusOK, conf.ORDServiceDefaultResponseType)
	})

	t.Run("404 when request to ORD Service for event spec have another tenant header value", func(t *testing.T) {
		params := urlpkg.Values{}

		params.Add("$select", "id,resourceDefinitions")
		params.Add("$format", "json")
		respBody := makeRequestWithHeadersAndQueryParams(t, intSystemHttpClient, conf.ORDServiceURL+"/events?", map[string][]string{tenantHeader: {defaultTestTenant}}, params)
		require.Equal(t, len(appInput.Bundles[0].EventDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()

		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, specURL, map[string][]string{tenantHeader: {secondaryTenant}}, http.StatusNotFound, conf.ORDServiceDefaultResponseType)
		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, specURL, map[string][]string{tenantHeader: {defaultTestTenant}}, http.StatusOK, conf.ORDServiceDefaultResponseType)
	})

	t.Run("Errors generate user-friendly message", func(t *testing.T) {
		params := urlpkg.Values{}

		params.Add("$format", "json")
		serviceURL := conf.ORDServiceURL + "/test?" + params.Encode()
		respBody := request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, serviceURL, map[string][]string{tenantHeader: {defaultTestTenant}}, http.StatusNotFound, conf.ORDServiceDefaultResponseType)

		require.Contains(t, gjson.Get(respBody, "error.message").String(), "Use odata-debug query parameter with value one of the following formats: json,html,download for more information")
	})

	t.Run("Additional non-ORD details about system instances are exposed", func(t *testing.T) {
		expectedProductType := fmt.Sprintf("SAP %s", "productType")
		appTmplInput := fixtures.FixApplicationTemplate(expectedProductType)

		t.Log("Creating integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, defaultTestTenant, "ord-service-non-ord-details")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, defaultTestTenant, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, defaultTestTenant, intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issuing a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		appTmpl, err := fixtures.CreateApplicationTemplateFromInputWithoutTenant(t, ctx, oauthGraphQLClient, appTmplInput)
		defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTmpl)
		require.NoError(t, err)

		appFromTmpl := directorSchema.ApplicationFromTemplateInput{
			TemplateName: expectedProductType, Values: []*directorSchema.TemplateValueInput{
				{
					Placeholder: "name",
					Value:       "new-value",
				},
				{
					Placeholder: "display-name",
					Value:       "new-value",
				},
			},
		}

		appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
		require.NoError(t, err)

		createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
		outputApp := directorSchema.ApplicationExt{}
		//WHEN
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, defaultTestTenant, createAppFromTmplRequest, &outputApp)
		defer fixtures.CleanupApplication(t, ctx, oauthGraphQLClient, defaultTestTenant, &outputApp)
		require.NoError(t, err)

		params := urlpkg.Values{}

		params.Add("$format", "json")
		getSystemInstanceURL := fmt.Sprintf("%s/systemInstances(%s)?", conf.ORDServiceURL, outputApp.ID)

		respBody := makeRequestWithHeadersAndQueryParams(t, intSystemHttpClient, getSystemInstanceURL, map[string][]string{tenantHeader: {defaultTestTenant}}, params)

		require.Equal(t, outputApp.Name, gjson.Get(respBody, "title").String())

		t.Run("systemNumber is exposed", func(t *testing.T) {
			require.True(t, gjson.Get(respBody, "systemNumber").Exists())
		})

		t.Run("productType is exposed", func(t *testing.T) {
			require.True(t, gjson.Get(respBody, "productType").Exists())
			require.Equal(t, expectedProductType, gjson.Get(respBody, "productType").String())
		})
	})
}

func TestORDServiceSystemDiscoveryByApplicationTenantID(t *testing.T) {
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetDefaultTenantID()

	certSubject := strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.ExternalCertProviderConfig.TestExternalCertCN, "ord-svc-system-discovery", -1)

	// We need an externally issued cert with a subject that is not part of the access level mappings
	externalCertProviderConfig := createExternalConfigProvider(certSubject)

	// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig, false)

	// HTTP client configured with certificate with patched subject, issued from cert-rotation job
	certHttpClient := CreateHttpClientWithCert(providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, "int-system-ord-service-consumption")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	// The external cert secret created by the NewExternalCertFromConfig above is used by the external-services-mock for the async formation status API call,
	// that's why in the function above there is a false parameter that don't delete it and an explicit defer deletion func is added here
	// so, the secret could be deleted at the end of the test. Otherwise, it will remain as leftover resource in the cluster
	defer func() {
		k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
		require.NoError(t, err)
		k8s.DeleteSecret(t, ctx, k8sClient, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace)
	}()

	// Create Application Template
	namePlaceholder := "name"
	displayNamePlaceholder := "display-name"
	localTenantID := "local-tenant-id-system-discovery"
	applicationType := "app-type-ord-svc-system-discovery"

	// Use oauthGraphQLClient so that the GQL resolver won't create a CertificateSubjectMapping object for that app template
	appTemplateInput := fixtures.FixApplicationTemplateWithoutWebhook(applicationType, localTenantID, "test-app-region", "compass.test", namePlaceholder, displayNamePlaceholder)
	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantID, appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTmpl)
	require.NoError(t, err)
	require.NotEmpty(t, appTmpl.ID)

	t.Logf("Create application from template %q", applicationType)
	appFromTmplSrc := fixtures.FixApplicationFromTemplateInput(applicationType, namePlaceholder, "app-ord-system-discovery-e2e-tests", displayNamePlaceholder, "App ORD service Display Name")
	appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
	require.NoError(t, err)
	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
	application := directorSchema.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantID, createAppFromTmplRequest, &application)
	defer fixtures.CleanupApplication(t, ctx, oauthGraphQLClient, tenantID, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)
	t.Logf("app ID: %q", application.ID)

	// Create certificate subject mapping with custom subject that was used to create a certificate for the graphql client above
	consumerType := string(consumer.ManagedApplicationProviderOperator)
	tenantAccessLevels := []string{"global"} // should be a valid tenant access level
	internalConsumerID := appTmpl.ID         // add application templated ID as certificate subject mapping internal consumer to satisfy the authorization checks in the formation assignment status API
	certSubjectMappingCustomSubjectWithCommaSeparator := strings.ReplaceAll(strings.TrimLeft(certSubject, "/"), "/", ",")

	csmInput := fixtures.FixCertificateSubjectMappingInput(certSubjectMappingCustomSubjectWithCommaSeparator, consumerType, &internalConsumerID, tenantAccessLevels)
	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", certSubjectMappingCustomSubjectWithCommaSeparator, consumerType, tenantAccessLevels)

	var csmCreate directorSchema.CertificateSubjectMapping
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreate)
	csmCreate = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput)
	t.Logf("Sleeping for %s, so the hydrator component could update the certificate subject mapping cache with the new data", conf.CertSubjectMappingResyncInterval.String())
	time.Sleep(conf.CertSubjectMappingResyncInterval)

	consumerAppType := string(util.ApplicationTypeC4C)
	consumerApp, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "consumer-app", conf.ApplicationTypeLabelKey, consumerAppType, tenantID)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &consumerApp)
	require.NoError(t, err)
	require.NotEmpty(t, consumerApp.ID)

	headers := map[string][]string{applicationTenantIDHeaderKey: {localTenantID}}
	// Make a request to the ORD service with http client containing custom certificate and application tenant ID header
	t.Log("Getting application using custom certificate and applicationTenantId header before a formation is created...")
	respBody := makeRequestWithHeaders(t, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)
	require.Empty(t, gjson.Get(respBody, "value").Array())
	t.Log("No system instance details are returned due to missing formation")

	formationTmplName := "e2e-test-formation-template-system-discovery"
	t.Logf("Creating formation template for the provider application tempalаte type %q with name %q", conf.SubscriptionProviderAppNameValue, formationTmplName)
	var ft directorSchema.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
	ft = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, directorSchema.FormationTemplateInput{
		Name:               formationTmplName,
		ApplicationTypes:   []string{applicationType, consumerAppType},
		DiscoveryConsumers: []string{applicationType},
	})

	systemDiscoveryFormationName := "e2e-tests-system-discovery-formation"
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, systemDiscoveryFormationName)
	formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, systemDiscoveryFormationName, &formationTmplName)
	require.NotEmpty(t, formation.ID)
	t.Logf("Successfully created formation: %s", systemDiscoveryFormationName)

	assignToFormation(t, ctx, application.ID, string(directorSchema.FormationObjectTypeApplication), systemDiscoveryFormationName, tenantID)
	defer unassignFromFormation(t, ctx, application.ID, string(directorSchema.FormationObjectTypeApplication), systemDiscoveryFormationName, tenantID)

	assignToFormation(t, ctx, consumerApp.ID, string(directorSchema.FormationObjectTypeApplication), systemDiscoveryFormationName, tenantID)
	defer unassignFromFormation(t, ctx, consumerApp.ID, string(directorSchema.FormationObjectTypeApplication), systemDiscoveryFormationName, tenantID)

	t.Log("Getting application using custom certificate and appplicationTenantId header after formation is created...")
	respBody = makeRequestWithHeaders(t, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)

	require.Len(t, gjson.Get(respBody, "value").Array(), 2)

	isSystemFound := false
	var systemInstanceDetails gjson.Result
	for _, element := range gjson.Get(respBody, "value").Array() {
		systemName := gjson.Get(element.String(), "title")
		if consumerApp.Name == systemName.String() {
			isSystemFound = true
			systemInstanceDetails = element
			break
		}
	}
	require.Equal(t, true, isSystemFound)
	t.Log("Successfully fetched system instance details using custom certificate and application tenant ID header")

	expectedFormationDetailsAssignmentID := getExpectedFormationDetailsAssignmentID(t, ctx, tenantID, consumerApp.ID, application.ID, formation.ID)
	verifyFormationDetails(t, systemInstanceDetails, formation.ID, expectedFormationDetailsAssignmentID, ft.ID)
}

func TestORDServiceSystemDiscoveryByApplicationTenantIDUsingProviderCSM(t *testing.T) {
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetDefaultTenantID()
	cn := "ord-svc-system-with-csm-discovery"

	technicalCertSubject := strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.ExternalCertProviderConfig.TestExternalCertCN, "csm-discovery-technical", -1)
	technicalCertProvider := createExternalConfigProvider(technicalCertSubject)
	technicalProviderClientKey, technicalProviderRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, technicalCertProvider, false)
	technicalCertDirectorGQLClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, technicalProviderClientKey, technicalProviderRawCertChain, conf.SkipSSLValidation)

	certSubject := strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.ExternalCertProviderConfig.TestExternalCertCN, cn, -1)
	// We need an externally issued cert with a subject that is not part of the access level mappings
	externalCertProviderConfig := createExternalConfigProvider(certSubject)

	// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig, false)

	certDirectorGQLClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)
	// HTTP client configured with certificate with patched subject, issued from cert-rotation job
	certHttpClient := CreateHttpClientWithCert(providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

	// The external cert secret created by the NewExternalCertFromConfig above is used by the external-services-mock for the async formation status API call,
	// that's why in the function above there is a false parameter that don't delete it and an explicit defer deletion func is added here
	// so, the secret could be deleted at the end of the test. Otherwise, it will remain as leftover resource in the cluster
	defer func() {
		k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
		require.NoError(t, err)
		k8s.DeleteSecret(t, ctx, k8sClient, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace)
	}()

	// Create Application Template
	namePlaceholder := "name"
	displayNamePlaceholder := "display-name"
	localTenantID := "local-tenant-id-system-discovery"
	applicationType := "app-type-ord-svc-system-discovery"
	productLabelValue := []interface{}{"productLabelValue1"}

	appTemplateInput := fixtures.FixApplicationTemplateWithoutWebhook(applicationType, localTenantID, "test-app-region", "compass.test", namePlaceholder, displayNamePlaceholder)
	appTemplateInput.Labels[conf.ApplicationTemplateProductLabel] = productLabelValue
	appTmpl, err := fixtures.CreateApplicationTemplateFromInputWithoutTenant(t, ctx, certDirectorGQLClient, appTemplateInput)

	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, certDirectorGQLClient, appTmpl)
	require.NoError(t, err)
	require.NotEmpty(t, appTmpl.ID)

	t.Logf("Create application from template %q", applicationType)
	appFromTmplSrc := fixtures.FixApplicationFromTemplateInput(applicationType, namePlaceholder, "app-ord-system-discovery-e2e-tests", displayNamePlaceholder, "App ORD service Display Name")
	appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
	require.NoError(t, err)
	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
	application := directorSchema.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, technicalCertDirectorGQLClient, conf.TestProviderSubaccountID, createAppFromTmplRequest, &application)
	defer fixtures.CleanupApplication(t, ctx, technicalCertDirectorGQLClient, conf.TestProviderSubaccountID, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)
	t.Logf("app ID: %q", application.ID)

	csm := fixtures.FindCertSubjectMappingForApplicationTemplate(t, ctx, certSecuredGraphQLClient, appTmpl.ID, cn)
	require.NotNil(t, csm)

	t.Logf("Sleeping for %s, so the hydrator component could update the certificate subject mapping cache with the new data", conf.CertSubjectMappingResyncInterval.String())
	time.Sleep(conf.CertSubjectMappingResyncInterval)

	consumerAppType := string(util.ApplicationTypeC4C)
	consumerApp, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "consumer-app", conf.ApplicationTypeLabelKey, consumerAppType, tenantID)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &consumerApp)
	require.NoError(t, err)
	require.NotEmpty(t, consumerApp.ID)

	headers := map[string][]string{applicationTenantIDHeaderKey: {localTenantID}}
	// Make a request to the ORD service with http client containing custom certificate and application tenant ID header
	t.Log("Getting application using custom certificate and appplicationTenantId header before a formation is created...")

	params := urlpkg.Values{}
	params.Add("$format", "json")
	respBody := makeRequestWithHeadersAndQueryParams(t, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/systemInstances?", headers, params)

	require.Empty(t, gjson.Get(respBody, "value").Array())
	t.Log("No system instance details are returned due to missing formation")

	formationTmplName := "e2e-test-formation-template-system-discovery"
	t.Logf("Creating formation template for the provider application tempalаte type %q with name %q", conf.SubscriptionProviderAppNameValue, formationTmplName)
	var ft directorSchema.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
	ft = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, directorSchema.FormationTemplateInput{
		Name:               formationTmplName,
		ApplicationTypes:   []string{applicationType, consumerAppType},
		DiscoveryConsumers: []string{applicationType},
	})

	systemDiscoveryFormationName := "e2e-tests-system-discovery-formation"
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, systemDiscoveryFormationName)
	formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, systemDiscoveryFormationName, &formationTmplName)
	require.NotEmpty(t, formation.ID)
	t.Logf("Successfully created formation: %s", systemDiscoveryFormationName)

	assignToFormation(t, ctx, application.ID, string(directorSchema.FormationObjectTypeApplication), systemDiscoveryFormationName, tenantID)
	defer unassignFromFormation(t, ctx, application.ID, string(directorSchema.FormationObjectTypeApplication), systemDiscoveryFormationName, tenantID)

	assignToFormation(t, ctx, consumerApp.ID, string(directorSchema.FormationObjectTypeApplication), systemDiscoveryFormationName, tenantID)
	defer unassignFromFormation(t, ctx, consumerApp.ID, string(directorSchema.FormationObjectTypeApplication), systemDiscoveryFormationName, tenantID)

	t.Log("Getting application using custom certificate and appplicationTenantId header after formation is created...")

	respBody = makeRequestWithHeaders(t, certHttpClient, conf.ORDExternalCertSecuredServiceURL+"/systemInstances?$format=json", headers)

	require.Len(t, gjson.Get(respBody, "value").Array(), 2)

	isSystemFound := false
	var systemInstanceDetails gjson.Result
	for _, element := range gjson.Get(respBody, "value").Array() {
		systemName := gjson.Get(element.String(), "title")
		if consumerApp.Name == systemName.String() {
			isSystemFound = true
			systemInstanceDetails = element
			break
		}
	}
	require.Equal(t, true, isSystemFound)
	t.Log("Successfully fetched system instance details using custom certificate and application tenant ID header")

	expectedFormationDetailsAssignmentID := getExpectedFormationDetailsAssignmentID(t, ctx, tenantID, consumerApp.ID, application.ID, formation.ID)
	verifyFormationDetails(t, systemInstanceDetails, formation.ID, expectedFormationDetailsAssignmentID, ft.ID)
}

func assertEqualAPIDefinitions(t *testing.T, expectedAPIDefinitions []*directorSchema.APIDefinitionInput, actualAPIDefinitions string, apisMap map[string]directorSchema.APIDefinitionInput, client *http.Client, headers map[string][]string) {
	require.Equal(t, len(expectedAPIDefinitions), len(gjson.Parse(actualAPIDefinitions).Array()))

	for i := range expectedAPIDefinitions {
		name := gjson.Get(actualAPIDefinitions, fmt.Sprintf("%d.title", i)).String()
		require.NotEmpty(t, name)

		expectedAPI, exists := apisMap[name]
		require.True(t, exists)

		require.Equal(t, *expectedAPI.Description, gjson.Get(actualAPIDefinitions, fmt.Sprintf("%d.description", i)).String())
		require.Equal(t, expectedAPI.TargetURL, gjson.Get(actualAPIDefinitions, fmt.Sprintf("%d.entryPoints.0.value", i)).String())
		require.NotEmpty(t, gjson.Get(actualAPIDefinitions, fmt.Sprintf("%d.partOfConsumptionBundles", i)).String())

		releaseStatus := gjson.Get(actualAPIDefinitions, fmt.Sprintf("%d.releaseStatus", i)).String()
		switch releaseStatus {
		case "decommissioned":
			require.True(t, *expectedAPI.Version.ForRemoval)
			require.True(t, *expectedAPI.Version.Deprecated)
		case "deprecated":
			require.False(t, *expectedAPI.Version.ForRemoval)
			require.True(t, *expectedAPI.Version.Deprecated)
		case "active":
			require.False(t, *expectedAPI.Version.ForRemoval)
			require.False(t, *expectedAPI.Version.Deprecated)
		default:
			panic(errors.New(fmt.Sprintf("Unknown release status: %s", releaseStatus)))
		}

		apiProtocol := gjson.Get(actualAPIDefinitions, fmt.Sprintf("%d.apiProtocol", i)).String()
		switch apiProtocol {
		case "odata-v2":
			require.Equal(t, expectedAPI.Spec.Type, directorSchema.APISpecTypeOdata)
		case "rest":
			require.Equal(t, expectedAPI.Spec.Type, directorSchema.APISpecTypeOpenAPI)
		default:
			t.Log(fmt.Sprintf("API Protocol for API %s is %s. It does not match a predefined spec type.", name, apiProtocol))
		}

		specs := gjson.Get(actualAPIDefinitions, fmt.Sprintf("%d.resourceDefinitions", i)).Array()
		require.Equal(t, 1, len(specs))

		specType := specs[0].Get("type").String()
		switch specType {
		case "edmx":
			require.Equal(t, expectedAPI.Spec.Type, directorSchema.APISpecTypeOdata)
		case "openapi-v3":
			require.Equal(t, expectedAPI.Spec.Type, directorSchema.APISpecTypeOpenAPI)
		default:
			panic(errors.New(fmt.Sprintf("Unknown spec type: %s", specType)))
		}

		specFormat := specs[0].Get("mediaType").String()
		switch specFormat {
		case "text/yaml":
			require.Equal(t, expectedAPI.Spec.Format, directorSchema.SpecFormatYaml)
		case "application/json":
			require.Equal(t, expectedAPI.Spec.Format, directorSchema.SpecFormatJSON)
		case "application/xml":
			require.Equal(t, expectedAPI.Spec.Format, directorSchema.SpecFormatXML)
		default:
			panic(errors.New(fmt.Sprintf("Unknown spec format: %s", specFormat)))
		}

		apiID := gjson.Get(actualAPIDefinitions, fmt.Sprintf("%d.id", i)).String()
		require.NotEmpty(t, apiID)

		specURL := specs[0].Get("url").String()
		specPath := fmt.Sprintf("/api/%s/specification", apiID)
		require.Contains(t, specURL, conf.ORDServiceStaticPrefix+specPath)

		respBody := makeRequestWithHeaders(t, client, specURL, headers)

		require.Equal(t, string(*expectedAPI.Spec.Data), respBody)

	}
}

func assertEqualEventDefinitions(t *testing.T, expectedEventDefinitions []*directorSchema.EventDefinitionInput, actualEventDefinitions string, eventsMap map[string]directorSchema.EventDefinitionInput, client *http.Client, headers map[string][]string) {
	require.Equal(t, len(expectedEventDefinitions), len(gjson.Parse(actualEventDefinitions).Array()))

	for i := range expectedEventDefinitions {
		name := gjson.Get(actualEventDefinitions, fmt.Sprintf("%d.title", i)).String()
		require.NotEmpty(t, name)

		expectedEvent, exists := eventsMap[name]
		require.True(t, exists)

		require.Equal(t, *expectedEvent.Description, gjson.Get(actualEventDefinitions, fmt.Sprintf("%d.description", i)).String())
		require.NotEmpty(t, gjson.Get(actualEventDefinitions, fmt.Sprintf("%d.partOfConsumptionBundles", i)).String())

		releaseStatus := gjson.Get(actualEventDefinitions, fmt.Sprintf("%d.releaseStatus", i)).String()
		switch releaseStatus {
		case "decommissioned":
			require.True(t, *expectedEvent.Version.ForRemoval)
			require.True(t, *expectedEvent.Version.Deprecated)
		case "deprecated":
			require.False(t, *expectedEvent.Version.ForRemoval)
			require.True(t, *expectedEvent.Version.Deprecated)
		case "active":
			require.False(t, *expectedEvent.Version.ForRemoval)
			require.False(t, *expectedEvent.Version.Deprecated)
		default:
			panic(errors.New(fmt.Sprintf("Unknown release status: %s", releaseStatus)))
		}

		specs := gjson.Get(actualEventDefinitions, fmt.Sprintf("%d.resourceDefinitions", i)).Array()
		require.Equal(t, 1, len(specs))

		specType := specs[0].Get("type").String()
		switch specType {
		case "asyncapi-v2":
			require.Equal(t, expectedEvent.Spec.Type, directorSchema.EventSpecTypeAsyncAPI)
		default:
			panic(errors.New(fmt.Sprintf("Unknown spec type: %s", specType)))
		}

		specFormat := specs[0].Get("mediaType").String()
		switch specFormat {
		case "text/yaml":
			require.Equal(t, expectedEvent.Spec.Format, directorSchema.SpecFormatYaml)
		case "application/json":
			require.Equal(t, expectedEvent.Spec.Format, directorSchema.SpecFormatJSON)
		case "application/xml":
			require.Equal(t, expectedEvent.Spec.Format, directorSchema.SpecFormatXML)
		default:
			panic(errors.New(fmt.Sprintf("Unknown spec format: %s", specFormat)))
		}

		eventID := gjson.Get(actualEventDefinitions, fmt.Sprintf("%d.id", i)).String()
		require.NotEmpty(t, eventID)

		specURL := specs[0].Get("url").String()
		specPath := fmt.Sprintf("/event/%s/specification", eventID)
		require.Contains(t, specURL, conf.ORDServiceStaticPrefix+specPath)

		respBody := makeRequestWithHeaders(t, client, specURL, headers)

		require.Equal(t, string(*expectedEvent.Spec.Data), respBody)
	}
}

func makeRequestWithHeaders(t require.TestingT, httpClient *http.Client, url string, headers map[string][]string) string {
	return request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, url, headers, http.StatusOK, conf.ORDServiceDefaultResponseType)
}

func makeRequestWithHeadersAndQueryParams(t require.TestingT, httpClient *http.Client, url string, headers map[string][]string, params urlpkg.Values) string {
	url = url + params.Encode()
	return request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, url, headers, http.StatusOK, conf.ORDServiceDefaultResponseType)
}

func makeRequestWithStatusExpect(t require.TestingT, httpClient *http.Client, url string, expectedHTTPStatus int) string {
	return request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, url, map[string][]string{}, expectedHTTPStatus, conf.ORDServiceDefaultResponseType)
}

// CreateHttpClientWithCert returns http client configured with provided client certificate and key
func CreateHttpClientWithCert(clientKey crypto.PrivateKey, rawCertChain [][]byte, skipSSLValidation bool) *http.Client {
	return &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{
					{
						Certificate: rawCertChain,
						PrivateKey:  clientKey,
					},
				},
				ClientAuth:         tls.RequireAndVerifyClientCert,
				InsecureSkipVerify: skipSSLValidation,
			},
		},
	}
}

func createExternalConfigProvider(subject string) certprovider.ExternalCertProviderConfig {
	return certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestSecretName:         conf.CertSvcInstanceSecretName,
		ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestExternalCertSubject:               subject,
		ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
		ExternalCertProvider:                  certprovider.CertificateService,
	}
}
