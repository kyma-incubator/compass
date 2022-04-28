package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

func FixRuntimeContextInput(key, value string) graphql.RuntimeContextInput {
	return graphql.RuntimeContextInput{
		Key:   key,
		Value: value,
	}
}

func FixAddRuntimeContextRequest(runtimeID, runtimeContextInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerRuntimeContext(runtimeID: "%s", in: %s) {
				%s
			}}`, runtimeID, runtimeContextInput, testctx.Tc.GQLFieldsProvider.ForRuntimeContext()))
}

func FixDeleteRuntimeContextRequest(runtimeContextID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: unregisterRuntimeContext(id: "%s") {
				%s
			}
		}`, runtimeContextID, testctx.Tc.GQLFieldsProvider.ForRuntimeContext()))
}

func FixGetRuntimeContextsRequest(runtimeID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
				%s
				}
			}`, runtimeID, testctx.Tc.GQLFieldsProvider.ForRuntime()))
}

func FixRuntimeContextRequest(runtimeID string, runtimeContextID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
				%s
				}
			}`, runtimeID, testctx.Tc.GQLFieldsProvider.ForRuntime(graphqlizer.FieldCtx{
			"Runtime.runtimeContext": fmt.Sprintf(`runtimeContext(id: "%s") {%s}`, runtimeContextID, testctx.Tc.GQLFieldsProvider.ForRuntimeContext()),
		})))
}

func FixUpdateRuntimeContextRequest(runtimeContextID, runtimeContextUpdateInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateRuntimeContext(id: "%s", in: %s) {
				%s
			}
		}`, runtimeContextID, runtimeContextUpdateInput, testctx.Tc.GQLFieldsProvider.ForRuntimeContext()))
}
