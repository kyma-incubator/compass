package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

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

func FixRefetchAPISpecRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: refetchAPISpec(apiID: "%s") {
						%s
					}
				}`,
			id, testctx.Tc.GQLFieldsProvider.ForAPISpec()))
}
