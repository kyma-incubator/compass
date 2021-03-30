package service_test

import (
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_DetailsToGraphQLCreateInput(t *testing.T) {
	additionalQueryParamsSerialized := graphql.QueryParamsSerialized(`{"q1":["a","b"],"q2":["c","d"]}`)
	additionalHeadersSerialized := graphql.HttpHeadersSerialized(`{"h1":["e","f"],"h2":["g","h"]}`)

	type testCase struct {
		given    model.ServiceDetails
		expected graphql.BundleCreateInput
	}

	conv := service.NewConverter()

	for name, tc := range map[string]testCase{
		"name and description propagated to api": {
			given: model.ServiceDetails{Name: "name", Description: "description", Api: &model.API{}},
			expected: graphql.BundleCreateInput{
				Name:                "name",
				Description:         ptrString("description"),
				DefaultInstanceAuth: &graphql.AuthInput{},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
						Name:        "name",
						Description: ptrString("description"),
					},
				},
			},
		},
		"API with only URL provided": {
			given: model.ServiceDetails{
				Api: &model.API{
					TargetUrl: "http://target.url",
				},
			},
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
						TargetURL: "http://target.url",
					},
				},
			},
		},
		"API with empty credentials": {
			given: model.ServiceDetails{
				Api: &model.API{
					TargetUrl:   "http://target.url",
					Credentials: &model.CredentialsWithCSRF{},
				},
			},
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
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
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
						TargetURL: "http://target.url",
						Spec: &graphql.APISpecInput{
							Type:   graphql.APISpecTypeOdata,
							Format: graphql.SpecFormatXML,
							FetchRequest: &graphql.FetchRequestInput{
								URL: "http://target.url/$metadata",
							},
						},
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
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
						Spec: &graphql.APISpecInput{
							Type:   graphql.APISpecTypeOpenAPI,
							Format: graphql.SpecFormatYaml,
						},
					},
				},
			},
		},

		"API with directly spec provided in YAML": {
			given: model.ServiceDetails{
				Api: &model.API{
					Spec: ptrSpecResponse(`openapi: "3.0.0"`),
				},
			},
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
						Spec: &graphql.APISpecInput{
							Data:   ptrClob(`openapi: "3.0.0"`),
							Type:   graphql.APISpecTypeOpenAPI,
							Format: graphql.SpecFormatYaml,
						},
					},
				},
			},
		},

		"API with directly spec provided in JSON": {
			given: model.ServiceDetails{
				Api: &model.API{
					Spec: ptrSpecResponse(`{"spec":"v0.0.1"}`),
				},
			},
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
						Spec: &graphql.APISpecInput{
							Data:   ptrClob(`{"spec":"v0.0.1"}`),
							Type:   graphql.APISpecTypeOpenAPI,
							Format: graphql.SpecFormatJSON,
						},
					},
				},
			},
		},

		"API with directly spec provided in XML": {
			given: model.ServiceDetails{
				Api: &model.API{
					Spec: ptrSpecResponse(`<spec></spec>"`),
				},
			},
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
						Spec: &graphql.APISpecInput{
							Data:   ptrClob(`<spec></spec>"`),
							Type:   graphql.APISpecTypeOpenAPI,
							Format: graphql.SpecFormatXML,
						},
					},
				},
			},
		},

		"API with query params and headers stored in old fields": {
			given: model.ServiceDetails{
				Name: "foo",
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
			expected: graphql.BundleCreateInput{
				Name: "foo",
				DefaultInstanceAuth: &graphql.AuthInput{
					AdditionalQueryParamsSerialized: &additionalQueryParamsSerialized,
					AdditionalHeadersSerialized:     &additionalHeadersSerialized,
				},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{Name: "foo"},
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
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{
					AdditionalQueryParamsSerialized: &additionalQueryParamsSerialized,
					AdditionalHeadersSerialized:     &additionalHeadersSerialized,
				},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{},
				},
			},
		},
		"API with query params and headers stored in old and new fields": {
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
						}},
					QueryParameters: &map[string][]string{
						"old": {"old"},
					},
					Headers: &map[string][]string{
						"old": {"old"},
					},
				},
			},
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{
					AdditionalQueryParamsSerialized: &additionalQueryParamsSerialized,
					AdditionalHeadersSerialized:     &additionalHeadersSerialized,
				},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{},
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
							CSRFInfo: &model.CSRFInfo{TokenEndpointURL: "foo.bar"},
						},
					},
				},
			},
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{
					Credential: &graphql.CredentialDataInput{
						Basic: &graphql.BasicCredentialDataInput{
							Username: "user",
							Password: "password",
						},
					},
					RequestAuth: &graphql.CredentialRequestAuthInput{
						Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{TokenEndpointURL: "foo.bar"},
					},
				},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{},
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
							CSRFInfo: &model.CSRFInfo{TokenEndpointURL: "foo.bar"},
						},
					},
				},
			},
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{
					Credential: &graphql.CredentialDataInput{
						Oauth: &graphql.OAuthCredentialDataInput{
							URL:          "http://oauth.url",
							ClientID:     "client_id",
							ClientSecret: "client_secret",
						},
					},
					RequestAuth: &graphql.CredentialRequestAuthInput{
						Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{TokenEndpointURL: "foo.bar"},
					},
				},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{},
				},
			},
		},
		"API specification mapped to fetch request": {
			given: model.ServiceDetails{
				Api: &model.API{
					SpecificationUrl: "http://specification.url",
				},
			},
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
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
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{},
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
							Type:   graphql.APISpecTypeOpenAPI,
							Format: graphql.SpecFormatJSON,
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
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{},
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
							Type:   graphql.APISpecTypeOpenAPI,
							Format: graphql.SpecFormatJSON,
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
			expected: graphql.BundleCreateInput{
				DefaultInstanceAuth: &graphql.AuthInput{},
				APIDefinitions: []*graphql.APIDefinitionInput{
					{
						Spec: &graphql.APISpecInput{
							FetchRequest: &graphql.FetchRequestInput{
								URL: "http://specification.url",
								Auth: &graphql.AuthInput{
									AdditionalQueryParamsSerialized: &additionalQueryParamsSerialized,
									AdditionalHeadersSerialized:     &additionalHeadersSerialized,
								},
							},
							Format: graphql.SpecFormatJSON,
							Type:   graphql.APISpecTypeOpenAPI,
						},
					},
				},
			},
		},
		"Event": {
			given: model.ServiceDetails{
				Name: "foo",
				Events: &model.Events{
					Spec: ptrSpecResponse(`asyncapi: "1.2.0"`),
				},
			},
			expected: graphql.BundleCreateInput{
				Name:                "foo",
				DefaultInstanceAuth: &graphql.AuthInput{},
				EventDefinitions: []*graphql.EventDefinitionInput{
					{
						Name: "foo",
						Spec: &graphql.EventSpecInput{
							Data:   ptrClob(`asyncapi: "1.2.0"`),
							Type:   graphql.EventSpecTypeAsyncAPI,
							Format: graphql.SpecFormatYaml,
						},
					},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			// WHEN
			actual, err := conv.DetailsToGraphQLCreateInput(tc.given)

			// THEN
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestConverter_GraphQLToServiceDetails(t *testing.T) {
	type testCase struct {
		given    graphql.BundleExt
		expected model.ServiceDetails
	}
	conv := service.NewConverter()

	testSvcRef := service.LegacyServiceReference{
		ID:         "foo",
		Identifier: "",
	}

	for name, tc := range map[string]testCase{
		"name and description is loaded from Bundle": {
			given: graphql.BundleExt{
				Bundle: graphql.Bundle{Name: "foo", Description: ptrString("description")},
				APIDefinitions: graphql.APIDefinitionPageExt{
					Data: []*graphql.APIDefinitionExt{
						{
							APIDefinition: graphql.APIDefinition{},
						},
					},
				},
			},
			expected: model.ServiceDetails{
				Name:        "foo",
				Description: "description",
				Api:         &model.API{},
				Labels:      emptyLabels(),
			},
		},
		"simple API": {
			given: graphql.BundleExt{
				APIDefinitions: graphql.APIDefinitionPageExt{
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
				Labels: emptyLabels(),
			},
		},
		"simple API with additional headers and query params": {
			given: graphql.BundleExt{
				Bundle: graphql.Bundle{
					DefaultInstanceAuth: &graphql.Auth{
						AdditionalQueryParams: graphql.QueryParams{
							"q1": []string{"a", "b"},
							"q2": []string{"c", "d"},
						},
						AdditionalHeaders: graphql.HttpHeaders{
							"h1": []string{"e", "f"},
							"h2": []string{"g", "h"},
						},
					},
				},
				APIDefinitions: graphql.APIDefinitionPageExt{
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
			given: graphql.BundleExt{
				Bundle: graphql.Bundle{
					DefaultInstanceAuth: &graphql.Auth{
						Credential: &graphql.BasicCredentialData{
							Username: "username",
							Password: "password",
						},
					},
				},
				APIDefinitions: graphql.APIDefinitionPageExt{
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
			given: graphql.BundleExt{
				Bundle: graphql.Bundle{
					DefaultInstanceAuth: &graphql.Auth{
						Credential: &graphql.OAuthCredentialData{
							URL:          "http://oauth.url",
							ClientID:     "client_id",
							ClientSecret: "client_secret",
						},
					},
				},
				APIDefinitions: graphql.APIDefinitionPageExt{
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
			given: graphql.BundleExt{
				APIDefinitions: graphql.APIDefinitionPageExt{
					Data: []*graphql.APIDefinitionExt{
						{
							Spec: &graphql.APISpecExt{
								FetchRequest: &graphql.FetchRequest{
									URL: "http://apispec.url",
									Auth: &graphql.Auth{
										AdditionalQueryParams: graphql.QueryParams{
											"q1": {"a", "b"},
											"q2": {"c", "d"},
										},
										AdditionalHeaders: graphql.HttpHeaders{
											"h1": {"e", "f"},
											"h2": {"g", "h"},
										},
									},
								}}},
					},
				}},
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
			given: graphql.BundleExt{
				APIDefinitions: graphql.APIDefinitionPageExt{
					Data: []*graphql.APIDefinitionExt{
						{
							Spec: &graphql.APISpecExt{
								FetchRequest: &graphql.FetchRequest{
									URL: "http://apispec.url",
									Auth: &graphql.Auth{
										Credential: &graphql.BasicCredentialData{
											Username: "username",
											Password: "password",
										},
									},
								}}},
					},
				}},
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
			given: graphql.BundleExt{
				APIDefinitions: graphql.APIDefinitionPageExt{
					Data: []*graphql.APIDefinitionExt{
						{
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
								}}},
					},
				}},
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
			given: graphql.BundleExt{
				EventDefinitions: graphql.EventAPIDefinitionPageExt{
					Data: []*graphql.EventAPIDefinitionExt{{
						Spec: &graphql.EventAPISpecExt{
							EventSpec: graphql.EventSpec{
								Data: ptrClob(`{"asyncapi": "1.2.0"}`),
							},
						}},
					},
				},
			},
			expected: model.ServiceDetails{
				Events: &model.Events{
					Spec: ptrSpecResponse(`{"asyncapi": "1.2.0"}`),
				},
				Labels: emptyLabels(),
			},
		},
		"events with XML spec": {
			given: graphql.BundleExt{
				EventDefinitions: graphql.EventAPIDefinitionPageExt{
					Data: []*graphql.EventAPIDefinitionExt{{
						Spec: &graphql.EventAPISpecExt{
							EventSpec: graphql.EventSpec{
								Data: ptrClob(`<xml></xml>`),
							},
						}},
					},
				},
			},
			expected: model.ServiceDetails{
				Events: &model.Events{
					Spec: ptrSpecResponse(`<xml></xml>`),
				},
				Labels: emptyLabels(),
			},
		},
		"API with XML spec": {
			given: graphql.BundleExt{
				APIDefinitions: graphql.APIDefinitionPageExt{
					Data: []*graphql.APIDefinitionExt{{
						Spec: &graphql.APISpecExt{
							APISpec: graphql.APISpec{
								Data: ptrClob(`<xml></xml>`),
							},
						}},
					},
				},
			},
			expected: model.ServiceDetails{
				Api: &model.API{
					Spec: ptrSpecResponse(`<xml></xml>`),
				},
				Labels: emptyLabels(),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			// WHEN
			actual, err := conv.GraphQLToServiceDetails(tc.given, testSvcRef)
			// THEN
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
			_, err = json.Marshal(actual)
			require.NoError(t, err)
		})
	}

	t.Run("identifier provided", func(t *testing.T) {
		in := graphql.BundleExt{
			APIDefinitions: graphql.APIDefinitionPageExt{
				Data: []*graphql.APIDefinitionExt{
					{
						APIDefinition: graphql.APIDefinition{
							TargetURL: "http://target.url",
						},
					},
				},
			},
		}
		inSvcRef := service.LegacyServiceReference{
			ID:         "foo",
			Identifier: "test",
		}
		expected := model.ServiceDetails{
			Identifier: "test",
			Api: &model.API{
				TargetUrl: "http://target.url",
			},
			Labels: emptyLabels(),
		}
		// WHEN
		actual, err := conv.GraphQLToServiceDetails(in, inSvcRef)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestConverter_ServiceDetailsToService(t *testing.T) {
	//GIVEN
	input := model.ServiceDetails{
		Provider:         "provider",
		Name:             "name",
		Description:      "description",
		ShortDescription: "short description",
		Identifier:       "identifier",
		Labels:           &map[string]string{"blalb": "blalba"},
	}
	id := "id"

	//WHEN
	conv := service.NewConverter()
	output, err := conv.ServiceDetailsToService(input, id)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, input.Provider, output.Provider)
	assert.Equal(t, input.Name, output.Name)
	assert.Equal(t, input.Description, output.Description)
	assert.Equal(t, input.Identifier, output.Identifier)
	assert.Equal(t, input.Labels, output.Labels)
}

func TestConverter_GraphQLCreateInputToUpdateInput(t *testing.T) {
	desc := "Desc"
	schema := graphql.JSONSchema("foo")
	auth := graphql.AuthInput{Credential: &graphql.CredentialDataInput{Basic: &graphql.BasicCredentialDataInput{
		Username: "foo",
		Password: "bar",
	}}}
	in := graphql.BundleCreateInput{
		Name:                           "foo",
		Description:                    &desc,
		InstanceAuthRequestInputSchema: &schema,
		DefaultInstanceAuth:            &auth,
	}
	expected := graphql.BundleUpdateInput{
		Name:                           "foo",
		Description:                    &desc,
		InstanceAuthRequestInputSchema: &schema,
		DefaultInstanceAuth:            &auth,
	}

	conv := service.NewConverter()

	res := conv.GraphQLCreateInputToUpdateInput(in)

	assert.Equal(t, expected, res)
}

func TestConverter_DetailsToGraphQLInput_TestSpecsRecognition(t *testing.T) {
	// GIVEN
	conv := service.NewConverter()

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
			out, err := conv.DetailsToGraphQLCreateInput(in)

			// THEN
			require.NoError(t, err)
			require.Len(t, out.APIDefinitions, 1)
			require.NotNil(t, out.APIDefinitions[0].Spec)
			assert.Equal(t, testCase.ExpectedType, out.APIDefinitions[0].Spec.Type)
			assert.Equal(t, testCase.ExpectedFormat, out.APIDefinitions[0].Spec.Format)
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
			out, err := conv.DetailsToGraphQLCreateInput(in)

			// THEN
			require.NoError(t, err)
			require.Len(t, out.EventDefinitions, 1)
			require.NotNil(t, out.EventDefinitions[0].Spec)
			assert.Equal(t, testCase.ExpectedType, out.EventDefinitions[0].Spec.Type)
			assert.Equal(t, testCase.ExpectedFormat, out.EventDefinitions[0].Spec.Format)
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

func ptrSpecResponse(in model.SpecResponse) *model.SpecResponse {
	return &in
}
