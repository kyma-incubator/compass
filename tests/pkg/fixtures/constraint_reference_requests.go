package fixtures

import (
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

func FixAttachConstraintToFormationTemplateRequest(constraintID, formationTemplateID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				  result: attachConstraintToFormationTemplate(constraintID: %s, formationTemplateID: %s) {
    					%s
					}
				}`, constraintID, formationTemplateID, testctx.Tc.GQLFieldsProvider.ForFormationTemplateConstraintReference()))
}

func FixDetachConstraintFromFormationTemplateRequest(constraintID, formationTemplateID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				  result: detachConstraintFromFormationTemplate(constraintID: %s, formationTemplateID: %s) {
    					%s
					}
				}`, constraintID, formationTemplateID, testctx.Tc.GQLFieldsProvider.ForFormationTemplateConstraintReference()))
}
