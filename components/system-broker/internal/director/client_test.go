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
	"regexp"
	"testing"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director/directorfakes"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGraphQLClient_FetchApplications(t *testing.T) {
	type fields struct {
		getGCLI           func() *directorfakes.FakeClient
		inputGraphqlizer  director.GraphQLizer
		outputGraphqlizer director.GqlFieldsProvider
	}
	type testCase struct {
		name               string
		fields             fields
		expectedErr        string
		expectedProperties map[string]int
	}

	inputGraphqlizer := &graphqlizer.Graphqlizer{}
	outputGraphqlizer := &graphqlizer.GqlFieldsProvider{}

	tests := []testCase{
		{
			name: "success",
			fields: fields{
				getGCLI: func() *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoReturns(nil)
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
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
			name: "when gql client returns an error",
			fields: fields{
				getGCLI: func() *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoReturns(errors.New("some error"))
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			expectedErr: "while fetching applications in gqlclient: some error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcli := tt.fields.getGCLI()
			c := director.NewGraphQLClient(
				gcli,
				tt.fields.inputGraphqlizer,
				tt.fields.outputGraphqlizer,
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

func TestGraphQLClient_RequestPackageInstanceCredentialsCreation(t *testing.T) {
	type fields struct {
		getGCLI           func() *directorfakes.FakeClient
		inputGraphqlizer  director.GraphQLizer
		outputGraphqlizer director.GqlFieldsProvider
	}
	type testCase struct {
		name             string
		fields           fields
		credentialsInput *director.PackageInstanceCredentialsInput
		expectedErr      string
		expectedQuery    string
	}

	inputGraphqlizer := &graphqlizer.Graphqlizer{}
	outputGraphqlizer := &graphqlizer.GqlFieldsProvider{}

	tests := []testCase{
		{
			name: "success",
			fields: fields{
				getGCLI: func() *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoReturns(nil)
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceCredentialsInput{
				PackageID: "packageID",
				AuthID:    "authID",
			},
			expectedQuery: "mutation {\n\t\t\t  result: requestPackageInstanceAuthCreation(\n\t\t\t\tpackageID: \"packageID\"\n\t\t\t\tin: {\n\t\t\t\t  id: \"authID\"\n\t\t\t\t  context: \"null\"\n    \t\t\t  inputParams: \"null\"\n\t\t\t\t}\n\t\t\t  ) {\n\t\t\t\t\tstatus {\n\t\t\t\t\t  condition\n\t\t\t\t\t  timestamp\n\t\t\t\t\t  message\n\t\t\t\t\t  reason\n\t\t\t\t\t}\n\t\t\t  \t }\n\t\t\t\t}",
		},
		{
			name: "when gql client returns an error",
			fields: fields{
				getGCLI: func() *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoReturns(errors.New("some error"))
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceCredentialsInput{
				PackageID: "packageID",
				AuthID:    "authID",
			},
			expectedErr: "while executing GraphQL call to create package instance auth: some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcli := tt.fields.getGCLI()
			c := director.NewGraphQLClient(
				gcli,
				tt.fields.inputGraphqlizer,
				tt.fields.outputGraphqlizer,
			)
			_, err := c.RequestPackageInstanceCredentialsCreation(context.TODO(), tt.credentialsInput)
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

func TestGraphQLClient_FetchPackageInstanceCredentials(t *testing.T) {
	type fields struct {
		getGCLI           func(*testing.T) *directorfakes.FakeClient
		inputGraphqlizer  director.GraphQLizer
		outputGraphqlizer director.GqlFieldsProvider
	}
	type testCase struct {
		name             string
		fields           fields
		credentialsInput *director.PackageInstanceInput
		expectedErr      string
		expectedQuery    string
	}

	inputGraphqlizer := &graphqlizer.Graphqlizer{}
	outputGraphqlizer := &graphqlizer.GqlFieldsProvider{}

	tests := []testCase{
		{
			name: "success",
			fields: fields{
				getGCLI: func(*testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoReturns(nil)
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedQuery: "query{\n\t\t\t  result:packageByInstanceAuth(authID:\"authID\"){\n\t\t\t\tapiDefinitions{\n\t\t\t\t  data{\n\t\t\t\t\tname\n\t\t\t\t\ttargetURL\n\t\t\t\t  }\n\t\t\t\t}\n\t\t\t\tinstanceAuth(id: \"authID\"){\n\t\t\t\t  \n\t\tid\n\t\tcontext\n\t\tinputParams\n\t\tauth {credential {\n\t\t\t\t... on BasicCredentialData {\n\t\t\t\t\tusername\n\t\t\t\t\tpassword\n\t\t\t\t}\n\t\t\t\t...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t}\n\t\t\t}\n\t\t\tadditionalHeaders\n\t\t\tadditionalQueryParams\n\t\t\trequestAuth { \n\t\t\t  csrf {\n\t\t\t\ttokenEndpointURL\n\t\t\t\tcredential {\n\t\t\t\t  ... on BasicCredentialData {\n\t\t\t\t  \tusername\n\t\t\t\t\tpassword\n\t\t\t\t  }\n\t\t\t\t  ...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t  }\n\t\t\t    }\n\t\t\t\tadditionalHeaders\n\t\t\t\tadditionalQueryParams\n\t\t\t}\n\t\t\t}\n\t\t}\n\t\tstatus {\n\t\tcondition\n\t\ttimestamp\n\t\tmessage\n\t\treason}\n\t\t\t\t}\n\t\t\t  }\n\t}",
		},
		{
			name: "when gql client returns an error",
			fields: fields{
				getGCLI: func(t *testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoReturns(errors.New("some error"))
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "while executing GraphQL call to get package instance auth: some error",
		},
		{
			name: "when no package is returned",
			fields: fields{
				getGCLI: func(*testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoStub = func(c context.Context, g *graphql.Request, i interface{}) error {
						bytesString := `{}`
						err := json.Unmarshal([]byte(bytesString), i)
						assert.NoError(t, err)
						return nil
					}
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "NotFound",
		},
		{
			name: "when no package instance auth is returned",
			fields: fields{
				getGCLI: func(*testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoStub = func(c context.Context, g *graphql.Request, i interface{}) error {
						bytesString := `{
							"result": {}
						}`
						err := json.Unmarshal([]byte(bytesString), i)
						assert.NoError(t, err)
						return nil
					}
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "NotFound",
		},
		{
			name: "when no package instance auth context is returned",
			fields: fields{
				getGCLI: func(*testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoStub = func(c context.Context, g *graphql.Request, i interface{}) error {
						bytesString := `{
							"result": {
								"instanceAuth": {}
							}
						}`
						err := json.Unmarshal([]byte(bytesString), i)
						assert.NoError(t, err)
						return nil
					}
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "NotFound",
		},
		{
			name: "when package instance auth context is not a JSON",
			fields: fields{
				getGCLI: func(*testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoStub = func(c context.Context, g *graphql.Request, i interface{}) error {
						bytesString := `{
							"result": {
								"instanceAuth": {
									"context": "not a json"
								}
							}
						}`
						err := json.Unmarshal([]byte(bytesString), i)
						assert.NoError(t, err)
						return nil
					}
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "while unmarshaling auth context",
		},
		{
			name: "when instance id is different than the one provided",
			fields: fields{
				getGCLI: func(*testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoStub = func(c context.Context, g *graphql.Request, i interface{}) error {
						bytesString := `{
							"result": {
								"instanceAuth": {
									"context": "{\"instance_id\": \"db_id\"}"
								}
							}
						}`
						err := json.Unmarshal([]byte(bytesString), i)
						assert.NoError(t, err)
						return nil
					}
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
				InstanceAuthID: "authID",
				Context: map[string]string{
					"instance_id": "inInstanceID",
				},
			},
			expectedErr: "found binding with mismatched context coordinates",
		},
		{
			name: "when binding id is different than the one provided",
			fields: fields{
				getGCLI: func(*testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoStub = func(c context.Context, g *graphql.Request, i interface{}) error {
						bytesString := `{
							"result": {
								"instanceAuth": {
									"context": "{\"instance_id\": \"inInstanceID\", \"binding_id\": \"db_id\"}"
								}
							}
						}`
						err := json.Unmarshal([]byte(bytesString), i)
						assert.NoError(t, err)
						return nil
					}
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
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
			gcli := tt.fields.getGCLI(t)
			c := director.NewGraphQLClient(
				gcli,
				tt.fields.inputGraphqlizer,
				tt.fields.outputGraphqlizer,
			)
			_, err := c.FetchPackageInstanceCredentials(context.TODO(), tt.credentialsInput)
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

func TestGraphQLClient_FetchPackageInstanceAuth(t *testing.T) {
	type fields struct {
		getGCLI           func(*testing.T) *directorfakes.FakeClient
		inputGraphqlizer  director.GraphQLizer
		outputGraphqlizer director.GqlFieldsProvider
	}
	type testCase struct {
		name             string
		fields           fields
		credentialsInput *director.PackageInstanceInput
		expectedErr      string
		expectedQuery    string
	}

	inputGraphqlizer := &graphqlizer.Graphqlizer{}
	outputGraphqlizer := &graphqlizer.GqlFieldsProvider{}

	tests := []testCase{
		{
			name: "success",
			fields: fields{
				getGCLI: func(*testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoReturns(nil)
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedQuery: "query {\n\t\t\t  result: packageInstanceAuth(id: \"authID\") {\n\t\t\t\tid\n\t\t\t\tcontext\n\t\t\t\tstatus {\n\t\t\t\t  condition\n\t\t\t\t  timestamp\n\t\t\t\t  message\n\t\t\t\t  reason\n\t\t\t\t}\n\t\t\t  }\n\t}",
		},
		{
			name: "when gql client returns an error",
			fields: fields{
				getGCLI: func(t *testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoReturns(errors.New("some error"))
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "while executing GraphQL call to get package instance auth: some error",
		},
		{
			name: "when no package instance auth is returned",
			fields: fields{
				getGCLI: func(*testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoStub = func(c context.Context, g *graphql.Request, i interface{}) error {
						bytesString := `{}`
						err := json.Unmarshal([]byte(bytesString), i)
						assert.NoError(t, err)
						return nil
					}
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "NotFound",
		},
		{
			name: "when no package instance auth context is returned",
			fields: fields{
				getGCLI: func(*testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoStub = func(c context.Context, g *graphql.Request, i interface{}) error {
						bytesString := `{
							"result": {}
						}`
						err := json.Unmarshal([]byte(bytesString), i)
						assert.NoError(t, err)
						return nil
					}
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "NotFound",
		},
		{
			name: "when package instance auth context is not a JSON",
			fields: fields{
				getGCLI: func(*testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoStub = func(c context.Context, g *graphql.Request, i interface{}) error {
						bytesString := `{
							"result": {
								"context": "not a json"
							}
						}`
						err := json.Unmarshal([]byte(bytesString), i)
						assert.NoError(t, err)
						return nil
					}
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
				InstanceAuthID: "authID",
			},
			expectedErr: "while unmarshaling auth context",
		},
		{
			name: "when instance id is different than the one provided",
			fields: fields{
				getGCLI: func(*testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoStub = func(c context.Context, g *graphql.Request, i interface{}) error {
						bytesString := `{
							"result": {
								"context": "{\"instance_id\": \"db_id\"}"
							}
						}`
						err := json.Unmarshal([]byte(bytesString), i)
						assert.NoError(t, err)
						return nil
					}
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
				InstanceAuthID: "authID",
				Context: map[string]string{
					"instance_id": "inInstanceID",
				},
			},
			expectedErr: "found binding with mismatched context coordinates",
		},
		{
			name: "when binding id is different than the one provided",
			fields: fields{
				getGCLI: func(*testing.T) *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoStub = func(c context.Context, g *graphql.Request, i interface{}) error {
						bytesString := `{
							"result": {
								"context": "{\"instance_id\": \"inInstanceID\", \"binding_id\": \"db_id\"}"
							}
						}`
						err := json.Unmarshal([]byte(bytesString), i)
						assert.NoError(t, err)
						return nil
					}
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceInput{
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
			gcli := tt.fields.getGCLI(t)
			c := director.NewGraphQLClient(
				gcli,
				tt.fields.inputGraphqlizer,
				tt.fields.outputGraphqlizer,
			)
			_, err := c.FetchPackageInstanceAuth(context.TODO(), tt.credentialsInput)
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

func TestGraphQLClient_RequestPackageInstanceCredentialsDeletion(t *testing.T) {
	type fields struct {
		getGCLI           func() *directorfakes.FakeClient
		inputGraphqlizer  director.GraphQLizer
		outputGraphqlizer director.GqlFieldsProvider
	}
	type testCase struct {
		name             string
		fields           fields
		credentialsInput *director.PackageInstanceAuthDeletionInput
		expectedErr      string
		expectedQuery    string
	}

	inputGraphqlizer := &graphqlizer.Graphqlizer{}
	outputGraphqlizer := &graphqlizer.GqlFieldsProvider{}

	tests := []testCase{
		{
			name: "success",
			fields: fields{
				getGCLI: func() *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoReturns(nil)
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceAuthDeletionInput{
				InstanceAuthID: "instanceAuthID",
			},
			expectedQuery: "mutation {\n\t\t\t  result: requestPackageInstanceAuthDeletion(authID: \"instanceAuthID\") {\n\t\t\t\t\t\tid\n\t\t\t\t\t\tstatus {\n\t\t\t\t\t\t  condition\n\t\t\t\t\t\t  timestamp\n\t\t\t\t\t\t  message\n\t\t\t\t\t\t  reason\n\t\t\t\t\t\t}\n\t\t\t\t\t  }\n\t\t\t\t\t}",
		},
		{
			name: "when gql client returns an error",
			fields: fields{
				getGCLI: func() *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoReturns(errors.New("some error"))
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceAuthDeletionInput{
				InstanceAuthID: "instanceAuthID",
			},
			expectedErr: "while executing GraphQL call to delete the package instance auth: some error",
		},
		{
			name: "when gql client returns object not found",
			fields: fields{
				getGCLI: func() *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoReturns(errors.New("Object not found"))
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageInstanceAuthDeletionInput{
				InstanceAuthID: "instanceAuthID",
			},
			expectedErr: "NotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcli := tt.fields.getGCLI()
			c := director.NewGraphQLClient(
				gcli,
				tt.fields.inputGraphqlizer,
				tt.fields.outputGraphqlizer,
			)
			_, err := c.RequestPackageInstanceCredentialsDeletion(context.TODO(), tt.credentialsInput)
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
	type fields struct {
		getGCLI           func() *directorfakes.FakeClient
		inputGraphqlizer  director.GraphQLizer
		outputGraphqlizer director.GqlFieldsProvider
	}
	type testCase struct {
		name             string
		fields           fields
		credentialsInput *director.PackageSpecificationInput
		expectedSpec     *director.PackageSpecificationOutput
		expectedErr      string
		expectedQuery    string
	}

	specData := schema.CLOB("data")

	inputGraphqlizer := &graphqlizer.Graphqlizer{}
	outputGraphqlizer := &graphqlizer.GqlFieldsProvider{}

	tests := []testCase{
		{
			name: "success when api spec",
			fields: fields{
				getGCLI: func() *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoStub = func(c context.Context, g *graphql.Request, i interface{}) error {
						bytesString := `{
							"result": {
								"package": {
									"apiDefinition": {
										"name": "apiDefName",
										"spec": {
											"data": "data",
											"format":"format"
										}
									}
								}
							}
						}`
						err := json.Unmarshal([]byte(bytesString), i)
						assert.NoError(t, err)
						return nil
					}
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			expectedSpec: &director.PackageSpecificationOutput{
				Name:   "apiDefName",
				Data:   &specData,
				Format: "format",
			},
			credentialsInput: &director.PackageSpecificationInput{
				ApplicationID: "appID",
				PackageID:     "packageID",
				DefinitionID:  "defID",
			},
			expectedQuery: "query {\n\t\t\t  result: application(id: \"appID\") {\n\t\t\t\t\t\tpackage(id: \"packageID\") {\n\t\t\t\t\t\t  apiDefinition(id: \"defID\") {\n\t\t\t\t\t\t\t  spec {\n\t\t\t\t\t\t\t\tdata\n\t\t\t\t\t\t\t\ttype\n\t\t\t\t\t\t\t\tformat\n\t\t\t\t\t\t\t  }\n\t\t\t\t\t\t  }\n\t\t\t\t\t\t  eventDefinition(id: \"defID\") {\n\t\t\t\t\t\t\t  spec {\n\t\t\t\t\t\t\t\tdata\n\t\t\t\t\t\t\t\ttype\n\t\t\t\t\t\t\t\tformat\n\t\t\t\t\t\t\t  }\n\t\t\t\t\t\t  }\n\t\t\t\t\t\t}\n\t\t\t\t\t  }\n\t\t\t\t\t}",
		},
		{
			name: "success when event spec",
			fields: fields{
				getGCLI: func() *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoStub = func(c context.Context, g *graphql.Request, i interface{}) error {
						bytesString := `{
							"result": {
								"package": {
									"eventDefinition": {
										"name": "eventDefName",
										"spec": {
											"data": "data",
											"format":"format"
										}
									}
								}
							}
						}`
						err := json.Unmarshal([]byte(bytesString), i)
						assert.NoError(t, err)
						return nil
					}
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			expectedSpec: &director.PackageSpecificationOutput{
				Name:   "eventDefName",
				Data:   &specData,
				Format: "format",
			},
			credentialsInput: &director.PackageSpecificationInput{
				ApplicationID: "appID",
				PackageID:     "packageID",
				DefinitionID:  "defID",
			},
			expectedQuery: "query {\n\t\t\t  result: application(id: \"appID\") {\n\t\t\t\t\t\tpackage(id: \"packageID\") {\n\t\t\t\t\t\t  apiDefinition(id: \"defID\") {\n\t\t\t\t\t\t\t  spec {\n\t\t\t\t\t\t\t\tdata\n\t\t\t\t\t\t\t\ttype\n\t\t\t\t\t\t\t\tformat\n\t\t\t\t\t\t\t  }\n\t\t\t\t\t\t  }\n\t\t\t\t\t\t  eventDefinition(id: \"defID\") {\n\t\t\t\t\t\t\t  spec {\n\t\t\t\t\t\t\t\tdata\n\t\t\t\t\t\t\t\ttype\n\t\t\t\t\t\t\t\tformat\n\t\t\t\t\t\t\t  }\n\t\t\t\t\t\t  }\n\t\t\t\t\t\t}\n\t\t\t\t\t  }\n\t\t\t\t\t}",
		},
		{
			name: "when gql client returns an error",
			fields: fields{
				getGCLI: func() *directorfakes.FakeClient {
					fakeGCLI := &directorfakes.FakeClient{}
					fakeGCLI.DoReturns(errors.New("some error"))
					return fakeGCLI
				},
				inputGraphqlizer:  inputGraphqlizer,
				outputGraphqlizer: outputGraphqlizer,
			},
			credentialsInput: &director.PackageSpecificationInput{
				ApplicationID: "appID",
				PackageID:     "packageID",
				DefinitionID:  "defID",
			},
			expectedErr: "while executing GraphQL call to get package instance auth: some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcli := tt.fields.getGCLI()
			c := director.NewGraphQLClient(
				gcli,
				tt.fields.inputGraphqlizer,
				tt.fields.outputGraphqlizer,
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
