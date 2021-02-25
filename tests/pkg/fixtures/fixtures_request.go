package fixtures

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

const (
	auditlogTokenEndpoint        = "audit-log/v2/oauth/token"
	auditlogSearchEndpoint       = "audit-log/v2/configuration-changes/search"
	auditlogDeleteEndpointFormat = "audit-log/v2/configuration-changes/%s"

	webhookURL = "https://kyma-project.io"

	integrationSystemID = "69230297-3c81-4711-aac2-3afa8cb42e2d"
)

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

//Auth
func FixBasicAuth(t *testing.T) *graphql.AuthInput {
	additionalHeaders, err := graphql.NewHttpHeadersSerialized(map[string][]string{
		"header-A": []string{"ha1", "ha2"},
		"header-B": []string{"hb1", "hb2"},
	})
	require.NoError(t, err)

	additionalQueryParams, err := graphql.NewQueryParamsSerialized(map[string][]string{
		"qA": []string{"qa1", "qa2"},
		"qB": []string{"qb1", "qb2"},
	})
	require.NoError(t, err)

	return &graphql.AuthInput{
		Credential:                      FixBasicCredential(),
		AdditionalHeadersSerialized:     &additionalHeaders,
		AdditionalQueryParamsSerialized: &additionalQueryParams,
	}
}

func FixOauthAuth() *graphql.AuthInput {
	return &graphql.AuthInput{
		Credential: FixOAuthCredential(),
	}
}

func FixBasicCredential() *graphql.CredentialDataInput {
	return &graphql.CredentialDataInput{
		Basic: &graphql.BasicCredentialDataInput{
			Username: "admin",
			Password: "secret",
		}}
}

func FixOAuthCredential() *graphql.CredentialDataInput {
	return &graphql.CredentialDataInput{
		Oauth: &graphql.OAuthCredentialDataInput{
			URL:          "url.net",
			ClientSecret: "grazynasecret",
			ClientID:     "clientid",
		}}
}

func FixDeprecatedVersion() *graphql.VersionInput {
	return &graphql.VersionInput{
		Value:           "v1",
		Deprecated:      ptr.Bool(true),
		ForRemoval:      ptr.Bool(false),
		DeprecatedSince: ptr.String("v5"),
	}
}

func FixDecommissionedVersion() *graphql.VersionInput {
	return &graphql.VersionInput{
		Value:      "v1",
		Deprecated: ptr.Bool(true),
		ForRemoval: ptr.Bool(true),
	}
}

func FixActiveVersion() *graphql.VersionInput {
	return &graphql.VersionInput{
		Value:      "v2",
		Deprecated: ptr.Bool(false),
		ForRemoval: ptr.Bool(false),
	}
}

// Application

func FixSampleApplicationRegisterInputWithName(placeholder, name string) graphql.ApplicationRegisterInput {
	sampleInput := FixSampleApplicationRegisterInput(placeholder)
	sampleInput.Name = name
	return sampleInput
}

func FixSampleApplicationRegisterInput(placeholder string) graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name:         placeholder,
		ProviderName: ptr.String("compass"),
		Labels:       &graphql.Labels{placeholder: []interface{}{placeholder}},
	}
}

func FixSampleApplicationRegisterInputWithWebhooks(placeholder string) graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name:         placeholder,
		ProviderName: ptr.String("compass"),
		Webhooks: []*graphql.WebhookInput{{
			Type: graphql.WebhookTypeConfigurationChanged,
			URL:  ptr.String(webhookURL),
		},
		},
	}
}

func FixSampleApplicationRegisterInputWithNameAndWebhooks(placeholder, name string) graphql.ApplicationRegisterInput {
	sampleInput := FixSampleApplicationRegisterInputWithWebhooks(placeholder)
	sampleInput.Name = name
	return sampleInput
}

func FixSampleApplicationCreateInputWithIntegrationSystem(placeholder string) graphql.ApplicationRegisterInput {
	sampleInput := FixSampleApplicationRegisterInputWithWebhooks(placeholder)
	sampleInput.IntegrationSystemID = ptr.String(integrationSystemID)
	return sampleInput
}

func FixSampleApplicationUpdateInput(placeholder string) graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Description:    &placeholder,
		HealthCheckURL: ptr.String(webhookURL),
		ProviderName:   &placeholder,
	}
}

func FixSampleApplicationUpdateInputWithIntegrationSystem(placeholder string) graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Description:         &placeholder,
		HealthCheckURL:      ptr.String(webhookURL),
		IntegrationSystemID: ptr.String(integrationSystemID),
		ProviderName:        ptr.String(placeholder),
	}
}

func FixApplicationTemplate(name string) graphql.ApplicationTemplateInput {
	appTemplateDesc := "app-template-desc"
	placeholderDesc := "new-placeholder-desc"
	providerName := "compass-tests"
	appTemplateInput := graphql.ApplicationTemplateInput{
		Name:        name,
		Description: &appTemplateDesc,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:         "app",
			ProviderName: &providerName,
			Description:  ptr.String("test {{new-placeholder}}"),
			Labels: &graphql.Labels{
				"a": []string{"b", "c"},
				"d": []string{"e", "f"},
			},
			Webhooks: []*graphql.WebhookInput{{
				Type: graphql.WebhookTypeConfigurationChanged,
				URL:  ptr.String("http://url.com"),
			}},
			HealthCheckURL: ptr.String("http://url.valid"),
		},
		Placeholders: []*graphql.PlaceholderDefinitionInput{
			{
				Name:        "new-placeholder",
				Description: &placeholderDesc,
			},
		},
		AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal,
	}
	return appTemplateInput
}

func FixCreateApplicationTemplateRequest(applicationTemplateInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplicationTemplate(in: %s) {
					%s
				}
			}`,
			applicationTemplateInGQL, testctx.Tc.GQLFieldsProvider.ForApplicationTemplate()))
}

func FixApplicationRegisterInputWithBundles(t *testing.T) graphql.ApplicationRegisterInput {
	bndl1 := FixBundleCreateInputWithRelatedObjects(t, "foo")
	bndl2 := FixBundleCreateInputWithRelatedObjects(t, "bar")
	return graphql.ApplicationRegisterInput{
		Name:         "create-application-with-documents",
		ProviderName: ptr.String("compass"),
		Bundles: []*graphql.BundleCreateInput{
			&bndl1, &bndl2,
		},
		Labels: &graphql.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}
}

func FixRegisterApplicationRequest(applicationInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplication(in: %s) {
					%s
				}
			}`,
			applicationInGQL, testctx.Tc.GQLFieldsProvider.ForApplication()))
}

func FixGetApplicationRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForApplication()))
}

func FixUpdateApplicationRequest(id, updateInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: updateApplication(id: "%s", in: %s) {
    					%s
					}
				}`, id, updateInputGQL, testctx.Tc.GQLFieldsProvider.ForApplication()))
}

func FixUnregisterApplicationRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		unregisterApplication(id: "%s") {
			%s
		}	
	}`, id, testctx.Tc.GQLFieldsProvider.ForApplication()))
}

func FixApplicationTemplateRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: applicationTemplate(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForApplicationTemplate()))
}

func FixUpdateApplicationTemplateRequest(id, updateInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: updateApplicationTemplate(id: "%s", in: %s) {
    					%s
					}
				}`, id, updateInputGQL, testctx.Tc.GQLFieldsProvider.ForApplicationTemplate()))
}

func FixDeleteApplicationTemplateRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteApplicationTemplate(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForApplicationTemplate()))
}

func FixRequestClientCredentialsForApplication(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestClientCredentialsForApplication(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForSystemAuth()))
}

func FixRequestOneTimeTokenForApplication(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestOneTimeTokenForApplication(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForOneTimeTokenForApplication()))
}

func FixApplicationForRuntimeRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
  			result: applicationsForRuntime(runtimeID: "%s", first:%d, after:"") { 
					%s 
				}
			}`, runtimeID, 4, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForApplication())))
}

func FixGetApplicationsRequestWithPagination() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications {
						%s
					}
				}`,
			testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForApplication())))
}

func FixApplicationsFilteredPageableRequest(labelFilterInGQL string, first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications(filter: %s, first: %d, after: "%s") {
						%s
					}
				}`,
			labelFilterInGQL, first, after, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForApplication())))
}

func FixApplicationsPageableRequest(first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications(first: %d, after: "%s") {
						%s
					}
				}`,
			first, after, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForApplication())))
}

func FixGetApplicationTemplatesWithPagination(first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applicationTemplates(first: %d, after: "%s") {
						%s
					}
				}`,
			first, after, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForApplicationTemplate())))
}

func FixDeleteSystemAuthForApplicationRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteSystemAuthForApplication(authID: "%s") {
					%s
				}
			}`, authID, testctx.Tc.GQLFieldsProvider.ForSystemAuth()))
}

func FixRegisterApplicationFromTemplate(applicationFromTemplateInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplicationFromTemplate(in: %s) {
					%s
				}
			}`,
			applicationFromTemplateInputInGQL, testctx.Tc.GQLFieldsProvider.ForApplication()))
}

func FixSetDefaultEventingForApplication(appID string, runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setDefaultEventingForApplication(runtimeID: "%s", appID: "%s") {
					%s
				}
			}`,
			runtimeID, appID, testctx.Tc.GQLFieldsProvider.ForEventingConfiguration()))
}

func FixDeleteDefaultEventingForApplication(appID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: deleteDefaultEventingForApplication(appID: "%s") {
						%s
					}
				}`,
			appID, testctx.Tc.GQLFieldsProvider.ForEventingConfiguration()))
}

//API
func FixUpdateAPIRequest(apiID, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: updateAPIDefinition(id: "%s", in: %s) {
				%s
			}
		}
		`, apiID, APIInputGQL, testctx.Tc.GQLFieldsProvider.ForAPIDefinition()))
}

func FixDeleteAPIRequest(apiID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: deleteAPIDefinition(id: "%s") {
				id
			}
		}`, apiID))
}

func FixUpdateEventAPIRequest(eventAPIID, eventAPIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: updateEventDefinition(id: "%s", in: %s) {
				%s
			}
		}
		`, eventAPIID, eventAPIInputGQL, testctx.Tc.GQLFieldsProvider.ForEventDefinition()))
}

func FixDeleteEventAPIRequest(eventAPIID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteEventDefinition(id: "%s") {
				id
			}
		}`, eventAPIID))
}

//API Definition
func FixAPIDefinitionInputWithName(name string) graphql.APIDefinitionInput {
	return graphql.APIDefinitionInput{
		Name:      name,
		TargetURL: "https://target.url",
		Spec: &graphql.APISpecInput{
			Format: graphql.SpecFormatJSON,
			Type:   graphql.APISpecTypeOpenAPI,
			FetchRequest: &graphql.FetchRequestInput{
				URL: "https://foo.bar",
			},
		},
	}
}

func FixEventAPIDefinitionInputWithName(name string) graphql.EventDefinitionInput {
	data := graphql.CLOB("data")
	return graphql.EventDefinitionInput{Name: name,
		Spec: &graphql.EventSpecInput{
			Data:   &data,
			Type:   graphql.EventSpecTypeAsyncAPI,
			Format: graphql.SpecFormatJSON,
		}}
}

func FixDocumentInputWithName(t *testing.T, name string) graphql.DocumentInput {
	return graphql.DocumentInput{
		Title:       name,
		Description: "Detailed description of project",
		Format:      graphql.DocumentFormatMarkdown,
		DisplayName: "display-name",
		FetchRequest: &graphql.FetchRequestInput{
			URL:    "kyma-project.io",
			Mode:   ptr.FetchMode(graphql.FetchModeBundle),
			Filter: ptr.String("/docs/README.md"),
			Auth:   FixBasicAuth(t),
		},
	}
}

func FixAPIDefinitionInBundleRequest(appID, bndlID, apiID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							apiDefinition(id: "%s"){
						%s
						}					
					}
				}
			}`, appID, bndlID, apiID, testctx.Tc.GQLFieldsProvider.ForAPIDefinition()))
}

func FixEventDefinitionInBundleRequest(appID, bndlID, eventID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							eventDefinition(id: "%s"){
						%s
						}					
					}
				}
			}`, appID, bndlID, eventID, testctx.Tc.GQLFieldsProvider.ForEventDefinition()))
}

//API Spec
func FixRefetchAPISpecRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: refetchAPISpec(apiID: "%s") {
						%s
					}
				}`,
			id, testctx.Tc.GQLFieldsProvider.ForApiSpec()))
}

// External services mock
func GetAuditlogMockToken(t *testing.T, client *http.Client, baseURL string) Token {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", baseURL, auditlogTokenEndpoint), nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", "client_id", "client_secret"))))
	resp, err := client.Do(req)
	require.NoError(t, err)

	var auditlogToken Token
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &auditlogToken)
	require.NoError(t, err)

	return auditlogToken
}

func SearchForAuditlogByString(t *testing.T, client *http.Client, baseURL string, auditlogToken Token, search string) []model.ConfigurationChange {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", baseURL, auditlogSearchEndpoint), nil)
	require.NoError(t, err)

	req.URL.RawQuery = fmt.Sprintf("query=%s", search)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auditlogToken.AccessToken))
	resp, err := client.Do(req)
	require.NoError(t, err)

	var auditlogs []model.ConfigurationChange
	body, err := ioutil.ReadAll(resp.Body)

	require.NoError(t, err)
	err = json.Unmarshal(body, &auditlogs)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	return auditlogs
}

func DeleteAuditlogByID(t *testing.T, client *http.Client, baseURL string, auditlogToken Token, id string) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s%s", baseURL, fmt.Sprintf(auditlogDeleteEndpointFormat, id)), nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auditlogToken.AccessToken))
	resp, err := client.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

//Bundle
func FixBundleCreateInput(name string) graphql.BundleCreateInput {
	return graphql.BundleCreateInput{
		Name: name,
	}
}

func FixAddAPIToBundleRequest(bundleID, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addAPIDefinitionToBundle(bundleID: "%s", in: %s) {
				%s
			}
		}
		`, bundleID, APIInputGQL, testctx.Tc.GQLFieldsProvider.ForAPIDefinition()))
}

func FixBundleInstanceAuthRequestInput(ctx, inputParams *graphql.JSON) graphql.BundleInstanceAuthRequestInput {
	return graphql.BundleInstanceAuthRequestInput{
		Context:     ctx,
		InputParams: inputParams,
	}
}

func FixBundleInstanceAuthSetInputSucceeded(auth *graphql.AuthInput) graphql.BundleInstanceAuthSetInput {
	return graphql.BundleInstanceAuthSetInput{
		Auth: auth,
	}
}

func FixAddBundleRequest(appID, bundleCreateInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addBundle(applicationID: "%s", in: %s) {
				%s
			}}`, appID, bundleCreateInput, testctx.Tc.GQLFieldsProvider.ForBundle()))
}

func FixUpdateBundleRequest(bundleID, bndlUpdateInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateBundle(id: "%s", in: %s) {
				%s
			}
		}`, bundleID, bndlUpdateInput, testctx.Tc.GQLFieldsProvider.ForBundle()))
}

func FixDeleteBundleRequest(bundleID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteBundle(id: "%s") {
				%s
			}
		}`, bundleID, testctx.Tc.GQLFieldsProvider.ForBundle()))
}

func FixBundleRequest(applicationID string, bundleID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID, testctx.Tc.GQLFieldsProvider.ForApplication(graphqlizer.FieldCtx{
			"Application.bundle": fmt.Sprintf(`bundle(id: "%s") {%s}`, bundleID, testctx.Tc.GQLFieldsProvider.ForBundle()),
		})))
}

func FixAddDocumentToBundleRequest(bundleID, documentInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addDocumentToBundle(bundleID: "%s", in: %s) {
 				%s
			}				
		}`, bundleID, documentInputInGQL, testctx.Tc.GQLFieldsProvider.ForDocument()))
}

func FixAddEventAPIToBundleRequest(bndlID, eventAPIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addEventDefinitionToBundle(bundleID: "%s", in: %s) {
				%s
			}
		}
		`, bndlID, eventAPIInputGQL, testctx.Tc.GQLFieldsProvider.ForEventDefinition()))
}

func FixBundleCreateInputWithRelatedObjects(t *testing.T, name string) graphql.BundleCreateInput {
	desc := "Foo bar"
	return graphql.BundleCreateInput{
		Name:        name,
		Description: &desc,
		APIDefinitions: []*graphql.APIDefinitionInput{
			{
				Name:        "comments-v1",
				Description: ptr.String("api for adding comments"),
				TargetURL:   "http://mywordpress.com/comments",
				Group:       ptr.String("comments"),
				Version:     FixDeprecatedVersion(),
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOpenAPI,
					Format: graphql.SpecFormatYaml,
					Data:   ptr.CLOB(`{"openapi":"3.0.2"}`),
				},
			},
			{
				Name:      "reviews-v1",
				TargetURL: "http://mywordpress.com/reviews",
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOdata,
					Format: graphql.SpecFormatJSON,
					FetchRequest: &graphql.FetchRequestInput{
						URL:    "http://mywordpress.com/apis",
						Mode:   ptr.FetchMode(graphql.FetchModeBundle),
						Filter: ptr.String("odata.json"),
						Auth:   FixBasicAuth(t),
					},
				},
			},
			{
				Name:      "xml",
				TargetURL: "http://mywordpress.com/xml",
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOdata,
					Format: graphql.SpecFormatXML,
					Data:   ptr.CLOB("odata"),
				},
			},
		},
		EventDefinitions: []*graphql.EventDefinitionInput{
			{
				Name:        "comments-v1",
				Description: ptr.String("comments events"),
				Version:     FixDeprecatedVersion(),
				Group:       ptr.String("comments"),
				Spec: &graphql.EventSpecInput{
					Type:   graphql.EventSpecTypeAsyncAPI,
					Format: graphql.SpecFormatYaml,
					Data:   ptr.CLOB(`{"asyncapi":"1.2.0"}`),
				},
			},
			{
				Name:        "reviews-v1",
				Description: ptr.String("review events"),
				Spec: &graphql.EventSpecInput{
					Type:   graphql.EventSpecTypeAsyncAPI,
					Format: graphql.SpecFormatYaml,
					FetchRequest: &graphql.FetchRequestInput{
						URL:    "http://mywordpress.com/events",
						Mode:   ptr.FetchMode(graphql.FetchModeBundle),
						Filter: ptr.String("async.json"),
						Auth:   FixOauthAuth(),
					},
				},
			},
		},
		Documents: []*graphql.DocumentInput{
			{
				Title:       "Readme",
				Description: "Detailed description of project",
				Format:      graphql.DocumentFormatMarkdown,
				DisplayName: "display-name",
				FetchRequest: &graphql.FetchRequestInput{
					URL:    "kyma-project.io",
					Mode:   ptr.FetchMode(graphql.FetchModeBundle),
					Filter: ptr.String("/docs/README.md"),
					Auth:   FixBasicAuth(t),
				},
			},
			{
				Title:       "Troubleshooting",
				Description: "Troubleshooting description",
				Format:      graphql.DocumentFormatMarkdown,
				DisplayName: "display-name",
				Data:        ptr.CLOB("No problems, everything works on my machine"),
			},
		},
	}
}

func FixBundleCreateInputWithDefaultAuth(name string, authInput *graphql.AuthInput) graphql.BundleCreateInput {
	return graphql.BundleCreateInput{
		Name:                name,
		DefaultInstanceAuth: authInput,
	}
}

func FixBundleUpdateInput(name string) graphql.BundleUpdateInput {
	return graphql.BundleUpdateInput{
		Name: name,
	}
}

func FixDocumentInBundleRequest(appID, bndlID, docID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							document(id: "%s"){
						%s
						}					
					}
				}
			}`, appID, bndlID, docID, testctx.Tc.GQLFieldsProvider.ForDocument()))
}

func FixAPIDefinitionsInBundleRequest(appID, bndlID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							apiDefinitions{
						%s
						}					
					}
				}
			}`, appID, bndlID, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForAPIDefinition())))

}

func FixEventDefinitionsInBundleRequest(appID, bndlID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							eventDefinitions{
						%s
						}					
					}
				}
			}`, appID, bndlID, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForEventDefinition())))
}

func FixDocumentsInBundleRequest(appID, bndlID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							documents{
						%s
						}					
					}
				}
			}`, appID, bndlID, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForDocument())))
}

func FixSetBundleInstanceAuthRequest(authID, apiAuthInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setBundleInstanceAuth(authID: "%s", in: %s) {
				%s
			}
		}`, authID, apiAuthInput, testctx.Tc.GQLFieldsProvider.ForBundleInstanceAuth()))
}

func FixDeleteBundleInstanceAuthRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteBundleInstanceAuth(authID: "%s") {
				%s
			}
		}`, authID, testctx.Tc.GQLFieldsProvider.ForBundleInstanceAuth()))
}

func FixRequestBundleInstanceAuthCreationRequest(bundleID, bndlInstanceAuthRequestInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestBundleInstanceAuthCreation(bundleID: "%s", in: %s) {
				%s
			}
		}`, bundleID, bndlInstanceAuthRequestInput, testctx.Tc.GQLFieldsProvider.ForBundleInstanceAuth()))
}

func FixRequestBundleInstanceAuthDeletionRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestBundleInstanceAuthDeletion(authID: "%s") {
				%s
			}
		}`, authID, testctx.Tc.GQLFieldsProvider.ForBundleInstanceAuth()))
}

func FixBundleByInstanceAuthIDRequest(packageInstanceAuthID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: packageByInstanceAuth(authID: "%s") {
				%s
				}
			}`, packageInstanceAuthID, testctx.Tc.GQLFieldsProvider.ForBundle()))
}

func FixGetBundleWithInstanceAuthRequest(applicationID string, bundleID string, instanceAuthID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID,
			testctx.Tc.GQLFieldsProvider.ForApplication(
				graphqlizer.FieldCtx{"Application.bundle": fmt.Sprintf(`bundle(id: "%s") {%s}`,
					bundleID,
					testctx.Tc.GQLFieldsProvider.ForBundle(graphqlizer.FieldCtx{
						"Bundle.instanceAuth": fmt.Sprintf(`instanceAuth(id: "%s") {%s}`,
							instanceAuthID,
							testctx.Tc.GQLFieldsProvider.ForBundleInstanceAuth()),
					})),
				})))
}

func FixGetBundlesRequest(applicationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID, testctx.Tc.GQLFieldsProvider.ForApplication()))
}

func FixBundleInstanceAuthRequest(packageInstanceAuthID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: bundleInstanceAuth(id: "%s") {
					%s
				}
			}`, packageInstanceAuthID, testctx.Tc.GQLFieldsProvider.ForBundleInstanceAuth()))
}

func FixBundleInstanceAuthContextAndInputParams(t *testing.T) (*graphql.JSON, *graphql.JSON) {
	authCtxPayload := map[string]interface{}{
		"ContextData": "ContextValue",
	}
	var authCtxData interface{} = authCtxPayload

	inputParamsPayload := map[string]interface{}{
		"InKey": "InValue",
	}
	var inputParamsData interface{} = inputParamsPayload

	authCtx := pkg.MarshalJSON(t, authCtxData)
	inputParams := pkg.MarshalJSON(t, inputParamsData)

	return authCtx, inputParams
}

// Integration system
func FixRegisterIntegrationSystemRequest(integrationSystemInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerIntegrationSystem(in: %s) {
					%s
				}
			}`,
			integrationSystemInGQL, testctx.Tc.GQLFieldsProvider.ForIntegrationSystem()))
}

func FixGetIntegrationSystemRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: integrationSystem(id: "%s") {
					%s
				}
			}`,
			id, testctx.Tc.GQLFieldsProvider.ForIntegrationSystem()))
}

func FixRequestClientCredentialsForIntegrationSystem(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestClientCredentialsForIntegrationSystem(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForSystemAuth()))
}

func FixUnregisterIntegrationSystem(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: unregisterIntegrationSystem(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForIntegrationSystem()))
}

func FixDeleteSystemAuthForIntegrationSystemRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteSystemAuthForIntegrationSystem(authID: "%s") {
					%s
				}
			}`, authID, testctx.Tc.GQLFieldsProvider.ForSystemAuth()))
}

func FixUpdateIntegrationSystemRequest(id, integrationSystemInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateIntegrationSystem(id: "%s", in: %s) {
					%s
				}
			}`, id, integrationSystemInGQL, testctx.Tc.GQLFieldsProvider.ForIntegrationSystem()))
}

func FixGetIntegrationSystemsRequestWithPagination(first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: integrationSystems(first: %d, after: "%s") {
						%s
					}
				}`,
			first, after, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForIntegrationSystem())))
}

//Runtime
func FixRuntimeInput(placeholder string) graphql.RuntimeInput {
	return graphql.RuntimeInput{
		Name:        placeholder,
		Description: ptr.String(fmt.Sprintf("%s-description", placeholder)),
		Labels:      &graphql.Labels{"placeholder": []interface{}{"placeholder"}},
	}
}

func FixRegisterRuntimeRequest(runtimeInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerRuntime(in: %s) {
					%s
				}
			}`,
			runtimeInGQL, testctx.Tc.GQLFieldsProvider.ForRuntime()))
}

func FixRequestClientCredentialsForRuntime(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestClientCredentialsForRuntime(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForSystemAuth()))
}

func FixUnregisterRuntimeRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{unregisterRuntime(id: "%s") {
				%s
			}
		}`, id, testctx.Tc.GQLFieldsProvider.ForRuntime()))
}

func FixGetRuntimeRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
					%s
				}}`, id, testctx.Tc.GQLFieldsProvider.ForRuntime()))
}

func FixRequestOneTimeTokenForRuntime(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestOneTimeTokenForRuntime(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForOneTimeTokenForRuntime()))
}

func FixUpdateRuntimeRequest(id, updateInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateRuntime(id: "%s", in: %s) {
					%s
				}
			}`,
			id, updateInputInGQL, testctx.Tc.GQLFieldsProvider.ForRuntime()))
}

func FixRuntimeRequestWithPaginationRequest(after int, cursor string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtimes(first:%d, after:"%s") {
					%s
				}
			}`, after, cursor, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForRuntime())))
}

func FixGetRuntimesRequestWithPagination() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: runtimes {
						%s
					}
				}`,
			testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForRuntime())))
}

func FixRuntimesFilteredPageableRequest(labelFilterInGQL string, first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: runtimes(filter: %s, first: %d, after: "%s") {
						%s
					}
				}`,
			labelFilterInGQL, first, after, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForRuntime())))
}

func FixDeleteSystemAuthForRuntimeRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteSystemAuthForRuntime(authID: "%s") {
					%s
				}
			}`, authID, testctx.Tc.GQLFieldsProvider.ForSystemAuth()))
}

// Viewer
func FixGetViewerRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: viewer {
					%s
				}
			}`,
			testctx.Tc.GQLFieldsProvider.ForViewer()))
}

//Label
func FixCreateLabelDefinitionRequest(labelDefinitionInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: createLabelDefinition(in: %s) {
						%s
					}
				}`,
			labelDefinitionInputGQL, testctx.Tc.GQLFieldsProvider.ForLabelDefinition()))
}

func FixUpdateLabelDefinitionRequest(ldInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: updateLabelDefinition(in: %s) {
						%s
					}
				}`, ldInputGQL, testctx.Tc.GQLFieldsProvider.ForLabelDefinition()))
}

func FixSetApplicationLabelRequest(appID, labelKey string, labelValue interface{}) *gcli.Request {
	jsonValue, err := json.Marshal(labelValue)
	if err != nil {
		panic(errors.New("label value can not be marshalled"))
	}
	value := removeDoubleQuotesFromJSONKeys(string(jsonValue))

	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setApplicationLabel(applicationID: "%s", key: "%s", value: %s) {
					%s
				}
			}`,
			appID, labelKey, value, testctx.Tc.GQLFieldsProvider.ForLabel()))
}

func FixSetRuntimeLabelRequest(runtimeID, labelKey string, labelValue interface{}) *gcli.Request {
	jsonValue, err := json.Marshal(labelValue)
	if err != nil {
		panic(errors.New("label value can not be marshalled"))
	}
	value := removeDoubleQuotesFromJSONKeys(string(jsonValue))

	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: setRuntimeLabel(runtimeID: "%s", key: "%s", value: %s) {
						%s
					}
				}`, runtimeID, labelKey, value, testctx.Tc.GQLFieldsProvider.ForLabel()))
}

func FixLabelDefinitionRequest(labelKey string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: labelDefinition(key: "%s") {
						%s
					}
				}`,
			labelKey, testctx.Tc.GQLFieldsProvider.ForLabelDefinition()))
}

func FixLabelDefinitionsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result:	labelDefinitions() {
					key
					schema
				}
			}`))
}

func FixDeleteLabelDefinitionRequest(labelDefinitionKey string, deleteRelatedLabels bool) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteLabelDefinition(key: "%s", deleteRelatedLabels: %t) {
					%s
				}
			}`, labelDefinitionKey, deleteRelatedLabels, testctx.Tc.GQLFieldsProvider.ForLabelDefinition()))
}

func FixDeleteRuntimeLabelRequest(runtimeID, labelKey string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteRuntimeLabel(runtimeID: "%s", key: "%s") {
					%s
				}
			}`, runtimeID, labelKey, testctx.Tc.GQLFieldsProvider.ForLabel()))
}

func FixDeleteApplicationLabelRequest(applicationID, labelKey string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteApplicationLabel(applicationID: "%s", key: "%s") {
					%s
				}
			}`, applicationID, labelKey, testctx.Tc.GQLFieldsProvider.ForLabel()))
}

//Document
func FixDeleteDocumentRequest(docID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteDocument(id: "%s") {
					id
				}
			}`, docID))
}

//Webhook
func FixAddWebhookRequest(applicationID, webhookInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addWebhook(applicationID: "%s", in: %s) {
					%s
				}
			}`,
			applicationID, webhookInGQL, testctx.Tc.GQLFieldsProvider.ForWebhooks()))
}

func FixDeleteWebhookRequest(webhookID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteWebhook(webhookID: "%s") {
				%s
			}
		}`, webhookID, testctx.Tc.GQLFieldsProvider.ForWebhooks()))
}

func FixUpdateWebhookRequest(webhookID, webhookInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateWebhook(webhookID: "%s", in: %s) {
					%s
				}
			}`,
			webhookID, webhookInGQL, testctx.Tc.GQLFieldsProvider.ForWebhooks()))
}

func FixTenantsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: tenants {
						%s
					}
				}`, testctx.Tc.GQLFieldsProvider.ForTenant()))
}

//Scenario Assignment
func FixAutomaticScenarioAssigmentInput(automaticScenario, selectorKey, selectorValue string) graphql.AutomaticScenarioAssignmentSetInput {
	return graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: automaticScenario,
		Selector: &graphql.LabelSelectorInput{
			Key:   selectorKey,
			Value: selectorValue,
		},
	}
}

func FixCreateAutomaticScenarioAssignmentRequest(automaticScenarioAssignmentInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: createAutomaticScenarioAssignment(in: %s) {
						%s
					}
				}`,
			automaticScenarioAssignmentInput, testctx.Tc.GQLFieldsProvider.ForAutomaticScenarioAssignment()))
}

func FixDeleteAutomaticScenarioAssignmentForScenarioRequest(scenario string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
            result: deleteAutomaticScenarioAssignmentForScenario(scenarioName: "%s") {
                  %s
               }
            }`,
			scenario, testctx.Tc.GQLFieldsProvider.ForAutomaticScenarioAssignment()))
}

func FixDeleteAutomaticScenarioAssignmentsForSelectorRequest(labelSelectorInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
            result: deleteAutomaticScenarioAssignmentsForSelector(selector: %s) {
                  %s
               }
            }`,
			labelSelectorInput, testctx.Tc.GQLFieldsProvider.ForAutomaticScenarioAssignment()))
}

func FixAutomaticScenarioAssignmentsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: automaticScenarioAssignments {
						%s
					}
				}`,
			testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForAutomaticScenarioAssignment())))
}

func FixAutomaticScenarioAssignmentsForSelectorRequest(labelSelectorInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: automaticScenarioAssignmentsForSelector(selector: %s) {
						%s
					}
				}`,
			labelSelectorInput, testctx.Tc.GQLFieldsProvider.ForAutomaticScenarioAssignment()))
}

func FixAutomaticScenarioAssignmentForScenarioRequest(scenarioName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: automaticScenarioAssignmentForScenario(scenarioName: "%s") {
						%s
					}
				}`,
			scenarioName, testctx.Tc.GQLFieldsProvider.ForAutomaticScenarioAssignment()))
}

func removeDoubleQuotesFromJSONKeys(in string) string {
	var validRegex = regexp.MustCompile(`"(\w+|\$\w+)"\s*:`)
	return validRegex.ReplaceAllString(in, `$1:`)
}

func FixEventAPIDefinitionInput() graphql.EventDefinitionInput {
	data := graphql.CLOB("data")
	return graphql.EventDefinitionInput{Name: "name",
		Spec: &graphql.EventSpecInput{
			Data:   &data,
			Type:   graphql.EventSpecTypeAsyncAPI,
			Format: graphql.SpecFormatJSON,
		},
	}
}

func FixAPIDefinitionInput() graphql.APIDefinitionInput {
	return graphql.APIDefinitionInput{
		Name:      "new-api-name",
		TargetURL: "https://target.url",
		Spec: &graphql.APISpecInput{
			Format: graphql.SpecFormatJSON,
			Type:   graphql.APISpecTypeOpenAPI,
			FetchRequest: &graphql.FetchRequestInput{
				URL: "https://foo.bar",
			},
		},
	}
}

func FixDocumentInput(t *testing.T) graphql.DocumentInput {
	return graphql.DocumentInput{
		Title:       "Readme",
		Description: "Detailed description of project",
		Format:      graphql.DocumentFormatMarkdown,
		DisplayName: "display-name",
		FetchRequest: &graphql.FetchRequestInput{
			URL:    "kyma-project.io",
			Mode:   ptr.FetchMode(graphql.FetchModeBundle),
			Filter: ptr.String("/docs/README.md"),
			Auth:   FixBasicAuth(t),
		},
	}
}

// TODO: Delete after bundles are adopted
func FixRegisterApplicationWithPackagesRequest(name string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			  result: registerApplication(
				in: {
				  name: "%s"
				  providerName: "compass"
				  labels: { scenarios: ["DEFAULT"] }
				  packages: [
					{
					  name: "foo"
					  description: "Foo bar"
					  apiDefinitions: [
						{
						  name: "comments-v1"
						  description: "api for adding comments"
						  targetURL: "http://mywordpress.com/comments"
						  group: "comments"
						  spec: {
							data: "{\"openapi\":\"3.0.2\"}"
							type: OPEN_API
							format: YAML
						  }
						  version: {
							value: "v1"
							deprecated: true
							deprecatedSince: "v5"
							forRemoval: true
						  }
						}
						{
						  name: "reviews-v1"
						  targetURL: "http://mywordpress.com/reviews"
						  spec: {
							type: ODATA
							format: JSON
							fetchRequest: {
							  url: "http://mywordpress.com/apis"
							  auth: {
								credential: {
								  basic: { username: "admin", password: "secret" }
								}
								additionalHeadersSerialized: "{\"header-A\":[\"ha1\",\"ha2\"],\"header-B\":[\"hb1\",\"hb2\"]}"
								additionalQueryParamsSerialized: "{\"qA\":[\"qa1\",\"qa2\"],\"qB\":[\"qb1\",\"qb2\"]}"
							  }
							  mode: PACKAGE
							  filter: "odata.json"
							}
						  }
						}
						{
						  name: "xml"
						  targetURL: "http://mywordpress.com/xml"
						  spec: { data: "odata", type: ODATA, format: XML }
						}
					  ]
					  eventDefinitions: [
						{
						  name: "comments-v1"
						  description: "comments events"
						  spec: {
							data: "{\"asyncapi\":\"1.2.0\"}"
							type: ASYNC_API
							format: YAML
						  }
						  group: "comments"
						  version: {
							value: "v1"
							deprecated: true
							deprecatedSince: "v5"
							forRemoval: true
						  }
						}
						{
						  name: "reviews-v1"
						  description: "review events"
						  spec: {
							type: ASYNC_API
							fetchRequest: {
							  url: "http://mywordpress.com/events"
							  auth: {
								credential: {
								  oauth: {
									clientId: "clientid"
									clientSecret: "grazynasecret"
									url: "url.net"
								  }
								}
							  }
							  mode: PACKAGE
							  filter: "async.json"
							}
							format: YAML
						  }
						}
					  ]
					  documents: [
						{
						  title: "Readme"
						  displayName: "display-name"
						  description: "Detailed description of project"
						  format: MARKDOWN
						  fetchRequest: {
							url: "kyma-project.io"
							auth: {
							  credential: {
								basic: { username: "admin", password: "secret" }
							  }
							  additionalHeadersSerialized: "{\"header-A\":[\"ha1\",\"ha2\"],\"header-B\":[\"hb1\",\"hb2\"]}"
							  additionalQueryParamsSerialized: "{\"qA\":[\"qa1\",\"qa2\"],\"qB\":[\"qb1\",\"qb2\"]}"
							}
							mode: PACKAGE
							filter: "/docs/README.md"
						  }
						}
						{
						  title: "Troubleshooting"
						  displayName: "display-name"
						  description: "Troubleshooting description"
						  format: MARKDOWN
						  data: "No problems, everything works on my machine"
						}
					  ]
					}
					{
					  name: "bar"
					  description: "Foo bar"
					  apiDefinitions: [
						{
						  name: "comments-v1"
						  description: "api for adding comments"
						  targetURL: "http://mywordpress.com/comments"
						  group: "comments"
						  spec: {
							data: "{\"openapi\":\"3.0.2\"}"
							type: OPEN_API
							format: YAML
						  }
						  version: {
							value: "v1"
							deprecated: true
							deprecatedSince: "v5"
							forRemoval: true
						  }
						}
						{
						  name: "reviews-v1"
						  targetURL: "http://mywordpress.com/reviews"
						  spec: {
							type: ODATA
							format: JSON
							fetchRequest: {
							  url: "http://mywordpress.com/apis"
							  auth: {
								credential: {
								  basic: { username: "admin", password: "secret" }
								}
								additionalHeadersSerialized: "{\"header-A\":[\"ha1\",\"ha2\"],\"header-B\":[\"hb1\",\"hb2\"]}"
								additionalQueryParamsSerialized: "{\"qA\":[\"qa1\",\"qa2\"],\"qB\":[\"qb1\",\"qb2\"]}"
							  }
							  mode: PACKAGE
							  filter: "odata.json"
							}
						  }
						}
						{
						  name: "xml"
						  targetURL: "http://mywordpress.com/xml"
						  spec: { data: "odata", type: ODATA, format: XML }
						}
					  ]
					  eventDefinitions: [
						{
						  name: "comments-v1"
						  description: "comments events"
						  spec: {
							data: "{\"asyncapi\":\"1.2.0\"}"
							type: ASYNC_API
							format: YAML
						  }
						  group: "comments"
						  version: {
							value: "v1"
							deprecated: true
							deprecatedSince: "v5"
							forRemoval: true
						  }
						}
						{
						  name: "reviews-v1"
						  description: "review events"
						  spec: {
							type: ASYNC_API
							fetchRequest: {
							  url: "http://mywordpress.com/events"
							  auth: {
								credential: {
								  oauth: {
									clientId: "clientid"
									clientSecret: "grazynasecret"
									url: "url.net"
								  }
								}
							  }
							  mode: PACKAGE
							  filter: "async.json"
							}
							format: YAML
						  }
						}
					  ]
					  documents: [
						{
						  title: "Readme"
						  displayName: "display-name"
						  description: "Detailed description of project"
						  format: MARKDOWN
						  fetchRequest: {
							url: "kyma-project.io"
							auth: {
							  credential: {
								basic: { username: "admin", password: "secret" }
							  }
							  additionalHeadersSerialized: "{\"header-A\":[\"ha1\",\"ha2\"],\"header-B\":[\"hb1\",\"hb2\"]}"
							  additionalQueryParamsSerialized: "{\"qA\":[\"qa1\",\"qa2\"],\"qB\":[\"qb1\",\"qb2\"]}"
							}
							mode: PACKAGE
							filter: "/docs/README.md"
						  }
						}
						{
						  title: "Troubleshooting"
						  displayName: "display-name"
						  description: "Troubleshooting description"
						  format: MARKDOWN
						  data: "No problems, everything works on my machine"
						}
					  ]
					}
				  ]
				}
			  ) {
				id
				name
				providerName
				description
				integrationSystemID
				labels
				status {
				  condition
				  timestamp
				}
				webhooks {
				  id
				  applicationID
				  type
				  url
				  auth {
					credential {
					  ... on BasicCredentialData {
						username
						password
					  }
					  ... on OAuthCredentialData {
						clientId
						clientSecret
						url
					  }
					}
					additionalHeaders
					additionalQueryParams
					requestAuth {
					  csrf {
						tokenEndpointURL
						credential {
						  ... on BasicCredentialData {
							username
							password
						  }
						  ... on OAuthCredentialData {
							clientId
							clientSecret
							url
						  }
						}
						additionalHeaders
						additionalQueryParams
					  }
					}
				  }
				}
				healthCheckURL
				packages {
				  data {
					id
					name
					description
					instanceAuthRequestInputSchema
					instanceAuths {
					  id
					  context
					  inputParams
					  auth {
						credential {
						  ... on BasicCredentialData {
							username
							password
						  }
						  ... on OAuthCredentialData {
							clientId
							clientSecret
							url
						  }
						}
						additionalHeaders
						additionalQueryParams
						requestAuth {
						  csrf {
							tokenEndpointURL
							credential {
							  ... on BasicCredentialData {
								username
								password
							  }
							  ... on OAuthCredentialData {
								clientId
								clientSecret
								url
							  }
							}
							additionalHeaders
							additionalQueryParams
						  }
						}
					  }
					  status {
						condition
						timestamp
						message
						reason
					  }
					}
					defaultInstanceAuth {
					  credential {
						... on BasicCredentialData {
						  username
						  password
						}
						... on OAuthCredentialData {
						  clientId
						  clientSecret
						  url
						}
					  }
					  additionalHeaders
					  additionalQueryParams
					  requestAuth {
						csrf {
						  tokenEndpointURL
						  credential {
							... on BasicCredentialData {
							  username
							  password
							}
							... on OAuthCredentialData {
							  clientId
							  clientSecret
							  url
							}
						  }
						  additionalHeaders
						  additionalQueryParams
						}
					  }
					}
					apiDefinitions {
					  data {
						id
						name
						description
						spec {
						  data
						  format
						  type
						  fetchRequest {
							url
							auth {
							  credential {
								... on BasicCredentialData {
								  username
								  password
								}
								... on OAuthCredentialData {
								  clientId
								  clientSecret
								  url
								}
							  }
							  additionalHeaders
							  additionalQueryParams
							  requestAuth {
								csrf {
								  tokenEndpointURL
								  credential {
									... on BasicCredentialData {
									  username
									  password
									}
									... on OAuthCredentialData {
									  clientId
									  clientSecret
									  url
									}
								  }
								  additionalHeaders
								  additionalQueryParams
								}
							  }
							}
							mode
							filter
							status {
							  condition
							  message
							  timestamp
							}
						  }
						}
						targetURL
						group
						version {
						  value
						  deprecated
						  deprecatedSince
						  forRemoval
						}
					  }
					  pageInfo {
						startCursor
						endCursor
						hasNextPage
					  }
					  totalCount
					}
					eventDefinitions {
					  data {
						id
						name
						description
						group
						spec {
						  data
						  type
						  format
						  fetchRequest {
							url
							auth {
							  credential {
								... on BasicCredentialData {
								  username
								  password
								}
								... on OAuthCredentialData {
								  clientId
								  clientSecret
								  url
								}
							  }
							  additionalHeaders
							  additionalQueryParams
							  requestAuth {
								csrf {
								  tokenEndpointURL
								  credential {
									... on BasicCredentialData {
									  username
									  password
									}
									... on OAuthCredentialData {
									  clientId
									  clientSecret
									  url
									}
								  }
								  additionalHeaders
								  additionalQueryParams
								}
							  }
							}
							mode
							filter
							status {
							  condition
							  message
							  timestamp
							}
						  }
						}
						version {
						  value
						  deprecated
						  deprecatedSince
						  forRemoval
						}
					  }
					  pageInfo {
						startCursor
						endCursor
						hasNextPage
					  }
					  totalCount
					}
					documents {
					  data {
						id
						title
						displayName
						description
						format
						kind
						data
						fetchRequest {
						  url
						  auth {
							credential {
							  ... on BasicCredentialData {
								username
								password
							  }
							  ... on OAuthCredentialData {
								clientId
								clientSecret
								url
							  }
							}
							additionalHeaders
							additionalQueryParams
							requestAuth {
							  csrf {
								tokenEndpointURL
								credential {
								  ... on BasicCredentialData {
									username
									password
								  }
								  ... on OAuthCredentialData {
									clientId
									clientSecret
									url
								  }
								}
								additionalHeaders
								additionalQueryParams
							  }
							}
						  }
						  mode
						  filter
						  status {
							condition
							message
							timestamp
						  }
						}
					  }
					  pageInfo {
						startCursor
						endCursor
						hasNextPage
					  }
					  totalCount
					}
				  }
				  pageInfo {
					startCursor
					endCursor
					hasNextPage
				  }
				  totalCount
				}
				auths {
				  id
				  auth {
					credential {
					  ... on BasicCredentialData {
						username
						password
					  }
					  ... on OAuthCredentialData {
						clientId
						clientSecret
						url
					  }
					}
					additionalHeaders
					additionalQueryParams
					requestAuth {
					  csrf {
						tokenEndpointURL
						credential {
						  ... on BasicCredentialData {
							username
							password
						  }
						  ... on OAuthCredentialData {
							clientId
							clientSecret
							url
						  }
						}
						additionalHeaders
						additionalQueryParams
					  }
					}
				  }
				}
				eventingConfiguration {
				  defaultURL
				}
			  }
			}
		`, name))
}

// TODO: Delete after bundles are adopted
func FixGetApplicationWithPackageRequest(appID, packageID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			  result: application(id: "%s") {
				id
				name
				providerName
				description
				integrationSystemID
				labels
				status {
				  condition
				  timestamp
				}
				webhooks {
				  id
				  applicationID
				  type
				  url
				  auth {
					credential {
					  ... on BasicCredentialData {
						username
						password
					  }
					  ... on OAuthCredentialData {
						clientId
						clientSecret
						url
					  }
					}
					additionalHeaders
					additionalQueryParams
					requestAuth {
					  csrf {
						tokenEndpointURL
						credential {
						  ... on BasicCredentialData {
							username
							password
						  }
						  ... on OAuthCredentialData {
							clientId
							clientSecret
							url
						  }
						}
						additionalHeaders
						additionalQueryParams
					  }
					}
				  }
				}
				healthCheckURL
				package(id: "%s") {
					id
					name
					description
					instanceAuthRequestInputSchema
					instanceAuths {
					  id
					  context
					  inputParams
					  auth {
						credential {
						  ... on BasicCredentialData {
							username
							password
						  }
						  ... on OAuthCredentialData {
							clientId
							clientSecret
							url
						  }
						}
						additionalHeaders
						additionalQueryParams
						requestAuth {
						  csrf {
							tokenEndpointURL
							credential {
							  ... on BasicCredentialData {
								username
								password
							  }
							  ... on OAuthCredentialData {
								clientId
								clientSecret
								url
							  }
							}
							additionalHeaders
							additionalQueryParams
						  }
						}
					  }
					  status {
						condition
						timestamp
						message
						reason
					  }
					}
					defaultInstanceAuth {
					  credential {
						... on BasicCredentialData {
						  username
						  password
						}
						... on OAuthCredentialData {
						  clientId
						  clientSecret
						  url
						}
					  }
					  additionalHeaders
					  additionalQueryParams
					  requestAuth {
						csrf {
						  tokenEndpointURL
						  credential {
							... on BasicCredentialData {
							  username
							  password
							}
							... on OAuthCredentialData {
							  clientId
							  clientSecret
							  url
							}
						  }
						  additionalHeaders
						  additionalQueryParams
						}
					  }
					}
					apiDefinitions {
					  data {
						id
						name
						description
						spec {
						  data
						  format
						  type
						  fetchRequest {
							url
							auth {
							  credential {
								... on BasicCredentialData {
								  username
								  password
								}
								... on OAuthCredentialData {
								  clientId
								  clientSecret
								  url
								}
							  }
							  additionalHeaders
							  additionalQueryParams
							  requestAuth {
								csrf {
								  tokenEndpointURL
								  credential {
									... on BasicCredentialData {
									  username
									  password
									}
									... on OAuthCredentialData {
									  clientId
									  clientSecret
									  url
									}
								  }
								  additionalHeaders
								  additionalQueryParams
								}
							  }
							}
							mode
							filter
							status {
							  condition
							  message
							  timestamp
							}
						  }
						}
						targetURL
						group
						version {
						  value
						  deprecated
						  deprecatedSince
						  forRemoval
						}
					  }
					  pageInfo {
						startCursor
						endCursor
						hasNextPage
					  }
					  totalCount
					}
					eventDefinitions {
					  data {
						id
						name
						description
						group
						spec {
						  data
						  type
						  format
						  fetchRequest {
							url
							auth {
							  credential {
								... on BasicCredentialData {
								  username
								  password
								}
								... on OAuthCredentialData {
								  clientId
								  clientSecret
								  url
								}
							  }
							  additionalHeaders
							  additionalQueryParams
							  requestAuth {
								csrf {
								  tokenEndpointURL
								  credential {
									... on BasicCredentialData {
									  username
									  password
									}
									... on OAuthCredentialData {
									  clientId
									  clientSecret
									  url
									}
								  }
								  additionalHeaders
								  additionalQueryParams
								}
							  }
							}
							mode
							filter
							status {
							  condition
							  message
							  timestamp
							}
						  }
						}
						version {
						  value
						  deprecated
						  deprecatedSince
						  forRemoval
						}
					  }
					  pageInfo {
						startCursor
						endCursor
						hasNextPage
					  }
					  totalCount
					}
					documents {
					  data {
						id
						title
						displayName
						description
						format
						kind
						data
						fetchRequest {
						  url
						  auth {
							credential {
							  ... on BasicCredentialData {
								username
								password
							  }
							  ... on OAuthCredentialData {
								clientId
								clientSecret
								url
							  }
							}
							additionalHeaders
							additionalQueryParams
							requestAuth {
							  csrf {
								tokenEndpointURL
								credential {
								  ... on BasicCredentialData {
									username
									password
								  }
								  ... on OAuthCredentialData {
									clientId
									clientSecret
									url
								  }
								}
								additionalHeaders
								additionalQueryParams
							  }
							}
						  }
						  mode
						  filter
						  status {
							condition
							message
							timestamp
						  }
						}
					  }
					  pageInfo {
						startCursor
						endCursor
						hasNextPage
					  }
					  totalCount
					}
				}
				auths {
				  id
				  auth {
					credential {
					  ... on BasicCredentialData {
						username
						password
					  }
					  ... on OAuthCredentialData {
						clientId
						clientSecret
						url
					  }
					}
					additionalHeaders
					additionalQueryParams
					requestAuth {
					  csrf {
						tokenEndpointURL
						credential {
						  ... on BasicCredentialData {
							username
							password
						  }
						  ... on OAuthCredentialData {
							clientId
							clientSecret
							url
						  }
						}
						additionalHeaders
						additionalQueryParams
					  }
					}
				  }
				}
				eventingConfiguration {
				  defaultURL
				}
			  }
			}`, appID, packageID))
}

// TODO: Delete after bundles are adopted
func FixApplicationsForRuntimeWithPackagesRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				  result: applicationsForRuntime(runtimeID: "%s") {
					data {
					  id
					  name
					  providerName
					  description
					  integrationSystemID
					  labels
					  status {
						condition
						timestamp
					  }
					  webhooks {
						id
						applicationID
						type
						url
						auth {
						  credential {
							... on BasicCredentialData {
							  username
							  password
							}
							... on OAuthCredentialData {
							  clientId
							  clientSecret
							  url
							}
						  }
						  additionalHeaders
						  additionalQueryParams
						  requestAuth {
							csrf {
							  tokenEndpointURL
							  credential {
								... on BasicCredentialData {
								  username
								  password
								}
								... on OAuthCredentialData {
								  clientId
								  clientSecret
								  url
								}
							  }
							  additionalHeaders
							  additionalQueryParams
							}
						  }
						}
					  }
					  healthCheckURL
					  packages {
						data {
						  id
						  name
						  description
						  instanceAuthRequestInputSchema
						  instanceAuths {
							id
							context
							inputParams
							auth {
							  credential {
								... on BasicCredentialData {
								  username
								  password
								}
								... on OAuthCredentialData {
								  clientId
								  clientSecret
								  url
								}
							  }
							  additionalHeaders
							  additionalQueryParams
							  requestAuth {
								csrf {
								  tokenEndpointURL
								  credential {
									... on BasicCredentialData {
									  username
									  password
									}
									... on OAuthCredentialData {
									  clientId
									  clientSecret
									  url
									}
								  }
								  additionalHeaders
								  additionalQueryParams
								}
							  }
							}
							status {
							  condition
							  timestamp
							  message
							  reason
							}
						  }
						  defaultInstanceAuth {
							credential {
							  ... on BasicCredentialData {
								username
								password
							  }
							  ... on OAuthCredentialData {
								clientId
								clientSecret
								url
							  }
							}
							additionalHeaders
							additionalQueryParams
							requestAuth {
							  csrf {
								tokenEndpointURL
								credential {
								  ... on BasicCredentialData {
									username
									password
								  }
								  ... on OAuthCredentialData {
									clientId
									clientSecret
									url
								  }
								}
								additionalHeaders
								additionalQueryParams
							  }
							}
						  }
						  apiDefinitions {
							data {
							  id
							  name
							  description
							  spec {
								data
								format
								type
								fetchRequest {
								  url
								  auth {
									credential {
									  ... on BasicCredentialData {
										username
										password
									  }
									  ... on OAuthCredentialData {
										clientId
										clientSecret
										url
									  }
									}
									additionalHeaders
									additionalQueryParams
									requestAuth {
									  csrf {
										tokenEndpointURL
										credential {
										  ... on BasicCredentialData {
											username
											password
										  }
										  ... on OAuthCredentialData {
											clientId
											clientSecret
											url
										  }
										}
										additionalHeaders
										additionalQueryParams
									  }
									}
								  }
								  mode
								  filter
								  status {
									condition
									message
									timestamp
								  }
								}
							  }
							  targetURL
							  group
							  version {
								value
								deprecated
								deprecatedSince
								forRemoval
							  }
							}
							pageInfo {
							  startCursor
							  endCursor
							  hasNextPage
							}
							totalCount
						  }
						  eventDefinitions {
							data {
							  id
							  name
							  description
							  group
							  spec {
								data
								type
								format
								fetchRequest {
								  url
								  auth {
									credential {
									  ... on BasicCredentialData {
										username
										password
									  }
									  ... on OAuthCredentialData {
										clientId
										clientSecret
										url
									  }
									}
									additionalHeaders
									additionalQueryParams
									requestAuth {
									  csrf {
										tokenEndpointURL
										credential {
										  ... on BasicCredentialData {
											username
											password
										  }
										  ... on OAuthCredentialData {
											clientId
											clientSecret
											url
										  }
										}
										additionalHeaders
										additionalQueryParams
									  }
									}
								  }
								  mode
								  filter
								  status {
									condition
									message
									timestamp
								  }
								}
							  }
							  version {
								value
								deprecated
								deprecatedSince
								forRemoval
							  }
							}
							pageInfo {
							  startCursor
							  endCursor
							  hasNextPage
							}
							totalCount
						  }
						  documents {
							data {
							  id
							  title
							  displayName
							  description
							  format
							  kind
							  data
							  fetchRequest {
								url
								auth {
								  credential {
									... on BasicCredentialData {
									  username
									  password
									}
									... on OAuthCredentialData {
									  clientId
									  clientSecret
									  url
									}
								  }
								  additionalHeaders
								  additionalQueryParams
								  requestAuth {
									csrf {
									  tokenEndpointURL
									  credential {
										... on BasicCredentialData {
										  username
										  password
										}
										... on OAuthCredentialData {
										  clientId
										  clientSecret
										  url
										}
									  }
									  additionalHeaders
									  additionalQueryParams
									}
								  }
								}
								mode
								filter
								status {
								  condition
								  message
								  timestamp
								}
							  }
							}
							pageInfo {
							  startCursor
							  endCursor
							  hasNextPage
							}
							totalCount
						  }
						}
						pageInfo {
						  startCursor
						  endCursor
						  hasNextPage
						}
						totalCount
					  }
					  auths {
						id
						auth {
						  credential {
							... on BasicCredentialData {
							  username
							  password
							}
							... on OAuthCredentialData {
							  clientId
							  clientSecret
							  url
							}
						  }
						  additionalHeaders
						  additionalQueryParams
						  requestAuth {
							csrf {
							  tokenEndpointURL
							  credential {
								... on BasicCredentialData {
								  username
								  password
								}
								... on OAuthCredentialData {
								  clientId
								  clientSecret
								  url
								}
							  }
							  additionalHeaders
							  additionalQueryParams
							}
						  }
						}
					  }
					  eventingConfiguration {
						defaultURL
					  }
					}
					pageInfo {
					  startCursor
					  endCursor
					  hasNextPage
					}
					totalCount
				  }
				}`, runtimeID))
}
