package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

func FixGetOperationByIDRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: operation(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForOperation()))
}

func FixScheduleOperationByIDRequest(id string, priority int) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: scheduleOperation(operationID: "%s", priority: %d) {
					%s
				}
			}`, id, priority, testctx.Tc.GQLFieldsProvider.ForOperation()))
}
