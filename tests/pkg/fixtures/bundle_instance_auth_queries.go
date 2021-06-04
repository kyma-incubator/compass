package fixtures

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func DeleteBundleInstanceAuth(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.BundleInstanceAuth {
	if id == "" {
		return graphql.BundleInstanceAuth{}
	}
	deleteRequest := FixDeleteBundleInstanceAuthRequest(id)
	output := graphql.BundleInstanceAuth{}

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, deleteRequest, &output)
	require.NoError(t, err)
	return output
}
