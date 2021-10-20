package fixtures

import (
	"fmt"

	gcli "github.com/machinebox/graphql"
)

func FixCreateFormationRequest(formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
				  result: createFormation(formation: {name: "%s"}){
					name
				  }
				}`, formationName))
}

func FixDeleteFormationRequest(formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
				  result: deleteFormation(formation: {name: "%s"}){
					name
				  }
				}`, formationName))
}

func FixAssignFormationRequest(objID, objType, formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
			  result: assignFormation(objectID:"%s",objectType: %s ,formation: {name: "%s"}){
				name
			  }
			}`, objID, objType, formationName))
}

func FixUnassignFormationRequest(objID, objType, formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
			  result: unassignFormation(objectID:"%s",objectType: %s ,formation: {name: "%s"}){
				name
			  }
			}`, objID, objType, formationName))
}
