package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

// FixGetLatestAssignmentOperation creates a new GraphQL request for getting the latest assignment operation
func FixGetLatestAssignmentOperation(formationID, assignmentID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query{
				  result: getLatestOperation(formationID: "%s", assignmentID: "%s"){
						%s
				  }
				}`, formationID, assignmentID, testctx.Tc.GQLFieldsProvider.ForAssignmentOperation()))
}
