package director

import (
	"fmt"

	gcli "github.com/machinebox/graphql"
)

func fixCreateApplicationRequest(applicationInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplication(in: %s) {
					%s
				}
			}`,
			applicationInGQL, tc.gqlFieldsProvider.ForApplication()))
}

func fixApplicationForRuntimeRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(
			`query {
  			result: applicationsForRuntime(runtimeID: "%s", first:%d, after:"") { 
					%s 
				}
			}`,
			runtimeID, 4, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication()),
		))
}

func fixCreateRuntimeRequest(runtimeInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createRuntime(in: %s) {
					%s
				}
			}`,
			runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
}

func fixRuntimeRequestWithPagination(after int, cursor string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtimes(first:%d, after:"%s") {
					%s
				}
			}`, after, cursor, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForRuntime())))
}
