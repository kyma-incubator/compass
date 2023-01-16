package fixtures

import (
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

func FixCreateFormationConstraintRequest(createFormationConstraintGQLInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				  result: createFormationConstraint(formationConstraint: %s) {
    					%s
					}
				}`, createFormationConstraintGQLInput, testctx.Tc.GQLFieldsProvider.ForFormationConstraint()))
}

func FixDeleteFormationConstraintRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				  result: deleteFormationConstraint(id: "%s") {
    					%s
					}
				}`, id, testctx.Tc.GQLFieldsProvider.ForFormationConstraint()))
}

func FixQueryFormationConstraintRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				  result: formationConstraint(id: "%s") {
    					%s
					}
				}`, id, testctx.Tc.GQLFieldsProvider.ForFormationConstraint()))
}

func FixQueryFormationConstraintsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				  result: formationConstraints() {
    					%s
					}
				}`, testctx.Tc.GQLFieldsProvider.ForFormationConstraint()))
}
