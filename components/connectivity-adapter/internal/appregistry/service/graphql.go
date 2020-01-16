package service

import (
	"fmt"

	gcli "github.com/machinebox/graphql"
)

func prepareUnregisterApplicationRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		unregisterApplication(id: "%s") {
			id
		}	
	}`, id))
}
