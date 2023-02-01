package fixtures

import (
	"fmt"
	gcli "github.com/machinebox/graphql"
)

func FixAttachConstraintToFormationTemplateRequest(constraintID, formationTemplateID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				  result: attachConstraintToFormationTemplate(constraintID: "%s", formationTemplateID: "%s") {
    					constraintID
			            formationTemplateID
					}
				}`, constraintID, formationTemplateID))
}

func FixDetachConstraintFromFormationTemplateRequest(constraintID, formationTemplateID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				  result: detachConstraintFromFormationTemplate(constraintID: "%s", formationTemplateID: "%s") {
    					constraintID
			            formationTemplateID
					}
				}`, constraintID, formationTemplateID))
}
