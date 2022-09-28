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

package director_test

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/director"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/director/directorfakes"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGraphQLClient_FetchApplication(t *testing.T) {
	const appID = "0191fcfd-ae7e-4d1a-8027-520a96d5319f"
	testErr := errors.New("some error")

	type testCase struct {
		name               string
		GQLClient          *directorfakes.FakeClient
		expectedErr        error
		expectedProperties map[string]int
		expectedApp        *director.ApplicationOutput
	}

	tests := []testCase{
		{
			name:      "success",
			GQLClient: getGCLI(t, fmt.Sprintf(`{"result":{"id":"%s"}}`, appID), nil),
			expectedApp: &director.ApplicationOutput{Result: &schema.ApplicationExt{Application: schema.Application{
				BaseEntity: &schema.BaseEntity{ID: appID},
			}}},
			expectedProperties: map[string]int{
				"webhooks":              1,
				"providerName":          0,
				"description":           0,
				"integrationSystemID":   0,
				"labels":                0,
				"bundles":               0,
				"auths":                 0,
				"status":                0,
				"instanceAuths":         0,
				"documents":             0,
				"fetchRequest":          0,
				"healthCheckURL":        0,
				"eventingConfiguration": 0,
			},
		},
		{
			name:        "returns error when app does not exist",
			GQLClient:   getGCLI(t, "", nil),
			expectedApp: nil,
			expectedErr: &director.NotFoundError{},
			expectedProperties: map[string]int{
				"webhooks":              1,
				"providerName":          0,
				"description":           0,
				"integrationSystemID":   0,
				"labels":                0,
				"bundles":               0,
				"auths":                 0,
				"status":                0,
				"instanceAuths":         0,
				"documents":             0,
				"fetchRequest":          0,
				"healthCheckURL":        0,
				"eventingConfiguration": 0,
			},
		},
		{
			name:        "when gql client returns an error",
			GQLClient:   getGCLI(t, "", testErr),
			expectedErr: testErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcli := tt.GQLClient
			c := director.NewGraphQLClient(
				gcli,
				&graphqlizer.Graphqlizer{},
				&graphqlizer.GqlFieldsProvider{},
			)
			app, err := c.FetchApplication(context.TODO(), "test-id")
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, app)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedApp, app)
				assert.Equal(t, 1, gcli.DoCallCount())

				_, graphqlReq, _ := gcli.DoArgsForCall(0)
				query := graphqlReq.Query()
				for expectedProp, expectedCount := range tt.expectedProperties {
					fieldRegex := regexp.MustCompile(`\b` + expectedProp + `\b`)

					matches := fieldRegex.FindAllStringIndex(query, -1)
					actualCount := len(matches)

					assert.Equal(t, expectedCount, actualCount, expectedProp)
				}
			}
		})
	}
}

func TestGraphQLClient_FetchApplications(t *testing.T) {
	type testCase struct {
		name               string
		GQLClient          *directorfakes.FakeClient
		expectedErr        string
		expectedProperties map[string]int
	}

	tests := []testCase{
		{
			name:      "success",
			GQLClient: getGCLI(t, "", nil),
			expectedProperties: map[string]int{
				"auths":         0,
				"webhooks":      0,
				"status":        0,
				"instanceAuths": 0,
				"documents":     0,
				"fetchRequest":  0,
			},
		},
		{
			name:        "when gql client returns an error",
			GQLClient:   getGCLI(t, "", errors.New("some error")),
			expectedErr: "while fetching applications in gqlclient: some error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcli := tt.GQLClient
			c := director.NewGraphQLClient(
				gcli,
				&graphqlizer.Graphqlizer{},
				&graphqlizer.GqlFieldsProvider{},
			)
			_, err := c.FetchApplications(context.TODO())
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.Equal(t, 1, gcli.DoCallCount())

				_, graphqlReq, _ := gcli.DoArgsForCall(0)
				query := graphqlReq.Query()
				for expectedProp, expectedCount := range tt.expectedProperties {
					fieldRegex := regexp.MustCompile(`\b` + expectedProp + `\b`)

					matches := fieldRegex.FindAllStringIndex(query, -1)
					actualCount := len(matches)

					assert.Equal(t, expectedCount, actualCount, expectedProp)
				}
			}
		})
	}
}

func TestGraphQLClient_RequestBundleInstanceCredentialsCreation(t *testing.T) {

	type testCase struct {
		name          string
		GQLClient     *directorfakes.FakeClient
		expectedErr   string
		expectedQuery string
	}

	tests := []testCase{
		{
			name:          "success",
			GQLClient:     getGCLI(t, "", nil),
			expectedQuery: "mutation {\n\t\t\t  result: requestBundleInstanceAuthCreation(\n\t\t\t\tbundleID: \"bundleID\"\n\t\t\t\tin: {\n\t\t\t\t  id: \"authID\"\n\t\t\t\t  context: \"null\"\n    \t\t\t  inputParams: \"null\"\n\t\t\t\t}\n\t\t\t  ) {\n\t\t\t\t\tstatus {\n\t\t\t\t\t  condition\n\t\t\t\t\t  timestamp\n\t\t\t\t\t  message\n\t\t\t\t\t  reason\n\t\t\t\t\t}\n\t\t\t  \t }\n\t\t\t\t}",
		},
		{
			name:        "when gql client returns an error",
			GQLClient:   getGCLI(t, "", errors.New("some error")),
			expectedErr: "while executing GraphQL call to create bundle instance auth: some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcli := tt.GQLClient
			c := director.NewGraphQLClient(
				gcli,
				&graphqlizer.Graphqlizer{},
				&graphqlizer.GqlFieldsProvider{},
			)
			_, err := c.RequestBundleInstanceCredentialsCreation(context.TODO(), &director.BundleInstanceCredentialsInput{
				BundleID: "bundleID",
				AuthID:   "authID",
			})
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.Equal(t, 1, gcli.DoCallCount())

				_, graphqlReq, _ := gcli.DoArgsForCall(0)
				query := graphqlReq.Query()
				assert.Equal(t, tt.expectedQuery, query)
			}
		})
	}
}

func TestGraphQLClient_FetchBundleInstanceCredentials(t *testing.T) {

	tests := []testCase{
		{
			name:      "success",
			GQLClient: getGCLI(t, "", nil),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedQuery: "query{\n\t\t\t  result:bundleByInstanceAuth(authID:\"authID\"){\n\t\t\t\tapiDefinitions{\n\t\t\t\t  data{\n\t\t\t\t\tname\n\t\t\t\t\ttargetURL\n\t\t\t\t  }\n\t\t\t\t}\n\t\t\t\tinstanceAuth(id: \"authID\"){\n\t\t\t\t  \n\t\tid\n\t\tcontext\n\t\tinputParams\n\t\tauth {credential {\n\t\t\t\t... on BasicCredentialData {\n\t\t\t\t\tusername\n\t\t\t\t\tpassword\n\t\t\t\t}\n\t\t\t\t...  on CertificateOAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tcertificate\n\t\t\t\t\turl\n\t\t\t\t}\n   \t\t\t\t...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t}\n\t\t\t}\n\t\t\toneTimeToken {\n\t\t\t\t__typename\n\t\t\t\ttoken\n\t\t\t\tused\n\t\t\t\texpiresAt\n\t\t\t}\n\t\t\tcertCommonName\n\t\t\taccessStrategy\n\t\t\tadditionalHeaders\n\t\t\tadditionalQueryParams\n\t\t\trequestAuth { \n\t\t\t  csrf {\n\t\t\t\ttokenEndpointURL\n\t\t\t\tcredential {\n\t\t\t\t  ... on BasicCredentialData {\n\t\t\t\t  \tusername\n\t\t\t\t\tpassword\n\t\t\t\t  }\n\t\t\t\t  ...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t  }\n\t\t\t\t  ...  on CertificateOAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tcertificate\n\t\t\t\t\turl\n\t\t\t\t  }\n\t\t\t    }\n\t\t\t\tadditionalHeaders\n\t\t\t\tadditionalQueryParams\n\t\t\t  }\n\t\t\t}\n\t\t}\n\t\tstatus {\n\t\tcondition\n\t\ttimestamp\n\t\tmessage\n\t\treason}\n\t\truntimeID\n\t\truntimeContextID\n\t\t\t\t}\n\t\t\t  }\n\t}",
		},
		{
			name:      "when gql client returns an error",
			GQLClient: getGCLI(t, "", errors.New("some error")),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "while executing GraphQL call to get bundle instance auth: some error",
		},
		{
			name:      "when no bundle is returned",
			GQLClient: getGCLI(t, `{}`, nil),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "NotFound",
		},
		{
			name:      "when no bundle instance auth is returned",
			GQLClient: getGCLI(t, `{"result":{}}`, nil),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "NotFound",
		},
		{
			name:      "when no bundle instance auth context is returned",
			GQLClient: getGCLI(t, `{"result":{"instanceAuth":{}}}`, nil),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "NotFound",
		},
		{
			name:      "when bundle instance auth context is not a JSON",
			GQLClient: getGCLI(t, `{"result":{"instanceAuth":{"context":"not a json"}}}`, nil),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "while unmarshaling auth context",
		},
		{
			name:      "when instance id is different than the one provided",
			GQLClient: getGCLI(t, `{"result":{"instanceAuth":{"context":"{\"instance_id\": \"db_id\"}"}}}`, nil),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
				Context: map[string]string{
					"instance_id": "inInstanceID",
				},
			},
			expectedErr: "found binding with mismatched context coordinates",
		},
		{
			name:      "when binding id is different than the one provided",
			GQLClient: getGCLI(t, `{"result": {"instanceAuth": {"context": "{\"instance_id\": \"inInstanceID\",\"binding_id\": \"db_id\"}"}}}`, nil),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
				Context: map[string]string{
					"instance_id": "inInstanceID",
					"binding_id":  "inBindingID",
				},
			},
			expectedErr: "found binding with mismatched context coordinates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcli := tt.GQLClient
			c := director.NewGraphQLClient(
				gcli,
				&graphqlizer.Graphqlizer{},
				&graphqlizer.GqlFieldsProvider{},
			)
			_, err := c.FetchBundleInstanceCredentials(context.TODO(), tt.credentialsInput)
			testCommonLogic(t, tt, err)
		})
	}
}

func TestGraphQLClient_FetchBundleInstanceAuth(t *testing.T) {

	tests := []testCase{
		{
			name:      "success",
			GQLClient: getGCLI(t, "", nil),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedQuery: "query {\n\t\t\t  result: bundleInstanceAuth(id: \"authID\") {\n\t\t\t\tid\n\t\t\t\tcontext\n\t\t\t\tstatus {\n\t\t\t\t  condition\n\t\t\t\t  timestamp\n\t\t\t\t  message\n\t\t\t\t  reason\n\t\t\t\t}\n\t\t\t  }\n\t}",
		},
		{
			name:      "when gql client returns an error",
			GQLClient: getGCLI(t, "", errors.New("some error")),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "while executing GraphQL call to get bundle instance auth: some error",
		},
		{
			name:      "when no bundle instance auth is returned",
			GQLClient: getGCLI(t, `{}`, nil),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "NotFound",
		},
		{
			name: "when no bundle instance auth context is returned",
			GQLClient: getGCLI(t, `{
							"result": {}
						}`, nil),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "NotFound",
		},
		{
			name: "when bundle instance auth context is not a JSON",
			GQLClient: getGCLI(t, `{
							"result": {
								"context": "not a json"
							}
						}`, nil),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "while unmarshaling auth context",
		},
		{
			name: "when instance id is different than the one provided",
			GQLClient: getGCLI(t, `{
							"result": {
								"context": "{\"instance_id\": \"db_id\"}"
							}
						}`, nil),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
				Context: map[string]string{
					"instance_id": "inInstanceID",
				},
			},
			expectedErr: "found binding with mismatched context coordinates",
		},
		{
			name: "when binding id is different than the one provided",
			GQLClient: getGCLI(t, `{
							"result": {
								"context": "{\"instance_id\": \"inInstanceID\", \"binding_id\": \"db_id\"}"
							}
						}`, nil),
			credentialsInput: &director.BundleInstanceInput{
				InstanceAuthID: "authID",
				Context: map[string]string{
					"instance_id": "inInstanceID",
					"binding_id":  "inBindingID",
				},
			},
			expectedErr: "found binding with mismatched context coordinates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcli := tt.GQLClient
			c := director.NewGraphQLClient(
				gcli,
				&graphqlizer.Graphqlizer{},
				&graphqlizer.GqlFieldsProvider{},
			)
			_, err := c.FetchBundleInstanceAuth(context.TODO(), tt.credentialsInput)
			testCommonLogic(t, tt, err)
		})
	}
}

func TestGraphQLClient_RequestBundleInstanceCredentialsDeletion(t *testing.T) {
	type testCase struct {
		name             string
		GQLClient        *directorfakes.FakeClient
		credentialsInput *director.BundleInstanceAuthDeletionInput
		expectedErr      string
		expectedQuery    string
	}

	tests := []testCase{
		{
			name:      "success",
			GQLClient: getGCLI(t, "", nil),
			credentialsInput: &director.BundleInstanceAuthDeletionInput{
				InstanceAuthID: "instanceAuthID",
			},
			expectedQuery: "mutation {\n\t\t\t  result: requestBundleInstanceAuthDeletion(" +
				"authID: \"instanceAuthID\") {\n\t\t\t\t\t\tid\n\t\t\t\t\t\tstatus {\n\t\t\t\t\t\t  condition\n\t\t\t\t\t\t  timestamp\n\t\t\t\t\t\t  message\n\t\t\t\t\t\t  reason\n\t\t\t\t\t\t}\n\t\t\t\t\t  }\n\t\t\t\t\t}",
		},
		{
			name:      "when gql client returns an error",
			GQLClient: getGCLI(t, "", errors.New("some error")),
			credentialsInput: &director.BundleInstanceAuthDeletionInput{
				InstanceAuthID: "instanceAuthID",
			},
			expectedErr: "while executing GraphQL call to delete the bundle instance auth: some error",
		},
		{
			name:      "when gql client returns object not found",
			GQLClient: getGCLI(t, "", errors.New("Object not found")),
			credentialsInput: &director.BundleInstanceAuthDeletionInput{
				InstanceAuthID: "instanceAuthID",
			},
			expectedErr: "NotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcli := tt.GQLClient
			c := director.NewGraphQLClient(
				gcli,
				&graphqlizer.Graphqlizer{},
				&graphqlizer.GqlFieldsProvider{},
			)
			_, err := c.RequestBundleInstanceCredentialsDeletion(context.TODO(), tt.credentialsInput)
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.Equal(t, 1, gcli.DoCallCount())

				_, graphqlReq, _ := gcli.DoArgsForCall(0)
				query := graphqlReq.Query()
				assert.Equal(t, tt.expectedQuery, query)
			}
		})
	}
}

func TestGraphQLClient_FindSpecification(t *testing.T) {
	type testCase struct {
		name             string
		GQLClient        *directorfakes.FakeClient
		credentialsInput *director.BundleSpecificationInput
		expectedSpec     *director.BundleSpecificationOutput
		expectedErr      string
		expectedQuery    string
	}

	specData := schema.CLOB("data")

	tests := []testCase{
		{
			name: "success when api spec",
			GQLClient: getGCLI(t, `{
							"result": {
								"bundle": {
									"apiDefinition": {
										"name": "apiDefName",
										"spec": {
											"data": "data",
											"format":"format"
										}
									}
								}
							}
						}`, nil),
			expectedSpec: &director.BundleSpecificationOutput{
				Name:   "apiDefName",
				Data:   &specData,
				Format: "format",
			},
			credentialsInput: &director.BundleSpecificationInput{
				ApplicationID: "appID",
				BundleID:      "bundleID",
				DefinitionID:  "defID",
			},
			expectedQuery: "query {\n\t\t\t  result: application(id: \"appID\") {\n\t\t\t\t\t\tbundle(id: \"bundleID\") {\n\t\t\t\t\t\t  apiDefinition(id: \"defID\") {\n\t\t\t\t\t\t\t  spec {\n\t\t\t\t\t\t\t\tdata\n\t\t\t\t\t\t\t\ttype\n\t\t\t\t\t\t\t\tformat\n\t\t\t\t\t\t\t  }\n\t\t\t\t\t\t  }\n\t\t\t\t\t\t  eventDefinition(id: \"defID\") {\n\t\t\t\t\t\t\t  spec {\n\t\t\t\t\t\t\t\tdata\n\t\t\t\t\t\t\t\ttype\n\t\t\t\t\t\t\t\tformat\n\t\t\t\t\t\t\t  }\n\t\t\t\t\t\t  }\n\t\t\t\t\t\t}\n\t\t\t\t\t  }\n\t\t\t\t\t}",
		},
		{
			name: "success when event spec",
			GQLClient: getGCLI(t, `{
							"result": {
								"bundle": {
									"eventDefinition": {
										"name": "eventDefName",
										"spec": {
											"data": "data",
											"format":"format"
										}
									}
								}
							}
						}`, nil),
			expectedSpec: &director.BundleSpecificationOutput{
				Name:   "eventDefName",
				Data:   &specData,
				Format: "format",
			},
			credentialsInput: &director.BundleSpecificationInput{
				ApplicationID: "appID",
				BundleID:      "bundleID",
				DefinitionID:  "defID",
			},
			expectedQuery: "query {\n\t\t\t  result: application(id: \"appID\") {\n\t\t\t\t\t\tbundle(id: \"bundleID\") {\n\t\t\t\t\t\t  apiDefinition(id: \"defID\") {\n\t\t\t\t\t\t\t  spec {\n\t\t\t\t\t\t\t\tdata\n\t\t\t\t\t\t\t\ttype\n\t\t\t\t\t\t\t\tformat\n\t\t\t\t\t\t\t  }\n\t\t\t\t\t\t  }\n\t\t\t\t\t\t  eventDefinition(id: \"defID\") {\n\t\t\t\t\t\t\t  spec {\n\t\t\t\t\t\t\t\tdata\n\t\t\t\t\t\t\t\ttype\n\t\t\t\t\t\t\t\tformat\n\t\t\t\t\t\t\t  }\n\t\t\t\t\t\t  }\n\t\t\t\t\t\t}\n\t\t\t\t\t  }\n\t\t\t\t\t}",
		},
		{
			name:      "when gql client returns an error",
			GQLClient: getGCLI(t, "", errors.New("some error")),
			credentialsInput: &director.BundleSpecificationInput{
				ApplicationID: "appID",
				BundleID:      "bundleID",
				DefinitionID:  "defID",
			},
			expectedErr: "while executing GraphQL call to get bundle instance auth: some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcli := tt.GQLClient
			c := director.NewGraphQLClient(
				gcli,
				&graphqlizer.Graphqlizer{},
				&graphqlizer.GqlFieldsProvider{},
			)
			spec, err := c.FindSpecification(context.TODO(), tt.credentialsInput)
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.Equal(t, 1, gcli.DoCallCount())

				_, graphqlReq, _ := gcli.DoArgsForCall(0)
				query := graphqlReq.Query()
				assert.Equal(t, tt.expectedQuery, query)
				assert.Equal(t, tt.expectedSpec, spec)
			}
		})
	}
}

func getGCLI(t *testing.T, response string, err error) *directorfakes.FakeClient {
	fakeGCLI := &directorfakes.FakeClient{}
	fakeGCLI.DoStub = func(_ context.Context, g *graphql.Request, i interface{}) error {
		if err != nil {
			return err
		}
		if response != "" {
			err := json.Unmarshal([]byte(response), i)
			assert.NoError(t, err)
		}
		return nil
	}
	return fakeGCLI
}

type testCase struct {
	name             string
	GQLClient        *directorfakes.FakeClient
	credentialsInput *director.BundleInstanceInput
	expectedErr      string
	expectedQuery    string
}

func testCommonLogic(t *testing.T, tt testCase, err error) {
	gcli := tt.GQLClient

	if tt.expectedErr != "" {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), tt.expectedErr)
	} else {
		assert.Equal(t, 1, gcli.DoCallCount())

		_, graphqlReq, _ := gcli.DoArgsForCall(0)
		query := graphqlReq.Query()
		assert.Equal(t, tt.expectedQuery, query)
	}
}
