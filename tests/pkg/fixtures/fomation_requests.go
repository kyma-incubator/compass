package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

func FixGetFormationRequest(formationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query{
				  result: formation(id: "%s"){
					%s
				  }
				}`, formationID, testctx.Tc.GQLFieldsProvider.ForFormationWithStatus()))
}

func FixGetFormationByNameRequest(formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query{
				  result: formationByName(name: "%s"){
					%s
				  }
				}`, formationName, testctx.Tc.GQLFieldsProvider.ForFormationWithStatus()))
}

func FixListFormationsRequestWithPageSize(pageSize int) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				  result: formations(first:%d, after:"") {
    					%s
					}
				}`, pageSize, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForFormationWithStatus())))
}

func FixCreateFormationRequest(formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
				  result: createFormation(formation: {name: "%s"}){
					%s
				  }
				}`, formationName, testctx.Tc.GQLFieldsProvider.ForFormation()))
}

func FixCreateFormationWithTemplateRequest(formationInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
				  result: createFormation(formation: %s){
					%s
				  }
				}`, formationInput, testctx.Tc.GQLFieldsProvider.ForFormation()))
}

func FixDeleteFormationRequest(formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
				  result: deleteFormation(formation: {name: "%s"}){
					%s
				  }
				}`, formationName, testctx.Tc.GQLFieldsProvider.ForFormation()))
}

func FixDeleteFormationWithTemplateRequest(formationInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
				  result: deleteFormation(formation: %s){
					%s
				  }
				}`, formationInput, testctx.Tc.GQLFieldsProvider.ForFormation()))
}

func FixAssignFormationRequest(objID, objType, formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
			  result: assignFormation(objectID:"%s",objectType: %s ,formation: {name: "%s"}){
				%s
			  }
			}`, objID, objType, formationName, testctx.Tc.GQLFieldsProvider.ForFormation()))
}

func FixUnassignFormationRequest(objID, objType, formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
			  result: unassignFormation(objectID:"%s",objectType: %s ,formation: {name: "%s"}){
				%s
			  }
			}`, objID, objType, formationName, testctx.Tc.GQLFieldsProvider.ForFormation()))
}

func FixFormationInput(formationName string, formationTemplateName *string) graphql.FormationInput {
	return graphql.FormationInput{
		Name:         formationName,
		TemplateName: formationTemplateName,
	}
}
