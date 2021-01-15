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
	"go/ast"
	"io/ioutil"
	"net/http"
	urlpkg "net/url"
	"regexp"
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

func TestORDService(t *testing.T) {
	appInput := directorSchema.ApplicationRegisterInput{
		Name:        "test-app",
		Description: ptr.String("my application"),
		Packages: []*directorSchema.PackageCreateInput{
			{
				Name:        "foo-pkg",
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
	}

	apisMap := make(map[string]directorSchema.APIDefinitionInput, 0)
	for _, apiDefinition := range appInput.Packages[0].APIDefinitions {
		apisMap[apiDefinition.Name] = *apiDefinition
	}

	eventsMap := make(map[string]directorSchema.EventDefinitionInput, 0)
	for _, eventDefinition := range appInput.Packages[0].EventDefinitions {
		eventsMap[eventDefinition.Name] = *eventDefinition
	}

	ctx := context.Background()

	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	app, err := pkg.RegisterApplicationWithinTenant(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, appInput)
	require.NoError(t, err)

	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)

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

	t.Run("Requesting entities without specifying response format falls back to configured default response type", func(t *testing.T) {
		makeRequest(t, httpClient, testConfig.ORDServiceURL+"/packages")
	})

	t.Run("Requesting Packages returns them as expected", func(t *testing.T) {
		respBody := makeRequest(t, httpClient, testConfig.ORDServiceURL+"/packages?$format=json")

		require.Equal(t, len(appInput.Packages), len(gjson.Get(respBody, "value").Array()))
		require.Equal(t, appInput.Packages[0].Name, gjson.Get(respBody, "value.0.title").String())
		require.Equal(t, *appInput.Packages[0].Description, gjson.Get(respBody, "value.0.description").String())
	})

	t.Run("Requesting APIs and their specs returns them as expected", func(t *testing.T) {
		respBody := makeRequest(t, httpClient, testConfig.ORDServiceURL+"/apis?$format=json")

		require.Equal(t, len(appInput.Packages[0].APIDefinitions), len(gjson.Get(respBody, "value").Array()))

		for i := range appInput.Packages[0].APIDefinitions {
			name := gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String()
			require.NotEmpty(t, name)

			expectedAPI, exists := apisMap[name]
			require.True(t, exists)

			require.Equal(t, *expectedAPI.Description, gjson.Get(respBody, fmt.Sprintf("value.%d.description", i)).String())
			require.Equal(t, expectedAPI.TargetURL, gjson.Get(respBody, fmt.Sprintf("value.%d.entryPoint", i)).String())
			require.NotEmpty(t, gjson.Get(respBody, fmt.Sprintf("value.%d.partOfPackage", i)).String())

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
			require.Equal(t, testConfig.ORDServiceURL+specPath, specURL)

			respBody := makeRequest(t, httpClient, specURL)

			require.Equal(t, string(*expectedAPI.Spec.Data), respBody)
		}
	})

	t.Run("Requesting Events and their specs returns them as expected", func(t *testing.T) {
		respBody := makeRequest(t, httpClient, testConfig.ORDServiceURL+"/events?$format=json")

		require.Equal(t, len(appInput.Packages[0].EventDefinitions), len(gjson.Get(respBody, "value").Array()))

		for i := range appInput.Packages[0].EventDefinitions {
			name := gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String()
			require.NotEmpty(t, name)

			expectedEvent, exists := eventsMap[name]
			require.True(t, exists)

			require.Equal(t, *expectedEvent.Description, gjson.Get(respBody, fmt.Sprintf("value.%d.description", i)).String())
			require.NotEmpty(t, gjson.Get(respBody, fmt.Sprintf("value.%d.partOfPackage", i)).String())

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
			require.Equal(t, testConfig.ORDServiceURL+specPath, specURL)

			respBody := makeRequest(t, httpClient, specURL)

			require.Equal(t, string(*expectedEvent.Spec.Data), respBody)
		}
	})

	// Paging:
	t.Run("Requesting paging of Packages returns them as expected", func(t *testing.T) {
		totalCount := len(appInput.Packages)

		respBody := makeRequest(t, httpClient, testConfig.ORDServiceURL+"/packages?$top=10&$skip=0&$format=json")
		require.Equal(t, totalCount, len(gjson.Get(respBody, "value").Array()))

		respBody = makeRequest(t, httpClient, fmt.Sprintf("%s/packages?$top=10&$skip=%d&$format=json", testConfig.ORDServiceURL, totalCount))
		require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
	})

	t.Run("Requesting paging of Package APIs returns them as expected", func(t *testing.T) {
		totalCount := len(appInput.Packages[0].APIDefinitions)

		respBody := makeRequest(t, httpClient, testConfig.ORDServiceURL+"/packages?$expand=apis($top=10;$skip=0)&$format=json")
		require.Equal(t, totalCount, len(gjson.Get(respBody, "value.0.apis").Array()))

		expectedItemCount := 1
		respBody = makeRequest(t, httpClient, fmt.Sprintf("%s/packages?$expand=apis($top=10;$skip=%d)&$format=json", testConfig.ORDServiceURL, totalCount-expectedItemCount))
		require.Equal(t, expectedItemCount, len(gjson.Get(respBody, "value").Array()))
	})

	t.Run("Requesting paging of Package Events returns them as expected", func(t *testing.T) {
		totalCount := len(appInput.Packages[0].EventDefinitions)

		respBody := makeRequest(t, httpClient, testConfig.ORDServiceURL+"/packages?$expand=events($top=10;$skip=0)&$format=json")
		require.Equal(t, totalCount, len(gjson.Get(respBody, "value.0.events").Array()))

		expectedItemCount := 1
		respBody = makeRequest(t, httpClient, fmt.Sprintf("%s/packages?$expand=events($top=10;$skip=%d)&$format=json", testConfig.ORDServiceURL, totalCount-expectedItemCount))
		require.Equal(t, expectedItemCount, len(gjson.Get(respBody, "value").Array()))
	})

	// Filtering:
	t.Run("Requesting filtering of Packages returns them as expected", func(t *testing.T) {
		pkgName := appInput.Packages[0].Name

		escapedFilterValue := urlpkg.PathEscape(fmt.Sprintf("title eq '%s'", pkgName))
		respBody := makeRequest(t, httpClient, fmt.Sprintf("%s/packages?$filter=(%s)&$format=json", testConfig.ORDServiceURL, escapedFilterValue))
		require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

		escapedFilterValue = urlpkg.PathEscape(fmt.Sprintf("title ne '%s'", pkgName))
		respBody = makeRequest(t, httpClient, fmt.Sprintf("%s/packages?$filter=(%s)&$format=json", testConfig.ORDServiceURL, escapedFilterValue))
		require.Equal(t, 0, len(gjson.Get(respBody, "value").Array()))
	})

	t.Run("Requesting filtering of Package APIs returns them as expected", func(t *testing.T) {
		totalCount := len(appInput.Packages[0].APIDefinitions)
		apiName := appInput.Packages[0].APIDefinitions[0].Name

		escapedFilterValue := urlpkg.PathEscape(fmt.Sprintf("title eq '%s'", apiName))
		respBody := makeRequest(t, httpClient, fmt.Sprintf("%s/packages?$expand=apis($filter=(%s))&$format=json", testConfig.ORDServiceURL, escapedFilterValue))
		require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

		escapedFilterValue = urlpkg.PathEscape(fmt.Sprintf("title ne '%s'", apiName))
		respBody = makeRequest(t, httpClient, fmt.Sprintf("%s/packages?$expand=apis($filter=(%s))&$format=json", testConfig.ORDServiceURL, escapedFilterValue))
		require.Equal(t, totalCount-1, len(gjson.Get(respBody, "value.0.apis").Array()))
	})

	t.Run("Requesting filtering of Package Events returns them as expected", func(t *testing.T) {
		totalCount := len(appInput.Packages[0].EventDefinitions)
		eventName := appInput.Packages[0].EventDefinitions[0].Name

		escapedFilterValue := urlpkg.PathEscape(fmt.Sprintf("title eq '%s'", eventName))
		respBody := makeRequest(t, httpClient, fmt.Sprintf("%s/packages?$expand=events($filter=(%s))&$format=json", testConfig.ORDServiceURL, escapedFilterValue))
		require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))

		escapedFilterValue = urlpkg.PathEscape(fmt.Sprintf("title ne '%s'", eventName))
		respBody = makeRequest(t, httpClient, fmt.Sprintf("%s/packages?$expand=events($filter=(%s))&$format=json", testConfig.ORDServiceURL, escapedFilterValue))
		require.Equal(t, totalCount-1, len(gjson.Get(respBody, "value.0.events").Array()))
	})

	// Projection:
	t.Run("Requesting projection of Packages returns them as expected", func(t *testing.T) {
		respBody := makeRequest(t, httpClient, testConfig.ORDServiceURL+"/packages?$select=title&$format=json")
		require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
		require.Equal(t, appInput.Packages[0].Name, gjson.Get(respBody, "value.0.title").String())
		require.Equal(t, false, gjson.Get(respBody, "value.0.description").Exists())
	})

	t.Run("Requesting projection of Package APIs returns them as expected", func(t *testing.T) {
		respBody := makeRequest(t, httpClient, testConfig.ORDServiceURL+"/packages?$expand=apis($select=title)&$format=json")

		apis := gjson.Get(respBody, "value.0.apis").Array()
		require.Len(t, apis, len(appInput.Packages[0].APIDefinitions))

		for i := range appInput.Packages[0].APIDefinitions {
			name := apis[i].Get("title").String()
			require.NotEmpty(t, name)

			_, exists := apisMap[name]
			require.True(t, exists)
			require.False(t, apis[i].Get("description").Exists())
		}
	})

	t.Run("Requesting projection of Package Events returns them as expected", func(t *testing.T) {
		respBody := makeRequest(t, httpClient, testConfig.ORDServiceURL+"/packages?$expand=events($select=title)&$format=json")

		events := gjson.Get(respBody, "value.0.events").Array()
		require.Len(t, events, len(appInput.Packages[0].EventDefinitions))

		for i := range appInput.Packages[0].EventDefinitions {
			name := events[i].Get("title").String()
			require.NotEmpty(t, name)

			_, exists := eventsMap[name]
			require.True(t, exists)
			require.False(t, events[i].Get("description").Exists())
		}
	})

	//Ordering:
	t.Run("Requesting ordering of Packages returns them as expected", func(t *testing.T) {
		escapedOrderByValue := urlpkg.PathEscape("title asc,description desc")
		respBody := makeRequest(t, httpClient, fmt.Sprintf("%s/packages?$orderby=%s&$format=json", testConfig.ORDServiceURL, escapedOrderByValue))
		require.Equal(t, 1, len(gjson.Get(respBody, "value").Array()))
		require.Equal(t, appInput.Packages[0].Name, gjson.Get(respBody, "value.0.title").String())
		require.Equal(t, *appInput.Packages[0].Description, gjson.Get(respBody, "value.0.description").String())
	})

	t.Run("Requesting ordering of Package APIs returns them as expected", func(t *testing.T) {
		escapedOrderByValue := urlpkg.PathEscape("title asc,description desc")
		respBody := makeRequest(t, httpClient, fmt.Sprintf("%s/packages?$expand=apis($orderby=%s)&$format=json", testConfig.ORDServiceURL, escapedOrderByValue))

		apis := gjson.Get(respBody, "value.0.apis").Array()
		require.Len(t, apis, len(appInput.Packages[0].APIDefinitions))

		for i := range appInput.Packages[0].APIDefinitions {
			name := apis[i].Get("title").String()
			require.NotEmpty(t, name)

			expectedAPI, exists := apisMap[name]
			require.True(t, exists)

			require.Equal(t, *expectedAPI.Description, apis[i].Get("description").String())
		}
	})

	t.Run("Requesting ordering of Package Events returns them as expected", func(t *testing.T) {
		escapedOrderByValue := urlpkg.PathEscape("title asc,description desc")
		respBody := makeRequest(t, httpClient, fmt.Sprintf("%s/packages?$expand=events($orderby=%s)&$format=json", testConfig.ORDServiceURL, escapedOrderByValue))

		events := gjson.Get(respBody, "value.0.events").Array()
		require.Len(t, events, len(appInput.Packages[0].EventDefinitions))

		for i := range appInput.Packages[0].EventDefinitions {
			name := events[i].Get("title").String()
			require.NotEmpty(t, name)

			expectedEvent, exists := eventsMap[name]
			require.True(t, exists)

			require.Equal(t, *expectedEvent.Description, events[i].Get("description").String())
		}
	})

	t.Run("Errors generate user-friendly message", func(t *testing.T) {
		respBody := makeRequestWithStatusExpect(t, httpClient, testConfig.ORDServiceURL+"/test?$format=json", http.StatusNotFound)

		require.Contains(t, gjson.Get(respBody, "error.message").String(), "Use odata-debug query parameter with value one of the following formats: json,html,download for more information")
	})
}

func makeRequest(t *testing.T, httpClient *http.Client, url string) string {
	return makeRequestWithStatusExpect(t, httpClient, url, http.StatusOK)
}

func makeRequestWithStatusExpect(t *testing.T, httpClient *http.Client, url string, expectedHTTPStatus int) string {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	response, err := httpClient.Do(request)

	require.NoError(t, err)
	require.Equal(t, expectedHTTPStatus, response.StatusCode)

	reqFormatPattern := regexp.MustCompile(fmt.Sprintf("^.*$format=(.*)$"))
	matches := reqFormatPattern.FindStringSubmatch(url)

	contentType := response.Header.Get("Content-Type")
	if len(matches) > 1 {
		require.Contains(t, contentType, matches[1])
	} else {
		require.Contains(t, contentType, testConfig.ORDServiceDefaultResponseType)
	}

	body, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err)

	return string(body)
}
