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

func fixGetIntegrationSystemRequest(integrationSystemID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: integrationSystem(id: "%s") {
					%s
				}
			}`,
			integrationSystemID, tc.gqlFieldsProvider.ForIntegrationSystem()))
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

// DELETE
func fixDeleteApplicationRequest(t *testing.T, id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		unregisterApplication(id: "%s") {
			%s
		}	
	}`, id, tc.gqlFieldsProvider.ForApplication()))
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
			}`, appID, tc.gqlFieldsProvider.ForOneTimeToken()))
}
