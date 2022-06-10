package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/require"
)

func TestAddRuntimeContext(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	in := fixRuntimeInput("addRuntimeContext")

	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, &in)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	rtmCtxInput := fixtures.FixRuntimeContextInput("create", "create")
	rtmCtx, err := testctx.Tc.Graphqlizer.RuntimeContextInputToGQL(rtmCtxInput)
	require.NoError(t, err)

	addRtmCtxRequest := fixtures.FixAddRuntimeContextRequest(runtime.ID, rtmCtx)
	output := graphql.RuntimeContextExt{}

	// WHEN
	t.Log("Create runtimeContext")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addRtmCtxRequest, &output)
	defer fixtures.DeleteRuntimeContext(t, ctx, certSecuredGraphQLClient, tenantId, output.ID)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	assertions.AssertRuntimeContext(t, &rtmCtxInput, &output)

	saveExample(t, addRtmCtxRequest.Query(), "register runtime context")

	rtmCtxRequest := fixtures.FixRuntimeContextRequest(runtime.ID, output.ID)
	runtimeFromAPI := graphql.RuntimeExt{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, rtmCtxRequest, &runtimeFromAPI)
	require.NoError(t, err)

	assertions.AssertRuntimeContext(t, &rtmCtxInput, &runtimeFromAPI.RuntimeContext)
	saveExample(t, rtmCtxRequest.Query(), "query runtimeContext")
}

func TestQueryRuntimeContexts(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	in := fixRuntimeInput("addRuntimeContext")

	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, &in)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	rtmCtx1 := fixtures.CreateRuntimeContext(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID, "queryRuntimeContexts1", "queryRuntimeContexts1")
	defer fixtures.DeleteRuntimeContext(t, ctx, certSecuredGraphQLClient, tenantId, rtmCtx1.ID)

	rtmCtx2 := fixtures.CreateRuntimeContext(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID, "queryRuntimeContexts2", "queryRuntimeContexts2")
	defer fixtures.DeleteRuntimeContext(t, ctx, certSecuredGraphQLClient, tenantId, rtmCtx2.ID)

	rtmCtxsRequest := fixtures.FixGetRuntimeContextsRequest(runtime.ID)
	runtimeGql := graphql.RuntimeExt{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, rtmCtxsRequest, &runtimeGql)
	require.NoError(t, err)
	require.Equal(t, 2, len(runtimeGql.RuntimeContexts.Data))
	require.ElementsMatch(t, []*graphql.RuntimeContextExt{&rtmCtx1, &rtmCtx2}, runtimeGql.RuntimeContexts.Data)

	saveExample(t, rtmCtxsRequest.Query(), "query runtime contexts")
}

func TestUpdateRuntimeContext(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	in := fixRuntimeInput("addRuntimeContext")

	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, &in)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	rtmCtx := fixtures.CreateRuntimeContext(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID, "runtimeContext", "runtimeContext")
	defer fixtures.DeleteRuntimeContext(t, ctx, certSecuredGraphQLClient, tenantId, rtmCtx.ID)

	rtmCtxUpdateInput := fixtures.FixRuntimeContextInput("updateRuntimeContext", "updateRuntimeContext")
	rtmCtxUpdate, err := testctx.Tc.Graphqlizer.RuntimeContextInputToGQL(rtmCtxUpdateInput)
	require.NoError(t, err)

	updateRtmCtxReq := fixtures.FixUpdateRuntimeContextRequest(rtmCtx.ID, rtmCtxUpdate)
	runtimeContext := graphql.RuntimeContextExt{}

	// WHEN
	t.Log("Update runtime context")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateRtmCtxReq, &runtimeContext)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, runtimeContext.ID)

	assertions.AssertRuntimeContext(t, &rtmCtxUpdateInput, &runtimeContext)
	saveExample(t, updateRtmCtxReq.Query(), "update runtime context")
}

func TestDeleteRuntimeContext(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	in := fixRuntimeInput("addRuntimeContext")

	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, &in)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	rtmCtx := fixtures.CreateRuntimeContext(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID, "deleteRuntimeContext", "deleteRuntimeContext")

	rtmCtxDeleteReq := fixtures.FixDeleteRuntimeContextRequest(rtmCtx.ID)
	rtmCtxGql := graphql.RuntimeContext{}

	// WHEN
	t.Log("Delete runtimeContext")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, rtmCtxDeleteReq, &rtmCtxGql)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, rtmCtxGql.ID)
	require.Equal(t, "deleteRuntimeContext", rtmCtxGql.Key)
	require.Equal(t, "deleteRuntimeContext", rtmCtxGql.Value)

	saveExample(t, rtmCtxDeleteReq.Query(), "delete runtime context")
}
