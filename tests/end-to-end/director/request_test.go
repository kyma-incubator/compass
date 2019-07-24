package director

import (
	"fmt"

	gcli "github.com/machinebox/graphql"
)

func fixCreateApplicationRequest(inStr string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplication(in: %s) {
					%s
				}
			}`,
			inStr, tc.gqlFieldsProvider.ForApplication()))
}
