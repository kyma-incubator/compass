package service

import (
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConversionServiceDetailsToApplicationRegisterInput(t *testing.T) {

	type testCase struct {
		given    model.ServiceDetails
		expected graphql.ApplicationRegisterInput
	}
	// GIVEN
	sut := NewConverter()
	// WHEN

	for name, tc := range map[string]testCase{
		"minimal number of fields set": {
			given: model.ServiceDetails{
				Name:        "serviceName",
				Description: "description",
			},
			expected: graphql.ApplicationRegisterInput{
				Name:        "serviceName",
				Description: ptrStringOrNilForEmpty("description"),
			},
		},
		"labels": {
			given: model.ServiceDetails{
				Labels: &map[string]string{"some-label": "some-value"},
			},
			expected: graphql.ApplicationRegisterInput{
				Labels: getLabelsOrNil(map[string]interface{}{"some-label": "some-value"}),
			},
		},
		"labels and our custom labels": {
			given: model.ServiceDetails{
				Labels:     &map[string]string{"some-label": "some-value"},
				Identifier: "identifier",
			},
			expected: graphql.ApplicationRegisterInput{
				Labels: getLabelsOrNil(map[string]interface{}{
					"some-label":            "some-value",
					unmappedFieldIdentifier: "identifier"}),
			},
		},
		"only our custom labels": {
			given: model.ServiceDetails{
				Identifier: "identifier",
			},
			expected: graphql.ApplicationRegisterInput{
				Labels: getLabelsOrNil(map[string]interface{}{unmappedFieldIdentifier: "identifier"}),
			},
		},
		"all basic attributes provided": {
			given: model.ServiceDetails{
				Identifier:       "identifier",
				Name:             "name",
				Description:      "description",
				Provider:         "provider",
				ShortDescription: "shortDescription",
			},
			expected: graphql.ApplicationRegisterInput{
				Name:         "name",
				Description:  ptrStringOrNilForEmpty("description"),
				ProviderName: ptrStringOrNilForEmpty("provider"),
				Labels: getLabelsOrNil(map[string]interface{}{
					unmappedFieldIdentifier:       "identifier",
					unmappedFieldShortDescription: "shortDescription",
				}),
			},
		},
		"API with only URL provided": {
			given: model.ServiceDetails{
				Api: &model.API{
					TargetUrl: "http://target.url",
				},
			},
			expected: graphql.ApplicationRegisterInput{
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
						// TODO what about name?
						TargetURL: "http://target.url",
					},
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
			expected: graphql.ApplicationRegisterInput{
				APIDefinitions: []*graphql.APIDefinitionInput{
					{TargetURL: "http://target.url",
						Spec: &graphql.APISpecInput{
							Type:   graphql.APISpecTypeOdata,
							Format: graphql.SpecFormatJSON,
						}},
				},
			},
		},

		"API other than ODATA provided": {
			given: model.ServiceDetails{
				Api: &model.API{
					ApiType: "anything else",
				},
			},
			expected: graphql.ApplicationRegisterInput{
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
						Spec: &graphql.APISpecInput{
							Type:   graphql.APISpecTypeOpenAPI,
							Format: graphql.SpecFormatJSON,
						},
					},
				},
			},
		},

		"API with directly spec provided": {
			given: model.ServiceDetails{
				Api: &model.API{
					Spec: json.RawMessage(`openapi: "3.0.0"`),
				},
			},
			expected: graphql.ApplicationRegisterInput{
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
						Spec: &graphql.APISpecInput{
							Data:   ptrClob(graphql.CLOB(`openapi: "3.0.0"`)),
							Type:   graphql.APISpecTypeOpenAPI,
							Format: graphql.SpecFormatJSON,
						},
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
			expected: graphql.ApplicationRegisterInput{
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
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
			expected: graphql.ApplicationRegisterInput{
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
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
			expected: graphql.ApplicationRegisterInput{
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
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
			expected: graphql.ApplicationRegisterInput{
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
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
			expected: graphql.ApplicationRegisterInput{
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
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
		},
		"API protected with certificate": {
			// TODO this is not mapped
		},
		//
		"API specification mapped to fetch request": {
			given: model.ServiceDetails{
				Api: &model.API{
					SpecificationUrl: "http://specification.url",
				},
			},
			expected: graphql.ApplicationRegisterInput{
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
						Spec: &graphql.APISpecInput{
							FetchRequest: &graphql.FetchRequestInput{
								URL: "http://specification.url",
							},
						},
					},
				},
			},
		},
		"API specification with basic to fetch request": {
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
			expected: graphql.ApplicationRegisterInput{
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
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
		},
		"API specification with oauth to fetch request": {
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
			expected: graphql.ApplicationRegisterInput{
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
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
		},
		"API specification with request parameters": {
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
			expected: graphql.ApplicationRegisterInput{
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
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
		},
		"event": {
			given: model.ServiceDetails{
				Events: &model.Events{
					Spec: json.RawMessage(`asyncapi: "1.2.0"`),
				},
			},
			expected: graphql.ApplicationRegisterInput{
				EventDefinitions: []*graphql.EventDefinitionInput{
					{
						//TODO what about name
						Spec: &graphql.EventSpecInput{
							Data: ptrClob(graphql.CLOB(`asyncapi: "1.2.0"`)),
						},
					},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			actual, err := sut.DetailsToGraphQLInput(tc.given)

			// THEN
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestConversionApplicationExtToServiceDetails(t *testing.T) {
	// GIVEN
	type testCase struct {
		given    graphql.ApplicationExt
		expected model.ServiceDetails
	}
	sut := NewConverter()
	// WHEN

	for name, tc := range map[string]testCase{
		"simple attributes": {
			given: graphql.ApplicationExt{
				Application: graphql.Application{
					ID:          "id",
					Name:        "name",
					Description: ptrStringOrNilForEmpty("description"),
					//TODO IntegrationSystemID
					ProviderName: ptrStringOrNilForEmpty("providerName"),
				},
			},
			expected: model.ServiceDetails{
				Name:        "name",
				Description: "description",
				Provider:    "providerName",
			},
		},
		"custom mapping labels": {
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
			actual, err := sut.GraphQLToDetailsModel(tc.given)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
	// THEN
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

		actualOut, err := sut.GraphQLToModel(givenIn)

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

		actualOut, err := sut.GraphQLToModel(givenIn)

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
