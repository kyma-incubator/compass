package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

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
