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
	"io/ioutil"
	"net/http"
	urlpkg "net/url"
	"strings"
	"testing"
	"time"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/ord-service/pkg"
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
	appInput := createApp("tenant1")
	appInput2 := createApp("tenant2")

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

	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	app, err := pkg.RegisterApplicationWithinTenant(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, appInput)
	require.NoError(t, err)
	apiDefinitionIDDefaultTenant := app.Bundles.Data[0].APIDefinitions.Data[0].ID
	eventDefinitionIDDefaultTenant := app.Bundles.Data[0].EventDefinitions.Data[0].ID

	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)

	app2, err := pkg.RegisterApplicationWithinTenant(t, ctx, dexGraphQLClient, testConfig.Tenant, appInput2)
	require.NoError(t, err)

	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, testConfig.Tenant, app2.ID)

	t.Log("Create integration system")
	intSys := pkg.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, "test-int-system")
	require.NotEmpty(t, intSys)
	defer pkg.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, intSys.ID)

	intSystemCredentials := pkg.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, intSys.ID)
	defer pkg.DeleteSystemAuthForIntegrationSystem(t, ctx, dexGraphQLClient, intSystemCredentials.ID)

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
		makeRequestWithStatusExpect(t, httpClient, testConfig.ORDServiceURL+"/bundles?$format=json", http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service have wrong tenant header", func(t *testing.T) {
		makeRequestWithHeadersAndStatusExpect(t, httpClient, testConfig.ORDServiceURL+"/bundles?$format=json", map[string][]string{tenantHeader: {"wrong-tenant"}}, http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service api specification do not have tenant header", func(t *testing.T) {
		makeRequestWithStatusExpect(t, httpClient, fmt.Sprintf(testConfig.ORDServiceStaticURL+"/api/%s/specification", apiDefinitionIDDefaultTenant), http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service event specification do not have tenant header", func(t *testing.T) {
		makeRequestWithStatusExpect(t, httpClient, fmt.Sprintf(testConfig.ORDServiceStaticURL+"/event/%s/specification", eventDefinitionIDDefaultTenant), http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service api specification have wrong tenant header", func(t *testing.T) {
		makeRequestWithHeadersAndStatusExpect(t, httpClient, fmt.Sprintf(testConfig.ORDServiceStaticURL+"/api/%s/specification", apiDefinitionIDDefaultTenant), map[string][]string{tenantHeader: {"wrong-tenant"}}, http.StatusBadRequest)
	})

	t.Run("400 when requests to ORD Service event specification have wrong tenant header", func(t *testing.T) {
		makeRequestWithHeadersAndStatusExpect(t, httpClient, fmt.Sprintf(testConfig.ORDServiceStaticURL+"/event/%s/specification", eventDefinitionIDDefaultTenant), map[string][]string{tenantHeader: {"wrong-tenant"}}, http.StatusBadRequest)
	})

	t.Run("Requesting entities without specifying response format falls back to configured default response type when Accept header allows everything", func(t *testing.T) {
		makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/bundles", map[string][]string{acceptHeader: {"*/*"}, tenantHeader: {testConfig.DefaultTenant}})
	})

	t.Run("Requesting entities without specifying response format falls back to response type specified by Accept header when it provides a specific type", func(t *testing.T) {
		makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/bundles", map[string][]string{acceptHeader: {"application/json"}, tenantHeader: {testConfig.DefaultTenant}})
	})

	t.Run("Requesting Packages returns empty", func(t *testing.T) {
		respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/packages?$expand=apis,events&$format=json", testConfig.ORDServiceURL), map[string][]string{tenantHeader: {testConfig.DefaultTenant}})
		require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
	})

	for _, testData := range []struct {
		tenant    string
		appInput  directorSchema.ApplicationRegisterInput
		apisMap   map[string]directorSchema.APIDefinitionInput
		eventsMap map[string]directorSchema.EventDefinitionInput
	}{
		{
			tenant:    testConfig.DefaultTenant,
			appInput:  appInput,
			apisMap:   apisMap,
			eventsMap: eventsMap,
		}, {
			tenant:    testConfig.Tenant,
			appInput:  appInput2,
			apisMap:   apisMap2,
			eventsMap: eventsMap2,
		},
	} {
		t.Run(fmt.Sprintf("Requesting Bundles for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/bundles?$format=json", map[string][]string{tenantHeader: {testData.tenant}})

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
				require.Equal(t, expectedAPI.TargetURL, gjson.Get(respBody, fmt.Sprintf("value.%d.entryPoint", i)).String())
				require.NotEmpty(t, gjson.Get(respBody, fmt.Sprintf("value.%d.partOfConsumptionBundle", i)).String())

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

				specs := gjson.Get(respBody, fmt.Sprintf("value.%d.apiDefinitions", i)).Array()
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
				require.Equal(t, testConfig.ORDServiceStaticURL+specPath, specURL)

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
				require.NotEmpty(t, gjson.Get(respBody, fmt.Sprintf("value.%d.partOfConsumptionBundle", i)).String())

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

				specs := gjson.Get(respBody, fmt.Sprintf("value.%d.eventDefinitions", i)).Array()
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
				require.Equal(t, testConfig.ORDServiceStaticURL+specPath, specURL)

				respBody := makeRequestWithHeaders(t, httpClient, specURL, map[string][]string{tenantHeader: {testData.tenant}})

				require.Equal(t, string(*expectedEvent.Spec.Data), respBody)
			}
		})

		// Paging:
		t.Run(fmt.Sprintf("Requesting paging of Bundles for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles)

			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/bundles?$top=10&$skip=0&$format=json", map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, totalCount, len(gjson.Get(respBody, "value").Array()))

			respBody = makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/bundles?$top=10&$skip=%d&$format=json", testConfig.ORDServiceURL, totalCount), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
		})

		t.Run(fmt.Sprintf("Requesting paging of Bundle APIs for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].APIDefinitions)

			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/bundles?$expand=apis($top=10;$skip=0)&$format=json", map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, totalCount, len(gjson.Get(respBody, "value.0.apis").Array()))

			expectedItemCount := 1
			respBody = makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/bundles?$expand=apis($top=10;$skip=%d)&$format=json", testConfig.ORDServiceURL, totalCount-expectedItemCount), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, expectedItemCount, len(gjson.Get(respBody, "value").Array()))
		})

		t.Run(fmt.Sprintf("Requesting paging of Bundle Events for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].EventDefinitions)

			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/bundles?$expand=events($top=10;$skip=0)&$format=json", map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, totalCount, len(gjson.Get(respBody, "value.0.events").Array()))

			expectedItemCount := 1
			respBody = makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/bundles?$expand=events($top=10;$skip=%d)&$format=json", testConfig.ORDServiceURL, totalCount-expectedItemCount), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, expectedItemCount, len(gjson.Get(respBody, "value").Array()))
		})

		// Filtering:
		t.Run(fmt.Sprintf("Requesting filtering of Bundles for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			bndlName := testData.appInput.Bundles[0].Name

			escapedFilterValue := urlpkg.PathEscape(fmt.Sprintf("title eq '%s'", bndlName))
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/bundles?$filter=(%s)&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

			escapedFilterValue = urlpkg.PathEscape(fmt.Sprintf("title ne '%s'", bndlName))
			respBody = makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/bundles?$filter=(%s)&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
		})

		t.Run(fmt.Sprintf("Requesting filtering of Bundle APIs for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].APIDefinitions)
			apiName := testData.appInput.Bundles[0].APIDefinitions[0].Name

			escapedFilterValue := urlpkg.PathEscape(fmt.Sprintf("title eq '%s'", apiName))
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/bundles?$expand=apis($filter=(%s))&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

			escapedFilterValue = urlpkg.PathEscape(fmt.Sprintf("title ne '%s'", apiName))
			respBody = makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/bundles?$expand=apis($filter=(%s))&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, totalCount-1, len(gjson.Get(respBody, "value.0.apis").Array()))
		})

		t.Run(fmt.Sprintf("Requesting filtering of Bundle Events for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			totalCount := len(testData.appInput.Bundles[0].EventDefinitions)
			eventName := testData.appInput.Bundles[0].EventDefinitions[0].Name

			escapedFilterValue := urlpkg.PathEscape(fmt.Sprintf("title eq '%s'", eventName))
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/bundles?$expand=events($filter=(%s))&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

			escapedFilterValue = urlpkg.PathEscape(fmt.Sprintf("title ne '%s'", eventName))
			respBody = makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/bundles?$expand=events($filter=(%s))&$format=json", testConfig.ORDServiceURL, escapedFilterValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, totalCount-1, len(gjson.Get(respBody, "value.0.events").Array()))
		})

		// Projection:
		t.Run(fmt.Sprintf("Requesting projection of Bundles for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/bundles?$select=title&$format=json", map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Bundles[0].Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, false, gjson.Get(respBody, "value.0.description").Exists())
		})

		t.Run(fmt.Sprintf("Requesting projection of Bundle APIs for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/bundles?$expand=apis($select=title)&$format=json", map[string][]string{tenantHeader: {testData.tenant}})

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
			respBody := makeRequestWithHeaders(t, httpClient, testConfig.ORDServiceURL+"/bundles?$expand=events($select=title)&$format=json", map[string][]string{tenantHeader: {testData.tenant}})

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
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/bundles?$orderby=%s&$format=json", testConfig.ORDServiceURL, escapedOrderByValue), map[string][]string{tenantHeader: {testData.tenant}})
			require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
			require.Equal(t, testData.appInput.Bundles[0].Name, gjson.Get(respBody, "value.0.title").String())
			require.Equal(t, *testData.appInput.Bundles[0].Description, gjson.Get(respBody, "value.0.description").String())
		})

		t.Run(fmt.Sprintf("Requesting ordering of Bundle APIs for tenant %s returns them as expected", testData.tenant), func(t *testing.T) {
			escapedOrderByValue := urlpkg.PathEscape("title asc,description desc")
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/bundles?$expand=apis($orderby=%s)&$format=json", testConfig.ORDServiceURL, escapedOrderByValue), map[string][]string{tenantHeader: {testData.tenant}})

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
			respBody := makeRequestWithHeaders(t, httpClient, fmt.Sprintf("%s/bundles?$expand=events($orderby=%s)&$format=json", testConfig.ORDServiceURL, escapedOrderByValue), map[string][]string{tenantHeader: {testData.tenant}})

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
		makeRequestWithHeadersAndStatusExpect(t, httpClient, fmt.Sprintf(testConfig.ORDServiceStaticURL+"/api/%s/specification", apiDefinitionIDDefaultTenant), map[string][]string{tenantHeader: {testConfig.Tenant}}, http.StatusNotFound)
		makeRequestWithHeadersAndStatusExpect(t, httpClient, fmt.Sprintf(testConfig.ORDServiceStaticURL+"/api/%s/specification", apiDefinitionIDDefaultTenant), map[string][]string{tenantHeader: {testConfig.DefaultTenant}}, http.StatusOK)
	})

	t.Run("404 when request to ORD Service for event spec have another tenant header value", func(t *testing.T) {
		makeRequestWithHeadersAndStatusExpect(t, httpClient, fmt.Sprintf(testConfig.ORDServiceStaticURL+"/event/%s/specification", eventDefinitionIDDefaultTenant), map[string][]string{tenantHeader: {testConfig.Tenant}}, http.StatusNotFound)
		makeRequestWithHeadersAndStatusExpect(t, httpClient, fmt.Sprintf(testConfig.ORDServiceStaticURL+"/event/%s/specification", eventDefinitionIDDefaultTenant), map[string][]string{tenantHeader: {testConfig.DefaultTenant}}, http.StatusOK)
	})

	t.Run("Errors generate user-friendly message", func(t *testing.T) {
		respBody := makeRequestWithHeadersAndStatusExpect(t, httpClient, testConfig.ORDServiceURL+"/test?$format=json", map[string][]string{tenantHeader: {testConfig.DefaultTenant}}, http.StatusNotFound)

		require.Contains(t, gjson.Get(respBody, "error.message").String(), "Use odata-debug query parameter with value one of the following formats: json,html,download for more information")
	})
}

func makeRequest(t *testing.T, httpClient *http.Client, url string) string {
	return makeRequestWithHeadersAndStatusExpect(t, httpClient, url, map[string][]string{}, http.StatusOK)
}

func makeRequestWithHeaders(t *testing.T, httpClient *http.Client, url string, headers map[string][]string) string {
	return makeRequestWithHeadersAndStatusExpect(t, httpClient, url, headers, http.StatusOK)
}

func makeRequestWithStatusExpect(t *testing.T, httpClient *http.Client, url string, expectedHTTPStatus int) string {
	return makeRequestWithHeadersAndStatusExpect(t, httpClient, url, map[string][]string{}, expectedHTTPStatus)
}

func makeRequestWithHeadersAndStatusExpect(t *testing.T, httpClient *http.Client, url string, headers map[string][]string, expectedHTTPStatus int) string {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	for key, values := range headers {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}

	response, err := httpClient.Do(request)

	require.NoError(t, err)
	require.Equal(t, expectedHTTPStatus, response.StatusCode)

	parsedURL, err := urlpkg.Parse(url)
	require.NoError(t, err)

	if !strings.Contains(parsedURL.Path, "/specification") {
		formatParam := parsedURL.Query().Get("$format")
		acceptHeader, acceptHeaderProvided := headers[acceptHeader]

		contentType := response.Header.Get("Content-Type")
		if formatParam != "" {
			require.Contains(t, contentType, formatParam)
		} else if acceptHeaderProvided && acceptHeader[0] != "*/*" {
			require.Contains(t, contentType, acceptHeader[0])
		} else {
			require.Contains(t, contentType, testConfig.ORDServiceDefaultResponseType)
		}
	}

	body, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err)

	return string(body)
}

func createApp(suffix string) directorSchema.ApplicationRegisterInput {
	return generateAppInputForDifferentTenants(directorSchema.ApplicationRegisterInput{
		Name:        "test-app",
		Description: ptr.String("my application"),
		Bundles: []*directorSchema.BundleCreateInput{
			{
				Name:        "foo-bndl",
				Description: ptr.String("foo-descr"),
				APIDefinitions: []*directorSchema.APIDefinitionInput{
					{
						Name:        "comments-v1",
						Description: ptr.String("api for adding comments"),
						TargetURL:   "http://mywordpress.com/comments",
						Group:       ptr.String("comments"),
						Version:     pkg.FixDepracatedVersion(),
						Spec: &directorSchema.APISpecInput{
							Type:   directorSchema.APISpecTypeOpenAPI,
							Format: directorSchema.SpecFormatYaml,
							Data:   ptr.CLOB(`{"openapi":"3.0.2"}`),
						},
					},
					{
						Name:        "reviews-v1",
						Description: ptr.String("api for adding reviews"),
						TargetURL:   "http://mywordpress.com/reviews",
						Version:     pkg.FixActiveVersion(),
						Spec: &directorSchema.APISpecInput{
							Type:   directorSchema.APISpecTypeOdata,
							Format: directorSchema.SpecFormatJSON,
							Data:   ptr.CLOB(`{"openapi":"3.0.1"}`),
						},
					},
					{
						Name:        "xml",
						Description: ptr.String("xml api"),
						Version:     pkg.FixDecomissionedVersion(),
						TargetURL:   "http://mywordpress.com/xml",
						Spec: &directorSchema.APISpecInput{
							Type:   directorSchema.APISpecTypeOdata,
							Format: directorSchema.SpecFormatXML,
							Data:   ptr.CLOB("odata"),
						},
					},
				},
				EventDefinitions: []*directorSchema.EventDefinitionInput{
					{
						Name:        "comments-v1",
						Description: ptr.String("comments events"),
						Version:     pkg.FixDepracatedVersion(),
						Group:       ptr.String("comments"),
						Spec: &directorSchema.EventSpecInput{
							Type:   directorSchema.EventSpecTypeAsyncAPI,
							Format: directorSchema.SpecFormatYaml,
							Data:   ptr.CLOB(`{"asyncapi":"1.2.0"}`),
						},
					},
					{
						Name:        "reviews-v1",
						Description: ptr.String("review events"),
						Version:     pkg.FixActiveVersion(),
						Spec: &directorSchema.EventSpecInput{
							Type:   directorSchema.EventSpecTypeAsyncAPI,
							Format: directorSchema.SpecFormatYaml,
							Data:   ptr.CLOB(`{"asyncapi":"1.1.0"}`),
						},
					},
				},
			},
		},
	}, suffix)
}

func generateAppInputForDifferentTenants(appInput directorSchema.ApplicationRegisterInput, suffix string) directorSchema.ApplicationRegisterInput {
	appInput.Name += "-" + suffix
	for _, bndl := range appInput.Bundles {
		bndl.Name = bndl.Name + "-" + suffix

		for _, apiDef := range bndl.APIDefinitions {
			apiDef.Name = apiDef.Name + "-" + suffix
		}

		for _, eventDef := range bndl.EventDefinitions {
			eventDef.Name = eventDef.Name + "-" + suffix
		}
	}
	return appInput
}
