package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

func FixListFormationAssignmentRequest(formationID string, pageSize int) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query{
				  result: formation(id: "%s"){
						formationAssignments(first:%d, after:"") {
							%s
						}
				  }
				}`, formationID, pageSize, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForFormationAssignment())))
}
