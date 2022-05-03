package fixtures

import (
	"context"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func RegisterRuntime(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, name, tenant string) (graphql.RuntimeExt, error) {
	in := FixRuntimeInput(name)
	return RegisterRuntimeFromInputWithinTenant(t, ctx, gqlClient, tenant, &in)
}

func RegisterRuntimeFromInputWithinTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string, input *graphql.RuntimeInput) (graphql.RuntimeExt, error) {
	inputGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(*input)
	require.NoError(t, err)

	registerRuntimeRequest := FixRegisterRuntimeRequest(inputGQL)
	var runtime graphql.RuntimeExt

	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, registerRuntimeRequest, &runtime)
	return runtime, err
}

func RegisterRuntimeFromInputWithoutTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, input *graphql.RuntimeInput) graphql.RuntimeExt {
	inputGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(*input)
	require.NoError(t, err)

	registerRuntimeRequest := FixRegisterRuntimeRequest(inputGQL)
	var runtime graphql.RuntimeExt

	err = testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, registerRuntimeRequest, &runtime)
	require.NoError(t, err)
	return runtime
}

func UpdateRuntimeWithinTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, id string, in graphql.RuntimeInput) (graphql.RuntimeExt, error) {
	inputGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(in)
	require.NoError(t, err)

	updateRequest := FixUpdateRuntimeRequest(id, inputGQL)
	runtime := graphql.RuntimeExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, updateRequest, &runtime)
	return runtime, err
}

func RequestClientCredentialsForRuntime(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.RuntimeSystemAuth {
	req := FixRequestClientCredentialsForRuntime(id)
	systemAuth := graphql.RuntimeSystemAuth{}

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &systemAuth)
	require.NoError(t, err)
	return systemAuth
}

func UnregisterRuntime(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	if id == "" {
		return
	}
	delReq := FixUnregisterRuntimeRequest(id)

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, delReq, nil)
	require.NoError(t, err)
}

func UnregisterRuntimeWithoutTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) {
	if id == "" {
		return
	}
	delReq := FixUnregisterRuntimeRequest(id)

	err := testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, delReq, nil)
	require.NoError(t, err)
}

func CleanupRuntime(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string, rtm *graphql.RuntimeExt) {
	if rtm == nil || rtm.ID == "" {
		return
	}
	delReq := FixUnregisterRuntimeRequest(rtm.ID)

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, delReq, nil)
	assertions.AssertNoErrorForOtherThanNotFound(t, err)
}

func CleanupRuntimeWithoutTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, rtm *graphql.RuntimeExt) {
	if rtm == nil || rtm.ID == "" {
		return
	}
	delReq := FixUnregisterRuntimeRequest(rtm.ID)

	err := testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, delReq, nil)
	assertions.AssertNoErrorForOtherThanNotFound(t, err)
}

func GetRuntime(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.RuntimeExt {
	req := FixGetRuntimeRequest(id)
	runtime := graphql.RuntimeExt{}

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &runtime)
	require.NoError(t, err)
	return runtime
}

func ListRuntimes(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string) graphql.RuntimePageExt {
	runtimesPage := graphql.RuntimePageExt{}
	queryReq := FixGetRuntimesRequestWithPagination()
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, queryReq, &runtimesPage)
	require.NoError(t, err)
	return runtimesPage
}

func SetRuntimeLabel(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, runtimeID string, labelKey string, labelValue interface{}) *graphql.Label {
	setLabelRequest := FixSetRuntimeLabelRequest(runtimeID, labelKey, labelValue)
	label := graphql.Label{}
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, setLabelRequest, &label)
	require.NoError(t, err)

	return &label
}

func DeleteSystemAuthForRuntime(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) {
	req := FixDeleteSystemAuthForRuntimeRequest(id)
	err := testctx.Tc.RunOperation(ctx, gqlClient, req, nil)
	require.NoError(t, err)
}

func RequestOneTimeTokenForRuntime(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) graphql.OneTimeTokenForRuntimeExt {
	tokenRequest := FixRequestOneTimeTokenForRuntime(id)
	token := graphql.OneTimeTokenForRuntimeExt{}
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, tokenRequest, &token)
	require.NoError(t, err)
	require.NotEmpty(t, token.ConnectorURL)
	require.NotEmpty(t, token.Token)
	require.NotEmpty(t, token.Raw)
	require.NotEmpty(t, token.RawEncoded)
	return token
}
