package e2e

import (
	"fmt"
	"testing"

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

func fixCreateIntegrationSystemRequest(integrationSystemInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createIntegrationSystem(in: %s) {
					%s
				}
			}`,
			integrationSystemInGQL, tc.gqlFieldsProvider.ForIntegrationSystem()))
}

// ADD
func fixAddApiRequest(applicationId, apiInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addAPI(applicationID: "%s", in: %s) {
					%s
				}
			}`, applicationId, apiInputInGQL, tc.gqlFieldsProvider.ForAPIDefinition()))
}

// UPDATE
func fixGenerateClientCredentialsForIntegrationSystem(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: generateClientCredentialsForIntegrationSystem(id: "%s") {
						%s
					}
				}`, id, tc.gqlFieldsProvider.ForSystemAuth()))
}

// DELETE
func fixDeleteApplicationRequest(t *testing.T, id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteApplication(id: "%s") {
			%s
		}	
	}`, id, tc.gqlFieldsProvider.ForApplication()))
}

func fixDeleteIntegrationSystem(intSysID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteIntegrationSystem(id: "%s") {
					%s
				}
			}`, intSysID, tc.gqlFieldsProvider.ForIntegrationSystem()))
}
