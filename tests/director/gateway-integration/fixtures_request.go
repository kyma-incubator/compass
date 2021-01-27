package gateway_integration

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

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
			applicationInGQL, tc.GQLFieldsProvider.ForApplication()))
}

func fixUpdateApplicationRequest(id, updateInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: updateApplication(id: "%s", in: %s) {
    					%s
					}
				}`, id, updateInputGQL, tc.GQLFieldsProvider.ForApplication()))
}

func fixCreateIntegrationSystemRequest(integrationSystemInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerIntegrationSystem(in: %s) {
					%s
				}
			}`,
			integrationSystemInGQL, tc.GQLFieldsProvider.ForIntegrationSystem()))
}

func fixRegisterRuntimeRequest(runtimeInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerRuntime(in: %s) {
					%s
				}
			}`,
			runtimeInGQL, tc.GQLFieldsProvider.ForRuntime()))
}

func fixCreateApplicationTemplateRequest(applicationTemplateInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplicationTemplate(in: %s) {
					%s
				}
			}`,
			applicationTemplateInGQL, tc.GQLFieldsProvider.ForApplicationTemplate()))
}

func fixGetIntegrationSystemRequest(integrationSystemID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: integrationSystem(id: "%s") {
					%s
				}
			}`,
			integrationSystemID, tc.GQLFieldsProvider.ForIntegrationSystem()))
}

func fixGetViewerRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: viewer {
					%s
				}
			}`,
			tc.GQLFieldsProvider.ForViewer()))
}

// ADD
func fixAddAPIToBundleRequest(bndlID, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addAPIDefinitionToBundle(bundleID: "%s", in: %s) {
				%s
			}
		}
		`, bndlID, APIInputGQL, tc.GQLFieldsProvider.ForAPIDefinition()))
}

func fixBundleCreateInput(name string) graphql.BundleCreateInput {
	return graphql.BundleCreateInput{
		Name: name,
	}
}

func fixAddBundleRequest(appID, bndlCreateInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addBundle(applicationID: "%s", in: %s) {
				%s
			}}`, appID, bndlCreateInput, tc.GQLFieldsProvider.ForBundle()))
}

// UPDATE
func fixGenerateClientCredentialsForIntegrationSystem(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestClientCredentialsForIntegrationSystem(id: "%s") {
						%s
					}
				}`, id, tc.GQLFieldsProvider.ForSystemAuth()))
}

func fixRequestClientCredentialsForApplication(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestClientCredentialsForApplication(id: "%s") {
						%s
					}
				}`, id, tc.GQLFieldsProvider.ForSystemAuth()))
}

func fixRequestClientCredentialsForRuntime(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestClientCredentialsForRuntime(id: "%s") {
						%s
					}
				}`, id, tc.GQLFieldsProvider.ForSystemAuth()))
}

// DELETE
func fixDeleteApplicationRequest(t *testing.T, id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		unregisterApplication(id: "%s") {
			%s
		}	
	}`, id, tc.GQLFieldsProvider.ForApplication()))
}

func fixUnregisterRuntimeRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{unregisterRuntime(id: "%s") {
				%s
			}
		}`, id, tc.GQLFieldsProvider.ForRuntime()))
}

func fixUnregisterIntegrationSystem(intSysID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: unregisterIntegrationSystem(id: "%s") {
					%s
				}
			}`, intSysID, tc.GQLFieldsProvider.ForIntegrationSystem()))
}

func fixGenerateOneTimeTokenForApplication(appID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestOneTimeTokenForApplication(id: "%s") {
					%s
				}
			}`, appID, tc.GQLFieldsProvider.ForOneTimeTokenForApplication()))
}

func fixGenerateOneTimeTokenForRuntime(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestOneTimeTokenForRuntime(id: "%s") {
					%s
				}
			}`, runtimeID, tc.GQLFieldsProvider.ForOneTimeTokenForRuntime()))
}

func fixDeleteApplicationTemplateRequest(appTemplateID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteApplicationTemplate(id: "%s") {
					%s
				}
			}`, appTemplateID, tc.GQLFieldsProvider.ForApplicationTemplate()))
}

func fixDeleteBundleRequest(bundleID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteBundle(id: "%s") {
				%s
			}
		}`, bundleID, tc.GQLFieldsProvider.ForBundle()))
}

//GET
func fixGetApplicationRequest(applicationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
					%s
				}
			}`, applicationID, tc.GQLFieldsProvider.ForApplication()))
}

func fixIntegrationSystemRequest(integrationSystemID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: integrationSystem(id: "%s") {
					%s
				}
			}`,
			integrationSystemID, tc.GQLFieldsProvider.ForIntegrationSystem()))
}

func fixApplicationTemplateRequest(applicationTemplateID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: applicationTemplate(id: "%s") {
					%s
				}
			}`, applicationTemplateID, tc.GQLFieldsProvider.ForApplicationTemplate()))
}

func fixRuntimeRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
					%s
				}}`, runtimeID, tc.GQLFieldsProvider.ForRuntime()))
}
