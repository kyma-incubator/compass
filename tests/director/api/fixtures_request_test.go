package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	gcli "github.com/machinebox/graphql"
)

func fixRegisterApplicationRequest(applicationInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplication(in: %s) {
					%s
				}
			}`,
			applicationInGQL, tc.gqlFieldsProvider.ForApplication()))
}

func fixUnregisterApplicationRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			unregisterApplication(id: "%s") {
				id
			}
		}`, id))
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

func fixRegisterRuntimeRequest(runtimeInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerRuntime(in: %s) {
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

func fixRegisterIntegrationSystemRequest(integrationSystemInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerIntegrationSystem(in: %s) {
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

func fixDeleteDocumentRequest(docID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteDocument(id: "%s") {
					id
				}
			}`, docID))
}

func fixAddDocumentToPackageRequest(packageID, documentInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addDocumentToPackage(packageID: "%s", in: %s) {
 				%s
			}				
		}`, packageID, documentInputInGQL, tc.gqlFieldsProvider.ForDocument()))
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

func fixDeleteWebhookRequest(webhookID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteWebhook(webhookID: "%s") {
				%s
			}
		}`, webhookID, tc.gqlFieldsProvider.ForWebhooks()))
}

func fixAddAPIRequest(appID, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addAPIDefinition(applicationID: "%s", in: %s) {
				%s
			}
		}
		`, appID, APIInputGQL, tc.gqlFieldsProvider.ForAPIDefinition()))
}

func fixAddAPIToPackageRequest(pkgID, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addAPIDefinitionToPackage(packageID: "%s", in: %s) {
				%s
			}
		}
		`, pkgID, APIInputGQL, tc.gqlFieldsProvider.ForAPIDefinition()))
}

func fixUpdateAPIRequest(apiID, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: updateAPIDefinition(id: "%s", in: %s) {
				%s
			}
		}
		`, apiID, APIInputGQL, tc.gqlFieldsProvider.ForAPIDefinition()))
}

func fixDeleteAPIRequest(apiID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: deleteAPIDefinition(id: "%s") {
				id
			}
		}`, apiID))
}

func fixAddEventAPIRequest(appID, eventAPIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addEventDefinition(applicationID: "%s", in: %s) {
				%s
			}
		}
		`, appID, eventAPIInputGQL, tc.gqlFieldsProvider.ForEventDefinition()))
}

func fixAddEventAPIToPackageRequest(pkgID, eventAPIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addEventDefinitionToPackage(packageID: "%s", in: %s) {
				%s
			}
		}
		`, pkgID, eventAPIInputGQL, tc.gqlFieldsProvider.ForEventDefinition()))
}

func fixUpdateEventAPIRequest(eventAPIID, eventAPIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: updateEventDefinition(id: "%s", in: %s) {
				%s
			}
		}
		`, eventAPIID, eventAPIInputGQL, tc.gqlFieldsProvider.ForEventDefinition()))
}

func fixDeleteEventAPIRequest(eventAPIID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteEventDefinition(id: "%s") {
				id
			}
		}`, eventAPIID))
}

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

func fixRequestOneTimeTokenForRuntime(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestOneTimeTokenForRuntime(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForOneTimeTokenForRuntime()))
}

func fixRequestOneTimeTokenForApp(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestOneTimeTokenForApplication(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForOneTimeTokenForApplication()))
}

func fixUpdateIntegrationSystemRequest(id, integrationSystemInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: updateIntegrationSystem(id: "%s", in: %s) {
    					%s
					}
				}`, id, integrationSystemInGQL, tc.gqlFieldsProvider.ForIntegrationSystem()))
}

func fixRequestClientCredentialsForApplication(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestClientCredentialsForApplication(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForSystemAuth()))
}

func fixRequestClientCredentialsForRuntime(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestClientCredentialsForRuntime(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForSystemAuth()))
}

func fixRequestClientCredentialsForIntegrationSystem(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestClientCredentialsForIntegrationSystem(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForSystemAuth()))
}

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
				}`, runtimeID, labelKey, value, tc.gqlFieldsProvider.ForLabel()))
}

func fixSetAPIAuthRequest(apiID string, rtmID string, authInStr string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setAPIAuth(apiID: "%s", runtimeID: "%s", in: %s) {
					%s
				}
			}`, apiID, rtmID, authInStr, tc.gqlFieldsProvider.ForAPIRuntimeAuth()))
}

func fixApplicationForRuntimeRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
  			result: applicationsForRuntime(runtimeID: "%s", first:%d, after:"") { 
					%s 
				}
			}`, runtimeID, 4, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication())))
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
		}`, applicationID, tc.gqlFieldsProvider.ForApplication(graphqlizer.FieldCtx{
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

func fixUnregisterRuntimeRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{unregisterRuntime(id: "%s") {
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

func fixApplicationsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications {
						%s
					}
				}`,
			tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication())))
}

func fixApplicationsFilteredPageableRequest(labelFilterInGQL string, first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications(filter: %s, first: %d, after: "%s") {
						%s
					}
				}`,
			labelFilterInGQL, first, after, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication())))
}

func fixApplicationsPageableRequest(first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications(first: %d, after: "%s") {
						%s
					}
				}`,
			first, after, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication())))
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

func fixRuntimesRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: runtimes {
						%s
					}
				}`,
			tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForRuntime())))
}

func fixRuntimesFilteredPageableRequest(labelFilterInGQL string, first int, after string) *gcli.Request {
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

func fixTenantsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: tenants {
						%s
					}
				}`, tc.gqlFieldsProvider.ForTenant()))
}

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

func fixunregisterIntegrationSystem(intSysID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: unregisterIntegrationSystem(id: "%s") {
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

func fixSetDefaultEventingForApplication(appID string, runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setDefaultEventingForApplication(runtimeID: "%s", appID: "%s") {
					%s
				}
			}`,
			runtimeID, appID, tc.gqlFieldsProvider.ForEventingConfiguration()))
}

func fixAPIDefinitionInPackageRequest(appID, pkgID, apiID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							apiDefinition(id: "%s"){
						%s
						}					
					}
				}
			}`, appID, pkgID, apiID, tc.gqlFieldsProvider.ForAPIDefinition()))
}

func fixEventDefinitionInPackageRequest(appID, pkgID, eventID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							eventDefinition(id: "%s"){
						%s
						}					
					}
				}
			}`, appID, pkgID, eventID, tc.gqlFieldsProvider.ForEventDefinition()))
}

func fixDocumentInPackageRequest(appID, pkgID, docID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							document(id: "%s"){
						%s
						}					
					}
				}
			}`, appID, pkgID, docID, tc.gqlFieldsProvider.ForDocument()))
}

func fixAPIDefinitionsInPackageRequest(appID, pkgID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							apiDefinitions{
						%s
						}					
					}
				}
			}`, appID, pkgID, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForAPIDefinition())))

}

func fixEventDefinitionsInPackageRequest(appID, pkgID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							eventDefinitions{
						%s
						}					
					}
				}
			}`, appID, pkgID, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForEventDefinition())))
}

func fixDocumentsInPackageRequest(appID, pkgID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							documents{
						%s
						}					
					}
				}
			}`, appID, pkgID, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForDocument())))
}

func fixAddPackageRequest(appID, pkgCreateInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addPackage(applicationID: "%s", in: %s) {
				%s
			}}`, appID, pkgCreateInput, tc.gqlFieldsProvider.ForPackage()))
}

func fixUpdatePackageRequest(packageID, pkgUpdateInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updatePackage(id: "%s", in: %s) {
				%s
			}
		}`, packageID, pkgUpdateInput, tc.gqlFieldsProvider.ForPackage()))
}

func fixDeletePackageRequest(packageID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deletePackage(id: "%s") {
				%s
			}
		}`, packageID, tc.gqlFieldsProvider.ForPackage()))
}

func fixSetPackageInstanceAuthRequest(authID, apiAuthInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setPackageInstanceAuth(authID: "%s", in: %s) {
				%s
			}
		}`, authID, apiAuthInput, tc.gqlFieldsProvider.ForPackageInstanceAuth()))
}

func fixDeletePackageInstanceAuthRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deletePackageInstanceAuth(authID: "%s") {
				%s
			}
		}`, authID, tc.gqlFieldsProvider.ForPackageInstanceAuth()))
}

func fixRequestPackageInstanceAuthCreationRequest(packageID, pkgInstanceAuthRequestInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestPackageInstanceAuthCreation(packageID: "%s", in: %s) {
				%s
			}
		}`, packageID, pkgInstanceAuthRequestInput, tc.gqlFieldsProvider.ForPackageInstanceAuth()))
}

func fixRequestPackageInstanceAuthDeletionRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestPackageInstanceAuthDeletion(authID: "%s") {
				%s
			}
		}`, authID, tc.gqlFieldsProvider.ForPackageInstanceAuth()))
}

func fixPackageRequest(applicationID string, packageID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID, tc.gqlFieldsProvider.ForApplication(graphqlizer.FieldCtx{
			"Application.package": fmt.Sprintf(`package(id: "%s") {%s}`, packageID, tc.gqlFieldsProvider.ForPackage()),
		})))
}

func fixAPIDefinitionRequest(applicationID string, apiID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID, tc.gqlFieldsProvider.ForApplication(graphqlizer.FieldCtx{
			"Application.apiDefinition": fmt.Sprintf(`apiDefinition(id: "%s") {%s}`, apiID, tc.gqlFieldsProvider.ForAPIDefinition()),
		})))
}

func fixEventDefinitionRequest(applicationID string, eventDefID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID, tc.gqlFieldsProvider.ForApplication(graphqlizer.FieldCtx{
			"Application.eventDefinition": fmt.Sprintf(`eventDefinition(id: "%s") {%s}`, eventDefID, tc.gqlFieldsProvider.ForEventDefinition()),
		})))
}

func fixPackageWithInstanceAuthRequest(applicationID string, packageID string, instanceAuthID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID,
			tc.gqlFieldsProvider.ForApplication(
				graphqlizer.FieldCtx{"Application.package": fmt.Sprintf(`package(id: "%s") {%s}`,
					packageID,
					tc.gqlFieldsProvider.ForPackage(graphqlizer.FieldCtx{
						"Package.instanceAuth": fmt.Sprintf(`instanceAuth(id: "%s") {%s}`,
							instanceAuthID,
							tc.gqlFieldsProvider.ForPackageInstanceAuth()),
					})),
				})))
}

func fixPackagesRequest(applicationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID, tc.gqlFieldsProvider.ForApplication()))
}

func fixDeleteDefaultEventingForApplication(appID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: deleteDefaultEventingForApplication(appID: "%s") {
						%s
					}
				}`,
			appID, tc.gqlFieldsProvider.ForEventingConfiguration()))
}

func fixSetAutomaticScenarioAssignmentRequest(automaticScenarioAssignmentInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: setAutomaticScenarioAssignment(in: %s) {
						%s
					}
				}`,
			automaticScenarioAssignmentInput, tc.gqlFieldsProvider.ForAutomaticScenarioAssignment()))
}

func fixDeleteAutomaticScenarioAssignmentForScenarioRequest(scenario string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
            result: deleteAutomaticScenarioAssignmentForScenario(scenarioName: "%s") {
                  %s
               }
            }`,
			scenario, tc.gqlFieldsProvider.ForAutomaticScenarioAssignment()))
}

func fixDeleteAutomaticScenarioAssignmentsForSelectorRequest(labelSelectorInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
            result: deleteAutomaticScenarioAssignmentForSelector(selector: %s) {
                  %s
               }
            }`,
			labelSelectorInput, tc.gqlFieldsProvider.ForAutomaticScenarioAssignment()))
}
func fixAutomaticScenarioAssignmentsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: automaticScenarioAssignments {
						%s
					}
				}`,
			tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForAutomaticScenarioAssignment())))
}

func fixAutomaticScenarioAssignmentForSelectorRequest(labelSelectorInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: automaticScenarioAssignmentForSelector(selector: %s) {
						%s
					}
				}`,
			labelSelectorInput, tc.gqlFieldsProvider.ForAutomaticScenarioAssignment()))
}

func fixAutomaticScenarioAssignmentForScenarioRequest(scenarioName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: automaticScenarioAssignmentForScenario(scenarioName: "%s") {
						%s
					}
				}`,
			scenarioName, tc.gqlFieldsProvider.ForAutomaticScenarioAssignment()))
}
