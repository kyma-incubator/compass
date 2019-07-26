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

func fixCreateApplicationForRuntimeRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(
			`query {
  			result: applicationsForRuntime(runtimeID: "%s", first:%d, after:"%s") { 
					%s 
				}
			}`,
			runtimeID, 2, "next", tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication()),
		))
}
