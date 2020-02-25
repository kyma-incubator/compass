package service_test

import (
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConversionDeprecatedServiceDetailsToGraphQLInput(t *testing.T) {

	type testCase struct {
		given    model.ServiceDetails
		expected model.GraphQLServiceDetailsInput
	}
	sut := service.NewConverter()

	for name, tc := range map[string]testCase{
		"input ID propagated to output": {
			given: model.ServiceDetails{},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
			},
		},
		"name and description propagated to api": {
			given: model.ServiceDetails{Name: "name", Description: "description", Api: &model.API{}},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					Name:        "name",
					Description: ptrString("description"),
				},
			},
		},
		"API with only URL provided": {
			given: model.ServiceDetails{
				Api: &model.API{
					TargetUrl: "http://target.url",
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					// TODO out name?
					TargetURL: "http://target.url",
				},
			},
		},
		"ODATA API provided": {
			given: model.ServiceDetails{
				Api: &model.API{
					TargetUrl: "http://target.url",
					ApiType:   "ODATA",
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					TargetURL: "http://target.url",
					Spec: &graphql.APISpecInput{
						Type:   graphql.APISpecTypeOdata,
						Format: graphql.SpecFormatYaml,
					},
				},
			},
		},

		"API other than ODATA provided": {
			given: model.ServiceDetails{
				Api: &model.API{
					ApiType: "anything else",
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					Spec: &graphql.APISpecInput{
						Type:   graphql.APISpecTypeOpenAPI,
						Format: graphql.SpecFormatYaml,
					},
				},
			},
		},

		"API with directly spec provided in YAML": {
			given: model.ServiceDetails{
				Api: &model.API{
					Spec: json.RawMessage(`openapi: "3.0.0"`),
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					Spec: &graphql.APISpecInput{
						Data:   ptrClob(graphql.CLOB(`openapi: "3.0.0"`)),
						Type:   graphql.APISpecTypeOpenAPI,
						Format: graphql.SpecFormatYaml,
					},
				},
			},
		},

		"API with directly spec provided in JSON": {
			given: model.ServiceDetails{
				Api: &model.API{
					Spec: json.RawMessage(`{"spec":"v0.0.1"}`),
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					Spec: &graphql.APISpecInput{
						Data:   ptrClob(graphql.CLOB(`{"spec":"v0.0.1"}`)),
						Type:   graphql.APISpecTypeOpenAPI,
						Format: graphql.SpecFormatJSON,
					},
				},
			},
		},

		"API with directly spec provided in XML": {
			given: model.ServiceDetails{
				Api: &model.API{
					Spec: json.RawMessage(`<spec></spec>"`),
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					Spec: &graphql.APISpecInput{
						Data:   ptrClob(graphql.CLOB(`<spec></spec>"`)),
						Type:   graphql.APISpecTypeOpenAPI,
						Format: graphql.SpecFormatXML,
					},
				},
			},
		},

		"API with query params and headers stored in old fields": {
			given: model.ServiceDetails{
				Api: &model.API{
					QueryParameters: &map[string][]string{
						"q1": {"a", "b"},
						"q2": {"c", "d"},
					},
					Headers: &map[string][]string{
						"h1": {"e", "f"},
						"h2": {"g", "h"},
					},
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					DefaultAuth: &graphql.AuthInput{
						AdditionalQueryParams: &graphql.QueryParams{
							"q1": {"a", "b"},
							"q2": {"c", "d"},
						},
						AdditionalHeaders: &graphql.HttpHeaders{
							"h1": {"e", "f"},
							"h2": {"g", "h"},
						},
					},
				},
			},
		},
		"API with query params and headers stored in the new fields": {
			given: model.ServiceDetails{
				Api: &model.API{
					RequestParameters: &model.RequestParameters{
						QueryParameters: &map[string][]string{
							"q1": {"a", "b"},
							"q2": {"c", "d"},
						},
						Headers: &map[string][]string{
							"h1": {"e", "f"},
							"h2": {"g", "h"},
						},
					},
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					DefaultAuth: &graphql.AuthInput{
						AdditionalQueryParams: &graphql.QueryParams{
							"q1": {"a", "b"},
							"q2": {"c", "d"},
						},
						AdditionalHeaders: &graphql.HttpHeaders{
							"h1": {"e", "f"},
							"h2": {"g", "h"},
						},
					},
				},
			},
		},
		"API with query params and headers stored in old and new fields": {
			given: model.ServiceDetails{
				Api: &model.API{
					RequestParameters: &model.RequestParameters{
						QueryParameters: &map[string][]string{
							"new": {"new"},
						},
						Headers: &map[string][]string{
							"new": {"new"},
						}},
					QueryParameters: &map[string][]string{
						"old": {"old"},
					},
					Headers: &map[string][]string{
						"old": {"old"},
					},
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					DefaultAuth: &graphql.AuthInput{
						AdditionalQueryParams: &graphql.QueryParams{
							"new": {"new"},
						},
						AdditionalHeaders: &graphql.HttpHeaders{
							"new": {"new"},
						},
					},
				},
			},
		},
		"API protected with basic": {
			given: model.ServiceDetails{
				Api: &model.API{
					Credentials: &model.CredentialsWithCSRF{
						BasicWithCSRF: &model.BasicAuthWithCSRF{
							BasicAuth: model.BasicAuth{
								Username: "user",
								Password: "password",
							},
						},
					},
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					DefaultAuth: &graphql.AuthInput{
						Credential: &graphql.CredentialDataInput{
							Basic: &graphql.BasicCredentialDataInput{
								Username: "user",
								Password: "password",
							},
						},
					},
				},
			},
		},
		"API protected with oauth": {
			given: model.ServiceDetails{
				Api: &model.API{
					Credentials: &model.CredentialsWithCSRF{
						OauthWithCSRF: &model.OauthWithCSRF{
							Oauth: model.Oauth{
								ClientID:     "client_id",
								ClientSecret: "client_secret",
								URL:          "http://oauth.url",
								RequestParameters: &model.RequestParameters{ // TODO this field is not mapped at all
									QueryParameters: &map[string][]string{
										"q1": {"a", "b"},
										"q2": {"c", "d"},
									},
									Headers: &map[string][]string{
										"h1": {"e", "f"},
										"h2": {"g", "h"},
									},
								},
							},
						},
					},
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					DefaultAuth: &graphql.AuthInput{
						Credential: &graphql.CredentialDataInput{
							Oauth: &graphql.OAuthCredentialDataInput{
								URL:          "http://oauth.url",
								ClientID:     "client_id",
								ClientSecret: "client_secret",
							},
						},
					},
				},
			},
		},
		"API specification mapped to fetch request": {
			given: model.ServiceDetails{
				Api: &model.API{
					SpecificationUrl: "http://specification.url",
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					Spec: &graphql.APISpecInput{
						FetchRequest: &graphql.FetchRequestInput{
							URL: "http://specification.url",
						},
						Format: graphql.SpecFormatJSON,
						Type:   graphql.APISpecTypeOpenAPI,
					},
				},
			},
		},
		"API specification with basic auth converted to fetch request": {
			given: model.ServiceDetails{
				Api: &model.API{
					SpecificationUrl: "http://specification.url",
					SpecificationCredentials: &model.Credentials{
						Basic: &model.BasicAuth{
							Username: "username",
							Password: "password",
						},
					},
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					Spec: &graphql.APISpecInput{
						FetchRequest: &graphql.FetchRequestInput{
							URL: "http://specification.url",
							Auth: &graphql.AuthInput{
								Credential: &graphql.CredentialDataInput{
									Basic: &graphql.BasicCredentialDataInput{
										Username: "username",
										Password: "password",
									},
								},
							},
						},
						Type:   graphql.APISpecTypeOpenAPI,
						Format: graphql.SpecFormatJSON,
					},
				},
			},
		},
		"API specification with oauth converted to fetch request": {
			given: model.ServiceDetails{
				Api: &model.API{
					SpecificationUrl: "http://specification.url",
					SpecificationCredentials: &model.Credentials{
						Oauth: &model.Oauth{
							URL:               "http://oauth.url",
							ClientID:          "client_id",
							ClientSecret:      "client_secret",
							RequestParameters: nil, // TODO not supported
						},
					},
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					Spec: &graphql.APISpecInput{
						FetchRequest: &graphql.FetchRequestInput{
							URL: "http://specification.url",
							Auth: &graphql.AuthInput{
								Credential: &graphql.CredentialDataInput{
									Oauth: &graphql.OAuthCredentialDataInput{
										URL:          "http://oauth.url",
										ClientID:     "client_id",
										ClientSecret: "client_secret",
									},
								},
							},
						},
						Type:   graphql.APISpecTypeOpenAPI,
						Format: graphql.SpecFormatJSON,
					},
				},
			},
		},
		"API specification with request parameters converted to fetch request": {
			given: model.ServiceDetails{
				Api: &model.API{
					SpecificationUrl: "http://specification.url",
					SpecificationRequestParameters: &model.RequestParameters{
						QueryParameters: &map[string][]string{
							"q1": {"a", "b"},
							"q2": {"c", "d"},
						},
						Headers: &map[string][]string{
							"h1": {"e", "f"},
							"h2": {"g", "h"},
						},
					},
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				API: &graphql.APIDefinitionInput{
					Spec: &graphql.APISpecInput{
						FetchRequest: &graphql.FetchRequestInput{
							URL: "http://specification.url",
							Auth: &graphql.AuthInput{
								AdditionalQueryParams: &graphql.QueryParams{
									"q1": {"a", "b"},
									"q2": {"c", "d"},
								},
								AdditionalHeaders: &graphql.HttpHeaders{
									"h1": {"e", "f"},
									"h2": {"g", "h"},
								},
							},
						},
						Format: graphql.SpecFormatJSON,
						Type:   graphql.APISpecTypeOpenAPI,
					},
				},
			},
		},
		"Event": {
			given: model.ServiceDetails{
				Events: &model.Events{
					Spec: json.RawMessage(`asyncapi: "1.2.0"`),
				},
			},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
				Event: &graphql.EventDefinitionInput{
					//TODO what about name
					Spec: &graphql.EventSpecInput{
						Data:   ptrClob(graphql.CLOB(`asyncapi: "1.2.0"`)),
						Type:   graphql.EventSpecTypeAsyncAPI,
						Format: graphql.SpecFormatYaml,
					},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			// WHEN
			actual, err := sut.DetailsToGraphQLInput("id", tc.given)

			// THEN
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestGraphQLToServiceDetails(t *testing.T) {

	type testCase struct {
		given    model.GraphQLServiceDetails
		expected model.ServiceDetails
	}
	sut := service.NewConverter()

	for name, tc := range map[string]testCase{
		"name and description is loaded from api/event": {
			given: model.GraphQLServiceDetails{
				API: &graphql.APIDefinitionExt{
					APIDefinition: graphql.APIDefinition{
						Name:        "name",
						Description: ptrString("description"),
					},
				},
			},
			expected: model.ServiceDetails{
				Name:        "name",
				Description: "description",
				Api:         &model.API{},
				Labels:      emptyLabels(),
			},
		},
		"simple API": {
			given: model.GraphQLServiceDetails{
				API: &graphql.APIDefinitionExt{
					APIDefinition: graphql.APIDefinition{
						TargetURL: "http://target.url",
					},
				},
			},
			expected: model.ServiceDetails{
				Api: &model.API{
					TargetUrl: "http://target.url",
				},
				Labels: emptyLabels(),
			},
		},
		"simple API with additional headers and query params": {
			given: model.GraphQLServiceDetails{
				API: &graphql.APIDefinitionExt{
					APIDefinition: graphql.APIDefinition{
						TargetURL: "http://target.url",
						DefaultAuth: &graphql.Auth{
							AdditionalQueryParams: &graphql.QueryParams{
								"q1": []string{"a", "b"},
								"q2": []string{"c", "d"},
							},
							AdditionalHeaders: &graphql.HttpHeaders{
								"h1": []string{"e", "f"},
								"h2": []string{"g", "h"},
							},
						},
					},
				},
			},
			expected: model.ServiceDetails{
				Api: &model.API{
					TargetUrl: "http://target.url",
					Headers: &map[string][]string{
						"h1": {"e", "f"},
						"h2": {"g", "h"}},
					QueryParameters: &map[string][]string{
						"q1": {"a", "b"},
						"q2": {"c", "d"},
					},
					RequestParameters: &model.RequestParameters{
						Headers: &map[string][]string{
							"h1": {"e", "f"},
							"h2": {"g", "h"}},
						QueryParameters: &map[string][]string{
							"q1": {"a", "b"},
							"q2": {"c", "d"},
						},
					},
				},
				Labels: emptyLabels(),
			},
		},
		"simple API with Basic Auth": {
			given: model.GraphQLServiceDetails{
				API: &graphql.APIDefinitionExt{
					APIDefinition: graphql.APIDefinition{
						TargetURL: "http://target.url",
						DefaultAuth: &graphql.Auth{
							Credential: &graphql.BasicCredentialData{
								Username: "username",
								Password: "password",
							},
						},
					},
				},
			},
			expected: model.ServiceDetails{
				Api: &model.API{
					TargetUrl: "http://target.url",
					Credentials: &model.CredentialsWithCSRF{
						BasicWithCSRF: &model.BasicAuthWithCSRF{
							BasicAuth: model.BasicAuth{
								Username: "username",
								Password: "password",
							},
						},
					},
				},
				Labels: emptyLabels(),
			},
		},
		"simple API with Oauth": {
			given: model.GraphQLServiceDetails{
				API: &graphql.APIDefinitionExt{

					APIDefinition: graphql.APIDefinition{
						TargetURL: "http://target.url",
						DefaultAuth: &graphql.Auth{
							Credential: &graphql.OAuthCredentialData{
								URL:          "http://oauth.url",
								ClientID:     "client_id",
								ClientSecret: "client_secret",
							},
						},
					},
				},
			},
			expected: model.ServiceDetails{
				Api: &model.API{
					TargetUrl: "http://target.url",
					Credentials: &model.CredentialsWithCSRF{
						OauthWithCSRF: &model.OauthWithCSRF{
							Oauth: model.Oauth{
								URL:          "http://oauth.url",
								ClientID:     "client_id",
								ClientSecret: "client_secret",
							},
						},
					},
				},
				Labels: emptyLabels(),
			},
		},
		"simple API with FetchRequest (query params and headers)": {
			given: model.GraphQLServiceDetails{
				API: &graphql.APIDefinitionExt{
					Spec: &graphql.APISpecExt{
						FetchRequest: &graphql.FetchRequest{
							URL: "http://apispec.url",
							Auth: &graphql.Auth{
								AdditionalQueryParams: &graphql.QueryParams{
									"q1": {"a", "b"},
									"q2": {"c", "d"},
								},
								AdditionalHeaders: &graphql.HttpHeaders{
									"h1": {"e", "f"},
									"h2": {"g", "h"},
								},
							},
						}}}},
			expected: model.ServiceDetails{
				Api: &model.API{
					SpecificationUrl: "http://apispec.url",
					SpecificationRequestParameters: &model.RequestParameters{
						Headers: &map[string][]string{
							"h1": {"e", "f"},
							"h2": {"g", "h"}},
						QueryParameters: &map[string][]string{
							"q1": {"a", "b"},
							"q2": {"c", "d"},
						},
					},
				},
				Labels: emptyLabels(),
			},
		},
		"simple API with Fetch Request protected with Basic Auth": {
			given: model.GraphQLServiceDetails{
				API: &graphql.APIDefinitionExt{

					Spec: &graphql.APISpecExt{
						FetchRequest: &graphql.FetchRequest{
							URL: "http://apispec.url",
							Auth: &graphql.Auth{
								Credential: &graphql.BasicCredentialData{
									Username: "username",
									Password: "password",
								},
							},
						}}}},
			expected: model.ServiceDetails{
				Api: &model.API{
					SpecificationUrl: "http://apispec.url",
					SpecificationCredentials: &model.Credentials{
						Basic: &model.BasicAuth{
							Username: "username",
							Password: "password",
						},
					},
				},
				Labels: emptyLabels(),
			},
		},
		"simple API with Fetch Request protected with Oauth": {
			given: model.GraphQLServiceDetails{
				API: &graphql.APIDefinitionExt{
					Spec: &graphql.APISpecExt{
						FetchRequest: &graphql.FetchRequest{
							URL: "http://apispec.url",
							Auth: &graphql.Auth{
								Credential: &graphql.OAuthCredentialData{
									URL:          "http://oauth.url",
									ClientID:     "client_id",
									ClientSecret: "client_secret",
								},
							},
						}}}},
			expected: model.ServiceDetails{
				Api: &model.API{
					SpecificationUrl: "http://apispec.url",
					SpecificationCredentials: &model.Credentials{
						Oauth: &model.Oauth{
							URL:          "http://oauth.url",
							ClientID:     "client_id",
							ClientSecret: "client_secret",
						},
					},
				},
				Labels: emptyLabels(),
			},
		},
		"events": {
			given: model.GraphQLServiceDetails{
				Event: &graphql.EventAPIDefinitionExt{
					Spec: &graphql.EventAPISpecExt{
						EventSpec: graphql.EventSpec{
							Data: ptrClob(`asyncapi: "1.2.0"`),
						},
					},
				},
			},
			expected: model.ServiceDetails{
				Events: &model.Events{
					Spec: json.RawMessage(`asyncapi: "1.2.0"`),
				},
				Labels: emptyLabels(),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			// WHEN
			actual, err := sut.GraphQLToServiceDetails(tc.given)
			// THEN
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestConverter_ServiceDetailsToService(t *testing.T) {
	//GIVEN
	input := model.ServiceDetails{
		Provider:         "provider",
		Name:             "name",
		Description:      "description",
		ShortDescription: "short description",
		Identifier:       "identifie",
		Labels:           &map[string]string{"blalb": "blalba"},
	}
	id := "id"

	//WHEN
	sut := service.NewConverter()
	output, err := sut.ServiceDetailsToService(input, id)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, input.Provider, output.Provider)
	assert.Equal(t, input.Name, output.Name)
	assert.Equal(t, input.Description, output.Description)
	assert.Equal(t, input.Identifier, output.Identifier)
	assert.Equal(t, input.Labels, output.Labels)
}

func TestConverter_DetailsToGraphQLInput_TestSpecsRecognition(t *testing.T) {
	// GIVEN
	sut := service.NewConverter()

	// API
	apiCases := []struct {
		Name           string
		InputAPI       model.API
		ExpectedType   graphql.APISpecType
		ExpectedFormat graphql.SpecFormat
	}{
		{
			Name:           "OpenAPI + YAML",
			InputAPI:       fixAPIOpenAPIYAML(),
			ExpectedType:   graphql.APISpecTypeOpenAPI,
			ExpectedFormat: graphql.SpecFormatYaml,
		},
		{
			Name:           "OpenAPI + JSON",
			InputAPI:       fixAPIOpenAPIJSON(),
			ExpectedType:   graphql.APISpecTypeOpenAPI,
			ExpectedFormat: graphql.SpecFormatJSON,
		},
		{
			Name:           "OData + XML",
			InputAPI:       fixAPIODataXML(),
			ExpectedType:   graphql.APISpecTypeOdata,
			ExpectedFormat: graphql.SpecFormatXML,
		},
	}

	for _, testCase := range apiCases {
		t.Run(testCase.Name, func(t *testing.T) {
			in := model.ServiceDetails{Api: &testCase.InputAPI}

			// WHEN
			out, err := sut.DetailsToGraphQLInput("id", in)

			// THEN
			require.NoError(t, err)
			require.NotNil(t, out.API)
			require.NotNil(t, out.API.Spec)
			assert.Equal(t, testCase.ExpectedType, out.API.Spec.Type)
			assert.Equal(t, testCase.ExpectedFormat, out.API.Spec.Format)
		})
	}

	// Events
	eventsCases := []struct {
		Name           string
		InputEvents    model.Events
		ExpectedType   graphql.EventSpecType
		ExpectedFormat graphql.SpecFormat
	}{
		{
			Name:           "Async API + JSON",
			InputEvents:    fixEventsAsyncAPIJSON(),
			ExpectedType:   graphql.EventSpecTypeAsyncAPI,
			ExpectedFormat: graphql.SpecFormatJSON,
		},
		{
			Name:           "Async API + YAML",
			InputEvents:    fixEventsAsyncAPIYAML(),
			ExpectedType:   graphql.EventSpecTypeAsyncAPI,
			ExpectedFormat: graphql.SpecFormatYaml,
		},
	}

	for _, testCase := range eventsCases {
		t.Run(testCase.Name, func(t *testing.T) {
			in := model.ServiceDetails{Events: &testCase.InputEvents}

			// WHEN
			out, err := sut.DetailsToGraphQLInput("id", in)

			// THEN
			require.NoError(t, err)
			require.NotNil(t, out.Event)
			require.NotNil(t, out.Event.Spec)
			assert.Equal(t, testCase.ExpectedType, out.Event.Spec.Type)
			assert.Equal(t, testCase.ExpectedFormat, out.Event.Spec.Format)
		})
	}
}

func emptyLabels() *map[string]string {
	return &map[string]string{}
}

func ptrString(in string) *string {
	return &in
}

func ptrClob(in graphql.CLOB) *graphql.CLOB {
	return &in
}
