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

	"github.com/kyma-incubator/compass/tests/pkg/request"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	acceptHeader = "Accept"
	tenantHeader = "Tenant"
)

func TestORDService(t *testing.T) {
	// Cannot use tenant constants as the names become too long and cannot be inserted
	appInput := fixtures.CreateApp("tenant1")
	appInput2 := fixtures.CreateApp("tenant2")

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

	ctx := context.Background()

	app, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, appInput)
	require.NoError(t, err)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTestTenant, app.ID)

	app2, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, testConfig.SecondaryTenant, appInput2)
	require.NoError(t, err)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, testConfig.SecondaryTenant, app2.ID)

	t.Log("Create integration system")
	intSys := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, "", "test-int-system")
	require.NotEmpty(t, intSys)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, "", intSys.ID)

	intSystemCredentials := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, "", intSys.ID)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, dexGraphQLClient, intSystemCredentials.ID)

	unsecuredHttpClient := http.DefaultClient
	unsecuredHttpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	oauthCredentialData, ok := intSystemCredentials.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	conf := &clientcredentials.Config{
		ClientID:     oauthCredentialData.ClientID,
		ClientSecret: oauthCredentialData.ClientSecret,
		TokenURL:     oauthCredentialData.URL,
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, unsecuredHttpClient)
	httpClient := conf.Client(ctx)
	httpClient.Timeout = 10 * time.Second

	t.Run("401 when requests to ORD Service are unsecured", func(t *testing.T) {
		makeRequestWithStatusExpect(t, unsecuredHttpClient, testConfig.ORDServiceURL+"/$metadata?$format=json", http.StatusUnauthorized)
	})

	t.Run("400 when requests to ORD Service do not have tenant header", func(t *testing.T) {
		makeRequestWithStatusExpect(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$format=json", http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service have wrong tenant header", func(t *testing.T) {
		request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$format=json", map[string][]string{tenantHeader: {"wrong-tenant"}}, http.StatusBadRequest, testConfig.ORDServiceDefaultResponseType)
	})

	t.Run("400 when requests to ORD Service api specification do not have tenant header", func(t *testing.T) {
		respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/apis?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].APIDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()
		makeRequestWithStatusExpect(t, httpClient, specURL, http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service event specification do not have tenant header", func(t *testing.T) {
		respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/events?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].EventDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()
		makeRequestWithStatusExpect(t, httpClient, specURL, http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service api specification have wrong tenant header", func(t *testing.T) {
		respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/apis?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].APIDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()
		request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, specURL, map[string][]string{tenantHeader: {"wrong-tenant"}}, http.StatusBadRequest, testConfig.ORDServiceDefaultResponseType)
	})

	t.Run("400 when requests to ORD Service event specification have wrong tenant header", func(t *testing.T) {
		respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/events?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].EventDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()
		request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, specURL, map[string][]string{tenantHeader: {"wrong-tenant"}}, http.StatusBadRequest, testConfig.ORDServiceDefaultResponseType)
	})

	t.Run("Requesting entities without specifying response format falls back to configured default response type when Accept header allows everything", func(t *testing.T) {
		makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles", map[string][]string{acceptHeader: {"*/*"}, tenantHeader: {testConfig.DefaultTestTenant}})
	})

	t.Run("Requesting entities without specifying response format falls back to response type specified by Accept header when it provides a specific type", func(t *testing.T) {
		makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles", map[string][]string{acceptHeader: {"application/json"}, tenantHeader: {testConfig.DefaultTestTenant}})
	})

	t.Run("Requesting Packages returns empty", func(t *testing.T) {
		respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/packages?$expand=apis,events&$format=json", testConfig.ORDServiceURL), map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
		require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
	})

	for _, resource := range []string{"vendors", "tombstones", "products"} { // This tests assert integrity between ORD Service JPA model and our Database model
		t.Run(fmt.Sprintf("Requesting %s returns empty", resource), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/%s?$format=json", testConfig.ORDServiceURL, resource), map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
			require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
		})
	}

	for _, testData := range []struct {
		tenant    string
		appInput  directorSchema.ApplicationRegisterInput
		apisMap   map[string]directorSchema.APIDefinitionInput
		eventsMap map[string]directorSchema.EventDefinitionInput
	}{
		{
			tenant:    testConfig.DefaultTestTenant,
			appInput:  appInput,
			apisMap:   apisMap,
			eventsMap: eventsMap,
		}, {
			tenant:    testConfig.SecondaryTenant,
			appInput:  appInput2,
			apisMap:   apisMap2,
			eventsMap: eventsMap2,
		},
	} {

		t.Run(fmt.Sprintf("Requesting System Instances for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/systemInstances?$format=json", map[string][]string{tenantHeader: {testData.tenant}})

			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, *testData.appInput.Description, gjson.Get(respBody, "value.0.description").String())
		})

		t.Run(fmt.Sprintf("Requesting Bundles for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$format=json", map[string][]string{tenantHeader: {testData.tenant}})

			require.Equal(t, len(testData.appInput.Bundles), len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Bundles[0].Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, *testData.appInput.Bundles[0].Description, gjson.Get(respBody, "value.0.description").String())
		})

		t.Run(fmt.Sprintf("Requesting APIs and their specs for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/apis?$format=json", map[string][]string{tenantHeader: {testData.tenant}})

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
				require.Contains(t, specURL, testConfig.ORDServiceStaticURL+specPath)

				respBody := makeRequestWithHeaders(t, httpClient, specURL, map[string][]string{tenantHeader: {testData.tenant}})

				require.Equal(t, string(*expectedAPI.Spec.Data), respBody)
			}
		})

		t.Run(fmt.Sprintf("Requesting Events and their specs for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/events?$format=json", map[string][]string{tenantHeader: {testData.tenant}})

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
				require.Contains(t, specURL, testConfig.ORDServiceStaticURL+specPath)

				respBody := makeRequestWithHeaders(t, httpClient, specURL, map[string][]string{tenantHeader: {testData.tenant}})

				require.Equal(t, string(*expectedEvent.Spec.Data), respBody)
			}
		})

		// Paging:
		t.Run(fmt.Sprintf("Requesting paging of Bundles for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles)

			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$top=10&$skip=0&$format=json", map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, totalCount, len(gjson.Get(respBody, "value").Array()))

			respBody = makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/consumptionBundles?$top=10&$skip=%d&$format=json", testConfig.ORDServiceURL, totalCount), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
		})

		t.Run(fmt.Sprintf("Requesting paging of Bundle APIs for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].APIDefinitions)

			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$expand=apis($top=10;$skip=0)&$format=json", map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, totalCount, len(gjson.Get(respBody, "value.0.apis").Array()))

			expectedItemCount := 1
			respBody = makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/consumptionBundles?$expand=apis($top=10;$skip=%d)&$format=json", testConfig.ORDServiceURL, totalCount-expectedItemCount), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, expectedItemCount, len(gjson.Get(respBody, "value").Array()))
		})

		t.Run(fmt.Sprintf("Requesting paging of Bundle Events for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].EventDefinitions)

			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$expand=events($top=10;$skip=0)&$format=json", map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, totalCount, len(gjson.Get(respBody, "value.0.events").Array()))

			expectedItemCount := 1
			respBody = makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/consumptionBundles?$expand=events($top=10;$skip=%d)&$format=json", testConfig.ORDServiceURL, totalCount-expectedItemCount), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, expectedItemCount, len(gjson.Get(respBody, "value").Array()))
		})

		// Filtering:
		t.Run(fmt.Sprintf("Requesting filtering of Bundles for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			bndlName := testData.appInput.Bundles[0].Name

			escapedFilterValue := urlpkg.PathEscape(fmt.Sprintf("title eq '%s'", bndlName))
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/consumptionBundles?$filter=(%s)&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

			escapedFilterValue = urlpkg.PathEscape(fmt.Sprintf("title ne '%s'", bndlName))
			respBody = makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/consumptionBundles?$filter=(%s)&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
		})

		t.Run(fmt.Sprintf("Requesting filtering of Bundle APIs for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].APIDefinitions)
			apiName := testData.appInput.Bundles[0].APIDefinitions[0].Name

			escapedFilterValue := urlpkg.PathEscape(fmt.Sprintf("title eq '%s'", apiName))
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/consumptionBundles?$expand=apis($filter=(%s))&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

			escapedFilterValue = urlpkg.PathEscape(fmt.Sprintf("title ne '%s'", apiName))
			respBody = makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/consumptionBundles?$expand=apis($filter=(%s))&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, totalCount-1, len(gjson.Get(respBody, "value.0.apis").Array()))
		})

		t.Run(fmt.Sprintf("Requesting filtering of Bundle Events for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].EventDefinitions)
			eventName := testData.appInput.Bundles[0].EventDefinitions[0].Name

			escapedFilterValue := urlpkg.PathEscape(fmt.Sprintf("title eq '%s'", eventName))
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/consumptionBundles?$expand=events($filter=(%s))&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

			escapedFilterValue = urlpkg.PathEscape(fmt.Sprintf("title ne '%s'", eventName))
			respBody = makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/consumptionBundles?$expand=events($filter=(%s))&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, totalCount-1, len(gjson.Get(respBody, "value.0.events").Array()))
		})

		// Projection:
		t.Run(fmt.Sprintf("Requesting projection of Bundles for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$select=title&$format=json", map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Bundles[0].Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, false, gjson.Get(respBody, "value.0.description").Exists())
		})

		t.Run(fmt.Sprintf("Requesting projection of Bundle APIs for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$expand=apis($select=title)&$format=json", map[string][]string{tenantHeader: {testData.tenant}})

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

		t.Run(fmt.Sprintf("Requesting projection of Bundle Events for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/consumptionBundles?$expand=events($select=title)&$format=json", map[string][]string{tenantHeader: {testData.tenant}})

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
		t.Run(fmt.Sprintf("Requesting ordering of Bundles for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			escapedOrderByValue := urlpkg.PathEscape("title asc,description desc")
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/consumptionBundles?$orderby=%s&$format=json", testConfig.ORDServiceURL, escapedOrderByValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Bundles[0].Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, *testData.appInput.Bundles[0].Description, gjson.Get(respBody, "value.0.description").String())
		})

		t.Run(fmt.Sprintf("Requesting ordering of Bundle APIs for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			escapedOrderByValue := urlpkg.PathEscape("title asc,description desc")
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/consumptionBundles?$expand=apis($orderby=%s)&$format=json", testConfig.ORDServiceURL, escapedOrderByValue), map[string][]string{tenantHeader: {testData.tenant}})

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

		t.Run(fmt.Sprintf("Requesting ordering of Bundle Events for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			escapedOrderByValue := urlpkg.PathEscape("title asc,description desc")
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/consumptionBundles?$expand=events($orderby=%s)&$format=json", testConfig.ORDServiceURL, escapedOrderByValue), map[string][]string{tenantHeader: {testData.tenant}})

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
		respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/apis?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].APIDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()

		request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, specURL, map[string][]string{tenantHeader: {testConfig.SecondaryTenant}}, http.StatusNotFound, testConfig.ORDServiceDefaultResponseType)
		request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, specURL, map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}}, http.StatusOK, testConfig.ORDServiceDefaultResponseType)
	})

	t.Run("404 when request to ORD Service for event spec have another tenant header value", func(t *testing.T) {
		respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/events?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}})
		require.Equal(t, len(appInput.Bundles[0].EventDefinitions), len(gjson.Get(respBody, "value").Array()))

		specs := gjson.Get(respBody, fmt.Sprintf("value.%d.resourceDefinitions", 0)).Array()
		require.Equal(t, 1, len(specs))

		specURL := specs[0].Get("url").String()

		request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, specURL, map[string][]string{tenantHeader: {testConfig.SecondaryTenant}}, http.StatusNotFound, testConfig.ORDServiceDefaultResponseType)
		request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, specURL, map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}}, http.StatusOK, testConfig.ORDServiceDefaultResponseType)
	})

	t.Run("Errors generate user-friendly message", func(t *testing.T) {
		respBody := request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, testConfig.ORDServiceURL+"/test?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTestTenant}}, http.StatusNotFound, testConfig.ORDServiceDefaultResponseType)

		require.Contains(t, gjson.Get(respBody, "error.message").String(), "Use odata-debug query parameter with value one of the following formats: json,html,download for more information")
	})
}

func makeRequest(t *testing.T, httpClient *http.Client, url string) string {
	return request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, url, map[string][]string{}, http.StatusOK, testConfig.ORDServiceDefaultResponseType)
}

func makeRequestWithHeaders(t *testing.T, httpClient *http.Client, url string, headers map[string][]string) string {
	return request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, url, headers, http.StatusOK, testConfig.ORDServiceDefaultResponseType)
}

func makeRequestWithStatusExpect(t *testing.T, httpClient *http.Client, url string, expectedHTTPStatus int) string {
	return request.MakeRequestWithHeadersAndStatusExpect(t, httpClient, url, map[string][]string{}, expectedHTTPStatus, testConfig.ORDServiceDefaultResponseType)
}
