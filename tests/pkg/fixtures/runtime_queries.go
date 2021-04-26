package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func RegisterRuntimeFromInputWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, input *graphql.RuntimeInput) graphql.RuntimeExt {
	inputGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(*input)
	require.NoError(t, err)

	registerRuntimeRequest := FixRegisterRuntimeRequest(inputGQL)
	var runtime graphql.RuntimeExt

	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, registerRuntimeRequest, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)
	return runtime
}

func RequestClientCredentialsForRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.RuntimeSystemAuth {
	req := FixRequestClientCredentialsForRuntime(id)
	systemAuth := graphql.RuntimeSystemAuth{}

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &systemAuth)
	require.NoError(t, err)
	return systemAuth
}

func UnregisterRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	if id == "" {
		return
	}
	delReq := FixUnregisterRuntimeRequest(id)

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, delReq, nil)
	require.NoError(t, err)
}

func UnregisterGracefullyRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	if id == "" {
		return
	}
	delReq := FixUnregisterRuntimeRequest(id)

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, delReq, nil)

	if err != nil && !strings.Contains(err.Error(), "Object not found [object=runtime]") {
		require.NoError(t, err)
	}
}

func GetRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.RuntimeExt {
	req := FixGetRuntimeRequest(id)
	runtime := graphql.RuntimeExt{}

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &runtime)
	require.NoError(t, err)
	return runtime
}

func ListRuntimes(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string) graphql.RuntimePageExt {
	runtimesPage := graphql.RuntimePageExt{}
	queryReq := FixGetRuntimesRequestWithPagination()
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, queryReq, &runtimesPage)
	require.NoError(t, err)
	return runtimesPage
}

func SetRuntimeLabel(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, runtimeID string, labelKey string, labelValue interface{}) *graphql.Label {
	setLabelRequest := FixSetRuntimeLabelRequest(runtimeID, labelKey, labelValue)
	label := graphql.Label{}
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, setLabelRequest, &label)
	require.NoError(t, err)

	return &label
}

func DeleteSystemAuthForRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string) {
	req := FixDeleteSystemAuthForRuntimeRequest(id)
	err := testctx.Tc.RunOperation(ctx, gqlClient, req, nil)
	require.NoError(t, err)
}

func RequestOneTimeTokenForRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) graphql.OneTimeTokenForRuntimeExt {
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
