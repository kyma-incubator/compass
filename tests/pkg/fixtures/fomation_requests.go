package fixtures

import (
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

func FixGetFormationRequest(formationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query{
				  result: formation(id: "%s"){
					%s
				  }
				}`, formationID, testctx.Tc.GQLFieldsProvider.ForFormation()))
}

func FixListFormationsRequestWithPageSize(pageSize int) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				  result: formations(first:%d, after:"") {
    					%s
					}
				}`, pageSize, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForFormation())))
}

func FixCreateFormationRequest(formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
				  result: createFormation(formation: {name: "%s"}){
					%s
				  }
				}`, formationName, testctx.Tc.GQLFieldsProvider.ForFormation()))
}

// todo:: delete
//func FixCreateFormationWithTemplateRequest(formationInput string) *gcli.Request {
//	return gcli.NewRequest(
//		fmt.Sprintf(`mutation{
//				  result: createFormation(formation: {formation: "%s"}){
//					%s
//				  }
//				}`, formationInput, testctx.Tc.GQLFieldsProvider.ForFormation()))
//}
//
//func FixFormationInput(formationName, formationTemplateName string) graphql.FormationInput {
//	return graphql.FormationInput{
//		Name: "",
//		// todo:: update director component and add formation template name
//	}
//
//}

func FixDeleteFormationRequest(formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
				  result: deleteFormation(formation: {name: "%s"}){
					%s
				  }
				}`, formationName, testctx.Tc.GQLFieldsProvider.ForFormation()))
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
