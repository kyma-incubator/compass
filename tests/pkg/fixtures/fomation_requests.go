package fixtures

import (
	"fmt"

	gcli "github.com/machinebox/graphql"
)

func FixCreateFormationRequest(formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
				  createFormation(formation: {name: %s}){
					name
				  }
				}`, formationName))
}
