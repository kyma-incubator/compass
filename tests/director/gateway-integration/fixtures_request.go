package gateway_integration

import (
	"fmt"
	"testing"

	gcli "github.com/machinebox/graphql"
)

// CREATE
func fixRegisterApplicationRequest(applicationInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplication(in: %s) {
					%s
				}
			}`,
			applicationInGQL, tc.gqlFieldsProvider.ForApplication()))
}

func fixCreateIntegrationSystemRequest(integrationSystemInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerIntegrationSystem(in: %s) {
					%s
				}
			}`,
			integrationSystemInGQL, tc.gqlFieldsProvider.ForIntegrationSystem()))
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

func fixCreateApplicationTemplateRequest(applicationTemplateInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplicationTemplate(in: %s) {
					%s
				}
			}`,
			applicationTemplateInGQL, tc.gqlFieldsProvider.ForApplicationTemplate()))
}

func fixGetIntegrationSystemRequest(integrationSystemID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: integrationSystem(id: "%s") {
					%s
				}
			}`,
			integrationSystemID, tc.gqlFieldsProvider.ForIntegrationSystem()))
}

func fixGetViewerRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: viewer {
					%s
				}
			}`,
			tc.gqlFieldsProvider.ForViewer()))
}

// ADD
func fixAddApiRequest(applicationId, apiInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addAPIDefinition(applicationID: "%s", in: %s) {
					%s
				}
			}`, applicationId, apiInputInGQL, tc.gqlFieldsProvider.ForAPIDefinition()))
}

// UPDATE
func fixGenerateClientCredentialsForIntegrationSystem(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestClientCredentialsForIntegrationSystem(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForSystemAuth()))
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

// DELETE
func fixDeleteApplicationRequest(t *testing.T, id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		unregisterApplication(id: "%s") {
			%s
		}	
	}`, id, tc.gqlFieldsProvider.ForApplication()))
}

func fixUnregisterRuntime(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{unregisterRuntime(id: "%s") {
				%s
			}
		}`, id, tc.gqlFieldsProvider.ForRuntime()))
}

func fixUnregisterIntegrationSystem(intSysID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: unregisterIntegrationSystem(id: "%s") {
					%s
				}
			}`, intSysID, tc.gqlFieldsProvider.ForIntegrationSystem()))
}

func fixGenerateOneTimeTokenForApplication(appID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestOneTimeTokenForApplication(id: "%s") {
					%s
				}
			}`, appID, tc.gqlFieldsProvider.ForOneTimeTokenForApplication()))
}

func fixDeleteApplicationTemplateRequest(appTemplateID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteApplicationTemplate(id: "%s") {
					%s
				}
			}`, appTemplateID, tc.gqlFieldsProvider.ForApplicationTemplate()))
}

//GET
func fixGetApplicationRequest(applicationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
					%s
				}
			}`, applicationID, tc.gqlFieldsProvider.ForApplication()))
}

func fixIntegrationSystemRequest(integrationSystemID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: integrationSystem(id: "%s") {
					%s
				}
			}`,
			integrationSystemID, tc.gqlFieldsProvider.ForIntegrationSystem()))
}

func fixApplicationTemplateRequest(applicationTemplateID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: applicationTemplate(id: "%s") {
					%s
				}
			}`, applicationTemplateID, tc.gqlFieldsProvider.ForApplicationTemplate()))
}

func fixRuntimeRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
					%s
				}}`, runtimeID, tc.gqlFieldsProvider.ForRuntime()))
}
