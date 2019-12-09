package director

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/gql"

	gcli "github.com/machinebox/graphql"
)

// CREATE
func fixCreateApplicationRequest(applicationInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplication(in: %s) {
					%s
				}
			}`,
			applicationInGQL, tc.gqlFieldsProvider.ForApplication()))
}

func fixCreateApplicationTemplateRequest(applicationTemplateInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplicationTemplate(in: %s) {
					%s
				}
			}`,
			applicationTemplateInGQL, tc.gqlFieldsProvider.ForApplicationTemplate()))
}

func fixCreateRuntimeRequest(runtimeInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createRuntime(in: %s) {
					%s
				}
			}`,
			runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
}

func fixCreateLabelDefinitionRequest(labelDefinitionInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: createLabelDefinition(in: %s) {
						%s
					}
				}`,
			labelDefinitionInputGQL, tc.gqlFieldsProvider.ForLabelDefinition()))
}

func fixCreateIntegrationSystemRequest(integrationSystemInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createIntegrationSystem(in: %s) {
					%s
				}
			}`,
			integrationSystemInGQL, tc.gqlFieldsProvider.ForIntegrationSystem()))
}

func fixAddDocumentRequest(appID, documentInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addDocument(applicationID: "%s", in: %s) {
 				%s
			}				
		}`, appID, documentInputInGQL, tc.gqlFieldsProvider.ForDocument()))
}

func fixAddWebhookRequest(applicationID, webhookInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addWebhook(applicationID: "%s", in: %s) {
					%s
				}
			}`,
			applicationID, webhookInGQL, tc.gqlFieldsProvider.ForWebhooks()))
}

func fixAddAPIRequest(appID, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addAPI(applicationID: "%s", in: %s) {
				%s
			}
		}
		`, appID, APIInputGQL, tc.gqlFieldsProvider.ForAPIDefinition()))
}

func fixUpdateAPIRequest(appID, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: updateAPI(id: "%s", in: %s) {
				%s
			}
		}
		`, appID, APIInputGQL, tc.gqlFieldsProvider.ForAPIDefinition()))
}

func fixAddEventAPIRequest(appID, eventAPIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addEventAPI(applicationID: "%s", in: %s) {
				%s
			}
		}
		`, appID, eventAPIInputGQL, tc.gqlFieldsProvider.ForEventAPI()))
}

func fixUpdateEventAPIRequest(appID, eventAPIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: updateEventAPI(id: "%s", in: %s) {
				%s
			}
		}
		`, appID, eventAPIInputGQL, tc.gqlFieldsProvider.ForEventAPI()))
}

//UPDATE
func fixUpdateRuntimeRequest(id, updateInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateRuntime(id: "%s", in: %s) {
					%s
				}
			}`,
			id, updateInputInGQL, tc.gqlFieldsProvider.ForRuntime()))
}

func fixUpdateWebhookRequest(webhookID, webhookInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateWebhook(webhookID: "%s", in: %s) {
					%s
				}
			}`,
			webhookID, webhookInGQL, tc.gqlFieldsProvider.ForWebhooks()))
}

func fixUpdateApplicationRequest(id, updateInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: updateApplication(id: "%s", in: %s) {
    					%s
					}
				}`, id, updateInputGQL, tc.gqlFieldsProvider.ForApplication()))
}

func fixUpdateApplicationTemplateRequest(id, updateInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: updateApplicationTemplate(id: "%s", in: %s) {
    					%s
					}
				}`, id, updateInputGQL, tc.gqlFieldsProvider.ForApplicationTemplate()))
}

func fixUpdateLabelDefinitionRequest(ldInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: updateLabelDefinition(in: %s) {
						%s
					}
				}`, ldInputGQL, tc.gqlFieldsProvider.ForLabelDefinition()))
}

func fixGenerateOneTimeTokenForRuntimeRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: generateOneTimeTokenForRuntime(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForOneTimeToken()))
}

func fixGenerateOneTimeTokenForAppRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: generateOneTimeTokenForApplication(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForOneTimeToken()))
}

func fixUpdateIntegrationSystemRequest(id, integrationSystemInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: updateIntegrationSystem(id: "%s", in: %s) {
    					%s
					}
				}`, id, integrationSystemInGQL, tc.gqlFieldsProvider.ForIntegrationSystem()))
}

func fixGenerateClientCredentialsForApplicationRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: generateClientCredentialsForApplication(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForSystemAuth()))
}

func fixGenerateClientCredentialsForRuntimeRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: generateClientCredentialsForRuntime(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForSystemAuth()))
}

func fixGenerateClientCredentialsForIntegrationSystemRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: generateClientCredentialsForIntegrationSystem(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForSystemAuth()))
}

// SET
func fixSetApplicationLabelRequest(appID, labelKey string, labelValue interface{}) *gcli.Request {
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
			appID, labelKey, value, tc.gqlFieldsProvider.ForLabel()))
}

func fixSetRuntimeLabelRequest(runtimeID, labelKey string, labelValue interface{}) *gcli.Request {
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
				}`,
			runtimeID, labelKey, value, tc.gqlFieldsProvider.ForLabel()))
}

func fixSetAPIAuthRequest(apiID string, rtmID string, authInStr string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setAPIAuth(apiID: "%s", runtimeID: "%s", in: %s) {
					%s
				}
			}`, apiID, rtmID, authInStr, tc.gqlFieldsProvider.ForAPIRuntimeAuth()))
}

// QUERY
func fixApplicationForRuntimeRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(
			`query {
  			result: applicationsForRuntime(runtimeID: "%s", first:%d, after:"") { 
					%s 
				}
			}`,
			runtimeID, 4, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication()),
		))
}

func fixRuntimeRequestWithPaginationRequest(after int, cursor string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtimes(first:%d, after:"%s") {
					%s
				}
			}`, after, cursor, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForRuntime())))
}

func fixAPIRuntimeAuthRequest(applicationID string, runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
							result: application(id: "%s") {
								%s
							}
						}`, applicationID, tc.gqlFieldsProvider.ForApplication(gql.FieldCtx{
			"APIDefinition.auth": fmt.Sprintf(`auth(runtimeID: "%s") {%s}`, runtimeID, tc.gqlFieldsProvider.ForAPIRuntimeAuth()),
		})))
}

func fixRuntimeRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
					%s
				}}`, runtimeID, tc.gqlFieldsProvider.ForRuntime()))
}

func fixDeleteRuntimeRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{deleteRuntime(id: "%s") {
				id
			}
		}`, id))
}

func fixLabelDefinitionRequest(labelKey string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: labelDefinition(key: "%s") {
						%s
					}
				}`,
			labelKey, tc.gqlFieldsProvider.ForLabelDefinition()))
}

func fixApplicationRequest(applicationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
					%s
				}
			}`, applicationID, tc.gqlFieldsProvider.ForApplication()))
}

func fixApplicationTemplateRequest(applicationTemplateID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: applicationTemplate(id: "%s") {
					%s
				}
			}`, applicationTemplateID, tc.gqlFieldsProvider.ForApplicationTemplate()))
}

func fixLabelDefinitionsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
  		result:	labelDefinitions() {
			key
    		schema
  				}
			}`))
}

func fixApplicationsRequest(labelFilterInGQL string, first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications(filter: %s, first: %d, after: "%s") {
						%s
					}
				}`,
			labelFilterInGQL, first, after, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication())))
}

func fixApplicationTemplates(first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applicationTemplates(first: %d, after: "%s") {
						%s
					}
				}`,
			first, after, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplicationTemplate())))
}

func fixRuntimesRequest(labelFilterInGQL string, first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: runtimes(filter: %s, first: %d, after: "%s") {
						%s
					}
				}`,
			labelFilterInGQL, first, after, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForRuntime())))
}

func fixIntegrationSystemsRequest(first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: integrationSystems(first: %d, after: "%s") {
						%s
					}
				}`,
			first, after, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForIntegrationSystem())))
}

func fixIntegrationSystemRequest(intSysID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: integrationSystem(id: "%s") {
						%s
					}
				}`,
			intSysID, tc.gqlFieldsProvider.ForIntegrationSystem()))
}

// DELETE
func fixDeleteLabelDefinitionRequest(labelDefinitionKey string, deleteRelatedLabels bool) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteLabelDefinition(key: "%s", deleteRelatedLabels: %t) {
					%s
				}
			}`, labelDefinitionKey, deleteRelatedLabels, tc.gqlFieldsProvider.ForLabelDefinition()))
}

func fixDeleteApplicationLabelRequest(applicationID, labelKey string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteApplicationLabel(applicationID: "%s", key: "%s") {
					%s
				}
			}`, applicationID, labelKey, tc.gqlFieldsProvider.ForLabel()))
}

func fixDeleteAPIAuthRequestRequest(apiID string, rtmID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteAPIAuth(apiID: "%s",runtimeID: "%s") {
					%s
				} 
			}`, apiID, rtmID, tc.gqlFieldsProvider.ForAPIRuntimeAuth()))
}

func fixDeleteIntegrationSystemRequest(intSysID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteIntegrationSystem(id: "%s") {
					%s
				}
			}`, intSysID, tc.gqlFieldsProvider.ForIntegrationSystem()))
}

func fixDeleteSystemAuthForApplicationRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteSystemAuthForApplication(authID: "%s") {
					%s
				}
			}`, authID, tc.gqlFieldsProvider.ForSystemAuth()))
}

func fixDeleteSystemAuthForRuntimeRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteSystemAuthForRuntime(authID: "%s") {
					%s
				}
			}`, authID, tc.gqlFieldsProvider.ForSystemAuth()))
}

func fixDeleteSystemAuthForIntegrationSystemRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteSystemAuthForIntegrationSystem(authID: "%s") {
					%s
				}
			}`, authID, tc.gqlFieldsProvider.ForSystemAuth()))
}

func fixDeleteApplicationTemplate(appTemplateID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteApplicationTemplate(id: "%s") {
					%s
				}
			}`, appTemplateID, tc.gqlFieldsProvider.ForApplicationTemplate()))
}

func removeDoubleQuotesFromJSONKeys(in string) string {
	var validRegex = regexp.MustCompile(`"(\w+|\$\w+)"\s*:`)
	return validRegex.ReplaceAllString(in, `$1:`)
}

func fixRegisterApplicationFromTemplate(applicationFromTemplateInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplicationFromTemplate(in: %s) {
					%s
				}
			}`,
			applicationFromTemplateInputInGQL, tc.gqlFieldsProvider.ForApplication()))
}
