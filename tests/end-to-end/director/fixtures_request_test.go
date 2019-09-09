package director

import (
	"encoding/json"
	"errors"
	"fmt"

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

//UPDATE
func fixUpdateLabelDefinitionRequest(ldInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: updateLabelDefinition(in: %s) {
						%s
					}
				}`, ldInputGQL, tc.gqlFieldsProvider.ForLabelDefinition()))
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
			}`, apiID, rtmID, authInStr, tc.gqlFieldsProvider.ForRuntimeAuth()))
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

func fixRuntimeRequestWithPagination(after int, cursor string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtimes(first:%d, after:"%s") {
					%s
				}
			}`, after, cursor, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForRuntime())))
}

func fixRuntimeAuthRequest(applicationID string, runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
							result: application(id: "%s") {
								%s
							}
						}`, applicationID, tc.gqlFieldsProvider.ForApplication(fieldCtx{
			"APIDefinition.auth": fmt.Sprintf(`auth(runtimeID: "%s") {%s}`, runtimeID, tc.gqlFieldsProvider.ForRuntimeAuth()),
		})))
}

func fixRuntimeQuery(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
					%s
				}}`, runtimeID, tc.gqlFieldsProvider.ForRuntime()))
}

func fixDeleteRuntime(id string) *gcli.Request {
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

func fixLabelDefinitionsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
  		result:	labelDefinitions() {
			key
    		schema
  				}
			}`))
}

func fixApplications(labelFilterInGQL string, first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: applications(filter: %s, first: %d, after: "%s") {
						%s
					}
				}`,
			labelFilterInGQL, first, after, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication())))
}

func fixRuntimes(labelFilterInGQL string, first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: runtimes(filter: %s, first: %d, after: "%s") {
						%s
					}
				}`,
			labelFilterInGQL, first, after, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForRuntime())))
}

// DELETE
func fixDeleteLabelDefinition(labelDefinitionKey string, deleteRelatedLabels bool) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteLabelDefinition(key: "%s", deleteRelatedLabels: %t) {
					%s
				}
			}`, labelDefinitionKey, deleteRelatedLabels, tc.gqlFieldsProvider.ForLabelDefinition()))
}

func fixDeleteApplicationLabel(applicationID, labelKey string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteApplicationLabel(applicationID: "%s", key: "%s") {
					%s
				}
			}`, applicationID, labelKey, tc.gqlFieldsProvider.ForLabel()))
}

func fixDeleteAPIAuthRequest(apiID string, rtmID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteAPIAuth(apiID: "%s",runtimeID: "%s") {
					%s
				} 
			}`, apiID, rtmID, tc.gqlFieldsProvider.ForRuntimeAuth()))
}
