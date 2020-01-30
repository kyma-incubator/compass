package service

import (
	"encoding/json"
	"testing"

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
	sut := NewConverter()

	for name, tc := range map[string]testCase{
		"input ID propagated to output": {
			given: model.ServiceDetails{},
			expected: model.GraphQLServiceDetailsInput{
				ID: "id",
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
						Type: graphql.APISpecTypeOdata,
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
						Type: graphql.APISpecTypeOpenAPI,
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
						Data: ptrClob(graphql.CLOB(`asyncapi: "1.2.0"`)),
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

func TestConversionApplicationExtToServiceDetails2(t *testing.T) {

	type testCase struct {
		given    graphql.ApplicationExt
		expected model.ServiceDetails
	}
	sut := NewConverter()

	for name, tc := range map[string]testCase{
		"simple attributes": {
			given: graphql.ApplicationExt{
				Application: graphql.Application{
					ID:           "id",
					Name:         "name",
					Description:  ptrStringOrNilForEmpty("description"),
					ProviderName: ptrStringOrNilForEmpty("providerName"),
				},
			},
			expected: model.ServiceDetails{
				Name:        "name",
				Description: "description",
				Provider:    "providerName",
			},
		},
		"custom labels": {
			given: graphql.ApplicationExt{
				Labels: graphql.Labels{
					unmappedFieldIdentifier:       "identifier",
					unmappedFieldShortDescription: "short description",
				},
			},
			expected: model.ServiceDetails{
				Identifier:       "identifier",
				ShortDescription: "short description",
			},
		},
		"labels": {
			given: graphql.ApplicationExt{
				Labels: graphql.Labels{
					"label-a": "a",
					"label-b": "b",
				},
			},
			expected: model.ServiceDetails{
				Labels: &map[string]string{
					"label-a": "a",
					"label-b": "b",
				},
			},
		},
		"simple API": {
			given: graphql.ApplicationExt{
				APIDefinitions: graphql.APIDefinitionPageExt{
					APIDefinitionPage: fixPageWithTotalCountOne(),
					Data: []*graphql.APIDefinitionExt{
						{
							APIDefinition: graphql.APIDefinition{
								TargetURL: "http://target.url",
							},
						},
					},
				},
			},
			expected: model.ServiceDetails{
				Api: &model.API{
					TargetUrl: "http://target.url",
				},
			},
		},
		"simple API with additional headers and query params": {
			given: graphql.ApplicationExt{
				APIDefinitions: graphql.APIDefinitionPageExt{
					APIDefinitionPage: fixPageWithTotalCountOne(),
					Data: []*graphql.APIDefinitionExt{
						{
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
			},
		},
		"simple API with Basic Auth": {
			given: graphql.ApplicationExt{
				APIDefinitions: graphql.APIDefinitionPageExt{
					APIDefinitionPage: fixPageWithTotalCountOne(),
					Data: []*graphql.APIDefinitionExt{
						{
							APIDefinition: graphql.APIDefinition{
								TargetURL: "http://target.url",
								DefaultAuth: &graphql.Auth{
									Credential: graphql.BasicCredentialData{
										Username: "username",
										Password: "password",
									},
								},
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
			},
		},
		"simple API with Oauth": {
			given: graphql.ApplicationExt{
				APIDefinitions: graphql.APIDefinitionPageExt{
					APIDefinitionPage: fixPageWithTotalCountOne(),
					Data: []*graphql.APIDefinitionExt{
						{
							APIDefinition: graphql.APIDefinition{
								TargetURL: "http://target.url",
								DefaultAuth: &graphql.Auth{
									Credential: graphql.OAuthCredentialData{
										URL:          "http://oauth.url",
										ClientID:     "client_id",
										ClientSecret: "client_secret",
									},
								},
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
			},
		},
		"simple API with FetchRequest (query params and headers)": {
			given: graphql.ApplicationExt{
				APIDefinitions: graphql.APIDefinitionPageExt{
					APIDefinitionPage: fixPageWithTotalCountOne(),
					Data: []*graphql.APIDefinitionExt{
						{
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
			}},
		"simple API with Fetch Request protected with Basic Auth": {
			given: graphql.ApplicationExt{
				APIDefinitions: graphql.APIDefinitionPageExt{
					APIDefinitionPage: fixPageWithTotalCountOne(),
					Data: []*graphql.APIDefinitionExt{
						{
							Spec: &graphql.APISpecExt{
								FetchRequest: &graphql.FetchRequest{
									URL: "http://apispec.url",
									Auth: &graphql.Auth{
										Credential: graphql.BasicCredentialData{
											Username: "username",
											Password: "password",
										},
									},
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
			},
		},
		"simple API with Fetch Request protected with Oauth": {
			given: graphql.ApplicationExt{
				APIDefinitions: graphql.APIDefinitionPageExt{
					APIDefinitionPage: fixPageWithTotalCountOne(),
					Data: []*graphql.APIDefinitionExt{
						{
							Spec: &graphql.APISpecExt{
								FetchRequest: &graphql.FetchRequest{
									URL: "http://apispec.url",
									Auth: &graphql.Auth{
										Credential: graphql.OAuthCredentialData{
											URL:          "http://oauth.url",
											ClientID:     "client_id",
											ClientSecret: "client_secret",
										},
									},
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
			},
		},
		"events": {
			given: graphql.ApplicationExt{
				EventDefinitions: graphql.EventAPIDefinitionPageExt{
					EventDefinitionPage: graphql.EventDefinitionPage{
						TotalCount: 1,
					},
					Data: []*graphql.EventAPIDefinitionExt{
						{
							Spec: &graphql.EventAPISpecExt{
								EventSpec: graphql.EventSpec{
									Data: ptrClob(graphql.CLOB(`asyncapi: "1.2.0"`)),
								},
							},
						},
					},
				},
			},
			expected: model.ServiceDetails{
				Events: &model.Events{
					Spec: json.RawMessage(`asyncapi: "1.2.0"`),
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			// WHEN
			actual, err := sut.GraphQLToDetailsModel(tc.given)
			// THEN
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}

}

func Test2(t *testing.T) {

	type testCase struct {
		given    model.GraphQLServiceDetails
		expected model.ServiceDetails
	}
	sut := NewConverter()

	for name, tc := range map[string]testCase{

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
			},
		},
		"simple API with Basic Auth": {
			given: model.GraphQLServiceDetails{
				API: &graphql.APIDefinitionExt{
					APIDefinition: graphql.APIDefinition{
						TargetURL: "http://target.url",
						DefaultAuth: &graphql.Auth{
							Credential: graphql.BasicCredentialData{
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
			},
		},
		"simple API with Oauth": {
			given: model.GraphQLServiceDetails{
				API: &graphql.APIDefinitionExt{

					APIDefinition: graphql.APIDefinition{
						TargetURL: "http://target.url",
						DefaultAuth: &graphql.Auth{
							Credential: graphql.OAuthCredentialData{
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
			},
		},
		"simple API with Fetch Request protected with Basic Auth": {
			given: model.GraphQLServiceDetails{
				API: &graphql.APIDefinitionExt{

					Spec: &graphql.APISpecExt{
						FetchRequest: &graphql.FetchRequest{
							URL: "http://apispec.url",
							Auth: &graphql.Auth{
								Credential: graphql.BasicCredentialData{
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
			},
		},
		"simple API with Fetch Request protected with Oauth": {
			given: model.GraphQLServiceDetails{
				API: &graphql.APIDefinitionExt{
					Spec: &graphql.APISpecExt{
						FetchRequest: &graphql.FetchRequest{
							URL: "http://apispec.url",
							Auth: &graphql.Auth{
								Credential: graphql.OAuthCredentialData{
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
			},
		},
		"events": {
			given: model.GraphQLServiceDetails{
				Event: &graphql.EventDefinition{
					Spec: &graphql.EventSpec{
						Data: ptrClob(graphql.CLOB(`asyncapi: "1.2.0"`)),
					},
				},
			},
			expected: model.ServiceDetails{
				Events: &model.Events{
					Spec: json.RawMessage(`asyncapi: "1.2.0"`),
				},
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

func TestConvertGraphQLToModel(t *testing.T) {

	t.Run("all fields provided", func(t *testing.T) {
		// GIVEN
		var err error

		givenIn := graphql.ApplicationExt{
			Application: graphql.Application{
				ID:           "id",
				Name:         "name",
				Description:  ptrStringOrNilForEmpty("description"),
				ProviderName: ptrStringOrNilForEmpty("providerName"),
			},
		}
		givenIn.Labels, err = fixLabels()
		givenIn.Labels[unmappedFieldIdentifier] = "identifier-dont-confuse-with-id-please"
		require.NoError(t, err)

		expectedOut := model.Service{
			ID:          "id",
			Name:        "name",
			Provider:    "providerName",
			Description: "description",
			Identifier:  "identifier-dont-confuse-with-id-please",
			Labels: &map[string]string{
				"simple-label": "simple-value",
			},
		}
		sut := NewConverter()

		// WHEN
		actualOut, err := sut.GraphQLToModel(givenIn)
		// THEN

		require.NoError(t, err)
		assert.Equal(t, expectedOut, actualOut)
	})

	t.Run("only required fields provided", func(t *testing.T) {
		// GIVEN
		var err error

		givenIn := graphql.ApplicationExt{
			Application: graphql.Application{
				ID:   "id",
				Name: "name",
			},
		}

		expectedOut := model.Service{
			ID:   "id",
			Name: "name",
		}
		sut := NewConverter()
		// WHEN
		actualOut, err := sut.GraphQLToModel(givenIn)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedOut, actualOut)
	})

}

func fixLabels() (graphql.Labels, error) {
	l := graphql.Labels{}
	j := `{ "ignored-group": ["production", "experimental"], "ignored-scenarios": ["DEFAULT"], "simple-label":"simple-value", "embedded":{"key":"value"} }`
	err := json.Unmarshal(([]byte)(j), &l)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func fixPageWithTotalCountOne() graphql.APIDefinitionPage {
	return graphql.APIDefinitionPage{TotalCount: 1}
}
