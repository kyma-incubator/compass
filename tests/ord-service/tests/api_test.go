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
	"errors"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"testing"
	"time"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/request"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	acceptHeader       = "Accept"
	tenantHeader       = "Tenant"
	scenarioName       = "formation-filtering"
	scenariosLabel     = "scenarios"
	selectorKey        = "global_subaccount_id"
	subTenantID        = "123e4567-e89b-12d3-a456-426614174001"
	restAPIProtocol    = "rest"
	odataV2APIProtocol = "odata-v2"
)

func TestORDService(stdT *testing.T) {
	t := testingx.NewT(stdT)

	ctx := context.Background()

	defaultTestTenant := tenant.TestTenants.GetDefaultTenantID()
	secondaryTenant := tenant.TestTenants.GetIDByName(t, tenant.ApplicationsForRuntimeTenantName)
	tenantFilteringTenant := tenant.TestTenants.GetIDByName(t, tenant.TenantSeparationTenantName)
	tenantAPIProtocolFiltering := tenant.TestTenants.GetIDByName(t, tenant.ListLabelDefinitionsTenantName)

	// Cannot use tenant constants as the names become too long and cannot be inserted
	appInput := fixtures.CreateApp("tenant1")
	appInput2 := fixtures.CreateApp("tenant2")
	appInputInScenario := fixtures.CreateApp("tenant3-in-scenario")
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

	app, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, defaultTestTenant, appInput)
	defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, defaultTestTenant, &app)
	require.NoError(t, err)

	app2, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, secondaryTenant, appInput2)
	defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, secondaryTenant, &app2)
	require.NoError(t, err)

	appInScenario, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, tenantFilteringTenant, appInputInScenario)
	defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, tenantFilteringTenant, &appInScenario)
	require.NoError(t, err)

	appNotInScenario, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, tenantFilteringTenant, appInputNotInScenario)
	defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, tenantFilteringTenant, &appNotInScenario)
	require.NoError(t, err)

	appAPIProtocolFiltering, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, tenantAPIProtocolFiltering, appInputAPIProtocolFiltering)
	defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, tenantAPIProtocolFiltering, &appAPIProtocolFiltering)
	require.NoError(t, err)

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, "", "test-int-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, dexGraphQLClient, "", intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSystemCredentials := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, "", intSys.ID)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, dexGraphQLClient, intSystemCredentials.ID)

	unsecuredHttpClient := http.DefaultClient
	unsecuredHttpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	intSystemHttpClient, err := clients.NewIntegrationSystemClient(ctx, intSystemCredentials)
	require.NoError(t, err)
	extIssuerCertHttpClient := extIssuerCertClient(t, subTenantID)

	t.Run("401 when requests to ORD Service are unsecured", func(t *testing.T) {
		makeRequestWithStatusExpect(t, unsecuredHttpClient, testConfig.ORDServiceURL+"/$metadata?$format=json", http.StatusUnauthorized)
	})

	t.Run("400 when requests to ORD Service do not have tenant header", func(t *testing.T) {
		makeRequestWithStatusExpect(t, intSystemHttpClient, testConfig.ORDServiceURL+"/consumptionBundles?$format=json", http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service have wrong tenant header", func(t *testing.T) {
		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, testConfig.ORDServiceURL+"/consumptionBundles?$format=json", map[string][]string{tenantHeader: {" "}}, http.StatusBadRequest, testConfig.ORDServiceDefaultResponseType)
	})

	t.Run("400 when requests to ORD Service api specification do not have tenant header", func(t *testing.T) {
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, testConfig.ORDServiceURL+"/apis?$format=json", map[string][]string{tenantHeader: {defaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].APIDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()
		makeRequestWithStatusExpect(t, intSystemHttpClient, specURL, http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service event specification do not have tenant header", func(t *testing.T) {
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, testConfig.ORDServiceURL+"/events?$format=json", map[string][]string{tenantHeader: {defaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].EventDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()
		makeRequestWithStatusExpect(t, intSystemHttpClient, specURL, http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service api specification have wrong tenant header", func(t *testing.T) {
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, testConfig.ORDServiceURL+"/apis?$format=json", map[string][]string{tenantHeader: {defaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].APIDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()
		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, specURL, map[string][]string{tenantHeader: {" "}}, http.StatusBadRequest, testConfig.ORDServiceDefaultResponseType)
	})

	t.Run("400 when requests to ORD Service event specification have wrong tenant header", func(t *testing.T) {
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, testConfig.ORDServiceURL+"/events?$format=json", map[string][]string{tenantHeader: {defaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].EventDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()
		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, specURL, map[string][]string{tenantHeader: {" "}}, http.StatusBadRequest, testConfig.ORDServiceDefaultResponseType)
	})

	t.Run("Requesting entities without specifying response format falls back to configured default response type when Accept header allows everything", func(t *testing.T) {
		makeRequestWithHeaders(t, intSystemHttpClient, testConfig.ORDServiceURL+"/consumptionBundles", map[string][]string{acceptHeader: {"*/*"}, tenantHeader: {defaultTestTenant}})
	})

	t.Run("Requesting entities without specifying response format falls back to response type specified by Accept header when it provides a specific type", func(t *testing.T) {
		makeRequestWithHeaders(t, intSystemHttpClient, testConfig.ORDServiceURL+"/consumptionBundles", map[string][]string{acceptHeader: {"application/json"}, tenantHeader: {defaultTestTenant}})
	})

	t.Run("Requesting Packages returns empty", func(t *testing.T) {
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, fmt.Sprintf("%s/packages?$expand=apis,events&$format=json", testConfig.ORDServiceURL), map[string][]string{tenantHeader: {defaultTestTenant}})
		require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
	})

	t.Run("Requesting filtering of Bundles that do not have only ODATA APIs", func(t *testing.T) {
		escapedFilterValue := urlpkg.PathEscape("apis/any(d:d/apiProtocol ne 'odata-v2')")

		respBody := makeRequestWithHeaders(t, intSystemHttpClient, fmt.Sprintf("%s/consumptionBundles?$filter=%s&$expand=apis&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {tenantAPIProtocolFiltering}})

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
		escapedFilterValue := urlpkg.PathEscape("apis/all(d:d/apiProtocol eq 'odata-v2')")

		respBody := makeRequestWithHeaders(t, intSystemHttpClient, fmt.Sprintf("%s/consumptionBundles?$filter=%s&$expand=apis&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {tenantAPIProtocolFiltering}})

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
		t.Run(fmt.Sprintf("Requesting %s returns empty", resource), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, intSystemHttpClient, fmt.Sprintf("%s/%s?$format=json", testConfig.ORDServiceURL, resource), map[string][]string{tenantHeader: {defaultTestTenant}})
			require.True(t, gjson.Get(respBody, "value").Exists())
			require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
		})
	}

	// create label definition
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantFilteringTenant, []string{"DEFAULT", scenarioName})
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantFilteringTenant, []string{"DEFAULT"})

	// create automatic scenario assigment for subTenant
	asaInput := fixtures.FixAutomaticScenarioAssigmentInput(scenarioName, selectorKey, subTenantID)
	fixtures.CreateAutomaticScenarioAssignmentInTenant(t, ctx, dexGraphQLClient, asaInput, tenantFilteringTenant)
	defer fixtures.DeleteAutomaticScenarioAssignmentForScenarioWithinTenant(t, ctx, dexGraphQLClient, tenantFilteringTenant, scenarioName)

	// assert no system instances are visible without formation
	respBody := makeRequestWithHeaders(t, intSystemHttpClient, testConfig.ORDServiceURL+"/systemInstances?$format=json", map[string][]string{tenantHeader: {subTenantID}})
	require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))

	// assign application to scenario
	appLabelRequest := fixtures.FixSetApplicationLabelRequest(appInScenario.ID, scenariosLabel, []string{scenarioName})
	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantFilteringTenant, appLabelRequest, nil))
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, dexGraphQLClient, tenantFilteringTenant, appInScenario.ID, testConfig.DefaultScenarioEnabled)

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
			url:       testConfig.ORDServiceURL,
		},
		{
			msg:       secondaryTenant + " as integration system",
			headers:   map[string][]string{tenantHeader: {secondaryTenant}},
			appInput:  appInput2,
			apisMap:   apisMap2,
			eventsMap: eventsMap2,
			client:    intSystemHttpClient,
			url:       testConfig.ORDServiceURL,
		},
		{
			msg:       subTenantID + " as integration system",
			headers:   map[string][]string{tenantHeader: {subTenantID}},
			appInput:  appInputInScenario,
			apisMap:   apisMapInScenario,
			eventsMap: eventsMapInScenario,
			client:    intSystemHttpClient,
			url:       testConfig.ORDServiceURL,
		},
		{
			msg:       subTenantID + " as Runtime using externally issued certificate",
			headers:   map[string][]string{}, // The tenant comes from the certificate
			appInput:  appInputInScenario,
			apisMap:   apisMapInScenario,
			eventsMap: eventsMapInScenario,
			client:    extIssuerCertHttpClient,
			url:       testConfig.ORDExternalCertSecuredServiceURL,
		},
	} {

		t.Run(fmt.Sprintf("Requesting System Instances for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, testData.client, testData.url+"/systemInstances?$format=json", testData.headers)

			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, *testData.appInput.Description, gjson.Get(respBody, "value.0.description").String())
		})

		t.Run(fmt.Sprintf("Requesting Bundles for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, testData.client, testData.url+"/consumptionBundles?$format=json", testData.headers)

			require.Equal(t, len(testData.appInput.Bundles), len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Bundles[0].Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, *testData.appInput.Bundles[0].Description, gjson.Get(respBody, "value.0.description").String())
		})

		t.Run(fmt.Sprintf("Requesting APIs and their specs for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, testData.client, testData.url+"/apis?$format=json", testData.headers)

			require.Equal(t, len(testData.appInput.Bundles[0].APIDefinitions), len(gjson.Get(respBody, "value").Array()))

			for i := range testData.appInput.Bundles[0].APIDefinitions {
				name := gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String()
				require.NotEmpty(t, name)

				expectedAPI, exists := testData.apisMap[name]
				require.True(t, exists)

				require.Equal(t, *expectedAPI.Description, gjson.Get(respBody, fmt.Sprintf("value.%d.description", i)).String())
				require.Equal(t, expectedAPI.TargetURL, gjson.Get(respBody, fmt.Sprintf("value.%d.entryPoints.0.value", i)).String())
				require.NotEmpty(t, gjson.Get(respBody, fmt.Sprintf("value.%d.partOfConsumptionBundles", i)).String())

				releaseStatus := gjson.Get(respBody, fmt.Sprintf("value.%d.releaseStatus", i)).String()
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

				apiProtocol := gjson.Get(respBody, fmt.Sprintf("value.%d.apiProtocol", i)).String()
				switch apiProtocol {
				case "odata-v2":
					require.Equal(t, expectedAPI.Spec.Type, directorSchema.APISpecTypeOdata)
				case "rest":
					require.Equal(t, expectedAPI.Spec.Type, directorSchema.APISpecTypeOpenAPI)
				default:
					t.Log(fmt.Sprintf("API Protocol for API %s is %s. It does not match a predefined spec type.", name, apiProtocol))
				}

				specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", i)).Array()
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

				apiID := gjson.Get(respBody, fmt.Sprintf("value.%d.id", i)).String()
				require.NotEmpty(t, apiID)

				specURL := specs[0].Get("url").String()
				specPath := fmt.Sprintf("/api/%s/specification", apiID)
				require.Contains(t, specURL, testConfig.ORDServiceStaticPrefix+specPath)

				respBody := makeRequestWithHeaders(t, testData.client, specURL, testData.headers)

				require.Equal(t, string(*expectedAPI.Spec.Data), respBody)
			}
		})

		t.Run(fmt.Sprintf("Requesting Events and their specs for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, testData.client, testData.url+"/events?$format=json", testData.headers)

			require.Equal(t, len(testData.appInput.Bundles[0].EventDefinitions), len(gjson.Get(respBody, "value").Array()))

			for i := range testData.appInput.Bundles[0].EventDefinitions {
				name := gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String()
				require.NotEmpty(t, name)

				expectedEvent, exists := testData.eventsMap[name]
				require.True(t, exists)

				require.Equal(t, *expectedEvent.Description, gjson.Get(respBody, fmt.Sprintf("value.%d.description", i)).String())
				require.NotEmpty(t, gjson.Get(respBody, fmt.Sprintf("value.%d.partOfConsumptionBundles", i)).String())

				releaseStatus := gjson.Get(respBody, fmt.Sprintf("value.%d.releaseStatus", i)).String()
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

				specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", i)).Array()
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

				eventID := gjson.Get(respBody, fmt.Sprintf("value.%d.id", i)).String()
				require.NotEmpty(t, eventID)

				specURL := specs[0].Get("url").String()
				specPath := fmt.Sprintf("/event/%s/specification", eventID)
				require.Contains(t, specURL, testConfig.ORDServiceStaticPrefix+specPath)

				respBody := makeRequestWithHeaders(t, testData.client, specURL, testData.headers)

				require.Equal(t, string(*expectedEvent.Spec.Data), respBody)
			}
		})

		// Paging:
		t.Run(fmt.Sprintf("Requesting paging of Bundles for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles)

			respBody := makeRequestWithHeaders(t, testData.client, testData.url+"/consumptionBundles?$top=10&$skip=0&$format=json", testData.headers)
			require.Equal(t, totalCount, len(gjson.Get(respBody, "value").Array()))

			respBody = makeRequestWithHeaders(t, testData.client, fmt.Sprintf("%s/consumptionBundles?$top=10&$skip=%d&$format=json", testData.url, totalCount), testData.headers)
			require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
		})

		t.Run(fmt.Sprintf("Requesting paging of Bundle APIs for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].APIDefinitions)

			respBody := makeRequestWithHeaders(t, testData.client, testData.url+"/consumptionBundles?$expand=apis($top=10;$skip=0)&$format=json", testData.headers)
			require.Equal(t, totalCount, len(gjson.Get(respBody, "value.0.apis").Array()))

			expectedItemCount := 1
			respBody = makeRequestWithHeaders(t, testData.client, fmt.Sprintf("%s/consumptionBundles?$expand=apis($top=10;$skip=%d)&$format=json", testData.url, totalCount-expectedItemCount), testData.headers)
			require.Equal(t, expectedItemCount, len(gjson.Get(respBody, "value").Array()))
		})

		t.Run(fmt.Sprintf("Requesting paging of Bundle Events for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].EventDefinitions)

			respBody := makeRequestWithHeaders(t, testData.client, testData.url+"/consumptionBundles?$expand=events($top=10;$skip=0)&$format=json", testData.headers)
			require.Equal(t, totalCount, len(gjson.Get(respBody, "value.0.events").Array()))

			expectedItemCount := 1
			respBody = makeRequestWithHeaders(t, testData.client, fmt.Sprintf("%s/consumptionBundles?$expand=events($top=10;$skip=%d)&$format=json", testData.url, totalCount-expectedItemCount), testData.headers)
			require.Equal(t, expectedItemCount, len(gjson.Get(respBody, "value").Array()))
		})

		// Filtering:
		t.Run(fmt.Sprintf("Requesting filtering of Bundles for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			bndlName := testData.appInput.Bundles[0].Name

			escapedFilterValue := urlpkg.PathEscape(fmt.Sprintf("title eq '%s'", bndlName))
			respBody := makeRequestWithHeaders(t, testData.client, fmt.Sprintf("%s/consumptionBundles?$filter=(%s)&$format=json", testData.url, escapedFilterValue), testData.headers)
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

			escapedFilterValue = urlpkg.PathEscape(fmt.Sprintf("title ne '%s'", bndlName))
			respBody = makeRequestWithHeaders(t, testData.client, fmt.Sprintf("%s/consumptionBundles?$filter=(%s)&$format=json", testData.url, escapedFilterValue), testData.headers)
			require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
		})

		t.Run(fmt.Sprintf("Requesting filtering of Bundle APIs for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].APIDefinitions)
			apiName := testData.appInput.Bundles[0].APIDefinitions[0].Name

			escapedFilterValue := urlpkg.PathEscape(fmt.Sprintf("title eq '%s'", apiName))
			respBody := makeRequestWithHeaders(t, testData.client, fmt.Sprintf("%s/consumptionBundles?$expand=apis($filter=(%s))&$format=json", testData.url, escapedFilterValue), testData.headers)
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

			escapedFilterValue = urlpkg.PathEscape(fmt.Sprintf("title ne '%s'", apiName))
			respBody = makeRequestWithHeaders(t, testData.client, fmt.Sprintf("%s/consumptionBundles?$expand=apis($filter=(%s))&$format=json", testData.url, escapedFilterValue), testData.headers)
			require.Equal(t, totalCount-1, len(gjson.Get(respBody, "value.0.apis").Array()))
		})

		t.Run(fmt.Sprintf("Requesting filtering of Bundle Events for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].EventDefinitions)
			eventName := testData.appInput.Bundles[0].EventDefinitions[0].Name

			escapedFilterValue := urlpkg.PathEscape(fmt.Sprintf("title eq '%s'", eventName))
			respBody := makeRequestWithHeaders(t, testData.client, fmt.Sprintf("%s/consumptionBundles?$expand=events($filter=(%s))&$format=json", testData.url, escapedFilterValue), testData.headers)
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

			escapedFilterValue = urlpkg.PathEscape(fmt.Sprintf("title ne '%s'", eventName))
			respBody = makeRequestWithHeaders(t, testData.client, fmt.Sprintf("%s/consumptionBundles?$expand=events($filter=(%s))&$format=json", testData.url, escapedFilterValue), testData.headers)
			require.Equal(t, totalCount-1, len(gjson.Get(respBody, "value.0.events").Array()))
		})

		// Projection:
		t.Run(fmt.Sprintf("Requesting projection of Bundles for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, testData.client, testData.url+"/consumptionBundles?$select=title&$format=json", testData.headers)
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Bundles[0].Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, false, gjson.Get(respBody, "value.0.description").Exists())
		})

		t.Run(fmt.Sprintf("Requesting projection of Bundle APIs for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, testData.client, testData.url+"/consumptionBundles?$expand=apis($select=title)&$format=json", testData.headers)

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
			respBody := makeRequestWithHeaders(t, testData.client, testData.url+"/consumptionBundles?$expand=events($select=title)&$format=json", testData.headers)

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
			escapedOrderByValue := urlpkg.PathEscape("title asc,description desc")
			respBody := makeRequestWithHeaders(t, testData.client, fmt.Sprintf("%s/consumptionBundles?$orderby=%s&$format=json", testData.url, escapedOrderByValue), testData.headers)
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Bundles[0].Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, *testData.appInput.Bundles[0].Description, gjson.Get(respBody, "value.0.description").String())
		})

		t.Run(fmt.Sprintf("Requesting ordering of Bundle APIs for tenant %s returns them as expected", testData.msg), func(t *testing.T) {
			escapedOrderByValue := urlpkg.PathEscape("title asc,description desc")
			respBody := makeRequestWithHeaders(t, testData.client, fmt.Sprintf("%s/consumptionBundles?$expand=apis($orderby=%s)&$format=json", testData.url, escapedOrderByValue), testData.headers)

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
			escapedOrderByValue := urlpkg.PathEscape("title asc,description desc")
			respBody := makeRequestWithHeaders(t, testData.client, fmt.Sprintf("%s/consumptionBundles?$expand=events($orderby=%s)&$format=json", testData.url, escapedOrderByValue), testData.headers)

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
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, testConfig.ORDServiceURL+"/apis?$format=json", map[string][]string{tenantHeader: {defaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].APIDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()

		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, specURL, map[string][]string{tenantHeader: {secondaryTenant}}, http.StatusNotFound, testConfig.ORDServiceDefaultResponseType)
		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, specURL, map[string][]string{tenantHeader: {defaultTestTenant}}, http.StatusOK, testConfig.ORDServiceDefaultResponseType)
	})

	t.Run("404 when request to ORD Service for event spec have another tenant header value", func(t *testing.T) {
		respBody := makeRequestWithHeaders(t, intSystemHttpClient, testConfig.ORDServiceURL+"/events?$format=json", map[string][]string{tenantHeader: {defaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].EventDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()

		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, specURL, map[string][]string{tenantHeader: {secondaryTenant}}, http.StatusNotFound, testConfig.ORDServiceDefaultResponseType)
		request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, specURL, map[string][]string{tenantHeader: {defaultTestTenant}}, http.StatusOK, testConfig.ORDServiceDefaultResponseType)
	})

	t.Run("Errors generate user-friendly message", func(t *testing.T) {
		respBody := request.MakeRequestWithHeadersAndStatusExpect(t, intSystemHttpClient, testConfig.ORDServiceURL+"/test?$format=json", map[string][]string{tenantHeader: {defaultTestTenant}}, http.StatusNotFound, testConfig.ORDServiceDefaultResponseType)

		require.Contains(t, gjson.Get(respBody, "error.message").String(), "Use odata-debug query parameter with value one of the following formats: json,html,download for more information")
	})

	t.Run("Additional non-ORD details about system instances are exposed", func(t *testing.T) {
		expectedProductType := "productType"
		appTmplInput := fixtures.FixApplicationTemplate(expectedProductType)
		placeholderKey := "new-placeholder"
		appTmplInput.ApplicationInput.Description = ptr.String("test {{new-placeholder}}")
		appTmplInput.Placeholders = []*directorSchema.PlaceholderDefinitionInput{
			{
				Name:        placeholderKey,
				Description: ptr.String("description"),
			},
		}

		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, defaultTestTenant, appTmplInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, dexGraphQLClient, defaultTestTenant, &appTmpl)
		require.NoError(t, err)

		appFromTmpl := directorSchema.ApplicationFromTemplateInput{
			TemplateName: expectedProductType, Values: []*directorSchema.TemplateValueInput{
				{
					Placeholder: placeholderKey,
					Value:       "new-value",
				},
			},
		}

		appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
		require.NoError(t, err)

		createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
		outputApp := directorSchema.ApplicationExt{}
		//WHEN
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, defaultTestTenant, createAppFromTmplRequest, &outputApp)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, defaultTestTenant, &outputApp)
		require.NoError(t, err)

		getSystemInstanceURL := fmt.Sprintf("%s/systemInstances(%s)?$format=json", testConfig.ORDServiceURL, outputApp.ID)

		respBody := makeRequestWithHeaders(t, intSystemHttpClient, getSystemInstanceURL, map[string][]string{tenantHeader: {defaultTestTenant}})

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

func makeRequestWithHeaders(t require.TestingT, httpClient *http.Client, url string, headers map[string][]string) string {
	return request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, url, headers, http.StatusOK, testConfig.ORDServiceDefaultResponseType)
}

func makeRequestWithStatusExpect(t require.TestingT, httpClient *http.Client, url string, expectedHTTPStatus int) string {
	return request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, url, map[string][]string{}, expectedHTTPStatus, testConfig.ORDServiceDefaultResponseType)
}

// extIssuerCertClient returns http client configured with client certificate manually signed by connector's CA
// and a subject matching external issuer's subject contract.
func extIssuerCertClient(t require.TestingT, subTenantID string) *http.Client {
	clientKey, certChain := certs.IssueExternalIssuerCertificate(t, testConfig.CA.Certificate, testConfig.CA.Key, subTenantID)

	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{
					{
						Certificate: certChain,
						PrivateKey:  clientKey,
					},
				},
				ClientAuth:         tls.RequireAndVerifyClientCert,
				InsecureSkipVerify: true,
			},
		},
	}
}
