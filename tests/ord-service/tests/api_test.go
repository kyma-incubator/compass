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
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/ord-service/pkg"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const TestApp = "mytestapp"

func TestODService(t *testing.T) {
	appInput := directorSchema.ApplicationRegisterInput{
		Name:        TestApp,
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

	httpClient := http.DefaultClient
	httpClient.Timeout = 10 * time.Second
	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	t.Run("Requesting Packages returns them as expected", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, testConfig.ORDServiceURL+"/packages?$format=json", nil)
		require.NoError(t, err)

		response, err := httpClient.Do(request)

		// THEN
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		body, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)

		respBody := string(body)
		require.Equal(t, len(appInput.Packages), len(gjson.Get(respBody, "value").Array()))
		require.Equal(t, appInput.Packages[0].Name, gjson.Get(respBody, "value.0.title").String())
		require.Equal(t, *appInput.Packages[0].Description, gjson.Get(respBody, "value.0.description").String())
	})

	t.Run("Requesting APIs and their specs returns them as expected", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, testConfig.ORDServiceURL+"/apis?$format=json", nil)
		require.NoError(t, err)

		response, err := httpClient.Do(request)

		// THEN
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		body, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)

		respBody := string(body)

		require.Equal(t, len(appInput.Packages[0].APIDefinitions), len(gjson.Get(respBody, "value").Array()))

		for i := range appInput.Packages[0].APIDefinitions {
			name := gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String()
			require.NotEmpty(t, name)

			expectedAPI := apisMap[name]

			require.Equal(t, *expectedAPI.Description, gjson.Get(respBody, fmt.Sprintf("value.%d.description", i)).String())
			require.Equal(t, expectedAPI.TargetURL, gjson.Get(respBody, fmt.Sprintf("value.%d.entryPoint", i)).String())
			require.NotEmpty(t, gjson.Get(respBody, fmt.Sprintf("value.%d.partOfPackage", i)).String())
			releaseStatus := gjson.Get(respBody, fmt.Sprintf("value.%d.releaseStatus", i)).String()
			if *expectedAPI.Version.ForRemoval {
				require.Equal(t, "decommissioned", releaseStatus)
			} else if *expectedAPI.Version.Deprecated {
				require.Equal(t, "deprecated", releaseStatus)
			} else {
				require.Equal(t, "active", releaseStatus)
			}

			specs := gjson.Get(respBody, fmt.Sprintf("value.%d.apiDefinitions", i)).Array()
			require.Equal(t, 1, len(specs))

			specType := specs[0].Get("type").String()
			if expectedAPI.Spec.Type == directorSchema.APISpecTypeOdata {
				require.Equal(t, "edmx", specType)
			} else if expectedAPI.Spec.Type == directorSchema.APISpecTypeOpenAPI {
				require.Equal(t, "openapi-v3", specType)
			}

			specFormat := specs[0].Get("mediaType").String()
			if expectedAPI.Spec.Format == directorSchema.SpecFormatYaml {
				require.Equal(t, "text/yaml", specFormat)
			} else if expectedAPI.Spec.Format == directorSchema.SpecFormatJSON {
				require.Equal(t, "application/json", specFormat)
			} else if expectedAPI.Spec.Format == directorSchema.SpecFormatXML {
				require.Equal(t, "application/xml", specFormat)
			}

			apiID := gjson.Get(respBody, fmt.Sprintf("value.%d.id", i)).String()
			require.NotEmpty(t, apiID)

			specURL := specs[0].Get("url").String()
			require.Equal(t, fmt.Sprintf("%s/api/%s/specification", testConfig.ORDServiceURL, apiID), specURL)

			specReq, err := http.NewRequest(http.MethodGet, specURL, nil)
			require.NoError(t, err)

			specResp, err := httpClient.Do(specReq)

			require.NoError(t, err)
			require.Equal(t, http.StatusOK, specResp.StatusCode)

			body, err := ioutil.ReadAll(specResp.Body)
			require.NoError(t, err)

			require.Equal(t, string(*expectedAPI.Spec.Data), string(body))
		}
	})

	t.Run("Requesting Events and their specs returns them as expected", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, testConfig.ORDServiceURL+"/events?$format=json", nil)
		require.NoError(t, err)

		response, err := httpClient.Do(request)

		// THEN
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		body, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)

		respBody := string(body)

		require.Equal(t, len(appInput.Packages[0].EventDefinitions), len(gjson.Get(respBody, "value").Array()))

		for i := range appInput.Packages[0].EventDefinitions {
			name := gjson.Get(respBody, fmt.Sprintf("value.%d.title", i)).String()
			require.NotEmpty(t, name)

			expectedEvent := eventsMap[name]

			require.Equal(t, *expectedEvent.Description, gjson.Get(respBody, fmt.Sprintf("value.%d.description", i)).String())
			require.NotEmpty(t, gjson.Get(respBody, fmt.Sprintf("value.%d.partOfPackage", i)).String())
			releaseStatus := gjson.Get(respBody, fmt.Sprintf("value.%d.releaseStatus", i)).String()
			if *expectedEvent.Version.ForRemoval {
				require.Equal(t, "decommissioned", releaseStatus)
			} else if *expectedEvent.Version.Deprecated {
				require.Equal(t, "deprecated", releaseStatus)
			} else {
				require.Equal(t, "active", releaseStatus)
			}

			specs := gjson.Get(respBody, fmt.Sprintf("value.%d.eventDefinitions", i)).Array()
			require.Equal(t, 1, len(specs))

			specType := specs[0].Get("type").String()
			if expectedEvent.Spec.Type == directorSchema.EventSpecTypeAsyncAPI {
				require.Equal(t, "asyncapi-v2", specType)
			}

			specFormat := specs[0].Get("mediaType").String()
			if expectedEvent.Spec.Format == directorSchema.SpecFormatYaml {
				require.Equal(t, "text/yaml", specFormat)
			} else if expectedEvent.Spec.Format == directorSchema.SpecFormatJSON {
				require.Equal(t, "application/json", specFormat)
			} else if expectedEvent.Spec.Format == directorSchema.SpecFormatXML {
				require.Equal(t, "application/xml", specFormat)
			}

			eventID := gjson.Get(respBody, fmt.Sprintf("value.%d.id", i)).String()
			require.NotEmpty(t, eventID)

			specURL := specs[0].Get("url").String()
			require.Equal(t, fmt.Sprintf("%s/event/%s/specification", testConfig.ORDServiceURL, eventID), specURL)

			specReq, err := http.NewRequest(http.MethodGet, specURL, nil)
			require.NoError(t, err)

			specResp, err := httpClient.Do(specReq)

			require.NoError(t, err)
			require.Equal(t, http.StatusOK, specResp.StatusCode)

			body, err := ioutil.ReadAll(specResp.Body)
			require.NoError(t, err)

			require.Equal(t, string(*expectedEvent.Spec.Data), string(body))
		}
	})

	t.Run("Errors generate user-friendly message", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, testConfig.ORDServiceURL+"/test?$format=json", nil)
		require.NoError(t, err)

		response, err := httpClient.Do(request)

		// THEN
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, response.StatusCode)

		body, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)

		respBody := string(body)

		require.Contains(t, gjson.Get(respBody, "error.message").String(), "Use odata-debug query parameter with value one of the following formats: json,html,download for more information")
	})

}
