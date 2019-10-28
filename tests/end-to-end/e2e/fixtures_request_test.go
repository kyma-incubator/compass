package e2e

import (
	"fmt"
	"regexp"
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

func FixCreateIntegrationSystemRequest(integrationSystemInGQL string) *gcli.Request {
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

// QUERY
func fixApplicationRequest(applicationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
					%s
				}
			}`, applicationID, tc.gqlFieldsProvider.ForApplication()))
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

func fixDeleteSystemAuthForIntegrationSystem(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteSystemAuthForIntegrationSystem(authID: "%s") {
					%s
				}
			}`, authID, tc.gqlFieldsProvider.ForSystemAuth()))
}

func removeDoubleQuotesFromJSONKeys(in string) string {
	var validRegex = regexp.MustCompile(`"(\w+|\$\w+)"\s*:`)
	return validRegex.ReplaceAllString(in, `$1:`)
}
