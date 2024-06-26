package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

func FixGetFormationsForObjectRequest(objectID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query{
				  result: formationsForObject(objectID: "%s"){
					%s
					formationAssignments(first:%d, after:"") {
						%s
					}
					status {%s}
				  }
				}`, objectID, testctx.Tc.GQLFieldsProvider.ForFormation(), 200, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForFormationAssignment()), testctx.Tc.GQLFieldsProvider.ForFormationStatus()))
}

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
				}`, formationName, testctx.Tc.GQLFieldsProvider.ForFormationWithStatus()))
}

func FixCreateFormationWithTemplateRequest(formationInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
				  result: createFormation(formation: %s){
					%s
				  }
				}`, formationInput, testctx.Tc.GQLFieldsProvider.ForFormationWithStatus()))
}

func FixDeleteFormationRequest(formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
				  result: deleteFormation(formation: {name: "%s"}){
					%s
				  }
				}`, formationName, testctx.Tc.GQLFieldsProvider.ForFormationWithStatus()))
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
			}`, objID, objType, formationName, testctx.Tc.GQLFieldsProvider.ForFormationWithStatus()))
}

func FixAssignFormationRequestWithInitialConfigurations(objID, objType, formationName, initialConfigurations string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
			  result: assignFormation(objectID:"%s",objectType: %s ,formation: {name: "%s"}, initialConfigurations: %s){
				%s
			  }
			}`, objID, objType, formationName, initialConfigurations, testctx.Tc.GQLFieldsProvider.ForFormationWithStatus()))
}

func FixUnassignFormationRequest(objID, objType, formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
			  result: unassignFormation(objectID:"%s",objectType: %s ,formation: {name: "%s"}){
				%s
			  }
			}`, objID, objType, formationName, testctx.Tc.GQLFieldsProvider.ForFormationWithStatus()))
}

func FixUnassignFormationGlobalRequest(objID, objType, formationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
			result: unassignFormationGlobal(objectID:"%s", objectType: %s, formation: "%s"){
					%s
					formationAssignments(first:%d, after:"") {
						%s
					}
					status {%s}
				}
			}`, objID, objType, formationID, testctx.Tc.GQLFieldsProvider.ForFormation(), 200, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForFormationAssignment()), testctx.Tc.GQLFieldsProvider.ForFormationStatus()))
}

func FixResynchronizeFormationNotificationsRequest(formationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
			  result: resynchronizeFormationNotifications(formationID:"%s"){
				%s
			  }
			}`, formationID, testctx.Tc.GQLFieldsProvider.ForFormationWithStatus()))
}

func FixFinalizeDraftFormationRequest(formationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
			  result: finalizeDraftFormation(formationID:"%s"){
				%s
			  }
			}`, formationID, testctx.Tc.GQLFieldsProvider.ForFormationWithStatus()))
}

func FixResynchronizeFormationNotificationsRequestWithResetOption(formationID string, reset bool) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
			  result: resynchronizeFormationNotifications(formationID:"%s", reset:%t){
				%s
			  }
			}`, formationID, reset, testctx.Tc.GQLFieldsProvider.ForFormationWithStatus()))
}

func FixFormationInput(formationName string, formationTemplateName *string) graphql.FormationInput {
	return graphql.FormationInput{
		Name:         formationName,
		TemplateName: formationTemplateName,
	}
}

func FixFormationInputWithState(formationName string, formationTemplateName, formationState *string) graphql.FormationInput {
	return graphql.FormationInput{
		Name:         formationName,
		TemplateName: formationTemplateName,
		State:        formationState,
	}
}
