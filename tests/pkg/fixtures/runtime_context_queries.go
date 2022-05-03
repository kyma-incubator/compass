package fixtures

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func CreateRuntimeContext(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, runtimeID, key, value string) graphql.RuntimeContextExt {
	rtmCtx, err := testctx.Tc.Graphqlizer.RuntimeContextInputToGQL(FixRuntimeContextInput(key, value))
	require.NoError(t, err)

	addRtmCtxRequest := FixAddRuntimeContextRequest(runtimeID, rtmCtx)
	var resp graphql.RuntimeContextExt

	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, addRtmCtxRequest, &resp)
	require.NoError(t, err)

	return resp
}

func DeleteRuntimeContext(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	req := FixDeleteRuntimeContextRequest(id)

	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil))
}

func GetRuntimeContext(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, runtimeID, rtmCtxID string) graphql.RuntimeContextExt {
	req := FixRuntimeContextRequest(runtimeID, rtmCtxID)
	runtime := graphql.RuntimeExt{}
	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &runtime))
	return runtime.RuntimeContext
}
