package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
)

func DeleteBundleInstanceAuth(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.BundleInstanceAuth {
	if id == "" {
		return graphql.BundleInstanceAuth{}
	}
	deleteRequest := FixDeleteBundleInstanceAuthRequest(id)
	output := graphql.BundleInstanceAuth{}

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, deleteRequest, &output)
	if err != nil && strings.Contains(err.Error(), "Object not found [object=bundleInstanceAuth]") {
		return output
	}
	require.NoError(t, err)
	return output
}
